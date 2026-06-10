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
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
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
	mockService *workflowservicetest.MockClient
}

func newScheduleClientTestData(t *testing.T) *scheduleClientTestData {
	ctrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(ctrl)
	sc := NewClient(mockService, scheduleTestDomain, &ClientOptions{
		Identity: scheduleTestIdent,
	}).ScheduleClient()
	return &scheduleClientTestData{sc: sc, mockService: mockService}
}

// ── Constructor ───────────────────────────────────────────────────────────────

func TestScheduleClient(t *testing.T) {
	testcases := []struct {
		name    string
		options *ClientOptions
	}{
		{name: "nil options", options: nil},
		{name: "custom data converter", options: &ClientOptions{DataConverter: getDefaultDataConverter()}},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			sc := NewClient(workflowservicetest.NewMockClient(ctrl), "domain", tt.options).ScheduleClient()
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
					WorkflowType:                 "my-workflow",
					TaskList:                     "my-task-list",
					ExecutionStartToCloseTimeout: time.Hour,
				},
			},
		}
	}

	testcases := []struct {
		name               string
		request            *CreateScheduleRequest
		rpcResponse        *s.CreateScheduleResponse
		rpcError           error
		expectedScheduleID string
		verify             func(*testing.T, *s.CreateScheduleRequest)
	}{
		{
			name:               "success - basic",
			request:            basicRequest(),
			rpcResponse:        &s.CreateScheduleResponse{ScheduleId: common.StringPtr(scheduleTestID)},
			expectedScheduleID: scheduleTestID,
			verify: func(t *testing.T, req *s.CreateScheduleRequest) {
				assert.Equal(t, scheduleTestDomain, req.GetDomain())
				assert.Equal(t, scheduleTestID, req.GetScheduleId())
				require.NotNil(t, req.Spec)
				assert.Equal(t, "0 * * * *", req.Spec.GetCronExpression())
				require.NotNil(t, req.Action)
				require.NotNil(t, req.Action.StartWorkflow)
				assert.Equal(t, "my-workflow", req.Action.StartWorkflow.WorkflowType.GetName())
			},
		},
		{
			name: "success - full request",
			request: &CreateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType:                 "my-workflow",
						TaskList:                     "my-task-list",
						ExecutionStartToCloseTimeout: time.Hour,
						Input:                        []byte(`"hello"`),
						RetryPolicy:                  &RetryPolicy{MaximumAttempts: 3},
						Memo:                         map[string]interface{}{"mk": "mv"},
						SearchAttributes:             map[string]interface{}{"sk": "sv"},
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
			rpcResponse:        &s.CreateScheduleResponse{ScheduleId: common.StringPtr(scheduleTestID)},
			expectedScheduleID: scheduleTestID,
			verify: func(t *testing.T, req *s.CreateScheduleRequest) {
				assert.NotNil(t, req.Memo)
				assert.NotNil(t, req.SearchAttributes)
				require.NotNil(t, req.Policies)
				require.NotNil(t, req.Policies.OverlapPolicy)
				assert.Equal(t, s.ScheduleOverlapPolicyBuffer, *req.Policies.OverlapPolicy)
				require.NotNil(t, req.Policies.CatchUpPolicy)
				assert.Equal(t, s.ScheduleCatchUpPolicyOne, *req.Policies.CatchUpPolicy)
				assert.Equal(t, true, req.Policies.GetPauseOnFailure())
				require.NotNil(t, req.Action.StartWorkflow.RetryPolicy)
				assert.Equal(t, int32(3), req.Action.StartWorkflow.RetryPolicy.GetMaximumAttempts())
				// zero BackoffCoefficient must be defaulted, not sent as 0
				assert.Equal(t, backoff.DefaultBackoffCoefficient, req.Action.StartWorkflow.RetryPolicy.GetBackoffCoefficient())
				assert.Equal(t, []byte(`"hello"`), req.Action.StartWorkflow.Input)
			},
		},
		{
			name:     "rpc failure",
			request:  basicRequest(),
			rpcError: errScheduleNonRetryable,
		},
		{
			name: "sub-second retry interval is ceiling'd to 1 second",
			request: &CreateScheduleRequest{
				ScheduleID: scheduleTestID,
				Spec:       &ScheduleSpec{CronExpression: "0 * * * *"},
				Action: &ScheduleAction{
					StartWorkflow: &ScheduleStartWorkflowAction{
						WorkflowType:                 "my-workflow",
						TaskList:                     "my-task-list",
						ExecutionStartToCloseTimeout: time.Hour,
						RetryPolicy: &RetryPolicy{
							InitialInterval:    500 * time.Millisecond,
							MaximumInterval:    1500 * time.Millisecond,
							ExpirationInterval: 100 * time.Millisecond,
							BackoffCoefficient: 1.5,
						},
					},
				},
			},
			rpcResponse:        &s.CreateScheduleResponse{ScheduleId: common.StringPtr(scheduleTestID)},
			expectedScheduleID: scheduleTestID,
			verify: func(t *testing.T, req *s.CreateScheduleRequest) {
				rp := req.Action.StartWorkflow.RetryPolicy
				require.NotNil(t, rp)
				assert.Equal(t, int32(1), rp.GetInitialIntervalInSeconds())    // ceil(0.5) = 1
				assert.Equal(t, int32(2), rp.GetMaximumIntervalInSeconds())    // ceil(1.5) = 2
				assert.Equal(t, int32(1), rp.GetExpirationIntervalInSeconds()) // ceil(0.1) = 1
				assert.Equal(t, 1.5, rp.GetBackoffCoefficient())
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			td.mockService.EXPECT().
				CreateSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *s.CreateScheduleRequest, _ ...interface{}) {
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
					WorkflowType:                 "my-workflow",
					TaskList:                     "my-task-list",
					ExecutionStartToCloseTimeout: time.Hour,
				},
			},
		}
	}

	testcases := []struct {
		name    string
		request *CreateScheduleRequest
		wantErr error
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: ErrRequestRequired,
		},
		{
			name: "empty ScheduleID",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.ScheduleID = ""
				return r
			}(),
			wantErr: ErrScheduleIDRequired,
		},
		{
			name: "nil Spec",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Spec = nil
				return r
			}(),
			wantErr: ErrCronExpressionRequired,
		},
		{
			name: "empty CronExpression",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Spec.CronExpression = ""
				return r
			}(),
			wantErr: ErrCronExpressionRequired,
		},
		{
			name: "nil Action",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action = nil
				return r
			}(),
			wantErr: ErrActionStartWorkflowRequired,
		},
		{
			name: "nil Action.StartWorkflow",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action.StartWorkflow = nil
				return r
			}(),
			wantErr: ErrActionStartWorkflowRequired,
		},
		{
			name: "empty WorkflowType",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action.StartWorkflow.WorkflowType = ""
				return r
			}(),
			wantErr: ErrWorkflowTypeRequired,
		},
		{
			name: "empty TaskList",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action.StartWorkflow.TaskList = ""
				return r
			}(),
			wantErr: ErrTaskListRequired,
		},
		{
			name: "zero ExecutionStartToCloseTimeout",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action.StartWorkflow.ExecutionStartToCloseTimeout = 0
				return r
			}(),
			wantErr: ErrExecutionTimeoutRequired,
		},
		{
			name: "negative DecisionTaskStartToCloseTimeout",
			request: func() *CreateScheduleRequest {
				r := validRequest()
				r.Action.StartWorkflow.DecisionTaskStartToCloseTimeout = -time.Second
				return r
			}(),
			wantErr: ErrNegativeDecisionTimeout,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			_, err := td.sc.Create(context.Background(), tt.request)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestScheduleClient_Create_EncodingErrors(t *testing.T) {
	ch := make(chan int)
	baseAction := func() *ScheduleAction {
		return &ScheduleAction{
			StartWorkflow: &ScheduleStartWorkflowAction{
				WorkflowType:                 "my-workflow",
				TaskList:                     "my-task-list",
				ExecutionStartToCloseTimeout: time.Hour,
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
						WorkflowType:                 "my-workflow",
						TaskList:                     "my-task-list",
						ExecutionStartToCloseTimeout: time.Hour,
						Memo:                         map[string]interface{}{"key": ch},
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
						WorkflowType:                 "my-workflow",
						TaskList:                     "my-task-list",
						ExecutionStartToCloseTimeout: time.Hour,
						SearchAttributes:             map[string]interface{}{"key": ch},
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
		rpcResponse *s.DescribeScheduleResponse
	}{
		{
			name:        "success",
			rpcResponse: &s.DescribeScheduleResponse{},
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
				Do(func(_ context.Context, req *s.DescribeScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.GetDomain())
					assert.Equal(t, scheduleTestID, req.GetScheduleId())
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
	require.ErrorIs(t, err, ErrScheduleIDRequired)
}

func TestScheduleClient_Describe_FullResponse(t *testing.T) {
	td := newScheduleClientTestData(t)

	// NOTE: This test uses a hand-crafted synthetic Thrift response to verify the
	// conversion logic for all fields. In practice, the server does not populate
	// CreateTimeNano, LastUpdateTimeNano, OngoingBackfills, or PausedTimeNano —
	// those fields always arrive as nil and convert to zero values.
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	startNano := start.UnixNano()
	endNano := end.UnixNano()
	jitterSec := int32(300)
	exSec := int32(3600)
	taskSec := int32(10)
	totalRuns := int64(42)
	runsCompleted := int32(3)
	runsTotal := int32(5)
	paused := true
	coeff := 2.0
	maxAttempts := int32(3)
	bufferLimit := int32(5)
	concurrencyLimit := int32(2)
	catchUpWindowSec := int32(86400)
	pauseOnFailure := true

	overlapPolicy := s.ScheduleOverlapPolicyBuffer
	catchUpPolicy := s.ScheduleCatchUpPolicyOne

	thriftResp := &s.DescribeScheduleResponse{
		Spec: &s.ScheduleSpec{
			CronExpression:  common.StringPtr("0 * * * *"),
			StartTimeNano:   &startNano,
			EndTimeNano:     &endNano,
			JitterInSeconds: &jitterSec,
		},
		Action: &s.ScheduleAction{
			StartWorkflow: &s.ScheduleStartWorkflowAction{
				WorkflowType:                        &s.WorkflowType{Name: common.StringPtr("my-workflow")},
				TaskList:                            &s.TaskList{Name: common.StringPtr("my-task-list")},
				Input:                               []byte("input"),
				WorkflowIdPrefix:                    common.StringPtr("prefix-"),
				ExecutionStartToCloseTimeoutSeconds: &exSec,
				TaskStartToCloseTimeoutSeconds:      &taskSec,
				RetryPolicy: &s.RetryPolicy{
					InitialIntervalInSeconds: common.Int32Ptr(1),
					BackoffCoefficient:       &coeff,
					MaximumIntervalInSeconds: common.Int32Ptr(60),
					MaximumAttempts:          &maxAttempts,
				},
			},
		},
		Policies: &s.SchedulePolicies{
			OverlapPolicy:          &overlapPolicy,
			CatchUpPolicy:          &catchUpPolicy,
			CatchUpWindowInSeconds: &catchUpWindowSec,
			PauseOnFailure:         &pauseOnFailure,
			BufferLimit:            &bufferLimit,
			ConcurrencyLimit:       &concurrencyLimit,
		},
		State: &s.ScheduleState{
			Paused: &paused,
			PauseInfo: &s.SchedulePauseInfo{
				Reason:         common.StringPtr("maintenance"),
				PausedTimeNano: &startNano,
			},
		},
		Info: &s.ScheduleInfo{
			LastRunTimeNano:    &startNano,
			NextRunTimeNano:    &endNano,
			TotalRuns:          &totalRuns,
			CreateTimeNano:     &startNano,
			LastUpdateTimeNano: &endNano,
			OngoingBackfills: []*s.BackfillInfo{
				{
					BackfillId:    common.StringPtr("bf-1"),
					StartTimeNano: &startNano,
					EndTimeNano:   &endNano,
					RunsCompleted: &runsCompleted,
					RunsTotal:     &runsTotal,
				},
			},
		},
	}

	td.mockService.EXPECT().
		DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(thriftResp, nil)

	resp, err := td.sc.Describe(context.Background(), scheduleTestID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotNil(t, resp.Spec)
	assert.Equal(t, "0 * * * *", resp.Spec.CronExpression)
	assert.Equal(t, start, resp.Spec.StartTime)
	assert.Equal(t, end, resp.Spec.EndTime)
	assert.Equal(t, 300*time.Second, resp.Spec.Jitter)

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
}

// ── Update ────────────────────────────────────────────────────────────────────

func TestScheduleClient_Update(t *testing.T) {
	// baseline is the schedule's current thrift state returned by DescribeSchedule.
	baseline := func() *s.DescribeScheduleResponse {
		op := s.ScheduleOverlapPolicyBuffer
		return &s.DescribeScheduleResponse{
			Spec: &s.ScheduleSpec{CronExpression: common.StringPtr("0 * * * *")},
			Action: &s.ScheduleAction{StartWorkflow: &s.ScheduleStartWorkflowAction{
				WorkflowType:                        &s.WorkflowType{Name: common.StringPtr("my-wf")},
				TaskList:                            &s.TaskList{Name: common.StringPtr("my-tl")},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(3600),
			}},
			Policies: &s.SchedulePolicies{
				OverlapPolicy:  &op,
				BufferLimit:    common.Int32Ptr(5),
				PauseOnFailure: common.BoolPtr(false),
			},
		}
	}

	testcases := []struct {
		name         string
		mutate       func(*ScheduleUpdate) error
		expectUpdate bool
		rpcError     error
		verify       func(*testing.T, *s.UpdateScheduleRequest)
	}{
		{
			name:         "change one policy sub-field preserves the others",
			mutate:       func(u *ScheduleUpdate) error { u.Policies.PauseOnFailure = true; return nil },
			expectUpdate: true,
			verify: func(t *testing.T, req *s.UpdateScheduleRequest) {
				require.NotNil(t, req.Policies)
				assert.True(t, req.Policies.GetPauseOnFailure())
				assert.Equal(t, int32(5), req.Policies.GetBufferLimit()) // preserved, not reset
				assert.Nil(t, req.Spec, "unchanged Spec must be omitted")
				assert.Nil(t, req.Action, "unchanged Action must be omitted")
			},
		},
		{
			name:         "change spec only",
			mutate:       func(u *ScheduleUpdate) error { u.Spec.CronExpression = "0 2 * * *"; return nil },
			expectUpdate: true,
			verify: func(t *testing.T, req *s.UpdateScheduleRequest) {
				require.NotNil(t, req.Spec)
				assert.Equal(t, "0 2 * * *", req.Spec.GetCronExpression())
				assert.Nil(t, req.Policies, "unchanged Policies must be omitted")
				assert.Nil(t, req.Action, "unchanged Action must be omitted")
			},
		},
		{
			// Mutating the action (here via SetActionMemo) sends the whole action through
			// scheduleActionDescriptionToThrift: the carried-over required fields survive and the
			// missing DecisionTaskStartToCloseTimeout is defaulted to 10s.
			name:         "change action via SetActionMemo",
			mutate:       func(u *ScheduleUpdate) error { return u.SetActionMemo(map[string]interface{}{"k": "v"}) },
			expectUpdate: true,
			verify: func(t *testing.T, req *s.UpdateScheduleRequest) {
				require.NotNil(t, req.Action)
				require.NotNil(t, req.Action.StartWorkflow)
				sw := req.Action.StartWorkflow
				assert.Equal(t, "my-wf", sw.WorkflowType.GetName(), "required field carried over from Describe")
				assert.Equal(t, "my-tl", sw.TaskList.GetName(), "required field carried over from Describe")
				assert.Equal(t, int32(3600), sw.GetExecutionStartToCloseTimeoutSeconds(), "carried over")
				assert.Equal(t, int32(defaultDecisionTaskTimeoutInSecs), sw.GetTaskStartToCloseTimeoutSeconds(), "defaulted")
				require.NotNil(t, sw.Memo)
				wantV, encErr := encodeArg(getDefaultDataConverter(), "v")
				require.NoError(t, encErr)
				assert.Equal(t, wantV, sw.Memo.Fields["k"])
				assert.Nil(t, req.Spec, "unchanged Spec must be omitted")
				assert.Nil(t, req.Policies, "unchanged Policies must be omitted")
			},
		},
		{
			// SearchAttributes can be added/replaced via Update (baseline has none here).
			name:         "add search attributes",
			mutate:       func(u *ScheduleUpdate) error { return u.SetSearchAttributes(map[string]interface{}{"sk": "sv"}) },
			expectUpdate: true,
			verify: func(t *testing.T, req *s.UpdateScheduleRequest) {
				require.NotNil(t, req.SearchAttributes)
				wantSV, _ := json.Marshal("sv")
				assert.Equal(t, wantSV, req.SearchAttributes.IndexedFields["sk"])
				assert.Nil(t, req.Spec, "unchanged Spec must be omitted")
				assert.Nil(t, req.Action, "unchanged Action must be omitted")
				assert.Nil(t, req.Policies, "unchanged Policies must be omitted")
			},
		},
		{
			name:         "no change does not send UpdateSchedule",
			mutate:       func(u *ScheduleUpdate) error { return nil },
			expectUpdate: false,
		},
		{
			// Nil-ing a top-level field cannot clear it (an omitted wire field is preserved
			// by the server's top-level merge), so it is treated as "no change" rather than
			// firing a silent no-op RPC.
			name:         "nil-ing a top-level field is a no-op",
			mutate:       func(u *ScheduleUpdate) error { u.Policies = nil; return nil },
			expectUpdate: false,
		},
		{
			name:         "rpc failure is returned",
			mutate:       func(u *ScheduleUpdate) error { u.Spec.CronExpression = "0 2 * * *"; return nil },
			expectUpdate: true,
			rpcError:     errScheduleNonRetryable,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			td.mockService.EXPECT().
				DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(baseline(), nil)
			if tt.expectUpdate {
				td.mockService.EXPECT().
					UpdateSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, req *s.UpdateScheduleRequest, _ ...interface{}) {
						assert.Equal(t, scheduleTestDomain, req.GetDomain())
						assert.Equal(t, scheduleTestID, req.GetScheduleId())
						if tt.verify != nil {
							tt.verify(t, req)
						}
					}).
					Return(&s.UpdateScheduleResponse{}, tt.rpcError)
			}
			err := td.sc.Update(context.Background(), scheduleTestID, tt.mutate)
			if tt.rpcError != nil {
				assert.Equal(t, tt.rpcError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestScheduleClient_Update_SearchAttributesCannotClear(t *testing.T) {
	// Baseline already has a schedule-level search attribute.
	describe := func() *s.DescribeScheduleResponse {
		return &s.DescribeScheduleResponse{
			Spec:             &s.ScheduleSpec{CronExpression: common.StringPtr("0 * * * *")},
			SearchAttributes: &s.SearchAttributes{IndexedFields: map[string][]byte{"sk": []byte(`"sv"`)}},
		}
	}

	// Clearing search attributes (nil-ing them in the callback) cannot be expressed on the wire —
	// an omitted field is preserved by the server's top-level merge — so it is a no-op: no RPC.
	t.Run("clearing is a no-op", func(t *testing.T) {
		td := newScheduleClientTestData(t)
		td.mockService.EXPECT().
			DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(describe(), nil)
		// No UpdateSchedule expectation: gomock fails if it is called.
		err := td.sc.Update(context.Background(), scheduleTestID, func(u *ScheduleUpdate) error {
			u.SearchAttributes = nil
			return nil
		})
		require.NoError(t, err)
	})

	// Modifying an existing attribute does send the full (replacement) set.
	t.Run("modifying replaces", func(t *testing.T) {
		td := newScheduleClientTestData(t)
		td.mockService.EXPECT().
			DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(describe(), nil)
		td.mockService.EXPECT().
			UpdateSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, req *s.UpdateScheduleRequest, _ ...interface{}) {
				require.NotNil(t, req.SearchAttributes)
				wantSV, _ := json.Marshal("sv2")
				assert.Equal(t, wantSV, req.SearchAttributes.IndexedFields["sk"])
			}).
			Return(&s.UpdateScheduleResponse{}, nil)
		err := td.sc.Update(context.Background(), scheduleTestID, func(u *ScheduleUpdate) error {
			return u.SetSearchAttributes(map[string]interface{}{"sk": "sv2"})
		})
		require.NoError(t, err)
	})
}

func TestScheduleClient_Update_Validation(t *testing.T) {
	specOnlyDescribe := func() *s.DescribeScheduleResponse {
		return &s.DescribeScheduleResponse{Spec: &s.ScheduleSpec{CronExpression: common.StringPtr("0 * * * *")}}
	}

	t.Run("empty scheduleID errors before any RPC", func(t *testing.T) {
		td := newScheduleClientTestData(t)
		err := td.sc.Update(context.Background(), "", func(*ScheduleUpdate) error { return nil })
		require.ErrorIs(t, err, ErrScheduleIDRequired)
	})

	t.Run("nil mutate errors before any RPC", func(t *testing.T) {
		td := newScheduleClientTestData(t)
		err := td.sc.Update(context.Background(), scheduleTestID, nil)
		require.ErrorIs(t, err, ErrMutateFunctionRequired)
	})

	t.Run("describe failure aborts the update", func(t *testing.T) {
		td := newScheduleClientTestData(t)
		td.mockService.EXPECT().
			DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errScheduleNonRetryable)
		err := td.sc.Update(context.Background(), scheduleTestID, func(*ScheduleUpdate) error { return nil })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "describe failed")
	})

	t.Run("mutate error aborts without UpdateSchedule", func(t *testing.T) {
		td := newScheduleClientTestData(t)
		td.mockService.EXPECT().
			DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(specOnlyDescribe(), nil)
		wantErr := errors.New("boom")
		err := td.sc.Update(context.Background(), scheduleTestID, func(*ScheduleUpdate) error { return wantErr })
		assert.Equal(t, wantErr, err)
	})

	t.Run("cleared cron is rejected", func(t *testing.T) {
		td := newScheduleClientTestData(t)
		td.mockService.EXPECT().
			DescribeSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(specOnlyDescribe(), nil)
		err := td.sc.Update(context.Background(), scheduleTestID, func(u *ScheduleUpdate) error {
			u.Spec.CronExpression = ""
			return nil
		})
		require.ErrorIs(t, err, ErrCronExpressionRequired)
	})
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
				DeleteSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *s.DeleteScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.GetDomain())
					assert.Equal(t, scheduleTestID, req.GetScheduleId())
				}).
				Return(&s.DeleteScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Delete(context.Background(), scheduleTestID))
		})
	}
}

