package workflow

import (
	"go.uber.org/cadence/internal/executor"
)

// Executor is a interface that allows to add futures to an executor and execute them in batch
type Executor interface {
	AddFuture(factory func(ctx Context) Future, valuePtr interface{}) error
	Execute(ctx Context) error
}

// NewBatchExecutor creates a new batch executor
func NewBatchExecutor(ctx Context, batchSize int) Executor {
	return executor.NewBatchExecutor(ctx, batchSize)
}
