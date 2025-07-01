// Copyright (c) 2018 Uber Technologies, Inc.
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
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/metrics"
	"go.uber.org/cadence/internal/common/serializer"
)

const (
	domain                = "some random domain"
	workflowID            = "some random workflow ID"
	workflowType          = "some random workflow type"
	runID                 = "some random run ID"
	activityID            = "1234"
	tasklist              = "some random tasklist"
	identity              = "some random identity"
	taskToken             = "some random task-token"
	timeoutInSeconds      = 20
	workflowIDReusePolicy = WorkflowIDReusePolicyAllowDuplicateFailedOnly
	testHeader            = "test-header"
)

// historyEventIteratorSuite

type (
	historyEventIteratorSuite struct {
		suite.Suite
		mockCtrl              *gomock.Controller
		workflowServiceClient *workflowservicetest.MockClient
		wfClient              *workflowClient
	}
)

// stringMapPropagator propagates the list of keys across a workflow,
// interpreting the payloads as strings.
type stringMapPropagator struct {
	keys map[string]struct{}
}

// NewStringMapPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewStringMapPropagator(keys []string) ContextPropagator {
	keyMap := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keyMap[key] = struct{}{}
	}
	return &stringMapPropagator{keyMap}
}

// Inject injects values from context into headers for propagation
func (s *stringMapPropagator) Inject(ctx context.Context, writer HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(contextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *stringMapPropagator) InjectFromWorkflow(ctx Context, writer HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(contextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *stringMapPropagator) Extract(ctx context.Context, reader HeaderReader) (context.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = context.WithValue(ctx, contextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *stringMapPropagator) ExtractToWorkflow(ctx Context, reader HeaderReader) (Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = WithValue(ctx, contextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

func TestHistoryEventIteratorSuite(t *testing.T) {
	s := new(historyEventIteratorSuite)
	suite.Run(t, s)
}

func (s *historyEventIteratorSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

func (s *historyEventIteratorSuite) SetupTest() {
	// Create service endpoint
	s.mockCtrl = gomock.NewController(s.T())
	s.workflowServiceClient = workflowservicetest.NewMockClient(s.mockCtrl)

	s.wfClient = &workflowClient{
		workflowService: s.workflowServiceClient,
		domain:          domain,
	}
}

func (s *historyEventIteratorSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *historyEventIteratorSuite) TestIterator_NoError() {
	filterType := shared.HistoryEventFilterTypeAllEvent
	request1 := getGetWorkflowExecutionHistoryRequest(filterType)
	response1 := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				// dummy history event
				&shared.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	request2 := getGetWorkflowExecutionHistoryRequest(filterType)
	request2.NextPageToken = response1.NextPageToken
	response2 := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				// dummy history event
				&shared.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	dummyEvent := []*shared.HistoryEvent{
		// dummy history event
		&shared.HistoryEvent{},
	}

	blobData := serializeEvents(dummyEvent)
	request3 := getGetWorkflowExecutionHistoryRequest(filterType)
	request3.NextPageToken = response2.NextPageToken
	response3 := &shared.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*shared.DataBlob{
			// dummy history event
			blobData,
		},
		NextPageToken: nil,
	}

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request1, gomock.Any()).Return(response1, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request2, gomock.Any()).Return(response2, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request3, gomock.Any()).Return(response3, nil).Times(1)

	events := []*shared.HistoryEvent{}
	iter := s.wfClient.GetWorkflowHistory(context.Background(), workflowID, runID, true, shared.HistoryEventFilterTypeAllEvent)
	for iter.HasNext() {
		event, err := iter.Next()
		s.NoError(err)
		events = append(events, event)
	}
	s.Equal(3, len(events))
}

func (s *historyEventIteratorSuite) TestIterator_NoError_EmptyPage() {
	filterType := shared.HistoryEventFilterTypeAllEvent
	request1 := getGetWorkflowExecutionHistoryRequest(filterType)
	response1 := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	request2 := getGetWorkflowExecutionHistoryRequest(filterType)
	request2.NextPageToken = response1.NextPageToken
	response2 := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				// dummy history event
				&shared.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	dummyEvent := []*shared.HistoryEvent{
		// dummy history event
		&shared.HistoryEvent{},
	}

	blobData := serializeEvents(dummyEvent)
	request3 := getGetWorkflowExecutionHistoryRequest(filterType)
	request3.NextPageToken = response2.NextPageToken
	response3 := &shared.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*shared.DataBlob{
			// dummy history event
			blobData,
		},
		NextPageToken: nil,
	}

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request1, gomock.Any()).Return(response1, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request2, gomock.Any()).Return(response2, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request3, gomock.Any()).Return(response3, nil).Times(1)

	events := []*shared.HistoryEvent{}
	iter := s.wfClient.GetWorkflowHistory(context.Background(), workflowID, runID, true, shared.HistoryEventFilterTypeAllEvent)
	for iter.HasNext() {
		event, err := iter.Next()
		s.NoError(err)
		events = append(events, event)
	}
	s.Equal(2, len(events))
}

func (s *historyEventIteratorSuite) TestIterator_RPCError() {
	filterType := shared.HistoryEventFilterTypeAllEvent
	request1 := getGetWorkflowExecutionHistoryRequest(filterType)
	response1 := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				// dummy history event
				&shared.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	request2 := getGetWorkflowExecutionHistoryRequest(filterType)
	request2.NextPageToken = response1.NextPageToken

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request1, gomock.Any()).Return(response1, nil).Times(1)

	iter := s.wfClient.GetWorkflowHistory(context.Background(), workflowID, runID, true, shared.HistoryEventFilterTypeAllEvent)

	s.True(iter.HasNext())
	event, err := iter.Next()
	s.NotNil(event)
	s.NoError(err)

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request2, gomock.Any()).Return(nil, &shared.EntityNotExistsError{}).Times(1)

	s.True(iter.HasNext())
	event, err = iter.Next()
	s.Nil(event)
	s.Error(err)
}

func (s *historyEventIteratorSuite) TestIterator_StopsTryingNearTimeout() {
	// ensuring "when GetWorkflow().Get(...) times out while waiting", we return a timed-out error of some kind,
	// and stop sending requests rather than trying again and getting some other kind of error.
	// historically this led to a "bad request", insufficient time left for long poll, which was confusing and noisy.

	filterType := shared.HistoryEventFilterTypeCloseEvent
	reqNormal := getGetWorkflowExecutionHistoryRequest(filterType)
	reqFinal := getGetWorkflowExecutionHistoryRequest(filterType)

	// all items filtered out for both requests
	resEmpty := &shared.GetWorkflowExecutionHistoryResponse{
		History:       &shared.History{Events: nil}, // this or RawHistory must be non-nil, but they can be empty
		NextPageToken: []byte{1, 2, 3, 4, 5},
	}
	reqFinal.NextPageToken = resEmpty.NextPageToken

	s.True(time.Second < (defaultGetHistoryTimeoutInSecs*time.Second), "sanity check: default timeout must be longer than how long we extend the timeout for tests")

	// begin with a deadline that is long enough to allow 2 requests
	d := time.Now().Add((defaultGetHistoryTimeoutInSecs * time.Second) + time.Second)
	baseCtx, cancel := context.WithDeadline(context.Background(), d)
	defer cancel()
	ctx := &fakeDeadlineContext{baseCtx, d}

	// prep the iterator
	iter := s.wfClient.GetWorkflowHistory(ctx, workflowID, runID, true, filterType)

	// first attempt should occur, and trigger a second request with less than the requested timeout
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), reqNormal, gomock.Any()).DoAndReturn(func(_ context.Context, _ *shared.GetWorkflowExecutionHistoryRequest, _ ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
		// first request is being sent, modify the context to simulate time passing,
		// and give the second request insufficient time to trigger another poll.
		//
		// without this, you should see an attempt at a third call, as the context is not canceled,
		// and this mock took no time to respond.
		ctx.d = time.Now().Add(time.Second)
		return resEmpty, nil
	}).Times(1)
	// second request should occur, but not another
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), reqFinal, gomock.Any()).Return(resEmpty, nil).Times(1)

	// trigger both paginated requests as part of the single HasNext call
	s.True(iter.HasNext())
	event, err := iter.Next()
	s.Nil(event, "iterator should not have returned any events")
	s.Error(err, "iterator should have errored")
	// canceled may also be appropriate, but currently this is true
	s.Truef(errors.Is(err, context.DeadlineExceeded), "iterator should have returned a deadline-exceeded error, but returned a: %#v", err)
	s.Contains(err.Error(), "waiting for the workflow to finish", "should be descriptive of what happened")
}

