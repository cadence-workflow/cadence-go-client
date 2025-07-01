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
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/uber/tchannel-go/thrift"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
)

type (
	serviceWrapperSuite struct {
		suite.Suite
		Service    *workflowservicetest.MockClient
		controller *gomock.Controller
	}
)

func TestAPICalls(t *testing.T) {

	tests := map[string]struct {
		action           func(ctx context.Context, p workflowserviceclient.Interface) (interface{}, error)
		affordance       func(m *workflowservicetest.MockClient)
		expectedResponse interface{}
		expectedErr      error
	}{
		"DeprecateDomain": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.DeprecateDomain(ctx, &shared.DeprecateDomainRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().DeprecateDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
			},
		},
		"ListDomains": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ListDomains(ctx, &shared.ListDomainsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ListDomains(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.ListDomainsResponse{}, nil)
			},
			expectedResponse: &shared.ListDomainsResponse{},
		},
		"DescribeDomain": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.DescribeDomain(ctx, &shared.DescribeDomainRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.DescribeDomainResponse{}, nil)
			},
			expectedResponse: &shared.DescribeDomainResponse{},
		},
		"DescribeWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.DescribeWorkflowExecution(ctx, &shared.DescribeWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().DescribeWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.DescribeWorkflowExecutionResponse{}, nil)
			},
			expectedResponse: &shared.DescribeWorkflowExecutionResponse{},
		},
		"DiagnoseWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.DiagnoseWorkflowExecution(ctx, &shared.DiagnoseWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().DiagnoseWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.DiagnoseWorkflowExecutionResponse{}, nil)
			},
			expectedResponse: &shared.DiagnoseWorkflowExecutionResponse{},
		},
		"ListOpenWorkflowExecutions": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ListOpenWorkflowExecutions(ctx, &shared.ListOpenWorkflowExecutionsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.ListOpenWorkflowExecutionsResponse{}, nil)
			},
			expectedResponse: &shared.ListOpenWorkflowExecutionsResponse{},
		},
		"ListClosedWorkflowExecutions": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ListClosedWorkflowExecutions(ctx, &shared.ListClosedWorkflowExecutionsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.ListClosedWorkflowExecutionsResponse{}, nil)
			},
			expectedResponse: &shared.ListClosedWorkflowExecutionsResponse{},
		},
		"ListWorkflowExecutions": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ListWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.ListWorkflowExecutionsResponse{}, nil)
			},
			expectedResponse: &shared.ListWorkflowExecutionsResponse{},
		},
		"ListArchivedWorkflowExecutions": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ListArchivedWorkflowExecutions(ctx, &shared.ListArchivedWorkflowExecutionsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ListArchivedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(&shared.ListArchivedWorkflowExecutionsResponse{}, nil)
			},
			expectedResponse: &shared.ListArchivedWorkflowExecutionsResponse{},
		},
		"ScanWorkflowExecutions": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ScanWorkflowExecutions(ctx, &shared.ListWorkflowExecutionsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ScanWorkflowExecutions(gomock.Any(), &shared.ListWorkflowExecutionsRequest{}, gomock.Any()).Times(1).Return(&shared.ListWorkflowExecutionsResponse{}, nil)
			},
			expectedResponse: &shared.ListWorkflowExecutionsResponse{},
		},
		"CountWorkflowExecutions": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.CountWorkflowExecutions(ctx, &shared.CountWorkflowExecutionsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().CountWorkflowExecutions(gomock.Any(), &shared.CountWorkflowExecutionsRequest{}, gomock.Any()).Times(1).Return(&shared.CountWorkflowExecutionsResponse{}, nil)
			},
			expectedResponse: &shared.CountWorkflowExecutionsResponse{},
		},
		"PollForActivityTask": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.PollForActivityTask(ctx, &shared.PollForActivityTaskRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().PollForActivityTask(gomock.Any(), &shared.PollForActivityTaskRequest{}, gomock.Any()).Times(1).Return(&shared.PollForActivityTaskResponse{}, nil)
			},
			expectedResponse: &shared.PollForActivityTaskResponse{},
		},
		"PollForDecisionTask": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.PollForDecisionTask(ctx, &shared.PollForDecisionTaskRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().PollForDecisionTask(gomock.Any(), &shared.PollForDecisionTaskRequest{}, gomock.Any()).Times(1).Return(&shared.PollForDecisionTaskResponse{}, nil)
			},
			expectedResponse: &shared.PollForDecisionTaskResponse{},
		},
		"PollForRecordHeartbeatTask": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.RecordActivityTaskHeartbeat(ctx, &shared.RecordActivityTaskHeartbeatRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), &shared.RecordActivityTaskHeartbeatRequest{}, gomock.Any()).Times(1).Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil)
			},
			expectedResponse: &shared.RecordActivityTaskHeartbeatResponse{},
		},
		"PollForRecordHeartbeatTaskByID": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.RecordActivityTaskHeartbeatByID(ctx, &shared.RecordActivityTaskHeartbeatByIDRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RecordActivityTaskHeartbeatByID(gomock.Any(), &shared.RecordActivityTaskHeartbeatByIDRequest{}, gomock.Any()).Times(1).Return(&shared.RecordActivityTaskHeartbeatResponse{}, nil)
			},
			expectedResponse: &shared.RecordActivityTaskHeartbeatResponse{},
		},
		"RegisterDomain": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RegisterDomain(ctx, &shared.RegisterDomainRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RegisterDomain(gomock.Any(), &shared.RegisterDomainRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RequestCancelWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RequestCancelWorkflowExecution(ctx, &shared.RequestCancelWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), &shared.RequestCancelWorkflowExecutionRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RespondActivityTaskCanceled": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondActivityTaskCanceled(ctx, &shared.RespondActivityTaskCanceledRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondActivityTaskCanceled(gomock.Any(), &shared.RespondActivityTaskCanceledRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RespondActivityTaskCompleted": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondActivityTaskCompleted(ctx, &shared.RespondActivityTaskCompletedRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondActivityTaskCompleted(gomock.Any(), &shared.RespondActivityTaskCompletedRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RespondActivityTaskFailed": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondActivityTaskFailed(ctx, &shared.RespondActivityTaskFailedRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondActivityTaskFailed(gomock.Any(), &shared.RespondActivityTaskFailedRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RespondActivityTaskCompletedByID": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondActivityTaskCompletedByID(ctx, &shared.RespondActivityTaskCompletedByIDRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondActivityTaskCompletedByID(gomock.Any(), &shared.RespondActivityTaskCompletedByIDRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RespondActivityTaskCanceledByID": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondActivityTaskCanceledByID(ctx, &shared.RespondActivityTaskCanceledByIDRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondActivityTaskCanceledByID(gomock.Any(), &shared.RespondActivityTaskCanceledByIDRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RespondActivityTaskFailedByID": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondActivityTaskFailedByID(ctx, &shared.RespondActivityTaskFailedByIDRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondActivityTaskFailedByID(gomock.Any(), &shared.RespondActivityTaskFailedByIDRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"RespondDecisionTaskCompleted": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.RespondDecisionTaskCompleted(ctx, &shared.RespondDecisionTaskCompletedRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), &shared.RespondDecisionTaskCompletedRequest{}, gomock.Any()).Times(1).Return(&shared.RespondDecisionTaskCompletedResponse{}, nil)
			},
			expectedResponse: &shared.RespondDecisionTaskCompletedResponse{},
		},
		"RespondDecisionTaskFailed": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondDecisionTaskFailed(ctx, &shared.RespondDecisionTaskFailedRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondDecisionTaskFailed(gomock.Any(), &shared.RespondDecisionTaskFailedRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"SignalWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.SignalWorkflowExecution(ctx, &shared.SignalWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().SignalWorkflowExecution(gomock.Any(), &shared.SignalWorkflowExecutionRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"SignalWithStartWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.SignalWithStartWorkflowExecution(ctx, &shared.SignalWithStartWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), &shared.SignalWithStartWorkflowExecutionRequest{}, gomock.Any()).Times(1).Return(&shared.StartWorkflowExecutionResponse{}, nil)
			},
			expectedResponse: &shared.StartWorkflowExecutionResponse{},
		},
		"SignalWithStartWorkflowExecutionAsync": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.SignalWithStartWorkflowExecutionAsync(ctx, &shared.SignalWithStartWorkflowExecutionAsyncRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().SignalWithStartWorkflowExecutionAsync(gomock.Any(), &shared.SignalWithStartWorkflowExecutionAsyncRequest{}, gomock.Any()).Times(1).Return(&shared.SignalWithStartWorkflowExecutionAsyncResponse{}, nil)
			},
			expectedResponse: &shared.SignalWithStartWorkflowExecutionAsyncResponse{},
		},
		"StartWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.StartWorkflowExecution(ctx, &shared.StartWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().StartWorkflowExecution(gomock.Any(), &shared.StartWorkflowExecutionRequest{}, gomock.Any()).Times(1).Return(&shared.StartWorkflowExecutionResponse{}, nil)
			},
			expectedResponse: &shared.StartWorkflowExecutionResponse{},
		},
		"StartWorkflowExecutionAsync": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.StartWorkflowExecutionAsync(ctx, &shared.StartWorkflowExecutionAsyncRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().StartWorkflowExecutionAsync(gomock.Any(), &shared.StartWorkflowExecutionAsyncRequest{}, gomock.Any()).Times(1).Return(&shared.StartWorkflowExecutionAsyncResponse{}, nil)
			},
			expectedResponse: &shared.StartWorkflowExecutionAsyncResponse{},
		},
		"TerminateWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.TerminateWorkflowExecution(ctx, &shared.TerminateWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().TerminateWorkflowExecution(gomock.Any(), &shared.TerminateWorkflowExecutionRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"ResetWorkflowExecution": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ResetWorkflowExecution(ctx, &shared.ResetWorkflowExecutionRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ResetWorkflowExecution(gomock.Any(), &shared.ResetWorkflowExecutionRequest{}, gomock.Any()).Times(1).Return(&shared.ResetWorkflowExecutionResponse{}, nil)
			},
			expectedResponse: &shared.ResetWorkflowExecutionResponse{},
		},
		"UpdateDomain": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.UpdateDomain(ctx, &shared.UpdateDomainRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().UpdateDomain(gomock.Any(), &shared.UpdateDomainRequest{}, gomock.Any()).Times(1).Return(&shared.UpdateDomainResponse{}, nil)
			},
			expectedResponse: &shared.UpdateDomainResponse{},
		},
		"QueryWorkflow": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.QueryWorkflow(ctx, &shared.QueryWorkflowRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().QueryWorkflow(gomock.Any(), &shared.QueryWorkflowRequest{}, gomock.Any()).Times(1).Return(&shared.QueryWorkflowResponse{}, nil)
			},
			expectedResponse: &shared.QueryWorkflowResponse{},
		},
		"ResetStickyTaskList": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ResetStickyTaskList(ctx, &shared.ResetStickyTaskListRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ResetStickyTaskList(gomock.Any(), &shared.ResetStickyTaskListRequest{}, gomock.Any()).Times(1).Return(&shared.ResetStickyTaskListResponse{}, nil)
			},
			expectedResponse: &shared.ResetStickyTaskListResponse{},
		},
		"DescribeTaskList": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.DescribeTaskList(ctx, &shared.DescribeTaskListRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().DescribeTaskList(gomock.Any(), &shared.DescribeTaskListRequest{}, gomock.Any()).Times(1).Return(&shared.DescribeTaskListResponse{}, nil)
			},
			expectedResponse: &shared.DescribeTaskListResponse{},
		},
		"RespondQueryTaskCompleted": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return nil, sw.RespondQueryTaskCompleted(ctx, &shared.RespondQueryTaskCompletedRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().RespondQueryTaskCompleted(gomock.Any(), &shared.RespondQueryTaskCompletedRequest{}, gomock.Any()).Times(1).Return(nil)
			},
		},
		"GetSearchAttributes": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.GetSearchAttributes(ctx)
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().GetSearchAttributes(gomock.Any(), gomock.Any()).Times(1).Return(&shared.GetSearchAttributesResponse{}, nil)
			},
			expectedResponse: &shared.GetSearchAttributesResponse{},
		},
		"ListTaskListPartitions": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.ListTaskListPartitions(ctx, &shared.ListTaskListPartitionsRequest{})
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().ListTaskListPartitions(gomock.Any(), &shared.ListTaskListPartitionsRequest{}, gomock.Any()).Times(1).Return(&shared.ListTaskListPartitionsResponse{}, nil)
			},
			expectedResponse: &shared.ListTaskListPartitionsResponse{},
		},
		"GetClusterInfo": {
			action: func(ctx context.Context, sw workflowserviceclient.Interface) (interface{}, error) {
				return sw.GetClusterInfo(ctx)
			},
			affordance: func(m *workflowservicetest.MockClient) {
				m.EXPECT().GetClusterInfo(gomock.Any(), gomock.Any()).Times(1).Return(&shared.ClusterInfo{}, nil)
			},
			expectedResponse: &shared.ClusterInfo{},
		},
	}

	for name, td := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockClient := workflowservicetest.NewMockClient(ctrl)
			td.affordance(mockClient)
			sw := NewWorkflowServiceWrapper(mockClient, "zone-1")
			ctx, _ := thrift.NewContext(time.Minute)
			res, err := td.action(ctx, sw)
			assert.Equal(t, td.expectedResponse, res)
			assert.Equal(t, td.expectedErr, err)
		})
	}
}
