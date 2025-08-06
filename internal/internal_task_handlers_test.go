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
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

const (
	testDomain = "test-domain"
)

type (
	TaskHandlersTestSuite struct {
		suite.Suite
		logger   *zap.Logger
		service  *workflowservicetest.MockClient
		registry *registry
	}
)

func registerWorkflows(r *registry) {
	r.RegisterWorkflowWithOptions(
		helloWorldWorkflowFunc,
		RegisterWorkflowOptions{Name: "HelloWorld_Workflow"},
	)
	r.RegisterWorkflowWithOptions(
		helloWorldWorkflowCancelFunc,
		RegisterWorkflowOptions{Name: "HelloWorld_WorkflowCancel"},
	)
	r.RegisterWorkflowWithOptions(
		returnPanicWorkflowFunc,
		RegisterWorkflowOptions{Name: "ReturnPanicWorkflow"},
	)
	r.RegisterWorkflowWithOptions(
		panicWorkflowFunc,
		RegisterWorkflowOptions{Name: "PanicWorkflow"},
	)
	r.RegisterWorkflowWithOptions(
		getWorkflowInfoWorkflowFunc,
		RegisterWorkflowOptions{Name: "GetWorkflowInfoWorkflow"},
	)
	r.RegisterWorkflowWithOptions(
		querySignalWorkflowFunc,
		RegisterWorkflowOptions{Name: "QuerySignalWorkflow"},
	)
	r.RegisterActivityWithOptions(
		greeterActivityFunc,
		RegisterActivityOptions{Name: "Greeter_Activity"},
	)
	r.RegisterWorkflowWithOptions(
		binaryChecksumWorkflowFunc,
		RegisterWorkflowOptions{Name: "BinaryChecksumWorkflow"},
	)
}

func returnPanicWorkflowFunc(ctx Context, input []byte) error {
	return newPanicError("panicError", "stackTrace")
}

func panicWorkflowFunc(ctx Context, input []byte) error {
	panic("panicError")
}

func getWorkflowInfoWorkflowFunc(ctx Context, expectedLastCompletionResult string) (info *WorkflowInfo, err error) {
	result := GetWorkflowInfo(ctx)
	var lastCompletionResult string
	err = getDefaultDataConverter().FromData(result.lastCompletionResult, &lastCompletionResult)
	if err != nil {
		return nil, err
	}
	if lastCompletionResult != expectedLastCompletionResult {
		return nil, errors.New("lastCompletionResult is not " + expectedLastCompletionResult)
	}
	return result, nil
}

// Test suite.
func (t *TaskHandlersTestSuite) SetupTest() {
	t.logger = testlogger.NewZap(t.T())
}

func (t *TaskHandlersTestSuite) SetupSuite() {
	t.logger = testlogger.NewZap(t.T())
	registerWorkflows(t.registry)
}

func TestTaskHandlersTestSuite(t *testing.T) {
	suite.Run(t, &TaskHandlersTestSuite{
		registry: newRegistry(),
	})
}

func createTestEventWorkflowExecutionCompleted(eventID int64, attr *s.WorkflowExecutionCompletedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{EventId: common.Int64Ptr(eventID), EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionCompleted), WorkflowExecutionCompletedEventAttributes: attr}
}

func createTestEventWorkflowExecutionContinuedAsNew(eventID int64, attr *s.WorkflowExecutionContinuedAsNewEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{EventId: common.Int64Ptr(eventID), EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionContinuedAsNew), WorkflowExecutionContinuedAsNewEventAttributes: attr}
}

func createTestEventWorkflowExecutionStarted(eventID int64, attr *s.WorkflowExecutionStartedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{EventId: common.Int64Ptr(eventID), EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionStarted), WorkflowExecutionStartedEventAttributes: attr}
}

func createTestEventLocalActivity(eventID int64, attr *s.MarkerRecordedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                       common.Int64Ptr(eventID),
		EventType:                     common.EventTypePtr(s.EventTypeMarkerRecorded),
		MarkerRecordedEventAttributes: attr}
}

func createTestEventActivityTaskScheduled(eventID int64, attr *s.ActivityTaskScheduledEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeActivityTaskScheduled),
		ActivityTaskScheduledEventAttributes: attr}
}

func createTestEventActivityTaskStarted(eventID int64, attr *s.ActivityTaskStartedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                            common.Int64Ptr(eventID),
		EventType:                          common.EventTypePtr(s.EventTypeActivityTaskStarted),
		ActivityTaskStartedEventAttributes: attr}
}

func createTestEventActivityTaskCompleted(eventID int64, attr *s.ActivityTaskCompletedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeActivityTaskCompleted),
		ActivityTaskCompletedEventAttributes: attr}
}

func createTestEventActivityTaskTimedOut(eventID int64, attr *s.ActivityTaskTimedOutEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                             common.Int64Ptr(eventID),
		EventType:                           common.EventTypePtr(s.EventTypeActivityTaskTimedOut),
		ActivityTaskTimedOutEventAttributes: attr}
}

func createTestEventDecisionTaskScheduled(eventID int64, attr *s.DecisionTaskScheduledEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeDecisionTaskScheduled),
		DecisionTaskScheduledEventAttributes: attr}
}

func createTestEventDecisionTaskStarted(eventID int64) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:   common.Int64Ptr(eventID),
		EventType: common.EventTypePtr(s.EventTypeDecisionTaskStarted)}
}

func createTestEventWorkflowExecutionSignaled(eventID int64, signalName string) *s.HistoryEvent {
	return createTestEventWorkflowExecutionSignaledWithPayload(eventID, signalName, nil)
}

func createTestEventWorkflowExecutionSignaledWithPayload(eventID int64, signalName string, payload []byte) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:   common.Int64Ptr(eventID),
		EventType: common.EventTypePtr(s.EventTypeWorkflowExecutionSignaled),
		WorkflowExecutionSignaledEventAttributes: &s.WorkflowExecutionSignaledEventAttributes{
			SignalName: common.StringPtr(signalName),
			Input:      payload,
			Identity:   common.StringPtr("test-identity"),
		},
	}
}

func createTestEventDecisionTaskCompleted(eventID int64, attr *s.DecisionTaskCompletedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                              common.Int64Ptr(eventID),
		EventType:                            common.EventTypePtr(s.EventTypeDecisionTaskCompleted),
		DecisionTaskCompletedEventAttributes: attr}
}

func createTestEventDecisionTaskFailed(eventID int64, attr *s.DecisionTaskFailedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:                           common.Int64Ptr(eventID),
		EventType:                         common.EventTypePtr(s.EventTypeDecisionTaskFailed),
		DecisionTaskFailedEventAttributes: attr}
}

func createTestEventSignalExternalWorkflowExecutionFailed(eventID int64, attr *s.SignalExternalWorkflowExecutionFailedEventAttributes) *s.HistoryEvent {
	return &s.HistoryEvent{
		EventId:   common.Int64Ptr(eventID),
		EventType: common.EventTypePtr(s.EventTypeSignalExternalWorkflowExecutionFailed),
		SignalExternalWorkflowExecutionFailedEventAttributes: attr,
	}
}

