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
	"go.uber.org/cadence/internal/common"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

func CountWorkflowExecutionsRequest(t *apiv1.CountWorkflowExecutionsRequest) *shared.CountWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &shared.CountWorkflowExecutionsRequest{
		Domain: &t.Domain,
		Query:  &t.Query,
	}
}

func DeprecateDomainRequest(t *apiv1.DeprecateDomainRequest) *shared.DeprecateDomainRequest {
	if t == nil {
		return nil
	}
	return &shared.DeprecateDomainRequest{
		Name:          &t.Name,
		SecurityToken: &t.SecurityToken,
	}
}

func DescribeDomainRequest(t *apiv1.DescribeDomainRequest) *shared.DescribeDomainRequest {
	if t == nil {
		return nil
	}

	switch describeBy := t.DescribeBy.(type) {
	case *apiv1.DescribeDomainRequest_Id:
		return &shared.DescribeDomainRequest{UUID: &describeBy.Id}
	case *apiv1.DescribeDomainRequest_Name:
		return &shared.DescribeDomainRequest{Name: &describeBy.Name}
	}
	panic("unknown oneof field is set for DescribeDomainRequest")
}

func DescribeTaskListRequest(t *apiv1.DescribeTaskListRequest) *shared.DescribeTaskListRequest {
	if t == nil {
		return nil
	}
	return &shared.DescribeTaskListRequest{
		Domain:                &t.Domain,
		TaskList:              TaskList(t.TaskList),
		TaskListType:          TaskListType(t.TaskListType),
		IncludeTaskListStatus: &t.IncludeTaskListStatus,
	}
}

func DescribeWorkflowExecutionRequest(t *apiv1.DescribeWorkflowExecutionRequest) *shared.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &shared.DescribeWorkflowExecutionRequest{
		Domain:    &t.Domain,
		Execution: WorkflowExecution(t.WorkflowExecution),
	}
}

func DiagnoseWorkflowExecutionRequest(t *apiv1.DiagnoseWorkflowExecutionRequest) *shared.DiagnoseWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &shared.DiagnoseWorkflowExecutionRequest{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		Identity:          &t.Identity,
	}
}

func GetWorkflowExecutionHistoryRequest(t *apiv1.GetWorkflowExecutionHistoryRequest) *shared.GetWorkflowExecutionHistoryRequest {
	if t == nil {
		return nil
	}
	return &shared.GetWorkflowExecutionHistoryRequest{
		Domain:                 &t.Domain,
		Execution:              WorkflowExecution(t.WorkflowExecution),
		MaximumPageSize:        &t.PageSize,
		NextPageToken:          t.NextPageToken,
		WaitForNewEvent:        &t.WaitForNewEvent,
		HistoryEventFilterType: EventFilterType(t.HistoryEventFilterType),
		SkipArchival:           &t.SkipArchival,
	}
}

func ListArchivedWorkflowExecutionsRequest(t *apiv1.ListArchivedWorkflowExecutionsRequest) *shared.ListArchivedWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &shared.ListArchivedWorkflowExecutionsRequest{
		Domain:        &t.Domain,
		PageSize:      &t.PageSize,
		NextPageToken: t.NextPageToken,
		Query:         &t.Query,
	}
}

func ListDomainsRequest(t *apiv1.ListDomainsRequest) *shared.ListDomainsRequest {
	if t == nil {
		return nil
	}
	return &shared.ListDomainsRequest{
		PageSize:      &t.PageSize,
		NextPageToken: t.NextPageToken,
	}
}

func ListTaskListPartitionsRequest(t *apiv1.ListTaskListPartitionsRequest) *shared.ListTaskListPartitionsRequest {
	if t == nil {
		return nil
	}
	return &shared.ListTaskListPartitionsRequest{
		Domain:   &t.Domain,
		TaskList: TaskList(t.TaskList),
	}
}

func ListWorkflowExecutionsRequest(t *apiv1.ListWorkflowExecutionsRequest) *shared.ListWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &shared.ListWorkflowExecutionsRequest{
		Domain:        &t.Domain,
		PageSize:      &t.PageSize,
		NextPageToken: t.NextPageToken,
		Query:         &t.Query,
	}
}

func PollForActivityTaskRequest(t *apiv1.PollForActivityTaskRequest) *shared.PollForActivityTaskRequest {
	if t == nil {
		return nil
	}
	return &shared.PollForActivityTaskRequest{
		Domain:           &t.Domain,
		TaskList:         TaskList(t.TaskList),
		Identity:         &t.Identity,
		TaskListMetadata: TaskListMetadata(t.TaskListMetadata),
	}
}

