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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	shared "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
)

// ── Enum converters ──────────────────────────────────────────────────────────

func scheduleOverlapPolicyToThrift(p ScheduleOverlapPolicy) *shared.ScheduleOverlapPolicy {
	switch p {
	case ScheduleOverlapPolicyUnspecified:
		return nil
	case ScheduleOverlapPolicySkipNew:
		v := shared.ScheduleOverlapPolicySkipNew
		return &v
	case ScheduleOverlapPolicyBuffer:
		v := shared.ScheduleOverlapPolicyBuffer
		return &v
	case ScheduleOverlapPolicyConcurrent:
		v := shared.ScheduleOverlapPolicyConcurrent
		return &v
	case ScheduleOverlapPolicyCancelPrevious:
		v := shared.ScheduleOverlapPolicyCancelPrevious
		return &v
	case ScheduleOverlapPolicyTerminatePrevious:
		v := shared.ScheduleOverlapPolicyTerminatePrevious
		return &v
	default:
		v := shared.ScheduleOverlapPolicyInvalid
		return &v
	}
}

func scheduleOverlapPolicyFromThrift(p *shared.ScheduleOverlapPolicy) ScheduleOverlapPolicy {
	if p == nil {
		return ScheduleOverlapPolicyUnspecified
	}
	switch *p {
	case shared.ScheduleOverlapPolicySkipNew:
		return ScheduleOverlapPolicySkipNew
	case shared.ScheduleOverlapPolicyBuffer:
		return ScheduleOverlapPolicyBuffer
	case shared.ScheduleOverlapPolicyConcurrent:
		return ScheduleOverlapPolicyConcurrent
	case shared.ScheduleOverlapPolicyCancelPrevious:
		return ScheduleOverlapPolicyCancelPrevious
	case shared.ScheduleOverlapPolicyTerminatePrevious:
		return ScheduleOverlapPolicyTerminatePrevious
	default:
		return ScheduleOverlapPolicyUnspecified
	}
}

func scheduleCatchUpPolicyToThrift(p ScheduleCatchUpPolicy) *shared.ScheduleCatchUpPolicy {
	switch p {
	case ScheduleCatchUpPolicyUnspecified:
		return nil
	case ScheduleCatchUpPolicySkip:
		v := shared.ScheduleCatchUpPolicySkip
		return &v
	case ScheduleCatchUpPolicyOne:
		v := shared.ScheduleCatchUpPolicyOne
		return &v
	case ScheduleCatchUpPolicyAll:
		v := shared.ScheduleCatchUpPolicyAll
		return &v
	default:
		v := shared.ScheduleCatchUpPolicyInvalid
		return &v
	}
}

func scheduleCatchUpPolicyFromThrift(p *shared.ScheduleCatchUpPolicy) ScheduleCatchUpPolicy {
	if p == nil {
		return ScheduleCatchUpPolicyUnspecified
	}
	switch *p {
	case shared.ScheduleCatchUpPolicySkip:
		return ScheduleCatchUpPolicySkip
	case shared.ScheduleCatchUpPolicyOne:
		return ScheduleCatchUpPolicyOne
	case shared.ScheduleCatchUpPolicyAll:
		return ScheduleCatchUpPolicyAll
	default:
		return ScheduleCatchUpPolicyUnspecified
	}
}

// ── Time helpers ─────────────────────────────────────────────────────────────

func timeToThriftNano(t time.Time) *int64 {
	if t.IsZero() {
		return nil
	}
	return common.Int64Ptr(t.UnixNano())
}

func thriftNanoToTime(n *int64) time.Time {
	if n == nil {
		return time.Time{}
	}
	return time.Unix(0, *n).UTC()
}

func durationToThriftSeconds(d time.Duration) *int32 {
	if d == 0 {
		return nil
	}
	return common.Int32Ptr(common.Int32Ceil(d.Seconds()))
}

func thriftSecondsToDuration(n *int32) time.Duration {
	if n == nil {
		return 0
	}
	return time.Duration(*n) * time.Second
}

// ── Write-side converters (SDK → Thrift) ────────────────────────────────────

