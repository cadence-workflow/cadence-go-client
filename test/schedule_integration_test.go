// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package test

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/internal/common"
)

// scheduleTestWorkflowType is the registered name of SimplestWorkflow used as the schedule action target.
// Matches the name derived by worker.RegisterWorkflow for a method receiver func.
const scheduleTestWorkflowType = "go.uber.org/cadence/test.(*Workflows).SimplestWorkflow"

// newScheduleID returns a unique schedule ID for the current test.
func (ts *IntegrationTestSuite) newScheduleID() string {
	return fmt.Sprintf("integ-schedule-%v-%v", ts.seq, time.Now().UnixNano())
}

// minimalCreateRequest builds a valid CreateScheduleRequest whose cron fires at most once a year
// (Jan 1 at midnight), so no workflow runs are triggered during normal test execution.
func (ts *IntegrationTestSuite) minimalCreateRequest(id string) *internal.CreateScheduleRequest {
	return &internal.CreateScheduleRequest{
		ScheduleID: id,
		Spec: &internal.ScheduleSpec{
			CronExpression: "0 0 1 1 *",
		},
		Action: &internal.ScheduleAction{
			StartWorkflow: &internal.ScheduleStartWorkflowAction{
				WorkflowType:                 scheduleTestWorkflowType,
				TaskList:                     ts.taskListName,
				ExecutionStartToCloseTimeout: 15 * time.Second,
			},
		},
	}
}

// skipIfScheduleNotSupported skips the test when the server indicates the schedule
// feature is unavailable (e.g. the feature flag is not enabled in the test cluster).
func (ts *IntegrationTestSuite) skipIfScheduleNotSupported(err error) {
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "not supported") ||
		strings.Contains(msg, "not implemented") ||
		strings.Contains(msg, "not enabled") ||
		strings.Contains(msg, "feature") {
		ts.T().Skipf("schedule feature not supported on this server: %v", err)
	}
}

// TestSchedule_CreateAndDescribe verifies that a schedule can be created and that
// Describe returns the spec and action fields that were supplied on creation.
func (ts *IntegrationTestSuite) TestSchedule_CreateAndDescribe() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp)
	ts.Require().NotNil(resp.Spec)
	ts.Equal("0 0 1 1 *", resp.Spec.CronExpression)
	ts.Require().NotNil(resp.Action)
	ts.Require().NotNil(resp.Action.StartWorkflow)
	ts.Equal(scheduleTestWorkflowType, resp.Action.StartWorkflow.WorkflowType)
	ts.Equal(ts.taskListName, resp.Action.StartWorkflow.TaskList)
}

// TestSchedule_CreateDuplicate verifies that creating a schedule with a duplicate ID
// returns an error (Create is not idempotent).
func (ts *IntegrationTestSuite) TestSchedule_CreateDuplicate() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	_, err = sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.Error(err, "second Create with same ID must fail")
}

// TestSchedule_Update verifies that Update replaces the schedule's spec and that
// the new spec is reflected in a subsequent Describe.
func (ts *IntegrationTestSuite) TestSchedule_Update() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	// Update to a monthly cron.
	err = sc.Update(ctx, &internal.UpdateScheduleRequest{
		ScheduleID: id,
		Spec:       &internal.ScheduleSpec{CronExpression: "0 0 1 * *"},
	})
	ts.NoError(err)

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp.Spec)
	ts.Equal("0 0 1 * *", resp.Spec.CronExpression)
}

// TestSchedule_UpdatePolicies verifies that Update can modify policies, including
// the *bool PauseOnFailure field, without disturbing fields that are left nil.
func (ts *IntegrationTestSuite) TestSchedule_UpdatePolicies() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	// Explicitly set PauseOnFailure = true.
	err = sc.Update(ctx, &internal.UpdateScheduleRequest{
		ScheduleID: id,
		Policies: &internal.SchedulePolicies{
			PauseOnFailure: common.BoolPtr(true),
		},
	})
	ts.NoError(err)

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp.Policies)
	ts.Equal(common.BoolPtr(true), resp.Policies.PauseOnFailure)

	// Update again with PauseOnFailure = nil: server should keep its stored value (true).
	err = sc.Update(ctx, &internal.UpdateScheduleRequest{
		ScheduleID: id,
		Policies: &internal.SchedulePolicies{
			// PauseOnFailure is nil — should not overwrite the server's stored true.
			OverlapPolicy: internal.ScheduleOverlapPolicySkipNew,
		},
	})
	ts.NoError(err)

	resp, err = sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp.Policies)
	ts.Equal(common.BoolPtr(true), resp.Policies.PauseOnFailure, "PauseOnFailure must remain true after nil update")
}

