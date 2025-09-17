// Copyright (c) 2017-2020 Uber Technologies Inc.
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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/serializer"
)

const (
	replayDomainName             = "ReplayDomain"
	replayTaskListName           = "ReplayTaskList"
	replayWorkflowID             = "ReplayId"
	replayWorkerIdentity         = "replayID"
	replayPreviousStartedEventID = math.MaxInt64
	replayTaskToken              = "ReplayTaskToken"
)

var (
	errReplayEmptyHistory          = errors.New("empty events")
	errReplayHistoryTooShort       = errors.New("at least 3 events expected in the history")
	errReplayInvalidFirstEvent     = errors.New("first event is not WorkflowExecutionStarted")
	errReplayCorruptedStartedEvent = errors.New("corrupted WorkflowExecutionStarted")
)

// WorkflowReplayer is used to replay workflow code from an event history
type WorkflowReplayer struct {
	registry *registry
	options  ReplayOptions
}

// ReplayOptions is used to configure the replay decision task worker.
type ReplayOptions struct {
	// Optional: Sets DataConverter to customize serialization/deserialization of arguments in Cadence
	// default: defaultDataConverter, an combination of thriftEncoder and jsonEncoder
	DataConverter DataConverter

	// Optional: Specifies factories used to instantiate workflow interceptor chain
	// The chain is instantiated per each replay of a workflow execution
	WorkflowInterceptorChainFactories []WorkflowInterceptorFactory

	// Optional: Sets ContextPropagators that allows users to control the context information passed through a workflow
	// default: no ContextPropagators
	ContextPropagators []ContextPropagator

	// Optional: Sets opentracing Tracer that is to be used to emit tracing information
	// default: no tracer - opentracing.NoopTracer
	Tracer opentracing.Tracer

	// Optional: flags to turn on/off some features on server side
	// default: all features under the struct is turned off
	FeatureFlags FeatureFlags
}

// IsReplayDomain checks if the domainName is from replay
func IsReplayDomain(dn string) bool {
	return replayDomainName == dn
}

// NewWorkflowReplayer creates an instance of the WorkflowReplayer
func NewWorkflowReplayer() *WorkflowReplayer {
	return NewWorkflowReplayerWithOptions(ReplayOptions{})
}

// NewWorkflowReplayerWithOptions creates an instance of the WorkflowReplayer
// with provided replay worker options
func NewWorkflowReplayerWithOptions(
	options ReplayOptions,
) *WorkflowReplayer {
	augmentReplayOptions(&options)
	return &WorkflowReplayer{
		registry: newRegistry(),
		options:  options,
	}
}

// RegisterWorkflow registers workflow function to replay
func (r *WorkflowReplayer) RegisterWorkflow(w interface{}) {
	r.registry.RegisterWorkflow(w)
}

// RegisterWorkflowWithOptions registers workflow function with custom workflow name to replay
func (r *WorkflowReplayer) RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	r.registry.RegisterWorkflowWithOptions(w, options)
}

// RegisterActivity registers an activity function for this replayer
func (r *WorkflowReplayer) RegisterActivity(a interface{}) {
	r.registry.RegisterActivity(a)
}

// RegisterActivityWithOptions registers an activity function for this replayer with custom options, e.g. an explicit name.
func (r *WorkflowReplayer) RegisterActivityWithOptions(a interface{}, options RegisterActivityOptions) {
	r.registry.RegisterActivityWithOptions(a, options)
}

// GetRegisteredWorkflows retrieves the registered workflows on the replayer
func (r *WorkflowReplayer) GetRegisteredWorkflows() []RegistryWorkflowInfo {
	workflows := r.registry.GetRegisteredWorkflows()
	var result []RegistryWorkflowInfo
	for _, wf := range workflows {
		result = append(result, wf)
	}
	return result
}

// GetRegisteredActivities retrieves the registered activities on the replayer
func (r *WorkflowReplayer) GetRegisteredActivities() []RegistryActivityInfo {
	activities := r.registry.getRegisteredActivities()
	var result []RegistryActivityInfo
	for _, a := range activities {
		result = append(result, a)
	}
	return result
}

