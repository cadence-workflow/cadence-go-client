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

package metrics

import (
	"time"

	"github.com/uber-go/tally"
)

// EmitLatency records latency to both timer and histogram.
// This helper function supports the dual-emit pattern during migration from timers to histograms.
// Both timer and histogram metrics are ALWAYS emitted to support gradual dashboard/alert migration.
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
	// Emit timer (existing behavior, will be deprecated in future release)
	scope.Timer(name).Record(latency)

	// Emit histogram (new, will replace timers in future release)
	RecordTimer(scope, name, latency, buckets)
}

// DualStopwatch is a stopwatch that emits to both timer and histogram.
// This supports the dual-emit pattern during migration from timers to histograms.
// Both timer and histogram metrics are ALWAYS recorded to support gradual dashboard/alert migration.
type DualStopwatch struct {
	timerSW     tally.Stopwatch
	histogramSW tally.Stopwatch
}

// StartLatency creates a stopwatch that emits to both timer and histogram.
// Call .Stop() on the returned stopwatch to record the duration.
// Both timer and histogram metrics are ALWAYS emitted.
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
	return &DualStopwatch{
		timerSW:     scope.Timer(name).Start(),
		histogramSW: StartTimer(scope, name, buckets),
	}
}

// Stop records the elapsed time to both timer and histogram.
func (sw *DualStopwatch) Stop() {
	sw.timerSW.Stop()
	sw.histogramSW.Stop()
}
