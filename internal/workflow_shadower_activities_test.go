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
	"math"
	"math/rand"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shadower"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

type workflowShadowerActivitiesSuite struct {
	*require.Assertions
	suite.Suite
	WorkflowTestSuite

	controller  *gomock.Controller
	mockService *workflowservicetest.MockClient

	env                 *TestActivityEnvironment
	testReplayer        *WorkflowReplayer
	testWorkflowHistory *shared.History
}

func TestWorkflowShadowerActivitiesSuite(t *testing.T) {
	s := new(workflowShadowerActivitiesSuite)
	suite.Run(t, s)
}

func (s *workflowShadowerActivitiesSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockService = workflowservicetest.NewMockClient(s.controller)
	s.SetLogger(testlogger.NewZap(s.T()))

	s.env = s.NewTestActivityEnvironment()
	s.testReplayer = NewWorkflowReplayer()
	s.testReplayer.RegisterWorkflow(testReplayWorkflow)
	s.testWorkflowHistory = getTestReplayWorkflowFullHistory(s.T())

	activityContext := context.Background()
	activityContext = context.WithValue(activityContext, serviceClientContextKey, s.mockService)
	activityContext = context.WithValue(activityContext, workflowReplayerContextKey, s.testReplayer)
	s.env.SetWorkerOptions(WorkerOptions{
		BackgroundActivityContext: activityContext,
		Logger:                    testlogger.NewZap(s.T()),
	})
	s.env.RegisterActivityWithOptions(scanWorkflowActivity, RegisterActivityOptions{
		Name: shadower.ScanWorkflowActivityName,
	})
	s.env.RegisterActivityWithOptions(replayWorkflowActivity, RegisterActivityOptions{
		Name: shadower.ReplayWorkflowActivityName,
	})
}

func (s *workflowShadowerActivitiesSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *workflowShadowerActivitiesSuite) TestScanWorkflowActivity_Succeed() {
	numExecutions := 1000
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(numExecutions),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)

	params := shadower.ScanWorkflowActivityParams{
		Domain:        common.StringPtr(defaultTestDomain),
		WorkflowQuery: common.StringPtr("some random workflow visibility query"),
		SamplingRate:  common.Float64Ptr(0.5),
	}

	resultValue, err := s.env.ExecuteActivity(shadower.ScanWorkflowActivityName, params)
	s.NoError(err)

	var result shadower.ScanWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.NotNil(result.NextPageToken)
	s.True(len(result.Executions) > 0)
	s.True(len(result.Executions) < numExecutions)
}

func (s *workflowShadowerActivitiesSuite) TestScanWorkflowActivity_MinResultSize() {
	numExecutionsPerScan := 3
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(numExecutionsPerScan),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(int(math.Ceil(float64(minScanWorkflowResultSize) / float64(numExecutionsPerScan))))

	params := shadower.ScanWorkflowActivityParams{
		Domain:        common.StringPtr(defaultTestDomain),
		WorkflowQuery: common.StringPtr("some random workflow visibility query"),
		SamplingRate:  common.Float64Ptr(1),
	}

	resultValue, err := s.env.ExecuteActivity(shadower.ScanWorkflowActivityName, params)
	s.NoError(err)

	var result shadower.ScanWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.NotNil(result.NextPageToken)
	s.True(len(result.Executions) >= minScanWorkflowResultSize)
}

func (s *workflowShadowerActivitiesSuite) TestScanWorkflowActivity_CompletionTime() {
	activityTimeoutSeconds := int32(1)
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(1),
		NextPageToken: []byte{1, 2, 3},
	}, nil).MaxTimes(int(time.Duration(activityTimeoutSeconds) * time.Second / scanWorkflowWaitPeriod))

	params := shadower.ScanWorkflowActivityParams{
		Domain:        common.StringPtr(defaultTestDomain),
		WorkflowQuery: common.StringPtr("some random workflow visibility query"),
		SamplingRate:  common.Float64Ptr(0.00000001),
	}

	resultValue, err := s.env.impl.executeActivityWithOptions(
		activityOptions{
			ScheduleToCloseTimeoutSeconds: activityTimeoutSeconds,
			StartToCloseTimeoutSeconds:    activityTimeoutSeconds,
		},
		shadower.ScanWorkflowActivityName,
		params,
	)
	s.NoError(err)

	var result shadower.ScanWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.NotNil(result.NextPageToken)
	s.Empty(result.Executions)
}

