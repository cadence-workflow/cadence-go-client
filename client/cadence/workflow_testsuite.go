package cadence

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/facebookgo/clock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	"github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
	"github.com/uber-go/cadence-client/mocks"
	"github.com/uber/tchannel-go/thrift"
	"go.uber.org/zap"
)

const (
	defaultTestTaskList   = "default-test-tasklist"
	defaultTestWorkflowID = "default-test-workflow-id"
	defaultTestRunID      = "default-test-run-id"
)

type (
	// EncodedValue is type alias used to encapsulate/extract encoded result from workflow/activity.
	EncodedValue []byte

	// EncodedValues is a type alias used to encapsulate/extract encoded arguments from workflow/activity.
	EncodedValues []byte

	timerHandle struct {
		callback resultHandler
		timer    *clock.Timer
		duration time.Duration
		timerId  int
	}

	activityHandle struct {
		callback     resultHandler
		activityType string
	}

	callbackHandle struct {
		callback          func()
		startDecisionTask bool // start a new decision task after callback() is handled.
	}

	activityExecutorWrapper struct {
		activity
		env *TestWorkflowEnvironment
	}

	taskListSpecificActivity struct {
		fn        interface{}
		taskLists map[string]struct{}
	}

	// WorkflowTestSuite is the test suite to run unit tests for workflow/activity.
	WorkflowTestSuite struct {
		suite.Suite
		hostEnv                    *hostEnvImpl
		taskListSpecificActivities map[string]*taskListSpecificActivity
	}

	// TestWorkflowEnvironment is the environment that runs the workflow/activity unit tests.
	TestWorkflowEnvironment struct {
		testSuite          *WorkflowTestSuite
		overrodeActivities map[string]interface{} // map of registered-fnName -> fakeActivityFn

		service       m.TChanWorkflowService
		workerOptions WorkerOptions
		logger        *zap.Logger
		clock         clock.Clock

		workflowInfo          *WorkflowInfo
		workflowDef           workflowDefinition
		counterID             int
		workflowCancelHandler func()

		locker              *sync.Mutex
		scheduledActivities map[string]*activityHandle
		scheduledTimers     map[string]*timerHandle

		onActivityStartedListener func(ctx context.Context, args EncodedValues, activityType string)
		onActivityEndedListener   func(result EncodedValue, err error, activityType string)
		onTimerScheduledListener  func(timerID string, duration time.Duration)
		onTimerFiredListener      func(timerID string)
		onTimerCancelledListener  func(timerID string)

		callbackChannel chan callbackHandle
		isTestCompleted bool
		testResult      EncodedValue
		testError       error

		autoStartDecisionTask  bool
		enableFastForwardClock bool
	}

	// testWorkflowEnvironmentInternal is a wrapper around TestWorkflowEnvironment, it implements workflowEnvironment.
	// Using this wrapper to hide the real implementation methods on TestWorkflowEnvironment so those methods are not exposed.
	testWorkflowEnvironmentInternal struct {
		*TestWorkflowEnvironment
	}
)

// Get extract data from encoded data to desired value type. valuePtr is pointer to the actual value type.
func (b EncodedValue) Get(valuePtr interface{}) error {
	return getHostEnvironment().decodeArg(b, valuePtr)
}

// Get extract values from encoded data to desired value type. valuePtrs are pointers to the actual value types.
func (b EncodedValues) Get(valuePtrs ...interface{}) error {
	return getHostEnvironment().decode(b, valuePtrs)
}

// SetT sets the testing.T instance. This method is called by testify to setup the testing.T for test suite.
func (s *WorkflowTestSuite) SetT(t *testing.T) {
	s.Suite.SetT(t)
	if s.hostEnv == nil {
		s.hostEnv = &hostEnvImpl{
			workflowFuncMap: make(map[string]interface{}),
			activityFuncMap: make(map[string]interface{}),
			encoding:        gobEncoding{},
		}
		s.taskListSpecificActivities = make(map[string]*taskListSpecificActivity)
	}
}

