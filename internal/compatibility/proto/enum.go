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

func TaskListKind(t *shared.TaskListKind) apiv1.TaskListKind {
	if t == nil {
		return apiv1.TaskListKind_TASK_LIST_KIND_INVALID
	}
	switch *t {
	case shared.TaskListKindNormal:
		return apiv1.TaskListKind_TASK_LIST_KIND_NORMAL
	case shared.TaskListKindSticky:
		return apiv1.TaskListKind_TASK_LIST_KIND_STICKY
	}
	panic("unexpected enum value")
}

func TaskListType(t *shared.TaskListType) apiv1.TaskListType {
	if t == nil {
		return apiv1.TaskListType_TASK_LIST_TYPE_INVALID
	}
	switch *t {
	case shared.TaskListTypeDecision:
		return apiv1.TaskListType_TASK_LIST_TYPE_DECISION
	case shared.TaskListTypeActivity:
		return apiv1.TaskListType_TASK_LIST_TYPE_ACTIVITY
	}
	panic("unexpected enum value")
}

func EventFilterType(t *shared.HistoryEventFilterType) apiv1.EventFilterType {
	if t == nil {
		return apiv1.EventFilterType_EVENT_FILTER_TYPE_INVALID
	}
	switch *t {
	case shared.HistoryEventFilterTypeAllEvent:
		return apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT
	case shared.HistoryEventFilterTypeCloseEvent:
		return apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	}
	panic("unexpected enum value")
}

func QueryRejectCondition(t *shared.QueryRejectCondition) apiv1.QueryRejectCondition {
	if t == nil {
		return apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_INVALID
	}
	switch *t {
	case shared.QueryRejectConditionNotOpen:
		return apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_OPEN
	case shared.QueryRejectConditionNotCompletedCleanly:
		return apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
	}
	panic("unexpected enum value")
}

func QueryConsistencyLevel(t *shared.QueryConsistencyLevel) apiv1.QueryConsistencyLevel {
	if t == nil {
		return apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID
	}
	switch *t {
	case shared.QueryConsistencyLevelEventual:
		return apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL
	case shared.QueryConsistencyLevelStrong:
		return apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG
	}
	panic("unexpected enum value")
}

func ContinueAsNewInitiator(t *shared.ContinueAsNewInitiator) apiv1.ContinueAsNewInitiator {
	if t == nil {
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID
	}
	switch *t {
	case shared.ContinueAsNewInitiatorDecider:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER
	case shared.ContinueAsNewInitiatorRetryPolicy:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY
	case shared.ContinueAsNewInitiatorCronSchedule:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE
	}
	panic("unexpected enum value")
}

func WorkflowIDReusePolicy(t *shared.WorkflowIdReusePolicy) apiv1.WorkflowIdReusePolicy {
	if t == nil {
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_INVALID
	}
	switch *t {
	case shared.WorkflowIdReusePolicyAllowDuplicateFailedOnly:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY
	case shared.WorkflowIdReusePolicyAllowDuplicate:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	case shared.WorkflowIdReusePolicyRejectDuplicate:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE
	case shared.WorkflowIdReusePolicyTerminateIfRunning:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
	}
	panic("unexpected enum value")
}

func QueryResultType(t *shared.QueryResultType) apiv1.QueryResultType {
	if t == nil {
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID
	}
	switch *t {
	case shared.QueryResultTypeAnswered:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED
	case shared.QueryResultTypeFailed:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED
	}
	panic("unexpected enum value")
}

func ArchivalStatus(t *shared.ArchivalStatus) apiv1.ArchivalStatus {
	if t == nil {
		return apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID
	}
	switch *t {
	case shared.ArchivalStatusDisabled:
		return apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED
	case shared.ArchivalStatusEnabled:
		return apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED
	}
	panic("unexpected enum value")
}

