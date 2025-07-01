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

func CountWorkflowExecutionsRequest(t *shared.CountWorkflowExecutionsRequest) *apiv1.CountWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.CountWorkflowExecutionsRequest{
		Domain: t.GetDomain(),
		Query:  t.GetQuery(),
	}
}

func DeprecateDomainRequest(t *shared.DeprecateDomainRequest) *apiv1.DeprecateDomainRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DeprecateDomainRequest{
		Name:          t.GetName(),
		SecurityToken: t.GetSecurityToken(),
	}
}

func DescribeDomainRequest(t *shared.DescribeDomainRequest) *apiv1.DescribeDomainRequest {
	if t == nil {
		return nil
	}
	if t.UUID != nil {
		return &apiv1.DescribeDomainRequest{DescribeBy: &apiv1.DescribeDomainRequest_Id{Id: *t.UUID}}
	}
	if t.Name != nil {
		return &apiv1.DescribeDomainRequest{DescribeBy: &apiv1.DescribeDomainRequest_Name{Name: *t.Name}}
	}
	panic("neither oneof field is set for DescribeDomainRequest")
}

func DescribeTaskListRequest(t *shared.DescribeTaskListRequest) *apiv1.DescribeTaskListRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DescribeTaskListRequest{
		Domain:                t.GetDomain(),
		TaskList:              TaskList(t.TaskList),
		TaskListType:          TaskListType(t.TaskListType),
		IncludeTaskListStatus: t.GetIncludeTaskListStatus(),
	}
}

func DescribeWorkflowExecutionRequest(t *shared.DescribeWorkflowExecutionRequest) *apiv1.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DescribeWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.Execution),
	}
}

func DiagnoseWorkflowExecutionRequest(t *shared.DiagnoseWorkflowExecutionRequest) *apiv1.DiagnoseWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DiagnoseWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.GetWorkflowExecution()),
		Identity:          t.GetIdentity(),
	}
}

func GetWorkflowExecutionHistoryRequest(t *shared.GetWorkflowExecutionHistoryRequest) *apiv1.GetWorkflowExecutionHistoryRequest {
	if t == nil {
		return nil
	}
	return &apiv1.GetWorkflowExecutionHistoryRequest{
		Domain:                 t.GetDomain(),
		WorkflowExecution:      WorkflowExecution(t.Execution),
		PageSize:               t.GetMaximumPageSize(),
		NextPageToken:          t.GetNextPageToken(),
		WaitForNewEvent:        t.GetWaitForNewEvent(),
		HistoryEventFilterType: EventFilterType(t.HistoryEventFilterType),
		SkipArchival:           t.GetSkipArchival(),
	}
}

func ListArchivedWorkflowExecutionsRequest(t *shared.ListArchivedWorkflowExecutionsRequest) *apiv1.ListArchivedWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListArchivedWorkflowExecutionsRequest{
		Domain:        t.GetDomain(),
		PageSize:      t.GetPageSize(),
		NextPageToken: t.GetNextPageToken(),
		Query:         t.GetQuery(),
	}
}

func ListDomainsRequest(t *shared.ListDomainsRequest) *apiv1.ListDomainsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListDomainsRequest{
		PageSize:      t.GetPageSize(),
		NextPageToken: t.NextPageToken,
	}
}

func ListTaskListPartitionsRequest(t *shared.ListTaskListPartitionsRequest) *apiv1.ListTaskListPartitionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListTaskListPartitionsRequest{
		Domain:   t.GetDomain(),
		TaskList: TaskList(t.TaskList),
	}
}

func ListWorkflowExecutionsRequest(t *shared.ListWorkflowExecutionsRequest) *apiv1.ListWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListWorkflowExecutionsRequest{
		Domain:        t.GetDomain(),
		PageSize:      t.GetPageSize(),
		NextPageToken: t.NextPageToken,
		Query:         t.GetQuery(),
	}
}

