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

package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
)

// ── Enum converters ─────────────────────────────────────────────────────────

func TestScheduleOverlapPolicyRoundTrip(t *testing.T) {
	cases := []struct {
		sdk    ScheduleOverlapPolicy
		thrift s.ScheduleOverlapPolicy
	}{
		{ScheduleOverlapPolicySkipNew, s.ScheduleOverlapPolicySkipNew},
		{ScheduleOverlapPolicyBuffer, s.ScheduleOverlapPolicyBuffer},
		{ScheduleOverlapPolicyConcurrent, s.ScheduleOverlapPolicyConcurrent},
		{ScheduleOverlapPolicyCancelPrevious, s.ScheduleOverlapPolicyCancelPrevious},
		{ScheduleOverlapPolicyTerminatePrevious, s.ScheduleOverlapPolicyTerminatePrevious},
	}
	for _, tc := range cases {
		got := scheduleOverlapPolicyFromThrift(scheduleOverlapPolicyToThrift(tc.sdk))
		assert.Equal(t, tc.sdk, got, "overlap policy %v did not round-trip", tc.sdk)
		gotThrift := scheduleOverlapPolicyToThrift(scheduleOverlapPolicyFromThrift(&tc.thrift))
		require.NotNil(t, gotThrift)
		assert.Equal(t, tc.thrift, *gotThrift, "thrift overlap policy %v did not round-trip", tc.thrift)
	}
}

func TestScheduleCatchUpPolicyRoundTrip(t *testing.T) {
	cases := []struct {
		sdk    ScheduleCatchUpPolicy
		thrift s.ScheduleCatchUpPolicy
	}{
		{ScheduleCatchUpPolicySkip, s.ScheduleCatchUpPolicySkip},
		{ScheduleCatchUpPolicyOne, s.ScheduleCatchUpPolicyOne},
		{ScheduleCatchUpPolicyAll, s.ScheduleCatchUpPolicyAll},
	}
	for _, tc := range cases {
		got := scheduleCatchUpPolicyFromThrift(scheduleCatchUpPolicyToThrift(tc.sdk))
		assert.Equal(t, tc.sdk, got, "catch-up policy %v did not round-trip", tc.sdk)
		gotThrift := scheduleCatchUpPolicyToThrift(scheduleCatchUpPolicyFromThrift(&tc.thrift))
		require.NotNil(t, gotThrift)
		assert.Equal(t, tc.thrift, *gotThrift, "thrift catch-up policy %v did not round-trip", tc.thrift)
	}
}

func TestScheduleOverlapPolicyUnknownValues(t *testing.T) {
	unknown := s.ScheduleOverlapPolicy(999)
	got := scheduleOverlapPolicyFromThrift(&unknown)
	assert.Equal(t, ScheduleOverlapPolicyUnspecified, got)

	gotThrift := scheduleOverlapPolicyToThrift(ScheduleOverlapPolicy(999))
	require.NotNil(t, gotThrift)
	assert.Equal(t, s.ScheduleOverlapPolicyInvalid, *gotThrift)
}

// ── Primitive helpers ────────────────────────────────────────────────────────

func TestDurationToThriftSeconds(t *testing.T) {
	assert.Nil(t, durationToThriftSeconds(0), "zero duration must return nil")
	assert.Equal(t, int32(1), *durationToThriftSeconds(time.Second))
	assert.Equal(t, int32(1), *durationToThriftSeconds(500 * time.Millisecond), "sub-second must ceil to 1")
	assert.Equal(t, int32(2), *durationToThriftSeconds(1500 * time.Millisecond), "1.5s must ceil to 2")
	assert.Equal(t, int32(60), *durationToThriftSeconds(60 * time.Second))
}

func TestDurationRoundTrip(t *testing.T) {
	for _, secs := range []int32{1, 10, 60, 3600} {
		got := durationToThriftSeconds(thriftSecondsToDuration(&secs))
		require.NotNil(t, got)
		assert.Equal(t, secs, *got)
	}
}

