// Copyright (c) 2017 Uber Technologies, Inc.
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
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/pborman/uuid"
	"github.com/uber-go/tally"
	"go.uber.org/yarpc"

	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/metrics"
)

const (
	// libraryVersionHeaderName refers to the name of the
	// tchannel / http header that contains the client
	// library version
	libraryVersionHeaderName = "cadence-client-library-version"

	// featureVersionHeaderName refers to the name of the
	// tchannel / http header that contains the client
	// feature version
	featureVersionHeaderName = "cadence-client-feature-version"

	// clientImplHeaderName refers to the name of the
	// header that contains the client implementation
	clientImplHeaderName  = "cadence-client-name"
	clientImplHeaderValue = "uber-go"

	clientFeatureFlagsHeaderName = "cadence-client-feature-flags"

	// defaultRPCTimeout is the default tchannel rpc call timeout
	defaultRPCTimeout = 10 * time.Second
	//minRPCTimeout is minimum rpc call timeout allowed
	minRPCTimeout = 1 * time.Second
	//maxRPCTimeout is maximum rpc call timeout allowed
	maxRPCTimeout = 5 * time.Second
	// maxQueryRPCTimeout is the maximum rpc call timeout allowed for query
	maxQueryRPCTimeout = 20 * time.Second
)

type (
	FeatureFlags struct {
		WorkflowExecutionAlreadyCompletedErrorEnabled bool
		PollerAutoScalerEnabled                       bool
	}
)

var (
	// call header to cadence server
	_yarpcCallOptions = []yarpc.CallOption{
		yarpc.WithHeader(libraryVersionHeaderName, LibraryVersion),
		yarpc.WithHeader(featureVersionHeaderName, FeatureVersion),
		yarpc.WithHeader(clientImplHeaderName, clientImplHeaderValue),
	}
)

func fromInternalFeatureFlags(featureFlags FeatureFlags) s.FeatureFlags {
	// if we are using client-side-only flags in client.FeatureFlags;
	// don't include them in shared.FeatureFlags and drop them here
	return s.FeatureFlags{
		WorkflowExecutionAlreadyCompletedErrorEnabled: common.BoolPtr(featureFlags.WorkflowExecutionAlreadyCompletedErrorEnabled),
	}
}

func toInternalFeatureFlags(featureFlags *s.FeatureFlags) FeatureFlags {
	flags := FeatureFlags{}
	if featureFlags != nil {
		if featureFlags.WorkflowExecutionAlreadyCompletedErrorEnabled != nil {
			flags.WorkflowExecutionAlreadyCompletedErrorEnabled = *featureFlags.WorkflowExecutionAlreadyCompletedErrorEnabled
		}
	}
	return flags
}

func featureFlagsHeader(featureFlags FeatureFlags) string {
	serialized := ""
	buf, err := json.Marshal(fromInternalFeatureFlags(featureFlags))
	if err == nil {
		serialized = string(buf)
	}
	return serialized
}

func getYarpcCallOptions(featureFlags FeatureFlags) []yarpc.CallOption {
	return append(
		_yarpcCallOptions,
		yarpc.WithHeader(clientFeatureFlagsHeaderName, featureFlagsHeader(featureFlags)),
	)
}

// ContextBuilder stores all Channel-specific parameters that will
// be stored inside of a context.
type contextBuilder struct {
	// If Timeout is zero, Build will default to defaultTimeout.
	Timeout time.Duration

	// ParentContext to build the new context from. If empty, context.Background() is used.
	// The new (child) context inherits a number of properties from the parent context:
	//   - context fields, accessible via `ctx.Value(key)`
	ParentContext context.Context
}

func (cb *contextBuilder) Build() (context.Context, context.CancelFunc) {
	parent := cb.ParentContext
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, cb.Timeout)
}

// sets the rpc timeout for a context
func chanTimeout(timeout time.Duration) func(builder *contextBuilder) {
	return func(b *contextBuilder) {
		b.Timeout = timeout
	}
}

// newChannelContext - Get a rpc channel context for query
func newChannelContextForQuery(
	ctx context.Context,
	featureFlags FeatureFlags,
	options ...func(builder *contextBuilder),
) (context.Context, context.CancelFunc, []yarpc.CallOption) {
	return newChannelContextHelper(ctx, true, featureFlags, options...)
}

// newChannelContext - Get a rpc channel context
func newChannelContext(
	ctx context.Context,
	featureFlags FeatureFlags,
	options ...func(builder *contextBuilder),
) (context.Context, context.CancelFunc, []yarpc.CallOption) {
	return newChannelContextHelper(ctx, false, featureFlags, options...)
}

