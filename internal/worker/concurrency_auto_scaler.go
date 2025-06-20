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
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/metrics"
)

const (
	defaultAutoScalerUpdateTick   = time.Second
	numberOfPollsInRollingAverage = 5 // control flakiness of the auto scaler signal

	autoScalerEventPollerScaleUp              autoScalerEvent = "poller-limit-scale-up"
	autoScalerEventPollerScaleDown            autoScalerEvent = "poller-limit-scale-down"
	autoScalerEventPollerSkipUpdateCooldown   autoScalerEvent = "poller-limit-skip-update-cooldown"
	autoScalerEventPollerSkipUpdateNoChange   autoScalerEvent = "poller-limit-skip-update-no-change"
	autoScalerEventPollerSkipUpdateNotEnabled autoScalerEvent = "poller-limit-skip-update-not-enabled"
	autoScalerEventEmitMetrics                autoScalerEvent = "auto-scaler-emit-metrics"
	autoScalerEventEnable                     autoScalerEvent = "auto-scaler-enable"
	autoScalerEventDisable                    autoScalerEvent = "auto-scaler-disable"
	autoScalerEventStart                      autoScalerEvent = "auto-scaler-start"
	autoScalerEventStop                       autoScalerEvent = "auto-scaler-stop"
	autoScalerEventLogMsg                     string          = "concurrency auto scaler event"

	metricsEnabled        = "enabled"
	metricsDisabled       = "disabled"
	metricsPollerQuota    = "poller-quota"
	metricsPollerWaitTime = "poller-wait-time"
)

var (
	metricsPollerQuotaBuckets    = tally.MustMakeExponentialValueBuckets(1, 2, 10)                     // 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024
	metricsPollerWaitTimeBuckets = tally.MustMakeExponentialDurationBuckets(1*time.Millisecond, 2, 14) // 1ms, 2ms, 4ms, 8ms, 16ms, 32ms, 64ms, 128ms, 256ms, 512ms, 1024ms, 2048ms, 4096ms, 8192ms, 16384ms
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

		// state of autoscaler
		lock    sync.RWMutex
		enabled bool

		// poller
		pollerInitCount          int
		pollerMaxCount           int
		pollerMinCount           int
		pollerWaitTime           *rollingAverage[time.Duration]
		pollerPermitLastUpdate   time.Time
		pollerWaitTimeUpperBound time.Duration
		pollerWaitTimeLowerBound time.Duration
	}

	ConcurrencyAutoScalerInput struct {
		Concurrency              *ConcurrencyLimit
		Cooldown                 time.Duration // cooldown time of update
		Tick                     time.Duration // frequency of update check
		PollerMaxCount           int
		PollerMinCount           int
		PollerWaitTimeUpperBound time.Duration
		PollerWaitTimeLowerBound time.Duration
		Logger                   *zap.Logger
		Scope                    tally.Scope
		Clock                    clockwork.Clock
	}

	autoScalerEvent string
)

func NewConcurrencyAutoScaler(input ConcurrencyAutoScalerInput) *ConcurrencyAutoScaler {
	tick := defaultAutoScalerUpdateTick
	if input.Tick != 0 {
		tick = input.Tick
	}
	return &ConcurrencyAutoScaler{
		shutdownChan:             make(chan struct{}),
		concurrency:              input.Concurrency,
		cooldown:                 input.Cooldown,
		log:                      input.Logger.Named(metrics.ConcurrencyAutoScalerScope),
		scope:                    input.Scope.SubScope(metrics.ConcurrencyAutoScalerScope),
		clock:                    input.Clock,
		updateTick:               tick,
		enabled:                  false, // initial value should be false and is only turned on from auto config hint
		pollerInitCount:          input.Concurrency.PollerPermit.Quota(),
		pollerMaxCount:           input.PollerMaxCount,
		pollerMinCount:           input.PollerMinCount,
		pollerWaitTime:           newRollingAverage[time.Duration](numberOfPollsInRollingAverage),
		pollerWaitTimeUpperBound: input.PollerWaitTimeUpperBound,
		pollerWaitTimeLowerBound: input.PollerWaitTimeLowerBound,
		pollerPermitLastUpdate:   input.Clock.Now(),
	}
}

func (c *ConcurrencyAutoScaler) Start() {
	if c == nil {
		return // no-op if auto scaler is not set
	}
	c.logEvent(autoScalerEventStart)

	c.wg.Add(1)

	go func() {
		defer c.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				c.log.Error("panic in concurrency auto scaler, stopping the auto scaler", zap.Any("error", r))
			}
		}()
		ticker := c.clock.NewTicker(c.updateTick)
		defer ticker.Stop()
		for {
			select {
			case <-c.shutdownChan:
				return
			case <-ticker.Chan():
				c.lock.Lock()
				c.logEvent(autoScalerEventEmitMetrics)
				c.updatePollerPermit()
				c.lock.Unlock()
			}
		}
	}()
}

func (c *ConcurrencyAutoScaler) Stop() {
	if c == nil {
		return // no-op if auto scaler is not set
	}
	c.lock.Lock()
	c.logEvent(autoScalerEventStop)
	c.lock.Unlock()
	close(c.shutdownChan)
	c.wg.Wait()
}

