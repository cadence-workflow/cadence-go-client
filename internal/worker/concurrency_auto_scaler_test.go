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
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"go.uber.org/atomic"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

const (
	testTickTime   = 1 * time.Second
	testTimeFormat = time.TimeOnly
)

func createTestConcurrencyAutoScaler(t *testing.T, logger *zap.Logger, clock clockwork.Clock) *ConcurrencyAutoScaler {
	return NewConcurrencyAutoScaler(ConcurrencyAutoScalerInput{
		Concurrency: &ConcurrencyLimit{
			PollerPermit: NewResizablePermit(100),
			TaskPermit:   NewResizablePermit(1000),
		},
		Cooldown:                 2 * testTickTime,
		Tick:                     testTickTime,
		PollerMaxCount:           200,
		PollerMinCount:           50,
		Logger:                   logger,
		Scope:                    tally.NoopScope,
		Clock:                    clock,
		PollerWaitTimeLowerBound: 16 * time.Millisecond,
		PollerWaitTimeUpperBound: 256 * time.Millisecond,
	})
}

func TestConcurrencyAutoScaler(t *testing.T) {

	type eventLog struct {
		eventType   autoScalerEvent
		enabled     bool
		pollerQuota int64
	}

	for _, tt := range []struct {
		name               string
		pollAutoConfigHint []*shared.AutoConfigHint
		expectedEvents     []eventLog
	}{
		{
			"start and stop immediately",
			[]*shared.AutoConfigHint{},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventStop, false, 100},
			},
		},
		{
			"just enough pollers",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(16))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(16))}, // <- tick, no update
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerSkipUpdateNoChange, true, 100},
				{autoScalerEventStop, true, 100},
			},
		},
		{
			"poller slightly idle but no change",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(100))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(100))}, // <- tick, no update
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerSkipUpdateNoChange, true, 100},
				{autoScalerEventStop, true, 100},
			},
		},
		{
			"busy pollers",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(10))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(10))}, // <- tick, scale up
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerScaleUp, true, 116},
				{autoScalerEventStop, true, 116},
			},
		},
		{
			"very busy pollers, scale up to maximum",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(0))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(0))}, // <- tick, scale up significantly to maximum
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerScaleUp, true, 200},
				{autoScalerEventStop, true, 200},
			},
		},
		{
			"busy pollers, scale up, and then scale down slowly",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(10))},    // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(10))},    // <- tick, scale up
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, skip due to cooldown
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, scale down
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, skip due to cooldown
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, scale down further
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerScaleUp, true, 116},
				{autoScalerEventPollerSkipUpdateCooldown, true, 116},
				{autoScalerEventPollerScaleDown, true, 70},
				{autoScalerEventPollerSkipUpdateCooldown, true, 70},
				{autoScalerEventPollerScaleDown, true, 50},
				{autoScalerEventStop, true, 50},
			},
		},
		{
			"pollers, scale up and down multiple times",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(10))},    // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(10))},    // <- tick, scale up
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, skip due to cooldown
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, scale down
				{common.PtrOf(true), common.PtrOf(int64(10))},    // <- tick, skip due to cooldown
				{common.PtrOf(true), common.PtrOf(int64(10))},    // <- tick, scale up again
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, skip due to cooldown
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, scale down again
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerScaleUp, true, 116},
				{autoScalerEventPollerSkipUpdateCooldown, true, 116},
				{autoScalerEventPollerScaleDown, true, 70},
				{autoScalerEventPollerSkipUpdateCooldown, true, 70},
				{autoScalerEventPollerScaleUp, true, 81},
				{autoScalerEventPollerSkipUpdateCooldown, true, 81},
				{autoScalerEventPollerScaleDown, true, 50},
				{autoScalerEventStop, true, 50},
			},
		},
		{
			"idle pollers waiting for tasks",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(1000))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(1000))}, // <- tick, scale down
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerScaleDown, true, 80},
				{autoScalerEventStop, true, 80},
			},
		},
		{
			"idle pollers, scale down to minimum",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(60000))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(60000))}, // <- tick, scale down
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerScaleDown, true, 50},
				{autoScalerEventStop, true, 50},
			},
		},
		{
			"idle pollers but disabled scaling",
			[]*shared.AutoConfigHint{
				{common.PtrOf(false), common.PtrOf(int64(60000))}, // <- tick, in cool down
				{common.PtrOf(false), common.PtrOf(int64(60000))}, // <- tick, no update due to disabled
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100},
				{autoScalerEventStop, false, 100},
			},
		},
		{
			"idle pollers but disabled scaling at a later time",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(1000))},  // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(1000))},  // <- tick, scale down
				{common.PtrOf(false), common.PtrOf(int64(1000))}, // <- disable
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100},
				{autoScalerEventPollerScaleDown, true, 80},
				{autoScalerEventDisable, false, 100},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100},
				{autoScalerEventStop, false, 100},
			},
		},
		{
			"idle pollers and enabled at a later time",
			[]*shared.AutoConfigHint{
				{common.PtrOf(false), common.PtrOf(int64(1000))}, // <- tick, in cool down
				{common.PtrOf(false), common.PtrOf(int64(1000))}, // <- tick, not enabled
				{common.PtrOf(true), common.PtrOf(int64(1000))},  // <- tick, enable scale up
			},
			[]eventLog{
				{autoScalerEventStart, false, 100},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100},
				{autoScalerEventEnable, true, 100},
				{autoScalerEventPollerScaleDown, true, 80},
				{autoScalerEventStop, true, 80},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			core, obs := observer.New(zap.DebugLevel)
			logger := zap.New(core)
			clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
			scaler := createTestConcurrencyAutoScaler(t, logger, clock)

			// mock poller every 1 tick time
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				clock.Sleep(testTickTime / 2) // poll delay by 0.5 unit of time to avoid test flakiness
				for _, hint := range tt.pollAutoConfigHint {
					t.Log("hint process time: ", clock.Now().Format(testTimeFormat))
					for i := 0; i < numberOfPollsInRollingAverage; i++ { // simulate polling multiple times to avoid flakiness
						scaler.ProcessPollerHint(hint)
					}
					clock.Sleep(testTickTime)
				}
			}()

			scaler.Start()
			clock.BlockUntil(2)

			// advance clock
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < len(tt.pollAutoConfigHint)*2+1; i++ {
					clock.Advance(testTickTime / 2)
					time.Sleep(100 * time.Millisecond) // process non-time logic
				}
			}()

			wg.Wait()
			scaler.Stop()

			var actualEvents []eventLog
			for _, event := range obs.FilterMessage(autoScalerEventLogMsg).All() {
				t.Log("event: ", event.ContextMap())

				actualEvents = append(actualEvents, eventLog{
					eventType:   autoScalerEvent(event.ContextMap()["event"].(string)),
					enabled:     event.ContextMap()["enabled"].(bool),
					pollerQuota: event.ContextMap()["poller_quota"].(int64),
				})
			}
			assert.ElementsMatch(t, tt.expectedEvents, actualEvents)
		})
	}
}

