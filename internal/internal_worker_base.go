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
	"os"
	"sync"
	"syscall"
	"time"

	"go.uber.org/cadence/internal/common/debug"
	"go.uber.org/cadence/internal/worker"

	"github.com/jonboulle/clockwork"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/metrics"
	"go.uber.org/cadence/internal/common/util"
)

const (
	retryPollOperationInitialInterval = 20 * time.Millisecond
	retryPollOperationMaxInterval     = 10 * time.Second
)

var (
	pollOperationRetryPolicy = createPollRetryPolicy()
)

var errShutdown = errors.New("worker shutting down")

type (
	// resultHandler that returns result
	resultHandler   func(result []byte, err error)
	laResultHandler func(lar *localActivityResultWrapper)

	localActivityResultWrapper struct {
		err     error // internal error type, possibly containing encoded user-error data
		result  []byte
		attempt int32
		backoff time.Duration
	}

	// workflowEnvironment Represents the environment for workflow/decider.
	// Should only be used within the scope of workflow definition
	workflowEnvironment interface {
		asyncActivityClient
		localActivityClient
		workflowTimerClient
		SideEffect(f func() ([]byte, error), callback resultHandler)
		GetVersion(changeID string, minSupported, maxSupported Version) Version
		WorkflowInfo() *WorkflowInfo
		Complete(result []byte, err error)
		RegisterCancelHandler(handler func())
		RequestCancelChildWorkflow(domainName, workflowID string)
		RequestCancelExternalWorkflow(domainName, workflowID, runID string, callback resultHandler)
		ExecuteChildWorkflow(params executeWorkflowParams, callback resultHandler, startedHandler func(r WorkflowExecution, e error)) error
		GetLogger() *zap.Logger
		GetMetricsScope() tally.Scope
		RegisterSignalHandler(handler func(name string, input []byte))
		SignalExternalWorkflow(domainName, workflowID, runID, signalName string, input []byte, arg interface{}, childWorkflowOnly bool, callback resultHandler)
		RegisterQueryHandler(handler func(queryType string, queryArgs []byte) ([]byte, error))
		IsReplaying() bool
		MutableSideEffect(id string, f func() interface{}, equals func(a, b interface{}) bool) Value
		GetDataConverter() DataConverter
		AddSession(sessionInfo *SessionInfo)
		RemoveSession(sessionID string)
		GetContextPropagators() []ContextPropagator
		UpsertSearchAttributes(attributes map[string]interface{}) error
		GetRegistry() *registry
		GetWorkflowInterceptors() []WorkflowInterceptorFactory
	}

	// WorkflowDefinition wraps the code that can execute a workflow.
	workflowDefinition interface {
		Execute(env workflowEnvironment, header *shared.Header, input []byte)
		// OnDecisionTaskStarted is called for each non timed out startDecision event.
		// Executed after all history events since the previous decision are applied to workflowDefinition
		OnDecisionTaskStarted()
		StackTrace() string // Stack trace of all coroutines owned by the Dispatcher instance

		// KnownQueryTypes returns a list of known query types of the workflowOptions with BuiltinQueryTypes
		KnownQueryTypes() []string
		Close()
	}

	// baseWorkerOptions options to configure base worker.
	baseWorkerOptions struct {
		pollerAutoScaler  pollerAutoScalerOptions
		pollerCount       int
		pollerRate        int
		maxConcurrentTask int
		maxTaskPerSecond  float64
		taskWorker        taskPoller
		identity          string
		workerType        string
		shutdownTimeout   time.Duration
		userContextCancel context.CancelFunc
		host              string
		pollerTracker     debug.PollerTracker
	}

	// baseWorker that wraps worker activities.
	baseWorker struct {
		options              baseWorkerOptions
		isWorkerStarted      bool
		shutdownCh           chan struct{}  // Channel used to shut down the go routines.
		shutdownWG           sync.WaitGroup // The WaitGroup for shutting down existing routines.
		pollLimiter          *rate.Limiter
		taskLimiter          *rate.Limiter
		limiterContext       context.Context
		limiterContextCancel func()
		retrier              *backoff.ConcurrentRetrier // Service errors back off retrier
		logger               *zap.Logger
		metricsScope         tally.Scope

		concurrency           *worker.ConcurrencyLimit
		concurrencyAutoScaler *worker.ConcurrencyAutoScaler
		taskQueueCh           chan interface{}
		sessionTokenBucket    *sessionTokenBucket
	}

	polledTask struct {
		task interface{}
	}
)

func createPollRetryPolicy() backoff.RetryPolicy {
	policy := backoff.NewExponentialRetryPolicy(retryPollOperationInitialInterval)
	policy.SetMaximumInterval(retryPollOperationMaxInterval)

	// NOTE: We don't use expiration interval since we don't use retries from retrier class.
	// We use it to calculate next backoff. We have additional layer that is built on poller
	// in the worker layer for to add some middleware for any poll retry that includes
	// (a) rate limiting across pollers (b) back-off across pollers when server is busy
	policy.SetExpirationInterval(backoff.NoInterval) // We don't ever expire
	return policy
}

