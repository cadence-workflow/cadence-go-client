package executor

import (
	"errors"

	"go.uber.org/multierr"

	"go.uber.org/cadence/internal"
)

type Executor interface {
	AddFuture(factory func(ctx internal.Context) internal.Future, valuePtr interface{}) error
	Execute(ctx internal.Context) error
}

type batchExecutor struct {
	isRunning bool
	factories []func(ctx internal.Context) internal.Future
	valuePtrs []interface{}

	batchSize int
}

type futureWithResult struct {
	future   internal.Future
	valuePtr interface{}
}

func (e *batchExecutor) AddFuture(factory func(ctx internal.Context) internal.Future, valuePtr interface{}) error {
	if e.isRunning {
		return errors.New("executor is already running")
	}
	e.factories = append(e.factories, factory)
	e.valuePtrs = append(e.valuePtrs, valuePtr)
	return nil
}

func (e *batchExecutor) Execute(ctx internal.Context) error {
	if e.isRunning {
		return errors.New("executor is already running")
	}
	e.isRunning = true

	futuresToProcess := internal.NewNamedChannel(ctx, "futures")
	wg := internal.NewWaitGroup(ctx)

	buffered := internal.NewBufferedChannel(ctx, e.batchSize)
	var errs error
	wg.Add(1)

	// future processor
	internal.Go(ctx, func(ctx internal.Context) {
		defer wg.Done()
		for {
			var futureWithResult futureWithResult
			ok := futuresToProcess.Receive(ctx, &futureWithResult)
			if !ok {
				break
			}
			wg.Add(1)
			internal.Go(ctx, func(ctx internal.Context) {
				defer wg.Done()
				err := futureWithResult.future.Get(ctx, futureWithResult.valuePtr)
				errs = multierr.Append(errs, err)
				buffered.Receive(ctx, nil)
			})
		}
	})

	// submit all futures
	for i := range e.factories {
		buffered.Send(ctx, nil)
		futuresToProcess.Send(ctx, futureWithResult{
			future:   e.factories[i](ctx),
			valuePtr: e.valuePtrs[i],
		})
	}

	// close the channel to signal the task result collector that no more tasks are coming
	futuresToProcess.Close()

	wg.Wait(ctx)

	return errs
}

func NewBatchExecutor(ctx internal.Context, batchSize int) Executor {
	return &batchExecutor{
		batchSize: batchSize,
	}
}
