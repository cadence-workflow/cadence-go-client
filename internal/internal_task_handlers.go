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

// All code in this file is private to the package.

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/cache"
	"go.uber.org/cadence/internal/common/metrics"
)

const (
	defaultHeartBeatIntervalInSec = 10 * 60

	defaultStickyCacheSize = 10000

	noRetryBackoff = time.Duration(-1)

	defaultInstantLivedWorkflowTimeoutUpperLimitInSec = 1

	defaultShortLivedWorkflowTimeoutUpperLimitInSec = 1 * 1800

	defaultMediumLivedWorkflowTimeoutUpperLimitInSec = 8 * 3600
)

var (
	_ autoConfigHintAwareTask = (*workflowTask)(nil)
	_ autoConfigHintAwareTask = (*activityTask)(nil)
)

type (
	// workflowExecutionEventHandler process a single event.
	workflowExecutionEventHandler interface {
		// Process a single event and return the assosciated decisions.
		// Return List of decisions made, any error.
		ProcessEvent(event *s.HistoryEvent, isReplay bool, isLast bool) error
		// ProcessQuery process a query request.
		ProcessQuery(queryType string, queryArgs []byte) ([]byte, error)
		StackTrace() string
		// Close for cleaning up resources on this event handler
		Close()
	}

	// autoConfigHintAwareTask is a task that can provide auto config hint
	autoConfigHintAwareTask interface {
		getAutoConfigHint() *s.AutoConfigHint
	}

	// workflowTask wraps a decision task.
	workflowTask struct {
		task            *s.PollForDecisionTaskResponse
		autoConfigHint  *s.AutoConfigHint
		historyIterator HistoryIterator
		doneCh          chan struct{}
		laResultCh      chan *localActivityResult
	}

	// activityTask wraps a activity task.
	activityTask struct {
		task           *s.PollForActivityTaskResponse
		autoConfigHint *s.AutoConfigHint
		pollStartTime  time.Time
	}

	// resetStickinessTask wraps a ResetStickyTaskListRequest.
	resetStickinessTask struct {
		task *s.ResetStickyTaskListRequest
	}

	// workflowExecutionContextImpl is the cached workflow state for sticky execution
	workflowExecutionContextImpl struct {
		mutex             sync.Mutex
		workflowStartTime time.Time
		workflowInfo      *WorkflowInfo
		wth               *workflowTaskHandlerImpl

		// eventHandler is changed to a atomic.Value as a temporally bug fix for local activity
		// retry issue (github issue #915). Therefore, when accessing/modifying this field, the
		// mutex should still be held.
		eventHandler atomic.Value

		isWorkflowCompleted bool
		result              []byte
		err                 error

		previousStartedEventID int64

		newDecisions        []*s.Decision
		currentDecisionTask *s.PollForDecisionTaskResponse
		laTunnel            *localActivityTunnel
		decisionStartTime   time.Time
	}

	// workflowTaskHandlerImpl is the implementation of WorkflowTaskHandler
	workflowTaskHandlerImpl struct {
		domain                         string
		metricsScope                   *metrics.TaggedScope
		ppMgr                          pressurePointMgr
		logger                         *zap.Logger
		identity                       string
		enableLoggingInReplay          bool
		disableStickyExecution         bool
		registry                       *registry
		laTunnel                       *localActivityTunnel
		nonDeterministicWorkflowPolicy NonDeterministicWorkflowPolicy
		dataConverter                  DataConverter
		contextPropagators             []ContextPropagator
		tracer                         opentracing.Tracer
		workflowInterceptorFactories   []WorkflowInterceptorFactory
		disableStrictNonDeterminism    bool
	}

	activityProvider func(name string) activity

	// history wrapper method to help information about events.
	history struct {
		workflowTask   *workflowTask
		eventsHandler  *workflowExecutionEventHandlerImpl
		loadedEvents   []*s.HistoryEvent
		currentIndex   int
		nextEventID    int64 // next expected eventID for sanity
		lastEventID    int64 // last expected eventID, zero indicates read until end of stream
		next           []*s.HistoryEvent
		binaryChecksum *string
	}

	decisionHeartbeatError struct {
		Message string
	}
)

func (t *workflowTask) getAutoConfigHint() *s.AutoConfigHint {
	if t.task != nil {
		return t.task.AutoConfigHint
	}
	return t.autoConfigHint
}

func (t *activityTask) getAutoConfigHint() *s.AutoConfigHint {
	if t.task != nil {
		return t.task.AutoConfigHint
	}
	return t.autoConfigHint
}

func newHistory(task *workflowTask, eventsHandler *workflowExecutionEventHandlerImpl) *history {
	result := &history{
		workflowTask:  task,
		eventsHandler: eventsHandler,
		loadedEvents:  task.task.History.Events,
		currentIndex:  0,
		// don't set lastEventID to task.GetNextEventId()
		// as for sticky query, the history in workflow task will be empty
		// and query will be run based on existing workflow state.
		// so the sanity check in verifyAllEventsProcessed will fail
		lastEventID: task.task.GetStartedEventId(),
	}
	if len(result.loadedEvents) > 0 {
		result.nextEventID = result.loadedEvents[0].GetEventId()
	}
	return result
}

func (e decisionHeartbeatError) Error() string {
	return e.Message
}

// Get workflow start event.
func (eh *history) GetWorkflowStartedEvent() (*s.HistoryEvent, error) {
	events := eh.workflowTask.task.History.Events
	if len(events) == 0 || events[0].GetEventType() != s.EventTypeWorkflowExecutionStarted {
		return nil, errors.New("unable to find WorkflowExecutionStartedEventAttributes in the history")
	}
	return events[0], nil
}

func (eh *history) IsReplayEvent(event *s.HistoryEvent) bool {
	return event.GetEventId() <= eh.workflowTask.task.GetPreviousStartedEventId() || isDecisionEvent(event.GetEventType())
}

func (eh *history) IsNextDecisionFailed() (isFailed bool, binaryChecksum *string, err error) {

	nextIndex := eh.currentIndex + 1
	if nextIndex >= len(eh.loadedEvents) && eh.hasMoreEvents() { // current page ends and there is more pages
		if err := eh.loadMoreEvents(); err != nil {
			return false, nil, err
		}
	}

	if nextIndex < len(eh.loadedEvents) {
		nextEvent := eh.loadedEvents[nextIndex]
		nextEventType := nextEvent.GetEventType()
		isFailed := nextEventType == s.EventTypeDecisionTaskTimedOut || nextEventType == s.EventTypeDecisionTaskFailed
		var binaryChecksum *string
		if nextEventType == s.EventTypeDecisionTaskCompleted {
			binaryChecksum = nextEvent.DecisionTaskCompletedEventAttributes.BinaryChecksum
		}
		return isFailed, binaryChecksum, nil
	}
	return false, nil, nil
}

func (eh *history) loadMoreEvents() error {
	historyPage, err := eh.getMoreEvents()
	if err != nil {
		return err
	}
	eh.loadedEvents = append(eh.loadedEvents, historyPage.Events...)
	if eh.nextEventID == 0 && len(eh.loadedEvents) > 0 {
		eh.nextEventID = eh.loadedEvents[0].GetEventId()
	}
	return nil
}

func isDecisionEvent(eventType s.EventType) bool {
	switch eventType {
	case s.EventTypeWorkflowExecutionCompleted,
		s.EventTypeWorkflowExecutionFailed,
		s.EventTypeWorkflowExecutionCanceled,
		s.EventTypeWorkflowExecutionContinuedAsNew,
		s.EventTypeActivityTaskScheduled,
		s.EventTypeActivityTaskCancelRequested,
		s.EventTypeTimerStarted,
		s.EventTypeTimerCanceled,
		s.EventTypeCancelTimerFailed,
		s.EventTypeMarkerRecorded,
		s.EventTypeStartChildWorkflowExecutionInitiated,
		s.EventTypeRequestCancelExternalWorkflowExecutionInitiated,
		s.EventTypeSignalExternalWorkflowExecutionInitiated,
		s.EventTypeUpsertWorkflowSearchAttributes:
		return true
	default:
		return false
	}
}

