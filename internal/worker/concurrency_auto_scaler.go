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
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/shared"
)

const (
	defaultAutoScalerUpdateTick = time.Second
	// concurrencyAutoScalerObservabilityTick = time.Millisecond * 500
	targetPollerWaitTimeInMsLog2  = 4 // 16 ms
	numberOfPollsInRollingAverage = 20

	autoScalerEventPollerUpdate               autoScalerEvent = "update-poller-limit"
	autoScalerEventPollerSkipUpdateCooldown                   = "skip-update-poller-limit-cooldown"
	autoScalerEventPollerSkipUpdateNoChange                   = "skip-update-poller-limit-no-change"
	autoScalerEventPollerSkipUpdateNotEnabled                 = "skip-update-poller-limit-not-enabled"
	autoScalerEventMetrics                                    = "metrics"
	autoScalerEventEnable                                     = "enable"
	autoScalerEventDisable                                    = "disable"
	autoScalerEventStart                                      = "start"
	autoScalerEventStop                                       = "stop"
	autoScalerEventLogMsg                     string          = "concurrency auto scaler event"
	testTimeFormat                            string          = "15:04:05"
)

type (
	ConcurrencyAutoScaler struct {
		shutdownChan chan struct{}
		wg           sync.WaitGroup
		log          *zap.Logger
		scope        tally.Scope
		clock        clockwork.Clock

		concurrency *ConcurrencyLimit
		cooldown    time.Duration
		updateTick  time.Duration

		// enable auto scaler on concurrency or not
		enable atomic.Bool

		// poller
		pollerInitCount        int
		pollerMaxCount         int
		pollerMinCount         int
		pollerWaitTimeInMsLog2 *rollingAverage // log2(pollerWaitTimeInMs+1) for smoothing (ideal value is 0)
		pollerPermitLastUpdate time.Time
	}

	ConcurrencyAutoScalerInput struct {
		Concurrency    *ConcurrencyLimit
		Cooldown       time.Duration // cooldown time of update
		Tick           time.Duration // frequency of update check
		PollerMaxCount int
		PollerMinCount int
		Logger         *zap.Logger
		Scope          tally.Scope
		Clock          clockwork.Clock
	}

	autoScalerEvent string
)

func NewConcurrencyAutoScaler(input ConcurrencyAutoScalerInput) *ConcurrencyAutoScaler {
	tick := defaultAutoScalerUpdateTick
	if input.Tick != 0 {
		tick = input.Tick
	}
	return &ConcurrencyAutoScaler{
		shutdownChan:           make(chan struct{}),
		concurrency:            input.Concurrency,
		cooldown:               input.Cooldown,
		log:                    input.Logger,
		scope:                  input.Scope,
		clock:                  input.Clock,
		updateTick:             tick,
		enable:                 atomic.Bool{}, // initial value should be false and is only turned on from auto config hint
		pollerInitCount:        input.Concurrency.PollerPermit.Quota(),
		pollerMaxCount:         input.PollerMaxCount,
		pollerMinCount:         input.PollerMinCount,
		pollerWaitTimeInMsLog2: newRollingAverage(numberOfPollsInRollingAverage),
		pollerPermitLastUpdate: input.Clock.Now(),
	}
}

func (c *ConcurrencyAutoScaler) Start() {
	c.logEvent(autoScalerEventStart)

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		ticker := c.clock.NewTicker(c.updateTick)
		defer ticker.Stop()
		for {
			select {
			case <-c.shutdownChan:
				return
			case <-ticker.Chan():
				c.logEvent(autoScalerEventMetrics)
				c.updatePollerPermit()
			}
		}
	}()
}

func (c *ConcurrencyAutoScaler) Stop() {
	close(c.shutdownChan)
	c.wg.Wait()
	c.logEvent(autoScalerEventStop)
}

