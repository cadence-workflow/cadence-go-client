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
	"strings"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/debug"
	"go.uber.org/cadence/internal/common/metrics"
)

type (
	// activityTaskHandlerImpl is the implementation of ActivityTaskHandler
	activityTaskHandlerImpl struct {
		clock              clockwork.Clock
		taskListName       string
		identity           string
		service            workflowserviceclient.Interface
		metricsScope       *metrics.TaggedScope
		logger             *zap.Logger
		userContext        context.Context
		registry           *registry
		activityProvider   activityProvider
		dataConverter      DataConverter
		workerStopCh       <-chan struct{}
		contextPropagators []ContextPropagator
		tracer             opentracing.Tracer
		featureFlags       FeatureFlags
		activityTracker    debug.ActivityTracker
	}
)

func newActivityTaskHandler(
	service workflowserviceclient.Interface,
	params workerExecutionParameters,
	registry *registry,
) ActivityTaskHandler {
	return newActivityTaskHandlerWithCustomProvider(service, params, registry, nil, clockwork.NewRealClock())
}

func newActivityTaskHandlerWithCustomProvider(
	service workflowserviceclient.Interface,
	params workerExecutionParameters,
	registry *registry,
	activityProvider activityProvider,
	clock clockwork.Clock,
) ActivityTaskHandler {
	if params.Tracer == nil {
		params.Tracer = opentracing.NoopTracer{}
	}
	if params.WorkerStats.ActivityTracker == nil {
		params.WorkerStats.ActivityTracker = debug.NewNoopActivityTracker()
	}
	return &activityTaskHandlerImpl{
		clock:              clock,
		taskListName:       params.TaskList,
		identity:           params.Identity,
		service:            service,
		logger:             params.Logger,
		metricsScope:       metrics.NewTaggedScope(params.MetricsScope),
		userContext:        params.UserContext,
		registry:           registry,
		activityProvider:   activityProvider,
		dataConverter:      params.DataConverter,
		workerStopCh:       params.WorkerStopChannel,
		contextPropagators: params.ContextPropagators,
		tracer:             params.Tracer,
		featureFlags:       params.FeatureFlags,
		activityTracker:    params.WorkerStats.ActivityTracker,
	}
}

// Execute executes an implementation of the activity.
func (ath *activityTaskHandlerImpl) Execute(taskList string, t *s.PollForActivityTaskResponse) (result interface{}, err error) {
	traceLog(func() {
		ath.logger.Debug("Processing new activity task",
			zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
			zap.String(tagActivityType, t.ActivityType.GetName()))
	})

	rootCtx := ath.userContext
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	canCtx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	workflowType := t.WorkflowType.GetName()
	activityType := t.ActivityType.GetName()
	invoker := newServiceInvoker(t.TaskToken, ath.identity, ath.service, cancel, t.GetHeartbeatTimeoutSeconds(), ath.workerStopCh, ath.featureFlags, ath.logger, workflowType, activityType)
	defer func() {
		_, activityCompleted := result.(*s.RespondActivityTaskCompletedRequest)
		invoker.Close(!activityCompleted) // flush buffered heartbeat if activity was not successfully completed.
	}()

	metricsScope := getMetricsScopeForActivity(ath.metricsScope, workflowType, activityType)
	ctx := WithActivityTask(canCtx, t, taskList, invoker, ath.logger, metricsScope, ath.dataConverter, ath.workerStopCh, ath.contextPropagators, ath.tracer)

	activityImplementation := ath.getActivity(activityType)
	if activityImplementation == nil {
		// Couldn't find the activity implementation.
		supported := strings.Join(ath.getRegisteredActivityNames(), ", ")
		return nil, fmt.Errorf("unable to find activityType=%v. Supported types: [%v]", activityType, supported)
	}

	// panic handler
	defer func() {
		if p := recover(); p != nil {
			topLine := fmt.Sprintf("activity for %s [panic]:", ath.taskListName)
			st := getStackTraceRaw(topLine, 7, 0)
			ath.logger.Error("Activity panic.",
				zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
				zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
				zap.String(tagActivityType, activityType),
				zap.String(tagPanicError, fmt.Sprintf("%v", p)),
				zap.String(tagPanicStack, st))
			metricsScope.Counter(metrics.ActivityTaskPanicCounter).Inc(1)
			panicErr := newPanicError(p, st)
			result, err = convertActivityResultToRespondRequest(ath.identity, t.TaskToken, nil, panicErr, ath.dataConverter), nil
		}
	}()

	// propagate context information into the activity context from the headers
	for _, ctxProp := range ath.contextPropagators {
		var err error
		if ctx, err = ctxProp.Extract(ctx, NewHeaderReader(t.Header)); err != nil {
			return nil, fmt.Errorf("unable to propagate context %w", err)
		}
	}

	info := ctx.Value(activityEnvContextKey).(*activityEnvironment)
	ctx, dlCancelFunc := context.WithDeadline(ctx, info.deadline)
	defer dlCancelFunc()

	ctx, span := createOpenTracingActivitySpan(ctx, ath.tracer, time.Now(), activityType, t.WorkflowExecution.GetWorkflowId(), t.WorkflowExecution.GetRunId())
	defer span.Finish()

	if activityImplementation.GetOptions().EnableAutoHeartbeat && t.HeartbeatTimeoutSeconds != nil && *t.HeartbeatTimeoutSeconds > 0 {
		heartBeater := newHeartbeater(ath.workerStopCh, invoker, ath.logger, ath.clock, activityType, t.WorkflowExecution)
		go heartBeater.Run(ctx, time.Duration(*t.HeartbeatTimeoutSeconds)*time.Second)
	}
	activityInfo := debug.ActivityInfo{
		TaskList:     ath.taskListName,
		ActivityType: activityType,
	}
	defer ath.activityTracker.Start(activityInfo).Stop()
	output, err := activityImplementation.Execute(ctx, t.Input)

	dlCancelFunc()
	if <-ctx.Done(); ctx.Err() == context.DeadlineExceeded {
		ath.logger.Warn("Activity timeout.",
			zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
			zap.String(tagActivityType, activityType),
		)
		return nil, ctx.Err()
	}
	if err != nil && err != ErrActivityResultPending {
		ath.logger.Error("Activity error.",
			zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
			zap.String(tagActivityType, activityType),
			zap.Error(err),
		)
	}
	return convertActivityResultToRespondRequest(ath.identity, t.TaskToken, output, err, ath.dataConverter), nil
}

func (ath *activityTaskHandlerImpl) getActivity(name string) activity {
	if ath.activityProvider != nil {
		return ath.activityProvider(name)
	}

	if a, ok := ath.registry.GetActivity(name); ok {
		return a
	}

	return nil
}

func (ath *activityTaskHandlerImpl) getRegisteredActivityNames() (activityNames []string) {
	for _, a := range ath.registry.getRegisteredActivities() {
		activityNames = append(activityNames, a.ActivityType().Name)
	}
	return
}
