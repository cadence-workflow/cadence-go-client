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

	gogo "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	yarpc "go.uber.org/yarpc"

	s "go.uber.org/cadence/.gen/go/shared"
)

const (
	scheduleTestDomain = "test-domain"
	scheduleTestID     = "test-schedule-id"
	scheduleTestIdent  = "test-identity"
)

var errScheduleNonRetryable = &s.BadRequestError{Message: "bad request"}

// ── Test helpers ──────────────────────────────────────────────────────────────

type scheduleClientTestData struct {
	sc          ScheduleClient
	mockService *ScheduleAPIYARPCClient
}

func newScheduleClientTestData(t *testing.T) *scheduleClientTestData {
	mockService := NewScheduleAPIYARPCClient(t)
	sc := NewScheduleClient(mockService, scheduleTestDomain, &ClientOptions{
		Identity: scheduleTestIdent,
	})
	return &scheduleClientTestData{sc: sc, mockService: mockService}
}

func mustProtoTimestamp(t time.Time) *gogo.Timestamp {
	ts, err := gogo.TimestampProto(t)
	if err != nil {
		panic(err)
	}
	return ts
}

// ── Constructor ───────────────────────────────────────────────────────────────

func TestNewScheduleClient(t *testing.T) {
	testcases := []struct {
		name    string
		options *ClientOptions
	}{
		{name: "nil options", options: nil},
		{name: "custom data converter", options: &ClientOptions{DataConverter: getDefaultDataConverter()}},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			mockService := NewScheduleAPIYARPCClient(t)
			sc := NewScheduleClient(mockService, "domain", tt.options)
			require.NotNil(t, sc)
		})
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestScheduleClient_Create(t *testing.T) {
	basicRequest := func() *CreateScheduleRequest {
		return &CreateScheduleRequest{
			ScheduleID: scheduleTestID,
			Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
			Action: &ScheduleAction{
				StartWorkflow: &ScheduleStartWorkflowAction{
					WorkflowType: "my-workflow",
					TaskList:     "my-task-list",
				},
			},
		}
	}

	testcases := []struct {
		name               string
		request            *CreateScheduleRequest
		rpcResponse        *apiv1.CreateScheduleResponse
		rpcError           error
		expectedScheduleID string
		verify             func(*testing.T, *apiv1.CreateScheduleRequest)
	}{
		{
			name:               "success - basic",
			request:            basicRequest(),
			rpcResponse:        &apiv1.CreateScheduleResponse{ScheduleId: scheduleTestID},
			expectedScheduleID: scheduleTestID,
			verify: func(t *testing.T, req *apiv1.CreateScheduleRequest) {
				assert.Equal(t, scheduleTestDomain, req.Domain)
				assert.Equal(t, scheduleTestID, req.ScheduleId)
				assert.Equal(t, "0 * * * *", req.Spec.CronExpression)
				assert.Equal(t, "my-workflow", req.Action.StartWorkflow.WorkflowType.Name)
			},
		},
		{
			name: "success - full request",
			request: &CreateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType:     "my-workflow",
						TaskList:         "my-task-list",
						Input:            []byte(`"hello"`),
						RetryPolicy:      &RetryPolicy{MaximumAttempts: 3},
						Memo:             map[string]interface{}{"mk": "mv"},
						SearchAttributes: map[string]interface{}{"sk": "sv"},
					},
				},
				Policies: &SchedulePolicies{
					OverlapPolicy:  ScheduleOverlapPolicyBuffer,
					CatchUpPolicy:  ScheduleCatchUpPolicyOne,
					PauseOnFailure: true,
					BufferLimit:    5,
				},
				Memo:             map[string]interface{}{"m": "mv"},
				SearchAttributes: map[string]interface{}{"sa": "sav"},
			},
			rpcResponse:        &apiv1.CreateScheduleResponse{ScheduleId: scheduleTestID},
			expectedScheduleID: scheduleTestID,
			verify: func(t *testing.T, req *apiv1.CreateScheduleRequest) {
				assert.NotNil(t, req.Memo)
				assert.NotNil(t, req.SearchAttributes)
				require.NotNil(t, req.Policies)
				assert.Equal(t, apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_BUFFER, req.Policies.OverlapPolicy)
				assert.Equal(t, apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ONE, req.Policies.CatchUpPolicy)
				assert.True(t, req.Policies.PauseOnFailure)
				require.NotNil(t, req.Action.StartWorkflow.RetryPolicy)
				assert.Equal(t, int32(3), req.Action.StartWorkflow.RetryPolicy.MaximumAttempts)
				require.NotNil(t, req.Action.StartWorkflow.Input)
				assert.Equal(t, []byte(`"hello"`), req.Action.StartWorkflow.Input.Data)
			},
		},
		{
			name:     "rpc failure",
			request:  basicRequest(),
			rpcError: errScheduleNonRetryable,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			td.mockService.EXPECT().
				CreateSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.CreateScheduleRequest, _ ...yarpc.CallOption) {
					if tt.verify != nil {
						tt.verify(t, req)
					}
				}).
				Return(tt.rpcResponse, tt.rpcError)

			scheduleID, err := td.sc.Create(context.Background(), tt.request)
			assert.Equal(t, tt.rpcError, err)
			assert.Equal(t, tt.expectedScheduleID, scheduleID)
		})
	}
}

