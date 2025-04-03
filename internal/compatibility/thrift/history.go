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

func History(t *apiv1.History) *shared.History {
	if t == nil {
		return nil
	}
	return &shared.History{
		Events: HistoryEventArray(t.Events),
	}
}

func HistoryEventArray(t []*apiv1.HistoryEvent) []*shared.HistoryEvent {
	if t == nil {
		return nil
	}
	v := make([]*shared.HistoryEvent, len(t))
	for i := range t {
		v[i] = HistoryEvent(t[i])
	}
	return v
}

func HistoryEvent(e *apiv1.HistoryEvent) *shared.HistoryEvent {
	if e == nil {
		return nil
	}
	event := shared.HistoryEvent{
		EventId:   &e.EventId,
		Timestamp: timeToUnixNano(e.EventTime),
		Version:   &e.Version,
		TaskId:    &e.TaskId,
	}
	switch attr := e.Attributes.(type) {
	case *apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionStarted.Ptr()
		event.WorkflowExecutionStartedEventAttributes = WorkflowExecutionStartedEventAttributes(attr.WorkflowExecutionStartedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionCompleted.Ptr()
		event.WorkflowExecutionCompletedEventAttributes = WorkflowExecutionCompletedEventAttributes(attr.WorkflowExecutionCompletedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionFailed.Ptr()
		event.WorkflowExecutionFailedEventAttributes = WorkflowExecutionFailedEventAttributes(attr.WorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionTimedOutEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionTimedOut.Ptr()
		event.WorkflowExecutionTimedOutEventAttributes = WorkflowExecutionTimedOutEventAttributes(attr.WorkflowExecutionTimedOutEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskScheduled.Ptr()
		event.DecisionTaskScheduledEventAttributes = DecisionTaskScheduledEventAttributes(attr.DecisionTaskScheduledEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskStartedEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskStarted.Ptr()
		event.DecisionTaskStartedEventAttributes = DecisionTaskStartedEventAttributes(attr.DecisionTaskStartedEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskCompleted.Ptr()
		event.DecisionTaskCompletedEventAttributes = DecisionTaskCompletedEventAttributes(attr.DecisionTaskCompletedEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskTimedOutEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskTimedOut.Ptr()
		event.DecisionTaskTimedOutEventAttributes = DecisionTaskTimedOutEventAttributes(attr.DecisionTaskTimedOutEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskFailedEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskFailed.Ptr()
		event.DecisionTaskFailedEventAttributes = DecisionTaskFailedEventAttributes(attr.DecisionTaskFailedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes:
		event.EventType = shared.EventTypeActivityTaskScheduled.Ptr()
		event.ActivityTaskScheduledEventAttributes = ActivityTaskScheduledEventAttributes(attr.ActivityTaskScheduledEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskStartedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskStarted.Ptr()
		event.ActivityTaskStartedEventAttributes = ActivityTaskStartedEventAttributes(attr.ActivityTaskStartedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskCompletedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskCompleted.Ptr()
		event.ActivityTaskCompletedEventAttributes = ActivityTaskCompletedEventAttributes(attr.ActivityTaskCompletedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskFailedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskFailed.Ptr()
		event.ActivityTaskFailedEventAttributes = ActivityTaskFailedEventAttributes(attr.ActivityTaskFailedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskTimedOutEventAttributes:
		event.EventType = shared.EventTypeActivityTaskTimedOut.Ptr()
		event.ActivityTaskTimedOutEventAttributes = ActivityTaskTimedOutEventAttributes(attr.ActivityTaskTimedOutEventAttributes)
	case *apiv1.HistoryEvent_TimerStartedEventAttributes:
		event.EventType = shared.EventTypeTimerStarted.Ptr()
		event.TimerStartedEventAttributes = TimerStartedEventAttributes(attr.TimerStartedEventAttributes)
	case *apiv1.HistoryEvent_TimerFiredEventAttributes:
		event.EventType = shared.EventTypeTimerFired.Ptr()
		event.TimerFiredEventAttributes = TimerFiredEventAttributes(attr.TimerFiredEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskCancelRequestedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskCancelRequested.Ptr()
		event.ActivityTaskCancelRequestedEventAttributes = ActivityTaskCancelRequestedEventAttributes(attr.ActivityTaskCancelRequestedEventAttributes)
	case *apiv1.HistoryEvent_RequestCancelActivityTaskFailedEventAttributes:
		event.EventType = shared.EventTypeRequestCancelActivityTaskFailed.Ptr()
		event.RequestCancelActivityTaskFailedEventAttributes = RequestCancelActivityTaskFailedEventAttributes(attr.RequestCancelActivityTaskFailedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskCanceledEventAttributes:
		event.EventType = shared.EventTypeActivityTaskCanceled.Ptr()
		event.ActivityTaskCanceledEventAttributes = ActivityTaskCanceledEventAttributes(attr.ActivityTaskCanceledEventAttributes)
	case *apiv1.HistoryEvent_TimerCanceledEventAttributes:
		event.EventType = shared.EventTypeTimerCanceled.Ptr()
		event.TimerCanceledEventAttributes = TimerCanceledEventAttributes(attr.TimerCanceledEventAttributes)
	case *apiv1.HistoryEvent_CancelTimerFailedEventAttributes:
		event.EventType = shared.EventTypeCancelTimerFailed.Ptr()
		event.CancelTimerFailedEventAttributes = CancelTimerFailedEventAttributes(attr.CancelTimerFailedEventAttributes)
	case *apiv1.HistoryEvent_MarkerRecordedEventAttributes:
		event.EventType = shared.EventTypeMarkerRecorded.Ptr()
		event.MarkerRecordedEventAttributes = MarkerRecordedEventAttributes(attr.MarkerRecordedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionSignaled.Ptr()
		event.WorkflowExecutionSignaledEventAttributes = WorkflowExecutionSignaledEventAttributes(attr.WorkflowExecutionSignaledEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionTerminatedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionTerminated.Ptr()
		event.WorkflowExecutionTerminatedEventAttributes = WorkflowExecutionTerminatedEventAttributes(attr.WorkflowExecutionTerminatedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionCancelRequested.Ptr()
		event.WorkflowExecutionCancelRequestedEventAttributes = WorkflowExecutionCancelRequestedEventAttributes(attr.WorkflowExecutionCancelRequestedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionCanceledEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionCanceled.Ptr()
		event.WorkflowExecutionCanceledEventAttributes = WorkflowExecutionCanceledEventAttributes(attr.WorkflowExecutionCanceledEventAttributes)
	case *apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiatedEventAttributes:
		event.EventType = shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated.Ptr()
		event.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes = RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(attr.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes)
	case *apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeRequestCancelExternalWorkflowExecutionFailed.Ptr()
		event.RequestCancelExternalWorkflowExecutionFailedEventAttributes = RequestCancelExternalWorkflowExecutionFailedEventAttributes(attr.RequestCancelExternalWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ExternalWorkflowExecutionCancelRequestedEventAttributes:
		event.EventType = shared.EventTypeExternalWorkflowExecutionCancelRequested.Ptr()
		event.ExternalWorkflowExecutionCancelRequestedEventAttributes = ExternalWorkflowExecutionCancelRequestedEventAttributes(attr.ExternalWorkflowExecutionCancelRequestedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionContinuedAsNew.Ptr()
		event.WorkflowExecutionContinuedAsNewEventAttributes = WorkflowExecutionContinuedAsNewEventAttributes(attr.WorkflowExecutionContinuedAsNewEventAttributes)
	case *apiv1.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes:
		event.EventType = shared.EventTypeStartChildWorkflowExecutionInitiated.Ptr()
		event.StartChildWorkflowExecutionInitiatedEventAttributes = StartChildWorkflowExecutionInitiatedEventAttributes(attr.StartChildWorkflowExecutionInitiatedEventAttributes)
	case *apiv1.HistoryEvent_StartChildWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeStartChildWorkflowExecutionFailed.Ptr()
		event.StartChildWorkflowExecutionFailedEventAttributes = StartChildWorkflowExecutionFailedEventAttributes(attr.StartChildWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionStartedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionStarted.Ptr()
		event.ChildWorkflowExecutionStartedEventAttributes = ChildWorkflowExecutionStartedEventAttributes(attr.ChildWorkflowExecutionStartedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionCompletedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionCompleted.Ptr()
		event.ChildWorkflowExecutionCompletedEventAttributes = ChildWorkflowExecutionCompletedEventAttributes(attr.ChildWorkflowExecutionCompletedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionFailed.Ptr()
		event.ChildWorkflowExecutionFailedEventAttributes = ChildWorkflowExecutionFailedEventAttributes(attr.ChildWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionCanceledEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionCanceled.Ptr()
		event.ChildWorkflowExecutionCanceledEventAttributes = ChildWorkflowExecutionCanceledEventAttributes(attr.ChildWorkflowExecutionCanceledEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionTimedOutEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionTimedOut.Ptr()
		event.ChildWorkflowExecutionTimedOutEventAttributes = ChildWorkflowExecutionTimedOutEventAttributes(attr.ChildWorkflowExecutionTimedOutEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionTerminatedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionTerminated.Ptr()
		event.ChildWorkflowExecutionTerminatedEventAttributes = ChildWorkflowExecutionTerminatedEventAttributes(attr.ChildWorkflowExecutionTerminatedEventAttributes)
	case *apiv1.HistoryEvent_SignalExternalWorkflowExecutionInitiatedEventAttributes:
		event.EventType = shared.EventTypeSignalExternalWorkflowExecutionInitiated.Ptr()
		event.SignalExternalWorkflowExecutionInitiatedEventAttributes = SignalExternalWorkflowExecutionInitiatedEventAttributes(attr.SignalExternalWorkflowExecutionInitiatedEventAttributes)
	case *apiv1.HistoryEvent_SignalExternalWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeSignalExternalWorkflowExecutionFailed.Ptr()
		event.SignalExternalWorkflowExecutionFailedEventAttributes = SignalExternalWorkflowExecutionFailedEventAttributes(attr.SignalExternalWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ExternalWorkflowExecutionSignaledEventAttributes:
		event.EventType = shared.EventTypeExternalWorkflowExecutionSignaled.Ptr()
		event.ExternalWorkflowExecutionSignaledEventAttributes = ExternalWorkflowExecutionSignaledEventAttributes(attr.ExternalWorkflowExecutionSignaledEventAttributes)
	case *apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes:
		event.EventType = shared.EventTypeUpsertWorkflowSearchAttributes.Ptr()
		event.UpsertWorkflowSearchAttributesEventAttributes = UpsertWorkflowSearchAttributesEventAttributes(attr.UpsertWorkflowSearchAttributesEventAttributes)
	default:
		panic("unknown event type")
	}

	return &event
}

func ActivityTaskCancelRequestedEventAttributes(t *apiv1.ActivityTaskCancelRequestedEventAttributes) *shared.ActivityTaskCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskCancelRequestedEventAttributes{
		ActivityId:                   &t.ActivityId,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func ActivityTaskCanceledEventAttributes(t *apiv1.ActivityTaskCanceledEventAttributes) *shared.ActivityTaskCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskCanceledEventAttributes{
		Details:                      Payload(t.Details),
		LatestCancelRequestedEventId: &t.LatestCancelRequestedEventId,
		ScheduledEventId:             &t.ScheduledEventId,
		StartedEventId:               &t.StartedEventId,
		Identity:                     &t.Identity,
	}
}

func ActivityTaskCompletedEventAttributes(t *apiv1.ActivityTaskCompletedEventAttributes) *shared.ActivityTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskCompletedEventAttributes{
		Result:           Payload(t.Result),
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Identity:         &t.Identity,
	}
}

func ActivityTaskFailedEventAttributes(t *apiv1.ActivityTaskFailedEventAttributes) *shared.ActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskFailedEventAttributes{
		Reason:           FailureReason(t.Failure),
		Details:          FailureDetails(t.Failure),
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Identity:         &t.Identity,
	}
}

func ActivityTaskScheduledEventAttributes(t *apiv1.ActivityTaskScheduledEventAttributes) *shared.ActivityTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskScheduledEventAttributes{
		ActivityId:                    &t.ActivityId,
		ActivityType:                  ActivityType(t.ActivityType),
		Domain:                        &t.Domain,
		TaskList:                      TaskList(t.TaskList),
		Input:                         Payload(t.Input),
		ScheduleToCloseTimeoutSeconds: durationToSeconds(t.ScheduleToCloseTimeout),
		ScheduleToStartTimeoutSeconds: durationToSeconds(t.ScheduleToStartTimeout),
		StartToCloseTimeoutSeconds:    durationToSeconds(t.StartToCloseTimeout),
		HeartbeatTimeoutSeconds:       durationToSeconds(t.HeartbeatTimeout),
		DecisionTaskCompletedEventId:  &t.DecisionTaskCompletedEventId,
		RetryPolicy:                   RetryPolicy(t.RetryPolicy),
		Header:                        Header(t.Header),
	}
}

func ActivityTaskStartedEventAttributes(t *apiv1.ActivityTaskStartedEventAttributes) *shared.ActivityTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskStartedEventAttributes{
		ScheduledEventId:   &t.ScheduledEventId,
		Identity:           &t.Identity,
		RequestId:          &t.RequestId,
		Attempt:            &t.Attempt,
		LastFailureReason:  FailureReason(t.LastFailure),
		LastFailureDetails: FailureDetails(t.LastFailure),
	}
}

func ActivityTaskTimedOutEventAttributes(t *apiv1.ActivityTaskTimedOutEventAttributes) *shared.ActivityTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskTimedOutEventAttributes{
		Details:            Payload(t.Details),
		ScheduledEventId:   &t.ScheduledEventId,
		StartedEventId:     &t.StartedEventId,
		TimeoutType:        TimeoutType(t.TimeoutType),
		LastFailureReason:  FailureReason(t.LastFailure),
		LastFailureDetails: FailureDetails(t.LastFailure),
	}
}

func CancelTimerFailedEventAttributes(t *apiv1.CancelTimerFailedEventAttributes) *shared.CancelTimerFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.CancelTimerFailedEventAttributes{
		TimerId:                      &t.TimerId,
		Cause:                        &t.Cause,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Identity:                     &t.Identity,
	}
}

func ChildWorkflowExecutionCanceledEventAttributes(t *apiv1.ChildWorkflowExecutionCanceledEventAttributes) *shared.ChildWorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionCanceledEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		Details:           Payload(t.Details),
	}
}

func ChildWorkflowExecutionCompletedEventAttributes(t *apiv1.ChildWorkflowExecutionCompletedEventAttributes) *shared.ChildWorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionCompletedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		Result:            Payload(t.Result),
	}
}

func ChildWorkflowExecutionFailedEventAttributes(t *apiv1.ChildWorkflowExecutionFailedEventAttributes) *shared.ChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionFailedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		Reason:            FailureReason(t.Failure),
		Details:           FailureDetails(t.Failure),
	}
}

func ChildWorkflowExecutionStartedEventAttributes(t *apiv1.ChildWorkflowExecutionStartedEventAttributes) *shared.ChildWorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionStartedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		Header:            Header(t.Header),
	}
}

func ChildWorkflowExecutionTerminatedEventAttributes(t *apiv1.ChildWorkflowExecutionTerminatedEventAttributes) *shared.ChildWorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionTerminatedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
	}
}

func ChildWorkflowExecutionTimedOutEventAttributes(t *apiv1.ChildWorkflowExecutionTimedOutEventAttributes) *shared.ChildWorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionTimedOutEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		WorkflowType:      WorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		TimeoutType:       TimeoutType(t.TimeoutType),
	}
}

func DecisionTaskFailedEventAttributes(t *apiv1.DecisionTaskFailedEventAttributes) *shared.DecisionTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskFailedEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Cause:            DecisionTaskFailedCause(t.Cause),
		Reason:           FailureReason(t.Failure),
		Details:          FailureDetails(t.Failure),
		Identity:         &t.Identity,
		BaseRunId:        &t.BaseRunId,
		NewRunId:         &t.NewRunId,
		ForkEventVersion: &t.ForkEventVersion,
		BinaryChecksum:   &t.BinaryChecksum,
	}
}

func DecisionTaskScheduledEventAttributes(t *apiv1.DecisionTaskScheduledEventAttributes) *shared.DecisionTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskScheduledEventAttributes{
		TaskList:                   TaskList(t.TaskList),
		StartToCloseTimeoutSeconds: durationToSeconds(t.StartToCloseTimeout),
		Attempt:                    common.Int64Ptr(int64(t.Attempt)),
	}
}

func DecisionTaskStartedEventAttributes(t *apiv1.DecisionTaskStartedEventAttributes) *shared.DecisionTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskStartedEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		Identity:         &t.Identity,
		RequestId:        &t.RequestId,
	}
}

func DecisionTaskCompletedEventAttributes(t *apiv1.DecisionTaskCompletedEventAttributes) *shared.DecisionTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskCompletedEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Identity:         &t.Identity,
		BinaryChecksum:   &t.BinaryChecksum,
		ExecutionContext: t.ExecutionContext,
	}
}

func DecisionTaskTimedOutEventAttributes(t *apiv1.DecisionTaskTimedOutEventAttributes) *shared.DecisionTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskTimedOutEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		TimeoutType:      TimeoutType(t.TimeoutType),
		BaseRunId:        &t.BaseRunId,
		NewRunId:         &t.NewRunId,
		ForkEventVersion: &t.ForkEventVersion,
		Reason:           &t.Reason,
		Cause:            DecisionTaskTimedOutCause(t.Cause),
	}
}

func ExternalWorkflowExecutionCancelRequestedEventAttributes(t *apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes) *shared.ExternalWorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ExternalWorkflowExecutionCancelRequestedEventAttributes{
		InitiatedEventId:  &t.InitiatedEventId,
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
	}
}

