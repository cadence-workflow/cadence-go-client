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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/metrics"
)

var errNonRetryable = &s.BadRequestError{Message: "bad request"}

const testScheduleID = "test-schedule-id"

type scheduleClientTestData struct {
	sc          ScheduleClient
	mockService *MockScheduleAPIYARPCClient
}

func newScheduleClientTestData(t *testing.T) *scheduleClientTestData {
	ctrl := gomock.NewController(t)
	mockService := NewMockScheduleAPIYARPCClient(ctrl)
	sc := NewScheduleClient(mockService, testDomain, &ClientOptions{
		MetricsScope: metrics.NewTaggedScope(nil),
		Identity:     identity,
	})
	return &scheduleClientTestData{sc: sc, mockService: mockService}
}

func TestScheduleClient_Create(t *testing.T) {
	testcases := []struct {
		name               string
		rpcError           error
		response           *apiv1.CreateScheduleResponse
		expectedScheduleID string
	}{
		{
			name:               "success",
			rpcError:           nil,
			response:           &apiv1.CreateScheduleResponse{ScheduleId: testScheduleID},
			expectedScheduleID: testScheduleID,
		},
		{
			name:     "rpc failure",
			rpcError: errNonRetryable,
			response: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			request := &apiv1.CreateScheduleRequest{ScheduleId: testScheduleID}

			td.mockService.EXPECT().
				CreateSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.CreateScheduleRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
				}).
				Return(tt.response, tt.rpcError)

			scheduleID, err := td.sc.Create(context.Background(), request)
			assert.Equal(t, tt.rpcError, err)
			assert.Equal(t, tt.expectedScheduleID, scheduleID)
		})
	}
}

func TestScheduleClient_Describe(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
		response *apiv1.DescribeScheduleResponse
	}{
		{
			name:     "success",
			rpcError: nil,
			response: &apiv1.DescribeScheduleResponse{},
		},
		{
			name:     "rpc failure",
			rpcError: errNonRetryable,
			response: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.DescribeScheduleRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
					assert.Equal(t, testScheduleID, req.ScheduleId)
				}).
				Return(tt.response, tt.rpcError)

			resp, err := td.sc.Describe(context.Background(), testScheduleID)
			assert.Equal(t, tt.rpcError, err)
			assert.Equal(t, tt.response, resp)
		})
	}
}

func TestScheduleClient_Update(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success", rpcError: nil},
		{name: "rpc failure", rpcError: errNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			request := &apiv1.UpdateScheduleRequest{ScheduleId: testScheduleID}

			td.mockService.EXPECT().
				UpdateSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.UpdateScheduleRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
				}).
				Return(&apiv1.UpdateScheduleResponse{}, tt.rpcError)

			err := td.sc.Update(context.Background(), request)
			assert.Equal(t, tt.rpcError, err)
		})
	}
}

func TestScheduleClient_Delete(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success", rpcError: nil},
		{name: "rpc failure", rpcError: errNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				DeleteSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.DeleteScheduleRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
					assert.Equal(t, testScheduleID, req.ScheduleId)
				}).
				Return(&apiv1.DeleteScheduleResponse{}, tt.rpcError)

			err := td.sc.Delete(context.Background(), testScheduleID)
			assert.Equal(t, tt.rpcError, err)
		})
	}
}

func TestScheduleClient_Pause(t *testing.T) {
	const reason = "pausing for maintenance"
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success", rpcError: nil},
		{name: "rpc failure", rpcError: errNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				PauseSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.PauseScheduleRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
					assert.Equal(t, testScheduleID, req.ScheduleId)
					assert.Equal(t, reason, req.Reason)
					assert.Equal(t, identity, req.Identity)
				}).
				Return(&apiv1.PauseScheduleResponse{}, tt.rpcError)

			err := td.sc.Pause(context.Background(), testScheduleID, reason)
			assert.Equal(t, tt.rpcError, err)
		})
	}
}

func TestScheduleClient_Unpause(t *testing.T) {
	const reason = "resuming after maintenance"
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success", rpcError: nil},
		{name: "rpc failure", rpcError: errNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			request := &apiv1.UnpauseScheduleRequest{
				ScheduleId: testScheduleID,
				Reason:     reason,
			}

			td.mockService.EXPECT().
				UnpauseSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.UnpauseScheduleRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
					assert.Equal(t, testScheduleID, req.ScheduleId)
					assert.Equal(t, reason, req.Reason)
				}).
				Return(&apiv1.UnpauseScheduleResponse{}, tt.rpcError)

			err := td.sc.Unpause(context.Background(), request)
			assert.Equal(t, tt.rpcError, err)
		})
	}
}

func TestScheduleClient_Backfill(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success", rpcError: nil},
		{name: "rpc failure", rpcError: errNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			request := &apiv1.BackfillScheduleRequest{ScheduleId: testScheduleID}

			td.mockService.EXPECT().
				BackfillSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.BackfillScheduleRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
				}).
				Return(&apiv1.BackfillScheduleResponse{}, tt.rpcError)

			err := td.sc.Backfill(context.Background(), request)
			assert.Equal(t, tt.rpcError, err)
		})
	}
}

func TestScheduleClient_List(t *testing.T) {
	testcases := []struct {
		name          string
		rpcError      error
		response      *apiv1.ListSchedulesResponse
		pageSize      int32
		nextPageToken []byte
	}{
		{
			name:          "success",
			rpcError:      nil,
			response:      &apiv1.ListSchedulesResponse{},
			pageSize:      10,
			nextPageToken: []byte("token"),
		},
		{
			name:     "rpc failure",
			rpcError: errNonRetryable,
			response: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				ListSchedules(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.ListSchedulesRequest, _ ...interface{}) {
					assert.Equal(t, testDomain, req.Domain)
					assert.Equal(t, tt.pageSize, req.PageSize)
					assert.Equal(t, tt.nextPageToken, req.NextPageToken)
				}).
				Return(tt.response, tt.rpcError)

			resp, err := td.sc.List(context.Background(), tt.pageSize, tt.nextPageToken)
			assert.Equal(t, tt.rpcError, err)
			assert.Equal(t, tt.response, resp)
		})
	}
}
