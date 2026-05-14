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

func CreateScheduleResponse(t *apiv1.CreateScheduleResponse) *shared.CreateScheduleResponse {
	if t == nil {
		return nil
	}
	return &shared.CreateScheduleResponse{
		ScheduleId: &t.ScheduleId,
	}
}

func DescribeScheduleResponse(t *apiv1.DescribeScheduleResponse) *shared.DescribeScheduleResponse {
	if t == nil {
		return nil
	}
	return &shared.DescribeScheduleResponse{
		Spec:             ScheduleSpec(t.Spec),
		Action:           ScheduleAction(t.Action),
		Policies:         SchedulePolicies(t.Policies),
		State:            ScheduleState(t.State),
		Info:             ScheduleInfo(t.Info),
		Memo:             Memo(t.Memo),
		SearchAttributes: SearchAttributes(t.SearchAttributes),
	}
}

func UpdateScheduleResponse(t *apiv1.UpdateScheduleResponse) *shared.UpdateScheduleResponse {
	if t == nil {
		return nil
	}
	return &shared.UpdateScheduleResponse{}
}

func DeleteScheduleResponse(t *apiv1.DeleteScheduleResponse) *shared.DeleteScheduleResponse {
	if t == nil {
		return nil
	}
	return &shared.DeleteScheduleResponse{}
}

func PauseScheduleResponse(t *apiv1.PauseScheduleResponse) *shared.PauseScheduleResponse {
	if t == nil {
		return nil
	}
	return &shared.PauseScheduleResponse{}
}

func UnpauseScheduleResponse(t *apiv1.UnpauseScheduleResponse) *shared.UnpauseScheduleResponse {
	if t == nil {
		return nil
	}
	return &shared.UnpauseScheduleResponse{}
}

func BackfillScheduleResponse(t *apiv1.BackfillScheduleResponse) *shared.BackfillScheduleResponse {
	if t == nil {
		return nil
	}
	return &shared.BackfillScheduleResponse{}
}

func ListSchedulesResponse(t *apiv1.ListSchedulesResponse) *shared.ListSchedulesResponse {
	if t == nil {
		return nil
	}
	return &shared.ListSchedulesResponse{
		Schedules:     ScheduleListEntryArray(t.Schedules),
		NextPageToken: t.NextPageToken,
	}
}

func ScheduleSpec(t *apiv1.ScheduleSpec) *shared.ScheduleSpec {
	if t == nil {
		return nil
	}
	return &shared.ScheduleSpec{
		CronExpression:  common.StringPtr(t.CronExpression),
		StartTimeNano:   timeToUnixNano(t.StartTime),
		EndTimeNano:     timeToUnixNano(t.EndTime),
		JitterInSeconds: durationToSeconds(t.Jitter),
	}
}

func ScheduleAction(t *apiv1.ScheduleAction) *shared.ScheduleAction {
	if t == nil {
		return nil
	}
	return &shared.ScheduleAction{
		StartWorkflow: ScheduleStartWorkflowAction(t.StartWorkflow),
	}
}

func ScheduleStartWorkflowAction(t *apiv1.ScheduleAction_StartWorkflowAction) *shared.ScheduleStartWorkflowAction {
	if t == nil {
		return nil
	}
	return &shared.ScheduleStartWorkflowAction{
		WorkflowType:                        WorkflowType(t.WorkflowType),
		TaskList:                            TaskList(t.TaskList),
		Input:                               Payload(t.Input),
		WorkflowIdPrefix:                    common.StringPtr(t.WorkflowIdPrefix),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		RetryPolicy:                         RetryPolicy(t.RetryPolicy),
		Memo:                                Memo(t.Memo),
		SearchAttributes:                    SearchAttributes(t.SearchAttributes),
	}
}

func SchedulePolicies(t *apiv1.SchedulePolicies) *shared.SchedulePolicies {
	if t == nil {
		return nil
	}
	return &shared.SchedulePolicies{
		OverlapPolicy:          ScheduleOverlapPolicy(t.OverlapPolicy),
		CatchUpPolicy:          ScheduleCatchUpPolicy(t.CatchUpPolicy),
		CatchUpWindowInSeconds: durationToSeconds(t.CatchUpWindow),
		PauseOnFailure:         common.BoolPtr(t.PauseOnFailure),
		BufferLimit:            common.Int32Ptr(t.BufferLimit),
		ConcurrencyLimit:       common.Int32Ptr(t.ConcurrencyLimit),
	}
}

