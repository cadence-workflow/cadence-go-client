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

func History(t *shared.History) *apiv1.History {
	if t == nil {
		return nil
	}
	return &apiv1.History{
		Events: HistoryEventArray(t.Events),
	}
}

func HistoryEventArray(t []*shared.HistoryEvent) []*apiv1.HistoryEvent {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.HistoryEvent, len(t))
	for i := range t {
		v[i] = HistoryEvent(t[i])
	}
	return v
}

func HistoryEvent(e *shared.HistoryEvent) *apiv1.HistoryEvent {
	if e == nil {
		return nil
	}
	event := apiv1.HistoryEvent{
		EventId:   e.GetEventId(),
		EventTime: unixNanoToTime(e.Timestamp),
		Version:   e.GetVersion(),
		TaskId:    e.GetTaskId(),
	}
	switch *e.EventType {
	case shared.EventTypeWorkflowExecutionStarted:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{WorkflowExecutionStartedEventAttributes: WorkflowExecutionStartedEventAttributes(e.WorkflowExecutionStartedEventAttributes)}
	case shared.EventTypeWorkflowExecutionCompleted:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{WorkflowExecutionCompletedEventAttributes: WorkflowExecutionCompletedEventAttributes(e.WorkflowExecutionCompletedEventAttributes)}
	case shared.EventTypeWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionFailedEventAttributes{WorkflowExecutionFailedEventAttributes: WorkflowExecutionFailedEventAttributes(e.WorkflowExecutionFailedEventAttributes)}
	case shared.EventTypeWorkflowExecutionTimedOut:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionTimedOutEventAttributes{WorkflowExecutionTimedOutEventAttributes: WorkflowExecutionTimedOutEventAttributes(e.WorkflowExecutionTimedOutEventAttributes)}
	case shared.EventTypeDecisionTaskScheduled:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes{DecisionTaskScheduledEventAttributes: DecisionTaskScheduledEventAttributes(e.DecisionTaskScheduledEventAttributes)}
	case shared.EventTypeDecisionTaskStarted:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskStartedEventAttributes{DecisionTaskStartedEventAttributes: DecisionTaskStartedEventAttributes(e.DecisionTaskStartedEventAttributes)}
	case shared.EventTypeDecisionTaskCompleted:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes{DecisionTaskCompletedEventAttributes: DecisionTaskCompletedEventAttributes(e.DecisionTaskCompletedEventAttributes)}
	case shared.EventTypeDecisionTaskTimedOut:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskTimedOutEventAttributes{DecisionTaskTimedOutEventAttributes: DecisionTaskTimedOutEventAttributes(e.DecisionTaskTimedOutEventAttributes)}
	case shared.EventTypeDecisionTaskFailed:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskFailedEventAttributes{DecisionTaskFailedEventAttributes: DecisionTaskFailedEventAttributes(e.DecisionTaskFailedEventAttributes)}
	case shared.EventTypeActivityTaskScheduled:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: ActivityTaskScheduledEventAttributes(e.ActivityTaskScheduledEventAttributes)}
	case shared.EventTypeActivityTaskStarted:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskStartedEventAttributes{ActivityTaskStartedEventAttributes: ActivityTaskStartedEventAttributes(e.ActivityTaskStartedEventAttributes)}
	case shared.EventTypeActivityTaskCompleted:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskCompletedEventAttributes{ActivityTaskCompletedEventAttributes: ActivityTaskCompletedEventAttributes(e.ActivityTaskCompletedEventAttributes)}
	case shared.EventTypeActivityTaskFailed:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskFailedEventAttributes{ActivityTaskFailedEventAttributes: ActivityTaskFailedEventAttributes(e.ActivityTaskFailedEventAttributes)}
	case shared.EventTypeActivityTaskTimedOut:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskTimedOutEventAttributes{ActivityTaskTimedOutEventAttributes: ActivityTaskTimedOutEventAttributes(e.ActivityTaskTimedOutEventAttributes)}
	case shared.EventTypeTimerStarted:
		event.Attributes = &apiv1.HistoryEvent_TimerStartedEventAttributes{TimerStartedEventAttributes: TimerStartedEventAttributes(e.TimerStartedEventAttributes)}
	case shared.EventTypeTimerFired:
		event.Attributes = &apiv1.HistoryEvent_TimerFiredEventAttributes{TimerFiredEventAttributes: TimerFiredEventAttributes(e.TimerFiredEventAttributes)}
	case shared.EventTypeActivityTaskCancelRequested:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskCancelRequestedEventAttributes{ActivityTaskCancelRequestedEventAttributes: ActivityTaskCancelRequestedEventAttributes(e.ActivityTaskCancelRequestedEventAttributes)}
	case shared.EventTypeRequestCancelActivityTaskFailed:
		event.Attributes = &apiv1.HistoryEvent_RequestCancelActivityTaskFailedEventAttributes{RequestCancelActivityTaskFailedEventAttributes: RequestCancelActivityTaskFailedEventAttributes(e.RequestCancelActivityTaskFailedEventAttributes)}
	case shared.EventTypeActivityTaskCanceled:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskCanceledEventAttributes{ActivityTaskCanceledEventAttributes: ActivityTaskCanceledEventAttributes(e.ActivityTaskCanceledEventAttributes)}
	case shared.EventTypeTimerCanceled:
		event.Attributes = &apiv1.HistoryEvent_TimerCanceledEventAttributes{TimerCanceledEventAttributes: TimerCanceledEventAttributes(e.TimerCanceledEventAttributes)}
	case shared.EventTypeCancelTimerFailed:
		event.Attributes = &apiv1.HistoryEvent_CancelTimerFailedEventAttributes{CancelTimerFailedEventAttributes: CancelTimerFailedEventAttributes(e.CancelTimerFailedEventAttributes)}
	case shared.EventTypeMarkerRecorded:
		event.Attributes = &apiv1.HistoryEvent_MarkerRecordedEventAttributes{MarkerRecordedEventAttributes: MarkerRecordedEventAttributes(e.MarkerRecordedEventAttributes)}
	case shared.EventTypeWorkflowExecutionSignaled:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes{WorkflowExecutionSignaledEventAttributes: WorkflowExecutionSignaledEventAttributes(e.WorkflowExecutionSignaledEventAttributes)}
	case shared.EventTypeWorkflowExecutionTerminated:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionTerminatedEventAttributes{WorkflowExecutionTerminatedEventAttributes: WorkflowExecutionTerminatedEventAttributes(e.WorkflowExecutionTerminatedEventAttributes)}
	case shared.EventTypeWorkflowExecutionCancelRequested:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes{WorkflowExecutionCancelRequestedEventAttributes: WorkflowExecutionCancelRequestedEventAttributes(e.WorkflowExecutionCancelRequestedEventAttributes)}
	case shared.EventTypeWorkflowExecutionCanceled:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCanceledEventAttributes{WorkflowExecutionCanceledEventAttributes: WorkflowExecutionCanceledEventAttributes(e.WorkflowExecutionCanceledEventAttributes)}
	case shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated:
		event.Attributes = &apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{RequestCancelExternalWorkflowExecutionInitiatedEventAttributes: RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(e.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes)}
	case shared.EventTypeRequestCancelExternalWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionFailedEventAttributes{RequestCancelExternalWorkflowExecutionFailedEventAttributes: RequestCancelExternalWorkflowExecutionFailedEventAttributes(e.RequestCancelExternalWorkflowExecutionFailedEventAttributes)}
	case shared.EventTypeExternalWorkflowExecutionCancelRequested:
		event.Attributes = &apiv1.HistoryEvent_ExternalWorkflowExecutionCancelRequestedEventAttributes{ExternalWorkflowExecutionCancelRequestedEventAttributes: ExternalWorkflowExecutionCancelRequestedEventAttributes(e.ExternalWorkflowExecutionCancelRequestedEventAttributes)}
	case shared.EventTypeWorkflowExecutionContinuedAsNew:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes{WorkflowExecutionContinuedAsNewEventAttributes: WorkflowExecutionContinuedAsNewEventAttributes(e.WorkflowExecutionContinuedAsNewEventAttributes)}
	case shared.EventTypeStartChildWorkflowExecutionInitiated:
		event.Attributes = &apiv1.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes{StartChildWorkflowExecutionInitiatedEventAttributes: StartChildWorkflowExecutionInitiatedEventAttributes(e.StartChildWorkflowExecutionInitiatedEventAttributes)}
	case shared.EventTypeStartChildWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_StartChildWorkflowExecutionFailedEventAttributes{StartChildWorkflowExecutionFailedEventAttributes: StartChildWorkflowExecutionFailedEventAttributes(e.StartChildWorkflowExecutionFailedEventAttributes)}
	case shared.EventTypeChildWorkflowExecutionStarted:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionStartedEventAttributes{ChildWorkflowExecutionStartedEventAttributes: ChildWorkflowExecutionStartedEventAttributes(e.ChildWorkflowExecutionStartedEventAttributes)}
	case shared.EventTypeChildWorkflowExecutionCompleted:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionCompletedEventAttributes{ChildWorkflowExecutionCompletedEventAttributes: ChildWorkflowExecutionCompletedEventAttributes(e.ChildWorkflowExecutionCompletedEventAttributes)}
	case shared.EventTypeChildWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionFailedEventAttributes{ChildWorkflowExecutionFailedEventAttributes: ChildWorkflowExecutionFailedEventAttributes(e.ChildWorkflowExecutionFailedEventAttributes)}
	case shared.EventTypeChildWorkflowExecutionCanceled:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionCanceledEventAttributes{ChildWorkflowExecutionCanceledEventAttributes: ChildWorkflowExecutionCanceledEventAttributes(e.ChildWorkflowExecutionCanceledEventAttributes)}
	case shared.EventTypeChildWorkflowExecutionTimedOut:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionTimedOutEventAttributes{ChildWorkflowExecutionTimedOutEventAttributes: ChildWorkflowExecutionTimedOutEventAttributes(e.ChildWorkflowExecutionTimedOutEventAttributes)}
	case shared.EventTypeChildWorkflowExecutionTerminated:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionTerminatedEventAttributes{ChildWorkflowExecutionTerminatedEventAttributes: ChildWorkflowExecutionTerminatedEventAttributes(e.ChildWorkflowExecutionTerminatedEventAttributes)}
	case shared.EventTypeSignalExternalWorkflowExecutionInitiated:
		event.Attributes = &apiv1.HistoryEvent_SignalExternalWorkflowExecutionInitiatedEventAttributes{SignalExternalWorkflowExecutionInitiatedEventAttributes: SignalExternalWorkflowExecutionInitiatedEventAttributes(e.SignalExternalWorkflowExecutionInitiatedEventAttributes)}
	case shared.EventTypeSignalExternalWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_SignalExternalWorkflowExecutionFailedEventAttributes{SignalExternalWorkflowExecutionFailedEventAttributes: SignalExternalWorkflowExecutionFailedEventAttributes(e.SignalExternalWorkflowExecutionFailedEventAttributes)}
	case shared.EventTypeExternalWorkflowExecutionSignaled:
		event.Attributes = &apiv1.HistoryEvent_ExternalWorkflowExecutionSignaledEventAttributes{ExternalWorkflowExecutionSignaledEventAttributes: ExternalWorkflowExecutionSignaledEventAttributes(e.ExternalWorkflowExecutionSignaledEventAttributes)}
	case shared.EventTypeUpsertWorkflowSearchAttributes:
		event.Attributes = &apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes{UpsertWorkflowSearchAttributesEventAttributes: UpsertWorkflowSearchAttributesEventAttributes(e.UpsertWorkflowSearchAttributesEventAttributes)}
	default:
		panic("unknown event type")
	}
	return &event
}

