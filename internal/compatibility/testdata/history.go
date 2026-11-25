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
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

var (
	History = apiv1.History{
		Events: HistoryEventArray,
	}
	HistoryEventArray = []*apiv1.HistoryEvent{
		&HistoryEvent_WorkflowExecutionStarted,
		&HistoryEvent_WorkflowExecutionStarted,
		&HistoryEvent_WorkflowExecutionCompleted,
		&HistoryEvent_WorkflowExecutionFailed,
		&HistoryEvent_WorkflowExecutionTimedOut,
		&HistoryEvent_DecisionTaskScheduled,
		&HistoryEvent_DecisionTaskStarted,
		&HistoryEvent_DecisionTaskCompleted,
		&HistoryEvent_DecisionTaskTimedOut,
		&HistoryEvent_DecisionTaskFailed,
		&HistoryEvent_ActivityTaskScheduled,
		&HistoryEvent_ActivityTaskStarted,
		&HistoryEvent_ActivityTaskCompleted,
		&HistoryEvent_ActivityTaskFailed,
		&HistoryEvent_ActivityTaskTimedOut,
		&HistoryEvent_ActivityTaskCancelRequested,
		&HistoryEvent_RequestCancelActivityTaskFailed,
		&HistoryEvent_ActivityTaskCanceled,
		&HistoryEvent_TimerStarted,
		&HistoryEvent_TimerFired,
		&HistoryEvent_CancelTimerFailed,
		&HistoryEvent_TimerCanceled,
		&HistoryEvent_WorkflowExecutionCancelRequested,
		&HistoryEvent_WorkflowExecutionCanceled,
		&HistoryEvent_RequestCancelExternalWorkflowExecutionInitiated,
		&HistoryEvent_RequestCancelExternalWorkflowExecutionFailed,
		&HistoryEvent_ExternalWorkflowExecutionCancelRequested,
		&HistoryEvent_MarkerRecorded,
		&HistoryEvent_WorkflowExecutionSignaled,
		&HistoryEvent_WorkflowExecutionTerminated,
		&HistoryEvent_WorkflowExecutionContinuedAsNew,
		&HistoryEvent_StartChildWorkflowExecutionInitiated,
		&HistoryEvent_StartChildWorkflowExecutionFailed,
		&HistoryEvent_ChildWorkflowExecutionStarted,
		&HistoryEvent_ChildWorkflowExecutionCompleted,
		&HistoryEvent_ChildWorkflowExecutionFailed,
		&HistoryEvent_ChildWorkflowExecutionCanceled,
		&HistoryEvent_ChildWorkflowExecutionTimedOut,
		&HistoryEvent_ChildWorkflowExecutionTerminated,
		&HistoryEvent_SignalExternalWorkflowExecutionInitiated,
		&HistoryEvent_SignalExternalWorkflowExecutionFailed,
		&HistoryEvent_ExternalWorkflowExecutionSignaled,
		&HistoryEvent_UpsertWorkflowSearchAttributes,
		&HistoryEvent_SemaphoreAcquireInitiated,
		&HistoryEvent_SemaphoreAcquired,
		&HistoryEvent_SemaphoreReleased,
		&HistoryEvent_SemaphoreAcquireFailed,
		&HistoryEvent_SemaphoreTokenLost,
	}

	HistoryEvent_WorkflowExecutionStarted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{WorkflowExecutionStartedEventAttributes: &WorkflowExecutionStartedEventAttributes}
	})
	HistoryEvent_WorkflowExecutionCompleted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{WorkflowExecutionCompletedEventAttributes: &WorkflowExecutionCompletedEventAttributes}
	})
	HistoryEvent_WorkflowExecutionFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionFailedEventAttributes{WorkflowExecutionFailedEventAttributes: &WorkflowExecutionFailedEventAttributes}
	})
	HistoryEvent_WorkflowExecutionTimedOut = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionTimedOutEventAttributes{WorkflowExecutionTimedOutEventAttributes: &WorkflowExecutionTimedOutEventAttributes}
	})
	HistoryEvent_DecisionTaskScheduled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes{DecisionTaskScheduledEventAttributes: &DecisionTaskScheduledEventAttributes}
	})
	HistoryEvent_DecisionTaskStarted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_DecisionTaskStartedEventAttributes{DecisionTaskStartedEventAttributes: &DecisionTaskStartedEventAttributes}
	})
	HistoryEvent_DecisionTaskCompleted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes{DecisionTaskCompletedEventAttributes: &DecisionTaskCompletedEventAttributes}
	})
	HistoryEvent_DecisionTaskTimedOut = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_DecisionTaskTimedOutEventAttributes{DecisionTaskTimedOutEventAttributes: &DecisionTaskTimedOutEventAttributes}
	})
	HistoryEvent_DecisionTaskFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_DecisionTaskFailedEventAttributes{DecisionTaskFailedEventAttributes: &DecisionTaskFailedEventAttributes}
	})
	HistoryEvent_ActivityTaskScheduled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: &ActivityTaskScheduledEventAttributes}
	})
	HistoryEvent_ActivityTaskStarted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ActivityTaskStartedEventAttributes{ActivityTaskStartedEventAttributes: &ActivityTaskStartedEventAttributes}
	})
	HistoryEvent_ActivityTaskCompleted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ActivityTaskCompletedEventAttributes{ActivityTaskCompletedEventAttributes: &ActivityTaskCompletedEventAttributes}
	})
	HistoryEvent_ActivityTaskFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ActivityTaskFailedEventAttributes{ActivityTaskFailedEventAttributes: &ActivityTaskFailedEventAttributes}
	})
	HistoryEvent_ActivityTaskTimedOut = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ActivityTaskTimedOutEventAttributes{ActivityTaskTimedOutEventAttributes: &ActivityTaskTimedOutEventAttributes}
	})
	HistoryEvent_ActivityTaskCancelRequested = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ActivityTaskCancelRequestedEventAttributes{ActivityTaskCancelRequestedEventAttributes: &ActivityTaskCancelRequestedEventAttributes}
	})
	HistoryEvent_RequestCancelActivityTaskFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_RequestCancelActivityTaskFailedEventAttributes{RequestCancelActivityTaskFailedEventAttributes: &RequestCancelActivityTaskFailedEventAttributes}
	})
	HistoryEvent_ActivityTaskCanceled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ActivityTaskCanceledEventAttributes{ActivityTaskCanceledEventAttributes: &ActivityTaskCanceledEventAttributes}
	})
	HistoryEvent_TimerStarted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_TimerStartedEventAttributes{TimerStartedEventAttributes: &TimerStartedEventAttributes}
	})
	HistoryEvent_TimerFired = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_TimerFiredEventAttributes{TimerFiredEventAttributes: &TimerFiredEventAttributes}
	})
	HistoryEvent_CancelTimerFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_CancelTimerFailedEventAttributes{CancelTimerFailedEventAttributes: &CancelTimerFailedEventAttributes}
	})
	HistoryEvent_TimerCanceled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_TimerCanceledEventAttributes{TimerCanceledEventAttributes: &TimerCanceledEventAttributes}
	})
	HistoryEvent_WorkflowExecutionCancelRequested = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes{WorkflowExecutionCancelRequestedEventAttributes: &WorkflowExecutionCancelRequestedEventAttributes}
	})
	HistoryEvent_WorkflowExecutionCanceled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCanceledEventAttributes{WorkflowExecutionCanceledEventAttributes: &WorkflowExecutionCanceledEventAttributes}
	})
	HistoryEvent_RequestCancelExternalWorkflowExecutionInitiated = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{RequestCancelExternalWorkflowExecutionInitiatedEventAttributes: &RequestCancelExternalWorkflowExecutionInitiatedEventAttributes}
	})
	HistoryEvent_RequestCancelExternalWorkflowExecutionFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionFailedEventAttributes{RequestCancelExternalWorkflowExecutionFailedEventAttributes: &RequestCancelExternalWorkflowExecutionFailedEventAttributes}
	})
	HistoryEvent_ExternalWorkflowExecutionCancelRequested = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ExternalWorkflowExecutionCancelRequestedEventAttributes{ExternalWorkflowExecutionCancelRequestedEventAttributes: &ExternalWorkflowExecutionCancelRequestedEventAttributes}
	})
	HistoryEvent_MarkerRecorded = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_MarkerRecordedEventAttributes{MarkerRecordedEventAttributes: &MarkerRecordedEventAttributes}
	})
	HistoryEvent_WorkflowExecutionSignaled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes{WorkflowExecutionSignaledEventAttributes: &WorkflowExecutionSignaledEventAttributes}
	})
	HistoryEvent_WorkflowExecutionTerminated = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionTerminatedEventAttributes{WorkflowExecutionTerminatedEventAttributes: &WorkflowExecutionTerminatedEventAttributes}
	})
	HistoryEvent_WorkflowExecutionContinuedAsNew = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes{WorkflowExecutionContinuedAsNewEventAttributes: &WorkflowExecutionContinuedAsNewEventAttributes}
	})
	HistoryEvent_StartChildWorkflowExecutionInitiated = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes{StartChildWorkflowExecutionInitiatedEventAttributes: &StartChildWorkflowExecutionInitiatedEventAttributes}
	})
	HistoryEvent_StartChildWorkflowExecutionFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_StartChildWorkflowExecutionFailedEventAttributes{StartChildWorkflowExecutionFailedEventAttributes: &StartChildWorkflowExecutionFailedEventAttributes}
	})
	HistoryEvent_ChildWorkflowExecutionStarted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionStartedEventAttributes{ChildWorkflowExecutionStartedEventAttributes: &ChildWorkflowExecutionStartedEventAttributes}
	})
	HistoryEvent_ChildWorkflowExecutionCompleted = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionCompletedEventAttributes{ChildWorkflowExecutionCompletedEventAttributes: &ChildWorkflowExecutionCompletedEventAttributes}
	})
	HistoryEvent_ChildWorkflowExecutionFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionFailedEventAttributes{ChildWorkflowExecutionFailedEventAttributes: &ChildWorkflowExecutionFailedEventAttributes}
	})
	HistoryEvent_ChildWorkflowExecutionCanceled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionCanceledEventAttributes{ChildWorkflowExecutionCanceledEventAttributes: &ChildWorkflowExecutionCanceledEventAttributes}
	})
	HistoryEvent_ChildWorkflowExecutionTimedOut = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionTimedOutEventAttributes{ChildWorkflowExecutionTimedOutEventAttributes: &ChildWorkflowExecutionTimedOutEventAttributes}
	})
	HistoryEvent_ChildWorkflowExecutionTerminated = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionTerminatedEventAttributes{ChildWorkflowExecutionTerminatedEventAttributes: &ChildWorkflowExecutionTerminatedEventAttributes}
	})
	HistoryEvent_SignalExternalWorkflowExecutionInitiated = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_SignalExternalWorkflowExecutionInitiatedEventAttributes{SignalExternalWorkflowExecutionInitiatedEventAttributes: &SignalExternalWorkflowExecutionInitiatedEventAttributes}
	})
	HistoryEvent_SignalExternalWorkflowExecutionFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_SignalExternalWorkflowExecutionFailedEventAttributes{SignalExternalWorkflowExecutionFailedEventAttributes: &SignalExternalWorkflowExecutionFailedEventAttributes}
	})
	HistoryEvent_ExternalWorkflowExecutionSignaled = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_ExternalWorkflowExecutionSignaledEventAttributes{ExternalWorkflowExecutionSignaledEventAttributes: &ExternalWorkflowExecutionSignaledEventAttributes}
	})
	HistoryEvent_UpsertWorkflowSearchAttributes = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes{UpsertWorkflowSearchAttributesEventAttributes: &UpsertWorkflowSearchAttributesEventAttributes}
	})
	HistoryEvent_SemaphoreAcquireInitiated = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_SemaphoreAcquireInitiatedEventAttributes{SemaphoreAcquireInitiatedEventAttributes: &SemaphoreAcquireInitiatedEventAttributes}
	})
	HistoryEvent_SemaphoreAcquired = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_SemaphoreAcquiredEventAttributes{SemaphoreAcquiredEventAttributes: &SemaphoreAcquiredEventAttributes}
	})
	HistoryEvent_SemaphoreReleased = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_SemaphoreReleasedEventAttributes{SemaphoreReleasedEventAttributes: &SemaphoreReleasedEventAttributes}
	})
	HistoryEvent_SemaphoreAcquireFailed = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_SemaphoreAcquireFailedEventAttributes{SemaphoreAcquireFailedEventAttributes: &SemaphoreAcquireFailedEventAttributes}
	})
	HistoryEvent_SemaphoreTokenLost = generateEvent(func(e *apiv1.HistoryEvent) {
		e.Attributes = &apiv1.HistoryEvent_SemaphoreTokenLostEventAttributes{SemaphoreTokenLostEventAttributes: &SemaphoreTokenLostEventAttributes}
	})

	WorkflowExecutionStartedEventAttributes = apiv1.WorkflowExecutionStartedEventAttributes{
		WorkflowType: &WorkflowType,
		ParentExecutionInfo: &apiv1.ParentExecutionInfo{
			DomainName:        DomainName,
			WorkflowExecution: &WorkflowExecution,
			InitiatedId:       EventID1,
		},
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		ContinuedExecutionRunId:      RunID1,
		Initiator:                    ContinueAsNewInitiator,
		ContinuedFailure:             &Failure,
		LastCompletionResult:         &Payload2,
		OriginalExecutionRunId:       RunID2,
		Identity:                     Identity,
		FirstExecutionRunId:          RunID3,
		RetryPolicy:                  &RetryPolicy,
		Attempt:                      Attempt,
		ExpirationTime:               Timestamp1,
		CronSchedule:                 CronSchedule,
		FirstDecisionTaskBackoff:     Duration3,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
		PrevAutoResetPoints:          &ResetPoints,
		Header:                       &Header,
	}
	WorkflowExecutionCompletedEventAttributes = apiv1.WorkflowExecutionCompletedEventAttributes{
		Result:                       &Payload1,
		DecisionTaskCompletedEventId: EventID1,
	}
	WorkflowExecutionFailedEventAttributes = apiv1.WorkflowExecutionFailedEventAttributes{
		Failure:                      &Failure,
		DecisionTaskCompletedEventId: EventID1,
	}
	WorkflowExecutionTimedOutEventAttributes = apiv1.WorkflowExecutionTimedOutEventAttributes{
		TimeoutType: TimeoutType,
	}
	DecisionTaskScheduledEventAttributes = apiv1.DecisionTaskScheduledEventAttributes{
		TaskList:            &TaskList,
		StartToCloseTimeout: Duration1,
		Attempt:             Attempt,
	}
	DecisionTaskStartedEventAttributes = apiv1.DecisionTaskStartedEventAttributes{
		ScheduledEventId: EventID1,
		Identity:         Identity,
		RequestId:        RequestID,
	}
	DecisionTaskCompletedEventAttributes = apiv1.DecisionTaskCompletedEventAttributes{
		ExecutionContext: ExecutionContext,
		ScheduledEventId: EventID1,
		StartedEventId:   EventID2,
		Identity:         Identity,
		BinaryChecksum:   Checksum,
	}
	DecisionTaskTimedOutEventAttributes = apiv1.DecisionTaskTimedOutEventAttributes{
		ScheduledEventId: EventID1,
		StartedEventId:   EventID2,
		TimeoutType:      TimeoutType,
		BaseRunId:        RunID1,
		NewRunId:         RunID2,
		ForkEventVersion: Version1,
		Reason:           Reason,
		Cause:            DecisionTaskTimedOutCause,
	}
	DecisionTaskFailedEventAttributes = apiv1.DecisionTaskFailedEventAttributes{
		ScheduledEventId: EventID1,
		StartedEventId:   EventID2,
		Cause:            DecisionTaskFailedCause,
		Failure:          &Failure,
		Identity:         Identity,
		BaseRunId:        RunID1,
		NewRunId:         RunID2,
		ForkEventVersion: Version1,
		BinaryChecksum:   Checksum,
	}
	ActivityTaskScheduledEventAttributes = apiv1.ActivityTaskScheduledEventAttributes{
		ActivityId:                   ActivityID,
		ActivityType:                 &ActivityType,
		Domain:                       DomainName,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ScheduleToCloseTimeout:       Duration1,
		ScheduleToStartTimeout:       Duration2,
		StartToCloseTimeout:          Duration3,
		HeartbeatTimeout:             Duration4,
		DecisionTaskCompletedEventId: EventID1,
		RetryPolicy:                  &RetryPolicy,
		Header:                       &Header,
	}
	ActivityTaskStartedEventAttributes = apiv1.ActivityTaskStartedEventAttributes{
		ScheduledEventId: EventID1,
		Identity:         Identity,
		RequestId:        RequestID,
		Attempt:          Attempt,
		LastFailure:      &Failure,
	}
	ActivityTaskCompletedEventAttributes = apiv1.ActivityTaskCompletedEventAttributes{
		Result:           &Payload1,
		ScheduledEventId: EventID1,
		StartedEventId:   EventID2,
		Identity:         Identity,
	}
	ActivityTaskFailedEventAttributes = apiv1.ActivityTaskFailedEventAttributes{
		Failure:          &Failure,
		ScheduledEventId: EventID1,
		StartedEventId:   EventID2,
		Identity:         Identity,
	}
	ActivityTaskTimedOutEventAttributes = apiv1.ActivityTaskTimedOutEventAttributes{
		Details:          &Payload1,
		ScheduledEventId: EventID1,
		StartedEventId:   EventID2,
		TimeoutType:      TimeoutType,
		LastFailure:      &Failure,
	}
	TimerStartedEventAttributes = apiv1.TimerStartedEventAttributes{
		TimerId:                      TimerID,
		StartToFireTimeout:           Duration1,
		DecisionTaskCompletedEventId: EventID1,
	}
	TimerFiredEventAttributes = apiv1.TimerFiredEventAttributes{
		TimerId:        TimerID,
		StartedEventId: EventID1,
	}
	ActivityTaskCancelRequestedEventAttributes = apiv1.ActivityTaskCancelRequestedEventAttributes{
		ActivityId:                   ActivityID,
		DecisionTaskCompletedEventId: EventID1,
	}
	RequestCancelActivityTaskFailedEventAttributes = apiv1.RequestCancelActivityTaskFailedEventAttributes{
		ActivityId:                   ActivityID,
		Cause:                        Cause,
		DecisionTaskCompletedEventId: EventID1,
	}
	ActivityTaskCanceledEventAttributes = apiv1.ActivityTaskCanceledEventAttributes{
		Details:                      &Payload1,
		LatestCancelRequestedEventId: EventID1,
		ScheduledEventId:             EventID2,
		StartedEventId:               EventID3,
		Identity:                     Identity,
	}
	TimerCanceledEventAttributes = apiv1.TimerCanceledEventAttributes{
		TimerId:                      TimerID,
		StartedEventId:               EventID1,
		DecisionTaskCompletedEventId: EventID2,
		Identity:                     Identity,
	}
	CancelTimerFailedEventAttributes = apiv1.CancelTimerFailedEventAttributes{
		TimerId:                      TimerID,
		Cause:                        Cause,
		DecisionTaskCompletedEventId: EventID1,
		Identity:                     Identity,
	}
	MarkerRecordedEventAttributes = apiv1.MarkerRecordedEventAttributes{
		MarkerName:                   MarkerName,
		Details:                      &Payload1,
		DecisionTaskCompletedEventId: EventID1,
		Header:                       &Header,
	}
	WorkflowExecutionSignaledEventAttributes = apiv1.WorkflowExecutionSignaledEventAttributes{
		SignalName: SignalName,
		Input:      &Payload1,
		Identity:   Identity,
	}
	WorkflowExecutionTerminatedEventAttributes = apiv1.WorkflowExecutionTerminatedEventAttributes{
		Reason:   Reason,
		Details:  &Payload1,
		Identity: Identity,
	}
	WorkflowExecutionCancelRequestedEventAttributes = apiv1.WorkflowExecutionCancelRequestedEventAttributes{
		Cause: Cause,
		ExternalExecutionInfo: &apiv1.ExternalExecutionInfo{
			WorkflowExecution: &WorkflowExecution,
			InitiatedId:       EventID1,
		},
		Identity: Identity,
	}
	WorkflowExecutionCanceledEventAttributes = apiv1.WorkflowExecutionCanceledEventAttributes{
		DecisionTaskCompletedEventId: EventID1,
		Details:                      &Payload1,
	}
	RequestCancelExternalWorkflowExecutionInitiatedEventAttributes = apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: EventID1,
		Domain:                       DomainName,
		WorkflowExecution:            &WorkflowExecution,
		Control:                      Control,
		ChildWorkflowOnly:            true,
	}
	RequestCancelExternalWorkflowExecutionFailedEventAttributes = apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        CancelExternalWorkflowExecutionFailedCause,
		DecisionTaskCompletedEventId: EventID1,
		Domain:                       DomainName,
		WorkflowExecution:            &WorkflowExecution,
		InitiatedEventId:             EventID2,
		Control:                      Control,
	}
	ExternalWorkflowExecutionCancelRequestedEventAttributes = apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes{
		InitiatedEventId:  EventID1,
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
	}
	WorkflowExecutionContinuedAsNewEventAttributes = apiv1.WorkflowExecutionContinuedAsNewEventAttributes{
		NewExecutionRunId:            RunID1,
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		DecisionTaskCompletedEventId: EventID1,
		BackoffStartInterval:         Duration3,
		Initiator:                    ContinueAsNewInitiator,
		Failure:                      &Failure,
		LastCompletionResult:         &Payload2,
		Header:                       &Header,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
	}
	StartChildWorkflowExecutionInitiatedEventAttributes = apiv1.StartChildWorkflowExecutionInitiatedEventAttributes{
		Domain:                       DomainName,
		WorkflowId:                   WorkflowID,
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		ParentClosePolicy:            ParentClosePolicy,
		Control:                      Control,
		DecisionTaskCompletedEventId: EventID1,
		WorkflowIdReusePolicy:        WorkflowIDReusePolicy,
		RetryPolicy:                  &RetryPolicy,
		CronSchedule:                 CronSchedule,
		Header:                       &Header,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
	}
	StartChildWorkflowExecutionFailedEventAttributes = apiv1.StartChildWorkflowExecutionFailedEventAttributes{
		Domain:                       DomainName,
		WorkflowId:                   WorkflowID,
		WorkflowType:                 &WorkflowType,
		Cause:                        ChildWorkflowExecutionFailedCause,
		Control:                      Control,
		InitiatedEventId:             EventID1,
		DecisionTaskCompletedEventId: EventID2,
	}
	ChildWorkflowExecutionStartedEventAttributes = apiv1.ChildWorkflowExecutionStartedEventAttributes{
		Domain:            DomainName,
		InitiatedEventId:  EventID1,
		WorkflowExecution: &WorkflowExecution,
		WorkflowType:      &WorkflowType,
		Header:            &Header,
	}
	ChildWorkflowExecutionCompletedEventAttributes = apiv1.ChildWorkflowExecutionCompletedEventAttributes{
		Result:            &Payload1,
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		WorkflowType:      &WorkflowType,
		InitiatedEventId:  EventID1,
		StartedEventId:    EventID2,
	}
	ChildWorkflowExecutionFailedEventAttributes = apiv1.ChildWorkflowExecutionFailedEventAttributes{
		Failure:           &Failure,
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		WorkflowType:      &WorkflowType,
		InitiatedEventId:  EventID1,
		StartedEventId:    EventID2,
	}
	ChildWorkflowExecutionCanceledEventAttributes = apiv1.ChildWorkflowExecutionCanceledEventAttributes{
		Details:           &Payload1,
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		WorkflowType:      &WorkflowType,
		InitiatedEventId:  EventID1,
		StartedEventId:    EventID2,
	}
	ChildWorkflowExecutionTimedOutEventAttributes = apiv1.ChildWorkflowExecutionTimedOutEventAttributes{
		TimeoutType:       TimeoutType,
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		WorkflowType:      &WorkflowType,
		InitiatedEventId:  EventID1,
		StartedEventId:    EventID2,
	}
	ChildWorkflowExecutionTerminatedEventAttributes = apiv1.ChildWorkflowExecutionTerminatedEventAttributes{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		WorkflowType:      &WorkflowType,
		InitiatedEventId:  EventID1,
		StartedEventId:    EventID2,
	}
	SignalExternalWorkflowExecutionInitiatedEventAttributes = apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: EventID1,
		Domain:                       DomainName,
		WorkflowExecution:            &WorkflowExecution,
		SignalName:                   SignalName,
		Input:                        &Payload1,
		Control:                      Control,
		ChildWorkflowOnly:            true,
	}
	SignalExternalWorkflowExecutionFailedEventAttributes = apiv1.SignalExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        SignalExternalWorkflowExecutionFailedCause,
		DecisionTaskCompletedEventId: EventID1,
		Domain:                       DomainName,
		WorkflowExecution:            &WorkflowExecution,
		InitiatedEventId:             EventID2,
		Control:                      Control,
	}
	ExternalWorkflowExecutionSignaledEventAttributes = apiv1.ExternalWorkflowExecutionSignaledEventAttributes{
		InitiatedEventId:  EventID1,
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		Control:           Control,
	}
	UpsertWorkflowSearchAttributesEventAttributes = apiv1.UpsertWorkflowSearchAttributesEventAttributes{
		DecisionTaskCompletedEventId: EventID1,
		SearchAttributes:             &SearchAttributes,
	}
	SemaphoreAcquireInitiatedEventAttributes = apiv1.SemaphoreAcquireInitiatedEventAttributes{
		SemaphoreName:                SemaphoreName,
		WaitTimeout:                  Duration1,
		DecisionTaskCompletedEventId: EventID1,
	}
	SemaphoreAcquiredEventAttributes = apiv1.SemaphoreAcquiredEventAttributes{
		TokenId:          1,
		InitiatedEventId: EventID1,
	}
	SemaphoreReleasedEventAttributes = apiv1.SemaphoreReleasedEventAttributes{
		TokenId:                      1,
		InitiatedEventId:             EventID1,
		DecisionTaskCompletedEventId: EventID2,
	}
	SemaphoreAcquireFailedEventAttributes = apiv1.SemaphoreAcquireFailedEventAttributes{
		Cause:            apiv1.SemaphoreAcquireFailedCause_SEMAPHORE_ACQUIRE_FAILED_CAUSE_TIMED_OUT,
		InitiatedEventId: EventID1,
	}
	SemaphoreTokenLostEventAttributes = apiv1.SemaphoreTokenLostEventAttributes{
		TokenId:          1,
		InitiatedEventId: EventID1,
		Cause:            apiv1.SemaphoreTokenLostCause_SEMAPHORE_TOKEN_LOST_CAUSE_RESET,
	}
)

func generateEvent(modifier func(e *apiv1.HistoryEvent)) apiv1.HistoryEvent {
	e := apiv1.HistoryEvent{
		EventId:   EventID1,
		EventTime: Timestamp1,
		Version:   Version1,
		TaskId:    TaskID,
	}
	modifier(&e)
	return e
}