func newChannelContextHelper(
	ctx context.Context,
	isQuery bool,
	featureFlags FeatureFlags,
	options ...func(builder *contextBuilder),
) (context.Context, context.CancelFunc, []yarpc.CallOption) {
	rpcTimeout := defaultRPCTimeout
	if ctx != nil {
		// Set rpc timeout less than context timeout to allow for retries when call gets lost
		now := time.Now()
		if expiration, ok := ctx.Deadline(); ok && expiration.After(now) {
			rpcTimeout = expiration.Sub(now) / 2
			// Make sure to not set rpc timeout lower than minRPCTimeout
			if rpcTimeout < minRPCTimeout {
				rpcTimeout = minRPCTimeout
			} else if rpcTimeout > maxRPCTimeout && !isQuery {
				rpcTimeout = maxRPCTimeout
			} else if rpcTimeout > maxQueryRPCTimeout && isQuery {
				rpcTimeout = maxQueryRPCTimeout
			}
		}
	}
	builder := &contextBuilder{Timeout: rpcTimeout}
	if ctx != nil {
		builder.ParentContext = ctx
	}
	for _, opt := range options {
		opt(builder)
	}
	ctx, cancelFn := builder.Build()

	return ctx, cancelFn, getYarpcCallOptions(featureFlags)
}

// GetWorkerIdentity gets a default identity for the worker.
//
// This contains a random UUID, generated each time it is called, to prevent identity collisions when workers share
// other host/pid/etc information.  These alone are not guaranteed to be unique, especially when Docker is involved.
// Take care to retrieve this only once per worker.
func getWorkerIdentity(tasklistName string) string {
	return fmt.Sprintf("%d@%s@%s@%s", os.Getpid(), getHostName(), tasklistName, uuid.New())
}

func getHostName() string {
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "UnKnown"
	}
	return hostName
}

func getWorkerTaskList(stickyUUID string) string {
	// includes hostname for debuggability, stickyUUID guarantees the uniqueness
	return fmt.Sprintf("%s:%s", getHostName(), stickyUUID)
}

// ActivityTypePtr makes a copy and returns the pointer to a ActivityType.
func activityTypePtr(v ActivityType) *s.ActivityType {
	return &s.ActivityType{Name: common.StringPtr(v.Name)}
}

func flowWorkflowTypeFrom(v s.WorkflowType) WorkflowType {
	return WorkflowType{Name: v.GetName()}
}

// WorkflowTypePtr makes a copy and returns the pointer to a WorkflowType.
func workflowTypePtr(t WorkflowType) *s.WorkflowType {
	return &s.WorkflowType{Name: common.StringPtr(t.Name)}
}

// getErrorDetails gets reason and details.
func getErrorDetails(err error, dataConverter DataConverter) (string, []byte) {
	switch err := err.(type) {
	case *CustomError:
		var data []byte
		var err0 error
		switch details := err.details.(type) {
		case ErrorDetailsValues:
			data, err0 = encodeArgs(dataConverter, details)
		case *EncodedValues:
			data = details.values
		default:
			panic("unknown error type")
		}
		if err0 != nil {
			panic(err0)
		}
		return err.Reason(), data
	case *CanceledError:
		var data []byte
		var err0 error
		switch details := err.details.(type) {
		case ErrorDetailsValues:
			data, err0 = encodeArgs(dataConverter, details)
		case *EncodedValues:
			data = details.values
		default:
			panic("unknown error type")
		}
		if err0 != nil {
			panic(err0)
		}
		return errReasonCanceled, data
	case *PanicError:
		data, err0 := encodeArgs(dataConverter, []interface{}{err.Error(), err.StackTrace()})
		if err0 != nil {
			panic(err0)
		}
		return errReasonPanic, data
	case *TimeoutError:
		var data []byte
		var err0 error
		switch details := err.details.(type) {
		case ErrorDetailsValues:
			data, err0 = encodeArgs(dataConverter, details)
		case *EncodedValues:
			data = details.values
		default:
			panic("unknown error type")
		}
		if err0 != nil {
			panic(err0)
		}
		return fmt.Sprintf("%v %v", errReasonTimeout, err.timeoutType), data
	default:
		// will be convert to GenericError when receiving from server.
		return errReasonGeneric, []byte(err.Error())
	}
}