func scheduleSpecToThrift(spec *ScheduleSpec) *shared.ScheduleSpec {
	if spec == nil {
		return nil
	}
	return &shared.ScheduleSpec{
		CronExpression:  common.StringPtr(spec.CronExpression),
		StartTimeNano:   timeToThriftNano(spec.StartTime),
		EndTimeNano:     timeToThriftNano(spec.EndTime),
		JitterInSeconds: durationToThriftSeconds(spec.Jitter),
	}
}

func scheduleRetryPolicyToThrift(p *RetryPolicy) *shared.RetryPolicy {
	if p == nil {
		return nil
	}
	backoffCoeff := p.BackoffCoefficient
	if backoffCoeff == 0 {
		backoffCoeff = backoff.DefaultBackoffCoefficient
	}
	return &shared.RetryPolicy{
		InitialIntervalInSeconds:    durationToThriftSeconds(p.InitialInterval),
		BackoffCoefficient:          &backoffCoeff,
		MaximumIntervalInSeconds:    durationToThriftSeconds(p.MaximumInterval),
		ExpirationIntervalInSeconds: durationToThriftSeconds(p.ExpirationInterval),
		MaximumAttempts:             common.Int32Ptr(p.MaximumAttempts),
		NonRetriableErrorReasons:    p.NonRetriableErrorReasons,
	}
}

func scheduleStartWorkflowActionToThrift(a *ScheduleStartWorkflowAction, dc DataConverter) (*shared.ScheduleStartWorkflowAction, error) {
	if a == nil {
		return nil, nil
	}
	if a.WorkflowType == "" {
		return nil, errors.New("StartWorkflow: WorkflowType is required")
	}
	if a.TaskList == "" {
		return nil, errors.New("StartWorkflow: TaskList is required")
	}
	if a.ExecutionStartToCloseTimeout <= 0 {
		return nil, errors.New("StartWorkflow: ExecutionStartToCloseTimeout is required")
	}
	decisionTaskTimeout := a.DecisionTaskStartToCloseTimeout
	if decisionTaskTimeout < 0 {
		return nil, errors.New("StartWorkflow: DecisionTaskStartToCloseTimeout must not be negative")
	}
	if decisionTaskTimeout == 0 {
		decisionTaskTimeout = time.Duration(defaultDecisionTaskTimeoutInSecs) * time.Second
	}
	var memo *shared.Memo
	if len(a.Memo) > 0 {
		fields := make(map[string][]byte, len(a.Memo))
		for k, v := range a.Memo {
			b, err := encodeArg(dc, v)
			if err != nil {
				return nil, fmt.Errorf("encode memo field %q: %w", k, err)
			}
			fields[k] = b
		}
		memo = &shared.Memo{Fields: fields}
	}
	var searchAttr *shared.SearchAttributes
	if len(a.SearchAttributes) > 0 {
		fields := make(map[string][]byte, len(a.SearchAttributes))
		for k, v := range a.SearchAttributes {
			b, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("encode search attribute %q: %w", k, err)
			}
			fields[k] = b
		}
		searchAttr = &shared.SearchAttributes{IndexedFields: fields}
	}
	var input []byte
	if len(a.Input) > 0 {
		input = a.Input
	}
	return &shared.ScheduleStartWorkflowAction{
		WorkflowType:                        &shared.WorkflowType{Name: common.StringPtr(a.WorkflowType)},
		TaskList:                            &shared.TaskList{Name: common.StringPtr(a.TaskList)},
		Input:                               input,
		WorkflowIdPrefix:                    common.StringPtr(a.WorkflowIDPrefix),
		ExecutionStartToCloseTimeoutSeconds: durationToThriftSeconds(a.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToThriftSeconds(decisionTaskTimeout),
		RetryPolicy:                         scheduleRetryPolicyToThrift(a.RetryPolicy),
		Memo:                                memo,
		SearchAttributes:                    searchAttr,
	}, nil
}

func scheduleActionToThrift(a *ScheduleAction, dc DataConverter) (*shared.ScheduleAction, error) {
	if a == nil {
		return nil, nil
	}
	if a.StartWorkflow == nil {
		return nil, errors.New("Action.StartWorkflow is required when Action is set")
	}
	sw, err := scheduleStartWorkflowActionToThrift(a.StartWorkflow, dc)
	if err != nil {
		return nil, err
	}
	return &shared.ScheduleAction{StartWorkflow: sw}, nil
}