func createWorkflowTask(
	events []*s.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
) *s.PollForDecisionTaskResponse {
	return createWorkflowTaskWithQueries(events, previousStartEventID, workflowName, nil)
}

func createWorkflowTaskWithQueries(
	events []*s.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
	queries map[string]*s.WorkflowQuery,
) *s.PollForDecisionTaskResponse {
	eventsCopy := make([]*s.HistoryEvent, len(events))
	copy(eventsCopy, events)
	return &s.PollForDecisionTaskResponse{
		PreviousStartedEventId: common.Int64Ptr(previousStartEventID),
		WorkflowType:           workflowTypePtr(WorkflowType{Name: workflowName}),
		History:                &s.History{Events: eventsCopy},
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr("fake-workflow-id"),
			RunId:      common.StringPtr(uuid.New()),
		},
		Queries: queries,
	}
}

func createQueryTask(
	events []*s.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
	queryType string,
) *s.PollForDecisionTaskResponse {
	task := createWorkflowTask(events, previousStartEventID, workflowName)
	task.Query = &s.WorkflowQuery{
		QueryType: common.StringPtr(queryType),
	}
	return task
}

func createTestEventTimerStarted(eventID int64, id int) *s.HistoryEvent {
	timerID := fmt.Sprintf("%v", id)
	attr := &s.TimerStartedEventAttributes{
		TimerId:                      common.StringPtr(timerID),
		StartToFireTimeoutSeconds:    nil,
		DecisionTaskCompletedEventId: nil,
	}
	return &s.HistoryEvent{
		EventId:                     common.Int64Ptr(eventID),
		EventType:                   common.EventTypePtr(s.EventTypeTimerStarted),
		TimerStartedEventAttributes: attr}
}

func createTestEventTimerFired(eventID int64, id int) *s.HistoryEvent {
	timerID := fmt.Sprintf("%v", id)
	attr := &s.TimerFiredEventAttributes{
		TimerId: common.StringPtr(timerID),
	}

	return &s.HistoryEvent{
		EventId:                   common.Int64Ptr(eventID),
		EventType:                 common.EventTypePtr(s.EventTypeTimerFired),
		TimerFiredEventAttributes: attr}
}

func findLogField(entry observer.LoggedEntry, fieldName string) *zapcore.Field {
	for _, field := range entry.Context {
		if field.Key == fieldName {
			return &field
		}
	}
	return nil
}

var testWorkflowTaskTasklist = "tl1"

func (t *TaskHandlersTestSuite) testWorkflowTaskWorkflowExecutionStartedHelper(params workerExecutionParameters) {
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: &testWorkflowTaskTasklist}}),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_Workflow")
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeScheduleActivityTask, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].ScheduleActivityTaskDecisionAttributes)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowExecutionStarted() {
	params := workerExecutionParameters{
		TaskList: &s.TaskList{Name: common.StringPtr(testWorkflowTaskTasklist), Kind: s.TaskListKindNormal.Ptr()},
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}
	t.testWorkflowTaskWorkflowExecutionStartedHelper(params)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowExecutionStarted_WithDataConverter() {
	params := workerExecutionParameters{
		TaskList: &s.TaskList{Name: common.StringPtr(testWorkflowTaskTasklist), Kind: s.TaskListKindNormal.Ptr()},
		WorkerOptions: WorkerOptions{
			Identity:      "test-id-1",
			Logger:        t.logger,
			DataConverter: newTestDataConverter(),
		},
	}
	t.testWorkflowTaskWorkflowExecutionStartedHelper(params)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_BinaryChecksum() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	checksum1 := "chck1"
	checksum2 := "chck2"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{
			ScheduledEventId: common.Int64Ptr(2), BinaryChecksum: common.StringPtr(checksum1)}),
		createTestEventTimerStarted(5, 0),
		createTestEventTimerFired(6, 0),
		createTestEventDecisionTaskScheduled(7, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(8),
		createTestEventDecisionTaskCompleted(9, &s.DecisionTaskCompletedEventAttributes{
			ScheduledEventId: common.Int64Ptr(7), BinaryChecksum: common.StringPtr(checksum2)}),
		createTestEventTimerStarted(10, 1),
		createTestEventTimerFired(11, 1),
		createTestEventDecisionTaskScheduled(12, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(13),
	}
	task := createWorkflowTask(testEvents, 8, "BinaryChecksumWorkflow")
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)

	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeCompleteWorkflowExecution, response.Decisions[0].GetDecisionType())
	checksumsBytes := response.Decisions[0].CompleteWorkflowExecutionDecisionAttributes.Result
	var checksums []string
	json.Unmarshal(checksumsBytes, &checksums)
	t.Equal(3, len(checksums))
	t.Equal("chck1", checksums[0])
	t.Equal("chck2", checksums[1])
	t.Equal(getBinaryChecksum(), checksums[2])
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_ActivityTaskScheduled() {
	// Schedule an activity and see if we complete workflow.
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("Greeter_Activity")},
			TaskList:     taskList,
		}),
		createTestEventActivityTaskStarted(6, &s.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &s.ActivityTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(5)}),
		createTestEventDecisionTaskStarted(8),
	}
	task := createWorkflowTask(testEvents[0:3], 0, "HelloWorld_Workflow")
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)

	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeScheduleActivityTask, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].ScheduleActivityTaskDecisionAttributes)

	// Schedule an activity and see if we complete workflow, Having only one last decision.
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response = request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeCompleteWorkflowExecution, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].CompleteWorkflowExecutionDecisionAttributes)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_QueryWorkflow() {
	// Schedule an activity and see if we complete workflow.
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	execution := &s.WorkflowExecution{
		WorkflowId: common.StringPtr("fake-workflow-id"),
		RunId:      common.StringPtr(uuid.New()),
	}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("Greeter_Activity")},
			TaskList:     taskList,
		}),
		createTestEventActivityTaskStarted(6, &s.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &s.ActivityTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(5)}),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)

	// first make progress on the workflow
	task := createWorkflowTask(testEvents[0:1], 0, "HelloWorld_Workflow")
	task.StartedEventId = common.Int64Ptr(1)
	task.WorkflowExecution = execution
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeScheduleActivityTask, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].ScheduleActivityTaskDecisionAttributes)

	// then check the current state using query task
	task = createQueryTask([]*s.HistoryEvent{}, 6, "HelloWorld_Workflow", queryType)
	task.WorkflowExecution = execution
	queryResp, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.verifyQueryResult(queryResp, "waiting-activity-result")
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_QueryWorkflow_2() {
	// Schedule an activity and see if we complete workflow.

	// This test appears to be just a finer-grained version of TestWorkflowTask_QueryWorkflow, though the older names
	// for them implied entirely different purposes.  Likely it can be combined with TestWorkflowTask_QueryWorkflow
	// without losing anything useful.
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("Greeter_Activity")},
			TaskList:     taskList,
		}),
		createTestEventActivityTaskStarted(6, &s.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &s.ActivityTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(5)}),
		// TODO: below seems irrational.  there's a start without a schedule, and this workflow does not respond to signals.
		// aside from this, the list of tasks is the same as TestWorkflowTask_QueryWorkflow
		createTestEventDecisionTaskStarted(8),
		createTestEventWorkflowExecutionSignaled(9, "test-signal"),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}

	// TODO: the following comment is not true, previousStartEventID is either 0 or points to a valid decisionTaskStartedID
	// we need to fix the test
	//
	// query after first decision task (notice the previousStartEventID is always the last eventID for query task)
	task := createQueryTask(testEvents[0:3], 3, "HelloWorld_Workflow", queryType)
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "waiting-activity-result")

	// query after activity task complete but before second decision task started
	task = createQueryTask(testEvents[0:7], 7, "HelloWorld_Workflow", queryType)
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "waiting-activity-result")

	// query after second decision task
	task = createQueryTask(testEvents[0:8], 8, "HelloWorld_Workflow", queryType)
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "done")

	// query after second decision task with extra events
	task = createQueryTask(testEvents[0:9], 9, "HelloWorld_Workflow", queryType)
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "done")

	task = createQueryTask(testEvents[0:9], 9, "HelloWorld_Workflow", "invalid-query-type")
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NotNil(response)
	queryResp, ok := response.(*s.RespondQueryTaskCompletedRequest)
	t.True(ok)
	t.NotNil(queryResp.ErrorMessage)
	t.Contains(*queryResp.ErrorMessage, "unknown queryType")
}

