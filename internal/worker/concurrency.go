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
	"sync"
	"time"

	"github.com/marusama/semaphore/v2"
)

var _ Permit = (*permit)(nil)

// ConcurrencyLimit contains synchronization primitives for dynamically controlling the concurrencies in workers
type ConcurrencyLimit struct {
	PollerPermit Permit // controls concurrency of pollers
	TaskPermit   Permit // controlls concurrency of task processings
}

// Permit is an adaptive permit issuer to control concurrency
type Permit interface {
	Acquire(context.Context, int) error
	AcquireChan(context.Context, *sync.WaitGroup) <-chan struct{}
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
// Remember to call Release(count) to release the permit after usage
func (p *permit) Acquire(ctx context.Context, count int) error {
	if err := p.sem.Acquire(ctx, count); err != nil {
		return fmt.Errorf("failed to acquire permit before context is done: %w", err)
	}
	return nil
}

// AcquireChan returns a permit ready channel. Similar to Acquire, but non-blocking.
// Remember to call Release(1) to release the permit after usage
func (p *permit) AcquireChan(ctx context.Context, wg *sync.WaitGroup) <-chan struct{} {
	ch := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.sem.Acquire(ctx, 1); err != nil {
			return
		}
		select { // try to send to channel, but don't block if listener is gone
		case ch <- struct{}{}:
		case <-time.After(10 * time.Millisecond): // wait time is needed to avoid race condition of channel sending
			p.sem.Release(1)
		}
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