func PollForActivityTaskRequest(t *shared.PollForActivityTaskRequest) *apiv1.PollForActivityTaskRequest {
	if t == nil {
		return nil
	}
	return &apiv1.PollForActivityTaskRequest{
		Domain:           t.GetDomain(),
		TaskList:         TaskList(t.TaskList),
		Identity:         t.GetIdentity(),
		TaskListMetadata: TaskListMetadata(t.TaskListMetadata),
	}
}

func PollForDecisionTaskRequest(t *shared.PollForDecisionTaskRequest) *apiv1.PollForDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &apiv1.PollForDecisionTaskRequest{
		Domain:         t.GetDomain(),
		TaskList:       TaskList(t.TaskList),
		Identity:       t.GetIdentity(),
		BinaryChecksum: t.GetBinaryChecksum(),
	}
}

func QueryWorkflowRequest(t *shared.QueryWorkflowRequest) *apiv1.QueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &apiv1.QueryWorkflowRequest{
		Domain:                t.GetDomain(),
		WorkflowExecution:     WorkflowExecution(t.Execution),
		Query:                 WorkflowQuery(t.Query),
		QueryRejectCondition:  QueryRejectCondition(t.QueryRejectCondition),
		QueryConsistencyLevel: QueryConsistencyLevel(t.QueryConsistencyLevel),
	}
}

func RecordActivityTaskHeartbeatByIDRequest(t *shared.RecordActivityTaskHeartbeatByIDRequest) *apiv1.RecordActivityTaskHeartbeatByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RecordActivityTaskHeartbeatByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Details:           Payload(t.Details),
		Identity:          t.GetIdentity(),
	}
}

func RecordActivityTaskHeartbeatRequest(t *shared.RecordActivityTaskHeartbeatRequest) *apiv1.RecordActivityTaskHeartbeatRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RecordActivityTaskHeartbeatRequest{
		TaskToken: t.TaskToken,
		Details:   Payload(t.Details),
		Identity:  t.GetIdentity(),
	}
}

func RegisterDomainRequest(t *shared.RegisterDomainRequest) *apiv1.RegisterDomainRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RegisterDomainRequest{
		Name:                             t.GetName(),
		Description:                      t.GetDescription(),
		OwnerEmail:                       t.GetOwnerEmail(),
		WorkflowExecutionRetentionPeriod: daysToDuration(t.WorkflowExecutionRetentionPeriodInDays),
		Clusters:                         ClusterReplicationConfigurationArray(t.Clusters),
		ActiveClusterName:                t.GetActiveClusterName(),
		Data:                             t.Data,
		SecurityToken:                    t.GetSecurityToken(),
		IsGlobalDomain:                   t.GetIsGlobalDomain(),
		HistoryArchivalStatus:            ArchivalStatus(t.HistoryArchivalStatus),
		HistoryArchivalUri:               t.GetHistoryArchivalURI(),
		VisibilityArchivalStatus:         ArchivalStatus(t.VisibilityArchivalStatus),
		VisibilityArchivalUri:            t.GetVisibilityArchivalURI(),
	}
}

func RequestCancelWorkflowExecutionRequest(t *shared.RequestCancelWorkflowExecutionRequest) *apiv1.RequestCancelWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		Identity:          t.GetIdentity(),
		RequestId:         t.GetRequestId(),
	}
}

func ResetStickyTaskListRequest(t *shared.ResetStickyTaskListRequest) *apiv1.ResetStickyTaskListRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ResetStickyTaskListRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.Execution),
	}
}

func ResetWorkflowExecutionRequest(t *shared.ResetWorkflowExecutionRequest) *apiv1.ResetWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ResetWorkflowExecutionRequest{
		Domain:                t.GetDomain(),
		WorkflowExecution:     WorkflowExecution(t.WorkflowExecution),
		Reason:                t.GetReason(),
		DecisionFinishEventId: t.GetDecisionFinishEventId(),
		RequestId:             t.GetRequestId(),
		SkipSignalReapply:     t.GetSkipSignalReapply(),
	}
}

