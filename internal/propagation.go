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
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
)

// Span names and tag keys align with the Cadence Java client's TracingPropagator.
// See: https://github.com/cadence-workflow/cadence-java-client/blob/master/src/main/java/com/uber/cadence/internal/tracing/TracingPropagator.java
const (
	spanNameExecuteWorkflow      = "cadence-ExecuteWorkflow"
	spanNameExecuteActivity      = "cadence-ExecuteActivity"
	spanNameExecuteLocalActivity = "cadence-ExecuteLocalActivity"

	tagCadenceWorkflowType = "cadenceWorkflowType"
	tagCadenceActivityType = "cadenceActivityType"
	tagCadenceWorkflowID   = "cadenceWorkflowID"
	tagCadenceRunID        = "cadenceRunID"
)

// createOpenTracingWorkflowSpan creates a new context with a workflow started span
func createOpenTracingWorkflowSpan(
	ctx context.Context,
	tracer opentracing.Tracer,
	start time.Time,
	workflowType, workflowID string,
) (context.Context, opentracing.Span) {
	tags := opentracing.Tags{
		tagCadenceWorkflowID: workflowID,
	}
	var parent opentracing.SpanContext
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		parent = parentSpan.Context()
	} else if spanCtx, ok := ctx.Value(activeSpanContextKey).(opentracing.SpanContext); ok {
		parent = spanCtx
	}
	return startOpenTracingSpan(ctx, tracer, start, workflowType, tags, parent)
}

// createOpenTracingSpanFromHeaders creates a new context with a started span, linking FollowsFrom
// only span context in activeSpanContextKey (set by ContextPropagator.Extract). This matches Java's
// ignoreActiveSpan behavior for worker execute workflow/activity/local-activity spans.
func createOpenTracingSpanFromHeaders(
	ctx context.Context,
	tracer opentracing.Tracer,
	start time.Time,
	name string,
	tags opentracing.Tags,
) (context.Context, opentracing.Span) {
	var parent opentracing.SpanContext
	if spanCtx, ok := ctx.Value(activeSpanContextKey).(opentracing.SpanContext); ok {
		parent = spanCtx
	}
	return startOpenTracingSpan(ctx, tracer, start, name, tags, parent)
}

func startOpenTracingSpan(
	ctx context.Context,
	tracer opentracing.Tracer,
	start time.Time,
	name string,
	tags opentracing.Tags,
	parent opentracing.SpanContext,
) (context.Context, opentracing.Span) {
	opts := []opentracing.StartSpanOption{
		opentracing.StartTime(start),
		tags,
	}
	if parent != nil {
		opts = append(opts, opentracing.FollowsFrom(parent))
	}

	span := tracer.StartSpan(name, opts...)
	ctx = opentracing.ContextWithSpan(ctx, span)
	return ctx, span
}
