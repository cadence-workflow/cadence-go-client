// Copyright (c) 2017 Uber Technologies, Inc.
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

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/thriftrw/ptr"

	s "go.uber.org/cadence/.gen/go/shared"
)

func Test_anyToString(t *testing.T) {
	testCases := []struct {
		name             string
		thingToSerialize interface{}
		expected         string
	}{
		{
			name:             "nil",
			thingToSerialize: nil,
			expected:         "<nil>",
		},
		{
			name:             "int",
			thingToSerialize: 1,
			expected:         "1",
		},
		{
			name:             "string",
			thingToSerialize: "test",
			expected:         "test",
		},
		{
			name: "struct",
			thingToSerialize: struct {
				A string
				B int
				C bool
			}{A: "test", B: 1, C: true},
			expected: "(A:test, B:1, C:true)",
		},
		{
			name: "pointer",
			thingToSerialize: &struct {
				A string
				B int
				C bool
			}{A: "test", B: 1, C: true},
			expected: "(A:test, B:1, C:true)",
		},
		{
			name:             "slice",
			thingToSerialize: []int{1, 2, 3},
			expected:         "[1 2 3]",
		},
		{
			name:             "struct with slice",
			thingToSerialize: struct{ A []int }{A: []int{1, 2, 3}},
			expected:         "(A:[len=3])",
		},
		{
			name:             "struct with blob",
			thingToSerialize: struct{ A []byte }{A: []byte("blob-data")},
			expected:         "(A:[blob-data])",
		},
		{
			name:             "struct with struct",
			thingToSerialize: struct{ A struct{ B int } }{A: struct{ B int }{B: 1}},
			expected:         "(A:(B:1))",
		},
		{
			name:             "struct with pointer",
			thingToSerialize: struct{ A *int }{A: new(int)},
			expected:         "(A:0)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := anyToString(tc.thingToSerialize)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func Test_byteSliceToString(t *testing.T) {
	data := []byte("blob-data")
	v := reflect.ValueOf(data)
	strVal := valueToString(v)

	require.Equal(t, "[blob-data]", strVal)

	intBlob := []int32{1, 2, 3}
	v2 := reflect.ValueOf(intBlob)
	strVal2 := valueToString(v2)

	require.Equal(t, "[len=3]", strVal2)
}

func TestHistoryEventToString(t *testing.T) {
	event := &s.HistoryEvent{
		EventType: toPtr(s.EventTypeWorkflowExecutionStarted),
		WorkflowExecutionStartedEventAttributes: &s.WorkflowExecutionStartedEventAttributes{
			WorkflowType: &s.WorkflowType{
				Name: toPtr("test-workflow"),
			},
			TaskList: &s.TaskList{
				Name: toPtr("task-list"),
			},
			ExecutionStartToCloseTimeoutSeconds: ptr.Int32(60),
		},
	}

	strVal := HistoryEventToString(event)

	expected := `WorkflowExecutionStarted: (WorkflowType:(Name:test-workflow), TaskList:(Name:task-list), Input:[], ExecutionStartToCloseTimeoutSeconds:60, ContinuedFailureDetails:[], LastCompletionResult:[], PartitionConfig:map[])`

	assert.Equal(t, expected, strVal)
}

// This just tests that we pick the right attributes to return
// the other attributes will be nil
func Test_getHistoryEventData(t *testing.T) {
	cases := []struct {
		event    *s.HistoryEvent
		expected interface{}
	}{
		{
			event: &s.HistoryEvent{
				EventType:                               toPtr(s.EventTypeWorkflowExecutionStarted),
				WorkflowExecutionStartedEventAttributes: &s.WorkflowExecutionStartedEventAttributes{},
			},
			expected: &s.WorkflowExecutionStartedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType: toPtr(s.EventTypeWorkflowExecutionCompleted),
				WorkflowExecutionCompletedEventAttributes: &s.WorkflowExecutionCompletedEventAttributes{},
			},
			expected: &s.WorkflowExecutionCompletedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                              toPtr(s.EventTypeWorkflowExecutionFailed),
				WorkflowExecutionFailedEventAttributes: &s.WorkflowExecutionFailedEventAttributes{},
			},
			expected: &s.WorkflowExecutionFailedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                                toPtr(s.EventTypeWorkflowExecutionTimedOut),
				WorkflowExecutionTimedOutEventAttributes: &s.WorkflowExecutionTimedOutEventAttributes{},
			},
			expected: &s.WorkflowExecutionTimedOutEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                            toPtr(s.EventTypeDecisionTaskScheduled),
				DecisionTaskScheduledEventAttributes: &s.DecisionTaskScheduledEventAttributes{},
			},
			expected: &s.DecisionTaskScheduledEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                          toPtr(s.EventTypeDecisionTaskStarted),
				DecisionTaskStartedEventAttributes: &s.DecisionTaskStartedEventAttributes{},
			},
			expected: &s.DecisionTaskStartedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                            toPtr(s.EventTypeDecisionTaskCompleted),
				DecisionTaskCompletedEventAttributes: &s.DecisionTaskCompletedEventAttributes{},
			},
			expected: &s.DecisionTaskCompletedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                           toPtr(s.EventTypeDecisionTaskTimedOut),
				DecisionTaskTimedOutEventAttributes: &s.DecisionTaskTimedOutEventAttributes{},
			},
			expected: &s.DecisionTaskTimedOutEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                            toPtr(s.EventTypeActivityTaskScheduled),
				ActivityTaskScheduledEventAttributes: &s.ActivityTaskScheduledEventAttributes{},
			},
			expected: &s.ActivityTaskScheduledEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                          toPtr(s.EventTypeActivityTaskStarted),
				ActivityTaskStartedEventAttributes: &s.ActivityTaskStartedEventAttributes{},
			},
			expected: &s.ActivityTaskStartedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                            toPtr(s.EventTypeActivityTaskCompleted),
				ActivityTaskCompletedEventAttributes: &s.ActivityTaskCompletedEventAttributes{},
			},
			expected: &s.ActivityTaskCompletedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                         toPtr(s.EventTypeActivityTaskFailed),
				ActivityTaskFailedEventAttributes: &s.ActivityTaskFailedEventAttributes{},
			},
			expected: &s.ActivityTaskFailedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                           toPtr(s.EventTypeActivityTaskTimedOut),
				ActivityTaskTimedOutEventAttributes: &s.ActivityTaskTimedOutEventAttributes{},
			},
			expected: &s.ActivityTaskTimedOutEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType: toPtr(s.EventTypeActivityTaskCancelRequested),
				ActivityTaskCancelRequestedEventAttributes: &s.ActivityTaskCancelRequestedEventAttributes{},
			},
			expected: &s.ActivityTaskCancelRequestedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType: toPtr(s.EventTypeRequestCancelActivityTaskFailed),
				RequestCancelActivityTaskFailedEventAttributes: &s.RequestCancelActivityTaskFailedEventAttributes{},
			},
			expected: &s.RequestCancelActivityTaskFailedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                           toPtr(s.EventTypeActivityTaskCanceled),
				ActivityTaskCanceledEventAttributes: &s.ActivityTaskCanceledEventAttributes{},
			},
			expected: &s.ActivityTaskCanceledEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                   toPtr(s.EventTypeTimerStarted),
				TimerStartedEventAttributes: &s.TimerStartedEventAttributes{},
			},
			expected: &s.TimerStartedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                 toPtr(s.EventTypeTimerFired),
				TimerFiredEventAttributes: &s.TimerFiredEventAttributes{},
			},
			expected: &s.TimerFiredEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                        toPtr(s.EventTypeCancelTimerFailed),
				CancelTimerFailedEventAttributes: &s.CancelTimerFailedEventAttributes{},
			},
			expected: &s.CancelTimerFailedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                    toPtr(s.EventTypeTimerCanceled),
				TimerCanceledEventAttributes: &s.TimerCanceledEventAttributes{},
			},
			expected: &s.TimerCanceledEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType:                     toPtr(s.EventTypeMarkerRecorded),
				MarkerRecordedEventAttributes: &s.MarkerRecordedEventAttributes{},
			},
			expected: &s.MarkerRecordedEventAttributes{},
		},
		{
			event: &s.HistoryEvent{
				EventType: toPtr(s.EventTypeWorkflowExecutionTerminated),
				WorkflowExecutionTerminatedEventAttributes: &s.WorkflowExecutionTerminatedEventAttributes{},
			},
			expected: &s.WorkflowExecutionTerminatedEventAttributes{},
		},
		// In the default case, we should return the event itself
		{
			event: &s.HistoryEvent{
				EventType: toPtr(s.EventType(123456789)),
			},
			expected: &s.HistoryEvent{
				EventType: toPtr(s.EventType(123456789)),
			},
		},
	}

	for _, tc := range cases {
		name, err := tc.event.GetEventType().MarshalText()
		require.NoError(t, err)
		t.Run(string(name), func(t *testing.T) {
			require.Equal(t, tc.expected, getHistoryEventData(tc.event))
		})
	}
}