// minor helper type to allow faking deadlines between calls, as we cannot normally modify a context that way.
type fakeDeadlineContext struct {
	context.Context

	d time.Time
}

func (f *fakeDeadlineContext) Deadline() (time.Time, bool) {
	return f.d, true
}

// workflowRunSuite

type (
	workflowRunSuite struct {
		suite.Suite
		mockCtrl              *gomock.Controller
		workflowServiceClient *workflowservicetest.MockClient
		workflowClient        Client
	}
)

func TestWorkflowRunSuite(t *testing.T) {
	s := new(workflowRunSuite)
	suite.Run(t, s)
}

func (s *workflowRunSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

func (s *workflowRunSuite) TearDownSuite() {

}

func (s *workflowRunSuite) SetupTest() {
	// Create service endpoint
	s.mockCtrl = gomock.NewController(s.T())
	s.workflowServiceClient = workflowservicetest.NewMockClient(s.mockCtrl)

	metricsScope := metrics.NewTaggedScope(nil)
	options := &ClientOptions{
		MetricsScope: metricsScope,
		Identity:     identity,
	}
	s.workflowClient = NewClient(s.workflowServiceClient, domain, options)
}

func (s *workflowRunSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Success() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType := shared.EventTypeWorkflowExecutionCompleted
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType,
					WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
						Result: encodedResult,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions()...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_RawHistory_Success() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType := shared.EventTypeWorkflowExecutionCompleted
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	events := []*shared.HistoryEvent{
		&shared.HistoryEvent{
			EventType: &eventType,
			WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
				Result: encodedResult,
			},
		},
	}

	blobData := serializeEvents(events)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*shared.DataBlob{
			blobData,
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions()...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflowWorkflowExecutionAlreadyStartedError() {
	alreadyStartedErr := &shared.WorkflowExecutionAlreadyStartedError{
		RunId:          common.StringPtr(runID),
		Message:        common.StringPtr("Already Started"),
		StartRequestId: common.StringPtr(uuid.NewRandom().String()),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).
		Return(nil, alreadyStartedErr).Times(1)

	eventType := shared.EventTypeWorkflowExecutionCompleted
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				{
					EventType: &eventType,
					WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
						Result: encodedResult,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).
		Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *shared.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		workflowID := getRequest.Execution.WorkflowId
		s.NotNil(workflowID)
		s.NotEmpty(*workflowID)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflowWorkflowExecutionAlreadyStartedError_RawHistory() {
	alreadyStartedErr := &shared.WorkflowExecutionAlreadyStartedError{
		RunId:          common.StringPtr(runID),
		Message:        common.StringPtr("Already Started"),
		StartRequestId: common.StringPtr(uuid.NewRandom().String()),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).
		Return(nil, alreadyStartedErr).Times(1)

	eventType := shared.EventTypeWorkflowExecutionCompleted
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	events := []*shared.HistoryEvent{
		{
			EventType: &eventType,
			WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
				Result: encodedResult,
			},
		},
	}

	blobData := serializeEvents(events)

	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*shared.DataBlob{
			blobData,
		},
		NextPageToken: nil,
	}
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).
		Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *shared.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		workflowID := getRequest.Execution.WorkflowId
		s.NotNil(workflowID)
		s.NotEmpty(*workflowID)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
}

// Test for the bug in ExecuteWorkflow.
// When Options.ID was empty then GetWorkflowExecutionHistory was called with an empty WorkflowID.
func (s *workflowRunSuite) TestExecuteWorkflow_NoIdInOptions() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	eventType := shared.EventTypeWorkflowExecutionCompleted
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				{
					EventType: &eventType,
					WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
						Result: encodedResult,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	var wid *string
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *shared.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		wid = getRequest.Execution.WorkflowId
		s.NotNil(wid)
		s.NotEmpty(*wid)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
	s.Equal(workflowRun.GetID(), *wid)
}