// NewWorkflowTestSuite creates a WorkflowTestSuite
func NewWorkflowTestSuite() *WorkflowTestSuite {
	return &WorkflowTestSuite{
		hostEnv: &hostEnvImpl{
			workflowFuncMap: make(map[string]interface{}),
			activityFuncMap: make(map[string]interface{}),
			encoding:        gobEncoding{},
		},
		taskListSpecificActivities: make(map[string]*taskListSpecificActivity),
	}
}

// NewTestWorkflowEnvironment create a new instance of TestWorkflowEnvironment
func (s *WorkflowTestSuite) NewTestWorkflowEnvironment() *TestWorkflowEnvironment {
	env := &TestWorkflowEnvironment{
		testSuite: s,

		overrodeActivities: make(map[string]interface{}),

		workflowInfo: &WorkflowInfo{
			WorkflowExecution: WorkflowExecution{
				ID:    defaultTestWorkflowID,
				RunID: defaultTestRunID,
			},
			WorkflowType: WorkflowType{Name: "workflow-type-not-specified"},
			TaskListName: defaultTestTaskList,
		},

		locker:                 &sync.Mutex{},
		scheduledActivities:    make(map[string]*activityHandle),
		scheduledTimers:        make(map[string]*timerHandle),
		callbackChannel:        make(chan callbackHandle, 100),
		autoStartDecisionTask:  true,
		enableFastForwardClock: true,
	}

	if env.logger == nil {
		logger, _ := zap.NewDevelopment()
		env.logger = logger
	}

	if env.service == nil {
		mockService := new(mocks.TChanWorkflowService)
		mockHeartbeatFn := func(c thrift.Context, r *shared.RecordActivityTaskHeartbeatRequest) error {
			activityID := string(r.TaskToken)
			env.locker.Lock() // need lock as this is running in activity worker's goroutinue
			_, ok := env.scheduledActivities[activityID]
			env.locker.Unlock()
			if !ok {
				env.logger.Debug("RecordActivityTaskHeartbeat: ActivityID not found, could be already completed or cancelled.",
					zap.String(tagActivityID, activityID))
				return shared.NewEntityNotExistsError()
			}
			env.logger.Debug("RecordActivityTaskHeartbeat", zap.String("ActivityID", string(r.TaskToken)))
			return nil
		}

		mockService.On("RecordActivityTaskHeartbeat", mock.Anything, mock.Anything).Return(
			&shared.RecordActivityTaskHeartbeatResponse{CancelRequested: common.BoolPtr(false)},
			mockHeartbeatFn)
		env.service = mockService
	}
	if env.workerOptions == nil {
		env.workerOptions = NewWorkerOptions().SetLogger(env.logger)
	}

	if env.clock == nil {
		env.clock = clock.NewMock()
	}

	return env
}

// RegisterWorkflow registers a workflow that could be used by tests of this WorkflowTestSuite instance. All workflow registered
// via cadence.RegisterWorkflow() are still valid and will be available to all tests of all instance of WorkflowTestSuite.
func (s *WorkflowTestSuite) RegisterWorkflow(workflowFunc interface{}) {
	err := s.hostEnv.RegisterWorkflow(workflowFunc)
	if err != nil {
		panic(err)
	}
}

// RegisterActivity registers an activity with given task list that could be used by tests of this WorkflowTestSuite instance
// for that given tasklist. Activities registered via cadence.RegisterActivity() are still valid and will be available
// to all tests of all instances of WorkflowTestSuite for any tasklist. However, if an activity is registered via this
// WorkflowTestSuite.RegisterActivity() with a given tasklist, that activity will only be visible to that registered tasklist
// in tests of this WorkflowTestSuite instance, even if that activity was also registered via cadence.RegisterActivity().
// Example: ActivityA, ActivityB are registered via cadence.RegisterActivity(). These 2 activities will be visible to all
// tests of any instance of WorkflowTestSuite. If testSuite1 register ActivityA to task-list-1, then ActivityA will only
// be visible to task-list-1 for tests of testSuite1, and ActivityB will be visible to all tasklist for tests of testSuite1.
func (s *WorkflowTestSuite) RegisterActivity(activityFn interface{}, taskList string) {
	fnName := getFunctionName(activityFn)

	_, ok := s.hostEnv.activityFuncMap[fnName]
	if !ok {
		// activity not registered yet, register now
		err := s.hostEnv.RegisterActivity(activityFn)
		if err != nil {
			panic(err)
		}
	}

	taskListActivity, ok := s.taskListSpecificActivities[fnName]
	if !ok {
		taskListActivity = &taskListSpecificActivity{fn: activityFn, taskLists: make(map[string]struct{})}
		s.taskListSpecificActivities[fnName] = taskListActivity
	}
	taskListActivity.taskLists[taskList] = struct{}{}
}

