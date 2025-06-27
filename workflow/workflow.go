// Copyright (c) 2017-2020 Uber Technologies, Inc.
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

package workflow

import (
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/internal"
)

type (

	// ChildWorkflowFuture represents the result of a child workflow execution
	ChildWorkflowFuture = internal.ChildWorkflowFuture

	// Type identifies a workflow type.
	Type = internal.WorkflowType

	// Execution Details.
	Execution = internal.WorkflowExecution

	// ExecutionAsync Details.
	ExecutionAsync = internal.WorkflowExecutionAsync

	// Version represents a change version. See GetVersion call.
	Version = internal.Version

	// ChildWorkflowOptions stores all child workflow specific parameters that will be stored inside of a Context.
	ChildWorkflowOptions = internal.ChildWorkflowOptions

	// Bugports contains opt-in flags for backports of old, buggy behavior that has been fixed.
	// See the internal docs for details.
	Bugports = internal.Bugports

	// RegisterOptions consists of options for registering a workflow
	RegisterOptions = internal.RegisterWorkflowOptions

	// Info information about currently executing workflow
	Info = internal.WorkflowInfo

	RegistryInfo = internal.RegistryWorkflowInfo

	// GetVersionOption is used to specify options for GetVersion
	GetVersionOption = internal.GetVersionOption
)

// Register - registers a workflow function with the framework.
// A workflow takes a workflow context and input and returns a (result, error) or just error.
// Examples:
//
//	func sampleWorkflow(ctx workflow.Context, input []byte) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context, arg1 int, arg2 string) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context, arg1 int) (result string, err error)
//
// Serialization of all primitive types, structures is supported ... except channels, functions, variadic, unsafe pointer.
// This method calls panic if workflowFunc doesn't comply with the expected format.
// Deprecated: Global workflow registration methods are replaced by equivalent Worker instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func Register(workflowFunc interface{}) {
	internal.RegisterWorkflow(workflowFunc)
}

// RegisterWithOptions registers the workflow function with options
// The user can use options to provide an external name for the workflow or leave it empty if no
// external name is required. This can be used as
//
//	client.RegisterWorkflow(sampleWorkflow, RegisterOptions{})
//	client.RegisterWorkflow(sampleWorkflow, RegisterOptions{Name: "foo"})
//
// A workflow takes a workflow context and input and returns a (result, error) or just error.
// Examples:
//
//	func sampleWorkflow(ctx workflow.Context, input []byte) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context, arg1 int, arg2 string) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context, arg1 int) (result string, err error)
//
// Serialization of all primitive types, structures is supported ... except channels, functions, variadic, unsafe pointer.
// This method calls panic if workflowFunc doesn't comply with the expected format.
// Deprecated: Global workflow registration methods are replaced by equivalent Worker instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func RegisterWithOptions(workflowFunc interface{}, opts RegisterOptions) {
	internal.RegisterWorkflowWithOptions(workflowFunc, opts)
}

// GetRegisteredWorkflowTypes returns the registered workflow function/alias names.
func GetRegisteredWorkflowTypes() []string {
	return internal.GetRegisteredWorkflowTypes()
}

// ExecuteActivity requests activity execution in the context of a workflow.
// Context can be used to pass the settings for this activity.
// For example: task list that this need to be routed, timeouts that need to be configured.
// Use ActivityOptions to pass down the options.
//
//	 ao := ActivityOptions{
//		    TaskList: "exampleTaskList",
//		    ScheduleToStartTimeout: 10 * time.Second,
//		    StartToCloseTimeout: 5 * time.Second,
//		    ScheduleToCloseTimeout: 10 * time.Second,
//		    HeartbeatTimeout: 0,
//		}
//		ctx := WithActivityOptions(ctx, ao)
//
// Or to override a single option
//
//	ctx := WithTaskList(ctx, "exampleTaskList")
//
// Input activity is either an activity name (string) or a function representing an activity that is getting scheduled.
// Input args are the arguments that need to be passed to the scheduled activity.
//
// If the activity failed to complete then the future get error would indicate the failure, and it can be one of
// CustomError, TimeoutError, CanceledError, PanicError, GenericError.
// You can cancel the pending activity using context(workflow.WithCancel(ctx)) and that will fail the activity with
// error CanceledError.
//
// ExecuteActivity returns Future with activity result or failure.
func ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	return internal.ExecuteActivity(ctx, activity, args...)
}

