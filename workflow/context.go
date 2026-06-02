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

package workflow

import (
	"github.com/opentracing/opentracing-go"

	"go.uber.org/cadence/internal"
)

// Context is a clone of context.Context with Done() returning Channel instead
// of native channel.
// A Context carries a deadline, a cancellation signal, and other values across
// API boundaries.
//
// Context's methods may be called by multiple goroutines simultaneously.
type Context = internal.Context

// ErrCanceled is the error returned by Context.Err when the context is canceled.
var ErrCanceled = internal.ErrCanceled

// ErrDeadlineExceeded is the error returned by Context.Err when the context's
// deadline passes.
var ErrDeadlineExceeded = internal.ErrDeadlineExceeded

// A CancelFunc tells an operation to abandon its work.
// A CancelFunc does not wait for the work to stop.
// After the first call, subsequent calls to a CancelFunc do nothing.
type CancelFunc = internal.CancelFunc

// WithCancel returns a copy of parent with a new Done channel. The returned
// context's Done channel is closed when the returned cancel function is called
// or when the parent context's Done channel is closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	return internal.WithCancel(parent)
}

// WithValue returns a copy of parent in which the value associated with key is
// val.
//
// Use context Values only for request-scoped data that transits processes and
// APIs, not for passing optional parameters to functions.
func WithValue(parent Context, key interface{}, val interface{}) Context {
	return internal.WithValue(parent, key, val)
}

// NewDisconnectedContext returns a new context that won't propagate parent's cancellation to the new child context.
// One common use case is to do cleanup work after workflow is cancelled.
//
//	err := workflow.ExecuteActivity(ctx, ActivityFoo).Get(ctx, &activityFooResult)
//	if err != nil && cadence.IsCanceledError(ctx.Err()) {
//	  // activity failed, and workflow context is canceled
//	  disconnectedCtx, _ := workflow.newDisconnectedContext(ctx);
//	  workflow.ExecuteActivity(disconnectedCtx, handleCancellationActivity).Get(disconnectedCtx, nil)
//	  return err // workflow return CanceledError
//	}
func NewDisconnectedContext(parent Context) (ctx Context, cancel CancelFunc) {
	return internal.NewDisconnectedContext(parent)
}

// GetSpanContext returns the [opentracing.SpanContext] from [Context].
// Returns nil if tracer is not set in [go.uber.org/cadence/worker.Options].
//
// When a tracer is configured, the worker sets span context during
// non-replay workflow execution ("cadence-ExecuteWorkflow" span) and activities
// receive propagated context from workflow headers.
//
// Example Usage:
//
//	span := GetSpanContext(ctx)
//	if span != nil {
//		span.SetTag("foo", "bar")
//	}
func GetSpanContext(ctx Context) opentracing.SpanContext {
	return internal.GetSpanContext(ctx)
}

// WithSpanContext returns [Context] with override [opentracing.SpanContext].
// This is useful to modify baggage items of current workflow and pass it to activities and child workflows.
//
// Example Usage:
//
//	func goodWorkflow(ctx Context) (string, error) {
//		// start a short lived new workflow span within SideEffect to avoid duplicate span creation during replay
//		type spanActivationResult struct {
//			Carrier map[string]string // exported field so it's json encoded
//			Err     error
//		}
//		resultValue := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
//			wSpan := w.tracer.StartSpan("workflow-operation-with-new-span", opentracing.ChildOf(workflow.GetSpanContext(ctx)))
//			defer wSpan.Finish()
//			wSpan.SetBaggageItem("some-key", "some-value")
//			carrier := make(map[string]string)
//			err := w.tracer.Inject(wSpan.Context(), opentracing.TextMap, opentracing.TextMapCarrier(carrier))
//			return spanActivationResult{Carrier: carrier, Err: err}
//		})
//		var activationResult spanActivationResult
//		if err := resultValue.Get(&activationResult); err != nil {
//			return nil, fmt.Errorf("failed to decode span activation result: %v", err)
//		}
//		if activationResult.Err != nil {
//			return nil, fmt.Errorf("failed to activate new span: %v", activationResult.Err)
//		}
//		spanContext, err := w.tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(activationResult.Carrier))
//		if err != nil {
//			return nil, fmt.Errorf("failed to extract span context: %v", err)
//		}
//		ctx = workflow.WithSpanContext(ctx, spanContext)
//		var activityFooResult string
//		aCtx := workflow.WithActivityOptions(ctx, opts)
//		err = workflow.ExecuteActivity(aCtx, activityFoo).Get(aCtx, &activityFooResult)
//		return activityFooResult, err
//	}
//
// Bad Example:
//
//	func badWorkflow(ctx Context) (string, error) {
//		// start a new workflow span for EVERY REPLAY
//		wSpan := opentracing.StartSpan("workflow-operation", opentracing.ChildOf(GetSpanContext(ctx)))
//		wSpan.SetBaggageItem("some-key", "some-value")
//		// pass the new span context to activity
//		ctx = WithSpanContext(ctx, wSpan.Context())
//	    aCtx := workflow.WithActivityOptions(ctx, opts)
//		var activityFooResult string
//		err := ExecuteActivity(aCtx, activityFoo).Get(aCtx, &activityFooResult)
//		wSpan.Finish()
//		return activityFooResult, err
//	}
//
//	func activityFoo(ctx Context) (string, error) {
//		return "activity-foo-result", nil
//	}
func WithSpanContext(ctx Context, spanContext opentracing.SpanContext) Context {
	return internal.WithSpanContext(ctx, spanContext)
}