// ExecuteWorkflow executes a workflow, wait until workflow complete or idleTimeout. It returns whether workflow is completed,
// the workflow result, and error. Returned isCompleted could be false if the workflow is blocked by activity or timer and
// cannot make progress within idleTimeout. If isCompleted is true, caller should use EncodedValue.Get() to extract
// strong typed result value.
func (s *WorkflowTestSuite) ExecuteWorkflow(
	idleTimeout time.Duration,
	workflowFn interface{},
	args ...interface{},
) (isCompleted bool, result EncodedValue, err error) {
	env := s.StartWorkflow(workflowFn, args...)
	env.StartDispatcherLoop(idleTimeout)
	return env.IsTestCompleted(), env.GetTestResult(), env.GetTestError()
}

// StartWorkflow creates a new TestWorkflowEnvironment that is prepared and ready to run the given workflow.
func (s *WorkflowTestSuite) StartWorkflow(workflowFn interface{}, args ...interface{}) *TestWorkflowEnvironment {
	var workflowType string
	fnType := reflect.TypeOf(workflowFn)
	switch fnType.Kind() {
	case reflect.String:
		workflowType = workflowFn.(string)
	case reflect.Func:
		workflowType = getFunctionName(workflowFn)
	default:
		panic("unsupported workflowFn")
	}

	env := s.NewTestWorkflowEnvironment()
	env.workflowInfo.WorkflowType.Name = workflowType
	factory := getWorkflowDefinitionFactory(s.hostEnv.newRegisteredWorkflowFactory())
	workflowDefinition, err := factory(env.workflowInfo.WorkflowType)
	if err != nil {
		// try to get workflow from global registered workflows
		factory = getWorkflowDefinitionFactory(getHostEnvironment().newRegisteredWorkflowFactory())
		workflowDefinition, err = factory(env.workflowInfo.WorkflowType)
		if err != nil {
			panic(err)
		}
	}
	env.workflowDef = workflowDefinition

	input, err := s.hostEnv.encodeArgs(args)
	if err != nil {
		panic(err)
	}
	env.workflowDef.Execute(&testWorkflowEnvironmentInternal{env}, input)
	return env
}

// StartWorkflowPart wraps a function and test it just as if it is a workflow. You don's need to register workflowPartFn.
func (s *WorkflowTestSuite) StartWorkflowPart(workflowPartFn interface{}, args ...interface{}) *TestWorkflowEnvironment {
	// auto register workflow
	fnName := getFunctionName(workflowPartFn)
	if _, ok := s.hostEnv.getWorkflowFn(fnName); !ok {
		s.RegisterWorkflow(workflowPartFn)
	}

	return s.StartWorkflow(workflowPartFn, args)
}

// Override overrides an actual activity with a fake activity. The fake activity will be invoked in place where the
// actual activity should have been invoked.
func (env *TestWorkflowEnvironment) Override(activityFn, fakeActivityFn interface{}) {
	// verify both functions are valid activity func
	actualFnType := reflect.TypeOf(activityFn)
	if err := validateFnFormat(actualFnType, false); err != nil {
		panic(err)
	}
	fakeFnType := reflect.TypeOf(fakeActivityFn)
	if err := validateFnFormat(fakeFnType, false); err != nil {
		panic(err)
	}

	// verify signature of registeredActivityFn and fakeActivityFn are the same.
	if actualFnType != fakeFnType {
		panic("activityFn and fakeActivityFn have different func signature")
	}

	fnName := getFunctionName(activityFn)
	env.overrodeActivities[fnName] = fakeActivityFn
}

