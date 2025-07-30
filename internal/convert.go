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
	"fmt"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
)

func convertRetryPolicy(retryPolicy *RetryPolicy) *s.RetryPolicy {
	if retryPolicy == nil {
		return nil
	}
	thriftRetryPolicy := s.RetryPolicy{
		InitialIntervalInSeconds:    common.Int32Ptr(common.Int32Ceil(retryPolicy.InitialInterval.Seconds())),
		MaximumIntervalInSeconds:    common.Int32Ptr(common.Int32Ceil(retryPolicy.MaximumInterval.Seconds())),
		BackoffCoefficient:          &retryPolicy.BackoffCoefficient,
		MaximumAttempts:             &retryPolicy.MaximumAttempts,
		NonRetriableErrorReasons:    retryPolicy.NonRetriableErrorReasons,
		ExpirationIntervalInSeconds: common.Int32Ptr(common.Int32Ceil(retryPolicy.ExpirationInterval.Seconds())),
	}
	if *thriftRetryPolicy.BackoffCoefficient == 0 {
		thriftRetryPolicy.BackoffCoefficient = common.Float64Ptr(backoff.DefaultBackoffCoefficient)
	}
	return &thriftRetryPolicy
}

func convertActiveClusterSelectionPolicy(policy *ActiveClusterSelectionPolicy) (*s.ActiveClusterSelectionPolicy, error) {
	if policy == nil {
		return nil, nil
	}

	switch policy.Strategy {
	case ActiveClusterSelectionStrategyRegionSticky:
		return &s.ActiveClusterSelectionPolicy{
			Strategy: s.ActiveClusterSelectionStrategyRegionSticky.Ptr(),
		}, nil
	case ActiveClusterSelectionStrategyExternalEntity:
		if policy.ExternalEntityType == "" {
			return nil, fmt.Errorf("external entity type is required for external entity strategy")
		}
		if policy.ExternalEntityKey == "" {
			return nil, fmt.Errorf("external entity key is required for external entity strategy")
		}
		return &s.ActiveClusterSelectionPolicy{
			Strategy:           s.ActiveClusterSelectionStrategyExternalEntity.Ptr(),
			ExternalEntityType: common.StringPtr(policy.ExternalEntityType),
			ExternalEntityKey:  common.StringPtr(policy.ExternalEntityKey),
		}, nil
	default:
		return nil, fmt.Errorf("invalid active cluster selection strategy: %d", policy.Strategy)
	}
}

func convertQueryConsistencyLevel(level QueryConsistencyLevel) *s.QueryConsistencyLevel {
	switch level {
	case QueryConsistencyLevelEventual:
		return s.QueryConsistencyLevelEventual.Ptr()
	case QueryConsistencyLevelStrong:
		return s.QueryConsistencyLevelStrong.Ptr()
	default:
		return nil
	}
}