func TestThriftNanoToTime(t *testing.T) {
	assert.True(t, thriftNanoToTime(nil).IsZero())

	ns := int64(1_700_000_000_000_000_000)
	got := thriftNanoToTime(&ns)
	assert.Equal(t, time.UTC, got.Location())
	assert.Equal(t, ns, got.UnixNano())
}

func TestTimeToThriftNanoRoundTrip(t *testing.T) {
	ns := int64(1_700_000_000_123_456_789)
	got := timeToThriftNano(thriftNanoToTime(&ns))
	require.NotNil(t, got)
	assert.Equal(t, ns, *got)
}

// ── Struct converters — thrift→SDK→thrift round-trips ───────────────────────

func TestScheduleSpecRoundTrip(t *testing.T) {
	assert.Nil(t, scheduleSpecFromThrift(nil))
	assert.Nil(t, scheduleSpecToThrift(nil))

	startNs := int64(1_700_000_000_000_000_000)
	endNs := int64(1_800_000_000_000_000_000)
	jitter := int32(30)
	orig := &s.ScheduleSpec{
		CronExpression:  common.StringPtr("0 * * * *"),
		StartTimeNano:   &startNs,
		EndTimeNano:     &endNs,
		JitterInSeconds: &jitter,
	}
	assert.Equal(t, orig, scheduleSpecToThrift(scheduleSpecFromThrift(orig)))
}

func TestScheduleRetryPolicyRoundTrip(t *testing.T) {
	assert.Nil(t, scheduleRetryPolicyFromThrift(nil))
	assert.Nil(t, scheduleRetryPolicyToThrift(nil))

	init := int32(5)
	maxI := int32(60)
	exp := int32(300)
	coeff := 1.5
	maxA := int32(3)
	orig := &s.RetryPolicy{
		InitialIntervalInSeconds:    &init,
		BackoffCoefficient:          &coeff,
		MaximumIntervalInSeconds:    &maxI,
		ExpirationIntervalInSeconds: &exp,
		MaximumAttempts:             &maxA,
		NonRetriableErrorReasons:    []string{"BadRequest"},
	}
	assert.Equal(t, orig, scheduleRetryPolicyToThrift(scheduleRetryPolicyFromThrift(orig)))
}

func TestScheduleRetryPolicyZeroCoefficientDefaulted(t *testing.T) {
	zero := float64(0)
	in := &s.RetryPolicy{BackoffCoefficient: &zero}
	out := scheduleRetryPolicyToThrift(scheduleRetryPolicyFromThrift(in))
	require.NotNil(t, out)
	assert.Equal(t, backoff.DefaultBackoffCoefficient, out.GetBackoffCoefficient())
}

func TestSchedulePoliciesRoundTrip(t *testing.T) {
	assert.Nil(t, schedulePoliciesFromThrift(nil))
	assert.Nil(t, schedulePoliciesToThrift(nil))

	overlap := s.ScheduleOverlapPolicyBuffer
	catchUp := s.ScheduleCatchUpPolicyOne
	window := int32(86400)
	pause := true
	buf := int32(5)
	conc := int32(2)
	orig := &s.SchedulePolicies{
		OverlapPolicy:          &overlap,
		CatchUpPolicy:          &catchUp,
		CatchUpWindowInSeconds: &window,
		PauseOnFailure:         &pause,
		BufferLimit:            &buf,
		ConcurrencyLimit:       &conc,
	}
	assert.Equal(t, orig, schedulePoliciesToThrift(schedulePoliciesFromThrift(orig)))
}

// ── Response-only converters (no SDK→thrift path) ───────────────────────────

func TestSchedulePauseInfoFromThrift(t *testing.T) {
	assert.Nil(t, schedulePauseInfoFromThrift(nil))

	ns := int64(1_700_000_000_000_000_000)
	in := &s.SchedulePauseInfo{
		Reason:         common.StringPtr("maintenance"),
		PausedTimeNano: &ns,
		PausedBy:       common.StringPtr("user@example.com"),
	}
	got := schedulePauseInfoFromThrift(in)
	require.NotNil(t, got)
	assert.Equal(t, "maintenance", got.Reason)
	assert.Equal(t, "user@example.com", got.PausedBy)
	assert.Equal(t, ns, got.PausedAt.UnixNano())
}

