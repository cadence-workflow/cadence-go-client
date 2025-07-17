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
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/cadence/internal/common/debug"

	"go.uber.org/cadence/internal/common/isolationgroup"

	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/auth"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/metrics"
	"go.uber.org/cadence/internal/common/util"
)

var startVersionMetric sync.Once
var StopMetrics = make(chan struct{})

const (
	// Set to 2 pollers for now, can adjust later if needed. The typical RTT (round-trip time) is below 1ms within data
	// center. And the poll API latency is about 5ms. With 2 poller, we could achieve around 300~400 RPS.
	defaultConcurrentPollRoutineSize = 2

	defaultMaxConcurrentActivityExecutionSize = 1000   // Large concurrent activity execution size (1k)
	defaultWorkerActivitiesPerSecond          = 100000 // Large activity executions/sec (unlimited)

	defaultMaxConcurrentLocalActivityExecutionSize = 1000   // Large concurrent activity execution size (1k)
	defaultWorkerLocalActivitiesPerSecond          = 100000 // Large activity executions/sec (unlimited)

	defaultTaskListActivitiesPerSecond = 100000.0 // Large activity executions/sec (unlimited)

	defaultMaxConcurrentTaskExecutionSize = 1000   // hardcoded max task execution size.
	defaultWorkerTaskExecutionRate        = 100000 // Large task execution rate (unlimited)

	defaultPollerRate = 1000

	defaultMaxConcurrentSessionExecutionSize = 1000 // Large concurrent session execution size (1k)

	testTagsContextKey = "cadence-testTags"
	clientVersionTag   = "cadence_client_version"
	clientGauge        = "client_version_metric"
)

type (
	// WorkflowWorker wraps the code for hosting workflow types.
	// And worker is mapped 1:1 with task list. If the user want's to poll multiple
	// task list names they might have to manage 'n' workers for 'n' task lists.
	workflowWorker struct {
		executionParameters workerExecutionParameters
		workflowService     workflowserviceclient.Interface
		domain              string
		poller              taskPoller // taskPoller to poll and process the tasks.
		worker              *baseWorker
		localActivityWorker *baseWorker
		identity            string
		stopC               chan struct{}
	}

	// ActivityWorker wraps the code for hosting activity types.
	// TODO: Worker doing heartbeating automatically while activity task is running
	activityWorker struct {
		executionParameters workerExecutionParameters
		workflowService     workflowserviceclient.Interface
		domain              string
		poller              taskPoller
		worker              *baseWorker
		identity            string
		stopC               chan struct{}
	}

	// sessionWorker wraps the code for hosting session creation, completion and
	// activities within a session. The creationWorker polls from a global tasklist,
	// while the activityWorker polls from a resource specific tasklist.
	sessionWorker struct {
		creationWorker *activityWorker
		activityWorker *activityWorker
	}

	// Worker overrides.
	workerOverrides struct {
		workflowTaskHandler                WorkflowTaskHandler
		activityTaskHandler                ActivityTaskHandler
		useLocallyDispatchedActivityPoller bool
	}

	// workerExecutionParameters defines worker configure/execution options.
	workerExecutionParameters struct {
		WorkerOptions

		// Task list name to poll.
		TaskList string

		// Context to store user provided key/value pairs
		UserContext context.Context

		// Context cancel function to cancel user context
		UserContextCancel context.CancelFunc

		// WorkerStopTimeout is the time delay before hard terminate worker
		WorkerStopTimeout time.Duration

		// WorkerStopChannel is a read only channel listen on worker close. The worker will close the channel before exit.
		WorkerStopChannel <-chan struct{}

		// SessionResourceID is a unique identifier of the resource the session will consume
		SessionResourceID string
	}
)

// newWorkflowWorker returns an instance of the workflow worker.
func newWorkflowWorker(
	service workflowserviceclient.Interface,
	domain string,
	params workerExecutionParameters,
	ppMgr pressurePointMgr,
	registry *registry,
	ldaTunnel *locallyDispatchedActivityTunnel,
) *workflowWorker {
	return newWorkflowWorkerInternal(service, domain, params, ppMgr, nil, registry, ldaTunnel)
}

func ensureRequiredParams(params *workerExecutionParameters) {
	if params.Tracer == nil {
		params.Tracer = opentracing.NoopTracer{}
	}
	if params.Identity == "" {
		params.Identity = getWorkerIdentity(params.TaskList)
	}
	if params.Logger == nil {
		// create default logger if user does not supply one.
		config := zap.NewProductionConfig()
		// set default time formatter to "2006-01-02T15:04:05.000Z0700"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// config.Level.SetLevel(zapcore.DebugLevel)
		logger, _ := config.Build()
		params.Logger = logger
		params.Logger.Info("No logger configured for cadence worker. Created default one.")
	}
	if params.MetricsScope == nil {
		params.MetricsScope = tally.NoopScope
		params.Logger.Info("No metrics scope configured for cadence worker. Use NoopScope as default.")
	}
	if params.DataConverter == nil {
		params.DataConverter = getDefaultDataConverter()
		params.Logger.Info("No DataConverter configured for cadence worker. Use default one.")
	}
	if params.UserContext == nil {
		params.UserContext = context.Background()
	}
	if params.WorkerStats.PollerTracker == nil {
		params.WorkerStats.PollerTracker = debug.NewNoopPollerTracker()
		params.Logger.Debug("No PollerTracker configured for WorkerStats option. Will use the default.")
	}
	if params.WorkerStats.ActivityTracker == nil {
		params.WorkerStats.ActivityTracker = debug.NewNoopActivityTracker()
		params.Logger.Debug("No ActivityTracker configured for WorkerStats option. Will use the default.")
	}
}