func TestRollingAverage(t *testing.T) {
	for _, tt := range []struct {
		name     string
		cap      int
		input    []float64
		expected []float64
	}{
		{
			"cap is 0",
			0,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{0, 0, 0, 0, 0, 0, 0},
		},
		{
			"cap is 1",
			1,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 2, 3, 4, 5, 6, 7},
		},
		{
			"cap is 2",
			2,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 1.5, 2.5, 3.5, 4.5, 5.5, 6.5},
		},
		{
			"cap is 3",
			3,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 1.5, 2, 3, 4, 5, 6},
		},
		{
			"cap is 4",
			4,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 1.5, 2, 2.5, 3.5, 4.5, 5.5},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			r := newRollingAverage[float64](tt.cap)
			for i := range tt.input {
				r.Add(tt.input[i])
				assert.Equal(t, tt.expected[i], r.Average())
			}
		})
	}
}

func TestRollingAverage_Race(t *testing.T) {
	total := 100000
	r := newRollingAverage[float64](total)
	trueSum := atomic.NewFloat64(0)
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
			v := rand.Float64()
			r.Add(v)
			trueSum.Add(v)
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
			r.Average()
		}()
	}

	wg.Wait()

	// sanity check
	assert.InDelta(t, trueSum.Load()/float64(total), r.Average(), 0.001)
}