func PollForDecisionTaskRequest(t *apiv1.PollForDecisionTaskRequest) *shared.PollForDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &shared.PollForDecisionTaskRequest{
		Domain:         &t.Domain,
		TaskList:       TaskList(t.TaskList),
		Identity:       &t.Identity,
		BinaryChecksum: &t.BinaryChecksum,
	}
}

func QueryWorkflowRequest(t *apiv1.QueryWorkflowRequest) *shared.QueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &shared.QueryWorkflowRequest{
		Domain:                &t.Domain,
		Execution:             WorkflowExecution(t.WorkflowExecution),
		Query:                 WorkflowQuery(t.Query),
		QueryRejectCondition:  QueryRejectCondition(t.QueryRejectCondition),
		QueryConsistencyLevel: QueryConsistencyLevel(t.QueryConsistencyLevel),
	}
}

func RecordActivityTaskHeartbeatByIDRequest(t *apiv1.RecordActivityTaskHeartbeatByIDRequest) *shared.RecordActivityTaskHeartbeatByIDRequest {
	if t == nil {
		return nil
	}
	return &shared.RecordActivityTaskHeartbeatByIDRequest{
		Domain:     &t.Domain,
		WorkflowID: WorkflowID(t.WorkflowExecution),
		RunID:      RunID(t.WorkflowExecution),
		ActivityID: &t.ActivityId,
		Details:    Payload(t.Details),
		Identity:   &t.Identity,
	}
}

func RecordActivityTaskHeartbeatRequest(t *apiv1.RecordActivityTaskHeartbeatRequest) *shared.RecordActivityTaskHeartbeatRequest {
	if t == nil {
		return nil
	}
	return &shared.RecordActivityTaskHeartbeatRequest{
		TaskToken: t.TaskToken,
		Details:   Payload(t.Details),
		Identity:  &t.Identity,
	}
}

func RegisterDomainRequest(t *apiv1.RegisterDomainRequest) *shared.RegisterDomainRequest {
	if t == nil {
		return nil
	}
	return &shared.RegisterDomainRequest{
		Name:                                   &t.Name,
		Description:                            &t.Description,
		OwnerEmail:                             &t.OwnerEmail,
		WorkflowExecutionRetentionPeriodInDays: durationToDays(t.WorkflowExecutionRetentionPeriod),
		Clusters:                               ClusterReplicationConfigurationArray(t.Clusters),
		ActiveClusterName:                      &t.ActiveClusterName,
		ActiveClustersByRegion:                 t.ActiveClustersByRegion,
		Data:                                   t.Data,
		SecurityToken:                          &t.SecurityToken,
		IsGlobalDomain:                         &t.IsGlobalDomain,
		HistoryArchivalStatus:                  ArchivalStatus(t.HistoryArchivalStatus),
		HistoryArchivalURI:                     &t.HistoryArchivalUri,
		VisibilityArchivalStatus:               ArchivalStatus(t.VisibilityArchivalStatus),
		VisibilityArchivalURI:                  &t.VisibilityArchivalUri,
	}
}

func RequestCancelWorkflowExecutionRequest(t *apiv1.RequestCancelWorkflowExecutionRequest) *shared.RequestCancelWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &shared.RequestCancelWorkflowExecutionRequest{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		Identity:          &t.Identity,
		RequestId:         &t.RequestId,
	}
}

func ResetStickyTaskListRequest(t *apiv1.ResetStickyTaskListRequest) *shared.ResetStickyTaskListRequest {
	if t == nil {
		return nil
	}
	return &shared.ResetStickyTaskListRequest{
		Domain:    &t.Domain,
		Execution: WorkflowExecution(t.WorkflowExecution),
	}
}

func ResetWorkflowExecutionRequest(t *apiv1.ResetWorkflowExecutionRequest) *shared.ResetWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &shared.ResetWorkflowExecutionRequest{
		Domain:                &t.Domain,
		WorkflowExecution:     WorkflowExecution(t.WorkflowExecution),
		Reason:                &t.Reason,
		DecisionFinishEventId: &t.DecisionFinishEventId,
		RequestId:             &t.RequestId,
		SkipSignalReapply:     &t.SkipSignalReapply,
	}
}

func RespondActivityTaskCanceledByIDRequest(t *apiv1.RespondActivityTaskCanceledByIDRequest) *shared.RespondActivityTaskCanceledByIDRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondActivityTaskCanceledByIDRequest{
		Domain:     &t.Domain,
		WorkflowID: WorkflowID(t.WorkflowExecution),
		RunID:      RunID(t.WorkflowExecution),
		ActivityID: &t.ActivityId,
		Details:    Payload(t.Details),
		Identity:   &t.Identity,
	}
}

