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
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/facebookgo/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

type workflowShadowerSuite struct {
	*require.Assertions
	suite.Suite

	controller  *gomock.Controller
	mockService *workflowservicetest.MockClient

	testShadower        *WorkflowShadower
	testWorkflowHistory *shared.History
	testTimestamp       time.Time
}

func TestWorkflowShadowerSuite(t *testing.T) {
	s := new(workflowShadowerSuite)
	suite.Run(t, s)
}

func (s *workflowShadowerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockService = workflowservicetest.NewMockClient(s.controller)

	var err error
	s.testShadower, err = NewWorkflowShadower(s.mockService, "testDomain", ShadowOptions{}, ReplayOptions{}, nil)
	s.NoError(err)

	// overwrite shadower clock to be a mock clock
	s.testShadower.clock = clock.NewMock()
	// register test workflow
	s.testShadower.RegisterWorkflow(testReplayWorkflow)

	s.testWorkflowHistory = getTestReplayWorkflowFullHistory(s.T())

	s.testTimestamp = time.Now()
}

func (s *workflowShadowerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *workflowShadowerSuite) TestTimeFilterValidation() {
	testCases := []struct {
		msg          string
		timeFilter   TimeFilter
		expectErr    bool
		validationFn func(TimeFilter)
	}{
		{
			msg: "maxTimestamp before minTimestamp",
			timeFilter: TimeFilter{
				MinTimestamp: s.testTimestamp.Add(time.Hour),
				MaxTimestamp: s.testTimestamp,
			},
			expectErr: true,
		},
		{
			msg:        "neither timestamp is specified",
			timeFilter: TimeFilter{},
			expectErr:  false,
			validationFn: func(f TimeFilter) {
				s.True(f.MinTimestamp.IsZero())
				s.True(f.MaxTimestamp.Equal(maxTimestamp))
			},
		},
		{
			msg: "only min timestamp is specified",
			timeFilter: TimeFilter{
				MinTimestamp: s.testTimestamp,
			},
			expectErr: false,
			validationFn: func(f TimeFilter) {
				s.True(f.MinTimestamp.Equal(s.testTimestamp))
				s.True(f.MaxTimestamp.Equal(maxTimestamp))
			},
		},
		{
			msg: "only max timestamp is specified",
			timeFilter: TimeFilter{
				MaxTimestamp: s.testTimestamp,
			},
			expectErr: false,
			validationFn: func(f TimeFilter) {
				s.True(f.MinTimestamp.IsZero())
				s.True(f.MaxTimestamp.Equal(s.testTimestamp))
			},
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			err := test.timeFilter.validateAndPopulateFields()
			if test.expectErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			test.validationFn(test.timeFilter)
		})
	}
}

func (s *workflowShadowerSuite) TestTimeFilterIsEmpty() {
	testCases := []struct {
		msg     string
		filter  *TimeFilter
		isEmpty bool
	}{
		{
			msg:     "nil pointer",
			filter:  nil,
			isEmpty: true,
		},
		{
			msg:     "neither field is specified",
			filter:  &TimeFilter{},
			isEmpty: true,
		},
		{
			msg: "not empty",
			filter: &TimeFilter{
				MaxTimestamp: time.Now(),
			},
			isEmpty: false,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			s.Equal(test.isEmpty, test.filter.isEmpty())
		})
	}
}

