package cadence

import (
	"errors"
	"fmt"

	"github.com/uber-go/cadence-client/.gen/go/shared"
	"reflect"
)

type (
	// Marker functions are used to ensure that interfaces never implement each other.
	// For example without marker an implementation of ErrorWithDetails matches
	// CanceledError interface as well.

	// ErrorWithDetails to return from Workflow and activity implementations.
	ErrorWithDetails interface {
		error
		Reason() string
		Details(d ...interface{}) // Extracts details into passed pointers
		errorWithDetails()        // interface marker
	}

	// TimeoutError returned when activity or child workflow timed out
	TimeoutError interface {
		error
		TimeoutType() shared.TimeoutType
		Details(d ...interface{}) // Present only for HEARTBEAT TimeoutType
		timeoutError()            // interface marker
	}

	// CanceledError returned when operation was canceled
	CanceledError interface {
		error
		Details(d ...interface{}) // Extracts details into passed pointers
		canceledError()           // interface marker
	}

	// PanicError contains information about panicked workflow
	PanicError interface {
		error
		Value(v interface{}) // Value passed to panic call
		StackTrace() string  // Stack trace of a panicked coroutine
		panicError()         // interface marker
	}
)

var _ ErrorWithDetails = (*errorWithDetails)(nil)
var _ CanceledError = (*canceledError)(nil)
var _ TimeoutError = (*timeoutError)(nil)
var _ PanicError = (*panicError)(nil)

// ErrActivityResultPending is returned from activity's Execute method to indicate the activity is not completed when
// Execute method returns. activity will be completed asynchronously when Client.CompleteActivity() is called.
var ErrActivityResultPending = errors.New("not error: do not autocomplete, " +
	"using Client.CompleteActivity() to complete")

// NewErrorWithDetails creates ErrorWithDetails instance
// Create standard error through errors.New or fmt.Errorf() if no details are provided
func NewErrorWithDetails(reason string, details ...interface{}) ErrorWithDetails {
	data, err := getHostEnvironment().encodeArgs(details)
	if err != nil {
		panic(err)
	}
	return &errorWithDetails{reason: reason, details: data}
}

// NewTimeoutError creates TimeoutError instance.
// Use NewHeartbeatTimeoutError to create heartbeat TimeoutError
// WARNING: This function is public only to support unit testing of workflows.
// It shouldn't be used by application level code.
func NewTimeoutError(timeoutType shared.TimeoutType) TimeoutError {
	return &timeoutError{timeoutType: timeoutType}
}

// NewHeartbeatTimeoutError creates TimeoutError instance
// WARNING: This function is public only to support unit testing of workflows.
// It shouldn't be used by application level code.
func NewHeartbeatTimeoutError(details ...interface{}) TimeoutError {
	data, err := getHostEnvironment().encodeArgs(details)
	if err != nil {
		panic(err)
	}
	return &timeoutError{timeoutType: shared.TimeoutType_HEARTBEAT, details: data}
}

// NewCanceledError creates CanceledError instance
func NewCanceledError(details ...interface{}) CanceledError {
	data, err := getHostEnvironment().encodeArgs(details)
	if err != nil {
		panic(err)
	}
	return &canceledError{details: data}
}

// errorWithDetails implements ErrorWithDetails
type errorWithDetails struct {
	reason  string
	details []byte
}

// Error from error interface
func (e *errorWithDetails) Error() string {
	return e.reason
}

// Reason is from ErrorWithDetails interface
func (e *errorWithDetails) Reason() string {
	return e.reason
}

// Details is from ErrorWithDetails interface
func (e *errorWithDetails) Details(d ...interface{}) {
	if err := getHostEnvironment().decodeArgsTo(e.details, d); err != nil {
		panic(err)
	}
}

// errorWithDetails is from ErrorWithDetails interface
func (e *errorWithDetails) errorWithDetails() {}

// timeoutError implements TimeoutError
type timeoutError struct {
	timeoutType shared.TimeoutType
	details     []byte
}

// Error from error interface
func (e *timeoutError) Error() string {
	return fmt.Sprintf("TimeoutType: %v", e.timeoutType)
}

func (e *timeoutError) TimeoutType() shared.TimeoutType {
	return e.timeoutType
}

// Details is from TimeoutError interface
func (e *timeoutError) Details(d ...interface{}) {
	if err := getHostEnvironment().decodeArgsTo(e.details, d); err != nil {
		panic(err)
	}

}

func (e *timeoutError) timeoutError() {}

type canceledError struct {
	details []byte
}

// Error from error interface
func (e *canceledError) Error() string {
	return "CanceledError"
}

// Details is from CanceledError interface
func (e *canceledError) Details(d ...interface{}) {
	if err := getHostEnvironment().decodeArgsTo(e.details, d); err != nil {
		panic(err)
	}
}

func (e *canceledError) canceledError() {}

type panicError struct {
	value      interface{}
	stackTrace string
}

func newPanicError(value interface{}, stackTrace string) PanicError {
	return &panicError{value: value, stackTrace: stackTrace}
}

func (e *panicError) Error() string {
	return fmt.Sprintf("%v", e.value)
}

func (e *panicError) Value(v interface{}) {
	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(e.value))
}

func (e *panicError) StackTrace() string {
	return e.stackTrace
}

func (e *panicError) panicError() {}