func RespondActivityTaskCanceledRequest(t *apiv1.RespondActivityTaskCanceledRequest) *shared.RespondActivityTaskCanceledRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondActivityTaskCanceledRequest{
		TaskToken: t.TaskToken,
		Details:   Payload(t.Details),
		Identity:  &t.Identity,
	}
}

func RespondActivityTaskCompletedByIDRequest(t *apiv1.RespondActivityTaskCompletedByIDRequest) *shared.RespondActivityTaskCompletedByIDRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondActivityTaskCompletedByIDRequest{
		Domain:     &t.Domain,
		WorkflowID: WorkflowID(t.WorkflowExecution),
		RunID:      RunID(t.WorkflowExecution),
		ActivityID: &t.ActivityId,
		Result:     Payload(t.Result),
		Identity:   &t.Identity,
	}
}

func RespondActivityTaskCompletedRequest(t *apiv1.RespondActivityTaskCompletedRequest) *shared.RespondActivityTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondActivityTaskCompletedRequest{
		TaskToken: t.TaskToken,
		Result:    Payload(t.Result),
		Identity:  &t.Identity,
	}
}

func RespondActivityTaskFailedByIDRequest(t *apiv1.RespondActivityTaskFailedByIDRequest) *shared.RespondActivityTaskFailedByIDRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondActivityTaskFailedByIDRequest{
		Domain:     &t.Domain,
		WorkflowID: WorkflowID(t.WorkflowExecution),
		RunID:      RunID(t.WorkflowExecution),
		ActivityID: &t.ActivityId,
		Reason:     FailureReason(t.Failure),
		Details:    FailureDetails(t.Failure),
		Identity:   &t.Identity,
	}
}

func RespondActivityTaskFailedRequest(t *apiv1.RespondActivityTaskFailedRequest) *shared.RespondActivityTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondActivityTaskFailedRequest{
		TaskToken: t.TaskToken,
		Reason:    FailureReason(t.Failure),
		Details:   FailureDetails(t.Failure),
		Identity:  &t.Identity,
	}
}

func RespondDecisionTaskCompletedRequest(t *apiv1.RespondDecisionTaskCompletedRequest) *shared.RespondDecisionTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondDecisionTaskCompletedRequest{
		TaskToken:                  t.TaskToken,
		Decisions:                  DecisionArray(t.Decisions),
		ExecutionContext:           t.ExecutionContext,
		Identity:                   &t.Identity,
		StickyAttributes:           StickyExecutionAttributes(t.StickyAttributes),
		ReturnNewDecisionTask:      &t.ReturnNewDecisionTask,
		ForceCreateNewDecisionTask: &t.ForceCreateNewDecisionTask,
		BinaryChecksum:             &t.BinaryChecksum,
		QueryResults:               WorkflowQueryResultMap(t.QueryResults),
	}
}

func RespondDecisionTaskFailedRequest(t *apiv1.RespondDecisionTaskFailedRequest) *shared.RespondDecisionTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &shared.RespondDecisionTaskFailedRequest{
		TaskToken:      t.TaskToken,
		Cause:          DecisionTaskFailedCause(t.Cause),
		Details:        Payload(t.Details),
		Identity:       &t.Identity,
		BinaryChecksum: &t.BinaryChecksum,
	}
}

func RespondQueryTaskCompletedRequest(t *apiv1.RespondQueryTaskCompletedRequest) *shared.RespondQueryTaskCompletedRequest {
	if t == nil {
		return nil
	}
	request := &shared.RespondQueryTaskCompletedRequest{
		TaskToken:         t.TaskToken,
		WorkerVersionInfo: WorkerVersionInfo(t.WorkerVersionInfo),
	}
	if t.Result != nil {
		request.CompletedType = QueryTaskCompletedType(t.Result.ResultType)
		request.QueryResult = Payload(t.Result.Answer)
		request.ErrorMessage = &t.Result.ErrorMessage
	}
	return request
}

func ScanWorkflowExecutionsRequest(t *apiv1.ScanWorkflowExecutionsRequest) *shared.ListWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &shared.ListWorkflowExecutionsRequest{
		Domain:        &t.Domain,
		PageSize:      &t.PageSize,
		NextPageToken: t.NextPageToken,
		Query:         &t.Query,
	}
}

