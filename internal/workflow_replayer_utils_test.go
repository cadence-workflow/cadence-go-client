// Copyright (c) 2017-2021 Uber Technologies Inc.
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

package internal

import (
	"testing"

	"go.uber.org/cadence/internal/compatibility/testdata"

	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

type matchReplayWithHistoryTestCase struct {
	name            string
	replayDecisions []*shared.Decision
	historyEvents   []*shared.HistoryEvent
	expectError     bool
}

func TestMatchReplayWithHistory(t *testing.T) {
	testCases := []matchReplayWithHistoryTestCase{
		{
			name: "Success matching",
			replayDecisions: []*shared.Decision{
				mockDecision(shared.DecisionTypeScheduleActivityTask),
			},
			historyEvents: []*shared.HistoryEvent{
				mockHistoryEvent(shared.EventTypeActivityTaskScheduled),
			},
			expectError: false,
		},
		{
			name: "Mismatch error",
			replayDecisions: []*shared.Decision{
				mockDecision(shared.DecisionTypeStartTimer),
			},
			historyEvents: []*shared.HistoryEvent{
				mockHistoryEvent(shared.EventTypeTimerCanceled),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := &WorkflowInfo{} // Assuming WorkflowInfo is a struct you have defined
			err := matchReplayWithHistory(info, tc.replayDecisions, tc.historyEvents)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsDecisionMatchEvent(t *testing.T) {
	tests := []struct {
		name       string
		decision   *shared.Decision
		event      *shared.HistoryEvent
		expected   bool
		strictMode bool
	}{
		{
			name:     "case DecisionTypeScheduleActivityTask - event type not equal EventTypeActivityTaskScheduled",
			decision: &shared.Decision{DecisionType: common.DecisionTypePtr(shared.DecisionTypeScheduleActivityTask)},
		},
		{
			name:     "case DecisionTypeRequestCancelActivityTask - event type not equal EventTypeActivityTaskCancelRequested",
			decision: &shared.Decision{DecisionType: common.DecisionTypePtr(shared.DecisionTypeRequestCancelActivityTask)},
		},
		{
			name: "case DecisionTypeRequestCancelActivityTask - different activityIDs",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeRequestCancelActivityTask),
				RequestCancelActivityTaskDecisionAttributes: &shared.RequestCancelActivityTaskDecisionAttributes{ActivityId: common.StringPtr("ActivityID-1")},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeActivityTaskCancelRequested),
				ActivityTaskCancelRequestedEventAttributes: &shared.ActivityTaskCancelRequestedEventAttributes{ActivityId: common.StringPtr("ActivityID-2")},
			},
		},
		{
			name: "case DecisionTypeRequestCancelActivityTask - equal activityIDs",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeRequestCancelActivityTask),
				RequestCancelActivityTaskDecisionAttributes: &shared.RequestCancelActivityTaskDecisionAttributes{ActivityId: common.StringPtr(testdata.ActivityID)},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeActivityTaskCancelRequested),
				ActivityTaskCancelRequestedEventAttributes: &shared.ActivityTaskCancelRequestedEventAttributes{ActivityId: common.StringPtr(testdata.ActivityID)},
			},
			expected: true,
		},
		{
			name: "case DecisionTypeStartTimer - event type != EventTypeTimerStarted",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeStartTimer),
			},
			expected: false,
		},
		{
			name: "case DecisionTypeStartTimer - different timerIDs",
			decision: &shared.Decision{
				DecisionType:                 common.DecisionTypePtr(shared.DecisionTypeStartTimer),
				StartTimerDecisionAttributes: &shared.StartTimerDecisionAttributes{TimerId: common.StringPtr("timerID-1")},
			},
			event: &shared.HistoryEvent{
				EventType:                   common.EventTypePtr(shared.EventTypeTimerStarted),
				TimerStartedEventAttributes: &shared.TimerStartedEventAttributes{TimerId: common.StringPtr("timerID-2")},
			},
		},
		{
			name: "case DecisionTypeStartTimer - equal timerIDs",
			decision: &shared.Decision{
				DecisionType:                 common.DecisionTypePtr(shared.DecisionTypeStartTimer),
				StartTimerDecisionAttributes: &shared.StartTimerDecisionAttributes{TimerId: common.StringPtr("timerID")},
			},
			event: &shared.HistoryEvent{
				EventType:                   common.EventTypePtr(shared.EventTypeTimerStarted),
				TimerStartedEventAttributes: &shared.TimerStartedEventAttributes{TimerId: common.StringPtr("timerID")},
			},
			expected: true,
		},
		{
			name: "case DecisionTypeCancelTimer - event not of type EventTypeTimerCanceled",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeCancelTimer),
			},
		},
		{
			name: "case DecisionTypeCancelTimer - event type EventTypeTimerCanceled with different timer IDs",
			decision: &shared.Decision{
				DecisionType:                  common.DecisionTypePtr(shared.DecisionTypeCancelTimer),
				CancelTimerDecisionAttributes: &shared.CancelTimerDecisionAttributes{TimerId: common.StringPtr("timerID-1")},
			},
			event: &shared.HistoryEvent{
				EventType:                    common.EventTypePtr(shared.EventTypeTimerCanceled),
				TimerCanceledEventAttributes: &shared.TimerCanceledEventAttributes{TimerId: common.StringPtr("timerID-2")},
			},
		},
		{
			name: "case DecisionTypeCancelTimer - event type EventTypeCancelTimerFailed with different timer IDs",
			decision: &shared.Decision{
				DecisionType:                  common.DecisionTypePtr(shared.DecisionTypeCancelTimer),
				CancelTimerDecisionAttributes: &shared.CancelTimerDecisionAttributes{TimerId: common.StringPtr("timerID-1")},
			},
			event: &shared.HistoryEvent{
				EventType:                        common.EventTypePtr(shared.EventTypeCancelTimerFailed),
				CancelTimerFailedEventAttributes: &shared.CancelTimerFailedEventAttributes{TimerId: common.StringPtr("timerID-2")},
			},
		},
		{
			name: "case DecisionTypeCancelTimer - event type EventTypeCancelTimerFailed with equal timer IDs",
			decision: &shared.Decision{
				DecisionType:                  common.DecisionTypePtr(shared.DecisionTypeCancelTimer),
				CancelTimerDecisionAttributes: &shared.CancelTimerDecisionAttributes{TimerId: common.StringPtr("timerID")},
			},
			event: &shared.HistoryEvent{
				EventType:                        common.EventTypePtr(shared.EventTypeCancelTimerFailed),
				CancelTimerFailedEventAttributes: &shared.CancelTimerFailedEventAttributes{TimerId: common.StringPtr("timerID")},
			},
			expected: true,
		},
		{
			name: "case DecisionTypeCompleteWorkflowExecution - match event",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeCompleteWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionCompleted),
			},
			expected: true,
		},
		{
			name: "case DecisionTypeCompleteWorkflowExecution - event not of type EventTypeWorkflowExecutionCompleted",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeCompleteWorkflowExecution),
			},
		},
		{
			name: "case DecisionTypeCompleteWorkflowExecution - event result not equal decision result",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeCompleteWorkflowExecution),
				CompleteWorkflowExecutionDecisionAttributes: &shared.CompleteWorkflowExecutionDecisionAttributes{Result: []byte("some-result-1")},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionCompleted),
				WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{Result: []byte("some-result-2")},
			},
			strictMode: true,
		},
		{
			name: "case DecisionTypeFailWorkflowExecution - match event",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeFailWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionFailed),
			},
			expected: true,
		},
		{
			name: "case DecisionTypeFailWorkflowExecution - event not of type EventTypeWorkflowExecutionFailed",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeFailWorkflowExecution),
			},
		},
		{
			name: "case DecisionTypeFailWorkflowExecution - event reason not equal decision reason",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeFailWorkflowExecution),
				FailWorkflowExecutionDecisionAttributes: &shared.FailWorkflowExecutionDecisionAttributes{
					Reason:  common.StringPtr("some-reason-1"),
					Details: []byte("some-details-1"),
				},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionFailed),
				WorkflowExecutionFailedEventAttributes: &shared.WorkflowExecutionFailedEventAttributes{
					Reason:  common.StringPtr("some-reason-2"),
					Details: []byte("some-details-2"),
				},
			},
			strictMode: true,
		},
		{
			name: "case DecisionTypeRecordMarker - equal markerNames",
			decision: &shared.Decision{
				DecisionType:                   common.DecisionTypePtr(shared.DecisionTypeRecordMarker),
				RecordMarkerDecisionAttributes: &shared.RecordMarkerDecisionAttributes{MarkerName: common.StringPtr("some-name")},
			},
			event: &shared.HistoryEvent{
				EventType:                     common.EventTypePtr(shared.EventTypeMarkerRecorded),
				MarkerRecordedEventAttributes: &shared.MarkerRecordedEventAttributes{MarkerName: common.StringPtr("some-name")},
			},
			expected: true,
		},
		{
			name: "case DecisionTypeRecordMarker - different markerNames",
			decision: &shared.Decision{
				DecisionType:                   common.DecisionTypePtr(shared.DecisionTypeRecordMarker),
				RecordMarkerDecisionAttributes: &shared.RecordMarkerDecisionAttributes{MarkerName: common.StringPtr("some-name-1")},
			},
			event: &shared.HistoryEvent{
				EventType:                     common.EventTypePtr(shared.EventTypeMarkerRecorded),
				MarkerRecordedEventAttributes: &shared.MarkerRecordedEventAttributes{MarkerName: common.StringPtr("some-name-2")},
			},
			expected: false,
		},
		{
			name: "case DecisionTypeRecordMarker - event not of type EventTypeMarkerRecorded",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeRecordMarker),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventType(-24)), // non existent event type
			},
			expected: false,
		},
		{
			name: "case DecisionTypeRequestCancelExternalWorkflowExecution - event not of type EventTypeRequestCancelExternalWorkflowExecutionInitiated",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeRequestCancelExternalWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventType(-24)), // non existent event type
			},
		},
		{
			name: "case DecisionTypeRequestCancelExternalWorkflowExecution - different workflowIDs",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeRequestCancelExternalWorkflowExecution),
				RequestCancelExternalWorkflowExecutionDecisionAttributes: &shared.RequestCancelExternalWorkflowExecutionDecisionAttributes{
					WorkflowId: common.StringPtr("some-worklfow-id-1"),
				},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated),
				RequestCancelExternalWorkflowExecutionInitiatedEventAttributes: &shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
					WorkflowExecution: &shared.WorkflowExecution{WorkflowId: common.StringPtr("some-workflow-id-2")},
				},
			},
		},
		{
			name: "case DecisionTypeRequestCancelExternalWorkflowExecution - equal workflowIDs",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeRequestCancelExternalWorkflowExecution),
				RequestCancelExternalWorkflowExecutionDecisionAttributes: &shared.RequestCancelExternalWorkflowExecutionDecisionAttributes{
					WorkflowId: common.StringPtr("some-workflow-id"),
				},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated),
				RequestCancelExternalWorkflowExecutionInitiatedEventAttributes: &shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
					WorkflowExecution: &shared.WorkflowExecution{WorkflowId: common.StringPtr("some-workflow-id")},
				},
			},
			expected: true,
		},
		{
			name: "case DecisionTypeSignalExternalWorkflowExecution - event not of type EventTypeSignalExternalWorkflowExecutionInitiated",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeSignalExternalWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventType(-24)), // non existent event type
			},
		},
		{
			name: "case DecisionTypeSignalExternalWorkflowExecution - different workflowIDs",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeSignalExternalWorkflowExecution),
				SignalExternalWorkflowExecutionDecisionAttributes: &shared.SignalExternalWorkflowExecutionDecisionAttributes{
					Execution: &shared.WorkflowExecution{WorkflowId: common.StringPtr("some-workflow-id-1")},
				},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeSignalExternalWorkflowExecutionInitiated),
				SignalExternalWorkflowExecutionInitiatedEventAttributes: &shared.SignalExternalWorkflowExecutionInitiatedEventAttributes{
					WorkflowExecution: &shared.WorkflowExecution{WorkflowId: common.StringPtr("some-workflow-id-2")},
				},
			},
		},
		{
			name: "case DecisionTypeCancelWorkflowExecution - event not of type EventTypeSignalExternalWorkflowExecutionInitiated",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeCancelWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventType(-24)), // non existent event type
			},
		},
		{
			name: "case DecisionTypeCancelWorkflowExecution - different attribute details",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeCancelWorkflowExecution),
				CancelWorkflowExecutionDecisionAttributes: &shared.CancelWorkflowExecutionDecisionAttributes{Details: []byte("some-details-1")},
			},
			event: &shared.HistoryEvent{
				EventType:                                common.EventTypePtr(shared.EventTypeWorkflowExecutionCanceled),
				WorkflowExecutionCanceledEventAttributes: &shared.WorkflowExecutionCanceledEventAttributes{Details: []byte("some-details-2")},
			},
			strictMode: true,
		},
		{
			name: "case DecisionTypeCancelWorkflowExecution - equal attribute details",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeCancelWorkflowExecution),
				CancelWorkflowExecutionDecisionAttributes: &shared.CancelWorkflowExecutionDecisionAttributes{Details: []byte("some-details")},
			},
			event: &shared.HistoryEvent{
				EventType:                                common.EventTypePtr(shared.EventTypeWorkflowExecutionCanceled),
				WorkflowExecutionCanceledEventAttributes: &shared.WorkflowExecutionCanceledEventAttributes{Details: []byte("some-details")},
			},
			strictMode: true,
			expected:   true,
		},
		{
			name: "case DecisionTypeContinueAsNewWorkflowExecution - event not of type EventTypeWorkflowExecutionContinuedAsNew",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeContinueAsNewWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventType(-24)), // non existent event type
			},
		},
		{
			name: "case DecisionTypeContinueAsNewWorkflowExecution - match event",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeContinueAsNewWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeWorkflowExecutionContinuedAsNew),
			},
			expected: true,
		},
		{
			name: "case DecisionTypeStartChildWorkflowExecution - event not of type EventTypeStartChildWorkflowExecutionInitiated",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeStartChildWorkflowExecution),
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventType(-24)), // non existent event type
			},
		},
		{
			name: "case DecisionTypeStartChildWorkflowExecution - different taskList names",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionTypeStartChildWorkflowExecution),
				StartChildWorkflowExecutionDecisionAttributes: &shared.StartChildWorkflowExecutionDecisionAttributes{
					TaskList: &shared.TaskList{Name: common.StringPtr("some-tasklist-1")},
				},
			},
			event: &shared.HistoryEvent{
				EventType: common.EventTypePtr(shared.EventTypeStartChildWorkflowExecutionInitiated),
				StartChildWorkflowExecutionInitiatedEventAttributes: &shared.StartChildWorkflowExecutionInitiatedEventAttributes{
					TaskList: &shared.TaskList{Name: common.StringPtr("some-tasklist-2")},
				},
			},
			strictMode: true,
		},
		{
			name: "default switch case",
			decision: &shared.Decision{
				DecisionType: common.DecisionTypePtr(shared.DecisionType(-23)), // some random decision type
			},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.decision == nil {
				test.decision = &shared.Decision{}
			}
			if test.event == nil {
				test.event = &shared.HistoryEvent{}
			}
			ok := isDecisionMatchEvent(test.decision, test.event, test.strictMode)
			assert.Equal(t, test.expected, ok)
		})
	}
}