func ActivityTaskCancelRequestedEventAttributes(t *shared.ActivityTaskCancelRequestedEventAttributes) *apiv1.ActivityTaskCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskCancelRequestedEventAttributes{
		ActivityId:                   t.GetActivityId(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func ActivityTaskCanceledEventAttributes(t *shared.ActivityTaskCanceledEventAttributes) *apiv1.ActivityTaskCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskCanceledEventAttributes{
		Details:                      Payload(t.Details),
		LatestCancelRequestedEventId: t.GetLatestCancelRequestedEventId(),
		ScheduledEventId:             t.GetScheduledEventId(),
		StartedEventId:               t.GetStartedEventId(),
		Identity:                     t.GetIdentity(),
	}
}

func ActivityTaskCompletedEventAttributes(t *shared.ActivityTaskCompletedEventAttributes) *apiv1.ActivityTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskCompletedEventAttributes{
		Result:           Payload(t.Result),
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Identity:         t.GetIdentity(),
	}
}

func ActivityTaskFailedEventAttributes(t *shared.ActivityTaskFailedEventAttributes) *apiv1.ActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskFailedEventAttributes{
		Failure:          Failure(t.Reason, t.Details),
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Identity:         t.GetIdentity(),
	}
}

func ActivityTaskScheduledEventAttributes(t *shared.ActivityTaskScheduledEventAttributes) *apiv1.ActivityTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskScheduledEventAttributes{
		ActivityId:                   t.GetActivityId(),
		ActivityType:                 ActivityType(t.ActivityType),
		Domain:                       t.GetDomain(),
		TaskList:                     TaskList(t.TaskList),
		Input:                        Payload(t.Input),
		ScheduleToCloseTimeout:       secondsToDuration(t.ScheduleToCloseTimeoutSeconds),
		ScheduleToStartTimeout:       secondsToDuration(t.ScheduleToStartTimeoutSeconds),
		StartToCloseTimeout:          secondsToDuration(t.StartToCloseTimeoutSeconds),
		HeartbeatTimeout:             secondsToDuration(t.HeartbeatTimeoutSeconds),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		RetryPolicy:                  RetryPolicy(t.RetryPolicy),
		Header:                       Header(t.Header),
	}
}

