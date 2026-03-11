// Copyright (c) 2017 Uber Technologies, Inc.
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

// Package metrics provides utilities for emitting latency metrics during the timer→histogram migration.
//
// Concurrency Model:
//   - The global currentEmitMode variable is accessed via atomic operations (atomic.LoadInt32/StoreInt32)
//     to ensure thread-safety when multiple workers are created concurrently.
//   - SetEmitMode() is called during worker initialization (newAggregatedWorker) and may be called
//     concurrently by multiple goroutines creating workers.
//   - getCurrentEmitMode() is called on every metric emission from worker pollers and activity/decision handlers,
//     which run concurrently across many goroutines.
//   - Using atomics ensures proper memory ordering and prevents data races without the overhead of a mutex.
package metrics

import (
	"sync/atomic"
	"time"

	"github.com/uber-go/tally"
)

// MetricEmitMode controls which metrics are emitted for latency measurements.
type MetricEmitMode int

const (
	// EmitModeUnset indicates the mode has not been explicitly set (will use default)
	EmitModeUnset MetricEmitMode = iota
	// EmitTimersOnly emits only timer metrics (legacy OSS behavior)
	EmitTimersOnly
	// EmitBoth emits both timer and histogram metrics (default for migration)
	EmitBoth
	// EmitHistogramsOnly emits only histogram metrics (post-migration)
	EmitHistogramsOnly
)

// currentEmitMode is the active emission mode. Default is EmitBoth for migration.
// This should be set during application initialization (e.g., in init() or before starting workers).
// It should NOT be changed dynamically after workers have started.
// Access via atomic operations for thread-safety.
var currentEmitMode = int32(EmitBoth)

// SetEmitMode configures the metric emission strategy.
// This should be called during application initialization, before any metrics are emitted.
// Alternatively, use WorkerOptions.FeatureFlags.MetricEmitMode.
// This function is safe for concurrent use.
//
// Example usage:
//
//	import "go.uber.org/cadence/internal/common/metrics"
//
//	func init() {
//	    // To use only timers (legacy behavior)
//	    metrics.SetEmitMode(metrics.EmitTimersOnly)
//	}
func SetEmitMode(mode MetricEmitMode) {
	atomic.StoreInt32(&currentEmitMode, int32(mode))
}

// getCurrentEmitMode returns the current emission mode.
// This function is safe for concurrent use.
func getCurrentEmitMode() MetricEmitMode {
	return MetricEmitMode(atomic.LoadInt32(&currentEmitMode))
}

// GetCurrentEmitMode returns the current emission mode (exported for testing).
// This function is safe for concurrent use.
func GetCurrentEmitMode() MetricEmitMode {
	return MetricEmitMode(atomic.LoadInt32(&currentEmitMode))
}

// EmitLatency records latency based on the current emit mode setting.
// This helper function supports flexible metric emission during timer→histogram migration.
//
// Parameters:
//   - scope: The tally scope to emit metrics to
//   - name: The metric name (without suffix)
//   - latency: The duration to record
//   - buckets: The histogram bucket configuration to use
//
// Example:
//
//	EmitLatency(scope, "decision-poll-latency", duration, Default1ms100s)
func EmitLatency(scope tally.Scope, name string, latency time.Duration, buckets SubsettableHistogram) {
	switch getCurrentEmitMode() {
	case EmitTimersOnly:
		scope.Timer(name).Record(latency)
	case EmitBoth:
		scope.Timer(name).Record(latency)
		RecordHistogram(scope, name, latency, buckets)
	case EmitHistogramsOnly:
		RecordHistogram(scope, name, latency, buckets)
	}
}

// DualStopwatch is a stopwatch that emits metrics based on the current emit mode setting.
// This supports flexible metric emission during timer→histogram migration.
type DualStopwatch struct {
	timerSW     tally.Stopwatch
	histogramSW tally.Stopwatch
	mode        MetricEmitMode
}

// StartLatency creates a stopwatch that emits based on current emit mode setting.
// Call .Stop() on the returned stopwatch to record the duration.
//
// Parameters:
//   - scope: The tally scope to emit metrics to
//   - name: The metric name (without suffix)
//   - buckets: The histogram bucket configuration to use
//
// Example:
//
//	sw := StartLatency(scope, "activity-execution-latency", Default1ms100s)
//	// ... do work ...
//	sw.Stop()
func StartLatency(scope tally.Scope, name string, buckets SubsettableHistogram) *DualStopwatch {
	mode := getCurrentEmitMode()
	sw := &DualStopwatch{mode: mode}

	switch mode {
	case EmitTimersOnly:
		sw.timerSW = scope.Timer(name).Start()
	case EmitBoth:
		sw.timerSW = scope.Timer(name).Start()
		sw.histogramSW = StartHistogram(scope, name, buckets)
	case EmitHistogramsOnly:
		sw.histogramSW = StartHistogram(scope, name, buckets)
	}

	return sw
}

// Stop records the elapsed time based on current emit mode setting.
func (sw *DualStopwatch) Stop() {
	switch sw.mode {
	case EmitTimersOnly:
		sw.timerSW.Stop()
	case EmitBoth:
		sw.timerSW.Stop()
		sw.histogramSW.Stop()
	case EmitHistogramsOnly:
		sw.histogramSW.Stop()
	}
}
