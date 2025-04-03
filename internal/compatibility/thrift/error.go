// Copyright (c) 2021 Uber Technologies, Inc.
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

package thrift

import (
	"errors"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/yarpc/encoding/protobuf"
	"go.uber.org/yarpc/yarpcerrors"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

func Error(err error) error {
	status := yarpcerrors.FromError(err)
	if status == nil || status.Code() == yarpcerrors.CodeOK {
		return nil
	}

	switch status.Code() {
	case yarpcerrors.CodePermissionDenied:
		return &shared.AccessDeniedError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeInternal:
		return &shared.InternalServiceError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeNotFound:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.EntityNotExistsError:
			return &shared.EntityNotExistsError{
				Message:        status.Message(),
				CurrentCluster: &details.CurrentCluster,
				ActiveCluster:  &details.ActiveCluster,
			}
		case *apiv1.WorkflowExecutionAlreadyCompletedError:
			return &shared.WorkflowExecutionAlreadyCompletedError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeInvalidArgument:
		switch getErrorDetails(err).(type) {
		case nil:
			return &shared.BadRequestError{
				Message: status.Message(),
			}
		case *apiv1.QueryFailedError:
			return &shared.QueryFailedError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeAlreadyExists:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.CancellationAlreadyRequestedError:
			return &shared.CancellationAlreadyRequestedError{
				Message: status.Message(),
			}
		case *apiv1.DomainAlreadyExistsError:
			return &shared.DomainAlreadyExistsError{
				Message: status.Message(),
			}
		case *apiv1.WorkflowExecutionAlreadyStartedError:
			return &shared.WorkflowExecutionAlreadyStartedError{
				Message:        common.StringPtr(status.Message()),
				StartRequestId: &details.StartRequestId,
				RunId:          &details.RunId,
			}
		}
	case yarpcerrors.CodeDataLoss:
		return &shared.InternalDataInconsistencyError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeFailedPrecondition:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.ClientVersionNotSupportedError:
			return &shared.ClientVersionNotSupportedError{
				FeatureVersion:    details.FeatureVersion,
				ClientImpl:        details.ClientImpl,
				SupportedVersions: details.SupportedVersions,
			}
		case *apiv1.FeatureNotEnabledError:
			return &shared.FeatureNotEnabledError{
				FeatureFlag: details.FeatureFlag,
			}
		case *apiv1.DomainNotActiveError:
			return &shared.DomainNotActiveError{
				Message:        status.Message(),
				DomainName:     details.Domain,
				CurrentCluster: details.CurrentCluster,
				ActiveCluster:  details.ActiveCluster,
			}
		}
	case yarpcerrors.CodeResourceExhausted:
		switch getErrorDetails(err).(type) {
		case *apiv1.LimitExceededError:
			return &shared.LimitExceededError{
				Message: status.Message(),
			}
		case *apiv1.ServiceBusyError:
			return &shared.ServiceBusyError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeUnknown:
		return errors.New(status.Message())
	}

	// If error does not match anything, return raw yarpc status error
	// There are some code that casts error to yarpc status to check for deadline exceeded status
	return status
}

func getErrorDetails(err error) interface{} {
	details := protobuf.GetErrorDetails(err)
	if len(details) > 0 {
		return details[0]
	}
	return nil
}