func ExternalWorkflowExecutionSignaledEventAttributes(t *apiv1.ExternalWorkflowExecutionSignaledEventAttributes) *shared.ExternalWorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ExternalWorkflowExecutionSignaledEventAttributes{
		InitiatedEventId:  &t.InitiatedEventId,
		Domain:            &t.Domain,
		WorkflowExecution: WorkflowExecution(t.WorkflowExecution),
		Control:           t.Control,
	}
}

func MarkerRecordedEventAttributes(t *apiv1.MarkerRecordedEventAttributes) *shared.MarkerRecordedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.MarkerRecordedEventAttributes{
		MarkerName:                   &t.MarkerName,
		Details:                      Payload(t.Details),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Header:                       Header(t.Header),
	}
}

func RequestCancelActivityTaskFailedEventAttributes(t *apiv1.RequestCancelActivityTaskFailedEventAttributes) *shared.RequestCancelActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.RequestCancelActivityTaskFailedEventAttributes{
		ActivityId:                   &t.ActivityId,
		Cause:                        &t.Cause,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func RequestCancelExternalWorkflowExecutionFailedEventAttributes(t *apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes) *shared.RequestCancelExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.RequestCancelExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        CancelExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             &t.InitiatedEventId,
		Control:                      t.Control,
	}
}

func RequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) *shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		Control:                      t.Control,
		ChildWorkflowOnly:            &t.ChildWorkflowOnly,
	}
}

func SignalExternalWorkflowExecutionFailedEventAttributes(t *apiv1.SignalExternalWorkflowExecutionFailedEventAttributes) *shared.SignalExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.SignalExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        SignalExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             &t.InitiatedEventId,
		Control:                      t.Control,
	}
}

func SignalExternalWorkflowExecutionInitiatedEventAttributes(t *apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes) *shared.SignalExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.SignalExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            WorkflowExecution(t.WorkflowExecution),
		SignalName:                   &t.SignalName,
		Input:                        Payload(t.Input),
		Control:                      t.Control,
		ChildWorkflowOnly:            &t.ChildWorkflowOnly,
	}
}

func StartChildWorkflowExecutionFailedEventAttributes(t *apiv1.StartChildWorkflowExecutionFailedEventAttributes) *shared.StartChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.StartChildWorkflowExecutionFailedEventAttributes{
		Domain:                       &t.Domain,
		WorkflowId:                   &t.WorkflowId,
		WorkflowType:                 WorkflowType(t.WorkflowType),
		Cause:                        ChildWorkflowExecutionFailedCause(t.Cause),
		Control:                      t.Control,
		InitiatedEventId:             &t.InitiatedEventId,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func StartChildWorkflowExecutionInitiatedEventAttributes(t *apiv1.StartChildWorkflowExecutionInitiatedEventAttributes) *shared.StartChildWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.StartChildWorkflowExecutionInitiatedEventAttributes{
		Domain:                              &t.Domain,
		WorkflowId:                          &t.WorkflowId,
		WorkflowType:                        WorkflowType(t.WorkflowType),
		TaskList:                            TaskList(t.TaskList),
		Input:                               Payload(t.Input),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		ParentClosePolicy:                   ParentClosePolicy(t.ParentClosePolicy),
		Control:                             t.Control,
		DecisionTaskCompletedEventId:        &t.DecisionTaskCompletedEventId,
		WorkflowIdReusePolicy:               WorkflowIdReusePolicy(t.WorkflowIdReusePolicy),
		RetryPolicy:                         RetryPolicy(t.RetryPolicy),
		CronSchedule:                        &t.CronSchedule,
		Header:                              Header(t.Header),
		Memo:                                Memo(t.Memo),
		SearchAttributes:                    SearchAttributes(t.SearchAttributes),
		DelayStartSeconds:                   durationToSeconds(t.DelayStart),
		JitterStartSeconds:                  durationToSeconds(t.JitterStart),
		FirstRunAtTimestamp:                 timeToUnixNano(t.FirstRunAt),
	}
}

