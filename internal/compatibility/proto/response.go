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
	"go.uber.org/cadence/.gen/go/shared"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

func CountWorkflowExecutionsResponse(t *shared.CountWorkflowExecutionsResponse) *apiv1.CountWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.CountWorkflowExecutionsResponse{
		Count: t.GetCount(),
	}
}

func DescribeDomainResponse(t *shared.DescribeDomainResponse) *apiv1.DescribeDomainResponse {
	if t == nil {
		return nil
	}
	domain := apiv1.Domain{
		FailoverVersion: t.GetFailoverVersion(),
		IsGlobalDomain:  t.GetIsGlobalDomain(),
	}
	if info := t.DomainInfo; info != nil {
		domain.Id = info.GetUUID()
		domain.Name = info.GetName()
		domain.Status = DomainStatus(info.Status)
		domain.Description = info.GetDescription()
		domain.OwnerEmail = info.GetOwnerEmail()
		domain.Data = info.Data
	}
	if config := t.Configuration; config != nil {
		domain.WorkflowExecutionRetentionPeriod = daysToDuration(config.WorkflowExecutionRetentionPeriodInDays)
		domain.BadBinaries = BadBinaries(config.BadBinaries)
		domain.HistoryArchivalStatus = ArchivalStatus(config.HistoryArchivalStatus)
		domain.HistoryArchivalUri = config.GetHistoryArchivalURI()
		domain.VisibilityArchivalStatus = ArchivalStatus(config.VisibilityArchivalStatus)
		domain.VisibilityArchivalUri = config.GetVisibilityArchivalURI()
	}
	if repl := t.ReplicationConfiguration; repl != nil {
		domain.ActiveClusterName = repl.GetActiveClusterName()
		domain.Clusters = ClusterReplicationConfigurationArray(repl.Clusters)
		domain.ActiveClusters = ActiveClusters(repl.ActiveClusters)
	}
	return &apiv1.DescribeDomainResponse{Domain: &domain}
}

func DescribeTaskListResponse(t *shared.DescribeTaskListResponse) *apiv1.DescribeTaskListResponse {
	if t == nil {
		return nil
	}
	return &apiv1.DescribeTaskListResponse{
		Pollers:        PollerInfoArray(t.Pollers),
		TaskListStatus: TaskListStatus(t.TaskListStatus),
	}
}

func DescribeWorkflowExecutionResponse(t *shared.DescribeWorkflowExecutionResponse) *apiv1.DescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &apiv1.DescribeWorkflowExecutionResponse{
		ExecutionConfiguration: WorkflowExecutionConfiguration(t.ExecutionConfiguration),
		WorkflowExecutionInfo:  WorkflowExecutionInfo(t.WorkflowExecutionInfo),
		PendingActivities:      PendingActivityInfoArray(t.PendingActivities),
		PendingChildren:        PendingChildExecutionInfoArray(t.PendingChildren),
		PendingDecision:        PendingDecisionInfo(t.PendingDecision),
	}
}

func DiagnoseWorkflowExecutionResponse(t *shared.DiagnoseWorkflowExecutionResponse) *apiv1.DiagnoseWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &apiv1.DiagnoseWorkflowExecutionResponse{
		Domain:                      t.GetDomain(),
		DiagnosticWorkflowExecution: WorkflowExecution(t.DiagnosticWorkflowExecution),
	}
}

func GetClusterInfoResponse(t *shared.ClusterInfo) *apiv1.GetClusterInfoResponse {
	if t == nil {
		return nil
	}
	return &apiv1.GetClusterInfoResponse{
		SupportedClientVersions: SupportedClientVersions(t.SupportedClientVersions),
	}
}

func GetSearchAttributesResponse(t *shared.GetSearchAttributesResponse) *apiv1.GetSearchAttributesResponse {
	if t == nil {
		return nil
	}
	return &apiv1.GetSearchAttributesResponse{
		Keys: IndexedValueTypeMap(t.Keys),
	}
}

func GetWorkflowExecutionHistoryResponse(t *shared.GetWorkflowExecutionHistoryResponse) *apiv1.GetWorkflowExecutionHistoryResponse {
	if t == nil {
		return nil
	}
	return &apiv1.GetWorkflowExecutionHistoryResponse{
		History:       History(t.History),
		RawHistory:    DataBlobArray(t.RawHistory),
		NextPageToken: t.NextPageToken,
		Archived:      t.GetArchived(),
	}
}