// ExecuteActivity executes an activity. The tested activity will be executed synchronously in the calling goroutinue.
// Caller should use EncodedValue.Get() to extract strong typed result value.
func (env *TestWorkflowEnvironment) ExecuteActivity(activityFn interface{}, args ...interface{}) (EncodedValue, error) {
	fnName := getFunctionName(activityFn)

	input, err := getHostEnvironment().encodeArgs(args)
	if err != nil {
		panic(err)
	}

	task := newTestActivityTask(
		defaultTestWorkflowID,
		defaultTestRunID,
		"0",
		fnName,
		input,
	)

	// ensure activityFn is registered to defaultTestTaskList
	env.testSuite.RegisterActivity(activityFn, defaultTestTaskList)
	taskHandler := env.newTestActivityTaskHandler(defaultTestTaskList)
	result, err := taskHandler.Execute(task)
	switch request := result.(type) {
	case *shared.RespondActivityTaskCanceledRequest:
		return nil, NewCanceledError(request.Details)
	case *shared.RespondActivityTaskFailedRequest:
		return nil, NewErrorWithDetails(*request.Reason, request.Details)
	case *shared.RespondActivityTaskCompletedRequest:
		return EncodedValue(request.Result_), nil
	default:
		// will never happen
		return nil, fmt.Errorf("unsupported respond type %T", result)
	}
}

// SetLogger sets the logger for TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) SetLogger(logger *zap.Logger) *TestWorkflowEnvironment {
	env.logger = logger
	return env
}

// SetService sets the m.TChanWorkflowService for TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) SetService(service m.TChanWorkflowService) *TestWorkflowEnvironment {
	env.service = service
	return env
}

// SetClock sets the clock that will be used for TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) SetClock(clock clock.Clock) *TestWorkflowEnvironment {
	env.clock = clock
	return env
}

// SetWorkerOption sets the WorkerOptions for WorkflowTestSuite
func (env *TestWorkflowEnvironment) SetWorkerOption(options WorkerOptions) *TestWorkflowEnvironment {
	env.workerOptions = options
	return env
}

// StartDecisionTask will trigger OnDecisionTaskStart() on the workflow which will execute the dispatcher until all
// coroutinues are blocked. This method is only necessary when you disable the auto start decision task to have full
// control of the workflow execution on when to start a decision task.
func (t *TestWorkflowEnvironment) StartDecisionTask() {
	// post an empty callback to event loop, and request OnDecisionTaskStarted to be triggered after that empty callback
	t.postCallback(func() {}, true /* to start decision task */)
}

// EnableAutoStartDecisionTask enables auto start decision task on every events. It is equivalent to immediately schedule
// a new decision task on new history events (like activity completed/failed/cancelled, timer fired/cancelled, etc).
// Default is true. Only set this to false if you need to precisely control when to schedule new decision task. For example,
// if your workflow code starts 2 activities, and you only want a new decision task scheduled after both of them are completed.
// Your test can set a listener by using SetOnActivityEndedListener() to know when your activities are done to determine
// when to schedule a new decision task.
func (t *TestWorkflowEnvironment) EnableAutoStartDecisionTask(enable bool) *TestWorkflowEnvironment {
	t.autoStartDecisionTask = enable
	return t
}

// EnableClockFastForwardWhenBlockedByTimer enables auto clock fast forward when dispatcher is blocked by timer.
// Default is true. If you set this flag to false, your timer will be fired by below 2 cases:
//  1) Use real clock, and timer is fired when time goes by.
//  2) Use mock clock, and you need to manually move forward mock clock to fire timer.
func (t *TestWorkflowEnvironment) EnableClockFastForwardWhenBlockedByTimer(enable bool) *TestWorkflowEnvironment {
	t.enableFastForwardClock = enable
	return t
}

