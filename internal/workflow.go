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

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/zap"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
)

var (
	errDomainNotSet                  = errors.New("domain is not set")
	errTaskListNotSet                = errors.New("task list is not set")
	errWorkflowIDNotSet              = errors.New("workflowId is not set")
	errLocalActivityParamsBadRequest = errors.New("missing local activity parameters through context, check LocalActivityOptions")
	errActivityParamsBadRequest      = errors.New("missing activity parameters through context, check ActivityOptions")
	errWorkflowOptionBadRequest      = errors.New("missing workflow options through context, check WorkflowOptions")
	errSearchAttributesNotSet        = errors.New("search attributes is empty")
)

type (
	// Channel must be used in workflows instead of a native Go chan.
	//
	// Use workflow.NewChannel(ctx) to create an unbuffered Channel instance,
	// workflow.NewBufferedChannel(ctx, size) to create a Channel which has a buffer,
	// or workflow.GetSignalChannel(ctx, "name") to get a Channel that contains data sent to this workflow by a call to
	// SignalWorkflow (e.g. on the Client, or similar methods like SignalExternalWorkflow or SignalChildWorkflow).
	//
	// Both NewChannel and NewBufferedChannel have "Named" constructors as well.
	// These names will be visible in stack-trace queries, so they can help with debugging, but they do not otherwise
	// impact behavior at all, and are not recorded anywhere (so you can change them without versioning your code).
	//
	// Also note that channels created by NewChannel and NewBufferedChannel do not do any serialization or
	// deserialization - you will receive whatever value was sent, and non-(de)serializable values like function
	// references and interfaces are fine, the same as using a normal Go channel.
	//
	// Signal channels, however, contain whatever bytes were sent to your workflow, and the values must be decoded into
	// the output value.  By default, this means that Receive(ctx, &out) will use json.Unmarshal(data, &out), but this
	// can be overridden at a worker level (worker.Options) or at a context level (workflow.WithDataConverter(ctx, dc)).
	//
	// You are able to send values to your own signal channels, and these values will behave the same as they do in
	// normal channels (i.e. they will not be (de)serialized).  However, doing so is not generally recommended, as
	// mixing the value types can increase the risk that you fail to read a value, causing values to be lost.  See
	// Receive for more details about that behavior.
	Channel interface {
		// Receive blocks until it receives a value, and then assigns the received value to the provided pointer.
		// It returns false when the Channel is closed and all data has already been consumed from the Channel, in the
		// same way as Go channel reads work, but the assignment only occurs if there was a value in the Channel.
		//
		// This is technically equivalent to:
		//  received, ok := <- aChannel:
		//  if ok {
		//  	*valuePtr = received
		//  }
		//
		// But if your output values are zero values, this is equivalent to a normal channel read:
		//  value, ok <- aChannel
		//
		// valuePtr must be assignable, and will be used to assign (for in-memory data in regular channels) or decode
		// (for signal channels) the data in the channel.
		//
		// If decoding or assigning fails:
		// - an error will be logged
		// - the value will be dropped from the channel
		// - Receive will automatically try again
		// - This will continue until a successful value is found, or the channel is emptied and it resumes blocking.
		//   Closed channels with no values will always succeed, but they will not change valuePtr.
		//
		// Go would normally prevent incorrect-type failures like this at compile time, but the same cannot be done
		// here.  If you need to "try" to assign to multiple things, similar to a Future you can use:
		// - for signal channels, a []byte pointer.  This will give you the raw data that Cadence received, and no
		//   decoding will be attempted, so you can try it yourself.
		// - for other channels, an interface{} pointer.  All values are interfaces, so this will never fail, and you
		//   can inspect the type with reflection or type assertions.
		Receive(ctx Context, valuePtr interface{}) (ok bool)

		// ReceiveAsync tries to Receive from Channel without blocking.
		// If there is data available from the Channel, it assigns the data to valuePtr and returns true.
		// Otherwise, it returns false immediately.
		//
		// This is technically equivalent to:
		//  select {
		//  case received, ok := <- aChannel:
		//  	if ok {
		//  		*valuePtr = received
		//  	}
		//  default:
		//  	// no value was read
		//  	ok = false
		//  }
		//
		// But if your output values are zero values, this is equivalent to a simpler form:
		//  select {
		//  case value, ok := <- aChannel:
		//  default:
		//  	// no value was read
		//  	ok = false
		//  }
		//
		// Decoding or assigning failures are handled like Receive.
		ReceiveAsync(valuePtr interface{}) (ok bool)

		// ReceiveAsyncWithMoreFlag is the same as ReceiveAsync, with an extra return to indicate if there could be
		// more values from the Channel in the future.
		// `more` is false only when Channel is closed and the read failed (empty).
		//
		// This is technically equivalent to:
		//  select {
		//  case received, ok := <- aChannel:
		//  	if ok {
		//  		*valuePtr = received
		//  	}
		//  	more = ok
		//  default:
		//  	// no value was read
		//  	ok = false
		//  	// but the read would have blocked, so the channel is not closed
		//  	more = true
		//  }
		//
		// But if your output values are zero values, this is equivalent to a simpler form:
		//  select {
		//  case value, ok := <- aChannel:
		//  	more = ok
		//  default:
		//  	// no value was read
		//  	ok = false
		//  	// but the read would have blocked, so the channel is not closed
		//  	more = true
		//  }
		//
		// Decoding or assigning failures are handled like Receive.
		ReceiveAsyncWithMoreFlag(valuePtr interface{}) (ok bool, more bool)

		// Send blocks until the data is sent.
		//
		// This is equivalent to `aChannel <- v`.
		Send(ctx Context, v interface{})

		// SendAsync will try to send without blocking.
		// It returns true if the data was sent (i.e. there was room in the buffer, or a reader was waiting to receive
		// it), otherwise it returns false.
		//
		// This is equivalent to:
		//  select {
		//  case aChannel <- v: ok = true
		//  default: ok = false
		//  }
		SendAsync(v interface{}) (ok bool)

		// Close closes the Channel, and prohibits subsequent sends.
		// As with a normal Go channel that has been closed, sending to a closed channel will panic.
		Close()
	}

	// Selector must be used in workflows instead of a native Go select statement.
	//
	// Use workflow.NewSelector(ctx) to create a Selector instance, and then add cases to it with its methods.
	// The interface is intended to simulate Go's select statement, and any Go select can be fairly trivially rewritten
	// for a Selector with effectively identical behavior.
	//
	// For example, normal Go code like below (which will receive values forever, until idle for an hour):
	//  chA := make(chan int)
	//  chB := make(chan int)
	//  counter := 0
	//  for {
	//  	select {
	//  	case x := <- chA:
	//  		counter += i
	//  	case y := <- chB:
	//  		counter += i
	//  	case <- time.After(time.Hour):
	//  		break
	//  	}
	//  }
	// can be written as:
	//  chA := workflow.NewChannel(ctx)
	//  chB := workflow.NewChannel(ctx)
	//  counter := 0
	//  for {
	//  	timedout := false
	//  	s := workflow.NewSelector(ctx)
	//  	s.AddReceive(chA, func(c workflow.Channel, more bool) {
	//  		var x int
	//  		c.Receive(ctx, &x)
	//  		counter += i
	//  	})
	//  	s.AddReceive(chB, func(c workflow.Channel, more bool) {
	//  		var y int
	//  		c.Receive(ctx, &y)
	//  		counter += i
	//  	})
	//  	s.AddFuture(workflow.NewTimer(ctx, time.Hour), func(f workflow.Future) {
	//  		timedout = true
	//  	})
	//  	s.Select(ctx)
	//  	if timedout {
	//  		break
	//  	}
	//  }
	//
	// You can create a new Selector as needed or mutate one and call Select multiple times, but note that:
	//
	// 1. AddFuture will not behave the same across both patterns.  Read AddFuture for more details.
	//
	// 2. There is no way to remove a case from a Selector, so you must make a new Selector to "remove" them.
	//
	// Finally, note that Select will not return until a condition's needs are met, like a Go selector - canceling the
	// Context used to construct the Selector, or the Context used to Select, will not (directly) unblock a Select call.
	// Read Select for more details.
	Selector interface {
		// AddReceive waits until a value can be received from a channel.
		// f is invoked when the channel has data or is closed.
		//
		// This is equivalent to `case v, ok := <- aChannel`, and `ok` will only be false when
		// the channel is both closed and no data was received.
		//
		// When f is invoked, the data (or closed state) remains untouched in the channel, so
		// you need to `c.Receive(ctx, &out)` (or `c.ReceiveAsync(&out)`) to remove and decode the value.
		// Failure to do this is not an error - the value will simply remain in the channel until a future
		// Receive retrieves it.
		//
		// The `ok` argument will match what a call to c.Receive would return (on a successful read), so it
		// may be used to check for closed + empty channels without needing to try to read from the channel.
		// See Channel.Receive for additional details about reading from channels.
		AddReceive(c Channel, f func(c Channel, ok bool)) Selector
		// AddSend waits to send a value to a channel.
		// f is invoked when the value was successfully sent to the channel.
		//
		// This is equivalent to `case aChannel <- value`.
		//
		// Unlike AddReceive, the value has already been sent on the channel when f is invoked.
		AddSend(c Channel, v interface{}, f func()) Selector
		// AddFuture waits until a Future is ready, and then invokes f only once.
		// If the Future is ready before Select is called, it is eligible to be invoked immediately.
		//
		// There is no direct equivalent in a native Go select statement.
		// It was added because Futures are common in Cadence code, and some patterns are much simpler with it.
		//
		// Each call to AddFuture will invoke its f at most one time, regardless of how many times Select is called.
		// This means, for a Future that is (or will be) ready:
		// - Adding the Future once, then calling Select twice, will invoke the callback once with the first Select
		//   call, and then wait for other Selector conditions in the second Select call (or block forever if there are
		//   no other eligible conditions).
		// - Adding the same Future twice, then calling Select twice, will invoke each callback once.
		// - Adding the same Future to two different Selectors, then calling Select once on each Selector, will invoke
		//   each Selector's callback once.
		//
		// Therefore, with a Future "f" that is or will become ready, this is an infinite loop that will consume as much
		// CPU as possible:
		//  for {
		//  	workflow.NewSelector(ctx).AddFuture(f, func(f workflow.Future){}).Select(ctx)
		//  }
		// While this will loop once, and then wait idle forever:
		//  s := workflow.NewSelector(ctx).AddFuture(f, func(f workflow.Future){})
		//  for {
		//  	s.Select(ctx)
		//  }
		AddFuture(future Future, f func(f Future)) Selector
		// AddDefault adds a default branch to the selector.
		// f is invoked immediately when none of the other conditions (AddReceive, AddSend, AddFuture) are met for a
		// Select call.
		//
		// This is equivalent to a `default:` case.
		//
		// Note that this applies to each Select call.  If you create a Selector with only one AddDefault, and then call
		// Select on it twice, f will be invoked twice.
		AddDefault(f func())
		// Select waits for one of the added conditions to be met and invokes the callback as described above.
		// If no condition is met, Select will block until one or more are available, then one callback will be invoked.
		// If no condition is ever met, Select will block forever.
		//
		// Note that Select does not return an error, and does not stop waiting if its Context is canceled.
		// This mimics a native Go select statement, which has no way to be interrupted except for its listed cases.
		//
		// If you wish to stop Selecting when the Context is canceled, use AddReceive with the Context's Done() channel,
		// in the same way as you would use a `case <- ctx.Done():` in a Go select statement.  E.g.:
		//  cancelled := false
		//  s := workflow.NewSelector(ctx)
		//  s.AddFuture(f, func(f workflow.Future) {}) // assume this is never ready
		//  s.AddReceive(ctx.Done(), func(c workflow.Channel, more bool) {
		//  	// this will be invoked when the Context is cancelled for any reason,
		//  	// and more will be false.
		//  	cancelled = true
		//  })
		//  s.Select(ctx)
		//  if cancelled {
		//  	// this will be executed
		//  }
		Select(ctx Context)
	}

	// WaitGroup must be used instead of native go sync.WaitGroup by
	// workflow code.  Use workflow.NewWaitGroup(ctx) method to create
	// a new WaitGroup instance
	WaitGroup interface {
		Add(delta int)
		Done()
		Wait(ctx Context)
	}

	// Future represents the result of an asynchronous computation.
	Future interface {
		// Get blocks until the future is ready.
		// When ready it either returns the Future's contained error, or assigns the contained value to the output var.
		// Failures to assign or decode the value will panic.
		//
		// Two common patterns to retrieve data are:
		//  var out string
		//  // this will assign the string value, which may be "", or an error and leave out as "".
		//  err := f.Get(ctx, &out)
		// and
		//  var out *string
		//  // this will assign the string value, which may be "" or nil, or an error and leave out as nil.
		//  err := f.Get(ctx, &out)
		//
		// The valuePtr parameter can be nil when the encoded result value is not needed:
		//  err := f.Get(ctx, nil)
		//
		// Futures with values set in-memory via a call to their Settable's methods can be retrieved without knowing the
		// type with an interface, i.e. this will not ever panic:
		//  var out interface{}
		//  // this will assign the same value that was set,
		//  // and you can check its type with reflection or type assertions.
		//  err := f.Get(ctx, &out)
		//
		// Futures with encoded data from e.g. activities or child workflows can bypass decoding with a byte slice, and
		// similarly this will not ever panic:
		//  var out []byte
		//  // out will contain the raw bytes given to Cadence's servers, you should decode it however is necessary
		//  err := f.Get(ctx, &out) // err can only be the Future's contained error
		Get(ctx Context, valuePtr interface{}) error

		// IsReady will return true Get is guaranteed to not block.
		IsReady() bool
	}

	// Settable is used to set value or error on a future.
	// See more: workflow.NewFuture(ctx).
	Settable interface {
		Set(value interface{}, err error)
		SetValue(value interface{})
		SetError(err error)
		Chain(future Future) // Value (or error) of the future become the same of the chained one.
	}

	// ChildWorkflowFuture represents the result of a child workflow execution
	ChildWorkflowFuture interface {
		Future
		// GetChildWorkflowExecution returns a future that will be ready when child workflow execution started. You can
		// get the WorkflowExecution of the child workflow from the future. Then you can use Workflow ID and RunID of
		// child workflow to cancel or send signal to child workflow.
		//  childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, child, ...)
		//  var childWE WorkflowExecution
		//  if err := childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childWE); err == nil {
		//      // child workflow started, you can use childWE to get the WorkflowID and RunID of child workflow
		//  }
		GetChildWorkflowExecution() Future

		// SignalWorkflowByID sends a signal to the child workflow. This call will block until child workflow is started.
		SignalChildWorkflow(ctx Context, signalName string, data interface{}) Future
	}

	// RegistryWorkflowInfo
	RegistryWorkflowInfo interface {
		WorkflowType() WorkflowType
		GetFunction() interface{}
	}

	// WorkflowType identifies a workflow type.
	WorkflowType struct {
		Name string
		Path string
	}

	// WorkflowExecution Details.
	WorkflowExecution struct {
		ID    string
		RunID string
	}

	// WorkflowExecutionAsync Details.
	WorkflowExecutionAsync struct {
		ID string
	}

	// EncodedValue is type alias used to encapsulate/extract encoded result from workflow/activity.
	EncodedValue struct {
		value         []byte
		dataConverter DataConverter
	}
	// Version represents a change version. See GetVersion call.
	Version int

	// ChildWorkflowOptions stores all child workflow specific parameters that will be stored inside of a Context.
	// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
	// subjected to change in the future.
	ChildWorkflowOptions struct {
		// Domain of the child workflow.
		// Optional: the current workflow (parent)'s domain will be used if this is not provided.
		Domain string

		// WorkflowID of the child workflow to be scheduled.
		// Optional: an auto generated workflowID will be used if this is not provided.
		WorkflowID string

		// TaskList that the child workflow needs to be scheduled on.
		// Optional: the parent workflow task list will be used if this is not provided.
		TaskList string

		// ExecutionStartToCloseTimeout - The end to end timeout for the child workflow execution.
		// Mandatory: no default
		ExecutionStartToCloseTimeout time.Duration

		// TaskStartToCloseTimeout - The decision task timeout for the child workflow.
		// Optional: default is 10s if this is not provided (or if 0 is provided).
		TaskStartToCloseTimeout time.Duration

		// WaitForCancellation - Whether to wait for cancelled child workflow to be ended (child workflow can be ended
		// as: completed/failed/timedout/terminated/canceled)
		// Optional: default false
		WaitForCancellation bool

		// WorkflowIDReusePolicy - Whether server allow reuse of workflow ID, can be useful
		// for dedup logic if set to WorkflowIdReusePolicyRejectDuplicate
		WorkflowIDReusePolicy WorkflowIDReusePolicy

		// RetryPolicy specify how to retry child workflow if error happens.
		// Optional: default is no retry
		RetryPolicy *RetryPolicy

		// CronSchedule - Optional cron schedule for workflow. If a cron schedule is specified, the workflow will run
		// as a cron based on the schedule. The scheduling will be based on UTC time. Schedule for next run only happen
		// after the current run is completed/failed/timeout. If a RetryPolicy is also supplied, and the workflow failed
		// or timeout, the workflow will be retried based on the retry policy. While the workflow is retrying, it won't
		// schedule its next run. If next schedule is due while workflow is running (or retrying), then it will skip that
		// schedule. Cron workflow will not stop until it is terminated or cancelled (by returning cadence.CanceledError).
		// The cron spec is as following:
		// ┌───────────── minute (0 - 59)
		// │ ┌───────────── hour (0 - 23)
		// │ │ ┌───────────── day of the month (1 - 31)
		// │ │ │ ┌───────────── month (1 - 12)
		// │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
		// │ │ │ │ │
		// │ │ │ │ │
		// * * * * *
		CronSchedule string

		// Memo - Optional non-indexed info that will be shown in list workflow.
		Memo map[string]interface{}

		// SearchAttributes - Optional indexed info that can be used in query of List/Scan/Count workflow APIs (only
		// supported when Cadence server is using ElasticSearch). The key and value type must be registered on Cadence server side.
		// Use GetSearchAttributes API to get valid key and corresponding value type.
		SearchAttributes map[string]interface{}

		// ParentClosePolicy - Optional policy to decide what to do for the child.
		// Default is Terminate (if onboarded to this feature)
		ParentClosePolicy ParentClosePolicy

		// Bugports allows opt-in enabling of older, possibly buggy behavior, primarily intended to allow temporarily
		// emulating old behavior until a fix is deployed.
		//
		// Bugports are always deprecated and may be removed in future versions.
		// Generally speaking they will *likely* remain in place for one minor version, and then they may be removed to
		// allow cleaning up the additional code complexity that they cause.
		//
		// Deprecated: All bugports are always deprecated and may be removed at any time.
		Bugports Bugports
	}

	// Bugports allows opt-in enabling of older, possibly buggy behavior, primarily intended to allow temporarily
	// emulating old behavior until a fix is deployed.
	// By default, bugs (especially rarely-occurring ones) are fixed and all users are opted into the new behavior.
	// Back-ported buggy behavior *may* be available via these flags.
	//
	// Fields in here are NOT guaranteed to be stable.  They will almost certainly be removed in the next major
	// release, and might be removed earlier if a need arises, e.g. if the historical behavior causes too much of an
	// increase in code complexity.
	//
	// See each individual field for details.
	//
	// Bugports are always deprecated and may be removed in future versions.
	// Generally speaking they will *likely* remain in place for one minor version, and then they may be removed to
	// allow cleaning up the additional code complexity that they cause.
	//
	// DEPRECATED: All bugports are always deprecated and may be removed at any time.
	Bugports struct {
		// StartChildWorkflowsOnCanceledContext allows emulating older, buggy behavior that existed prior to v0.18.4.
		//
		// Prior to the fix, child workflows would be started and keep running when their context was canceled in two
		// situations:
		// 1) when the context was canceled before ExecuteChildWorkflow is called, and
		// 2) when the context was canceled after ExecuteChildWorkflow but before the child workflow was started.
		//
		// 1 is unfortunately easy to trigger, though many workflows will encounter an error earlier and not reach the
		// child-workflow-executing code.  2 is expected to be very rare in practice.
		//
		// To permanently emulate old behavior, use a disconnected context when starting child workflows, and
		// cancel it only after `childfuture.GetWorkflowExecution().Get(...)` returns.  This can be used when this flag
		// is removed in the future.
		//
		// If you have currently-broken workflows and need to repair them, there are two primary options:
		//
		// 1: Check the BinaryChecksum value of your new deploy and/or of the decision that is currently failing
		// workflows.  Then set this flag when replaying history on those not-fixed checksums.  Concretely, this means
		// checking both `workflow.GetInfo(ctx).BinaryChecksum` (note that sufficiently old clients may not have
		// recorded a value, and it may be nil) and `workflow.IsReplaying(ctx)`.
		//
		// 2: Reset broken workflows back to either before the buggy behavior was recorded, or before the fixed behavior
		// was deployed.  A "bad binary" reset type can do the latter in bulk, see the CLI's
		// `cadence workflow reset-batch --reset_type BadBinary --help` for details.  For the former, check the failing
		// histories, identify the point at which the bug occurred, and reset to prior to that decision task.
		//
		// Added in 0.18.4, this may be removed in or after v0.19.0, so please migrate off of it ASAP.
		//
		// Deprecated: All bugports are always deprecated and may be removed at any time.
		StartChildWorkflowsOnCanceledContext bool
	}
)