// NextDecisionEvents returns events that there processed as new by the next decision.
// TODO(maxim): Refactor to return a struct instead of multiple parameters
func (eh *history) NextDecisionEvents() (result []*s.HistoryEvent, markers []*s.HistoryEvent, binaryChecksum *string, err error) {
	if eh.next == nil {
		eh.next, _, err = eh.nextDecisionEvents()
		if err != nil {
			return result, markers, eh.binaryChecksum, err
		}
	}

	result = eh.next
	checksum := eh.binaryChecksum
	if len(result) > 0 {
		eh.next, markers, err = eh.nextDecisionEvents()
	}
	return result, markers, checksum, err
}

func (eh *history) HasNextDecisionEvents() bool {
	return len(eh.next) != 0 || eh.currentIndex != len(eh.loadedEvents) || eh.hasMoreEvents()
}

func (eh *history) hasMoreEvents() bool {
	historyIterator := eh.workflowTask.historyIterator
	return historyIterator != nil && historyIterator.HasNextPage()
}

func (eh *history) getMoreEvents() (*s.History, error) {
	return eh.workflowTask.historyIterator.GetNextPage()
}

func (eh *history) verifyAllEventsProcessed() error {
	if eh.lastEventID > 0 && eh.nextEventID <= eh.lastEventID {
		return fmt.Errorf(
			"history_events: premature end of stream, expectedLastEventID=%v but no more events after eventID=%v",
			eh.lastEventID,
			eh.nextEventID-1)
	}
	if eh.lastEventID > 0 && eh.nextEventID != (eh.lastEventID+1) {
		eh.eventsHandler.logger.Warn(
			"history_events: processed events past the expected lastEventID",
			zap.Int64("expectedLastEventID", eh.lastEventID),
			zap.Int64("processedLastEventID", eh.nextEventID-1))
	}
	return nil
}

func (eh *history) nextDecisionEvents() (nextEvents []*s.HistoryEvent, markers []*s.HistoryEvent, err error) {
	if eh.currentIndex == len(eh.loadedEvents) && !eh.hasMoreEvents() {
		if err := eh.verifyAllEventsProcessed(); err != nil {
			return nil, nil, err
		}
		return []*s.HistoryEvent{}, []*s.HistoryEvent{}, nil
	}

	// Process events

OrderEvents:
	for {
		// load more history events if needed
		for eh.currentIndex == len(eh.loadedEvents) {
			if !eh.hasMoreEvents() {
				if err = eh.verifyAllEventsProcessed(); err != nil {
					return
				}
				break OrderEvents
			}
			if err = eh.loadMoreEvents(); err != nil {
				return
			}
		}

		event := eh.loadedEvents[eh.currentIndex]
		eventID := event.GetEventId()
		if eventID != eh.nextEventID {
			err = fmt.Errorf(
				"missing history events, expectedNextEventID=%v but receivedNextEventID=%v",
				eh.nextEventID, eventID)
			return
		}

		eh.nextEventID++

		switch event.GetEventType() {
		case s.EventTypeDecisionTaskStarted:
			isFailed, binaryChecksum, err1 := eh.IsNextDecisionFailed()
			if err1 != nil {
				err = err1
				return
			}
			if !isFailed {
				eh.binaryChecksum = binaryChecksum
				eh.currentIndex++
				nextEvents = append(nextEvents, event)
				break OrderEvents
			}
		case s.EventTypeDecisionTaskScheduled,
			s.EventTypeDecisionTaskTimedOut,
			s.EventTypeDecisionTaskFailed:
			// Skip
		default:
			if isPreloadMarkerEvent(event) {
				markers = append(markers, event)
			}
			nextEvents = append(nextEvents, event)
		}
		eh.currentIndex++
	}

	// shrink loaded events so it can be GCed
	eh.loadedEvents = eh.loadedEvents[eh.currentIndex:]
	eh.currentIndex = 0

	return nextEvents, markers, nil
}

func isPreloadMarkerEvent(event *s.HistoryEvent) bool {
	return event.GetEventType() == s.EventTypeMarkerRecorded
}

// newWorkflowTaskHandler returns an implementation of workflow task handler.
func newWorkflowTaskHandler(
	domain string,
	params workerExecutionParameters,
	ppMgr pressurePointMgr,
	registry *registry,
) WorkflowTaskHandler {
	ensureRequiredParams(&params)
	wth := &workflowTaskHandlerImpl{
		domain:                         domain,
		logger:                         params.Logger,
		ppMgr:                          ppMgr,
		metricsScope:                   metrics.NewTaggedScope(params.MetricsScope),
		identity:                       params.Identity,
		enableLoggingInReplay:          params.EnableLoggingInReplay,
		disableStickyExecution:         params.DisableStickyExecution,
		registry:                       registry,
		nonDeterministicWorkflowPolicy: params.NonDeterministicWorkflowPolicy,
		dataConverter:                  params.DataConverter,
		contextPropagators:             params.ContextPropagators,
		tracer:                         params.Tracer,
		workflowInterceptorFactories:   params.WorkflowInterceptorChainFactories,
		disableStrictNonDeterminism:    params.WorkerBugPorts.DisableStrictNonDeterminismCheck,
	}

	traceLog(func() {
		wth.logger.Debug("Workflow task handler is created.",
			zap.String(tagDomain, wth.domain),
			zap.Bool("disableStrictNonDeterminism", wth.disableStrictNonDeterminism))
	})

	return wth
}

// TODO: need a better eviction policy based on memory usage
var workflowCache cache.Cache
var stickyCacheSize = defaultStickyCacheSize
var initCacheOnce sync.Once
var stickyCacheLock sync.Mutex

// SetStickyWorkflowCacheSize sets the cache size for sticky workflow cache. Sticky workflow execution is the affinity
// between decision tasks of a specific workflow execution to a specific worker. The affinity is set if sticky execution
// is enabled via Worker.Options (It is enabled by default unless disabled explicitly). The benefit of sticky execution
// is that workflow does not have to reconstruct the state by replaying from beginning of history events. But the cost
// is it consumes more memory as it rely on caching workflow execution's running state on the worker. The cache is shared
// between workers running within same process. This must be called before any worker is started. If not called, the
// default size of 10K (might change in future) will be used.
func SetStickyWorkflowCacheSize(cacheSize int) {
	stickyCacheLock.Lock()
	defer stickyCacheLock.Unlock()
	if workflowCache != nil {
		panic("cache already created, please set cache size before worker starts.")
	}
	stickyCacheSize = cacheSize
}

func getWorkflowCache() cache.Cache {
	initCacheOnce.Do(func() {
		stickyCacheLock.Lock()
		defer stickyCacheLock.Unlock()
		workflowCache = cache.New(stickyCacheSize, &cache.Options{
			RemovedFunc: func(cachedEntity interface{}) {
				wc := cachedEntity.(*workflowExecutionContextImpl)
				wc.onEviction()
			},
		})
	})
	return workflowCache
}

func getWorkflowContext(runID string) *workflowExecutionContextImpl {
	o := getWorkflowCache().Get(runID)
	if o == nil {
		return nil
	}
	wc := o.(*workflowExecutionContextImpl)
	return wc
}

func putWorkflowContext(runID string, wc *workflowExecutionContextImpl) (*workflowExecutionContextImpl, error) {
	existing, err := getWorkflowCache().PutIfNotExist(runID, wc)
	if err != nil {
		return nil, err
	}
	return existing.(*workflowExecutionContextImpl), nil
}