func TimerCanceledEventAttributes(t *apiv1.TimerCanceledEventAttributes) *shared.TimerCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.TimerCanceledEventAttributes{
		TimerId:                      &t.TimerId,
		StartedEventId:               &t.StartedEventId,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Identity:                     &t.Identity,
	}
}

func TimerFiredEventAttributes(t *apiv1.TimerFiredEventAttributes) *shared.TimerFiredEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.TimerFiredEventAttributes{
		TimerId:        &t.TimerId,
		StartedEventId: &t.StartedEventId,
	}
}

func TimerStartedEventAttributes(t *apiv1.TimerStartedEventAttributes) *shared.TimerStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.TimerStartedEventAttributes{
		TimerId:                      &t.TimerId,
		StartToFireTimeoutSeconds:    int32To64(durationToSeconds(t.StartToFireTimeout)),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func UpsertWorkflowSearchAttributesEventAttributes(t *apiv1.UpsertWorkflowSearchAttributesEventAttributes) *shared.UpsertWorkflowSearchAttributesEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.UpsertWorkflowSearchAttributesEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		SearchAttributes:             SearchAttributes(t.SearchAttributes),
	}
}

func WorkflowExecutionCancelRequestedEventAttributes(t *apiv1.WorkflowExecutionCancelRequestedEventAttributes) *shared.WorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionCancelRequestedEventAttributes{
		Cause:                     &t.Cause,
		ExternalInitiatedEventId:  ExternalInitiatedId(t.ExternalExecutionInfo),
		ExternalWorkflowExecution: ExternalWorkflowExecution(t.ExternalExecutionInfo),
		Identity:                  &t.Identity,
	}
}

