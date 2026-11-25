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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
)

func TestHistogramBuckets(t *testing.T) {
	tests := []struct {
		name      string
		histogram SubsettableHistogram
		wantMin   int
		wantMax   int
	}{
		{
			name:      "Default1ms100s",
			histogram: Default1ms100s,
			wantMin:   80,
			wantMax:   85,
		},
		{
			name:      "Low1ms100s",
			histogram: Low1ms100s,
			wantMin:   40,
			wantMax:   45,
		},
		{
			name:      "High1ms24h",
			histogram: High1ms24h,
			wantMin:   112,
			wantMax:   115,
		},
		{
			name:      "Mid1ms24h",
			histogram: Mid1ms24h,
			wantMin:   56,
			wantMax:   60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buckets := tt.histogram.Buckets()

			// Check bucket count
			if len(buckets) < tt.wantMin || len(buckets) > tt.wantMax {
				t.Errorf("got %d buckets, want between %d and %d", len(buckets), tt.wantMin, tt.wantMax)
			}

			// First bucket should always be 0
			if buckets[0] != 0 {
				t.Errorf("first bucket should be 0, got %v", buckets[0])
			}

			// Buckets should be monotonically increasing
			for i := 1; i < len(buckets); i++ {
				if buckets[i] <= buckets[i-1] {
					t.Errorf("buckets not monotonic at index %d: %v <= %v",
						i, buckets[i], buckets[i-1])
				}
			}
		})
	}
}

func TestHistogramValues(t *testing.T) {
	t.Run("Default1ms100s values", func(t *testing.T) {
		buckets := Default1ms100s.Buckets()
		// Check a few key values
		assert.Equal(t, time.Duration(0), buckets[0], "first bucket should be 0")
		assert.Equal(t, time.Millisecond, buckets[1], "second bucket should be 1ms")

		// Last bucket should exceed 100s by at least 2x
		last := buckets[len(buckets)-1]
		assert.True(t, last >= 200*time.Second, "last bucket should be >= 200s, got %v", last)

		// Print buckets for visual inspection (in verbose mode)
		if testing.Verbose() {
			t.Logf("Default1ms100s buckets (%d total):", len(buckets))
			printBuckets(t, buckets, 4) // scale=2 means 4 buckets per doubling
		}
	})

	t.Run("Low1ms100s values", func(t *testing.T) {
		buckets := Low1ms100s.Buckets()
		assert.Equal(t, time.Duration(0), buckets[0])
		assert.Equal(t, time.Millisecond, buckets[1])

		if testing.Verbose() {
			t.Logf("Low1ms100s buckets (%d total):", len(buckets))
			printBuckets(t, buckets, 2) // scale=1 means 2 buckets per doubling
		}
	})
}

func TestSubset(t *testing.T) {
	tests := []struct {
		name     string
		original SubsettableHistogram
		newScale int
	}{
		{
			name:     "Default1ms100s scale 2 to 1",
			original: Default1ms100s,
			newScale: 1,
		},
		{
			name:     "Default1ms100s scale 2 to 0",
			original: Default1ms100s,
			newScale: 0,
		},
		{
			name:     "High1ms24h scale 2 to 1",
			original: High1ms24h,
			newScale: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subset := tt.original.subsetTo(tt.newScale)
			originalBuckets := tt.original.Buckets()
			subsetBuckets := subset.Buckets()

			// Subset should have fewer buckets
			assert.Less(t, len(subsetBuckets), len(originalBuckets),
				"subset should have fewer buckets")

			// First bucket should be the same
			assert.Equal(t, originalBuckets[0], subsetBuckets[0],
				"first bucket should match")

			// Last bucket should be the same
			assert.Equal(t, originalBuckets[len(originalBuckets)-1],
				subsetBuckets[len(subsetBuckets)-1],
				"last bucket should match")

			// All subset buckets should exist in original
			for _, sb := range subsetBuckets {
				found := false
				for _, ob := range originalBuckets {
					if sb == ob {
						found = true
						break
					}
				}
				assert.True(t, found, "subset bucket %v not found in original", sb)
			}

			if testing.Verbose() {
				t.Logf("Original: %d buckets, Subset: %d buckets",
					len(originalBuckets), len(subsetBuckets))
			}
		})
	}
}