// verifyDomainExist does a DescribeDomain operation on the specified domain with backoff/retry
// It returns an error, if the server returns an EntityNotExist or BadRequest error
// On any other transient error, this method will just return success
func verifyDomainExist(
	client workflowserviceclient.Interface,
	domain string,
	logger *zap.Logger,
	featureFlags FeatureFlags,
) error {
	ctx := context.Background()
	descDomainOp := func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
		defer cancel()
		_, err := client.DescribeDomain(tchCtx, &shared.DescribeDomainRequest{Name: &domain}, opt...)
		if err != nil {
			if _, ok := err.(*shared.EntityNotExistsError); ok {
				logger.Error("domain does not exist", zap.String("domain", domain), zap.Error(err))
				return err
			}
			if _, ok := err.(*shared.BadRequestError); ok {
				logger.Error("domain does not exist", zap.String("domain", domain), zap.Error(err))
				return err
			}
			// on any other error, just return true
			logger.Warn("unable to verify if domain exist", zap.String("domain", domain), zap.Error(err))
		}
		return nil
	}

	if len(domain) == 0 {
		return errors.New("domain cannot be empty")
	}

	// exponential backoff retry for upto a minute
	return backoff.Retry(ctx, descDomainOp, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
}

func newWorkflowWorkerInternal(
	service workflowserviceclient.Interface,
	domain string,
	params workerExecutionParameters,
	ppMgr pressurePointMgr,
	overrides *workerOverrides,
	registry *registry,
	ldaTunnel *locallyDispatchedActivityTunnel,
) *workflowWorker {
	workerStopChannel := make(chan struct{})
	params.WorkerStopChannel = getReadOnlyChannel(workerStopChannel)
	// Get a workflow task handler.
	ensureRequiredParams(&params)
	var taskHandler WorkflowTaskHandler
	if overrides != nil && overrides.workflowTaskHandler != nil {
		taskHandler = overrides.workflowTaskHandler
	} else {
		taskHandler = newWorkflowTaskHandler(domain, params, ppMgr, registry)
	}
	return newWorkflowTaskWorkerInternal(taskHandler, service, domain, params, workerStopChannel, ldaTunnel)
}

func newWorkflowTaskWorkerInternal(
	taskHandler WorkflowTaskHandler,
	service workflowserviceclient.Interface,
	domain string,
	params workerExecutionParameters,
	stopC chan struct{},
	ldaTunnel *locallyDispatchedActivityTunnel,
) *workflowWorker {
	ensureRequiredParams(&params)
	poller := newWorkflowTaskPoller(
		taskHandler,
		ldaTunnel,
		service,
		domain,
		params,
	)
	worker := newBaseWorker(baseWorkerOptions{
		pollerAutoScaler:              params.AutoScalerOptions,
		pollerCountWithoutAutoScaling: params.MaxConcurrentDecisionTaskPollers,
		pollerRate:                    defaultPollerRate,
		maxConcurrentTask:             params.MaxConcurrentDecisionTaskExecutionSize,
		maxTaskPerSecond:              params.WorkerDecisionTasksPerSecond,
		taskWorker:                    poller,
		identity:                      params.Identity,
		workerType:                    "DecisionWorker",
		shutdownTimeout:               params.WorkerStopTimeout,
		pollerTracker:                 params.WorkerStats.PollerTracker,
	},
		params.Logger,
		params.MetricsScope,
		nil,
	)

	// laTunnel is the glue that hookup 3 parts
	laTunnel := newLocalActivityTunnel(params.WorkerStopChannel)

	// 1) workflow handler will send local activity task to laTunnel
	if handlerImpl, ok := taskHandler.(*workflowTaskHandlerImpl); ok {
		handlerImpl.laTunnel = laTunnel
	}

	// 2) local activity task poller will poll from laTunnel, and result will be pushed to laTunnel
	localActivityTaskPoller := newLocalActivityPoller(params, laTunnel)
	localActivityWorker := newBaseWorker(baseWorkerOptions{
		pollerCountWithoutAutoScaling: 1, // 1 poller (from local channel) is enough for local activity
		maxConcurrentTask:             params.MaxConcurrentLocalActivityExecutionSize,
		maxTaskPerSecond:              params.WorkerLocalActivitiesPerSecond,
		taskWorker:                    localActivityTaskPoller,
		identity:                      params.Identity,
		workerType:                    "LocalActivityWorker",
		shutdownTimeout:               params.WorkerStopTimeout,
		pollerTracker:                 params.WorkerStats.PollerTracker,
	},
		params.Logger,
		params.MetricsScope,
		nil,
	)

	// 3) the result pushed to laTunnel will be send as task to workflow worker to process.
	worker.taskQueueCh = laTunnel.resultCh

	return &workflowWorker{
		executionParameters: params,
		workflowService:     service,
		poller:              poller,
		worker:              worker,
		localActivityWorker: localActivityWorker,
		identity:            params.Identity,
		domain:              domain,
		stopC:               stopC,
	}
}

// Start the worker.
func (ww *workflowWorker) Start() error {
	err := verifyDomainExist(ww.workflowService, ww.domain, ww.worker.logger, ww.executionParameters.FeatureFlags)
	if err != nil {
		return err
	}
	ww.localActivityWorker.Start()
	ww.worker.Start()
	return nil // TODO: propagate error
}

