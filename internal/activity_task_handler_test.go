// Copyright (c) 2017-2021 Uber Technologies Inc.
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
	"errors"
	"fmt"
	"testing"
	"time"

	"go.uber.org/cadence/internal/common/testlogger"

	"github.com/jonboulle/clockwork"

	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

func TestActivityTaskHandler_Execute_deadline(t *testing.T) {
	now := time.Now()
	deadlineTests := []deadlineTest{
		{time.Duration(0), now, 3, now, 3, nil},
		{time.Duration(0), now, 4, now, 3, nil},
		{time.Duration(0), now, 3, now, 4, nil},
		{time.Duration(0), now.Add(-1 * time.Second), 1, now, 1, context.DeadlineExceeded},
		{time.Duration(0), now, 1, now.Add(-1 * time.Second), 1, context.DeadlineExceeded},
		{time.Duration(0), now.Add(-1 * time.Second), 1, now.Add(-1 * time.Second), 1, context.DeadlineExceeded},
		{1 * time.Second, now, 1, now, 1, context.DeadlineExceeded},
		{1 * time.Second, now, 2, now, 1, context.DeadlineExceeded},
		{1 * time.Second, now, 1, now, 2, context.DeadlineExceeded},
	}

	for i, d := range deadlineTests {
		t.Run(fmt.Sprintf("testIndex: %v, testDetails: %v", i, d), func(t *testing.T) {
			logger := testlogger.NewZap(t)
			a := &testActivityDeadline{logger: logger}
			registry := newRegistry()
			registry.addActivityWithLock(a.ActivityType().Name, a)

			mockCtrl := gomock.NewController(t)
			mockService := workflowservicetest.NewMockClient(mockCtrl)
			a.d = d.actWaitDuration
			wep := workerExecutionParameters{
				WorkerOptions: WorkerOptions{
					Logger:        logger,
					DataConverter: getDefaultDataConverter(),
					Tracer:        opentracing.NoopTracer{},
				},
			}
			ensureRequiredParams(&wep)
			activityHandler := newActivityTaskHandler(mockService, wep, registry)
			pats := &s.PollForActivityTaskResponse{
				TaskToken: []byte("token"),
				WorkflowExecution: &s.WorkflowExecution{
					WorkflowId: common.StringPtr("wID"),
					RunId:      common.StringPtr("rID")},
				ActivityType:                    &s.ActivityType{Name: common.StringPtr("test")},
				ActivityId:                      common.StringPtr(uuid.New()),
				ScheduledTimestamp:              common.Int64Ptr(d.ScheduleTS.UnixNano()),
				ScheduledTimestampOfThisAttempt: common.Int64Ptr(d.ScheduleTS.UnixNano()),
				ScheduleToCloseTimeoutSeconds:   common.Int32Ptr(d.ScheduleDuration),
				StartedTimestamp:                common.Int64Ptr(d.StartTS.UnixNano()),
				StartToCloseTimeoutSeconds:      common.Int32Ptr(d.StartDuration),
				Attempt:                         common.Int32Ptr(0),
				WorkflowType: &s.WorkflowType{
					Name: common.StringPtr("wType"),
				},
				WorkflowDomain: common.StringPtr("domain"),
			}
			r, err := activityHandler.Execute(tasklist, pats)
			assert.Equal(t, d.err, err)
			if err != nil {
				assert.Nil(t, r)
			}
		})
	}
}