func removeWorkflowContext(runID string) {
	getWorkflowCache().Delete(runID)
}

func newWorkflowExecutionContext(
	startTime time.Time,
	workflowInfo *WorkflowInfo,
	taskHandler *workflowTaskHandlerImpl,
) *workflowExecutionContextImpl {
	workflowContext := &workflowExecutionContextImpl{
		workflowStartTime: startTime,
		workflowInfo:      workflowInfo,
		wth:               taskHandler,
	}
	workflowContext.createEventHandler()
	return workflowContext
}

func (w *workflowExecutionContextImpl) Lock() {
	w.mutex.Lock()
}

func (w *workflowExecutionContextImpl) Unlock(err error) {
	cleared := false
	cached := getWorkflowCache().Exist(w.workflowInfo.WorkflowExecution.RunID)
	if err != nil || w.err != nil || w.isWorkflowCompleted || (w.wth.disableStickyExecution && !w.hasPendingLocalActivityWork()) {
		// TODO: in case of closed, it assumes the close decision always succeed. need server side change to return
		// error to indicate the close failure case. This should be rare case. For now, always remove the cache, and
		// if the close decision failed, the next decision will have to rebuild the state.
		if cached {
			// also clears state asynchronously via cache eviction
			removeWorkflowContext(w.workflowInfo.WorkflowExecution.RunID)
		} else {
			w.clearState()
		}
		cleared = true
	}
	// there are a variety of reasons a workflow may not have been put into the cache.
	// all of them mean we need to clear the state at this point, or any running goroutines will be orphaned.
	if !cleared && !cached {
		w.clearState()
	}

	w.mutex.Unlock()
}

func (w *workflowExecutionContextImpl) getEventHandler() *workflowExecutionEventHandlerImpl {
	eventHandler := w.eventHandler.Load()
	if eventHandler == nil {
		return nil
	}
	eventHandlerImpl, ok := eventHandler.(*workflowExecutionEventHandlerImpl)
	if !ok {
		panic("unknown type for workflow execution event handler")
	}
	return eventHandlerImpl
}

func (w *workflowExecutionContextImpl) completeWorkflow(result []byte, err error) {
	w.isWorkflowCompleted = true
	w.result = result
	w.err = err
}

func (w *workflowExecutionContextImpl) shouldResetStickyOnEviction() bool {
	// Not all evictions from the cache warrant a call to the server
	// to reset stickiness.
	// Cases when this is redundant or unnecessary include
	// when an error was encountered during execution
	// or workflow simply completed successfully.
	return w.err == nil && !w.isWorkflowCompleted
}

func (w *workflowExecutionContextImpl) onEviction() {
	// onEviction is run by LRU cache's removeFunc in separate goroutinue
	w.mutex.Lock()

	// Queue a ResetStickiness request *BEFORE* calling clearState
	// because once destroyed, no sensible information
	// may be ascertained about the execution context's state,
	// nor should any of its methods be invoked.
	if w.shouldResetStickyOnEviction() {
		w.queueResetStickinessTask()
	}

	w.clearState()
	w.mutex.Unlock()
}

func (w *workflowExecutionContextImpl) IsDestroyed() bool {
	return w.getEventHandler() == nil
}

func (w *workflowExecutionContextImpl) queueResetStickinessTask() {
	var task resetStickinessTask
	task.task = &s.ResetStickyTaskListRequest{
		Domain: common.StringPtr(w.workflowInfo.Domain),
		Execution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr(w.workflowInfo.WorkflowExecution.ID),
			RunId:      common.StringPtr(w.workflowInfo.WorkflowExecution.RunID),
		},
	}
	// w.laTunnel could be nil for worker.ReplayHistory() because there is no worker started, in that case we don't
	// care about resetStickinessTask.
	if w.laTunnel != nil && w.laTunnel.resultCh != nil {
		w.laTunnel.resultCh <- &task
	}
}

func (w *workflowExecutionContextImpl) clearState() {
	w.clearCurrentTask()
	w.isWorkflowCompleted = false
	w.result = nil
	w.err = nil
	w.previousStartedEventID = 0
	w.newDecisions = nil

	eventHandler := w.getEventHandler()
	if eventHandler != nil {
		// Set isReplay to true to prevent user code in defer guarded by !isReplaying() from running
		eventHandler.isReplay = true
		eventHandler.Close()
		w.eventHandler.Store((*workflowExecutionEventHandlerImpl)(nil))
	}
}

func (w *workflowExecutionContextImpl) createEventHandler() {
	w.clearState()
	eventHandler := newWorkflowExecutionEventHandler(
		w.workflowInfo,
		w.completeWorkflow,
		w.wth.logger,
		w.wth.enableLoggingInReplay,
		w.wth.metricsScope,
		w.wth.registry,
		w.wth.dataConverter,
		w.wth.contextPropagators,
		w.wth.tracer,
		w.wth.workflowInterceptorFactories,
	)
	w.eventHandler.Store(eventHandler)
}

func resetHistory(task *s.PollForDecisionTaskResponse, historyIterator HistoryIterator) (*s.History, error) {
	historyIterator.Reset()
	firstPageHistory, err := historyIterator.GetNextPage()
	if err != nil {
		return nil, err
	}
	task.History = firstPageHistory
	return firstPageHistory, nil
}

func (wth *workflowTaskHandlerImpl) createWorkflowContext(task *s.PollForDecisionTaskResponse) (*workflowExecutionContextImpl, error) {
	h := task.History
	attributes := h.Events[0].WorkflowExecutionStartedEventAttributes
	if attributes == nil {
		return nil, errors.New("first history event is not WorkflowExecutionStarted")
	}
	taskList := attributes.TaskList
	if taskList == nil {
		return nil, errors.New("nil TaskList in WorkflowExecutionStarted event")
	}

	runID := task.WorkflowExecution.GetRunId()
	workflowID := task.WorkflowExecution.GetWorkflowId()

	// Setup workflow Info
	var parentWorkflowExecution *WorkflowExecution
	if attributes.ParentWorkflowExecution != nil {
		parentWorkflowExecution = &WorkflowExecution{
			ID:    attributes.ParentWorkflowExecution.GetWorkflowId(),
			RunID: attributes.ParentWorkflowExecution.GetRunId(),
		}
	}
	workflowInfo := &WorkflowInfo{
		WorkflowExecution: WorkflowExecution{
			ID:    workflowID,
			RunID: runID,
		},
		OriginalRunId:                       attributes.GetOriginalExecutionRunId(),
		WorkflowType:                        flowWorkflowTypeFrom(*task.WorkflowType),
		TaskListName:                        taskList.GetName(),
		ExecutionStartToCloseTimeoutSeconds: attributes.GetExecutionStartToCloseTimeoutSeconds(),
		TaskStartToCloseTimeoutSeconds:      attributes.GetTaskStartToCloseTimeoutSeconds(),
		Domain:                              wth.domain,
		Attempt:                             attributes.GetAttempt(),
		lastCompletionResult:                attributes.LastCompletionResult,
		CronSchedule:                        attributes.CronSchedule,
		ContinuedExecutionRunID:             attributes.ContinuedExecutionRunId,
		ParentWorkflowDomain:                attributes.ParentWorkflowDomain,
		ParentWorkflowExecution:             parentWorkflowExecution,
		Memo:                                attributes.Memo,
		SearchAttributes:                    attributes.SearchAttributes,
		RetryPolicy:                         attributes.RetryPolicy,
	}

	wfStartTime := time.Unix(0, h.Events[0].GetTimestamp())
	return newWorkflowExecutionContext(wfStartTime, workflowInfo, wth), nil
}

