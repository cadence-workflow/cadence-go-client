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
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/metrics"
)

const (
	scheduleTestDomain = "test-domain"
	scheduleTestID     = "test-schedule-id"
	scheduleTestIdent  = "test-identity"
)

var errScheduleNonRetryable = &s.BadRequestError{Message: "bad request"}

type scheduleClientTestData struct {
	sc          ScheduleClient
	mockService *MockScheduleAPIYARPCClient
}

func newScheduleClientTestData(t *testing.T) *scheduleClientTestData {
	ctrl := gomock.NewController(t)
	mockService := NewMockScheduleAPIYARPCClient(ctrl)
	sc := NewScheduleClient(mockService, scheduleTestDomain, &ClientOptions{
		MetricsScope: metrics.NewTaggedScope(nil),
		Identity:     scheduleTestIdent,
	})
	return &scheduleClientTestData{sc: sc, mockService: mockService}
}

func TestScheduleClient_Create(t *testing.T) {
	testcases := []struct {
		name               string
		rpcError           error
		rpcResponse        *apiv1.CreateScheduleResponse
		expectedScheduleID string
	}{
		{
			name:               "success",
			rpcResponse:        &apiv1.CreateScheduleResponse{ScheduleId: scheduleTestID},
			expectedScheduleID: scheduleTestID,
		},
		{
			name:     "rpc failure",
			rpcError: errScheduleNonRetryable,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			request := &CreateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType: "my-workflow",
						TaskList:     "my-task-list",
					},
				},
			}

			td.mockService.EXPECT().
				CreateSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.CreateScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
					assert.Equal(t, "0 * * * *", req.Spec.CronExpression)
					assert.Equal(t, "my-workflow", req.Action.StartWorkflow.WorkflowType.Name)
				}).
				Return(tt.rpcResponse, tt.rpcError)

			scheduleID, err := td.sc.Create(context.Background(), request)
			assert.Equal(t, tt.rpcError, err)
			assert.Equal(t, tt.expectedScheduleID, scheduleID)
		})
	}
}

func TestScheduleClient_Describe(t *testing.T) {
	testcases := []struct {
		name        string
		rpcError    error
		rpcResponse *apiv1.DescribeScheduleResponse
	}{
		{
			name:        "success",
			rpcResponse: &apiv1.DescribeScheduleResponse{},
		},
		{
			name:     "rpc failure",
			rpcError: errScheduleNonRetryable,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.DescribeScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
				}).
				Return(tt.rpcResponse, tt.rpcError)

			resp, err := td.sc.Describe(context.Background(), scheduleTestID)
			assert.Equal(t, tt.rpcError, err)
			if tt.rpcError == nil {
				require.NotNil(t, resp)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func TestScheduleClient_Update(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success"},
		{name: "rpc failure", rpcError: errScheduleNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			request := &UpdateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 2 * * *"},
			}

			td.mockService.EXPECT().
				UpdateSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.UpdateScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
					assert.Equal(t, "0 2 * * *", req.Spec.CronExpression)
				}).
				Return(&apiv1.UpdateScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Update(context.Background(), request))
		})
	}
}

func TestScheduleClient_Delete(t *testing.T) {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success"},
		{name: "rpc failure", rpcError: errScheduleNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				DeleteSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.DeleteScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
				}).
				Return(&apiv1.DeleteScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Delete(context.Background(), scheduleTestID))
		})
	}
}

func TestScheduleClient_Pause(t *testing.T) {
	const reason = "pausing for maintenance"
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success"},
		{name: "rpc failure", rpcError: errScheduleNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				PauseSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.PauseScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
					assert.Equal(t, reason, req.Reason)
					assert.Equal(t, scheduleTestIdent, req.Identity)
				}).
				Return(&apiv1.PauseScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Pause(context.Background(), scheduleTestID, reason))
		})
	}
}

func TestScheduleClient_Unpause(t *testing.T) {
	const reason = "resuming after maintenance"
	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success"},
		{name: "rpc failure", rpcError: errScheduleNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				UnpauseSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.UnpauseScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
					assert.Equal(t, reason, req.Reason)
				}).
				Return(&apiv1.UnpauseScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Unpause(context.Background(), scheduleTestID, reason))
		})
	}
}

func TestScheduleClient_Backfill(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	backfillID := "bf-001"

	testcases := []struct {
		name     string
		rpcError error
	}{
		{name: "success"},
		{name: "rpc failure", rpcError: errScheduleNonRetryable},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			request := &BackfillRequest{
				StartTime:  start,
				EndTime:    end,
				BackfillID: backfillID,
			}

			td.mockService.EXPECT().
				BackfillSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.BackfillScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
					assert.Equal(t, backfillID, req.BackfillId)
				}).
				Return(&apiv1.BackfillScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Backfill(context.Background(), scheduleTestID, request))
		})
	}
}

func TestScheduleClient_List(t *testing.T) {
	testcases := []struct {
		name          string
		rpcError      error
		rpcResponse   *apiv1.ListSchedulesResponse
		pageSize      int32
		nextPageToken []byte
	}{
		{
			name:          "success",
			rpcResponse:   &apiv1.ListSchedulesResponse{},
			pageSize:      10,
			nextPageToken: []byte("token"),
		},
		{
			name:     "rpc failure",
			rpcError: errScheduleNonRetryable,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				ListSchedules(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *apiv1.ListSchedulesRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, tt.pageSize, req.PageSize)
					assert.Equal(t, tt.nextPageToken, req.NextPageToken)
				}).
				Return(tt.rpcResponse, tt.rpcError)

			resp, err := td.sc.List(context.Background(), tt.pageSize, tt.nextPageToken)
			assert.Equal(t, tt.rpcError, err)
			if tt.rpcError == nil {
				require.NotNil(t, resp)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}