func TestSubsetPanics(t *testing.T) {
	t.Run("scale not less than current", func(t *testing.T) {
		assert.Panics(t, func() {
			Default1ms100s.subsetTo(2) // same scale
		})
		assert.Panics(t, func() {
			Default1ms100s.subsetTo(3) // higher scale
		})
	})

	t.Run("negative scale", func(t *testing.T) {
		assert.Panics(t, func() {
			Default1ms100s.subsetTo(-1)
		})
	})
}

func TestMakeSubsettableHistogramPanics(t *testing.T) {
	t.Run("invalid scale", func(t *testing.T) {
		assert.Panics(t, func() {
			makeSubsettableHistogram(-1, time.Millisecond, 100*time.Second, 80)
		})
		assert.Panics(t, func() {
			makeSubsettableHistogram(4, time.Millisecond, 100*time.Second, 80)
		})
	})

	t.Run("invalid start", func(t *testing.T) {
		assert.Panics(t, func() {
			makeSubsettableHistogram(2, 0, 100*time.Second, 80)
		})
		assert.Panics(t, func() {
			makeSubsettableHistogram(2, -time.Millisecond, 100*time.Second, 80)
		})
	})

	t.Run("start >= end", func(t *testing.T) {
		assert.Panics(t, func() {
			makeSubsettableHistogram(2, 100*time.Second, 100*time.Second, 80)
		})
		assert.Panics(t, func() {
			makeSubsettableHistogram(2, 200*time.Second, 100*time.Second, 80)
		})
	})

	t.Run("invalid length", func(t *testing.T) {
		assert.Panics(t, func() {
			makeSubsettableHistogram(2, time.Millisecond, 100*time.Second, 5) // too small
		})
		assert.Panics(t, func() {
			makeSubsettableHistogram(2, time.Millisecond, 100*time.Second, 200) // too large
		})
	})

	t.Run("length not aligned to power of 2", func(t *testing.T) {
		assert.Panics(t, func() {
			// scale=2 needs length divisible by 4
			makeSubsettableHistogram(2, time.Millisecond, 100*time.Second, 81)
		})
	})

	t.Run("not enough buckets for range", func(t *testing.T) {
		assert.Panics(t, func() {
			// 12 buckets won't cover 1ms to 100s
			makeSubsettableHistogram(2, time.Millisecond, 100*time.Second, 12)
		})
	})
}

func TestRecordHistogram(t *testing.T) {
	scope := tally.NewTestScope("test", nil)

	RecordHistogram(scope, "test-metric", 50*time.Millisecond, Default1ms100s)

	snapshot := scope.Snapshot()
	histograms := snapshot.Histograms()

	// Should have recorded to histogram with _ns suffix
	hist, ok := histograms["test.test-metric_ns+"]
	assert.True(t, ok, "histogram should exist with _ns suffix")
	assert.NotNil(t, hist)

	// Should have values in buckets
	assert.NotNil(t, hist.Durations())
}

func TestStartHistogram(t *testing.T) {
	scope := tally.NewTestScope("test", nil)

	sw := StartHistogram(scope, "test-metric", Default1ms100s)
	time.Sleep(10 * time.Millisecond)
	sw.Stop()

	snapshot := scope.Snapshot()
	histograms := snapshot.Histograms()

	// Should have recorded to histogram
	hist, ok := histograms["test.test-metric_ns+"]
	assert.True(t, ok, "histogram should exist")
	assert.NotNil(t, hist)
}

// printBuckets prints histogram buckets in a readable format, grouped by rows
func printBuckets(t *testing.T, buckets tally.DurationBuckets, width int) {
	t.Logf("[%v]", buckets[0]) // zero value on its own row
	for rowStart := 1; rowStart < len(buckets); rowStart += width {
		end := rowStart + width
		if end > len(buckets) {
			end = len(buckets)
		}
		row := buckets[rowStart:end]
		t.Logf("%v", formatDurations(row))
	}
}

// formatDurations formats a slice of durations as a string
func formatDurations(durations []time.Duration) string {
	var parts []string
	for _, d := range durations {
		parts = append(parts, d.String())
	}
	return "[" + strings.Join(parts, " ") + "]"
}
