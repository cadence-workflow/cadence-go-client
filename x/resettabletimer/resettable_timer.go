package resettabletimer

import (
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
)

type (
	// ResettableTimer represents a timer that can be reset to restart its countdown.
	ResettableTimer interface {
		workflow.Future

		// Reset - Cancels the current timer and starts a new one with the given duration.
		// If the timer has already fired, Reset has no effect.
		Reset(d time.Duration)
	}

	Timer struct {
		ctx         workflow.Context
		timerCtx    workflow.Context
		cancelTimer workflow.CancelFunc
		// This is suboptimal, but we cannot implement the internal asyncFuture interface because it is not exported. It is what it is.
		Future   workflow.Future
		settable workflow.Settable
		duration time.Duration
		isReady  bool
	}
)

// New returns a timer that can be reset to restart its countdown. The timer becomes ready after the
// specified duration d. The timer can be reset using timer.Reset(duration) with a new duration. This is useful for
// implementing timeout patterns that should restart based on external events. The workflow needs to use this
// New() instead of creating new timers repeatedly. The current timer resolution implementation is in
// seconds and uses math.Ceil(d.Seconds()) as the duration. But is subjected to change in the future.
func New(ctx workflow.Context, d time.Duration) *Timer {
	rt := &Timer{
		ctx:      ctx,
		duration: d,
	}
	rt.Future, rt.settable = workflow.NewFuture(ctx)
	rt.startTimer(d)
	return rt
}

func (rt *Timer) startTimer(d time.Duration) {
	rt.duration = d

	if rt.cancelTimer != nil {
		rt.cancelTimer()
	}

	rt.timerCtx, rt.cancelTimer = workflow.WithCancel(rt.ctx)

	timer := workflow.NewTimer(rt.timerCtx, d)

	workflow.Go(rt.ctx, func(ctx workflow.Context) {
		err := timer.Get(ctx, nil)

		if !cadence.IsCanceledError(err) && !rt.isReady {
			rt.isReady = true
			rt.settable.Set(nil, err)
		}
	})
}

// Reset - Cancels the current timer and starts a new one with the given duration.
// If the timer has already fired, Reset has no effect.
func (rt *Timer) Reset(d time.Duration) {
	if !rt.isReady {
		rt.startTimer(d)
	}
}