func ScheduleState(t *apiv1.ScheduleState) *shared.ScheduleState {
	if t == nil {
		return nil
	}
	return &shared.ScheduleState{
		Paused:    common.BoolPtr(t.Paused),
		PauseInfo: SchedulePauseInfo(t.PauseInfo),
	}
}

func SchedulePauseInfo(t *apiv1.SchedulePauseInfo) *shared.SchedulePauseInfo {
	if t == nil {
		return nil
	}
	return &shared.SchedulePauseInfo{
		Reason:         common.StringPtr(t.Reason),
		PausedTimeNano: timeToUnixNano(t.PausedAt),
		PausedBy:       common.StringPtr(t.PausedBy),
	}
}

func ScheduleInfo(t *apiv1.ScheduleInfo) *shared.ScheduleInfo {
	if t == nil {
		return nil
	}
	return &shared.ScheduleInfo{
		LastRunTimeNano:    timeToUnixNano(t.LastRunTime),
		NextRunTimeNano:    timeToUnixNano(t.NextRunTime),
		TotalRuns:          common.Int64Ptr(t.TotalRuns),
		CreateTimeNano:     timeToUnixNano(t.CreateTime),
		LastUpdateTimeNano: timeToUnixNano(t.LastUpdateTime),
		OngoingBackfills:   BackfillInfoArray(t.OngoingBackfills),
	}
}

func BackfillInfo(t *apiv1.BackfillInfo) *shared.BackfillInfo {
	if t == nil {
		return nil
	}
	return &shared.BackfillInfo{
		BackfillId:    common.StringPtr(t.BackfillId),
		StartTimeNano: timeToUnixNano(t.StartTime),
		EndTimeNano:   timeToUnixNano(t.EndTime),
		RunsCompleted: common.Int32Ptr(t.RunsCompleted),
		RunsTotal:     common.Int32Ptr(t.RunsTotal),
	}
}

func BackfillInfoArray(t []*apiv1.BackfillInfo) []*shared.BackfillInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.BackfillInfo, len(t))
	for i, e := range t {
		v[i] = BackfillInfo(e)
	}
	return v
}

func ScheduleListEntry(t *apiv1.ScheduleListEntry) *shared.ScheduleListEntry {
	if t == nil {
		return nil
	}
	return &shared.ScheduleListEntry{
		ScheduleId:     common.StringPtr(t.ScheduleId),
		WorkflowType:   WorkflowType(t.WorkflowType),
		State:          ScheduleState(t.State),
		CronExpression: common.StringPtr(t.CronExpression),
	}
}

func ScheduleListEntryArray(t []*apiv1.ScheduleListEntry) []*shared.ScheduleListEntry {
	if t == nil {
		return nil
	}
	v := make([]*shared.ScheduleListEntry, len(t))
	for i, e := range t {
		v[i] = ScheduleListEntry(e)
	}
	return v
}

func ScheduleOverlapPolicy(t apiv1.ScheduleOverlapPolicy) *shared.ScheduleOverlapPolicy {
	var v shared.ScheduleOverlapPolicy
	switch t {
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_SKIP_NEW:
		v = shared.ScheduleOverlapPolicySkipNew
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_BUFFER:
		v = shared.ScheduleOverlapPolicyBuffer
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CONCURRENT:
		v = shared.ScheduleOverlapPolicyConcurrent
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CANCEL_PREVIOUS:
		v = shared.ScheduleOverlapPolicyCancelPrevious
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_TERMINATE_PREVIOUS:
		v = shared.ScheduleOverlapPolicyTerminatePrevious
	default:
		v = shared.ScheduleOverlapPolicyInvalid
	}
	return &v
}

func ScheduleCatchUpPolicy(t apiv1.ScheduleCatchUpPolicy) *shared.ScheduleCatchUpPolicy {
	var v shared.ScheduleCatchUpPolicy
	switch t {
	case apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_SKIP:
		v = shared.ScheduleCatchUpPolicySkip
	case apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ONE:
		v = shared.ScheduleCatchUpPolicyOne
	case apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ALL:
		v = shared.ScheduleCatchUpPolicyAll
	default:
		v = shared.ScheduleCatchUpPolicyInvalid
	}
	return &v
}
