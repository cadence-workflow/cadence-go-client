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
	"math"
	"strings"
	"time"

	"go.uber.org/cadence/.gen/go/shared"
)

type (
	// WorkflowStatus represents the status of a workflow
	WorkflowStatus string
)

const (
	// WorkflowStatusOpen is the WorkflowStatus for open workflows
	WorkflowStatusOpen WorkflowStatus = "OPEN"
	// WorkflowStatusClosed is the WorkflowStatus for closed workflows
	WorkflowStatusClosed WorkflowStatus = "CLOSED"
	// WorkflowStatusALL is the WorkflowStatus for all workflows
	WorkflowStatusALL WorkflowStatus = "ALL"
)

var (
	// WorkflowStatusCompleted is the WorkflowStatus for completed workflow
	WorkflowStatusCompleted = WorkflowStatus(shared.WorkflowExecutionCloseStatusCompleted.String())
	// WorkflowStatusFailed is the WorkflowStatus for failed workflows
	WorkflowStatusFailed = WorkflowStatus(shared.WorkflowExecutionCloseStatusFailed.String())
	// WorkflowStatusCanceled is the WorkflowStatus for canceled workflows
	WorkflowStatusCanceled = WorkflowStatus(shared.WorkflowExecutionCloseStatusCanceled.String())
	// WorkflowStatusTerminated is the WorkflowStatus for terminated workflows
	WorkflowStatusTerminated = WorkflowStatus(shared.WorkflowExecutionCloseStatusTerminated.String())
	// WorkflowStatusContinuedAsNew is the WorkflowStatus for continuedAsNew workflows
	WorkflowStatusContinuedAsNew = WorkflowStatus(shared.WorkflowExecutionCloseStatusContinuedAsNew.String())
	// WorkflowStatusTimedOut is the WorkflowStatus for timedout workflows
	WorkflowStatusTimedOut = WorkflowStatus(shared.WorkflowExecutionCloseStatusTimedOut.String())
)

const (
	keyWorkflowType = "WorkflowType"
	keyCloseStatus  = "CloseStatus"
	keyStartTime    = "StartTime"
	keyCloseTime    = "CloseTime"
)

var (
	maxTimestamp = time.Unix(0, math.MaxInt64)
)

type (
	// QueryBuilder builds visibility query. It's shadower's own Query builders that processes the shadow filter
	// options into a query to pull the required workflows.

	QueryBuilder interface {
		WorkflowTypes([]string) QueryBuilder
		ExcludeWorkflowTypes([]string) QueryBuilder
		WorkflowStatus([]WorkflowStatus) QueryBuilder
		StartTime(time.Time, time.Time) QueryBuilder
		CloseTime(time.Time, time.Time) QueryBuilder
		Build() string
	}

	queryBuilderImpl struct {
		builder strings.Builder
	}
)

// NewQueryBuilder creates a new visibility QueryBuilder
func NewQueryBuilder() QueryBuilder {
	return &queryBuilderImpl{}
}

func (q *queryBuilderImpl) WorkflowTypes(types []string) QueryBuilder {
	workflowTypeQueries := make([]string, 0, len(types))
	for _, workflowType := range types {
		workflowTypeQueries = append(workflowTypeQueries, fmt.Sprintf(keyWorkflowType+` = "%v"`, workflowType))
	}
	q.appendPartialQuery(strings.Join(workflowTypeQueries, " or "))
	return q
}

func (q *queryBuilderImpl) ExcludeWorkflowTypes(types []string) QueryBuilder {
	if len(types) == 0 {
		return q
	}
	excludeTypeQueries := make([]string, 0, len(types))
	for _, workflowType := range types {
		excludeTypeQueries = append(excludeTypeQueries, fmt.Sprintf(keyWorkflowType+` != "%v"`, workflowType))
	}
	q.appendPartialQuery(strings.Join(excludeTypeQueries, " and "))
	return q
}

func (q *queryBuilderImpl) WorkflowStatus(statuses []WorkflowStatus) QueryBuilder {
	workflowStatusQueries := make([]string, 0, len(statuses))
	for _, status := range statuses {
		var statusQuery string
		switch status {
		case WorkflowStatusOpen:
			statusQuery = keyCloseTime + " = missing"
		case WorkflowStatusClosed:
			statusQuery = keyCloseTime + " != missing"
		case WorkflowStatusALL:
			// no query needed
			return q
		default:
			statusQuery = keyCloseStatus + ` = "` + string(status) + `"`
		}
		workflowStatusQueries = append(workflowStatusQueries, statusQuery)
	}
	q.appendPartialQuery(strings.Join(workflowStatusQueries, " or "))
	return q
}

func (q *queryBuilderImpl) StartTime(minStartTime, maxStartTime time.Time) QueryBuilder {
	startTimeQueries := make([]string, 0, 2)
	if !minStartTime.IsZero() {
		startTimeQueries = append(startTimeQueries, fmt.Sprintf(keyStartTime+` >= %v`, minStartTime.UnixNano()))
	}
	if !maxStartTime.Equal(maxTimestamp) {
		startTimeQueries = append(startTimeQueries, fmt.Sprintf(keyStartTime+` <= %v`, maxStartTime.UnixNano()))
	}

	q.appendPartialQuery(strings.Join(startTimeQueries, " and "))
	return q
}

func (q *queryBuilderImpl) CloseTime(minCloseTime, maxCloseTime time.Time) QueryBuilder {
	CloseTimeQueries := make([]string, 0, 2)
	if !minCloseTime.IsZero() {
		CloseTimeQueries = append(CloseTimeQueries, fmt.Sprintf(keyCloseTime+` >= %v`, minCloseTime.UnixNano()))
	}
	if !maxCloseTime.Equal(maxTimestamp) {
		CloseTimeQueries = append(CloseTimeQueries, fmt.Sprintf(keyCloseTime+` <= %v`, maxCloseTime.UnixNano()))
	}

	q.appendPartialQuery(strings.Join(CloseTimeQueries, " and "))
	return q
}

func (q *queryBuilderImpl) Build() string {
	return q.builder.String()
}

func (q *queryBuilderImpl) appendPartialQuery(query string) {
	if len(query) == 0 {
		return
	}

	if q.builder.Len() != 0 {
		q.builder.WriteString(" and ")
	}

	q.builder.WriteRune('(')
	q.builder.WriteString(query)
	q.builder.WriteRune(')')
}

// ToWorkflowStatus converts workflow status from string type to WorkflowStatus type
func ToWorkflowStatus(statusString string) (WorkflowStatus, error) {
	status := WorkflowStatus(strings.ToUpper(statusString))
	switch status {
	case WorkflowStatusOpen, WorkflowStatusClosed, WorkflowStatusCompleted,
		WorkflowStatusFailed, WorkflowStatusCanceled, WorkflowStatusTerminated,
		WorkflowStatusContinuedAsNew, WorkflowStatusTimedOut, WorkflowStatusALL:
		return status, nil
	default:
		return "", fmt.Errorf("unknown workflow status: %v", statusString)
	}
}
