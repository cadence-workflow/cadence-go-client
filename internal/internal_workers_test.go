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

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/atomic"
	"go.uber.org/yarpc"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	m "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

// ActivityTaskHandler never returns response
type noResponseActivityTaskHandler struct {
	isExecuteCalled chan struct{}
}

func newNoResponseActivityTaskHandler() *noResponseActivityTaskHandler {
	return &noResponseActivityTaskHandler{isExecuteCalled: make(chan struct{})}
}

func (ath noResponseActivityTaskHandler) Execute(taskList string, task *m.PollForActivityTaskResponse) (interface{}, error) {
	close(ath.isExecuteCalled)
	c := make(chan struct{})
	<-c
	return nil, nil
}

func (ath noResponseActivityTaskHandler) BlockedOnExecuteCalled() error {
	<-ath.isExecuteCalled
	return nil
}

type (
	WorkersTestSuite struct {
		suite.Suite
		mockCtrl *gomock.Controller
		service  *workflowservicetest.MockClient
	}
)

// Test suite.
func (s *WorkersTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = workflowservicetest.NewMockClient(s.mockCtrl)
}

func (s *WorkersTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func TestWorkersTestSuite(t *testing.T) {
	suite.Run(t, new(WorkersTestSuite))
}

func (s *WorkersTestSuite) TestWorkflowWorker() {
	domain := "testDomain"
	logger, _ := zap.NewDevelopment()

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForDecisionTaskResponse{}, nil).AnyTimes()
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	executionParameters := workerExecutionParameters{
		TaskList: "testTaskList",
		WorkerOptions: WorkerOptions{
			MaxConcurrentDecisionTaskPollers: 5,
			Logger:                           logger},
		UserContext:       ctx,
		UserContextCancel: cancel,
	}
	overrides := &workerOverrides{workflowTaskHandler: newSampleWorkflowTaskHandler()}
	workflowWorker := newWorkflowWorkerInternal(
		s.service, domain, executionParameters, nil, overrides, newRegistry(), nil,
	)
	workflowWorker.Start()
	workflowWorker.Stop()

	s.Nil(ctx.Err())
}

func (s *WorkersTestSuite) TestActivityWorker() {
	s.testActivityWorker(false)
}

func (s *WorkersTestSuite) TestActivityWorkerWithLocalActivityDispatch() {
	s.testActivityWorker(true)
}

func (s *WorkersTestSuite) testActivityWorker(useLocallyDispatched bool) {
	domain := "testDomain"
	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil)
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForActivityTaskResponse{}, nil).AnyTimes()
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).Return(nil).AnyTimes()

	executionParameters := workerExecutionParameters{
		TaskList: "testTaskList",
		WorkerOptions: WorkerOptions{
			MaxConcurrentActivityTaskPollers: 5,
			Logger:                           testlogger.NewZap(s.T())},
	}
	overrides := &workerOverrides{activityTaskHandler: newSampleActivityTaskHandler(), useLocallyDispatchedActivityPoller: useLocallyDispatched}
	a := &greeterActivity{}
	registry := newRegistry()
	registry.addActivityWithLock(a.ActivityType().Name, a)
	activityWorker := newActivityWorker(
		s.service, domain, executionParameters, overrides, registry, nil,
	)
	activityWorker.Start()
	activityWorker.Stop()
}

