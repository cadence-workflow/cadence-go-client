package cadence

// All code in this file is private to the package.

import (
	"errors"
	"time"

	"golang.org/x/net/context"

	"fmt"
	"github.com/uber-go/cadence-client/common"
	"reflect"
)

// Assert that structs do indeed implement the interfaces
var _ ActivityOptions = (*activityOptions)(nil)

type (
	activityInfo struct {
		activityID string
	}

	// executeActivityParameters configuration parameters for scheduling an activity
	executeActivityParameters struct {
		ActivityID                    *string // Users can choose IDs but our framework makes it optional to decrease the crust.
		ActivityType                  ActivityType
		TaskListName                  string
		Input                         []byte
		ScheduleToCloseTimeoutSeconds int32
		ScheduleToStartTimeoutSeconds int32
		StartToCloseTimeoutSeconds    int32
		HeartbeatTimeoutSeconds       int32
		WaitForCancellation           bool
	}

	// asyncActivityClient for requesting activity execution
	asyncActivityClient interface {
		// The ExecuteActivity schedules an activity with a callback handler.
		// If the activity failed to complete the callback error would indicate the failure
		// and it can be one of ActivityTaskFailedError, ActivityTaskTimeoutError, ActivityTaskCanceledError
		ExecuteActivity(parameters executeActivityParameters, callback resultHandler) *activityInfo

		// This only initiates cancel request for activity. if the activity is configured to not waitForCancellation then
		// it would invoke the callback handler immediately with error code ActivityTaskCanceledError.
		// If the activity is not running(either scheduled or started) then it is a no-operation.
		RequestCancelActivity(activityID string)
	}

	activityEnvironment struct {
		taskToken         []byte
		workflowExecution WorkflowExecution
		activityID        string
		activityType      ActivityType
		serviceInvoker    ServiceInvoker
	}
)

const activityEnvContextKey = "activityEnv"

func getActivityEnv(ctx context.Context) *activityEnvironment {
	env := ctx.Value(activityEnvContextKey)
	if env == nil {
		panic("getActivityEnv: Not an activity context")
	}
	return env.(*activityEnvironment)
}

const activityOptionsContextKey = "activityOptions"

func getActivityOptions(ctx Context) *executeActivityParameters {
	eap := ctx.Value(activityOptionsContextKey)
	if eap == nil {
		return nil
	}
	return eap.(*executeActivityParameters)
}

func getValidatedActivityOptions(ctx Context) (*executeActivityParameters, error) {
	p := getActivityOptions(ctx)
	if p == nil {
		// We need task list as a compulsory parameter. This can be removed after registration
		return nil, errActivityParamsBadRequest
	}
	if p.ScheduleToStartTimeoutSeconds <= 0 {
		return nil, errors.New("missing or negative ScheduleToStartTimeoutSeconds")
	}
	if p.ScheduleToCloseTimeoutSeconds <= 0 {
		return nil, errors.New("missing or negative ScheduleToCloseTimeoutSeconds")
	}
	if p.StartToCloseTimeoutSeconds <= 0 {
		return nil, errors.New("missing or negative StartToCloseTimeoutSeconds")
	}
	return p, nil
}

func marshalFunctionArgs(f interface{}, args ...interface{}) ([]byte, error) {
	fnName := getFunctionName(f)
	s := fnSignature{FnName: fnName, Args: args}
	input, err := getHostEnvironment().Encoder().Marshal(s)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func validateFunctionArgs(f interface{}, args ...interface{}) error {
	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		return fmt.Errorf("Provided type: %v is not a function type", f)
	}
	fnName := getFunctionName(f)
	// Validate provided args match with function order match.
	if fType.NumIn() != len(args) {
		return fmt.Errorf(
			"expected %d function: %v args but found %v",
			fType.NumIn(), fnName, len(args))
	}
	for i := 0; i < fType.NumIn(); i++ {
		argType := reflect.TypeOf(args[i])
		if !argType.AssignableTo(fType.In(i)) {
			return fmt.Errorf(
				"cannot assign function argument: %d from type: %s to type: %s",
				i+1, argType, fType.In(i),
			)
		}
	}

	return nil
}

func getValidatedActivityFunction(f interface{}, args ...interface{}) (ActivityType, []byte, error) {
	var activityType ActivityType

	fType := reflect.TypeOf(f)
	switch fType.Kind() {
	case reflect.String:
		activityType.Name = reflect.ValueOf(f).String()

	case reflect.Func:
		if err := validateFunctionArgs(f, args); err != nil {
			return activityType, nil, err
		}
		fnName := getFunctionName(f)
		activityType.Name = fnName
	default:
		return activityType, nil, fmt.Errorf(
			"Invalid type 'f' parameter provided, it can be either functor (or) name of the activity type: %v", f)
	}

	input, err := marshalFunctionArgs(f, args)
	if err != nil {
		return activityType, nil, err
	}
	return activityType, input, nil
}

func setActivityParametersIfNotExist(ctx Context) Context {
	if valCtx := getActivityOptions(ctx); valCtx == nil {
		return WithValue(ctx, activityOptionsContextKey, &executeActivityParameters{})
	}
	return ctx
}

// activityOptions stores all activity-specific parameters that will
// be stored inside of a context.
type activityOptions struct {
	activityID                    *string
	taskListName                  *string
	scheduleToCloseTimeoutSeconds *int32
	scheduleToStartTimeoutSeconds *int32
	startToCloseTimeoutSeconds    *int32
	heartbeatTimeoutSeconds       *int32
	waitForCancellation           *bool
}

// WithTaskList sets the task list name for this Context.
func (ab *activityOptions) WithTaskList(name string) ActivityOptions {
	ab.taskListName = common.StringPtr(name)
	return ab
}

// WithScheduleToCloseTimeout sets timeout for this Context.
func (ab *activityOptions) WithScheduleToCloseTimeout(d time.Duration) ActivityOptions {
	ab.scheduleToCloseTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithScheduleToStartTimeout sets timeout for this Context.
func (ab *activityOptions) WithScheduleToStartTimeout(d time.Duration) ActivityOptions {
	ab.scheduleToStartTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithStartToCloseTimeout sets timeout for this Context.
func (ab *activityOptions) WithStartToCloseTimeout(d time.Duration) ActivityOptions {
	ab.startToCloseTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithHeartbeatTimeout sets timeout for this Context.
func (ab *activityOptions) WithHeartbeatTimeout(d time.Duration) ActivityOptions {
	ab.heartbeatTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithWaitForCancellation sets timeout for this Context.
func (ab *activityOptions) WithWaitForCancellation(wait bool) ActivityOptions {
	ab.waitForCancellation = &wait
	return ab
}

// WithActivityID sets the activity task list ID for this Context.
// NOTE: We don't expose configuring Activity ID to the user, This is something will be done in future
// so they have end to end scenario of how to use this ID to complete and fail an activity(business use case).
func (ab *activityOptions) WithActivityID(activityID string) ActivityOptions {
	ab.activityID = common.StringPtr(activityID)
	return ab
}