func RespondActivityTaskCanceledByIDRequest(t *shared.RespondActivityTaskCanceledByIDRequest) *apiv1.RespondActivityTaskCanceledByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCanceledByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Details:           Payload(t.Details),
		Identity:          t.GetIdentity(),
	}
}

func RespondActivityTaskCanceledRequest(t *shared.RespondActivityTaskCanceledRequest) *apiv1.RespondActivityTaskCanceledRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCanceledRequest{
		TaskToken: t.TaskToken,
		Details:   Payload(t.Details),
		Identity:  t.GetIdentity(),
	}
}

func RespondActivityTaskCompletedByIDRequest(t *shared.RespondActivityTaskCompletedByIDRequest) *apiv1.RespondActivityTaskCompletedByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCompletedByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Result:            Payload(t.Result),
		Identity:          t.GetIdentity(),
	}
}

func RespondActivityTaskCompletedRequest(t *shared.RespondActivityTaskCompletedRequest) *apiv1.RespondActivityTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCompletedRequest{
		TaskToken: t.TaskToken,
		Result:    Payload(t.Result),
		Identity:  t.GetIdentity(),
	}
}

func RespondActivityTaskFailedByIDRequest(t *shared.RespondActivityTaskFailedByIDRequest) *apiv1.RespondActivityTaskFailedByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskFailedByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Failure:           Failure(t.Reason, t.Details),
		Identity:          t.GetIdentity(),
	}
}

func RespondActivityTaskFailedRequest(t *shared.RespondActivityTaskFailedRequest) *apiv1.RespondActivityTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskFailedRequest{
		TaskToken: t.TaskToken,
		Failure:   Failure(t.Reason, t.Details),
		Identity:  t.GetIdentity(),
	}
}

func RespondDecisionTaskCompletedRequest(t *shared.RespondDecisionTaskCompletedRequest) *apiv1.RespondDecisionTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondDecisionTaskCompletedRequest{
		TaskToken:                  t.TaskToken,
		Decisions:                  DecisionArray(t.Decisions),
		ExecutionContext:           t.ExecutionContext,
		Identity:                   t.GetIdentity(),
		StickyAttributes:           StickyExecutionAttributes(t.StickyAttributes),
		ReturnNewDecisionTask:      t.GetReturnNewDecisionTask(),
		ForceCreateNewDecisionTask: t.GetForceCreateNewDecisionTask(),
		BinaryChecksum:             t.GetBinaryChecksum(),
		QueryResults:               WorkflowQueryResultMap(t.QueryResults),
	}
}

func RespondDecisionTaskFailedRequest(t *shared.RespondDecisionTaskFailedRequest) *apiv1.RespondDecisionTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondDecisionTaskFailedRequest{
		TaskToken:      t.TaskToken,
		Cause:          DecisionTaskFailedCause(t.Cause),
		Details:        Payload(t.Details),
		Identity:       t.GetIdentity(),
		BinaryChecksum: t.GetBinaryChecksum(),
	}
}

func RespondQueryTaskCompletedRequest(t *shared.RespondQueryTaskCompletedRequest) *apiv1.RespondQueryTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondQueryTaskCompletedRequest{
		TaskToken: t.TaskToken,
		Result: &apiv1.WorkflowQueryResult{
			ResultType:   QueryTaskCompletedType(t.CompletedType),
			Answer:       Payload(t.QueryResult),
			ErrorMessage: t.GetErrorMessage(),
		},
		WorkerVersionInfo: WorkerVersionInfo(t.WorkerVersionInfo),
	}
}