func TestActivityTaskHandler_Execute_worker_stop(t *testing.T) {
	logger := testlogger.NewZap(t)

	a := &testActivityDeadline{logger: logger}
	registry := newRegistry()
	registry.RegisterActivityWithOptions(
		activityWithWorkerStop,
		RegisterActivityOptions{Name: a.ActivityType().Name, DisableAlreadyRegisteredCheck: true},
	)

	now := time.Now()

	mockCtrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(mockCtrl)
	workerStopCh := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	wep := workerExecutionParameters{
		WorkerOptions: WorkerOptions{
			Logger:        logger,
			DataConverter: getDefaultDataConverter(),
		},
		UserContext:       ctx,
		UserContextCancel: cancel,
		WorkerStopChannel: workerStopCh,
	}
	activityHandler := newActivityTaskHandler(mockService, wep, registry)
	pats := &s.PollForActivityTaskResponse{
		TaskToken: []byte("token"),
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr("wID"),
			RunId:      common.StringPtr("rID")},
		ActivityType:                    &s.ActivityType{Name: common.StringPtr("test")},
		ActivityId:                      common.StringPtr(uuid.New()),
		ScheduledTimestamp:              common.Int64Ptr(now.UnixNano()),
		ScheduledTimestampOfThisAttempt: common.Int64Ptr(now.UnixNano()),
		ScheduleToCloseTimeoutSeconds:   common.Int32Ptr(1),
		StartedTimestamp:                common.Int64Ptr(now.UnixNano()),
		StartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Attempt:                         common.Int32Ptr(0),
		WorkflowType: &s.WorkflowType{
			Name: common.StringPtr("wType"),
		},
		WorkflowDomain: common.StringPtr("domain"),
	}
	close(workerStopCh)
	r, err := activityHandler.Execute(tasklist, pats)
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestActivityTaskHandler_Execute_with_propagators(t *testing.T) {
	logger := testlogger.NewZap(t)

	now := time.Now()

	testKey, testValue := testHeader, "test_value"

	// inline activity using value passing through user context.
	activityWithUserContext := func(ctx context.Context) (string, error) {
		value := ctx.Value(contextKey(testKey))
		if value != nil {
			return value.(string), nil
		}
		return "", errors.New("value not found from ctx")
	}
	registry := newRegistry()
	err := registry.registerActivityFunction(activityWithUserContext, RegisterActivityOptions{Name: "activityWithUserContext"})
	require.NoError(t, err)

	mockCtrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(mockCtrl)
	wep := workerExecutionParameters{
		WorkerOptions: WorkerOptions{
			Logger:             logger,
			DataConverter:      getDefaultDataConverter(),
			Tracer:             opentracing.NoopTracer{},
			ContextPropagators: []ContextPropagator{NewStringMapPropagator([]string{testHeader})},
		},
	}
	ensureRequiredParams(&wep)
	activityHandler := newActivityTaskHandler(mockService, wep, registry)
	pats := &s.PollForActivityTaskResponse{
		TaskToken: []byte("token"),
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr("wID"),
			RunId:      common.StringPtr("rID")},
		ActivityType:                    &s.ActivityType{Name: common.StringPtr("activityWithUserContext")},
		ActivityId:                      common.StringPtr(uuid.New()),
		ScheduledTimestamp:              common.Int64Ptr(now.UnixNano()),
		ScheduledTimestampOfThisAttempt: common.Int64Ptr(now.UnixNano()),
		ScheduleToCloseTimeoutSeconds:   common.Int32Ptr(1),
		StartedTimestamp:                common.Int64Ptr(now.UnixNano()),
		StartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Attempt:                         common.Int32Ptr(0),
		WorkflowType: &s.WorkflowType{
			Name: common.StringPtr("wType"),
		},
		WorkflowDomain: common.StringPtr("domain"),
		Header: &s.Header{
			Fields: map[string][]byte{testHeader: []byte(testValue)},
		},
	}
	res, err := activityHandler.Execute(tasklist, pats)
	require.NoError(t, err)
	response, ok := res.(*s.RespondActivityTaskCompletedRequest)
	require.True(t, ok, "response is not of type *s.RespondActivityTaskCompletedRequest")
	assert.Equal(t, fmt.Sprintf("\"%s\"\n", testValue), string(response.Result))
}

