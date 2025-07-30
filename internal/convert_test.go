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

	"github.com/stretchr/testify/assert"

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
			name: "region sticky policy",
			policy: &ActiveClusterSelectionPolicy{
				Strategy: ActiveClusterSelectionStrategyRegionSticky,
			},
			thriftPolicy: &s.ActiveClusterSelectionPolicy{
				Strategy: s.ActiveClusterSelectionStrategyRegionSticky.Ptr(),
			},
		},
		{
			name: "external entity policy - success",
			policy: &ActiveClusterSelectionPolicy{
				Strategy:           ActiveClusterSelectionStrategyExternalEntity,
				ExternalEntityType: "test-type",
				ExternalEntityKey:  "test-key",
			},
			thriftPolicy: &s.ActiveClusterSelectionPolicy{
				Strategy:           s.ActiveClusterSelectionStrategyExternalEntity.Ptr(),
				ExternalEntityType: common.StringPtr("test-type"),
				ExternalEntityKey:  common.StringPtr("test-key"),
			},
		},
		{
			name: "external entity policy - missing type",
			policy: &ActiveClusterSelectionPolicy{
				Strategy:          ActiveClusterSelectionStrategyExternalEntity,
				ExternalEntityKey: "test-key",
			},
			wantErr: true,
		},
		{
			name: "external entity policy - missing key",
			policy: &ActiveClusterSelectionPolicy{
				Strategy:           ActiveClusterSelectionStrategyExternalEntity,
				ExternalEntityType: "test-type",
			},
			wantErr: true,
		},
		{
			name: "invalid strategy",
			policy: &ActiveClusterSelectionPolicy{
				Strategy: ActiveClusterSelectionStrategy(-1),
			},
			wantErr: true,
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