// ReplayWorkflowHistory executes a single decision task for the given history.
// Use for testing backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
func (r *WorkflowReplayer) ReplayWorkflowHistory(logger *zap.Logger, history *shared.History) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	testReporter := logger.Sugar()
	controller := gomock.NewController(testReporter)
	service := workflowservicetest.NewMockClient(controller)

	return r.replayWorkflowHistory(logger, service, replayDomainName, nil, history, nil)
}

func (r *WorkflowReplayer) ReplayWorkflowHistoryFromJSON(logger *zap.Logger, reader io.Reader) error {
	return r.ReplayPartialWorkflowHistoryFromJSON(logger, reader, 0)
}

func (r *WorkflowReplayer) ReplayPartialWorkflowHistoryFromJSON(logger *zap.Logger, reader io.Reader, lastEventID int64) error {
	history, err := extractHistoryFromReader(reader, lastEventID)

	if err != nil {
		return err
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	testReporter := logger.Sugar()
	controller := gomock.NewController(testReporter)
	service := workflowservicetest.NewMockClient(controller)

	return r.replayWorkflowHistory(logger, service, replayDomainName, nil, history, nil)
}

// ReplayWorkflowHistoryFromJSONFile executes a single decision task for the given json history file.
// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
func (r *WorkflowReplayer) ReplayWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string) error {
	return r.ReplayPartialWorkflowHistoryFromJSONFile(logger, jsonfileName, 0)
}

// ReplayPartialWorkflowHistoryFromJSONFile executes a single decision task for the given json history file up to provided
// lastEventID(inclusive).
// Use for testing backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
func (r *WorkflowReplayer) ReplayPartialWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string, lastEventID int64) error {
	file, err := os.Open(jsonfileName)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()
	return r.ReplayPartialWorkflowHistoryFromJSON(logger, file, lastEventID)
}

// ReplayWorkflowExecution replays workflow execution loading it from Cadence service.
// The logger is an optional parameter. Defaults to the noop logger.
func (r *WorkflowReplayer) ReplayWorkflowExecution(
	ctx context.Context,
	service workflowserviceclient.Interface,
	logger *zap.Logger,
	domain string,
	execution WorkflowExecution,
) error {
	sharedExecution := &shared.WorkflowExecution{
		RunId:      common.StringPtr(execution.RunID),
		WorkflowId: common.StringPtr(execution.ID),
	}
	request := &shared.GetWorkflowExecutionHistoryRequest{
		Domain:    common.StringPtr(domain),
		Execution: sharedExecution,
	}

	var hResponse *shared.GetWorkflowExecutionHistoryResponse
	if err := backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, r.options.FeatureFlags)

			var err error
			hResponse, err = service.GetWorkflowExecutionHistory(tchCtx, request, opt...)
			cancel()

			return err
		},
		createDynamicServiceRetryPolicy(ctx),
		func(err error) bool {
			if _, ok := err.(*shared.InternalServiceError); ok {
				// treat InternalServiceError as non-retryable, as the workflow history may be corrupted
				return false
			}
			return isServiceTransientError(err)
		},
	); err != nil {
		return err
	}

	if hResponse.RawHistory != nil {
		history, err := serializer.DeserializeBlobDataToHistoryEvents(hResponse.RawHistory, shared.HistoryEventFilterTypeAllEvent)
		if err != nil {
			return err
		}

		hResponse.History = history
	}

	return r.replayWorkflowHistory(logger, service, domain, &execution, hResponse.History, hResponse.NextPageToken)
}

