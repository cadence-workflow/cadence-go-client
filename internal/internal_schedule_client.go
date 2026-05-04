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
	"context"

	"github.com/uber-go/tally"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

type (
	// ScheduleClient is the client for managing Cadence schedules within a domain.
	ScheduleClient interface {
		// Create creates a new schedule. The Domain field of the request is set automatically.
		// Returns the server-assigned schedule ID.
		Create(ctx context.Context, request *apiv1.CreateScheduleRequest) (string, error)

		// Describe returns the current configuration and state of a schedule.
		Describe(ctx context.Context, scheduleID string) (*apiv1.DescribeScheduleResponse, error)

		// Update updates the spec, action, and/or policies of an existing schedule.
		// The Domain field of the request is set automatically.
		Update(ctx context.Context, request *apiv1.UpdateScheduleRequest) error

		// Delete deletes a schedule.
		Delete(ctx context.Context, scheduleID string) error

		// Pause pauses a running schedule. reason is recorded in the schedule's pause info.
		Pause(ctx context.Context, scheduleID string, reason string) error

		// Unpause resumes a paused schedule. The Domain field of the request is set automatically.
		// Use request.CatchUpPolicy to control backfill behavior on resume.
		Unpause(ctx context.Context, request *apiv1.UnpauseScheduleRequest) error

		// Backfill triggers runs for a historical time range.
		// The Domain field of the request is set automatically.
		Backfill(ctx context.Context, request *apiv1.BackfillScheduleRequest) error

		// List returns all schedules in the domain with optional pagination.
		List(ctx context.Context, pageSize int32, nextPageToken []byte) (*apiv1.ListSchedulesResponse, error)
	}
)

var _ ScheduleClient = (*scheduleClient)(nil)

type scheduleClient struct {
	scheduleService apiv1.ScheduleAPIYARPCClient
	domain          string
	metricsScope    tally.Scope
	identity        string
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
	var metricsScope tally.Scope
	if options != nil {
		metricsScope = options.MetricsScope
	}
	metricsScope = tagScope(metricsScope, tagDomain, domain, clientImplHeaderName, clientImplHeaderValue, callerTypeHeaderName, callerTypeHeaderValue)
	return &scheduleClient{
		scheduleService: service,
		domain:          domain,
		metricsScope:    metricsScope,
		identity:        identity,
		featureFlags:    getFeatureFlags(options),
	}
}

func (sc *scheduleClient) Create(ctx context.Context, request *apiv1.CreateScheduleRequest) (string, error) {
	request.Domain = sc.domain
	var scheduleID string
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		resp, err := sc.scheduleService.CreateSchedule(tchCtx, request, opt...)
		if err == nil && resp != nil {
			scheduleID = resp.ScheduleId
		}
		return err
	})
	return scheduleID, err
}

func (sc *scheduleClient) Describe(ctx context.Context, scheduleID string) (*apiv1.DescribeScheduleResponse, error) {
	request := &apiv1.DescribeScheduleRequest{
		Domain:     sc.domain,
		ScheduleId: scheduleID,
	}
	var response *apiv1.DescribeScheduleResponse
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		var err error
		response, err = sc.scheduleService.DescribeSchedule(tchCtx, request, opt...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (sc *scheduleClient) Update(ctx context.Context, request *apiv1.UpdateScheduleRequest) error {
	request.Domain = sc.domain
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, err := sc.scheduleService.UpdateSchedule(tchCtx, request, opt...)
		return err
	})
}

func (sc *scheduleClient) Delete(ctx context.Context, scheduleID string) error {
	request := &apiv1.DeleteScheduleRequest{
		Domain:     sc.domain,
		ScheduleId: scheduleID,
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, err := sc.scheduleService.DeleteSchedule(tchCtx, request, opt...)
		return err
	})
}

func (sc *scheduleClient) Pause(ctx context.Context, scheduleID string, reason string) error {
	request := &apiv1.PauseScheduleRequest{
		Domain:     sc.domain,
		ScheduleId: scheduleID,
		Reason:     reason,
		Identity:   sc.identity,
	}
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, err := sc.scheduleService.PauseSchedule(tchCtx, request, opt...)
		return err
	})
}

func (sc *scheduleClient) Unpause(ctx context.Context, request *apiv1.UnpauseScheduleRequest) error {
	request.Domain = sc.domain
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, err := sc.scheduleService.UnpauseSchedule(tchCtx, request, opt...)
		return err
	})
}

func (sc *scheduleClient) Backfill(ctx context.Context, request *apiv1.BackfillScheduleRequest) error {
	request.Domain = sc.domain
	return retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		_, err := sc.scheduleService.BackfillSchedule(tchCtx, request, opt...)
		return err
	})
}

func (sc *scheduleClient) List(ctx context.Context, pageSize int32, nextPageToken []byte) (*apiv1.ListSchedulesResponse, error) {
	request := &apiv1.ListSchedulesRequest{
		Domain:        sc.domain,
		PageSize:      pageSize,
		NextPageToken: nextPageToken,
	}
	var response *apiv1.ListSchedulesResponse
	err := retryWhileTransientError(ctx, func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sc.featureFlags)
		defer cancel()
		var err error
		response, err = sc.scheduleService.ListSchedules(tchCtx, request, opt...)
		return err
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}
