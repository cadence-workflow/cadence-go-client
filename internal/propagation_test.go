package internal

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
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

	returnedCtx, span := createOpenTracingWorkflowSpan(ctx, tracer, time.Now(), "test-workflow", "wf-id", true)

	assert.NotNil(t, span)
	assert.Nil(t, opentracing.SpanFromContext(returnedCtx))
}

func TestCreateOpenTracingWorkflowSpan_WithActiveSpan(t *testing.T) {
	t.Parallel()
	tracer := mocktracer.New()
	parentSpan := tracer.StartSpan("parent")
	ctx := opentracing.ContextWithSpan(context.Background(), parentSpan)

	returnedCtx, span := createOpenTracingWorkflowSpan(ctx, tracer, time.Now(), "test-workflow", "wf-id", true)

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

	_, span := createOpenTracingWorkflowSpan(ctx, tracer, time.Now(), "test-workflow", "wf-id", true)

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

// TestStartWorkflowCronTagPropagation verifies end-to-end that starting a
// workflow records the cadenceIsCron span tag on the tracer, reflecting whether
// a cron schedule was configured on the start options.
func TestStartWorkflowCronTagPropagation(t *testing.T) {
	tests := []struct {
		name           string
		cronSchedule   string
		expectedIsCron bool
	}{
		{
			name:           "cron workflow sets tag to true",
			cronSchedule:   "* * * * *",
			expectedIsCron: true,
		},
		{
			name:           "non-cron workflow sets tag to false",
			cronSchedule:   "",
			expectedIsCron: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			service := workflowservicetest.NewMockClient(mockCtrl)
			tracer := mocktracer.New()
			client := NewClient(service, domain, &ClientOptions{
				Identity: identity,
				Tracer:   tracer,
			})

			service.EXPECT().
				StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&shared.StartWorkflowExecutionResponse{RunId: common.StringPtr(runID)}, nil).
				Times(1)

			options := StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        tasklist,
				ExecutionStartToCloseTimeout:    timeoutInSeconds,
				DecisionTaskStartToCloseTimeout: timeoutInSeconds,
				CronSchedule:                    tt.cronSchedule,
			}
			f1 := func(ctx Context, r []byte) string { return "result" }

			resp, err := client.StartWorkflow(context.Background(), options, f1, []byte("test"))
			require.NoError(t, err)
			require.Equal(t, runID, resp.RunID)

			finished := tracer.FinishedSpans()
			require.Len(t, finished, 1, "starting a workflow should produce exactly one span")
			span := finished[0]

			require.Equal(t, tt.expectedIsCron, span.Tag(tagCadenceIsCron),
				"cadenceIsCron tag should reflect the cron schedule")
			require.Equal(t, workflowID, span.Tag(tagCadenceWorkflowID))
		})
	}
}