func TestScheduleClient_Create_Validation(t *testing.T) {
	validRequest := func() *CreateScheduleRequest {
		return &CreateScheduleRequest{
			ScheduleID: scheduleTestID,
			Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
			Action: &ScheduleAction{
				StartWorkflow: &ScheduleStartWorkflowAction{
					WorkflowType: "my-workflow",
					TaskList:     "my-task-list",
				},
			},
		}
	}

	testcases := []struct {
		name    string
		request *CreateScheduleRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "Create: request is required",
		},
		{
			name: "empty ScheduleID",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.ScheduleID = ""
				return r
			}(),
			wantErr: "Create: ScheduleID is required",
		},
		{
			name: "nil Spec",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Spec = nil
				return r
			}(),
			wantErr: "Create: Spec.CronExpression is required",
		},
		{
			name: "empty CronExpression",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Spec.CronExpression = ""
				return r
			}(),
			wantErr: "Create: Spec.CronExpression is required",
		},
		{
			name: "nil Action",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action = nil
				return r
			}(),
			wantErr: "Create: Action.StartWorkflow is required",
		},
		{
			name: "nil Action.StartWorkflow",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action.StartWorkflow = nil
				return r
			}(),
			wantErr: "Create: Action.StartWorkflow is required",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			_, err := td.sc.Create(context.Background(), tt.request)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestScheduleClient_Create_EncodingErrors(t *testing.T) {
	ch := make(chan int)
	baseAction := func() *ScheduleAction {
		return &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType: "my-workflow",
				TaskList:     "my-task-list",
			},
		}
	}

	testcases := []struct {
		name    string
		request *CreateScheduleRequest
		wantErr string
	}{
		{
			name: "action memo encoding error",
			request: &CreateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType: "my-workflow",
						TaskList:     "my-task-list",
						Memo:         map[string]interface{}{"key": ch},
					},
				},
			},
			wantErr: "encode memo field",
		},
		{
			name: "action search attribute encoding error",
			request: &CreateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType:     "my-workflow",
						TaskList:         "my-task-list",
						SearchAttributes: map[string]interface{}{"key": ch},
					},
				},
			},
			wantErr: "encode search attribute",
		},
		{
			name: "schedule search attribute encoding error",
			request: &CreateScheduleRequest{
				ScheduleID:       scheduleTestID,
				Spec:             &ScheduleSpec{CronExpression: "0 * * * *"},
				Action:           baseAction(),
				SearchAttributes: map[string]interface{}{"key": ch},
			},
			wantErr: "encode search attribute",
		},
		{
			name: "schedule memo encoding error",
			request: &CreateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
				Action:     baseAction(),
				Memo:       map[string]interface{}{"key": ch},
			},
			wantErr: "encode memo field",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			_, err := td.sc.Create(context.Background(), tt.request)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// ── Describe ──────────────────────────────────────────────────────────────────

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
				DescribeSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.DescribeScheduleRequest, _ ...yarpc.CallOption) {
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

func TestScheduleClient_Describe_Validation(t *testing.T) {
	td := newScheduleClientTestData(t)
	_, err := td.sc.Describe(context.Background(), "")
	require.EqualError(t, err, "Describe: scheduleID is required")
}

func TestScheduleClient_Describe_FullResponse(t *testing.T) {
	td := newScheduleClientTestData(t)

	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	startTs := mustProtoTimestamp(start)
	endTs := mustProtoTimestamp(end)

	protoResp := &apiv1.DescribeScheduleResponse{
		Spec: &apiv1.ScheduleSpec{
			CronExpression: "0 * * * *",
			StartTime:      startTs,
			EndTime:        endTs,
			Jitter:         gogo.DurationProto(5 * time.Minute),
		},
		Action: &apiv1.ScheduleAction{
			StartWorkflow: &apiv1.ScheduleAction_StartWorkflowAction{
				WorkflowType:                 &apiv1.WorkflowType{Name: "my-workflow"},
				TaskList:                     &apiv1.TaskList{Name: "my-task-list"},
				Input:                        &apiv1.Payload{Data: []byte("input")},
				WorkflowIdPrefix:             "prefix-",
				ExecutionStartToCloseTimeout: gogo.DurationProto(time.Hour),
				TaskStartToCloseTimeout:      gogo.DurationProto(10 * time.Second),
				RetryPolicy: &apiv1.RetryPolicy{
					InitialInterval:    gogo.DurationProto(time.Second),
					BackoffCoefficient: 2.0,
					MaximumInterval:    gogo.DurationProto(time.Minute),
					MaximumAttempts:    3,
				},
			},
		},
		Policies: &apiv1.SchedulePolicies{
			OverlapPolicy:    apiv1.ScheduleOverlapPolicy_SCHEDULE_OVERLAP_POLICY_BUFFER,
			CatchUpPolicy:    apiv1.ScheduleCatchUpPolicy_SCHEDULE_CATCH_UP_POLICY_ONE,
			CatchUpWindow:    gogo.DurationProto(24 * time.Hour),
			PauseOnFailure:   true,
			BufferLimit:      5,
			ConcurrencyLimit: 2,
		},
		State: &apiv1.ScheduleState{
			Paused: true,
			PauseInfo: &apiv1.SchedulePauseInfo{
				Reason:   "maintenance",
				PausedAt: startTs,
				PausedBy: "user",
			},
		},
		Info: &apiv1.ScheduleInfo{
			LastRunTime:    startTs,
			NextRunTime:    endTs,
			TotalRuns:      42,
			CreateTime:     startTs,
			LastUpdateTime: endTs,
			OngoingBackfills: []*apiv1.BackfillInfo{
				{
					BackfillId:    "bf-1",
					StartTime:     startTs,
					EndTime:       endTs,
					RunsCompleted: 3,
					RunsTotal:     5,
				},
			},
		},
		Memo:             &apiv1.Memo{Fields: map[string]*apiv1.Payload{"m": {Data: []byte("mv")}}},
		SearchAttributes: &apiv1.SearchAttributes{IndexedFields: map[string]*apiv1.Payload{"sa": {Data: []byte(`"sav"`)}}},
	}

	td.mockService.EXPECT().
		DescribeSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(protoResp, nil)

	resp, err := td.sc.Describe(context.Background(), scheduleTestID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotNil(t, resp.Spec)
	assert.Equal(t, "0 * * * *", resp.Spec.CronExpression)
	assert.Equal(t, start, resp.Spec.StartTime)
	assert.Equal(t, end, resp.Spec.EndTime)
	assert.Equal(t, 5*time.Minute, resp.Spec.Jitter)

	require.NotNil(t, resp.Action)
	sw := resp.Action.StartWorkflow
	require.NotNil(t, sw)
	assert.Equal(t, "my-workflow", sw.WorkflowType)
	assert.Equal(t, "my-task-list", sw.TaskList)
	assert.Equal(t, []byte("input"), sw.Input)
	assert.Equal(t, "prefix-", sw.WorkflowIDPrefix)
	assert.Equal(t, time.Hour, sw.ExecutionStartToCloseTimeout)
	assert.Equal(t, 10*time.Second, sw.DecisionTaskStartToCloseTimeout)
	require.NotNil(t, sw.RetryPolicy)
	assert.Equal(t, time.Second, sw.RetryPolicy.InitialInterval)
	assert.Equal(t, 2.0, sw.RetryPolicy.BackoffCoefficient)
	assert.Equal(t, int32(3), sw.RetryPolicy.MaximumAttempts)

	require.NotNil(t, resp.Policies)
	assert.Equal(t, ScheduleOverlapPolicyBuffer, resp.Policies.OverlapPolicy)
	assert.Equal(t, ScheduleCatchUpPolicyOne, resp.Policies.CatchUpPolicy)
	assert.Equal(t, 24*time.Hour, resp.Policies.CatchUpWindow)
	assert.True(t, resp.Policies.PauseOnFailure)
	assert.Equal(t, int32(5), resp.Policies.BufferLimit)
	assert.Equal(t, int32(2), resp.Policies.ConcurrencyLimit)

	require.NotNil(t, resp.State)
	assert.True(t, resp.State.Paused)
	require.NotNil(t, resp.State.PauseInfo)
	assert.Equal(t, "maintenance", resp.State.PauseInfo.Reason)
	assert.Equal(t, start, resp.State.PauseInfo.PausedAt)
	assert.Equal(t, "user", resp.State.PauseInfo.PausedBy)

	require.NotNil(t, resp.Info)
	assert.Equal(t, start, resp.Info.LastRunTime)
	assert.Equal(t, end, resp.Info.NextRunTime)
	assert.Equal(t, int64(42), resp.Info.TotalRuns)
	require.Len(t, resp.Info.OngoingBackfills, 1)
	bf := resp.Info.OngoingBackfills[0]
	assert.Equal(t, "bf-1", bf.BackfillID)
	assert.Equal(t, start, bf.StartTime)
	assert.Equal(t, end, bf.EndTime)
	assert.Equal(t, int32(3), bf.RunsCompleted)
	assert.Equal(t, int32(5), bf.RunsTotal)

	assert.Equal(t, []byte("mv"), resp.Memo["m"])
	assert.Equal(t, []byte(`"sav"`), resp.SearchAttributes["sa"])
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestScheduleClient_Update(t *testing.T) {
	testcases := []struct {
		name     string
		request  *UpdateScheduleRequest
		rpcError error
		verify   func(*testing.T, *apiv1.UpdateScheduleRequest)
	}{
		{
			name: "success - spec only",
			request: &UpdateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 2 * * *"},
			},
			verify: func(t *testing.T, req *apiv1.UpdateScheduleRequest) {
				assert.Equal(t, scheduleTestDomain, req.Domain)
				assert.Equal(t, scheduleTestID, req.ScheduleId)
				assert.Equal(t, "0 2 * * *", req.Spec.CronExpression)
			},
		},
		{
			name: "success - with action and search attrs",
			request: &UpdateScheduleRequest{
				ScheduleID: scheduleTestID,
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType: "my-workflow",
						TaskList:     "my-task-list",
					},
				},
				SearchAttributes: map[string]interface{}{"attr": "val"},
			},
			verify: func(t *testing.T, req *apiv1.UpdateScheduleRequest) {
				assert.NotNil(t, req.Action)
				assert.NotNil(t, req.SearchAttributes)
			},
		},
		{
			name: "rpc failure",
			request: &UpdateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 2 * * *"},
			},
			rpcError: errScheduleNonRetryable,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				UpdateSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.UpdateScheduleRequest, _ ...yarpc.CallOption) {
					if tt.verify != nil {
						tt.verify(t, req)
					}
				}).
				Return(&apiv1.UpdateScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Update(context.Background(), tt.request))
		})
	}
}

