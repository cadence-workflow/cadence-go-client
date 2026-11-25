package batch

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.uber.org/multierr"

	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/testsuite"
)

// TODO: add clock simulation to speed up the test

type batchWorkflowInput struct {
	Concurrency int
	TotalSize   int
}

func batchWorkflow(ctx internal.Context, input batchWorkflowInput) ([]int, error) {
	factories := make([]func(ctx internal.Context) internal.Future, input.TotalSize)
	for i := 0; i < input.TotalSize; i++ {
		i := i
		factories[i] = func(ctx internal.Context) internal.Future {
			aCtx := internal.WithActivityOptions(ctx, internal.ActivityOptions{
				ScheduleToStartTimeout: time.Second * 10,
				StartToCloseTimeout:    time.Second * 10,
			})
			return internal.ExecuteActivity(aCtx, batchActivity, i)
		}
	}

	batchFuture, err := NewBatchFuture(ctx, input.Concurrency, factories)
	if err != nil {
		return nil, err
	}

	result := make([]int, input.TotalSize)
	err = batchFuture.Get(ctx, &result)
	return result, err
}

func batchWorkflowUsingFutures(ctx internal.Context, input batchWorkflowInput) ([]int, error) {
	factories := make([]func(ctx internal.Context) internal.Future, input.TotalSize)
	for i := 0; i < input.TotalSize; i++ {
		i := i
		factories[i] = func(ctx internal.Context) internal.Future {
			aCtx := internal.WithActivityOptions(ctx, internal.ActivityOptions{
				ScheduleToStartTimeout: time.Second * 10,
				StartToCloseTimeout:    time.Second * 10,
			})
			return internal.ExecuteActivity(aCtx, batchActivity, i)
		}
	}

	batchFuture, err := NewBatchFuture(ctx, input.Concurrency, factories)
	if err != nil {
		return nil, err
	}
	result := make([]int, input.TotalSize)

	for i, f := range batchFuture.GetFutures() {
		err = f.Get(ctx, &result[i])
		if err != nil {
			return nil, err
		}
	}

	return result, err
}

func batchActivity(ctx context.Context, taskID int) (int, error) {
	select {
	case <-ctx.Done():
		return taskID, fmt.Errorf("batch activity %d failed: %w", taskID, ctx.Err())
	case <-time.After(time.Duration(rand.Int63n(100))*time.Millisecond + 900*time.Millisecond):
		return taskID, nil
	}
}

func Test_BatchWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(batchWorkflow)
	env.RegisterActivity(batchActivity)

	totalSize := 100
	concurrency := 20

	startTime := time.Now()
	env.ExecuteWorkflow(batchWorkflow, batchWorkflowInput{
		Concurrency: concurrency,
		TotalSize:   totalSize,
	})

	assert.Less(t, time.Since(startTime), time.Second*time.Duration(float64(totalSize)/float64(concurrency)))
	assert.True(t, env.IsWorkflowCompleted())

	assert.Nil(t, env.GetWorkflowError())
	var result []int
	assert.Nil(t, env.GetWorkflowResult(&result))
	var expected []int
	for i := 0; i < totalSize; i++ {
		expected = append(expected, i)
	}
	assert.Equal(t, expected, result)
}

func Test_BatchWorkflow_Cancel(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(batchWorkflow)
	env.RegisterActivity(batchActivity)

	totalSize := 100
	concurrency := 20

	totalExpectedTime := time.Second * time.Duration(1+totalSize/concurrency)
	env.ExecuteWorkflow(batchWorkflow, batchWorkflowInput{
		Concurrency: concurrency,
		TotalSize:   totalSize,
	})

	env.RegisterDelayedCallback(func() {
		env.CancelWorkflow()
	}, totalExpectedTime/2)

	assert.True(t, env.IsWorkflowCompleted())

	err := env.GetWorkflowError()
	errs := multierr.Errors(errors.Unwrap(err))
	assert.Less(t, len(errs), totalSize, "expect at least some to succeed")
	for _, e := range errs {
		assert.Contains(t, e.Error(), "Canceled")
	}
}

