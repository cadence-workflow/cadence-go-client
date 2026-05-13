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

func CreateScheduleRequest(t *shared.CreateScheduleRequest) *apiv1.CreateScheduleRequest {
	if t == nil {
		return nil
	}
	return &apiv1.CreateScheduleRequest{
		Domain:           t.GetDomain(),
		ScheduleId:       t.GetScheduleId(),
		Spec:             ScheduleSpec(t.Spec),
		Action:           ScheduleAction(t.Action),
		Policies:         SchedulePolicies(t.Policies),
		Memo:             Memo(t.Memo),
		SearchAttributes: SearchAttributes(t.SearchAttributes),
	}
}

func DescribeScheduleRequest(t *shared.DescribeScheduleRequest) *apiv1.DescribeScheduleRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DescribeScheduleRequest{
		Domain:     t.GetDomain(),
		ScheduleId: t.GetScheduleId(),
	}
}

func UpdateScheduleRequest(t *shared.UpdateScheduleRequest) *apiv1.UpdateScheduleRequest {
	if t == nil {
		return nil
	}
	return &apiv1.UpdateScheduleRequest{
		Domain:           t.GetDomain(),
		ScheduleId:       t.GetScheduleId(),
		Spec:             ScheduleSpec(t.Spec),
		Action:           ScheduleAction(t.Action),
		Policies:         SchedulePolicies(t.Policies),
		SearchAttributes: SearchAttributes(t.SearchAttributes),
	}
}

func DeleteScheduleRequest(t *shared.DeleteScheduleRequest) *apiv1.DeleteScheduleRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DeleteScheduleRequest{
		Domain:     t.GetDomain(),
		ScheduleId: t.GetScheduleId(),
	}
}

func PauseScheduleRequest(t *shared.PauseScheduleRequest) *apiv1.PauseScheduleRequest {
	if t == nil {
		return nil
	}
	return &apiv1.PauseScheduleRequest{
		Domain:     t.GetDomain(),
		ScheduleId: t.GetScheduleId(),
		Reason:     t.GetReason(),
		Identity:   t.GetIdentity(),
	}
}

func UnpauseScheduleRequest(t *shared.UnpauseScheduleRequest) *apiv1.UnpauseScheduleRequest {
	if t == nil {
		return nil
	}
	return &apiv1.UnpauseScheduleRequest{
		Domain:        t.GetDomain(),
		ScheduleId:    t.GetScheduleId(),
		Reason:        t.GetReason(),
		CatchUpPolicy: ScheduleCatchUpPolicy(t.CatchUpPolicy),
	}
}

func BackfillScheduleRequest(t *shared.BackfillScheduleRequest) *apiv1.BackfillScheduleRequest {
	if t == nil {
		return nil
	}
	return &apiv1.BackfillScheduleRequest{
		Domain:        t.GetDomain(),
		ScheduleId:    t.GetScheduleId(),
		StartTime:     unixNanoToTime(t.StartTimeNano),
		EndTime:       unixNanoToTime(t.EndTimeNano),
		OverlapPolicy: ScheduleOverlapPolicy(t.OverlapPolicy),
		BackfillId:    t.GetBackfillId(),
	}
}

func ListSchedulesRequest(t *shared.ListSchedulesRequest) *apiv1.ListSchedulesRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListSchedulesRequest{
		Domain:        t.GetDomain(),
		PageSize:      t.GetPageSize(),
		NextPageToken: t.NextPageToken,
	}
}

func ScheduleSpec(t *shared.ScheduleSpec) *apiv1.ScheduleSpec {
	if t == nil {
		return nil
	}
	return &apiv1.ScheduleSpec{
		CronExpression: t.GetCronExpression(),
		StartTime:      unixNanoToTime(t.StartTimeNano),
		EndTime:        unixNanoToTime(t.EndTimeNano),
		Jitter:         secondsToDuration(t.JitterInSeconds),
	}
}

func ScheduleAction(t *shared.ScheduleAction) *apiv1.ScheduleAction {
	if t == nil {
		return nil
	}
	return &apiv1.ScheduleAction{
		StartWorkflow: ScheduleStartWorkflowAction(t.StartWorkflow),
	}
}

func ScheduleStartWorkflowAction(t *shared.ScheduleStartWorkflowAction) *apiv1.ScheduleAction_StartWorkflowAction {
	if t == nil {
		return nil
	}
	return &apiv1.ScheduleAction_StartWorkflowAction{
		WorkflowType:                 WorkflowType(t.WorkflowType),
		TaskList:                     TaskList(t.TaskList),
		Input:                        Payload(t.Input),
		WorkflowIdPrefix:             t.GetWorkflowIdPrefix(),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		RetryPolicy:                  RetryPolicy(t.RetryPolicy),
		Memo:                         Memo(t.Memo),
		SearchAttributes:             SearchAttributes(t.SearchAttributes),
	}
}

