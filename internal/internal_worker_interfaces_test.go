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
	"errors"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	m "go.uber.org/cadence/.gen/go/shared"
)

const (
	queryType    = "test-query"
	errQueryType = "test-err-query"
	signalCh     = "signal-chan"

	startingQueryValue = ""
	finishedQueryValue = "done"
	queryErr           = "error handling query"
)

type (
	// Greeter activity
	greeterActivity struct {
	}

	InterfacesTestSuite struct {
		suite.Suite
		mockCtrl *gomock.Controller
		service  *workflowservicetest.MockClient
	}
)

func helloWorldWorkflowFunc(ctx Context, input []byte) error {
	queryResult := startingQueryValue
	SetQueryHandler(ctx, queryType, func() (string, error) {
		return queryResult, nil
	})

	activityName := "Greeter_Activity"
	ao := ActivityOptions{
		TaskList:               "taskList",
		ActivityID:             "0",
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       20 * time.Second,
	}
	ctx = WithActivityOptions(ctx, ao)
	var result []byte
	queryResult = "waiting-activity-result"
	err := ExecuteActivity(ctx, activityName).Get(ctx, &result)
	if err == nil {
		queryResult = finishedQueryValue
		return nil
	}

	queryResult = "error:" + err.Error()
	return err
}

func querySignalWorkflowFunc(ctx Context, numSignals int) error {
	queryResult := startingQueryValue
	SetQueryHandler(ctx, queryType, func() (string, error) {
		return queryResult, nil
	})

	SetQueryHandler(ctx, errQueryType, func() (string, error) {
		return "", errors.New(queryErr)
	})

	ch := GetSignalChannel(ctx, signalCh)
	for i := 0; i < numSignals; i++ {
		// update queryResult when signal is received
		ch.Receive(ctx, &queryResult)

		// schedule activity to verify decisions are produced
		ao := ActivityOptions{
			TaskList:               "taskList",
			ActivityID:             "0",
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       20 * time.Second,
		}
		ExecuteActivity(WithActivityOptions(ctx, ao), "Greeter_Activity")
	}
	return nil
}

func binaryChecksumWorkflowFunc(ctx Context) ([]string, error) {
	var result []string
	result = append(result, GetWorkflowInfo(ctx).GetBinaryChecksum())
	Sleep(ctx, time.Hour)
	result = append(result, GetWorkflowInfo(ctx).GetBinaryChecksum())
	Sleep(ctx, time.Hour)
	result = append(result, GetWorkflowInfo(ctx).GetBinaryChecksum())
	return result, nil
}

func helloWorldWorkflowCancelFunc(ctx Context, input []byte) error {
	activityName := "Greeter_Activity"
	ao := ActivityOptions{
		TaskList:               "taskList",
		ActivityID:             "0",
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       20 * time.Second,
	}
	ctx = WithActivityOptions(ctx, ao)
	ExecuteActivity(ctx, activityName)
	getWorkflowEnvironment(ctx).RequestCancelActivity("0")
	return nil
}

// Greeter activity methods
func (ga greeterActivity) ActivityType() ActivityType {
	activityName := "Greeter_Activity"
	return ActivityType{Name: activityName}
}

func (ga greeterActivity) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return []byte("World"), nil
}

func (ga greeterActivity) GetFunction() interface{} {
	return ga.Execute
}

func (ga greeterActivity) GetOptions() RegisterActivityOptions {
	return RegisterActivityOptions{}
}

// Greeter activity func
func greeterActivityFunc(ctx context.Context, input []byte) ([]byte, error) {
	return []byte("Hello world"), nil
}

// Test suite.
func TestInterfacesTestSuite(t *testing.T) {
	suite.Run(t, new(InterfacesTestSuite))
}

func (s *InterfacesTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = workflowservicetest.NewMockClient(s.mockCtrl)
}

func (s *InterfacesTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *InterfacesTestSuite) TestInterface() {
	domain := "testDomain"
	// Workflow execution parameters.
	workflowExecutionParameters := workerExecutionParameters{
		TaskList: &m.TaskList{Name: common.StringPtr("testTaskList"), Kind: m.TaskListKindNormal.Ptr()},
		WorkerOptions: WorkerOptions{
			MaxConcurrentActivityTaskPollers: 4,
			MaxConcurrentDecisionTaskPollers: 4,
			Logger:                           testlogger.NewZap(s.T()),
		},
	}

	domainStatus := m.DomainStatusRegistered
	domainDesc := &m.DescribeDomainResponse{
		DomainInfo: &m.DomainInfo{
			Name:   &domain,
			Status: &domainStatus,
		},
	}

	// mocks
	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions()...).Return(domainDesc, nil).AnyTimes()
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptionsWithIsolationGroupHeader()...).Return(&m.PollForActivityTaskResponse{}, nil).AnyTimes()
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptionsWithIsolationGroupHeader()...).Return(nil).AnyTimes()
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptionsWithIsolationGroupHeader()...).Return(&m.PollForDecisionTaskResponse{}, nil).AnyTimes()
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptionsWithIsolationGroupHeader()...).Return(nil, nil).AnyTimes()
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptionsWithIsolationGroupHeader()...).Return(&m.StartWorkflowExecutionResponse{}, nil).AnyTimes()

	registry := newRegistry()
	// Launch worker.
	workflowWorker := newWorkflowWorker(s.service, domain, workflowExecutionParameters, nil, registry, nil)
	defer workflowWorker.Stop()
	workflowWorker.Start()

	// Create activity execution parameters.
	activityExecutionParameters := workerExecutionParameters{
		TaskList: &m.TaskList{Name: common.StringPtr("testTaskList"), Kind: m.TaskListKindNormal.Ptr()},
		WorkerOptions: WorkerOptions{
			MaxConcurrentActivityTaskPollers: 10,
			MaxConcurrentDecisionTaskPollers: 10,
			Logger:                           testlogger.NewZap(s.T()),
		},
	}

	// Register activity instances and launch the worker.
	activityWorker := newActivityWorker(s.service, domain, activityExecutionParameters, nil, registry, nil)
	defer activityWorker.Stop()
	activityWorker.Start()

	// Start a workflow.
	workflowOptions := StartWorkflowOptions{
		ID:                              "HelloWorld_Workflow",
		TaskList:                        "testTaskList",
		ExecutionStartToCloseTimeout:    10 * time.Second,
		DecisionTaskStartToCloseTimeout: 10 * time.Second,
	}
	workflowClient := NewClient(s.service, domain, &ClientOptions{IsolationGroup: "zone-2"})
	_, err := workflowClient.StartWorkflow(context.Background(), workflowOptions, "workflowType")
	s.NoError(err)
}
