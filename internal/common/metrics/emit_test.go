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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
)

func TestEmitLatency_DefaultMode(t *testing.T) {
	// Reset to default (should be EmitBoth)
	SetEmitMode(EmitBoth)

	scope := tally.NewTestScope("test", nil)
	EmitLatency(scope, "test-metric", 50*time.Millisecond, Default1ms100s)

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	// Should have both timer and histogram in default mode
	timer, timerExists := timers["test.test-metric+"]
	hist, histExists := histograms["test.test-metric_ns+"]

	assert.True(t, timerExists, "timer should exist in EmitBoth (default) mode")
	assert.NotNil(t, timer)
	assert.True(t, histExists, "histogram should exist in EmitBoth (default) mode")
	assert.NotNil(t, hist)
}

func TestEmitLatency_DualMode(t *testing.T) {
	// Set to dual-emit mode (this is also the default)
	SetEmitMode(EmitBoth)
	defer SetEmitMode(EmitBoth) // Reset after test

	scope := tally.NewTestScope("test", nil)
	EmitLatency(scope, "test-metric", 50*time.Millisecond, Default1ms100s)

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	// Should have both timer and histogram
	timer, timerExists := timers["test.test-metric+"]
	hist, histExists := histograms["test.test-metric_ns+"]

	assert.True(t, timerExists, "timer should exist in EmitBoth mode")
	assert.NotNil(t, timer)
	assert.True(t, histExists, "histogram should exist in EmitBoth mode")
	assert.NotNil(t, hist)
}

func TestEmitLatency_HistogramOnlyMode(t *testing.T) {
	// Set to histogram-only mode
	SetEmitMode(EmitHistogramsOnly)
	defer SetEmitMode(EmitBoth) // Reset after test

	scope := tally.NewTestScope("test", nil)
	EmitLatency(scope, "test-metric", 50*time.Millisecond, Default1ms100s)

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	// Should have histogram, not timer
	_, timerExists := timers["test.test-metric+"]
	hist, histExists := histograms["test.test-metric_ns+"]

	assert.False(t, timerExists, "timer should not exist in EmitHistogramsOnly mode")
	assert.True(t, histExists, "histogram should exist in EmitHistogramsOnly mode")
	assert.NotNil(t, hist)
}

func TestEmitLatency_TimersOnlyMode(t *testing.T) {
	// Set to timers-only mode (legacy behavior)
	SetEmitMode(EmitTimersOnly)
	defer SetEmitMode(EmitBoth) // Reset after test

	scope := tally.NewTestScope("test", nil)
	EmitLatency(scope, "test-metric", 50*time.Millisecond, Default1ms100s)

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	// Should have timer, not histogram
	timer, timerExists := timers["test.test-metric+"]
	_, histExists := histograms["test.test-metric_ns+"]

	assert.True(t, timerExists, "timer should exist in EmitTimersOnly mode")
	assert.NotNil(t, timer)
	assert.False(t, histExists, "histogram should not exist in EmitTimersOnly mode")
}

func TestStartLatency_DefaultMode(t *testing.T) {
	// Reset to default (should be EmitBoth)
	SetEmitMode(EmitBoth)

	scope := tally.NewTestScope("test", nil)
	sw := StartLatency(scope, "test-metric", Default1ms100s)
	time.Sleep(10 * time.Millisecond)
	sw.Stop()

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	timer, timerExists := timers["test.test-metric+"]
	hist, histExists := histograms["test.test-metric_ns+"]

	assert.True(t, timerExists, "timer should exist in EmitBoth (default) mode")
	assert.NotNil(t, timer)
	assert.True(t, histExists, "histogram should exist in EmitBoth (default) mode")
	assert.NotNil(t, hist)
}

func TestStartLatency_DualMode(t *testing.T) {
	SetEmitMode(EmitBoth)
	defer SetEmitMode(EmitBoth)

	scope := tally.NewTestScope("test", nil)
	sw := StartLatency(scope, "test-metric", Default1ms100s)
	time.Sleep(10 * time.Millisecond)
	sw.Stop()

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	timer, timerExists := timers["test.test-metric+"]
	hist, histExists := histograms["test.test-metric_ns+"]

	assert.True(t, timerExists, "timer should exist in EmitBoth mode")
	assert.NotNil(t, timer)
	assert.True(t, histExists, "histogram should exist in EmitBoth mode")
	assert.NotNil(t, hist)
}

func TestStartLatency_HistogramOnlyMode(t *testing.T) {
	SetEmitMode(EmitHistogramsOnly)
	defer SetEmitMode(EmitBoth)

	scope := tally.NewTestScope("test", nil)
	sw := StartLatency(scope, "test-metric", Default1ms100s)
	time.Sleep(10 * time.Millisecond)
	sw.Stop()

	snapshot := scope.Snapshot()
	timers := snapshot.Timers()
	histograms := snapshot.Histograms()

	_, timerExists := timers["test.test-metric+"]
	hist, histExists := histograms["test.test-metric_ns+"]

	assert.False(t, timerExists, "timer should not exist in EmitHistogramsOnly mode")
	assert.True(t, histExists, "histogram should exist in EmitHistogramsOnly mode")
	assert.NotNil(t, hist)
}

func TestDualStopwatch_MultipleStops(t *testing.T) {
	SetEmitMode(EmitBoth)

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
		{"Default1ms100s", Default1ms100s},
		{"Low1ms100s", Low1ms100s},
		{"High1ms24h", High1ms24h},
		{"Mid1ms24h", Mid1ms24h},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetEmitMode(EmitBoth)
			defer SetEmitMode(EmitBoth)

			scope := tally.NewTestScope("test", nil)
			EmitLatency(scope, "test-metric", 50*time.Millisecond, tt.histogram)

			snapshot := scope.Snapshot()
			timer, timerOk := snapshot.Timers()["test.test-metric+"]
			hist, histOk := snapshot.Histograms()["test.test-metric_ns+"]

			assert.True(t, timerOk, "timer should exist")
			assert.NotNil(t, timer)
			assert.True(t, histOk, "histogram should exist")
			assert.NotNil(t, hist)
		})
	}
}

func TestMetricEmitMode_ConcurrentAccess(t *testing.T) {
	// This test verifies that SetEmitMode and GetCurrentEmitMode are thread-safe
	// and can be called concurrently without race conditions

	const numGoroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines that concurrently read and write the emit mode
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Cycle through different modes
				mode := MetricEmitMode((id + j) % 3)
				switch mode {
				case 0:
					SetEmitMode(EmitTimersOnly)
				case 1:
					SetEmitMode(EmitBoth)
				case 2:
					SetEmitMode(EmitHistogramsOnly)
				}

				// Read the mode (may see any valid value due to concurrent writes)
				currentMode := GetCurrentEmitMode()
				// Verify it's a valid mode
				assert.True(t, currentMode >= EmitModeUnset && currentMode <= EmitHistogramsOnly,
					"Invalid emit mode: %d", currentMode)
			}
		}(i)
	}

	wg.Wait()

	// Clean up: restore to default
	SetEmitMode(EmitBoth)
}