func Test_BatchWorkflowUsingFutures(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(batchWorkflowUsingFutures)
	env.RegisterActivity(batchActivity)

	totalSize := 100
	concurrency := 20

	startTime := time.Now()
	env.ExecuteWorkflow(batchWorkflowUsingFutures, batchWorkflowInput{
		Concurrency: concurrency,
		TotalSize:   totalSize,
	})
	assert.Less(t, time.Since(startTime), time.Second*time.Duration(float64(totalSize)/float64(concurrency)))
	assert.True(t, env.IsWorkflowCompleted())

	assert.Nil(t, env.GetWorkflowError())
	var result []int
	assert.Nil(t, env.GetWorkflowResult(&result))
	var expected []int
	for i := 0; i < totalSize; i++ {
		expected = append(expected, i)
	}
	assert.Equal(t, expected, result)
}

func batchWorkflowAssignWithSlice(ctx internal.Context) ([]int, error) {
	totalSize := 5
	concurrency := 2
	factories := make([]func(ctx internal.Context) internal.Future, totalSize)
	for i := 0; i < totalSize; i++ {
		i := i
		factories[i] = func(ctx internal.Context) internal.Future {
			aCtx := internal.WithActivityOptions(ctx, internal.ActivityOptions{
				ScheduleToStartTimeout: time.Second * 10,
				StartToCloseTimeout:    time.Second * 10,
			})
			return internal.ExecuteActivity(aCtx, batchActivity, i)
		}
	}

	batchFuture, err := NewBatchFuture(ctx, concurrency, factories)
	if err != nil {
		return nil, err
	}

	var valuePtr []int
	if err := batchFuture.Get(ctx, &valuePtr); err != nil {
		return nil, err
	}
	return valuePtr, nil
}

func batchWorkflowAssignWithSliceOfPointers(ctx internal.Context) ([]int, error) {
	totalSize := 5
	concurrency := 2
	factories := make([]func(ctx internal.Context) internal.Future, totalSize)
	for i := 0; i < totalSize; i++ {
		i := i
		factories[i] = func(ctx internal.Context) internal.Future {
			aCtx := internal.WithActivityOptions(ctx, internal.ActivityOptions{
				ScheduleToStartTimeout: time.Second * 10,
				StartToCloseTimeout:    time.Second * 10,
			})
			return internal.ExecuteActivity(aCtx, batchActivity, i)
		}
	}
	batchFuture, err := NewBatchFuture(ctx, concurrency, factories)
	if err != nil {
		return nil, err
	}
	var valuePtr []*int
	if err := batchFuture.Get(ctx, &valuePtr); err != nil {
		return nil, err
	}

	var result []int
	for _, v := range valuePtr {
		result = append(result, *v)
	}
	return result, nil
}

func batchWorkflowAssignWithNil(ctx internal.Context) ([]int, error) {
	totalSize := 5
	concurrency := 2
	factories := make([]func(ctx internal.Context) internal.Future, totalSize)
	for i := 0; i < totalSize; i++ {
		i := i
		factories[i] = func(ctx internal.Context) internal.Future {
			aCtx := internal.WithActivityOptions(ctx, internal.ActivityOptions{
				ScheduleToStartTimeout: time.Second * 10,
				StartToCloseTimeout:    time.Second * 10,
			})
			return internal.ExecuteActivity(aCtx, batchActivity, i)
		}
	}

	batchFuture, err := NewBatchFuture(ctx, concurrency, factories)
	if err != nil {
		return nil, err
	}

	if err := batchFuture.Get(ctx, nil); err != nil {
		return nil, err
	}
	return nil, nil
}

func Test_BatchFuture_Get(t *testing.T) {
	tests := []struct {
		name     string
		workflow func(ctx internal.Context) ([]int, error)
		want     interface{}
	}{
		{
			name:     "success with nil slice",
			workflow: batchWorkflowAssignWithSlice,
			want:     []int{0, 1, 2, 3, 4},
		},
		{
			name:     "success with non-nil slice",
			workflow: batchWorkflowAssignWithSliceOfPointers,
			want:     []int{0, 1, 2, 3, 4},
		},
		{
			name:     "success with nil",
			workflow: batchWorkflowAssignWithNil,
			want:     []int(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testSuite := &testsuite.WorkflowTestSuite{}
			env := testSuite.NewTestWorkflowEnvironment()
			env.RegisterWorkflow(tt.workflow)
			env.RegisterActivity(batchActivity)
			env.ExecuteWorkflow(tt.workflow)
			assert.True(t, env.IsWorkflowCompleted())
			assert.Nil(t, env.GetWorkflowError())
			var result []int
			assert.Nil(t, env.GetWorkflowResult(&result))
			assert.Equal(t, tt.want, result)
		})
	}
}