func ActivityTaskStartedEventAttributes(t *shared.ActivityTaskStartedEventAttributes) *apiv1.ActivityTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskStartedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		Identity:         t.GetIdentity(),
		RequestId:        t.GetRequestId(),
		Attempt:          t.GetAttempt(),
		LastFailure:      Failure(t.LastFailureReason, t.LastFailureDetails),
	}
}

func ActivityTaskTimedOutEventAttributes(t *shared.ActivityTaskTimedOutEventAttributes) *apiv1.ActivityTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskTimedOutEventAttributes{
		Details:          Payload(t.Details),
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		TimeoutType:      TimeoutType(t.TimeoutType),
		LastFailure:      Failure(t.LastFailureReason, t.LastFailureDetails),
	}
}

func CancelTimerFailedEventAttributes(t *shared.CancelTimerFailedEventAttributes) *apiv1.CancelTimerFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.CancelTimerFailedEventAttributes{
		TimerId:                      t.GetTimerId(),
		Cause:                        t.GetCause(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Identity:                     t.GetIdentity(),
	}
}

func ChildWorkflowExecutionCanceledEventAttributes(t *shared.ChildWorkflowExecutionCanceledEventAttributes) *apiv1.ChildWorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionCanceledEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		Details:           Payload(t.Details),
	}
}