// Test for the bug in ExecuteWorkflow in the case of raw history returned from API.
// When Options.ID was empty then GetWorkflowExecutionHistory was called with an empty WorkflowID.
func (s *workflowRunSuite) TestExecuteWorkflow_NoIdInOptions_RawHistory() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	eventType := shared.EventTypeWorkflowExecutionCompleted
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	events := []*shared.HistoryEvent{
		{
			EventType: &eventType,
			WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
				Result: encodedResult,
			},
		},
	}

	blobData := serializeEvents(events)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*shared.DataBlob{
			blobData,
		},
		NextPageToken: nil,
	}
	var wid *string
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions()...).Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *shared.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		wid = getRequest.Execution.WorkflowId
		s.NotNil(wid)
		s.NotEmpty(*wid)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
	s.Equal(workflowRun.GetID(), *wid)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Cancelled() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType := shared.EventTypeWorkflowExecutionCanceled
	details := "some details"
	encodedDetails, _ := encodeArg(getDefaultDataConverter(), details)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType,
					WorkflowExecutionCanceledEventAttributes: &shared.WorkflowExecutionCanceledEventAttributes{
						Details: encodedDetails,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions()...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Error(err)
	_, ok := err.(*CanceledError)
	s.True(ok)
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Failed() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType := shared.EventTypeWorkflowExecutionFailed
	reason := "some reason"
	details := "some details"
	dataConverter := getDefaultDataConverter()
	encodedDetails, _ := encodeArg(dataConverter, details)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType,
					WorkflowExecutionFailedEventAttributes: &shared.WorkflowExecutionFailedEventAttributes{
						Reason:  common.StringPtr(reason),
						Details: encodedDetails,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions()...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Equal(constructError(reason, encodedDetails, dataConverter), err)
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Terminated() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType := shared.EventTypeWorkflowExecutionTerminated
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType,
					WorkflowExecutionTerminatedEventAttributes: &shared.WorkflowExecutionTerminatedEventAttributes{},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions()...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Equal(newTerminatedError(), err)
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_TimedOut() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType := shared.EventTypeWorkflowExecutionTimedOut
	timeType := shared.TimeoutTypeScheduleToStart
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType,
					WorkflowExecutionTimedOutEventAttributes: &shared.WorkflowExecutionTimedOutEventAttributes{
						TimeoutType: &timeType,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions()...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Error(err)
	_, ok := err.(*TimeoutError)
	s.True(ok)
	s.Equal(timeType, err.(*TimeoutError).TimeoutType())
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_ContinueAsNew() {
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions()...).Return(createResponse, nil).Times(1)

	newRunID := "some other random run ID"
	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType1 := shared.EventTypeWorkflowExecutionContinuedAsNew
	getRequest1 := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse1 := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType1,
					WorkflowExecutionContinuedAsNewEventAttributes: &shared.WorkflowExecutionContinuedAsNewEventAttributes{
						NewExecutionRunId: common.StringPtr(newRunID),
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest1, callOptions()...).Return(getResponse1, nil).Times(1)

	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	eventType2 := shared.EventTypeWorkflowExecutionCompleted
	getRequest2 := getGetWorkflowExecutionHistoryRequest(filterType)
	getRequest2.Execution.RunId = common.StringPtr(newRunID)
	getResponse2 := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType2,
					WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
						Result: encodedResult,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest2, callOptions()...).Return(getResponse2, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.NoError(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestGetWorkflow() {
	filterType := shared.HistoryEventFilterTypeCloseEvent
	eventType := shared.EventTypeWorkflowExecutionCompleted
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{
			Events: []*shared.HistoryEvent{
				&shared.HistoryEvent{
					EventType: &eventType,
					WorkflowExecutionCompletedEventAttributes: &shared.WorkflowExecutionCompletedEventAttributes{
						Result: encodedResult,
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions()...).Return(getResponse, nil).Times(1)

	workflowID := workflowID
	runID := runID

	workflowRun := s.workflowClient.GetWorkflow(
		context.Background(),
		workflowID,
		runID,
	)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err := workflowRun.Get(context.Background(), &decodedResult)
	s.NoError(err)
	s.Equal(workflowResult, decodedResult)
}

func getGetWorkflowExecutionHistoryRequest(filterType shared.HistoryEventFilterType) *shared.GetWorkflowExecutionHistoryRequest {
	isLongPoll := true

	request := &shared.GetWorkflowExecutionHistoryRequest{
		Domain: common.StringPtr(domain),
		Execution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      getRunID(runID),
		},
		WaitForNewEvent:        common.BoolPtr(isLongPoll),
		HistoryEventFilterType: &filterType,
		SkipArchival:           common.BoolPtr(true),
	}

	return request
}

// workflow client test suite
type (
	workflowClientTestSuite struct {
		suite.Suite
		mockCtrl *gomock.Controller
		service  *workflowservicetest.MockClient
		client   Client
	}
)

func TestWorkflowClientSuite(t *testing.T) {
	suite.Run(t, new(workflowClientTestSuite))
}

func (s *workflowClientTestSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

func (s *workflowClientTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = workflowservicetest.NewMockClient(s.mockCtrl)
	s.client = NewClient(s.service, domain, &ClientOptions{
		Identity: identity,
	})
}

func (s *workflowClientTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *workflowClientTestSuite) SetupSubTest() {
	// required for s.Run in case of table tests and mocks
	s.SetupTest()
}

func (s *workflowClientTestSuite) TearDownSubTest() {
	// required for s.Run in case of table tests and mocks
	s.TearDownTest()
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflow() {
	signalName := "my signal"
	signalInput := []byte("my signal input")
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}

	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.service.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil).Times(2)

	resp, err := s.client.SignalWithStartWorkflow(context.Background(), workflowID, signalName, signalInput,
		options, workflowType)
	s.NoError(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)

	resp, err = s.client.SignalWithStartWorkflow(context.Background(), "", signalName, signalInput,
		options, workflowType)
	s.NoError(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflow_RPCError() {
	signalName := "my signal"
	signalInput := []byte("my signal input")
	options := StartWorkflowOptions{}

	// Pass a context with a deadline so error retry doesn't take forever
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := s.client.SignalWithStartWorkflow(ctx, workflowID, signalName, signalInput, options, workflowType)
	s.Equal(errors.New("missing TaskList"), err)
	s.Nil(resp)

	// Pass a context with a deadline so error retry doesn't take forever
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	options.TaskList = tasklist
	resp, err = s.client.SignalWithStartWorkflow(ctx, workflowID, signalName, signalInput, options, workflowType)
	s.Error(err)
	s.Nil(resp)

	options.ExecutionStartToCloseTimeout = timeoutInSeconds
	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.service.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil)
	resp, err = s.client.SignalWithStartWorkflow(context.Background(), workflowID, signalName, signalInput,
		options, workflowType)
	s.NoError(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow() {
	client := s.client.(*workflowClient)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}

	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil)

	resp, err := client.StartWorkflow(context.Background(), options, f1, []byte("test"))
	s.Equal(getDefaultDataConverter(), client.dataConverter)
	s.NoError(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow_WithContext() {
	s.client = NewClient(s.service, domain, &ClientOptions{ContextPropagators: []ContextPropagator{NewStringMapPropagator([]string{testHeader})}})
	client := s.client.(*workflowClient)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) error {
		value := ctx.Value(contextKey(testHeader))
		if val, ok := value.([]byte); ok {
			s.Equal("test-data", string(val))
			return nil
		}
		return fmt.Errorf("context did not propagate to workflow")
	}

	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil)

	resp, err := client.StartWorkflow(context.Background(), options, f1, []byte("test"))
	s.NoError(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow_WithDataConverter() {
	dc := newTestDataConverter()
	client := NewClient(s.service, domain, &ClientOptions{DataConverter: dc})
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r string) string {
		return "result"
	}
	input := "test" // note: not []byte as that bypasses encoding, and we are exercising the dataconverter

	correctlyEncoded, err := dc.ToData(input) // []any is spread to ToData args
	require.NoError(s.T(), err, "test data converter should not fail on simple inputs")
	defaultEncoded, err := getDefaultDataConverter().ToData(input)
	require.NoError(s.T(), err, "default data converter should not fail on simple inputs")

	// sanity check: we must be able to tell right from wrong
	require.NotEqual(s.T(), correctlyEncoded, defaultEncoded, "test data converter should encode differently or the test is not valid")

	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil).
		Do(func(_ interface{}, req *shared.StartWorkflowExecutionRequest, _ ...interface{}) {
			assert.Equal(s.T(), correctlyEncoded, req.Input, "client-encoded data should use the customized data converter")
		})

	resp, err := client.StartWorkflow(context.Background(), options, f1, input)
	s.NoError(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow_ByteBypass() {
	// default DataConverter checks for []byte args, and does not re-encode them.
	// this is NOT a general feature of DataConverter, though perhaps it should be
	// as it's a quite useful escape hatch when making incompatible type changes.
	client := NewClient(s.service, domain, nil)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}
	input := []byte("test") // intentionally bypassing dataconverter

	createResponse := &shared.StartWorkflowExecutionResponse{
		RunId: common.StringPtr(runID),
	}
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil).
		Do(func(_ interface{}, req *shared.StartWorkflowExecutionRequest, _ ...interface{}) {
			assert.Equal(s.T(), input, req.Input, "[]byte inputs should not be re-encoded by the default converter")
		})

	resp, err := client.StartWorkflow(context.Background(), options, f1, input)
	s.NoError(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow_WithMemoAndSearchAttr() {
	memo := map[string]interface{}{
		"testMemo": "memo value",
	}
	searchAttributes := map[string]interface{}{
		"testAttr": "attr value",
	}
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
		Memo:                            memo,
		SearchAttributes:                searchAttributes,
	}
	wf := func(ctx Context) string {
		return "result"
	}
	startResp := &shared.StartWorkflowExecutionResponse{}

	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(startResp, nil).
		Do(func(_ interface{}, req *shared.StartWorkflowExecutionRequest, _ ...interface{}) {
			// trailing newline is added by the dataconverter
			s.Equal([]byte("\"memo value\"\n"), req.Memo.Fields["testMemo"], "memos use the dataconverter because they are user data")
			s.Equal([]byte(`"attr value"`), req.SearchAttributes.IndexedFields["testAttr"], "search attributes must be JSON-encoded, not using dataconverter")
		})

	_, err := s.client.StartWorkflow(context.Background(), options, wf)
	s.NoError(err)
}

func (s *workflowClientTestSuite) TestStartWorkflow_RequestCreationFails() {
	client := s.client.(*workflowClient)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        "", // this causes error
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}

	_, err := client.StartWorkflow(context.Background(), options, f1, []byte("test"))
	s.ErrorContains(err, "missing TaskList")
}

func (s *workflowClientTestSuite) TestStartWorkflow_RPCError() {
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	wf := func(ctx Context) string {
		return "result"
	}
	startResp := &shared.StartWorkflowExecutionResponse{}

	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(startResp, errors.New("failed")).MinTimes(1)

	// Pass a context with a deadline so error retry doesn't take forever
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.client.StartWorkflow(ctx, options, wf)
	s.Error(err)
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflow_WithMemoAndSearchAttr() {
	memo := map[string]interface{}{
		"testMemo": "memo value",
	}
	searchAttributes := map[string]interface{}{
		"testAttr": "attr value",
	}
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
		Memo:                            memo,
		SearchAttributes:                searchAttributes,
	}
	wf := func(ctx Context) string {
		return "result"
	}
	startResp := &shared.StartWorkflowExecutionResponse{}

	s.service.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any()).Return(startResp, nil).
		Do(func(_ interface{}, req *shared.SignalWithStartWorkflowExecutionRequest, _ ...interface{}) {
			// trailing newline is added by the dataconverter
			s.Equal([]byte("\"memo value\"\n"), req.Memo.Fields["testMemo"], "memos use the dataconverter because they are user data")
			s.Equal([]byte(`"attr value"`), req.SearchAttributes.IndexedFields["testAttr"], "search attributes must be JSON-encoded")
		})

	_, err := s.client.SignalWithStartWorkflow(context.Background(), "wid", "signal", "value", options, wf)
	s.NoError(err)
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflowAsync_WithMemoAndSearchAttr() {
	memo := map[string]interface{}{
		"testMemo": "memo value",
	}
	searchAttributes := map[string]interface{}{
		"testAttr": "attr value",
	}
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
		Memo:                            memo,
		SearchAttributes:                searchAttributes,
	}
	wf := func(ctx Context) string {
		return "result"
	}

	s.service.EXPECT().SignalWithStartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).
		Do(func(_ interface{}, asyncReq *shared.SignalWithStartWorkflowExecutionAsyncRequest, _ ...interface{}) {
			req := asyncReq.Request
			// trailing newline is added by the dataconverter
			s.Equal([]byte("\"memo value\"\n"), req.Memo.Fields["testMemo"], "memos use the dataconverter because they are user data")
			s.Equal([]byte(`"attr value"`), req.SearchAttributes.IndexedFields["testAttr"], "search attributes must be JSON-encoded")
		})

	_, err := s.client.SignalWithStartWorkflowAsync(context.Background(), "wid", "signal", "value", options, wf)
	s.NoError(err)
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflow_RequestCreationFails() {
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        "", // this causes error
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	wf := func(ctx Context) string {
		return "result"
	}

	_, err := s.client.SignalWithStartWorkflow(context.Background(), "wid", "signal", "value", options, wf)
	s.ErrorContains(err, "missing TaskList")
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflowAsync_RequestCreationFails() {
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        "", // this causes error
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	wf := func(ctx Context) string {
		return "result"
	}

	_, err := s.client.SignalWithStartWorkflowAsync(context.Background(), "wid", "signal", "value", options, wf)
	s.ErrorContains(err, "missing TaskList")
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflowAsync_RPCError() {
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	wf := func(ctx Context) string {
		return "result"
	}

	s.service.EXPECT().SignalWithStartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed")).MinTimes(1)

	// Pass a context with a deadline so error retry doesn't take forever
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.client.SignalWithStartWorkflowAsync(ctx, "wid", "signal", "value", options, wf)
	s.Error(err)
}

func (s *workflowClientTestSuite) TestStartWorkflowAsync() {
	client := s.client.(*workflowClient)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}

	s.service.EXPECT().StartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	_, err := client.StartWorkflowAsync(context.Background(), options, f1, []byte("test"))
	s.Equal(getDefaultDataConverter(), client.dataConverter)
	s.NoError(err)
}

func (s *workflowClientTestSuite) TestStartWorkflowAsync_WithDataConverter() {
	dc := newTestDataConverter()
	client := NewClient(s.service, domain, &ClientOptions{DataConverter: dc})
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r string) string {
		return "result"
	}
	input := "test" // note: not []byte as that bypasses encoding, and we are exercising the dataconverter

	// this test requires that the overridden test-data-converter is used, so we need to make sure the encoded input looks correct.
	// the test data converter's output doesn't actually matter as long as it's noticeably different.
	correctlyEncoded, err := dc.ToData(input) // []any is spread to ToData args
	require.NoError(s.T(), err, "test data converter should not fail on simple inputs")
	wrongDefaultEncoding, err := getDefaultDataConverter().ToData(input)
	require.NoError(s.T(), err, "default data converter should not fail on simple inputs")

	// sanity check: we must be able to tell right from wrong
	require.NotEqual(s.T(), correctlyEncoded, wrongDefaultEncoding, "test data converter should encode differently or the test is not valid")

	s.service.EXPECT().StartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).
		Do(func(_ interface{}, asyncReq *shared.StartWorkflowExecutionAsyncRequest, _ ...interface{}) {
			assert.Equal(s.T(), correctlyEncoded, asyncReq.Request.Input, "client-encoded data should use the customized data converter")
		})

	_, err = client.StartWorkflowAsync(context.Background(), options, f1, input)
	s.NoError(err)
}

func (s *workflowClientTestSuite) TestStartWorkflowAsync_ByteBypass() {
	// default DataConverter checks for []byte args, and does not re-encode them.
	// this is NOT a general feature of DataConverter, though perhaps it should be
	// as it's a quite useful escape hatch when making incompatible type changes.
	client := NewClient(s.service, domain, nil)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}
	input := []byte("test") // intentionally bypassing dataconverter

	s.service.EXPECT().StartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).
		Do(func(_ interface{}, req *shared.StartWorkflowExecutionAsyncRequest, _ ...interface{}) {
			assert.Equal(s.T(), input, req.Request.Input, "[]byte inputs should not be re-encoded by the default converter")
		})

	_, err := client.StartWorkflowAsync(context.Background(), options, f1, input)
	s.NoError(err)
}

func (s *workflowClientTestSuite) TestStartWorkflowAsync_WithMemoAndSearchAttr() {
	memo := map[string]interface{}{
		"testMemo": "memo value",
	}
	searchAttributes := map[string]interface{}{
		"testAttr": "attr value",
	}
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
		Memo:                            memo,
		SearchAttributes:                searchAttributes,
	}
	wf := func(ctx Context) string {
		return "result"
	}

	s.service.EXPECT().StartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).
		Do(func(_ interface{}, asyncReq *shared.StartWorkflowExecutionAsyncRequest, _ ...interface{}) {
			req := asyncReq.Request
			// trailing newline is added by the dataconverter
			s.Equal([]byte("\"memo value\"\n"), req.Memo.Fields["testMemo"], "memos use the dataconverter because they are user data")
			s.Equal([]byte(`"attr value"`), req.SearchAttributes.IndexedFields["testAttr"], "search attributes must be JSON-encoded")
		})

	_, err := s.client.StartWorkflowAsync(context.Background(), options, wf)
	s.NoError(err)
}

func (s *workflowClientTestSuite) TestStartWorkflowAsync_RequestCreationFails() {
	client := s.client.(*workflowClient)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        "", // this causes error
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}
	_, err := client.StartWorkflowAsync(context.Background(), options, f1, []byte("test"))
	s.ErrorContains(err, "missing TaskList")
}