// ExecuteLocalActivity requests to run a local activity. A local activity is like a regular activity with some key
// differences:
//
// • Local activity is scheduled and run by the workflow worker locally.
//
// • Local activity does not need Cadence server to schedule activity task and does not rely on activity worker.
//
// • No need to register local activity.
//
// • The parameter activity to ExecuteLocalActivity() must be a function.
//
// • Local activity is for short living activities (usually finishes within seconds).
//
// • Local activity cannot heartbeat.
//
// Context can be used to pass the settings for this local activity.
// For now there is only one setting for timeout to be set:
//
//	lao := LocalActivityOptions{
//		ScheduleToCloseTimeout: 5 * time.Second,
//	}
//	ctx := WithLocalActivityOptions(ctx, lao)
//
// The timeout here should be relative shorter than the DecisionTaskStartToCloseTimeout of the workflow. If you need a
// longer timeout, you probably should not use local activity and instead should use regular activity. Local activity is
// designed to be used for short living activities (usually finishes within seconds).
//
// Input args are the arguments that will to be passed to the local activity. The input args will be hand over directly
// to local activity function without serialization/deserialization because we don't need to pass the input across process
// boundary. However, the result will still go through serialization/deserialization because we need to record the result
// as history to cadence server so if the workflow crashes, a different worker can replay the history without running
// the local activity again.
//
// If the activity failed to complete then the future get error would indicate the failure, and it can be one of
// CustomError, TimeoutError, CanceledError, PanicError, GenericError.
// You can cancel the pending activity by cancel the context(workflow.WithCancel(ctx)) and that will fail the activity
// with error CanceledError.
//
// ExecuteLocalActivity returns Future with local activity result or failure.
func ExecuteLocalActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	return internal.ExecuteLocalActivity(ctx, activity, args...)
}

// ExecuteChildWorkflow requests child workflow execution in the context of a workflow.
// Context can be used to pass the settings for the child workflow.
// For example: task list that this child workflow should be routed, timeouts that need to be configured.
// Use ChildWorkflowOptions to pass down the options.
//
//	 cwo := ChildWorkflowOptions{
//		    ExecutionStartToCloseTimeout: 10 * time.Minute,
//		    TaskStartToCloseTimeout: time.Minute,
//		}
//	 ctx := WithChildOptions(ctx, cwo)
//
// Input childWorkflow is either a workflow name or a workflow function that is getting scheduled.
// Input args are the arguments that need to be passed to the child workflow function represented by childWorkflow.
// If the child workflow failed to complete then the future get error would indicate the failure and it can be one of
// CustomError, TimeoutError, CanceledError, GenericError.
// You can cancel the pending child workflow using context(workflow.WithCancel(ctx)) and that will fail the workflow with
// error CanceledError.
// ExecuteChildWorkflow returns ChildWorkflowFuture.
func ExecuteChildWorkflow(ctx Context, childWorkflow interface{}, args ...interface{}) ChildWorkflowFuture {
	return internal.ExecuteChildWorkflow(ctx, childWorkflow, args...)
}

// GetInfo extracts info of a current workflow from a context.
func GetInfo(ctx Context) *Info {
	return internal.GetWorkflowInfo(ctx)
}

// GetLogger returns a logger to be used in workflow's context
func GetLogger(ctx Context) *zap.Logger {
	return internal.GetLogger(ctx)
}

// GetUnhandledSignalNames returns signal names that have  unconsumed signals.
func GetUnhandledSignalNames(ctx Context) []string {
	return internal.GetUnhandledSignalNames(ctx)
}

// GetMetricsScope returns a metrics scope to be used in workflow's context
func GetMetricsScope(ctx Context) tally.Scope {
	return internal.GetMetricsScope(ctx)
}

// GetTotalHistoryBytes returns the current history size of that workflow
func GetTotalHistoryBytes(ctx Context) int64 {
	return internal.GetTotalEstimatedHistoryBytes(ctx)
}