func (r *WorkflowReplayer) replayWorkflowHistory(
	logger *zap.Logger,
	service workflowserviceclient.Interface,
	domain string,
	execution *WorkflowExecution,
	history *shared.History,
	nextPageToken []byte,
) error {
	events := history.Events
	if events == nil {
		return errReplayEmptyHistory
	}
	if len(events) < 3 {
		return errReplayHistoryTooShort
	}
	first := events[0]
	if first.GetEventType() != shared.EventTypeWorkflowExecutionStarted {
		return errReplayInvalidFirstEvent
	}
	last := events[len(events)-1]

	attr := first.WorkflowExecutionStartedEventAttributes
	if attr == nil {
		return errReplayCorruptedStartedEvent
	}
	workflowType := attr.WorkflowType
	if execution == nil {
		execution = &WorkflowExecution{
			ID:    replayWorkflowID,
			RunID: uuid.NewRandom().String(),
		}
		if first.WorkflowExecutionStartedEventAttributes.GetOriginalExecutionRunId() != "" {
			execution.RunID = first.WorkflowExecutionStartedEventAttributes.GetOriginalExecutionRunId()
		}
	}

	task := &shared.PollForDecisionTaskResponse{
		Attempt:      common.Int64Ptr(int64(attr.GetAttempt())),
		TaskToken:    []byte(replayTaskToken),
		WorkflowType: workflowType,
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(execution.ID),
			RunId:      common.StringPtr(execution.RunID),
		},
		History:                history,
		PreviousStartedEventId: common.Int64Ptr(replayPreviousStartedEventID),
		NextPageToken:          nextPageToken,
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	workerParams := workerExecutionParameters{
		WorkerOptions: WorkerOptions{
			Identity:                          replayWorkerIdentity,
			DataConverter:                     r.options.DataConverter,
			ContextPropagators:                r.options.ContextPropagators,
			WorkflowInterceptorChainFactories: r.options.WorkflowInterceptorChainFactories,
			Tracer:                            r.options.Tracer,
			Logger:                            logger,
			DisableStickyExecution:            true,
		},
		TaskList: &shared.TaskList{
			Name: common.StringPtr(replayTaskListName),
			Kind: shared.TaskListKindNormal.Ptr(),
		},
	}

	metricScope := tally.NoopScope
	iterator := &historyIteratorImpl{
		nextPageToken:  task.NextPageToken,
		execution:      task.WorkflowExecution,
		domain:         domain,
		service:        service,
		metricsScope:   metricScope,
		startedEventID: task.GetStartedEventId(),
		featureFlags:   r.options.FeatureFlags,
	}
	taskHandler := newWorkflowTaskHandler(domain, workerParams, nil, r.registry)
	resp, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task, historyIterator: iterator}, nil)
	if err != nil {
		return err
	}

	//Technically speaking we do not need extra validations for the continue as new cases as they are equivalent to that case getting completed.
	if last.GetEventType() != shared.EventTypeWorkflowExecutionCompleted {
		return nil
	}
	// TODO: the following result will not be executed if nextPageToken is not nil, which is probably fine as the actual workflow task
	// processing logic does not have such check. If we want to always execute this check for closed workflows, we need to dump the
	// entire history before starting the replay as otherwise we can't get the last event here.
	// compare workflow results
	if resp != nil {
		completeReq, ok := resp.(*shared.RespondDecisionTaskCompletedRequest)
		if ok {
			for _, d := range completeReq.Decisions {
				if d.GetDecisionType() == shared.DecisionTypeCompleteWorkflowExecution &&
					last.GetEventType() == shared.EventTypeWorkflowExecutionCompleted {
					resultA := last.WorkflowExecutionCompletedEventAttributes.Result
					resultB := d.CompleteWorkflowExecutionDecisionAttributes.Result
					if bytes.Compare(resultA, resultB) == 0 {
						return nil
					}
				}
			}
		}
	}
	return fmt.Errorf("replay workflow doesn't return the same result as the last event, resp: %v, last: %v", resp, last)
}

func extractHistoryFromReader(r io.Reader, lastEventID int64) (*shared.History, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	var deserializedEvents []*shared.HistoryEvent
	err = json.Unmarshal(raw, &deserializedEvents)

	if err != nil {
		return nil, fmt.Errorf("invalid json contents: %w", err)
	}

	if lastEventID <= 0 {
		return &shared.History{Events: deserializedEvents}, nil
	}

	// Caller is potentially asking for subset of history instead of all history events
	var events []*shared.HistoryEvent
	for _, event := range deserializedEvents {
		events = append(events, event)
		if event.GetEventId() == lastEventID {
			// Copy history up to last event (inclusive)
			break
		}
	}

	return &shared.History{Events: events}, nil
}

func augmentReplayOptions(
	options *ReplayOptions,
) {
	// if the user passes in a tracer then add a tracing context propagator
	if options.Tracer != nil {
		options.ContextPropagators = append(options.ContextPropagators, NewTracingContextPropagator(zap.NewNop(), options.Tracer))
	} else {
		options.Tracer = opentracing.NoopTracer{}
	}
}