func (ww *workflowWorker) Run() error {
	err := verifyDomainExist(ww.workflowService, ww.domain, ww.worker.logger, ww.executionParameters.FeatureFlags)
	if err != nil {
		return err
	}
	ww.localActivityWorker.Start()
	ww.worker.Run()
	return nil
}

// Shutdown the worker.
func (ww *workflowWorker) Stop() {
	select {
	case <-ww.stopC:
		// channel is already closed
	default:
		close(ww.stopC)
	}
	// TODO: remove the stop methods in favor of the workerStopChannel
	ww.localActivityWorker.Stop()
	ww.worker.Stop()
}

func newSessionWorker(service workflowserviceclient.Interface,
	domain string,
	params workerExecutionParameters,
	overrides *workerOverrides,
	env *registry,
	maxConcurrentSessionExecutionSize int,
) *sessionWorker {
	ensureRequiredParams(&params)
	// For now resourceID is hidden from user so we will always create a unique one for each worker.
	if params.SessionResourceID == "" {
		params.SessionResourceID = uuid.New()
	}
	sessionEnvironment := newSessionEnvironment(params.SessionResourceID, maxConcurrentSessionExecutionSize)

	creationTasklist := getCreationTasklist(params.TaskList)
	params.UserContext = context.WithValue(params.UserContext, sessionEnvironmentContextKey, sessionEnvironment)
	params.TaskList = sessionEnvironment.GetResourceSpecificTasklist()
	activityWorker := newActivityWorker(service, domain, params, overrides, env, nil)

	params.MaxConcurrentActivityTaskPollers = 1
	params.TaskList = creationTasklist
	creationWorker := newActivityWorker(service, domain, params, overrides, env, sessionEnvironment.GetTokenBucket())

	return &sessionWorker{
		creationWorker: creationWorker,
		activityWorker: activityWorker,
	}
}

func (sw *sessionWorker) Start() error {
	err := sw.creationWorker.Start()
	if err != nil {
		return err
	}

	err = sw.activityWorker.Start()
	if err != nil {
		sw.creationWorker.Stop()
		return err
	}
	return nil
}

func (sw *sessionWorker) Run() error {
	err := sw.creationWorker.Start()
	if err != nil {
		return err
	}
	return sw.activityWorker.Run()
}

func (sw *sessionWorker) Stop() {
	sw.creationWorker.Stop()
	sw.activityWorker.Stop()
}

func newActivityWorker(
	service workflowserviceclient.Interface,
	domain string,
	params workerExecutionParameters,
	overrides *workerOverrides,
	env *registry,
	sessionTokenBucket *sessionTokenBucket,
) *activityWorker {
	workerStopChannel := make(chan struct{}, 1)
	params.WorkerStopChannel = getReadOnlyChannel(workerStopChannel)
	ensureRequiredParams(&params)
	workerType := "ActivityWorker"
	// Get a activity task handler.
	var taskHandler ActivityTaskHandler
	if overrides != nil && overrides.activityTaskHandler != nil {
		taskHandler = overrides.activityTaskHandler
	} else {
		taskHandler = newActivityTaskHandler(service, params, env)
	}
	// Get an activity task poller.
	var taskPoller taskPoller
	if overrides != nil && overrides.useLocallyDispatchedActivityPoller {
		taskPoller = newLocallyDispatchedActivityTaskPoller(taskHandler, service, domain, params)
		workerType = "LocallyDispatchedActivityWorker"
	} else {
		taskPoller = newActivityTaskPoller(
			taskHandler,
			service,
			domain,
			params,
		)
	}
	return newActivityTaskWorker(service, domain, params, sessionTokenBucket, workerStopChannel, taskPoller, workerType)
}

func newActivityTaskWorker(
	service workflowserviceclient.Interface,
	domain string,
	workerParams workerExecutionParameters,
	sessionTokenBucket *sessionTokenBucket,
	stopC chan struct{},
	poller taskPoller,
	workerType string,
) (worker *activityWorker) {
	ensureRequiredParams(&workerParams)
	base := newBaseWorker(
		baseWorkerOptions{
			pollerAutoScaler:              workerParams.AutoScalerOptions,
			pollerCountWithoutAutoScaling: workerParams.MaxConcurrentActivityTaskPollers,
			pollerRate:                    defaultPollerRate,
			maxConcurrentTask:             workerParams.MaxConcurrentActivityExecutionSize,
			maxTaskPerSecond:              workerParams.WorkerActivitiesPerSecond,
			taskWorker:                    poller,
			identity:                      workerParams.Identity,
			workerType:                    workerType,
			shutdownTimeout:               workerParams.WorkerStopTimeout,
			userContextCancel:             workerParams.UserContextCancel,
			pollerTracker:                 workerParams.WorkerStats.PollerTracker,
		},

		workerParams.Logger,
		workerParams.MetricsScope,
		sessionTokenBucket,
	)

	return &activityWorker{
		executionParameters: workerParams,
		workflowService:     service,
		worker:              base,
		poller:              poller,
		identity:            workerParams.Identity,
		domain:              domain,
		stopC:               stopC,
	}
}

// Start the worker.
func (aw *activityWorker) Start() error {
	err := verifyDomainExist(aw.workflowService, aw.domain, aw.worker.logger, aw.executionParameters.FeatureFlags)
	if err != nil {
		return err
	}
	aw.worker.Start()
	return nil // TODO: propagate errors
}