func SignalWithStartWorkflowExecutionRequest(t *apiv1.SignalWithStartWorkflowExecutionRequest) *shared.SignalWithStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	request := &shared.SignalWithStartWorkflowExecutionRequest{
		SignalName:  &t.SignalName,
		SignalInput: Payload(t.SignalInput),
		Control:     t.Control,
	}

	if t.StartRequest != nil {
		request.Domain = &t.StartRequest.Domain
		request.WorkflowId = &t.StartRequest.WorkflowId
		request.WorkflowType = WorkflowType(t.StartRequest.WorkflowType)
		request.TaskList = TaskList(t.StartRequest.TaskList)
		request.Input = Payload(t.StartRequest.Input)
		request.ExecutionStartToCloseTimeoutSeconds = durationToSeconds(t.StartRequest.ExecutionStartToCloseTimeout)
		request.TaskStartToCloseTimeoutSeconds = durationToSeconds(t.StartRequest.TaskStartToCloseTimeout)
		request.Identity = &t.StartRequest.Identity
		request.RequestId = &t.StartRequest.RequestId
		request.WorkflowIdReusePolicy = WorkflowIDReusePolicy(t.StartRequest.WorkflowIdReusePolicy)
		request.RetryPolicy = RetryPolicy(t.StartRequest.RetryPolicy)
		request.CronSchedule = &t.StartRequest.CronSchedule
		request.Memo = Memo(t.StartRequest.Memo)
		request.SearchAttributes = SearchAttributes(t.StartRequest.SearchAttributes)
		request.Header = Header(t.StartRequest.Header)
		request.CronOverlapPolicy = CronOverlapPolicy(t.StartRequest.CronOverlapPolicy)
		request.ActiveClusterSelectionPolicy = ActiveClusterSelectionPolicy(t.StartRequest.ActiveClusterSelectionPolicy)
	}

	return request
}

func SignalWorkflowExecutionRequest(t *apiv1.SignalWorkflowExecutionRequest) *shared.SignalWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &shared.SignalWorkflowExecutionRequest{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		SignalName:        &t.SignalName,
		Input:             Payload(t.SignalInput),
		Identity:          &t.Identity,
		RequestId:         &t.RequestId,
		Control:           t.Control,
	}
}

func StartWorkflowExecutionRequest(t *apiv1.StartWorkflowExecutionRequest) *shared.StartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &shared.StartWorkflowExecutionRequest{
		Domain:                              &t.Domain,
		WorkflowId:                          &t.WorkflowId,
		WorkflowType:                        WorkflowType(t.WorkflowType),
		TaskList:                            TaskList(t.TaskList),
		Input:                               Payload(t.Input),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		Identity:                            &t.Identity,
		RequestId:                           &t.RequestId,
		WorkflowIdReusePolicy:               WorkflowIDReusePolicy(t.WorkflowIdReusePolicy),
		RetryPolicy:                         RetryPolicy(t.RetryPolicy),
		CronSchedule:                        &t.CronSchedule,
		Memo:                                Memo(t.Memo),
		SearchAttributes:                    SearchAttributes(t.SearchAttributes),
		Header:                              Header(t.Header),
		DelayStartSeconds:                   durationToSeconds(t.DelayStart),
		JitterStartSeconds:                  durationToSeconds(t.JitterStart),
		FirstRunAtTimestamp:                 timeToUnixNano(t.FirstRunAt),
		CronOverlapPolicy:                   CronOverlapPolicy(t.CronOverlapPolicy),
		ActiveClusterSelectionPolicy:        ActiveClusterSelectionPolicy(t.ActiveClusterSelectionPolicy),
	}
}

func TerminateWorkflowExecutionRequest(t *apiv1.TerminateWorkflowExecutionRequest) *shared.TerminateWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &shared.TerminateWorkflowExecutionRequest{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		Reason:            &t.Reason,
		Details:           Payload(t.Details),
		Identity:          &t.Identity,
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
	DomainUpdateActiveClustersField           = "active_clusters"
	DomainUpdateDeleteBadBinaryField          = "delete_bad_binary"
	DomainUpdateFailoverTimeoutField          = "failover_timeout"
)