// SetOnActivityStartedListener sets a listener that will be called when an activity task started.
func (t *TestWorkflowEnvironment) SetOnActivityStartedListener(listener func(ctx context.Context, args EncodedValues, activityType string)) {
	t.onActivityStartedListener = listener
}

// SetOnActivityEndedListener sets a listener that will be called when an activity task ended.
func (t *TestWorkflowEnvironment) SetOnActivityEndedListener(listener func(result EncodedValue, err error, activityType string)) {
	t.onActivityEndedListener = listener
}

// SetOnTimerScheduledListener sets a listener that will be called when a timer is scheduled.
func (t *TestWorkflowEnvironment) SetOnTimerScheduledListener(listener func(timerID string, duration time.Duration)) {
	t.onTimerScheduledListener = listener
}

// SetOnTimerFiredListener sets a listener that will be called when a timer is fired
func (t *TestWorkflowEnvironment) SetOnTimerFiredListener(listener func(timerID string)) {
	t.onTimerFiredListener = listener
}

// SetOnTimerCancelledListener sets a listener that will be called when a timer is cancelled
func (t *TestWorkflowEnvironment) SetOnTimerCancelledListener(listener func(timerID string)) {
	t.onTimerCancelledListener = listener
}

// StartDispatcherLoop starts the main loop that drives workflow dispatcher. The main loop runs in the calling goroutinue
// and it blocked until tested workflow is completed.
func (t *TestWorkflowEnvironment) StartDispatcherLoop(idleTimeout time.Duration) bool {
	// kick off the initial decision task
	t.StartDecisionTask()

	for {
		if t.isTestCompleted {
			t.logger.Debug("Workflow completed, stop callback processing...")
			return true
		}

		// use non-blocking-select to check if there is anything pending in the main thread.
		select {
		case c := <-t.callbackChannel:
			// this will drain the callbackChannel
			t.processCallback(c)
		default:
			// nothing to process, main thread is blocked at this moment, now check if we should auto fire next timer
			if !t.autoFireNextTimer() {
				// no timer to fire, wait for things to do or timeout.
				select {
				case c := <-t.callbackChannel:
					t.processCallback(c)
				case <-time.After(idleTimeout):
					t.logger.Info("Idle timeout, existing.",
						zap.Duration("Timeout", idleTimeout),
						zap.String("DispatcherStack", t.workflowDef.StackTrace()))
					return false
				}
			}
		}
	}
}

func (t *TestWorkflowEnvironment) processCallback(c callbackHandle) {
	// locker is needed to prevent race condition between dispatcher loop goroutinue and activity worker goroutinues.
	// The activity workers could call into Heartbeat which by default is mocked in this test suite. The mock needs to
	// access s.scheduledActivities map, that could cause data race warning.
	t.locker.Lock()
	defer t.locker.Unlock()
	c.callback()
	if c.startDecisionTask {
		t.workflowDef.OnDecisionTaskStarted() // this will execute dispatcher
	}
}

func (t *TestWorkflowEnvironment) autoFireNextTimer() bool {
	if !t.enableFastForwardClock || len(t.scheduledTimers) == 0 {
		return false
	}

	// find next timer
	var tofire *timerHandle
	for _, t := range t.scheduledTimers {
		if tofire == nil {
			tofire = t
		} else if t.duration < tofire.duration ||
			(t.duration == tofire.duration && t.timerId < tofire.timerId) {
			tofire = t
		}
	}

	mockClock, ok := t.clock.(*clock.Mock)
	if !ok {
		panic("configured clock does not support fast forward, must use a mock clock for fast forward")
	}
	d := tofire.duration
	t.logger.Sugar().Debugf("Auto fire timer %d, moving clock forward by %v.\n", tofire.timerId, tofire.duration)

	// Move mock clock forward, this will fire the timer, and the timer callback will remove timer from scheduledTimers.
	mockClock.Add(d)

	// reduce all pending timer's duration by d
	for _, t := range t.scheduledTimers {
		t.duration -= d
	}
	return true
}