// RegisterWorkflowOptions consists of options for registering a workflow
type RegisterWorkflowOptions struct {
	Name string
	// Workflow type name is equal to function name instead of fully qualified name including function package.
	// This option has no effect when explicit Name is provided.
	EnableShortName               bool
	DisableAlreadyRegisteredCheck bool
}

// RegisterWorkflow - registers a workflow function with the framework.
// The public form is: workflow.Register(...)
// A workflow takes a cadence context and input and returns a (result, error) or just error.
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
func RegisterWorkflow(workflowFunc interface{}) {
	RegisterWorkflowWithOptions(workflowFunc, RegisterWorkflowOptions{})
}

// RegisterWorkflowWithOptions registers the workflow function with options.
// The public form is: workflow.RegisterWithOptions(...)
// The user can use options to provide an external name for the workflow or leave it empty if no
// external name is required. This can be used as
//
//	workflow.RegisterWithOptions(sampleWorkflow, RegisterWorkflowOptions{})
//	workflow.RegisterWithOptions(sampleWorkflow, RegisterWorkflowOptions{Name: "foo"})
//
// A workflow takes a cadence context and input and returns a (result, error) or just error.
// Examples:
//
//	func sampleWorkflow(ctx workflow.Context, input []byte) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context, arg1 int, arg2 string) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context) (result []byte, err error)
//	func sampleWorkflow(ctx workflow.Context, arg1 int) (result string, err error)
//
// Serialization of all primitive types, structures is supported ... except channels, functions, variadic, unsafe pointer.
// This method calls panic if workflowFunc doesn't comply with the expected format or tries to register the same workflow
// type name twice. Use workflow.RegisterOptions.DisableAlreadyRegisteredCheck to allow multiple registrations.
// Deprecated: Global workflow registration methods are replaced by equivalent Worker instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func RegisterWorkflowWithOptions(workflowFunc interface{}, opts RegisterWorkflowOptions) {
	registry := getGlobalRegistry()
	registry.RegisterWorkflowWithOptions(workflowFunc, opts)
}