func TestScheduleClient_Update_Validation(t *testing.T) {
	validRequest := func() *UpdateScheduleRequest {
		return &UpdateScheduleRequest{
			ScheduleID: scheduleTestID,
			Spec:       &ScheduleSpec{CronExpression: "0 2 * * *"},
		}
	}

	testcases := []struct {
		name    string
		request *UpdateScheduleRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "Update: request is required",
		},
		{
			name: "empty ScheduleID",
			request: func() *UpdateScheduleRequest {
				r := validRequest()
				r.ScheduleID = ""
				return r
			}(),
			wantErr: "Update: ScheduleID is required",
		},
		{
			name: "all fields nil",
			request: func() *UpdateScheduleRequest {
				r := validRequest()
				r.Spec = nil
				r.Action = nil
				r.Policies = nil
				return r
			}(),
			wantErr: "Update: at least one of Spec, Action, or Policies must be set",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			require.EqualError(t, td.sc.Update(context.Background(), tt.request), tt.wantErr)
		})
	}
}

func TestScheduleClient_Update_EncodingErrors(t *testing.T) {
	ch := make(chan int)

	testcases := []struct {
		name    string
		request *UpdateScheduleRequest
	}{
		{
			name: "action memo encoding error",
			request: &UpdateScheduleRequest{
				ScheduleID: scheduleTestID,
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType: "my-workflow",
						TaskList:     "my-task-list",
						Memo:         map[string]interface{}{"key": ch},
					},
				},
			},
		},
		{
			name: "search attribute encoding error",
			request: &UpdateScheduleRequest{
				ScheduleID:       scheduleTestID,
				Spec:             &ScheduleSpec{CronExpression: "0 2 * * *"},
				SearchAttributes: map[string]interface{}{"key": ch},
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			require.Error(t, td.sc.Update(context.Background(), tt.request))
		})
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

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
				DeleteSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.DeleteScheduleRequest, _ ...yarpc.CallOption) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
				}).
				Return(&apiv1.DeleteScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Delete(context.Background(), scheduleTestID))
		})
	}
}

