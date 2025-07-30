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
	"go.uber.org/cadence/.gen/go/shared"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

func CountWorkflowExecutionsResponse(t *apiv1.CountWorkflowExecutionsResponse) *shared.CountWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.CountWorkflowExecutionsResponse{
		Count: &t.Count,
	}
}

func DescribeDomainResponse(t *apiv1.DescribeDomainResponse) *shared.DescribeDomainResponse {
	if t == nil || t.Domain == nil {
		return nil
	}
	return &shared.DescribeDomainResponse{
		DomainInfo: &shared.DomainInfo{
			Name:        &t.Domain.Name,
			Status:      DomainStatus(t.Domain.Status),
			Description: &t.Domain.Description,
			OwnerEmail:  &t.Domain.OwnerEmail,
			Data:        t.Domain.Data,
			UUID:        &t.Domain.Id,
		},
		Configuration: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: durationToDays(t.Domain.WorkflowExecutionRetentionPeriod),
			EmitMetric:                             boolPtr(true),
			BadBinaries:                            BadBinaries(t.Domain.BadBinaries),
			HistoryArchivalStatus:                  ArchivalStatus(t.Domain.HistoryArchivalStatus),
			HistoryArchivalURI:                     &t.Domain.HistoryArchivalUri,
			VisibilityArchivalStatus:               ArchivalStatus(t.Domain.VisibilityArchivalStatus),
			VisibilityArchivalURI:                  &t.Domain.VisibilityArchivalUri,
		},
		ReplicationConfiguration: &shared.DomainReplicationConfiguration{
			ActiveClusterName: &t.Domain.ActiveClusterName,
			Clusters:          ClusterReplicationConfigurationArray(t.Domain.Clusters),
			ActiveClusters:    ActiveClusters(t.Domain.ActiveClusters),
		},
		FailoverVersion: &t.Domain.FailoverVersion,
		IsGlobalDomain:  &t.Domain.IsGlobalDomain,
	}
}

func DescribeTaskListResponse(t *apiv1.DescribeTaskListResponse) *shared.DescribeTaskListResponse {
	if t == nil {
		return nil
	}
	return &shared.DescribeTaskListResponse{
		Pollers:        PollerInfoArray(t.Pollers),
		TaskListStatus: TaskListStatus(t.TaskListStatus),
	}
}

func DescribeWorkflowExecutionResponse(t *apiv1.DescribeWorkflowExecutionResponse) *shared.DescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.DescribeWorkflowExecutionResponse{
		ExecutionConfiguration: WorkflowExecutionConfiguration(t.ExecutionConfiguration),
		WorkflowExecutionInfo:  WorkflowExecutionInfo(t.WorkflowExecutionInfo),
		PendingActivities:      PendingActivityInfoArray(t.PendingActivities),
		PendingChildren:        PendingChildExecutionInfoArray(t.PendingChildren),
		PendingDecision:        PendingDecisionInfo(t.PendingDecision),
	}
}

func DiagnoseWorkflowExecutionResponse(t *apiv1.DiagnoseWorkflowExecutionResponse) *shared.DiagnoseWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.DiagnoseWorkflowExecutionResponse{
		Domain:                      Domain(t.Domain),
		DiagnosticWorkflowExecution: WorkflowExecution(t.DiagnosticWorkflowExecution),
	}
}

func GetClusterInfoResponse(t *apiv1.GetClusterInfoResponse) *shared.ClusterInfo {
	if t == nil {
		return nil
	}
	return &shared.ClusterInfo{
		SupportedClientVersions: SupportedClientVersions(t.SupportedClientVersions),
	}
}

func GetSearchAttributesResponse(t *apiv1.GetSearchAttributesResponse) *shared.GetSearchAttributesResponse {
	if t == nil {
		return nil
	}
	return &shared.GetSearchAttributesResponse{
		Keys: IndexedValueTypeMap(t.Keys),
	}
}

func GetWorkflowExecutionHistoryResponse(t *apiv1.GetWorkflowExecutionHistoryResponse) *shared.GetWorkflowExecutionHistoryResponse {
	if t == nil {
		return nil
	}
	return &shared.GetWorkflowExecutionHistoryResponse{
		History:       History(t.History),
		RawHistory:    DataBlobArray(t.RawHistory),
		NextPageToken: t.NextPageToken,
		Archived:      &t.Archived,
	}
}