// GetRegisteredWorkflowTypes returns the registered workflow function/alias names.
// The public form is: workflow.GetRegisteredWorkflowTypes(...)
func GetRegisteredWorkflowTypes() []string {
	registry := getGlobalRegistry()
	return registry.GetRegisteredWorkflowTypes()
}

// Await blocks the calling thread until condition() returns true
// Returns CanceledError if the ctx is canceled.
func Await(ctx Context, condition func() bool) error {
	state := getState(ctx)
	defer state.unblocked()

	for !condition() {
		doneCh := ctx.Done()
		// TODO: Consider always returning a channel
		if doneCh != nil {
			if _, more := doneCh.ReceiveAsyncWithMoreFlag(nil); !more {
				return NewCanceledError("Await context cancelled")
			}
		}
		state.yield("Await")
	}
	return nil
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
	env := getWorkflowEnvironment(ctx)
	return &channelImpl{name: name, dataConverter: getDataConverterFromWorkflowContext(ctx), env: env}
}

// NewBufferedChannel create new buffered Channel instance
func NewBufferedChannel(ctx Context, size int) Channel {
	env := getWorkflowEnvironment(ctx)
	return &channelImpl{size: size, dataConverter: getDataConverterFromWorkflowContext(ctx), env: env}
}

// NewNamedBufferedChannel create new BufferedChannel instance with a given human readable name.
// Name appears in stack traces that are blocked on this Channel.
func NewNamedBufferedChannel(ctx Context, name string, size int) Channel {
	env := getWorkflowEnvironment(ctx)
	return &channelImpl{name: name, size: size, dataConverter: getDataConverterFromWorkflowContext(ctx), env: env}
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

// NewWaitGroup creates a new WaitGroup instance.
func NewWaitGroup(ctx Context) WaitGroup {
	f, s := NewFuture(ctx)
	return &waitGroupImpl{future: f, settable: s}
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
	impl := &futureImpl{channel: NewChannel(ctx).(*channelImpl)}
	return impl, impl
}

func (wc *workflowEnvironmentInterceptor) ExecuteWorkflow(ctx Context, workflowType string, inputArgs ...interface{}) (results []interface{}) {
	args := []reflect.Value{reflect.ValueOf(ctx)}
	for _, arg := range inputArgs {
		// []byte arguments are not serialized
		switch arg.(type) {
		case []byte:
			args = append(args, reflect.ValueOf(arg))
		default:
			args = append(args, reflect.ValueOf(arg).Elem())
		}
	}
	fnValue := reflect.ValueOf(wc.fn)
	retValues := fnValue.Call(args)
	for _, r := range retValues {
		results = append(results, r.Interface())
	}
	return
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
	i := getWorkflowInterceptor(ctx)
	registry := getRegistryFromWorkflowContext(ctx)
	activityType := getActivityFunctionName(registry, activity)
	return i.ExecuteActivity(ctx, activityType, args...)
}

func (wc *workflowEnvironmentInterceptor) ExecuteActivity(ctx Context, typeName string, args ...interface{}) Future {
	// Validate type and its arguments.
	dataConverter := getDataConverterFromWorkflowContext(ctx)
	registry := getRegistryFromWorkflowContext(ctx)
	future, settable := newDecodeFuture(ctx, typeName)
	activityType, err := getValidatedActivityFunction(typeName, args, registry)
	if err != nil {
		settable.Set(nil, err)
		return future
	}
	// Validate context options.
	options, err := getValidatedActivityOptions(ctx)
	if err != nil {
		settable.Set(nil, err)
		return future
	}

	// Validate session state.
	if sessionInfo := getSessionInfo(ctx); sessionInfo != nil {
		isCreationActivity := isSessionCreationActivity(typeName)
		if sessionInfo.sessionState == sessionStateFailed && !isCreationActivity {
			settable.Set(nil, ErrSessionFailed)
			return future
		}
		if sessionInfo.sessionState == sessionStateOpen && !isCreationActivity {
			// Use session tasklist
			oldTaskListName := options.TaskListName
			options.TaskListName = sessionInfo.tasklist
			defer func() {
				options.TaskListName = oldTaskListName
			}()
		}
	}

	// Retrieve headers from context to pass them on
	header := getHeadersFromContext(ctx)

	input, err := encodeArgs(dataConverter, args)
	if err != nil {
		panic(err)
	}

	params := executeActivityParams{
		activityOptions: *options,
		ActivityType:    *activityType,
		Input:           input,
		DataConverter:   dataConverter,
		Header:          header,
	}

	ctxDone, cancellable := ctx.Done().(*channelImpl)
	cancellationCallback := &receiveCallback{}
	a := getWorkflowEnvironment(ctx).ExecuteActivity(params, func(r []byte, e error) {
		settable.Set(r, e)
		if cancellable {
			// future is done, we don't need the cancellation callback anymore.
			ctxDone.removeReceiveCallback(cancellationCallback)
		}
	})

	if cancellable {
		cancellationCallback.fn = func(v interface{}, more bool) bool {
			if ctx.Err() == ErrCanceled {
				wc.env.RequestCancelActivity(a.activityID)
			}
			return false
		}
		_, ok, more := ctxDone.receiveAsyncImpl(cancellationCallback)
		if ok || !more {
			cancellationCallback.fn(nil, more)
		}
	}
	return future
}

// ExecuteLocalActivity requests to run a local activity. A local activity is like a regular activity with some key
// differences:
// * Local activity is scheduled and run by the workflow worker locally.
// * Local activity does not need Cadence server to schedule activity task and does not rely on activity worker.
// * No need to register local activity.
// * The parameter activity to ExecuteLocalActivity() must be a function.
// * Local activity is for short living activities (usually finishes within seconds).
// * Local activity cannot heartbeat.
//
// Context can be used to pass the settings for this local activity.
// For now there is only one setting for timeout to be set:
//
//	 lao := LocalActivityOptions{
//		    ScheduleToCloseTimeout: 5 * time.Second,
//		}
//		ctx := WithLocalActivityOptions(ctx, lao)
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
	i := getWorkflowInterceptor(ctx)
	env := getWorkflowEnvironment(ctx)
	activityType := getActivityFunctionName(env.GetRegistry(), activity)
	ctx = WithValue(ctx, localActivityFnContextKey, activity)
	return i.ExecuteLocalActivity(ctx, activityType, args...)
}