func (wth *workflowTaskHandlerImpl) getOrCreateWorkflowContext(
	task *s.PollForDecisionTaskResponse,
	historyIterator HistoryIterator,
) (workflowContext *workflowExecutionContextImpl, err error) {
	metricsScope := wth.metricsScope.GetTaggedScope(tagWorkflowType, task.WorkflowType.GetName())
	defer func(metricsScope tally.Scope) {
		if err == nil && workflowContext != nil && workflowContext.laTunnel == nil {
			workflowContext.laTunnel = wth.laTunnel
		}
		metricsScope.Gauge(metrics.StickyCacheSize).Update(float64(getWorkflowCache().Size()))
	}(metricsScope)

	runID := task.WorkflowExecution.GetRunId()

	history := task.History
	isFullHistory := isFullHistory(history)

	workflowContext = nil
	if task.Query == nil || (task.Query != nil && !isFullHistory) {
		workflowContext = getWorkflowContext(runID)
	}

	if workflowContext != nil {
		workflowContext.Lock()
		// add new tag on metrics scope with workflow runtime length category
		scope := metricsScope.Tagged(map[string]string{tagWorkflowRuntimeLength: workflowCategorizedByTimeout(workflowContext)})
		if task.Query != nil && !isFullHistory {
			// query task and we have a valid cached state
			scope.Counter(metrics.StickyCacheHit).Inc(1)
		} else if history.Events[0].GetEventId() == workflowContext.previousStartedEventID+1 {
			// non query task and we have a valid cached state
			scope.Counter(metrics.StickyCacheHit).Inc(1)
		} else {
			// non query task and cached state is missing events, we need to discard the cached state and rebuild one.
			workflowContext.ResetIfStale(task, historyIterator)
		}
	} else {
		if !isFullHistory {
			// we are getting partial history task, but cached state was already evicted.
			// we need to reset history so we get events from beginning to replay/rebuild the state
			metricsScope.Counter(metrics.StickyCacheMiss).Inc(1)
			if history, err = resetHistory(task, historyIterator); err != nil {
				return
			}
		}

		if workflowContext, err = wth.createWorkflowContext(task); err != nil {
			return
		}

		if !wth.disableStickyExecution && task.Query == nil {
			workflowContext, _ = putWorkflowContext(runID, workflowContext)
		}
		workflowContext.Lock()
	}

	err = workflowContext.resetStateIfDestroyed(task, historyIterator)
	if err != nil {
		workflowContext.Unlock(err)
	}

	return
}

func isFullHistory(history *s.History) bool {
	if len(history.Events) == 0 || history.Events[0].GetEventType() != s.EventTypeWorkflowExecutionStarted {
		return false
	}
	return true
}

func (w *workflowExecutionContextImpl) resetStateIfDestroyed(task *s.PollForDecisionTaskResponse, historyIterator HistoryIterator) error {
	// It is possible that 2 threads (one for decision task and one for query task) that both are getting this same
	// cached workflowContext. If one task finished with err, it would destroy the cached state. In that case, the
	// second task needs to reset the cache state and start from beginning of the history.
	if w.IsDestroyed() {
		w.createEventHandler()
		// reset history events if necessary
		if !isFullHistory(task.History) {
			if _, err := resetHistory(task, historyIterator); err != nil {
				return err
			}
		}
	}
	return nil
}

// ProcessWorkflowTask processes all the events of the workflow task.
func (wth *workflowTaskHandlerImpl) ProcessWorkflowTask(
	workflowTask *workflowTask,
	heartbeatFunc decisionHeartbeatFunc,
) (completeRequest interface{}, errRet error) {
	if workflowTask == nil || workflowTask.task == nil {
		return nil, errors.New("nil workflow task provided")
	}
	task := workflowTask.task
	if task.History == nil || len(task.History.Events) == 0 {
		task.History = &s.History{
			Events: []*s.HistoryEvent{},
		}
	}
	if task.Query == nil && len(task.History.Events) == 0 {
		return nil, errors.New("nil or empty history")
	}

	if task.Query != nil && len(task.Queries) != 0 {
		return nil, errors.New("invalid query decision task")
	}

	runID := task.WorkflowExecution.GetRunId()
	workflowID := task.WorkflowExecution.GetWorkflowId()
	traceLog(func() {
		wth.logger.Debug("Processing new workflow task.",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, workflowID),
			zap.String(tagRunID, runID),
			zap.Int64("PreviousStartedEventId", task.GetPreviousStartedEventId()))
	})

	workflowContext, err := wth.getOrCreateWorkflowContext(task, workflowTask.historyIterator)
	if err != nil {
		return nil, err
	}

	defer func() {
		workflowContext.Unlock(errRet)
	}()

	var response interface{}
process_Workflow_Loop:
	for {
		startTime := time.Now()
		response, err = workflowContext.ProcessWorkflowTask(workflowTask)
		if err == nil && response == nil {
		wait_LocalActivity_Loop:
			for {
				deadlineToTrigger := time.Duration(float32(ratioToForceCompleteDecisionTaskComplete) * float32(workflowContext.GetDecisionTimeout()))
				delayDuration := startTime.Add(deadlineToTrigger).Sub(time.Now())
				select {
				case <-time.After(delayDuration):
					// force complete, call the decision heartbeat function
					workflowTask, err = heartbeatFunc(
						workflowContext.CompleteDecisionTask(workflowTask, false),
						startTime,
					)
					if err != nil {
						return nil, &decisionHeartbeatError{Message: fmt.Sprintf("error sending decision heartbeat %v", err)}
					}
					if workflowTask == nil {
						return nil, nil
					}
					continue process_Workflow_Loop

				case lar := <-workflowTask.laResultCh:
					// local activity result ready
					response, err = workflowContext.ProcessLocalActivityResult(workflowTask, lar)
					if err == nil && response == nil {
						// decision task is not done yet, still waiting for more local activities
						continue wait_LocalActivity_Loop
					}
					break process_Workflow_Loop
				}
			}
		} else {
			break process_Workflow_Loop
		}
	}
	return response, err
}

