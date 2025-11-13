package resettabletimer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/internal/common/testlogger"
	"go.uber.org/cadence/x/resettabletimer"
)

type ResettableTimerTestSuite struct {
	suite.Suite
	internal.WorkflowTestSuite
}

func (s *ResettableTimerTestSuite) SetupTest() {
	s.SetLogger(testlogger.NewZap(s.T()))
}

func TestResettableTimerSuite(t *testing.T) {
	suite.Run(t, new(ResettableTimerTestSuite))
}

func (s *ResettableTimerTestSuite) TestTimerFiresAfterExpiration() {
	wf := func(ctx internal.Context) error {
		startTime := internal.Now(ctx)
		timer := resettabletimer.New(ctx, 5*time.Second)

		timerFired := false
		internal.NewSelector(ctx).AddFuture(timer.Future, func(f internal.Future) {
			err := f.Get(ctx, nil)
			s.NoError(err)
			timerFired = true
		}).Select(ctx)

		elapsedTime := internal.Now(ctx).Sub(startTime)
		s.Equal(5*time.Second, elapsedTime, "5 Seconds should have elapsed")
		s.True(timerFired, "Timer should fire after 5 seconds with no resets")
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.ExecuteWorkflow(wf)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *ResettableTimerTestSuite) TestTimerDoesNotFireWhenKeptReset() {
	wf := func(ctx internal.Context) error {
		d := 2 * time.Second
		timer := resettabletimer.New(ctx, d)
		activityCh := internal.GetSignalChannel(ctx, "activity")
		stopCh := internal.GetSignalChannel(ctx, "stop")

		timerFired := false
		activityCount := 0
		stopped := false

		for !stopped {
			selector := internal.NewSelector(ctx)

			selector.AddFuture(timer.Future, func(f internal.Future) {
				timerFired = true
			})

			selector.AddReceive(activityCh, func(c internal.Channel, more bool) {
				var signal string
				c.Receive(ctx, &signal)
				activityCount++
				timer.Reset(d)
			})

			selector.AddReceive(stopCh, func(c internal.Channel, more bool) {
				var stop bool
				c.Receive(ctx, &stop)
				stopped = true
			})

			selector.Select(ctx)
		}

		s.Equal(3, activityCount, "Should have received 3 activity signals")
		s.False(timerFired, "Timer should NOT fire when continuously reset")
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("activity", "scan1")
	}, 500*time.Millisecond)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("activity", "scan2")
	}, 1500*time.Millisecond)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("activity", "scan3")
	}, 2500*time.Millisecond)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("stop", true)
	}, 4*time.Second)

	env.ExecuteWorkflow(wf)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *ResettableTimerTestSuite) TestTimerFiresAfterResetsStop() {
	wf := func(ctx internal.Context) error {
		d := 5 * time.Second
		timer := resettabletimer.New(ctx, d)

		timer.Reset(d)               // Still 5s
		timer.Reset(2 * time.Second) // Now 2s

		timerFired := false
		internal.NewSelector(ctx).AddFuture(timer.Future, func(f internal.Future) {
			err := f.Get(ctx, nil)
			s.NoError(err)
			timerFired = true
		}).Select(ctx)

		s.True(timerFired, "Timer should eventually fire after resets stop")
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.ExecuteWorkflow(wf)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *ResettableTimerTestSuite) TestTimerResetAfterFireIsNoop() {
	wf := func(ctx internal.Context) error {
		d := 1 * time.Second
		timer := resettabletimer.New(ctx, d)

		err := timer.Future.Get(ctx, nil)
		s.NoError(err)
		s.True(timer.Future.IsReady(), "Timer should be ready after firing")

		timer.Reset(5 * time.Second)
		s.True(timer.Future.IsReady(), "Timer should still be ready (reset was no-op)")

		timer.Reset(d)
		s.True(timer.Future.IsReady(), "Multiple resets after fire are all no-ops")
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.ExecuteWorkflow(wf)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *ResettableTimerTestSuite) TestTimerResetWithDifferentDurations() {
	wf := func(ctx internal.Context) error {
		timer := resettabletimer.New(ctx, 10*time.Second)

		timer.Reset(1 * time.Second)

		timerFired := false
		startTime := internal.Now(ctx)

		internal.NewSelector(ctx).AddFuture(timer.Future, func(f internal.Future) {
			err := f.Get(ctx, nil)
			s.NoError(err)
			timerFired = true
		}).Select(ctx)

		elapsed := internal.Now(ctx).Sub(startTime)

		s.True(timerFired, "Timer should fire")
		s.Less(elapsed, 5*time.Second, "Should fire with reset duration, not original")
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.ExecuteWorkflow(wf)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
