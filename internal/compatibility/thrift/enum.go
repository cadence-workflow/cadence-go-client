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

func TaskListKind(t apiv1.TaskListKind) *shared.TaskListKind {
	switch t {
	case apiv1.TaskListKind_TASK_LIST_KIND_INVALID:
		return nil
	case apiv1.TaskListKind_TASK_LIST_KIND_NORMAL:
		return shared.TaskListKindNormal.Ptr()
	case apiv1.TaskListKind_TASK_LIST_KIND_STICKY:
		return shared.TaskListKindSticky.Ptr()
	}
	panic("unexpected enum value")
}

func TaskListType(t apiv1.TaskListType) *shared.TaskListType {
	if t == apiv1.TaskListType_TASK_LIST_TYPE_INVALID {
		return nil
	}
	switch t {
	case apiv1.TaskListType_TASK_LIST_TYPE_DECISION:
		return shared.TaskListTypeDecision.Ptr()
	case apiv1.TaskListType_TASK_LIST_TYPE_ACTIVITY:
		return shared.TaskListTypeActivity.Ptr()
	}
	panic("unexpected enum value")
}

func EventFilterType(t apiv1.EventFilterType) *shared.HistoryEventFilterType {
	if t == apiv1.EventFilterType_EVENT_FILTER_TYPE_INVALID {
		return nil
	}
	switch t {
	case apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT:
		return shared.HistoryEventFilterTypeAllEvent.Ptr()
	case apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT:
		return shared.HistoryEventFilterTypeCloseEvent.Ptr()
	}
	panic("unexpected enum value")
}

func QueryRejectCondition(t apiv1.QueryRejectCondition) *shared.QueryRejectCondition {
	if t == apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_INVALID {
		return nil
	}
	switch t {
	case apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_OPEN:
		return shared.QueryRejectConditionNotOpen.Ptr()
	case apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY:
		return shared.QueryRejectConditionNotCompletedCleanly.Ptr()
	}
	panic("unexpected enum value")
}

func QueryConsistencyLevel(t apiv1.QueryConsistencyLevel) *shared.QueryConsistencyLevel {
	if t == apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID {
		return nil
	}
	switch t {
	case apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL:
		return shared.QueryConsistencyLevelEventual.Ptr()
	case apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG:
		return shared.QueryConsistencyLevelStrong.Ptr()
	}
	panic("unexpected enum value")
}

func ContinueAsNewInitiator(t apiv1.ContinueAsNewInitiator) *shared.ContinueAsNewInitiator {
	switch t {
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID:
		return nil
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER:
		return shared.ContinueAsNewInitiatorDecider.Ptr()
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY:
		return shared.ContinueAsNewInitiatorRetryPolicy.Ptr()
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE:
		return shared.ContinueAsNewInitiatorCronSchedule.Ptr()
	}
	panic("unexpected enum value")
}

func WorkflowIDReusePolicy(t apiv1.WorkflowIdReusePolicy) *shared.WorkflowIdReusePolicy {
	switch t {
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_INVALID:
		return nil
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY:
		return shared.WorkflowIdReusePolicyAllowDuplicateFailedOnly.Ptr()
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE:
		return shared.WorkflowIdReusePolicyAllowDuplicate.Ptr()
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE:
		return shared.WorkflowIdReusePolicyRejectDuplicate.Ptr()
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING:
		return shared.WorkflowIdReusePolicyTerminateIfRunning.Ptr()
	}
	panic("unexpected enum value")
}

func QueryResultType(t apiv1.QueryResultType) *shared.QueryResultType {
	if t == apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID {
		return nil
	}
	switch t {
	case apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED:
		return shared.QueryResultTypeAnswered.Ptr()
	case apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED:
		return shared.QueryResultTypeFailed.Ptr()
	}
	panic("unexpected enum value")
}

func ArchivalStatus(t apiv1.ArchivalStatus) *shared.ArchivalStatus {
	switch t {
	case apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID:
		return nil
	case apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED:
		return shared.ArchivalStatusDisabled.Ptr()
	case apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED:
		return shared.ArchivalStatusEnabled.Ptr()
	}
	panic("unexpected enum value")
}