func (w *workflowExecutionContextImpl) ProcessWorkflowTask(workflowTask *workflowTask) (interface{}, error) {
	task := workflowTask.task
	historyIterator := workflowTask.historyIterator
	w.workflowInfo.HistoryBytesServer = task.GetTotalHistoryBytes()
	w.workflowInfo.HistoryCount = task.GetNextEventId() - 1
	if err := w.ResetIfStale(task, historyIterator); err != nil {
		return nil, err
	}
	w.SetCurrentTask(task)

	eventHandler := w.getEventHandler()
	reorderedHistory := newHistory(workflowTask, eventHandler)
	var replayDecisions []*s.Decision
	var respondEvents []*s.HistoryEvent

	skipReplayCheck := w.skipReplayCheck()
	isReplayTest := task.GetPreviousStartedEventId() == replayPreviousStartedEventID
	if isReplayTest {
		w.wth.logger.Info("Processing workflow task in replay test mode",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
		)
	}
	// Process events
ProcessEvents:
	for {
		reorderedEvents, markers, binaryChecksum, err := reorderedHistory.NextDecisionEvents()
		w.wth.metricsScope.GetTaggedScope("workflowtype", w.workflowInfo.WorkflowType.Name).Gauge(metrics.EstimatedHistorySize).Update(float64(w.workflowInfo.TotalHistoryBytes))
		w.wth.metricsScope.GetTaggedScope("workflowtype", w.workflowInfo.WorkflowType.Name).Gauge(metrics.ServerSideHistorySize).Update(float64(w.workflowInfo.HistoryBytesServer))
		if err != nil {
			return nil, err
		}

		if len(reorderedEvents) == 0 {
			break ProcessEvents
		}
		if binaryChecksum == nil {
			w.workflowInfo.BinaryChecksum = common.StringPtr(getBinaryChecksum())
		} else {
			w.workflowInfo.BinaryChecksum = binaryChecksum
		}
		// Markers are from the events that are produced from the current decision
		for _, m := range markers {
			if m.MarkerRecordedEventAttributes.GetMarkerName() != localActivityMarkerName {
				// local activity marker needs to be applied after decision task started event
				err := eventHandler.ProcessEvent(m, true, false)
				if err != nil {
					return nil, err
				}
				if w.isWorkflowCompleted {
					break ProcessEvents
				}
			}
		}

		for i, event := range reorderedEvents {
			isInReplay := reorderedHistory.IsReplayEvent(event)
			isLast := !isInReplay && i == len(reorderedEvents)-1
			if !skipReplayCheck && isDecisionEvent(event.GetEventType()) {
				respondEvents = append(respondEvents, event)
			}

			if isPreloadMarkerEvent(event) {
				// marker events are processed separately
				continue
			}

			// Any pressure points.
			err := w.wth.executeAnyPressurePoints(event, isInReplay)
			if err != nil {
				return nil, err
			}

			err = eventHandler.ProcessEvent(event, isInReplay, isLast)
			if err != nil {
				return nil, err
			}
			if w.isWorkflowCompleted {
				break ProcessEvents
			}
		}

		// now apply local activity markers
		for _, m := range markers {
			if m.MarkerRecordedEventAttributes.GetMarkerName() == localActivityMarkerName {
				err := eventHandler.ProcessEvent(m, true, false)
				if err != nil {
					return nil, err
				}
				if w.isWorkflowCompleted {
					break ProcessEvents
				}
			}
		}
		isReplay := len(reorderedEvents) > 0 && reorderedHistory.IsReplayEvent(reorderedEvents[len(reorderedEvents)-1])
		lastDecisionEventsForReplayTest := isReplayTest && !reorderedHistory.HasNextDecisionEvents()
		if isReplay && !lastDecisionEventsForReplayTest {
			eventDecisions := eventHandler.decisionsHelper.getDecisions(true)
			if len(eventDecisions) > 0 && !skipReplayCheck {
				replayDecisions = append(replayDecisions, eventDecisions...)
			}
		}
	}

	// Non-deterministic error could happen in 2 different places:
	//   1) the replay decisions does not match to history events. This is usually due to non backwards compatible code
	// change to decider logic. For example, change calling one activity to a different activity.
	//   2) the decision state machine is trying to make illegal state transition while replay a history event (like
	// activity task completed), but the corresponding decider code that start the event has been removed. In that case
	// the replay of that event will panic on the decision state machine and the workflow will be marked as completed
	// with the panic error.
	var nonDeterministicErr error
	var nonDeterminismType nonDeterminismDetectionType
	if !skipReplayCheck && !w.isWorkflowCompleted || isReplayTest {
		// check if decisions from reply matches to the history events
		if err := matchReplayWithHistory(w.workflowInfo, replayDecisions, respondEvents); err != nil {
			nonDeterministicErr = err
			nonDeterminismType = nonDeterminismDetectionTypeReplayComparison
		}
	} else if panicErr, ok := w.getWorkflowPanicIfIllegaleStatePanic(); ok {
		// This is a nondeterministic execution which ended up panicking
		nonDeterministicErr = panicErr
		nonDeterminismType = nonDeterminismDetectionTypeIllegalStatePanic
		// Since we know there is an error, we do the replay check to give more context in the log
		replayErr := matchReplayWithHistory(w.workflowInfo, replayDecisions, respondEvents)
		w.wth.logger.Error("Illegal state caused panic",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.Error(nonDeterministicErr),
			zap.NamedError("ReplayError", replayErr),
		)
	}

	if nonDeterministicErr != nil {
		scope := w.wth.metricsScope.GetTaggedScope(tagWorkflowType, task.WorkflowType.GetName(), tagNonDeterminismDetectionType, string(nonDeterminismType))
		scope.Counter(metrics.NonDeterministicError).Inc(1)
		w.wth.logger.Error("non-deterministic-error",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.Error(nonDeterministicErr))

		switch w.wth.nonDeterministicWorkflowPolicy {
		case NonDeterministicWorkflowPolicyFailWorkflow:
			// complete workflow with custom error will fail the workflow
			eventHandler.Complete(nil, NewCustomError("NonDeterministicWorkflowPolicyFailWorkflow", nonDeterministicErr.Error()))
		case NonDeterministicWorkflowPolicyBlockWorkflow:
			// return error here will be convert to DecisionTaskFailed for the first time, and ignored for subsequent
			// attempts which will cause DecisionTaskTimeout and server will retry forever until issue got fixed or
			// workflow timeout.
			return nil, nonDeterministicErr
		default:
			panic("unknown mismatched workflow history policy.")
		}
	}

	return w.CompleteDecisionTask(workflowTask, true), nil
}

func (w *workflowExecutionContextImpl) ProcessLocalActivityResult(workflowTask *workflowTask, lar *localActivityResult) (interface{}, error) {
	if lar.err != nil && w.retryLocalActivity(lar) {
		return nil, nil // nothing to do here as we are retrying...
	}

	err := w.getEventHandler().ProcessLocalActivityResult(lar)
	if err != nil {
		return nil, err
	}

	return w.CompleteDecisionTask(workflowTask, true), nil
}

func (w *workflowExecutionContextImpl) retryLocalActivity(lar *localActivityResult) bool {
	if lar.task.retryPolicy == nil || lar.err == nil || IsCanceledError(lar.err) {
		return false
	}

	backoff := getRetryBackoff(lar, time.Now())
	if backoff > 0 && backoff <= w.GetDecisionTimeout() {
		// we need a local retry
		time.AfterFunc(backoff, func() {
			// TODO: this should not be a separate goroutine as it introduces race condition when accessing eventHandler.
			// currently this is solved by changing eventHandler to an atomic.Value. Ideally, this retry timer should be
			// part of the event loop for processing the workflow task.
			eventHandler := w.getEventHandler()

			// if decision heartbeat failed, the workflow execution context will be cleared and eventHandler will be nil
			if eventHandler == nil {
				return
			}

			if _, ok := eventHandler.pendingLaTasks[lar.task.activityID]; !ok {
				return
			}

			lar.task.attempt++

			if !w.laTunnel.sendTask(lar.task) {
				lar.task.attempt--
			}
		})
		return true
	}
	// Backoff could be large and potentially much larger than DecisionTaskTimeout. We cannot just sleep locally for
	// retry. Because it will delay the local activity from complete which keeps the decision task open. In order to
	// keep decision task open, we have to keep "heartbeating" current decision task.
	// In that case, it is more efficient to create a server timer with backoff duration and retry when that backoff
	// timer fires. So here we will return false to indicate we don't need local retry anymore. However, we have to
	// store the current attempt and backoff to the same LocalActivityResultMarker so the replay can do the right thing.
	// The backoff timer will be created by workflow.ExecuteLocalActivity().
	lar.backoff = backoff

	return false
}

func getRetryBackoff(lar *localActivityResult, now time.Time) time.Duration {
	p := lar.task.retryPolicy
	var errReason string
	if len(p.NonRetriableErrorReasons) > 0 {
		if lar.err == ErrDeadlineExceeded {
			errReason = "timeout:" + s.TimeoutTypeScheduleToClose.String()
		} else {
			errReason, _ = getErrorDetails(lar.err, nil)
		}
	}
	return getRetryBackoffWithNowTime(p, lar.task.attempt, errReason, now, lar.task.expireTime)
}