func (t *TestWorkflowEnvironment) postCallback(cb func(), startDecisionTask bool) {
	t.callbackChannel <- callbackHandle{callback: cb, startDecisionTask: startDecisionTask}
}

func (t *TestWorkflowEnvironment) requestCancelActivity(activityID string) {
	t.logger.Sugar().Debugf("RequestCancelActivity %v", activityID)
	handle, ok := t.scheduledActivities[activityID]
	if !ok {
		t.logger.Debug("RequestCancelActivity failed, ActivityID not exists.", zap.String(tagActivityID, activityID))
		return
	}

	delete(t.scheduledActivities, activityID)
	t.postCallback(func() {
		handle.callback(nil, NewCanceledError())
	}, t.autoStartDecisionTask)
}

func (t *TestWorkflowEnvironment) requestCancelTimer(timerID string) {
	t.logger.Sugar().Debugf("RequestCancelTimer %v", timerID)
	timerHandle, ok := t.scheduledTimers[timerID]
	if !ok {
		t.logger.Debug("RequestCancelTimer failed, TimerID not exists.", zap.String(tagTimerID, timerID))
	}

	delete(t.scheduledTimers, timerID)
	timerHandle.timer.Stop()
	t.postCallback(func() {
		timerHandle.callback(nil, NewCanceledError())
		if t.onTimerCancelledListener != nil {
			t.onTimerCancelledListener(timerID)
		}
	}, t.autoStartDecisionTask)
}

func (t *TestWorkflowEnvironment) complete(result []byte, err error) {
	if t.isTestCompleted {
		t.logger.Debug("Workflow already completed.")
		return
	}
	t.isTestCompleted = true
	t.testResult = EncodedValue(result)
	t.testError = err

	if err == ErrCanceled && t.workflowCancelHandler != nil {
		t.workflowCancelHandler()
	}
}

// IsTestCompleted check if test is completed or not
func (t *TestWorkflowEnvironment) IsTestCompleted() bool {
	return t.isTestCompleted
}

// GetTestResult return the encoded result from test workflow
func (t *TestWorkflowEnvironment) GetTestResult() EncodedValue {
	return t.testResult
}

// GetTestError return the error from test workflow
func (t *TestWorkflowEnvironment) GetTestError() error {
	return t.testError
}

func (t *TestWorkflowEnvironment) CompleteActivity(taskToken []byte, result interface{}, err error) error {
	if taskToken == nil {
		return errors.New("nil task token provided")
	}
	var data []byte
	if result != nil {
		var encodeErr error
		data, encodeErr = getHostEnvironment().encodeArg(result)
		if encodeErr != nil {
			return encodeErr
		}
	}

	activityID := string(taskToken)
	t.postCallback(func() {
		activityHandle, ok := t.scheduledActivities[activityID]
		if !ok {
			t.logger.Debug("CompleteActivity: ActivityID not found, could be already completed or cancelled.",
				zap.String(tagActivityID, activityID))
			return
		}
		request := convertActivityResultToRespondRequest("test-identity", taskToken, data, err)
		t.handleActivityResult(activityID, request, activityHandle.activityType)
	}, false /* do not auto schedule decision task, because activity might be still pending */)

	return nil
}

func (t *TestWorkflowEnvironment) getLogger() *zap.Logger {
	return t.logger
}

func (t *TestWorkflowEnvironment) executeActivity(parameters executeActivityParameters, callback resultHandler) *activityInfo {
	activityInfo := &activityInfo{fmt.Sprintf("%d", t.nextId())}

	task := newTestActivityTask(
		defaultTestWorkflowID,
		defaultTestRunID,
		activityInfo.activityID,
		parameters.ActivityType.Name,
		parameters.Input,
	)

	taskHandler := t.newTestActivityTaskHandler(parameters.TaskListName)
	activityHandle := &activityHandle{callback: callback, activityType: parameters.ActivityType.Name}
	t.scheduledActivities[activityInfo.activityID] = activityHandle

	// activity runs in separate goroutinue outside of workflow dispatcher
	go func() {
		result, err := taskHandler.Execute(task)
		if err != nil {
			panic(err)
		}
		// post activity result to workflow dispatcher
		t.postCallback(func() {
			t.handleActivityResult(activityInfo.activityID, result, parameters.ActivityType.Name)
		}, false /* do not auto schedule decision task, because activity might be still pending */)
	}()

	return activityInfo
}