// ProcessPollerHint reads the poller response hint and take actions
// 1. update poller wait time
// 2. enable/disable auto scaler
func (c *ConcurrencyAutoScaler) ProcessPollerHint(hint *shared.AutoConfigHint) {
	if hint == nil {
		c.log.Warn("auto config hint is nil, this results in no action")
		return
	}
	if hint.PollerWaitTimeInMs != nil {
		waitTimeInMs := *hint.PollerWaitTimeInMs
		c.pollerWaitTimeInMsLog2.Add(math.Log2(float64(waitTimeInMs + 1)))
	}

	/*
		Atomically compare and switch the auto scaler enable flag. If auto scaler is turned off, IMMEDIATELY reset the concurrency limits.
	*/
	var shouldEnable bool
	if hint.EnableAutoConfig != nil && *hint.EnableAutoConfig {
		shouldEnable = true
	}
	if switched := c.enable.CompareAndSwap(!shouldEnable, shouldEnable); switched {
		if shouldEnable {
			c.logEvent(autoScalerEventEnable)
		} else {
			c.resetConcurrency()
			c.logEvent(autoScalerEventDisable)
		}
	}
}

// resetConcurrency reset poller quota to the max value. This will be used for gracefully switching the auto scaler off to avoid workers stuck in the wrong state
func (c *ConcurrencyAutoScaler) resetConcurrency() {
	c.concurrency.PollerPermit.SetQuota(c.pollerInitCount)
}

func (c *ConcurrencyAutoScaler) logEvent(event autoScalerEvent) {
	if c.enable.Load() {
		c.scope.Counter("concurrency_auto_scaler.enabled").Inc(1)
	} else {
		c.scope.Counter("concurrency_auto_scaler.disabled").Inc(1)
	}
	c.scope.Gauge("poller_in_action").Update(float64(c.concurrency.PollerPermit.Count()))
	c.scope.Gauge("poller_quota").Update(float64(c.concurrency.PollerPermit.Quota()))
	c.scope.Gauge("poller_wait_time").Update(math.Exp2(c.pollerWaitTimeInMsLog2.Average()))
	c.log.Debug(autoScalerEventLogMsg,
		zap.Time("time", c.clock.Now()),
		zap.String("event", string(event)),
		zap.Bool("enabled", c.enable.Load()),
		zap.Int("poller_quota", c.concurrency.PollerPermit.Quota()),
		zap.Int("poller_in_action", c.concurrency.PollerPermit.Count()),
	)
}

func (c *ConcurrencyAutoScaler) updatePollerPermit() {
	if !c.enable.Load() { // skip update if auto scaler is disabled
		c.logEvent(autoScalerEventPollerSkipUpdateNotEnabled)
		return
	}
	updateTime := c.clock.Now()
	if updateTime.Before(c.pollerPermitLastUpdate.Add(c.cooldown)) { // before cooldown
		c.logEvent(autoScalerEventPollerSkipUpdateCooldown)
		return
	}
	currentQuota := c.concurrency.PollerPermit.Quota()
	newQuota := int(math.Round(float64(currentQuota) * c.pollerWaitTimeInMsLog2.Average() / targetPollerWaitTimeInMsLog2))
	if newQuota < c.pollerMinCount {
		newQuota = c.pollerMinCount
	}
	if newQuota > c.pollerMaxCount {
		newQuota = c.pollerMaxCount
	}
	if newQuota == currentQuota {
		c.logEvent(autoScalerEventPollerSkipUpdateNoChange)
		return
	}
	c.concurrency.PollerPermit.SetQuota(newQuota)
	c.pollerPermitLastUpdate = updateTime
	c.logEvent(autoScalerEventPollerUpdate)
}

type rollingAverage struct {
	mu     sync.RWMutex
	window []float64
	index  int
	sum    float64
	count  int
}

func newRollingAverage(capacity int) *rollingAverage {
	return &rollingAverage{
		window: make([]float64, capacity),
	}
}

// Add always add positive numbers
func (r *rollingAverage) Add(value float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// no op on zero rolling window
	if len(r.window) == 0 {
		return
	}

	// replace the old value with the new value
	r.index %= len(r.window)
	r.sum += value - r.window[r.index]
	r.window[r.index] = value
	r.index++

	if r.count < len(r.window) {
		r.count++
	}
}

func (r *rollingAverage) Average() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.count == 0 {
		return 0
	}
	return r.sum / float64(r.count)
}
