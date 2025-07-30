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
	gogo "github.com/gogo/protobuf/types"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

var (
	RegisterDomainRequest = apiv1.RegisterDomainRequest{
		Name:                             DomainName,
		Description:                      DomainDescription,
		OwnerEmail:                       DomainOwnerEmail,
		WorkflowExecutionRetentionPeriod: DomainRetention,
		Clusters:                         ClusterReplicationConfigurationArray,
		ActiveClusterName:                ClusterName1,
		ActiveClustersByRegion:           ActiveClustersByRegion,
		Data:                             DomainData,
		SecurityToken:                    SecurityToken,
		IsGlobalDomain:                   true,
		HistoryArchivalStatus:            ArchivalStatus,
		HistoryArchivalUri:               HistoryArchivalURI,
		VisibilityArchivalStatus:         ArchivalStatus,
		VisibilityArchivalUri:            VisibilityArchivalURI,
	}
	DescribeDomainRequest_ID = apiv1.DescribeDomainRequest{
		DescribeBy: &apiv1.DescribeDomainRequest_Id{Id: DomainID},
	}
	DescribeDomainRequest_Name = apiv1.DescribeDomainRequest{
		DescribeBy: &apiv1.DescribeDomainRequest_Name{Name: DomainName},
	}
	DescribeDomainResponse = apiv1.DescribeDomainResponse{
		Domain: &Domain,
	}
	ListDomainsRequest = apiv1.ListDomainsRequest{
		PageSize:      PageSize,
		NextPageToken: NextPageToken,
	}
	ListDomainsResponse = apiv1.ListDomainsResponse{
		Domains:       []*apiv1.Domain{&Domain},
		NextPageToken: NextPageToken,
	}
	UpdateDomainRequest = apiv1.UpdateDomainRequest{
		Name:                             DomainName,
		Description:                      DomainDescription,
		OwnerEmail:                       DomainOwnerEmail,
		Data:                             DomainData,
		WorkflowExecutionRetentionPeriod: DomainRetention,
		BadBinaries:                      &BadBinaries,
		HistoryArchivalStatus:            ArchivalStatus,
		HistoryArchivalUri:               HistoryArchivalURI,
		VisibilityArchivalStatus:         ArchivalStatus,
		VisibilityArchivalUri:            VisibilityArchivalURI,
		ActiveClusterName:                ClusterName1,
		ActiveClusters:                   ActiveClusters,
		Clusters:                         ClusterReplicationConfigurationArray,
		SecurityToken:                    SecurityToken,
		DeleteBadBinary:                  DeleteBadBinary,
		FailoverTimeout:                  Duration1,
		UpdateMask: &gogo.FieldMask{
			Paths: []string{
				"description",
				"owner_email",
				"data",
				"workflow_execution_retention_period",
				"bad_binaries",
				"history_archival_status",
				"history_archival_uri",
				"visibility_archival_status",
				"visibility_archival_uri",
				"active_cluster_name",
				"clusters",
				"active_clusters",
				"delete_bad_binary",
				"failover_timeout",
			},
		},
	}
	UpdateDomainResponse = apiv1.UpdateDomainResponse{
		Domain: &Domain,
	}
	DeprecateDomainRequest = apiv1.DeprecateDomainRequest{
		Name:          DomainName,
		SecurityToken: SecurityToken,
	}
	ListWorkflowExecutionsRequest = apiv1.ListWorkflowExecutionsRequest{
		Domain:        DomainName,
		PageSize:      PageSize,
		NextPageToken: NextPageToken,
		Query:         VisibilityQuery,
	}
	ListWorkflowExecutionsResponse = apiv1.ListWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray,
		NextPageToken: NextPageToken,
	}
	ScanWorkflowExecutionsRequest = apiv1.ScanWorkflowExecutionsRequest{
		Domain:        DomainName,
		PageSize:      PageSize,
		NextPageToken: NextPageToken,
		Query:         VisibilityQuery,
	}
	ScanWorkflowExecutionsResponse = apiv1.ScanWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray,
		NextPageToken: NextPageToken,
	}
	ListOpenWorkflowExecutionsRequest_ExecutionFilter = apiv1.ListOpenWorkflowExecutionsRequest{
		Domain:          DomainName,
		PageSize:        PageSize,
		NextPageToken:   NextPageToken,
		StartTimeFilter: &StartTimeFilter,
		Filters: &apiv1.ListOpenWorkflowExecutionsRequest_ExecutionFilter{
			ExecutionFilter: &WorkflowExecutionFilter,
		},
	}
	ListOpenWorkflowExecutionsRequest_TypeFilter = apiv1.ListOpenWorkflowExecutionsRequest{
		Domain:          DomainName,
		PageSize:        PageSize,
		NextPageToken:   NextPageToken,
		StartTimeFilter: &StartTimeFilter,
		Filters: &apiv1.ListOpenWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: &WorkflowTypeFilter,
		},
	}
	ListOpenWorkflowExecutionsResponse = apiv1.ListOpenWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray,
		NextPageToken: NextPageToken,
	}
	ListClosedWorkflowExecutionsRequest_ExecutionFilter = apiv1.ListClosedWorkflowExecutionsRequest{
		Domain:          DomainName,
		PageSize:        PageSize,
		NextPageToken:   NextPageToken,
		StartTimeFilter: &StartTimeFilter,
		Filters: &apiv1.ListClosedWorkflowExecutionsRequest_ExecutionFilter{
			ExecutionFilter: &WorkflowExecutionFilter,
		},
	}
	ListClosedWorkflowExecutionsRequest_TypeFilter = apiv1.ListClosedWorkflowExecutionsRequest{
		Domain:          DomainName,
		PageSize:        PageSize,
		NextPageToken:   NextPageToken,
		StartTimeFilter: &StartTimeFilter,
		Filters: &apiv1.ListClosedWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: &WorkflowTypeFilter,
		},
	}
	ListClosedWorkflowExecutionsRequest_StatusFilter = apiv1.ListClosedWorkflowExecutionsRequest{
		Domain:          DomainName,
		PageSize:        PageSize,
		NextPageToken:   NextPageToken,
		StartTimeFilter: &StartTimeFilter,
		Filters: &apiv1.ListClosedWorkflowExecutionsRequest_StatusFilter{
			StatusFilter: &StatusFilter,
		},
	}
	ListClosedWorkflowExecutionsResponse = apiv1.ListClosedWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray,
		NextPageToken: NextPageToken,
	}
	ListArchivedWorkflowExecutionsRequest = apiv1.ListArchivedWorkflowExecutionsRequest{
		Domain:        DomainName,
		PageSize:      PageSize,
		NextPageToken: NextPageToken,
		Query:         VisibilityQuery,
	}
	ListArchivedWorkflowExecutionsResponse = apiv1.ListArchivedWorkflowExecutionsResponse{
		Executions:    WorkflowExecutionInfoArray,
		NextPageToken: NextPageToken,
	}
	CountWorkflowExecutionsRequest = apiv1.CountWorkflowExecutionsRequest{
		Domain: DomainName,
		Query:  VisibilityQuery,
	}
	CountWorkflowExecutionsResponse = apiv1.CountWorkflowExecutionsResponse{
		Count: int64(8),
	}
	GetSearchAttributesResponse = apiv1.GetSearchAttributesResponse{
		Keys: IndexedValueTypeMap,
	}
	PollForDecisionTaskRequest = apiv1.PollForDecisionTaskRequest{
		Domain:         DomainName,
		TaskList:       &TaskList,
		Identity:       Identity,
		BinaryChecksum: Checksum,
	}
	PollForDecisionTaskResponse = apiv1.PollForDecisionTaskResponse{
		TaskToken:                 TaskToken,
		WorkflowExecution:         &WorkflowExecution,
		WorkflowType:              &WorkflowType,
		PreviousStartedEventId:    &gogo.Int64Value{Value: EventID1},
		StartedEventId:            EventID2,
		Attempt:                   Attempt,
		BacklogCountHint:          BacklogCountHint,
		History:                   &History,
		NextPageToken:             NextPageToken,
		Query:                     &WorkflowQuery,
		WorkflowExecutionTaskList: &TaskList,
		ScheduledTime:             Timestamp1,
		StartedTime:               Timestamp2,
		Queries:                   WorkflowQueryMap,
		NextEventId:               EventID3,
		AutoConfigHint:            &AutoConfigHint,
	}
	RespondDecisionTaskCompletedRequest = apiv1.RespondDecisionTaskCompletedRequest{
		TaskToken:                  TaskToken,
		Decisions:                  DecisionArray,
		ExecutionContext:           ExecutionContext,
		Identity:                   Identity,
		StickyAttributes:           &StickyExecutionAttributes,
		ReturnNewDecisionTask:      true,
		ForceCreateNewDecisionTask: true,
		BinaryChecksum:             Checksum,
		QueryResults:               WorkflowQueryResultMap,
	}
	RespondDecisionTaskCompletedResponse = apiv1.RespondDecisionTaskCompletedResponse{
		DecisionTask:                &PollForDecisionTaskResponse,
		ActivitiesToDispatchLocally: ActivityLocalDispatchInfoMap,
	}
	RespondDecisionTaskFailedRequest = apiv1.RespondDecisionTaskFailedRequest{
		TaskToken:      TaskToken,
		Cause:          DecisionTaskFailedCause,
		Details:        &Payload1,
		Identity:       Identity,
		BinaryChecksum: Checksum,
	}
	PollForActivityTaskRequest = apiv1.PollForActivityTaskRequest{
		Domain:           DomainName,
		TaskList:         &TaskList,
		Identity:         Identity,
		TaskListMetadata: &TaskListMetadata,
	}
	PollForActivityTaskResponse = apiv1.PollForActivityTaskResponse{
		TaskToken:                  TaskToken,
		WorkflowExecution:          &WorkflowExecution,
		ActivityId:                 ActivityID,
		ActivityType:               &ActivityType,
		Input:                      &Payload1,
		ScheduledTime:              Timestamp1,
		ScheduleToCloseTimeout:     Duration1,
		StartedTime:                Timestamp2,
		StartToCloseTimeout:        Duration2,
		HeartbeatTimeout:           Duration3,
		Attempt:                    Attempt,
		ScheduledTimeOfThisAttempt: Timestamp3,
		HeartbeatDetails:           &Payload2,
		WorkflowType:               &WorkflowType,
		WorkflowDomain:             DomainName,
		Header:                     &Header,
		AutoConfigHint:             &AutoConfigHint,
	}
	RespondActivityTaskCompletedRequest = apiv1.RespondActivityTaskCompletedRequest{
		TaskToken: TaskToken,
		Result:    &Payload1,
		Identity:  Identity,
	}
	RespondActivityTaskCompletedByIDRequest = apiv1.RespondActivityTaskCompletedByIDRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		ActivityId:        ActivityID,
		Result:            &Payload1,
		Identity:          Identity,
	}
	RespondActivityTaskFailedRequest = apiv1.RespondActivityTaskFailedRequest{
		TaskToken: TaskToken,
		Failure:   &Failure,
		Identity:  Identity,
	}
	RespondActivityTaskFailedByIDRequest = apiv1.RespondActivityTaskFailedByIDRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		ActivityId:        ActivityID,
		Failure:           &Failure,
		Identity:          Identity,
	}
	RespondActivityTaskCanceledRequest = apiv1.RespondActivityTaskCanceledRequest{
		TaskToken: TaskToken,
		Details:   &Payload1,
		Identity:  Identity,
	}
	RespondActivityTaskCanceledByIDRequest = apiv1.RespondActivityTaskCanceledByIDRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		ActivityId:        ActivityID,
		Details:           &Payload1,
		Identity:          Identity,
	}
	RecordActivityTaskHeartbeatRequest = apiv1.RecordActivityTaskHeartbeatRequest{
		TaskToken: TaskToken,
		Details:   &Payload1,
		Identity:  Identity,
	}
	RecordActivityTaskHeartbeatResponse = apiv1.RecordActivityTaskHeartbeatResponse{
		CancelRequested: true,
	}
	RecordActivityTaskHeartbeatByIDResponse = apiv1.RecordActivityTaskHeartbeatByIDResponse{
		CancelRequested: true,
	}
	RecordActivityTaskHeartbeatByIDRequest = apiv1.RecordActivityTaskHeartbeatByIDRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		ActivityId:        ActivityID,
		Details:           &Payload1,
		Identity:          Identity,
	}
	RespondQueryTaskCompletedRequest = apiv1.RespondQueryTaskCompletedRequest{
		TaskToken: TaskToken,
		Result: &apiv1.WorkflowQueryResult{
			ResultType:   QueryResultType,
			Answer:       &Payload1,
			ErrorMessage: ErrorMessage,
		},
		WorkerVersionInfo: &WorkerVersionInfo,
	}
	RequestCancelWorkflowExecutionRequest = apiv1.RequestCancelWorkflowExecutionRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		Identity:          Identity,
		RequestId:         RequestID,
	}
	StartWorkflowExecutionRequest = apiv1.StartWorkflowExecutionRequest{
		Domain:                       DomainName,
		WorkflowId:                   WorkflowID,
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		Identity:                     Identity,
		RequestId:                    RequestID,
		WorkflowIdReusePolicy:        WorkflowIDReusePolicy,
		RetryPolicy:                  &RetryPolicy,
		CronSchedule:                 CronSchedule,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
		Header:                       &Header,
	}
	StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy1 = apiv1.StartWorkflowExecutionRequest{
		Domain:                       DomainName,
		WorkflowId:                   WorkflowID,
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		Identity:                     Identity,
		RequestId:                    RequestID,
		WorkflowIdReusePolicy:        WorkflowIDReusePolicy,
		RetryPolicy:                  &RetryPolicy,
		CronSchedule:                 CronSchedule,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
		Header:                       &Header,
		CronOverlapPolicy:            apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_SKIPPED,
		ActiveClusterSelectionPolicy: ActiveClusterSelectionPolicyRegionSticky,
	}
	StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy2 = apiv1.StartWorkflowExecutionRequest{
		Domain:                       DomainName,
		WorkflowId:                   WorkflowID,
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		Identity:                     Identity,
		RequestId:                    RequestID,
		WorkflowIdReusePolicy:        WorkflowIDReusePolicy,
		RetryPolicy:                  &RetryPolicy,
		CronSchedule:                 CronSchedule,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
		Header:                       &Header,
		CronOverlapPolicy:            apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_BUFFER_ONE,
		ActiveClusterSelectionPolicy: ActiveClusterSelectionPolicyExternalEntity,
	}
	StartWorkflowExecutionResponse = apiv1.StartWorkflowExecutionResponse{
		RunId: RunID,
	}
	SignalWorkflowExecutionRequest = apiv1.SignalWorkflowExecutionRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		SignalName:        SignalName,
		SignalInput:       &Payload1,
		Identity:          Identity,
		RequestId:         RequestID,
		Control:           Control,
	}
	SignalWithStartWorkflowExecutionRequest = apiv1.SignalWithStartWorkflowExecutionRequest{
		StartRequest: &StartWorkflowExecutionRequest,
		SignalName:   SignalName,
		SignalInput:  &Payload2,
		Control:      Control,
	}
	SignalWithStartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy1 = apiv1.SignalWithStartWorkflowExecutionRequest{
		StartRequest: &StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy1,
		SignalName:   SignalName,
		SignalInput:  &Payload2,
		Control:      Control,
	}
	SignalWithStartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy2 = apiv1.SignalWithStartWorkflowExecutionRequest{
		StartRequest: &StartWorkflowExecutionRequestWithCronAndActiveClusterSelectionPolicy2,
		SignalName:   SignalName,
		SignalInput:  &Payload2,
		Control:      Control,
	}
	SignalWithStartWorkflowExecutionResponse = apiv1.SignalWithStartWorkflowExecutionResponse{
		RunId: RunID,
	}
	ResetWorkflowExecutionRequest = apiv1.ResetWorkflowExecutionRequest{
		Domain:                DomainName,
		WorkflowExecution:     &WorkflowExecution,
		Reason:                Reason,
		DecisionFinishEventId: EventID1,
		RequestId:             RequestID,
		SkipSignalReapply:     true,
	}
	ResetWorkflowExecutionResponse = apiv1.ResetWorkflowExecutionResponse{
		RunId: RunID,
	}
	TerminateWorkflowExecutionRequest = apiv1.TerminateWorkflowExecutionRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		Reason:            Reason,
		Details:           &Payload1,
		Identity:          Identity,
	}
	DescribeWorkflowExecutionRequest = apiv1.DescribeWorkflowExecutionRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
	}
	DescribeWorkflowExecutionResponse = apiv1.DescribeWorkflowExecutionResponse{
		ExecutionConfiguration: &WorkflowExecutionConfiguration,
		WorkflowExecutionInfo:  &WorkflowExecutionInfo,
		PendingActivities:      PendingActivityInfoArray,
		PendingChildren:        PendingChildExecutionInfoArray,
		PendingDecision:        &PendingDecisionInfo,
	}
	DiagnoseWorkflowExecutionRequest = apiv1.DiagnoseWorkflowExecutionRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		Identity:          Identity,
	}
	DiagnoseWorkflowExecutionResponse = apiv1.DiagnoseWorkflowExecutionResponse{
		Domain:                      DomainName,
		DiagnosticWorkflowExecution: &WorkflowExecution,
	}
	QueryWorkflowRequest = apiv1.QueryWorkflowRequest{
		Domain:                DomainName,
		WorkflowExecution:     &WorkflowExecution,
		Query:                 &WorkflowQuery,
		QueryRejectCondition:  QueryRejectCondition,
		QueryConsistencyLevel: QueryConsistencyLevel,
	}
	QueryWorkflowResponse = apiv1.QueryWorkflowResponse{
		QueryResult:   &Payload1,
		QueryRejected: &QueryRejected,
	}
	DescribeTaskListRequest = apiv1.DescribeTaskListRequest{
		Domain:                DomainName,
		TaskList:              &TaskList,
		TaskListType:          TaskListType,
		IncludeTaskListStatus: true,
	}
	DescribeTaskListResponse = apiv1.DescribeTaskListResponse{
		Pollers:        PollerInfoArray,
		TaskListStatus: &TaskListStatus,
	}
	ListTaskListPartitionsRequest = apiv1.ListTaskListPartitionsRequest{
		Domain:   DomainName,
		TaskList: &TaskList,
	}
	ListTaskListPartitionsResponse = apiv1.ListTaskListPartitionsResponse{
		ActivityTaskListPartitions: TaskListPartitionMetadataArray,
		DecisionTaskListPartitions: TaskListPartitionMetadataArray,
	}
	ResetStickyTaskListRequest = apiv1.ResetStickyTaskListRequest{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
	}
	ResetStickyTaskListResponse        = apiv1.ResetStickyTaskListResponse{}
	GetWorkflowExecutionHistoryRequest = apiv1.GetWorkflowExecutionHistoryRequest{
		Domain:                 DomainName,
		WorkflowExecution:      &WorkflowExecution,
		PageSize:               PageSize,
		NextPageToken:          NextPageToken,
		WaitForNewEvent:        true,
		HistoryEventFilterType: HistoryEventFilterType,
		SkipArchival:           true,
	}
	GetWorkflowExecutionHistoryResponse = apiv1.GetWorkflowExecutionHistoryResponse{
		History:       &History,
		RawHistory:    DataBlobArray,
		NextPageToken: NextPageToken,
		Archived:      true,
	}

	DomainArray = []*apiv1.Domain{
		&Domain,
	}
	GetClusterInfoResponse = apiv1.GetClusterInfoResponse{
		SupportedClientVersions: &SupportedClientVersions,
	}
)
