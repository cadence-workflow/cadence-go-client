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
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/compatibility/proto"
	"go.uber.org/cadence/internal/compatibility/thrift"
)

const UnknownValue = 9999

func TestArchivalStatus(t *testing.T) {
	for _, item := range []apiv1.ArchivalStatus{
		apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID,
		apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED,
		apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED,
	} {
		assert.Equal(t, item, proto.ArchivalStatus(thrift.ArchivalStatus(item)))
	}
	assert.Panics(t, func() { proto.ArchivalStatus(shared.ArchivalStatus(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.ArchivalStatus(apiv1.ArchivalStatus(UnknownValue)) })
}
func TestCancelExternalWorkflowExecutionFailedCause(t *testing.T) {
	for _, item := range []apiv1.CancelExternalWorkflowExecutionFailedCause{
		apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID,
		apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION,
		apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED,
	} {
		assert.Equal(t, item, proto.CancelExternalWorkflowExecutionFailedCause(thrift.CancelExternalWorkflowExecutionFailedCause(item)))
	}
	assert.Panics(t, func() {
		proto.CancelExternalWorkflowExecutionFailedCause(shared.CancelExternalWorkflowExecutionFailedCause(UnknownValue).Ptr())
	})
	assert.Panics(t, func() {
		thrift.CancelExternalWorkflowExecutionFailedCause(apiv1.CancelExternalWorkflowExecutionFailedCause(UnknownValue))
	})
}
func TestChildWorkflowExecutionFailedCause(t *testing.T) {
	for _, item := range []apiv1.ChildWorkflowExecutionFailedCause{
		apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID,
		apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING,
	} {
		assert.Equal(t, item, proto.ChildWorkflowExecutionFailedCause(thrift.ChildWorkflowExecutionFailedCause(item)))
	}
	assert.Panics(t, func() {
		proto.ChildWorkflowExecutionFailedCause(shared.ChildWorkflowExecutionFailedCause(UnknownValue).Ptr())
	})
	assert.Panics(t, func() {
		thrift.ChildWorkflowExecutionFailedCause(apiv1.ChildWorkflowExecutionFailedCause(UnknownValue))
	})
}
func TestContinueAsNewInitiator(t *testing.T) {
	for _, item := range []apiv1.ContinueAsNewInitiator{
		apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID,
		apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER,
		apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY,
		apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE,
	} {
		assert.Equal(t, item, proto.ContinueAsNewInitiator(thrift.ContinueAsNewInitiator(item)))
	}
	assert.Panics(t, func() { proto.ContinueAsNewInitiator(shared.ContinueAsNewInitiator(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.ContinueAsNewInitiator(apiv1.ContinueAsNewInitiator(UnknownValue)) })
}
func TestDecisionTaskFailedCause(t *testing.T) {
	for _, item := range []apiv1.DecisionTaskFailedCause{
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_INVALID,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_RECORD_MARKER_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CONTINUE_AS_NEW_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_START_TIMER_DUPLICATE_ID,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_STICKY_TASK_LIST,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_CHILD_EXECUTION_ATTRIBUTES,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FORCE_CLOSE_DECISION,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FAILOVER_CLOSE_DECISION,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_INPUT_SIZE,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_WORKFLOW,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_BINARY,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_SCHEDULE_ACTIVITY_DUPLICATE_ID,
		apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SEARCH_ATTRIBUTES,
	} {
		assert.Equal(t, item, proto.DecisionTaskFailedCause(thrift.DecisionTaskFailedCause(item)))
	}
	assert.Panics(t, func() { proto.DecisionTaskFailedCause(shared.DecisionTaskFailedCause(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.DecisionTaskFailedCause(apiv1.DecisionTaskFailedCause(UnknownValue)) })
}
func TestDomainStatus(t *testing.T) {
	for _, item := range []apiv1.DomainStatus{
		apiv1.DomainStatus_DOMAIN_STATUS_INVALID,
		apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED,
		apiv1.DomainStatus_DOMAIN_STATUS_DEPRECATED,
		apiv1.DomainStatus_DOMAIN_STATUS_DELETED,
	} {
		assert.Equal(t, item, proto.DomainStatus(thrift.DomainStatus(item)))
	}
	assert.Panics(t, func() { proto.DomainStatus(shared.DomainStatus(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.DomainStatus(apiv1.DomainStatus(UnknownValue)) })
}
func TestEncodingType(t *testing.T) {
	for _, item := range []apiv1.EncodingType{
		apiv1.EncodingType_ENCODING_TYPE_INVALID,
		apiv1.EncodingType_ENCODING_TYPE_THRIFTRW,
		apiv1.EncodingType_ENCODING_TYPE_JSON,
	} {
		assert.Equal(t, item, proto.EncodingType(thrift.EncodingType(item)))
	}
	assert.Panics(t, func() { proto.EncodingType(shared.EncodingType(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.EncodingType(apiv1.EncodingType(UnknownValue)) })
	assert.Panics(t, func() { thrift.EncodingType(apiv1.EncodingType_ENCODING_TYPE_PROTO3) })
}
func TestEventFilterType(t *testing.T) {
	for _, item := range []apiv1.EventFilterType{
		apiv1.EventFilterType_EVENT_FILTER_TYPE_INVALID,
		apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT,
		apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT,
	} {
		assert.Equal(t, item, proto.EventFilterType(thrift.EventFilterType(item)))
	}
	assert.Panics(t, func() { proto.EventFilterType(shared.HistoryEventFilterType(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.EventFilterType(apiv1.EventFilterType(UnknownValue)) })
}
func TestIndexedValueType(t *testing.T) {
	for _, item := range []apiv1.IndexedValueType{
		apiv1.IndexedValueType_INDEXED_VALUE_TYPE_STRING,
		apiv1.IndexedValueType_INDEXED_VALUE_TYPE_KEYWORD,
		apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INT,
		apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DOUBLE,
		apiv1.IndexedValueType_INDEXED_VALUE_TYPE_BOOL,
		apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DATETIME,
	} {
		assert.Equal(t, item, proto.IndexedValueType(thrift.IndexedValueType(item)))
	}
	assert.Panics(t, func() { thrift.IndexedValueType(apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INVALID) })
	assert.Panics(t, func() { proto.IndexedValueType(shared.IndexedValueType(UnknownValue)) })
	assert.Panics(t, func() { thrift.IndexedValueType(apiv1.IndexedValueType(UnknownValue)) })
}
func TestParentClosePolicy(t *testing.T) {
	for _, item := range []apiv1.ParentClosePolicy{
		apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_INVALID,
		apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_ABANDON,
		apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_REQUEST_CANCEL,
		apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE,
	} {
		assert.Equal(t, item, proto.ParentClosePolicy(thrift.ParentClosePolicy(item)))
	}
	assert.Panics(t, func() { proto.ParentClosePolicy(shared.ParentClosePolicy(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.ParentClosePolicy(apiv1.ParentClosePolicy(UnknownValue)) })
}
func TestPendingActivityState(t *testing.T) {
	for _, item := range []apiv1.PendingActivityState{
		apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_INVALID,
		apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_SCHEDULED,
		apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_STARTED,
		apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_CANCEL_REQUESTED,
	} {
		assert.Equal(t, item, proto.PendingActivityState(thrift.PendingActivityState(item)))
	}
	assert.Panics(t, func() { proto.PendingActivityState(shared.PendingActivityState(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.PendingActivityState(apiv1.PendingActivityState(UnknownValue)) })
}
func TestPendingDecisionState(t *testing.T) {
	for _, item := range []apiv1.PendingDecisionState{
		apiv1.PendingDecisionState_PENDING_DECISION_STATE_INVALID,
		apiv1.PendingDecisionState_PENDING_DECISION_STATE_SCHEDULED,
		apiv1.PendingDecisionState_PENDING_DECISION_STATE_STARTED,
	} {
		assert.Equal(t, item, proto.PendingDecisionState(thrift.PendingDecisionState(item)))
	}
	assert.Panics(t, func() { proto.PendingDecisionState(shared.PendingDecisionState(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.PendingDecisionState(apiv1.PendingDecisionState(UnknownValue)) })
}
func TestQueryConsistencyLevel(t *testing.T) {
	for _, item := range []apiv1.QueryConsistencyLevel{
		apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID,
		apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL,
		apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG,
	} {
		assert.Equal(t, item, proto.QueryConsistencyLevel(thrift.QueryConsistencyLevel(item)))
	}
	assert.Panics(t, func() { proto.QueryConsistencyLevel(shared.QueryConsistencyLevel(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.QueryConsistencyLevel(apiv1.QueryConsistencyLevel(UnknownValue)) })
}
func TestQueryRejectCondition(t *testing.T) {
	for _, item := range []apiv1.QueryRejectCondition{
		apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_INVALID,
		apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_OPEN,
		apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY,
	} {
		assert.Equal(t, item, proto.QueryRejectCondition(thrift.QueryRejectCondition(item)))
	}
	assert.Panics(t, func() { proto.QueryRejectCondition(shared.QueryRejectCondition(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.QueryRejectCondition(apiv1.QueryRejectCondition(UnknownValue)) })
}
func TestQueryResultType(t *testing.T) {
	for _, item := range []apiv1.QueryResultType{
		apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID,
		apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
		apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
	} {
		assert.Equal(t, item, proto.QueryResultType(thrift.QueryResultType(item)))
	}
	assert.Panics(t, func() { proto.QueryResultType(shared.QueryResultType(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.QueryResultType(apiv1.QueryResultType(UnknownValue)) })
}
func TestQueryTaskCompletedType(t *testing.T) {
	for _, item := range []apiv1.QueryResultType{
		apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID,
		apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
		apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
	} {
		assert.Equal(t, item, proto.QueryTaskCompletedType(thrift.QueryTaskCompletedType(item)))
	}
	assert.Panics(t, func() { thrift.QueryTaskCompletedType(apiv1.QueryResultType(UnknownValue)) })
	assert.Panics(t, func() { proto.QueryTaskCompletedType(shared.QueryTaskCompletedType(UnknownValue).Ptr()) })
}
func TestSignalExternalWorkflowExecutionFailedCause(t *testing.T) {
	for _, item := range []apiv1.SignalExternalWorkflowExecutionFailedCause{
		apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID,
		apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION,
		apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_COMPLETED,
	} {
		assert.Equal(t, item, proto.SignalExternalWorkflowExecutionFailedCause(thrift.SignalExternalWorkflowExecutionFailedCause(item)))
	}
	assert.Panics(t, func() {
		proto.SignalExternalWorkflowExecutionFailedCause(shared.SignalExternalWorkflowExecutionFailedCause(UnknownValue).Ptr())
	})
	assert.Panics(t, func() {
		thrift.SignalExternalWorkflowExecutionFailedCause(apiv1.SignalExternalWorkflowExecutionFailedCause(UnknownValue))
	})
}
func TestTaskListKind(t *testing.T) {
	for _, item := range []apiv1.TaskListKind{
		apiv1.TaskListKind_TASK_LIST_KIND_INVALID,
		apiv1.TaskListKind_TASK_LIST_KIND_NORMAL,
		apiv1.TaskListKind_TASK_LIST_KIND_STICKY,
	} {
		assert.Equal(t, item, proto.TaskListKind(thrift.TaskListKind(item)))
	}
	assert.Panics(t, func() { proto.TaskListKind(shared.TaskListKind(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.TaskListKind(apiv1.TaskListKind(UnknownValue)) })
}
func TestTaskListType(t *testing.T) {
	for _, item := range []apiv1.TaskListType{
		apiv1.TaskListType_TASK_LIST_TYPE_INVALID,
		apiv1.TaskListType_TASK_LIST_TYPE_DECISION,
		apiv1.TaskListType_TASK_LIST_TYPE_ACTIVITY,
	} {
		assert.Equal(t, item, proto.TaskListType(thrift.TaskListType(item)))
	}
	assert.Panics(t, func() { proto.TaskListType(shared.TaskListType(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.TaskListType(apiv1.TaskListType(UnknownValue)) })
}
func TestTimeoutType(t *testing.T) {
	for _, item := range []apiv1.TimeoutType{
		apiv1.TimeoutType_TIMEOUT_TYPE_INVALID,
		apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE,
		apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START,
		apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE,
		apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT,
	} {
		assert.Equal(t, item, proto.TimeoutType(thrift.TimeoutType(item)))
	}
	assert.Panics(t, func() { proto.TimeoutType(shared.TimeoutType(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.TimeoutType(apiv1.TimeoutType(UnknownValue)) })
}
func TestDecisionTaskTimedOutCause(t *testing.T) {
	for _, item := range []apiv1.DecisionTaskTimedOutCause{
		apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_INVALID,
		apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_TIMEOUT,
		apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET,
	} {
		assert.Equal(t, item, proto.DecisionTaskTimedOutCause(thrift.DecisionTaskTimedOutCause(item)))
	}
	assert.Panics(t, func() { proto.DecisionTaskTimedOutCause(shared.DecisionTaskTimedOutCause(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.DecisionTaskTimedOutCause(apiv1.DecisionTaskTimedOutCause(UnknownValue)) })
}
func TestWorkflowExecutionCloseStatus(t *testing.T) {
	for _, item := range []apiv1.WorkflowExecutionCloseStatus{
		apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID,
		apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED,
		apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED,
		apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED,
		apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED,
		apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW,
		apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT,
	} {
		assert.Equal(t, item, proto.WorkflowExecutionCloseStatus(thrift.WorkflowExecutionCloseStatus(item)))
	}
	assert.Panics(t, func() { proto.WorkflowExecutionCloseStatus(shared.WorkflowExecutionCloseStatus(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.WorkflowExecutionCloseStatus(apiv1.WorkflowExecutionCloseStatus(UnknownValue)) })
}
func TestWorkflowIDReusePolicy(t *testing.T) {
	for _, item := range []apiv1.WorkflowIdReusePolicy{
		apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_INVALID,
		apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
		apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	} {
		assert.Equal(t, item, proto.WorkflowIDReusePolicy(thrift.WorkflowIDReusePolicy(item)))
	}
	assert.Panics(t, func() { proto.WorkflowIDReusePolicy(shared.WorkflowIdReusePolicy(UnknownValue).Ptr()) })
	assert.Panics(t, func() { thrift.WorkflowIDReusePolicy(apiv1.WorkflowIdReusePolicy(UnknownValue)) })
}