func TestScheduleClient_Delete_Validation(t *testing.T) {
	td := newScheduleClientTestData(t)
	require.ErrorIs(t, td.sc.Delete(context.Background(), ""), ErrScheduleIDRequired)
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
				PauseSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *s.PauseScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.GetDomain())
					assert.Equal(t, scheduleTestID, req.GetScheduleId())
					assert.Equal(t, reason, req.GetReason())
					assert.Equal(t, scheduleTestIdent, req.GetIdentity())
				}).
				Return(&s.PauseScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Pause(context.Background(), scheduleTestID, reason))
		})
	}
}

func TestScheduleClient_Pause_Validation(t *testing.T) {
	td := newScheduleClientTestData(t)
	require.ErrorIs(t, td.sc.Pause(context.Background(), "", "reason"), ErrScheduleIDRequired)
}

// ── Unpause ───────────────────────────────────────────────────────────────────

func TestScheduleClient_Unpause(t *testing.T) {
	const reason = "resuming after maintenance"
	allPolicy := s.ScheduleCatchUpPolicyAll
	testcases := []struct {
		name          string
		reason        string
		catchUpPolicy ScheduleCatchUpPolicy
		wantCUP       *s.ScheduleCatchUpPolicy
		rpcError      error
	}{
		{
			name:          "success - unspecified policy omitted on wire",
			reason:        reason,
			catchUpPolicy: ScheduleCatchUpPolicyUnspecified,
			wantCUP:       nil,
		},
		{
			name:          "success - catch-up policy All",
			reason:        reason,
			catchUpPolicy: ScheduleCatchUpPolicyAll,
			wantCUP:       &allPolicy,
		},
		{
			name:     "rpc failure",
			reason:   reason,
			rpcError: errScheduleNonRetryable,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)

			td.mockService.EXPECT().
				UnpauseSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *s.UnpauseScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.GetDomain())
					assert.Equal(t, scheduleTestID, req.GetScheduleId())
					assert.Equal(t, tt.reason, req.GetReason())
					assert.Equal(t, tt.wantCUP, req.CatchUpPolicy)
				}).
				Return(&s.UnpauseScheduleResponse{}, tt.rpcError)

			assert.Equal(t, tt.rpcError, td.sc.Unpause(context.Background(), scheduleTestID, tt.reason, tt.catchUpPolicy))
		})
	}
}

