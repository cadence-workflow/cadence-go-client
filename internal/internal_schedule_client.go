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

//go:generate mockery --srcpkg . --name ScheduleClient --output ../mocks --with-expecter

package internal

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	shared "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

// ScheduleClient is the client for managing Cadence schedules within a domain.
type ScheduleClient interface {
	// Create creates a new schedule and returns the schedule ID (echoed from the request).
	//
	// Create is not idempotent. If the server processes the request successfully but the
	// response is lost (e.g. due to a network failure), the automatic retry will receive a
	// BadRequestError indicating the schedule already exists. Callers should call Describe
	// after a Create failure to determine whether the schedule was actually created.
	Create(ctx context.Context, request *CreateScheduleRequest) (string, error)

	// Describe returns the current configuration and state of a schedule.
	Describe(ctx context.Context, scheduleID string) (*DescribeScheduleResponse, error)

	// Update reads the schedule's current state, passes it to mutate for in-place editing,
	// and applies the result. Only the fields changed are sent; untouched fields are preserved.
	Update(ctx context.Context, scheduleID string, mutate func(*ScheduleUpdate) error) error

	// Delete deletes a schedule.
	Delete(ctx context.Context, scheduleID string) error

	// Pause pauses a running schedule. reason is recorded in the schedule's pause info.
	Pause(ctx context.Context, scheduleID string, reason string) error

	// Unpause resumes a paused schedule. reason is recorded in the schedule's pause info.
	// catchUpPolicy overrides the schedule's configured catch-up policy for this unpause only;
	// pass ScheduleCatchUpPolicyUnspecified to defer to the schedule's configured policy.
	Unpause(ctx context.Context, scheduleID string, reason string, catchUpPolicy ScheduleCatchUpPolicy) error

	// Backfill triggers workflow runs for a historical time range.
	Backfill(ctx context.Context, scheduleID string, request *BackfillRequest) error

	// List returns all schedules in the domain with optional pagination.
	List(ctx context.Context, pageSize int32, nextPageToken []byte) (*ListSchedulesResponse, error)
}

var _ ScheduleClient = (*scheduleClient)(nil)

type scheduleClient struct {
	workflowService workflowserviceclient.Interface
	domain          string
	identity        string
	dataConverter   DataConverter
	featureFlags    FeatureFlags
}

// ScheduleClient implements Client. It returns a ScheduleClient scoped to this client's domain and connection.
func (wc *workflowClient) ScheduleClient() ScheduleClient {
	return &scheduleClient{
		workflowService: wc.workflowService,
		domain:          wc.domain,
		identity:        wc.identity,
		dataConverter:   wc.dataConverter,
		featureFlags:    wc.featureFlags,
	}
}

func (sc *scheduleClient) Create(ctx context.Context, request *CreateScheduleRequest) (string, error) {
	thriftReq, err := scheduleCreateRequestToThrift(sc.domain, request, sc.dataConverter)
	if err != nil {
		return "", err
	}
	var scheduleID string
	err = retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		resp, rpcErr := sc.workflowService.CreateSchedule(tchCtx, thriftReq, opt...)
		if rpcErr == nil && resp != nil {
			scheduleID = resp.GetScheduleId()
		}
		return rpcErr
	})
	return scheduleID, err
}

func (sc *scheduleClient) Describe(ctx context.Context, scheduleID string) (*DescribeScheduleResponse, error) {
	if scheduleID == "" {
		return nil, errors.New("Describe: scheduleID is required")
	}
	req := &shared.DescribeScheduleRequest{
		Domain:     common.StringPtr(sc.domain),
		ScheduleId: common.StringPtr(scheduleID),
	}
	var thriftResp *shared.DescribeScheduleResponse
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		var rpcErr error
		thriftResp, rpcErr = sc.workflowService.DescribeSchedule(tchCtx, req, opt...)
		return rpcErr
	})
	if err != nil {
		return nil, err
	}
	return describeScheduleResponseFromThrift(thriftResp), nil
}