func ParentClosePolicy(t apiv1.ParentClosePolicy) *shared.ParentClosePolicy {
	switch t {
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_INVALID:
		return nil
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_ABANDON:
		return shared.ParentClosePolicyAbandon.Ptr()
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_REQUEST_CANCEL:
		return shared.ParentClosePolicyRequestCancel.Ptr()
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE:
		return shared.ParentClosePolicyTerminate.Ptr()
	}
	panic("unexpected enum value")
}

func DecisionTaskFailedCause(t apiv1.DecisionTaskFailedCause) *shared.DecisionTaskFailedCause {
	switch t {
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION:
		return shared.DecisionTaskFailedCauseUnhandledDecision.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadScheduleActivityAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadRequestCancelActivityAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadStartTimerAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadCancelTimerAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_RECORD_MARKER_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadRecordMarkerAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CONTINUE_AS_NEW_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadContinueAsNewAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_START_TIMER_DUPLICATE_ID:
		return shared.DecisionTaskFailedCauseStartTimerDuplicateID.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_STICKY_TASK_LIST:
		return shared.DecisionTaskFailedCauseResetStickyTasklist.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE:
		return shared.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_CHILD_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadStartChildExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FORCE_CLOSE_DECISION:
		return shared.DecisionTaskFailedCauseForceCloseDecision.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FAILOVER_CLOSE_DECISION:
		return shared.DecisionTaskFailedCauseFailoverCloseDecision.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_INPUT_SIZE:
		return shared.DecisionTaskFailedCauseBadSignalInputSize.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_WORKFLOW:
		return shared.DecisionTaskFailedCauseResetWorkflow.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_BINARY:
		return shared.DecisionTaskFailedCauseBadBinary.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_SCHEDULE_ACTIVITY_DUPLICATE_ID:
		return shared.DecisionTaskFailedCauseScheduleActivityDuplicateID.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SEARCH_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadSearchAttributes.Ptr()
	}
	panic("unexpected enum value")
}

func WorkflowExecutionCloseStatus(t apiv1.WorkflowExecutionCloseStatus) *shared.WorkflowExecutionCloseStatus {
	switch t {
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID:
		return nil
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED:
		return shared.WorkflowExecutionCloseStatusCompleted.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED:
		return shared.WorkflowExecutionCloseStatusFailed.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED:
		return shared.WorkflowExecutionCloseStatusCanceled.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED:
		return shared.WorkflowExecutionCloseStatusTerminated.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW:
		return shared.WorkflowExecutionCloseStatusContinuedAsNew.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT:
		return shared.WorkflowExecutionCloseStatusTimedOut.Ptr()
	}
	panic("unexpected enum value")
}

func QueryTaskCompletedType(t apiv1.QueryResultType) *shared.QueryTaskCompletedType {
	if t == apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID {
		return nil
	}
	switch t {
	case apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED:
		return shared.QueryTaskCompletedTypeCompleted.Ptr()
	case apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED:
		return shared.QueryTaskCompletedTypeFailed.Ptr()
	}
	panic("unexpected enum value")
}

func DomainStatus(t apiv1.DomainStatus) *shared.DomainStatus {
	switch t {
	case apiv1.DomainStatus_DOMAIN_STATUS_INVALID:
		return nil
	case apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED:
		return shared.DomainStatusRegistered.Ptr()
	case apiv1.DomainStatus_DOMAIN_STATUS_DEPRECATED:
		return shared.DomainStatusDeprecated.Ptr()
	case apiv1.DomainStatus_DOMAIN_STATUS_DELETED:
		return shared.DomainStatusDeleted.Ptr()
	}
	panic("unexpected enum value")
}

func PendingActivityState(t apiv1.PendingActivityState) *shared.PendingActivityState {
	switch t {
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_INVALID:
		return nil
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_SCHEDULED:
		return shared.PendingActivityStateScheduled.Ptr()
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_STARTED:
		return shared.PendingActivityStateStarted.Ptr()
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_CANCEL_REQUESTED:
		return shared.PendingActivityStateCancelRequested.Ptr()
	}
	panic("unexpected enum value")
}

