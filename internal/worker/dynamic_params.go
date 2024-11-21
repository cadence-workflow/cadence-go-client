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

package worker

import (
	"context"
	"fmt"

	"github.com/marusama/semaphore/v2"
)

var _ Permit = (*permit)(nil)

// Synchronization contains synchronization primitives for dynamic configuration.
type DynamicParams struct {
	PollerPermit Permit // controls concurrency of pollers
	TaskPermit   Permit // controlls concurrency of task processings
}

// Permit is an adaptive
type Permit interface {
	Acquire(context.Context, int) error
	AcquireChan(context.Context) <-chan struct{}
	Quota() int
	SetQuota(int)
	Count() int
	Release(count int)
}

type permit struct {
	sem semaphore.Semaphore
}

// NewPermit creates a dynamic permit that's resizable
func NewPermit(initCount int) Permit {
	return &permit{sem: semaphore.New(initCount)}
}

// Acquire is blocking until a permit is acquired or returns error after context is done
func (p *permit) Acquire(ctx context.Context, count int) error {
	if err := p.sem.Acquire(ctx, count); err != nil {
		return fmt.Errorf("failed to acquire permit before context is done: %w", err)
	}
	return nil
}

// AcquireChan returns a permit ready channel. It's closed then permit is acquired
func (p *permit) AcquireChan(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		if err := p.sem.Acquire(ctx, 1); err != nil {
			close(ch)
			return
		}
		ch <- struct{}{}
		close(ch)
	}()
	return ch
}

func (p *permit) Release(count int) {
	p.sem.Release(count)
}

func (p *permit) Quota() int {
	return p.sem.GetLimit()
}

func (p *permit) SetQuota(c int) {
	p.sem.SetLimit(c)
}

func (p *permit) Count() int {
	return p.sem.GetCount()
}