func (t *TaskHandlersTestSuite) verifyQueryResult(response interface{}, expectedResult string) {
	t.NotNil(response)
	queryResp, ok := response.(*s.RespondQueryTaskCompletedRequest)
	t.True(ok)
	t.Nil(queryResp.ErrorMessage)
	t.NotNil(queryResp.QueryResult)
	encodedValue := newEncodedValue(queryResp.QueryResult, nil)
	var queryResult string
	err := encodedValue.Get(&queryResult)
	t.NoError(err)
	t.Equal(expectedResult, queryResult)
}

func (t *TaskHandlersTestSuite) TestCacheEvictionWhenErrorOccurs() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("pkg.Greeter_Activity")},
			TaskList:     taskList,
		}),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         zap.NewNop(),
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	// now change the history event so it does not match to decision produced via replay
	testEvents[4].ActivityTaskScheduledEventAttributes.ActivityType.Name = common.StringPtr("some-other-activity")
	task := createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
	// will fail as it can't find laTunnel in getWorkflowCache().
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "nondeterministic")

	// There should be nothing in the cache.
	t.EqualValues(getWorkflowCache().Size(), 0)
}

func (t *TaskHandlersTestSuite) TestWithMissingHistoryEvents() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventDecisionTaskScheduled(6, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(7),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         zap.NewNop(),
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
	}

	for _, startEventID := range []int64{0, 3} {
		taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
		task := createWorkflowTask(testEvents, startEventID, "HelloWorld_Workflow")
		// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
		// will fail as it can't find laTunnel in getWorkflowCache().
		newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
		request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

		t.Error(err)
		t.Nil(request)
		t.Contains(err.Error(), "missing history events")

		// There should be nothing in the cache.
		t.EqualValues(getWorkflowCache().Size(), 0)
	}
}

func (t *TaskHandlersTestSuite) TestWithTruncatedHistory() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskFailed(4, &s.DecisionTaskFailedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventDecisionTaskScheduled(5, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(6),
		createTestEventDecisionTaskCompleted(7, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(5)}),
		createTestEventActivityTaskScheduled(8, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("pkg.Greeter_Activity")},
			TaskList:     taskList,
		}),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         zap.NewNop(),
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
	}

	testCases := []struct {
		startedEventID         int64
		previousStartedEventID int64
		isResultErr            bool
	}{
		{0, 6, false},
		{10, 0, true},
		{10, 6, true},
	}

	for i, tc := range testCases {
		taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
		task := createWorkflowTask(testEvents, tc.previousStartedEventID, "HelloWorld_Workflow")
		task.StartedEventId = common.Int64Ptr(tc.startedEventID)
		// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
		// will fail as it can't find laTunnel in getWorkflowCache().
		newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
		request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

		if tc.isResultErr {
			t.Error(err, "testcase %v failed", i)
			t.Nil(request)
			t.Contains(err.Error(), "premature end of stream")
			t.EqualValues(getWorkflowCache().Size(), 1)
			continue
		}

		t.NoError(err, "testcase %v failed", i)
	}
}

func (t *TaskHandlersTestSuite) TestSideEffectDefer_Sticky() {
	t.testSideEffectDeferHelper(false)
}

func (t *TaskHandlersTestSuite) TestSideEffectDefer_NonSticky() {
	t.testSideEffectDeferHelper(true)
}