func (s *workflowClientTestSuite) TestStartWorkflowAsync_RPCError() {
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	wf := func(ctx Context) string {
		return "result"
	}

	s.service.EXPECT().StartWorkflowExecutionAsync(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed")).MinTimes(1)

	// Pass a context with a deadline so error retry doesn't take forever
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.client.StartWorkflowAsync(ctx, options, wf)
	s.Error(err)
}

func (s *workflowClientTestSuite) TestGetWorkflowMemo() {
	var input1 map[string]interface{}
	result1, err := getWorkflowMemo(input1, nil)
	s.NoError(err)
	s.Nil(result1)

	input1 = make(map[string]interface{})
	result2, err := getWorkflowMemo(input1, nil)
	s.NoError(err)
	s.NotNil(result2)
	s.Equal(0, len(result2.Fields))

	input1["t1"] = "v1"
	result3, err := getWorkflowMemo(input1, nil)
	s.NoError(err)
	s.NotNil(result3)
	s.Equal(1, len(result3.Fields))
	var resultString string
	decodeArg(nil, result3.Fields["t1"], &resultString)
	s.Equal("v1", resultString)

	input1["non-serializable"] = make(chan int)
	_, err = getWorkflowMemo(input1, nil)
	s.Error(err)
}

func (s *workflowClientTestSuite) TestSerializeSearchAttributes() {
	var input1 map[string]interface{}
	result1, err := serializeSearchAttributes(input1)
	s.NoError(err)
	s.Nil(result1)

	input1 = make(map[string]interface{})
	result2, err := serializeSearchAttributes(input1)
	s.NoError(err)
	s.NotNil(result2)
	s.Equal(0, len(result2.IndexedFields))

	input1["t1"] = "v1"
	result3, err := serializeSearchAttributes(input1)
	s.NoError(err)
	s.NotNil(result3)
	s.Equal(1, len(result3.IndexedFields))
	var resultString string
	decodeArg(nil, result3.IndexedFields["t1"], &resultString)
	s.Equal("v1", resultString)

	input1["non-serializable"] = make(chan int)
	_, err = serializeSearchAttributes(input1)
	s.Error(err)
}

func (s *workflowClientTestSuite) TestListWorkflow() {
	request := &shared.ListWorkflowExecutionsRequest{}
	response := &shared.ListWorkflowExecutionsResponse{}
	s.service.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *shared.ListWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	resp, err := s.client.ListWorkflow(context.Background(), request)
	s.NoError(err)
	s.Equal(response, resp)

	responseErr := &shared.BadRequestError{}
	request.Domain = common.StringPtr("another")
	s.service.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *shared.ListWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	_, err = s.client.ListWorkflow(context.Background(), request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestListArchivedWorkflow() {
	request := &shared.ListArchivedWorkflowExecutionsRequest{}
	response := &shared.ListArchivedWorkflowExecutionsResponse{}
	s.service.EXPECT().ListArchivedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *shared.ListArchivedWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := s.client.ListArchivedWorkflow(ctxWithTimeout, request)
	s.NoError(err)
	s.Equal(response, resp)

	responseErr := &shared.BadRequestError{}
	request.Domain = common.StringPtr("another")
	s.service.EXPECT().ListArchivedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *shared.ListArchivedWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	resp, err = s.client.ListArchivedWorkflow(ctxWithTimeout, request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestScanWorkflow() {
	request := &shared.ListWorkflowExecutionsRequest{}
	response := &shared.ListWorkflowExecutionsResponse{}
	s.service.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *shared.ListWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	resp, err := s.client.ScanWorkflow(context.Background(), request)
	s.NoError(err)
	s.Equal(response, resp)

	responseErr := &shared.BadRequestError{}
	request.Domain = common.StringPtr("another")
	s.service.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *shared.ListWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	resp, err = s.client.ScanWorkflow(context.Background(), request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestCountWorkflow() {
	request := &shared.CountWorkflowExecutionsRequest{}
	response := &shared.CountWorkflowExecutionsResponse{}
	s.service.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *shared.CountWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	resp, err := s.client.CountWorkflow(context.Background(), request)
	s.NoError(err)
	s.Equal(response, resp)

	responseErr := &shared.BadRequestError{}
	request.Domain = common.StringPtr("another")
	s.service.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *shared.CountWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	resp, err = s.client.CountWorkflow(context.Background(), request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestGetSearchAttributes() {
	response := &shared.GetSearchAttributesResponse{}
	s.service.EXPECT().GetSearchAttributes(gomock.Any(), gomock.Any()).Return(response, nil)
	resp, err := s.client.GetSearchAttributes(context.Background())
	s.NoError(err)
	s.Equal(response, resp)

	responseErr := &shared.BadRequestError{}
	s.service.EXPECT().GetSearchAttributes(gomock.Any(), gomock.Any()).Return(nil, responseErr)
	resp, err = s.client.GetSearchAttributes(context.Background())
	s.Equal(responseErr, err)
}

func serializeEvents(events []*shared.HistoryEvent) *shared.DataBlob {

	blob, _ := serializer.SerializeBatchEvents(events, shared.EncodingTypeThriftRW)

	return &shared.DataBlob{
		EncodingType: shared.EncodingTypeThriftRW.Ptr(),
		Data:         blob.Data,
	}
}

func (s *workflowClientTestSuite) TestCancelWorkflow() {
	s.service.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), newPartialCancelRequestMatcher(common.StringPtr("testWf"), common.StringPtr("test reason")), gomock.All(gomock.Any())).Return(nil)

	err := s.client.CancelWorkflow(context.Background(), "testWf", "testRun", WithCancelReason("test reason"))

	s.NoError(err)
}

func (s *workflowClientTestSuite) TestCancelWorkflowBackwardsCompatible() {
	s.service.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), newPartialCancelRequestMatcher(common.StringPtr("testWf"), nil), gomock.All(gomock.Any())).Return(nil)

	err := s.client.CancelWorkflow(context.Background(), "testWf", "testRun")

	s.NoError(err)
}

type PartialCancelRequestMatcher struct {
	wfID  *string
	cause *string
}

func newPartialCancelRequestMatcher(wfID *string, cause *string) gomock.Matcher {
	return &PartialCancelRequestMatcher{
		wfID:  wfID,
		cause: cause,
	}
}

func (m *PartialCancelRequestMatcher) Matches(a interface{}) bool {
	aEx, ok := a.(*shared.RequestCancelWorkflowExecutionRequest)
	if !ok {
		return false
	}

	return (aEx.Cause == m.cause || *aEx.Cause == *m.cause) && *aEx.WorkflowExecution.WorkflowId == *m.wfID
}

func (m *PartialCancelRequestMatcher) String() string {
	return "partial cancellation request matcher matches cause and wfId fields"
}

func (s *workflowClientTestSuite) TestTerminateWorkflow() {
	expectedRequest := &shared.TerminateWorkflowExecutionRequest{
		Domain: common.StringPtr(domain),
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		Reason:   common.StringPtr("test reason"),
		Details:  []byte("test details"),
		Identity: common.StringPtr(identity),
	}

	// We consider unknown errors retryable, so we expect the call to be retried
	s.service.EXPECT().TerminateWorkflowExecution(gomock.Any(), expectedRequest, gomock.All(gomock.Any())).Return(assert.AnError)
	s.service.EXPECT().TerminateWorkflowExecution(gomock.Any(), expectedRequest, gomock.All(gomock.Any())).Return(nil)

	err := s.client.TerminateWorkflow(context.Background(), workflowID, runID, "test reason", []byte("test details"))

	s.NoError(err)
}

func (s *workflowClientTestSuite) TestDescribeTaskList() {
	testcases := []struct {
		name     string
		rpcError error
		response *shared.DescribeTaskListResponse
	}{
		{
			name:     "success",
			rpcError: nil,
			response: &shared.DescribeTaskListResponse{},
		},
		{
			name:     "failure",
			rpcError: &shared.AccessDeniedError{},
			response: nil,
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			expectedRequest := &shared.DescribeTaskListRequest{
				Domain:       common.StringPtr(domain),
				TaskList:     &shared.TaskList{Name: common.StringPtr(tasklist)},
				TaskListType: shared.TaskListTypeActivity.Ptr(),
			}
			s.service.EXPECT().
				DescribeTaskList(gomock.Any(), expectedRequest, gomock.Any()).
				Return(tt.response, tt.rpcError)

			r, err := s.client.DescribeTaskList(context.Background(), tasklist, shared.TaskListTypeActivity)
			s.Equal(tt.rpcError, err, "error should be returned as-is")
			s.Equal(tt.response, r)
		})
	}
}

func (s *workflowClientTestSuite) TestRefreshWorkflowTasks() {
	testcases := []struct {
		name     string
		rpcError error
	}{
		{
			name:     "success",
			rpcError: nil,
		},
		{
			name:     "failure",
			rpcError: &shared.AccessDeniedError{},
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			expectedRequest := &shared.RefreshWorkflowTasksRequest{
				Domain: common.StringPtr(domain),
				Execution: &shared.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
					RunId:      common.StringPtr(runID),
				},
			}
			s.service.EXPECT().
				RefreshWorkflowTasks(gomock.Any(), expectedRequest, gomock.Any()).
				Return(tt.rpcError)

			err := s.client.RefreshWorkflowTasks(context.Background(), workflowID, runID)
			s.Equal(tt.rpcError, err, "error should be returned as-is")
		})
	}
}

func (s *workflowClientTestSuite) TestListOpenWorkflow() {
	testcases := []struct {
		name    string
		request *shared.ListOpenWorkflowExecutionsRequest
		// both error and response should be returned as-is
		err      error
		response *shared.ListOpenWorkflowExecutionsResponse
	}{
		{
			name: "success with domain-name",
			request: &shared.ListOpenWorkflowExecutionsRequest{
				Domain: common.StringPtr(domain),
			},
			err:      nil,
			response: &shared.ListOpenWorkflowExecutionsResponse{},
		},
		{
			// domain name should be taken from workflow client
			name:     "success without domain-name",
			request:  &shared.ListOpenWorkflowExecutionsRequest{},
			err:      nil,
			response: &shared.ListOpenWorkflowExecutionsResponse{},
		},
		{
			name: "failed RPC",
			request: &shared.ListOpenWorkflowExecutionsRequest{
				Domain: common.StringPtr(domain),
			},
			err:      &shared.AccessDeniedError{},
			response: nil,
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			s.service.EXPECT().
				ListOpenWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *shared.ListOpenWorkflowExecutionsRequest, _ ...yarpc.CallOption) {
					s.Equal(domain, *req.Domain)
				}).
				Return(tt.response, tt.err)

			resp, err := s.client.ListOpenWorkflow(context.Background(), tt.request)
			s.Equal(tt.err, err)
			s.Equal(tt.response, resp)
		})
	}
}

