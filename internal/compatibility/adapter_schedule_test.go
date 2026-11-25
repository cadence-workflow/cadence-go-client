// Copyright (c) 2021 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package compatibility

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/shared"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
)

// scheduleStub is a minimal stub of apiv1.ScheduleAPIYARPCClient for adapter tests.
type scheduleStub struct{}

func (s *scheduleStub) BackfillSchedule(_ context.Context, _ *apiv1.BackfillScheduleRequest, _ ...yarpc.CallOption) (*apiv1.BackfillScheduleResponse, error) {
	return &apiv1.BackfillScheduleResponse{}, nil
}
func (s *scheduleStub) CreateSchedule(_ context.Context, _ *apiv1.CreateScheduleRequest, _ ...yarpc.CallOption) (*apiv1.CreateScheduleResponse, error) {
	return &apiv1.CreateScheduleResponse{}, nil
}
func (s *scheduleStub) DeleteSchedule(_ context.Context, _ *apiv1.DeleteScheduleRequest, _ ...yarpc.CallOption) (*apiv1.DeleteScheduleResponse, error) {
	return &apiv1.DeleteScheduleResponse{}, nil
}
func (s *scheduleStub) DescribeSchedule(_ context.Context, _ *apiv1.DescribeScheduleRequest, _ ...yarpc.CallOption) (*apiv1.DescribeScheduleResponse, error) {
	return &apiv1.DescribeScheduleResponse{}, nil
}
func (s *scheduleStub) ListSchedules(_ context.Context, _ *apiv1.ListSchedulesRequest, _ ...yarpc.CallOption) (*apiv1.ListSchedulesResponse, error) {
	return &apiv1.ListSchedulesResponse{}, nil
}
func (s *scheduleStub) PauseSchedule(_ context.Context, _ *apiv1.PauseScheduleRequest, _ ...yarpc.CallOption) (*apiv1.PauseScheduleResponse, error) {
	return &apiv1.PauseScheduleResponse{}, nil
}
func (s *scheduleStub) UnpauseSchedule(_ context.Context, _ *apiv1.UnpauseScheduleRequest, _ ...yarpc.CallOption) (*apiv1.UnpauseScheduleResponse, error) {
	return &apiv1.UnpauseScheduleResponse{}, nil
}
func (s *scheduleStub) UpdateSchedule(_ context.Context, _ *apiv1.UpdateScheduleRequest, _ ...yarpc.CallOption) (*apiv1.UpdateScheduleResponse, error) {
	return &apiv1.UpdateScheduleResponse{}, nil
}

func TestAdapterScheduleNilGuard(t *testing.T) {
	a := thrift2protoAdapter{} // schedule is nil
	ctx := context.Background()

	checkErr := func(err error) {
		t.Helper()
		assert.True(t, strings.Contains(err.Error(), "schedule API not configured"))
		var badReqErr *shared.BadRequestError
		assert.True(t, errors.As(err, &badReqErr), "expected *shared.BadRequestError, got %T", err)
	}

	_, err := a.BackfillSchedule(ctx, &shared.BackfillScheduleRequest{})
	checkErr(err)

	_, err = a.CreateSchedule(ctx, &shared.CreateScheduleRequest{})
	checkErr(err)

	_, err = a.DeleteSchedule(ctx, &shared.DeleteScheduleRequest{})
	checkErr(err)

	_, err = a.DescribeSchedule(ctx, &shared.DescribeScheduleRequest{})
	checkErr(err)

	_, err = a.ListSchedules(ctx, &shared.ListSchedulesRequest{})
	checkErr(err)

	_, err = a.PauseSchedule(ctx, &shared.PauseScheduleRequest{})
	checkErr(err)

	_, err = a.UnpauseSchedule(ctx, &shared.UnpauseScheduleRequest{})
	checkErr(err)

	_, err = a.UpdateSchedule(ctx, &shared.UpdateScheduleRequest{})
	checkErr(err)
}

func TestAdapterScheduleDelegates(t *testing.T) {
	a := thrift2protoAdapter{schedule: &scheduleStub{}}
	ctx := context.Background()

	resp, err := a.BackfillSchedule(ctx, &shared.BackfillScheduleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, resp)

	cresp, err := a.CreateSchedule(ctx, &shared.CreateScheduleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, cresp)

	dresp, err := a.DeleteSchedule(ctx, &shared.DeleteScheduleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, dresp)

	desresp, err := a.DescribeSchedule(ctx, &shared.DescribeScheduleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, desresp)

	lresp, err := a.ListSchedules(ctx, &shared.ListSchedulesRequest{})
	require.NoError(t, err)
	assert.NotNil(t, lresp)

	presp, err := a.PauseSchedule(ctx, &shared.PauseScheduleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, presp)

	uresp, err := a.UnpauseSchedule(ctx, &shared.UnpauseScheduleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, uresp)

	upresp, err := a.UpdateSchedule(ctx, &shared.UpdateScheduleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, upresp)
}