// Run the worker.
func (aw *activityWorker) Run() error {
	err := verifyDomainExist(aw.workflowService, aw.domain, aw.worker.logger, aw.executionParameters.FeatureFlags)
	if err != nil {
		return err
	}
	aw.worker.Run()
	return nil
}

// Shutdown the worker.
func (aw *activityWorker) Stop() {
	select {
	case <-aw.stopC:
		// channel is already closed
	default:
		close(aw.stopC)
	}
	aw.worker.Stop()
}

// Validate function parameters.
func validateFnFormat(fnType reflect.Type, isWorkflow bool) error {
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("expected a func as input but was %s", fnType.Kind())
	}
	if isWorkflow {
		if fnType.NumIn() < 1 {
			return fmt.Errorf(
				"expected at least one argument of type workflow.Context in function, found %d input arguments",
				fnType.NumIn(),
			)
		}
		if !isWorkflowContext(fnType.In(0)) {
			return fmt.Errorf("expected first argument to be workflow.Context but found %s", fnType.In(0))
		}
	}

	// Return values
	// We expect either
	// 	<result>, error
	//	(or) just error
	if fnType.NumOut() < 1 || fnType.NumOut() > 2 {
		return fmt.Errorf(
			"expected function to return result, error or just error, but found %d return values", fnType.NumOut(),
		)
	}
	if fnType.NumOut() > 1 && !isValidResultType(fnType.Out(0)) {
		return fmt.Errorf(
			"expected function first return value to return valid type but found: %v", fnType.Out(0).Kind(),
		)
	}
	if !isError(fnType.Out(fnType.NumOut() - 1)) {
		return fmt.Errorf(
			"expected function second return value to return error but found %v", fnType.Out(fnType.NumOut()-1).Kind(),
		)
	}
	return nil
}

// encode multiple arguments(arguments to a function).
func encodeArgs(dc DataConverter, args []interface{}) ([]byte, error) {
	if dc == nil {
		return getDefaultDataConverter().ToData(args...)
	}
	return dc.ToData(args...)
}

// decode multiple arguments(arguments to a function).
func decodeArgs(dc DataConverter, fnType reflect.Type, data []byte) (result []reflect.Value, err error) {
	r, err := decodeArgsToValues(dc, fnType, data)
	if err != nil {
		return
	}
	for i := 0; i < len(r); i++ {
		result = append(result, reflect.ValueOf(r[i]).Elem())
	}
	return
}

func decodeArgsToValues(dc DataConverter, fnType reflect.Type, data []byte) (result []interface{}, err error) {
	if dc == nil {
		dc = getDefaultDataConverter()
	}
argsLoop:
	for i := 0; i < fnType.NumIn(); i++ {
		argT := fnType.In(i)
		if i == 0 && (isActivityContext(argT) || isWorkflowContext(argT)) {
			continue argsLoop
		}
		arg := reflect.New(argT).Interface()
		result = append(result, arg)
	}
	err = dc.FromData(data, result...)
	if err != nil {
		return
	}
	return
}

// encode single value(like return parameter).
func encodeArg(dc DataConverter, arg interface{}) ([]byte, error) {
	if dc == nil {
		return getDefaultDataConverter().ToData(arg)
	}
	return dc.ToData(arg)
}

// decode single value(like return parameter).
func decodeArg(dc DataConverter, data []byte, to interface{}) error {
	if dc == nil {
		return getDefaultDataConverter().FromData(data, to)
	}
	return dc.FromData(data, to)
}

func decodeAndAssignValue(dc DataConverter, from interface{}, toValuePtr interface{}) error {
	if toValuePtr == nil {
		return nil
	}
	if rf := reflect.ValueOf(toValuePtr); rf.Type().Kind() != reflect.Ptr {
		return errors.New("value parameter provided is not a pointer")
	}
	if data, ok := from.([]byte); ok {
		if err := decodeArg(dc, data, toValuePtr); err != nil {
			return err
		}
	} else if fv := reflect.ValueOf(from); fv.IsValid() {
		fromType := fv.Type()
		toType := reflect.TypeOf(toValuePtr).Elem()
		assignable := fromType.AssignableTo(toType)
		if !assignable {
			return fmt.Errorf("%s is not assignable to  %s", fromType.Name(), toType.Name())
		}
		reflect.ValueOf(toValuePtr).Elem().Set(fv)
	}
	return nil
}

// Wrapper to execute workflow functions.
type workflowExecutor struct {
	workflowType string
	fn           interface{}
	path         string
}

func (we *workflowExecutor) Execute(ctx Context, input []byte) ([]byte, error) {
	var args []interface{}
	dataConverter := getWorkflowEnvOptions(ctx).dataConverter
	fnType := reflect.TypeOf(we.fn)
	if fnType.NumIn() == 2 && util.IsTypeByteSlice(fnType.In(1)) {
		// Do not deserialize input if workflow has a single byte slice argument (besides ctx)
		args = append(args, input)
	} else {
		decoded, err := decodeArgsToValues(dataConverter, fnType, input)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to decode the workflow function input bytes with error: %v, function name: %v",
				err, we.workflowType)
		}
		args = append(args, decoded...)
	}
	envInterceptor := getEnvInterceptor(ctx)
	envInterceptor.fn = we.fn
	results := envInterceptor.interceptorChainHead.ExecuteWorkflow(ctx, we.workflowType, args...)
	return serializeResults(we.fn, results, dataConverter)
}

