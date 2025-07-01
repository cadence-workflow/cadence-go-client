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

package compatibility

import (
	"context"

	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/compatibility/proto"
	"go.uber.org/cadence/internal/compatibility/thrift"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

type thrift2protoAdapter struct {
	domain     apiv1.DomainAPIYARPCClient
	workflow   apiv1.WorkflowAPIYARPCClient
	worker     apiv1.WorkerAPIYARPCClient
	visibility apiv1.VisibilityAPIYARPCClient
}

func NewThrift2ProtoAdapter(
	domain apiv1.DomainAPIYARPCClient,
	workflow apiv1.WorkflowAPIYARPCClient,
	worker apiv1.WorkerAPIYARPCClient,
	visibility apiv1.VisibilityAPIYARPCClient,
) workflowserviceclient.Interface {
	return thrift2protoAdapter{domain, workflow, worker, visibility}
}

func (a thrift2protoAdapter) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error) {
	response, err := a.visibility.CountWorkflowExecutions(ctx, proto.CountWorkflowExecutionsRequest(request), opts...)
	return thrift.CountWorkflowExecutionsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	_, err := a.domain.DeprecateDomain(ctx, proto.DeprecateDomainRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest, opts ...yarpc.CallOption) (*shared.DescribeDomainResponse, error) {
	response, err := a.domain.DescribeDomain(ctx, proto.DescribeDomainRequest(request), opts...)
	return thrift.DescribeDomainResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
	response, err := a.workflow.DescribeTaskList(ctx, proto.DescribeTaskListRequest(request), opts...)
	return thrift.DescribeTaskListResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	response, err := a.workflow.DescribeWorkflowExecution(ctx, proto.DescribeWorkflowExecutionRequest(request), opts...)
	return thrift.DescribeWorkflowExecutionResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) DiagnoseWorkflowExecution(ctx context.Context, request *shared.DiagnoseWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DiagnoseWorkflowExecutionResponse, error) {
	response, err := a.workflow.DiagnoseWorkflowExecution(ctx, proto.DiagnoseWorkflowExecutionRequest(request), opts...)
	return thrift.DiagnoseWorkflowExecutionResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*shared.ClusterInfo, error) {
	response, err := a.workflow.GetClusterInfo(ctx, &apiv1.GetClusterInfoRequest{}, opts...)
	return thrift.GetClusterInfoResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error) {
	response, err := a.visibility.GetSearchAttributes(ctx, &apiv1.GetSearchAttributesRequest{}, opts...)
	return thrift.GetSearchAttributesResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	response, err := a.workflow.GetWorkflowExecutionHistory(ctx, proto.GetWorkflowExecutionHistoryRequest(request), opts...)
	return thrift.GetWorkflowExecutionHistoryResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListArchivedWorkflowExecutions(ctx, proto.ListArchivedWorkflowExecutionsRequest(request), opts...)
	return thrift.ListArchivedWorkflowExecutionsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListClosedWorkflowExecutions(ctx, proto.ListClosedWorkflowExecutionsRequest(request), opts...)
	return thrift.ListClosedWorkflowExecutionsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) ListDomains(ctx context.Context, request *shared.ListDomainsRequest, opts ...yarpc.CallOption) (*shared.ListDomainsResponse, error) {
	response, err := a.domain.ListDomains(ctx, proto.ListDomainsRequest(request), opts...)
	return thrift.ListDomainsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListOpenWorkflowExecutions(ctx, proto.ListOpenWorkflowExecutionsRequest(request), opts...)
	return thrift.ListOpenWorkflowExecutionsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
	response, err := a.workflow.ListTaskListPartitions(ctx, proto.ListTaskListPartitionsRequest(request), opts...)
	return thrift.ListTaskListPartitionsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListWorkflowExecutions(ctx, proto.ListWorkflowExecutionsRequest(request), opts...)
	return thrift.ListWorkflowExecutionsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
	response, err := a.worker.PollForActivityTask(ctx, proto.PollForActivityTaskRequest(request), opts...)
	return thrift.PollForActivityTaskResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error) {
	response, err := a.worker.PollForDecisionTask(ctx, proto.PollForDecisionTaskRequest(request), opts...)
	return thrift.PollForDecisionTaskResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
	response, err := a.workflow.QueryWorkflow(ctx, proto.QueryWorkflowRequest(request), opts...)
	return thrift.QueryWorkflowResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	response, err := a.worker.RecordActivityTaskHeartbeat(ctx, proto.RecordActivityTaskHeartbeatRequest(request), opts...)
	return thrift.RecordActivityTaskHeartbeatResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	response, err := a.worker.RecordActivityTaskHeartbeatByID(ctx, proto.RecordActivityTaskHeartbeatByIDRequest(request), opts...)
	return thrift.RecordActivityTaskHeartbeatByIDResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	_, err := a.domain.RegisterDomain(ctx, proto.RegisterDomainRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := a.workflow.RequestCancelWorkflowExecution(ctx, proto.RequestCancelWorkflowExecutionRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error) {
	_, err := a.worker.ResetStickyTaskList(ctx, proto.ResetStickyTaskListRequest(request), opts...)
	return &shared.ResetStickyTaskListResponse{}, thrift.Error(err)
}

func (a thrift2protoAdapter) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
	response, err := a.workflow.ResetWorkflowExecution(ctx, proto.ResetWorkflowExecutionRequest(request), opts...)
	return thrift.ResetWorkflowExecutionResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCanceled(ctx, proto.RespondActivityTaskCanceledRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCanceledByID(ctx, proto.RespondActivityTaskCanceledByIDRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCompleted(ctx, proto.RespondActivityTaskCompletedRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCompletedByID(ctx, proto.RespondActivityTaskCompletedByIDRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskFailed(ctx, proto.RespondActivityTaskFailedRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskFailedByID(ctx, proto.RespondActivityTaskFailedByIDRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error) {
	response, err := a.worker.RespondDecisionTaskCompleted(ctx, proto.RespondDecisionTaskCompletedRequest(request), opts...)
	return thrift.RespondDecisionTaskCompletedResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondDecisionTaskFailed(ctx, proto.RespondDecisionTaskFailedRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondQueryTaskCompleted(ctx, proto.RespondQueryTaskCompletedRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ScanWorkflowExecutions(ctx, proto.ScanWorkflowExecutionsRequest(request), opts...)
	return thrift.ScanWorkflowExecutionsResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := a.workflow.SignalWithStartWorkflowExecution(ctx, proto.SignalWithStartWorkflowExecutionRequest(request), opts...)
	return thrift.SignalWithStartWorkflowExecutionResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) SignalWithStartWorkflowExecutionAsync(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionAsyncRequest, opts ...yarpc.CallOption) (*shared.SignalWithStartWorkflowExecutionAsyncResponse, error) {
	response, err := a.workflow.SignalWithStartWorkflowExecutionAsync(ctx, proto.SignalWithStartWorkflowExecutionAsyncRequest(request), opts...)
	return thrift.SignalWithStartWorkflowExecutionAsyncResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := a.workflow.SignalWorkflowExecution(ctx, proto.SignalWorkflowExecutionRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := a.workflow.StartWorkflowExecution(ctx, proto.StartWorkflowExecutionRequest(request), opts...)
	return thrift.StartWorkflowExecutionResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) StartWorkflowExecutionAsync(ctx context.Context, request *shared.StartWorkflowExecutionAsyncRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionAsyncResponse, error) {
	response, err := a.workflow.StartWorkflowExecutionAsync(ctx, proto.StartWorkflowExecutionAsyncRequest(request), opts...)
	return thrift.StartWorkflowExecutionAsyncResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := a.workflow.TerminateWorkflowExecution(ctx, proto.TerminateWorkflowExecutionRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest, opts ...yarpc.CallOption) (*shared.UpdateDomainResponse, error) {
	response, err := a.domain.UpdateDomain(ctx, proto.UpdateDomainRequest(request), opts...)
	return thrift.UpdateDomainResponse(response), thrift.Error(err)
}

func (a thrift2protoAdapter) GetTaskListsByDomain(ctx context.Context, Request *shared.GetTaskListsByDomainRequest, opts ...yarpc.CallOption) (*shared.GetTaskListsByDomainResponse, error) {
	panic("implement me")
}

func (a thrift2protoAdapter) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	_, err := a.workflow.RefreshWorkflowTasks(ctx, proto.RefreshWorkflowTasksRequest(request), opts...)
	return thrift.Error(err)
}

func (a thrift2protoAdapter) RestartWorkflowExecution(ctx context.Context, request *shared.RestartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.RestartWorkflowExecutionResponse, error) {
	response, err := a.workflow.RestartWorkflowExecution(ctx, proto.RestartWorkflowExecutionRequest(request), opts...)
	return thrift.RestartWorkflowExecutionResponse(response), proto.Error(err)
}

type domainAPIthriftAdapter struct {
	service workflowserviceclient.Interface
}

func NewDomainAPITriftAdapter(thrift workflowserviceclient.Interface) domainAPIthriftAdapter { //revive:disable-line:unexported-return this is not used at all, consider removing
	return domainAPIthriftAdapter{thrift}
}

func (a domainAPIthriftAdapter) RegisterDomain(ctx context.Context, request *apiv1.RegisterDomainRequest, opts ...yarpc.CallOption) (*apiv1.RegisterDomainResponse, error) {
	err := a.service.RegisterDomain(ctx, thrift.RegisterDomainRequest(request), opts...)
	return &apiv1.RegisterDomainResponse{}, proto.Error(err)
}
func (a domainAPIthriftAdapter) DescribeDomain(ctx context.Context, request *apiv1.DescribeDomainRequest, opts ...yarpc.CallOption) (*apiv1.DescribeDomainResponse, error) {
	response, err := a.service.DescribeDomain(ctx, thrift.DescribeDomainRequest(request), opts...)
	return proto.DescribeDomainResponse(response), proto.Error(err)
}
func (a domainAPIthriftAdapter) ListDomains(ctx context.Context, request *apiv1.ListDomainsRequest, opts ...yarpc.CallOption) (*apiv1.ListDomainsResponse, error) {
	response, err := a.service.ListDomains(ctx, thrift.ListDomainsRequest(request), opts...)
	return proto.ListDomainsResponse(response), proto.Error(err)
}
func (a domainAPIthriftAdapter) UpdateDomain(ctx context.Context, request *apiv1.UpdateDomainRequest, opts ...yarpc.CallOption) (*apiv1.UpdateDomainResponse, error) {
	response, err := a.service.UpdateDomain(ctx, thrift.UpdateDomainRequest(request), opts...)
	return proto.UpdateDomainResponse(response), proto.Error(err)
}
func (a domainAPIthriftAdapter) DeprecateDomain(ctx context.Context, request *apiv1.DeprecateDomainRequest, opts ...yarpc.CallOption) (*apiv1.DeprecateDomainResponse, error) {
	err := a.service.DeprecateDomain(ctx, thrift.DeprecateDomainRequest(request), opts...)
	return &apiv1.DeprecateDomainResponse{}, proto.Error(err)
}

type workflowAPIthriftAdapter struct {
	service workflowserviceclient.Interface
}

func NewWorkflowAPITriftAdapter(thrift workflowserviceclient.Interface) workflowAPIthriftAdapter { //revive:disable-line:unexported-return this is not used at all, consider removing
	return workflowAPIthriftAdapter{thrift}
}

func (a workflowAPIthriftAdapter) StartWorkflowExecution(ctx context.Context, request *apiv1.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.StartWorkflowExecutionResponse, error) {
	response, err := a.service.StartWorkflowExecution(ctx, thrift.StartWorkflowExecutionRequest(request), opts...)
	return proto.StartWorkflowExecutionResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) SignalWorkflowExecution(ctx context.Context, request *apiv1.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.SignalWorkflowExecutionResponse, error) {
	err := a.service.SignalWorkflowExecution(ctx, thrift.SignalWorkflowExecutionRequest(request), opts...)
	return &apiv1.SignalWorkflowExecutionResponse{}, proto.Error(err)
}
func (a workflowAPIthriftAdapter) SignalWithStartWorkflowExecution(ctx context.Context, request *apiv1.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.SignalWithStartWorkflowExecutionResponse, error) {
	response, err := a.service.SignalWithStartWorkflowExecution(ctx, thrift.SignalWithStartWorkflowExecutionRequest(request), opts...)
	return proto.SignalWithStartWorkflowExecutionResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) ResetWorkflowExecution(ctx context.Context, request *apiv1.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.ResetWorkflowExecutionResponse, error) {
	response, err := a.service.ResetWorkflowExecution(ctx, thrift.ResetWorkflowExecutionRequest(request), opts...)
	return proto.ResetWorkflowExecutionResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) RequestCancelWorkflowExecution(ctx context.Context, request *apiv1.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.RequestCancelWorkflowExecutionResponse, error) {
	err := a.service.RequestCancelWorkflowExecution(ctx, thrift.RequestCancelWorkflowExecutionRequest(request), opts...)
	return &apiv1.RequestCancelWorkflowExecutionResponse{}, proto.Error(err)
}
func (a workflowAPIthriftAdapter) TerminateWorkflowExecution(ctx context.Context, request *apiv1.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.TerminateWorkflowExecutionResponse, error) {
	err := a.service.TerminateWorkflowExecution(ctx, thrift.TerminateWorkflowExecutionRequest(request), opts...)
	return &apiv1.TerminateWorkflowExecutionResponse{}, proto.Error(err)
}
func (a workflowAPIthriftAdapter) DescribeWorkflowExecution(ctx context.Context, request *apiv1.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.DescribeWorkflowExecutionResponse, error) {
	response, err := a.service.DescribeWorkflowExecution(ctx, thrift.DescribeWorkflowExecutionRequest(request), opts...)
	return proto.DescribeWorkflowExecutionResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) DiagnoseWorkflowExecution(ctx context.Context, request *apiv1.DiagnoseWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.DiagnoseWorkflowExecutionResponse, error) {
	response, err := a.service.DiagnoseWorkflowExecution(ctx, thrift.DiagnoseWorkflowExecutionRequest(request), opts...)
	return proto.DiagnoseWorkflowExecutionResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) QueryWorkflow(ctx context.Context, request *apiv1.QueryWorkflowRequest, opts ...yarpc.CallOption) (*apiv1.QueryWorkflowResponse, error) {
	response, err := a.service.QueryWorkflow(ctx, thrift.QueryWorkflowRequest(request), opts...)
	return proto.QueryWorkflowResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) DescribeTaskList(ctx context.Context, request *apiv1.DescribeTaskListRequest, opts ...yarpc.CallOption) (*apiv1.DescribeTaskListResponse, error) {
	response, err := a.service.DescribeTaskList(ctx, thrift.DescribeTaskListRequest(request), opts...)
	return proto.DescribeTaskListResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) ListTaskListPartitions(ctx context.Context, request *apiv1.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*apiv1.ListTaskListPartitionsResponse, error) {
	response, err := a.service.ListTaskListPartitions(ctx, thrift.ListTaskListPartitionsRequest(request), opts...)
	return proto.ListTaskListPartitionsResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) GetClusterInfo(ctx context.Context, request *apiv1.GetClusterInfoRequest, opts ...yarpc.CallOption) (*apiv1.GetClusterInfoResponse, error) {
	response, err := a.service.GetClusterInfo(ctx, opts...)
	return proto.GetClusterInfoResponse(response), proto.Error(err)
}
func (a workflowAPIthriftAdapter) GetWorkflowExecutionHistory(ctx context.Context, request *apiv1.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*apiv1.GetWorkflowExecutionHistoryResponse, error) {
	response, err := a.service.GetWorkflowExecutionHistory(ctx, thrift.GetWorkflowExecutionHistoryRequest(request), opts...)
	return proto.GetWorkflowExecutionHistoryResponse(response), proto.Error(err)
}

type workerAPIthriftAdapter struct {
	service workflowserviceclient.Interface
}

func NewWorkerAPITriftAdapter(thrift workflowserviceclient.Interface) workerAPIthriftAdapter { //revive:disable-line:unexported-return this is not used at all, consider removing
	return workerAPIthriftAdapter{thrift}
}

func (a workerAPIthriftAdapter) PollForDecisionTask(ctx context.Context, request *apiv1.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*apiv1.PollForDecisionTaskResponse, error) {
	response, err := a.service.PollForDecisionTask(ctx, thrift.PollForDecisionTaskRequest(request), opts...)
	return proto.PollForDecisionTaskResponse(response), proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondDecisionTaskCompleted(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*apiv1.RespondDecisionTaskCompletedResponse, error) {
	response, err := a.service.RespondDecisionTaskCompleted(ctx, thrift.RespondDecisionTaskCompletedRequest(request), opts...)
	return proto.RespondDecisionTaskCompletedResponse(response), proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondDecisionTaskFailed(ctx context.Context, request *apiv1.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) (*apiv1.RespondDecisionTaskFailedResponse, error) {
	err := a.service.RespondDecisionTaskFailed(ctx, thrift.RespondDecisionTaskFailedRequest(request), opts...)
	return &apiv1.RespondDecisionTaskFailedResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) PollForActivityTask(ctx context.Context, request *apiv1.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*apiv1.PollForActivityTaskResponse, error) {
	response, err := a.service.PollForActivityTask(ctx, thrift.PollForActivityTaskRequest(request), opts...)
	return proto.PollForActivityTaskResponse(response), proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondActivityTaskCompleted(ctx context.Context, request *apiv1.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCompletedResponse, error) {
	err := a.service.RespondActivityTaskCompleted(ctx, thrift.RespondActivityTaskCompletedRequest(request), opts...)
	return &apiv1.RespondActivityTaskCompletedResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondActivityTaskCompletedByID(ctx context.Context, request *apiv1.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCompletedByIDResponse, error) {
	err := a.service.RespondActivityTaskCompletedByID(ctx, thrift.RespondActivityTaskCompletedByIDRequest(request), opts...)
	return &apiv1.RespondActivityTaskCompletedByIDResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondActivityTaskFailed(ctx context.Context, request *apiv1.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskFailedResponse, error) {
	err := a.service.RespondActivityTaskFailed(ctx, thrift.RespondActivityTaskFailedRequest(request), opts...)
	return &apiv1.RespondActivityTaskFailedResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondActivityTaskFailedByID(ctx context.Context, request *apiv1.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskFailedByIDResponse, error) {
	err := a.service.RespondActivityTaskFailedByID(ctx, thrift.RespondActivityTaskFailedByIDRequest(request), opts...)
	return &apiv1.RespondActivityTaskFailedByIDResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondActivityTaskCanceled(ctx context.Context, request *apiv1.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCanceledResponse, error) {
	err := a.service.RespondActivityTaskCanceled(ctx, thrift.RespondActivityTaskCanceledRequest(request), opts...)
	return &apiv1.RespondActivityTaskCanceledResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondActivityTaskCanceledByID(ctx context.Context, request *apiv1.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCanceledByIDResponse, error) {
	err := a.service.RespondActivityTaskCanceledByID(ctx, thrift.RespondActivityTaskCanceledByIDRequest(request), opts...)
	return &apiv1.RespondActivityTaskCanceledByIDResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) RecordActivityTaskHeartbeat(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*apiv1.RecordActivityTaskHeartbeatResponse, error) {
	response, err := a.service.RecordActivityTaskHeartbeat(ctx, thrift.RecordActivityTaskHeartbeatRequest(request), opts...)
	return proto.RecordActivityTaskHeartbeatResponse(response), proto.Error(err)
}
func (a workerAPIthriftAdapter) RecordActivityTaskHeartbeatByID(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*apiv1.RecordActivityTaskHeartbeatByIDResponse, error) {
	response, err := a.service.RecordActivityTaskHeartbeatByID(ctx, thrift.RecordActivityTaskHeartbeatByIDRequest(request), opts...)
	return proto.RecordActivityTaskHeartbeatByIDResponse(response), proto.Error(err)
}
func (a workerAPIthriftAdapter) RespondQueryTaskCompleted(ctx context.Context, request *apiv1.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) (*apiv1.RespondQueryTaskCompletedResponse, error) {
	err := a.service.RespondQueryTaskCompleted(ctx, thrift.RespondQueryTaskCompletedRequest(request), opts...)
	return &apiv1.RespondQueryTaskCompletedResponse{}, proto.Error(err)
}
func (a workerAPIthriftAdapter) ResetStickyTaskList(ctx context.Context, request *apiv1.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*apiv1.ResetStickyTaskListResponse, error) {
	_, err := a.service.ResetStickyTaskList(ctx, thrift.ResetStickyTaskListRequest(request), opts...)
	return &apiv1.ResetStickyTaskListResponse{}, proto.Error(err)
}

type visibilityAPIthriftAdapter struct {
	service workflowserviceclient.Interface
}

func NewVisibilityAPITriftAdapter(thrift workflowserviceclient.Interface) visibilityAPIthriftAdapter { //revive:disable-line:unexported-return this is not used at all, consider removing
	return visibilityAPIthriftAdapter{thrift}
}

func (a visibilityAPIthriftAdapter) ListWorkflowExecutions(ctx context.Context, request *apiv1.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListWorkflowExecutionsResponse, error) {
	response, err := a.service.ListWorkflowExecutions(ctx, thrift.ListWorkflowExecutionsRequest(request), opts...)
	return proto.ListWorkflowExecutionsResponse(response), proto.Error(err)
}
func (a visibilityAPIthriftAdapter) ListOpenWorkflowExecutions(ctx context.Context, request *apiv1.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListOpenWorkflowExecutionsResponse, error) {
	response, err := a.service.ListOpenWorkflowExecutions(ctx, thrift.ListOpenWorkflowExecutionsRequest(request), opts...)
	return proto.ListOpenWorkflowExecutionsResponse(response), proto.Error(err)
}
func (a visibilityAPIthriftAdapter) ListClosedWorkflowExecutions(ctx context.Context, request *apiv1.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListClosedWorkflowExecutionsResponse, error) {
	response, err := a.service.ListClosedWorkflowExecutions(ctx, thrift.ListClosedWorkflowExecutionsRequest(request), opts...)
	return proto.ListClosedWorkflowExecutionsResponse(response), proto.Error(err)
}
func (a visibilityAPIthriftAdapter) ListArchivedWorkflowExecutions(ctx context.Context, request *apiv1.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListArchivedWorkflowExecutionsResponse, error) {
	response, err := a.service.ListArchivedWorkflowExecutions(ctx, thrift.ListArchivedWorkflowExecutionsRequest(request), opts...)
	return proto.ListArchivedWorkflowExecutionsResponse(response), proto.Error(err)
}
func (a visibilityAPIthriftAdapter) ScanWorkflowExecutions(ctx context.Context, request *apiv1.ScanWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ScanWorkflowExecutionsResponse, error) {
	response, err := a.service.ScanWorkflowExecutions(ctx, thrift.ScanWorkflowExecutionsRequest(request), opts...)
	return proto.ScanWorkflowExecutionsResponse(response), proto.Error(err)
}
func (a visibilityAPIthriftAdapter) CountWorkflowExecutions(ctx context.Context, request *apiv1.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.CountWorkflowExecutionsResponse, error) {
	response, err := a.service.CountWorkflowExecutions(ctx, thrift.CountWorkflowExecutionsRequest(request), opts...)
	return proto.CountWorkflowExecutionsResponse(response), proto.Error(err)
}
func (a visibilityAPIthriftAdapter) GetSearchAttributes(ctx context.Context, request *apiv1.GetSearchAttributesRequest, opts ...yarpc.CallOption) (*apiv1.GetSearchAttributesResponse, error) {
	response, err := a.service.GetSearchAttributes(ctx, opts...)
	return proto.GetSearchAttributesResponse(response), proto.Error(err)
}
