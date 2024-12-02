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
			expectFailuresRange: []int{1, 999}, // should at least pass some acquires
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
			expectFailuresRange: []int{1, 999},
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
			for i := 0; i < tt.goroutines; i++ {
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

			wg.Wait()
			// sanity check
			assert.Equal(t, 0, permit.Count())
			assert.Equal(t, tt.capacity[len(tt.capacity)-1], permit.Quota())

			// expect failures in range
			expectFailureMin := tt.expectFailuresRange[0]
			expectFailureMax := tt.expectFailuresRange[1]
			assert.GreaterOrEqual(t, int(failures.Load()), expectFailureMin)
			assert.LessOrEqual(t, int(failures.Load()), expectFailureMax)
		})
	}
}
