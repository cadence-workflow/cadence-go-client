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

	"github.com/marusama/semaphore/v2"
)

type resizablePermit struct {
	sem semaphore.Semaphore
}

// NewResizablePermit creates a dynamic permit that's resizable
func NewResizablePermit(initCount int) Permit {
	return &resizablePermit{sem: semaphore.New(initCount)}
}

// Acquire is blocking until a permit is acquired or returns error after context is done
// Remember to call Release(count) to release the permit after usage
func (p *resizablePermit) Acquire(ctx context.Context) error {
	if err := p.sem.Acquire(ctx, 1); err != nil {
		return fmt.Errorf("failed to acquire permit before context is done: %w", err)
	}
	return nil
}

// Release release one permit
func (p *resizablePermit) Release() {
	p.sem.Release(1)
}

// Quota returns the maximum number of permits that can be acquired
func (p *resizablePermit) Quota() int {
	return p.sem.GetLimit()
}

// SetQuota sets the maximum number of permits that can be acquired
func (p *resizablePermit) SetQuota(c int) {
	p.sem.SetLimit(c)
}

// Count returns the number of permits available
func (p *resizablePermit) Count() int {
	return p.sem.GetCount()
}

// AcquireChan returns a channel that could be used to wait for the permit and a close function when done
// Notes:
//  1. avoid goroutine leak by calling the done function
//  2. if the read succeeded, release permit by calling permit.Release()
func (p *resizablePermit) AcquireChan(ctx context.Context) (<-chan struct{}, func()) {
	ctx, cancel := context.WithCancel(ctx)
	pc := &permitChannel{
		p:      p,
		c:      make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
		wg:     &sync.WaitGroup{},
	}
	pc.Start()
	return pc.C(), func() {
		pc.Close()
	}
}

// permitChannel is an implementation to acquire a permit through channel
type permitChannel struct {
	p      Permit
	c      chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func (ch *permitChannel) C() <-chan struct{} {
	return ch.c
}

func (ch *permitChannel) Start() {
	ch.wg.Add(1)
	go func() {
		defer ch.wg.Done()
		if err := ch.p.Acquire(ch.ctx); err != nil {
			return
		}
		// avoid blocking on sending to the channel
		select {
		case ch.c <- struct{}{}:
		case <-ch.ctx.Done(): // release if acquire is successful but fail to send to the channel
			ch.p.Release()
		}
	}()
}

func (ch *permitChannel) Close() {
	ch.cancel()
	ch.wg.Wait()
}
