package batch

import (
	"fmt"
	"reflect"

	"go.uber.org/multierr"

	"go.uber.org/cadence/internal"
)

type BatchFuture interface {
	internal.Future
	GetFutures() []internal.Future
}

type batchFutureImpl struct {
	futures   []internal.Future
	settables []internal.Settable
	factories []func(ctx internal.Context) internal.Future
	batchSize int

	// state
	wg internal.WaitGroup
}

func NewBatchFuture(ctx internal.Context, batchSize int, factories []func(ctx internal.Context) internal.Future) (BatchFuture, error) {
	var futures []internal.Future
	var settables []internal.Settable
	for range factories {
		future, settable := internal.NewFuture(ctx)
		futures = append(futures, future)
		settables = append(settables, settable)
	}

	batchFuture := &batchFutureImpl{
		futures:   futures,
		settables: settables,
		factories: factories,
		batchSize: batchSize,

		wg: internal.NewWaitGroup(ctx),
	}
	batchFuture.start(ctx)
	return batchFuture, nil
}

func (b *batchFutureImpl) GetFutures() []internal.Future {
	return b.futures
}

func (b *batchFutureImpl) start(ctx internal.Context) {

	buffered := internal.NewBufferedChannel(ctx, b.batchSize) // buffered channel to limit the number of concurrent futures
	channel := internal.NewNamedChannel(ctx, "batch-future-channel")
	b.wg.Add(1)
	internal.GoNamed(ctx, "batch-future-submitter", func(ctx internal.Context) {
		defer b.wg.Done()

		for i := range b.factories {
			buffered.Send(ctx, nil)
			channel.Send(ctx, i)
		}
		channel.Close()
	})

	b.wg.Add(1)
	internal.GoNamed(ctx, "batch-future-processor", func(ctx internal.Context) {
		defer b.wg.Done()

		wgForFutures := internal.NewWaitGroup(ctx)

		var idx int
		for channel.Receive(ctx, &idx) {
			idx := idx

			wgForFutures.Add(1)
			internal.GoNamed(ctx, "batch-future-processor-one-future", func(ctx internal.Context) {
				defer wgForFutures.Done()

				// fork a future and chain it to the processed future for user to get the result
				f := b.factories[idx](ctx)
				b.settables[idx].Chain(f)

				// error handling is not needed here because the result is chained to the settable
				f.Get(ctx, nil)
				buffered.Receive(ctx, nil)
			})
		}
		wgForFutures.Wait(ctx)
	})
}

func (b *batchFutureImpl) IsReady() bool {
	for _, future := range b.futures {
		if !future.IsReady() {
			return false
		}
	}
	return true
}

func (b *batchFutureImpl) Get(ctx internal.Context, valuePtr interface{}) error {
	// ensure valuePtr is a slice
	var sliceValue reflect.Value
	if valuePtr != nil {

		switch v := reflect.ValueOf(valuePtr); v.Kind() {
		case reflect.Ptr:
			if v.Elem().Kind() != reflect.Slice {
				return fmt.Errorf("valuePtr must be a pointer to a slice, got %v", v)
			}
			sliceValue = v.Elem()
		case reflect.Slice:
			sliceValue = v
		default:
			return fmt.Errorf("valuePtr must be a slice or a pointer to a slice, got %v", v.Kind())
		}
		// ensure slice size is the same as the number of futures
		if sliceValue.Len() != len(b.futures) {
			return fmt.Errorf("slice size must be the same as the number of futures, got %d, expected %d", sliceValue.Len(), len(b.futures))
		}
	}

	// wait for all futures to be ready
	b.wg.Wait(ctx)

	// loop through all elements of valuePtr
	var errs error
	for i := range b.futures {
		if valuePtr == nil {
			errs = multierr.Append(errs, b.futures[i].Get(ctx, nil))
		} else {
			value := sliceValue.Index(i)
			if value.Kind() != reflect.Ptr {
				value = value.Addr()
			}
			// if value is nil, initialize it
			if value.IsNil() {
				value.Set(reflect.New(value.Type().Elem()))
			}

			e := b.futures[i].Get(ctx, value.Interface())
			errs = multierr.Append(errs, e)
		}
	}

	return errs
}
