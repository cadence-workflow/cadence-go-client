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
)

type channelPermit struct {
	channel chan struct{}
}

// NewChannelPermit creates a static permit that's not resizable
func NewChannelPermit(count int) ChannelPermit {
	channel := make(chan struct{}, count)
	for i := 0; i < count; i++ {
		channel <- struct{}{}
	}
	return &channelPermit{channel: channel}
}

func (p *channelPermit) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to acquire permit before context is done")
	case p.channel <- struct{}{}:
		return nil
	}
}

// AcquireChan returns a permit ready channel
func (p *channelPermit) GetChan() <-chan struct{} {
	return p.channel
}

func (p *channelPermit) Release() {
	p.channel <- struct{}{}
}

// Count returns the number of permits available
func (p *channelPermit) Count() int {
	return len(p.channel)
}

func (p *channelPermit) Quota() int {
	return cap(p.channel)
}

// SetQuota on static permit doesn't take effect
func (p *channelPermit) SetQuota(_ int) {
}