func (s *workflowClientTestSuite) TestListClosedWorkflow() {
	testcases := []struct {
		name    string
		request *shared.ListClosedWorkflowExecutionsRequest
		// both error and response should be returned as-is
		err      error
		response *shared.ListClosedWorkflowExecutionsResponse
	}{
		{
			name: "success with domain-name",
			request: &shared.ListClosedWorkflowExecutionsRequest{
				Domain: common.StringPtr(domain),
			},
			err:      nil,
			response: &shared.ListClosedWorkflowExecutionsResponse{},
		},
		{
			// domain name should be taken from workflow client
			name:     "success without domain-name",
			request:  &shared.ListClosedWorkflowExecutionsRequest{},
			err:      nil,
			response: &shared.ListClosedWorkflowExecutionsResponse{},
		},
		{
			name: "failed RPC",
			request: &shared.ListClosedWorkflowExecutionsRequest{
				Domain: common.StringPtr(domain),
			},
			err:      &shared.AccessDeniedError{},
			response: nil,
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			s.service.EXPECT().
				ListClosedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *shared.ListClosedWorkflowExecutionsRequest, _ ...yarpc.CallOption) {
					s.Equal(domain, *req.Domain)
				}).
				Return(tt.response, tt.err)

			resp, err := s.client.ListClosedWorkflow(context.Background(), tt.request)
			s.Equal(tt.err, err)
			s.Equal(tt.response, resp)
		})
	}
}

