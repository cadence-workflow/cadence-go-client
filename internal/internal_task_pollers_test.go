// Copyright (c) 2017-2021 Uber Technologies Inc.
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
	"errors"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/metrics"
)

const (
	_testDomainName = "test-domain"
	_testTaskList   = "test-tasklist"
	_testIdentity   = "test-worker"
)

func Test_newWorkflowTaskPoller(t *testing.T) {
	t.Run("success with nil ldaTunnel", func(t *testing.T) {
		poller := newWorkflowTaskPoller(
			nil,
			nil,
			nil,
			_testDomainName,
			workerExecutionParameters{})
		assert.NotNil(t, poller)
		if poller.ldaTunnel != nil {
			t.Error("unexpected not nil ldaTunnel")
		}
	})
}

func TestLocalActivityPanic(t *testing.T) {
	// regression: panics in local activities should not terminate the process
	s := WorkflowTestSuite{logger: testlogger.NewZap(t)}
	env := s.NewTestWorkflowEnvironment()

	wf := "panicky_local_activity"
	env.RegisterWorkflowWithOptions(func(ctx Context) error {
		ctx = WithLocalActivityOptions(ctx, LocalActivityOptions{
			ScheduleToCloseTimeout: time.Second,
		})
		return ExecuteLocalActivity(ctx, func(ctx context.Context) error {
			panic("should not kill process")
		}).Get(ctx, nil)
	}, RegisterWorkflowOptions{Name: wf})

	env.ExecuteWorkflow(wf)
	err := env.GetWorkflowError()
	require.Error(t, err)
	var perr *PanicError
	require.True(t, errors.As(err, &perr), "error should be a panic error")
	assert.Contains(t, perr.StackTrace(), "panic")
	assert.Contains(t, perr.StackTrace(), t.Name(), "should mention the source location of the local activity that panicked")
}

