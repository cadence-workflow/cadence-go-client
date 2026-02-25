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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
)

func TestEmitLatency_DualEmit(t *testing.T) {
	scope := tally.NewTestScope("test", nil)

	EmitLatency(scope, "test-metric", 50*time.Millisecond, Default1ms100s)

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	// Should have timer
	timer, ok := timers["test.test-metric+"]
	assert.True(t, ok, "timer should exist")
	assert.NotNil(t, timer)

	// Should have histogram
	hist, ok := histograms["test.test-metric_ns+"]
	assert.True(t, ok, "histogram should exist")
	assert.NotNil(t, hist)
}

func TestStartLatency_DualEmit(t *testing.T) {
	scope := tally.NewTestScope("test", nil)

	sw := StartLatency(scope, "test-metric", Default1ms100s)
	time.Sleep(10 * time.Millisecond)
	sw.Stop()

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	// Should have timer
	timer, ok := timers["test.test-metric+"]
	assert.True(t, ok, "timer should exist")
	assert.NotNil(t, timer)

	// Should have histogram
	hist, ok := histograms["test.test-metric_ns+"]
	assert.True(t, ok, "histogram should exist")
	assert.NotNil(t, hist)
}

func TestDualStopwatch_MultipleStops(t *testing.T) {
	scope := tally.NewTestScope("test", nil)

	sw := StartLatency(scope, "test-metric", Default1ms100s)
	time.Sleep(10 * time.Millisecond)
	sw.Stop()

	// Multiple stops should not panic
	assert.NotPanics(t, func() {
		sw.Stop()
		sw.Stop()
	})
}

func TestEmitLatency_DifferentHistograms(t *testing.T) {
	tests := []struct {
		name      string
		histogram SubsettableHistogram
	}{
		{
			name:      "Default1ms100s",
			histogram: Default1ms100s,
		},
		{
			name:      "Low1ms100s",
			histogram: Low1ms100s,
		},
		{
			name:      "High1ms24h",
			histogram: High1ms24h,
		},
		{
			name:      "Mid1ms24h",
			histogram: Mid1ms24h,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := tally.NewTestScope("test", nil)

			EmitLatency(scope, "test-metric", 50*time.Millisecond, tt.histogram)

			snapshot := scope.Snapshot()
			histograms := snapshot.Histograms()

			// Should have histogram with correct buckets
			hist, ok := histograms["test.test-metric_ns+"]
			assert.True(t, ok, "histogram should exist")
			assert.NotNil(t, hist)
		})
	}
}