func (we *workflowExecutor) WorkflowType() WorkflowType {
	return WorkflowType{
		Name: we.workflowType,
		Path: we.path,
	}
}

func (we *workflowExecutor) GetFunction() interface{} {
	return we.fn
}

// Wrapper to execute activity functions.
type activityExecutor struct {
	name    string
	fn      interface{}
	options RegisterActivityOptions
	path    string
}

func (ae *activityExecutor) ActivityType() ActivityType {
	return ActivityType{Name: ae.name, Path: ae.path}
}

func (ae *activityExecutor) GetFunction() interface{} {
	return ae.fn
}

func (ae *activityExecutor) GetOptions() RegisterActivityOptions {
	return ae.options
}

func (ae *activityExecutor) Execute(ctx context.Context, input []byte) ([]byte, error) {
	fnType := reflect.TypeOf(ae.fn)
	var args []reflect.Value
	dataConverter := getDataConverterFromActivityCtx(ctx)

	// activities optionally might not take context.
	if fnType.NumIn() > 0 && isActivityContext(fnType.In(0)) {
		args = append(args, reflect.ValueOf(ctx))
	}

	if fnType.NumIn() == 1 && util.IsTypeByteSlice(fnType.In(0)) {
		args = append(args, reflect.ValueOf(input))
	} else {
		decoded, err := decodeArgs(dataConverter, fnType, input)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to decode the activity function input bytes with error: %v for function name: %v",
				err, ae.name)
		}
		args = append(args, decoded...)
	}

	fnValue := reflect.ValueOf(ae.fn)
	retValues := fnValue.Call(args)
	return validateFunctionAndGetResults(ae.fn, retValues, dataConverter)
}

func (ae *activityExecutor) ExecuteWithActualArgs(ctx context.Context, actualArgs []interface{}) ([]byte, error) {
	retValues := ae.executeWithActualArgsWithoutParseResult(ctx, actualArgs)
	dataConverter := getDataConverterFromActivityCtx(ctx)

	return validateFunctionAndGetResults(ae.fn, retValues, dataConverter)
}

func (ae *activityExecutor) executeWithActualArgsWithoutParseResult(ctx context.Context, actualArgs []interface{}) []reflect.Value {
	fnType := reflect.TypeOf(ae.fn)
	args := []reflect.Value{}

	// activities optionally might not take context.
	argsOffeset := 0
	if fnType.NumIn() > 0 && isActivityContext(fnType.In(0)) {
		args = append(args, reflect.ValueOf(ctx))
		argsOffeset = 1
	}

	for i, arg := range actualArgs {
		if arg == nil {
			args = append(args, reflect.New(fnType.In(i+argsOffeset)).Elem())
		} else {
			args = append(args, reflect.ValueOf(arg))
		}
	}

	fnValue := reflect.ValueOf(ae.fn)
	retValues := fnValue.Call(args)
	return retValues
}

func getDataConverterFromActivityCtx(ctx context.Context) DataConverter {
	if ctx == nil || ctx.Value(activityEnvContextKey) == nil {
		return getDefaultDataConverter()
	}
	info := ctx.Value(activityEnvContextKey).(*activityEnvironment)
	if info.dataConverter == nil {
		return getDefaultDataConverter()
	}
	return info.dataConverter
}

// aggregatedWorker combines management of both workflowWorker and activityWorker worker lifecycle.
type aggregatedWorker struct {
	workflowWorker                  *workflowWorker
	activityWorker                  *activityWorker
	locallyDispatchedActivityWorker *activityWorker
	sessionWorker                   *sessionWorker
	shadowWorker                    *shadowWorker
	logger                          *zap.Logger
	registry                        *registry
	workerstats                     debug.WorkerStats
}

var _ debug.Debugger = &aggregatedWorker{}

func (aw *aggregatedWorker) GetRegisteredWorkflows() []RegistryWorkflowInfo {
	workflows := aw.registry.GetRegisteredWorkflows()
	var result []RegistryWorkflowInfo
	for _, wf := range workflows {
		result = append(result, wf)
	}
	return result
}

func (aw *aggregatedWorker) GetRegisteredActivities() []RegistryActivityInfo {
	activities := aw.registry.getRegisteredActivities()
	var result []RegistryActivityInfo
	for _, a := range activities {
		result = append(result, a)
	}
	return result
}

func (aw *aggregatedWorker) RegisterWorkflow(w interface{}) {
	aw.registry.RegisterWorkflow(w)
}

func (aw *aggregatedWorker) RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	aw.registry.RegisterWorkflowWithOptions(w, options)
}

func (aw *aggregatedWorker) RegisterActivity(a interface{}) {
	aw.registry.RegisterActivity(a)
}

func (aw *aggregatedWorker) RegisterActivityWithOptions(a interface{}, options RegisterActivityOptions) {
	aw.registry.RegisterActivityWithOptions(a, options)
}

