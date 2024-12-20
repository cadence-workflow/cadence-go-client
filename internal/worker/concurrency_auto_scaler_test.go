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
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

const (
	testTickTime = 1 * time.Second
)

func createTestConcurrencyAutoScaler(t *testing.T, logger *zap.Logger, clock clockwork.Clock) *ConcurrencyAutoScaler {
	return NewConcurrencyAutoScaler(ConcurrencyAutoScalerInput{
		Concurrency: &ConcurrencyLimit{
			PollerPermit: NewResizablePermit(100),
			TaskPermit:   NewResizablePermit(1000),
		},
		Cooldown:       2 * testTickTime,
		Tick:           testTickTime,
		PollerMaxCount: 200,
		PollerMinCount: 50,
		Logger:         logger,
		Scope:          tally.NoopScope,
		Clock:          clock,
	})
}

func TestConcurrencyAutoScaler(t *testing.T) {

	type eventLog struct {
		eventType   autoScalerEvent
		enabled     bool
		pollerQuota int64
		time        string
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
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventStop, false, 100, "00:00:00"},
			},
		},
		{
			"just enough pollers",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(15))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(15))}, // <- tick, no update
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventEnable, true, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100, "00:00:01"},
				{autoScalerEventPollerSkipUpdateNoChange, true, 100, "00:00:02"},
				{autoScalerEventStop, true, 100, "00:00:02"},
			},
		},
		{
			"too many pollers",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(10))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(10))}, // <- tick, scale up
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventEnable, true, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100, "00:00:01"},
				{autoScalerEventPollerUpdate, true, 86, "00:00:02"},
				{autoScalerEventStop, true, 86, "00:00:02"},
			},
		},
		{
			"too many pollers, scale down to minimum",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(0))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(0))}, // <- tick, scale down to minimum
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventEnable, true, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100, "00:00:01"},
				{autoScalerEventPollerUpdate, true, 50, "00:00:02"},
				{autoScalerEventStop, true, 50, "00:00:02"},
			},
		},
		{
			"lack pollers",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(100))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(100))}, // <- tick, scale up
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventEnable, true, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100, "00:00:01"},
				{autoScalerEventPollerUpdate, true, 166, "00:00:02"},
				{autoScalerEventStop, true, 166, "00:00:02"},
			},
		},
		{
			"lack pollers, scale up to maximum",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(10000))}, // <- tick, scale up
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventEnable, true, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100, "00:00:01"},
				{autoScalerEventPollerUpdate, true, 200, "00:00:02"},
				{autoScalerEventStop, true, 200, "00:00:02"},
			},
		},
		{
			"lack pollers but disabled",
			[]*shared.AutoConfigHint{
				{common.PtrOf(false), common.PtrOf(int64(100))}, // <- tick, in cool down
				{common.PtrOf(false), common.PtrOf(int64(100))}, // <- tick, scale up
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100, "00:00:01"},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100, "00:00:02"},
				{autoScalerEventStop, false, 100, "00:00:02"},
			},
		},
		{
			"too many pollers but disabled at a later time",
			[]*shared.AutoConfigHint{
				{common.PtrOf(true), common.PtrOf(int64(10))},  // <- tick, in cool down
				{common.PtrOf(true), common.PtrOf(int64(10))},  // <- tick, scale up
				{common.PtrOf(false), common.PtrOf(int64(10))}, // <- disable
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventEnable, true, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateCooldown, true, 100, "00:00:01"},
				{autoScalerEventPollerUpdate, true, 86, "00:00:02"},
				{autoScalerEventDisable, false, 100, "00:00:02"},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100, "00:00:03"},
				{autoScalerEventStop, false, 100, "00:00:03"},
			},
		},
		{
			"lack pollers and enabled at a later time",
			[]*shared.AutoConfigHint{
				{common.PtrOf(false), common.PtrOf(int64(100))}, // <- tick, in cool down
				{common.PtrOf(false), common.PtrOf(int64(100))}, // <- tick, not enabled
				{common.PtrOf(true), common.PtrOf(int64(100))},  // <- tick, enable scale up
			},
			[]eventLog{
				{autoScalerEventStart, false, 100, "00:00:00"},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100, "00:00:01"},
				{autoScalerEventPollerSkipUpdateNotEnabled, false, 100, "00:00:02"},
				{autoScalerEventEnable, true, 100, "00:00:02"},
				{autoScalerEventPollerUpdate, true, 166, "00:00:03"},
				{autoScalerEventStop, true, 166, "00:00:03"},
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
					scaler.ProcessPollerHint(hint)
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
					time.Sleep(10 * time.Millisecond) // process non-time logic
				}
			}()

			wg.Wait()
			scaler.Stop()

			var actualEvents []eventLog
			for _, event := range obs.FilterMessage(autoScalerEventLogMsg).All() {
				if event.ContextMap()["event"] != autoScalerEventMetrics {
					t.Log("event: ", event.ContextMap())
					actualEvents = append(actualEvents, eventLog{
						eventType:   autoScalerEvent(event.ContextMap()["event"].(string)),
						enabled:     event.ContextMap()["enabled"].(bool),
						pollerQuota: event.ContextMap()["poller_quota"].(int64),
						time:        event.ContextMap()["time"].(time.Time).Format(testTimeFormat),
					})
				}
			}
			assert.ElementsMatch(t, tt.expectedEvents, actualEvents)
		})
	}
}

func TestRollingAverage(t *testing.T) {
	for _, tt := range []struct {
		name         string
		cap          int
		addGoroutine int
		input        []float64
		expected     []float64
	}{
		{
			"cap is 0",
			0,
			5,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{0, 0, 0, 0, 0, 0, 0},
		},
		{
			"cap is 1",
			1,
			5,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 2, 3, 4, 5, 6, 7},
		},
		{
			"cap is 2",
			2,
			5,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 1.5, 2.5, 3.5, 4.5, 5.5, 6.5},
		},
		{
			"cap is 3",
			3,
			5,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 1.5, 2, 3, 4, 5, 6},
		},
		{
			"cap is 4",
			4,
			5,
			[]float64{1, 2, 3, 4, 5, 6, 7},
			[]float64{1, 1.5, 2, 2.5, 3.5, 4.5, 5.5},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer goleak.VerifyNone(t)
			r := newRollingAverage(tt.cap)

			doneC := make(chan struct{})

			inputChan := make(chan float64)
			var wg sync.WaitGroup
			for i := 0; i < tt.addGoroutine; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for v := range inputChan {
						r.Add(v)
						doneC <- struct{}{}
					}
				}()
			}

			for i := range tt.input {
				inputChan <- tt.input[i]
				<-doneC
				assert.Equal(t, tt.expected[i], r.Average())
			}
			close(inputChan)
			wg.Wait()
		})
	}
}