func TestScheduleClient_Delete_Validation(t *testing.T) {
	td := newScheduleClientTestData(t)
	require.EqualError(t, td.sc.Delete(context.Background(), ""), "Delete: scheduleID is required")
}

// ── Pause ─────────────────────────────────────────────────────────────────────

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
				PauseSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.PauseScheduleRequest, _ ...yarpc.CallOption) {
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

func TestScheduleClient_Pause_Validation(t *testing.T) {
	td := newScheduleClientTestData(t)
	require.EqualError(t, td.sc.Pause(context.Background(), "", "reason"), "Pause: scheduleID is required")
}

// ── Unpause ───────────────────────────────────────────────────────────────────

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
				UnpauseSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.UnpauseScheduleRequest, _ ...yarpc.CallOption) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
					assert.Equal(t, reason, req.Reason)
				}).
				Return(&apiv1.UnpauseScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Unpause(context.Background(), scheduleTestID, reason))
		})
	}
}

func TestScheduleClient_Unpause_Validation(t *testing.T) {
	td := newScheduleClientTestData(t)
	require.EqualError(t, td.sc.Unpause(context.Background(), "", "reason"), "Unpause: scheduleID is required")
}

// ── Backfill ──────────────────────────────────────────────────────────────────

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
				BackfillSchedule(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.BackfillScheduleRequest, _ ...yarpc.CallOption) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, scheduleTestID, req.ScheduleId)
					assert.Equal(t, backfillID, req.BackfillId)
				}).
				Return(&apiv1.BackfillScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Backfill(context.Background(), scheduleTestID, request))
		})
	}
}

