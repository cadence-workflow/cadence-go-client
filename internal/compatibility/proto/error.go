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

package proto

import (
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/yarpc/encoding/protobuf"
	"go.uber.org/yarpc/yarpcerrors"

	"go.uber.org/cadence/.gen/go/shared"
)

func Error(err error) error {
	if err == nil {
		return protobuf.NewError(yarpcerrors.CodeOK, "")
	}

	switch e := err.(type) {
	case *shared.AccessDeniedError:
		return protobuf.NewError(yarpcerrors.CodePermissionDenied, e.Message)
	case *shared.InternalServiceError:
		return protobuf.NewError(yarpcerrors.CodeInternal, e.Message)
	case *shared.EntityNotExistsError:
		return protobuf.NewError(yarpcerrors.CodeNotFound, e.Message, protobuf.WithErrorDetails(&apiv1.EntityNotExistsError{
			CurrentCluster: e.GetCurrentCluster(),
			ActiveCluster:  e.GetActiveCluster(),
		}))
	case *shared.WorkflowExecutionAlreadyCompletedError:
		return protobuf.NewError(yarpcerrors.CodeNotFound, e.Message, protobuf.WithErrorDetails(&apiv1.WorkflowExecutionAlreadyCompletedError{}))
	case *shared.BadRequestError:
		return protobuf.NewError(yarpcerrors.CodeInvalidArgument, e.Message)
	case *shared.QueryFailedError:
		return protobuf.NewError(yarpcerrors.CodeInvalidArgument, e.Message, protobuf.WithErrorDetails(&apiv1.QueryFailedError{}))

	case *shared.CancellationAlreadyRequestedError:
		return protobuf.NewError(yarpcerrors.CodeAlreadyExists, e.Message, protobuf.WithErrorDetails(&apiv1.CancellationAlreadyRequestedError{}))
	case *shared.DomainAlreadyExistsError:
		return protobuf.NewError(yarpcerrors.CodeAlreadyExists, e.Message, protobuf.WithErrorDetails(&apiv1.DomainAlreadyExistsError{}))
	case *shared.WorkflowExecutionAlreadyStartedError:
		return protobuf.NewError(yarpcerrors.CodeAlreadyExists, e.GetMessage(), protobuf.WithErrorDetails(&apiv1.WorkflowExecutionAlreadyStartedError{
			StartRequestId: e.GetStartRequestId(),
			RunId:          e.GetRunId(),
		}))
	case *shared.ClientVersionNotSupportedError:
		return protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "Client version not supported", protobuf.WithErrorDetails(&apiv1.ClientVersionNotSupportedError{
			FeatureVersion:    e.FeatureVersion,
			ClientImpl:        e.ClientImpl,
			SupportedVersions: e.SupportedVersions,
		}))
	case *shared.FeatureNotEnabledError:
		return protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "Feature flag not enabled", protobuf.WithErrorDetails(&apiv1.FeatureNotEnabledError{
			FeatureFlag: e.FeatureFlag,
		}))
	case *shared.DomainNotActiveError:
		return protobuf.NewError(yarpcerrors.CodeFailedPrecondition, e.Message, protobuf.WithErrorDetails(&apiv1.DomainNotActiveError{
			Domain:         e.DomainName,
			CurrentCluster: e.CurrentCluster,
			ActiveCluster:  e.ActiveCluster,
		}))
	case *shared.LimitExceededError:
		return protobuf.NewError(yarpcerrors.CodeResourceExhausted, e.Message, protobuf.WithErrorDetails(&apiv1.LimitExceededError{}))
	case *shared.ServiceBusyError:
		return protobuf.NewError(yarpcerrors.CodeResourceExhausted, e.Message, protobuf.WithErrorDetails(&apiv1.ServiceBusyError{}))
	}

	return protobuf.NewError(yarpcerrors.CodeUnknown, err.Error())
}