func TestScheduleClient_Unpause_Validation(t *testing.T) {
	td := newScheduleClientTestData(t)
	require.ErrorIs(t, td.sc.Unpause(context.Background(), "", "reason", ScheduleCatchUpPolicyUnspecified), ErrScheduleIDRequired)
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
				StartTime:     start,
				EndTime:       end,
				BackfillID:    backfillID,
				OverlapPolicy: ScheduleOverlapPolicyBuffer,
			}

			td.mockService.EXPECT().
				BackfillSchedule(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *s.BackfillScheduleRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.GetDomain())
					assert.Equal(t, scheduleTestID, req.GetScheduleId())
					assert.Equal(t, backfillID, req.GetBackfillId())
					assert.Equal(t, start.UnixNano(), req.GetStartTimeNano())
					assert.Equal(t, end.UnixNano(), req.GetEndTimeNano())
					require.NotNil(t, req.OverlapPolicy)
					assert.Equal(t, s.ScheduleOverlapPolicyBuffer, *req.OverlapPolicy)
				}).
				Return(&s.BackfillScheduleResponse{}, tt.rpcError)

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
		wantErr    error
	}{
		{
			name:       "empty scheduleID",
			scheduleID: "",
			request:    validRequest(),
			wantErr:    ErrScheduleIDRequired,
		},
		{
			name:       "nil request",
			scheduleID: scheduleTestID,
			request:    nil,
			wantErr:    ErrRequestRequired,
		},
		{
			name:       "zero StartTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.StartTime = time.Time{}; return r }(),
			wantErr:    ErrStartTimeRequired,
		},
		{
			name:       "zero EndTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.EndTime = time.Time{}; return r }(),
			wantErr:    ErrEndTimeRequired,
		},
		{
			name:       "EndTime equal to StartTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.EndTime = start; return r }(),
			wantErr:    ErrEndTimeBeforeStartTime,
		},
		{
			name:       "EndTime before StartTime",
			scheduleID: scheduleTestID,
			request:    func() *BackfillRequest { r := validRequest(); r.EndTime = start.Add(-time.Hour); return r }(),
			wantErr:    ErrEndTimeBeforeStartTime,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			td := newScheduleClientTestData(t)
			require.ErrorIs(t, td.sc.Backfill(context.Background(), tt.scheduleID, tt.request), tt.wantErr)
		})
	}
}

// ── List ──────────────────────────────────────────────────────────────────────

func TestScheduleClient_List(t *testing.T) {
	paused := true
	testcases := []struct {
		name          string
		rpcError      error
		rpcResponse   *s.ListSchedulesResponse
		pageSize      int32
		nextPageToken []byte
		verify        func(*testing.T, *ListSchedulesResponse)
	}{
		{
			name:          "success - empty response",
			rpcResponse:   &s.ListSchedulesResponse{},
			pageSize:      10,
			nextPageToken: []byte("token"),
		},
		{
			name: "success - with entries",
			rpcResponse: &s.ListSchedulesResponse{
				Schedules: []*s.ScheduleListEntry{
					{
						ScheduleId:     common.StringPtr("sched-1"),
						WorkflowType:   &s.WorkflowType{Name: common.StringPtr("my-workflow")},
						CronExpression: common.StringPtr("0 * * * *"),
						State:          &s.ScheduleState{Paused: &paused},
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
				ListSchedules(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *s.ListSchedulesRequest, _ ...interface{}) {
					assert.Equal(t, scheduleTestDomain, req.GetDomain())
					assert.Equal(t, tt.pageSize, req.GetPageSize())
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
