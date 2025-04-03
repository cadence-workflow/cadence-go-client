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
	"bytes"
	"fmt"
	"reflect"

	s "go.uber.org/cadence/.gen/go/shared"
)

func anyToString(d interface{}) string {
	v := reflect.ValueOf(d)
	switch v.Kind() {
	case reflect.Ptr:
		return anyToString(v.Elem().Interface())
	case reflect.Struct:
		var buf bytes.Buffer
		t := reflect.TypeOf(d)
		buf.WriteString("(")
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if f.Kind() == reflect.Invalid {
				continue
			}
			fieldValue := valueToString(f)
			if len(fieldValue) == 0 {
				continue
			}
			if buf.Len() > 1 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%s:%s", t.Field(i).Name, fieldValue))
		}
		buf.WriteString(")")
		return buf.String()
	default:
		return fmt.Sprint(d)
	}
}

func valueToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Ptr:
		return valueToString(v.Elem())
	case reflect.Struct:
		return anyToString(v.Interface())
	case reflect.Invalid:
		return ""
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return fmt.Sprintf("[%v]", string(v.Bytes()))
		}
		return fmt.Sprintf("[len=%d]", v.Len())
	default:
		return fmt.Sprint(v.Interface())
	}
}

// HistoryEventToString convert HistoryEvent to string
func HistoryEventToString(e *s.HistoryEvent) string {
	data := getHistoryEventData(e)

	return e.GetEventType().String() + ": " + anyToString(data)
}

func getHistoryEventData(e *s.HistoryEvent) interface{} {
	switch e.GetEventType() {
	case s.EventTypeWorkflowExecutionStarted:
		return e.WorkflowExecutionStartedEventAttributes

	case s.EventTypeWorkflowExecutionCompleted:
		return e.WorkflowExecutionCompletedEventAttributes

	case s.EventTypeWorkflowExecutionFailed:
		return e.WorkflowExecutionFailedEventAttributes

	case s.EventTypeWorkflowExecutionTimedOut:
		return e.WorkflowExecutionTimedOutEventAttributes

	case s.EventTypeDecisionTaskScheduled:
		return e.DecisionTaskScheduledEventAttributes

	case s.EventTypeDecisionTaskStarted:
		return e.DecisionTaskStartedEventAttributes

	case s.EventTypeDecisionTaskCompleted:
		return e.DecisionTaskCompletedEventAttributes

	case s.EventTypeDecisionTaskTimedOut:
		return e.DecisionTaskTimedOutEventAttributes

	case s.EventTypeActivityTaskScheduled:
		return e.ActivityTaskScheduledEventAttributes

	case s.EventTypeActivityTaskStarted:
		return e.ActivityTaskStartedEventAttributes

	case s.EventTypeActivityTaskCompleted:
		return e.ActivityTaskCompletedEventAttributes

	case s.EventTypeActivityTaskFailed:
		return e.ActivityTaskFailedEventAttributes

	case s.EventTypeActivityTaskTimedOut:
		return e.ActivityTaskTimedOutEventAttributes

	case s.EventTypeActivityTaskCancelRequested:
		return e.ActivityTaskCancelRequestedEventAttributes

	case s.EventTypeRequestCancelActivityTaskFailed:
		return e.RequestCancelActivityTaskFailedEventAttributes

	case s.EventTypeActivityTaskCanceled:
		return e.ActivityTaskCanceledEventAttributes

	case s.EventTypeTimerStarted:
		return e.TimerStartedEventAttributes

	case s.EventTypeTimerFired:
		return e.TimerFiredEventAttributes

	case s.EventTypeCancelTimerFailed:
		return e.CancelTimerFailedEventAttributes

	case s.EventTypeTimerCanceled:
		return e.TimerCanceledEventAttributes

	case s.EventTypeMarkerRecorded:
		return e.MarkerRecordedEventAttributes

	case s.EventTypeWorkflowExecutionTerminated:
		return e.WorkflowExecutionTerminatedEventAttributes

	default:
		return e
	}
}

// DecisionToString convert Decision to string
func DecisionToString(d *s.Decision) string {
	data := decisionGetData(d)

	return d.GetDecisionType().String() + ": " + anyToString(data)
}

func decisionGetData(d *s.Decision) interface{} {
	switch d.GetDecisionType() {
	case s.DecisionTypeScheduleActivityTask:
		return d.ScheduleActivityTaskDecisionAttributes

	case s.DecisionTypeRequestCancelActivityTask:
		return d.RequestCancelActivityTaskDecisionAttributes

	case s.DecisionTypeStartTimer:
		return d.StartTimerDecisionAttributes

	case s.DecisionTypeCancelTimer:
		return d.CancelTimerDecisionAttributes

	case s.DecisionTypeCompleteWorkflowExecution:
		return d.CompleteWorkflowExecutionDecisionAttributes

	case s.DecisionTypeFailWorkflowExecution:
		return d.FailWorkflowExecutionDecisionAttributes

	case s.DecisionTypeRecordMarker:
		return d.RecordMarkerDecisionAttributes

	default:
		return d
	}
}
