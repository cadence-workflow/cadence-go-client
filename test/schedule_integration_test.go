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
	"reflect"
	"runtime"
	"strings"
	"time"

	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/internal"
)

// scheduleTestWorkflowType is the registered name of SimplestWorkflow, derived the same way
// the SDK does internally (runtime.FuncForPC + strip the "-fm" bound-method suffix).
// Using reflect keeps this in sync if the function is renamed — a rename is a compile error here.
var scheduleTestWorkflowType = strings.TrimSuffix(
	runtime.FuncForPC(reflect.ValueOf((*Workflows)(nil).SimplestWorkflow).Pointer()).Name(),
	"-fm",
)

// scheduleRunID is captured once when the package loads. Each docker compose retry attempt starts
// a fresh go test process at a different wall-clock nanosecond, so schedule IDs are unique across
// retry attempts even when the Cadence server retains state between retries on the same CI runner.
// os.Getpid() is not safe here: Docker containers reset PID namespaces from 1, so PIDs are
// identical across retry runs.
var scheduleRunID = time.Now().UnixNano()

// scheduleID returns a unique schedule ID for the current test using ts.seq (per-test counter)
// and scheduleRunID (per-process start, avoids collisions across CI retry attempts).
func (ts *IntegrationTestSuite) scheduleID(suffix ...string) string {
	if len(suffix) > 0 {
		return fmt.Sprintf("integ-schedule-%v-%v-%v", scheduleRunID, ts.seq, suffix[0])
	}
	return fmt.Sprintf("integ-schedule-%v-%v", scheduleRunID, ts.seq)
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

// fullCreateRequest builds a CreateScheduleRequest that exercises every settable field
// (rich Spec, Action, Policies, and both schedule- and action-level Memo) so a
// create -> describe round-trip validates field-by-field fidelity — including the action
// Memo/SearchAttributes that DescribeSchedule must return. Custom SearchAttributes are
// intentionally omitted: they must be pre-registered in the cluster or Create is rejected.
// The cron fires at most once a year, so no runs are triggered during the test.
func (ts *IntegrationTestSuite) fullCreateRequest(id string) *internal.CreateScheduleRequest {
	now := time.Now().Truncate(time.Second)
	return &internal.CreateScheduleRequest{
		ScheduleID: id,
		Spec: &internal.ScheduleSpec{
			CronExpression: "0 0 1 1 *",
			StartTime:      now,
			EndTime:        now.Add(365 * 24 * time.Hour),
			Jitter:         5 * time.Second,
		},
		Action: &internal.ScheduleAction{
			StartWorkflow: &internal.ScheduleStartWorkflowAction{
				WorkflowType:                    scheduleTestWorkflowType,
				TaskList:                        ts.taskListName,
				Input:                           []byte(`"hello"`),
				WorkflowIDPrefix:                "integ-full",
				ExecutionStartToCloseTimeout:    15 * time.Second,
				DecisionTaskStartToCloseTimeout: 10 * time.Second,
				RetryPolicy: &internal.RetryPolicy{
					InitialInterval: time.Second,
					MaximumAttempts: 3,
				},
				Memo: map[string]interface{}{"actionOwner": "integ-action"},
			},
		},
		Policies: &internal.SchedulePolicies{
			OverlapPolicy:    internal.ScheduleOverlapPolicySkipNew,
			CatchUpPolicy:    internal.ScheduleCatchUpPolicyOne,
			CatchUpWindow:    time.Hour,
			PauseOnFailure:   true,
			BufferLimit:      10,
			ConcurrencyLimit: 5,
		},
		Memo: map[string]interface{}{"scheduleOwner": "integ-schedule"},
	}
}

// decodeMemoString decodes a single raw memo field with the default DataConverter
// (DescribeSchedule returns Memo as raw bytes; the caller decodes).
func (ts *IntegrationTestSuite) decodeMemoString(memo map[string][]byte, key string) string {
	raw, ok := memo[key]
	ts.Require().True(ok, "memo key %q missing from Describe response", key)
	var v string
	ts.Require().NoError(encoded.GetDefaultDataConverter().FromData(raw, &v))
	return v
}

// TestSchedule_CreateAndDescribe creates a schedule with every settable field and verifies
// Describe returns them all (field-by-field round-trip), including action-level Memo.
func (ts *IntegrationTestSuite) TestSchedule_CreateAndDescribe() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	req := ts.fullCreateRequest(id)
	_, err := sc.Create(ctx, req)
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp)

	// Spec (times round-trip as UTC instants; compare the instant, not the location).
	ts.Require().NotNil(resp.Spec)
	ts.Equal(req.Spec.CronExpression, resp.Spec.CronExpression)
	ts.WithinDuration(req.Spec.StartTime, resp.Spec.StartTime, time.Second)
	ts.WithinDuration(req.Spec.EndTime, resp.Spec.EndTime, time.Second)
	ts.Equal(req.Spec.Jitter, resp.Spec.Jitter)

	// Action.
	ts.Require().NotNil(resp.Action)
	ts.Require().NotNil(resp.Action.StartWorkflow)
	a := resp.Action.StartWorkflow
	ts.Equal(scheduleTestWorkflowType, a.WorkflowType)
	ts.Equal(ts.taskListName, a.TaskList)
	ts.Equal([]byte(`"hello"`), a.Input)
	ts.Equal("integ-full", a.WorkflowIDPrefix)
	ts.Equal(15*time.Second, a.ExecutionStartToCloseTimeout)
	ts.Equal(10*time.Second, a.DecisionTaskStartToCloseTimeout)
	ts.Require().NotNil(a.RetryPolicy)
	ts.Equal(int32(3), a.RetryPolicy.MaximumAttempts)

	// Policies.
	ts.Require().NotNil(resp.Policies)
	ts.Equal(internal.ScheduleOverlapPolicySkipNew, resp.Policies.OverlapPolicy)
	ts.Equal(internal.ScheduleCatchUpPolicyOne, resp.Policies.CatchUpPolicy)
	ts.Equal(time.Hour, resp.Policies.CatchUpWindow)
	ts.True(resp.Policies.PauseOnFailure)
	ts.Equal(int32(10), resp.Policies.BufferLimit)
	ts.Equal(int32(5), resp.Policies.ConcurrencyLimit)

	// Memo round-trips as raw bytes at both levels (action-level Memo is the field this
	// branch's lossless-Describe change restored).
	ts.Equal("integ-schedule", ts.decodeMemoString(resp.Memo, "scheduleOwner"))
	ts.Equal("integ-action", ts.decodeMemoString(a.Memo, "actionOwner"))
}