func (t *TaskHandlersTestSuite) testSideEffectDeferHelper(disableSticky bool) {
	value := "should not be modified"
	expectedValue := value
	doneCh := make(chan struct{})

	workflowFunc := func(ctx Context) error {
		defer func() {
			if !IsReplaying(ctx) {
				// This is an side effect op
				value = ""
			}
			close(doneCh)
		}()
		Sleep(ctx, 1*time.Second)
		return nil
	}
	workflowName := fmt.Sprintf("SideEffectDeferWorkflow-Sticky=%v", disableSticky)
	t.registry.RegisterWorkflowWithOptions(
		workflowFunc,
		RegisterWorkflowOptions{Name: workflowName},
	)

	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
	}

	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:               "test-id-1",
			Logger:                 zap.NewNop(),
			DisableStickyExecution: disableSticky,
		},
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	task := createWorkflowTask(testEvents, 0, workflowName)
	_, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.Nil(err)

	if !params.DisableStickyExecution {
		// 1. We can't set cache size in the test to 1, otherwise other tests will break.
		// 2. We need to make sure cache is empty when the test is completed,
		// So manually trigger a delete.
		getWorkflowCache().Delete(task.WorkflowExecution.GetRunId())
	}
	// Make sure the workflow coroutine has exited.
	<-doneCh
	// The side effect op should not be executed.
	t.Equal(expectedValue, value)

	// There should be nothing in the cache.
	t.EqualValues(0, getWorkflowCache().Size())
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_NondeterministicDetection() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("pkg.Greeter_Activity")},
			TaskList:     taskList,
		}),
	}
	task := createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	stopC := make(chan struct{})
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         zap.NewNop(),
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
		WorkerStopChannel: stopC,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	// there should be no error as the history events matched the decisions.
	t.NoError(err)
	t.NotNil(response)

	// now change the history event so it does not match to decision produced via replay
	testEvents[4].ActivityTaskScheduledEventAttributes.ActivityType.Name = common.StringPtr("some-other-activity")
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
	// will fail as it can't find laTunnel in getWorkflowCache().
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, stopC, nil)
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "nondeterministic")

	// now, create a new task handler with fail nondeterministic workflow policy
	// and verify that it handles the mismatching history correctly.
	params.NonDeterministicWorkflowPolicy = NonDeterministicWorkflowPolicyFailWorkflow
	failOnNondeterminismTaskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, err = failOnNondeterminismTaskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	// When FailWorkflow policy is set, task handler does not return an error,
	// because it will indicate non determinism in the request.
	t.NoError(err)
	// Verify that request is a RespondDecisionTaskCompleteRequest
	response, ok := request.(*s.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	// Verify there's at least 1 decision
	// and the last last decision is to fail workflow
	// and contains proper justification.(i.e. nondeterminism).
	t.True(len(response.Decisions) > 0)
	closeDecision := response.Decisions[len(response.Decisions)-1]
	t.Equal(*closeDecision.DecisionType, s.DecisionTypeFailWorkflowExecution)
	t.Contains(*closeDecision.FailWorkflowExecutionDecisionAttributes.Reason, "NonDeterministicWorkflowPolicyFailWorkflow")

	// now with different package name to activity type
	testEvents[4].ActivityTaskScheduledEventAttributes.ActivityType.Name = common.StringPtr("new-package.Greeter_Activity")
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_NondeterministicLogNonexistingID() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			// Insert an ID which does not exist
			ActivityId:   common.StringPtr("NotAnActivityID"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("pkg.Greeter_Activity")},
			TaskList:     taskList,
		}),
	}

	obs, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(obs)

	task := createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	stopC := make(chan struct{})
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   logger,
		},
		WorkerStopChannel: stopC,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

	t.Nil(request)
	t.ErrorContains(err, "nondeterministic workflow")

	// Check that the error was logged
	illegalPanicLogs := logs.FilterMessage("Illegal state caused panic")
	require.Len(t.T(), illegalPanicLogs.All(), 1)

	replayErrorField := findLogField(illegalPanicLogs.All()[0], "ReplayError")
	require.NotNil(t.T(), replayErrorField)
	require.Equal(t.T(), zapcore.ErrorType, replayErrorField.Type)
	require.ErrorContains(t.T(), replayErrorField.Interface.(error),
		"nondeterministic workflow: "+
			"history event is ActivityTaskScheduled: (ActivityId:NotAnActivityID, ActivityType:(Name:pkg.Greeter_Activity), TaskList:(Name:taskList, Kind:NORMAL), Input:[]), "+
			"replay decision is ScheduleActivityTask: (ActivityId:0, ActivityType:(Name:Greeter_Activity), TaskList:(Name:taskList, Kind:NORMAL)")
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowReturnsPanicError() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, "ReturnPanicWorkflow")
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         zap.NewNop(),
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*s.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	t.EqualValues(s.DecisionTypeFailWorkflowExecution, r.Decisions[0].GetDecisionType())
	attr := r.Decisions[0].FailWorkflowExecutionDecisionAttributes
	t.EqualValues("cadenceInternal:Panic", attr.GetReason())
	details := string(attr.Details)
	t.True(strings.HasPrefix(details, "\"panicError"), details)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowPanics() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
	}

	obs, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(obs)

	task := createWorkflowTask(testEvents, 3, "PanicWorkflow")
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         logger,
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*s.RespondDecisionTaskFailedRequest)
	t.True(ok)
	t.EqualValues("WORKFLOW_WORKER_UNHANDLED_FAILURE", r.Cause.String())
	t.EqualValues("panicError", string(r.Details))

	// Check that the error was logged
	panicLogs := logs.FilterMessage("Workflow panic.")
	require.Len(t.T(), panicLogs.All(), 1)

	wfTypeField := findLogField(panicLogs.All()[0], tagWorkflowType)
	require.NotNil(t.T(), wfTypeField)
	require.Equal(t.T(), "PanicWorkflow", wfTypeField.String)
}

func (t *TaskHandlersTestSuite) TestGetWorkflowInfo() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	parentID := "parentID"
	parentRunID := "parentRun"
	cronSchedule := "5 4 * * *"
	continuedRunID := uuid.New()
	parentExecution := &s.WorkflowExecution{
		WorkflowId: &parentID,
		RunId:      &parentRunID,
	}
	parentDomain := "parentDomain"
	var attempt int32 = 123
	var executionTimeout int32 = 213456
	var taskTimeout int32 = 21
	workflowType := "GetWorkflowInfoWorkflow"
	lastCompletionResult, err := getDefaultDataConverter().ToData("lastCompletionData")
	t.NoError(err)
	retryPolicy := &s.RetryPolicy{
		InitialIntervalInSeconds: common.Int32Ptr(1),
		BackoffCoefficient:       common.Float64Ptr(1.0),
		MaximumIntervalInSeconds: common.Int32Ptr(1),
		MaximumAttempts:          common.Int32Ptr(3),
	}
	startedEventAttributes := &s.WorkflowExecutionStartedEventAttributes{
		Input:                               lastCompletionResult,
		TaskList:                            taskList,
		ParentWorkflowExecution:             parentExecution,
		CronSchedule:                        &cronSchedule,
		ContinuedExecutionRunId:             &continuedRunID,
		ParentWorkflowDomain:                &parentDomain,
		Attempt:                             &attempt,
		ExecutionStartToCloseTimeoutSeconds: &executionTimeout,
		TaskStartToCloseTimeoutSeconds:      &taskTimeout,
		LastCompletionResult:                lastCompletionResult,
		RetryPolicy:                         retryPolicy,
	}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, startedEventAttributes),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, workflowType)
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         zap.NewNop(),
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*s.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	t.EqualValues(s.DecisionTypeCompleteWorkflowExecution, r.Decisions[0].GetDecisionType())
	attr := r.Decisions[0].CompleteWorkflowExecutionDecisionAttributes
	var result WorkflowInfo
	t.NoError(getDefaultDataConverter().FromData(attr.Result, &result))
	t.EqualValues(taskList.GetName(), result.TaskListName)
	t.EqualValues(parentID, result.ParentWorkflowExecution.ID)
	t.EqualValues(parentRunID, result.ParentWorkflowExecution.RunID)
	t.EqualValues(cronSchedule, *result.CronSchedule)
	t.EqualValues(continuedRunID, *result.ContinuedExecutionRunID)
	t.EqualValues(parentDomain, *result.ParentWorkflowDomain)
	t.EqualValues(attempt, result.Attempt)
	t.EqualValues(executionTimeout, result.ExecutionStartToCloseTimeoutSeconds)
	t.EqualValues(taskTimeout, result.TaskStartToCloseTimeoutSeconds)
	t.EqualValues(workflowType, result.WorkflowType.Name)
	t.EqualValues(testDomain, result.Domain)
	t.EqualValues(retryPolicy, result.RetryPolicy)
}

func (t *TaskHandlersTestSuite) TestConsistentQuery_InvalidQueryTask() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:                       "test-id-1",
			Logger:                         zap.NewNop(),
			NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		},
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	task := createWorkflowTask(nil, 3, "HelloWorld_Workflow")
	task.Query = &s.WorkflowQuery{}
	task.Queries = map[string]*s.WorkflowQuery{"query_id": {}}
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
	// query and queries are both specified so this is an invalid task
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "invalid query decision task")

	// There should be nothing in the cache.
	t.EqualValues(getWorkflowCache().Size(), 0)
}

