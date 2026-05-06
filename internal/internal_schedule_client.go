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

//go:generate mockery --srcpkg github.com/uber/cadence-idl/go/proto/api/v1 --name ScheduleAPIYARPCClient --output . --outpkg internal --filename mock_schedule_api_yarpc_client_test.go --with-expecter

package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

// ScheduleClient is the client for managing Cadence schedules within a domain.
type ScheduleClient interface {
	// Create creates a new schedule and returns the server-assigned schedule ID.
	Create(ctx context.Context, request *CreateScheduleRequest) (string, error)

	// Describe returns the current configuration and state of a schedule.
	Describe(ctx context.Context, scheduleID string) (*DescribeScheduleResponse, error)

	// Update replaces the spec, action, and/or policies of an existing schedule.
	Update(ctx context.Context, request *UpdateScheduleRequest) error

	// Delete deletes a schedule.
	Delete(ctx context.Context, scheduleID string) error

	// Pause pauses a running schedule. reason is recorded in the schedule's pause info.
	Pause(ctx context.Context, scheduleID string, reason string) error

	// Unpause resumes a paused schedule. reason is recorded in the schedule's pause info.
	Unpause(ctx context.Context, scheduleID string, reason string) error

	// Backfill triggers workflow runs for a historical time range.
	Backfill(ctx context.Context, scheduleID string, request *BackfillRequest) error

	// List returns all schedules in the domain with optional pagination.
	List(ctx context.Context, pageSize int32, nextPageToken []byte) (*ListSchedulesResponse, error)
}

var _ ScheduleClient = (*scheduleClient)(nil)

type scheduleClient struct {
	scheduleService apiv1.ScheduleAPIYARPCClient
	domain          string
	identity        string
	dataConverter   DataConverter
	featureFlags    FeatureFlags
}

// NewScheduleClient creates a ScheduleClient that manages schedules in the given domain.
func NewScheduleClient(service apiv1.ScheduleAPIYARPCClient, domain string, options *ClientOptions) ScheduleClient {
	var identity string
	if options == nil || options.Identity == "" {
		identity = getWorkerIdentity("")
	} else {
		identity = options.Identity
	}
	var dc DataConverter
	if options != nil && options.DataConverter != nil {
		dc = options.DataConverter
	} else {
		dc = getDefaultDataConverter()
	}
	return &scheduleClient{
		scheduleService: service,
		domain:          domain,
		identity:        identity,
		dataConverter:   dc,
		featureFlags:    getFeatureFlags(options),
	}
}

func (sc *scheduleClient) Create(ctx context.Context, request *CreateScheduleRequest) (string, error) {
	protoReq, err := scheduleCreateRequestToProto(sc.domain, request, sc.dataConverter)
	if err != nil {
		return "", err
	}
	var scheduleID string
	err = retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		resp, rpcErr := sc.scheduleService.CreateSchedule(tchCtx, protoReq, opt...)
		if rpcErr == nil && resp != nil {
			scheduleID = resp.ScheduleId
		}
		return rpcErr
	})
	return scheduleID, err
}

func (sc *scheduleClient) Describe(ctx context.Context, scheduleID string) (*DescribeScheduleResponse, error) {
	if scheduleID == "" {
		return nil, errors.New("Describe: scheduleID is required")
	}
	req := &apiv1.DescribeScheduleRequest{
		Domain:     sc.domain,
		ScheduleId: scheduleID,
	}
	var protoResp *apiv1.DescribeScheduleResponse
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		var rpcErr error
		protoResp, rpcErr = sc.scheduleService.DescribeSchedule(tchCtx, req, opt...)
		return rpcErr
	})
	if err != nil {
		return nil, err
	}
	return describeScheduleResponseFromProto(protoResp), nil
}