func (aw *aggregatedWorker) Start() error {
	if _, err := initBinaryChecksum(); err != nil {
		return fmt.Errorf("failed to get executable checksum: %v", err)
	}

	if aw.workflowWorker != nil {
		if len(aw.registry.GetRegisteredWorkflowTypes()) == 0 {
			aw.logger.Info(
				"Worker has no workflows registered, so workflow worker will not be started.",
			)
		} else {
			if err := aw.workflowWorker.Start(); err != nil {
				return err
			}
		}
		aw.logger.Info("Started Workflow Worker")
	}
	if aw.activityWorker != nil {
		if len(aw.registry.getRegisteredActivities()) == 0 {
			aw.logger.Info(
				"Worker has no activities registered, so activity worker will not be started.",
			)
		} else {
			if err := aw.activityWorker.Start(); err != nil {
				// stop workflow worker.
				if aw.workflowWorker != nil {
					aw.workflowWorker.Stop()
				}
				return err
			}
			if aw.locallyDispatchedActivityWorker != nil {
				if err := aw.locallyDispatchedActivityWorker.Start(); err != nil {
					// stop workflow worker.
					if aw.workflowWorker != nil {
						aw.workflowWorker.Stop()
					}
					aw.activityWorker.Stop()
					return err
				}
			}
			aw.logger.Info("Started Activity Worker")
		}
	}

	if aw.sessionWorker != nil {
		if err := aw.sessionWorker.Start(); err != nil {
			// stop workflow worker and activity worker.
			if aw.workflowWorker != nil {
				aw.workflowWorker.Stop()
			}
			if aw.activityWorker != nil {
				aw.activityWorker.Stop()
			}
			if aw.locallyDispatchedActivityWorker != nil {
				aw.locallyDispatchedActivityWorker.Stop()
			}
			return err
		}
		aw.logger.Info("Started Session Worker")
	}

	if aw.shadowWorker != nil {
		if err := aw.shadowWorker.Start(); err != nil {
			if aw.workflowWorker != nil {
				aw.workflowWorker.Stop()
			}
			if aw.activityWorker != nil {
				aw.activityWorker.Stop()
			}
			if aw.locallyDispatchedActivityWorker != nil {
				aw.locallyDispatchedActivityWorker.Stop()
			}
			if aw.sessionWorker != nil {
				aw.sessionWorker.Stop()
			}
			return err
		}
		aw.logger.Info("Started Shadow Worker")
	}

	return nil
}

var binaryChecksum atomic.Value
var binaryChecksumLock sync.Mutex

// SetBinaryChecksum set binary checksum
func SetBinaryChecksum(checksum string) {
	binaryChecksum.Store(checksum)
}

func initBinaryChecksum() (string, error) {
	// initBinaryChecksum may be called multiple times concurrently during worker startup.
	// To avoid reading and hashing the contents of the binary multiple times acquire mutex here.
	binaryChecksumLock.Lock()
	defer binaryChecksumLock.Unlock()

	// check if binaryChecksum already set/initialized.
	if bcsVal, ok := binaryChecksum.Load().(string); ok {
		return bcsVal, nil
	}

	exec, err := os.Executable()
	if err != nil {
		return "", err
	}

	f, err := os.Open(exec)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close() // error is unimportant as it is read-only
	}()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	checksum := h.Sum(nil)
	bcsVal := hex.EncodeToString(checksum[:])
	binaryChecksum.Store(bcsVal)

	return bcsVal, err
}

func getBinaryChecksum() string {
	bcsVal, ok := binaryChecksum.Load().(string)
	if ok {
		return bcsVal
	}

	bcsVal, err := initBinaryChecksum()
	if err != nil {
		panic(err)
	}

	return bcsVal
}

func (aw *aggregatedWorker) Run() error {
	if err := aw.Start(); err != nil {
		return err
	}
	d := <-getKillSignal()
	aw.logger.Info("Worker has been killed", zap.String("Signal", d.String()))
	aw.Stop()
	return nil
}

func (aw *aggregatedWorker) Stop() {
	if aw.workflowWorker != nil {
		aw.workflowWorker.Stop()
	}
	if aw.activityWorker != nil {
		aw.activityWorker.Stop()
	}
	if aw.locallyDispatchedActivityWorker != nil {
		aw.locallyDispatchedActivityWorker.Stop()
	}
	if aw.sessionWorker != nil {
		aw.sessionWorker.Stop()
	}
	if aw.shadowWorker != nil {
		aw.shadowWorker.Stop()
	}
	aw.logger.Info("Stopped Worker")
}

func (aw *aggregatedWorker) GetWorkerStats() debug.WorkerStats {
	return aw.workerstats
}

