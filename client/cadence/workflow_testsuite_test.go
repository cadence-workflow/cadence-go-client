package cadence

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const testTaskList = "test-task-list"

type WorkflowTestSuiteUnitTest struct {
	WorkflowTestSuite
	activityOptions ActivityOptions
}

func (s *WorkflowTestSuiteUnitTest) SetupSuite() {
	ao := ActivityOptions{}
	ao.ScheduleToStartTimeout = time.Minute
	ao.StartToCloseTimeout = time.Minute
	ao.ScheduleToCloseTimeout = time.Minute
	ao.HeartbeatTimeout = 20 * time.Second
	s.activityOptions = ao
	s.RegisterActivity(testActivityHello)
	s.RegisterActivity(testActivityHeartbeat)
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuiteUnitTest))
}

func (s *WorkflowTestSuiteUnitTest) Test_ActivityOverride() {
	fakeActivity := func(ctx context.Context, msg string) (string, error) {
		return "fake_" + msg, nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.OverrideActivity(testActivityHello, fakeActivity)

	env.ExecuteWorkflow(testWorkflowHello)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	env.GetWorkflowResult(&result)
	s.Equal("fake_world", result)
}

func (s *WorkflowTestSuiteUnitTest) Test_OnActivityStartedListener() {
	workflowFn := func(ctx Context) error {
		ctx = WithActivityOptions(ctx, s.activityOptions)

		for i := 1; i <= 3; i++ {
			err := ExecuteActivity(ctx, testActivityHello, fmt.Sprintf("msg%d", i)).Get(ctx, nil)
			if err != nil {
				return err
			}
		}
		return nil
	} // END of workflow code

	env := s.NewTestWorkflowEnvironment()

	var activityCalls []string
	env.SetOnActivityStartedListener(func(ctx context.Context, args EncodedValues) {
		activityType := GetActivityInfo(ctx).ActivityType.Name
		var input string
		s.NoError(args.Get(&input))
		activityCalls = append(activityCalls, fmt.Sprintf("%s:%s", activityType, input))
	})
	expectedCalls := []string{
		"github.com/uber-go/cadence-client/client/cadence.testActivityHello:msg1",
		"github.com/uber-go/cadence-client/client/cadence.testActivityHello:msg2",
		"github.com/uber-go/cadence-client/client/cadence.testActivityHello:msg3",
	}

	env.ExecuteWorkflow(workflowFn)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(expectedCalls, activityCalls)
}

func (s *WorkflowTestSuiteUnitTest) Test_TimerWorkflow_ClockAutoFastForward() {
	var firedTimerRecord []string
	workflowFn := func(ctx Context) error {
		t1 := NewTimer(ctx, time.Second*5)
		t2 := NewTimer(ctx, time.Second*1)
		t3 := NewTimer(ctx, time.Second*2)
		t4 := NewTimer(ctx, time.Second*5)

		selector := NewSelector(ctx)
		selector.AddFuture(t1, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t1")
		}).AddFuture(t2, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t2")
		}).AddFuture(t3, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t3")
		}).AddFuture(t4, func(f Future) {
			firedTimerRecord = append(firedTimerRecord, "t4")
		})

		selector.Select(ctx)
		selector.Select(ctx)
		selector.Select(ctx)
		selector.Select(ctx)

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal([]string{"t2", "t3", "t1", "t4"}, firedTimerRecord)
}

func (s *WorkflowTestSuiteUnitTest) xTest_WorkflowAutoForwardClock() {
	/**
	TODO: update this test once we update the test workflow clock implementation.
	*/
	workflowFn := func(ctx Context) (string, error) {
		f1 := NewTimer(ctx, time.Second*2)
		ctx = WithActivityOptions(ctx, s.activityOptions)
		f2 := ExecuteActivity(ctx, testActivityHello, "controlled_execution")

		timerErr := f1.Get(ctx, nil) // wait until timer fires
		if timerErr != nil {
			return "", timerErr
		}

		if !f2.IsReady() {
			return "", errors.New("activity is not completed when timer fired")
		}

		var activityResult string
		activityErr := f2.Get(ctx, &activityResult)
		if activityErr != nil {
			return "", activityErr
		}

		return activityResult, nil
	} // END of workflow code

	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	env.GetWorkflowResult(&result)
	s.Equal("hello_controlled_execution", result)
}

func (s *WorkflowTestSuiteUnitTest) Test_WorkflowActivityCancellation() {
	workflowFn := func(ctx Context) (string, error) {
		ctx = WithActivityOptions(ctx, s.activityOptions)

		ctx, cancelHandler := WithCancel(ctx)
		f1 := ExecuteActivity(ctx, testActivityHeartbeat, "msg1", time.Millisecond) // fast activity
		f2 := ExecuteActivity(ctx, testActivityHeartbeat, "msg2", time.Second*3)    // slow activity

		selector := NewSelector(ctx)
		selector.AddFuture(f1, func(f Future) {
			cancelHandler()
		}).AddFuture(f2, func(f Future) {
			cancelHandler()
		})

		selector.Select(ctx)
		err := f2.Get(ctx, nil)
		if _, ok := err.(CanceledError); !ok {
			return "", err
		}

		GetLogger(ctx).Info("testWorkflowActivityCancellation completed.")
		return "result from testWorkflowActivityCancellation", nil
	}

	env := s.NewTestWorkflowEnvironment()
	counter := 0
	env.SetOnActivityEndedListener(func(result EncodedValue, err error, activityType string) {
		if err != nil {
			// assert err is CancelErr
			_, ok := err.(CanceledError)
			s.True(ok)
		} else {
			var msg string
			result.Get(&msg)
			s.Equal("ok_msg1", msg) // assert that fast activity finished
		}
		counter++
	})
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(2, counter) // assert listener get called twice
}