func ChildWorkflowExecutionCompletedEventAttributes(t *shared.ChildWorkflowExecutionCompletedEventAttributes) *apiv1.ChildWorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionCompletedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		Result:            Payload(t.Result),
	}
}

func ChildWorkflowExecutionFailedEventAttributes(t *shared.ChildWorkflowExecutionFailedEventAttributes) *apiv1.ChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionFailedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		Failure:           Failure(t.Reason, t.Details),
	}
}

func ChildWorkflowExecutionStartedEventAttributes(t *shared.ChildWorkflowExecutionStartedEventAttributes) *apiv1.ChildWorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionStartedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		Header:            Header(t.Header),
	}
}

func ChildWorkflowExecutionTerminatedEventAttributes(t *shared.ChildWorkflowExecutionTerminatedEventAttributes) *apiv1.ChildWorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionTerminatedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
	}
}

func ChildWorkflowExecutionTimedOutEventAttributes(t *shared.ChildWorkflowExecutionTimedOutEventAttributes) *apiv1.ChildWorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionTimedOutEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		TimeoutType:       TimeoutType(t.TimeoutType),
	}
}

func DecisionTaskFailedEventAttributes(t *shared.DecisionTaskFailedEventAttributes) *apiv1.DecisionTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskFailedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Cause:            DecisionTaskFailedCause(t.Cause),
		Failure:          Failure(t.Reason, t.Details),
		Identity:         t.GetIdentity(),
		BaseRunId:        t.GetBaseRunId(),
		NewRunId:         t.GetNewRunId(),
		ForkEventVersion: t.GetForkEventVersion(),
		BinaryChecksum:   t.GetBinaryChecksum(),
	}
}

