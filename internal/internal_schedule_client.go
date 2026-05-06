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
	"errors"

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