func ScanWorkflowExecutionsRequest(t *shared.ListWorkflowExecutionsRequest) *apiv1.ScanWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ScanWorkflowExecutionsRequest{
		Domain:        t.GetDomain(),
		PageSize:      t.GetPageSize(),
		NextPageToken: t.NextPageToken,
		Query:         t.GetQuery(),
	}
}

func SignalWithStartWorkflowExecutionRequest(t *shared.SignalWithStartWorkflowExecutionRequest) *apiv1.SignalWithStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.SignalWithStartWorkflowExecutionRequest{
		StartRequest: &apiv1.StartWorkflowExecutionRequest{
			Domain:                       t.GetDomain(),
			WorkflowId:                   t.GetWorkflowId(),
			WorkflowType:                 WorkflowType(t.WorkflowType),
			TaskList:                     TaskList(t.TaskList),
			Input:                        Payload(t.Input),
			ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
			TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
			Identity:                     t.GetIdentity(),
			RequestId:                    t.GetRequestId(),
			WorkflowIdReusePolicy:        WorkflowIDReusePolicy(t.WorkflowIdReusePolicy),
			RetryPolicy:                  RetryPolicy(t.RetryPolicy),
			CronSchedule:                 t.GetCronSchedule(),
			Memo:                         Memo(t.Memo),
			SearchAttributes:             SearchAttributes(t.SearchAttributes),
			Header:                       Header(t.Header),
			DelayStart:                   secondsToDuration(t.DelayStartSeconds),
			JitterStart:                  secondsToDuration(t.JitterStartSeconds),
			FirstRunAt:                   unixNanoToTime(t.FirstRunAtTimestamp),
		},
		SignalName:  t.GetSignalName(),
		SignalInput: Payload(t.SignalInput),
		Control:     t.Control,
	}
}

func SignalWithStartWorkflowExecutionAsyncRequest(t *shared.SignalWithStartWorkflowExecutionAsyncRequest) *apiv1.SignalWithStartWorkflowExecutionAsyncRequest {
	if t == nil {
		return nil
	}

	return &apiv1.SignalWithStartWorkflowExecutionAsyncRequest{
		Request: SignalWithStartWorkflowExecutionRequest(t.GetRequest()),
	}
}

func SignalWorkflowExecutionRequest(t *shared.SignalWorkflowExecutionRequest) *apiv1.SignalWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.SignalWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		SignalName:        t.GetSignalName(),
		SignalInput:       Payload(t.Input),
		Identity:          t.GetIdentity(),
		RequestId:         t.GetRequestId(),
		Control:           t.Control,
	}
}

func StartWorkflowExecutionRequest(t *shared.StartWorkflowExecutionRequest) *apiv1.StartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.StartWorkflowExecutionRequest{
		Domain:                       t.GetDomain(),
		WorkflowId:                   t.GetWorkflowId(),
		WorkflowType:                 WorkflowType(t.WorkflowType),
		TaskList:                     TaskList(t.TaskList),
		Input:                        Payload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		Identity:                     t.GetIdentity(),
		RequestId:                    t.GetRequestId(),
		WorkflowIdReusePolicy:        WorkflowIDReusePolicy(t.WorkflowIdReusePolicy),
		RetryPolicy:                  RetryPolicy(t.RetryPolicy),
		CronSchedule:                 t.GetCronSchedule(),
		Memo:                         Memo(t.Memo),
		SearchAttributes:             SearchAttributes(t.SearchAttributes),
		Header:                       Header(t.Header),
		DelayStart:                   secondsToDuration(t.DelayStartSeconds),
		JitterStart:                  secondsToDuration(t.JitterStartSeconds),
		FirstRunAt:                   unixNanoToTime(t.FirstRunAtTimestamp),
	}
}

func StartWorkflowExecutionAsyncRequest(t *shared.StartWorkflowExecutionAsyncRequest) *apiv1.StartWorkflowExecutionAsyncRequest {
	if t == nil {
		return nil
	}
	return &apiv1.StartWorkflowExecutionAsyncRequest{
		Request: StartWorkflowExecutionRequest(t.GetRequest()),
	}
}