func (wc *workflowEnvironmentInterceptor) ExecuteLocalActivity(ctx Context, activityType string, args ...interface{}) Future {
	header := getHeadersFromContext(ctx)
	activityFn := ctx.Value(localActivityFnContextKey)
	if activityFn == nil {
		panic("ExecuteLocalActivity: Expected context key " + localActivityFnContextKey + " is missing")
	}

	future, settable := newDecodeFuture(ctx, activityFn)
	if err := validateFunctionArgs(activityFn, args, false); err != nil {
		settable.Set(nil, err)
		return future
	}
	options, err := getValidatedLocalActivityOptions(ctx)
	if err != nil {
		settable.Set(nil, err)
		return future
	}
	params := &executeLocalActivityParams{
		localActivityOptions: *options,
		ActivityFn:           activityFn,
		ActivityType:         activityType,
		InputArgs:            args,
		WorkflowInfo:         GetWorkflowInfo(ctx),
		DataConverter:        getDataConverterFromWorkflowContext(ctx),
		ScheduledTime:        Now(ctx), // initial scheduled time
		Header:               header,
	}

	Go(ctx, func(ctx Context) {
		for {
			f := wc.scheduleLocalActivity(ctx, params)
			var result []byte
			err := f.Get(ctx, &result)
			if retryErr, ok := err.(*needRetryError); ok && retryErr.Backoff > 0 {
				// Backoff for retry
				Sleep(ctx, retryErr.Backoff)
				// increase the attempt, and retry the local activity
				params.Attempt = retryErr.Attempt + 1
				continue
			}

			// not more retry, return whatever is received.
			settable.Set(result, err)
			return
		}
	})

	return future
}

type needRetryError struct {
	Backoff time.Duration
	Attempt int32
}

func (e *needRetryError) Error() string {
	return fmt.Sprintf("Retry backoff: %v, Attempt: %v", e.Backoff, e.Attempt)
}

