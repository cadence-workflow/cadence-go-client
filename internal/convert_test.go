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

package internal

import (
	"testing"
	"time"

	gogo "github.com/gogo/protobuf/types"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
)

func TestConvertRetryPolicy(t *testing.T) {
	tests := []struct {
		name         string
		retryPolicy  *RetryPolicy
		thriftPolicy *s.RetryPolicy
	}{
		{
			name:         "nil retry policy - return nil",
			retryPolicy:  nil,
			thriftPolicy: nil,
		},
		{
			name: "non-zero backoff coefficient - use provided",
			retryPolicy: &RetryPolicy{
				InitialInterval:          1 * time.Second,
				MaximumInterval:          10 * time.Second,
				BackoffCoefficient:       2.0,
				MaximumAttempts:          3,
				NonRetriableErrorReasons: []string{"error1", "error2"},
				ExpirationInterval:       10 * time.Second,
			},
			thriftPolicy: &s.RetryPolicy{
				InitialIntervalInSeconds:    common.Int32Ptr(1),
				MaximumIntervalInSeconds:    common.Int32Ptr(10),
				BackoffCoefficient:          common.Float64Ptr(2.0),
				MaximumAttempts:             common.Int32Ptr(3),
				NonRetriableErrorReasons:    []string{"error1", "error2"},
				ExpirationIntervalInSeconds: common.Int32Ptr(10),
			},
		},
		{
			name: "zero backoff coefficient - use default",
			retryPolicy: &RetryPolicy{
				InitialInterval:          1 * time.Second,
				MaximumInterval:          10 * time.Second,
				BackoffCoefficient:       0.0,
				MaximumAttempts:          3,
				NonRetriableErrorReasons: []string{"error1", "error2"},
				ExpirationInterval:       10 * time.Second,
			},
			thriftPolicy: &s.RetryPolicy{
				InitialIntervalInSeconds:    common.Int32Ptr(1),
				MaximumIntervalInSeconds:    common.Int32Ptr(10),
				BackoffCoefficient:          common.Float64Ptr(backoff.DefaultBackoffCoefficient),
				MaximumAttempts:             common.Int32Ptr(3),
				NonRetriableErrorReasons:    []string{"error1", "error2"},
				ExpirationIntervalInSeconds: common.Int32Ptr(10),
			},
		},
	}

	for _, test := range tests {
		thriftPolicy := convertRetryPolicy(test.retryPolicy)
		assert.Equal(t, test.thriftPolicy, thriftPolicy)
	}
}

func TestConvertActiveClusterSelectionPolicy(t *testing.T) {
	tests := []struct {
		name         string
		policy       *ActiveClusterSelectionPolicy
		thriftPolicy *s.ActiveClusterSelectionPolicy
		wantErr      bool
	}{
		{
			name:         "nil policy - return nil",
			policy:       nil,
			thriftPolicy: nil,
		},
		{
			name:         "empty policy",
			policy:       &ActiveClusterSelectionPolicy{},
			thriftPolicy: &s.ActiveClusterSelectionPolicy{},
		},
		{
			name: "valid policy",
			policy: &ActiveClusterSelectionPolicy{
				ClusterAttribute: &ClusterAttribute{
					Scope: "region",
					Name:  "us-east-1",
				},
			},
			thriftPolicy: &s.ActiveClusterSelectionPolicy{
				ClusterAttribute: &s.ClusterAttribute{
					Scope: common.StringPtr("region"),
					Name:  common.StringPtr("us-east-1"),
				},
			},
		},
	}

	for _, test := range tests {
		thriftPolicy, err := convertActiveClusterSelectionPolicy(test.policy)
		if test.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.thriftPolicy, thriftPolicy)
		}
	}
}

func TestConvertQueryConsistencyLevel(t *testing.T) {
	tests := []struct {
		name           string
		level          QueryConsistencyLevel
		expectedThrift *s.QueryConsistencyLevel
	}{
		{
			name:           "QueryConsistencyLevelUnspecified - return nil",
			level:          QueryConsistencyLevelUnspecified,
			expectedThrift: nil,
		},
		{
			name:           "QueryConsistencyLevelEventual - return eventual",
			level:          QueryConsistencyLevelEventual,
			expectedThrift: s.QueryConsistencyLevelEventual.Ptr(),
		},
		{
			name:           "QueryConsistencyLevelStrong - return strong",
			level:          QueryConsistencyLevelStrong,
			expectedThrift: s.QueryConsistencyLevelStrong.Ptr(),
		},
		{
			name:           "invalid level (999) - return nil",
			level:          QueryConsistencyLevel(999),
			expectedThrift: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := convertQueryConsistencyLevel(test.level)
			assert.Equal(t, test.expectedThrift, result)
		})
	}
}

func TestToProtoTimestamp(t *testing.T) {
	testcases := []struct {
		name    string
		input   time.Time
		wantNil bool
	}{
		{name: "zero time", input: time.Time{}, wantNil: true},
		{name: "unix epoch", input: time.Unix(0, 0).UTC()},
		{name: "specific time with nanos", input: time.Date(2024, 1, 15, 10, 30, 0, 500, time.UTC)},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoTimestamp(tt.input)
			if tt.wantNil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result)
			assert.Equal(t, tt.input.Unix(), result.Seconds)
			assert.Equal(t, int32(tt.input.Nanosecond()), result.Nanos)
		})
	}
}