func DecisionTaskScheduledEventAttributes(t *shared.DecisionTaskScheduledEventAttributes) *apiv1.DecisionTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskScheduledEventAttributes{
		TaskList:            TaskList(t.TaskList),
		StartToCloseTimeout: secondsToDuration(t.StartToCloseTimeoutSeconds),
		Attempt:             int32(t.GetAttempt()),
	}
}

func DecisionTaskStartedEventAttributes(t *shared.DecisionTaskStartedEventAttributes) *apiv1.DecisionTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskStartedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		Identity:         t.GetIdentity(),
		RequestId:        t.GetRequestId(),
	}
}

func DecisionTaskCompletedEventAttributes(t *shared.DecisionTaskCompletedEventAttributes) *apiv1.DecisionTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskCompletedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Identity:         t.GetIdentity(),
		BinaryChecksum:   t.GetBinaryChecksum(),
		ExecutionContext: t.ExecutionContext,
	}
}

func DecisionTaskTimedOutEventAttributes(t *shared.DecisionTaskTimedOutEventAttributes) *apiv1.DecisionTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskTimedOutEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		TimeoutType:      TimeoutType(t.TimeoutType),
		BaseRunId:        t.GetBaseRunId(),
		NewRunId:         t.GetNewRunId(),
		ForkEventVersion: t.GetForkEventVersion(),
		Reason:           t.GetReason(),
		Cause:            DecisionTaskTimedOutCause(t.Cause),
	}
}

func ExternalWorkflowExecutionCancelRequestedEventAttributes(t *shared.ExternalWorkflowExecutionCancelRequestedEventAttributes) *apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes{
		InitiatedEventId:  t.GetInitiatedEventId(),
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
	}
}

func ExternalWorkflowExecutionSignaledEventAttributes(t *shared.ExternalWorkflowExecutionSignaledEventAttributes) *apiv1.ExternalWorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ExternalWorkflowExecutionSignaledEventAttributes{
		InitiatedEventId:  t.GetInitiatedEventId(),
		Domain:            t.GetDomain(),
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		Control:           t.Control,
	}
}

func MarkerRecordedEventAttributes(t *shared.MarkerRecordedEventAttributes) *apiv1.MarkerRecordedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.MarkerRecordedEventAttributes{
		MarkerName:                   t.GetMarkerName(),
		Details:                      Payload(t.Details),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Header:                       Header(t.Header),
	}
}

func RequestCancelActivityTaskFailedEventAttributes(t *shared.RequestCancelActivityTaskFailedEventAttributes) *apiv1.RequestCancelActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelActivityTaskFailedEventAttributes{
		ActivityId:                   t.GetActivityId(),
		Cause:                        t.GetCause(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func RequestCancelExternalWorkflowExecutionFailedEventAttributes(t *shared.RequestCancelExternalWorkflowExecutionFailedEventAttributes) *apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        CancelExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             t.GetInitiatedEventId(),
		Control:                      t.Control,
	}
}

func RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) *apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		Control:                      t.Control,
		ChildWorkflowOnly:            t.GetChildWorkflowOnly(),
	}
}