func TestActivityTaskHandler_Execute_with_propagator_failure(t *testing.T) {
	logger := testlogger.NewZap(t)

	now := time.Now()

	// inline activity using value passing through user context.
	activityWithUserContext := func(ctx context.Context) (string, error) {
		return "", nil
	}
	registry := newRegistry()
	err := registry.registerActivityFunction(activityWithUserContext, RegisterActivityOptions{Name: "activityWithUserContext"})
	require.NoError(t, err)

	mockCtrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(mockCtrl)
	wep := workerExecutionParameters{
		WorkerOptions: WorkerOptions{
			Logger:             logger,
			DataConverter:      getDefaultDataConverter(),
			Tracer:             opentracing.NoopTracer{},
			ContextPropagators: []ContextPropagator{failingContextPropagator{err: assert.AnError}},
		},
	}
	ensureRequiredParams(&wep)
	activityHandler := newActivityTaskHandler(mockService, wep, registry)
	pats := &s.PollForActivityTaskResponse{
		TaskToken: []byte("token"),
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr("wID"),
			RunId:      common.StringPtr("rID")},
		ActivityType:                    &s.ActivityType{Name: common.StringPtr("activityWithUserContext")},
		ActivityId:                      common.StringPtr(uuid.New()),
		ScheduledTimestamp:              common.Int64Ptr(now.UnixNano()),
		ScheduledTimestampOfThisAttempt: common.Int64Ptr(now.UnixNano()),
		ScheduleToCloseTimeoutSeconds:   common.Int32Ptr(1),
		StartedTimestamp:                common.Int64Ptr(now.UnixNano()),
		StartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		Attempt:                         common.Int32Ptr(0),
		WorkflowType: &s.WorkflowType{
			Name: common.StringPtr("wType"),
		},
		WorkflowDomain: common.StringPtr("domain"),
	}
	_, err = activityHandler.Execute(tasklist, pats)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestActivityTaskHandler_Execute_with_auto_heartbeat(t *testing.T) {
	logger := testlogger.NewZap(t)

	now := time.Now()

	clock := clockwork.NewFakeClock()

	// inline activity using value passing through user context.
	activityWithUserContext := func(ctx context.Context) error {
		// heartbeat once
		clock.Advance(time.Second / 2)
		return nil
	}
	registry := newRegistry()
	err := registry.registerActivityFunction(activityWithUserContext, RegisterActivityOptions{
		Name:                "activityWithUserContext",
		EnableAutoHeartbeat: true,
	})
	require.NoError(t, err)

	mockCtrl := gomock.NewController(t)
	mockService := workflowservicetest.NewMockClient(mockCtrl)
	wep := workerExecutionParameters{
		WorkerOptions: WorkerOptions{
			Logger:        logger,
			DataConverter: getDefaultDataConverter(),
			Tracer:        opentracing.NoopTracer{},
		},
	}
	ensureRequiredParams(&wep)
	activityHandler := newActivityTaskHandler(mockService, wep, registry)
	activityHandlerImpl := activityHandler.(*activityTaskHandlerImpl)
	activityHandlerImpl.clock = clock

	pats := &s.PollForActivityTaskResponse{
		TaskToken: []byte("token"),
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr("wID"),
			RunId:      common.StringPtr("rID")},
		ActivityType:                    &s.ActivityType{Name: common.StringPtr("activityWithUserContext")},
		ActivityId:                      common.StringPtr(uuid.New()),
		ScheduledTimestamp:              common.Int64Ptr(now.UnixNano()),
		ScheduledTimestampOfThisAttempt: common.Int64Ptr(now.UnixNano()),
		ScheduleToCloseTimeoutSeconds:   common.Int32Ptr(1),
		StartedTimestamp:                common.Int64Ptr(now.UnixNano()),
		StartToCloseTimeoutSeconds:      common.Int32Ptr(1),
		HeartbeatTimeoutSeconds:         common.Int32Ptr(1),
		Attempt:                         common.Int32Ptr(0),
		WorkflowType: &s.WorkflowType{
			Name: common.StringPtr("wType"),
		},
		WorkflowDomain: common.StringPtr("domain"),
	}
	res, err := activityHandler.Execute(tasklist, pats)
	require.NoError(t, err)
	_, ok := res.(*s.RespondActivityTaskCompletedRequest)
	require.True(t, ok, "response is not of type *s.RespondActivityTaskCompletedRequest but of type %T", res)
}

func activityWithWorkerStop(ctx context.Context) error {
	fmt.Println("Executing Activity with worker stop")
	workerStopCh := GetWorkerStopChannel(ctx)

	select {
	case <-workerStopCh:
		return nil
	case <-time.NewTimer(time.Second * 5).C:
		return fmt.Errorf("Activity failed to handle worker stop event")
	}
}

type deadlineTest struct {
	actWaitDuration  time.Duration
	ScheduleTS       time.Time
	ScheduleDuration int32
	StartTS          time.Time
	StartDuration    int32
	err              error
}

type failingContextPropagator struct {
	err error
}

func (f failingContextPropagator) Inject(ctx context.Context, writer HeaderWriter) error {
	return f.err
}

func (f failingContextPropagator) Extract(ctx context.Context, reader HeaderReader) (context.Context, error) {
	return nil, f.err
}

func (f failingContextPropagator) InjectFromWorkflow(ctx Context, writer HeaderWriter) error {
	return f.err
}

func (f failingContextPropagator) ExtractToWorkflow(ctx Context, reader HeaderReader) (Context, error) {
	return nil, f.err
}