func ListArchivedWorkflowExecutionsResponse(t *shared.ListArchivedWorkflowExecutionsResponse) *apiv1.ListArchivedWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ListArchivedWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func ListClosedWorkflowExecutionsResponse(t *shared.ListClosedWorkflowExecutionsResponse) *apiv1.ListClosedWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ListClosedWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func ListDomainsResponse(t *shared.ListDomainsResponse) *apiv1.ListDomainsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ListDomainsResponse{
		Domains:       DescribeDomainResponseArray(t.Domains),
		NextPageToken: t.NextPageToken,
	}
}

func ListOpenWorkflowExecutionsResponse(t *shared.ListOpenWorkflowExecutionsResponse) *apiv1.ListOpenWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ListOpenWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func ListTaskListPartitionsResponse(t *shared.ListTaskListPartitionsResponse) *apiv1.ListTaskListPartitionsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ListTaskListPartitionsResponse{
		ActivityTaskListPartitions: TaskListPartitionMetadataArray(t.ActivityTaskListPartitions),
		DecisionTaskListPartitions: TaskListPartitionMetadataArray(t.DecisionTaskListPartitions),
	}
}

func ListWorkflowExecutionsResponse(t *shared.ListWorkflowExecutionsResponse) *apiv1.ListWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ListWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func PollForActivityTaskResponse(t *shared.PollForActivityTaskResponse) *apiv1.PollForActivityTaskResponse {
	if t == nil {
		return nil
	}
	return &apiv1.PollForActivityTaskResponse{
		TaskToken:                  t.TaskToken,
		WorkflowExecution:          WorkflowExecution(t.WorkflowExecution),
		ActivityId:                 t.GetActivityId(),
		ActivityType:               ActivityType(t.ActivityType),
		Input:                      Payload(t.Input),
		ScheduledTime:              unixNanoToTime(t.ScheduledTimestamp),
		StartedTime:                unixNanoToTime(t.StartedTimestamp),
		ScheduleToCloseTimeout:     secondsToDuration(t.ScheduleToCloseTimeoutSeconds),
		StartToCloseTimeout:        secondsToDuration(t.StartToCloseTimeoutSeconds),
		HeartbeatTimeout:           secondsToDuration(t.HeartbeatTimeoutSeconds),
		Attempt:                    t.GetAttempt(),
		ScheduledTimeOfThisAttempt: unixNanoToTime(t.ScheduledTimestampOfThisAttempt),
		HeartbeatDetails:           Payload(t.HeartbeatDetails),
		WorkflowType:               WorkflowType(t.WorkflowType),
		WorkflowDomain:             t.GetWorkflowDomain(),
		Header:                     Header(t.Header),
		AutoConfigHint:             AutoConfigHint(t.AutoConfigHint),
	}
}

func PollForDecisionTaskResponse(t *shared.PollForDecisionTaskResponse) *apiv1.PollForDecisionTaskResponse {
	if t == nil {
		return nil
	}
	return &apiv1.PollForDecisionTaskResponse{
		TaskToken:                 t.TaskToken,
		WorkflowExecution:         WorkflowExecution(t.WorkflowExecution),
		WorkflowType:              WorkflowType(t.WorkflowType),
		PreviousStartedEventId:    fromInt64Value(t.PreviousStartedEventId),
		StartedEventId:            t.GetStartedEventId(),
		Attempt:                   t.GetAttempt(),
		BacklogCountHint:          t.GetBacklogCountHint(),
		History:                   History(t.History),
		NextPageToken:             t.NextPageToken,
		Query:                     WorkflowQuery(t.Query),
		WorkflowExecutionTaskList: TaskList(t.WorkflowExecutionTaskList),
		ScheduledTime:             unixNanoToTime(t.ScheduledTimestamp),
		StartedTime:               unixNanoToTime(t.StartedTimestamp),
		Queries:                   WorkflowQueryMap(t.Queries),
		NextEventId:               t.GetNextEventId(),
		TotalHistoryBytes:         t.GetTotalHistoryBytes(),
		AutoConfigHint:            AutoConfigHint(t.AutoConfigHint),
	}
}

