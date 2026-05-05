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

//go:generate mockery --srcpkg . --name ScheduleClient --output ../mocks

package client

import (
	"context"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"

	"go.uber.org/cadence/internal"
)

// ScheduleClient is the client for managing Cadence schedules within a domain.
type ScheduleClient interface {
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

// NewScheduleClient creates a ScheduleClient that manages schedules in the given domain.
func NewScheduleClient(service apiv1.ScheduleAPIYARPCClient, domain string, options *Options) ScheduleClient {
	return internal.NewScheduleClient(service, domain, options)
}

// compile-time interface compatibility checks
var _ ScheduleClient = internal.ScheduleClient(nil)
var _ internal.ScheduleClient = ScheduleClient(nil)