func TestScheduleStateFromThrift(t *testing.T) {
	assert.Nil(t, scheduleStateFromThrift(nil))

	paused := true
	in := &s.ScheduleState{
		Paused:    &paused,
		PauseInfo: &s.SchedulePauseInfo{Reason: common.StringPtr("test")},
	}
	got := scheduleStateFromThrift(in)
	require.NotNil(t, got)
	assert.True(t, got.Paused)
	require.NotNil(t, got.PauseInfo)
	assert.Equal(t, "test", got.PauseInfo.Reason)
}

func TestBackfillInfoFromThrift(t *testing.T) {
	assert.Nil(t, backfillInfoFromThrift(nil))

	startNs := int64(1_700_000_000_000_000_000)
	endNs := int64(1_800_000_000_000_000_000)
	completed := int32(3)
	total := int32(5)
	in := &s.BackfillInfo{
		BackfillId:    common.StringPtr("bf-1"),
		StartTimeNano: &startNs,
		EndTimeNano:   &endNs,
		RunsCompleted: &completed,
		RunsTotal:     &total,
	}
	got := backfillInfoFromThrift(in)
	require.NotNil(t, got)
	assert.Equal(t, "bf-1", got.BackfillID)
	assert.Equal(t, startNs, got.StartTime.UnixNano())
	assert.Equal(t, endNs, got.EndTime.UnixNano())
	assert.Equal(t, int32(3), got.RunsCompleted)
	assert.Equal(t, int32(5), got.RunsTotal)
}

func TestScheduleInfoFromThrift(t *testing.T) {
	assert.Nil(t, scheduleInfoFromThrift(nil))

	// nil OngoingBackfills must produce nil (not an empty non-nil slice)
	assert.Nil(t, scheduleInfoFromThrift(&s.ScheduleInfo{}).OngoingBackfills)

	startNs := int64(1_700_000_000_000_000_000)
	endNs := int64(1_800_000_000_000_000_000)
	total := int64(42)
	in := &s.ScheduleInfo{
		LastRunTimeNano:    &startNs,
		NextRunTimeNano:    &endNs,
		TotalRuns:          &total,
		CreateTimeNano:     &startNs,
		LastUpdateTimeNano: &endNs,
		OngoingBackfills: []*s.BackfillInfo{
			{BackfillId: common.StringPtr("bf-1")},
		},
	}
	got := scheduleInfoFromThrift(in)
	require.NotNil(t, got)
	assert.Equal(t, int64(42), got.TotalRuns)
	require.Len(t, got.OngoingBackfills, 1)
	assert.Equal(t, "bf-1", got.OngoingBackfills[0].BackfillID)
}

func TestScheduleListEntryFromThrift(t *testing.T) {
	assert.Nil(t, scheduleListEntryFromThrift(nil))

	paused := true
	in := &s.ScheduleListEntry{
		ScheduleId:     common.StringPtr("sched-1"),
		WorkflowType:   &s.WorkflowType{Name: common.StringPtr("my-wf")},
		State:          &s.ScheduleState{Paused: &paused},
		CronExpression: common.StringPtr("0 * * * *"),
	}
	got := scheduleListEntryFromThrift(in)
	require.NotNil(t, got)
	assert.Equal(t, "sched-1", got.ScheduleID)
	assert.Equal(t, "my-wf", got.WorkflowType)
	assert.True(t, got.State.Paused)
	assert.Equal(t, "0 * * * *", got.CronExpression)
}