func newBaseWorker(options baseWorkerOptions, logger *zap.Logger, metricsScope tally.Scope, sessionTokenBucket *sessionTokenBucket) *baseWorker {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.With(zapcore.Field{Key: tagWorkerType, Type: zapcore.StringType, String: options.workerType})
	metricsScope = tagScope(metricsScope, tagWorkerType, options.workerType)

	concurrency := &worker.ConcurrencyLimit{
		PollerPermit: worker.NewResizablePermit(options.pollerCount),
		TaskPermit:   worker.NewResizablePermit(options.maxConcurrentTask),
	}

	var concurrencyAS *worker.ConcurrencyAutoScaler
	if pollerOptions := options.pollerAutoScaler; pollerOptions.Enabled {
		concurrencyAS = worker.NewConcurrencyAutoScaler(worker.ConcurrencyAutoScalerInput{
			Concurrency:    concurrency,
			Cooldown:       pollerOptions.Cooldown,
			PollerMaxCount: pollerOptions.MaxCount,
			PollerMinCount: pollerOptions.MinCount,
			Logger:         logger,
			Scope:          metricsScope,
			Clock:          clockwork.NewRealClock(),
		})
	}

	bw := &baseWorker{
		options:               options,
		shutdownCh:            make(chan struct{}),
		taskLimiter:           rate.NewLimiter(rate.Limit(options.maxTaskPerSecond), 1),
		retrier:               backoff.NewConcurrentRetrier(pollOperationRetryPolicy),
		logger:                logger,
		metricsScope:          metricsScope,
		concurrency:           concurrency,
		concurrencyAutoScaler: concurrencyAS,
		taskQueueCh:           make(chan interface{}), // no buffer, so poller only able to poll new task after previous is dispatched.
		limiterContext:        ctx,
		limiterContextCancel:  cancel,
		sessionTokenBucket:    sessionTokenBucket,
	}
	if options.pollerRate > 0 {
		bw.pollLimiter = rate.NewLimiter(rate.Limit(options.pollerRate), 1)
	}
	return bw
}

// Start starts a fixed set of routines to do the work.
func (bw *baseWorker) Start() {
	if bw.isWorkerStarted {
		return
	}

	bw.metricsScope.Counter(metrics.WorkerStartCounter).Inc(1)

	if bw.concurrencyAutoScaler != nil {
		bw.concurrencyAutoScaler.Start()
	}

	for i := 0; i < bw.options.pollerCount; i++ {
		bw.shutdownWG.Add(1)
		go bw.runPoller()
	}

	bw.shutdownWG.Add(1)
	go bw.runTaskDispatcher()

	bw.isWorkerStarted = true
	traceLog(func() {
		bw.logger.Info("Started Worker",
			zap.Int("PollerCount", bw.options.pollerCount),
			zap.Int("MaxConcurrentTask", bw.options.maxConcurrentTask),
			zap.Float64("MaxTaskPerSecond", bw.options.maxTaskPerSecond),
		)
	})
}

func (bw *baseWorker) isShutdown() bool {
	select {
	case <-bw.shutdownCh:
		return true
	default:
		return false
	}
}

func (bw *baseWorker) runPoller() {
	defer bw.shutdownWG.Done()
	defer bw.options.pollerTracker.Start().Stop()

	bw.metricsScope.Counter(metrics.PollerStartCounter).Inc(1)

	for {
		permitChannel, channelDone := bw.concurrency.TaskPermit.AcquireChan(bw.limiterContext)
		select {
		case <-bw.shutdownCh:
			channelDone()
			return
		case <-permitChannel: // don't poll unless there is a task permit
			channelDone()
			// TODO move to a centralized place inside the worker
			// emit metrics on concurrent task permit quota and current task permit count
			// NOTE task permit doesn't mean there is a task running, it still needs to poll until it gets a task to process
			// thus the metrics is only an estimated value of how many tasks are running concurrently
			bw.metricsScope.Gauge(metrics.ConcurrentTaskQuota).Update(float64(bw.concurrency.TaskPermit.Quota()))
			bw.metricsScope.Gauge(metrics.PollerRequestBufferUsage).Update(float64(bw.concurrency.TaskPermit.Count()))
			if bw.sessionTokenBucket != nil {
				bw.sessionTokenBucket.waitForAvailableToken()
			}
			bw.pollTask()
		}
	}
}