func ListArchivedWorkflowExecutionsResponse(t *apiv1.ListArchivedWorkflowExecutionsResponse) *shared.ListArchivedWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListArchivedWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func ListClosedWorkflowExecutionsResponse(t *apiv1.ListClosedWorkflowExecutionsResponse) *shared.ListClosedWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListClosedWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func ListDomainsResponse(t *apiv1.ListDomainsResponse) *shared.ListDomainsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListDomainsResponse{
		Domains:       DescribeDomainResponseArray(t.Domains),
		NextPageToken: t.NextPageToken,
	}
}

func ListOpenWorkflowExecutionsResponse(t *apiv1.ListOpenWorkflowExecutionsResponse) *shared.ListOpenWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListOpenWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func ListTaskListPartitionsResponse(t *apiv1.ListTaskListPartitionsResponse) *shared.ListTaskListPartitionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListTaskListPartitionsResponse{
		ActivityTaskListPartitions: TaskListPartitionMetadataArray(t.ActivityTaskListPartitions),
		DecisionTaskListPartitions: TaskListPartitionMetadataArray(t.DecisionTaskListPartitions),
	}
}

func ListWorkflowExecutionsResponse(t *apiv1.ListWorkflowExecutionsResponse) *shared.ListWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func PollForActivityTaskResponse(t *apiv1.PollForActivityTaskResponse) *shared.PollForActivityTaskResponse {
	if t == nil {
		return nil
	}
	return &shared.PollForActivityTaskResponse{
		TaskToken:                       t.TaskToken,
		WorkflowExecution:               WorkflowExecution(t.WorkflowExecution),
		ActivityId:                      &t.ActivityId,
		ActivityType:                    ActivityType(t.ActivityType),
		Input:                           Payload(t.Input),
		ScheduledTimestamp:              timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:                timeToUnixNano(t.StartedTime),
		ScheduleToCloseTimeoutSeconds:   durationToSeconds(t.ScheduleToCloseTimeout),
		StartToCloseTimeoutSeconds:      durationToSeconds(t.StartToCloseTimeout),
		HeartbeatTimeoutSeconds:         durationToSeconds(t.HeartbeatTimeout),
		Attempt:                         &t.Attempt,
		ScheduledTimestampOfThisAttempt: timeToUnixNano(t.ScheduledTimeOfThisAttempt),
		HeartbeatDetails:                Payload(t.HeartbeatDetails),
		WorkflowType:                    WorkflowType(t.WorkflowType),
		WorkflowDomain:                  &t.WorkflowDomain,
		Header:                          Header(t.Header),
		AutoConfigHint:                  AutoConfigHint(t.AutoConfigHint),
	}
}

func PollForDecisionTaskResponse(t *apiv1.PollForDecisionTaskResponse) *shared.PollForDecisionTaskResponse {
	if t == nil {
		return nil
	}
	return &shared.PollForDecisionTaskResponse{
		TaskToken:                 t.TaskToken,
		WorkflowExecution:         WorkflowExecution(t.WorkflowExecution),
		WorkflowType:              WorkflowType(t.WorkflowType),
		PreviousStartedEventId:    toInt64Value(t.PreviousStartedEventId),
		StartedEventId:            &t.StartedEventId,
		Attempt:                   &t.Attempt,
		BacklogCountHint:          &t.BacklogCountHint,
		History:                   History(t.History),
		NextPageToken:             t.NextPageToken,
		Query:                     WorkflowQuery(t.Query),
		WorkflowExecutionTaskList: TaskList(t.WorkflowExecutionTaskList),
		ScheduledTimestamp:        timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:          timeToUnixNano(t.StartedTime),
		Queries:                   WorkflowQueryMap(t.Queries),
		NextEventId:               &t.NextEventId,
		TotalHistoryBytes:         &t.TotalHistoryBytes,
		AutoConfigHint:            AutoConfigHint(t.AutoConfigHint),
	}
}

func QueryWorkflowResponse(t *apiv1.QueryWorkflowResponse) *shared.QueryWorkflowResponse {
	if t == nil {
		return nil
	}
	return &shared.QueryWorkflowResponse{
		QueryResult:   Payload(t.QueryResult),
		QueryRejected: QueryRejected(t.QueryRejected),
	}
}