func WorkflowExecutionCanceledEventAttributes(t *apiv1.WorkflowExecutionCanceledEventAttributes) *shared.WorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionCanceledEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Details:                      Payload(t.Details),
	}
}

func WorkflowExecutionCompletedEventAttributes(t *apiv1.WorkflowExecutionCompletedEventAttributes) *shared.WorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionCompletedEventAttributes{
		Result:                       Payload(t.Result),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func WorkflowExecutionContinuedAsNewEventAttributes(t *apiv1.WorkflowExecutionContinuedAsNewEventAttributes) *shared.WorkflowExecutionContinuedAsNewEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionContinuedAsNewEventAttributes{
		NewExecutionRunId:                   &t.NewExecutionRunId,
		WorkflowType:                        WorkflowType(t.WorkflowType),
		TaskList:                            TaskList(t.TaskList),
		Input:                               Payload(t.Input),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		DecisionTaskCompletedEventId:        &t.DecisionTaskCompletedEventId,
		BackoffStartIntervalInSeconds:       durationToSeconds(t.BackoffStartInterval),
		Initiator:                           ContinueAsNewInitiator(t.Initiator),
		FailureReason:                       FailureReason(t.Failure),
		FailureDetails:                      FailureDetails(t.Failure),
		LastCompletionResult:                Payload(t.LastCompletionResult),
		Header:                              Header(t.Header),
		Memo:                                Memo(t.Memo),
		SearchAttributes:                    SearchAttributes(t.SearchAttributes),
	}
}

