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

// This file defines all Go-native types for the schedule API. It is intentionally proto-free:
// no proto imports appear here so that callers never encounter generated types at the API boundary.
// Conversion to/from proto wire types happens entirely inside internal_schedule_client.go.

import "time"

// ScheduleOverlapPolicy defines behavior when a new run is triggered while a previous run is still active.
type ScheduleOverlapPolicy int

const (
	// ScheduleOverlapPolicyUnspecified defers to the server default (equivalent to SkipNew).
	ScheduleOverlapPolicyUnspecified ScheduleOverlapPolicy = iota
	// ScheduleOverlapPolicySkipNew skips the new run if the previous is still active.
	ScheduleOverlapPolicySkipNew
	// ScheduleOverlapPolicyBuffer queues new runs and executes them sequentially.
	// The maximum queue depth is controlled by SchedulePolicies.BufferLimit.
	ScheduleOverlapPolicyBuffer
	// ScheduleOverlapPolicyConcurrent allows multiple runs to execute simultaneously.
	// The maximum number of concurrent runs is controlled by SchedulePolicies.ConcurrencyLimit.
	ScheduleOverlapPolicyConcurrent
	// ScheduleOverlapPolicyCancelPrevious cancels the active run gracefully and starts the new one.
	ScheduleOverlapPolicyCancelPrevious
	// ScheduleOverlapPolicyTerminatePrevious terminates the active run immediately and starts the new one.
	ScheduleOverlapPolicyTerminatePrevious
)

// ScheduleCatchUpPolicy controls how missed runs are handled when a schedule resumes after being paused.
type ScheduleCatchUpPolicy int

const (
	// ScheduleCatchUpPolicyUnspecified defers to the server default.
	ScheduleCatchUpPolicyUnspecified ScheduleCatchUpPolicy = iota
	// ScheduleCatchUpPolicySkip discards all missed runs.
	ScheduleCatchUpPolicySkip
	// ScheduleCatchUpPolicyOne executes at most one missed run on resume.
	ScheduleCatchUpPolicyOne
	// ScheduleCatchUpPolicyAll executes every missed run on resume.
	ScheduleCatchUpPolicyAll
)

// ScheduleSpec defines when a schedule triggers.
type ScheduleSpec struct {
	// CronExpression is a standard five-field cron expression (e.g. "0 2 * * *" for 2 AM daily).
	CronExpression string
	// StartTime is the earliest time at which the schedule may trigger. Zero means no lower bound.
	StartTime time.Time
	// EndTime is the latest time at which the schedule may trigger. Zero means no upper bound.
	EndTime time.Time
	// Jitter adds a random delay up to this duration after each nominal trigger time.
	Jitter time.Duration
}

// ScheduleStartWorkflowAction describes the workflow to start when the schedule triggers.
type ScheduleStartWorkflowAction struct {
	// WorkflowType is the registered name of the workflow function to execute. Required.
	WorkflowType string
	// TaskList is the task list on which the workflow will run. Required.
	TaskList string
	// Input is the pre-encoded binary payload passed as input to the workflow on each run.
	// Encode using your configured DataConverter or json.Marshal for simple types.
	Input []byte
	// WorkflowIDPrefix is prepended to a server-generated suffix to form each run's workflow ID.
	// If empty, the server uses the schedule ID as the prefix.
	WorkflowIDPrefix string
	// ExecutionStartToCloseTimeout is the maximum duration for the entire workflow execution. Required.
	ExecutionStartToCloseTimeout time.Duration
	// DecisionTaskStartToCloseTimeout is the maximum duration for a single decision task. Required.
	DecisionTaskStartToCloseTimeout time.Duration
	// RetryPolicy is the retry policy applied to each triggered workflow run. Optional.
	RetryPolicy *RetryPolicy
	// Memo is additional metadata attached to each triggered workflow run. Optional.
	// Values are encoded using the DataConverter configured on the ScheduleClient.
	// Note: Memo is not populated when reading back via DescribeSchedule, because
	// the encoded bytes cannot be decoded without the original DataConverter.
	Memo map[string]interface{}
	// SearchAttributes are indexed attributes attached to each triggered workflow run. Optional.
	// Values are JSON-encoded.
	// Note: SearchAttributes is not populated when reading back via DescribeSchedule,
	// because the encoded bytes cannot be decoded without the original DataConverter.
	SearchAttributes map[string]interface{}
}

// ScheduleAction defines what the schedule does when it triggers.
// Currently only StartWorkflow is supported.
type ScheduleAction struct {
	// StartWorkflow starts a new workflow execution on each trigger.
	StartWorkflow *ScheduleStartWorkflowAction
}

