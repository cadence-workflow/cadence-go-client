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
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/uber-go/tally"
)

// SubsettableHistogram is a duration-based histogram that is compatible with OpenTelemetry's
// exponential histogram specification: https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exponentialhistogram
//
// These histograms provide logarithmic precision - finer detail at lower values, coarser at higher values,
// which matches how latency/duration metrics typically behave.
//
// All histogram metric names MUST have a "_ns" suffix to differentiate them from timers.
type SubsettableHistogram struct {
	tallyBuckets tally.DurationBuckets
	scale        int
}

// Pre-defined histograms for common client-side metrics.
// Use these instead of creating custom histograms to encourage consistency.
var (
	// Default1ms100s is the default histogram for most client-side latency metrics.
	// Range: 1ms → ~15 minutes
	// Buckets: 80 (scale=2, ~4 buckets per doubling)
	// Use for: API calls, decision execution, activity execution, poll latencies
	Default1ms100s = makeSubsettableHistogram(2, time.Millisecond, 100*time.Second, 80)

	// Low1ms100s is a half-resolution version of Default1ms100s for high-cardinality metrics.
	// Range: 1ms → ~15 minutes
	// Buckets: 40 (scale=1, ~2 buckets per doubling)
	// Use for: Per-activity-type metrics, per-workflow-type metrics where cardinality is high
	Low1ms100s = Default1ms100s.subsetTo(1)

	// High1ms24h is for long-running operations like workflow end-to-end latency.
	// Range: 1ms → ~3 days
	// Buckets: 112 (scale=2, ~4 buckets per doubling)
	// Use for: Workflow end-to-end latency, long-running activity latency
	High1ms24h = makeSubsettableHistogram(2, time.Millisecond, 24*time.Hour, 112)

	// Mid1ms24h is a lower-resolution version of High1ms24h.
	// Range: 1ms → ~3 days
	// Buckets: 56 (scale=1, ~2 buckets per doubling)
	// Use for: When High1ms24h's cardinality is too high
	Mid1ms24h = High1ms24h.subsetTo(1)
)

// makeSubsettableHistogram creates an exponential histogram with the specified parameters.
//
// The bucket boundaries are calculated using the formula:
//
//	bucket[i] = start × 2^(i / 2^scale)
//
// Parameters:
//   - scale: Controls bucket density (0-3). Higher scale = more buckets = finer granularity
//   - scale=0: buckets double every 1 step (growth factor ≈ 2.00×)
//   - scale=1: buckets double every 2 steps (growth factor ≈ 1.41×)
//   - scale=2: buckets double every 4 steps (growth factor ≈ 1.19×)
//   - scale=3: buckets double every 8 steps (growth factor ≈ 1.09×)
//   - start: First bucket value (e.g., 1ms). Must be > 0.
//   - end: Target maximum value. Actual max will exceed this by at least 2x.
//   - length: Number of buckets. Must be divisible by 2^scale.
func makeSubsettableHistogram(scale int, start, end time.Duration, length int) SubsettableHistogram {
	if scale < 0 || scale > 3 {
		panic(fmt.Sprintf("scale must be between 0 (grows by *2) and 3 (grows by *2^1/8), got: %v", scale))
	}
	if start <= 0 {
		panic(fmt.Sprintf("start must be greater than 0, got %v", start))
	}
	if start >= end {
		panic(fmt.Sprintf("start must be less than end (%v < %v)", start, end))
	}
	if length < 12 || length > 160 {
		panic(fmt.Sprintf("length must be between 12 and 160, got %d", length))
	}

	// Ensure buckets complete a full "row" (power of 2 alignment)
	powerOfTwoWidth := int(math.Pow(2, float64(scale)))
	missing := length % powerOfTwoWidth
	if missing != 0 {
		panic(fmt.Sprintf("number of buckets must end at a power of 2. got %d, raise to %d",
			length, length+missing))
	}

	var buckets tally.DurationBuckets
	for i := 0; i < length; i++ {
		buckets = append(buckets, nextBucket(start, len(buckets), scale))
	}
	// Add one more to reach the next power of 2
	buckets = append(buckets, nextBucket(start, len(buckets), scale))

	if last(buckets) < end*2 {
		panic(fmt.Sprintf("not enough buckets (%d) to exceed the end target (%v) by at least 2x. "+
			"last bucket: %v. Consider increasing length or reducing end.",
			length, end, last(buckets)))
	}

	return SubsettableHistogram{
		tallyBuckets: append(
			tally.DurationBuckets{0}, // Always include 0 to detect negative values
			buckets...,
		),
		scale: scale,
	}
}