func TestFromProtoTimestamp(t *testing.T) {
	testcases := []struct {
		name  string
		input *gogo.Timestamp
		want  time.Time
	}{
		{name: "nil", input: nil, want: time.Time{}},
		{name: "unix epoch", input: &gogo.Timestamp{Seconds: 0, Nanos: 0}, want: time.Unix(0, 0).UTC()},
		{name: "seconds and nanos", input: &gogo.Timestamp{Seconds: 1705312200, Nanos: 500}, want: time.Unix(1705312200, 500).UTC()},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fromProtoTimestamp(tt.input))
		})
	}
}

func TestToProtoDuration(t *testing.T) {
	testcases := []struct {
		name    string
		input   time.Duration
		wantNil bool
	}{
		{name: "zero", input: 0, wantNil: true},
		{name: "one second", input: time.Second},
		{name: "sub-second", input: 500 * time.Millisecond},
		{name: "mixed seconds and nanos", input: 2*time.Second + 300*time.Nanosecond},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoDuration(tt.input)
			if tt.wantNil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result)
			got := time.Duration(result.Seconds)*time.Second + time.Duration(result.Nanos)
			assert.Equal(t, tt.input, got)
		})
	}
}

func TestFromProtoDuration(t *testing.T) {
	testcases := []struct {
		name  string
		input *gogo.Duration
		want  time.Duration
	}{
		{name: "nil", input: nil, want: 0},
		{name: "zero", input: &gogo.Duration{Seconds: 0, Nanos: 0}, want: 0},
		{name: "one second", input: &gogo.Duration{Seconds: 1, Nanos: 0}, want: time.Second},
		{name: "nanos only", input: &gogo.Duration{Seconds: 0, Nanos: 500}, want: 500 * time.Nanosecond},
		{name: "mixed", input: &gogo.Duration{Seconds: 2, Nanos: 300}, want: 2*time.Second + 300*time.Nanosecond},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fromProtoDuration(tt.input))
		})
	}
}

// TestProtoTimestampRoundTrip verifies fromProtoTimestamp(toProtoTimestamp(ts)) == ts
// using a seeded fuzzer to cover a wide range of timestamp values.
func TestProtoTimestampRoundTrip(t *testing.T) {
	const (
		seed       = 123
		iterations = 100
		// cap seconds to stay within the safe int64 nanosecond range (~year 2262)
		maxSeconds = int64(9223372036)
		maxNanos   = int32(1e9 - 1)
	)
	f := fuzz.NewWithSeed(seed).NilChance(0)
	f.Funcs(func(ts *gogo.Timestamp, c fuzz.Continue) {
		ts.Seconds = c.Int63n(maxSeconds)
		ts.Nanos = c.Int31n(maxNanos + 1)
	})

	for i := 0; i < iterations; i++ {
		var ts gogo.Timestamp
		f.Fuzz(&ts)

		got := toProtoTimestamp(fromProtoTimestamp(&ts))
		require.NotNil(t, got, "iteration %d", i)
		assert.Equal(t, ts.Seconds, got.Seconds, "iteration %d", i)
		assert.Equal(t, ts.Nanos, got.Nanos, "iteration %d", i)
	}
}

// TestProtoDurationRoundTrip verifies fromProtoDuration(toProtoDuration(d)) == d
// using a seeded fuzzer to cover a wide range of duration values.
func TestProtoDurationRoundTrip(t *testing.T) {
	const (
		seed       = 123
		iterations = 100
		maxSeconds = int64(1<<32 - 1)
		maxNanos   = int32(1e9 - 1)
	)
	f := fuzz.NewWithSeed(seed).NilChance(0)
	f.Funcs(func(d *gogo.Duration, c fuzz.Continue) {
		d.Seconds = c.Int63n(maxSeconds)
		d.Nanos = c.Int31n(maxNanos + 1)
		if d.Seconds == 0 && d.Nanos == 0 {
			d.Seconds = 1 // zero maps to nil, which doesn't round-trip back to {0,0}
		}
	})

	for i := 0; i < iterations; i++ {
		var d gogo.Duration
		f.Fuzz(&d)

		got := toProtoDuration(fromProtoDuration(&d))
		require.NotNil(t, got, "iteration %d", i)
		assert.Equal(t, d.Seconds, got.Seconds, "iteration %d", i)
		assert.Equal(t, d.Nanos, got.Nanos, "iteration %d", i)
	}
}

func TestFromProtoTimestamp_InvalidNanos(t *testing.T) {
	// gogo.TimestampFromProto returns an error when Nanos is out of range.
	invalid := &gogo.Timestamp{Seconds: 0, Nanos: -1}
	result := fromProtoTimestamp(invalid)
	assert.Equal(t, time.Time{}, result)
}

func TestFromProtoDuration_InvalidNanos(t *testing.T) {
	// gogo.DurationFromProto returns an error when Nanos is outside [-999999999, 999999999].
	invalid := &gogo.Duration{Seconds: 0, Nanos: 2000000000}
	result := fromProtoDuration(invalid)
	assert.Equal(t, time.Duration(0), result)
}