// GetHistoryCount returns the current number of history event of that workflow
func GetHistoryCount(ctx Context) int64 {
	return internal.GetHistoryCount(ctx)
}

// RequestCancelExternalWorkflow can be used to request cancellation of an external workflow.
// Input workflowID is the workflow ID of target workflow.
// Input runID indicates the instance of a workflow. Input runID is optional (default is ""). When runID is not specified,
// then the currently running instance of that workflowID will be used.
// By default, the current workflow's domain will be used as target domain. However, you can specify a different domain
// of the target workflow using the context like:
//
//	ctx := WithWorkflowDomain(ctx, "domain-name")
//
// RequestCancelExternalWorkflow return Future with failure or empty success result.
func RequestCancelExternalWorkflow(ctx Context, workflowID, runID string) Future {
	return internal.RequestCancelExternalWorkflow(ctx, workflowID, runID)
}

// SignalExternalWorkflow can be used to send signal info to an external workflow.
// Input workflowID is the workflow ID of target workflow.
// Input runID indicates the instance of a workflow. Input runID is optional (default is ""). When runID is not specified,
// then the currently running instance of that workflowID will be used.
// By default, the current workflow's domain will be used as target domain. However, you can specify a different domain
// of the target workflow using the context like:
//
//	ctx := WithWorkflowDomain(ctx, "domain-name")
//
// SignalExternalWorkflow return Future with failure or empty success result.
func SignalExternalWorkflow(ctx Context, workflowID, runID, signalName string, arg interface{}) Future {
	return internal.SignalExternalWorkflow(ctx, workflowID, runID, signalName, arg)
}

// GetSignalChannel returns channel corresponding to the signal name.
func GetSignalChannel(ctx Context, signalName string) Channel {
	return internal.GetSignalChannel(ctx, signalName)
}

// SideEffect executes the provided callback, and records its result into the workflow history.
// During replay the recorded result will be returned instead, so the callback can do non-deterministic
// things (such as reading files or getting random numbers) without breaking determinism during replay.
//
// As there is no error return, the callback's code is assumed to be reliable.
// If you cannot retrieve the value for some reason, panicking and failing the decision task
// will cause it to be retried, possibly succeeding on another machine.
// For better error handling, use ExecuteLocalActivity instead.
//
// Caution: the callback MUST NOT call any blocking or history-creating APIs,
// e.g. workflow.Sleep or ExecuteActivity or calling .Get on any future.
// This will (potentially) work during the initial recording of history, but will
// fail when replaying because the data is not available when the call occurs
// (it becomes available later).
//
// In other words, in general: use this *only* for code that does not need the workflow.Context.
// Incorrect context-using calls are not currently prevented in tests or during execution,
// but they should be eventually.
//
//	// Bad example: this will work until a replay occurs,
//	// but then the workflow will fail to replay with a non-deterministic error.
//	var out string
//	err := workflow.SideEffect(func(ctx workflow.Context) interface{} {
//		var signal data
//		err := workflow.GetSignalChannel(ctx, "signal").Receive(ctx, &signal)
//		_ = err // ignore err for the example
//		return signal
//	}).Get(&out)
//	workflow.GetLogger(ctx).Info(out)
//
// Caution: do not use SideEffect to modify values outside the callback.
// Always retrieve result from SideEffect's encoded return value.
// For example this code will break during replay:
//
//	// Bad example:
//	var random int
//	workflow.SideEffect(func(ctx workflow.Context) interface{} {
//		random = rand.Intn(100) // this only occurs when recording history, not during replay
//		return nil
//	})
//	// random will always be 0 in replay, thus this code is non-deterministic
//	if random < 50 {
//		....
//	} else {
//		....
//	}
//
// On replay the provided function is not executed, the `random` var will always be 0,
// and the workflow could take a different path and break determinism.
//
// The only safe way to use SideEffect is to read from the returned encoded.Value:
//
//	// Good example:
//	encodedRandom := SideEffect(func(ctx workflow.Context) interface{} {
//		return rand.Intn(100)
//	})
//	var random int
//	encodedRandom.Get(&random)
//	if random < 50 {
//		....
//	} else {
//		....
//	}
func SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value {
	return internal.SideEffect(ctx, f)
}

