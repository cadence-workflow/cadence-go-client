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
		name                  string
		capacity              []int // update every 50ms
		goroutines            int   // each would block on acquiring 2-6 tokens for 100ms
		goroutinesAcquireChan int   // each would block using AcquireChan for 100ms
		maxTestDuration       time.Duration
		expectFailures        int
		expectFailuresAtLeast int
	}{
		{
			name:                  "enough permit, no blocking",
			maxTestDuration:       200 * time.Millisecond,
			capacity:              []int{1000},
			goroutines:            100,
			goroutinesAcquireChan: 100,
			expectFailures:        0,
		},
		{
			name:                  "not enough permit, blocking but all acquire",
			maxTestDuration:       1 * time.Second,
			capacity:              []int{200},
			goroutines:            500,
			goroutinesAcquireChan: 500,
			expectFailures:        0,
		},
		{
			name:                  "not enough permit for some to acquire, fail some",
			maxTestDuration:       100 * time.Millisecond,
			capacity:              []int{100},
			goroutines:            500,
			goroutinesAcquireChan: 500,
			expectFailuresAtLeast: 1,
		},
		{
			name:                  "not enough permit at beginning but due to capacity change, blocking but all acquire",
			maxTestDuration:       100 * time.Second,
			capacity:              []int{100, 200, 300},
			goroutines:            500,
			goroutinesAcquireChan: 500,
			expectFailures:        0,
		},
		{
			name:            "not enough permit for any acquire, fail all",
			maxTestDuration: 1 * time.Second,
			capacity:        []int{0},
			goroutines:      1000,
			expectFailures:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := &sync.WaitGroup{}
			permit := NewPermit(tt.capacity[0])
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
					num := rand.Intn(2) + 2
					// num := 1
					if err := permit.Acquire(ctx, num); err != nil {
						failures.Add(1)
						return
					}
					time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
					permit.Release(num)
				}()
			}
			for i := 0; i < tt.goroutinesAcquireChan; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					select {
					case <-permit.AcquireChan(ctx, wg):
						time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
						permit.Release(1)
					case <-ctx.Done():
						failures.Add(1)
					}
				}()
			}

			wg.Wait()
			assert.Equal(t, 0, permit.Count())
			if tt.expectFailuresAtLeast > 0 {
				assert.LessOrEqual(t, tt.expectFailuresAtLeast, int(failures.Load()))
			} else {
				assert.Equal(t, tt.expectFailures, int(failures.Load()))
			}
			assert.Equal(t, tt.capacity[len(tt.capacity)-1], permit.Quota())
		})
	}
}