func getRetryBackoffWithNowTime(p *RetryPolicy, attempt int32, errReason string, now, expireTime time.Time) time.Duration {
	if p.MaximumAttempts == 0 && p.ExpirationInterval == 0 {
		return noRetryBackoff
	}

	if p.MaximumAttempts > 0 && attempt > p.MaximumAttempts-1 {
		return noRetryBackoff // max attempt reached
	}

	backoffInterval := time.Duration(float64(p.InitialInterval) * math.Pow(p.BackoffCoefficient, float64(attempt)))
	if backoffInterval <= 0 {
		// math.Pow() could overflow
		if p.MaximumInterval > 0 {
			backoffInterval = p.MaximumInterval
		} else {
			return noRetryBackoff
		}
	}

	if p.MaximumInterval > 0 && backoffInterval > p.MaximumInterval {
		// cap next interval to MaxInterval
		backoffInterval = p.MaximumInterval
	}

	nextScheduleTime := now.Add(backoffInterval)
	if !expireTime.IsZero() && nextScheduleTime.After(expireTime) {
		return noRetryBackoff
	}

	// check if error is non-retriable
	for _, er := range p.NonRetriableErrorReasons {
		if er == errReason {
			return noRetryBackoff
		}
	}

	return backoffInterval
}

func (w *workflowExecutionContextImpl) CompleteDecisionTask(workflowTask *workflowTask, waitLocalActivities bool) interface{} {
	if w.currentDecisionTask == nil {
		return nil
	}
	eventHandler := w.getEventHandler()

	// w.laTunnel could be nil for worker.ReplayHistory() because there is no worker started, in that case we don't
	// care about the pending local activities, and just return because the result is ignored anyway by the caller.
	if w.hasPendingLocalActivityWork() && w.laTunnel != nil {
		if len(eventHandler.unstartedLaTasks) > 0 {
			// start new local activity tasks
			unstartedLaTasks := make(map[string]struct{})
			for activityID := range eventHandler.unstartedLaTasks {
				task := eventHandler.pendingLaTasks[activityID]
				task.wc = w
				task.workflowTask = workflowTask
				if !w.laTunnel.sendTask(task) {
					unstartedLaTasks[activityID] = struct{}{}
					task.wc = nil
					task.workflowTask = nil
				}
			}
			eventHandler.unstartedLaTasks = unstartedLaTasks
		}
		// cannot complete decision task as there are pending local activities
		if waitLocalActivities {
			return nil
		}
	}

	eventDecisions := eventHandler.decisionsHelper.getDecisions(true)
	if len(eventDecisions) > 0 {
		w.newDecisions = append(w.newDecisions, eventDecisions...)
	}

	completeRequest := w.wth.completeWorkflow(eventHandler, w.currentDecisionTask, w, w.newDecisions, !waitLocalActivities)
	w.clearCurrentTask()

	return completeRequest
}

func (w *workflowExecutionContextImpl) hasPendingLocalActivityWork() bool {
	eventHandler := w.getEventHandler()
	return !w.isWorkflowCompleted &&
		w.currentDecisionTask != nil &&
		w.currentDecisionTask.Query == nil && // don't run local activity for query task
		eventHandler != nil &&
		len(eventHandler.pendingLaTasks) > 0
}

func (w *workflowExecutionContextImpl) clearCurrentTask() {
	w.newDecisions = nil
	w.currentDecisionTask = nil
}

func (w *workflowExecutionContextImpl) skipReplayCheck() bool {
	return w.currentDecisionTask.Query != nil || !isFullHistory(w.currentDecisionTask.History)
}

func (w *workflowExecutionContextImpl) SetCurrentTask(task *s.PollForDecisionTaskResponse) {
	w.currentDecisionTask = task
	// do not update the previousStartedEventID for query task
	if task.Query == nil {
		w.previousStartedEventID = task.GetStartedEventId()
	}
	w.decisionStartTime = time.Now()
}

func (w *workflowExecutionContextImpl) ResetIfStale(task *s.PollForDecisionTaskResponse, historyIterator HistoryIterator) error {
	if len(task.History.Events) > 0 && task.History.Events[0].GetEventId() != w.previousStartedEventID+1 {
		w.wth.logger.Debug("Cached state staled, new task has unexpected events",
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.Int64("CachedPreviousStartedEventID", w.previousStartedEventID),
			zap.Int64("TaskFirstEventID", task.History.Events[0].GetEventId()),
			zap.Int64("TaskStartedEventID", task.GetStartedEventId()),
			zap.Int64("TaskPreviousStartedEventID", task.GetPreviousStartedEventId()))

		w.wth.metricsScope.
			GetTaggedScope(tagWorkflowType, task.WorkflowType.GetName()).
			Counter(metrics.StickyCacheStall).Inc(1)

		w.clearState()
		return w.resetStateIfDestroyed(task, historyIterator)
	}
	return nil
}

func (w *workflowExecutionContextImpl) GetDecisionTimeout() time.Duration {
	return time.Second * time.Duration(w.workflowInfo.TaskStartToCloseTimeoutSeconds)
}

func (w *workflowExecutionContextImpl) getWorkflowPanicIfIllegaleStatePanic() (*workflowPanicError, bool) {
	if !w.isWorkflowCompleted || w.err == nil {
		return nil, false
	}

	panicErr, ok := w.err.(*workflowPanicError)
	if !ok || panicErr.value == nil {
		return nil, false
	}

	_, ok = panicErr.value.(stateMachineIllegalStatePanic)
	if !ok {
		return nil, false
	}

	return panicErr, true
}