func (s *workflowShadowerActivitiesSuite) TestScanWorkflowActivity_InvalidQuery() {
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, &shared.BadRequestError{
		Message: "invalid query",
	}).Times(1)

	params := shadower.ScanWorkflowActivityParams{
		Domain:        common.StringPtr(defaultTestDomain),
		WorkflowQuery: common.StringPtr("invalid workflow visibility query"),
	}

	_, err := s.env.ExecuteActivity(shadower.ScanWorkflowActivityName, params)
	s.Error(err)
	s.Equal(shadower.ErrReasonInvalidQuery, err.Error())
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_NoPreviousProgress() {
	numExecutions := 10
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(numExecutions)

	params := newTestReplayWorkflowActivityParams(numExecutions)

	resultValue, err := s.env.ExecuteActivity(shadower.ReplayWorkflowActivityName, params)
	s.NoError(err)

	var result shadower.ReplayWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.Equal(int32(numExecutions), result.GetSucceeded())
	s.Zero(result.GetSkipped())
	s.Zero(result.GetFailed())
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_WithPreviousProgress() {
	progress := replayWorkflowActivityProgress{
		Result: shadower.ReplayWorkflowActivityResult{
			Succeeded: common.Int32Ptr(3),
			Skipped:   common.Int32Ptr(2),
			Failed:    common.Int32Ptr(0),
		},
		NextExecutionIdx: 5,
	}
	s.env.SetHeartbeatDetails(progress)

	numExecutions := 10
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(numExecutions - progress.NextExecutionIdx)

	params := newTestReplayWorkflowActivityParams(numExecutions)

	resultValue, err := s.env.ExecuteActivity(shadower.ReplayWorkflowActivityName, params)
	s.NoError(err)

	var result shadower.ReplayWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.Equal(progress.Result.GetSucceeded()+int32(numExecutions-progress.NextExecutionIdx), result.GetSucceeded())
	s.Equal(progress.Result.GetSkipped(), result.GetSkipped())
	s.Equal(progress.Result.GetFailed(), result.GetFailed())
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_RandomReplayResult() {
	numSucceed := int32(0)
	numFailed := int32(0)
	numSkipped := int32(0)
	numExecutions := 100

	mismatchWorkflowHistory := getTestReplayWorkflowMismatchHistory(s.T())
	for i := 0; i != numExecutions; i++ {
		switch rand.Intn(3) {
		case 0:
			numSucceed++
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
				History: s.testWorkflowHistory,
			}, nil).Times(1)
		case 1:
			numSkipped++
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(nil, &shared.EntityNotExistsError{}).Times(1)
		case 2:
			numFailed++
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
				History: mismatchWorkflowHistory,
			}, nil).Times(1)
		}
	}

	params := newTestReplayWorkflowActivityParams(numExecutions)

	resultValue, err := s.env.ExecuteActivity(shadower.ReplayWorkflowActivityName, params)
	s.NoError(err)

	var result shadower.ReplayWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.Equal(numSucceed, result.GetSucceeded())
	s.Equal(numSkipped, result.GetSkipped())
	s.Equal(numFailed, result.GetFailed())
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_WorkflowNotRegistered() {
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: getTestReplayWorkflowLocalActivityHistory(s.T()), // this workflow type is not registered
	}, nil).Times(1)

	params := newTestReplayWorkflowActivityParams(5)

	_, err := s.env.ExecuteActivity(shadower.ReplayWorkflowActivityName, params)
	s.Error(err)
	s.Equal(shadower.ErrReasonWorkflowTypeNotRegistered, err.Error())
}

func newTestReplayWorkflowActivityParams(numExecutions int) shadower.ReplayWorkflowActivityParams {
	params := shadower.ReplayWorkflowActivityParams{
		Domain:     common.StringPtr(defaultTestDomain),
		Executions: make([]*shared.WorkflowExecution, 0, numExecutions),
	}
	for i := 0; i != numExecutions; i++ {
		params.Executions = append(params.Executions, &shared.WorkflowExecution{})
	}
	return params
}