func (s *workflowShadowerSuite) TestShadowOptionsValidation() {
	testCases := []struct {
		msg          string
		options      ShadowOptions
		expectErr    bool
		validationFn func(*ShadowOptions)
	}{
		{
			msg: "exit condition not specified in continuous mode",
			options: ShadowOptions{
				Mode: ShadowModeContinuous,
			},
			expectErr: true,
		},
		{
			msg: "both query and other filters are specified",
			options: ShadowOptions{
				WorkflowQuery: "some random query",
				WorkflowStartTimeFilter: TimeFilter{
					MinTimestamp: time.Now(),
				},
			},
			expectErr: true,
		},
		{
			msg:       "populate sampling rate, concurrency and status",
			options:   ShadowOptions{},
			expectErr: false,
			validationFn: func(options *ShadowOptions) {
				s.Equal("(CloseTime = missing)", options.WorkflowQuery)
				s.Equal(1.0, options.SamplingRate)
				s.Equal(1, options.Concurrency)
			},
		},
		{
			msg: "construct query",
			options: ShadowOptions{
				WorkflowTypes:  []string{"testWorkflowType"},
				WorkflowStatus: []string{"open"},
				WorkflowStartTimeFilter: TimeFilter{
					MinTimestamp: s.testTimestamp.Add(-time.Hour),
					MaxTimestamp: s.testTimestamp,
				},
			},
			expectErr: false,
			validationFn: func(options *ShadowOptions) {
				expectedQuery := NewQueryBuilder().
					WorkflowTypes([]string{"testWorkflowType"}).
					WorkflowStatus([]WorkflowStatus{WorkflowStatusOpen}).
					StartTime(
						s.testTimestamp.Add(-time.Hour),
						s.testTimestamp,
					).Build()

				s.Equal(expectedQuery, options.WorkflowQuery)
			},
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			err := test.options.validateAndPopulateFields()
			if test.expectErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			test.validationFn(&test.options)
		})
	}
}

