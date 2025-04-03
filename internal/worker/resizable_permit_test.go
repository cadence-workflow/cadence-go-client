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
	"sync"
	"testing"
	"time"

	"math/rand"

	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
)

func TestPermit_Simulation(t *testing.T) {
	tests := []struct {
		name                string
		capacity            []int // update every 50ms
		goroutines          int   // each would block on acquiring 1 token for 100-150ms
		maxTestDuration     time.Duration
		expectFailuresRange []int // range of failures, inclusive [min, max]
	}{
		{
			name:                "enough permit, no blocking",
			maxTestDuration:     200 * time.Millisecond, // at most need 150 ms, add 50 ms buffer
			capacity:            []int{10000},
			goroutines:          1000,
			expectFailuresRange: []int{0, 0},
		},
		{
			name:                "not enough permit, blocking but all acquire",
			maxTestDuration:     800 * time.Millisecond, // at most need 150ms * 1000 / 200 = 750ms to acquire all permit
			capacity:            []int{200},
			goroutines:          1000,
			expectFailuresRange: []int{0, 0},
		},
		{
			name:                "not enough permit for some to acquire, fail some",
			maxTestDuration:     250 * time.Millisecond, // at least need 100ms * 1000 / 200 = 500ms to acquire all permit
			capacity:            []int{200},
			goroutines:          1000,
			expectFailuresRange: []int{400, 600}, // should at least pass some acquires
		},
		{
			name:                "not enough permit at beginning but due to capacity change, blocking but all acquire",
			maxTestDuration:     250 * time.Millisecond,
			capacity:            []int{200, 400, 600},
			goroutines:          1000,
			expectFailuresRange: []int{0, 0},
		},
		{
			name:                "enough permit at beginning but due to capacity change, some would fail",
			maxTestDuration:     250 * time.Millisecond,
			capacity:            []int{600, 400, 200},
			goroutines:          1000,
			expectFailuresRange: []int{1, 500}, // the worst case with 200 capacity will at least pass 500 acquires
		},
		{
			name:                "not enough permit for any acquire, fail all",
			maxTestDuration:     300 * time.Millisecond,
			capacity:            []int{0},
			goroutines:          1000,
			expectFailuresRange: []int{1000, 1000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			wg := &sync.WaitGroup{}
			permit := NewResizablePermit(tt.capacity[0])
			wg.Add(1)
			go func() { // update quota every 50ms
				defer wg.Done()
				for i := 1; i < len(tt.capacity); i++ {
					time.Sleep(50 * time.Millisecond)
					permit.SetQuota(tt.capacity[i])
				}
			}()
			failures := atomic.NewInt32(0)
			ctx, cancel := context.WithTimeout(context.Background(), tt.maxTestDuration)
			defer cancel()

			aquireChan := tt.goroutines / 2
			for i := 0; i < tt.goroutines-aquireChan; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := permit.Acquire(ctx); err != nil {
						failures.Inc()
						return
					}
					time.Sleep(time.Duration(100+rand.Intn(50)) * time.Millisecond)
					permit.Release()
				}()
			}
			for i := 0; i < aquireChan; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					permitChan, done := permit.AcquireChan(ctx)
					select {
					case <-permitChan:
						time.Sleep(time.Duration(100+rand.Intn(50)) * time.Millisecond)
						permit.Release()
					case <-ctx.Done():
						failures.Inc()
					}
					done()
				}()
			}

			wg.Wait()
			// sanity check
			assert.Equal(t, 0, permit.Count(), "all permit should be released")
			assert.Equal(t, tt.capacity[len(tt.capacity)-1], permit.Quota())

			// expect failures in range
			expectFailureMin := tt.expectFailuresRange[0]
			expectFailureMax := tt.expectFailuresRange[1]
			assert.GreaterOrEqual(t, int(failures.Load()), expectFailureMin)
			assert.LessOrEqual(t, int(failures.Load()), expectFailureMax)
		})
	}
}

// Test_Permit_Acquire tests the basic acquire functionality
// before each acquire will wait 100ms
func Test_Permit_Acquire(t *testing.T) {

	t.Run("acquire 1 permit", func(t *testing.T) {
		permit := NewResizablePermit(1)
		err := permit.Acquire(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 1, permit.Count())
	})

	t.Run("acquire timeout", func(t *testing.T) {
		permit := NewResizablePermit(1)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		time.Sleep(100 * time.Millisecond)
		err := permit.Acquire(ctx)
		assert.ErrorContains(t, err, "context deadline exceeded")
		assert.Empty(t, permit.Count())
	})

	t.Run("cancel acquire", func(t *testing.T) {
		permit := NewResizablePermit(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := permit.Acquire(ctx)
		assert.ErrorContains(t, err, "canceled")
		assert.Empty(t, permit.Count())
	})

	t.Run("acquire more than quota", func(t *testing.T) {
		permit := NewResizablePermit(1)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		err := permit.Acquire(ctx)
		assert.NoError(t, err)
		err = permit.Acquire(ctx)
		assert.ErrorContains(t, err, "failed to acquire permit")
		assert.Equal(t, 1, permit.Count())
	})
}

func Test_Permit_Release(t *testing.T) {
	for _, tt := range []struct {
		name                    string
		quota, acquire, release int
		expectPanic             bool
	}{
		{"release all acquired permits", 10, 5, 5, false},
		{"release partial acquired permit", 10, 5, 1, false},
		{"release non acquired permit", 10, 5, 0, false},
		{"release more than acquired permit", 10, 5, 10, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			permit := NewResizablePermit(tt.quota)
			for i := 0; i < tt.acquire; i++ {
				err := permit.Acquire(context.Background())
				assert.NoError(t, err)
			}
			releaseOp := func() {
				for i := 0; i < tt.release; i++ {
					permit.Release()
				}
			}

			if tt.expectPanic {
				assert.Panics(t, releaseOp)
			} else {
				assert.NotPanics(t, releaseOp)
				assert.Equal(t, tt.acquire-tt.release, permit.Count())
			}
		})
	}
}

func Test_Permit_AcquireChan(t *testing.T) {
	t.Run("acquire 1 permit", func(t *testing.T) {
		permit := NewResizablePermit(1)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		channel, done := permit.AcquireChan(ctx)
		defer done()
		select {
		case <-channel:
			assert.Equal(t, 1, permit.Count())
		case <-ctx.Done():
			t.Errorf("permit not acquired")
		}
	})

	t.Run("acquire timeout", func(t *testing.T) {
		permit := NewResizablePermit(1)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		time.Sleep(100 * time.Millisecond)
		channel, done := permit.AcquireChan(ctx)
		defer done()
		select {
		case <-channel:
			t.Errorf("permit acquired")
		case <-ctx.Done():
			assert.Empty(t, permit.Count())
		}
	})

	t.Run("cancel acquire", func(t *testing.T) {
		permit := NewResizablePermit(1)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		channel, done := permit.AcquireChan(ctx)
		defer done()
		select {
		case <-channel:
			t.Errorf("permit acquired")
		case <-ctx.Done():
			assert.Empty(t, permit.Count())
		}
	})

	t.Run("acquire more than quota", func(t *testing.T) {
		permit := NewResizablePermit(4)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		for i := 0; i < 10; i++ {
			channel, done := permit.AcquireChan(ctx)
			select {
			case <-channel:
			case <-ctx.Done():
			}
			done()
		}

		assert.Equal(t, 4, permit.Count())
	})
}