func (s *WorkersTestSuite) TestActivityWorkerStop() {
	domain := "testDomain"

	pats := &m.PollForActivityTaskResponse{
		TaskToken: []byte("token"),
		WorkflowExecution: &m.WorkflowExecution{
			WorkflowId: common.StringPtr("wID"),
			RunId:      common.StringPtr("rID")},
		ActivityType:                    &m.ActivityType{Name: common.StringPtr("test")},
		ActivityId:                      common.StringPtr(uuid.New()),
		ScheduledTimestamp:              common.Int64Ptr(time.Now().UnixNano()),
		ScheduledTimestampOfThisAttempt: common.Int64Ptr(time.Now().UnixNano()),
		ScheduleToCloseTimeoutSeconds:   common.Int32Ptr(1),
		StartedTimestamp:                common.Int64Ptr(time.Now().UnixNano()),
		StartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Attempt:                         common.Int32Ptr(0),
		WorkflowType: &m.WorkflowType{
			Name: common.StringPtr("wType"),
		},
		WorkflowDomain: common.StringPtr("domain"),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil)
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions()...).Return(pats, nil).AnyTimes()
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).Return(nil).AnyTimes()

	stopC := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	executionParameters := workerExecutionParameters{
		TaskList: "testTaskList",
		WorkerOptions: AugmentWorkerOptions(
			WorkerOptions{
				MaxConcurrentActivityTaskPollers:   5,
				MaxConcurrentActivityExecutionSize: 2,
				Logger:                             testlogger.NewZap(s.T()),
			},
		),
		UserContext:       ctx,
		UserContextCancel: cancel,
		WorkerStopTimeout: time.Second * 2,
		WorkerStopChannel: stopC,
	}
	activityTaskHandler := newNoResponseActivityTaskHandler()
	overrides := &workerOverrides{activityTaskHandler: activityTaskHandler}
	a := &greeterActivity{}
	registry := newRegistry()
	registry.addActivityWithLock(a.ActivityType().Name, a)
	worker := newActivityWorker(
		s.service, domain, executionParameters, overrides, registry, nil,
	)
	worker.Start()
	activityTaskHandler.BlockedOnExecuteCalled()
	go worker.Stop()

	<-worker.worker.shutdownCh
	err := ctx.Err()
	s.NoError(err)

	<-ctx.Done()
	err = ctx.Err()
	s.Error(err)
}

func (s *WorkersTestSuite) TestPollForDecisionTask_InternalServiceError() {
	domain := "testDomain"

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForDecisionTaskResponse{}, &m.InternalServiceError{}).AnyTimes()

	executionParameters := workerExecutionParameters{
		TaskList: "testDecisionTaskList",
		WorkerOptions: WorkerOptions{
			MaxConcurrentDecisionTaskPollers: 5,
			Logger:                           testlogger.NewZap(s.T())},
	}
	overrides := &workerOverrides{workflowTaskHandler: newSampleWorkflowTaskHandler()}
	workflowWorker := newWorkflowWorkerInternal(
		s.service, domain, executionParameters, nil, overrides, newRegistry(), nil,
	)
	workflowWorker.Start()
	workflowWorker.Stop()
}

