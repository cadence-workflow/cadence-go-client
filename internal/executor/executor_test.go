package executor

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/testsuite"
)

type batchWorkflowInput struct {
	Concurrency int
	TotalSize   int
}

func batchWorkflow(ctx internal.Context, input batchWorkflowInput) ([]int, error) {
	executor := NewBatchExecutor(ctx, input.Concurrency)
	result := make([]int, input.TotalSize)

	for i := 0; i < input.TotalSize; i++ {
		i := i
		executor.AddFuture(func(ctx internal.Context) internal.Future {
			aCtx := internal.WithActivityOptions(ctx, internal.ActivityOptions{
				ScheduleToStartTimeout: time.Second * 10,
				StartToCloseTimeout:    time.Second * 10,
			})
			return internal.ExecuteActivity(aCtx, batchActivity, i)
		}, &result[i])
	}

	errs := executor.Execute(ctx)

	return result, errs
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

	env.ExecuteWorkflow(batchWorkflow, batchWorkflowInput{
		Concurrency: 200,
		TotalSize:   1000,
	})

	assert.True(t, env.IsWorkflowCompleted())
	assert.Nil(t, env.GetWorkflowError())
	var result []int
	assert.Nil(t, env.GetWorkflowResult(&result))
	var expected []int
	for i := 0; i < 1000; i++ {
		expected = append(expected, i)
	}
	assert.Equal(t, expected, result)
}
