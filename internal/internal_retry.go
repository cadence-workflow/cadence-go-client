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

// All code in this file is private to the package.

import (
	"context"
	"errors"
	"time"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/backoff"
)

const (
	retryServiceOperationInitialInterval    = 20 * time.Millisecond
	retryServiceOperationExpirationInterval = 60 * time.Second
	retryServiceOperationBackoff            = 1.2
)

// Creates a retry policy which allows appropriate retries for the deadline passed in as context.
// It uses the context deadline to set MaxInterval as 1/10th of context timeout
// MaxInterval = Max(context_timeout/10, 20ms)
// defaults to ExpirationInterval of 60 seconds, or uses context deadline as expiration interval
func createDynamicServiceRetryPolicy(ctx context.Context) backoff.RetryPolicy {
	timeout := retryServiceOperationExpirationInterval
	if ctx != nil {
		now := time.Now()
		if expiration, ok := ctx.Deadline(); ok && expiration.After(now) {
			timeout = expiration.Sub(now)
		}
	}
	initialInterval := retryServiceOperationInitialInterval
	maximumInterval := timeout / 10
	if maximumInterval < retryServiceOperationInitialInterval {
		maximumInterval = retryServiceOperationInitialInterval
	}

	policy := backoff.NewExponentialRetryPolicy(initialInterval)
	policy.SetBackoffCoefficient(retryServiceOperationBackoff)
	policy.SetMaximumInterval(maximumInterval)
	policy.SetExpirationInterval(timeout)
	return policy
}

func isServiceTransientError(err error) bool {
	// check intentionally-not-retried error types via errors.As.
	//
	// sadly we cannot build this into a list of error values / types to range over, as that
	// would turn the variable passed as the &target into e.g. an interface{} or an error,
	// which is compatible with ALL errors, so all are .As(err, &target).
	//
	// so we're left with this mess.  it's not even generics-friendly.
	// at least this pattern lets it be done inline without exposing the variable.
	if target := (*s.AccessDeniedError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.BadRequestError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.CancellationAlreadyRequestedError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.ClientVersionNotSupportedError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.DomainAlreadyExistsError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.DomainNotActiveError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.EntityNotExistsError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.FeatureNotEnabledError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.LimitExceededError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.QueryFailedError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.WorkflowExecutionAlreadyCompletedError)(nil); errors.As(err, &target) {
		return false
	}
	if target := (*s.WorkflowExecutionAlreadyStartedError)(nil); errors.As(err, &target) {
		return false
	}

	// shutdowns are not retryable, of course
	if errors.Is(err, errShutdown) {
		return false
	}

	if target := (*s.ServiceBusyError)(nil); errors.As(err, &target) {
		return true
	}

	// s.InternalServiceError
	// s.ServiceBusyError (must retry after a delay, but it is transient)
	// server-side-only error types (as they should not reach clients)
	// and all other `error` types
	return true
}

func retryWhileTransientError(ctx context.Context, fn func() error) error {
	return backoff.Retry(ctx, fn, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
}