func (wc *workflowEnvironmentInterceptor) scheduleLocalActivity(ctx Context, params *executeLocalActivityParams) Future {
	f := &futureImpl{channel: NewChannel(ctx).(*channelImpl)}
	ctxDone, cancellable := ctx.Done().(*channelImpl)
	cancellationCallback := &receiveCallback{}
	la := wc.env.ExecuteLocalActivity(*params, func(lar *localActivityResultWrapper) {
		if cancellable {
			// future is done, we don't need cancellation anymore
			ctxDone.removeReceiveCallback(cancellationCallback)
		}

		if lar.err == nil || IsCanceledError(lar.err) || lar.backoff <= 0 {
			f.Set(lar.result, lar.err)
			return
		}

		// set retry error, and it will be handled by workflow.ExecuteLocalActivity().
		f.Set(nil, &needRetryError{Backoff: lar.backoff, Attempt: lar.attempt})
		return
	})

	if cancellable {
		cancellationCallback.fn = func(v interface{}, more bool) bool {
			if ctx.Err() == ErrCanceled {
				getWorkflowEnvironment(ctx).RequestCancelLocalActivity(la.activityID)
			}
			return false
		}
		_, ok, more := ctxDone.receiveAsyncImpl(cancellationCallback)
		if ok || !more {
			cancellationCallback.fn(nil, more)
		}
	}

	return f
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
//	 ctx := WithChildWorkflowOptions(ctx, cwo)
//
// Input childWorkflow is either a workflow name or a workflow function that is getting scheduled.
// Input args are the arguments that need to be passed to the child workflow function represented by childWorkflow.
// If the child workflow failed to complete then the future get error would indicate the failure and it can be one of
// CustomError, TimeoutError, CanceledError, GenericError.
// You can cancel the pending child workflow using context(workflow.WithCancel(ctx)) and that will fail the workflow with
// error CanceledError.
// ExecuteChildWorkflow returns ChildWorkflowFuture.
func ExecuteChildWorkflow(ctx Context, childWorkflow interface{}, args ...interface{}) ChildWorkflowFuture {
	i := getWorkflowInterceptor(ctx)
	env := getWorkflowEnvironment(ctx)
	workflowType := getWorkflowFunctionName(env.GetRegistry(), childWorkflow)
	return i.ExecuteChildWorkflow(ctx, workflowType, args...)
}

func (wc *workflowEnvironmentInterceptor) ExecuteChildWorkflow(ctx Context, childWorkflowType string, args ...interface{}) ChildWorkflowFuture {
	mainFuture, mainSettable := newDecodeFuture(ctx, childWorkflowType)
	executionFuture, executionSettable := NewFuture(ctx)
	result := &childWorkflowFutureImpl{
		decodeFutureImpl: mainFuture.(*decodeFutureImpl),
		executionFuture:  executionFuture.(*futureImpl),
	}
	// clients prior to v0.18.4 would incorrectly start child workflows that were started with cancelled contexts,
	// and did not react to cancellation between requested and started.
	correctChildCancellation := true
	workflowOptionsFromCtx := getWorkflowEnvOptions(ctx)

	// Starting with a canceled context should immediately fail, no need to even try.
	if ctx.Err() != nil {
		if workflowOptionsFromCtx.bugports.StartChildWorkflowsOnCanceledContext {
			// backport the bug
			correctChildCancellation = false
		} else {
			mainSettable.SetError(ctx.Err())
			executionSettable.SetError(ctx.Err())
			return result
		}
	}

	dc := workflowOptionsFromCtx.dataConverter
	env := getWorkflowEnvironment(ctx)
	wfType, input, err := getValidatedWorkflowFunction(childWorkflowType, args, dc, env.GetRegistry())
	if err != nil {
		executionSettable.Set(nil, err)
		mainSettable.Set(nil, err)
		return result
	}
	options, err := getValidatedWorkflowOptions(ctx)
	if err != nil {
		executionSettable.Set(nil, err)
		mainSettable.Set(nil, err)
		return result
	}
	options.dataConverter = dc
	options.contextPropagators = workflowOptionsFromCtx.contextPropagators
	options.memo = workflowOptionsFromCtx.memo
	options.searchAttributes = workflowOptionsFromCtx.searchAttributes

	params := executeWorkflowParams{
		workflowOptions: *options,
		input:           input,
		workflowType:    wfType,
		header:          getWorkflowHeader(ctx, options.contextPropagators),
		scheduledTime:   Now(ctx), /* this is needed for test framework, and is not send to server */
	}

	var childWorkflowExecution *WorkflowExecution

	ctxDone, cancellable := ctx.Done().(*channelImpl)
	cancellationCallback := &receiveCallback{}
	shouldCancelAsync := false
	err = getWorkflowEnvironment(ctx).ExecuteChildWorkflow(params, func(r []byte, e error) {
		mainSettable.Set(r, e)
		if cancellable {
			// future is done, we don't need cancellation anymore
			ctxDone.removeReceiveCallback(cancellationCallback)
		}
	}, func(r WorkflowExecution, e error) {
		if e == nil {
			childWorkflowExecution = &r
		}
		executionSettable.Set(r, e)

		// forward the delayed cancellation if necessary
		if shouldCancelAsync && e == nil && !mainFuture.IsReady() {
			if workflowOptionsFromCtx.bugports.StartChildWorkflowsOnCanceledContext {
				// do nothing: buggy behavior did not forward the cancellation
			} else {
				getWorkflowEnvironment(ctx).RequestCancelChildWorkflow(*options.domain, childWorkflowExecution.ID)
			}
		}
	})

	if err != nil {
		executionSettable.Set(nil, err)
		mainSettable.Set(nil, err)
		return result
	}

	if cancellable {
		cancellationCallback.fn = func(v interface{}, more bool) bool {
			if ctx.Err() == ErrCanceled {
				if childWorkflowExecution != nil && !mainFuture.IsReady() {
					// child workflow started, and ctx cancelled.  forward cancel to the child.
					getWorkflowEnvironment(ctx).RequestCancelChildWorkflow(*options.domain, childWorkflowExecution.ID)
				} else if childWorkflowExecution == nil && correctChildCancellation {
					// decision to start the child has been made, but it has not yet started.

					// TODO: ideal, but not strictly necessary for correctness:
					// if it's in the same decision, revoke that cancel synchronously.

					// if the decision has already gone through: wait for it to be started, and then cancel it.
					shouldCancelAsync = true
				}
			}
			return false
		}
		_, ok, more := ctxDone.receiveAsyncImpl(cancellationCallback)
		if ok || !more {
			cancellationCallback.fn(nil, more)
		}
	}

	return result
}

func getWorkflowHeader(ctx Context, ctxProps []ContextPropagator) *s.Header {
	header := &s.Header{
		Fields: make(map[string][]byte),
	}
	writer := NewHeaderWriter(header)
	for _, ctxProp := range ctxProps {
		ctxProp.InjectFromWorkflow(ctx, writer)
	}
	return header
}

// WorkflowInfo information about currently executing workflow
type WorkflowInfo struct {
	WorkflowExecution WorkflowExecution
	// The original runID before resetting. Using it instead of current runID can make workflow decision determinstic after reset
	// revive:disable-next-line this is an public field that is already exposed
	OriginalRunId                       string
	WorkflowType                        WorkflowType
	TaskListName                        string
	ExecutionStartToCloseTimeoutSeconds int32
	TaskStartToCloseTimeoutSeconds      int32
	Domain                              string
	Attempt                             int32 // Attempt starts from 0 and increased by 1 for every retry if retry policy is specified.
	lastCompletionResult                []byte
	CronSchedule                        *string
	ContinuedExecutionRunID             *string
	ParentWorkflowDomain                *string
	ParentWorkflowExecution             *WorkflowExecution
	Memo                                *s.Memo             // Value can be decoded using data converter (DefaultDataConverter, or custom one if set).
	SearchAttributes                    *s.SearchAttributes // Value can be decoded using DefaultDataConverter.
	BinaryChecksum                      *string             // The identifier(generated by md5sum by default) of worker code that is making the current decision(can be used for auto-reset feature)
	DecisionStartedEventID              int64               // the eventID of DecisionStarted that is making the current decision(can be used for reset API)
	RetryPolicy                         *s.RetryPolicy
	TotalHistoryBytes                   int64
	HistoryBytesServer                  int64
	HistoryCount                        int64
}

// GetBinaryChecksum returns the binary checksum(identifier) of this worker
// It is the identifier(generated by md5sum by default) of worker code that is making the current decision(can be used for auto-reset feature)
// In replay mode, it's from DecisionTaskCompleted event. In non-replay mode, it's from the currently executing worker.
func (wInfo *WorkflowInfo) GetBinaryChecksum() string {
	if wInfo.BinaryChecksum == nil {
		return getBinaryChecksum()
	}
	return *wInfo.BinaryChecksum
}

// GetDecisionCompletedEventID returns the eventID of DecisionStartedEvent that is making the current decision(can be used for reset API: decisionFinishEventID = DecisionStartedEventID + 1)
func (wInfo *WorkflowInfo) GetDecisionStartedEventID() int64 {
	return wInfo.DecisionStartedEventID
}

// GetWorkflowInfo extracts info of a current workflow from a context.
func GetWorkflowInfo(ctx Context) *WorkflowInfo {
	i := getWorkflowInterceptor(ctx)
	return i.GetWorkflowInfo(ctx)
}

func (wc *workflowEnvironmentInterceptor) GetWorkflowInfo(ctx Context) *WorkflowInfo {
	return wc.env.WorkflowInfo()
}

// GetLogger returns a logger to be used in workflow's context
func GetLogger(ctx Context) *zap.Logger {
	i := getWorkflowInterceptor(ctx)
	return i.GetLogger(ctx)
}

func (wc *workflowEnvironmentInterceptor) GetLogger(ctx Context) *zap.Logger {
	return wc.env.GetLogger()
}

// GetMetricsScope returns a metrics scope to be used in workflow's context
func GetMetricsScope(ctx Context) tally.Scope {
	i := getWorkflowInterceptor(ctx)
	return i.GetMetricsScope(ctx)
}

func (wc *workflowEnvironmentInterceptor) GetMetricsScope(ctx Context) tally.Scope {
	return wc.env.GetMetricsScope()
}

// GetTotalEstimatedHistoryBytes returns the current history size of that workflow
func GetTotalEstimatedHistoryBytes(ctx Context) int64 {
	i := getWorkflowInterceptor(ctx)
	return i.GetWorkflowInfo(ctx).TotalHistoryBytes
}

// GetHistoryCount returns the current number of history events of that workflow
func GetHistoryCount(ctx Context) int64 {
	i := getWorkflowInterceptor(ctx)
	return i.GetWorkflowInfo(ctx).HistoryCount
}

// Now returns the current time in UTC. It corresponds to the time when the decision task is started or replayed.
// Workflow needs to use this method to get the wall clock time instead of the one from the golang library.
func Now(ctx Context) time.Time {
	i := getWorkflowInterceptor(ctx)
	return i.Now(ctx).UTC()
}

func (wc *workflowEnvironmentInterceptor) Now(ctx Context) time.Time {
	return wc.env.Now()
}

// NewTimer returns immediately and the future becomes ready after the specified duration d. The workflow needs to use
// this NewTimer() to get the timer instead of the Go lang library one(timer.NewTimer()). You can cancel the pending
// timer by cancel the Context (using context from workflow.WithCancel(ctx)) and that will cancel the timer. After timer
// is canceled, the returned Future become ready, and Future.Get() will return *CanceledError.
// The current timer resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func NewTimer(ctx Context, d time.Duration) Future {
	i := getWorkflowInterceptor(ctx)
	return i.NewTimer(ctx, d)
}

