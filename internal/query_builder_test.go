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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type queryBuilderSuite struct {
	*require.Assertions
	suite.Suite
}

func TestQueryBuilderSuite(t *testing.T) {
	s := new(queryBuilderSuite)
	suite.Run(t, s)
}

func (s *queryBuilderSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *queryBuilderSuite) TestWorkflowTypeQuery() {
	testCases := []struct {
		msg           string
		workflowTypes []string
		expectedQuery string
	}{
		{
			msg:           "empty workflowTypes",
			workflowTypes: nil,
			expectedQuery: "",
		},
		{
			msg:           "single workflowType",
			workflowTypes: []string{"testWorkflowType"},
			expectedQuery: `(WorkflowType = "testWorkflowType")`,
		},
		{
			msg:           "multiple workflowTypes",
			workflowTypes: []string{"testWorkflowType1", "testWorkflowType2"},
			expectedQuery: `(WorkflowType = "testWorkflowType1" or WorkflowType = "testWorkflowType2")`,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			builder := NewQueryBuilder()
			builder.WorkflowTypes(test.workflowTypes)
			s.Equal(test.expectedQuery, builder.Build())
		})
	}
}

func (s *queryBuilderSuite) TestWorkflowStatusQuery() {
	testCases := []struct {
		msg              string
		workflowStatuses []WorkflowStatus
		expectedQuery    string
	}{
		{
			msg:              "empty workflow status",
			workflowStatuses: []WorkflowStatus{},
			expectedQuery:    "",
		},
		{
			msg:              "open workflow",
			workflowStatuses: []WorkflowStatus{WorkflowStatusOpen},
			expectedQuery:    "(CloseTime = missing)",
		},
		{
			msg:              "closed workflow",
			workflowStatuses: []WorkflowStatus{WorkflowStatusClosed},
			expectedQuery:    "(CloseTime != missing)",
		},
		{
			msg:              "multiple workflow statuses",
			workflowStatuses: []WorkflowStatus{WorkflowStatusFailed, WorkflowStatusTimedOut},
			expectedQuery:    `(CloseStatus = "FAILED" or CloseStatus = "TIMED_OUT")`,
		},
		{
			msg:              "all workflows",
			workflowStatuses: []WorkflowStatus{WorkflowStatusALL},
			expectedQuery:    "",
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			builder := NewQueryBuilder()
			builder.WorkflowStatus(test.workflowStatuses)
			s.Equal(test.expectedQuery, builder.Build())
		})
	}
}

func (s *queryBuilderSuite) TestStartTimeQuery() {
	testTimestamp := time.Now()
	testCases := []struct {
		msg           string
		minStartTime  time.Time
		maxStartTime  time.Time
		expectedQuery string
	}{
		{
			msg:           "empty minTimestamp",
			maxStartTime:  testTimestamp,
			expectedQuery: fmt.Sprintf("(StartTime <= %v)", testTimestamp.UnixNano()),
		},
		{
			msg:           "max maxTimestamp",
			minStartTime:  testTimestamp,
			maxStartTime:  maxTimestamp,
			expectedQuery: fmt.Sprintf("(StartTime >= %v)", testTimestamp.UnixNano()),
		},
		{
			msg:           "both timestamps are used",
			minStartTime:  testTimestamp.Add(-time.Hour),
			maxStartTime:  testTimestamp,
			expectedQuery: fmt.Sprintf("(StartTime >= %v and StartTime <= %v)", testTimestamp.Add(-time.Hour).UnixNano(), testTimestamp.UnixNano()),
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			builder := NewQueryBuilder()
			builder.StartTime(test.minStartTime, test.maxStartTime)
			s.Equal(test.expectedQuery, builder.Build())
		})
	}
}

func (s *queryBuilderSuite) TestExcludeWorkflowTypesQuery() {
	testCases := []struct {
		msg           string
		excludeTypes  []string
		expectedQuery string
	}{
		{
			msg:           "empty excludeTypes",
			excludeTypes:  nil,
			expectedQuery: "",
		},
		{
			msg:           "single excludeType",
			excludeTypes:  []string{"excludedWorkflowType"},
			expectedQuery: `(WorkflowType != "excludedWorkflowType")`,
		},
		{
			msg:           "multiple excludeTypes",
			excludeTypes:  []string{"excludedWorkflowType1", "excludedWorkflowType2"},
			expectedQuery: `(WorkflowType != "excludedWorkflowType1" and WorkflowType != "excludedWorkflowType2")`,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			builder := NewQueryBuilder()
			builder.ExcludeWorkflowTypes(test.excludeTypes)
			s.Equal(test.expectedQuery, builder.Build())
		})
	}
}

func (s *queryBuilderSuite) TestMultipleFilters() {
	maxStartTime := time.Now()
	minStartTime := maxStartTime.Add(-time.Hour)

	builder := NewQueryBuilder().
		WorkflowTypes([]string{"testWorkflowType1", "testWorkflowType2"}).
		WorkflowStatus([]WorkflowStatus{WorkflowStatusOpen}).
		StartTime(minStartTime, maxStartTime)

	expectedQuery := fmt.Sprintf(`(WorkflowType = "testWorkflowType1" or WorkflowType = "testWorkflowType2") and (CloseTime = missing) and (StartTime >= %v and StartTime <= %v)`,
		minStartTime.UnixNano(),
		maxStartTime.UnixNano(),
	)
	s.Equal(expectedQuery, builder.Build())
}

func (s *queryBuilderSuite) TestToWorkflowStatus() {
	testCases := []struct {
		msg            string
		statusString   string
		expectErr      bool
		expectedStatus WorkflowStatus
	}{
		{
			msg:          "unknown status",
			statusString: "unknown",
			expectErr:    true,
		},
		{
			msg:            "lower case status string",
			statusString:   "open",
			expectErr:      false,
			expectedStatus: WorkflowStatusOpen,
		},
		{
			msg:            "mixed case status string",
			statusString:   "Timed_Out",
			expectErr:      false,
			expectedStatus: WorkflowStatusTimedOut,
		},
		{
			msg:            "all",
			statusString:   "ALL",
			expectErr:      false,
			expectedStatus: WorkflowStatusALL,
		},
		{
			msg:            "closed",
			statusString:   "CLOSED",
			expectErr:      false,
			expectedStatus: WorkflowStatusClosed,
		},
		{
			msg:            "failed",
			statusString:   "FAILED",
			expectErr:      false,
			expectedStatus: WorkflowStatusFailed,
		},
		{
			msg:            "completed",
			statusString:   "COMPLETED",
			expectErr:      false,
			expectedStatus: WorkflowStatusCompleted,
		},
		{
			msg:            "canceled",
			statusString:   "CANCELED",
			expectErr:      false,
			expectedStatus: WorkflowStatusCanceled,
		},
		{
			msg:            "continued as new",
			statusString:   "CONTINUED_AS_NEW",
			expectErr:      false,
			expectedStatus: WorkflowStatusContinuedAsNew,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			actualStatus, err := ToWorkflowStatus(test.statusString)
			if test.expectErr {
				s.Error(err)
				return
			}

			s.Equal(test.expectedStatus, actualStatus)
		})
	}
}