func TestRespondTaskCompleted_failed(t *testing.T) {
	t.Run("fail sends RespondDecisionTaskFailedRequest", func(t *testing.T) {
		testTaskToken := []byte("test-task-token")

		poller, client, _, _ := buildWorkflowTaskPoller(t)
		client.EXPECT().RespondDecisionTaskFailed(gomock.Any(), &s.RespondDecisionTaskFailedRequest{
			TaskToken:      testTaskToken,
			Cause:          s.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure.Ptr(),
			Details:        []byte(assert.AnError.Error()),
			Identity:       common.StringPtr(_testIdentity),
			BinaryChecksum: common.StringPtr(getBinaryChecksum()),
		}, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

		res, err := poller.RespondTaskCompletedWithMetrics(nil, assert.AnError, &s.PollForDecisionTaskResponse{
			TaskToken: testTaskToken,
			Attempt:   common.Int64Ptr(0),
		}, time.Now())
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
	t.Run("fail fails to send RespondDecisionTaskFailedRequest", func(t *testing.T) {
		testTaskToken := []byte("test-task-token")

		poller, client, _, _ := buildWorkflowTaskPoller(t)
		client.EXPECT().RespondDecisionTaskFailed(gomock.Any(), &s.RespondDecisionTaskFailedRequest{
			TaskToken:      testTaskToken,
			Cause:          s.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure.Ptr(),
			Details:        []byte(assert.AnError.Error()),
			Identity:       common.StringPtr(_testIdentity),
			BinaryChecksum: common.StringPtr(getBinaryChecksum()),
		}, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(assert.AnError)

		// We cannot test RespondTaskCompleted since it uses backoff and has a hardcoded retry mechanism for 60 seconds.
		_, err := poller.respondTaskCompletedAttempt(errorToFailDecisionTask(testTaskToken, assert.AnError, _testIdentity), &s.PollForDecisionTaskResponse{
			TaskToken: testTaskToken,
			Attempt:   common.Int64Ptr(0),
		})
		assert.ErrorIs(t, err, assert.AnError)
	})
	t.Run("fail skips sending for not the first attempt", func(t *testing.T) {
		poller, _, _, _ := buildWorkflowTaskPoller(t)

		res, err := poller.RespondTaskCompletedWithMetrics(nil, assert.AnError, &s.PollForDecisionTaskResponse{
			Attempt: common.Int64Ptr(1),
		}, time.Now())
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestRespondTaskCompleted_Unsupported(t *testing.T) {
	poller, _, _, _ := buildWorkflowTaskPoller(t)

	assert.PanicsWithValue(t, "unknown request type from ProcessWorkflowTask()", func() {
		_, _ = poller.RespondTaskCompletedWithMetrics(assert.AnError, nil, &s.PollForDecisionTaskResponse{}, time.Now())
	})
}

func TestProcessTask_failures(t *testing.T) {
	t.Run("shutdown", func(t *testing.T) {
		poller, _, _, _ := buildWorkflowTaskPoller(t)
		ch := make(chan struct{})
		poller.shutdownC = ch
		close(ch)

		err := poller.ProcessTask(&workflowTask{})
		assert.ErrorIs(t, err, errShutdown)
	})
	t.Run("unsupported task type", func(t *testing.T) {
		poller, _, _, _ := buildWorkflowTaskPoller(t)
		assert.PanicsWithValue(t, "unknown task type.", func() {
			_ = poller.ProcessTask(10)
		})
	})
	t.Run("nil task", func(t *testing.T) {
		poller, _, _, _ := buildWorkflowTaskPoller(t)

		err := poller.ProcessTask(&workflowTask{})
		assert.NoError(t, err)
	})
	t.Run("heartbeat error", func(t *testing.T) {
		poller, _, mockedTaskHandler, _ := buildWorkflowTaskPoller(t)
		hearbeatErr := &decisionHeartbeatError{}
		mockedTaskHandler.EXPECT().ProcessWorkflowTask(mock.Anything, mock.Anything).Return(nil, hearbeatErr)
		err := poller.ProcessTask(&workflowTask{
			task: &s.PollForDecisionTaskResponse{},
		})
		assert.ErrorIs(t, err, hearbeatErr)
	})
	t.Run("ResetStickyTaskList fail", func(t *testing.T) {
		poller, client, _, _ := buildWorkflowTaskPoller(t)
		client.EXPECT().ResetStickyTaskList(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, assert.AnError)
		err := poller.ProcessTask(&resetStickinessTask{
			task: &s.ResetStickyTaskListRequest{
				Execution: &s.WorkflowExecution{
					WorkflowId: common.StringPtr("test-workflow-id"),
					RunId:      common.StringPtr("test-run-id"),
				},
			},
		})
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestActivityTaskPoller_PollTask(t *testing.T) {
	tests := []struct {
		name           string
		setupPoller    func(t *testing.T) (*activityTaskPoller, *workflowservicetest.MockClient)
		setupMocks     func(*workflowservicetest.MockClient)
		expectedResult interface{}
		expectedError  error
		validateResult func(t *testing.T, result interface{})
	}{
		{
			name: "success with valid activity task",
			setupPoller: func(t *testing.T) (*activityTaskPoller, *workflowservicetest.MockClient) {
				return buildActivityTaskPoller(t, false)
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForActivityTask(
					gomock.Any(),
					&s.PollForActivityTaskRequest{
						Domain:           common.StringPtr(_testDomainName),
						TaskList:         common.TaskListPtr(s.TaskList{Name: common.StringPtr(_testTaskList)}),
						Identity:         common.StringPtr(_testIdentity),
						TaskListMetadata: &s.TaskListMetadata{MaxTasksPerSecond: common.Float64Ptr(0.0)},
					},
					gomock.Any(),
				).Return(&s.PollForActivityTaskResponse{
					TaskToken:                       []byte("test-task-token"),
					WorkflowExecution:               &s.WorkflowExecution{WorkflowId: common.StringPtr("test-workflow")},
					ActivityId:                      common.StringPtr("test-activity"),
					ActivityType:                    &s.ActivityType{Name: common.StringPtr("TestActivity")},
					WorkflowType:                    &s.WorkflowType{Name: common.StringPtr("TestWorkflow")},
					ScheduledTimestampOfThisAttempt: common.Int64Ptr(time.Now().UnixNano()),
					StartedTimestamp:                common.Int64Ptr(time.Now().UnixNano()),
					AutoConfigHint:                  &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)},
				}, nil)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
				activityTask, ok := result.(*activityTask)
				assert.True(t, ok, "result should be *activityTask")
				assert.NotNil(t, activityTask.task)
				assert.Equal(t, []byte("test-task-token"), activityTask.task.TaskToken)
				assert.NotNil(t, activityTask.autoConfigHint)
				assert.Equal(t, &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)}, activityTask.autoConfigHint)
			},
		},
		{
			name: "success with empty task (no work available)",
			setupPoller: func(t *testing.T) (*activityTaskPoller, *workflowservicetest.MockClient) {
				return buildActivityTaskPoller(t, false)
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForActivityTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(&s.PollForActivityTaskResponse{
					TaskToken:      []byte{}, // Empty task token indicates no work
					AutoConfigHint: &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)},
				}, nil)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
				activityTask, ok := result.(*activityTask)
				assert.True(t, ok, "result should be *activityTask")
				assert.Nil(t, activityTask.task)
				assert.NotNil(t, activityTask.autoConfigHint)
				assert.Equal(t, &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)}, activityTask.autoConfigHint)
			},
		},
		{
			name: "service error during poll",
			setupPoller: func(t *testing.T) (*activityTaskPoller, *workflowservicetest.MockClient) {
				return buildActivityTaskPoller(t, false)
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForActivityTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil, &s.InternalServiceError{Message: "service unavailable"})
			},
			expectedError: &s.InternalServiceError{Message: "service unavailable"},
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "service busy error during poll",
			setupPoller: func(t *testing.T) (*activityTaskPoller, *workflowservicetest.MockClient) {
				return buildActivityTaskPoller(t, false)
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForActivityTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil, &s.ServiceBusyError{Message: "service busy"})
			},
			expectedError: &s.ServiceBusyError{Message: "service busy"},
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "poller shutting down",
			setupPoller: func(t *testing.T) (*activityTaskPoller, *workflowservicetest.MockClient) {
				return buildActivityTaskPoller(t, true) // shutdown = true
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				// No mock setup needed as doPoll should return early due to shutdown
			},
			expectedError: errShutdown,
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "context timeout during poll",
			setupPoller: func(t *testing.T) (*activityTaskPoller, *workflowservicetest.MockClient) {
				return buildActivityTaskPoller(t, false)
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForActivityTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil, context.DeadlineExceeded)
			},
			expectedError: context.DeadlineExceeded,
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller, mockService := tt.setupPoller(t)
			tt.setupMocks(mockService)

			result, err := poller.PollTask()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, result)
		})
	}
}