func SignalExternalWorkflowExecutionFailedEventAttributes(t *shared.SignalExternalWorkflowExecutionFailedEventAttributes) *apiv1.SignalExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.SignalExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        SignalExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             t.GetInitiatedEventId(),
		Control:                      t.Control,
	}
}

func SignalExternalWorkflowExecutionInitiatedEventAttributes(t *shared.SignalExternalWorkflowExecutionInitiatedEventAttributes) *apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		SignalName:                   t.GetSignalName(),
		Input:                        Payload(t.Input),
		Control:                      t.Control,
		ChildWorkflowOnly:            t.GetChildWorkflowOnly(),
	}
}

func StartChildWorkflowExecutionFailedEventAttributes(t *shared.StartChildWorkflowExecutionFailedEventAttributes) *apiv1.StartChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.StartChildWorkflowExecutionFailedEventAttributes{
		Domain:                       t.GetDomain(),
		WorkflowId:                   t.GetWorkflowId(),
		WorkflowType:                 WorkflowType(t.WorkflowType),
		Cause:                        ChildWorkflowExecutionFailedCause(t.Cause),
		Control:                      t.Control,
		InitiatedEventId:             t.GetInitiatedEventId(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func StartChildWorkflowExecutionInitiatedEventAttributes(t *shared.StartChildWorkflowExecutionInitiatedEventAttributes) *apiv1.StartChildWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.StartChildWorkflowExecutionInitiatedEventAttributes{
		Domain:                       t.GetDomain(),
		WorkflowId:                   t.GetWorkflowId(),
		WorkflowType:                 WorkflowType(t.WorkflowType),
		TaskList:                     TaskList(t.TaskList),
		Input:                        Payload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		ParentClosePolicy:            ParentClosePolicy(t.ParentClosePolicy),
		Control:                      t.Control,
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		WorkflowIdReusePolicy:        WorkflowIdReusePolicy(t.WorkflowIdReusePolicy),
		RetryPolicy:                  RetryPolicy(t.RetryPolicy),
		CronSchedule:                 t.GetCronSchedule(),
		Header:                       Header(t.Header),
		Memo:                         Memo(t.Memo),
		SearchAttributes:             SearchAttributes(t.SearchAttributes),
		DelayStart:                   secondsToDuration(t.DelayStartSeconds),
		JitterStart:                  secondsToDuration(t.JitterStartSeconds),
		FirstRunAt:                   unixNanoToTime(t.FirstRunAtTimestamp),
	}
}

func TimerCanceledEventAttributes(t *shared.TimerCanceledEventAttributes) *apiv1.TimerCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.TimerCanceledEventAttributes{
		TimerId:                      t.GetTimerId(),
		StartedEventId:               t.GetStartedEventId(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Identity:                     t.GetIdentity(),
	}
}

func TimerFiredEventAttributes(t *shared.TimerFiredEventAttributes) *apiv1.TimerFiredEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.TimerFiredEventAttributes{
		TimerId:        t.GetTimerId(),
		StartedEventId: t.GetStartedEventId(),
	}
}

func TimerStartedEventAttributes(t *shared.TimerStartedEventAttributes) *apiv1.TimerStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.TimerStartedEventAttributes{
		TimerId:                      t.GetTimerId(),
		StartToFireTimeout:           secondsToDuration(int64To32(t.StartToFireTimeoutSeconds)),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func UpsertWorkflowSearchAttributesEventAttributes(t *shared.UpsertWorkflowSearchAttributesEventAttributes) *apiv1.UpsertWorkflowSearchAttributesEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.UpsertWorkflowSearchAttributesEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		SearchAttributes:             SearchAttributes(t.SearchAttributes),
	}
}

func WorkflowExecutionCancelRequestedEventAttributes(t *shared.WorkflowExecutionCancelRequestedEventAttributes) *apiv1.WorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionCancelRequestedEventAttributes{
		Cause:                 t.GetCause(),
		ExternalExecutionInfo: ExternalExecutionInfo(t.ExternalWorkflowExecution, t.ExternalInitiatedEventId),
		Identity:              t.GetIdentity(),
	}
}