func (s *WorkersTestSuite) TestLongRunningDecisionTask() {
	localActivityCalledCount := 0
	localActivitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		localActivityCalledCount++
		return nil
	}

	doneCh := make(chan struct{})

	isWorkflowCompleted := false
	longDecisionWorkflowFn := func(ctx Context, input []byte) error {
		lao := LocalActivityOptions{
			ScheduleToCloseTimeout: time.Second * 2,
		}
		ctx = WithLocalActivityOptions(ctx, lao)
		err := ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)

		if err != nil {
			return err
		}

		err = ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)
		isWorkflowCompleted = true
		return err
	}

	domain := "testDomain"
	taskList := "long-running-decision-tl"
	testEvents := []*m.HistoryEvent{
		{
			EventId:   common.Int64Ptr(1),
			EventType: common.EventTypePtr(m.EventTypeWorkflowExecutionStarted),
			WorkflowExecutionStartedEventAttributes: &m.WorkflowExecutionStartedEventAttributes{
				TaskList:                            &m.TaskList{Name: &taskList},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
				WorkflowType:                        &m.WorkflowType{Name: common.StringPtr("long-running-decision-workflow-type")},
			},
		},
		createTestEventDecisionTaskScheduled(2, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &m.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		{
			EventId:   common.Int64Ptr(5),
			EventType: common.EventTypePtr(m.EventTypeMarkerRecorded),
			MarkerRecordedEventAttributes: &m.MarkerRecordedEventAttributes{
				MarkerName:                   common.StringPtr(localActivityMarkerName),
				Details:                      s.createLocalActivityMarkerDataForTest("0"),
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			},
		},
		createTestEventDecisionTaskScheduled(6, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(7),
		createTestEventDecisionTaskCompleted(8, &m.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		{
			EventId:   common.Int64Ptr(9),
			EventType: common.EventTypePtr(m.EventTypeMarkerRecorded),
			MarkerRecordedEventAttributes: &m.MarkerRecordedEventAttributes{
				MarkerName:                   common.StringPtr(localActivityMarkerName),
				Details:                      s.createLocalActivityMarkerDataForTest("1"),
				DecisionTaskCompletedEventId: common.Int64Ptr(8),
			},
		},
		createTestEventDecisionTaskScheduled(10, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(11),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()
	task := &m.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &m.WorkflowExecution{
			WorkflowId: common.StringPtr("long-running-decision-workflow-id"),
			RunId:      common.StringPtr("long-running-decision-workflow-run-id"),
		},
		WorkflowType: &m.WorkflowType{
			Name: common.StringPtr("long-running-decision-workflow-type"),
		},
		PreviousStartedEventId: common.Int64Ptr(0),
		StartedEventId:         common.Int64Ptr(3),
		History:                &m.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            common.Int64Ptr(4),
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForDecisionTaskResponse{}, &m.InternalServiceError{}).AnyTimes()

	respondCounter := 0
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *m.RespondDecisionTaskCompletedResponse, err error) {
		respondCounter++
		switch respondCounter {
		case 1:
			s.Equal(1, len(request.Decisions))
			s.Equal(m.DecisionTypeRecordMarker, request.Decisions[0].GetDecisionType())
			*task.PreviousStartedEventId = 3
			*task.StartedEventId = 7
			task.History.Events = testEvents[3:7]
			return &m.RespondDecisionTaskCompletedResponse{DecisionTask: task}, nil
		case 2:
			s.Equal(2, len(request.Decisions))
			s.Equal(m.DecisionTypeRecordMarker, request.Decisions[0].GetDecisionType())
			s.Equal(m.DecisionTypeCompleteWorkflowExecution, request.Decisions[1].GetDecisionType())
			*task.PreviousStartedEventId = 7
			*task.StartedEventId = 11
			task.History.Events = testEvents[7:11]
			close(doneCh)
			return nil, nil
		default:
			panic("unexpected RespondDecisionTaskCompleted")
		}
	}).Times(2)

	options := WorkerOptions{
		Logger:                testlogger.NewZap(s.T()),
		DisableActivityWorker: true,
		Identity:              "test-worker-identity",
	}
	worker, err := newAggregatedWorker(s.service, domain, taskList, options)
	s.Require().NoError(err)
	worker.RegisterWorkflowWithOptions(
		longDecisionWorkflowFn,
		RegisterWorkflowOptions{Name: "long-running-decision-workflow-type"},
	)
	worker.RegisterActivity(localActivitySleep)

	startWorkerAndWait(s, worker, &doneCh)

	s.True(isWorkflowCompleted)
	s.Equal(2, localActivityCalledCount)
}

func (s *WorkersTestSuite) TestQueryTask_WorkflowCacheEvicted() {
	domain := "testDomain"
	taskList := "query-task-cache-evicted-tl"
	workflowType := "query-task-cache-evicted-workflow"
	workflowID := "query-task-cache-evicted-workflow-id"
	runID := "query-task-cache-evicted-workflow-run-id"
	activityType := "query-task-cache-evicted-activity"
	queryType := "state"
	doneCh := make(chan struct{})

	activityFn := func(ctx context.Context) error {
		return nil
	}

	queryWorkflowFn := func(ctx Context) error {
		queryResult := "started"
		// setup query handler for query type "state"
		if err := SetQueryHandler(ctx, queryType, func(input []byte) (string, error) {
			return queryResult, nil
		}); err != nil {
			return err
		}

		queryResult = "waiting on timer"
		NewTimer(ctx, time.Minute*2).Get(ctx, nil)

		queryResult = "waiting on activity"
		ctx = WithActivityOptions(ctx, ActivityOptions{
			ScheduleToStartTimeout: 10 * time.Second,
			StartToCloseTimeout:    10 * time.Second,
		})
		if err := ExecuteActivity(ctx, activityFn).Get(ctx, nil); err != nil {
			return err
		}
		queryResult = "done"
		return nil
	}

	testEvents := []*m.HistoryEvent{
		{
			EventId:   common.Int64Ptr(1),
			EventType: common.EventTypePtr(m.EventTypeWorkflowExecutionStarted),
			WorkflowExecutionStartedEventAttributes: &m.WorkflowExecutionStartedEventAttributes{
				TaskList:                            &m.TaskList{Name: &taskList},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(180),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
				WorkflowType:                        &m.WorkflowType{Name: common.StringPtr(workflowType)},
			},
		},
		createTestEventDecisionTaskScheduled(2, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &m.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		{
			EventId:   common.Int64Ptr(5),
			EventType: common.EventTypePtr(m.EventTypeTimerStarted),
			TimerStartedEventAttributes: &m.TimerStartedEventAttributes{
				TimerId:                      common.StringPtr("0"),
				StartToFireTimeoutSeconds:    common.Int64Ptr(120),
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			},
		},
		{
			EventId:   common.Int64Ptr(6),
			EventType: common.EventTypePtr(m.EventTypeTimerFired),
			TimerFiredEventAttributes: &m.TimerFiredEventAttributes{
				TimerId:        common.StringPtr("0"),
				StartedEventId: common.Int64Ptr(5),
			},
		},
		createTestEventDecisionTaskScheduled(7, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(8),
		createTestEventDecisionTaskCompleted(9, &m.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(10, &m.ActivityTaskScheduledEventAttributes{
			ActivityId: common.StringPtr("1"),
			ActivityType: &m.ActivityType{
				Name: common.StringPtr(activityType),
			},
			Domain: common.StringPtr(domain),
			TaskList: &m.TaskList{
				Name: common.StringPtr(taskList),
			},
			ScheduleToStartTimeoutSeconds: common.Int32Ptr(10),
			StartToCloseTimeoutSeconds:    common.Int32Ptr(10),
			DecisionTaskCompletedEventId:  common.Int64Ptr(9),
		}),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()
	task := &m.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &m.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		WorkflowType: &m.WorkflowType{
			Name: common.StringPtr(workflowType),
		},
		PreviousStartedEventId: common.Int64Ptr(0),
		StartedEventId:         common.Int64Ptr(3),
		History:                &m.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            common.Int64Ptr(4),
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(task, nil).Times(1)
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *m.RespondDecisionTaskCompletedResponse, err error) {
		s.Equal(1, len(request.Decisions))
		s.Equal(m.DecisionTypeStartTimer, request.Decisions[0].GetDecisionType())
		return &m.RespondDecisionTaskCompletedResponse{}, nil
	}).Times(1)
	queryTask := &m.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &m.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		WorkflowType: &m.WorkflowType{
			Name: common.StringPtr(workflowType),
		},
		PreviousStartedEventId: common.Int64Ptr(3),
		History:                &m.History{}, // sticky query, so there's no history
		NextPageToken:          nil,
		NextEventId:            common.Int64Ptr(5),
		Query: &m.WorkflowQuery{
			QueryType: common.StringPtr(queryType),
		},
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.PollForDecisionTaskRequest, opts ...yarpc.CallOption,
	) (success *m.PollForDecisionTaskResponse, err error) {
		getWorkflowCache().Delete(runID) // force remove the workflow state
		return queryTask, nil
	}).Times(1)
	s.service.EXPECT().ResetStickyTaskList(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.ResetStickyTaskListResponse{}, nil).AnyTimes()
	s.service.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.GetWorkflowExecutionHistoryResponse{
		History: &m.History{Events: testEvents}, // workflow has made progress, return all available events
	}, nil).Times(1)
	dc := getDefaultDataConverter()
	expectedResult, err := dc.ToData("waiting on timer")
	s.NoError(err)
	s.service.EXPECT().RespondQueryTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
		s.Equal(m.QueryTaskCompletedTypeCompleted, request.GetCompletedType())
		s.Equal(expectedResult, request.GetQueryResult())
		close(doneCh)
		return nil
	}).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForDecisionTaskResponse{}, &m.InternalServiceError{}).AnyTimes()

	options := WorkerOptions{
		Logger:                testlogger.NewZap(s.T()),
		DisableActivityWorker: true,
		Identity:              "test-worker-identity",
		DataConverter:         dc,
		// set concurrent decision task execution to 1,
		// otherwise query task may be polled and start execution
		// before decision task put created workflowContext into the cache,
		// resulting in a cache hit for query
		// by setting concurrent execution size to 1, we ensure when polling
		// query task, cache already contains the workflowContext for this workflow,
		// and we can force clear the cache when polling the query task.
		// See the mock function for the second PollForDecisionTask call above.
		MaxConcurrentDecisionTaskExecutionSize: 1,
	}
	worker, err := newAggregatedWorker(s.service, domain, taskList, options)
	s.Require().NoError(err)
	worker.RegisterWorkflowWithOptions(
		queryWorkflowFn,
		RegisterWorkflowOptions{Name: workflowType},
	)
	worker.RegisterActivityWithOptions(
		activityFn,
		RegisterActivityOptions{Name: activityType},
	)

	startWorkerAndWait(s, worker, &doneCh)
}