func (sc *scheduleClient) Update(ctx context.Context, request *UpdateScheduleRequest) error {
	protoReq, err := scheduleUpdateRequestToProto(sc.domain, request, sc.dataConverter)
	if err != nil {
		return err
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.scheduleService.UpdateSchedule(tchCtx, protoReq, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Delete(ctx context.Context, scheduleID string) error {
	if scheduleID == "" {
		return errors.New("Delete: scheduleID is required")
	}
	req := &apiv1.DeleteScheduleRequest{
		Domain:     sc.domain,
		ScheduleId: scheduleID,
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.scheduleService.DeleteSchedule(tchCtx, req, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Pause(ctx context.Context, scheduleID string, reason string) error {
	if scheduleID == "" {
		return errors.New("Pause: scheduleID is required")
	}
	req := &apiv1.PauseScheduleRequest{
		Domain:     sc.domain,
		ScheduleId: scheduleID,
		Reason:     reason,
		Identity:   sc.identity,
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.scheduleService.PauseSchedule(tchCtx, req, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Unpause(ctx context.Context, scheduleID string, reason string) error {
	if scheduleID == "" {
		return errors.New("Unpause: scheduleID is required")
	}
	req := &apiv1.UnpauseScheduleRequest{
		Domain:     sc.domain,
		ScheduleId: scheduleID,
		Reason:     reason,
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.scheduleService.UnpauseSchedule(tchCtx, req, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Backfill(ctx context.Context, scheduleID string, request *BackfillRequest) error {
	protoReq, err := backfillRequestToProto(sc.domain, scheduleID, request)
	if err != nil {
		return err
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.scheduleService.BackfillSchedule(tchCtx, protoReq, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) List(ctx context.Context, pageSize int32, nextPageToken []byte) (*ListSchedulesResponse, error) {
	req := &apiv1.ListSchedulesRequest{
		Domain:        sc.domain,
		PageSize:      pageSize,
		NextPageToken: nextPageToken,
	}
	var protoResp *apiv1.ListSchedulesResponse
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		var rpcErr error
		protoResp, rpcErr = sc.scheduleService.ListSchedules(tchCtx, req, opt...)
		return rpcErr
	})
	if err != nil {
		return nil, err
	}
	return listSchedulesResponseFromProto(protoResp), nil
}

// ── Enum converters ──────────────────────────────────────────────────────────

func (p ScheduleOverlapPolicy) toProto() apiv1.ScheduleOverlapPolicy {
	switch p {
	case ScheduleOverlapPolicySkipNew:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_SKIP_NEW
	case ScheduleOverlapPolicyBuffer:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_BUFFER
	case ScheduleOverlapPolicyConcurrent:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CONCURRENT
	case ScheduleOverlapPolicyCancelPrevious:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CANCEL_PREVIOUS
	case ScheduleOverlapPolicyTerminatePrevious:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_TERMINATE_PREVIOUS
	default:
		return apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_INVALID
	}
}

func scheduleOverlapPolicyFromProto(p apiv1.ScheduleOverlapPolicy) ScheduleOverlapPolicy {
	switch p {
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_SKIP_NEW:
		return ScheduleOverlapPolicySkipNew
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_BUFFER:
		return ScheduleOverlapPolicyBuffer
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CONCURRENT:
		return ScheduleOverlapPolicyConcurrent
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_CANCEL_PREVIOUS:
		return ScheduleOverlapPolicyCancelPrevious
	case apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_TERMINATE_PREVIOUS:
		return ScheduleOverlapPolicyTerminatePrevious
	default:
		return ScheduleOverlapPolicyUnspecified
	}
}

func (p ScheduleCatchUpPolicy) toProto() apiv1.ScheduleCatchUpPolicy {
	switch p {
	case ScheduleCatchUpPolicySkip:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_SKIP
	case ScheduleCatchUpPolicyOne:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ONE
	case ScheduleCatchUpPolicyAll:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ALL
	default:
		return apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_INVALID
	}
}

func scheduleCatchUpPolicyFromProto(p apiv1.ScheduleCatchUpPolicy) ScheduleCatchUpPolicy {
	switch p {
	case apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_SKIP:
		return ScheduleCatchUpPolicySkip
	case apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ONE:
		return ScheduleCatchUpPolicyOne
	case apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ALL:
		return ScheduleCatchUpPolicyAll
	default:
		return ScheduleCatchUpPolicyUnspecified
	}
}

// ── Memo / SearchAttributes ──────────────────────────────────────────────────

func scheduleMemoToProto(input map[string]interface{}, dc DataConverter) (*apiv1.Memo, error) {
	if len(input) == 0 {
		return nil, nil
	}
	fields := make(map[string]*apiv1.Payload, len(input))
	for k, v := range input {
		b, err := encodeArg(dc, v)
		if err != nil {
			return nil, fmt.Errorf("schedule: encode memo field %q: %w", k, err)
		}
		fields[k] = &apiv1.Payload{Data: b}
	}
	return &apiv1.Memo{Fields: fields}, nil
}

func scheduleSearchAttributesToProto(input map[string]interface{}) (*apiv1.SearchAttributes, error) {
	if len(input) == 0 {
		return nil, nil
	}
	fields := make(map[string]*apiv1.Payload, len(input))
	for k, v := range input {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("schedule: encode search attribute %q: %w", k, err)
		}
		fields[k] = &apiv1.Payload{Data: b}
	}
	return &apiv1.SearchAttributes{IndexedFields: fields}, nil
}

func scheduleMemoFromProto(m *apiv1.Memo) map[string][]byte {
	if m == nil || len(m.Fields) == 0 {
		return nil
	}
	result := make(map[string][]byte, len(m.Fields))
	for k, p := range m.Fields {
		if p != nil {
			result[k] = p.Data
		}
	}
	return result
}

func scheduleSearchAttributesFromProto(sa *apiv1.SearchAttributes) map[string][]byte {
	if sa == nil || len(sa.IndexedFields) == 0 {
		return nil
	}
	result := make(map[string][]byte, len(sa.IndexedFields))
	for k, p := range sa.IndexedFields {
		if p != nil {
			result[k] = p.Data
		}
	}
	return result
}

// ── Write-side converters (SDK → proto) ─────────────────────────────────────

func scheduleSpecToProto(s *ScheduleSpec) *apiv1.ScheduleSpec {
	if s == nil {
		return nil
	}
	return &apiv1.ScheduleSpec{
		CronExpression: s.CronExpression,
		StartTime:      toProtoTimestamp(s.StartTime),
		EndTime:        toProtoTimestamp(s.EndTime),
		Jitter:         toProtoDuration(s.Jitter),
	}
}

func scheduleStartWorkflowActionToProto(a *ScheduleStartWorkflowAction, dc DataConverter) (*apiv1.ScheduleAction_StartWorkflowAction, error) {
	if a == nil {
		return nil, nil
	}
	memo, err := scheduleMemoToProto(a.Memo, dc)
	if err != nil {
		return nil, err
	}
	searchAttr, err := scheduleSearchAttributesToProto(a.SearchAttributes)
	if err != nil {
		return nil, err
	}
	var input *apiv1.Payload
	if len(a.Input) > 0 {
		input = &apiv1.Payload{Data: a.Input}
	}
	return &apiv1.ScheduleAction_StartWorkflowAction{
		WorkflowType:                 &apiv1.WorkflowType{Name: a.WorkflowType},
		TaskList:                     &apiv1.TaskList{Name: a.TaskList},
		Input:                        input,
		WorkflowIdPrefix:             a.WorkflowIDPrefix,
		ExecutionStartToCloseTimeout: toProtoDuration(a.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeout:      toProtoDuration(a.DecisionTaskStartToCloseTimeout),
		RetryPolicy:                  scheduleRetryPolicyToProto(a.RetryPolicy),
		Memo:                         memo,
		SearchAttributes:             searchAttr,
	}, nil
}

func scheduleActionToProto(a *ScheduleAction, dc DataConverter) (*apiv1.ScheduleAction, error) {
	if a == nil {
		return nil, nil
	}
	sw, err := scheduleStartWorkflowActionToProto(a.StartWorkflow, dc)
	if err != nil {
		return nil, err
	}
	return &apiv1.ScheduleAction{StartWorkflow: sw}, nil
}

func schedulePoliciesToProto(p *SchedulePolicies) *apiv1.SchedulePolicies {
	if p == nil {
		return nil
	}
	return &apiv1.SchedulePolicies{
		OverlapPolicy:    p.OverlapPolicy.toProto(),
		CatchUpPolicy:    p.CatchUpPolicy.toProto(),
		CatchUpWindow:    toProtoDuration(p.CatchUpWindow),
		PauseOnFailure:   p.PauseOnFailure,
		BufferLimit:      p.BufferLimit,
		ConcurrencyLimit: p.ConcurrencyLimit,
	}
}

// scheduleRetryPolicyToProto converts the SDK RetryPolicy to the proto equivalent.
// Targets apiv1.RetryPolicy (proto), not s.RetryPolicy (Thrift); field names and units differ.
func scheduleRetryPolicyToProto(p *RetryPolicy) *apiv1.RetryPolicy {
	if p == nil {
		return nil
	}
	return &apiv1.RetryPolicy{
		InitialInterval:          toProtoDuration(p.InitialInterval),
		BackoffCoefficient:       p.BackoffCoefficient,
		MaximumInterval:          toProtoDuration(p.MaximumInterval),
		ExpirationInterval:       toProtoDuration(p.ExpirationInterval),
		MaximumAttempts:          p.MaximumAttempts,
		NonRetryableErrorReasons: p.NonRetriableErrorReasons,
	}
}

func scheduleCreateRequestToProto(domain string, r *CreateScheduleRequest, dc DataConverter) (*apiv1.CreateScheduleRequest, error) {
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
	action, err := scheduleActionToProto(r.Action, dc)
	if err != nil {
		return nil, err
	}
	memo, err := scheduleMemoToProto(r.Memo, dc)
	if err != nil {
		return nil, err
	}
	searchAttr, err := scheduleSearchAttributesToProto(r.SearchAttributes)
	if err != nil {
		return nil, err
	}
	return &apiv1.CreateScheduleRequest{
		Domain:           domain,
		ScheduleId:       r.ScheduleID,
		Spec:             scheduleSpecToProto(r.Spec),
		Action:           action,
		Policies:         schedulePoliciesToProto(r.Policies),
		Memo:             memo,
		SearchAttributes: searchAttr,
	}, nil
}

func scheduleUpdateRequestToProto(domain string, r *UpdateScheduleRequest, dc DataConverter) (*apiv1.UpdateScheduleRequest, error) {
	if r == nil {
		return nil, errors.New("Update: request is required")
	}
	if r.ScheduleID == "" {
		return nil, errors.New("Update: ScheduleID is required")
	}
	if r.Spec == nil && r.Action == nil && r.Policies == nil {
		return nil, errors.New("Update: at least one of Spec, Action, or Policies must be set")
	}
	action, err := scheduleActionToProto(r.Action, dc)
	if err != nil {
		return nil, err
	}
	searchAttr, err := scheduleSearchAttributesToProto(r.SearchAttributes)
	if err != nil {
		return nil, err
	}
	return &apiv1.UpdateScheduleRequest{
		Domain:           domain,
		ScheduleId:       r.ScheduleID,
		Spec:             scheduleSpecToProto(r.Spec),
		Action:           action,
		Policies:         schedulePoliciesToProto(r.Policies),
		SearchAttributes: searchAttr,
	}, nil
}

// ── Read-side converters (proto → SDK) ──────────────────────────────────────

func scheduleSpecFromProto(s *apiv1.ScheduleSpec) *ScheduleSpec {
	if s == nil {
		return nil
	}
	return &ScheduleSpec{
		CronExpression: s.GetCronExpression(),
		StartTime:      fromProtoTimestamp(s.GetStartTime()),
		EndTime:        fromProtoTimestamp(s.GetEndTime()),
		Jitter:         fromProtoDuration(s.GetJitter()),
	}
}

func scheduleStartWorkflowActionFromProto(a *apiv1.ScheduleAction_StartWorkflowAction) *ScheduleStartWorkflowAction {
	if a == nil {
		return nil
	}
	var input []byte
	if p := a.GetInput(); p != nil {
		input = p.Data
	}
	return &ScheduleStartWorkflowAction{
		WorkflowType:                    a.GetWorkflowType().GetName(),
		TaskList:                        a.GetTaskList().GetName(),
		Input:                           input,
		WorkflowIDPrefix:                a.GetWorkflowIdPrefix(),
		ExecutionStartToCloseTimeout:    fromProtoDuration(a.GetExecutionStartToCloseTimeout()),
		DecisionTaskStartToCloseTimeout: fromProtoDuration(a.GetTaskStartToCloseTimeout()),
		RetryPolicy:                     scheduleRetryPolicyFromProto(a.GetRetryPolicy()),
		Memo:                            nil,
		SearchAttributes:                nil,
	}
}

func scheduleActionFromProto(a *apiv1.ScheduleAction) *ScheduleAction {
	if a == nil {
		return nil
	}
	return &ScheduleAction{
		StartWorkflow: scheduleStartWorkflowActionFromProto(a.GetStartWorkflow()),
	}
}

func schedulePoliciesFromProto(p *apiv1.SchedulePolicies) *SchedulePolicies {
	if p == nil {
		return nil
	}
	return &SchedulePolicies{
		OverlapPolicy:    scheduleOverlapPolicyFromProto(p.GetOverlapPolicy()),
		CatchUpPolicy:    scheduleCatchUpPolicyFromProto(p.GetCatchUpPolicy()),
		CatchUpWindow:    fromProtoDuration(p.GetCatchUpWindow()),
		PauseOnFailure:   p.GetPauseOnFailure(),
		BufferLimit:      p.GetBufferLimit(),
		ConcurrencyLimit: p.GetConcurrencyLimit(),
	}
}

func schedulePauseInfoFromProto(pi *apiv1.SchedulePauseInfo) *SchedulePauseInfo {
	if pi == nil {
		return nil
	}
	return &SchedulePauseInfo{
		Reason:   pi.GetReason(),
		PausedAt: fromProtoTimestamp(pi.GetPausedAt()),
		PausedBy: pi.GetPausedBy(),
	}
}

func scheduleStateFromProto(s *apiv1.ScheduleState) *ScheduleState {
	if s == nil {
		return nil
	}
	return &ScheduleState{
		Paused:    s.GetPaused(),
		PauseInfo: schedulePauseInfoFromProto(s.GetPauseInfo()),
	}
}

func backfillInfoFromProto(b *apiv1.BackfillInfo) *BackfillInfo {
	if b == nil {
		return nil
	}
	return &BackfillInfo{
		BackfillID:    b.GetBackfillId(),
		StartTime:     fromProtoTimestamp(b.GetStartTime()),
		EndTime:       fromProtoTimestamp(b.GetEndTime()),
		RunsCompleted: b.GetRunsCompleted(),
		RunsTotal:     b.GetRunsTotal(),
	}
}

func scheduleInfoFromProto(si *apiv1.ScheduleInfo) *ScheduleInfo {
	if si == nil {
		return nil
	}
	ongoing := make([]*BackfillInfo, 0, len(si.GetOngoingBackfills()))
	for _, b := range si.GetOngoingBackfills() {
		ongoing = append(ongoing, backfillInfoFromProto(b))
	}
	return &ScheduleInfo{
		LastRunTime:      fromProtoTimestamp(si.GetLastRunTime()),
		NextRunTime:      fromProtoTimestamp(si.GetNextRunTime()),
		TotalRuns:        si.GetTotalRuns(),
		CreateTime:       fromProtoTimestamp(si.GetCreateTime()),
		LastUpdateTime:   fromProtoTimestamp(si.GetLastUpdateTime()),
		OngoingBackfills: ongoing,
	}
}

func scheduleRetryPolicyFromProto(p *apiv1.RetryPolicy) *RetryPolicy {
	if p == nil {
		return nil
	}
	return &RetryPolicy{
		InitialInterval:          fromProtoDuration(p.GetInitialInterval()),
		BackoffCoefficient:       p.GetBackoffCoefficient(),
		MaximumInterval:          fromProtoDuration(p.GetMaximumInterval()),
		ExpirationInterval:       fromProtoDuration(p.GetExpirationInterval()),
		MaximumAttempts:          p.GetMaximumAttempts(),
		NonRetriableErrorReasons: p.GetNonRetryableErrorReasons(),
	}
}

func scheduleListEntryFromProto(e *apiv1.ScheduleListEntry) *ScheduleListEntry {
	if e == nil {
		return nil
	}
	return &ScheduleListEntry{
		ScheduleID:     e.GetScheduleId(),
		WorkflowType:   e.GetWorkflowType().GetName(),
		State:          scheduleStateFromProto(e.GetState()),
		CronExpression: e.GetCronExpression(),
	}
}

func describeScheduleResponseFromProto(r *apiv1.DescribeScheduleResponse) *DescribeScheduleResponse {
	if r == nil {
		return nil
	}
	return &DescribeScheduleResponse{
		Spec:             scheduleSpecFromProto(r.GetSpec()),
		Action:           scheduleActionFromProto(r.GetAction()),
		Policies:         schedulePoliciesFromProto(r.GetPolicies()),
		State:            scheduleStateFromProto(r.GetState()),
		Info:             scheduleInfoFromProto(r.GetInfo()),
		Memo:             scheduleMemoFromProto(r.GetMemo()),
		SearchAttributes: scheduleSearchAttributesFromProto(r.GetSearchAttributes()),
	}
}

func listSchedulesResponseFromProto(r *apiv1.ListSchedulesResponse) *ListSchedulesResponse {
	if r == nil {
		return nil
	}
	entries := make([]*ScheduleListEntry, 0, len(r.GetSchedules()))
	for _, e := range r.GetSchedules() {
		entries = append(entries, scheduleListEntryFromProto(e))
	}
	return &ListSchedulesResponse{
		Schedules:     entries,
		NextPageToken: r.GetNextPageToken(),
	}
}

func backfillRequestToProto(domain, scheduleID string, r *BackfillRequest) (*apiv1.BackfillScheduleRequest, error) {
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
	return &apiv1.BackfillScheduleRequest{
		Domain:        domain,
		ScheduleId:    scheduleID,
		StartTime:     toProtoTimestamp(r.StartTime),
		EndTime:       toProtoTimestamp(r.EndTime),
		OverlapPolicy: r.OverlapPolicy.toProto(),
		BackfillId:    r.BackfillID,
	}, nil
}