// MutableSideEffect is similar to SideEffect, but it only records changed values per ID instead of every call.
// Changes are detected by calling the provided "equals" func - return true to mark a result
// as the same as before, and not record the new value.  The first call per ID is always "changed" and will always be recorded.
//
// This is intended for things you want to *check* frequently, but which *change* only occasionally, as non-changing
// results will not add anything to your history.  This helps keep those workflows under its size/count limits, and
// can help keep replays more efficient than when using SideEffect or ExecuteLocalActivity (due to not storing duplicated data).
//
// Caution: the callback MUST NOT block or call any history-modifying events, or replays will be broken.
// See SideEffect docs for more details.
//
// Caution: do not use MutableSideEffect to modify closures.
// Always retrieve result from MutableSideEffect's encoded return value.
// See SideEffect docs for more details.
func MutableSideEffect(ctx Context, id string, f func(ctx Context) interface{}, equals func(a, b interface{}) bool) encoded.Value {
	return internal.MutableSideEffect(ctx, id, f, equals)
}

// DefaultVersion is a version returned by GetVersion for code that wasn't versioned before
const DefaultVersion Version = internal.DefaultVersion

// ExecuteWithVersion forces a specific version to be returned when GetVersion is executed for the first time,
// instead of returning maxSupported version. This option can be used when you want to separate the versioning of the workflow code and
// activation of the new logic in the workflow code, to ensure that your changes can be safely rolled back, if needed.
//
// For example, initially a workflow has the following code:
//
//	err = workflow.ExecuteActivity(ctx, foo).Get(ctx, nil)
//
// It should be updated to:
//
//	err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//
// Following the steps below, your changes will be forward and backward compatible, keeping possible a safe rollback
// to the previous version of the workflow code:
//
// 1. Keep execution of foo activity, add support for bar activity
//
//	v := GetVersion(ctx, "fooChange", DefaultVersion, 1,  ExecuteWithVersion(DefaultVersion))
//	if v == DefaultVersion {
//		err = workflow.ExecuteActivity(ctx, foo).Get(ctx, nil)
//	} else {
//		err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//	}
//
// The code above supports replaying of workflow execution with both versions DefaultVersion and 1.
// All new workflow executions will execute foo activity, because ExecuteWithVersion(DefaultVersion)
// causes GetVersion to return DefaultVersion rather than 1.
// In this example, this is also exactly the same as using ExecuteWithMinVersion().
//
// 2. Enable execution of bar activity
//
//	v := GetVersion(ctx, "fooChange", DefaultVersion, 1)
//	if v == DefaultVersion {
//	    err = workflow.ExecuteActivity(ctx, foo).Get(ctx, nil)
//	} else {
//	    err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//	}
//
// The code above supports replaying of workflow execution with both versions DefaultVersion and 1.
// All new workflow executions will execute bar activity, because
// GetVersion returns the maximum supported version (1), which will be recorded into the workflow history.
//
// 3. Remove a support of foo activity:
//
//	GetVersion(ctx, "fooChange", 1, 1)
//	err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//
// When there are no workflow executions running DefaultVersion the support of foo activity can be removed.
func ExecuteWithVersion(version Version) GetVersionOption {
	return internal.ExecuteWithVersion(version)
}

// ExecuteWithMinVersion forces minSupported version to be returned when GetVersion is executed for the first time,
// instead of returning maxSupported version. The option is equivalent to ExecuteWithVersion(minSupportedVersion).
// Check the ExecuteWithVersion documentation for more details.
func ExecuteWithMinVersion() GetVersionOption {
	return internal.ExecuteWithMinVersion()
}