func schedulePoliciesToThrift(p *SchedulePolicies) *shared.SchedulePolicies {
	if p == nil {
		return nil
	}
	return &shared.SchedulePolicies{
		OverlapPolicy:          scheduleOverlapPolicyToThrift(p.OverlapPolicy),
		CatchUpPolicy:          scheduleCatchUpPolicyToThrift(p.CatchUpPolicy),
		CatchUpWindowInSeconds: durationToThriftSeconds(p.CatchUpWindow),
		PauseOnFailure:         p.PauseOnFailure,
		BufferLimit:            common.Int32Ptr(p.BufferLimit),
		ConcurrencyLimit:       common.Int32Ptr(p.ConcurrencyLimit),
	}
}

func scheduleCreateRequestToThrift(domain string, r *CreateScheduleRequest, dc DataConverter) (*shared.CreateScheduleRequest, error) {
	if r == nil {
		return nil, errors.New("Create: request is required")
	}
	if r.ScheduleID == "" {
		return nil, errors.New("Create: ScheduleID is required")
	}
	if r.Spec == nil || r.Spec.CronExpression == "" {
		return nil, errors.New("Create: Spec.CronExpression is required")
	}
	if r.Action == nil || r.Action.StartWorkflow == nil {
		return nil, errors.New("Create: Action.StartWorkflow is required")
	}
	action, err := scheduleActionToThrift(r.Action, dc)
	if err != nil {
		return nil, err
	}
	var memo *shared.Memo
	if len(r.Memo) > 0 {
		fields := make(map[string][]byte, len(r.Memo))
		for k, v := range r.Memo {
			b, encErr := encodeArg(dc, v)
			if encErr != nil {
				return nil, fmt.Errorf("encode memo field %q: %w", k, encErr)
			}
			fields[k] = b
		}
		memo = &shared.Memo{Fields: fields}
	}
	var searchAttr *shared.SearchAttributes
	if len(r.SearchAttributes) > 0 {
		fields := make(map[string][]byte, len(r.SearchAttributes))
		for k, v := range r.SearchAttributes {
			b, encErr := json.Marshal(v)
			if encErr != nil {
				return nil, fmt.Errorf("encode search attribute %q: %w", k, encErr)
			}
			fields[k] = b
		}
		searchAttr = &shared.SearchAttributes{IndexedFields: fields}
	}
	return &shared.CreateScheduleRequest{
		Domain:           common.StringPtr(domain),
		ScheduleId:       common.StringPtr(r.ScheduleID),
		Spec:             scheduleSpecToThrift(r.Spec),
		Action:           action,
		Policies:         schedulePoliciesToThrift(r.Policies),
		Memo:             memo,
		SearchAttributes: searchAttr,
	}, nil
}

func scheduleUpdateRequestToThrift(domain string, r *UpdateScheduleRequest, dc DataConverter) (*shared.UpdateScheduleRequest, error) {
	if r == nil {
		return nil, errors.New("Update: request is required")
	}
	if r.ScheduleID == "" {
		return nil, errors.New("Update: ScheduleID is required")
	}
	if r.Spec == nil && r.Action == nil && r.Policies == nil && len(r.SearchAttributes) == 0 {
		return nil, errors.New("Update: at least one of Spec, Action, Policies, or SearchAttributes must be set")
	}
	if r.Spec != nil && r.Spec.CronExpression == "" {
		return nil, errors.New("Update: Spec.CronExpression is required when Spec is set")
	}
	action, err := scheduleActionToThrift(r.Action, dc)
	if err != nil {
		return nil, err
	}
	var searchAttr *shared.SearchAttributes
	if len(r.SearchAttributes) > 0 {
		fields := make(map[string][]byte, len(r.SearchAttributes))
		for k, v := range r.SearchAttributes {
			b, encErr := json.Marshal(v)
			if encErr != nil {
				return nil, fmt.Errorf("encode search attribute %q: %w", k, encErr)
			}
			fields[k] = b
		}
		searchAttr = &shared.SearchAttributes{IndexedFields: fields}
	}
	return &shared.UpdateScheduleRequest{
		Domain:           common.StringPtr(domain),
		ScheduleId:       common.StringPtr(r.ScheduleID),
		Spec:             scheduleSpecToThrift(r.Spec),
		Action:           action,
		Policies:         schedulePoliciesToThrift(r.Policies),
		SearchAttributes: searchAttr,
	}, nil
}