func QueryWorkflowResponse(t *shared.QueryWorkflowResponse) *apiv1.QueryWorkflowResponse {
	if t == nil {
		return nil
	}
	return &apiv1.QueryWorkflowResponse{
		QueryResult:   Payload(t.QueryResult),
		QueryRejected: QueryRejected(t.QueryRejected),
	}
}

func RecordActivityTaskHeartbeatByIDResponse(t *shared.RecordActivityTaskHeartbeatResponse) *apiv1.RecordActivityTaskHeartbeatByIDResponse {
	if t == nil {
		return nil
	}
	return &apiv1.RecordActivityTaskHeartbeatByIDResponse{
		CancelRequested: t.GetCancelRequested(),
	}
}

func RecordActivityTaskHeartbeatResponse(t *shared.RecordActivityTaskHeartbeatResponse) *apiv1.RecordActivityTaskHeartbeatResponse {
	if t == nil {
		return nil
	}
	return &apiv1.RecordActivityTaskHeartbeatResponse{
		CancelRequested: t.GetCancelRequested(),
	}
}

func ResetWorkflowExecutionResponse(t *shared.ResetWorkflowExecutionResponse) *apiv1.ResetWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ResetWorkflowExecutionResponse{
		RunId: t.GetRunId(),
	}
}

func RespondDecisionTaskCompletedResponse(t *shared.RespondDecisionTaskCompletedResponse) *apiv1.RespondDecisionTaskCompletedResponse {
	if t == nil {
		return nil
	}
	return &apiv1.RespondDecisionTaskCompletedResponse{
		DecisionTask:                PollForDecisionTaskResponse(t.DecisionTask),
		ActivitiesToDispatchLocally: ActivityLocalDispatchInfoMap(t.ActivitiesToDispatchLocally),
	}
}

func ScanWorkflowExecutionsResponse(t *shared.ListWorkflowExecutionsResponse) *apiv1.ScanWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &apiv1.ScanWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func SignalWithStartWorkflowExecutionResponse(t *shared.StartWorkflowExecutionResponse) *apiv1.SignalWithStartWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &apiv1.SignalWithStartWorkflowExecutionResponse{
		RunId: t.GetRunId(),
	}
}

func StartWorkflowExecutionResponse(t *shared.StartWorkflowExecutionResponse) *apiv1.StartWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &apiv1.StartWorkflowExecutionResponse{
		RunId: t.GetRunId(),
	}
}

func UpdateDomainResponse(t *shared.UpdateDomainResponse) *apiv1.UpdateDomainResponse {
	if t == nil {
		return nil
	}
	domain := &apiv1.Domain{
		FailoverVersion: t.GetFailoverVersion(),
		IsGlobalDomain:  t.GetIsGlobalDomain(),
	}
	if info := t.DomainInfo; info != nil {
		domain.Id = info.GetUUID()
		domain.Name = info.GetName()
		domain.Status = DomainStatus(info.Status)
		domain.Description = info.GetDescription()
		domain.OwnerEmail = info.GetOwnerEmail()
		domain.Data = info.Data
	}
	if config := t.Configuration; config != nil {
		domain.WorkflowExecutionRetentionPeriod = daysToDuration(config.WorkflowExecutionRetentionPeriodInDays)
		domain.BadBinaries = BadBinaries(config.BadBinaries)
		domain.HistoryArchivalStatus = ArchivalStatus(config.HistoryArchivalStatus)
		domain.HistoryArchivalUri = config.GetHistoryArchivalURI()
		domain.VisibilityArchivalStatus = ArchivalStatus(config.VisibilityArchivalStatus)
		domain.VisibilityArchivalUri = config.GetVisibilityArchivalURI()
	}
	if repl := t.ReplicationConfiguration; repl != nil {
		domain.ActiveClusterName = repl.GetActiveClusterName()
		domain.Clusters = ClusterReplicationConfigurationArray(repl.Clusters)
		domain.ActiveClusters = ActiveClusters(repl.ActiveClusters)
	}
	return &apiv1.UpdateDomainResponse{
		Domain: domain,
	}
}
