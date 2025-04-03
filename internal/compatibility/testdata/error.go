// Copyright (c) 2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package testdata

import (
	"go.uber.org/yarpc/encoding/protobuf"
	"go.uber.org/yarpc/yarpcerrors"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

const (
	ErrorMessage = "ErrorMessage"
	FeatureFlag  = "FeatureFlag"
)

var (
	AccessDeniedError                 = protobuf.NewError(yarpcerrors.CodePermissionDenied, ErrorMessage)
	BadRequestError                   = protobuf.NewError(yarpcerrors.CodeInvalidArgument, ErrorMessage)
	CancellationAlreadyRequestedError = protobuf.NewError(yarpcerrors.CodeAlreadyExists, ErrorMessage, protobuf.WithErrorDetails(&apiv1.CancellationAlreadyRequestedError{}))
	ClientVersionNotSupportedError    = protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "Client version not supported", protobuf.WithErrorDetails(&apiv1.ClientVersionNotSupportedError{
		FeatureVersion:    FeatureVersion,
		ClientImpl:        ClientImpl,
		SupportedVersions: SupportedVersions,
	}))
	DomainAlreadyExistsError = protobuf.NewError(yarpcerrors.CodeAlreadyExists, ErrorMessage, protobuf.WithErrorDetails(&apiv1.DomainAlreadyExistsError{}))
	DomainNotActiveError     = protobuf.NewError(yarpcerrors.CodeFailedPrecondition, ErrorMessage, protobuf.WithErrorDetails(&apiv1.DomainNotActiveError{
		Domain:         DomainName,
		CurrentCluster: ClusterName1,
		ActiveCluster:  ClusterName2,
	}))
	EntityNotExistsError = protobuf.NewError(yarpcerrors.CodeNotFound, ErrorMessage, protobuf.WithErrorDetails(&apiv1.EntityNotExistsError{
		CurrentCluster: ClusterName1,
		ActiveCluster:  ClusterName2,
	}))
	FeatureNotEnabledError = protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "Feature flag not enabled", protobuf.WithErrorDetails(&apiv1.FeatureNotEnabledError{
		FeatureFlag: FeatureFlag,
	}))
	WorkflowExecutionAlreadyCompletedError = protobuf.NewError(yarpcerrors.CodeNotFound, ErrorMessage, protobuf.WithErrorDetails(&apiv1.WorkflowExecutionAlreadyCompletedError{}))
	InternalServiceError                   = protobuf.NewError(yarpcerrors.CodeInternal, ErrorMessage)
	LimitExceededError                     = protobuf.NewError(yarpcerrors.CodeResourceExhausted, ErrorMessage, protobuf.WithErrorDetails(&apiv1.LimitExceededError{}))
	QueryFailedError                       = protobuf.NewError(yarpcerrors.CodeInvalidArgument, ErrorMessage, protobuf.WithErrorDetails(&apiv1.QueryFailedError{}))
	ServiceBusyError                       = protobuf.NewError(yarpcerrors.CodeResourceExhausted, ErrorMessage, protobuf.WithErrorDetails(&apiv1.ServiceBusyError{}))
	WorkflowExecutionAlreadyStartedError   = protobuf.NewError(yarpcerrors.CodeAlreadyExists, ErrorMessage, protobuf.WithErrorDetails(&apiv1.WorkflowExecutionAlreadyStartedError{
		StartRequestId: RequestID,
		RunId:          RunID,
	}))
	UnknownError = protobuf.NewError(yarpcerrors.CodeUnknown, ErrorMessage)
)