// AggregatedWorker returns an instance to manage the workers. Use defaultConcurrentPollRoutineSize (which is 2) as
// poller size. The typical RTT (round-trip time) is below 1ms within data center. And the poll API latency is about 5ms.
// With 2 poller, we could achieve around 300~400 RPS.
func newAggregatedWorker(
	service workflowserviceclient.Interface,
	domain string,
	taskList string,
	options WorkerOptions,
) (worker *aggregatedWorker, err error) {
	wOptions := AugmentWorkerOptions(options)
	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("worker options validation error: %w", err)
	}

	ctx := wOptions.BackgroundActivityContext
	if ctx == nil {
		ctx = context.Background()
	}
	backgroundActivityContext, backgroundActivityContextCancel := context.WithCancel(ctx)

	workerParams := workerExecutionParameters{
		WorkerOptions:     wOptions,
		TaskList:          taskList,
		UserContext:       backgroundActivityContext,
		UserContextCancel: backgroundActivityContextCancel,
	}

	ensureRequiredParams(&workerParams)
	workerParams.MetricsScope = tagScope(workerParams.MetricsScope, tagDomain, domain, tagTaskList, taskList, clientImplHeaderName, clientImplHeaderValue)
	workerParams.Logger = workerParams.Logger.With(
		zapcore.Field{Key: tagDomain, Type: zapcore.StringType, String: domain},
		zapcore.Field{Key: tagTaskList, Type: zapcore.StringType, String: taskList},
		zapcore.Field{Key: tagWorkerID, Type: zapcore.StringType, String: workerParams.Identity},
	)
	logger := workerParams.Logger
	if options.Authorization != nil {
		service = auth.NewWorkflowServiceWrapper(service, options.Authorization)
	}
	if options.IsolationGroup != "" {
		service = isolationgroup.NewWorkflowServiceWrapper(service, options.IsolationGroup)
	}
	service = metrics.NewWorkflowServiceWrapper(service, workerParams.MetricsScope)
	processTestTags(&wOptions, &workerParams)

	// worker specific registry
	registry := newRegistry()

	// ldaTunnel is a one way tunnel to dispatch activity tasks from workflow poller to activity poller
	var ldaTunnel *locallyDispatchedActivityTunnel

	// activity types.
	var activityWorker, locallyDispatchedActivityWorker *activityWorker

	if !wOptions.DisableActivityWorker {
		activityWorker = newActivityWorker(
			service,
			domain,
			workerParams,
			nil,
			registry,
			nil,
		)

		// do not dispatch locally if TaskListActivitiesPerSecond is set
		if workerParams.TaskListActivitiesPerSecond == defaultTaskListActivitiesPerSecond {
			// TODO update taskPoller interface so one activity worker can multiplex on multiple pollers
			locallyDispatchedActivityWorker = newActivityWorker(
				service,
				domain,
				workerParams,
				&workerOverrides{useLocallyDispatchedActivityPoller: true},
				registry,
				nil,
			)
			ldaTunnel = locallyDispatchedActivityWorker.poller.(*locallyDispatchedActivityTaskPoller).ldaTunnel
			ldaTunnel.metricsScope = metrics.NewTaggedScope(workerParams.MetricsScope)
		}
	}

	// workflow factory.
	var workflowWorker *workflowWorker
	if !wOptions.DisableWorkflowWorker {
		testTags := getTestTags(wOptions.BackgroundActivityContext)
		if len(testTags) > 0 {
			workflowWorker = newWorkflowWorkerWithPressurePoints(
				service,
				domain,
				workerParams,
				testTags,
				registry,
			)
		} else {
			workflowWorker = newWorkflowWorker(
				service,
				domain,
				workerParams,
				nil,
				registry,
				ldaTunnel,
			)
		}

	}

	var sessionWorker *sessionWorker
	if wOptions.EnableSessionWorker {
		sessionWorker = newSessionWorker(
			service,
			domain,
			workerParams,
			nil,
			registry,
			wOptions.MaxConcurrentSessionExecutionSize,
		)
		registry.RegisterActivityWithOptions(sessionCreationActivity, RegisterActivityOptions{
			Name: sessionCreationActivityName,
		})
		registry.RegisterActivityWithOptions(sessionCompletionActivity, RegisterActivityOptions{
			Name: sessionCompletionActivityName,
		})

	}

	var shadowWorker *shadowWorker
	if wOptions.EnableShadowWorker {
		shadowWorker = newShadowWorker(
			service,
			domain,
			wOptions.ShadowOptions,
			workerParams,
			registry,
		)
	}

	return &aggregatedWorker{
		workflowWorker:                  workflowWorker,
		activityWorker:                  activityWorker,
		locallyDispatchedActivityWorker: locallyDispatchedActivityWorker,
		sessionWorker:                   sessionWorker,
		shadowWorker:                    shadowWorker,
		logger:                          logger,
		registry:                        registry,
		workerstats:                     workerParams.WorkerStats,
	}, nil
}

// tagScope with one or multiple tags, like
// tagScope(scope, tag1, val1, tag2, val2)
func tagScope(metricsScope tally.Scope, keyValueinPairs ...string) tally.Scope {
	if metricsScope == nil {
		metricsScope = tally.NoopScope
	}
	if len(keyValueinPairs)%2 != 0 {
		panic("tagScope key value are not in pairs")
	}
	tagsMap := map[string]string{}
	for i := 0; i < len(keyValueinPairs); i += 2 {
		tagsMap[keyValueinPairs[i]] = keyValueinPairs[i+1]
	}
	return metricsScope.Tagged(tagsMap)
}

func processTestTags(wOptions *WorkerOptions, ep *workerExecutionParameters) {
	testTags := getTestTags(wOptions.BackgroundActivityContext)
	if testTags != nil {
		if paramsOverride, ok := testTags[workerOptionsConfig]; ok {
			for key, val := range paramsOverride {
				switch key {
				case workerOptionsConfigConcurrentPollRoutineSize:
					if size, err := strconv.Atoi(val); err == nil {
						ep.MaxConcurrentActivityTaskPollers = size
						ep.MaxConcurrentDecisionTaskPollers = size
					}
				}
			}
		}
	}
}

func isWorkflowContext(inType reflect.Type) bool {
	// NOTE: We don't expect any one to derive from workflow context.
	return inType == reflect.TypeOf((*Context)(nil)).Elem()
}

func isValidResultType(inType reflect.Type) bool {
	// https://golang.org/pkg/reflect/#Kind
	switch inType.Kind() {
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return false
	}

	return true
}

func isError(inType reflect.Type) bool {
	errorElem := reflect.TypeOf((*error)(nil)).Elem()
	return inType != nil && inType.Implements(errorElem)
}