func (t *TaskHandlersTestSuite) TestConsistentQuery_Success() {
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	checksum1 := "chck1"
	numberOfSignalsToComplete, err := getDefaultDataConverter().ToData(2)
	t.NoError(err)
	signal, err := getDefaultDataConverter().ToData("signal data")
	t.NoError(err)
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{
			TaskList: taskList,
			Input:    numberOfSignalsToComplete,
		}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{
			ScheduledEventId: common.Int64Ptr(2), BinaryChecksum: common.StringPtr(checksum1)}),
		createTestEventWorkflowExecutionSignaledWithPayload(5, signalCh, signal),
		createTestEventDecisionTaskScheduled(6, &s.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(7),
	}

	queries := map[string]*s.WorkflowQuery{
		"id1": {QueryType: common.StringPtr(queryType)},
		"id2": {QueryType: common.StringPtr(errQueryType)},
	}
	task := createWorkflowTaskWithQueries(testEvents[0:3], 0, "QuerySignalWorkflow", queries)

	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Len(response.Decisions, 0)
	expectedQueryResults := map[string]*s.WorkflowQueryResult{
		"id1": {
			ResultType: common.QueryResultTypePtr(s.QueryResultTypeAnswered),
			Answer:     []byte(fmt.Sprintf("\"%v\"\n", startingQueryValue)),
		},
		"id2": {
			ResultType:   common.QueryResultTypePtr(s.QueryResultTypeFailed),
			ErrorMessage: common.StringPtr(queryErr),
		},
	}
	t.assertQueryResultsEqual(expectedQueryResults, response.QueryResults)

	secondTask := createWorkflowTaskWithQueries(testEvents, 3, "QuerySignalWorkflow", queries)
	secondTask.WorkflowExecution.RunId = task.WorkflowExecution.RunId
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: secondTask}, nil)
	response = request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Len(response.Decisions, 1)
	expectedQueryResults = map[string]*s.WorkflowQueryResult{
		"id1": {
			ResultType: common.QueryResultTypePtr(s.QueryResultTypeAnswered),
			Answer:     []byte(fmt.Sprintf("\"%v\"\n", "signal data")),
		},
		"id2": {
			ResultType:   common.QueryResultTypePtr(s.QueryResultTypeFailed),
			ErrorMessage: common.StringPtr(queryErr),
		},
	}
	t.assertQueryResultsEqual(expectedQueryResults, response.QueryResults)

	// clean up workflow left in cache
	getWorkflowCache().Delete(*task.WorkflowExecution.RunId)
}

func (t *TaskHandlersTestSuite) assertQueryResultsEqual(expected map[string]*s.WorkflowQueryResult, actual map[string]*s.WorkflowQueryResult) {
	t.Equal(len(expected), len(actual))
	for expectedID, expectedResult := range expected {
		t.Contains(actual, expectedID)
		t.True(expectedResult.Equals(actual[expectedID]))
	}
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_CancelActivityBeforeSent() {
	// Schedule an activity and see if we complete workflow.
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_WorkflowCancel")

	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.Equal(s.DecisionTypeCompleteWorkflowExecution, response.Decisions[0].GetDecisionType())
	t.NotNil(response.Decisions[0].CompleteWorkflowExecutionDecisionAttributes)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_PageToken() {
	// Schedule a decision activity and see if we complete workflow.
	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{}),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_Workflow")
	task.NextPageToken = []byte("token")

	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
	}

	nextEvents := []*s.HistoryEvent{
		createTestEventDecisionTaskStarted(3),
	}

	historyIterator := &historyIteratorImpl{
		iteratorFunc: func(nextToken []byte) (*s.History, []byte, error) {
			return &s.History{nextEvents}, nil, nil
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task, historyIterator: historyIterator}, nil)
	response := request.(*s.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
}

func (t *TaskHandlersTestSuite) TestLocalActivityRetry_DecisionHeartbeatFail() {
	backoffIntervalInSeconds := int32(1)
	backoffDuration := time.Second * time.Duration(backoffIntervalInSeconds)
	workflowComplete := false

	retryLocalActivityWorkflowFunc := func(ctx Context, intput []byte) error {
		ao := LocalActivityOptions{
			ScheduleToCloseTimeout: time.Minute,
			RetryPolicy: &RetryPolicy{
				InitialInterval:    backoffDuration,
				BackoffCoefficient: 1.1,
				MaximumInterval:    time.Minute,
				ExpirationInterval: time.Minute,
			},
		}
		ctx = WithLocalActivityOptions(ctx, ao)

		err := ExecuteLocalActivity(ctx, func() error {
			return errors.New("some random error")
		}).Get(ctx, nil)
		workflowComplete = true
		return err
	}
	t.registry.RegisterWorkflowWithOptions(
		retryLocalActivityWorkflowFunc,
		RegisterWorkflowOptions{Name: "RetryLocalActivityWorkflow"},
	)

	decisionTaskStartedEvent := createTestEventDecisionTaskStarted(3)
	decisionTaskStartedEvent.Timestamp = common.Int64Ptr(time.Now().UnixNano())
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{
			// make sure the timeout is same as the backoff interval
			TaskStartToCloseTimeoutSeconds: common.Int32Ptr(backoffIntervalInSeconds),
			TaskList:                       &s.TaskList{Name: &testWorkflowTaskTasklist}},
		),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{}),
		decisionTaskStartedEvent,
	}

	task := createWorkflowTask(testEvents, 0, "RetryLocalActivityWorkflow")
	stopCh := make(chan struct{})
	params := workerExecutionParameters{
		TaskList: &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()},
		WorkerOptions: WorkerOptions{
			Identity: "test-id-1",
			Logger:   t.logger,
		},
		WorkerStopChannel: stopCh,
	}
	defer close(stopCh)

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	laTunnel := newLocalActivityTunnel(params.WorkerStopChannel)
	taskHandlerImpl, ok := taskHandler.(*workflowTaskHandlerImpl)
	t.True(ok)
	taskHandlerImpl.laTunnel = laTunnel

	laTaskPoller := newLocalActivityPoller(params, laTunnel)
	doneCh := make(chan struct{})
	go func() {
		// laTaskPoller needs to poll the local activity and process it
		task, err := laTaskPoller.PollTask()
		t.NoError(err)
		err = laTaskPoller.ProcessTask(task)
		t.NoError(err)

		// before clearing workflow state, a reset sticky task will be sent to this chan,
		// drain the chan so that workflow state can be cleared
		<-laTunnel.resultCh

		close(doneCh)
	}()

	laResultCh := make(chan *localActivityResult)
	response, err := taskHandler.ProcessWorkflowTask(
		&workflowTask{
			task:       task,
			laResultCh: laResultCh,
		},
		func(response interface{}, startTime time.Time) (*workflowTask, error) {
			return nil, &s.EntityNotExistsError{Message: "Decision task not found."}
		})
	t.Nil(response)
	t.Error(err)

	// wait for the retry timer to fire
	time.Sleep(backoffDuration)
	t.False(workflowComplete)
	<-doneCh
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NoError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)

	cancelRequested := false
	heartbeatResponse := s.RecordActivityTaskHeartbeatResponse{CancelRequested: &cancelRequested}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).Return(&heartbeatResponse, nil)

	cadenceInvoker := &cadenceInvoker{
		identity:  "Test_Cadence_Invoker",
		service:   mockService,
		taskToken: nil,
	}

	heartbeatErr := cadenceInvoker.BatchHeartbeat(nil)
	t.NoError(heartbeatErr)
}