func buildActivityTaskPoller(t *testing.T, shutdown bool) (*activityTaskPoller, *workflowservicetest.MockClient) {
	ctrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(ctrl)

	var shutdownC <-chan struct{}
	if shutdown {
		ch := make(chan struct{})
		close(ch)
		shutdownC = ch
	} else {
		shutdownC = make(<-chan struct{})
	}

	return &activityTaskPoller{
		basePoller: basePoller{
			shutdownC: shutdownC,
		},
		domain:              _testDomainName,
		taskListName:        _testTaskList,
		identity:            _testIdentity,
		service:             mockService,
		metricsScope:        &metrics.TaggedScope{Scope: tally.NewTestScope("test", nil)},
		logger:              testlogger.NewZap(t),
		activitiesPerSecond: 0.0,
		featureFlags:        FeatureFlags{},
	}, mockService
}

func TestWorkflowTaskPoller_PollTask(t *testing.T) {
	tests := []struct {
		name           string
		setupPoller    func(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient)
		setupMocks     func(*workflowservicetest.MockClient)
		expectedResult interface{}
		expectedError  error
		validateResult func(t *testing.T, result interface{})
	}{
		{
			name: "success with valid workflow task",
			setupPoller: func(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient) {
				wP, mockService, _, _ := buildWorkflowTaskPoller(t)
				return wP, mockService
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForDecisionTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(&s.PollForDecisionTaskResponse{
					TaskToken:          []byte("test-task-token"),
					WorkflowExecution:  &s.WorkflowExecution{WorkflowId: common.StringPtr("test-workflow"), RunId: common.StringPtr("test-run")},
					WorkflowType:       &s.WorkflowType{Name: common.StringPtr("TestWorkflow")},
					StartedEventId:     common.Int64Ptr(1),
					NextEventId:        common.Int64Ptr(2),
					ScheduledTimestamp: common.Int64Ptr(time.Now().UnixNano()),
					StartedTimestamp:   common.Int64Ptr(time.Now().UnixNano()),
					History:            &s.History{Events: []*s.HistoryEvent{{EventId: common.Int64Ptr(1)}}},
					AutoConfigHint:     &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)},
				}, nil)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
				workflowTask, ok := result.(*workflowTask)
				assert.True(t, ok, "result should be *workflowTask")
				assert.NotNil(t, workflowTask.task)
				assert.Equal(t, []byte("test-task-token"), workflowTask.task.TaskToken)
				assert.NotNil(t, workflowTask.historyIterator)
				assert.Equal(t, &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)}, workflowTask.autoConfigHint)
			},
		},
		{
			name: "success with empty task (no work available)",
			setupPoller: func(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient) {
				wP, mockService, _, _ := buildWorkflowTaskPoller(t)
				return wP, mockService
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForDecisionTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(&s.PollForDecisionTaskResponse{
					TaskToken:      []byte{}, // Empty task token indicates no work
					AutoConfigHint: &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)},
				}, nil)
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
				workflowTask, ok := result.(*workflowTask)
				assert.True(t, ok, "result should be *workflowTask")
				assert.Nil(t, workflowTask.task)
				assert.NotNil(t, workflowTask.autoConfigHint)
				assert.Equal(t, &s.AutoConfigHint{PollerWaitTimeInMs: common.Int64Ptr(1000)}, workflowTask.autoConfigHint)
			},
		},
		{
			name: "service error during poll",
			setupPoller: func(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient) {
				wP, mockService, _, _ := buildWorkflowTaskPoller(t)
				return wP, mockService
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForDecisionTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil, &s.InternalServiceError{Message: "service unavailable"})
			},
			expectedError: &s.InternalServiceError{Message: "service unavailable"},
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "service busy error during poll",
			setupPoller: func(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient) {
				wP, mockService, _, _ := buildWorkflowTaskPoller(t)
				return wP, mockService
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForDecisionTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil, &s.ServiceBusyError{Message: "service busy"})
			},
			expectedError: &s.ServiceBusyError{Message: "service busy"},
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "context timeout during poll",
			setupPoller: func(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient) {
				wP, mockService, _, _ := buildWorkflowTaskPoller(t)
				return wP, mockService
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForDecisionTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil, context.DeadlineExceeded)
			},
			expectedError: context.DeadlineExceeded,
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
		{
			name: "domain not exists error during poll",
			setupPoller: func(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient) {
				wP, mockService, _, _ := buildWorkflowTaskPoller(t)
				return wP, mockService
			},
			setupMocks: func(mockService *workflowservicetest.MockClient) {
				mockService.EXPECT().PollForDecisionTask(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(nil, &s.EntityNotExistsError{Message: "domain does not exist"})
			},
			expectedError: &s.EntityNotExistsError{Message: "domain does not exist"},
			validateResult: func(t *testing.T, result interface{}) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poller, mockService := tt.setupPoller(t)
			tt.setupMocks(mockService)

			result, err := poller.PollTask()

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			tt.validateResult(t, result)
		})
	}
}

func buildWorkflowTaskPoller(t *testing.T) (*workflowTaskPoller, *workflowservicetest.MockClient, *MockWorkflowTaskHandler, *mockLocalDispatcher) {
	ctrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(ctrl)
	taskHandler := &MockWorkflowTaskHandler{}
	lda := &mockLocalDispatcher{}

	return &workflowTaskPoller{
		basePoller: basePoller{
			shutdownC: make(<-chan struct{}),
		},
		domain:                       _testDomainName,
		taskListName:                 _testTaskList,
		identity:                     _testIdentity,
		service:                      mockService,
		taskHandler:                  taskHandler,
		ldaTunnel:                    lda,
		metricsScope:                 &metrics.TaggedScope{Scope: tally.NewTestScope("test", nil)},
		logger:                       testlogger.NewZap(t),
		stickyUUID:                   "sticky-uuid",
		disableStickyExecution:       false,
		StickyScheduleToStartTimeout: time.Millisecond,
		featureFlags:                 FeatureFlags{},
	}, mockService, taskHandler, lda
}