// constructError construct error from reason and details sending down from server.
func constructError(reason string, details []byte, dataConverter DataConverter) error {
	if strings.HasPrefix(reason, errReasonTimeout) {
		details := newEncodedValues(details, dataConverter)
		timeoutType, err := getTimeoutTypeFromErrReason(reason)
		if err != nil {
			// prior client version uses details to indicate timeoutType
			if err := details.Get(&timeoutType); err != nil {
				panic(err)
			}
			return NewTimeoutError(timeoutType)
		}
		return NewTimeoutError(timeoutType, details)
	}

	switch reason {
	case errReasonPanic:
		// panic error
		var msg, st string
		details := newEncodedValues(details, dataConverter)
		details.Get(&msg, &st)
		return newPanicError(msg, st)
	case errReasonGeneric:
		// errors created other than using NewCustomError() API.
		return &GenericError{err: string(details)}
	case errReasonCanceled:
		details := newEncodedValues(details, dataConverter)
		return NewCanceledError(details)
	default:
		details := newEncodedValues(details, dataConverter)
		err := NewCustomError(reason, details)
		return err
	}
}

func getKillSignal() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

// getMetricsScopeForActivity return properly tagged tally scope for activity
func getMetricsScopeForActivity(ts *metrics.TaggedScope, workflowType, activityType string) tally.Scope {
	return ts.GetTaggedScope(tagWorkflowType, workflowType, tagActivityType, activityType)
}

// getMetricsScopeForLocalActivity return properly tagged tally scope for local activity
func getMetricsScopeForLocalActivity(ts *metrics.TaggedScope, workflowType, localActivityType string) tally.Scope {
	return ts.GetTaggedScope(tagWorkflowType, workflowType, tagLocalActivityType, localActivityType)
}

func getTimeoutTypeFromErrReason(reason string) (s.TimeoutType, error) {
	timeoutTypeStr := reason[strings.Index(reason, " ")+1:]
	var timeoutType s.TimeoutType
	if err := timeoutType.UnmarshalText([]byte(timeoutTypeStr)); err != nil {
		// this happens when the timeout error reason is constructed by an prior constructed by prior client version
		return 0, err
	}
	return timeoutType, nil
}

