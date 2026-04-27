package fanout

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"go.uber.org/multierr"

	"go.uber.org/cadence/workflow"
)

const (
	DefaultMaxConcurrentActivities = 10
	DefaultMaxConcurrentChildren   = 10
)

func init() {
	workflow.Register(executeActivityInBulkChildWorkflow[any, any])
}

type (
	Options struct {
		MaxConcurrentActivities int
		FanoutOptions           *FanoutOptions
	}

	Option func(*Options)

	// FanoutOptions controls child workflow fanout behavior
	FanoutOptions struct {
		// MaxActivitiesPerWorkflow is the threshold for fanning out to child workflows.
		// If len(items) > MaxActivitiesPerWorkflow, ExecuteManyActivities will fan out
		// to multiple child workflows instead of executing all activities in the parent.
		MaxActivitiesPerWorkflow int

		// MaxConcurrentChildren limits how many child workflows can run concurrently.
		// Only relevant when fanout is triggered.
		MaxConcurrentChildren int

		// ChildWorkflowOptions controls how each child workflow is scheduled.
		ChildWorkflowOptions workflow.ChildWorkflowOptions

		// ActivityOptions controls how each activity is executed.
		ActivityOptions workflow.ActivityOptions
	}

	BulkFuture[R any] interface {
		IsReady() bool
		Get(ctx workflow.Context) ([]R, error)
	}

	bulkFuture[R any] struct {
		futures []workflow.Future
	}

	bulkFutureFromChildren[R any] struct {
		childFutures []workflow.ChildWorkflowFuture
		numItems     int
	}
)

// IsReady returns true when all wrapped futures are ready
func (b *bulkFuture[R]) IsReady() bool {
	for _, f := range b.futures {
		if !f.IsReady() {
			return false
		}
	}
	return true
}

// Get waits for all futures to complete and returns the results
func (b *bulkFuture[R]) Get(ctx workflow.Context) ([]R, error) {
	results := make([]R, len(b.futures))
	var errs error

	for i, f := range b.futures {
		var result R
		err := f.Get(ctx, &result)
		if err != nil {
			errs = multierr.Append(errs, err)
		}
		results[i] = result
	}

	return results, errs
}

// IsReady returns true when all child workflow futures are ready
func (b *bulkFutureFromChildren[R]) IsReady() bool {
	for _, f := range b.childFutures {
		if !f.IsReady() {
			return false
		}
	}
	return true
}

// Get waits for all child workflows to complete and returns the flattened results
func (b *bulkFutureFromChildren[R]) Get(ctx workflow.Context) ([]R, error) {
	allResults := make([]R, 0, b.numItems)
	var errs error

	for _, childFuture := range b.childFutures {
		var chunkResults []R
		err := childFuture.Get(ctx, &chunkResults)
		if err != nil {
			errs = multierr.Append(errs, err)
		}

		allResults = append(allResults, chunkResults...)
	}

	return allResults, errs
}

// ExecuteActivityInBulk executes many activities with automatic fanout to child workflows when needed.
//
// It provides two execution modes:
//  1. Direct execution: All activities run in the parent workflow (default)
//  2. Child workflow fanout: When items exceed MaxActivitiesPerWorkflow, work is distributed
//     across multiple child workflows to avoid overloading a single workflow/shard
//
// Type parameters:
//   - T: Input type for each activity
//   - R: Result type from each activity
//
// Parameters:
//   - ctx: Workflow context with activity options already set
//   - activity: Activity function to execute
//   - items: Slice of items to process
//   - options: Configuration options (WithConcurrency, WithFanoutOptions)
//
// Returns:
//   - BulkFuture[R]: Future that resolves to []R when all activities complete
func ExecuteActivityInBulk[T, R any](
	ctx workflow.Context,
	activity any,
	items []T,
	options ...Option,
) BulkFuture[R] {
	opts := &Options{
		MaxConcurrentActivities: DefaultMaxConcurrentActivities,
	}

	for _, option := range options {
		option(opts)
	}

	shouldFanout := opts.FanoutOptions != nil &&
		opts.FanoutOptions.MaxActivitiesPerWorkflow > 0 &&
		len(items) > opts.FanoutOptions.MaxActivitiesPerWorkflow

	if shouldFanout {
		return executeActivityInBulkWithChildren[T, R](ctx, activity, opts.FanoutOptions.ActivityOptions, items, opts)
	}

	return executeActivityInBulkInPlace[T, R](ctx, activity, items, opts)
}