func newHeartbeatRequestMatcher(details []byte) gomock.Matcher {
	return &recordHeartbeatRequestMatcher{details: details}
}

type recordHeartbeatRequestMatcher struct {
	details []byte
}

func (r *recordHeartbeatRequestMatcher) String() string {
	panic("implement me")
}

func (r *recordHeartbeatRequestMatcher) Matches(x interface{}) bool {
	req, ok := x.(*s.RecordActivityTaskHeartbeatRequest)
	if !ok {
		return false
	}

	return reflect.DeepEqual(r.details, req.Details)
}

func (t *TaskHandlersTestSuite) TestHeartBeat_Interleaved() {
	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)

	cancelRequested := false
	heartbeatResponse := s.RecordActivityTaskHeartbeatResponse{CancelRequested: &cancelRequested}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), newHeartbeatRequestMatcher([]byte("1")), callOptions()...).Return(&heartbeatResponse, nil).Times(3)
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), newHeartbeatRequestMatcher([]byte("2")), callOptions()...).Return(&heartbeatResponse, nil).Times(3)

	cadenceInvoker := &cadenceInvoker{
		identity:              "Test_Cadence_Invoker",
		service:               mockService,
		taskToken:             nil,
		heartBeatTimeoutInSec: 3,
	}

	heartbeatErr := cadenceInvoker.BatchHeartbeat([]byte("1"))
	t.NoError(heartbeatErr)
	time.Sleep(1000 * time.Millisecond)

	for i := 0; i < 4; i++ {
		heartbeatErr = cadenceInvoker.BackgroundHeartbeat()
		t.NoError(heartbeatErr)
		time.Sleep(800 * time.Millisecond)
	}

	time.Sleep(time.Second)

	heartbeatErr = cadenceInvoker.BatchHeartbeat([]byte("2"))
	t.NoError(heartbeatErr)
	time.Sleep(1000 * time.Millisecond)

	for i := 0; i < 4; i++ {
		heartbeatErr = cadenceInvoker.BackgroundHeartbeat()
		t.NoError(heartbeatErr)
		time.Sleep(800 * time.Millisecond)
	}
	time.Sleep(1 * time.Second)
}

func (t *TaskHandlersTestSuite) TestHeartBeatLogNil() {
	core, obs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	cadenceInv := &cadenceInvoker{
		identity: "Test_Cadence_Invoker",
		logger:   logger,
	}

	cadenceInv.logFailedHeartBeat(nil)

	t.Empty(obs.All())
}

func (t *TaskHandlersTestSuite) TestHeartBeatLogCanceledError() {
	core, obs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	cadenceInv := &cadenceInvoker{
		identity: "Test_Cadence_Invoker",
		logger:   logger,
	}

	var workflowCompleatedErr CanceledError
	cadenceInv.logFailedHeartBeat(&workflowCompleatedErr)

	t.Empty(obs.All())
}

func (t *TaskHandlersTestSuite) TestHeartBeatLogNotCanceled() {
	core, obs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	cadenceInv := &cadenceInvoker{
		identity: "Test_Cadence_Invoker",
		logger:   logger,
	}

	var workflowCompleatedErr s.WorkflowExecutionAlreadyCompletedError
	cadenceInv.logFailedHeartBeat(&workflowCompleatedErr)

	t.Len(obs.All(), 1)
	t.Equal("Failed to send heartbeat", obs.All()[0].Message)
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NilResponseWithError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)
	logger := testlogger.NewZap(t.T())

	entityNotExistsError := &s.EntityNotExistsError{}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, entityNotExistsError)

	cadenceInvoker := newServiceInvoker(
		nil,
		"Test_Cadence_Invoker",
		mockService,
		func() {},
		0,
		make(chan struct{}),
		FeatureFlags{},
		logger,
		testWorkflowType,
		testActivityType,
	)

	heartbeatErr := cadenceInvoker.BatchHeartbeat(nil)
	t.NotNil(heartbeatErr)
	_, ok := (heartbeatErr).(*s.EntityNotExistsError)
	t.True(ok, "heartbeatErr must be EntityNotExistsError.")
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NilResponseWithDomainNotActiveError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := workflowservicetest.NewMockClient(mockCtrl)
	logger := testlogger.NewZap(t.T())

	domainNotActiveError := &s.DomainNotActiveError{}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, domainNotActiveError)

	called := false
	cancelHandler := func() { called = true }

	cadenceInvoker := newServiceInvoker(
		nil,
		"Test_Cadence_Invoker",
		mockService,
		cancelHandler,
		0,
		make(chan struct{}),
		FeatureFlags{},
		logger,
		testWorkflowType,
		testActivityType,
	)

	heartbeatErr := cadenceInvoker.BatchHeartbeat(nil)
	t.NotNil(heartbeatErr)
	_, ok := (heartbeatErr).(*s.DomainNotActiveError)
	t.True(ok, "heartbeatErr must be DomainNotActiveError.")
	t.True(called)
}

type testActivityDeadline struct {
	logger *zap.Logger
	d      time.Duration
}

func (t *testActivityDeadline) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if d, _ := ctx.Deadline(); d.IsZero() {
		panic("invalid deadline provided")
	}
	if t.d != 0 {
		// Wait till deadline expires.
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return nil, nil
}

func (t *testActivityDeadline) ActivityType() ActivityType {
	return ActivityType{Name: "test"}
}

func (t *testActivityDeadline) GetFunction() interface{} {
	return t.Execute
}

func (t *testActivityDeadline) GetOptions() RegisterActivityOptions {
	return RegisterActivityOptions{}
}

// a regrettably-hacky func to use goleak to count leaking goroutines.
// ideally there will be a structured way to do this in the future, rather than string parsing
func countLeaks(leaks error) int {
	if leaks == nil {
		return 0
	}
	// leak messages look something like:
	// Goroutine 23 in state chan receive, with go.uber.org/cadence/internal.(*coroutineState).initialYield on top of the stack:
	// ... stacktrace ...
	//
	// Goroutine 28 ... on top of the stack:
	// ... stacktrace ...
	return strings.Count(leaks.Error(), "on top of the stack")
}