func ParentClosePolicy(t *shared.ParentClosePolicy) apiv1.ParentClosePolicy {
	if t == nil {
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_INVALID
	}
	switch *t {
	case shared.ParentClosePolicyAbandon:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_ABANDON
	case shared.ParentClosePolicyRequestCancel:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_REQUEST_CANCEL
	case shared.ParentClosePolicyTerminate:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE
	}
	panic("unexpected enum value")
}

func DecisionTaskFailedCause(t *shared.DecisionTaskFailedCause) apiv1.DecisionTaskFailedCause {
	if t == nil {
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.DecisionTaskFailedCauseUnhandledDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION
	case shared.DecisionTaskFailedCauseBadScheduleActivityAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRequestCancelActivityAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadStartTimerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCancelTimerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRecordMarkerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_RECORD_MARKER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadContinueAsNewAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CONTINUE_AS_NEW_ATTRIBUTES
	case shared.DecisionTaskFailedCauseStartTimerDuplicateID:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_START_TIMER_DUPLICATE_ID
	case shared.DecisionTaskFailedCauseResetStickyTasklist:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_STICKY_TASK_LIST
	case shared.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE
	case shared.DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadStartChildExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_CHILD_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseForceCloseDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FORCE_CLOSE_DECISION
	case shared.DecisionTaskFailedCauseFailoverCloseDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FAILOVER_CLOSE_DECISION
	case shared.DecisionTaskFailedCauseBadSignalInputSize:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_INPUT_SIZE
	case shared.DecisionTaskFailedCauseResetWorkflow:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_WORKFLOW
	case shared.DecisionTaskFailedCauseBadBinary:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_BINARY
	case shared.DecisionTaskFailedCauseScheduleActivityDuplicateID:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_SCHEDULE_ACTIVITY_DUPLICATE_ID
	case shared.DecisionTaskFailedCauseBadSearchAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SEARCH_ATTRIBUTES
	}
	panic("unexpected enum value")
}

func WorkflowExecutionCloseStatus(t *shared.WorkflowExecutionCloseStatus) apiv1.WorkflowExecutionCloseStatus {
	if t == nil {
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID
	}
	switch *t {
	case shared.WorkflowExecutionCloseStatusCompleted:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED
	case shared.WorkflowExecutionCloseStatusFailed:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED
	case shared.WorkflowExecutionCloseStatusCanceled:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED
	case shared.WorkflowExecutionCloseStatusTerminated:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED
	case shared.WorkflowExecutionCloseStatusContinuedAsNew:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW
	case shared.WorkflowExecutionCloseStatusTimedOut:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT
	}
	panic("unexpected enum value")
}

func QueryTaskCompletedType(t *shared.QueryTaskCompletedType) apiv1.QueryResultType {
	if t == nil {
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID
	}
	switch *t {
	case shared.QueryTaskCompletedTypeCompleted:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED
	case shared.QueryTaskCompletedTypeFailed:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED
	}
	panic("unexpected enum value")
}

func DomainStatus(t *shared.DomainStatus) apiv1.DomainStatus {
	if t == nil {
		return apiv1.DomainStatus_DOMAIN_STATUS_INVALID
	}
	switch *t {
	case shared.DomainStatusRegistered:
		return apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED
	case shared.DomainStatusDeprecated:
		return apiv1.DomainStatus_DOMAIN_STATUS_DEPRECATED
	case shared.DomainStatusDeleted:
		return apiv1.DomainStatus_DOMAIN_STATUS_DELETED
	}
	panic("unexpected enum value")
}

func PendingActivityState(t *shared.PendingActivityState) apiv1.PendingActivityState {
	if t == nil {
		return apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_INVALID
	}
	switch *t {
	case shared.PendingActivityStateScheduled:
		return apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_SCHEDULED
	case shared.PendingActivityStateStarted:
		return apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_STARTED
	case shared.PendingActivityStateCancelRequested:
		return apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_CANCEL_REQUESTED
	}
	panic("unexpected enum value")
}