func (s *workflowShadowerSuite) TestShadowOptionsWithExcludeTypes() {
	excludeTypes := []string{"excludedType1", "excludedType2"}
	options := ShadowOptions{
		WorkflowTypes: []string{"includedType1", "includedType2"},
		ExcludeTypes:  excludeTypes,
		Mode:          ShadowModeNormal,
	}
	expectedQuery := fmt.Sprintf(
		`(WorkflowType = "includedType1" or WorkflowType = "includedType2") and (WorkflowType != "excludedType1" and WorkflowType != "excludedType2") and (CloseTime = missing)`,
	)
	shadower, err := NewWorkflowShadower(s.mockService, "testDomain", options, ReplayOptions{}, nil)
	s.NoError(err)
	s.mockService.EXPECT().
		ListWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...interface{}) (*shared.ListWorkflowExecutionsResponse, error) {
			s.Equal(expectedQuery, *request.Query)
			return &shared.ListWorkflowExecutionsResponse{
				Executions:    nil,
				NextPageToken: nil,
			}, nil
		}).Times(1)
	s.NoError(shadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_ExpirationTime() {
	totalWorkflows := 50
	timePerWorkflow := 7 * time.Second
	expirationTime := time.Minute

	s.testShadower.shadowOptions.ExitCondition = ShadowExitCondition{
		ExpirationInterval: expirationTime,
	}

	s.mockService.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(totalWorkflows),
		NextPageToken: nil,
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).DoAndReturn(func(...interface{}) (*shared.GetWorkflowExecutionHistoryResponse, error) {
		s.testShadower.clock.(*clock.Mock).Add(timePerWorkflow)
		return &shared.GetWorkflowExecutionHistoryResponse{
			History: s.testWorkflowHistory,
		}, nil
	}).Times(int(expirationTime/timePerWorkflow) + 1)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_MaxShadowingCount() {
	maxShadowCount := 50

	s.testShadower.shadowOptions.ExitCondition = ShadowExitCondition{
		ShadowCount: maxShadowCount,
	}

	s.mockService.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(maxShadowCount * 2),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(maxShadowCount)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorker_NormalMode() {
	workflowExecutions := newTestWorkflowExecutions(10)
	numScan := 3
	totalWorkflows := len(workflowExecutions) * numScan

	for i := 0; i != numScan; i++ {
		scanResp := &shared.ListWorkflowExecutionsResponse{
			Executions:    workflowExecutions,
			NextPageToken: []byte{1, 2, 3},
		}
		if i == numScan-1 {
			scanResp.NextPageToken = nil
		}
		s.mockService.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(scanResp, nil).Times(1)
	}

	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(totalWorkflows)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorker_ContinuousMode() {
	workflowExecutions := newTestWorkflowExecutions(10)
	numScan := 3
	totalWorkflows := len(workflowExecutions) * numScan

	s.testShadower.shadowOptions.Mode = ShadowModeContinuous
	s.testShadower.shadowOptions.ExitCondition = ShadowExitCondition{
		ShadowCount: totalWorkflows,
	}

	for i := 0; i != numScan; i++ {
		scanResp := &shared.ListWorkflowExecutionsResponse{
			Executions: workflowExecutions,
		}
		s.mockService.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(scanResp, nil).Times(1)
	}

	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(totalWorkflows)

	doneCh := make(chan struct{})
	var advanceTimeWG sync.WaitGroup
	advanceTimeWG.Add(1)
	go func() {
		defer advanceTimeWG.Done()
		for {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-doneCh:
				return
			default:
				s.testShadower.clock.(*clock.Mock).Add(defaultWaitDurationPerIteration)
			}
		}
	}()

	s.NoError(s.testShadower.shadowWorker())
	close(doneCh)
	advanceTimeWG.Wait()
}

func (s *workflowShadowerSuite) TestShadowWorker_ReplayFailed() {
	successfullyReplayed := 5
	s.mockService.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(successfullyReplayed * 2),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(successfullyReplayed)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: getTestReplayWorkflowMismatchHistory(s.T()),
	}, nil).Times(1)

	s.Error(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorker_ExpectedReplayError() {
	testCases := []struct {
		msg                string
		getHistoryErr      error
		getHistoryResponse *shared.GetWorkflowExecutionHistoryResponse
	}{
		{
			msg:           "only workflow started event", // for example cron workflow
			getHistoryErr: nil,
			getHistoryResponse: &shared.GetWorkflowExecutionHistoryResponse{
				History: &shared.History{Events: []*shared.HistoryEvent{
					createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
						WorkflowType: &shared.WorkflowType{Name: common.StringPtr("testWorkflow")},
						TaskList:     &shared.TaskList{Name: common.StringPtr("taskList")},
						Input:        testEncodeFunctionArgs(s.T(), getDefaultDataConverter()),
						CronSchedule: common.StringPtr("* * * * *"),
					}),
				},
				},
			},
		},
		{
			msg:                "workflow not exist",
			getHistoryErr:      &shared.EntityNotExistsError{Message: "Workflow passed retention date"},
			getHistoryResponse: nil,
		},
		{
			msg:                "corrupted workflow history", // for example cron workflow
			getHistoryErr:      &shared.InternalServiceError{Message: "History events not continuous"},
			getHistoryResponse: nil,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			s.mockService.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions()...).Return(&shared.ListWorkflowExecutionsResponse{
				Executions:    newTestWorkflowExecutions(1),
				NextPageToken: nil,
			}, nil).Times(1)
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(test.getHistoryResponse, test.getHistoryErr).Times(1)

			s.NoError(s.testShadower.shadowWorker())
		})
	}
}

func (s *workflowShadowerSuite) TestWorkflowRegistration() {
	wfName := s.testShadower.GetRegisteredWorkflows()[0].WorkflowType().Name
	fnName := getFunctionName(testReplayWorkflow)
	s.Equal(fnName, wfName)

	fn := s.testShadower.GetRegisteredWorkflows()[0].GetFunction()
	s.Equal(fnName, runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name())
}

func newTestWorkflowExecutions(size int) []*shared.WorkflowExecutionInfo {
	executions := make([]*shared.WorkflowExecutionInfo, size)
	for i := 0; i != size; i++ {
		executions[i] = &shared.WorkflowExecutionInfo{
			Execution: &shared.WorkflowExecution{
				WorkflowId: common.StringPtr("workflowID"),
				RunId:      common.StringPtr("runID"),
			},
		}
	}
	return executions
}