func (s *workflowClientTestSuite) TestResetWorkflow() {
	testcases := []struct {
		name    string
		request *shared.ResetWorkflowExecutionRequest
		// both error and response should be returned as-is
		err      error
		response *shared.ResetWorkflowExecutionResponse
	}{
		{
			name: "success with domain-name",
			request: &shared.ResetWorkflowExecutionRequest{
				Domain: common.StringPtr(domain),
			},
			err:      nil,
			response: &shared.ResetWorkflowExecutionResponse{},
		},
		{
			// domain name should be taken from workflow client
			name:     "success without domain-name",
			request:  &shared.ResetWorkflowExecutionRequest{},
			err:      nil,
			response: &shared.ResetWorkflowExecutionResponse{},
		},
		{
			name: "failed RPC",
			request: &shared.ResetWorkflowExecutionRequest{
				Domain: common.StringPtr(domain),
			},
			err:      &shared.AccessDeniedError{},
			response: nil,
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			s.service.EXPECT().
				ResetWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, req *shared.ResetWorkflowExecutionRequest, _ ...yarpc.CallOption) {
					s.Equal(domain, *req.Domain)
				}).
				Return(tt.response, tt.err)

			resp, err := s.client.ResetWorkflow(context.Background(), tt.request)
			s.Equal(tt.err, err)
			s.Equal(tt.response, resp)
		})
	}
}

func (s *workflowClientTestSuite) TestDescribeWorkflowExecution() {
	testcases := []struct {
		name     string
		rpcError error
		response *shared.DescribeWorkflowExecutionResponse
	}{
		{
			name:     "success",
			rpcError: nil,
			response: &shared.DescribeWorkflowExecutionResponse{},
		},
		{
			name:     "failure",
			rpcError: &shared.AccessDeniedError{},
			response: nil,
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			expectedRequest := &shared.DescribeWorkflowExecutionRequest{
				Domain: common.StringPtr(domain),
				Execution: &shared.WorkflowExecution{
					WorkflowId: common.StringPtr(workflowID),
					RunId:      common.StringPtr(runID),
				},
			}
			s.service.EXPECT().
				DescribeWorkflowExecution(gomock.Any(), expectedRequest, gomock.Any()).
				Return(tt.response, tt.rpcError)

			r, err := s.client.DescribeWorkflowExecution(context.Background(), workflowID, runID)
			s.Equal(tt.rpcError, err, "error should be returned as-is")
			s.Equal(tt.response, r)
		})
	}
}

func (s *workflowClientTestSuite) TestCompleteActivity() {
	testcases := []struct {
		name           string
		taskToken      []byte
		activityResult interface{}
		activityError  error
		mockRPC        func()
		checkError     func(err error)
	}{
		{
			name:           "ActivityTaskCompletedRequest success",
			taskToken:      []byte(taskToken),
			activityResult: "result string",
			activityError:  nil,

			mockRPC: func() {
				request := &shared.RespondActivityTaskCompletedRequest{
					TaskToken: []byte(taskToken),
					Result:    []byte("\"result string\"\n"), // JSON encoded activityResult
					Identity:  common.StringPtr(identity),
				}

				s.service.EXPECT().
					RespondActivityTaskCompleted(gomock.Any(), request, gomock.Any()).
					Return(nil)
			},
			checkError: func(err error) {
				s.Nil(err)
			},
		},
		{
			name:           "ActivityTaskCompletedRequest failure",
			taskToken:      []byte(taskToken),
			activityResult: nil,
			activityError:  nil,
			mockRPC: func() {
				s.service.EXPECT().
					RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&shared.AccessDeniedError{})
			},
			checkError: func(err error) {
				s.Equal(&shared.AccessDeniedError{}, err)
			},
		},
		{
			name:      "missing taskToken",
			taskToken: nil,

			mockRPC: func() { /* we don't expect any RPC here */ },
			checkError: func(err error) {
				s.ErrorContains(err, "invalid task token provided")
			},
		},
		{
			name:           "fail to encode activity result",
			taskToken:      []byte(taskToken),
			activityResult: make(chan int), // can't be represented as JSON

			mockRPC: func() { /* we don't expect any RPC here */ },
			checkError: func(err error) {
				s.ErrorContains(err, "unable to encode argument")
			},
		},
		{
			name:          "ActivityResultPending success",
			taskToken:     []byte(taskToken),
			activityError: ErrActivityResultPending,

			mockRPC: func() { /* pending activity result is not reported to server */ },
			checkError: func(err error) {
				s.Nil(err)
			},
		},

		{
			name:          "ActivityTaskCanceled success",
			taskToken:     []byte(taskToken),
			activityError: NewCanceledError("some details"),

			mockRPC: func() {
				request := &shared.RespondActivityTaskCanceledRequest{
					TaskToken: []byte(taskToken),
					Details:   []byte("\"some details\"\n"),
					Identity:  common.StringPtr(identity),
				}
				s.service.EXPECT().
					RespondActivityTaskCanceled(gomock.Any(), request, gomock.Any()).
					Return(nil)
			},
			checkError: func(err error) {
				s.Nil(err)
			},
		},
		{
			name:          "ActivityTaskCanceled failure",
			taskToken:     []byte(taskToken),
			activityError: context.Canceled, // canceled context is another reason for canceled activity

			mockRPC: func() {
				s.service.EXPECT().
					RespondActivityTaskCanceled(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&shared.AccessDeniedError{})
			},
			checkError: func(err error) {
				s.Equal(&shared.AccessDeniedError{}, err)
			},
		},
		{
			name:          "ActivityTaskFailed success",
			taskToken:     []byte(taskToken),
			activityError: NewCustomError("some reason", "some details"),

			mockRPC: func() {
				request := &shared.RespondActivityTaskFailedRequest{
					TaskToken: []byte(taskToken),
					Reason:    common.StringPtr("some reason"),
					Details:   []byte("\"some details\"\n"),
					Identity:  common.StringPtr(identity),
				}
				s.service.EXPECT().
					RespondActivityTaskFailed(gomock.Any(), request, gomock.Any()).
					Return(nil)
			},
			checkError: func(err error) {
				s.Nil(err)
			},
		},
		{
			name:          "ActivityTaskFailed failure",
			taskToken:     []byte(taskToken),
			activityError: NewCustomError("some reason"),

			mockRPC: func() {
				s.service.EXPECT().
					RespondActivityTaskFailed(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&shared.AccessDeniedError{})
			},
			checkError: func(err error) {
				s.Equal(&shared.AccessDeniedError{}, err)
			},
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			tt.mockRPC()
			err := s.client.CompleteActivity(context.Background(), tt.taskToken, tt.activityResult, tt.activityError)
			tt.checkError(err)
		})
	}
}