func (bw *baseWorker) runTaskDispatcher() {
	defer bw.shutdownWG.Done()

	for {
		// wait for new task or shutdown
		select {
		case <-bw.shutdownCh:
			return
		case task := <-bw.taskQueueCh:
			// for non-polled-task (local activity result as task), we don't need to rate limit
			_, isPolledTask := task.(*polledTask)
			if isPolledTask && bw.taskLimiter.Wait(bw.limiterContext) != nil {
				if bw.isShutdown() {
					return
				}
			}
			bw.shutdownWG.Add(1)
			go bw.processTask(task)
		}
	}
}

/*
There are three types of constraint on polling tasks:
1. poller auto scaler is to constraint number of concurrent pollers
2. retrier is a backoff constraint on errors
3. limiter is a per-second constraint
*/
func (bw *baseWorker) pollTask() {
	var err error
	var task interface{}

	if bw.concurrencyAutoScaler != nil {
		if pErr := bw.concurrency.PollerPermit.Acquire(bw.limiterContext); pErr == nil {
			defer bw.concurrency.PollerPermit.Release()
		} else {
			bw.logger.Warn("poller permit acquire error", zap.Error(pErr))
		}
	}

	bw.retrier.Throttle()
	if bw.pollLimiter == nil || bw.pollLimiter.Wait(bw.limiterContext) == nil {
		task, err = bw.options.taskWorker.PollTask()
		if err != nil && enableVerboseLogging {
			bw.logger.Debug("Failed to poll for task.", zap.Error(err))
		}
		if err != nil {
			if isNonRetriableError(err) {
				bw.logger.Error("Worker received non-retriable error. Shutting down.", zap.Error(err))
				p, _ := os.FindProcess(os.Getpid())
				p.Signal(syscall.SIGINT)
				return
			}
			bw.retrier.Failed()
		} else {
			bw.retrier.Succeeded()
			if t, ok := task.(autoConfigHintAwareTask); bw.concurrencyAutoScaler != nil && ok {
				bw.concurrencyAutoScaler.ProcessPollerHint(t.getAutoConfigHint())
			}
		}
	}

	if task != nil {
		select {
		case bw.taskQueueCh <- &polledTask{task}:
		case <-bw.shutdownCh:
		}
	} else {
		bw.concurrency.TaskPermit.Release() // poll failed, trigger a new poll by returning a task permit
	}
}

func isNonRetriableError(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *shared.BadRequestError,
		*shared.ClientVersionNotSupportedError:
		return true
	}
	return false
}

func (bw *baseWorker) processTask(task interface{}) {
	defer bw.shutdownWG.Done()
	// If the task is from poller, after processing it we would need to request a new poll. Otherwise, the task is from
	// local activity worker, we don't need a new poll from server.
	polledTask, isPolledTask := task.(*polledTask)
	if isPolledTask {
		task = polledTask.task
	}
	defer func() {
		if p := recover(); p != nil {
			bw.metricsScope.Counter(metrics.WorkerPanicCounter).Inc(1)
			topLine := fmt.Sprintf("base worker for %s [panic]:", bw.options.workerType)
			st := getStackTraceRaw(topLine, 7, 0)
			bw.logger.Error("Unhandled panic.",
				zap.String(tagPanicError, fmt.Sprintf("%v", p)),
				zap.String(tagPanicStack, st))
		}

		if isPolledTask {
			bw.concurrency.TaskPermit.Release() // task processed, trigger a new poll by returning a task permit
		}
	}()
	err := bw.options.taskWorker.ProcessTask(task)
	if err != nil {
		if isClientSideError(err) {
			bw.logger.Info("Task processing failed with client side error", zap.Error(err))
		} else {
			bw.logger.Info("Task processing failed with error", zap.Error(err))
		}
	}
}

func (bw *baseWorker) Run() {
	bw.Start()
	d := <-getKillSignal()
	traceLog(func() {
		bw.logger.Info("Worker has been killed", zap.String("Signal", d.String()))
	})
	bw.Stop()
}

// Stop is a blocking call and cleans up all the resources associated with worker.
func (bw *baseWorker) Stop() {
	if !bw.isWorkerStarted {
		return
	}
	close(bw.shutdownCh)
	bw.limiterContextCancel()
	if bw.concurrencyAutoScaler != nil {
		bw.concurrencyAutoScaler.Stop()
	}

	if success := util.AwaitWaitGroup(&bw.shutdownWG, bw.options.shutdownTimeout); !success {
		traceLog(func() {
			bw.logger.Info("Worker graceful shutdown timed out.", zap.Duration("Shutdown timeout", bw.options.shutdownTimeout))
		})
	}

	// Close context
	if bw.options.userContextCancel != nil {
		bw.options.userContextCancel()
	}
	return
}

func getAutoConfigHint(task interface{}) *shared.AutoConfigHint {
	switch t := task.(type) {
	case *workflowTask:
		if t.task == nil {
			return nil
		}
		return t.task.AutoConfigHint
	case *activityTask:
		if t.task == nil {
			return nil
		}
		return t.task.AutoConfigHint
	default:
		return nil
	}
}