func UpdateDomainRequest(t *apiv1.UpdateDomainRequest) *shared.UpdateDomainRequest {
	if t == nil {
		return nil
	}
	request := shared.UpdateDomainRequest{
		Name:                     &t.Name,
		SecurityToken:            &t.SecurityToken,
		UpdatedInfo:              &shared.UpdateDomainInfo{},
		Configuration:            &shared.DomainConfiguration{},
		ReplicationConfiguration: &shared.DomainReplicationConfiguration{},
	}
	fs := newFieldSet(t.UpdateMask)

	if fs.isSet(DomainUpdateDescriptionField) {
		request.UpdatedInfo.Description = &t.Description
	}
	if fs.isSet(DomainUpdateOwnerEmailField) {
		request.UpdatedInfo.OwnerEmail = &t.OwnerEmail
	}
	if fs.isSet(DomainUpdateDataField) {
		request.UpdatedInfo.Data = t.Data
	}
	if fs.isSet(DomainUpdateRetentionPeriodField) {
		request.Configuration.WorkflowExecutionRetentionPeriodInDays = durationToDays(t.WorkflowExecutionRetentionPeriod)
	}
	if fs.isSet(DomainUpdateBadBinariesField) {
		request.Configuration.BadBinaries = BadBinaries(t.BadBinaries)
	}
	if fs.isSet(DomainUpdateHistoryArchivalStatusField) {
		request.Configuration.HistoryArchivalStatus = ArchivalStatus(t.HistoryArchivalStatus)
	}
	if fs.isSet(DomainUpdateHistoryArchivalURIField) {
		request.Configuration.HistoryArchivalURI = &t.HistoryArchivalUri
	}
	if fs.isSet(DomainUpdateVisibilityArchivalStatusField) {
		request.Configuration.VisibilityArchivalStatus = ArchivalStatus(t.VisibilityArchivalStatus)
	}
	if fs.isSet(DomainUpdateVisibilityArchivalURIField) {
		request.Configuration.VisibilityArchivalURI = &t.VisibilityArchivalUri
	}
	if fs.isSet(DomainUpdateActiveClusterNameField) {
		request.ReplicationConfiguration.ActiveClusterName = &t.ActiveClusterName
	}
	if fs.isSet(DomainUpdateActiveClustersField) {
		request.ReplicationConfiguration.ActiveClusters = ActiveClusters(t.ActiveClusters)
	}
	if fs.isSet(DomainUpdateClustersField) {
		request.ReplicationConfiguration.Clusters = ClusterReplicationConfigurationArray(t.Clusters)
	}

	if fs.isSet(DomainUpdateDeleteBadBinaryField) {
		request.DeleteBadBinary = &t.DeleteBadBinary
	}
	if fs.isSet(DomainUpdateFailoverTimeoutField) {
		request.FailoverTimeoutInSeconds = durationToSeconds(t.FailoverTimeout)
	}

	return &request
}

func ListClosedWorkflowExecutionsRequest(r *apiv1.ListClosedWorkflowExecutionsRequest) *shared.ListClosedWorkflowExecutionsRequest {
	if r == nil {
		return nil
	}
	return &shared.ListClosedWorkflowExecutionsRequest{
		Domain:          &r.Domain,
		MaximumPageSize: &r.PageSize,
		NextPageToken:   r.NextPageToken,
		StartTimeFilter: StartTimeFilter(r.StartTimeFilter),
		ExecutionFilter: WorkflowExecutionFilter(r.GetExecutionFilter()),
		TypeFilter:      WorkflowTypeFilter(r.GetTypeFilter()),
		StatusFilter:    WorkflowExecutionCloseStatus(r.GetStatusFilter().GetStatus()),
	}
}

func ListOpenWorkflowExecutionsRequest(r *apiv1.ListOpenWorkflowExecutionsRequest) *shared.ListOpenWorkflowExecutionsRequest {
	if r == nil {
		return nil
	}
	return &shared.ListOpenWorkflowExecutionsRequest{
		Domain:          &r.Domain,
		MaximumPageSize: &r.PageSize,
		NextPageToken:   r.NextPageToken,
		StartTimeFilter: StartTimeFilter(r.StartTimeFilter),
		ExecutionFilter: WorkflowExecutionFilter(r.GetExecutionFilter()),
		TypeFilter:      WorkflowTypeFilter(r.GetTypeFilter()),
	}
}

func ActiveClusterSelectionPolicy(t *apiv1.ActiveClusterSelectionPolicy) *shared.ActiveClusterSelectionPolicy {
	if t == nil {
		return nil
	}
	plc := &shared.ActiveClusterSelectionPolicy{
		Strategy: ActiveClusterSelectionStrategy(t.Strategy),
	}

	if plc.Strategy == nil {
		return nil
	}

	switch *plc.Strategy {
	case shared.ActiveClusterSelectionStrategyRegionSticky:
		plc.StickyRegion = common.StringPtr(t.GetActiveClusterStickyRegionConfig().GetStickyRegion())
	case shared.ActiveClusterSelectionStrategyExternalEntity:
		plc.ExternalEntityType = common.StringPtr(t.GetActiveClusterExternalEntityConfig().GetExternalEntityType())
		plc.ExternalEntityKey = common.StringPtr(t.GetActiveClusterExternalEntityConfig().GetExternalEntityKey())
	}

	return plc
}