func (s *workflowClientTestSuite) TestCompleteActivityByID() {
	testcases := []struct {
		name           string
		activityID     string
		activityResult interface{}
		activityError  error
		mockRPC        func()
		checkError     func(err error)
	}{
		{
			name:           "ActivityTaskCompletedByID success",
			activityID:     activityID,
			activityResult: "result string",
			activityError:  nil,

			mockRPC: func() {
				request := &shared.RespondActivityTaskCompletedByIDRequest{
					Domain:     common.StringPtr(domain),
					WorkflowID: common.StringPtr(workflowID),
					RunID:      common.StringPtr(runID),
					ActivityID: common.StringPtr(activityID),
					Result:     []byte("\"result string\"\n"), // JSON encoded activityResult
					Identity:   common.StringPtr(identity),
				}

				s.service.EXPECT().
					RespondActivityTaskCompletedByID(gomock.Any(), request, gomock.Any()).
					Return(nil)
			},
			checkError: func(err error) {
				s.Nil(err)
			},
		},
		{
			name:           "ActivityTaskCompletedByID failure",
			activityID:     activityID,
			activityResult: nil,
			activityError:  nil,

			mockRPC: func() {
				s.service.EXPECT().
					RespondActivityTaskCompletedByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&shared.AccessDeniedError{})
			},
			checkError: func(err error) {
				s.Equal(&shared.AccessDeniedError{}, err)
			},
		},
		{
			name:       "missing activityID",
			activityID: "",

			mockRPC: func() { /* we don't expect any RPC here */ },
			checkError: func(err error) {
				s.ErrorContains(err, "empty activity")
			},
		},
		{
			name:           "fail to encode activity result",
			activityID:     activityID,
			activityResult: make(chan int), // can't be represented as JSON

			mockRPC: func() { /* we don't expect any RPC here */ },
			checkError: func(err error) {
				s.ErrorContains(err, "unable to encode argument")
			},
		},
		{
			name:          "ActivityResultPending success",
			activityID:    activityID,
			activityError: ErrActivityResultPending,

			mockRPC: func() { /* pending activity result is not reported to server */ },
			checkError: func(err error) {
				s.Nil(err)
			},
		},

		{
			name:          "ActivityTaskCanceledByID success",
			activityID:    activityID,
			activityError: NewCanceledError("some details"),

			mockRPC: func() {
				request := &shared.RespondActivityTaskCanceledByIDRequest{
					Domain:     common.StringPtr(domain),
					WorkflowID: common.StringPtr(workflowID),
					RunID:      common.StringPtr(runID),
					ActivityID: common.StringPtr(activityID),
					Details:    []byte("\"some details\"\n"),
					Identity:   common.StringPtr(identity),
				}
				s.service.EXPECT().
					RespondActivityTaskCanceledByID(gomock.Any(), request, gomock.Any()).
					Return(nil)
			},
			checkError: func(err error) {
				s.Nil(err)
			},
		},
		{
			name:          "ActivityTaskCanceledByID failure",
			activityID:    activityID,
			activityError: context.Canceled, // canceled context is another reason for canceled activity

			mockRPC: func() {
				s.service.EXPECT().
					RespondActivityTaskCanceledByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&shared.AccessDeniedError{})
			},
			checkError: func(err error) {
				s.Equal(&shared.AccessDeniedError{}, err)
			},
		},
		{
			name:          "ActivityTaskFailedByID success",
			activityID:    activityID,
			activityError: NewCustomError("some reason", "some details"),

			mockRPC: func() {
				request := &shared.RespondActivityTaskFailedByIDRequest{
					Domain:     common.StringPtr(domain),
					WorkflowID: common.StringPtr(workflowID),
					RunID:      common.StringPtr(runID),
					ActivityID: common.StringPtr(activityID),
					Reason:     common.StringPtr("some reason"),
					Details:    []byte("\"some details\"\n"),
					Identity:   common.StringPtr(identity),
				}
				s.service.EXPECT().
					RespondActivityTaskFailedByID(gomock.Any(), request, gomock.Any()).
					Return(nil)
			},
			checkError: func(err error) {
				s.Nil(err)
			},
		},
		{
			name:          "ActivityTaskFailedByID failure",
			activityID:    activityID,
			activityError: NewCustomError("some reason"),

			mockRPC: func() {
				s.service.EXPECT().
					RespondActivityTaskFailedByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&shared.AccessDeniedError{})
			},
			checkError: func(err error) {
				s.Equal(&shared.AccessDeniedError{}, err)
			},
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			tt.mockRPC()
			err := s.client.CompleteActivityByID(
				context.Background(),
				domain,
				workflowID,
				runID,
				tt.activityID,
				tt.activityResult, tt.activityError,
			)
			tt.checkError(err)
		})
	}
}

func (s *workflowClientTestSuite) TestQueryWorkflowWithOptions() {
	testcases := []struct {
		name              string
		queryArgs         []interface{}
		requestValidator  func(req *shared.QueryWorkflowRequest) // nil if RPC is not expected
		rpcResponse       *shared.QueryWorkflowResponse
		rpcError          error
		responseValidator func(resp *QueryWorkflowWithOptionsResponse, err error)
	}{
		{
			name:      "success without arguments",
			queryArgs: nil,
			requestValidator: func(req *shared.QueryWorkflowRequest) {
				// do common validation for common fields as well
				s.Equal(domain, req.GetDomain())
				s.Equal(workflowID, req.GetExecution().GetWorkflowId())
				s.Equal(runID, req.GetExecution().GetRunId())
				s.Equal(queryType, req.GetQuery().GetQueryType())

				s.Empty(req.GetQuery().GetQueryArgs(), "no input queryArgs provided")
			},

			rpcResponse: &shared.QueryWorkflowResponse{QueryResult: []byte("\"result\"")},
			rpcError:    nil,
			responseValidator: func(resp *QueryWorkflowWithOptionsResponse, err error) {
				s.Require().Nil(err)
				s.Nil(resp.QueryRejected)

				var res string
				s.NoError(resp.QueryResult.Get(&res))
				s.Equal("result", res)
			},
		},
		{
			name:      "success with arguments",
			queryArgs: []interface{}{"arg1", "arg2"},
			requestValidator: func(req *shared.QueryWorkflowRequest) {
				s.Equal("\"arg1\"\n\"arg2\"\n", string(req.GetQuery().GetQueryArgs()))
			},

			rpcResponse: &shared.QueryWorkflowResponse{QueryResult: []byte("\"result\"")},
			rpcError:    nil,
			responseValidator: func(resp *QueryWorkflowWithOptionsResponse, err error) {
				s.Require().Nil(err)
				s.Nil(resp.QueryRejected)

				var res string
				s.NoError(resp.QueryResult.Get(&res))
				s.Equal("result", res)
			},
		},
		{
			name:             "failed to encode arguments",
			queryArgs:        []interface{}{make(chan int)}, // you can't marshal this object to JSON
			requestValidator: nil,

			responseValidator: func(resp *QueryWorkflowWithOptionsResponse, err error) {
				s.ErrorContains(err, "unable to encode")
				s.Nil(resp)
			},
		},
		{
			name:             "RPC fails",
			queryArgs:        nil,
			requestValidator: func(req *shared.QueryWorkflowRequest) {},

			rpcResponse: nil,
			rpcError:    &shared.AccessDeniedError{},
			responseValidator: func(resp *QueryWorkflowWithOptionsResponse, err error) {
				s.Equal(&shared.AccessDeniedError{}, err)
				s.Nil(resp)
			},
		},
		{
			name:             "query rejected",
			queryArgs:        nil,
			requestValidator: func(req *shared.QueryWorkflowRequest) {},

			rpcResponse: &shared.QueryWorkflowResponse{
				QueryRejected: &shared.QueryRejected{
					CloseStatus: shared.WorkflowExecutionCloseStatusTerminated.Ptr(),
				},
			},
			rpcError: nil,
			responseValidator: func(resp *QueryWorkflowWithOptionsResponse, err error) {
				s.Require().Nil(err)
				s.Nil(resp.QueryResult, "should be nil when query rejected")

				s.Require().NotNil(resp.QueryRejected)
				s.Equal(shared.WorkflowExecutionCloseStatusTerminated, resp.QueryRejected.GetCloseStatus())
			},
		},
	}

	for _, tt := range testcases {
		s.Run(tt.name, func() {
			if tt.requestValidator != nil {
				s.service.EXPECT().
					QueryWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, req *shared.QueryWorkflowRequest, _ ...yarpc.CallOption) {
						tt.requestValidator(req)
					}).
					Return(tt.rpcResponse, tt.rpcError)
			}

			request := &QueryWorkflowWithOptionsRequest{
				WorkflowID: workflowID,
				QueryType:  queryType,
				RunID:      runID,
				Args:       tt.queryArgs,
			}
			resp, err := s.client.QueryWorkflowWithOptions(context.Background(), request)
			tt.responseValidator(resp, err)
		})
	}
}