func PendingDecisionState(t apiv1.PendingDecisionState) *shared.PendingDecisionState {
	switch t {
	case apiv1.PendingDecisionState_PENDING_DECISION_STATE_INVALID:
		return nil
	case apiv1.PendingDecisionState_PENDING_DECISION_STATE_SCHEDULED:
		return shared.PendingDecisionStateScheduled.Ptr()
	case apiv1.PendingDecisionState_PENDING_DECISION_STATE_STARTED:
		return shared.PendingDecisionStateStarted.Ptr()
	}
	panic("unexpected enum value")
}

func IndexedValueType(t apiv1.IndexedValueType) shared.IndexedValueType {
	switch t {
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INVALID:
		panic("received IndexedValueType_INDEXED_VALUE_TYPE_INVALID")
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_STRING:
		return shared.IndexedValueTypeString
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_KEYWORD:
		return shared.IndexedValueTypeKeyword
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INT:
		return shared.IndexedValueTypeInt
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DOUBLE:
		return shared.IndexedValueTypeDouble
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_BOOL:
		return shared.IndexedValueTypeBool
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DATETIME:
		return shared.IndexedValueTypeDatetime
	}
	panic("unexpected enum value")
}

func EncodingType(t apiv1.EncodingType) *shared.EncodingType {
	switch t {
	case apiv1.EncodingType_ENCODING_TYPE_INVALID:
		return nil
	case apiv1.EncodingType_ENCODING_TYPE_THRIFTRW:
		return shared.EncodingTypeThriftRW.Ptr()
	case apiv1.EncodingType_ENCODING_TYPE_JSON:
		return shared.EncodingTypeJSON.Ptr()
	case apiv1.EncodingType_ENCODING_TYPE_PROTO3:
		panic("not supported yet")
	}
	panic("unexpected enum value")
}

func TimeoutType(t apiv1.TimeoutType) *shared.TimeoutType {
	switch t {
	case apiv1.TimeoutType_TIMEOUT_TYPE_INVALID:
		return nil
	case apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE:
		return shared.TimeoutTypeStartToClose.Ptr()
	case apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START:
		return shared.TimeoutTypeScheduleToStart.Ptr()
	case apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE:
		return shared.TimeoutTypeScheduleToClose.Ptr()
	case apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT:
		return shared.TimeoutTypeHeartbeat.Ptr()
	}
	panic("unexpected enum value")
}

func DecisionTaskTimedOutCause(t apiv1.DecisionTaskTimedOutCause) *shared.DecisionTaskTimedOutCause {
	switch t {
	case apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_INVALID:
		return nil
	case apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_TIMEOUT:
		return shared.DecisionTaskTimedOutCauseTimeout.Ptr()
	case apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET:
		return shared.DecisionTaskTimedOutCauseReset.Ptr()
	}
	panic("unexpected enum value")
}

func CancelExternalWorkflowExecutionFailedCause(t apiv1.CancelExternalWorkflowExecutionFailedCause) *shared.CancelExternalWorkflowExecutionFailedCause {
	switch t {
	case apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION:
		return shared.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution.Ptr()
	case apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED:
		return shared.CancelExternalWorkflowExecutionFailedCauseWorkflowAlreadyCompleted.Ptr()
	}
	panic("unexpected enum value")
}

func SignalExternalWorkflowExecutionFailedCause(t apiv1.SignalExternalWorkflowExecutionFailedCause) *shared.SignalExternalWorkflowExecutionFailedCause {
	switch t {
	case apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION:
		return shared.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution.Ptr()
	case apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED:
		return shared.SignalExternalWorkflowExecutionFailedCauseWorkflowAlreadyCompleted.Ptr()
	}
	panic("unexpected enum value")
}

func ChildWorkflowExecutionFailedCause(t apiv1.ChildWorkflowExecutionFailedCause) *shared.ChildWorkflowExecutionFailedCause {
	switch t {
	case apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING:
		return shared.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning.Ptr()
	}
	panic("unexpected enum value")
}