func getFunctionName(i interface{}) string {
	if fullName, ok := i.(string); ok {
		return fullName
	}
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	// This allows to call activities by method pointer
	// Compiler adds -fm suffix to a function name which has a receiver
	// Note that this works even if struct pointer used to get the function is nil
	// It is possible because nil receivers are allowed.
	// For example:
	// var a *Activities
	// ExecuteActivity(ctx, a.Foo)
	// will call this function which is going to return "Foo"
	return strings.TrimSuffix(fullName, "-fm")
}

func getActivityFunctionName(r *registry, i interface{}) string {
	result := getFunctionName(i)
	if alias, ok := r.getActivityAlias(result); ok {
		result = alias
	}
	return result
}

func getWorkflowFunctionName(r *registry, i interface{}) string {
	result := getFunctionName(i)
	if alias, ok := r.getWorkflowAlias(result); ok {
		result = alias
	}
	return result
}

func getReadOnlyChannel(c chan struct{}) <-chan struct{} {
	return c
}

func AugmentWorkerOptions(options WorkerOptions) WorkerOptions {
	if options.MaxConcurrentActivityExecutionSize == 0 {
		options.MaxConcurrentActivityExecutionSize = defaultMaxConcurrentActivityExecutionSize
	}
	if options.WorkerActivitiesPerSecond == 0 {
		options.WorkerActivitiesPerSecond = defaultWorkerActivitiesPerSecond
	}
	if options.MaxConcurrentActivityTaskPollers <= 0 {
		options.MaxConcurrentActivityTaskPollers = defaultConcurrentPollRoutineSize
	}
	if options.MaxConcurrentDecisionTaskExecutionSize == 0 {
		options.MaxConcurrentDecisionTaskExecutionSize = defaultMaxConcurrentTaskExecutionSize
	}
	if options.WorkerDecisionTasksPerSecond == 0 {
		options.WorkerDecisionTasksPerSecond = defaultWorkerTaskExecutionRate
	}
	if options.MaxConcurrentDecisionTaskPollers <= 0 {
		options.MaxConcurrentDecisionTaskPollers = defaultConcurrentPollRoutineSize
	}
	if options.MaxConcurrentLocalActivityExecutionSize == 0 {
		options.MaxConcurrentLocalActivityExecutionSize = defaultMaxConcurrentLocalActivityExecutionSize
	}
	if options.WorkerLocalActivitiesPerSecond == 0 {
		options.WorkerLocalActivitiesPerSecond = defaultWorkerLocalActivitiesPerSecond
	}
	if options.TaskListActivitiesPerSecond == 0 {
		options.TaskListActivitiesPerSecond = defaultTaskListActivitiesPerSecond
	}
	if options.StickyScheduleToStartTimeout.Seconds() == 0 {
		options.StickyScheduleToStartTimeout = stickyDecisionScheduleToStartTimeoutSeconds * time.Second
	}
	if options.DataConverter == nil {
		options.DataConverter = getDefaultDataConverter()
	}
	if options.MaxConcurrentSessionExecutionSize == 0 {
		options.MaxConcurrentSessionExecutionSize = defaultMaxConcurrentSessionExecutionSize
	}
	options.AutoScalerOptions = augmentAutoScalerOptions(options.AutoScalerOptions)

	// if the user passes in a tracer then add a tracing context propagator
	if options.Tracer != nil {
		options.ContextPropagators = append(options.ContextPropagators, NewTracingContextPropagator(options.Logger, options.Tracer))
	} else {
		options.Tracer = opentracing.NoopTracer{}
	}

	if options.EnableShadowWorker {
		options.DisableActivityWorker = true
		options.DisableWorkflowWorker = true
		options.EnableSessionWorker = false
	}

	return options
}

func augmentAutoScalerOptions(options AutoScalerOptions) AutoScalerOptions {
	// need at least 2 for sticky task polling to work
	if options.PollerMinCount <= 1 {
		options.PollerMinCount = defaultMinPollerSize
	}
	if options.PollerMaxCount <= 1 {
		options.PollerMaxCount = defaultMaxPollerSize
	}
	if options.PollerInitCount <= 1 {
		options.PollerInitCount = defaultInitPollerSize
	}
	if options.Cooldown == 0 {
		options.Cooldown = defaultPollerAutoScalerCooldown
	}
	if options.PollerWaitTimeUpperBound == 0 {
		options.PollerWaitTimeUpperBound = defaultPollerAutoScalerWaitTimeUpperBound
	}
	if options.PollerWaitTimeLowerBound == 0 {
		options.PollerWaitTimeLowerBound = defaultPollerAutoScalerWaitTimeLowerBound
	}
	return options
}

// getTestTags returns the test tags in the context.
func getTestTags(ctx context.Context) map[string]map[string]string {
	if ctx != nil {
		env := ctx.Value(testTagsContextKey)
		if env != nil {
			return env.(map[string]map[string]string)
		}
	}
	return nil
}

// StartVersionMetrics starts emitting version metrics
func StartVersionMetrics(metricsScope tally.Scope) {
	startVersionMetric.Do(func() {
		go func() {
			ticker := time.NewTicker(time.Minute)
			versionTags := map[string]string{clientVersionTag: LibraryVersion}
			for {
				select {
				case <-StopMetrics:
					return
				case <-ticker.C:
					metricsScope.Tagged(versionTags).Gauge(clientGauge).Update(1)
				}
			}
		}()
	})
}