// TestSchedule_CreateDuplicate verifies that creating a schedule with a duplicate ID
// returns an error (Create is not idempotent).
func (ts *IntegrationTestSuite) TestSchedule_CreateDuplicate() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	_, err = sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.Error(err, "second Create with same ID must fail")
}

// TestSchedule_Update verifies the describe-then-update flow: changing one top-level field
// (Spec) is reflected on Describe, and the untouched top-level fields (Action, Policies) are
// preserved — i.e. the SDK sends only what changed and the server's top-level merge keeps the rest.
func (ts *IntegrationTestSuite) TestSchedule_Update() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	_, err := sc.Create(ctx, ts.fullCreateRequest(id))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	err = sc.Update(ctx, id, func(u *internal.ScheduleUpdate) error {
		u.Spec.CronExpression = "0 0 1 * *"
		return nil
	})
	ts.NoError(err)

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	// Changed field applied.
	ts.Require().NotNil(resp.Spec)
	ts.Equal("0 0 1 * *", resp.Spec.CronExpression)
	// Untouched top-level fields preserved.
	ts.Require().NotNil(resp.Policies)
	ts.True(resp.Policies.PauseOnFailure)
	ts.Equal(int32(10), resp.Policies.BufferLimit)
	ts.Require().NotNil(resp.Action)
	ts.Require().NotNil(resp.Action.StartWorkflow)
	ts.Equal("integ-full", resp.Action.StartWorkflow.WorkflowIDPrefix)
}