func (t *TaskHandlersTestSuite) TestRegression_QueriesDoNotLeakGoroutines() {
	// this test must not be run in parallel with most other tests, as it mutates global vars
	var ridsToCleanUp []string
	originalLeaks := goleak.Find()
	defer func(size int) {
		// empty the cache to clear out any newly-introduced leaks
		current := getWorkflowCache()
		for _, rid := range ridsToCleanUp {
			current.Delete(rid)
		}
		// check the cleanup
		currentLeaks := goleak.Find()
		if countLeaks(currentLeaks) != countLeaks(originalLeaks) {
			t.T().Errorf("failed to clean up goroutines.\nOriginal state:\n%v\n\nCurrent state:\n%v", originalLeaks, currentLeaks)
		}

		// reset everything to make it "normal".
		// this does NOT restore the original workflow cache - that cannot be done correctly, initCacheOnce is not safe to copy (thus restore).
		stickyCacheSize = size
		workflowCache = nil
		initCacheOnce = sync.Once{}
	}(stickyCacheSize)
	workflowCache = nil
	initCacheOnce = sync.Once{}
	// cache is intentionally not *disabled*, as that would go down no-cache code paths.
	// also, there is an LRU-cache bug where the size allows N to enter, but then removes until N-1 remain,
	// so a size of 2 actually means a size of 1.
	SetStickyWorkflowCacheSize(2)

	taskList := &s.TaskList{Name: common.StringPtr("taskList"), Kind: s.TaskListKindNormal.Ptr()}
	params := workerExecutionParameters{
		TaskList: taskList,
		WorkerOptions: WorkerOptions{
			Identity:               "test-id-1",
			Logger:                 t.logger,
			DisableStickyExecution: false,
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)

	// process a throw-away workflow to fill the cache.  this is copied from TestWorkflowTask_QueryWorkflow since it's
	// relatively simple, but any should work fine, as long as it can be queried.
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{TaskList: taskList}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{ScheduledEventId: common.Int64Ptr(2)}),
		createTestEventActivityTaskScheduled(5, &s.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &s.ActivityType{Name: common.StringPtr("Greeter_Activity")},
			TaskList:     taskList,
		}),
	}
	cachedTask := createWorkflowTask(testEvents[0:1], 1, "HelloWorld_Workflow")
	cachedTask.WorkflowExecution.WorkflowId = common.StringPtr("cache-filling workflow id")
	ridsToCleanUp = append(ridsToCleanUp, *cachedTask.WorkflowExecution.RunId)
	_, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: cachedTask}, nil)

	// sanity check that the cache was indeed filled, and that it has created a goroutine
	require.NoError(t.T(), err, "cache-filling must succeed")
	require.Equal(t.T(), 1, getWorkflowCache().Size(), "workflow should be cached, but was not")
	oneCachedLeak := goleak.Find()
	require.Error(t.T(), oneCachedLeak, "expected at least one leaking goroutine")
	require.Equal(t.T(), countLeaks(originalLeaks)+1, countLeaks(oneCachedLeak), // ideally == 1, but currently there are other leaks
		"expected the cached workflow to leak one goroutine.  original leaks:\n%v\n\nleaks after one workflow:\n%v", originalLeaks, oneCachedLeak)

	// now query a different workflow ID / run ID
	uncachedTask := createQueryTask(testEvents, 5, "HelloWorld_Workflow", queryType)
	uncachedTask.WorkflowExecution.WorkflowId = common.StringPtr("should not leak this workflow id")
	ridsToCleanUp = append(ridsToCleanUp, *uncachedTask.WorkflowExecution.RunId) // only necessary if the test fails
	result, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: uncachedTask}, nil)
	require.NoError(t.T(), err)
	t.verifyQueryResult(result, "waiting-activity-result") // largely a sanity check

	// and finally the purpose of this test:
	// verify that the cache has not been modified, and that there is no new leak
	t.Equal(1, getWorkflowCache().Size(), "workflow cache should be the same size")
	t.True(getWorkflowCache().Exist(cachedTask.WorkflowExecution.GetRunId()), "originally-cached workflow should still be cached")
	t.False(getWorkflowCache().Exist(uncachedTask.WorkflowExecution.GetRunId()), "queried workflow should not be cached")
	newLeaks := goleak.Find()
	t.Error(newLeaks, "expected at least one leaking goroutine")
	t.Equal(countLeaks(oneCachedLeak), countLeaks(newLeaks),
		"expected the query to leak no new goroutines.  before query:\n%v\n\nafter query:\n%v", oneCachedLeak, newLeaks)
}

func Test_NonDeterministicCheck(t *testing.T) {
	decisionTypes := s.DecisionType_Values()
	require.Equal(t, 13, len(decisionTypes), "If you see this error, you are adding new decision type. "+
		"Before updating the number to make this test pass, please make sure you update isDecisionMatchEvent() method "+
		"to check the new decision type. Otherwise the replay will fail on the new decision event.")

	eventTypes := s.EventType_Values()
	decisionEventTypeCount := 0
	for _, et := range eventTypes {
		if isDecisionEvent(et) {
			decisionEventTypeCount++
		}
	}
	// CancelTimer has 2 corresponding events.
	require.Equal(t, len(decisionTypes)+1, decisionEventTypeCount, "Every decision type must have one matching event type. "+
		"If you add new decision type, you need to update isDecisionEvent() method to include that new event type as well.")
}

func Test_IsDecisionMatchEvent_UpsertWorkflowSearchAttributes(t *testing.T) {
	diType := s.DecisionTypeUpsertWorkflowSearchAttributes
	eType := s.EventTypeUpsertWorkflowSearchAttributes
	strictMode := false

	testCases := []struct {
		name     string
		decision *s.Decision
		event    *s.HistoryEvent
		expected bool
	}{
		{
			name: "event type not match",
			decision: &s.Decision{
				DecisionType: &diType,
				UpsertWorkflowSearchAttributesDecisionAttributes: &s.UpsertWorkflowSearchAttributesDecisionAttributes{
					SearchAttributes: &s.SearchAttributes{},
				},
			},
			event:    &s.HistoryEvent{},
			expected: false,
		},
		{
			name: "attributes not match",
			decision: &s.Decision{
				DecisionType: &diType,
				UpsertWorkflowSearchAttributesDecisionAttributes: &s.UpsertWorkflowSearchAttributesDecisionAttributes{
					SearchAttributes: &s.SearchAttributes{},
				},
			},
			event: &s.HistoryEvent{
				EventType: &eType,
				UpsertWorkflowSearchAttributesEventAttributes: &s.UpsertWorkflowSearchAttributesEventAttributes{},
			},
			expected: true,
		},
		{
			name: "attributes match",
			decision: &s.Decision{
				DecisionType: &diType,
				UpsertWorkflowSearchAttributesDecisionAttributes: &s.UpsertWorkflowSearchAttributesDecisionAttributes{
					SearchAttributes: &s.SearchAttributes{},
				},
			},
			event: &s.HistoryEvent{
				EventType: &eType,
				UpsertWorkflowSearchAttributesEventAttributes: &s.UpsertWorkflowSearchAttributesEventAttributes{
					SearchAttributes: &s.SearchAttributes{},
				},
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, isDecisionMatchEvent(testCase.decision, testCase.event, strictMode))
		})
	}

	strictMode = true

	testCases = []struct {
		name     string
		decision *s.Decision
		event    *s.HistoryEvent
		expected bool
	}{
		{
			name: "attributes not match",
			decision: &s.Decision{
				DecisionType: &diType,
				UpsertWorkflowSearchAttributesDecisionAttributes: &s.UpsertWorkflowSearchAttributesDecisionAttributes{
					SearchAttributes: &s.SearchAttributes{},
				},
			},
			event: &s.HistoryEvent{
				EventType: &eType,
				UpsertWorkflowSearchAttributesEventAttributes: &s.UpsertWorkflowSearchAttributesEventAttributes{},
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, isDecisionMatchEvent(testCase.decision, testCase.event, strictMode))
		})
	}
}