func estimateHistorySize(logger *zap.Logger, event *s.HistoryEvent) int {
	sum := historySizeEstimationBuffer
	switch event.GetEventType() {
	case s.EventTypeWorkflowExecutionStarted:
		if event.WorkflowExecutionStartedEventAttributes != nil {
			sum += len(event.WorkflowExecutionStartedEventAttributes.Input)
			sum += len(event.WorkflowExecutionStartedEventAttributes.ContinuedFailureDetails)
			sum += len(event.WorkflowExecutionStartedEventAttributes.LastCompletionResult)
			sum += sizeOf(event.WorkflowExecutionStartedEventAttributes.Memo.GetFields())
			sum += sizeOf(event.WorkflowExecutionStartedEventAttributes.Header.GetFields())
			sum += sizeOf(event.WorkflowExecutionStartedEventAttributes.SearchAttributes.GetIndexedFields())
		}
	case s.EventTypeWorkflowExecutionCompleted:
		if event.WorkflowExecutionCompletedEventAttributes != nil {
			sum += len(event.WorkflowExecutionCompletedEventAttributes.Result)
		}
	case s.EventTypeWorkflowExecutionSignaled:
		if event.WorkflowExecutionSignaledEventAttributes != nil {
			sum += len(event.WorkflowExecutionSignaledEventAttributes.Input)
		}
	case s.EventTypeWorkflowExecutionFailed:
		if event.WorkflowExecutionFailedEventAttributes != nil {
			sum += len(event.WorkflowExecutionFailedEventAttributes.Details)
		}
	case s.EventTypeDecisionTaskStarted:
		if event.DecisionTaskStartedEventAttributes != nil {
			sum += getLengthOfStringPointer(event.DecisionTaskStartedEventAttributes.Identity)
		}
	case s.EventTypeDecisionTaskCompleted:
		if event.DecisionTaskCompletedEventAttributes != nil {
			sum += len(event.DecisionTaskCompletedEventAttributes.ExecutionContext)
			sum += getLengthOfStringPointer(event.DecisionTaskCompletedEventAttributes.Identity)
			sum += getLengthOfStringPointer(event.DecisionTaskCompletedEventAttributes.BinaryChecksum)
		}
	case s.EventTypeDecisionTaskFailed:
		if event.DecisionTaskFailedEventAttributes != nil {
			sum += len(event.DecisionTaskFailedEventAttributes.Details)
		}
	case s.EventTypeActivityTaskScheduled:
		if event.ActivityTaskScheduledEventAttributes != nil {
			sum += len(event.ActivityTaskScheduledEventAttributes.Input)
			sum += sizeOf(event.ActivityTaskScheduledEventAttributes.Header.GetFields())
		}
	case s.EventTypeActivityTaskStarted:
		if event.ActivityTaskStartedEventAttributes != nil {
			sum += len(event.ActivityTaskStartedEventAttributes.LastFailureDetails)
		}
	case s.EventTypeActivityTaskCompleted:
		if event.ActivityTaskCompletedEventAttributes != nil {
			sum += len(event.ActivityTaskCompletedEventAttributes.Result)
			sum += getLengthOfStringPointer(event.ActivityTaskCompletedEventAttributes.Identity)
		}
	case s.EventTypeActivityTaskFailed:
		if event.ActivityTaskFailedEventAttributes != nil {
			sum += len(event.ActivityTaskFailedEventAttributes.Details)
		}
	case s.EventTypeActivityTaskTimedOut:
		if event.ActivityTaskTimedOutEventAttributes != nil {
			sum += len(event.ActivityTaskTimedOutEventAttributes.Details)
			sum += len(event.ActivityTaskTimedOutEventAttributes.LastFailureDetails)
		}
	case s.EventTypeActivityTaskCanceled:
		if event.ActivityTaskCanceledEventAttributes != nil {
			sum += len(event.ActivityTaskCanceledEventAttributes.Details)
		}
	case s.EventTypeMarkerRecorded:
		if event.MarkerRecordedEventAttributes != nil {
			sum += len(event.MarkerRecordedEventAttributes.Details)
		}
	case s.EventTypeWorkflowExecutionTerminated:
		if event.WorkflowExecutionTerminatedEventAttributes != nil {
			sum += len(event.WorkflowExecutionTerminatedEventAttributes.Details)
		}
	case s.EventTypeWorkflowExecutionCanceled:
		if event.WorkflowExecutionCanceledEventAttributes != nil {
			sum += len(event.WorkflowExecutionCanceledEventAttributes.Details)
		}
	case s.EventTypeWorkflowExecutionContinuedAsNew:
		if event.WorkflowExecutionContinuedAsNewEventAttributes != nil {
			sum += len(event.WorkflowExecutionContinuedAsNewEventAttributes.Input)
			sum += len(event.WorkflowExecutionContinuedAsNewEventAttributes.FailureDetails)
			sum += len(event.WorkflowExecutionContinuedAsNewEventAttributes.LastCompletionResult)
			sum += sizeOf(event.WorkflowExecutionContinuedAsNewEventAttributes.Memo.GetFields())
			sum += sizeOf(event.WorkflowExecutionContinuedAsNewEventAttributes.Header.GetFields())
			sum += sizeOf(event.WorkflowExecutionContinuedAsNewEventAttributes.SearchAttributes.GetIndexedFields())
		}
	case s.EventTypeStartChildWorkflowExecutionInitiated:
		if event.StartChildWorkflowExecutionInitiatedEventAttributes != nil {
			sum += len(event.StartChildWorkflowExecutionInitiatedEventAttributes.Input)
			sum += len(event.StartChildWorkflowExecutionInitiatedEventAttributes.Control)
			sum += sizeOf(event.StartChildWorkflowExecutionInitiatedEventAttributes.Memo.GetFields())
			sum += sizeOf(event.StartChildWorkflowExecutionInitiatedEventAttributes.Header.GetFields())
			sum += sizeOf(event.StartChildWorkflowExecutionInitiatedEventAttributes.SearchAttributes.GetIndexedFields())
		}
	case s.EventTypeChildWorkflowExecutionCompleted:
		if event.ChildWorkflowExecutionCompletedEventAttributes != nil {
			sum += len(event.ChildWorkflowExecutionCompletedEventAttributes.Result)
		}
	case s.EventTypeChildWorkflowExecutionFailed:
		if event.ChildWorkflowExecutionFailedEventAttributes != nil {
			sum += len(event.ChildWorkflowExecutionFailedEventAttributes.Details)
			sum += getLengthOfStringPointer(event.ChildWorkflowExecutionFailedEventAttributes.Reason)
		}
	case s.EventTypeChildWorkflowExecutionCanceled:
		if event.ChildWorkflowExecutionCanceledEventAttributes != nil {
			sum += len(event.ChildWorkflowExecutionCanceledEventAttributes.Details)
		}
	case s.EventTypeSignalExternalWorkflowExecutionInitiated:
		if event.SignalExternalWorkflowExecutionInitiatedEventAttributes != nil {
			sum += len(event.SignalExternalWorkflowExecutionInitiatedEventAttributes.Control)
			sum += len(event.SignalExternalWorkflowExecutionInitiatedEventAttributes.Input)
		}
	default:
		logger.Debug("unsupported event type for history size estimation", zap.String("Event Type", event.GetEventType().String()))
	}

	return sum
}

// simple function to estimate the size of a map[string][]byte
func sizeOf(o map[string][]byte) int {
	sum := 0
	for k, v := range o {
		sum += len(k) + len(v)
	}
	return sum
}

// simple function to estimate the size of a string pointer
func getLengthOfStringPointer(s *string) int {
	if s == nil {
		return 0
	}
	return len(*s)
}