func TestScheduleClient_Backfill_Validation(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	validRequest := func() *BackfillRequest {
		return &BackfillRequest{StartTime: start, EndTime: end}
	}

	testcases := []struct {
		name       string
		scheduleID string
		request    *BackfillRequest
		wantErr    string
	}{
		{
			name:       "empty scheduleID",
			scheduleID: "",
			request:    validRequest(),
			wantErr:    "Backfill: scheduleID is required",
		},
		{
			name:       "nil request",
			scheduleID: scheduleTestID,
			request:    nil,
			wantErr:    "Backfill: request is required",
		},
		{
			name:       "zero StartTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.StartTime = time.Time{}; return r }(),
			wantErr:    "Backfill: StartTime is required",
		},
		{
			name:       "zero EndTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.EndTime = time.Time{}; return r }(),
			wantErr:    "Backfill: EndTime is required",
		},
		{
			name:       "EndTime equal to StartTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.EndTime = start; return r }(),
			wantErr:    "Backfill: EndTime must be after StartTime",
		},
		{
			name:       "EndTime before StartTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.EndTime = start.Add(-time.Hour); return r }(),
			wantErr:    "Backfill: EndTime must be after StartTime",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			require.EqualError(t, td.sc.Backfill(context.Background(), tt.scheduleID, tt.request), tt.wantErr)
		})
	}
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestScheduleClient_List(t *testing.T) {
	testcases := []struct {
		name          string
		rpcError      error
		rpcResponse   *apiv1.ListSchedulesResponse
		pageSize      int32
		nextPageToken []byte
		verify        func(*testing.T, *ListSchedulesResponse)
	}{
		{
			name:          "success - empty response",
			rpcResponse:   &apiv1.ListSchedulesResponse{},
			pageSize:      10,
			nextPageToken: []byte("token"),
		},
		{
			name: "success - with entries",
			rpcResponse: &apiv1.ListSchedulesResponse{
				Schedules: []*apiv1.ScheduleListEntry{
					{
						ScheduleId:     "sched-1",
						WorkflowType:   &apiv1.WorkflowType{Name: "my-workflow"},
						CronExpression: "0 * * * *",
						State:          &apiv1.ScheduleState{Paused: true},
					},
				},
				NextPageToken: []byte("next-token"),
			},
			pageSize: 10,
			verify: func(t *testing.T, resp *ListSchedulesResponse) {
				assert.Equal(t, []byte("next-token"), resp.NextPageToken)
				require.Len(t, resp.Schedules, 1)
				entry := resp.Schedules[0]
				assert.Equal(t, "sched-1", entry.ScheduleID)
				assert.Equal(t, "my-workflow", entry.WorkflowType)
				assert.Equal(t, "0 * * * *", entry.CronExpression)
				require.NotNil(t, entry.State)
				assert.True(t, entry.State.Paused)
			},
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
				ListSchedules(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Run(func(_ context.Context, req *apiv1.ListSchedulesRequest, _ ...yarpc.CallOption) {
					assert.Equal(t, scheduleTestDomain, req.Domain)
					assert.Equal(t, tt.pageSize, req.PageSize)
					assert.Equal(t, tt.nextPageToken, req.NextPageToken)
				}).
				Return(tt.rpcResponse, tt.rpcError)

			resp, err := td.sc.List(context.Background(), tt.pageSize, tt.nextPageToken)
			assert.Equal(t, tt.rpcError, err)
			if tt.rpcError == nil {
				require.NotNil(t, resp)
				if tt.verify != nil {
					tt.verify(t, resp)
				}
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}
