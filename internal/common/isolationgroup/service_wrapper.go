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

package isolationgroup

import (
	"context"

	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
)

const (
	isolationGroupHeader = "cadence-client-isolation-group"
)

type workflowServiceIsolationGroupWrapper struct {
	service                  workflowserviceclient.Interface
	isolationGroupIdentifier string
}

// NewWorkflowServiceWrapper creates
func NewWorkflowServiceWrapper(service workflowserviceclient.Interface, isolationGroup string) workflowserviceclient.Interface {
	return &workflowServiceIsolationGroupWrapper{
		service:                  service,
		isolationGroupIdentifier: isolationGroup,
	}
}

func (w *workflowServiceIsolationGroupWrapper) getIsolationGroupIdentifier() yarpc.CallOption {
	return yarpc.WithHeader(isolationGroupHeader, w.isolationGroupIdentifier)
}

func (w *workflowServiceIsolationGroupWrapper) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	return w.service.DeprecateDomain(ctx, request, opts...)
}

func (w *workflowServiceIsolationGroupWrapper) ListDomains(ctx context.Context, request *shared.ListDomainsRequest, opts ...yarpc.CallOption) (*shared.ListDomainsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListDomains(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest, opts ...yarpc.CallOption) (*shared.DescribeDomainResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.DescribeDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.DescribeWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) DiagnoseWorkflowExecution(ctx context.Context, request *shared.DiagnoseWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DiagnoseWorkflowExecutionResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.DiagnoseWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.GetWorkflowExecutionHistory(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListClosedWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListOpenWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListArchivedWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ScanWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.CountWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.PollForActivityTask(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.PollForDecisionTask(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.RecordActivityTaskHeartbeat(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RegisterDomain(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RequestCancelWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondActivityTaskCanceled(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondActivityTaskCompleted(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondActivityTaskFailed(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondActivityTaskCanceledByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondActivityTaskCompletedByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondActivityTaskFailedByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	response, err := w.service.RespondDecisionTaskCompleted(ctx, request, opts...)
	return response, err
}

func (w *workflowServiceIsolationGroupWrapper) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondDecisionTaskFailed(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.SignalWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.SignalWithStartWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) SignalWithStartWorkflowExecutionAsync(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionAsyncRequest, opts ...yarpc.CallOption) (*shared.SignalWithStartWorkflowExecutionAsyncResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.SignalWithStartWorkflowExecutionAsync(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.StartWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) StartWorkflowExecutionAsync(ctx context.Context, request *shared.StartWorkflowExecutionAsyncRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionAsyncResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.StartWorkflowExecutionAsync(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.TerminateWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ResetWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest, opts ...yarpc.CallOption) (*shared.UpdateDomainResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.UpdateDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.QueryWorkflow(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ResetStickyTaskList(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.DescribeTaskList(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RespondQueryTaskCompleted(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.GetSearchAttributes(ctx, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListTaskListPartitions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*shared.ClusterInfo, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.GetClusterInfo(ctx, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) GetTaskListsByDomain(ctx context.Context, request *shared.GetTaskListsByDomainRequest, opts ...yarpc.CallOption) (*shared.GetTaskListsByDomainResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.GetTaskListsByDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIsolationGroupWrapper) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	err := w.service.RefreshWorkflowTasks(ctx, request, opts...)
	return err
}

func (w *workflowServiceIsolationGroupWrapper) RestartWorkflowExecution(ctx context.Context, request *shared.RestartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.RestartWorkflowExecutionResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	return w.service.RestartWorkflowExecution(ctx, request, opts...)
}

func (w *workflowServiceIsolationGroupWrapper) DeleteDomain(ctx context.Context, DeleteRequest *shared.DeleteDomainRequest, opts ...yarpc.CallOption) error {
	opts = append(opts, w.getIsolationGroupIdentifier())
	return w.service.DeleteDomain(ctx, DeleteRequest, opts...)
}

func (w *workflowServiceIsolationGroupWrapper) FailoverDomain(ctx context.Context, request *shared.FailoverDomainRequest, opts ...yarpc.CallOption) (*shared.FailoverDomainResponse, error) {
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.FailoverDomain(ctx, request, opts...)
	return result, err
}