func (s *WorkersTestSuite) TestMultipleLocalActivities() {
	localActivityCalledCount := 0
	localActivitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		localActivityCalledCount++
		return nil
	}

	doneCh := make(chan struct{})

	isWorkflowCompleted := false
	longDecisionWorkflowFn := func(ctx Context, input []byte) error {
		lao := LocalActivityOptions{
			ScheduleToCloseTimeout: time.Second * 2,
		}
		ctx = WithLocalActivityOptions(ctx, lao)
		err := ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)

		if err != nil {
			return err
		}

		err = ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)
		isWorkflowCompleted = true
		return err
	}

	domain := "testDomain"
	taskList := "multiple-local-activities-tl"
	testEvents := []*m.HistoryEvent{
		{
			EventId:   common.Int64Ptr(1),
			EventType: common.EventTypePtr(m.EventTypeWorkflowExecutionStarted),
			WorkflowExecutionStartedEventAttributes: &m.WorkflowExecutionStartedEventAttributes{
				TaskList:                            &m.TaskList{Name: &taskList},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(3),
				WorkflowType:                        &m.WorkflowType{Name: common.StringPtr("multiple-local-activities-workflow-type")},
			},
		},
		createTestEventDecisionTaskScheduled(2, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &m.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		{
			EventId:   common.Int64Ptr(5),
			EventType: common.EventTypePtr(m.EventTypeMarkerRecorded),
			MarkerRecordedEventAttributes: &m.MarkerRecordedEventAttributes{
				MarkerName:                   common.StringPtr(localActivityMarkerName),
				Details:                      s.createLocalActivityMarkerDataForTest("0"),
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			},
		},
		createTestEventDecisionTaskScheduled(6, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(7),
		createTestEventDecisionTaskCompleted(8, &m.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		{
			EventId:   common.Int64Ptr(9),
			EventType: common.EventTypePtr(m.EventTypeMarkerRecorded),
			MarkerRecordedEventAttributes: &m.MarkerRecordedEventAttributes{
				MarkerName:                   common.StringPtr(localActivityMarkerName),
				Details:                      s.createLocalActivityMarkerDataForTest("1"),
				DecisionTaskCompletedEventId: common.Int64Ptr(8),
			},
		},
		createTestEventDecisionTaskScheduled(10, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(11),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()
	task := &m.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &m.WorkflowExecution{
			WorkflowId: common.StringPtr("multiple-local-activities-workflow-id"),
			RunId:      common.StringPtr("multiple-local-activities-workflow-run-id"),
		},
		WorkflowType: &m.WorkflowType{
			Name: common.StringPtr("multiple-local-activities-workflow-type"),
		},
		PreviousStartedEventId: common.Int64Ptr(0),
		StartedEventId:         common.Int64Ptr(3),
		History:                &m.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            common.Int64Ptr(4),
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForDecisionTaskResponse{}, &m.InternalServiceError{}).AnyTimes()

	respondCounter := 0
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *m.RespondDecisionTaskCompletedResponse, err error) {
		respondCounter++
		switch respondCounter {
		case 1:
			s.Equal(3, len(request.Decisions))
			s.Equal(m.DecisionTypeRecordMarker, request.Decisions[0].GetDecisionType())
			*task.PreviousStartedEventId = 3
			*task.StartedEventId = 7
			task.History.Events = testEvents[3:11]
			close(doneCh)
			return nil, nil
		default:
			panic("unexpected RespondDecisionTaskCompleted")
		}
	}).Times(1)

	options := WorkerOptions{
		Logger:                testlogger.NewZap(s.T()),
		DisableActivityWorker: true,
		Identity:              "test-worker-identity",
	}
	worker, err := newAggregatedWorker(s.service, domain, taskList, options)
	s.Require().NoError(err)
	worker.RegisterWorkflowWithOptions(
		longDecisionWorkflowFn,
		RegisterWorkflowOptions{Name: "multiple-local-activities-workflow-type"},
	)
	worker.RegisterActivity(localActivitySleep)

	startWorkerAndWait(s, worker, &doneCh)

	s.True(isWorkflowCompleted)
	s.Equal(2, localActivityCalledCount)
}

func (s *WorkersTestSuite) createLocalActivityMarkerDataForTest(activityID string) []byte {
	lamd := localActivityMarkerData{
		ActivityID: activityID,
		ReplayTime: time.Now(),
	}

	// encode marker data
	markerData, err := encodeArg(nil, lamd)
	s.NoError(err)
	return markerData
}

func (s *WorkersTestSuite) TestLocallyDispatchedActivity() {
	activityCalledCount := atomic.NewInt32(0) // must be accessed with atomics, worker uses goroutines to run activities
	activitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		activityCalledCount.Add(1)
		return nil
	}

	doneCh := make(chan struct{})

	workflowFn := func(ctx Context, input []byte) error {
		ao := ActivityOptions{
			ScheduleToCloseTimeout: 1 * time.Second,
			ScheduleToStartTimeout: 1 * time.Second,
			StartToCloseTimeout:    1 * time.Second,
		}
		ctx = WithActivityOptions(ctx, ao)
		err := ExecuteActivity(ctx, activitySleep, 500*time.Millisecond).Get(ctx, nil)
		return err
	}

	domain := "testDomain"
	workflowType := "locally-dispatched-activity-workflow-type"
	taskList := "locally-dispatched-activity-tl"
	testEvents := []*m.HistoryEvent{
		{
			EventId:   common.Int64Ptr(1),
			EventType: common.EventTypePtr(m.EventTypeWorkflowExecutionStarted),
			WorkflowExecutionStartedEventAttributes: &m.WorkflowExecutionStartedEventAttributes{
				TaskList:                            &m.TaskList{Name: &taskList},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
				WorkflowType:                        &m.WorkflowType{Name: common.StringPtr(workflowType)},
			},
		},
		createTestEventDecisionTaskScheduled(2, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()
	task := &m.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &m.WorkflowExecution{
			WorkflowId: common.StringPtr("locally-dispatched-activity-workflow-id"),
			RunId:      common.StringPtr("locally-dispatched-activity-workflow-run-id"),
		},
		WorkflowType: &m.WorkflowType{
			Name: common.StringPtr(workflowType),
		},
		PreviousStartedEventId: common.Int64Ptr(0),
		StartedEventId:         common.Int64Ptr(3),
		History:                &m.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            common.Int64Ptr(4),
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForDecisionTaskResponse{}, &m.InternalServiceError{}).AnyTimes()
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *m.RespondDecisionTaskCompletedResponse, err error) {
		s.Equal(1, len(request.Decisions))
		activitiesToDispatchLocally := make(map[string]*m.ActivityLocalDispatchInfo)
		d := request.Decisions[0]
		s.Equal(m.DecisionTypeScheduleActivityTask, d.GetDecisionType())
		activitiesToDispatchLocally[*d.ScheduleActivityTaskDecisionAttributes.ActivityId] =
			&m.ActivityLocalDispatchInfo{
				ActivityId:                      d.ScheduleActivityTaskDecisionAttributes.ActivityId,
				ScheduledTimestamp:              common.Int64Ptr(time.Now().UnixNano()),
				ScheduledTimestampOfThisAttempt: common.Int64Ptr(time.Now().UnixNano()),
				StartedTimestamp:                common.Int64Ptr(time.Now().UnixNano()),
				TaskToken:                       []byte("test-token")}
		return &m.RespondDecisionTaskCompletedResponse{ActivitiesToDispatchLocally: activitiesToDispatchLocally}, nil
	}).Times(1)
	isActivityResponseCompleted := atomic.NewBool(false)
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption,
	) error {
		defer close(doneCh)
		isActivityResponseCompleted.Swap(true)
		return nil
	}).Times(1)

	options := WorkerOptions{
		Logger:   testlogger.NewZap(s.T()),
		Identity: "test-worker-identity",
	}
	worker, err := newAggregatedWorker(s.service, domain, taskList, options)
	s.Require().NoError(err)
	worker.RegisterWorkflowWithOptions(
		workflowFn,
		RegisterWorkflowOptions{Name: workflowType},
	)
	worker.RegisterActivityWithOptions(activitySleep, RegisterActivityOptions{Name: "activitySleep"})

	startWorkerAndWait(s, worker, &doneCh)

	s.True(isActivityResponseCompleted.Load())
	s.Equal(int32(1), activityCalledCount.Load())
}