func WorkflowExecutionCanceledEventAttributes(t *shared.WorkflowExecutionCanceledEventAttributes) *apiv1.WorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionCanceledEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Details:                      Payload(t.Details),
	}
}

func WorkflowExecutionCompletedEventAttributes(t *shared.WorkflowExecutionCompletedEventAttributes) *apiv1.WorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionCompletedEventAttributes{
		Result:                       Payload(t.Result),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func WorkflowExecutionContinuedAsNewEventAttributes(t *shared.WorkflowExecutionContinuedAsNewEventAttributes) *apiv1.WorkflowExecutionContinuedAsNewEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionContinuedAsNewEventAttributes{
		NewExecutionRunId:            t.GetNewExecutionRunId(),
		WorkflowType:                 WorkflowType(t.WorkflowType),
		TaskList:                     TaskList(t.TaskList),
		Input:                        Payload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		BackoffStartInterval:         secondsToDuration(t.BackoffStartIntervalInSeconds),
		Initiator:                    ContinueAsNewInitiator(t.Initiator),
		Failure:                      Failure(t.FailureReason, t.FailureDetails),
		LastCompletionResult:         Payload(t.LastCompletionResult),
		Header:                       Header(t.Header),
		Memo:                         Memo(t.Memo),
		SearchAttributes:             SearchAttributes(t.SearchAttributes),
	}
}

func WorkflowExecutionFailedEventAttributes(t *shared.WorkflowExecutionFailedEventAttributes) *apiv1.WorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionFailedEventAttributes{
		Failure:                      Failure(t.Reason, t.Details),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func WorkflowExecutionSignaledEventAttributes(t *shared.WorkflowExecutionSignaledEventAttributes) *apiv1.WorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionSignaledEventAttributes{
		SignalName: t.GetSignalName(),
		Input:      Payload(t.Input),
		Identity:   t.GetIdentity(),
	}
}

func WorkflowExecutionStartedEventAttributes(t *shared.WorkflowExecutionStartedEventAttributes) *apiv1.WorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionStartedEventAttributes{
		WorkflowType:                 WorkflowType(t.WorkflowType),
		ParentExecutionInfo:          ParentExecutionInfo(nil, t.ParentWorkflowDomain, t.ParentWorkflowExecution, t.ParentInitiatedEventId),
		TaskList:                     TaskList(t.TaskList),
		Input:                        Payload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		ContinuedExecutionRunId:      t.GetContinuedExecutionRunId(),
		Initiator:                    ContinueAsNewInitiator(t.Initiator),
		ContinuedFailure:             Failure(t.ContinuedFailureReason, t.ContinuedFailureDetails),
		LastCompletionResult:         Payload(t.LastCompletionResult),
		OriginalExecutionRunId:       t.GetOriginalExecutionRunId(),
		Identity:                     t.GetIdentity(),
		FirstExecutionRunId:          t.GetFirstExecutionRunId(),
		RetryPolicy:                  RetryPolicy(t.RetryPolicy),
		Attempt:                      t.GetAttempt(),
		ExpirationTime:               unixNanoToTime(t.ExpirationTimestamp),
		CronSchedule:                 t.GetCronSchedule(),
		FirstDecisionTaskBackoff:     secondsToDuration(t.FirstDecisionTaskBackoffSeconds),
		Memo:                         Memo(t.Memo),
		SearchAttributes:             SearchAttributes(t.SearchAttributes),
		PrevAutoResetPoints:          ResetPoints(t.PrevAutoResetPoints),
		Header:                       Header(t.Header),
	}
}

func WorkflowExecutionTerminatedEventAttributes(t *shared.WorkflowExecutionTerminatedEventAttributes) *apiv1.WorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionTerminatedEventAttributes{
		Reason:   t.GetReason(),
		Details:  Payload(t.Details),
		Identity: t.GetIdentity(),
	}
}

func WorkflowExecutionTimedOutEventAttributes(t *shared.WorkflowExecutionTimedOutEventAttributes) *apiv1.WorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionTimedOutEventAttributes{
		TimeoutType: TimeoutType(t.TimeoutType),
	}
}
