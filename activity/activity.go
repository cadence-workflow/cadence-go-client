// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package activity

import (
	"context"

	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/internal"
)

type (
	// Type identifies an activity type.
	Type = internal.ActivityType

	// Info contains information about a currently executing activity.
	Info = internal.ActivityInfo

	// RegisterOptions consists of options for registering an activity
	RegisterOptions = internal.RegisterActivityOptions

	// RegistryInfo
	RegistryInfo = internal.RegistryActivityInfo
)

// ErrResultPending is returned from activity's implementation to indicate the activity is not completed when
// activity method returns.  The Activity needs to be completed by [go.uber.org/cadence/client.Client.CompleteActivity]
// separately.
//
// For example, if an activity requires human interaction (like approving an expense report), the activity could push
// its [Info.TaskToken] to an external system, and return [ErrResultPending].  Later on, a human can approve it, and the
// system can use the associated token with [go.uber.org/cadence/client.Client.CompleteActivity] to finish the
// still-"running" activity.
//
// Caution: since using this frequently implies "long" timeouts, but there is no actually-running activity function
// that can use [go.uber.org/cadence/client.Client.RecordActivityHeartbeat], activity-worker losses prior to recording
// the [Info.TaskToken] in an external system may not be noticed until the "long" timeout occurs.  This can be resolved
// by having another system call RecordActivityHeartbeat while that external action is running, but there is currently
// no way to mitigate this issue without these heartbeats.
//
// If you cannot heartbeat and cannot tolerate this kind of delayed-activity-loss detection, consider emulating a long
// activity via a signal channel instead: you can start a short-lived activity and wait for a "saved to external system"
// signal, retrying as necessary, and then wait for an "external system finished" signal containing the final result.
var ErrResultPending = internal.ErrActivityResultPending

// Register - calls RegisterWithOptions with default registration options.
// Deprecated: Global activity registration methods are replaced by equivalent Worker instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func Register(activityFunc interface{}) {
	internal.RegisterActivity(activityFunc)
}

// RegisterWithOptions registers the activity function with options
// The user can use options to provide an external name for the activity or leave it empty if no
// external name is required. This can be used as
//
//	client.Register(barActivity, RegisterOptions{})
//	client.Register(barActivity, RegisterOptions{Name: "barExternal"})
//
// An activity takes a context and input and returns a (result, error) or just error.
// Examples:
//
//	func sampleActivity(ctx context.Context, input []byte) (result []byte, err error)
//	func sampleActivity(ctx context.Context, arg1 int, arg2 string) (result *customerStruct, err error)
//	func sampleActivity(ctx context.Context) (err error)
//	func sampleActivity() (result string, err error)
//	func sampleActivity(arg1 bool) (result int, err error)
//	func sampleActivity(arg1 bool) (err error)
//
// Serialization of all primitive types, structures is supported ... except channels, functions, unsafe pointer.
// If function implementation returns activity.ErrResultPending then activity is not completed from the
// calling workflow point of view. See documentation of activity.ErrResultPending for more info.
// This method calls panic if activityFunc doesn't comply with the expected format.
// Deprecated: Global activity registration methods are replaced by equivalent Worker instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func RegisterWithOptions(activityFunc interface{}, opts RegisterOptions) {
	internal.RegisterActivityWithOptions(activityFunc, opts)
}

// GetInfo returns information about currently executing activity.
func GetInfo(ctx context.Context) Info {
	return internal.GetActivityInfo(ctx)
}

// HasInfo returns if the context contains activity information
func HasInfo(ctx context.Context) bool {
	return internal.HasActivityInfo(ctx)
}

// GetLogger returns a logger that can be used in activity
func GetLogger(ctx context.Context) *zap.Logger {
	return internal.GetActivityLogger(ctx)
}

// GetMetricsScope returns a metrics scope that can be used in activity
func GetMetricsScope(ctx context.Context) tally.Scope {
	return internal.GetActivityMetricsScope(ctx)
}

// RecordHeartbeat sends heartbeat for the currently executing activity
// If the activity is either cancelled (or) workflow/activity doesn't exist then we would cancel
// the context with error context.Canceled.
//
// details - the details that you provided here can be seen in the workflow when it receives TimeoutError, you
// can check error with TimeoutType()/Details().
func RecordHeartbeat(ctx context.Context, details ...interface{}) {
	internal.RecordActivityHeartbeat(ctx, details...)
}

// HasHeartbeatDetails checks if there is heartbeat details from last attempt.
func HasHeartbeatDetails(ctx context.Context) bool {
	return internal.HasHeartbeatDetails(ctx)
}

// GetHeartbeatDetails extract heartbeat details from last failed attempt. This is used in combination with retry policy.
// An activity could be scheduled with an optional retry policy on ActivityOptions. If the activity failed then server
// would attempt to dispatch another activity task to retry according to the retry policy. If there was heartbeat
// details reported by activity from the failed attempt, the details would be delivered along with the activity task for
// retry attempt. Activity could extract the details by GetHeartbeatDetails() and resume from the progress.
// See TestActivityEnvironment.SetHeartbeatDetails() for unit test support.
func GetHeartbeatDetails(ctx context.Context, d ...interface{}) error {
	return internal.GetHeartbeatDetails(ctx, d...)
}

// GetWorkerStopChannel returns a read-only channel. The closure of this channel indicates the activity worker is stopping.
// When the worker is stopping, it will close this channel and wait until the worker stop timeout finishes. After the timeout
// hit, the worker will cancel the activity context and then exit. The timeout can be defined by worker option: WorkerStopTimeout.
// Use this channel to handle activity graceful exit when the activity worker stops.
func GetWorkerStopChannel(ctx context.Context) <-chan struct{} {
	return internal.GetWorkerStopChannel(ctx)
}
