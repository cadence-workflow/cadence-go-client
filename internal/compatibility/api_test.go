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

package compatibility

import (
	"testing"

	gogo "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/compatibility/proto"
	"go.uber.org/cadence/internal/compatibility/testdata"
	"go.uber.org/cadence/internal/compatibility/thrift"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

func TestActivityLocalDispatchInfo(t *testing.T) {
	for _, item := range []*apiv1.ActivityLocalDispatchInfo{nil, {}, &testdata.ActivityLocalDispatchInfo} {
		assert.Equal(t, item, proto.ActivityLocalDispatchInfo(thrift.ActivityLocalDispatchInfo(item)))
	}
}
func TestActivityTaskCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCancelRequestedEventAttributes{nil, {}, &testdata.ActivityTaskCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCancelRequestedEventAttributes(thrift.ActivityTaskCancelRequestedEventAttributes(item)))
	}
}
func TestActivityTaskCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCanceledEventAttributes{nil, {}, &testdata.ActivityTaskCanceledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCanceledEventAttributes(thrift.ActivityTaskCanceledEventAttributes(item)))
	}
}
func TestActivityTaskCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskCompletedEventAttributes{nil, {}, &testdata.ActivityTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskCompletedEventAttributes(thrift.ActivityTaskCompletedEventAttributes(item)))
	}
}
func TestActivityTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskFailedEventAttributes{nil, {}, &testdata.ActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskFailedEventAttributes(thrift.ActivityTaskFailedEventAttributes(item)))
	}
}
func TestActivityTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskScheduledEventAttributes{nil, {}, &testdata.ActivityTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskScheduledEventAttributes(thrift.ActivityTaskScheduledEventAttributes(item)))
	}
}
func TestActivityTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskStartedEventAttributes{nil, {}, &testdata.ActivityTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskStartedEventAttributes(thrift.ActivityTaskStartedEventAttributes(item)))
	}
}
func TestActivityTaskTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ActivityTaskTimedOutEventAttributes{nil, {}, &testdata.ActivityTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ActivityTaskTimedOutEventAttributes(thrift.ActivityTaskTimedOutEventAttributes(item)))
	}
}
func TestActivityType(t *testing.T) {
	for _, item := range []*apiv1.ActivityType{nil, {}, &testdata.ActivityType} {
		assert.Equal(t, item, proto.ActivityType(thrift.ActivityType(item)))
	}
}
func TestBadBinaries(t *testing.T) {
	for _, item := range []*apiv1.BadBinaries{nil, {}, &testdata.BadBinaries} {
		assert.Equal(t, item, proto.BadBinaries(thrift.BadBinaries(item)))
	}
}
func TestBadBinaryInfo(t *testing.T) {
	for _, item := range []*apiv1.BadBinaryInfo{nil, {}, &testdata.BadBinaryInfo} {
		assert.Equal(t, item, proto.BadBinaryInfo(thrift.BadBinaryInfo(item)))
	}
}
func TestCancelTimerFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.CancelTimerFailedEventAttributes{nil, {}, &testdata.CancelTimerFailedEventAttributes} {
		assert.Equal(t, item, proto.CancelTimerFailedEventAttributes(thrift.CancelTimerFailedEventAttributes(item)))
	}
}
func TestChildWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCanceledEventAttributes(thrift.ChildWorkflowExecutionCanceledEventAttributes(item)))
	}
}
func TestChildWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionCompletedEventAttributes(thrift.ChildWorkflowExecutionCompletedEventAttributes(item)))
	}
}
func TestChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionFailedEventAttributes(thrift.ChildWorkflowExecutionFailedEventAttributes(item)))
	}
}
func TestChildWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionStartedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionStartedEventAttributes(thrift.ChildWorkflowExecutionStartedEventAttributes(item)))
	}
}
func TestChildWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTerminatedEventAttributes(thrift.ChildWorkflowExecutionTerminatedEventAttributes(item)))
	}
}
func TestChildWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ChildWorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.ChildWorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionTimedOutEventAttributes(thrift.ChildWorkflowExecutionTimedOutEventAttributes(item)))
	}
}
func TestClusterReplicationConfiguration(t *testing.T) {
	for _, item := range []*apiv1.ClusterReplicationConfiguration{nil, {}, &testdata.ClusterReplicationConfiguration} {
		assert.Equal(t, item, proto.ClusterReplicationConfiguration(thrift.ClusterReplicationConfiguration(item)))
	}
}
func TestCountWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsRequest{nil, {}, &testdata.CountWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsRequest(thrift.CountWorkflowExecutionsRequest(item)))
	}
}
func TestCountWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.CountWorkflowExecutionsResponse{nil, {}, &testdata.CountWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.CountWorkflowExecutionsResponse(thrift.CountWorkflowExecutionsResponse(item)))
	}
}
func TestDataBlob(t *testing.T) {
	for _, item := range []*apiv1.DataBlob{nil, {}, &testdata.DataBlob} {
		assert.Equal(t, item, proto.DataBlob(thrift.DataBlob(item)))
	}
}
func TestDecisionTaskCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskCompletedEventAttributes{nil, {}, &testdata.DecisionTaskCompletedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskCompletedEventAttributes(thrift.DecisionTaskCompletedEventAttributes(item)))
	}
}
func TestDecisionTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskFailedEventAttributes{nil, {}, &testdata.DecisionTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskFailedEventAttributes(thrift.DecisionTaskFailedEventAttributes(item)))
	}
}
func TestDecisionTaskScheduledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskScheduledEventAttributes{nil, {}, &testdata.DecisionTaskScheduledEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskScheduledEventAttributes(thrift.DecisionTaskScheduledEventAttributes(item)))
	}
}
func TestDecisionTaskStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskStartedEventAttributes{nil, {}, &testdata.DecisionTaskStartedEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskStartedEventAttributes(thrift.DecisionTaskStartedEventAttributes(item)))
	}
}
func TestDecisionTaskTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.DecisionTaskTimedOutEventAttributes{nil, {}, &testdata.DecisionTaskTimedOutEventAttributes} {
		assert.Equal(t, item, proto.DecisionTaskTimedOutEventAttributes(thrift.DecisionTaskTimedOutEventAttributes(item)))
	}
}
func TestDeprecateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.DeprecateDomainRequest{nil, {}, &testdata.DeprecateDomainRequest} {
		assert.Equal(t, item, proto.DeprecateDomainRequest(thrift.DeprecateDomainRequest(item)))
	}
}
func TestDescribeDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeDomainRequest{
		&testdata.DescribeDomainRequest_ID,
		&testdata.DescribeDomainRequest_Name,
	} {
		assert.Equal(t, item, proto.DescribeDomainRequest(thrift.DescribeDomainRequest(item)))
	}
	assert.Nil(t, proto.DescribeDomainRequest(nil))
	assert.Nil(t, thrift.DescribeDomainRequest(nil))
	assert.Panics(t, func() { proto.DescribeDomainRequest(&shared.DescribeDomainRequest{}) })
	assert.Panics(t, func() { thrift.DescribeDomainRequest(&apiv1.DescribeDomainRequest{}) })
}
func TestDescribeDomainResponse_Domain(t *testing.T) {
	for _, item := range []*apiv1.Domain{nil, &testdata.Domain} {
		assert.Equal(t, item, proto.DescribeDomainResponseDomain(thrift.DescribeDomainResponseDomain(item)))
	}
}
func TestDescribeDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeDomainResponse{nil, &testdata.DescribeDomainResponse} {
		assert.Equal(t, item, proto.DescribeDomainResponse(thrift.DescribeDomainResponse(item)))
	}
}
func TestDescribeTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListRequest{nil, {}, &testdata.DescribeTaskListRequest} {
		assert.Equal(t, item, proto.DescribeTaskListRequest(thrift.DescribeTaskListRequest(item)))
	}
}
func TestDescribeTaskListResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeTaskListResponse{nil, {}, &testdata.DescribeTaskListResponse} {
		assert.Equal(t, item, proto.DescribeTaskListResponse(thrift.DescribeTaskListResponse(item)))
	}
}
func TestDescribeWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionRequest{nil, {}, &testdata.DescribeWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionRequest(thrift.DescribeWorkflowExecutionRequest(item)))
	}
}
func TestDescribeWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DescribeWorkflowExecutionResponse{nil, {}, &testdata.DescribeWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DescribeWorkflowExecutionResponse(thrift.DescribeWorkflowExecutionResponse(item)))
	}
}
func TestDiagnoseWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionRequest{nil, {}, &testdata.DiagnoseWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionRequest(thrift.DiagnoseWorkflowExecutionRequest(item)))
	}
}
func TestDiagnoseWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.DiagnoseWorkflowExecutionResponse{nil, {}, &testdata.DiagnoseWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.DiagnoseWorkflowExecutionResponse(thrift.DiagnoseWorkflowExecutionResponse(item)))
	}
}
func TestExternalWorkflowExecutionCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionCancelRequestedEventAttributes(thrift.ExternalWorkflowExecutionCancelRequestedEventAttributes(item)))
	}
}
func TestExternalWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.ExternalWorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.ExternalWorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.ExternalWorkflowExecutionSignaledEventAttributes(thrift.ExternalWorkflowExecutionSignaledEventAttributes(item)))
	}
}
func TestGetClusterInfoResponse(t *testing.T) {
	for _, item := range []*apiv1.GetClusterInfoResponse{nil, {}, &testdata.GetClusterInfoResponse} {
		assert.Equal(t, item, proto.GetClusterInfoResponse(thrift.GetClusterInfoResponse(item)))
	}
}
func TestGetSearchAttributesResponse(t *testing.T) {
	for _, item := range []*apiv1.GetSearchAttributesResponse{nil, {}, &testdata.GetSearchAttributesResponse} {
		assert.Equal(t, item, proto.GetSearchAttributesResponse(thrift.GetSearchAttributesResponse(item)))
	}
}
func TestGetWorkflowExecutionHistoryRequest(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryRequest{nil, {}, &testdata.GetWorkflowExecutionHistoryRequest} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryRequest(thrift.GetWorkflowExecutionHistoryRequest(item)))
	}
}
func TestGetWorkflowExecutionHistoryResponse(t *testing.T) {
	for _, item := range []*apiv1.GetWorkflowExecutionHistoryResponse{nil, {}, &testdata.GetWorkflowExecutionHistoryResponse} {
		assert.Equal(t, item, proto.GetWorkflowExecutionHistoryResponse(thrift.GetWorkflowExecutionHistoryResponse(item)))
	}
}
func TestHeader(t *testing.T) {
	for _, item := range []*apiv1.Header{nil, {}, &testdata.Header} {
		assert.Equal(t, item, proto.Header(thrift.Header(item)))
	}
}
func TestHistory(t *testing.T) {
	for _, item := range []*apiv1.History{nil, {}, &testdata.History} {
		assert.Equal(t, item, proto.History(thrift.History(item)))
	}
}
func TestListArchivedWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsRequest{nil, {}, &testdata.ListArchivedWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsRequest(thrift.ListArchivedWorkflowExecutionsRequest(item)))
	}
}
func TestListArchivedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListArchivedWorkflowExecutionsResponse{nil, {}, &testdata.ListArchivedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListArchivedWorkflowExecutionsResponse(thrift.ListArchivedWorkflowExecutionsResponse(item)))
	}
}
func TestListClosedWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListClosedWorkflowExecutionsResponse{nil, {}, &testdata.ListClosedWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListClosedWorkflowExecutionsResponse(thrift.ListClosedWorkflowExecutionsResponse(item)))
	}
}
func TestListDomainsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsRequest{nil, {}, &testdata.ListDomainsRequest} {
		assert.Equal(t, item, proto.ListDomainsRequest(thrift.ListDomainsRequest(item)))
	}
}
func TestListDomainsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListDomainsResponse{nil, {}, &testdata.ListDomainsResponse} {
		assert.Equal(t, item, proto.ListDomainsResponse(thrift.ListDomainsResponse(item)))
	}
}
func TestListOpenWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListOpenWorkflowExecutionsResponse{nil, {}, &testdata.ListOpenWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListOpenWorkflowExecutionsResponse(thrift.ListOpenWorkflowExecutionsResponse(item)))
	}
}
func TestListTaskListPartitionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsRequest{nil, {}, &testdata.ListTaskListPartitionsRequest} {
		assert.Equal(t, item, proto.ListTaskListPartitionsRequest(thrift.ListTaskListPartitionsRequest(item)))
	}
}
func TestListTaskListPartitionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListTaskListPartitionsResponse{nil, {}, &testdata.ListTaskListPartitionsResponse} {
		assert.Equal(t, item, proto.ListTaskListPartitionsResponse(thrift.ListTaskListPartitionsResponse(item)))
	}
}
func TestListWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsRequest{nil, {}, &testdata.ListWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsRequest(thrift.ListWorkflowExecutionsRequest(item)))
	}
}
func TestListWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ListWorkflowExecutionsResponse{nil, {}, &testdata.ListWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ListWorkflowExecutionsResponse(thrift.ListWorkflowExecutionsResponse(item)))
	}
}
func TestMarkerRecordedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.MarkerRecordedEventAttributes{nil, {}, &testdata.MarkerRecordedEventAttributes} {
		assert.Equal(t, item, proto.MarkerRecordedEventAttributes(thrift.MarkerRecordedEventAttributes(item)))
	}
}
func TestMemo(t *testing.T) {
	for _, item := range []*apiv1.Memo{nil, {}, &testdata.Memo} {
		assert.Equal(t, item, proto.Memo(thrift.Memo(item)))
	}
}
func TestPendingActivityInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingActivityInfo{nil, {}, &testdata.PendingActivityInfo} {
		assert.Equal(t, item, proto.PendingActivityInfo(thrift.PendingActivityInfo(item)))
	}
}
func TestPendingChildExecutionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingChildExecutionInfo{nil, {}, &testdata.PendingChildExecutionInfo} {
		assert.Equal(t, item, proto.PendingChildExecutionInfo(thrift.PendingChildExecutionInfo(item)))
	}
}
func TestPendingDecisionInfo(t *testing.T) {
	for _, item := range []*apiv1.PendingDecisionInfo{nil, {}, &testdata.PendingDecisionInfo} {
		assert.Equal(t, item, proto.PendingDecisionInfo(thrift.PendingDecisionInfo(item)))
	}
}
func TestPollForActivityTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskRequest{nil, {}, &testdata.PollForActivityTaskRequest} {
		assert.Equal(t, item, proto.PollForActivityTaskRequest(thrift.PollForActivityTaskRequest(item)))
	}
}
func TestPollForActivityTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForActivityTaskResponse{nil, {}, &testdata.PollForActivityTaskResponse} {
		assert.Equal(t, item, proto.PollForActivityTaskResponse(thrift.PollForActivityTaskResponse(item)))
	}
}
func TestPollForDecisionTaskRequest(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskRequest{nil, {}, &testdata.PollForDecisionTaskRequest} {
		assert.Equal(t, item, proto.PollForDecisionTaskRequest(thrift.PollForDecisionTaskRequest(item)))
	}
}
func TestPollForDecisionTaskResponse(t *testing.T) {
	for _, item := range []*apiv1.PollForDecisionTaskResponse{nil, {}, &testdata.PollForDecisionTaskResponse} {
		assert.Equal(t, item, proto.PollForDecisionTaskResponse(thrift.PollForDecisionTaskResponse(item)))
	}
}
func TestPollerInfo(t *testing.T) {
	for _, item := range []*apiv1.PollerInfo{nil, {}, &testdata.PollerInfo} {
		assert.Equal(t, item, proto.PollerInfo(thrift.PollerInfo(item)))
	}
}
func TestQueryRejected(t *testing.T) {
	for _, item := range []*apiv1.QueryRejected{nil, {}, &testdata.QueryRejected} {
		assert.Equal(t, item, proto.QueryRejected(thrift.QueryRejected(item)))
	}
}
func TestQueryWorkflowRequest(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowRequest{nil, {}, &testdata.QueryWorkflowRequest} {
		assert.Equal(t, item, proto.QueryWorkflowRequest(thrift.QueryWorkflowRequest(item)))
	}
}
func TestQueryWorkflowResponse(t *testing.T) {
	for _, item := range []*apiv1.QueryWorkflowResponse{nil, {}, &testdata.QueryWorkflowResponse} {
		assert.Equal(t, item, proto.QueryWorkflowResponse(thrift.QueryWorkflowResponse(item)))
	}
}
func TestRecordActivityTaskHeartbeatByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIdRequest(thrift.RecordActivityTaskHeartbeatByIdRequest(item)))
	}
}
func TestRecordActivityTaskHeartbeatByIDResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatByIDResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatByIDResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatByIdResponse(thrift.RecordActivityTaskHeartbeatByIdResponse(item)))
	}
}
func TestRecordActivityTaskHeartbeatRequest(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatRequest{nil, {}, &testdata.RecordActivityTaskHeartbeatRequest} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatRequest(thrift.RecordActivityTaskHeartbeatRequest(item)))
	}
}
func TestRecordActivityTaskHeartbeatResponse(t *testing.T) {
	for _, item := range []*apiv1.RecordActivityTaskHeartbeatResponse{nil, {}, &testdata.RecordActivityTaskHeartbeatResponse} {
		assert.Equal(t, item, proto.RecordActivityTaskHeartbeatResponse(thrift.RecordActivityTaskHeartbeatResponse(item)))
	}
}
func TestRegisterDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.RegisterDomainRequest{nil, &testdata.RegisterDomainRequest} {
		assert.Equal(t, item, proto.RegisterDomainRequest(thrift.RegisterDomainRequest(item)))
	}
}
func TestRequestCancelActivityTaskFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelActivityTaskFailedEventAttributes{nil, {}, &testdata.RequestCancelActivityTaskFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelActivityTaskFailedEventAttributes(thrift.RequestCancelActivityTaskFailedEventAttributes(item)))
	}
}
func TestRequestCancelExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionFailedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionFailedEventAttributes(item)))
	}
}
func TestRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(thrift.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}
}
func TestRequestCancelWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.RequestCancelWorkflowExecutionRequest{nil, {}, &testdata.RequestCancelWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.RequestCancelWorkflowExecutionRequest(thrift.RequestCancelWorkflowExecutionRequest(item)))
	}
}
func TestResetPointInfo(t *testing.T) {
	for _, item := range []*apiv1.ResetPointInfo{nil, {}, &testdata.ResetPointInfo} {
		assert.Equal(t, item, proto.ResetPointInfo(thrift.ResetPointInfo(item)))
	}
}
func TestResetPoints(t *testing.T) {
	for _, item := range []*apiv1.ResetPoints{nil, {}, &testdata.ResetPoints} {
		assert.Equal(t, item, proto.ResetPoints(thrift.ResetPoints(item)))
	}
}
func TestResetStickyTaskListRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetStickyTaskListRequest{nil, {}, &testdata.ResetStickyTaskListRequest} {
		assert.Equal(t, item, proto.ResetStickyTaskListRequest(thrift.ResetStickyTaskListRequest(item)))
	}
}
func TestResetWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionRequest{nil, {}, &testdata.ResetWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionRequest(thrift.ResetWorkflowExecutionRequest(item)))
	}
}
func TestResetWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.ResetWorkflowExecutionResponse{nil, {}, &testdata.ResetWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.ResetWorkflowExecutionResponse(thrift.ResetWorkflowExecutionResponse(item)))
	}
}
func TestRespondActivityTaskCanceledByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledByIDRequest{nil, {}, &testdata.RespondActivityTaskCanceledByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledByIdRequest(thrift.RespondActivityTaskCanceledByIdRequest(item)))
	}
}
func TestRespondActivityTaskCanceledRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCanceledRequest{nil, {}, &testdata.RespondActivityTaskCanceledRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCanceledRequest(thrift.RespondActivityTaskCanceledRequest(item)))
	}
}
func TestRespondActivityTaskCompletedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedByIDRequest{nil, {}, &testdata.RespondActivityTaskCompletedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedByIdRequest(thrift.RespondActivityTaskCompletedByIdRequest(item)))
	}
}
func TestRespondActivityTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskCompletedRequest{nil, {}, &testdata.RespondActivityTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskCompletedRequest(thrift.RespondActivityTaskCompletedRequest(item)))
	}
}
func TestRespondActivityTaskFailedByIDRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedByIDRequest{nil, {}, &testdata.RespondActivityTaskFailedByIDRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedByIdRequest(thrift.RespondActivityTaskFailedByIdRequest(item)))
	}
}
func TestRespondActivityTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondActivityTaskFailedRequest{nil, {}, &testdata.RespondActivityTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondActivityTaskFailedRequest(thrift.RespondActivityTaskFailedRequest(item)))
	}
}
func TestRespondDecisionTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedRequest{nil, {}, &testdata.RespondDecisionTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedRequest(thrift.RespondDecisionTaskCompletedRequest(item)))
	}
}
func TestRespondDecisionTaskCompletedResponse(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskCompletedResponse{nil, {}, &testdata.RespondDecisionTaskCompletedResponse} {
		assert.Equal(t, item, proto.RespondDecisionTaskCompletedResponse(thrift.RespondDecisionTaskCompletedResponse(item)))
	}
}
func TestRespondDecisionTaskFailedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondDecisionTaskFailedRequest{nil, {}, &testdata.RespondDecisionTaskFailedRequest} {
		assert.Equal(t, item, proto.RespondDecisionTaskFailedRequest(thrift.RespondDecisionTaskFailedRequest(item)))
	}
}
func TestRespondQueryTaskCompletedRequest(t *testing.T) {
	for _, item := range []*apiv1.RespondQueryTaskCompletedRequest{nil, {Result: &apiv1.WorkflowQueryResult{}}, &testdata.RespondQueryTaskCompletedRequest} {
		assert.Equal(t, item, proto.RespondQueryTaskCompletedRequest(thrift.RespondQueryTaskCompletedRequest(item)))
	}
}
func TestRetryPolicy(t *testing.T) {
	for _, item := range []*apiv1.RetryPolicy{nil, {}, &testdata.RetryPolicy} {
		assert.Equal(t, item, proto.RetryPolicy(thrift.RetryPolicy(item)))
	}
}
func TestScanWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsRequest{nil, {}, &testdata.ScanWorkflowExecutionsRequest} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsRequest(thrift.ScanWorkflowExecutionsRequest(item)))
	}
}
func TestScanWorkflowExecutionsResponse(t *testing.T) {
	for _, item := range []*apiv1.ScanWorkflowExecutionsResponse{nil, {}, &testdata.ScanWorkflowExecutionsResponse} {
		assert.Equal(t, item, proto.ScanWorkflowExecutionsResponse(thrift.ScanWorkflowExecutionsResponse(item)))
	}
}
func TestSearchAttributes(t *testing.T) {
	for _, item := range []*apiv1.SearchAttributes{nil, {}, &testdata.SearchAttributes} {
		assert.Equal(t, item, proto.SearchAttributes(thrift.SearchAttributes(item)))
	}
}
func TestSignalExternalWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionFailedEventAttributes(thrift.SignalExternalWorkflowExecutionFailedEventAttributes(item)))
	}
}
func TestSignalExternalWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.SignalExternalWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionInitiatedEventAttributes(thrift.SignalExternalWorkflowExecutionInitiatedEventAttributes(item)))
	}
}
func TestSignalWithStartWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.SignalWithStartWorkflowExecutionRequest{nil, {StartRequest: &apiv1.StartWorkflowExecutionRequest{}}, &testdata.SignalWithStartWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.SignalWithStartWorkflowExecutionRequest(thrift.SignalWithStartWorkflowExecutionRequest(item)))
	}
}
func TestSignalWithStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.SignalWithStartWorkflowExecutionResponse{nil, {}, &testdata.SignalWithStartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.SignalWithStartWorkflowExecutionResponse(thrift.SignalWithStartWorkflowExecutionResponse(item)))
	}
}
func TestSignalWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.SignalWorkflowExecutionRequest{nil, {}, &testdata.SignalWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.SignalWorkflowExecutionRequest(thrift.SignalWorkflowExecutionRequest(item)))
	}
}
func TestStartChildWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionFailedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionFailedEventAttributes(thrift.StartChildWorkflowExecutionFailedEventAttributes(item)))
	}
}
func TestStartChildWorkflowExecutionInitiatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.StartChildWorkflowExecutionInitiatedEventAttributes{nil, {}, &testdata.StartChildWorkflowExecutionInitiatedEventAttributes} {
		assert.Equal(t, item, proto.StartChildWorkflowExecutionInitiatedEventAttributes(thrift.StartChildWorkflowExecutionInitiatedEventAttributes(item)))
	}
}
func TestStartTimeFilter(t *testing.T) {
	for _, item := range []*apiv1.StartTimeFilter{nil, {}, &testdata.StartTimeFilter} {
		assert.Equal(t, item, proto.StartTimeFilter(thrift.StartTimeFilter(item)))
	}
}
func TestStartWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.StartWorkflowExecutionRequest{nil, {}, &testdata.StartWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.StartWorkflowExecutionRequest(thrift.StartWorkflowExecutionRequest(item)))
	}
}
func TestStartWorkflowExecutionResponse(t *testing.T) {
	for _, item := range []*apiv1.StartWorkflowExecutionResponse{nil, {}, &testdata.StartWorkflowExecutionResponse} {
		assert.Equal(t, item, proto.StartWorkflowExecutionResponse(thrift.StartWorkflowExecutionResponse(item)))
	}
}
func TestStatusFilter(t *testing.T) {
	for _, item := range []*apiv1.StatusFilter{nil, &testdata.StatusFilter} {
		assert.Equal(t, item, proto.StatusFilter(thrift.StatusFilter(item)))
	}
}
func TestStickyExecutionAttributes(t *testing.T) {
	for _, item := range []*apiv1.StickyExecutionAttributes{nil, {}, &testdata.StickyExecutionAttributes} {
		assert.Equal(t, item, proto.StickyExecutionAttributes(thrift.StickyExecutionAttributes(item)))
	}
}
func TestSupportedClientVersions(t *testing.T) {
	for _, item := range []*apiv1.SupportedClientVersions{nil, {}, &testdata.SupportedClientVersions} {
		assert.Equal(t, item, proto.SupportedClientVersions(thrift.SupportedClientVersions(item)))
	}
}
func TestTaskIDBlock(t *testing.T) {
	for _, item := range []*apiv1.TaskIDBlock{nil, {}, &testdata.TaskIDBlock} {
		assert.Equal(t, item, proto.TaskIdBlock(thrift.TaskIdBlock(item)))
	}
}
func TestTaskList(t *testing.T) {
	for _, item := range []*apiv1.TaskList{nil, {}, &testdata.TaskList} {
		assert.Equal(t, item, proto.TaskList(thrift.TaskList(item)))
	}
}
func TestTaskListMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListMetadata{nil, {}, &testdata.TaskListMetadata} {
		assert.Equal(t, item, proto.TaskListMetadata(thrift.TaskListMetadata(item)))
	}
}
func TestTaskListPartitionMetadata(t *testing.T) {
	for _, item := range []*apiv1.TaskListPartitionMetadata{nil, {}, &testdata.TaskListPartitionMetadata} {
		assert.Equal(t, item, proto.TaskListPartitionMetadata(thrift.TaskListPartitionMetadata(item)))
	}
}
func TestTaskListStatus(t *testing.T) {
	for _, item := range []*apiv1.TaskListStatus{nil, {}, &testdata.TaskListStatus} {
		assert.Equal(t, item, proto.TaskListStatus(thrift.TaskListStatus(item)))
	}
}
func TestTerminateWorkflowExecutionRequest(t *testing.T) {
	for _, item := range []*apiv1.TerminateWorkflowExecutionRequest{nil, {}, &testdata.TerminateWorkflowExecutionRequest} {
		assert.Equal(t, item, proto.TerminateWorkflowExecutionRequest(thrift.TerminateWorkflowExecutionRequest(item)))
	}
}
func TestTimerCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerCanceledEventAttributes{nil, {}, &testdata.TimerCanceledEventAttributes} {
		assert.Equal(t, item, proto.TimerCanceledEventAttributes(thrift.TimerCanceledEventAttributes(item)))
	}
}
func TestTimerFiredEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerFiredEventAttributes{nil, {}, &testdata.TimerFiredEventAttributes} {
		assert.Equal(t, item, proto.TimerFiredEventAttributes(thrift.TimerFiredEventAttributes(item)))
	}
}
func TestTimerStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.TimerStartedEventAttributes{nil, {}, &testdata.TimerStartedEventAttributes} {
		assert.Equal(t, item, proto.TimerStartedEventAttributes(thrift.TimerStartedEventAttributes(item)))
	}
}
func TestUpdateDomainRequest(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainRequest{nil, {UpdateMask: &gogo.FieldMask{}}, &testdata.UpdateDomainRequest} {
		assert.Equal(t, item, proto.UpdateDomainRequest(thrift.UpdateDomainRequest(item)))
	}
}
func TestUpdateDomainResponse(t *testing.T) {
	for _, item := range []*apiv1.UpdateDomainResponse{nil, &testdata.UpdateDomainResponse} {
		assert.Equal(t, item, proto.UpdateDomainResponse(thrift.UpdateDomainResponse(item)))
	}
}
func TestUpsertWorkflowSearchAttributesEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.UpsertWorkflowSearchAttributesEventAttributes{nil, {}, &testdata.UpsertWorkflowSearchAttributesEventAttributes} {
		assert.Equal(t, item, proto.UpsertWorkflowSearchAttributesEventAttributes(thrift.UpsertWorkflowSearchAttributesEventAttributes(item)))
	}
}
func TestWorkerVersionInfo(t *testing.T) {
	for _, item := range []*apiv1.WorkerVersionInfo{nil, {}, &testdata.WorkerVersionInfo} {
		assert.Equal(t, item, proto.WorkerVersionInfo(thrift.WorkerVersionInfo(item)))
	}
}
func TestWorkflowExecution(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecution{nil, {}, &testdata.WorkflowExecution} {
		assert.Equal(t, item, proto.WorkflowExecution(thrift.WorkflowExecution(item)))
	}
	assert.Empty(t, thrift.WorkflowId(nil))
	assert.Empty(t, thrift.RunId(nil))
}
func TestExternalExecutionInfo(t *testing.T) {
	assert.Nil(t, proto.ExternalExecutionInfo(nil, nil))
	assert.Nil(t, thrift.ExternalWorkflowExecution(nil))
	assert.Nil(t, thrift.ExternalInitiatedId(nil))
	assert.Panics(t, func() { proto.ExternalExecutionInfo(nil, common.Int64Ptr(testdata.EventID1)) })
	assert.Panics(t, func() { proto.ExternalExecutionInfo(thrift.WorkflowExecution(&testdata.WorkflowExecution), nil) })
	info := proto.ExternalExecutionInfo(thrift.WorkflowExecution(&testdata.WorkflowExecution), common.Int64Ptr(testdata.EventID1))
	assert.Equal(t, testdata.WorkflowExecution, *info.WorkflowExecution)
	assert.Equal(t, testdata.EventID1, info.InitiatedId)
}
func TestWorkflowExecutionCancelRequestedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCancelRequestedEventAttributes{nil, {}, &testdata.WorkflowExecutionCancelRequestedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCancelRequestedEventAttributes(thrift.WorkflowExecutionCancelRequestedEventAttributes(item)))
	}
}
func TestWorkflowExecutionCanceledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCanceledEventAttributes{nil, {}, &testdata.WorkflowExecutionCanceledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCanceledEventAttributes(thrift.WorkflowExecutionCanceledEventAttributes(item)))
	}
}
func TestWorkflowExecutionCompletedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionCompletedEventAttributes{nil, {}, &testdata.WorkflowExecutionCompletedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionCompletedEventAttributes(thrift.WorkflowExecutionCompletedEventAttributes(item)))
	}
}
func TestWorkflowExecutionConfiguration(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionConfiguration{nil, {}, &testdata.WorkflowExecutionConfiguration} {
		assert.Equal(t, item, proto.WorkflowExecutionConfiguration(thrift.WorkflowExecutionConfiguration(item)))
	}
}
func TestWorkflowExecutionContinuedAsNewEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionContinuedAsNewEventAttributes{nil, {}, &testdata.WorkflowExecutionContinuedAsNewEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionContinuedAsNewEventAttributes(thrift.WorkflowExecutionContinuedAsNewEventAttributes(item)))
	}
}
func TestWorkflowExecutionFailedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFailedEventAttributes{nil, {}, &testdata.WorkflowExecutionFailedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionFailedEventAttributes(thrift.WorkflowExecutionFailedEventAttributes(item)))
	}
}
func TestWorkflowExecutionFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionFilter{nil, {}, &testdata.WorkflowExecutionFilter} {
		assert.Equal(t, item, proto.WorkflowExecutionFilter(thrift.WorkflowExecutionFilter(item)))
	}
}
func TestParentExecutionInfo(t *testing.T) {
	assert.Nil(t, proto.ParentExecutionInfo(nil, nil, nil, nil))
	assert.Panics(t, func() { proto.ParentExecutionInfo(nil, &testdata.ParentExecutionInfo.DomainName, nil, nil) })
	info := proto.ParentExecutionInfo(nil,
		&testdata.ParentExecutionInfo.DomainName,
		thrift.WorkflowExecution(testdata.ParentExecutionInfo.WorkflowExecution),
		&testdata.ParentExecutionInfo.InitiatedId)
	assert.Equal(t, "", info.DomainId)
	assert.Equal(t, testdata.ParentExecutionInfo.DomainName, info.DomainName)
	assert.Equal(t, testdata.ParentExecutionInfo.WorkflowExecution, info.WorkflowExecution)
	assert.Equal(t, testdata.ParentExecutionInfo.InitiatedId, info.InitiatedId)
}

