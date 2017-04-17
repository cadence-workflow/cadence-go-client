package cadence

import (
	"errors"
	"fmt"
	"time"
)

var (
	errActivityParamsBadRequest = errors.New("missing activity parameters through context, check ActivityOptions")
)

type (

	// Channel must be used instead of native go channel by workflow code.
	// Use Context.NewChannel method to create an instance.
	Channel interface {
		Receive(ctx Context) (v interface{})
		ReceiveWithMoreFlag(ctx Context) (v interface{}, more bool)    // more is false when channel is closed
		ReceiveAsync() (v interface{}, ok bool)                        // ok is true when value was returned
		ReceiveAsyncWithMoreFlag() (v interface{}, ok bool, more bool) // ok is true when value was returned, more is false when channel is closed
		Send(ctx Context, v interface{})
		SendAsync(v interface{}) (ok bool) // ok when value was sent
		Close()                            // prohibit sends
	}

	// Selector must be used instead of native go select by workflow code
	// Use Context.NewSelector method to create an instance.
	Selector interface {
		AddReceive(c Channel, f func(v interface{})) Selector
		AddReceiveWithMoreFlag(c Channel, f func(v interface{}, more bool)) Selector
		AddSend(c Channel, v interface{}, f func()) Selector
		AddFuture(future Future, f func(f Future)) Selector
		AddDefault(f func())
		Select(ctx Context)
	}

	// Future represents the result of an asynchronous computation.
	Future interface {
		Get(ctx Context, value interface{}) error
		IsReady() bool
	}

	// Settable is used to set value or error on a future.
	// See NewFuture function.
	Settable interface {
		Set(value interface{}, err error)
		SetValue(value interface{})
		SetError(err error)
		Chain(future Future) // Value (or error) of the future become the same of the chained one.
	}

	// WorkflowType identifies a workflow type.
	WorkflowType struct {
		Name string
	}

	// WorkflowExecution Details.
	WorkflowExecution struct {
		ID    string
		RunID string
	}
)

// RegisterWorkflow - registers a workflow function with the framework.
// A workflow takes a cadence context and input and returns a (result, error) or just error.
// Examples:
//	func sampleWorkflow(ctx cadence.Context, input []byte) (result []byte, err error)
//	func sampleWorkflow(ctx cadence.Context, arg1 int, arg2 string) (result []byte, err error)
//	func sampleWorkflow(ctx cadence.Context) (result []byte, err error)
//	func sampleWorkflow(ctx cadence.Context, arg1 int) (result string, err error)
// Serialization of all primitive types, structures is supported ... except channels, functions, variadic, unsafe pointer.
// This method calls panic if workflowFunc doesn't comply with the expected format.
func RegisterWorkflow(workflowFunc interface{}) {
	thImpl := getHostEnvironment()
	err := thImpl.RegisterWorkflow(workflowFunc)
	if err != nil {
		panic(err)
	}
}

// NewChannel create new Channel instance
func NewChannel(ctx Context) Channel {
	state := getState(ctx)
	state.dispatcher.channelSequence++
	return NewNamedChannel(ctx, fmt.Sprintf("chan-%v", state.dispatcher.channelSequence))
}

// NewNamedChannel create new Channel instance with a given human readable name.
// Name appears in stack traces that are blocked on this channel.
func NewNamedChannel(ctx Context, name string) Channel {
	return &channelImpl{name: name}
}

// NewBufferedChannel create new buffered Channel instance
func NewBufferedChannel(ctx Context, size int) Channel {
	return &channelImpl{size: size}
}

// NewNamedBufferedChannel create new BufferedChannel instance with a given human readable name.
// Name appears in stack traces that are blocked on this Channel.
func NewNamedBufferedChannel(ctx Context, name string, size int) Channel {
	return &channelImpl{name: name, size: size}
}

// NewSelector creates a new Selector instance.
func NewSelector(ctx Context) Selector {
	state := getState(ctx)
	state.dispatcher.selectorSequence++
	return NewNamedSelector(ctx, fmt.Sprintf("selector-%v", state.dispatcher.selectorSequence))
}

// NewNamedSelector creates a new Selector instance with a given human readable name.
// Name appears in stack traces that are blocked on this Selector.
func NewNamedSelector(ctx Context, name string) Selector {
	return &selectorImpl{name: name}
}

// Go creates a new coroutine. It has similar semantic to goroutine in a context of the workflow.
func Go(ctx Context, f func(ctx Context)) {
	state := getState(ctx)
	state.dispatcher.newCoroutine(ctx, f)
}

// GoNamed creates a new coroutine with a given human readable name.
// It has similar semantic to goroutine in a context of the workflow.
// Name appears in stack traces that are blocked on this Channel.
func GoNamed(ctx Context, name string, f func(ctx Context)) {
	state := getState(ctx)
	state.dispatcher.newNamedCoroutine(ctx, name, f)
}

