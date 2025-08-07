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
	DecisionArray = []*apiv1.Decision{
		&Decision_CancelTimer,
		&Decision_CancelWorkflowExecution,
		&Decision_CompleteWorkflowExecution,
		&Decision_ContinueAsNewWorkflowExecution,
		&Decision_FailWorkflowExecution,
		&Decision_RecordMarker,
		&Decision_RequestCancelActivityTask,
		&Decision_RequestCancelExternalWorkflowExecution,
		&Decision_ScheduleActivityTask,
		&Decision_SignalExternalWorkflowExecution,
		&Decision_StartChildWorkflowExecution,
		&Decision_StartTimer,
		&Decision_UpsertWorkflowSearchAttributes,
	}

	Decision_CancelTimer = apiv1.Decision{Attributes: &apiv1.Decision_CancelTimerDecisionAttributes{
		CancelTimerDecisionAttributes: &CancelTimerDecisionAttributes,
	}}
	Decision_CancelWorkflowExecution = apiv1.Decision{Attributes: &apiv1.Decision_CancelWorkflowExecutionDecisionAttributes{
		CancelWorkflowExecutionDecisionAttributes: &CancelWorkflowExecutionDecisionAttributes,
	}}
	Decision_CompleteWorkflowExecution = apiv1.Decision{Attributes: &apiv1.Decision_CompleteWorkflowExecutionDecisionAttributes{
		CompleteWorkflowExecutionDecisionAttributes: &CompleteWorkflowExecutionDecisionAttributes,
	}}
	Decision_ContinueAsNewWorkflowExecution = apiv1.Decision{Attributes: &apiv1.Decision_ContinueAsNewWorkflowExecutionDecisionAttributes{
		ContinueAsNewWorkflowExecutionDecisionAttributes: &ContinueAsNewWorkflowExecutionDecisionAttributes,
	}}
	Decision_FailWorkflowExecution = apiv1.Decision{Attributes: &apiv1.Decision_FailWorkflowExecutionDecisionAttributes{
		FailWorkflowExecutionDecisionAttributes: &FailWorkflowExecutionDecisionAttributes,
	}}
	Decision_RecordMarker = apiv1.Decision{Attributes: &apiv1.Decision_RecordMarkerDecisionAttributes{
		RecordMarkerDecisionAttributes: &RecordMarkerDecisionAttributes,
	}}
	Decision_RequestCancelActivityTask = apiv1.Decision{Attributes: &apiv1.Decision_RequestCancelActivityTaskDecisionAttributes{
		RequestCancelActivityTaskDecisionAttributes: &RequestCancelActivityTaskDecisionAttributes,
	}}
	Decision_RequestCancelExternalWorkflowExecution = apiv1.Decision{Attributes: &apiv1.Decision_RequestCancelExternalWorkflowExecutionDecisionAttributes{
		RequestCancelExternalWorkflowExecutionDecisionAttributes: &RequestCancelExternalWorkflowExecutionDecisionAttributes,
	}}
	Decision_ScheduleActivityTask = apiv1.Decision{Attributes: &apiv1.Decision_ScheduleActivityTaskDecisionAttributes{
		ScheduleActivityTaskDecisionAttributes: &ScheduleActivityTaskDecisionAttributes,
	}}
	Decision_SignalExternalWorkflowExecution = apiv1.Decision{Attributes: &apiv1.Decision_SignalExternalWorkflowExecutionDecisionAttributes{
		SignalExternalWorkflowExecutionDecisionAttributes: &SignalExternalWorkflowExecutionDecisionAttributes,
	}}
	Decision_StartChildWorkflowExecution = apiv1.Decision{Attributes: &apiv1.Decision_StartChildWorkflowExecutionDecisionAttributes{
		StartChildWorkflowExecutionDecisionAttributes: &StartChildWorkflowExecutionDecisionAttributes,
	}}
	Decision_StartTimer = apiv1.Decision{Attributes: &apiv1.Decision_StartTimerDecisionAttributes{
		StartTimerDecisionAttributes: &StartTimerDecisionAttributes,
	}}
	Decision_UpsertWorkflowSearchAttributes = apiv1.Decision{Attributes: &apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes{
		UpsertWorkflowSearchAttributesDecisionAttributes: &UpsertWorkflowSearchAttributesDecisionAttributes,
	}}

	CancelTimerDecisionAttributes = apiv1.CancelTimerDecisionAttributes{
		TimerId: TimerID,
	}
	CancelWorkflowExecutionDecisionAttributes = apiv1.CancelWorkflowExecutionDecisionAttributes{
		Details: &Payload1,
	}
	CompleteWorkflowExecutionDecisionAttributes = apiv1.CompleteWorkflowExecutionDecisionAttributes{
		Result: &Payload1,
	}
	ContinueAsNewWorkflowExecutionDecisionAttributes = apiv1.ContinueAsNewWorkflowExecutionDecisionAttributes{
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		BackoffStartInterval:         Duration3,
		RetryPolicy:                  &RetryPolicy,
		Initiator:                    ContinueAsNewInitiator,
		Failure:                      &Failure,
		LastCompletionResult:         &Payload2,
		CronSchedule:                 CronSchedule,
		Header:                       &Header,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
		JitterStart:                  Duration4,
		CronOverlapPolicy:            apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_SKIPPED,
		ActiveClusterSelectionPolicy: ActiveClusterSelectionPolicyRegionSticky,
	}
	FailWorkflowExecutionDecisionAttributes = apiv1.FailWorkflowExecutionDecisionAttributes{
		Failure: &Failure,
	}
	RecordMarkerDecisionAttributes = apiv1.RecordMarkerDecisionAttributes{
		MarkerName: MarkerName,
		Details:    &Payload1,
		Header:     &Header,
	}
	RequestCancelActivityTaskDecisionAttributes = apiv1.RequestCancelActivityTaskDecisionAttributes{
		ActivityId: ActivityID,
	}
	RequestCancelExternalWorkflowExecutionDecisionAttributes = apiv1.RequestCancelExternalWorkflowExecutionDecisionAttributes{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		Control:           Control,
		ChildWorkflowOnly: true,
	}
	ScheduleActivityTaskDecisionAttributes = apiv1.ScheduleActivityTaskDecisionAttributes{
		ActivityId:             ActivityID,
		ActivityType:           &ActivityType,
		Domain:                 DomainName,
		TaskList:               &TaskList,
		Input:                  &Payload1,
		ScheduleToCloseTimeout: Duration1,
		ScheduleToStartTimeout: Duration2,
		StartToCloseTimeout:    Duration3,
		HeartbeatTimeout:       Duration4,
		RetryPolicy:            &RetryPolicy,
		Header:                 &Header,
		RequestLocalDispatch:   true,
	}
	SignalExternalWorkflowExecutionDecisionAttributes = apiv1.SignalExternalWorkflowExecutionDecisionAttributes{
		Domain:            DomainName,
		WorkflowExecution: &WorkflowExecution,
		SignalName:        SignalName,
		Input:             &Payload1,
		Control:           Control,
		ChildWorkflowOnly: true,
	}
	StartChildWorkflowExecutionDecisionAttributes = apiv1.StartChildWorkflowExecutionDecisionAttributes{
		Domain:                       DomainName,
		WorkflowId:                   WorkflowID,
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		ExecutionStartToCloseTimeout: Duration1,
		TaskStartToCloseTimeout:      Duration2,
		ParentClosePolicy:            ParentClosePolicy,
		Control:                      Control,
		WorkflowIdReusePolicy:        WorkflowIDReusePolicy,
		RetryPolicy:                  &RetryPolicy,
		CronSchedule:                 CronSchedule,
		Header:                       &Header,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
		CronOverlapPolicy:            apiv1.CronOverlapPolicy_CRON_OVERLAP_POLICY_SKIPPED,
		ActiveClusterSelectionPolicy: ActiveClusterSelectionPolicyRegionSticky,
	}
	StartTimerDecisionAttributes = apiv1.StartTimerDecisionAttributes{
		TimerId:            TimerID,
		StartToFireTimeout: Duration1,
	}
	UpsertWorkflowSearchAttributesDecisionAttributes = apiv1.UpsertWorkflowSearchAttributesDecisionAttributes{
		SearchAttributes: &SearchAttributes,
	}
)