func RecordActivityTaskHeartbeatByIDResponse(t *apiv1.RecordActivityTaskHeartbeatByIDResponse) *shared.RecordActivityTaskHeartbeatResponse {
	if t == nil {
		return nil
	}
	return &shared.RecordActivityTaskHeartbeatResponse{
		CancelRequested: &t.CancelRequested,
	}
}

func RecordActivityTaskHeartbeatResponse(t *apiv1.RecordActivityTaskHeartbeatResponse) *shared.RecordActivityTaskHeartbeatResponse {
	if t == nil {
		return nil
	}
	return &shared.RecordActivityTaskHeartbeatResponse{
		CancelRequested: &t.CancelRequested,
	}
}

func ResetWorkflowExecutionResponse(t *apiv1.ResetWorkflowExecutionResponse) *shared.ResetWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.ResetWorkflowExecutionResponse{
		RunId: &t.RunId,
	}
}

func RespondDecisionTaskCompletedResponse(t *apiv1.RespondDecisionTaskCompletedResponse) *shared.RespondDecisionTaskCompletedResponse {
	if t == nil {
		return nil
	}
	return &shared.RespondDecisionTaskCompletedResponse{
		DecisionTask:                PollForDecisionTaskResponse(t.DecisionTask),
		ActivitiesToDispatchLocally: ActivityLocalDispatchInfoMap(t.ActivitiesToDispatchLocally),
	}
}

func ScanWorkflowExecutionsResponse(t *apiv1.ScanWorkflowExecutionsResponse) *shared.ListWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func SignalWithStartWorkflowExecutionResponse(t *apiv1.SignalWithStartWorkflowExecutionResponse) *shared.StartWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.StartWorkflowExecutionResponse{
		RunId: &t.RunId,
	}
}

func SignalWithStartWorkflowExecutionAsyncResponse(t *apiv1.SignalWithStartWorkflowExecutionAsyncResponse) *shared.SignalWithStartWorkflowExecutionAsyncResponse {
	if t == nil {
		return nil
	}
	return &shared.SignalWithStartWorkflowExecutionAsyncResponse{}
}

func StartWorkflowExecutionResponse(t *apiv1.StartWorkflowExecutionResponse) *shared.StartWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.StartWorkflowExecutionResponse{
		RunId: &t.RunId,
	}
}

func StartWorkflowExecutionAsyncResponse(t *apiv1.StartWorkflowExecutionAsyncResponse) *shared.StartWorkflowExecutionAsyncResponse {
	if t == nil {
		return nil
	}
	return &shared.StartWorkflowExecutionAsyncResponse{}
}

func UpdateDomainResponse(t *apiv1.UpdateDomainResponse) *shared.UpdateDomainResponse {
	if t == nil || t.Domain == nil {
		return nil
	}
	return &shared.UpdateDomainResponse{
		DomainInfo: &shared.DomainInfo{
			Name:        &t.Domain.Name,
			Status:      DomainStatus(t.Domain.Status),
			Description: &t.Domain.Description,
			OwnerEmail:  &t.Domain.OwnerEmail,
			Data:        t.Domain.Data,
			UUID:        &t.Domain.Id,
		},
		Configuration: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: durationToDays(t.Domain.WorkflowExecutionRetentionPeriod),
			EmitMetric:                             boolPtr(true),
			BadBinaries:                            BadBinaries(t.Domain.BadBinaries),
			HistoryArchivalStatus:                  ArchivalStatus(t.Domain.HistoryArchivalStatus),
			HistoryArchivalURI:                     &t.Domain.HistoryArchivalUri,
			VisibilityArchivalStatus:               ArchivalStatus(t.Domain.VisibilityArchivalStatus),
			VisibilityArchivalURI:                  &t.Domain.VisibilityArchivalUri,
		},
		ReplicationConfiguration: &shared.DomainReplicationConfiguration{
			ActiveClusterName: &t.Domain.ActiveClusterName,
			Clusters:          ClusterReplicationConfigurationArray(t.Domain.Clusters),
			ActiveClusters:    ActiveClusters(t.Domain.ActiveClusters),
		},
		FailoverVersion: &t.Domain.FailoverVersion,
		IsGlobalDomain:  &t.Domain.IsGlobalDomain,
	}
}

func RestartWorkflowExecutionResponse(t *apiv1.RestartWorkflowExecutionResponse) *shared.RestartWorkflowExecutionResponse {
	if t == nil {
		return nil
	}

	return &shared.RestartWorkflowExecutionResponse{
		RunId: &t.RunId,
	}
}