func (t *TestWorkflowEnvironment) handleActivityResult(activityID string, result interface{}, activityType string) {
	if result == nil {
		// In case activity returns ErrActivityResultPending, the respond will be nil, and we don's need to do anything.
		// Activity will need to complete asynchronously using CompleteActivity().
		if t.onActivityEndedListener != nil {
			t.onActivityEndedListener(nil, ErrActivityResultPending, activityType)
		}
		return
	}

	// this is running in dispatcher
	activityHandle, ok := t.scheduledActivities[activityID]
	if !ok {
		t.logger.Debug("handleActivityResult: ActivityID not exists, could be already completed or cancelled.",
			zap.String(tagActivityID, activityID))
		return
	}

	delete(t.scheduledActivities, activityID)

	var blob []byte
	var err error

	switch request := result.(type) {
	case *shared.RespondActivityTaskCanceledRequest:
		err = NewCanceledError(request.Details)
		activityHandle.callback(nil, err)
	case *shared.RespondActivityTaskFailedRequest:
		err = NewErrorWithDetails(*request.Reason, request.Details)
		activityHandle.callback(nil, err)
	case *shared.RespondActivityTaskCompletedRequest:
		blob = request.Result_
		activityHandle.callback(blob, nil)
	default:
		panic(fmt.Sprintf("unsupported respond type %T", result))
	}

	if t.onActivityEndedListener != nil {
		t.onActivityEndedListener(EncodedValue(blob), err, activityType)
	}
	if t.autoStartDecisionTask {
		t.StartDecisionTask()
	}
}

// Execute executes the activity code. This is the wrapper where we call ActivityTaskStartedListener hook.
func (a *activityExecutorWrapper) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if a.env.onActivityStartedListener != nil {
		a.env.postCallback(func() {
			a.env.onActivityStartedListener(ctx, EncodedValues(input), a.ActivityType().Name)
		}, false)
	}
	return a.activity.Execute(ctx, input)
}

func (t *TestWorkflowEnvironment) newTestActivityTaskHandler(taskList string) ActivityTaskHandler {
	wOptions := t.workerOptions.(*workerOptions)
	params := workerExecutionParameters{
		TaskList:     taskList,
		Identity:     wOptions.identity,
		MetricsScope: wOptions.metricsScope,
		Logger:       wOptions.logger,
		UserContext:  wOptions.userContext,
	}
	if params.Logger == nil && t.logger != nil {
		params.Logger = t.logger
	}
	ensureRequiredParams(&params)

	var activities []activity
	for fnName, tasklistActivity := range t.testSuite.taskListSpecificActivities {
		if _, ok := tasklistActivity.taskLists[taskList]; ok {
			activities = append(activities, t.wrapActivity(&activityExecutor{name: fnName, fn: tasklistActivity.fn}))
		}
	}

	// get global registered activities
	globalActivities := getHostEnvironment().getRegisteredActivities()
	for _, a := range globalActivities {
		fnName := a.ActivityType().Name
		if _, ok := t.testSuite.taskListSpecificActivities[fnName]; ok {
			// activity is registered to a specific taskList, so ignore it from the global registered activities.
			continue
		}
		activities = append(activities, t.wrapActivity(a))
	}

	if len(activities) == 0 {
		panic(fmt.Sprintf("no activity is registered for tasklist '%v'", taskList))
	}

	taskHandler := newActivityTaskHandler(activities, t.service, params)
	return taskHandler
}

