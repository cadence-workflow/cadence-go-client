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

func DecisionArray(t []*shared.Decision) []*apiv1.Decision {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.Decision, len(t))
	for i := range t {
		v[i] = Decision(t[i])
	}
	return v
}

func Decision(d *shared.Decision) *apiv1.Decision {
	if d == nil {
		return nil
	}
	decision := apiv1.Decision{}
	switch *d.DecisionType {
	case shared.DecisionTypeScheduleActivityTask:
		attr := d.ScheduleActivityTaskDecisionAttributes
		decision.Attributes = &apiv1.Decision_ScheduleActivityTaskDecisionAttributes{
			ScheduleActivityTaskDecisionAttributes: &apiv1.ScheduleActivityTaskDecisionAttributes{
				ActivityId:             attr.GetActivityId(),
				ActivityType:           ActivityType(attr.ActivityType),
				Domain:                 attr.GetDomain(),
				TaskList:               TaskList(attr.TaskList),
				Input:                  Payload(attr.Input),
				ScheduleToCloseTimeout: secondsToDuration(attr.ScheduleToCloseTimeoutSeconds),
				ScheduleToStartTimeout: secondsToDuration(attr.ScheduleToStartTimeoutSeconds),
				StartToCloseTimeout:    secondsToDuration(attr.StartToCloseTimeoutSeconds),
				HeartbeatTimeout:       secondsToDuration(attr.HeartbeatTimeoutSeconds),
				RetryPolicy:            RetryPolicy(attr.RetryPolicy),
				Header:                 Header(attr.Header),
				RequestLocalDispatch:   attr.GetRequestLocalDispatch(),
			},
		}
	case shared.DecisionTypeRequestCancelActivityTask:
		attr := d.RequestCancelActivityTaskDecisionAttributes
		decision.Attributes = &apiv1.Decision_RequestCancelActivityTaskDecisionAttributes{
			RequestCancelActivityTaskDecisionAttributes: &apiv1.RequestCancelActivityTaskDecisionAttributes{
				ActivityId: attr.GetActivityId(),
			},
		}
	case shared.DecisionTypeStartTimer:
		attr := d.StartTimerDecisionAttributes
		decision.Attributes = &apiv1.Decision_StartTimerDecisionAttributes{
			StartTimerDecisionAttributes: &apiv1.StartTimerDecisionAttributes{
				TimerId:            attr.GetTimerId(),
				StartToFireTimeout: secondsToDuration(int64To32(attr.StartToFireTimeoutSeconds)),
			},
		}
	case shared.DecisionTypeCompleteWorkflowExecution:
		attr := d.CompleteWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_CompleteWorkflowExecutionDecisionAttributes{
			CompleteWorkflowExecutionDecisionAttributes: &apiv1.CompleteWorkflowExecutionDecisionAttributes{
				Result: Payload(attr.Result),
			},
		}
	case shared.DecisionTypeFailWorkflowExecution:
		attr := d.FailWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_FailWorkflowExecutionDecisionAttributes{
			FailWorkflowExecutionDecisionAttributes: &apiv1.FailWorkflowExecutionDecisionAttributes{
				Failure: Failure(attr.Reason, attr.Details),
			},
		}
	case shared.DecisionTypeCancelTimer:
		attr := d.CancelTimerDecisionAttributes
		decision.Attributes = &apiv1.Decision_CancelTimerDecisionAttributes{
			CancelTimerDecisionAttributes: &apiv1.CancelTimerDecisionAttributes{
				TimerId: attr.GetTimerId(),
			},
		}
	case shared.DecisionTypeCancelWorkflowExecution:
		attr := d.CancelWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_CancelWorkflowExecutionDecisionAttributes{
			CancelWorkflowExecutionDecisionAttributes: &apiv1.CancelWorkflowExecutionDecisionAttributes{
				Details: Payload(attr.Details),
			},
		}
	case shared.DecisionTypeRequestCancelExternalWorkflowExecution:
		attr := d.RequestCancelExternalWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_RequestCancelExternalWorkflowExecutionDecisionAttributes{
			RequestCancelExternalWorkflowExecutionDecisionAttributes: &apiv1.RequestCancelExternalWorkflowExecutionDecisionAttributes{
				Domain:            attr.GetDomain(),
				WorkflowExecution: WorkflowRunPair(attr.GetWorkflowId(), attr.GetRunId()),
				Control:           attr.Control,
				ChildWorkflowOnly: attr.GetChildWorkflowOnly(),
			},
		}
	case shared.DecisionTypeRecordMarker:
		attr := d.RecordMarkerDecisionAttributes
		decision.Attributes = &apiv1.Decision_RecordMarkerDecisionAttributes{
			RecordMarkerDecisionAttributes: &apiv1.RecordMarkerDecisionAttributes{
				MarkerName: attr.GetMarkerName(),
				Details:    Payload(attr.Details),
				Header:     Header(attr.Header),
			},
		}
	case shared.DecisionTypeContinueAsNewWorkflowExecution:
		attr := d.ContinueAsNewWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_ContinueAsNewWorkflowExecutionDecisionAttributes{
			ContinueAsNewWorkflowExecutionDecisionAttributes: &apiv1.ContinueAsNewWorkflowExecutionDecisionAttributes{
				WorkflowType:                 WorkflowType(attr.WorkflowType),
				TaskList:                     TaskList(attr.TaskList),
				Input:                        Payload(attr.Input),
				ExecutionStartToCloseTimeout: secondsToDuration(attr.ExecutionStartToCloseTimeoutSeconds),
				TaskStartToCloseTimeout:      secondsToDuration(attr.TaskStartToCloseTimeoutSeconds),
				BackoffStartInterval:         secondsToDuration(attr.BackoffStartIntervalInSeconds),
				RetryPolicy:                  RetryPolicy(attr.RetryPolicy),
				Initiator:                    ContinueAsNewInitiator(attr.Initiator),
				Failure:                      Failure(attr.FailureReason, attr.FailureDetails),
				LastCompletionResult:         Payload(attr.LastCompletionResult),
				CronSchedule:                 attr.GetCronSchedule(),
				Header:                       Header(attr.Header),
				Memo:                         Memo(attr.Memo),
				SearchAttributes:             SearchAttributes(attr.SearchAttributes),
				JitterStart:                  secondsToDuration(attr.JitterStartSeconds),
				CronOverlapPolicy:            CronOverlapPolicy(attr.CronOverlapPolicy),
				ActiveClusterSelectionPolicy: ActiveClusterSelectionPolicy(attr.ActiveClusterSelectionPolicy),
			},
		}
	case shared.DecisionTypeStartChildWorkflowExecution:
		attr := d.StartChildWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_StartChildWorkflowExecutionDecisionAttributes{
			StartChildWorkflowExecutionDecisionAttributes: &apiv1.StartChildWorkflowExecutionDecisionAttributes{
				Domain:                       attr.GetDomain(),
				WorkflowId:                   attr.GetWorkflowId(),
				WorkflowType:                 WorkflowType(attr.WorkflowType),
				TaskList:                     TaskList(attr.TaskList),
				Input:                        Payload(attr.Input),
				ExecutionStartToCloseTimeout: secondsToDuration(attr.ExecutionStartToCloseTimeoutSeconds),
				TaskStartToCloseTimeout:      secondsToDuration(attr.TaskStartToCloseTimeoutSeconds),
				ParentClosePolicy:            ParentClosePolicy(attr.ParentClosePolicy),
				Control:                      attr.Control,
				WorkflowIdReusePolicy:        WorkflowIDReusePolicy(attr.WorkflowIdReusePolicy),
				RetryPolicy:                  RetryPolicy(attr.RetryPolicy),
				CronSchedule:                 attr.GetCronSchedule(),
				Header:                       Header(attr.Header),
				Memo:                         Memo(attr.Memo),
				SearchAttributes:             SearchAttributes(attr.SearchAttributes),
				CronOverlapPolicy:            CronOverlapPolicy(attr.CronOverlapPolicy),
				ActiveClusterSelectionPolicy: ActiveClusterSelectionPolicy(attr.ActiveClusterSelectionPolicy),
			},
		}
	case shared.DecisionTypeSignalExternalWorkflowExecution:
		attr := d.SignalExternalWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_SignalExternalWorkflowExecutionDecisionAttributes{
			SignalExternalWorkflowExecutionDecisionAttributes: &apiv1.SignalExternalWorkflowExecutionDecisionAttributes{
				Domain:            attr.GetDomain(),
				WorkflowExecution: WorkflowExecution(attr.Execution),
				SignalName:        attr.GetSignalName(),
				Input:             Payload(attr.Input),
				Control:           attr.Control,
				ChildWorkflowOnly: attr.GetChildWorkflowOnly(),
			},
		}
	case shared.DecisionTypeUpsertWorkflowSearchAttributes:
		attr := d.UpsertWorkflowSearchAttributesDecisionAttributes
		decision.Attributes = &apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes{
			UpsertWorkflowSearchAttributesDecisionAttributes: &apiv1.UpsertWorkflowSearchAttributesDecisionAttributes{
				SearchAttributes: SearchAttributes(attr.SearchAttributes),
			},
		}
	default:
		panic("unknown decision type")
	}
	return &decision
}