func mockHistoryEvent(eventType shared.EventType) *shared.HistoryEvent {
	event := &shared.HistoryEvent{
		EventType: &eventType,
	}

	switch eventType {
	case shared.EventTypeActivityTaskScheduled:
		event.ActivityTaskScheduledEventAttributes = &shared.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("mockActivityId"),
			ActivityType: &shared.ActivityType{Name: common.StringPtr("mockActivityType")},
			TaskList:     &shared.TaskList{Name: common.StringPtr("mockTaskList")},
			Input:        []byte("mockInput"),
		}
	case shared.EventTypeTimerCanceled: // Newly added case
		event.TimerCanceledEventAttributes = &shared.TimerCanceledEventAttributes{
			TimerId: common.StringPtr("mockTimerId"),
		}
	}

	return event
}

func mockDecision(decisionType shared.DecisionType) *shared.Decision {
	decision := &shared.Decision{
		DecisionType: &decisionType,
	}

	switch decisionType {
	case shared.DecisionTypeScheduleActivityTask:
		decision.ScheduleActivityTaskDecisionAttributes = &shared.ScheduleActivityTaskDecisionAttributes{
			ActivityId:   common.StringPtr("mockActivityId"),
			ActivityType: &shared.ActivityType{Name: common.StringPtr("mockActivityType")},
			TaskList:     &shared.TaskList{Name: common.StringPtr("mockTaskList")},
			Input:        []byte("mockInput"),
		}
	case shared.DecisionTypeStartTimer: // Newly added case
		decision.StartTimerDecisionAttributes = &shared.StartTimerDecisionAttributes{
			TimerId:                   common.StringPtr("mockTimerId"),
			StartToFireTimeoutSeconds: common.Int64Ptr(60), // Example value
		}
	}

	return decision
}