func (wth *workflowTaskHandlerImpl) completeWorkflow(
	eventHandler *workflowExecutionEventHandlerImpl,
	task *s.PollForDecisionTaskResponse,
	workflowContext *workflowExecutionContextImpl,
	decisions []*s.Decision,
	forceNewDecision bool) interface{} {

	// for query task
	if task.Query != nil {
		queryCompletedRequest := &s.RespondQueryTaskCompletedRequest{TaskToken: task.TaskToken}
		if panicErr, ok := workflowContext.err.(*PanicError); ok {
			// NOTE: this code path should never be executed, we should check for workflowPanicError instead of PanicError
			wth.logger.Warn("Encountered PanicError in workflow query task",
				zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
				zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
				zap.String(tagPanicError, panicErr.Error()),
				zap.String(tagPanicStack, panicErr.StackTrace()),
			)

			queryCompletedRequest.CompletedType = common.QueryTaskCompletedTypePtr(s.QueryTaskCompletedTypeFailed)
			queryCompletedRequest.ErrorMessage = common.StringPtr("Workflow panic: " + panicErr.Error())
			return queryCompletedRequest
		}

		if workflowPanicErr, ok := workflowContext.err.(*workflowPanicError); ok {
			// NOTE: in this case we should return complete query task with CompletedTypeFailed
			// but we didn't check for the right error type before, this may break existing customer
			wth.logger.Warn("Ignored workflow panic error for query, query result may be partial",
				zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
				zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
				zap.String(tagPanicError, workflowPanicErr.Error()),
				zap.String(tagPanicStack, workflowPanicErr.StackTrace()),
				zap.Int64("PreviousStartedEventID", task.GetPreviousStartedEventId()),
			)
		}

		result, err := eventHandler.ProcessQuery(task.Query.GetQueryType(), task.Query.QueryArgs)
		if err != nil {
			queryCompletedRequest.CompletedType = common.QueryTaskCompletedTypePtr(s.QueryTaskCompletedTypeFailed)
			queryCompletedRequest.ErrorMessage = common.StringPtr(err.Error())
		} else {
			queryCompletedRequest.CompletedType = common.QueryTaskCompletedTypePtr(s.QueryTaskCompletedTypeCompleted)
			queryCompletedRequest.QueryResult = result
		}
		return queryCompletedRequest
	}

	metricsScope := wth.metricsScope.GetTaggedScope(tagWorkflowType, eventHandler.workflowEnvironmentImpl.workflowInfo.WorkflowType.Name)

	// fail decision task on decider panic
	if panicErr, ok := workflowContext.err.(*workflowPanicError); ok {
		// Workflow panic
		metricsScope.Counter(metrics.DecisionTaskPanicCounter).Inc(1)
		wth.logger.Error("Workflow panic.",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.String(tagPanicError, panicErr.Error()),
			zap.String(tagPanicStack, panicErr.StackTrace()))
		return errorToFailDecisionTask(task.TaskToken, panicErr, wth.identity)
	}

	// complete decision task
	var closeDecision *s.Decision
	if canceledErr, ok := workflowContext.err.(*CanceledError); ok {
		// Workflow cancelled
		metricsScope.Counter(metrics.WorkflowCanceledCounter).Inc(1)
		closeDecision = createNewDecision(s.DecisionTypeCancelWorkflowExecution)
		_, details := getErrorDetails(canceledErr, wth.dataConverter)
		closeDecision.CancelWorkflowExecutionDecisionAttributes = &s.CancelWorkflowExecutionDecisionAttributes{
			Details: details,
		}
	} else if contErr, ok := workflowContext.err.(*ContinueAsNewError); ok {
		// Continue as new error.
		metricsScope.Counter(metrics.WorkflowContinueAsNewCounter).Inc(1)
		closeDecision = createNewDecision(s.DecisionTypeContinueAsNewWorkflowExecution)
		closeDecision.ContinueAsNewWorkflowExecutionDecisionAttributes = &s.ContinueAsNewWorkflowExecutionDecisionAttributes{
			WorkflowType:                        workflowTypePtr(*contErr.params.workflowType),
			Input:                               contErr.params.input,
			TaskList:                            common.TaskListPtr(s.TaskList{Name: contErr.params.taskListName}),
			ExecutionStartToCloseTimeoutSeconds: contErr.params.executionStartToCloseTimeoutSeconds,
			TaskStartToCloseTimeoutSeconds:      contErr.params.taskStartToCloseTimeoutSeconds,
			Header:                              contErr.params.header,
			Memo:                                workflowContext.workflowInfo.Memo,
			SearchAttributes:                    workflowContext.workflowInfo.SearchAttributes,
			RetryPolicy:                         workflowContext.workflowInfo.RetryPolicy,
		}
	} else if workflowContext.err != nil {
		// Workflow failures
		metricsScope.Counter(metrics.WorkflowFailedCounter).Inc(1)
		closeDecision = createNewDecision(s.DecisionTypeFailWorkflowExecution)
		reason, details := getErrorDetails(workflowContext.err, wth.dataConverter)
		closeDecision.FailWorkflowExecutionDecisionAttributes = &s.FailWorkflowExecutionDecisionAttributes{
			Reason:  common.StringPtr(reason),
			Details: details,
		}
	} else if workflowContext.isWorkflowCompleted {
		// Workflow completion
		metricsScope.Counter(metrics.WorkflowCompletedCounter).Inc(1)
		closeDecision = createNewDecision(s.DecisionTypeCompleteWorkflowExecution)
		closeDecision.CompleteWorkflowExecutionDecisionAttributes = &s.CompleteWorkflowExecutionDecisionAttributes{
			Result: workflowContext.result,
		}
	}

	if closeDecision != nil {
		decisions = append(decisions, closeDecision)
		elapsed := time.Since(workflowContext.workflowStartTime)
		metricsScope.Timer(metrics.WorkflowEndToEndLatency).Record(elapsed)
		forceNewDecision = false
	}

	var queryResults map[string]*s.WorkflowQueryResult
	if len(task.Queries) != 0 {
		queryResults = make(map[string]*s.WorkflowQueryResult)
		for queryID, query := range task.Queries {
			result, err := eventHandler.ProcessQuery(query.GetQueryType(), query.QueryArgs)
			if err != nil {
				queryResults[queryID] = &s.WorkflowQueryResult{
					ResultType:   common.QueryResultTypePtr(s.QueryResultTypeFailed),
					ErrorMessage: common.StringPtr(err.Error()),
				}
			} else {
				queryResults[queryID] = &s.WorkflowQueryResult{
					ResultType: common.QueryResultTypePtr(s.QueryResultTypeAnswered),
					Answer:     result,
				}
			}
		}
	}

	return &s.RespondDecisionTaskCompletedRequest{
		TaskToken:                  task.TaskToken,
		Decisions:                  decisions,
		Identity:                   common.StringPtr(wth.identity),
		ReturnNewDecisionTask:      common.BoolPtr(true),
		ForceCreateNewDecisionTask: common.BoolPtr(forceNewDecision),
		BinaryChecksum:             common.StringPtr(getBinaryChecksum()),
		QueryResults:               queryResults,
	}
}

func errorToFailDecisionTask(taskToken []byte, err error, identity string) *s.RespondDecisionTaskFailedRequest {
	failedCause := s.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure
	_, details := getErrorDetails(err, nil)
	return &s.RespondDecisionTaskFailedRequest{
		TaskToken:      taskToken,
		Cause:          &failedCause,
		Details:        details,
		Identity:       common.StringPtr(identity),
		BinaryChecksum: common.StringPtr(getBinaryChecksum()),
	}
}

func (wth *workflowTaskHandlerImpl) executeAnyPressurePoints(event *s.HistoryEvent, isInReplay bool) error {
	if wth.ppMgr != nil && !reflect.ValueOf(wth.ppMgr).IsNil() && !isInReplay {
		switch event.GetEventType() {
		case s.EventTypeDecisionTaskStarted:
			return wth.ppMgr.Execute(pressurePointTypeDecisionTaskStartTimeout)
		case s.EventTypeActivityTaskScheduled:
			return wth.ppMgr.Execute(pressurePointTypeActivityTaskScheduleTimeout)
		case s.EventTypeActivityTaskStarted:
			return wth.ppMgr.Execute(pressurePointTypeActivityTaskStartTimeout)
		case s.EventTypeDecisionTaskCompleted:
			return wth.ppMgr.Execute(pressurePointTypeDecisionTaskCompleted)
		}
	}
	return nil
}

type cadenceInvoker struct {
	sync.Mutex
	identity              string
	service               workflowserviceclient.Interface
	taskToken             []byte
	cancelHandler         func()
	heartBeatTimeoutInSec int32       // The heart beat interval configured for this activity.
	hbBatchEndTimer       *time.Timer // Whether we started a batch of operations that need to be reported in the cycle. This gets started on a user call.
	detailsToReport       *[]byte     // Details to be reported in the next reporting interval.
	lastDetailsReported   *[]byte     // Details that were reported in the last reporting interval.
	closeCh               chan struct{}
	workerStopChannel     <-chan struct{}
	featureFlags          FeatureFlags
	logger                *zap.Logger
	workflowType          string
	activityType          string
}

func (i *cadenceInvoker) Heartbeat(details []byte) error {
	i.Lock()
	defer i.Unlock()

	_, err := i.internalHeartBeat(details)
	return err
}

func (i *cadenceInvoker) BackgroundHeartbeat() error {
	i.Lock()
	defer i.Unlock()

	if i.hbBatchEndTimer != nil {
		if i.detailsToReport == nil {
			i.detailsToReport = i.lastDetailsReported
		}

		return nil
	}

	var details []byte
	if i.detailsToReport != nil {
		details = *i.detailsToReport
	} else if i.lastDetailsReported != nil {
		details = *i.lastDetailsReported
	}

	return i.heartbeatAndScheduleNextRun(details)
}

func (i *cadenceInvoker) BatchHeartbeat(details []byte) error {
	i.Lock()
	defer i.Unlock()

	if i.hbBatchEndTimer != nil {
		// If we have started batching window, keep track of last reported progress.
		i.detailsToReport = &details
		return nil
	}

	return i.heartbeatAndScheduleNextRun(details)
}