func TerminateWorkflowExecutionRequest(t *shared.TerminateWorkflowExecutionRequest) *apiv1.TerminateWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.TerminateWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		Reason:            t.GetReason(),
		Details:           Payload(t.Details),
		Identity:          t.GetIdentity(),
	}
}

const (
	DomainUpdateDescriptionField              = "description"
	DomainUpdateOwnerEmailField               = "owner_email"
	DomainUpdateDataField                     = "data"
	DomainUpdateRetentionPeriodField          = "workflow_execution_retention_period"
	DomainUpdateBadBinariesField              = "bad_binaries"
	DomainUpdateHistoryArchivalStatusField    = "history_archival_status"
	DomainUpdateHistoryArchivalURIField       = "history_archival_uri"
	DomainUpdateVisibilityArchivalStatusField = "visibility_archival_status"
	DomainUpdateVisibilityArchivalURIField    = "visibility_archival_uri"
	DomainUpdateActiveClusterNameField        = "active_cluster_name"
	DomainUpdateClustersField                 = "clusters"
	DomainUpdateDeleteBadBinaryField          = "delete_bad_binary"
	DomainUpdateFailoverTimeoutField          = "failover_timeout"
)

func UpdateDomainRequest(t *shared.UpdateDomainRequest) *apiv1.UpdateDomainRequest {
	if t == nil {
		return nil
	}
	request := apiv1.UpdateDomainRequest{
		Name:          t.GetName(),
		SecurityToken: t.GetSecurityToken(),
	}
	var fields []string

	if updatedInfo := t.GetUpdatedInfo(); updatedInfo != nil {
		if updatedInfo.Description != nil {
			request.Description = *updatedInfo.Description
			fields = append(fields, DomainUpdateDescriptionField)
		}
		if updatedInfo.OwnerEmail != nil {
			request.OwnerEmail = *updatedInfo.OwnerEmail
			fields = append(fields, DomainUpdateOwnerEmailField)
		}
		if updatedInfo.Data != nil {
			request.Data = updatedInfo.Data
			fields = append(fields, DomainUpdateDataField)
		}
	}
	if configuration := t.GetConfiguration(); configuration != nil {
		if configuration.WorkflowExecutionRetentionPeriodInDays != nil {
			request.WorkflowExecutionRetentionPeriod = daysToDuration(configuration.WorkflowExecutionRetentionPeriodInDays)
			fields = append(fields, DomainUpdateRetentionPeriodField)
		}
		//if t.EmitMetric != nil {} - DEPRECATED
		if configuration.BadBinaries != nil {
			request.BadBinaries = BadBinaries(configuration.BadBinaries)
			fields = append(fields, DomainUpdateBadBinariesField)
		}
		if configuration.HistoryArchivalStatus != nil {
			request.HistoryArchivalStatus = ArchivalStatus(configuration.HistoryArchivalStatus)
			fields = append(fields, DomainUpdateHistoryArchivalStatusField)
		}
		if configuration.HistoryArchivalURI != nil {
			request.HistoryArchivalUri = *configuration.HistoryArchivalURI
			fields = append(fields, DomainUpdateHistoryArchivalURIField)
		}
		if configuration.VisibilityArchivalStatus != nil {
			request.VisibilityArchivalStatus = ArchivalStatus(configuration.VisibilityArchivalStatus)
			fields = append(fields, DomainUpdateVisibilityArchivalStatusField)
		}
		if configuration.VisibilityArchivalURI != nil {
			request.VisibilityArchivalUri = *configuration.VisibilityArchivalURI
			fields = append(fields, DomainUpdateVisibilityArchivalURIField)
		}
	}
	if replicationConfiguration := t.GetReplicationConfiguration(); replicationConfiguration != nil {
		if replicationConfiguration.ActiveClusterName != nil {
			request.ActiveClusterName = *replicationConfiguration.ActiveClusterName
			fields = append(fields, DomainUpdateActiveClusterNameField)
		}
		if replicationConfiguration.Clusters != nil {
			request.Clusters = ClusterReplicationConfigurationArray(replicationConfiguration.Clusters)
			fields = append(fields, DomainUpdateClustersField)
		}
	}
	if t.DeleteBadBinary != nil {
		request.DeleteBadBinary = *t.DeleteBadBinary
		fields = append(fields, DomainUpdateDeleteBadBinaryField)
	}
	if t.FailoverTimeoutInSeconds != nil {
		request.FailoverTimeout = secondsToDuration(t.FailoverTimeoutInSeconds)
		fields = append(fields, DomainUpdateFailoverTimeoutField)
	}

	request.UpdateMask = newFieldMask(fields)

	return &request
}

