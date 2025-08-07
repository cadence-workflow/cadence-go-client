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
	"testing"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shadower"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

type shadowWorkerSuite struct {
	*require.Assertions
	suite.Suite

	controller  *gomock.Controller
	mockService *workflowservicetest.MockClient
}

func TestShadowWorkerSuite(t *testing.T) {
	s := new(shadowWorkerSuite)
	suite.Run(t, s)
}

func (s *shadowWorkerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockService = workflowservicetest.NewMockClient(s.controller)
}

func (s *shadowWorkerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *shadowWorkerSuite) TestNewShadowWorker() {
	registry := newRegistry()
	shadowWorker := newShadowWorker(
		s.mockService,
		testDomain,
		ShadowOptions{},
		workerExecutionParameters{
			TaskList: testTaskList,
			WorkerOptions: WorkerOptions{
				Logger: testlogger.NewZap(s.T())},
		},
		registry,
	)

	// check if scan and replay activities are registered
	_, ok := registry.GetActivity(shadower.ScanWorkflowActivityName)
	s.True(ok)
	_, ok = registry.GetActivity(shadower.ReplayWorkflowActivityName)
	s.True(ok)

	// check if background context is updated with necessary components
	userContext := shadowWorker.activityWorker.executionParameters.UserContext
	_, ok = userContext.Value(serviceClientContextKey).(workflowserviceclient.Interface)
	s.True(ok)
	_, ok = userContext.Value(workflowReplayerContextKey).(*WorkflowReplayer)
	s.True(ok)

	taskList := shadowWorker.activityWorker.executionParameters.TaskList
	s.Contains(taskList.GetName(), testDomain)
}

func (s *shadowWorkerSuite) TestStartShadowWorker_Failed_InvalidShadowOption() {
	shadowWorker := newShadowWorker(
		s.mockService,
		testDomain,
		ShadowOptions{
			Mode: ShadowModeContinuous, // exit condition is not specified
		},
		workerExecutionParameters{
			TaskList: testTaskList,
			WorkerOptions: WorkerOptions{
				Logger: testlogger.NewZap(s.T())},
		},
		newRegistry(),
	)

	s.Error(shadowWorker.Start())
}

func (s *shadowWorkerSuite) TestStartShadowWorker_Failed_DomainNotExist() {
	s.mockService.EXPECT().DescribeDomain(gomock.Any(), &shared.DescribeDomainRequest{
		Name: common.StringPtr(testDomain),
	}, callOptions()...).Return(nil, &shared.EntityNotExistsError{}).Times(1)

	shadowWorker := newShadowWorker(
		s.mockService,
		testDomain,
		ShadowOptions{},
		workerExecutionParameters{
			TaskList: testTaskList,
			WorkerOptions: WorkerOptions{
				Logger: testlogger.NewZap(s.T())},
		},
		newRegistry(),
	)

	s.Error(shadowWorker.Start())
}

func (s *shadowWorkerSuite) TestStartShadowWorker_Failed_TaskListNotSpecified() {
	s.mockService.EXPECT().DescribeDomain(gomock.Any(), &shared.DescribeDomainRequest{
		Name: common.StringPtr(testDomain),
	}, callOptions()...).Return(&shared.DescribeDomainResponse{}, nil).Times(1)

	shadowWorker := newShadowWorker(
		s.mockService,
		testDomain,
		ShadowOptions{},
		workerExecutionParameters{
			WorkerOptions: WorkerOptions{
				Logger: testlogger.NewZap(s.T())},
		},
		newRegistry(),
	)

	s.Equal(errTaskListNotSet, shadowWorker.Start())
}

func (s *shadowWorkerSuite) TestStartShadowWorker_Failed_StartWorkflowError() {
	s.mockService.EXPECT().DescribeDomain(gomock.Any(), &shared.DescribeDomainRequest{
		Name: common.StringPtr(testDomain),
	}, callOptions()...).Return(&shared.DescribeDomainResponse{}, nil).Times(1)
	// first return a retryable error to check if retry policy is configured
	s.mockService.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, &shared.InternalServiceError{}).Times(1)
	// then return a non-retryable error
	s.mockService.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, &shared.BadRequestError{}).Times(1)

	shadowWorker := newShadowWorker(
		s.mockService,
		testDomain,
		ShadowOptions{},
		workerExecutionParameters{
			TaskList: testTaskList,
			WorkerOptions: WorkerOptions{
				Logger: testlogger.NewZap(s.T())},
		},
		newRegistry(),
	)

	s.Error(shadowWorker.Start())
}

func (s *shadowWorkerSuite) TestStartShadowWorker_Succeed() {
	workflowQuery := "workflow query string"
	samplingRate := 0.5
	concurrency := 10
	shadowMode := ShadowModeContinuous
	exitCondition := ShadowExitCondition{
		ShadowCount: 100,
	}

	var startRequest *shared.StartWorkflowExecutionRequest
	s.mockService.EXPECT().DescribeDomain(gomock.Any(), &shared.DescribeDomainRequest{
		Name: common.StringPtr(testDomain),
	}, callOptions()...).Return(&shared.DescribeDomainResponse{}, nil).Times(1)
	s.mockService.EXPECT().DescribeDomain(gomock.Any(), &shared.DescribeDomainRequest{
		Name: common.StringPtr(shadower.LocalDomainName),
	}, callOptions()...).Return(&shared.DescribeDomainResponse{}, nil).Times(1)
	s.mockService.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(
		func(_ context.Context, request *shared.StartWorkflowExecutionRequest, _ ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
			startRequest = request
			return nil, &shared.WorkflowExecutionAlreadyStartedError{}
		},
	).Times(1)

	shadowWorker := newShadowWorker(
		s.mockService,
		testDomain,
		ShadowOptions{
			WorkflowQuery: workflowQuery,
			SamplingRate:  samplingRate,
			Mode:          shadowMode,
			ExitCondition: exitCondition,
			Concurrency:   concurrency,
		},
		workerExecutionParameters{
			TaskList: testTaskList,
			WorkerOptions: WorkerOptions{
				Logger: testlogger.NewZap(s.T())},
		},
		newRegistry(),
	)

	s.NoError(shadowWorker.Start())
	shadowWorker.Stop()

	s.Equal(shadower.LocalDomainName, startRequest.GetDomain())
	s.Equal(testDomain+shadower.WorkflowIDSuffix, startRequest.GetWorkflowId())
	s.Equal(shadower.WorkflowName, startRequest.WorkflowType.GetName())
	s.Equal(shadower.TaskList, startRequest.TaskList.GetName())
	s.NotZero(startRequest.GetExecutionStartToCloseTimeoutSeconds())
	s.Equal(shared.WorkflowIdReusePolicyAllowDuplicate, startRequest.GetWorkflowIdReusePolicy())

	var workflowParams shadower.WorkflowParams
	getDefaultDataConverter().FromData(startRequest.Input, &workflowParams)
	s.Equal(testDomain, workflowParams.GetDomain())
	s.Equal(generateShadowTaskList(testDomain, testTaskList.GetName()), workflowParams.GetTaskList())
	s.Equal(workflowQuery, workflowParams.GetWorkflowQuery())
	s.Equal(samplingRate, workflowParams.GetSamplingRate())
	s.Equal(shadowMode.toThriftPtr(), workflowParams.ShadowMode)
	s.Equal(exitCondition.toThriftPtr(), workflowParams.ExitCondition)
	s.Equal(int32(concurrency), workflowParams.GetConcurrency())
}