// NewFuture creates a new future as well as associated Settable that is used to set its value.
func NewFuture(ctx Context) (Future, Settable) {
	impl := &futureImpl{channel: NewChannel(ctx)}
	return impl, impl
}

// ExecuteActivityAsync requests activity execution in the context of a workflow.
//  - Context can be used to pass the settings for this activity.
// 	For example: task list that this need to be routed, timeouts that need to be configured.
//	Use ActivityOptions to pass down the options.
//			ctx1 := WithActivityOptions(ctx, NewActivityOptions().
//					WithTaskList("exampleTaskList").
//					WithScheduleToCloseTimeout(time.Second).
//					WithScheduleToStartTimeout(time.Second).
//					WithHeartbeatTimeout(0)
//			or
//			ctx1 := WithTaskList(ctx, "exampleTaskList")
//  - f - Either a activity name or a function that is getting scheduled.
//  - args - The arguments that need to be passed to the function represented by 'f'.
//  - If the activity failed to complete then the future get error would indicate the failure
// and it can be one of ActivityTaskFailedError, ActivityTaskTimeoutError, ActivityTaskCanceledError.
//  - You can also cancel the pending activity using context(WithCancel(ctx)) and that will fail the activity with
// error ActivityTaskCanceledError.
// - result - Result the activity returns, if there is no result the activity returns
//      then it will be nil, indicating no result.
func ExecuteActivity(ctx Context, f interface{}, args ...interface{}) Future {
	// Validate type and its arguments.
	future, settable := NewFuture(ctx)
	activityType, input, err := getValidatedActivityFunction(f, args)
	if err != nil {
		settable.Set(nil, err)
		return future
	}
	// Validate context options.
	parameters := getActivityOptions(ctx)
	parameters, err = getValidatedActivityOptions(ctx)
	if err != nil {
		settable.Set(nil, err)
		return future
	}
	parameters.ActivityType = *activityType
	parameters.Input = input

	a := getWorkflowEnvironment(ctx).ExecuteActivity(*parameters, func(r []byte, e error) {
		result, serializeErr := deSerializeFunctionResult(f, r)
		if serializeErr != nil {
			e = serializeErr
		}
		settable.Set(result, e)
		executeDispatcher(ctx, getDispatcher(ctx))
	})
	Go(ctx, func(ctx Context) {
		if ctx.Done() == nil {
			return // not cancellable.
		}
		if ctx.Done().Receive(ctx); ctx.Err() == ErrCanceled {
			getWorkflowEnvironment(ctx).RequestCancelActivity(a.activityID)
		}
	})
	return future
}

// WorkflowInfo information about currently executing workflow
type WorkflowInfo struct {
	WorkflowExecution WorkflowExecution
	WorkflowType      WorkflowType
	TaskListName      string
}

// GetWorkflowInfo extracts info of a current workflow from a context.
func GetWorkflowInfo(ctx Context) *WorkflowInfo {
	return getWorkflowEnvironment(ctx).WorkflowInfo()
}

// Now returns the current time when the decision is started or replayed.
// The workflow needs to use this Now() to get the wall clock time instead of the Go lang library one.
func Now(ctx Context) time.Time {
	return getWorkflowEnvironment(ctx).Now()
}

// NewTimer returns immediately and the future becomes ready after the specified timeout.
//  - The current timer resolution implementation is in seconds but is subjected to change.
//  - The workflow needs to use this NewTimer() to get the timer instead of the Go lang library one(timer.NewTimer())
//  - You can also cancel the pending timer using context(WithCancel(ctx)) and that will cancel the timer with
// error TimerCanceledError.
func NewTimer(ctx Context, d time.Duration) Future {
	future, settable := NewFuture(ctx)
	if d <= 0 {
		settable.Set(true, nil)
		return future
	}

	t := getWorkflowEnvironment(ctx).NewTimer(d, func(r []byte, e error) {
		settable.Set(nil, e)
		executeDispatcher(ctx, getDispatcher(ctx))
	})
	if t != nil {
		Go(ctx, func(ctx Context) {
			if ctx.Done() == nil {
				return // not cancellable.
			}
			// We will cancel the timer either it is explicit cancellation
			// (or) we are closed.
			ctx.Done().Receive(ctx)
			getWorkflowEnvironment(ctx).RequestCancelTimer(t.timerID)
		})
	}
	return future
}

// Sleep pauses the current goroutine for at least the duration d.
// A negative or zero duration causes Sleep to return immediately.
//  - The current timer resolution implementation is in seconds but is subjected to change.
//  - The workflow needs to use this Sleep() to sleep instead of the Go lang library one(timer.Sleep())
//  - You can also cancel the pending sleep using context(WithCancel(ctx)) and that will cancel the sleep with
//    error TimerCanceledError.
func Sleep(ctx Context, d time.Duration) (err error) {
	t := NewTimer(ctx, d)
	err = t.Get(ctx, nil)
	return
}