func Test_IsSearchAttributesMatched(t *testing.T) {
	testCases := []struct {
		name     string
		lhs      *s.SearchAttributes
		rhs      *s.SearchAttributes
		expected bool
	}{
		{
			name:     "both nil",
			lhs:      nil,
			rhs:      nil,
			expected: true,
		},
		{
			name:     "left nil",
			lhs:      nil,
			rhs:      &s.SearchAttributes{},
			expected: false,
		},
		{
			name:     "right nil",
			lhs:      &s.SearchAttributes{},
			rhs:      nil,
			expected: false,
		},
		{
			name: "not match",
			lhs: &s.SearchAttributes{
				IndexedFields: map[string][]byte{
					"key1": []byte("1"),
					"key2": []byte("abc"),
				},
			},
			rhs:      &s.SearchAttributes{},
			expected: false,
		},
		{
			name: "match",
			lhs: &s.SearchAttributes{
				IndexedFields: map[string][]byte{
					"key1": []byte("1"),
					"key2": []byte("abc"),
				},
			},
			rhs: &s.SearchAttributes{
				IndexedFields: map[string][]byte{
					"key2": []byte("abc"),
					"key1": []byte("1"),
				},
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, isSearchAttributesMatched(testCase.lhs, testCase.rhs))
		})
	}
}

func Test__GetWorkflowStartedEvent(t *testing.T) {
	wfStartedEvent := createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{TaskList: &s.TaskList{Name: common.StringPtr("tl1")}})
	h := &history{workflowTask: &workflowTask{task: &s.PollForDecisionTaskResponse{History: &s.History{Events: []*s.HistoryEvent{wfStartedEvent}}}}}
	result, err := h.GetWorkflowStartedEvent()
	require.NoError(t, err)
	require.Equal(t, wfStartedEvent, result)

	emptyHistory := &history{workflowTask: &workflowTask{task: &s.PollForDecisionTaskResponse{History: &s.History{}}}}
	result, err = emptyHistory.GetWorkflowStartedEvent()
	require.ErrorContains(t, err, "unable to find WorkflowExecutionStartedEventAttributes")
	require.Nil(t, result)
}

func Test__verifyAllEventsProcessed(t *testing.T) {
	testCases := []struct {
		name        string
		lastEventID int64
		nextEventID int64
		Message     string
	}{
		{
			name:        "error",
			lastEventID: 1,
			nextEventID: 1,
			Message:     "history_events: premature end of stream",
		},
		{
			name:        "warn",
			lastEventID: 1,
			nextEventID: 3,
			Message:     "history_events: processed events past the expected lastEventID",
		},
		{
			name:        "success",
			lastEventID: 1,
			nextEventID: 2,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			obs, logs := observer.New(zap.WarnLevel)
			logger := zap.New(obs)
			h := &history{
				lastEventID:   testCase.lastEventID,
				nextEventID:   testCase.nextEventID,
				eventsHandler: &workflowExecutionEventHandlerImpl{workflowEnvironmentImpl: &workflowEnvironmentImpl{logger: logger}}}
			err := h.verifyAllEventsProcessed()
			if testCase.name == "error" {
				require.ErrorContains(t, err, testCase.Message)
			} else if testCase.name == "warn" {
				warnLogs := logs.FilterMessage(testCase.Message)
				require.Len(t, warnLogs.All(), 1)
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func Test__workflowCategorizedByTimeout(t *testing.T) {
	testCases := []struct {
		timeout          int32
		expectedCategory string
	}{
		{
			timeout:          1,
			expectedCategory: "instant",
		},
		{
			timeout:          1000,
			expectedCategory: "short",
		},
		{
			timeout:          2000,
			expectedCategory: "intermediate",
		},
		{
			timeout:          30000,
			expectedCategory: "long",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.expectedCategory, func(t *testing.T) {
			wfContext := &workflowExecutionContextImpl{workflowInfo: &WorkflowInfo{ExecutionStartToCloseTimeoutSeconds: tt.timeout}}
			require.Equal(t, tt.expectedCategory, workflowCategorizedByTimeout(wfContext))
		})
	}
}

func Test__SignalWorkflow(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(mockCtrl)
	mockService.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(nil)
	cadenceInvoker := &cadenceInvoker{
		identity:  "Test_Cadence_Invoker",
		service:   mockService,
		taskToken: nil,
	}
	err := cadenceInvoker.SignalWorkflow(context.Background(), "test-domain", "test-workflow-id", "test-run-id", "test-signal-name", nil)
	require.NoError(t, err)
}

func Test__getRetryBackoffWithNowTime(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name            string
		maxAttempts     int32
		ExpInterval     time.Duration
		result          time.Duration
		attempt         int32
		errReason       string
		expireTime      time.Time
		initialInterval time.Duration
		maxInterval     time.Duration
	}{
		{
			name:        "no max attempts or expiration interval set",
			maxAttempts: 0,
			ExpInterval: 0,
			result:      noRetryBackoff,
		},
		{
			name:        "max attempts done",
			maxAttempts: 5,
			attempt:     5,
			result:      noRetryBackoff,
		},
		{
			name:            "non retryable error",
			maxAttempts:     5,
			attempt:         2,
			errReason:       "bad request",
			initialInterval: time.Minute,
			maxInterval:     time.Minute,
			result:          noRetryBackoff,
		},
		{
			name:            "fallback to max interval when calculated backoff is 0",
			maxAttempts:     5,
			attempt:         2,
			initialInterval: 0,
			maxInterval:     time.Minute,
			result:          time.Minute,
		},
		{
			name:            "fallback to no retry backoff when calculated backoff is 0 and max interval is not set",
			maxAttempts:     5,
			attempt:         2,
			initialInterval: 0,
			result:          noRetryBackoff,
		},
		{
			name:            "expiry time reached",
			maxAttempts:     5,
			attempt:         2,
			expireTime:      now.Add(time.Second),
			initialInterval: time.Minute,
			maxInterval:     time.Minute,
			result:          noRetryBackoff,
		},
		{
			name:            "retry after backoff",
			maxAttempts:     5,
			attempt:         2,
			errReason:       "timeout",
			initialInterval: time.Minute,
			maxInterval:     time.Minute,
			result:          time.Minute,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			policy := &RetryPolicy{
				MaximumAttempts:          tt.maxAttempts,
				ExpirationInterval:       tt.ExpInterval,
				BackoffCoefficient:       2,
				NonRetriableErrorReasons: []string{"bad request"},
				MaximumInterval:          tt.maxInterval,
				InitialInterval:          tt.initialInterval,
			}
			require.Equal(t, tt.result, getRetryBackoffWithNowTime(policy, tt.attempt, tt.errReason, now, tt.expireTime))
		})
	}
}
