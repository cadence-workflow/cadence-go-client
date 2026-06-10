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
// Conversion to/from Thrift wire types happens entirely inside internal_schedule_convert_thrift.go.

import (
	"errors"
	"time"
)

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
// This is the request-side (write) type: Memo/SearchAttributes are native Go values that
// the SDK encodes for you. The describe (read) counterpart is
// ScheduleStartWorkflowActionDescription, which returns them as raw bytes.
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
	// DecisionTaskStartToCloseTimeout is the maximum duration for a single decision task.
	// If zero, defaults to 10s. The server rejects values <= 0, so a zero value must be defaulted client-side.
	DecisionTaskStartToCloseTimeout time.Duration
	// RetryPolicy is the retry policy applied to each triggered workflow run. Optional.
	RetryPolicy *RetryPolicy
	// Memo is additional metadata attached to each triggered workflow run. Optional.
	// Values are native Go values; the SDK encodes each one with the configured DataConverter.
	Memo map[string]interface{}
	// SearchAttributes are indexed attributes attached to each triggered workflow run. Optional.
	// Values are native Go values; the SDK JSON-encodes each one.
	SearchAttributes map[string]interface{}
}

// ScheduleAction defines what the schedule does when it triggers (request-side / write).
// Currently only StartWorkflow is supported.
type ScheduleAction struct {
	// StartWorkflow starts a new workflow execution on each trigger.
	StartWorkflow *ScheduleStartWorkflowAction
}

// ScheduleStartWorkflowActionDescription is the describe (read) counterpart of
// ScheduleStartWorkflowAction. Memo/SearchAttributes are returned as raw encoded bytes —
// exactly as the server stores them — for you to decode with your DataConverter. The
// request-side type takes native values that the SDK encodes for you.
type ScheduleStartWorkflowActionDescription struct {
	WorkflowType                    string
	TaskList                        string
	Input                           []byte
	WorkflowIDPrefix                string
	ExecutionStartToCloseTimeout    time.Duration
	DecisionTaskStartToCloseTimeout time.Duration
	RetryPolicy                     *RetryPolicy
	// Memo contains the raw encoded bytes of each memo field; decode with your DataConverter.
	Memo map[string][]byte
	// SearchAttributes contains the raw encoded bytes of each attribute (typically JSON).
	SearchAttributes map[string][]byte
}

// ScheduleActionDescription is the describe (read) counterpart of ScheduleAction.
type ScheduleActionDescription struct {
	StartWorkflow *ScheduleStartWorkflowActionDescription
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
	PauseOnFailure bool
	// BufferLimit caps the number of buffered runs when OverlapPolicy is Buffer (0 = no user limit, server cap applies).
	BufferLimit int32
	// ConcurrencyLimit caps the number of concurrent runs when OverlapPolicy is Concurrent (0 = unlimited, server cap applies).
	ConcurrencyLimit int32
}

// SchedulePauseInfo records when and why a schedule was paused.
type SchedulePauseInfo struct {
	Reason string
	// PausedAt is not currently populated by the server; it will always be the zero time.Time.
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
	LastRunTime time.Time
	NextRunTime time.Time
	TotalRuns   int64
	// CreateTime is not currently populated by the server; it will always be the zero time.Time.
	CreateTime time.Time
	// LastUpdateTime is not currently populated by the server; it will always be the zero time.Time.
	LastUpdateTime time.Time
	// OngoingBackfills is not currently populated by the server; it will always be nil.
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

// ScheduleUpdate is the mutable view of a schedule's current state passed to the callback
// in ScheduleClient.Update. It is pre-populated from DescribeSchedule with the schedule's
// updatable fields. Mutate the fields you want to change and leave the rest untouched; the
// SDK sends only the top-level fields you actually changed (others are preserved server-side).
//
// Memo/SearchAttributes are raw encoded bytes (as DescribeSchedule returns them), so
// untouched values round-trip byte-for-byte. To set a new action Memo from native Go values,
// use SetActionMemo. (UpdateSchedule does not support changing schedule-level Memo.)
type ScheduleUpdate struct {
	// Spec is the schedule's trigger specification.
	Spec *ScheduleSpec
	// Action is the schedule's action (reuses the read type; Memo/SearchAttributes are raw bytes).
	Action *ScheduleActionDescription
	// Policies are the schedule's runtime policies.
	Policies *SchedulePolicies
	// SearchAttributes are the schedule-level indexed attributes, as raw encoded bytes.
	SearchAttributes map[string][]byte

	// dc encodes values for the Set* helpers; populated by the SDK.
	dc DataConverter
}

// SetActionMemo sets the action-level Memo from native Go values, encoding each with the
// client's DataConverter (the same encoding used on Create). Replaces any existing action Memo.
func (u *ScheduleUpdate) SetActionMemo(memo map[string]interface{}) error {
	if u.Action == nil || u.Action.StartWorkflow == nil {
		return errors.New("SetActionMemo: Action.StartWorkflow is nil")
	}
	encoded, err := encodeMemo(u.dc, memo)
	if err != nil {
		return err
	}
	u.Action.StartWorkflow.Memo = encoded
	return nil
}

// SetSearchAttributes sets the schedule-level SearchAttributes from native Go values,
// JSON-encoding each (the same encoding used on Create). Replaces any existing schedule-level
// SearchAttributes. Note: like UpdateSchedule itself, this can add or replace attributes but
// cannot clear them — passing an empty map is a no-op (the existing attributes are preserved).
func (u *ScheduleUpdate) SetSearchAttributes(searchAttributes map[string]interface{}) error {
	encoded, err := encodeSearchAttributes(searchAttributes)
	if err != nil {
		return err
	}
	u.SearchAttributes = encoded
	return nil
}

// SetActionSearchAttributes sets the action-level SearchAttributes from native Go values,
// JSON-encoding each (the same encoding used on Create). Replaces any existing action-level
// SearchAttributes; pass an empty map to clear them.
func (u *ScheduleUpdate) SetActionSearchAttributes(searchAttributes map[string]interface{}) error {
	if u.Action == nil || u.Action.StartWorkflow == nil {
		return errors.New("SetActionSearchAttributes: Action.StartWorkflow is nil")
	}
	encoded, err := encodeSearchAttributes(searchAttributes)
	if err != nil {
		return err
	}
	u.Action.StartWorkflow.SearchAttributes = encoded
	return nil
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
	Action   *ScheduleActionDescription
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
