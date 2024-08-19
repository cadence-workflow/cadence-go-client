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

package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPollerLifeCycle(t *testing.T) {
	lifeCycle := NewLifeCycle()

	// Initially, poller count should be 0
	assert.Equal(t, int32(0), lifeCycle.ReadPollerCount())

	// Start a poller and verify that the count increments
	run1 := lifeCycle.PollerStart("worker-1")
	assert.Equal(t, int32(1), lifeCycle.ReadPollerCount())

	// Start another poller and verify that the count increments again
	run2 := lifeCycle.PollerStart("worker-2")
	assert.Equal(t, int32(2), lifeCycle.ReadPollerCount())

	// Stop the poller twice and verify idempotency
	run1.Stop()
	assert.Equal(t, int32(1), lifeCycle.ReadPollerCount())
	run1.Stop()
	assert.Equal(t, int32(1), lifeCycle.ReadPollerCount())

	run2.Stop()
	assert.Equal(t, int32(0), lifeCycle.ReadPollerCount())
}