func TestDecisionToString(t *testing.T) {
	decision := &s.Decision{
		DecisionType: toPtr(s.DecisionTypeScheduleActivityTask),
		ScheduleActivityTaskDecisionAttributes: &s.ScheduleActivityTaskDecisionAttributes{
			ActivityId:   toPtr("activity-id"),
			ActivityType: &s.ActivityType{Name: toPtr("activity-type")},
		},
	}

	strVal := DecisionToString(decision)
	expected := "ScheduleActivityTask: (ActivityId:activity-id, ActivityType:(Name:activity-type), Input:[])"
	require.Equal(t, expected, strVal)
}

// This just tests that we pick the right attributes to return
// the other attributes will be nil
func Test_decisionGetData(t *testing.T) {
	cases := []struct {
		decision *s.Decision
		expected interface{}
	}{
		{
			decision: &s.Decision{
				DecisionType:                           toPtr(s.DecisionTypeScheduleActivityTask),
				ScheduleActivityTaskDecisionAttributes: &s.ScheduleActivityTaskDecisionAttributes{},
			},
			expected: &s.ScheduleActivityTaskDecisionAttributes{},
		},
		{
			decision: &s.Decision{
				DecisionType: toPtr(s.DecisionTypeRequestCancelActivityTask),
				RequestCancelActivityTaskDecisionAttributes: &s.RequestCancelActivityTaskDecisionAttributes{},
			},
			expected: &s.RequestCancelActivityTaskDecisionAttributes{},
		},
		{
			decision: &s.Decision{
				DecisionType:                 toPtr(s.DecisionTypeStartTimer),
				StartTimerDecisionAttributes: &s.StartTimerDecisionAttributes{},
			},
			expected: &s.StartTimerDecisionAttributes{},
		},
		{
			decision: &s.Decision{
				DecisionType:                  toPtr(s.DecisionTypeCancelTimer),
				CancelTimerDecisionAttributes: &s.CancelTimerDecisionAttributes{},
			},
			expected: &s.CancelTimerDecisionAttributes{},
		},
		{
			decision: &s.Decision{
				DecisionType: toPtr(s.DecisionTypeCompleteWorkflowExecution),
				CompleteWorkflowExecutionDecisionAttributes: &s.CompleteWorkflowExecutionDecisionAttributes{},
			},
			expected: &s.CompleteWorkflowExecutionDecisionAttributes{},
		},
		{
			decision: &s.Decision{
				DecisionType:                            toPtr(s.DecisionTypeFailWorkflowExecution),
				FailWorkflowExecutionDecisionAttributes: &s.FailWorkflowExecutionDecisionAttributes{},
			},
			expected: &s.FailWorkflowExecutionDecisionAttributes{},
		},
		{
			decision: &s.Decision{
				DecisionType:                   toPtr(s.DecisionTypeRecordMarker),
				RecordMarkerDecisionAttributes: &s.RecordMarkerDecisionAttributes{},
			},
			expected: &s.RecordMarkerDecisionAttributes{},
		},
		// In the default case, we should return the decision itself
		{
			decision: &s.Decision{
				DecisionType: toPtr(s.DecisionType(123456789)),
			},
			expected: &s.Decision{
				DecisionType: toPtr(s.DecisionType(123456789)),
			},
		},
	}

	for _, tc := range cases {
		name, err := tc.decision.GetDecisionType().MarshalText()
		require.NoError(t, err)
		t.Run(string(name), func(t *testing.T) {
			require.Equal(t, tc.expected, decisionGetData(tc.decision))
		})
	}
}

func toPtr[v any](x v) *v {
	return &x
}