// GetVersion is used to safely perform backwards incompatible changes to workflow definitions.
// It is not allowed to update workflow code while there are workflows running as it is going to break
// determinism. The solution is to have both old code that is used to replay existing workflows
// as well as the new one that is used when it is executed for the first time.
// GetVersion returns maxSupported version when is executed for the first time. This version is recorded into the
// workflow history as a marker event. Even if maxSupported version is changed the version that was recorded is
// returned on replay. DefaultVersion constant contains version of code that wasn't versioned before.
// Check documentation for ExecuteWithVersion and ExecuteWithMinVersion to make your changes forward and backward compatible.
//
// For example initially workflow has the following code:
//
//	err = workflow.ExecuteActivity(ctx, foo).Get(ctx, nil)
//
// it should be updated to
//
//	err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//
// The backwards compatible way to execute the update is
//
//	v :=  GetVersion(ctx, "fooChange", DefaultVersion, 1)
//	if v  == DefaultVersion {
//	    err = workflow.ExecuteActivity(ctx, foo).Get(ctx, nil)
//	} else {
//	    err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//	}
//
// Then bar has to be changed to baz:
//
//	v :=  GetVersion(ctx, "fooChange", DefaultVersion, 2)
//	if v  == DefaultVersion {
//	    err = workflow.ExecuteActivity(ctx, foo).Get(ctx, nil)
//	} else if v == 1 {
//	    err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//	} else {
//	    err = workflow.ExecuteActivity(ctx, baz).Get(ctx, nil)
//	}
//
// Later when there are no workflow executions running DefaultVersion the correspondent branch can be removed:
//
//	v :=  GetVersion(ctx, "fooChange", 1, 2)
//	if v == 1 {
//	    err = workflow.ExecuteActivity(ctx, bar).Get(ctx, nil)
//	} else {
//	    err = workflow.ExecuteActivity(ctx, baz).Get(ctx, nil)
//	}
//
// It is recommended to keep the GetVersion() call even if single branch is left:
//
//	GetVersion(ctx, "fooChange", 2, 2)
//	err = workflow.ExecuteActivity(ctx, baz).Get(ctx, nil)
//
// The reason to keep it is: 1) it ensures that if there is older version execution still running, it will fail here
// and not proceed; 2) if you ever need to make more changes for “fooChange”, for example change activity from baz to qux,
// you just need to update the maxVersion from 2 to 3.
//
// Note that, you only need to preserve the first call to GetVersion() for each changeID. All subsequent call to GetVersion()
// with same changeID are safe to remove. However, if you really want to get rid of the first GetVersion() call as well,
// you can do so, but you need to make sure: 1) all older version executions are completed; 2) you can no longer use “fooChange”
// as changeID. If you ever need to make changes to that same part like change from baz to qux, you would need to use a
// different changeID like “fooChange-fix2”, and start minVersion from DefaultVersion again. The code would looks like:
//
//	v := workflow.GetVersion(ctx, "fooChange-fix2", workflow.DefaultVersion, 1)
//	if v == workflow.DefaultVersion {
//	  err = workflow.ExecuteActivity(ctx, baz, data).Get(ctx, nil)
//	} else {
//	  err = workflow.ExecuteActivity(ctx, qux, data).Get(ctx, nil)
//	}
func GetVersion(ctx Context, changeID string, minSupported, maxSupported Version, opts ...GetVersionOption) Version {
	return internal.GetVersion(ctx, changeID, minSupported, maxSupported, opts...)
}

// SetQueryHandler sets the query handler to handle workflow query. The queryType specify which query type this handler
// should handle. The handler must be a function that returns 2 values. The first return value must be a serializable
// result. The second return value must be an error. The handler function could receive any number of input parameters.
// All the input parameter must be serializable. You should call workflow.SetQueryHandler() at the beginning of the workflow
// code. When client calls Client.QueryWorkflow() to cadence server, a task will be generated on server that will be dispatched
// to a workflow worker, which will replay the history events and then execute a query handler based on the query type.
// The query handler will be invoked out of the context of the workflow, meaning that the handler code must not use workflow
// context to do things like workflow.NewChannel(), workflow.Go() or to call any workflow blocking functions like
// Channel.Get() or Future.Get(). Trying to do so in query handler code will fail the query and client will receive
// QueryFailedError.
// Example of workflow code that support query type "current_state":
//
//	func MyWorkflow(ctx workflow.Context, input string) error {
//	  currentState := "started" // this could be any serializable struct
//	  err := workflow.SetQueryHandler(ctx, "current_state", func() (string, error) {
//	    return currentState, nil
//	  })
//	  if err != nil {
//	    currentState = "failed to register query handler"
//	    return err
//	  }
//	  // your normal workflow code begins here, and you update the currentState as the code makes progress.
//	  currentState = "waiting timer"
//	  err = NewTimer(ctx, time.Hour).Get(ctx, nil)
//	  if err != nil {
//	    currentState = "timer failed"
//	    return err
//	  }
//
//	  currentState = "waiting activity"
//	  ctx = WithActivityOptions(ctx, myActivityOptions)
//	  err = ExecuteActivity(ctx, MyActivity, "my_input").Get(ctx, nil)
//	  if err != nil {
//	    currentState = "activity failed"
//	    return err
//	  }
//	  currentState = "done"
//	  return nil
//	}
func SetQueryHandler(ctx Context, queryType string, handler interface{}) error {
	return internal.SetQueryHandler(ctx, queryType, handler)
}