func backfillRequestToThrift(domain, scheduleID string, r *BackfillRequest) (*shared.BackfillScheduleRequest, error) {
	if scheduleID == "" {
		return nil, errors.New("Backfill: scheduleID is required")
	}
	if r == nil {
		return nil, errors.New("Backfill: request is required")
	}
	if r.StartTime.IsZero() {
		return nil, errors.New("Backfill: StartTime is required")
	}
	if r.EndTime.IsZero() {
		return nil, errors.New("Backfill: EndTime is required")
	}
	if !r.EndTime.After(r.StartTime) {
		return nil, errors.New("Backfill: EndTime must be after StartTime")
	}
	return &shared.BackfillScheduleRequest{
		Domain:        common.StringPtr(domain),
		ScheduleId:    common.StringPtr(scheduleID),
		StartTimeNano: timeToThriftNano(r.StartTime),
		EndTimeNano:   timeToThriftNano(r.EndTime),
		OverlapPolicy: scheduleOverlapPolicyToThrift(r.OverlapPolicy),
		BackfillId:    common.StringPtr(r.BackfillID),
	}, nil
}

// ── Read-side converters (Thrift → SDK) ─────────────────────────────────────

func scheduleSpecFromThrift(spec *shared.ScheduleSpec) *ScheduleSpec {
	if spec == nil {
		return nil
	}
	return &ScheduleSpec{
		CronExpression: spec.GetCronExpression(),
		StartTime:      thriftNanoToTime(spec.StartTimeNano),
		EndTime:        thriftNanoToTime(spec.EndTimeNano),
		Jitter:         thriftSecondsToDuration(spec.JitterInSeconds),
	}
}

func scheduleRetryPolicyFromThrift(p *shared.RetryPolicy) *RetryPolicy {
	if p == nil {
		return nil
	}
	return &RetryPolicy{
		InitialInterval:          thriftSecondsToDuration(p.InitialIntervalInSeconds),
		BackoffCoefficient:       p.GetBackoffCoefficient(),
		MaximumInterval:          thriftSecondsToDuration(p.MaximumIntervalInSeconds),
		ExpirationInterval:       thriftSecondsToDuration(p.ExpirationIntervalInSeconds),
		MaximumAttempts:          p.GetMaximumAttempts(),
		NonRetriableErrorReasons: p.NonRetriableErrorReasons,
	}
}

func scheduleStartWorkflowActionFromThrift(a *shared.ScheduleStartWorkflowAction) *ScheduleStartWorkflowAction {
	if a == nil {
		return nil
	}
	// Memo and SearchAttributes are intentionally not populated here.
	// The server returns them as encoded bytes (map[string][]byte), but the SDK
	// type uses map[string]interface{}, which requires a DataConverter to decode.
	// This converter has no DataConverter, so those fields are omitted on the read path.
	return &ScheduleStartWorkflowAction{
		WorkflowType:                    a.GetWorkflowType().GetName(),
		TaskList:                        a.GetTaskList().GetName(),
		Input:                           a.Input,
		WorkflowIDPrefix:                a.GetWorkflowIdPrefix(),
		ExecutionStartToCloseTimeout:    thriftSecondsToDuration(a.ExecutionStartToCloseTimeoutSeconds),
		DecisionTaskStartToCloseTimeout: thriftSecondsToDuration(a.TaskStartToCloseTimeoutSeconds),
		RetryPolicy:                     scheduleRetryPolicyFromThrift(a.RetryPolicy),
	}
}

func scheduleActionFromThrift(a *shared.ScheduleAction) *ScheduleAction {
	if a == nil {
		return nil
	}
	return &ScheduleAction{
		StartWorkflow: scheduleStartWorkflowActionFromThrift(a.StartWorkflow),
	}
}