func (wc *workflowEnvironmentInterceptor) NewTimer(ctx Context, d time.Duration) Future {
	future, settable := NewFuture(ctx)
	if d <= 0 {
		settable.Set(true, nil)
		return future
	}

	ctxDone, cancellable := ctx.Done().(*channelImpl)
	cancellationCallback := &receiveCallback{}
	t := wc.env.NewTimer(d, func(r []byte, e error) {
		settable.Set(nil, e)
		if cancellable {
			// future is done, we don't need cancellation anymore
			ctxDone.removeReceiveCallback(cancellationCallback)
		}
	})

	if t != nil && cancellable {
		cancellationCallback.fn = func(v interface{}, more bool) bool {
			if !future.IsReady() {
				wc.env.RequestCancelTimer(t.timerID)
			}
			return false
		}
		_, ok, more := ctxDone.receiveAsyncImpl(cancellationCallback)
		if ok || !more {
			cancellationCallback.fn(nil, more)
		}
	}
	return future
}

// Sleep pauses the current workflow for at least the duration d. A negative or zero duration causes Sleep to return
// immediately. Workflow code needs to use this Sleep() to sleep instead of the Go lang library one(timer.Sleep()).
// You can cancel the pending sleep by cancel the Context (using context from workflow.WithCancel(ctx)).
// Sleep() returns nil if the duration d is passed, or it returns *CanceledError if the ctx is canceled. There are 2
// reasons the ctx could be canceled: 1) your workflow code cancel the ctx (with workflow.WithCancel(ctx));
// 2) your workflow itself is canceled by external request.
// The current timer resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func Sleep(ctx Context, d time.Duration) (err error) {
	i := getWorkflowInterceptor(ctx)
	return i.Sleep(ctx, d)
}

func (wc *workflowEnvironmentInterceptor) Sleep(ctx Context, d time.Duration) (err error) {
	t := NewTimer(ctx, d)
	err = t.Get(ctx, nil)
	return
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
	i := getWorkflowInterceptor(ctx)
	return i.RequestCancelExternalWorkflow(ctx, workflowID, runID)
}

func (wc *workflowEnvironmentInterceptor) RequestCancelExternalWorkflow(ctx Context, workflowID, runID string) Future {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	options := getWorkflowEnvOptions(ctx1)
	future, settable := NewFuture(ctx1)

	if options.domain == nil || *options.domain == "" {
		settable.Set(nil, errDomainNotSet)
		return future
	}

	if workflowID == "" {
		settable.Set(nil, errWorkflowIDNotSet)
		return future
	}

	resultCallback := func(result []byte, err error) {
		settable.Set(result, err)
	}

	wc.env.RequestCancelExternalWorkflow(
		*options.domain,
		workflowID,
		runID,
		resultCallback,
	)

	return future
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
	i := getWorkflowInterceptor(ctx)
	return i.SignalExternalWorkflow(ctx, workflowID, runID, signalName, arg)
}

func (wc *workflowEnvironmentInterceptor) SignalExternalWorkflow(ctx Context, workflowID, runID, signalName string, arg interface{}) Future {
	const childWorkflowOnly = false // this means we are not limited to child workflow
	return signalExternalWorkflow(ctx, workflowID, runID, signalName, arg, childWorkflowOnly)
}

func signalExternalWorkflow(ctx Context, workflowID, runID, signalName string, arg interface{}, childWorkflowOnly bool) Future {
	env := getWorkflowEnvironment(ctx)
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	options := getWorkflowEnvOptions(ctx1)
	future, settable := NewFuture(ctx1)

	if options.domain == nil || *options.domain == "" {
		settable.Set(nil, errDomainNotSet)
		return future
	}

	if workflowID == "" {
		settable.Set(nil, errWorkflowIDNotSet)
		return future
	}

	input, err := encodeArg(options.dataConverter, arg)
	if err != nil {
		settable.Set(nil, err)
		return future
	}

	resultCallback := func(result []byte, err error) {
		settable.Set(result, err)
	}
	env.SignalExternalWorkflow(
		*options.domain,
		workflowID,
		runID,
		signalName,
		input,
		arg,
		childWorkflowOnly,
		resultCallback,
	)

	return future
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
	i := getWorkflowInterceptor(ctx)
	return i.UpsertSearchAttributes(ctx, attributes)
}

func (wc *workflowEnvironmentInterceptor) UpsertSearchAttributes(ctx Context, attributes map[string]interface{}) error {
	if _, ok := attributes[CadenceChangeVersion]; ok {
		return errors.New("CadenceChangeVersion is a reserved key that cannot be set, please use other key")
	}
	return wc.env.UpsertSearchAttributes(attributes)
}