// IsReplaying returns whether the current workflow code is replaying.
//
// Warning! Never make decisions, like schedule activity/childWorkflow/timer or send/wait on future/channel, based on
// this flag as it is going to break workflow determinism requirement.
// The only reasonable use case for this flag is to avoid some external actions during replay, like custom logging or
// metric reporting. Please note that Cadence already provide standard logging/metric via workflow.GetLogger(ctx) and
// workflow.GetMetricsScope(ctx), and those standard mechanism are replay-aware and it will automatically suppress during
// replay. Only use this flag if you need custom logging/metrics reporting, for example if you want to log to kafka.
//
// Warning! Any action protected by this flag should not fail or if it does fail should ignore that failure or panic
// on the failure. If workflow don't want to be blocked on those failure, it should ignore those failure; if workflow do
// want to make sure it proceed only when that action succeed then it should panic on that failure. Panic raised from a
// workflow causes decision task to fail and cadence server will rescheduled later to retry.
func IsReplaying(ctx Context) bool {
	return internal.IsReplaying(ctx)
}

// HasLastCompletionResult checks if there is completion result from previous runs.
// This is used in combination with cron schedule. A workflow can be started with an optional cron schedule.
// If a cron workflow wants to pass some data to next schedule, it can return any data and that data will become
// available when next run starts.
// This HasLastCompletionResult() checks if there is such data available passing down from previous successful run.
func HasLastCompletionResult(ctx Context) bool {
	return internal.HasLastCompletionResult(ctx)
}

// GetLastCompletionResult extract last completion result from previous run for this cron workflow.
// This is used in combination with cron schedule. A workflow can be started with an optional cron schedule.
// If a cron workflow wants to pass some data to next schedule, it can return any data and that data will become
// available when next run starts.
// This GetLastCompletionResult() extract the data into expected data structure.
// See TestWorkflowEnvironment.SetLastCompletionResult() for unit test support.
func GetLastCompletionResult(ctx Context, d ...interface{}) error {
	return internal.GetLastCompletionResult(ctx, d...)
}

// UpsertSearchAttributes is used to add or update workflow search attributes.
// The search attributes can be used in query of List/Scan/Count workflow APIs.
// The key and value type must be registered on cadence server side;
// The value has to deterministic when replay;
// The value has to be Json serializable.
// UpsertSearchAttributes will merge attributes to existing map in workflow, for example workflow code:
//
//	  func MyWorkflow(ctx workflow.Context, input string) error {
//		   attr1 := map[string]interface{}{
//			   "CustomIntField": 1,
//			   "CustomBoolField": true,
//		   }
//		   workflow.UpsertSearchAttributes(ctx, attr1)
//
//		   attr2 := map[string]interface{}{
//			   "CustomIntField": 2,
//			   "CustomKeywordField": "seattle",
//		   }
//		   workflow.UpsertSearchAttributes(ctx, attr2)
//	  }
//
// will eventually have search attributes:
//
//	map[string]interface{}{
//		"CustomIntField": 2,
//		"CustomBoolField": true,
//		"CustomKeywordField": "seattle",
//	}
//
// This is only supported when using ElasticSearch.
func UpsertSearchAttributes(ctx Context, attributes map[string]interface{}) error {
	return internal.UpsertSearchAttributes(ctx, attributes)
}