// nextBucket calculates the next bucket boundary using the exponential formula.
// Calculating from start each time reduces floating point error, ensuring "clean" values
// (e.g., 2ms instead of 1.9999994ms).
func nextBucket(start time.Duration, num int, scale int) time.Duration {
	// Formula: start × 2^(num / 2^scale)
	return time.Duration(
		float64(start) *
			math.Pow(2, float64(num)/math.Pow(2, float64(scale))))
}

// subsetTo reduces the histogram's detail level by reducing the scale.
// This halves the number of buckets for each scale reduction.
//
// Example: If you have a scale=2 histogram with 80 buckets,
// subsetTo(1) will produce a scale=1 histogram with 40 buckets.
func (s SubsettableHistogram) subsetTo(newScale int) SubsettableHistogram {
	if newScale >= s.scale {
		panic(fmt.Sprintf("scale %v is not less than the current scale %v", newScale, s.scale))
	}
	if newScale < 0 {
		panic(fmt.Sprintf("negative scales (%v) are not supported yet", newScale))
	}

	dup := SubsettableHistogram{
		tallyBuckets: slices.Clone(s.tallyBuckets),
		scale:        s.scale,
	}

	// Compress every other bucket per -1 scale
	for dup.scale > newScale {
		if (len(dup.tallyBuckets)-2)%2 != 0 {
			panic(fmt.Sprintf("cannot subset from scale %v to %v, %v buckets is not divisible by 2",
				dup.scale, dup.scale-1, len(dup.tallyBuckets)-2))
		}
		if len(dup.tallyBuckets) <= 3 {
			panic(fmt.Sprintf("not enough buckets to subset from scale %d to %d, only have %d",
				dup.scale, dup.scale-1, len(dup.tallyBuckets)))
		}

		// Keep first and last buckets, compress the rest
		bucketsToCompress := dup.tallyBuckets[1 : len(dup.tallyBuckets)-1]

		half := make(tally.DurationBuckets, 0, (len(bucketsToCompress)/2)+2)
		half = append(half, dup.tallyBuckets[0]) // keep zero value
		for i := 0; i < len(bucketsToCompress); i += 2 {
			half = append(half, bucketsToCompress[i]) // keep every other bucket
		}
		half = append(half, dup.tallyBuckets[len(dup.tallyBuckets)-1]) // keep last bucket

		dup.tallyBuckets = half
		dup.scale--
	}
	return dup
}

// Buckets returns the tally.DurationBuckets for use with scope.Histogram()
func (s SubsettableHistogram) Buckets() tally.DurationBuckets {
	return s.tallyBuckets
}

// RecordHistogram is a convenience helper to record a duration to a histogram.
// This makes it easier to migrate from Timer to Histogram.
//
// Example:
//
//	RecordHistogram(scope, "my-operation-latency", duration, Default1ms100s)
func RecordHistogram(scope tally.Scope, name string, duration time.Duration, buckets SubsettableHistogram) {
	scope.Histogram(name+"_ns", buckets.Buckets()).RecordDuration(duration)
}

// StartHistogram returns a stopwatch for timing operations with a histogram.
// Call .Stop() on the returned stopwatch to record the duration.
//
// Example:
//
//	sw := StartHistogram(scope, "my-operation-latency", Default1ms100s)
//	// ... do work ...
//	sw.Stop()
func StartHistogram(scope tally.Scope, name string, buckets SubsettableHistogram) tally.Stopwatch {
	return scope.Histogram(name+"_ns", buckets.Buckets()).Start()
}

// last returns the last element in a slice
func last[T any, X ~[]T](s X) T {
	return s[len(s)-1]
}
