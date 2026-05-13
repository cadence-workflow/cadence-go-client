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
	gogo "github.com/gogo/protobuf/types"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

var (
	ScheduleSpec = apiv1.ScheduleSpec{
		CronExpression: "0 * * * *",
		StartTime:      Timestamp1,
		EndTime:        Timestamp2,
		Jitter:         &gogo.Duration{Seconds: 10},
	}

	SchedulePauseInfo = apiv1.SchedulePauseInfo{
		Reason:   "paused for maintenance",
		PausedAt: Timestamp3,
		PausedBy: "user@example.com",
	}

	ScheduleState = apiv1.ScheduleState{
		Paused:    true,
		PauseInfo: &SchedulePauseInfo,
	}

	SchedulePolicies = apiv1.SchedulePolicies{
		OverlapPolicy:    apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_SKIP_NEW,
		CatchUpPolicy:    apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_SKIP,
		CatchUpWindow:    &gogo.Duration{Seconds: 3600},
		PauseOnFailure:   true,
		BufferLimit:      10,
		ConcurrencyLimit: 5,
	}

	ScheduleStartWorkflowAction = apiv1.ScheduleAction_StartWorkflowAction{
		WorkflowType:                 &WorkflowType,
		TaskList:                     &TaskList,
		Input:                        &Payload1,
		WorkflowIdPrefix:             "my-schedule",
		ExecutionStartToCloseTimeout: &gogo.Duration{Seconds: 3600},
		TaskStartToCloseTimeout:      &gogo.Duration{Seconds: 10},
		RetryPolicy:                  &RetryPolicy,
		Memo:                         &Memo,
		SearchAttributes:             &SearchAttributes,
	}

	ScheduleAction = apiv1.ScheduleAction{
		StartWorkflow: &ScheduleStartWorkflowAction,
	}

	BackfillInfo = apiv1.BackfillInfo{
		BackfillId:    "backfill-1",
		StartTime:     Timestamp1,
		EndTime:       Timestamp2,
		RunsCompleted: 3,
		RunsTotal:     5,
	}

	ScheduleInfo = apiv1.ScheduleInfo{
		LastRunTime:      Timestamp1,
		NextRunTime:      Timestamp2,
		TotalRuns:        42,
		CreateTime:       Timestamp3,
		LastUpdateTime:   Timestamp4,
		OngoingBackfills: []*apiv1.BackfillInfo{&BackfillInfo},
	}

	ScheduleListEntry = apiv1.ScheduleListEntry{
		ScheduleId:     "my-schedule-id",
		WorkflowType:   &WorkflowType,
		State:          &ScheduleState,
		CronExpression: "0 * * * *",
	}
)