func ListClosedWorkflowExecutionsRequest(r *shared.ListClosedWorkflowExecutionsRequest) *apiv1.ListClosedWorkflowExecutionsRequest {
	if r == nil {
		return nil
	}
	request := apiv1.ListClosedWorkflowExecutionsRequest{
		Domain:          r.GetDomain(),
		PageSize:        r.GetMaximumPageSize(),
		NextPageToken:   r.NextPageToken,
		StartTimeFilter: StartTimeFilter(r.StartTimeFilter),
	}
	if r.ExecutionFilter != nil {
		request.Filters = &apiv1.ListClosedWorkflowExecutionsRequest_ExecutionFilter{
			ExecutionFilter: WorkflowExecutionFilter(r.ExecutionFilter),
		}
	}
	if r.TypeFilter != nil {
		request.Filters = &apiv1.ListClosedWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: WorkflowTypeFilter(r.TypeFilter),
		}
	}
	if r.StatusFilter != nil {
		request.Filters = &apiv1.ListClosedWorkflowExecutionsRequest_StatusFilter{
			StatusFilter: StatusFilter(r.StatusFilter),
		}
	}

	return &request
}

func ListOpenWorkflowExecutionsRequest(r *shared.ListOpenWorkflowExecutionsRequest) *apiv1.ListOpenWorkflowExecutionsRequest {
	if r == nil {
		return nil
	}
	request := apiv1.ListOpenWorkflowExecutionsRequest{
		Domain:          r.GetDomain(),
		PageSize:        r.GetMaximumPageSize(),
		NextPageToken:   r.NextPageToken,
		StartTimeFilter: StartTimeFilter(r.StartTimeFilter),
	}
	if r.ExecutionFilter != nil {
		request.Filters = &apiv1.ListOpenWorkflowExecutionsRequest_ExecutionFilter{
			ExecutionFilter: WorkflowExecutionFilter(r.ExecutionFilter),
		}
	}
	if r.TypeFilter != nil {
		request.Filters = &apiv1.ListOpenWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: WorkflowTypeFilter(r.TypeFilter),
		}
	}

	return &request
}

func RefreshWorkflowTasksRequest(r *shared.RefreshWorkflowTasksRequest) *apiv1.RefreshWorkflowTasksRequest {
	if r == nil {
		return nil
	}
	request := apiv1.RefreshWorkflowTasksRequest{
		Domain:            r.GetDomain(),
		WorkflowExecution: WorkflowExecution(r.Execution),
	}
	return &request
}

func RestartWorkflowExecutionRequest(r *shared.RestartWorkflowExecutionRequest) *apiv1.RestartWorkflowExecutionRequest {
	if r == nil {
		return nil
	}
	request := apiv1.RestartWorkflowExecutionRequest{
		Domain:            r.GetDomain(),
		WorkflowExecution: WorkflowExecution(r.GetWorkflowExecution()),
		Identity:          r.GetIdentity(),
		Reason:            r.GetReason(),
	}

	return &request
}