func (sc *scheduleClient) Update(ctx context.Context, scheduleID string, mutate func(*ScheduleUpdate) error) error {
	if scheduleID == "" {
		return errors.New("Update: scheduleID is required")
	}
	if mutate == nil {
		return errors.New("Update: mutate function is required")
	}

	// Read current state; keep `desc` as the baseline for change detection.
	desc, err := sc.Describe(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("Update: describe failed: %w", err)
	}

	cur := scheduleUpdateFromDescribe(desc, sc.dataConverter)
	if err := mutate(cur); err != nil {
		return err
	}

	// Send only the top-level fields the caller actually changed. Untouched fields are
	// omitted so the server preserves them (top-level merge).
	//
	// A field is sent only when its new value is non-nil/non-empty and differs from the
	// baseline. Top-level fields therefore cannot be cleared via the callback: nil-ing one
	// (e.g. cur.Policies = nil) is treated as "no change", not "clear" — clearing isn't
	// expressible on the wire (an omitted field is preserved by the server's top-level merge).
	req := &shared.UpdateScheduleRequest{
		Domain:     common.StringPtr(sc.domain),
		ScheduleId: common.StringPtr(scheduleID),
	}
	changed := false
	if cur.Spec != nil && !reflect.DeepEqual(cur.Spec, desc.Spec) {
		if cur.Spec.CronExpression == "" {
			return errors.New("Update: Spec.CronExpression must not be empty")
		}
		req.Spec = scheduleSpecToThrift(cur.Spec)
		changed = true
	}
	if cur.Action != nil && !reflect.DeepEqual(cur.Action, desc.Action) {
		action, aerr := scheduleActionDescriptionToThrift(cur.Action)
		if aerr != nil {
			return aerr
		}
		req.Action = action
		changed = true
	}
	if cur.Policies != nil && !reflect.DeepEqual(cur.Policies, desc.Policies) {
		req.Policies = schedulePoliciesToThrift(cur.Policies)
		changed = true
	}
	// SearchAttributes are only sent when non-empty (same can't-clear rationale as above).
	if len(cur.SearchAttributes) > 0 && !reflect.DeepEqual(cur.SearchAttributes, desc.SearchAttributes) {
		req.SearchAttributes = &shared.SearchAttributes{IndexedFields: cur.SearchAttributes}
		changed = true
	}
	if !changed {
		// Nothing changed — don't issue a no-op UpdateSchedule RPC.
		return nil
	}

	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.workflowService.UpdateSchedule(tchCtx, req, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Delete(ctx context.Context, scheduleID string) error {
	if scheduleID == "" {
		return errors.New("Delete: scheduleID is required")
	}
	req := &shared.DeleteScheduleRequest{
		Domain:     common.StringPtr(sc.domain),
		ScheduleId: common.StringPtr(scheduleID),
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.workflowService.DeleteSchedule(tchCtx, req, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Pause(ctx context.Context, scheduleID string, reason string) error {
	if scheduleID == "" {
		return errors.New("Pause: scheduleID is required")
	}
	req := &shared.PauseScheduleRequest{
		Domain:     common.StringPtr(sc.domain),
		ScheduleId: common.StringPtr(scheduleID),
		Reason:     common.StringPtr(reason),
		Identity:   common.StringPtr(sc.identity),
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.workflowService.PauseSchedule(tchCtx, req, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Unpause(ctx context.Context, scheduleID string, reason string, catchUpPolicy ScheduleCatchUpPolicy) error {
	if scheduleID == "" {
		return errors.New("Unpause: scheduleID is required")
	}
	req := &shared.UnpauseScheduleRequest{
		Domain:        common.StringPtr(sc.domain),
		ScheduleId:    common.StringPtr(scheduleID),
		Reason:        common.StringPtr(reason),
		CatchUpPolicy: scheduleCatchUpPolicyToThrift(catchUpPolicy),
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.workflowService.UnpauseSchedule(tchCtx, req, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) Backfill(ctx context.Context, scheduleID string, request *BackfillRequest) error {
	thriftReq, err := backfillRequestToThrift(sc.domain, scheduleID, request)
	if err != nil {
		return err
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, rpcErr := sc.workflowService.BackfillSchedule(tchCtx, thriftReq, opt...)
		return rpcErr
	})
}

func (sc *scheduleClient) List(ctx context.Context, pageSize int32, nextPageToken []byte) (*ListSchedulesResponse, error) {
	req := &shared.ListSchedulesRequest{
		Domain:        common.StringPtr(sc.domain),
		PageSize:      common.Int32Ptr(pageSize),
		NextPageToken: nextPageToken,
	}
	var thriftResp *shared.ListSchedulesResponse
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		var rpcErr error
		thriftResp, rpcErr = sc.workflowService.ListSchedules(tchCtx, req, opt...)
		return rpcErr
	})
	if err != nil {
		return nil, err
	}
	return listSchedulesResponseFromThrift(thriftResp), nil
}
