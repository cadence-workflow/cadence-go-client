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

func DecisionArray(t []*apiv1.Decision) []*shared.Decision {
	if t == nil {
		return nil
	}
	v := make([]*shared.Decision, len(t))
	for i := range t {
		v[i] = Decision(t[i])
	}
	return v
}

func Decision(d *apiv1.Decision) *shared.Decision {
	if d == nil {
		return nil
	}
	decision := shared.Decision{}
	switch attr := d.Attributes.(type) {
	case *apiv1.Decision_ScheduleActivityTaskDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeScheduleActivityTask.Ptr()
		a := attr.ScheduleActivityTaskDecisionAttributes
		decision.ScheduleActivityTaskDecisionAttributes = &shared.ScheduleActivityTaskDecisionAttributes{
			ActivityId:                    &a.ActivityId,
			ActivityType:                  ActivityType(a.ActivityType),
			Domain:                        &a.Domain,
			TaskList:                      TaskList(a.TaskList),
			Input:                         Payload(a.Input),
			ScheduleToCloseTimeoutSeconds: durationToSeconds(a.ScheduleToCloseTimeout),
			ScheduleToStartTimeoutSeconds: durationToSeconds(a.ScheduleToStartTimeout),
			StartToCloseTimeoutSeconds:    durationToSeconds(a.StartToCloseTimeout),
			HeartbeatTimeoutSeconds:       durationToSeconds(a.HeartbeatTimeout),
			RetryPolicy:                   RetryPolicy(a.RetryPolicy),
			Header:                        Header(a.Header),
			RequestLocalDispatch:          &a.RequestLocalDispatch,
		}
	case *apiv1.Decision_StartTimerDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeStartTimer.Ptr()
		a := attr.StartTimerDecisionAttributes
		decision.StartTimerDecisionAttributes = &shared.StartTimerDecisionAttributes{
			TimerId:                   &a.TimerId,
			StartToFireTimeoutSeconds: int32To64(durationToSeconds(a.StartToFireTimeout)),
		}
	case *apiv1.Decision_CompleteWorkflowExecutionDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeCompleteWorkflowExecution.Ptr()
		a := attr.CompleteWorkflowExecutionDecisionAttributes
		decision.CompleteWorkflowExecutionDecisionAttributes = &shared.CompleteWorkflowExecutionDecisionAttributes{
			Result: Payload(a.Result),
		}
	case *apiv1.Decision_FailWorkflowExecutionDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeFailWorkflowExecution.Ptr()
		a := attr.FailWorkflowExecutionDecisionAttributes
		decision.FailWorkflowExecutionDecisionAttributes = &shared.FailWorkflowExecutionDecisionAttributes{
			Reason:  FailureReason(a.Failure),
			Details: FailureDetails(a.Failure),
		}
	case *apiv1.Decision_RequestCancelActivityTaskDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeRequestCancelActivityTask.Ptr()
		a := attr.RequestCancelActivityTaskDecisionAttributes
		decision.RequestCancelActivityTaskDecisionAttributes = &shared.RequestCancelActivityTaskDecisionAttributes{
			ActivityId: &a.ActivityId,
		}
	case *apiv1.Decision_CancelTimerDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeCancelTimer.Ptr()
		a := attr.CancelTimerDecisionAttributes
		decision.CancelTimerDecisionAttributes = &shared.CancelTimerDecisionAttributes{
			TimerId: &a.TimerId,
		}
	case *apiv1.Decision_CancelWorkflowExecutionDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeCancelWorkflowExecution.Ptr()
		a := attr.CancelWorkflowExecutionDecisionAttributes
		decision.CancelWorkflowExecutionDecisionAttributes = &shared.CancelWorkflowExecutionDecisionAttributes{
			Details: Payload(a.Details),
		}
	case *apiv1.Decision_RequestCancelExternalWorkflowExecutionDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeRequestCancelExternalWorkflowExecution.Ptr()
		a := attr.RequestCancelExternalWorkflowExecutionDecisionAttributes
		decision.RequestCancelExternalWorkflowExecutionDecisionAttributes = &shared.RequestCancelExternalWorkflowExecutionDecisionAttributes{
			Domain:            &a.Domain,
			WorkflowId:        WorkflowID(a.WorkflowExecution),
			RunId:             RunID(a.WorkflowExecution),
			Control:           a.Control,
			ChildWorkflowOnly: &a.ChildWorkflowOnly,
		}
	case *apiv1.Decision_RecordMarkerDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeRecordMarker.Ptr()
		a := attr.RecordMarkerDecisionAttributes
		decision.RecordMarkerDecisionAttributes = &shared.RecordMarkerDecisionAttributes{
			MarkerName: &a.MarkerName,
			Details:    Payload(a.Details),
			Header:     Header(a.Header),
		}
	case *apiv1.Decision_ContinueAsNewWorkflowExecutionDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeContinueAsNewWorkflowExecution.Ptr()
		a := attr.ContinueAsNewWorkflowExecutionDecisionAttributes
		decision.ContinueAsNewWorkflowExecutionDecisionAttributes = &shared.ContinueAsNewWorkflowExecutionDecisionAttributes{
			WorkflowType:                        WorkflowType(a.WorkflowType),
			TaskList:                            TaskList(a.TaskList),
			Input:                               Payload(a.Input),
			ExecutionStartToCloseTimeoutSeconds: durationToSeconds(a.ExecutionStartToCloseTimeout),
			TaskStartToCloseTimeoutSeconds:      durationToSeconds(a.TaskStartToCloseTimeout),
			BackoffStartIntervalInSeconds:       durationToSeconds(a.BackoffStartInterval),
			RetryPolicy:                         RetryPolicy(a.RetryPolicy),
			Initiator:                           ContinueAsNewInitiator(a.Initiator),
			FailureReason:                       FailureReason(a.Failure),
			FailureDetails:                      FailureDetails(a.Failure),
			LastCompletionResult:                Payload(a.LastCompletionResult),
			CronSchedule:                        &a.CronSchedule,
			Header:                              Header(a.Header),
			Memo:                                Memo(a.Memo),
			SearchAttributes:                    SearchAttributes(a.SearchAttributes),
			JitterStartSeconds:                  durationToSeconds(a.JitterStart),
			CronOverlapPolicy:                   CronOverlapPolicy(a.CronOverlapPolicy),
			ActiveClusterSelectionPolicy:        ActiveClusterSelectionPolicy(a.ActiveClusterSelectionPolicy),
		}
	case *apiv1.Decision_StartChildWorkflowExecutionDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeStartChildWorkflowExecution.Ptr()
		a := attr.StartChildWorkflowExecutionDecisionAttributes
		decision.StartChildWorkflowExecutionDecisionAttributes = &shared.StartChildWorkflowExecutionDecisionAttributes{
			Domain:                              &a.Domain,
			WorkflowId:                          &a.WorkflowId,
			WorkflowType:                        WorkflowType(a.WorkflowType),
			TaskList:                            TaskList(a.TaskList),
			Input:                               Payload(a.Input),
			ExecutionStartToCloseTimeoutSeconds: durationToSeconds(a.ExecutionStartToCloseTimeout),
			TaskStartToCloseTimeoutSeconds:      durationToSeconds(a.TaskStartToCloseTimeout),
			ParentClosePolicy:                   ParentClosePolicy(a.ParentClosePolicy),
			Control:                             a.Control,
			WorkflowIdReusePolicy:               WorkflowIDReusePolicy(a.WorkflowIdReusePolicy),
			RetryPolicy:                         RetryPolicy(a.RetryPolicy),
			CronSchedule:                        &a.CronSchedule,
			Header:                              Header(a.Header),
			Memo:                                Memo(a.Memo),
			SearchAttributes:                    SearchAttributes(a.SearchAttributes),
			CronOverlapPolicy:                   CronOverlapPolicy(a.CronOverlapPolicy),
			ActiveClusterSelectionPolicy:        ActiveClusterSelectionPolicy(a.ActiveClusterSelectionPolicy),
		}
	case *apiv1.Decision_SignalExternalWorkflowExecutionDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeSignalExternalWorkflowExecution.Ptr()
		a := attr.SignalExternalWorkflowExecutionDecisionAttributes
		decision.SignalExternalWorkflowExecutionDecisionAttributes = &shared.SignalExternalWorkflowExecutionDecisionAttributes{
			Domain:            &a.Domain,
			Execution:         WorkflowExecution(a.WorkflowExecution),
			SignalName:        &a.SignalName,
			Input:             Payload(a.Input),
			Control:           a.Control,
			ChildWorkflowOnly: &a.ChildWorkflowOnly,
		}
	case *apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes:
		decision.DecisionType = shared.DecisionTypeUpsertWorkflowSearchAttributes.Ptr()
		a := attr.UpsertWorkflowSearchAttributesDecisionAttributes
		decision.UpsertWorkflowSearchAttributesDecisionAttributes = &shared.UpsertWorkflowSearchAttributesDecisionAttributes{
			SearchAttributes: SearchAttributes(a.SearchAttributes),
		}
	default:
		panic("unknown decision type")
	}
	return &decision
}
