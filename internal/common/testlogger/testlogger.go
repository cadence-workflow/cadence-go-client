// Copyright (c) 2017-2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package testlogger

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"go.uber.org/cadence/internal/common"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

type TestingT interface {
	zaptest.TestingT
	Cleanup(func()) // not currently part of zaptest.TestingT
}

// NewZap makes a new test-oriented logger that prevents bad-lifecycle logs from failing tests.
func NewZap(t TestingT) *zap.Logger {
	/*
		HORRIBLE HACK due to async shutdown, both in our code and in libraries:
		normally, logs produced after a test finishes will *intentionally* fail the test and/or
		cause data to race on the test's internal `t.done` field.

		that's a good thing, it reveals possibly-dangerously-flawed lifecycle management.

		unfortunately some of our code and some libraries do not have good lifecycle management,
		and this cannot easily be patched from the outside.

		so this logger cheats: after a test completes, it logs to stderr rather than TestingT.
		EVERY ONE of these logs is bad and we should not produce them, but it's causing many
		otherwise-useful tests to be flaky, and that's a larger interruption than is useful.
	*/
	logAfterComplete, err := zap.NewDevelopment()
	require.NoError(t, err, "could not build a fallback zap logger")
	replaced := &fallbackTestCore{
		mu:        &sync.RWMutex{},
		t:         t,
		fallback:  logAfterComplete.Core(),
		testing:   zaptest.NewLogger(t).Core(),
		completed: common.PtrOf(false),
	}

	t.Cleanup(replaced.UseFallback) // switch to fallback before ending the test

	return zap.New(replaced)
}

// NewObserved makes a new test logger that both logs to `t` and collects logged
// events for asserting in tests.
func NewObserved(t TestingT) (*zap.Logger, *observer.ObservedLogs) {
	obsCore, obs := observer.New(zapcore.DebugLevel)
	z := NewZap(t)
	z = z.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, obsCore)
	}))
	return z, obs
}

type fallbackTestCore struct {
	mu        *sync.RWMutex
	t         TestingT
	fallback  zapcore.Core
	testing   zapcore.Core
	completed *bool
}

var _ zapcore.Core = (*fallbackTestCore)(nil)

func (f *fallbackTestCore) UseFallback() {
	f.mu.Lock()
	defer f.mu.Unlock()
	*f.completed = true
}

func (f *fallbackTestCore) Enabled(level zapcore.Level) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.completed != nil && *f.completed {
		return f.fallback.Enabled(level)
	}
	return f.testing.Enabled(level)
}

func (f *fallbackTestCore) With(fields []zapcore.Field) zapcore.Core {
	f.mu.Lock()
	defer f.mu.Unlock()

	// need to copy and defer, else the returned core will be used at an
	// arbitrarily later point in time, possibly after the test has completed.
	return &fallbackTestCore{
		mu:        f.mu,
		t:         f.t,
		fallback:  f.fallback.With(fields),
		testing:   f.testing.With(fields),
		completed: f.completed,
	}
}

func (f *fallbackTestCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()
	// see other Check impls, all look similar.
	// this defers the "where to log" decision to Write, as `f` is the core that will write.
	if f.fallback.Enabled(entry.Level) {
		return checked.AddCore(entry, f)
	}
	return checked // do not add any cores
}

func (f *fallbackTestCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if common.ValueFromPtr(f.completed) {
		entry.Message = fmt.Sprintf("COULD FAIL TEST %q, logged too late: %v", f.t.Name(), entry.Message)

		hasStack := slices.ContainsFunc(fields, func(field zapcore.Field) bool {
			// no specific stack-trace type, so just look for probable fields.
			return strings.Contains(strings.ToLower(field.Key), "stack")
		})
		if !hasStack {
			fields = append(fields, zap.Stack("log_stack"))
		}
		return f.fallback.Write(entry, fields)
	}
	return f.testing.Write(entry, fields)
}

func (f *fallbackTestCore) Sync() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if common.ValueFromPtr(f.completed) {
		return f.fallback.Sync()
	}
	return f.testing.Sync()
}
