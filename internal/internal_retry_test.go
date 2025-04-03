// Copyright (c) 2017-2021 Uber Technologies Inc.
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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	s "go.uber.org/cadence/.gen/go/shared"
)

func TestErrRetries(t *testing.T) {
	t.Run("retryable", func(t *testing.T) {
		for _, err := range []error{
			// service-busy means "retry later", which is still transient/retryable.
			// callers with retries MUST detect this separately and delay before retrying,
			// and ideally we'll return a minimum time-to-wait in errors in the future.
			&s.ServiceBusyError{},

			&s.InternalServiceError{},  // fallback error type from server, "unknown" = "retry might work"
			errors.New("unrecognized"), // fallback error type elsewhere, "unknown" = "retry might work"

			// below are all server-side internal errors and should never be exposed, they're just included
			// to be explicit about current behavior (without intentionally deciding on any of them).
			// retry by default, like any other unknown error.  it shouldn't lead to incorrect behavior.
			&s.CurrentBranchChangedError{},
			&s.RetryTaskV2Error{},
			&s.RemoteSyncMatchedError{},
			&s.InternalDataInconsistencyError{},
		} {
			retryable := isServiceTransientError(err)
			assert.True(t, retryable, "%T should be transient", err)
		}
	})
	t.Run("terminal", func(t *testing.T) {
		for _, err := range []error{
			&s.AccessDeniedError{},                      // access won't be granted immediately
			&s.BadRequestError{},                        // bad requests don't become good
			&s.CancellationAlreadyRequestedError{},      // can only cancel once
			&s.ClientVersionNotSupportedError{},         // clients don't magically upgrade
			&s.DomainAlreadyExistsError{},               // re-creating again won't work
			&s.DomainNotActiveError{},                   // may actually succeed, but very unlikely to work immediately
			&s.EntityNotExistsError{},                   // may actually succeed, but very unlikely to work immediately
			&s.FeatureNotEnabledError{},                 // features won't magically enable
			&s.LimitExceededError{},                     // adding +1 to a value does not move it below its limit
			&s.QueryFailedError{},                       // arguable, this could be considered "unknown" and retryable
			&s.WorkflowExecutionAlreadyCompletedError{}, // completed workflows won't uncomplete
			&s.WorkflowExecutionAlreadyStartedError{},   // started workflows could complete quickly, but re-starting may not be desirable

			errShutdown, // shutdowns can't be stopped
		} {
			retryable := isServiceTransientError(err)
			assert.False(t, retryable, "%T should be fatal", err)
		}
	})
}

func TestRetryWhileTransientError(t *testing.T) {
	testcases := []struct {
		name          string
		errors        []error
		expectedError error
	}{
		{
			name:          "Immediately return success",
			errors:        []error{nil},
			expectedError: nil,
		},
		{
			name:          "Success after retriable error",
			errors:        []error{&s.InternalServiceError{}, nil},
			expectedError: nil,
		},
		{
			name:          "AccessDenied should not be retried",
			errors:        []error{&s.AccessDeniedError{}},
			expectedError: &s.AccessDeniedError{},
		},
		{
			name:          "Multiple errors until non-retriable",
			errors:        []error{&s.InternalServiceError{}, &s.InternalServiceError{}, &s.AccessDeniedError{}},
			expectedError: &s.AccessDeniedError{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			i := 0
			err := retryWhileTransientError(
				context.Background(),
				func() error {
					e := tt.errors[i]
					i++
					return e
				},
			)

			assert.Equal(t, len(tt.errors), i, "all errors expected to be consumed")
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