// executeActivityInBulkInPlace executes all activities in the current workflow using BatchFuture
func executeActivityInBulkInPlace[T, R any](
	ctx workflow.Context,
	activity any,
	items []T,
	opts *Options,
) BulkFuture[R] {
	factories := make([]func(workflow.Context) workflow.Future, len(items))
	for i, item := range items {
		item := item // Capture loop variable for closure
		factories[i] = func(ctx workflow.Context) workflow.Future {
			return workflow.ExecuteActivity(ctx, activity, item)
		}
	}

	batchFuture, err := workflow.NewBatchFuture(ctx, opts.MaxConcurrentActivities, factories)
	if err != nil {
		// This shouldn't happen in practice, but we need to handle it
		panic(err)
	}

	return &bulkFuture[R]{
		futures: batchFuture.GetFutures(),
	}
}

func getFunctionName(i any) string {
	if name, ok := i.(string); ok {
		return name
	}
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	return strings.TrimSuffix(fullName, "-fm")
}

// executeActivityInBulkWithChildren fans out work to multiple child workflows
func executeActivityInBulkWithChildren[T, R any](
	ctx workflow.Context,
	activity any,
	ao workflow.ActivityOptions,
	items []T,
	opts *Options,
) BulkFuture[R] {
	maxPerChild := opts.FanoutOptions.MaxActivitiesPerWorkflow
	maxConcurrentChildren := opts.FanoutOptions.MaxConcurrentChildren

	numChildren := (len(items) + maxPerChild - 1) / maxPerChild

	chunks := make([][]T, numChildren)
	for i, item := range items {
		childIdx := i * numChildren / len(items)
		chunks[childIdx] = append(chunks[childIdx], item)
	}

	childFutures := make([]workflow.ChildWorkflowFuture, numChildren)

	activityName := getFunctionName(activity)
	childWorkflowName := getFunctionName(executeActivityInBulkChildWorkflow[any, any])

	semaphore := workflow.NewBufferedChannel(ctx, maxConcurrentChildren)
	wg := workflow.NewWaitGroup(ctx)

	for i, chunk := range chunks {
		i, chunk := i, chunk // Capture loop variables

		wg.Add(1)
		workflow.Go(ctx, func(ctx workflow.Context) {
			defer wg.Done()

			semaphore.Send(ctx, nil)
			defer func() {
				semaphore.Receive(ctx, nil)
			}()

			parentInfo := workflow.GetInfo(ctx)
			childID := fmt.Sprintf("%s-fanout-child-%d", parentInfo.WorkflowExecution.ID, i)
			opts.FanoutOptions.ChildWorkflowOptions.WorkflowID = childID
			childCtx := workflow.WithChildOptions(ctx, opts.FanoutOptions.ChildWorkflowOptions)

			childFutures[i] = workflow.ExecuteChildWorkflow(
				childCtx,
				childWorkflowName,
				ao,
				activityName,
				chunk,
				opts.MaxConcurrentActivities,
			)
		})
	}

	wg.Wait(ctx)

	return &bulkFutureFromChildren[R]{
		childFutures: childFutures,
		numItems:     len(items),
	}
}

// executeActivityInBulkChildWorkflow is the internal workflow that each child runs
func executeActivityInBulkChildWorkflow[T, R any](
	ctx workflow.Context,
	ao workflow.ActivityOptions,
	activityName string,
	items []T,
	concurrency int,
) ([]R, error) {
	ctx = workflow.WithActivityOptions(ctx, ao)
	factories := make([]func(workflow.Context) workflow.Future, len(items))
	for i, item := range items {
		item := item
		factories[i] = func(ctx workflow.Context) workflow.Future {
			return workflow.ExecuteActivity(ctx, activityName, item)
		}
	}

	batchFuture, err := workflow.NewBatchFuture(ctx, concurrency, factories)
	if err != nil {
		// This shouldn't happen in practice, but we need to handle it
		return nil, err
	}

	results := make([]R, len(items))
	futures := batchFuture.GetFutures()
	var errs error

	for i, f := range futures {
		err := f.Get(ctx, &results[i])
		if err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return results, errs
}

// WithMaxConcurrentActivities sets the number of concurrent activities to run within a singular workflow.
func WithMaxConcurrentActivities(n int) Option {
	return func(opts *Options) {
		opts.MaxConcurrentActivities = n
	}
}

// WithFanoutOptions sets the child workflow fanout options.
func WithFanoutOptions(fanoutOpts FanoutOptions) Option {
	return func(opts *Options) {
		if fanoutOpts.ActivityOptions == (workflow.ActivityOptions{}) {
			panic("ActivityOptions must be set when using fanout")
		}
		if fanoutOpts.MaxConcurrentChildren <= 0 {
			fanoutOpts.MaxConcurrentChildren = DefaultMaxConcurrentChildren
		}
		opts.FanoutOptions = &fanoutOpts
	}
}