func TestWorkflowExecutionInfo(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionInfo{nil, {}, &testdata.WorkflowExecutionInfo} {
		assert.Equal(t, item, proto.WorkflowExecutionInfo(thrift.WorkflowExecutionInfo(item)))
	}
}
func TestWorkflowExecutionSignaledEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionSignaledEventAttributes{nil, {}, &testdata.WorkflowExecutionSignaledEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionSignaledEventAttributes(thrift.WorkflowExecutionSignaledEventAttributes(item)))
	}
}
func TestWorkflowExecutionStartedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionStartedEventAttributes{nil, {}, &testdata.WorkflowExecutionStartedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionStartedEventAttributes(thrift.WorkflowExecutionStartedEventAttributes(item)))
	}
}
func TestWorkflowExecutionTerminatedEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTerminatedEventAttributes{nil, {}, &testdata.WorkflowExecutionTerminatedEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTerminatedEventAttributes(thrift.WorkflowExecutionTerminatedEventAttributes(item)))
	}
}
func TestWorkflowExecutionTimedOutEventAttributes(t *testing.T) {
	for _, item := range []*apiv1.WorkflowExecutionTimedOutEventAttributes{nil, {}, &testdata.WorkflowExecutionTimedOutEventAttributes} {
		assert.Equal(t, item, proto.WorkflowExecutionTimedOutEventAttributes(thrift.WorkflowExecutionTimedOutEventAttributes(item)))
	}
}
func TestWorkflowQuery(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQuery{nil, {}, &testdata.WorkflowQuery} {
		assert.Equal(t, item, proto.WorkflowQuery(thrift.WorkflowQuery(item)))
	}
}
func TestWorkflowQueryResult(t *testing.T) {
	for _, item := range []*apiv1.WorkflowQueryResult{nil, {}, &testdata.WorkflowQueryResult} {
		assert.Equal(t, item, proto.WorkflowQueryResult(thrift.WorkflowQueryResult(item)))
	}
}
func TestWorkflowType(t *testing.T) {
	for _, item := range []*apiv1.WorkflowType{nil, {}, &testdata.WorkflowType} {
		assert.Equal(t, item, proto.WorkflowType(thrift.WorkflowType(item)))
	}
}
func TestWorkflowTypeFilter(t *testing.T) {
	for _, item := range []*apiv1.WorkflowTypeFilter{nil, {}, &testdata.WorkflowTypeFilter} {
		assert.Equal(t, item, proto.WorkflowTypeFilter(thrift.WorkflowTypeFilter(item)))
	}
}
func TestDataBlobArray(t *testing.T) {
	for _, item := range [][]*apiv1.DataBlob{nil, {}, testdata.DataBlobArray} {
		assert.Equal(t, item, proto.DataBlobArray(thrift.DataBlobArray(item)))
	}
}
func TestHistoryEventArray(t *testing.T) {
	for _, item := range [][]*apiv1.HistoryEvent{nil, {}, testdata.HistoryEventArray} {
		assert.Equal(t, item, proto.HistoryEventArray(thrift.HistoryEventArray(item)))
	}
}
func TestTaskListPartitionMetadataArray(t *testing.T) {
	for _, item := range [][]*apiv1.TaskListPartitionMetadata{nil, {}, testdata.TaskListPartitionMetadataArray} {
		assert.Equal(t, item, proto.TaskListPartitionMetadataArray(thrift.TaskListPartitionMetadataArray(item)))
	}
}
func TestDecisionArray(t *testing.T) {
	for _, item := range [][]*apiv1.Decision{nil, {}, testdata.DecisionArray} {
		assert.Equal(t, item, proto.DecisionArray(thrift.DecisionArray(item)))
	}
}
func TestPollerInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PollerInfo{nil, {}, testdata.PollerInfoArray} {
		assert.Equal(t, item, proto.PollerInfoArray(thrift.PollerInfoArray(item)))
	}
}
func TestPendingChildExecutionInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PendingChildExecutionInfo{nil, {}, testdata.PendingChildExecutionInfoArray} {
		assert.Equal(t, item, proto.PendingChildExecutionInfoArray(thrift.PendingChildExecutionInfoArray(item)))
	}
}
func TestWorkflowExecutionInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.WorkflowExecutionInfo{nil, {}, testdata.WorkflowExecutionInfoArray} {
		assert.Equal(t, item, proto.WorkflowExecutionInfoArray(thrift.WorkflowExecutionInfoArray(item)))
	}
}
func TestDescribeDomainResponseArray(t *testing.T) {
	for _, item := range [][]*apiv1.Domain{nil, {}, testdata.DomainArray} {
		assert.Equal(t, item, proto.DescribeDomainResponseArray(thrift.DescribeDomainResponseArray(item)))
	}
}
func TestResetPointInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.ResetPointInfo{nil, {}, testdata.ResetPointInfoArray} {
		assert.Equal(t, item, proto.ResetPointInfoArray(thrift.ResetPointInfoArray(item)))
	}
}
func TestPendingActivityInfoArray(t *testing.T) {
	for _, item := range [][]*apiv1.PendingActivityInfo{nil, {}, testdata.PendingActivityInfoArray} {
		assert.Equal(t, item, proto.PendingActivityInfoArray(thrift.PendingActivityInfoArray(item)))
	}
}
func TestClusterReplicationConfigurationArray(t *testing.T) {
	for _, item := range [][]*apiv1.ClusterReplicationConfiguration{nil, {}, testdata.ClusterReplicationConfigurationArray} {
		assert.Equal(t, item, proto.ClusterReplicationConfigurationArray(thrift.ClusterReplicationConfigurationArray(item)))
	}
}
func TestActivityLocalDispatchInfoMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.ActivityLocalDispatchInfo{nil, {}, testdata.ActivityLocalDispatchInfoMap} {
		assert.Equal(t, item, proto.ActivityLocalDispatchInfoMap(thrift.ActivityLocalDispatchInfoMap(item)))
	}
}
func TestBadBinaryInfoMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.BadBinaryInfo{nil, {}, testdata.BadBinaryInfoMap} {
		assert.Equal(t, item, proto.BadBinaryInfoMap(thrift.BadBinaryInfoMap(item)))
	}
}
func TestIndexedValueTypeMap(t *testing.T) {
	for _, item := range []map[string]apiv1.IndexedValueType{nil, {}, testdata.IndexedValueTypeMap} {
		assert.Equal(t, item, proto.IndexedValueTypeMap(thrift.IndexedValueTypeMap(item)))
	}
}
func TestWorkflowQueryMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.WorkflowQuery{nil, {}, testdata.WorkflowQueryMap} {
		assert.Equal(t, item, proto.WorkflowQueryMap(thrift.WorkflowQueryMap(item)))
	}
}
func TestWorkflowQueryResultMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.WorkflowQueryResult{nil, {}, testdata.WorkflowQueryResultMap} {
		assert.Equal(t, item, proto.WorkflowQueryResultMap(thrift.WorkflowQueryResultMap(item)))
	}
}
func TestPayload(t *testing.T) {
	for _, item := range []*apiv1.Payload{nil, &testdata.Payload1} {
		assert.Equal(t, item, proto.Payload(thrift.Payload(item)))
	}

	assert.Equal(t, &apiv1.Payload{Data: []byte{}}, proto.Payload(thrift.Payload(&apiv1.Payload{})))
}
func TestPayloadMap(t *testing.T) {
	for _, item := range []map[string]*apiv1.Payload{nil, {}, testdata.PayloadMap} {
		assert.Equal(t, item, proto.PayloadMap(thrift.PayloadMap(item)))
	}
}
func TestFailure(t *testing.T) {
	assert.Nil(t, proto.Failure(nil, nil))
	assert.Nil(t, thrift.FailureReason(nil))
	assert.Nil(t, thrift.FailureDetails(nil))
	failure := proto.Failure(&testdata.FailureReason, testdata.FailureDetails)
	assert.Equal(t, testdata.FailureReason, *thrift.FailureReason(failure))
	assert.Equal(t, testdata.FailureDetails, thrift.FailureDetails(failure))
}
func TestHistoryEvent(t *testing.T) {
	for _, item := range []*apiv1.HistoryEvent{
		nil,
		&testdata.HistoryEvent_WorkflowExecutionStarted,
		&testdata.HistoryEvent_WorkflowExecutionCompleted,
		&testdata.HistoryEvent_WorkflowExecutionFailed,
		&testdata.HistoryEvent_WorkflowExecutionTimedOut,
		&testdata.HistoryEvent_DecisionTaskScheduled,
		&testdata.HistoryEvent_DecisionTaskStarted,
		&testdata.HistoryEvent_DecisionTaskCompleted,
		&testdata.HistoryEvent_DecisionTaskTimedOut,
		&testdata.HistoryEvent_DecisionTaskFailed,
		&testdata.HistoryEvent_ActivityTaskScheduled,
		&testdata.HistoryEvent_ActivityTaskStarted,
		&testdata.HistoryEvent_ActivityTaskCompleted,
		&testdata.HistoryEvent_ActivityTaskFailed,
		&testdata.HistoryEvent_ActivityTaskTimedOut,
		&testdata.HistoryEvent_ActivityTaskCancelRequested,
		&testdata.HistoryEvent_RequestCancelActivityTaskFailed,
		&testdata.HistoryEvent_ActivityTaskCanceled,
		&testdata.HistoryEvent_TimerStarted,
		&testdata.HistoryEvent_TimerFired,
		&testdata.HistoryEvent_CancelTimerFailed,
		&testdata.HistoryEvent_TimerCanceled,
		&testdata.HistoryEvent_WorkflowExecutionCancelRequested,
		&testdata.HistoryEvent_WorkflowExecutionCanceled,
		&testdata.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiated,
		&testdata.HistoryEvent_RequestCancelExternalWorkflowExecutionFailed,
		&testdata.HistoryEvent_ExternalWorkflowExecutionCancelRequested,
		&testdata.HistoryEvent_MarkerRecorded,
		&testdata.HistoryEvent_WorkflowExecutionSignaled,
		&testdata.HistoryEvent_WorkflowExecutionTerminated,
		&testdata.HistoryEvent_WorkflowExecutionContinuedAsNew,
		&testdata.HistoryEvent_StartChildWorkflowExecutionInitiated,
		&testdata.HistoryEvent_StartChildWorkflowExecutionFailed,
		&testdata.HistoryEvent_ChildWorkflowExecutionStarted,
		&testdata.HistoryEvent_ChildWorkflowExecutionCompleted,
		&testdata.HistoryEvent_ChildWorkflowExecutionFailed,
		&testdata.HistoryEvent_ChildWorkflowExecutionCanceled,
		&testdata.HistoryEvent_ChildWorkflowExecutionTimedOut,
		&testdata.HistoryEvent_ChildWorkflowExecutionTerminated,
		&testdata.HistoryEvent_SignalExternalWorkflowExecutionInitiated,
		&testdata.HistoryEvent_SignalExternalWorkflowExecutionFailed,
		&testdata.HistoryEvent_ExternalWorkflowExecutionSignaled,
		&testdata.HistoryEvent_UpsertWorkflowSearchAttributes,
	} {
		assert.Equal(t, item, proto.HistoryEvent(thrift.HistoryEvent(item)))
	}
	assert.Panics(t, func() { proto.HistoryEvent(&shared.HistoryEvent{EventType: shared.EventType(UnknownValue).Ptr()}) })
	assert.Panics(t, func() { thrift.HistoryEvent(&apiv1.HistoryEvent{}) })
}
func TestDecision(t *testing.T) {
	for _, item := range []*apiv1.Decision{
		nil,
		&testdata.Decision_CancelTimer,
		&testdata.Decision_CancelWorkflowExecution,
		&testdata.Decision_CompleteWorkflowExecution,
		&testdata.Decision_ContinueAsNewWorkflowExecution,
		&testdata.Decision_FailWorkflowExecution,
		&testdata.Decision_RecordMarker,
		&testdata.Decision_RequestCancelActivityTask,
		&testdata.Decision_RequestCancelExternalWorkflowExecution,
		&testdata.Decision_ScheduleActivityTask,
		&testdata.Decision_SignalExternalWorkflowExecution,
		&testdata.Decision_StartChildWorkflowExecution,
		&testdata.Decision_StartTimer,
		&testdata.Decision_UpsertWorkflowSearchAttributes,
	} {
		assert.Equal(t, item, proto.Decision(thrift.Decision(item)))
	}
	assert.Panics(t, func() { proto.Decision(&shared.Decision{DecisionType: shared.DecisionType(UnknownValue).Ptr()}) })
	assert.Panics(t, func() { thrift.Decision(&apiv1.Decision{}) })
}
func TestListClosedWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListClosedWorkflowExecutionsRequest{
		nil,
		{},
		&testdata.ListClosedWorkflowExecutionsRequest_ExecutionFilter,
		&testdata.ListClosedWorkflowExecutionsRequest_StatusFilter,
		&testdata.ListClosedWorkflowExecutionsRequest_TypeFilter,
	} {
		assert.Equal(t, item, proto.ListClosedWorkflowExecutionsRequest(thrift.ListClosedWorkflowExecutionsRequest(item)))
	}
}
func TestListOpenWorkflowExecutionsRequest(t *testing.T) {
	for _, item := range []*apiv1.ListOpenWorkflowExecutionsRequest{
		nil,
		{},
		&testdata.ListOpenWorkflowExecutionsRequest_ExecutionFilter,
		&testdata.ListOpenWorkflowExecutionsRequest_TypeFilter,
	} {
		assert.Equal(t, item, proto.ListOpenWorkflowExecutionsRequest(thrift.ListOpenWorkflowExecutionsRequest(item)))
	}
}