// TestSchedule_UpdatePolicies verifies within-struct sub-field partial updates: changing one
// policy sub-field preserves the others. This is the core capability describe-then-update adds —
// the SDK reads the current Policies, applies the one change, and resends the whole struct so the
// server's within-struct full replacement doesn't reset the siblings.
func (ts *IntegrationTestSuite) TestSchedule_UpdatePolicies() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	// fullCreateRequest sets PauseOnFailure=true, BufferLimit=10, ConcurrencyLimit=5.
	_, err := sc.Create(ctx, ts.fullCreateRequest(id))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	// Flip a single policy sub-field; the others must survive.
	err = sc.Update(ctx, id, func(u *internal.ScheduleUpdate) error {
		u.Policies.PauseOnFailure = false
		return nil
	})
	ts.NoError(err)

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.Require().NotNil(resp.Policies)
	ts.False(resp.Policies.PauseOnFailure, "changed sub-field applied")
	ts.Equal(int32(10), resp.Policies.BufferLimit, "sibling sub-field preserved")
	ts.Equal(int32(5), resp.Policies.ConcurrencyLimit, "sibling sub-field preserved")
}

// TestSchedule_PauseAndUnpause verifies the full pause → describe → unpause → describe cycle.
// It also exercises the catchUpPolicy parameter on Unpause.
func (ts *IntegrationTestSuite) TestSchedule_PauseAndUnpause() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

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
// ScheduleCatchUpPolicyUnspecified succeeds. Unspecified maps to nil on the wire,
// so the field is omitted and the server defers to its configured policy.
func (ts *IntegrationTestSuite) TestSchedule_UnpauseUnspecifiedPolicy() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id) }()

	ts.NoError(sc.Pause(ctx, id, "test"))
	ts.NoError(sc.Unpause(ctx, id, "test", internal.ScheduleCatchUpPolicyUnspecified))

	resp, err := sc.Describe(ctx, id)
	ts.NoError(err)
	ts.False(resp.State.Paused)
}

// TestSchedule_Delete verifies that a deleted schedule no longer appears in List results.
// Deletion is verified via List (not Describe) because the server keeps the completed
// scheduler workflow queryable after deletion, so Describe returns nil error indefinitely.
// waitForWorkflowFinish blocks until the underlying scheduler workflow exits,
// since DeleteSchedule is async: it signals the workflow and returns immediately.
func (ts *IntegrationTestSuite) TestSchedule_Delete() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id))
	ts.NoError(err)

	ts.NoError(sc.Delete(ctx, id))

	// Wait for the scheduler workflow to exit before checking List.
	// The server workflow ID is "cadence-scheduler:<scheduleID>".
	ts.NoError(ts.waitForWorkflowFinish("cadence-scheduler:"+id, ""))

	listResp, err := sc.List(ctx, 100, nil)
	ts.NoError(err)
	for _, s := range listResp.Schedules {
		ts.NotEqual(id, s.ScheduleID, "deleted schedule must not appear in List")
	}
}

// TestSchedule_List verifies that created schedules appear in List results.
func (ts *IntegrationTestSuite) TestSchedule_List() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id1 := ts.scheduleID("1")
	id2 := ts.scheduleID("2")

	// Clean up any schedules left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id1)
	_ = sc.Delete(ctx, id2)

	_, err := sc.Create(ctx, ts.minimalCreateRequest(id1))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id1) }()

	_, err = sc.Create(ctx, ts.minimalCreateRequest(id2))
	ts.NoError(err)
	defer func() { _ = sc.Delete(context.Background(), id2) }()

	// Create is async: the scheduler workflow may not be indexed by List immediately.
	// Poll until both schedules appear or the test context expires.
	var found map[string]bool
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(500 * time.Millisecond) {
		resp, err := sc.List(ctx, 100, nil)
		ts.NoError(err)
		found = make(map[string]bool)
		for _, s := range resp.Schedules {
			found[s.ScheduleID] = true
		}
		if found[id1] && found[id2] {
			break
		}
	}
	ts.True(found[id1], "schedule %v not found in list", id1)
	ts.True(found[id2], "schedule %v not found in list", id2)
}

// TestSchedule_Backfill verifies that a Backfill call for a past time range completes
// without error. The hourly schedule fires once within the [-2h, -1h] range; actual
// workflow execution is picked up by the test worker on ts.taskListName.
func (ts *IntegrationTestSuite) TestSchedule_Backfill() {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()

	sc := ts.libClient.ScheduleClient()
	id := ts.scheduleID()

	req := ts.minimalCreateRequest(id)
	req.Spec = &internal.ScheduleSpec{CronExpression: "0 * * * *"}

	// Clean up any schedule left over from a previous (failed or retried) run.
	_ = sc.Delete(ctx, id)

	_, err := sc.Create(ctx, req)
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