func schedulePoliciesFromThrift(p *shared.SchedulePolicies) *SchedulePolicies {
	if p == nil {
		return nil
	}
	return &SchedulePolicies{
		OverlapPolicy:    scheduleOverlapPolicyFromThrift(p.OverlapPolicy),
		CatchUpPolicy:    scheduleCatchUpPolicyFromThrift(p.CatchUpPolicy),
		CatchUpWindow:    thriftSecondsToDuration(p.CatchUpWindowInSeconds),
		PauseOnFailure:   p.PauseOnFailure,
		BufferLimit:      p.GetBufferLimit(),
		ConcurrencyLimit: p.GetConcurrencyLimit(),
	}
}

func schedulePauseInfoFromThrift(pi *shared.SchedulePauseInfo) *SchedulePauseInfo {
	if pi == nil {
		return nil
	}
	return &SchedulePauseInfo{
		Reason:   pi.GetReason(),
		PausedAt: thriftNanoToTime(pi.PausedTimeNano),
		PausedBy: pi.GetPausedBy(),
	}
}

func scheduleStateFromThrift(st *shared.ScheduleState) *ScheduleState {
	if st == nil {
		return nil
	}
	return &ScheduleState{
		Paused:    st.GetPaused(),
		PauseInfo: schedulePauseInfoFromThrift(st.PauseInfo),
	}
}

func backfillInfoFromThrift(b *shared.BackfillInfo) *BackfillInfo {
	if b == nil {
		return nil
	}
	return &BackfillInfo{
		BackfillID:    b.GetBackfillId(),
		StartTime:     thriftNanoToTime(b.StartTimeNano),
		EndTime:       thriftNanoToTime(b.EndTimeNano),
		RunsCompleted: b.GetRunsCompleted(),
		RunsTotal:     b.GetRunsTotal(),
	}
}

func scheduleInfoFromThrift(si *shared.ScheduleInfo) *ScheduleInfo {
	if si == nil {
		return nil
	}
	var ongoing []*BackfillInfo
	if si.OngoingBackfills != nil {
		ongoing = make([]*BackfillInfo, len(si.OngoingBackfills))
		for i, b := range si.OngoingBackfills {
			ongoing[i] = backfillInfoFromThrift(b)
		}
	}
	return &ScheduleInfo{
		LastRunTime:      thriftNanoToTime(si.LastRunTimeNano),
		NextRunTime:      thriftNanoToTime(si.NextRunTimeNano),
		TotalRuns:        si.GetTotalRuns(),
		CreateTime:       thriftNanoToTime(si.CreateTimeNano),
		LastUpdateTime:   thriftNanoToTime(si.LastUpdateTimeNano),
		OngoingBackfills: ongoing,
	}
}

func scheduleListEntryFromThrift(e *shared.ScheduleListEntry) *ScheduleListEntry {
	if e == nil {
		return nil
	}
	return &ScheduleListEntry{
		ScheduleID:     e.GetScheduleId(),
		WorkflowType:   e.GetWorkflowType().GetName(),
		State:          scheduleStateFromThrift(e.State),
		CronExpression: e.GetCronExpression(),
	}
}

func describeScheduleResponseFromThrift(r *shared.DescribeScheduleResponse) *DescribeScheduleResponse {
	if r == nil {
		return nil
	}
	var memo map[string][]byte
	if r.Memo != nil {
		memo = r.Memo.Fields
	}
	var searchAttr map[string][]byte
	if r.SearchAttributes != nil {
		searchAttr = r.SearchAttributes.IndexedFields
	}
	return &DescribeScheduleResponse{
		Spec:             scheduleSpecFromThrift(r.Spec),
		Action:           scheduleActionFromThrift(r.Action),
		Policies:         schedulePoliciesFromThrift(r.Policies),
		State:            scheduleStateFromThrift(r.State),
		Info:             scheduleInfoFromThrift(r.Info),
		Memo:             memo,
		SearchAttributes: searchAttr,
	}
}

func listSchedulesResponseFromThrift(r *shared.ListSchedulesResponse) *ListSchedulesResponse {
	if r == nil {
		return nil
	}
	var entries []*ScheduleListEntry
	if r.Schedules != nil {
		entries = make([]*ScheduleListEntry, len(r.Schedules))
		for i, e := range r.Schedules {
			entries[i] = scheduleListEntryFromThrift(e)
		}
	}
	return &ListSchedulesResponse{
		Schedules:     entries,
		NextPageToken: r.NextPageToken,
	}
}