func TestDescribeScheduleResponseFromThrift(t *testing.T) {
	assert.Nil(t, describeScheduleResponseFromThrift(nil))

	paused := bool(true)
	in := &s.DescribeScheduleResponse{
		Spec: &s.ScheduleSpec{CronExpression: common.StringPtr("0 * * * *")},
		Action: &s.ScheduleAction{
			StartWorkflow: &s.ScheduleStartWorkflowAction{
				WorkflowType: &s.WorkflowType{Name: common.StringPtr("my-wf")},
				TaskList:     &s.TaskList{Name: common.StringPtr("my-tl")},
			},
		},
		State: &s.ScheduleState{Paused: &paused},
		Memo:  &s.Memo{Fields: map[string][]byte{"k": []byte(`"v"`)}},
		SearchAttributes: &s.SearchAttributes{
			IndexedFields: map[string][]byte{"sk": []byte(`"sv"`)},
		},
	}
	got := describeScheduleResponseFromThrift(in)
	require.NotNil(t, got)
	require.NotNil(t, got.Spec)
	assert.Equal(t, "0 * * * *", got.Spec.CronExpression)
	require.NotNil(t, got.Action)
	require.NotNil(t, got.Action.StartWorkflow)
	assert.Equal(t, "my-wf", got.Action.StartWorkflow.WorkflowType)
	assert.Equal(t, "my-tl", got.Action.StartWorkflow.TaskList)
	require.NotNil(t, got.State)
	assert.True(t, got.State.Paused)
	assert.Equal(t, map[string][]byte{"k": []byte(`"v"`)}, got.Memo)
	assert.Equal(t, map[string][]byte{"sk": []byte(`"sv"`)}, got.SearchAttributes)
}

func TestListSchedulesResponseFromThrift(t *testing.T) {
	assert.Nil(t, listSchedulesResponseFromThrift(nil))

	// nil Schedules must produce nil (not an empty non-nil slice)
	assert.Nil(t, listSchedulesResponseFromThrift(&s.ListSchedulesResponse{}).Schedules)

	token := []byte("next-page")
	in := &s.ListSchedulesResponse{
		Schedules: []*s.ScheduleListEntry{
			{ScheduleId: common.StringPtr("sched-1")},
			{ScheduleId: common.StringPtr("sched-2")},
		},
		NextPageToken: token,
	}
	got := listSchedulesResponseFromThrift(in)
	require.NotNil(t, got)
	require.Len(t, got.Schedules, 2)
	assert.Equal(t, "sched-1", got.Schedules[0].ScheduleID)
	assert.Equal(t, "sched-2", got.Schedules[1].ScheduleID)
	assert.Equal(t, token, got.NextPageToken)
}

// ── Write-path converters (SDK→thrift) ──────────────────────────────────────

func TestScheduleCreateRequestToThrift(t *testing.T) {
	dc := getDefaultDataConverter()
	_, err := scheduleCreateRequestToThrift("dom", nil, dc)
	require.Error(t, err, "nil request must error")

	_, err = scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{}, dc)
	require.Error(t, err, "missing ScheduleID must error")

	_, err = scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
	}, dc)
	require.Error(t, err, "missing Spec must error")

	_, err = scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
		Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
	}, dc)
	require.Error(t, err, "missing Action.StartWorkflow must error")

	_, err = scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
		Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
		Action: &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				TaskList:                     "my-tl",
				ExecutionStartToCloseTimeout: time.Hour,
			},
		},
	}, dc)
	require.Error(t, err, "missing WorkflowType must error")

	_, err = scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
		Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
		Action: &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType:                 "my-wf",
				ExecutionStartToCloseTimeout: time.Hour,
			},
		},
	}, dc)
	require.Error(t, err, "missing TaskList must error")

	_, err = scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
		Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
		Action: &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType: "my-wf",
				TaskList:     "my-tl",
			},
		},
	}, dc)
	require.Error(t, err, "missing ExecutionStartToCloseTimeout must error")

	_, err = scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
		Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
		Action: &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType:                    "my-wf",
				TaskList:                        "my-tl",
				ExecutionStartToCloseTimeout:    time.Hour,
				DecisionTaskStartToCloseTimeout: -time.Second,
			},
		},
	}, dc)
	require.Error(t, err, "negative DecisionTaskStartToCloseTimeout must error")

	// zero DecisionTaskStartToCloseTimeout defaults to 10s
	gotDefault, err := scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
		Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
		Action: &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType:                 "my-wf",
				TaskList:                     "my-tl",
				ExecutionStartToCloseTimeout: time.Hour,
			},
		},
	}, dc)
	require.NoError(t, err)
	assert.Equal(t, int32(defaultDecisionTaskTimeoutInSecs), gotDefault.Action.StartWorkflow.GetTaskStartToCloseTimeoutSeconds())

	got, err := scheduleCreateRequestToThrift("dom", &CreateScheduleRequest{
		ScheduleID: "id",
		Spec: &ScheduleSpec{
			CronExpression: "0 * * * *",
			Jitter:         500 * time.Millisecond,
		},
		Action: &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType:                 "my-wf",
				TaskList:                     "my-tl",
				ExecutionStartToCloseTimeout: time.Hour,
				RetryPolicy:                  &RetryPolicy{MaximumAttempts: 3},
			},
		},
	}, dc)
	require.NoError(t, err)
	assert.Equal(t, "dom", got.GetDomain())
	assert.Equal(t, "id", got.GetScheduleId())
	assert.Equal(t, "0 * * * *", got.Spec.GetCronExpression())
	assert.Equal(t, int32(1), got.Spec.GetJitterInSeconds(), "500ms jitter must ceil to 1s")
	assert.Equal(t, "my-wf", got.Action.StartWorkflow.WorkflowType.GetName())
	assert.Equal(t, backoff.DefaultBackoffCoefficient, got.Action.StartWorkflow.RetryPolicy.GetBackoffCoefficient())
}