// WithChildWorkflowOptions adds all workflow options to the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithChildWorkflowOptions(ctx Context, cwo ChildWorkflowOptions) Context {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	wfOptions := getWorkflowEnvOptions(ctx1)
	wfOptions.domain = common.StringPtr(cwo.Domain)
	wfOptions.taskListName = common.StringPtr(cwo.TaskList)
	wfOptions.workflowID = cwo.WorkflowID
	wfOptions.executionStartToCloseTimeoutSeconds = common.Int32Ptr(common.Int32Ceil(cwo.ExecutionStartToCloseTimeout.Seconds()))
	wfOptions.taskStartToCloseTimeoutSeconds = common.Int32Ptr(common.Int32Ceil(cwo.TaskStartToCloseTimeout.Seconds()))
	wfOptions.waitForCancellation = cwo.WaitForCancellation
	wfOptions.workflowIDReusePolicy = cwo.WorkflowIDReusePolicy
	wfOptions.retryPolicy = convertRetryPolicy(cwo.RetryPolicy)
	wfOptions.cronSchedule = cwo.CronSchedule
	wfOptions.memo = cwo.Memo
	wfOptions.searchAttributes = cwo.SearchAttributes
	wfOptions.parentClosePolicy = cwo.ParentClosePolicy
	wfOptions.bugports = cwo.Bugports

	return ctx1
}

// WithWorkflowDomain adds a domain to the context.
func WithWorkflowDomain(ctx Context, name string) Context {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	getWorkflowEnvOptions(ctx1).domain = common.StringPtr(name)
	return ctx1
}

// WithWorkflowTaskList adds a task list to the context.
func WithWorkflowTaskList(ctx Context, name string) Context {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	getWorkflowEnvOptions(ctx1).taskListName = common.StringPtr(name)
	return ctx1
}

// GetWorkflowTaskList retrieves current workflow tasklist from context
func GetWorkflowTaskList(ctx Context) *string {
	wo := getWorkflowEnvOptions(ctx)
	if wo == nil || wo.taskListName == nil {
		return nil
	}
	tl := *wo.taskListName // copy
	return &tl
}

// WithWorkflowID adds a workflowID to the context.
func WithWorkflowID(ctx Context, workflowID string) Context {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	getWorkflowEnvOptions(ctx1).workflowID = workflowID
	return ctx1
}

// WithExecutionStartToCloseTimeout adds a workflow execution timeout to the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithExecutionStartToCloseTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	getWorkflowEnvOptions(ctx1).executionStartToCloseTimeoutSeconds = common.Int32Ptr(common.Int32Ceil(d.Seconds()))
	return ctx1
}

// WithWorkflowTaskStartToCloseTimeout adds a decision timeout to the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithWorkflowTaskStartToCloseTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	getWorkflowEnvOptions(ctx1).taskStartToCloseTimeoutSeconds = common.Int32Ptr(common.Int32Ceil(d.Seconds()))
	return ctx1
}

// WithDataConverter adds DataConverter to the context.
func WithDataConverter(ctx Context, dc DataConverter) Context {
	if dc == nil {
		panic("data converter is nil for WithDataConverter")
	}
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	getWorkflowEnvOptions(ctx1).dataConverter = dc
	return ctx1
}

// withContextPropagators adds ContextPropagators to the context.
func withContextPropagators(ctx Context, contextPropagators []ContextPropagator) Context {
	ctx1 := setWorkflowEnvOptionsIfNotExist(ctx)
	getWorkflowEnvOptions(ctx1).contextPropagators = contextPropagators
	return ctx1
}

// GetSignalChannel returns channel corresponding to the signal name.
func GetSignalChannel(ctx Context, signalName string) Channel {
	i := getWorkflowInterceptor(ctx)
	return i.GetSignalChannel(ctx, signalName)
}

func (wc *workflowEnvironmentInterceptor) GetSignalChannel(ctx Context, signalName string) Channel {
	return getWorkflowEnvOptions(ctx).getSignalChannel(ctx, signalName)
}

func newEncodedValue(value []byte, dc DataConverter) Value {
	if dc == nil {
		dc = getDefaultDataConverter()
	}
	return &EncodedValue{value, dc}
}

// Get extract data from encoded data to desired value type. valuePtr is pointer to the actual value type.
func (b EncodedValue) Get(valuePtr interface{}) error {
	if !b.HasValue() {
		return ErrNoData
	}
	return decodeArg(b.dataConverter, b.value, valuePtr)
}

// HasValue return whether there is value
func (b EncodedValue) HasValue() bool {
	return b.value != nil
}

// SideEffect docs are in the public API to prevent duplication: [go.uber.org/cadence/workflow.SideEffect]
func SideEffect(ctx Context, f func(ctx Context) interface{}) Value {
	i := getWorkflowInterceptor(ctx)
	return i.SideEffect(ctx, f)
}

func (wc *workflowEnvironmentInterceptor) SideEffect(ctx Context, f func(ctx Context) interface{}) Value {
	dc := getDataConverterFromWorkflowContext(ctx)
	future, settable := NewFuture(ctx)
	wrapperFunc := func() ([]byte, error) {
		r := f(ctx)
		return encodeArg(dc, r)
	}
	resultCallback := func(result []byte, err error) {
		settable.Set(EncodedValue{result, dc}, err)
	}
	wc.env.SideEffect(wrapperFunc, resultCallback)
	var encoded EncodedValue
	if err := future.Get(ctx, &encoded); err != nil {
		panic(err)
	}
	return encoded
}

// MutableSideEffect docs are in the public API to prevent duplication: [go.uber.org/cadence/workflow.MutableSideEffect]
func MutableSideEffect(ctx Context, id string, f func(ctx Context) interface{}, equals func(a, b interface{}) bool) Value {
	i := getWorkflowInterceptor(ctx)
	return i.MutableSideEffect(ctx, id, f, equals)
}

func (wc *workflowEnvironmentInterceptor) MutableSideEffect(ctx Context, id string, f func(ctx Context) interface{}, equals func(a, b interface{}) bool) Value {
	wrapperFunc := func() interface{} {
		return f(ctx)
	}
	return wc.env.MutableSideEffect(id, wrapperFunc, equals)
}

// DefaultVersion is a version returned by GetVersion for code that wasn't versioned before
const DefaultVersion Version = -1

// CadenceChangeVersion is used as search attributes key to find workflows with specific change version.
const CadenceChangeVersion = "CadenceChangeVersion"

// GetVersionOption is an interface for functional options that can be applied to GetVersion
type GetVersionOption interface {
	apply(*getVersionConfig)
}

// getVersionConfig holds the configuration for GetVersion
type getVersionConfig struct {
	// CustomVersion is used to force GetVersion to return a specific version
	// instead of maxSupported version. Set up via ExecuteWithVersion option.
	CustomVersion *Version

	// UseMinVersion is used to force GetVersion to return minSupported version
	// instead of maxSupported version. Set up via ExecuteWithMinVersion option.
	UseMinVersion bool
}

// customVersionOption forces GetVersion to return a specific version
type customVersionOption Version

func (c customVersionOption) apply(config *getVersionConfig) {
	version := Version(c)
	config.CustomVersion = &version
}

// ExecuteWithVersion forces a specific version to be returned when GetVersion is executed for the first time,
// instead of returning maxSupported version. Check public documentation of workflow.ExecuteWithVersion
func ExecuteWithVersion(version Version) GetVersionOption {
	return customVersionOption(version)
}

// useMinVersionOption forces GetVersion to return minSupported version
type useMinVersionOption struct{}

func (useMinVersionOption) apply(config *getVersionConfig) {
	config.UseMinVersion = true
}

