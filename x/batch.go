package x

import (
	"go.uber.org/cadence/internal/batch"
	"go.uber.org/cadence/workflow"
)

// BatchFuture is a interface that extends workflow.Future and adds a method to get the futures created
type BatchFuture interface {
	workflow.Future
	GetFutures() []workflow.Future
}

// NewBatchFuture creates a new batch future
func NewBatchFuture(ctx workflow.Context, batchSize int, factories []func(ctx workflow.Context) workflow.Future) (BatchFuture, error) {
	return batch.NewBatchFuture(ctx, batchSize, factories)
}