func SchedulePolicies(t *shared.SchedulePolicies) *apiv1.SchedulePolicies {
	if t == nil {
		return nil
	}
	return &apiv1.SchedulePolicies{
		OverlapPolicy:    ScheduleOverlapPolicy(t.OverlapPolicy),
		CatchUpPolicy:    ScheduleCatchUpPolicy(t.CatchUpPolicy),
		CatchUpWindow:    secondsToDuration(t.CatchUpWindowInSeconds),
		PauseOnFailure:   t.GetPauseOnFailure(),
		BufferLimit:      t.GetBufferLimit(),
		ConcurrencyLimit: t.GetConcurrencyLimit(),
	}
}

func ScheduleOverlapPolicy(t *shared.ScheduleOverlapPolicy) apiv1.ScheduleOverlapPolicy {
	if t == nil {
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_INVALID
	}
	switch *t {
	case shared.ScheduleOverlapPolicySkipNew:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_SKIP_NEW
	case shared.ScheduleOverlapPolicyBuffer:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_BUFFER
	case shared.ScheduleOverlapPolicyConcurrent:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CONCURRENT
	case shared.ScheduleOverlapPolicyCancelPrevious:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CANCEL_PREVIOUS
	case shared.ScheduleOverlapPolicyTerminatePrevious:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_TERMINATE_PREVIOUS
	default:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_INVALID
	}
}

func ScheduleCatchUpPolicy(t *shared.ScheduleCatchUpPolicy) apiv1.ScheduleCatchUpPolicy {
	if t == nil {
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_INVALID
	}
	switch *t {
	case shared.ScheduleCatchUpPolicySkip:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_SKIP
	case shared.ScheduleCatchUpPolicyOne:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ONE
	case shared.ScheduleCatchUpPolicyAll:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ALL
	default:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_INVALID
	}
}

func SchedulePauseInfo(t *shared.SchedulePauseInfo) *apiv1.SchedulePauseInfo {
	if t == nil {
		return nil
	}
	return &apiv1.SchedulePauseInfo{
		Reason:   t.GetReason(),
		PausedAt: unixNanoToTime(t.PausedTimeNano),
		PausedBy: t.GetPausedBy(),
	}
}

func ScheduleState(t *shared.ScheduleState) *apiv1.ScheduleState {
	if t == nil {
		return nil
	}
	return &apiv1.ScheduleState{
		Paused:    t.GetPaused(),
		PauseInfo: SchedulePauseInfo(t.PauseInfo),
	}
}

func BackfillInfo(t *shared.BackfillInfo) *apiv1.BackfillInfo {
	if t == nil {
		return nil
	}
	return &apiv1.BackfillInfo{
		BackfillId:    t.GetBackfillId(),
		StartTime:     unixNanoToTime(t.StartTimeNano),
		EndTime:       unixNanoToTime(t.EndTimeNano),
		RunsCompleted: t.GetRunsCompleted(),
		RunsTotal:     t.GetRunsTotal(),
	}
}

func ScheduleInfo(t *shared.ScheduleInfo) *apiv1.ScheduleInfo {
	if t == nil {
		return nil
	}
	var ongoing []*apiv1.BackfillInfo
	if t.OngoingBackfills != nil {
		ongoing = make([]*apiv1.BackfillInfo, len(t.OngoingBackfills))
		for i, b := range t.OngoingBackfills {
			ongoing[i] = BackfillInfo(b)
		}
	}
	return &apiv1.ScheduleInfo{
		LastRunTime:      unixNanoToTime(t.LastRunTimeNano),
		NextRunTime:      unixNanoToTime(t.NextRunTimeNano),
		TotalRuns:        t.GetTotalRuns(),
		CreateTime:       unixNanoToTime(t.CreateTimeNano),
		LastUpdateTime:   unixNanoToTime(t.LastUpdateTimeNano),
		OngoingBackfills: ongoing,
	}
}

func ScheduleListEntry(t *shared.ScheduleListEntry) *apiv1.ScheduleListEntry {
	if t == nil {
		return nil
	}
	return &apiv1.ScheduleListEntry{
		ScheduleId:     t.GetScheduleId(),
		WorkflowType:   WorkflowType(t.WorkflowType),
		State:          ScheduleState(t.State),
		CronExpression: t.GetCronExpression(),
	}
}