func (s *WorkflowTestSuiteUnitTest) Test_ActivityWithUserContext() {
	testKey, testValue := "test_key", "test_value"
	userCtx := context.WithValue(context.Background(), testKey, testValue)
	workerOptions := WorkerOptions{}
	workerOptions.ActivityContext = userCtx

	// inline activity using value passing through user context.
	activityWithUserContext := func(ctx context.Context, keyName string) (string, error) {
		value := ctx.Value(keyName)
		if value != nil {
			return value.(string), nil
		}
		return "", errors.New("value not found from ctx")
	}

	env := s.NewTestActivityEnvironment()
	env.SetWorkerOption(workerOptions)
	blob, err := env.ExecuteActivity(activityWithUserContext, testKey)
	s.NoError(err)
	var value string
	blob.Get(&value)
	s.Equal(testValue, value)
}

func (s *WorkflowTestSuiteUnitTest) Test_CompleteActivity() {
	env := s.NewTestWorkflowEnvironment()
	var activityInfo ActivityInfo
	fakeActivity := func(ctx context.Context, msg string) (string, error) {
		activityInfo = GetActivityInfo(ctx)
		env.RegisterDelayedCallback(func() {
			err := env.CompleteActivity(activityInfo.TaskToken, "async_complete", nil)
			s.NoError(err)
		}, time.Minute)
		return "", ErrActivityResultPending
	}

	env.OverrideActivity(testActivityHello, fakeActivity)
	env.SetTestTimeout(time.Second * 2) // don't waist time waiting

	env.ExecuteWorkflow(testWorkflowHello) // workflow won't complete, as the fakeActivity returns ErrActivityResultPending
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result string
	env.GetWorkflowResult(&result)
	s.Equal("async_complete", result)
}

func (s *WorkflowTestSuiteUnitTest) xTest_WorkflowCancellation() {
	/**
	TODO: fix test workflow clock implementation.
	This test is not working for now because current implementation only auto forward clock for timer when there is no
	running activities. As long as there is running activities, the workflow clock will stall. We need to change this
	behavior to continue moving forward workflow clock at real wall clock pace while there is running activities.
	*/
	workflowFn := func(ctx Context) error {
		ctx = WithActivityOptions(ctx, s.activityOptions)
		f := ExecuteActivity(ctx, testActivityHeartbeat, "msg1", time.Second*10)
		err := f.Get(ctx, nil) // wait for result
		return err
	}

	env := s.NewTestWorkflowEnvironment()
	env.SetTestTimeout(time.Minute)

	// Register a delayed callback using workflow timer internally. By default, the test suite enables the auto clock
	// forwarding when workflow is blocked. So, when the workflow is blocked on the testActivityHeartbeat activity, the
	// mock clock will auto forward and fires timer which calls our registered callback. Our callback would cancel the
	// workflow, which will terminate the whole workflow.
	env.RegisterDelayedCallback(func() {
		env.CancelWorkflow()
	}, time.Hour)

	var activityErr error
	env.SetOnActivityEndedListener(func(result EncodedValue, err error, activityType string) {
		activityErr = err
	})

	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NotNil(env.GetWorkflowError())
	_, ok := env.GetWorkflowError().(CanceledError)
	s.True(ok)

	// verify activity was cancelled as well.
	s.NotNil(activityErr)
	_, ok = activityErr.(CanceledError)
	s.True(ok)
}

func testWorkflowHello(ctx Context) (string, error) {
	ao := ActivityOptions{}
	ao.ScheduleToStartTimeout = time.Minute
	ao.StartToCloseTimeout = time.Minute
	ao.ScheduleToCloseTimeout = time.Minute
	ao.HeartbeatTimeout = 20 * time.Second
	ctx = WithActivityOptions(ctx, ao)

	var result string
	err := ExecuteActivity(ctx, testActivityHello, "world").Get(ctx, &result)
	if err != nil {
		return "", err
	}
	return result, nil
}

func testActivityHello(ctx context.Context, msg string) (string, error) {
	return "hello" + "_" + msg, nil
}

func testActivityHeartbeat(ctx context.Context, msg string, waitTime time.Duration) (string, error) {
	GetActivityLogger(ctx).Info("testActivityHeartbeat start",
		zap.String("msg", msg), zap.Duration("waitTime", waitTime))

	currWaitTime := time.Duration(0)
	for currWaitTime < waitTime {
		RecordActivityHeartbeat(ctx)
		select {
		case <-ctx.Done():
			// We have been cancelled.
			return "", ctx.Err()
		default:
			// We are not cancelled yet.
		}

		sleepDuration := time.Second
		if currWaitTime+sleepDuration > waitTime {
			sleepDuration = waitTime - currWaitTime
		}
		time.Sleep(sleepDuration)
		currWaitTime += sleepDuration
	}

	return "ok_" + msg, nil
}