// ExecuteWithMinVersion forces minSupported version to be returned when GetVersion is executed for the first time,
// instead of returning maxSupported version. Check public documentation of workflow.ExecuteWithMinVersion
func ExecuteWithMinVersion() GetVersionOption {
	return useMinVersionOption{}
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
// and not proceed; 2) if you ever need to make more changes for "fooChange", for example change activity from baz to qux,
// you just need to update the maxVersion from 2 to 3.
//
// Note that, you only need to preserve the first call to GetVersion() for each changeID. All subsequent call to GetVersion()
// with same changeID are safe to remove. However, if you really want to get rid of the first GetVersion() call as well,
// you can do so, but you need to make sure: 1) all older version executions are completed; 2) you can no longer use "fooChange"
// as changeID. If you ever need to make changes to that same part like change from baz to qux, you would need to use a
// different changeID like "fooChange-fix2", and start minVersion from DefaultVersion again. The code would looks like:
//
//	v := workflow.GetVersion(ctx, "fooChange-fix2", workflow.DefaultVersion, 1)
//	if v == workflow.DefaultVersion {
//	  err = workflow.ExecuteActivity(ctx, baz, data).Get(ctx, nil)
//	} else {
//	  err = workflow.ExecuteActivity(ctx, qux, data).Get(ctx, nil)
//	}
func GetVersion(ctx Context, changeID string, minSupported, maxSupported Version, opts ...GetVersionOption) Version {
	i := getWorkflowInterceptor(ctx)
	return i.GetVersion(ctx, changeID, minSupported, maxSupported, opts...)
}

func (wc *workflowEnvironmentInterceptor) GetVersion(ctx Context, changeID string, minSupported, maxSupported Version, opts ...GetVersionOption) Version {
	return wc.env.GetVersion(changeID, minSupported, maxSupported, opts...)
}

// SetQueryHandler sets the query handler to handle workflow query. The queryType specify which query type this handler
// should handle. The handler must be a function that returns 2 values. The first return value must be a serializable
// result. The second return value must be an error. The handler function could receive any number of input parameters.
// All the input parameter must be serializable. You should call workflow.SetQueryHandler() at the beginning of the workflow
// code. When client calls Client.QueryWorkflow() to cadence server, a task will be generated on server that will be dispatched
// to a workflow worker, which will replay the history events and then execute a query handler based on the query type.
// The query handler will be invoked out of the context of the workflow, meaning that the handler code must not use cadence
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
	i := getWorkflowInterceptor(ctx)
	return i.SetQueryHandler(ctx, queryType, handler)
}

func (wc *workflowEnvironmentInterceptor) SetQueryHandler(ctx Context, queryType string, handler interface{}) error {
	if strings.HasPrefix(queryType, "__") {
		return errors.New("queryType starts with '__' is reserved for internal use")
	}
	return setQueryHandler(ctx, queryType, handler)
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
	i := getWorkflowInterceptor(ctx)
	return i.IsReplaying(ctx)
}

func (wc *workflowEnvironmentInterceptor) IsReplaying(ctx Context) bool {
	return wc.env.IsReplaying()
}

// HasLastCompletionResult checks if there is completion result from previous runs.
// This is used in combination with cron schedule. A workflow can be started with an optional cron schedule.
// If a cron workflow wants to pass some data to next schedule, it can return any data and that data will become
// available when next run starts.
// This HasLastCompletionResult() checks if there is such data available passing down from previous successful run.
func HasLastCompletionResult(ctx Context) bool {
	i := getWorkflowInterceptor(ctx)
	return i.HasLastCompletionResult(ctx)
}

func (wc *workflowEnvironmentInterceptor) HasLastCompletionResult(ctx Context) bool {
	info := wc.GetWorkflowInfo(ctx)
	return len(info.lastCompletionResult) > 0
}

// GetLastCompletionResult extract last completion result from previous run for this cron workflow.
// This is used in combination with cron schedule. A workflow can be started with an optional cron schedule.
// If a cron workflow wants to pass some data to next schedule, it can return any data and that data will become
// available when next run starts.
// This GetLastCompletionResult() extract the data into expected data structure.
func GetLastCompletionResult(ctx Context, d ...interface{}) error {
	i := getWorkflowInterceptor(ctx)
	return i.GetLastCompletionResult(ctx, d...)
}

func (wc *workflowEnvironmentInterceptor) GetLastCompletionResult(ctx Context, d ...interface{}) error {
	info := wc.GetWorkflowInfo(ctx)
	if len(info.lastCompletionResult) == 0 {
		return ErrNoData
	}

	encodedVal := newEncodedValues(info.lastCompletionResult, getDataConverterFromWorkflowContext(ctx))
	return encodedVal.Get(d...)
}

// WithActivityOptions adds all options to the copy of the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithActivityOptions(ctx Context, options ActivityOptions) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	eap := getActivityOptions(ctx1)

	eap.TaskListName = options.TaskList
	eap.ScheduleToCloseTimeoutSeconds = common.Int32Ceil(options.ScheduleToCloseTimeout.Seconds())
	eap.StartToCloseTimeoutSeconds = common.Int32Ceil(options.StartToCloseTimeout.Seconds())
	eap.ScheduleToStartTimeoutSeconds = common.Int32Ceil(options.ScheduleToStartTimeout.Seconds())
	eap.HeartbeatTimeoutSeconds = common.Int32Ceil(options.HeartbeatTimeout.Seconds())
	eap.WaitForCancellation = options.WaitForCancellation
	eap.ActivityID = common.StringPtr(options.ActivityID)
	eap.RetryPolicy = convertRetryPolicy(options.RetryPolicy)
	return ctx1
}

// WithLocalActivityOptions adds local activity options to the copy of the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithLocalActivityOptions(ctx Context, options LocalActivityOptions) Context {
	ctx1 := setLocalActivityParametersIfNotExist(ctx)
	opts := getLocalActivityOptions(ctx1)

	opts.ScheduleToCloseTimeoutSeconds = common.Int32Ceil(options.ScheduleToCloseTimeout.Seconds())
	opts.RetryPolicy = options.RetryPolicy
	return ctx1
}

// WithTaskList adds a task list to the copy of the context.
// Note this shall not confuse with WithWorkflowTaskList. This is the tasklist for activities
func WithTaskList(ctx Context, name string) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).TaskListName = name
	return ctx1
}

// GetActivityTaskList retrieves tasklist info from context
func GetActivityTaskList(ctx Context) *string {
	ao := getActivityOptions(ctx)
	if ao == nil {
		return nil
	}
	tl := ao.TaskListName // copy
	return &tl
}

// WithScheduleToCloseTimeout adds a timeout to the copy of the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithScheduleToCloseTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).ScheduleToCloseTimeoutSeconds = common.Int32Ceil(d.Seconds())
	return ctx1
}

// WithScheduleToStartTimeout adds a timeout to the copy of the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithScheduleToStartTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).ScheduleToStartTimeoutSeconds = common.Int32Ceil(d.Seconds())
	return ctx1
}

// WithStartToCloseTimeout adds a timeout to the copy of the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithStartToCloseTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).StartToCloseTimeoutSeconds = common.Int32Ceil(d.Seconds())
	return ctx1
}

// WithHeartbeatTimeout adds a timeout to the copy of the context.
// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
// subjected to change in the future.
func WithHeartbeatTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).HeartbeatTimeoutSeconds = common.Int32Ceil(d.Seconds())
	return ctx1
}

// WithWaitForCancellation adds wait for the cacellation to the copy of the context.
func WithWaitForCancellation(ctx Context, wait bool) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).WaitForCancellation = wait
	return ctx1
}

// WithRetryPolicy adds retry policy to the copy of the context
func WithRetryPolicy(ctx Context, retryPolicy RetryPolicy) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).RetryPolicy = convertRetryPolicy(&retryPolicy)
	return ctx1
}

func convertRetryPolicy(retryPolicy *RetryPolicy) *s.RetryPolicy {
	if retryPolicy == nil {
		return nil
	}
	thriftRetryPolicy := s.RetryPolicy{
		InitialIntervalInSeconds:    common.Int32Ptr(common.Int32Ceil(retryPolicy.InitialInterval.Seconds())),
		MaximumIntervalInSeconds:    common.Int32Ptr(common.Int32Ceil(retryPolicy.MaximumInterval.Seconds())),
		BackoffCoefficient:          &retryPolicy.BackoffCoefficient,
		MaximumAttempts:             &retryPolicy.MaximumAttempts,
		NonRetriableErrorReasons:    retryPolicy.NonRetriableErrorReasons,
		ExpirationIntervalInSeconds: common.Int32Ptr(common.Int32Ceil(retryPolicy.ExpirationInterval.Seconds())),
	}
	if *thriftRetryPolicy.BackoffCoefficient == 0 {
		thriftRetryPolicy.BackoffCoefficient = common.Float64Ptr(backoff.DefaultBackoffCoefficient)
	}
	return &thriftRetryPolicy
}