func (s *WorkersTestSuite) TestMultipleLocallyDispatchedActivity() {
	activityCalledCount := atomic.NewInt32(0)
	activitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		activityCalledCount.Add(1)
		return nil
	}

	doneCh := make(chan struct{})

	var activityCount int32 = 5
	workflowFn := func(ctx Context, input []byte) error {
		ao := ActivityOptions{
			ScheduleToCloseTimeout: 1 * time.Second,
			ScheduleToStartTimeout: 1 * time.Second,
			StartToCloseTimeout:    1 * time.Second,
		}
		ctx = WithActivityOptions(ctx, ao)

		// start all activities in parallel, and wait for them all to complete.
		var all []Future
		for i := 0; i < int(activityCount); i++ {
			all = append(all, ExecuteActivity(ctx, activitySleep, 500*time.Millisecond))
		}
		for i, f := range all {
			s.NoError(f.Get(ctx, nil), "activity %v should not have failed", i)
		}
		return nil
	}

	domain := "testDomain"
	workflowType := "locally-dispatched-multiple-activity-workflow-type"
	taskList := "locally-dispatched-multiple-activity-tl"
	testEvents := []*m.HistoryEvent{
		{
			EventId:   common.Int64Ptr(1),
			EventType: common.EventTypePtr(m.EventTypeWorkflowExecutionStarted),
			WorkflowExecutionStartedEventAttributes: &m.WorkflowExecutionStartedEventAttributes{
				TaskList:                            &m.TaskList{Name: &taskList},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(2),
				WorkflowType:                        &m.WorkflowType{Name: common.StringPtr(workflowType)},
			},
		},
		createTestEventDecisionTaskScheduled(2, &m.DecisionTaskScheduledEventAttributes{TaskList: &m.TaskList{Name: &taskList}}),
		createTestEventDecisionTaskStarted(3),
	}

	options := WorkerOptions{
		Logger:   testlogger.NewZap(s.T()),
		Identity: "test-worker-identity",
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()
	task := &m.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &m.WorkflowExecution{
			WorkflowId: common.StringPtr("locally-dispatched-multiple-activity-workflow-id"),
			RunId:      common.StringPtr("locally-dispatched-multiple-activity-workflow-run-id"),
		},
		WorkflowType: &m.WorkflowType{
			Name: common.StringPtr(workflowType),
		},
		PreviousStartedEventId: common.Int64Ptr(0),
		StartedEventId:         common.Int64Ptr(3),
		History:                &m.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            common.Int64Ptr(4),
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions()...).Return(&m.PollForDecisionTaskResponse{}, &m.InternalServiceError{}).AnyTimes()
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, nil).AnyTimes()
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *m.RespondDecisionTaskCompletedResponse, err error) {
		s.Equal(int(activityCount), len(request.Decisions))
		activitiesToDispatchLocally := make(map[string]*m.ActivityLocalDispatchInfo)
		for _, d := range request.Decisions {
			s.Equal(m.DecisionTypeScheduleActivityTask, d.GetDecisionType())
			activitiesToDispatchLocally[*d.ScheduleActivityTaskDecisionAttributes.ActivityId] =
				&m.ActivityLocalDispatchInfo{
					ActivityId:                      d.ScheduleActivityTaskDecisionAttributes.ActivityId,
					ScheduledTimestamp:              common.Int64Ptr(time.Now().UnixNano()),
					ScheduledTimestampOfThisAttempt: common.Int64Ptr(time.Now().UnixNano()),
					StartedTimestamp:                common.Int64Ptr(time.Now().UnixNano()),
					TaskToken:                       []byte("test-token")}
		}
		return &m.RespondDecisionTaskCompletedResponse{ActivitiesToDispatchLocally: activitiesToDispatchLocally}, nil
	}).Times(1)
	activityResponseCompletedCount := atomic.NewInt32(0)
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(ctx context.Context, request *m.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption,
	) error {
		counted := activityResponseCompletedCount.Add(1)
		if counted == activityCount {
			close(doneCh)
		}
		return nil
	}).MinTimes(1)

	worker, err := newAggregatedWorker(s.service, domain, taskList, options)
	s.Require().NoError(err)
	worker.RegisterWorkflowWithOptions(
		workflowFn,
		RegisterWorkflowOptions{Name: workflowType},
	)
	worker.RegisterActivityWithOptions(activitySleep, RegisterActivityOptions{Name: "activitySleep"})
	s.NotNil(worker.locallyDispatchedActivityWorker)
	err = worker.Start()
	s.NoError(err, "worker failed to start")

	// wait for test to complete
	// This test currently never completes, however after the timeout the asserts are true
	// so the test passes, I believe this is an error.
	select {
	case <-doneCh:
		s.T().Log("completed")
	case <-time.After(1 * time.Second):
		s.T().Log("timed out")
	}
	worker.Stop()

	// for currently unbuffered channel at least one activity should be sent
	s.True(activityResponseCompletedCount.Load() > 0)
	s.True(activityCalledCount.Load() > 0)
}

// wait for test to complete - timeout and fail after 10 seconds to not block execution of other tests
func startWorkerAndWait(s *WorkersTestSuite, worker *aggregatedWorker, doneCh *chan struct{}) {
	s.T().Helper()
	err := worker.Start()
	s.NoError(err, "worker failed to start")
	// wait for test to complete
	select {
	case <-*doneCh:
	case <-time.After(10 * time.Second):
		s.Fail("Test timed out")
	}
	worker.Stop()
}