// ProcessPollerHint reads the poller response hint and take actions in a transactional way
// 1. update poller wait time
// 2. enable/disable auto scaler
func (c *ConcurrencyAutoScaler) ProcessPollerHint(hint *shared.AutoConfigHint) {
	if c == nil {
		return // no-op if auto scaler is not set
	}
	c.lock.Lock()
	defer c.lock.Unlock()

	if hint == nil {
		return
	}
	if hint.PollerWaitTimeInMs != nil {
		waitTimeInMs := *hint.PollerWaitTimeInMs
		c.pollerWaitTime.Add(time.Millisecond * time.Duration(waitTimeInMs))
	}

	var shouldEnable bool
	if hint.EnableAutoConfig != nil && *hint.EnableAutoConfig {
		shouldEnable = true
	}
	if shouldEnable != c.enabled { // flag switched
		c.enabled = shouldEnable
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
	if c.enabled {
		c.scope.Counter(metricsEnabled).Inc(1)
	} else {
		c.scope.Counter(metricsDisabled).Inc(1)
	}
	c.scope.Histogram(metricsPollerQuota, metricsPollerQuotaBuckets).RecordValue(float64(c.concurrency.PollerPermit.Quota()))
	c.scope.Histogram(metricsPollerWaitTime, metricsPollerWaitTimeBuckets).RecordDuration(c.pollerWaitTime.Average())
	c.log.Debug(autoScalerEventLogMsg,
		zap.Time("time", c.clock.Now()),
		zap.String("event", string(event)),
		zap.Bool("enabled", c.enabled),
		zap.Int("poller_quota", c.concurrency.PollerPermit.Quota()),
	)
}

func (c *ConcurrencyAutoScaler) updatePollerPermit() {
	if !c.enabled { // skip update if auto scaler is disabled
		c.logEvent(autoScalerEventPollerSkipUpdateNotEnabled)
		return
	}
	updateTime := c.clock.Now()
	if updateTime.Before(c.pollerPermitLastUpdate.Add(c.cooldown)) { // before cooldown
		c.logEvent(autoScalerEventPollerSkipUpdateCooldown)
		return
	}

	var newQuota int
	pollerWaitTime := c.pollerWaitTime.Average()
	if pollerWaitTime < c.pollerWaitTimeLowerBound { // pollers are busy
		newQuota = c.scaleUpPollerPermit(pollerWaitTime)
		c.concurrency.PollerPermit.SetQuota(newQuota)
		c.pollerPermitLastUpdate = updateTime
		c.logEvent(autoScalerEventPollerScaleUp)
	} else if pollerWaitTime > c.pollerWaitTimeUpperBound { // pollers are idle
		newQuota = c.scaleDownPollerPermit(pollerWaitTime)
		c.concurrency.PollerPermit.SetQuota(newQuota)
		c.pollerPermitLastUpdate = updateTime
		c.logEvent(autoScalerEventPollerScaleDown)
	} else {
		c.logEvent(autoScalerEventPollerSkipUpdateNoChange)
		return
	}
}

func (c *ConcurrencyAutoScaler) scaleUpPollerPermit(pollerWaitTime time.Duration) int {
	currentQuota := c.concurrency.PollerPermit.Quota()

	// inverse scaling with edge case of 0 wait time
	// use logrithm to smooth the scaling to avoid drastic change
	newQuota := math.Round(
		float64(currentQuota) * smoothingFunc(c.pollerWaitTimeLowerBound) / smoothingFunc(pollerWaitTime))
	newQuota = math.Max(
		float64(c.pollerMinCount),
		math.Min(float64(c.pollerMaxCount), newQuota),
	)
	return int(newQuota)
}

func (c *ConcurrencyAutoScaler) scaleDownPollerPermit(pollerWaitTime time.Duration) int {
	currentQuota := c.concurrency.PollerPermit.Quota()

	// inverse scaling with edge case of 0 wait time
	// use logrithm to smooth the scaling to avoid drastic change
	newQuota := math.Round(
		float64(currentQuota) * smoothingFunc(c.pollerWaitTimeUpperBound) / smoothingFunc(pollerWaitTime))
	newQuota = math.Max(
		float64(c.pollerMinCount),
		math.Min(float64(c.pollerMaxCount), newQuota),
	)
	return int(newQuota)
}

// smoothingFunc is a log2 function with offset to smooth the scaling and address 0 values
func smoothingFunc(x time.Duration) float64 {
	return math.Log2(2 + float64(x/time.Millisecond))
}

type number interface {
	int64 | float64 | time.Duration
}

type rollingAverage[T number] struct {
	mu     sync.RWMutex
	window []T
	index  int
	sum    T
	count  int
}

func newRollingAverage[T number](capacity int) *rollingAverage[T] {
	return &rollingAverage[T]{
		window: make([]T, capacity),
	}
}

// Add always add positive numbers
func (r *rollingAverage[T]) Add(value T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// no op on zero rolling window
	if len(r.window) == 0 {
		return
	}

	// replace the old value with the new value
	r.sum += value - r.window[r.index]
	r.window[r.index] = value
	r.index++
	r.index %= len(r.window)

	if r.count < len(r.window) {
		r.count++
	}
}

func (r *rollingAverage[T]) Average() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.count == 0 {
		return 0
	}
	return r.sum / T(r.count)
}
