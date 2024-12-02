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

type resizablePermit struct {
	sem semaphore.Semaphore
}

// NewResizablePermit creates a dynamic permit that's resizable
func NewResizablePermit(initCount int) *resizablePermit {
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

func (p *resizablePermit) Release() {
	p.sem.Release(1)
}

func (p *resizablePermit) Quota() int {
	return p.sem.GetLimit()
}

func (p *resizablePermit) SetQuota(c int) {
	p.sem.SetLimit(c)
}

// Count returns the number of permits available
func (p *resizablePermit) Count() int {
	return p.sem.GetCount()
}
