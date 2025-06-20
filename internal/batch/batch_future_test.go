package batch

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.uber.org/multierr"

	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/testsuite"
)

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
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		env.ExecuteWorkflow(batchWorkflow, batchWorkflowInput{
			Concurrency: concurrency,
			TotalSize:   totalSize,
		})
	}()

	time.Sleep(totalExpectedTime / 2)
	env.CancelWorkflow()

	wg.Wait() // wait for the workflow to complete otherwise test env will have race condition
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

func futureTest(ctx internal.Context) error {
	f, s := internal.NewFuture(ctx)
	f2, s2 := internal.NewFuture(ctx)
	s2.Chain(f)

	wg := internal.NewWaitGroup(ctx)
	wg.Add(1)
	internal.GoNamed(ctx, "future-test", func(ctx internal.Context) {
		defer wg.Done()
		internal.Sleep(ctx, time.Second*10)
		s.Set(1, nil)
	})

	err := f2.Get(ctx, nil)
	if err != nil {
		return err
	}

	err = f.Get(ctx, nil)
	if err != nil {
		return err
	}

	wg.Wait(ctx)
	return err
}

func Test_Futures(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(futureTest)

	env.ExecuteWorkflow(futureTest)
}

func Test_valuePtr(t *testing.T) {
	slices := make([]int, 10)
	slicePtr := &slices

	fmt.Println(reflect.ValueOf(slicePtr).Elem().Len())
}
