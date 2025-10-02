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

package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
)

const (
	jwtHeaderName = "cadence-authorization"
)

type workflowServiceAuthWrapper struct {
	service      workflowserviceclient.Interface
	authProvider AuthorizationProvider
}

type AuthorizationProvider interface {
	// GetAuthToken provides the OAuth authorization token
	// It's called before every request to Cadence server, and sets the token in the request header.
	GetAuthToken() ([]byte, error)
}

type JWTClaims struct {
	jwt.RegisteredClaims

	Sub    string
	Name   string
	Groups string // separated by space
	Admin  bool
	TTL    int64
}

// NewWorkflowServiceWrapper creates
func NewWorkflowServiceWrapper(service workflowserviceclient.Interface, authorizationProvider AuthorizationProvider) workflowserviceclient.Interface {
	return &workflowServiceAuthWrapper{
		service:      service,
		authProvider: authorizationProvider,
	}
}

func (w *workflowServiceAuthWrapper) getYarpcJWTHeader() (*yarpc.CallOption, error) {
	token, err := w.authProvider.GetAuthToken()
	if err != nil {
		return nil, err
	}
	header := yarpc.WithHeader(jwtHeaderName, string(token))
	return &header, nil
}

func (w *workflowServiceAuthWrapper) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.DeprecateDomain(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) ListDomains(ctx context.Context, request *shared.ListDomainsRequest, opts ...yarpc.CallOption) (*shared.ListDomainsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ListDomains(ctx, request, opts...)

	return result, err
}

func (w *workflowServiceAuthWrapper) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest, opts ...yarpc.CallOption) (*shared.DescribeDomainResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.DescribeDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.DescribeWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) DiagnoseWorkflowExecution(ctx context.Context, request *shared.DiagnoseWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DiagnoseWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.DiagnoseWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.GetWorkflowExecutionHistory(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ListClosedWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ListOpenWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ListWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ListArchivedWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ScanWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.CountWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.PollForActivityTask(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.PollForDecisionTask(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.RecordActivityTaskHeartbeat(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RegisterDomain(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RequestCancelWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondActivityTaskCanceled(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondActivityTaskCompleted(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondActivityTaskFailed(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondActivityTaskCanceledByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondActivityTaskCompletedByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondActivityTaskFailedByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	response, err := w.service.RespondDecisionTaskCompleted(ctx, request, opts...)
	return response, err
}

func (w *workflowServiceAuthWrapper) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondDecisionTaskFailed(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.SignalWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.SignalWithStartWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) SignalWithStartWorkflowExecutionAsync(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionAsyncRequest, opts ...yarpc.CallOption) (*shared.SignalWithStartWorkflowExecutionAsyncResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.SignalWithStartWorkflowExecutionAsync(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.StartWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) StartWorkflowExecutionAsync(ctx context.Context, request *shared.StartWorkflowExecutionAsyncRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionAsyncResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.StartWorkflowExecutionAsync(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.TerminateWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ResetWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest, opts ...yarpc.CallOption) (*shared.UpdateDomainResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.UpdateDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.QueryWorkflow(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ResetStickyTaskList(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.DescribeTaskList(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RespondQueryTaskCompleted(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.GetSearchAttributes(ctx, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.ListTaskListPartitions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*shared.ClusterInfo, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.GetClusterInfo(ctx, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) GetTaskListsByDomain(ctx context.Context, request *shared.GetTaskListsByDomainRequest, opts ...yarpc.CallOption) (*shared.GetTaskListsByDomainResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.GetTaskListsByDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceAuthWrapper) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.RefreshWorkflowTasks(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RestartWorkflowExecution(ctx context.Context, request *shared.RestartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.RestartWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	return w.service.RestartWorkflowExecution(ctx, request, opts...)
}

func (w *workflowServiceAuthWrapper) DeleteDomain(ctx context.Context, DeleteRequest *shared.DeleteDomainRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	err = w.service.DeleteDomain(ctx, DeleteRequest, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) FailoverDomain(ctx context.Context, request *shared.FailoverDomainRequest, opts ...yarpc.CallOption) (*shared.FailoverDomainResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.FailoverDomain(ctx, request, opts...)
	return result, err
}