func PendingDecisionState(t *shared.PendingDecisionState) apiv1.PendingDecisionState {
	if t == nil {
		return apiv1.PendingDecisionState_PENDING_DECISION_STATE_INVALID
	}
	switch *t {
	case shared.PendingDecisionStateScheduled:
		return apiv1.PendingDecisionState_PENDING_DECISION_STATE_SCHEDULED
	case shared.PendingDecisionStateStarted:
		return apiv1.PendingDecisionState_PENDING_DECISION_STATE_STARTED
	}
	panic("unexpected enum value")
}

func IndexedValueType(t shared.IndexedValueType) apiv1.IndexedValueType {
	switch t {
	case shared.IndexedValueTypeString:
		return apiv1.IndexedValueType_INDEXED_VALUE_TYPE_STRING
	case shared.IndexedValueTypeKeyword:
		return apiv1.IndexedValueType_INDEXED_VALUE_TYPE_KEYWORD
	case shared.IndexedValueTypeInt:
		return apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INT
	case shared.IndexedValueTypeDouble:
		return apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DOUBLE
	case shared.IndexedValueTypeBool:
		return apiv1.IndexedValueType_INDEXED_VALUE_TYPE_BOOL
	case shared.IndexedValueTypeDatetime:
		return apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DATETIME
	}
	panic("unexpected enum value")
}

func EncodingType(t *shared.EncodingType) apiv1.EncodingType {
	if t == nil {
		return apiv1.EncodingType_ENCODING_TYPE_INVALID
	}
	switch *t {
	case shared.EncodingTypeThriftRW:
		return apiv1.EncodingType_ENCODING_TYPE_THRIFTRW
	case shared.EncodingTypeJSON:
		return apiv1.EncodingType_ENCODING_TYPE_JSON
	}
	panic("unexpected enum value")
}

func TimeoutType(t *shared.TimeoutType) apiv1.TimeoutType {
	if t == nil {
		return apiv1.TimeoutType_TIMEOUT_TYPE_INVALID
	}
	switch *t {
	case shared.TimeoutTypeStartToClose:
		return apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE
	case shared.TimeoutTypeScheduleToStart:
		return apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START
	case shared.TimeoutTypeScheduleToClose:
		return apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE
	case shared.TimeoutTypeHeartbeat:
		return apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT
	}
	panic("unexpected enum value")
}

func DecisionTaskTimedOutCause(t *shared.DecisionTaskTimedOutCause) apiv1.DecisionTaskTimedOutCause {
	if t == nil {
		return apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_INVALID
	}
	switch *t {
	case shared.DecisionTaskTimedOutCauseTimeout:
		return apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_TIMEOUT
	case shared.DecisionTaskTimedOutCauseReset:
		return apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET
	}
	panic("unexpected enum value")
}

func CancelExternalWorkflowExecutionFailedCause(t *shared.CancelExternalWorkflowExecutionFailedCause) apiv1.CancelExternalWorkflowExecutionFailedCause {
	if t == nil {
		return apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution:
		return apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION
	case shared.CancelExternalWorkflowExecutionFailedCauseWorkflowAlreadyCompleted:
		return apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED
	}
	panic("unexpected enum value")
}

func SignalExternalWorkflowExecutionFailedCause(t *shared.SignalExternalWorkflowExecutionFailedCause) apiv1.SignalExternalWorkflowExecutionFailedCause {
	if t == nil {
		return apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution:
		return apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION
	case shared.SignalExternalWorkflowExecutionFailedCauseWorkflowAlreadyCompleted:
		return apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED
	}
	panic("unexpected enum value")
}

func ChildWorkflowExecutionFailedCause(t *shared.ChildWorkflowExecutionFailedCause) apiv1.ChildWorkflowExecutionFailedCause {
	if t == nil {
		return apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning:
		return apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING
	}
	panic("unexpected enum value")
}