func (s *workflowClientTestSuite) TestGetWorkflowHistory() {
	// Page 1 of 2
	//// Events
	events, err := serializer.SerializeBatchEvents(
		[]*shared.HistoryEvent{
			{EventId: common.Int64Ptr(1)},
			{EventId: common.Int64Ptr(2)},
		},
		shared.EncodingTypeThriftRW,
	)
	s.NoError(err)

	//// Mock
	s.service.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&shared.GetWorkflowExecutionHistoryResponse{
			History:       nil,
			RawHistory:    []*shared.DataBlob{events},
			NextPageToken: []byte("token"),
			Archived:      nil,
		}, nil)

	// Page 2 of 2
	//// Events
	events, err = serializer.SerializeBatchEvents(
		[]*shared.HistoryEvent{
			{EventId: common.Int64Ptr(3)},
			{EventId: common.Int64Ptr(4)},
		},
		shared.EncodingTypeThriftRW,
	)
	s.NoError(err)

	//// Mock
	s.service.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&shared.GetWorkflowExecutionHistoryResponse{
			History:       nil,
			RawHistory:    []*shared.DataBlob{events},
			NextPageToken: nil,
			Archived:      nil,
		}, nil)

	// Act
	iterator := s.client.GetWorkflowHistory(
		context.Background(),
		workflowID,
		runID,
		true,
		shared.HistoryEventFilterTypeAllEvent,
	)

	s.NotNil(iterator)

	// Check that the iterator returns the correct events
	for i := 1; iterator.HasNext(); i++ {
		event, err := iterator.Next()
		s.NoError(err)
		s.Equal(int64(i), event.GetEventId())
	}
}

func TestGetWorkflowStartRequest(t *testing.T) {
	tests := []struct {
		name         string
		options      StartWorkflowOptions
		workflowFunc interface{}
		args         []interface{}
		wantRequest  *shared.StartWorkflowExecutionRequest
		wantErr      string
	}{
		{
			name: "success",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      0 * time.Second,
				JitterStart:                     0 * time.Second,
			},
			workflowFunc: func(ctx Context) {},
			wantRequest: &shared.StartWorkflowExecutionRequest{
				Domain:     common.StringPtr(domain),
				WorkflowId: common.StringPtr(workflowID),
				WorkflowType: &shared.WorkflowType{
					Name: common.StringPtr("go.uber.org/cadence/internal.TestGetWorkflowStartRequest.func1"),
				},
				TaskList: &shared.TaskList{
					Name: common.StringPtr(tasklist),
				},
				ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(10),
				TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(5),
				DelayStartSeconds:                   common.Int32Ptr(0),
				JitterStartSeconds:                  common.Int32Ptr(0),
				FirstRunAtTimestamp:                 common.Int64Ptr(0),
				CronSchedule:                        common.StringPtr(""),
				Header:                              &shared.Header{Fields: map[string][]byte{}},
				WorkflowIdReusePolicy:               shared.WorkflowIdReusePolicyAllowDuplicateFailedOnly.Ptr(),
			},
		},
		{
			name: "missing TaskList",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        "", // this causes error
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      0 * time.Second,
				JitterStart:                     0 * time.Second,
			},
			workflowFunc: func(ctx Context) {},
			wantErr:      "missing TaskList",
		},
		{
			name: "invalid ExecutionStartToCloseTimeout",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    0 * time.Second, // this causes error
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      0 * time.Second,
				JitterStart:                     0 * time.Second,
			},
			workflowFunc: func(ctx Context) {},
			wantErr:      "missing or invalid ExecutionStartToCloseTimeout",
		},
		{
			name: "negative DecisionTaskStartToCloseTimeout",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: -1 * time.Second, // this causes error
				DelayStart:                      0 * time.Second,
				JitterStart:                     0 * time.Second,
			},
			workflowFunc: func(ctx Context) {},
			wantErr:      "negative DecisionTaskStartToCloseTimeout provided",
		},
		{
			name: "negative DelayStart",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      -1 * time.Second, // this causes error
				JitterStart:                     0 * time.Second,
			},
			workflowFunc: func(ctx Context) {},
			wantErr:      "Invalid DelayStart option",
		},
		{
			name: "negative JitterStart",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      0 * time.Second,
				JitterStart:                     -1 * time.Second, // this causes error
			},
			workflowFunc: func(ctx Context) {},
			wantErr:      "Invalid JitterStart option",
		},
		{
			name: "negative firstRunAtTimestamp",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      0 * time.Second,
				FirstRunAt:                      time.Unix(-12, 0), // this causes error
			},
			workflowFunc: func(ctx Context) {},
			wantErr:      "Invalid FirstRunAt option",
		},
		{
			name: "invalid workflow func",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      0 * time.Second,
				JitterStart:                     0 * time.Second,
			},
			workflowFunc: func(ctx Context, a, b int) {}, // this causes error because args not provided
			args:         []interface{}{},
			wantErr:      "expected 2 args for function",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			service := workflowservicetest.NewMockClient(mockCtrl)
			wc, ok := NewClient(service, domain, &ClientOptions{
				Identity: "test-identity",
			}).(*workflowClient)

			if !ok {
				t.Fatalf("expected NewClient to return a *workflowClient, but got %T", wc)
			}

			gotReq, err := wc.getWorkflowStartRequest(context.Background(), "", tc.options, tc.workflowFunc, tc.args...)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)

			// set the randomized fields in the expected request before comparison
			tc.wantRequest.Identity = &wc.identity
			tc.wantRequest.RequestId = gotReq.RequestId

			assert.Equal(t, tc.wantRequest, gotReq)
		})
	}
}

func TestGetSignalWithStartRequest(t *testing.T) {
	tests := []struct {
		name         string
		workflowID   string
		signalName   string
		signalArg    interface{}
		options      StartWorkflowOptions
		workflowFunc interface{}
		args         []interface{}
		wantRequest  *shared.SignalWithStartWorkflowExecutionRequest
		wantErr      string
	}{
		{
			name: "first run at negative",
			options: StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    10 * time.Second,
				DecisionTaskStartToCloseTimeout: 5 * time.Second,
				DelayStart:                      0 * time.Second,
				JitterStart:                     0 * time.Second,
				FirstRunAt:                      time.Unix(-12, 0),
			},
			workflowFunc: func(ctx Context) {},
			wantErr:      "Invalid FirstRunAt option",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			service := workflowservicetest.NewMockClient(mockCtrl)
			wc, ok := NewClient(service, domain, &ClientOptions{
				Identity: "test-identity",
			}).(*workflowClient)

			if !ok {
				t.Fatalf("expected NewClient to return a *workflowClient, but got %T", wc)
			}

			gotReq, err := wc.getSignalWithStartRequest(context.Background(), "", tc.workflowID, tc.signalName, tc.signalArg, tc.options, tc.workflowFunc, tc.args...)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			assert.NoError(t, err)

			// set the randomized fields in the expected request before comparison
			tc.wantRequest.Identity = &wc.identity
			tc.wantRequest.RequestId = gotReq.RequestId

			assert.Equal(t, tc.wantRequest, gotReq)
		})
	}
}
