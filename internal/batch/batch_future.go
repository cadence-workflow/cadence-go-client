package batch

import (
	"fmt"
	"reflect"

	"go.uber.org/multierr"

	"go.uber.org/cadence/internal"
)

// BatchFuture is an implementation of public BatchFuture interface.
type BatchFuture struct {
	futures   []internal.Future
	settables []internal.Settable
	factories []func(ctx internal.Context) internal.Future
	batchSize int

	// state
	wg internal.WaitGroup
}

func NewBatchFuture(ctx internal.Context, batchSize int, factories []func(ctx internal.Context) internal.Future) (*BatchFuture, error) {
	var futures []internal.Future
	var settables []internal.Settable
	for range factories {
		future, settable := internal.NewFuture(ctx)
		futures = append(futures, future)
		settables = append(settables, settable)
	}

	batchFuture := &BatchFuture{
		futures:   futures,
		settables: settables,
		factories: factories,
		batchSize: batchSize,

		wg: internal.NewWaitGroup(ctx),
	}
	batchFuture.start(ctx)
	return batchFuture, nil
}

func (b *BatchFuture) GetFutures() []internal.Future {
	return b.futures
}

func (b *BatchFuture) start(ctx internal.Context) {

	semaphore := internal.NewBufferedChannel(ctx, b.batchSize) // buffered workChan to limit the number of concurrent futures
	workChan := internal.NewNamedChannel(ctx, "batch-future-channel")
	b.wg.Add(1)
	internal.GoNamed(ctx, "batch-future-submitter", func(ctx internal.Context) {
		defer b.wg.Done()

		for i := range b.factories {
			semaphore.Send(ctx, nil)
			workChan.Send(ctx, i)
		}
		workChan.Close()
	})

	b.wg.Add(1)
	internal.GoNamed(ctx, "batch-future-processor", func(ctx internal.Context) {
		defer b.wg.Done()

		wgForFutures := internal.NewWaitGroup(ctx)

		var idx int
		for workChan.Receive(ctx, &idx) {
			idx := idx

			wgForFutures.Add(1)
			internal.GoNamed(ctx, fmt.Sprintf("batch-future-processor-one-future-%d", idx), func(ctx internal.Context) {
				defer wgForFutures.Done()

				// fork a future and chain it to the processed future for user to get the result
				f := b.factories[idx](ctx)
				b.settables[idx].Chain(f)

				// error handling is not needed here because the result is chained to the settable
				f.Get(ctx, nil)
				semaphore.Receive(ctx, nil)
			})
		}
		wgForFutures.Wait(ctx)
	})
}

func (b *BatchFuture) IsReady() bool {
	for _, future := range b.futures {
		if !future.IsReady() {
			return false
		}
	}
	return true
}

// Get assigns the result of the futures to the valuePtr.
// NOTE: valuePtr must be a pointer to a slice, or nil.
// If valuePtr is a pointer to a slice, the slice will be resized to the length of the futures. Each element of the slice will be assigned with the underlying Future.Get() and thus behaves the same way.
// If valuePtr is nil, no assignment will be made.
// If error occurs, values will be set on successful futures and the errors of failed futures will be returned.
func (b *BatchFuture) Get(ctx internal.Context, valuePtr interface{}) error {
	// No assignment if valuePtr is nil
	if valuePtr == nil {
		b.wg.Wait(ctx)
		var errs error
		for i := range b.futures {
			errs = multierr.Append(errs, b.futures[i].Get(ctx, nil))
		}
		return errs
	}

	v := reflect.ValueOf(valuePtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("valuePtr must be a pointer to a slice, got %v", v.Kind())
	}

	// resize the slice to the length of the futures
	slice := v.Elem()
	if slice.Cap() < len(b.futures) {
		slice.Grow(len(b.futures) - slice.Cap())
	}
	slice.SetLen(len(b.futures))

	// wait for all futures to be ready
	b.wg.Wait(ctx)

	// loop through all elements of valuePtr
	var errs error
	for i := range b.futures {
		e := b.futures[i].Get(ctx, slice.Index(i).Addr().Interface())
		errs = multierr.Append(errs, e)
	}

	return errs
}