func TestScheduleUpdateRequestToThrift(t *testing.T) {
	dc := getDefaultDataConverter()

	_, err := scheduleUpdateRequestToThrift("dom", nil, dc)
	require.Error(t, err, "nil request must error")

	_, err = scheduleUpdateRequestToThrift("dom", &UpdateScheduleRequest{}, dc)
	require.Error(t, err, "missing ScheduleID must error")

	_, err = scheduleUpdateRequestToThrift("dom", &UpdateScheduleRequest{ScheduleID: "id"}, dc)
	require.Error(t, err, "no fields set must error")

	// Non-nil Action with nil StartWorkflow must error — not silently corrupt the schedule.
	_, err = scheduleUpdateRequestToThrift("dom", &UpdateScheduleRequest{
		ScheduleID: "id",
		Action:     &ScheduleAction{StartWorkflow: nil},
	}, dc)
	require.Error(t, err, "Action with nil StartWorkflow must error")

	// Non-nil Spec with empty CronExpression must error — the scheduler workflow silently
	// ignores updates with empty cron, so the client must catch this before the RPC.
	_, err = scheduleUpdateRequestToThrift("dom", &UpdateScheduleRequest{
		ScheduleID: "id",
		Spec:       &ScheduleSpec{},
	}, dc)
	require.Error(t, err, "Spec with empty CronExpression must error")

	got, err := scheduleUpdateRequestToThrift("dom", &UpdateScheduleRequest{
		ScheduleID: "id",
		Action: &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType:                 "my-wf",
				TaskList:                     "my-tl",
				ExecutionStartToCloseTimeout: time.Hour,
			},
		},
	}, dc)
	require.NoError(t, err)
	assert.Equal(t, "dom", got.GetDomain())
	assert.Equal(t, "id", got.GetScheduleId())
	assert.Equal(t, "my-wf", got.Action.StartWorkflow.WorkflowType.GetName())
}

func TestScheduleStartWorkflowActionFromThriftLossyFields(t *testing.T) {
	// Memo and SearchAttributes are intentionally dropped on the read path:
	// they are encoded bytes that cannot be reconstructed without a DataConverter.
	in := &s.ScheduleStartWorkflowAction{
		WorkflowType: &s.WorkflowType{Name: common.StringPtr("my-wf")},
		TaskList:     &s.TaskList{Name: common.StringPtr("my-tl")},
		Memo:         &s.Memo{Fields: map[string][]byte{"k": []byte(`"v"`)}},
		SearchAttributes: &s.SearchAttributes{
			IndexedFields: map[string][]byte{"sk": []byte(`"sv"`)},
		},
	}
	got := scheduleStartWorkflowActionFromThrift(in)
	require.NotNil(t, got)
	assert.Equal(t, "my-wf", got.WorkflowType)
	assert.Nil(t, got.Memo, "Memo must be dropped (no DataConverter on read path)")
	assert.Nil(t, got.SearchAttributes, "SearchAttributes must be dropped (no DataConverter on read path)")
}