func (t *TestWorkflowEnvironment) wrapActivity(a activity) *activityExecutorWrapper {
	fnName := a.ActivityType().Name
	if overrideFn, ok := t.overrodeActivities[fnName]; ok {
		// override activity
		a = &activityExecutor{name: fnName, fn: overrideFn}
	}

	activityWrapper := &activityExecutorWrapper{activity: a, env: t}
	return activityWrapper
}

func newTestActivityTask(workflowID, runID, activityID, activityType string, input []byte) *shared.PollForActivityTaskResponse {
	task := &shared.PollForActivityTaskResponse{
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		ActivityId:   common.StringPtr(activityID),
		TaskToken:    []byte(activityID), // use activityID as TaskToken so we can map TaskToken in heartbeat calls.
		ActivityType: &shared.ActivityType{Name: common.StringPtr(activityType)},
		Input:        input,
	}
	return task
}

func (t *TestWorkflowEnvironment) newTimer(d time.Duration, callback resultHandler) *timerInfo {
	nextId := t.nextId()
	timerInfo := &timerInfo{fmt.Sprintf("%d", nextId)}
	timer := t.clock.AfterFunc(d, func() {
		t.postCallback(func() {
			delete(t.scheduledTimers, timerInfo.timerID)
			callback(nil, nil)
			if t.onTimerFiredListener != nil {
				t.onTimerFiredListener(timerInfo.timerID)
			}
		}, t.autoStartDecisionTask)
	})
	t.scheduledTimers[timerInfo.timerID] = &timerHandle{timer: timer, callback: callback, duration: d, timerId: nextId}
	if t.onTimerScheduledListener != nil {
		t.onTimerScheduledListener(timerInfo.timerID, d)
	}
	return timerInfo
}

func (t *TestWorkflowEnvironment) now() time.Time {
	return t.clock.Now()
}

func (t *TestWorkflowEnvironment) getWorkflowInfo() *WorkflowInfo {
	return t.workflowInfo
}

func (t *TestWorkflowEnvironment) registerCancel(handler func()) {
	t.workflowCancelHandler = handler
}

func (t *TestWorkflowEnvironment) requestCancelWorkflow(domainName, workflowID, runID string) error {
	panic("not implemented yet")
}

func (t *TestWorkflowEnvironment) nextId() int {
	activityID := t.counterID
	t.counterID++
	return activityID
}

// bellow are the implementation methods of testWorkflowEnvironmentInternal. The actual methods are defined as private
// methods of TestWorkflowEnvironment so user won's accidentally call into those methods.
func (env *testWorkflowEnvironmentInternal) Now() time.Time {
	return env.now()
}
func (env *testWorkflowEnvironmentInternal) NewTimer(d time.Duration, callback resultHandler) *timerInfo {
	return env.newTimer(d, callback)
}
func (env *testWorkflowEnvironmentInternal) RequestCancelTimer(timerID string) {
	env.requestCancelTimer(timerID)
}
func (env *testWorkflowEnvironmentInternal) ExecuteActivity(parameters executeActivityParameters, callback resultHandler) *activityInfo {
	return env.executeActivity(parameters, callback)
}
func (env *testWorkflowEnvironmentInternal) RequestCancelActivity(activityID string) {
	env.requestCancelActivity(activityID)
}
func (env *testWorkflowEnvironmentInternal) WorkflowInfo() *WorkflowInfo {
	return env.getWorkflowInfo()
}
func (env *testWorkflowEnvironmentInternal) Complete(result []byte, err error) {
	env.complete(result, err)
}
func (env *testWorkflowEnvironmentInternal) RegisterCancel(handler func()) {
	env.registerCancel(handler)
}
func (env *testWorkflowEnvironmentInternal) RequestCancelWorkflow(domainName, workflowID, runID string) error {
	return env.requestCancelWorkflow(domainName, workflowID, runID)
}
func (env *testWorkflowEnvironmentInternal) GetLogger() *zap.Logger {
	return env.getLogger()
}

// make sure interface is implemented
var _ workflowEnvironment = (*testWorkflowEnvironmentInternal)(nil)