func WorkflowExecutionFailedEventAttributes(t *apiv1.WorkflowExecutionFailedEventAttributes) *shared.WorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionFailedEventAttributes{
		Reason:                       FailureReason(t.Failure),
		Details:                      FailureDetails(t.Failure),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func WorkflowExecutionSignaledEventAttributes(t *apiv1.WorkflowExecutionSignaledEventAttributes) *shared.WorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionSignaledEventAttributes{
		SignalName: &t.SignalName,
		Input:      Payload(t.Input),
		Identity:   &t.Identity,
	}
}

func WorkflowExecutionStartedEventAttributes(t *apiv1.WorkflowExecutionStartedEventAttributes) *shared.WorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionStartedEventAttributes{
		WorkflowType:                        WorkflowType(t.WorkflowType),
		ParentWorkflowDomain:                ParentDomainName(t.ParentExecutionInfo),
		ParentWorkflowExecution:             ParentWorkflowExecution(t.ParentExecutionInfo),
		ParentInitiatedEventId:              ParentInitiatedId(t.ParentExecutionInfo),
		TaskList:                            TaskList(t.TaskList),
		Input:                               Payload(t.Input),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		ContinuedExecutionRunId:             &t.ContinuedExecutionRunId,
		Initiator:                           ContinueAsNewInitiator(t.Initiator),
		ContinuedFailureReason:              FailureReason(t.ContinuedFailure),
		ContinuedFailureDetails:             FailureDetails(t.ContinuedFailure),
		LastCompletionResult:                Payload(t.LastCompletionResult),
		OriginalExecutionRunId:              &t.OriginalExecutionRunId,
		Identity:                            &t.Identity,
		FirstExecutionRunId:                 &t.FirstExecutionRunId,
		RetryPolicy:                         RetryPolicy(t.RetryPolicy),
		Attempt:                             &t.Attempt,
		ExpirationTimestamp:                 timeToUnixNano(t.ExpirationTime),
		CronSchedule:                        &t.CronSchedule,
		FirstDecisionTaskBackoffSeconds:     durationToSeconds(t.FirstDecisionTaskBackoff),
		Memo:                                Memo(t.Memo),
		SearchAttributes:                    SearchAttributes(t.SearchAttributes),
		PrevAutoResetPoints:                 ResetPoints(t.PrevAutoResetPoints),
		Header:                              Header(t.Header),
	}
}

func WorkflowExecutionTerminatedEventAttributes(t *apiv1.WorkflowExecutionTerminatedEventAttributes) *shared.WorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionTerminatedEventAttributes{
		Reason:   &t.Reason,
		Details:  Payload(t.Details),
		Identity: &t.Identity,
	}
}

func WorkflowExecutionTimedOutEventAttributes(t *apiv1.WorkflowExecutionTimedOutEventAttributes) *shared.WorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionTimedOutEventAttributes{
		TimeoutType: TimeoutType(t.TimeoutType),
	}
}
