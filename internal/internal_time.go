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

import (
	"time"
)

// All code in this file is private to the package.

type (
	timerInfo struct {
		timerID string
	}

	// workflowTimerClient wraps the async workflow timer functionality.
	workflowTimerClient interface {

		// Now - Current time when the decision task is started or replayed.
		// the workflow need to use this for wall clock to make the flow logic deterministic.
		Now() time.Time

		// NewTimer - Creates a new timer that will fire callback after d(resolution is in seconds).
		// The callback indicates the error(TimerCanceledError) if the timer is cancelled.
		NewTimer(d time.Duration, callback resultHandler) *timerInfo

		// RequestCancelTimer - Requests cancel of a timer, this one doesn't wait for cancellation request
		// to complete, instead invokes the resultHandler with TimerCanceledError
		// If the timer is not started then it is a no-operation.
		RequestCancelTimer(timerID string)
	}

	// ResettableTimer represents a timer that can be reset to restart its countdown.
	ResettableTimer interface {
		Future

		// Reset - Cancels the current timer and starts a new one with the given duration.
		// If duration is not provided, uses the previous duration.
		// If the timer has already fired, Reset has no effect.
		Reset(d ...time.Duration)
	}

	resettableTimerImpl struct {
		ctx         Context
		timerCtx    Context
		cancelTimer CancelFunc
		future      Future
		settable    Settable
		duration    time.Duration
		isReady     bool
	}
)

// NewResettableTimer creates a new resettable timer that fires after duration d.
// The timer can be reset using Reset() to restart the countdown.
func NewResettableTimer(ctx Context, d time.Duration) ResettableTimer {
	rt := &resettableTimerImpl{
		ctx:      ctx,
		duration: d,
	}
	rt.future, rt.settable = NewFuture(ctx)
	rt.startTimer(d)
	return rt
}

func (rt *resettableTimerImpl) startTimer(d time.Duration) {
	rt.duration = d

	if rt.cancelTimer != nil {
		rt.cancelTimer()
	}

	rt.timerCtx, rt.cancelTimer = WithCancel(rt.ctx)

	timer := NewTimer(rt.timerCtx, d)

	Go(rt.ctx, func(ctx Context) {
		err := timer.Get(ctx, nil)

		if !IsCanceledError(err) && !rt.isReady {
			rt.isReady = true
			rt.settable.Set(nil, err)
		}
	})
}

func (rt *resettableTimerImpl) Reset(d ...time.Duration) {
	if rt.isReady {
		return
	}

	duration := rt.duration
	if len(d) > 0 {
		duration = d[0]
	}

	rt.startTimer(duration)
}

// Future interface delegation methods

func (rt *resettableTimerImpl) Get(ctx Context, valuePtr interface{}) error {
	return rt.future.Get(ctx, valuePtr)
}

func (rt *resettableTimerImpl) IsReady() bool {
	return rt.future.IsReady()
}

// asyncFuture interface delegation methods (needed for Selector.AddFuture)

func (rt *resettableTimerImpl) GetAsync(callback *receiveCallback) (v interface{}, ok bool, err error) {
	return rt.future.(asyncFuture).GetAsync(callback)
}

func (rt *resettableTimerImpl) RemoveReceiveCallback(callback *receiveCallback) {
	rt.future.(asyncFuture).RemoveReceiveCallback(callback)
}

func (rt *resettableTimerImpl) ChainFuture(f Future) {
	rt.future.(asyncFuture).ChainFuture(f)
}

func (rt *resettableTimerImpl) GetValueAndError() (v interface{}, err error) {
	return rt.future.(asyncFuture).GetValueAndError()
}

func (rt *resettableTimerImpl) Set(value interface{}, err error) {
	rt.future.(asyncFuture).Set(value, err)
}