// SchedulePolicies controls the runtime behavior of a schedule.
type SchedulePolicies struct {
	// OverlapPolicy defines what happens when a new run is triggered while a previous run is active.
	OverlapPolicy ScheduleOverlapPolicy
	// CatchUpPolicy defines how missed runs are handled when the schedule resumes from pause.
	CatchUpPolicy ScheduleCatchUpPolicy
	// CatchUpWindow is the maximum time window the server looks back for missed runs on resume.
	// Zero means the server uses its configured default.
	CatchUpWindow time.Duration
	// PauseOnFailure automatically pauses the schedule if a triggered workflow run fails.
	// Use nil to leave the server's current value unchanged on Update; use common.BoolPtr(true/false) to set explicitly.
	PauseOnFailure *bool
	// BufferLimit caps the number of buffered runs when OverlapPolicy is Buffer (0 = unlimited).
	BufferLimit int32
	// ConcurrencyLimit caps the number of concurrent runs when OverlapPolicy is Concurrent (0 = unlimited).
	ConcurrencyLimit int32
}

// SchedulePauseInfo records when and why a schedule was paused.
type SchedulePauseInfo struct {
	Reason   string
	PausedAt time.Time
	PausedBy string
}

// ScheduleState is the runtime pause/unpause state of a schedule.
type ScheduleState struct {
	Paused    bool
	PauseInfo *SchedulePauseInfo
}

// BackfillInfo describes a single active or completed backfill operation.
type BackfillInfo struct {
	BackfillID    string
	StartTime     time.Time
	EndTime       time.Time
	RunsCompleted int32
	RunsTotal     int32
}

// ScheduleInfo contains runtime statistics for a schedule.
type ScheduleInfo struct {
	LastRunTime      time.Time
	NextRunTime      time.Time
	TotalRuns        int64
	CreateTime       time.Time
	LastUpdateTime   time.Time
	OngoingBackfills []*BackfillInfo
}

// ScheduleListEntry is a summary of a schedule returned by ScheduleClient.List.
type ScheduleListEntry struct {
	ScheduleID     string
	WorkflowType   string
	State          *ScheduleState
	CronExpression string
}

// CreateScheduleRequest is the request to ScheduleClient.Create.
type CreateScheduleRequest struct {
	// ScheduleID is the client-provided unique identifier for the schedule within the domain. Required.
	ScheduleID string
	// Spec defines when the schedule triggers. Required.
	Spec *ScheduleSpec
	// Action defines what the schedule does when it triggers. Required.
	Action *ScheduleAction
	// Policies controls the runtime behavior. Optional; server defaults apply for unset fields.
	Policies *SchedulePolicies
	// Memo is metadata attached to the schedule itself (not to each triggered run).
	// Values are encoded using the DataConverter configured on the ScheduleClient. Optional.
	Memo map[string]interface{}
	// SearchAttributes are indexed attributes on the schedule itself. Optional.
	// Values are JSON-encoded.
	SearchAttributes map[string]interface{}
}

// UpdateScheduleRequest is the request to ScheduleClient.Update.
// Only non-nil fields are applied; nil fields leave the existing value unchanged.
type UpdateScheduleRequest struct {
	// ScheduleID identifies the schedule to update. Required.
	ScheduleID string
	// Spec replaces the trigger specification when non-nil.
	Spec *ScheduleSpec
	// Action replaces the schedule action when non-nil.
	Action *ScheduleAction
	// Policies replaces the schedule policies when non-nil.
	Policies *SchedulePolicies
	// SearchAttributes replaces the schedule's search attributes when non-nil.
	SearchAttributes map[string]interface{}
}

// BackfillRequest triggers workflow runs for a historical time range.
type BackfillRequest struct {
	// StartTime is the beginning of the backfill window. Required.
	StartTime time.Time
	// EndTime is the end of the backfill window. Required.
	EndTime time.Time
	// OverlapPolicy controls how backfill runs interact with any currently active runs.
	// Unspecified defers to the schedule's configured OverlapPolicy.
	OverlapPolicy ScheduleOverlapPolicy
	// BackfillID is a client-provided identifier used for idempotency and progress tracking.
	// The server generates a UUID if empty.
	BackfillID string
}

// DescribeScheduleResponse is returned by ScheduleClient.Describe.
type DescribeScheduleResponse struct {
	Spec     *ScheduleSpec
	Action   *ScheduleAction
	Policies *SchedulePolicies
	State    *ScheduleState
	Info     *ScheduleInfo
	// Memo contains the raw DataConverter-encoded bytes of each memo field.
	// Decode individual values using your configured DataConverter.
	Memo map[string][]byte
	// SearchAttributes contains the raw JSON-encoded bytes of each search attribute.
	SearchAttributes map[string][]byte
}

// ListSchedulesResponse is returned by ScheduleClient.List.
type ListSchedulesResponse struct {
	Schedules     []*ScheduleListEntry
	NextPageToken []byte
}
