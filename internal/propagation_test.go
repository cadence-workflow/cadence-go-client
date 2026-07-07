package internal

import (
	"context"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartOpenTracingSpan_NoopTracer_DoesNotWrapContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tracer := opentracing.NoopTracer{}

	returnedCtx, span := startOpenTracingSpan(ctx, tracer, time.Now(), "test-op", nil, nil)

	assert.NotNil(t, span)
	assert.Nil(t, opentracing.SpanFromContext(returnedCtx), "NoopTracer should not inject span into context")
}

func TestStartOpenTracingSpan_RealTracer_WrapsContext(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tracer := mocktracer.New()

	returnedCtx, span := startOpenTracingSpan(ctx, tracer, time.Now(), "test-op", nil, nil)

	assert.NotNil(t, span)
	assert.Equal(t, span, opentracing.SpanFromContext(returnedCtx), "real tracer should inject span into context")
}

func TestStartOpenTracingSpan_WithParent(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	parentSpan := tracer.StartSpan("parent")

	_, childSpan := startOpenTracingSpan(
		context.Background(), tracer, time.Now(), "child",
		opentracing.Tags{"key": "val"}, parentSpan.Context(),
	)

	require.NotNil(t, childSpan)
	ms := childSpan.(*mocktracer.MockSpan)
	assert.Equal(t, parentSpan.(*mocktracer.MockSpan).SpanContext.SpanID, ms.ParentID)
}

func TestCreateOpenTracingWorkflowSpan_NoopTracer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tracer := opentracing.NoopTracer{}

	returnedCtx, span := createOpenTracingWorkflowSpan(ctx, tracer, time.Now(), "test-workflow", "wf-id")

	assert.NotNil(t, span)
	assert.Nil(t, opentracing.SpanFromContext(returnedCtx))
}

func TestCreateOpenTracingWorkflowSpan_WithActiveSpan(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	parentSpan := tracer.StartSpan("parent")
	ctx := opentracing.ContextWithSpan(context.Background(), parentSpan)

	returnedCtx, span := createOpenTracingWorkflowSpan(ctx, tracer, time.Now(), "test-workflow", "wf-id")

	require.NotNil(t, span)
	assert.NotNil(t, opentracing.SpanFromContext(returnedCtx))
	ms := span.(*mocktracer.MockSpan)
	assert.Equal(t, parentSpan.(*mocktracer.MockSpan).SpanContext.SpanID, ms.ParentID)
}

func TestCreateOpenTracingWorkflowSpan_WithSpanContextKey(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	parentSpan := tracer.StartSpan("parent")
	ctx := context.WithValue(context.Background(), activeSpanContextKey, parentSpan.Context())

	_, span := createOpenTracingWorkflowSpan(ctx, tracer, time.Now(), "test-workflow", "wf-id")

	require.NotNil(t, span)
	ms := span.(*mocktracer.MockSpan)
	assert.Equal(t, parentSpan.(*mocktracer.MockSpan).SpanContext.SpanID, ms.ParentID)
}

func TestCreateOpenTracingSpanFromHeaders_NoopTracer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tracer := opentracing.NoopTracer{}

	returnedCtx, span := createOpenTracingSpanFromHeaders(ctx, tracer, time.Now(), "test-activity", nil)

	assert.NotNil(t, span)
	assert.Nil(t, opentracing.SpanFromContext(returnedCtx))
}

func TestCreateOpenTracingSpanFromHeaders_WithSpanContextKey(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	parentSpan := tracer.StartSpan("parent")
	ctx := context.WithValue(context.Background(), activeSpanContextKey, parentSpan.Context())

	returnedCtx, span := createOpenTracingSpanFromHeaders(ctx, tracer, time.Now(), "test-activity", nil)

	require.NotNil(t, span)
	assert.NotNil(t, opentracing.SpanFromContext(returnedCtx))
	ms := span.(*mocktracer.MockSpan)
	assert.Equal(t, parentSpan.(*mocktracer.MockSpan).SpanContext.SpanID, ms.ParentID)
}

func TestCreateOpenTracingSpanFromHeaders_IgnoresActiveSpan(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	activeSpan := tracer.StartSpan("active")
	ctx := opentracing.ContextWithSpan(context.Background(), activeSpan)

	_, span := createOpenTracingSpanFromHeaders(ctx, tracer, time.Now(), "test-activity", nil)

	require.NotNil(t, span)
	ms := span.(*mocktracer.MockSpan)
	assert.Equal(t, 0, ms.ParentID, "should not use active span as parent — only uses activeSpanContextKey")
}