// TestSchedule_PauseAndUnpause verifies the full pause → describe → unpause → describe cycle.
// It also exercises the catchUpPolicy parameter on Unpause.
func (ts *IntegrationTestSuite) TestSchedule_PauseAndUnpause() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	// Pause the schedule.
	const pauseReason = "integration test pause"
	ts.NoError(sc.Pause(ctx, id, pauseReason))

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp.State)
	ts.True(resp.State.Paused, "schedule must be paused after Pause()")
	ts.Require().NotNil(resp.State.PauseInfo)
	ts.Equal(pauseReason, resp.State.PauseInfo.Reason)

	// Unpause with an explicit catch-up policy override.
	ts.NoError(sc.Unpause(ctx, id, "resuming after test", internal.ScheduleCatchUpPolicySkip))

	resp, err = sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp.State)
	ts.False(resp.State.Paused, "schedule must not be paused after Unpause()")
}

// TestSchedule_UnpauseUnspecifiedPolicy verifies that Unpause with
// ScheduleCatchUpPolicyUnspecified succeeds (policy field is omitted on the wire,
// deferring to the schedule's configured policy).
func (ts *IntegrationTestSuite) TestSchedule_UnpauseUnspecifiedPolicy() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	ts.NoError(sc.Pause(ctx, id, "test"))
	ts.NoError(sc.Unpause(ctx, id, "test", internal.ScheduleCatchUpPolicyUnspecified))

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.False(resp.State.Paused)
}

// TestSchedule_Delete verifies that a deleted schedule is no longer accessible via Describe.
func (ts *IntegrationTestSuite) TestSchedule_Delete() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)

	ts.NoError(sc.Delete(ctx, id))

	_, err = sc.Describe(ctx, id)
	ts.Error(err, "Describe after Delete must return an error")
}

// TestSchedule_List verifies that created schedules appear in List results.
func (ts *IntegrationTestSuite) TestSchedule_List() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()

	id1 := ts.newScheduleID()
	id2 := fmt.Sprintf("%v-b", ts.newScheduleID())

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id1))
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id1) }()

	_, err = sc.Create(ctx, ts.minimalCreateRequest(id2))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id2) }()

	// List with a generous page size; both schedules must appear.
	resp, err := sc.List(ctx, 100, nil)
	ts.NoError(err)
	ts.Require().NotNil(resp)

	found := make(map[string]bool)
	for _, s := range resp.Schedules {
		found[s.ScheduleID] = true
	}
	ts.True(found[id1], "schedule %v not found in list", id1)
	ts.True(found[id2], "schedule %v not found in list", id2)
}

// TestSchedule_Backfill verifies that a Backfill call for a past time range completes
// without error. Actual workflow execution from the backfill is not verified here
// because it requires waiting for async runs to complete.
func (ts *IntegrationTestSuite) TestSchedule_Backfill() {
	sc := ts.libClient.NewScheduleClient()
	ctx := context.Background()
	id := ts.newScheduleID()

	// Use an hourly schedule so the backfill range covers a meaningful number of windows.
	req := ts.minimalCreateRequest(id)
	req.Spec = &internal.ScheduleSpec{CronExpression: "0 * * * *"}

	_, err := sc.Create(ctx, req)
	ts.skipIfScheduleNotSupported(err)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	now := time.Now()
	err = sc.Backfill(ctx, id, &internal.BackfillRequest{
		StartTime:     now.Add(-2 * time.Hour),
		EndTime:       now.Add(-1 * time.Hour),
		OverlapPolicy: internal.ScheduleOverlapPolicySkipNew,
	})
	ts.NoError(err)
}