func (i *cadenceInvoker) heartbeatAndScheduleNextRun(details []byte) error {
	isActivityCancelled, err := i.internalHeartBeat(details)

	// If the activity is cancelled, the activity can ignore the cancellation and do its work
	// and complete. Our cancellation is co-operative, so we will try to heartbeat.
	if err == nil || isActivityCancelled {
		// We have successfully sent heartbeat, start next batching window.
		i.lastDetailsReported = &details
		i.detailsToReport = nil

		// Create timer to fire before the threshold to report.
		deadlineToTrigger := i.heartBeatTimeoutInSec
		if deadlineToTrigger <= 0 {
			// If we don't have any heartbeat timeout configured.
			deadlineToTrigger = defaultHeartBeatIntervalInSec
		}

		// We set a deadline at 80% of the timeout.
		duration := time.Duration(0.8*float32(deadlineToTrigger)) * time.Second
		i.hbBatchEndTimer = time.NewTimer(duration)

		go func() {
			select {
			case <-i.hbBatchEndTimer.C:
				// We are close to deadline.
			case <-i.workerStopChannel:
				// Activity worker is close to stop. This does the same steps as batch timer ends.
			case <-i.closeCh:
				// We got closed.
				return
			}

			// We close the batch and report the progress.
			var detailsToReport *[]byte

			i.Lock()
			detailsToReport = i.detailsToReport
			i.hbBatchEndTimer.Stop()
			i.hbBatchEndTimer = nil

			var err error
			if detailsToReport != nil {
				err = i.heartbeatAndScheduleNextRun(*detailsToReport)
			}
			i.Unlock()

			// Log the error outside the lock.
			i.logFailedHeartBeat(err)
		}()
	}

	return err
}

func (i *cadenceInvoker) logFailedHeartBeat(err error) {
	// If the error is a canceled error do not log, as this is expected.
	var canceledErr *CanceledError

	// We need to check for nil as errors.As returns false for nil. Which would cause us to log on nil.
	if err != nil && !errors.As(err, &canceledErr) {
		i.logger.Error("Failed to send heartbeat", zap.Error(err), zap.String(tagWorkflowType, i.workflowType), zap.String(tagActivityType, i.activityType))
	}
}

func (i *cadenceInvoker) internalHeartBeat(details []byte) (bool, error) {
	isActivityCancelled := false
	timeout := time.Duration(i.heartBeatTimeoutInSec) * time.Second
	if timeout <= 0 {
		timeout = time.Duration(defaultHeartBeatIntervalInSec) * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := recordActivityHeartbeat(ctx, i.service, i.identity, i.taskToken, details, i.featureFlags)

	switch err.(type) {
	case *CanceledError:
		// We are asked to cancel. inform the activity about cancellation through context.
		i.cancelHandler()
		isActivityCancelled = true

	case *s.EntityNotExistsError, *s.WorkflowExecutionAlreadyCompletedError, *s.DomainNotActiveError:
		// We will pass these through as cancellation for now but something we can change
		// later when we have setter on cancel handler.
		i.cancelHandler()
		isActivityCancelled = true
	}

	// We don't want to bubble temporary errors to the user.
	// This error won't be return to user check RecordActivityHeartbeat().
	return isActivityCancelled, err
}

func (i *cadenceInvoker) Close(flushBufferedHeartbeat bool) {
	i.Lock()
	defer i.Unlock()
	close(i.closeCh)
	if i.hbBatchEndTimer != nil {
		i.hbBatchEndTimer.Stop()
		if flushBufferedHeartbeat && i.detailsToReport != nil {
			i.internalHeartBeat(*i.detailsToReport)
			i.lastDetailsReported = i.detailsToReport
			i.detailsToReport = nil
		}
	}
}

func (i *cadenceInvoker) SignalWorkflow(ctx context.Context, domain, workflowID, runID, signalName string, signalInput []byte) error {
	return signalWorkflow(ctx, i.service, i.identity, domain, workflowID, runID, signalName, signalInput, i.featureFlags)
}

func newServiceInvoker(
	taskToken []byte,
	identity string,
	service workflowserviceclient.Interface,
	cancelHandler func(),
	heartBeatTimeoutInSec int32,
	workerStopChannel <-chan struct{},
	featureFlags FeatureFlags,
	logger *zap.Logger,
	workflowType string,
	activityType string,
) ServiceInvoker {
	return &cadenceInvoker{
		taskToken:             taskToken,
		identity:              identity,
		service:               service,
		cancelHandler:         cancelHandler,
		heartBeatTimeoutInSec: heartBeatTimeoutInSec,
		closeCh:               make(chan struct{}),
		workerStopChannel:     workerStopChannel,
		featureFlags:          featureFlags,
		logger:                logger,
		workflowType:          workflowType,
		activityType:          activityType,
	}
}

func createNewDecision(decisionType s.DecisionType) *s.Decision {
	return &s.Decision{
		DecisionType: common.DecisionTypePtr(decisionType),
	}
}
func signalWorkflow(
	ctx context.Context,
	service workflowserviceclient.Interface,
	identity string,
	domain string,
	workflowID string,
	runID string,
	signalName string,
	signalInput []byte,
	featureFlags FeatureFlags,
) error {
	request := &s.SignalWorkflowExecutionRequest{
		Domain: common.StringPtr(domain),
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      getRunID(runID),
		},
		SignalName: common.StringPtr(signalName),
		Input:      signalInput,
		Identity:   common.StringPtr(identity),
	}

	return backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
			defer cancel()
			return service.SignalWorkflowExecution(tchCtx, request, opt...)
		}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
}

func recordActivityHeartbeat(
	ctx context.Context,
	service workflowserviceclient.Interface,
	identity string,
	taskToken, details []byte,
	featureFlags FeatureFlags,
) error {
	request := &s.RecordActivityTaskHeartbeatRequest{
		TaskToken: taskToken,
		Details:   details,
		Identity:  common.StringPtr(identity)}

	var heartbeatResponse *s.RecordActivityTaskHeartbeatResponse
	heartbeatErr := backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
			defer cancel()

			var err error
			heartbeatResponse, err = service.RecordActivityTaskHeartbeat(tchCtx, request, opt...)
			return err
		}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)

	if heartbeatErr == nil && heartbeatResponse != nil && heartbeatResponse.GetCancelRequested() {
		return NewCanceledError()
	}

	return heartbeatErr
}

func recordActivityHeartbeatByID(
	ctx context.Context,
	service workflowserviceclient.Interface,
	identity string,
	domain, workflowID, runID, activityID string,
	details []byte,
	featureFlags FeatureFlags,
) error {
	request := &s.RecordActivityTaskHeartbeatByIDRequest{
		Domain:     common.StringPtr(domain),
		WorkflowID: common.StringPtr(workflowID),
		RunID:      common.StringPtr(runID),
		ActivityID: common.StringPtr(activityID),
		Details:    details,
		Identity:   common.StringPtr(identity)}

	var heartbeatResponse *s.RecordActivityTaskHeartbeatResponse
	heartbeatErr := backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
			defer cancel()

			var err error
			heartbeatResponse, err = service.RecordActivityTaskHeartbeatByID(tchCtx, request, opt...)
			return err
		}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)

	if heartbeatErr == nil && heartbeatResponse != nil && heartbeatResponse.GetCancelRequested() {
		return NewCanceledError()
	}

	return heartbeatErr
}

// This enables verbose logging in the client library.
// check worker.EnableVerboseLogging()
func traceLog(fn func()) {
	if enableVerboseLogging {
		fn()
	}
}

func workflowCategorizedByTimeout(wfContext *workflowExecutionContextImpl) string {
	executionTimeout := wfContext.workflowInfo.ExecutionStartToCloseTimeoutSeconds
	if executionTimeout <= defaultInstantLivedWorkflowTimeoutUpperLimitInSec {
		return "instant"
	} else if executionTimeout <= defaultShortLivedWorkflowTimeoutUpperLimitInSec {
		return "short"
	} else if executionTimeout <= defaultMediumLivedWorkflowTimeoutUpperLimitInSec {
		return "intermediate"
	} else {
		return "long"
	}
}
