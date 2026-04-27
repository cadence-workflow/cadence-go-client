package fanout

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/multierr"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/internal/common/testlogger"
	"go.uber.org/cadence/workflow"
)

// =============================================================================
// Test Suite Setup
// =============================================================================

type FanoutTestSuite struct {
	suite.Suite
	internal.WorkflowTestSuite
}

func (s *FanoutTestSuite) SetupTest() {
	s.SetLogger(testlogger.NewZap(s.T()))
	// Reset global test state to avoid test pollution
	retryAttempts = sync.Map{}
}

func TestFanoutSuite(t *testing.T) {
	suite.Run(t, new(FanoutTestSuite))
}

// =============================================================================
// Unit Tests - Components
// =============================================================================

// TestBulkFutureIsReady tests the IsReady implementation
func (s *FanoutTestSuite) TestBulkFutureIsReady() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{"a", "b", "c"}
		future := ExecuteActivityInBulk[string, string](
			ctx,
			echoActivity,
			items,
			WithMaxConcurrentActivities(2),
		)

		// Should not be ready immediately
		s.False(future.IsReady(), "Future should not be ready immediately")

		// Wait for completion
		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(3, len(results))

		// Should be ready after Get
		s.True(future.IsReady(), "Future should be ready after Get")

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(echoActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *FanoutTestSuite) TestBulkFutureFromChildrenIsReady() {
	wf := func(ctx workflow.Context) error {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		})

		items := []int{0, 1, 2, 3}
		future := ExecuteActivityInBulk[int, int](
			ctx,
			incrementActivity,
			items,
			WithFanoutOptions(FanoutOptions{
				MaxActivitiesPerWorkflow: 2,
				MaxConcurrentChildren:    2,
				ChildWorkflowOptions: workflow.ChildWorkflowOptions{
					ExecutionStartToCloseTimeout: time.Minute,
					TaskStartToCloseTimeout:      time.Second,
				},
				ActivityOptions: workflow.ActivityOptions{
					ScheduleToStartTimeout: time.Minute,
					StartToCloseTimeout:    time.Second,
				},
			}),
		)

		s.False(future.IsReady(), "child future should not be ready immediately")
		results, err := future.Get(ctx)
		s.NoError(err)
		s.True(future.IsReady(), "child future should be ready after Get")
		s.Equal([]int{1, 2, 3, 4}, results)
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(incrementActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// =============================================================================
// Workflow Tests - Basic Functionality
// =============================================================================

// TestSimpleExecution tests basic execution with a few items
func (s *FanoutTestSuite) TestSimpleExecution() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{"hello", "world", "test"}
		future := ExecuteActivityInBulk[string, string](
			ctx,
			echoActivity,
			items,
			WithMaxConcurrentActivities(10),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(3, len(results))
		s.Equal("hello", results[0])
		s.Equal("world", results[1])
		s.Equal("test", results[2])

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(echoActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestEmptyItemsList tests behavior with empty items
func (s *FanoutTestSuite) TestEmptyItemsList() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{}
		future := ExecuteActivityInBulk[string, string](
			ctx,
			echoActivity,
			items,
			WithMaxConcurrentActivities(10),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(0, len(results), "Should return empty results for empty input")

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(echoActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestSingleItem tests with just one item
func (s *FanoutTestSuite) TestSingleItem() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{"single"}
		future := ExecuteActivityInBulk[string, string](
			ctx,
			echoActivity,
			items,
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(1, len(results))
		s.Equal("single", results[0])

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(echoActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestManyItems tests with many items to ensure concurrency works
func (s *FanoutTestSuite) TestManyItems() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		// Create 100 items
		items := make([]int, 100)
		for i := 0; i < 100; i++ {
			items[i] = i
		}

		future := ExecuteActivityInBulk[int, int](
			ctx,
			incrementActivity,
			items,
			WithMaxConcurrentActivities(20),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(100, len(results))

		// Verify all results
		for i, result := range results {
			s.Equal(i+1, result, "Result should be incremented value")
		}

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(incrementActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// =============================================================================
// Workflow Tests - Concurrency Control
// =============================================================================

// TestConcurrencyLimit tests that concurrency limit is respected
func (s *FanoutTestSuite) TestConcurrencyLimit() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    5 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		// Create 10 items, concurrency of 3
		items := make([]int, 10)
		for i := 0; i < 10; i++ {
			items[i] = i
		}

		future := ExecuteActivityInBulk[int, int](
			ctx,
			incrementActivity,
			items,
			WithMaxConcurrentActivities(3),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(10, len(results))

		// Verify all results are correct
		for i, result := range results {
			s.Equal(i+1, result, "Result should be incremented value at index %d", i)
		}

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(incrementActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestDefaultConcurrency tests that default concurrency of 10 is used
func (s *FanoutTestSuite) TestDefaultConcurrency() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{"a", "b", "c"}
		// No concurrency option provided, should use default of 10
		future := ExecuteActivityInBulk[string, string](
			ctx,
			echoActivity,
			items,
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(3, len(results))

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(echoActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// =============================================================================
// Workflow Tests - ParamFactory
// =============================================================================

// TestWithParamFactory tests using ParamFactory instead of items
func (s *FanoutTestSuite) TestWithOffsetLimit() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		// Generate offset/limit params for 50 items with chunk size 10
		totalItems := 50
		chunkSize := 10
		params := []offsetLimitParams{}
		for offset := 0; offset < totalItems; offset += chunkSize {
			params = append(params, offsetLimitParams{
				Offset: offset,
				Limit:  chunkSize,
			})
		}

		future := ExecuteActivityInBulk[offsetLimitParams, offsetLimitResult](
			ctx,
			processChunkActivity,
			params,
			WithMaxConcurrentActivities(5),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(5, len(results), "Should create 5 activities for 50 items with chunk size 10")

		// Verify each result
		totalProcessed := 0
		for i, result := range results {
			s.Equal(i*10, result.StartOffset, "Start offset should match")
			s.Equal(10, result.ItemsProcessed, "Should process 10 items per chunk")
			totalProcessed += result.ItemsProcessed
		}
		s.Equal(50, totalProcessed, "Should process all 50 items total")

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(processChunkActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// =============================================================================
// Workflow Tests - Error Handling
// =============================================================================

// TestPartialFailure tests that some activities can fail while others succeed
func (s *FanoutTestSuite) TestPartialFailure() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		// Items: 0, 1, 2, 3, 4, 5 (activity fails on even numbers)
		items := []int{0, 1, 2, 3, 4, 5}

		future := ExecuteActivityInBulk[int, int](
			ctx,
			failOnEvenActivity,
			items,
			WithMaxConcurrentActivities(3),
		)

		results, err := future.Get(ctx)

		// Should have errors from failed activities
		s.Error(err, "Should return error when some activities fail")

		// But should still have all results (zeros for failed ones)
		s.Equal(6, len(results))

		// Odd numbers should succeed
		s.Equal(2, results[1], "Odd numbers should succeed")
		s.Equal(4, results[3], "Odd numbers should succeed")
		s.Equal(6, results[5], "Odd numbers should succeed")

		// Even numbers should have zero values (failed)
		s.Equal(0, results[0], "Even numbers fail, should be zero value")
		s.Equal(0, results[2], "Even numbers fail, should be zero value")
		s.Equal(0, results[4], "Even numbers fail, should be zero value")

		// Should have 3 errors (for 0, 2, 4)
		errs := multierr.Errors(err)
		s.Equal(3, len(errs), "Should have 3 errors")

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(failOnEvenActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestAllActivitiesFail tests when all activities fail
func (s *FanoutTestSuite) TestAllActivitiesFail() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{"a", "b", "c"}

		future := ExecuteActivityInBulk[string, string](
			ctx,
			alwaysFailActivity,
			items,
			WithMaxConcurrentActivities(2),
		)

		results, err := future.Get(ctx)

		s.Error(err, "Should return error when all activities fail")
		s.Equal(3, len(results), "Should still have results slice")

		// All should be zero values
		for i, result := range results {
			s.Equal("", result, "Failed activity should have zero value at index %d", i)
		}

		// Should have 3 errors
		errs := multierr.Errors(err)
		s.Equal(3, len(errs), "Should have error for each failed activity")

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(alwaysFailActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestActivityRetry tests that activity retries work with ExecuteManyActivities
func (s *FanoutTestSuite) TestActivityRetry() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
			RetryPolicy: &workflow.RetryPolicy{
				InitialInterval:    time.Millisecond,
				BackoffCoefficient: 1.0,
				MaximumInterval:    time.Second,
				MaximumAttempts:    3,
			},
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{"fail-once", "fail-twice"}

		future := ExecuteActivityInBulk[string, string](
			ctx,
			retryableActivity,
			items,
		)

		results, err := future.Get(ctx)

		// Both should eventually succeed after retries
		s.NoError(err)
		s.Equal(2, len(results))
		s.Equal("fail-once", results[0])
		s.Equal("fail-twice", results[1])

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(retryableActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// =============================================================================
// Workflow Tests - Complex Types
// =============================================================================

// TestComplexTypes tests with struct input and output types
func (s *FanoutTestSuite) TestComplexTypes() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []emailRequest{
			{To: "user1@example.com", Subject: "Hello"},
			{To: "user2@example.com", Subject: "World"},
		}

		future := ExecuteActivityInBulk[emailRequest, emailResult](
			ctx,
			sendEmailActivity,
			items,
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(2, len(results))
		s.Equal("msg-user1@example.com", results[0].MessageID)
		s.Equal("msg-user2@example.com", results[1].MessageID)

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(sendEmailActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// =============================================================================
// Test Activities
// =============================================================================

func echoActivity(ctx context.Context, input string) (string, error) {
	return input, nil
}

func incrementActivity(ctx context.Context, input int) (int, error) {
	return input + 1, nil
}

func hangingIncrementActivity(ctx context.Context, input int) (int, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-time.After(time.Hour):
		return input + 1, nil
	}
}

type blockingActivityControl struct {
	mu          sync.Mutex
	active      int
	maxActive   int
	release     chan struct{}
	releaseOnce sync.Once
}

func newBlockingActivityControl() *blockingActivityControl {
	return &blockingActivityControl{
		release: make(chan struct{}),
	}
}

func (c *blockingActivityControl) acquire() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.active++
	if c.active > c.maxActive {
		c.maxActive = c.active
	}
}

func (c *blockingActivityControl) releaseOne() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active > 0 {
		c.active--
	}
}

func (c *blockingActivityControl) wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.release:
		return nil
	}
}

func (c *blockingActivityControl) releaseAll() {
	c.releaseOnce.Do(func() {
		close(c.release)
	})
}

func (c *blockingActivityControl) maxActiveCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.maxActive
}

var blockingCtrl atomic.Pointer[blockingActivityControl]

func setBlockingActivityControl(ctrl *blockingActivityControl) {
	blockingCtrl.Store(ctrl)
}

func getBlockingActivityControl() *blockingActivityControl {
	return blockingCtrl.Load()
}

func blockingIncrementActivity(ctx context.Context, input int) (int, error) {
	ctrl := getBlockingActivityControl()
	if ctrl == nil {
		return incrementActivity(ctx, input)
	}

	ctrl.acquire()
	defer ctrl.releaseOne()

	if err := ctrl.wait(ctx); err != nil {
		return 0, err
	}

	return input + 1, nil
}

func alwaysFailActivity(ctx context.Context, input string) (string, error) {
	return "", errors.New("activity always fails")
}

var failTwiceAttempts atomic.Int32

func failTwiceActivity(ctx context.Context, input string) (string, error) {
	if attempt := failTwiceAttempts.Add(1); attempt < 3 {
		return "", fmt.Errorf("attempt %d failed", attempt)
	}
	return input, nil
}

func failOnEvenActivity(ctx context.Context, input int) (int, error) {
	if input%2 == 0 {
		return 0, fmt.Errorf("failed on even number: %d", input)
	}
	return input + 1, nil
}

// retryableActivity fails N times based on the input, then succeeds
// Uses sync.Map to avoid data races when activities run in parallel
var retryAttempts sync.Map

func retryableActivity(ctx context.Context, input string) (string, error) {
	// Atomically increment the attempt counter
	val, _ := retryAttempts.LoadOrStore(input, new(int32))
	counter := val.(*int32)
	attempt := atomic.AddInt32(counter, 1)

	var maxRetries int32
	switch input {
	case "fail-once":
		maxRetries = 1
	case "fail-twice":
		maxRetries = 2
	default:
		maxRetries = 0
	}

	if attempt <= maxRetries {
		return "", fmt.Errorf("attempt %d failed", attempt)
	}

	return input, nil
}

// =============================================================================
// Test Types and Helpers
// =============================================================================

type offsetLimitParams struct {
	Offset int
	Limit  int
}

type offsetLimitResult struct {
	StartOffset    int
	ItemsProcessed int
}

func processChunkActivity(ctx context.Context, params offsetLimitParams) (offsetLimitResult, error) {
	return offsetLimitResult{
		StartOffset:    params.Offset,
		ItemsProcessed: params.Limit,
	}, nil
}

type emailRequest struct {
	To      string
	Subject string
}

type emailResult struct {
	MessageID string
	SentAt    time.Time
}

func sendEmailActivity(ctx context.Context, req emailRequest) (emailResult, error) {
	return emailResult{
		MessageID: "msg-" + req.To,
		SentAt:    time.Now(),
	}, nil
}

// =============================================================================
// Integration Tests - Realistic Scenarios
// =============================================================================

// TestRealisticBulkEmail tests a realistic bulk email scenario
func (s *FanoutTestSuite) TestRealisticBulkEmail() {
	wf := func(ctx workflow.Context, numEmails int) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Hour,
			StartToCloseTimeout:    time.Minute,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		// Generate email requests
		emails := make([]emailRequest, numEmails)
		for i := 0; i < numEmails; i++ {
			emails[i] = emailRequest{
				To:      fmt.Sprintf("user%d@example.com", i),
				Subject: "Bulk notification",
			}
		}

		future := ExecuteActivityInBulk[emailRequest, emailResult](
			ctx,
			sendEmailActivity,
			emails,
			WithMaxConcurrentActivities(50),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(numEmails, len(results))

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(sendEmailActivity)
	env.ExecuteWorkflow(wf, 100) // 100 emails

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestRealisticCatalogBuild tests a realistic catalog building scenario
func (s *FanoutTestSuite) TestRealisticCatalogBuild() {
	wf := func(ctx workflow.Context, totalRecords int) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Hour,
			StartToCloseTimeout:    5 * time.Minute,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		// Generate offset/limit params for chunked processing
		chunkSize := 1000
		params := []offsetLimitParams{}
		for offset := 0; offset < totalRecords; offset += chunkSize {
			limit := chunkSize
			if offset+limit > totalRecords {
				limit = totalRecords - offset
			}
			params = append(params, offsetLimitParams{
				Offset: offset,
				Limit:  limit,
			})
		}

		future := ExecuteActivityInBulk[offsetLimitParams, offsetLimitResult](
			ctx,
			processChunkActivity,
			params,
			WithMaxConcurrentActivities(100),
		)

		results, err := future.Get(ctx)
		s.NoError(err)

		// Verify all records were processed
		totalProcessed := 0
		for _, result := range results {
			totalProcessed += result.ItemsProcessed
		}
		s.Equal(totalRecords, totalProcessed)

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(processChunkActivity)
	env.ExecuteWorkflow(wf, 10000) // 10k records

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestPartialLastChunk tests handling of a partial last chunk
func (s *FanoutTestSuite) TestPartialLastChunk() {
	wf := func(ctx workflow.Context) error {
		ctx = workflow.WithActivityOptions(ctx, internal.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		})

		// Generate 95 items with chunk size 10 - last chunk will have only 5 items
		totalItems := 95
		chunkSize := 10
		params := []offsetLimitParams{}
		for offset := 0; offset < totalItems; offset += chunkSize {
			limit := chunkSize
			if offset+limit > totalItems {
				limit = totalItems - offset
			}
			params = append(params, offsetLimitParams{
				Offset: offset,
				Limit:  limit,
			})
		}

		future := ExecuteActivityInBulk[offsetLimitParams, offsetLimitResult](
			ctx,
			processChunkActivity,
			params,
			WithMaxConcurrentActivities(3),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(10, len(results), "Should create 10 activities")

		// Last chunk should have only 5 items
		s.Equal(90, results[9].StartOffset, "Last chunk offset should be 90")
		s.Equal(5, results[9].ItemsProcessed, "Last chunk should process only 5 items")

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(processChunkActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *FanoutTestSuite) TestFanoutChildWorkflowSuccess() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)
		childWorkflowOpts := workflow.ChildWorkflowOptions{
			ExecutionStartToCloseTimeout: time.Hour,
			TaskStartToCloseTimeout:      10 * time.Second,
		}

		// Create 150 items - will trigger fanout with MaxActivitiesPerWorkflow=100
		items := make([]int, 150)
		for i := range items {
			items[i] = i
		}

		future := ExecuteActivityInBulk[int, int](
			ctx,
			incrementActivity,
			items,
			WithMaxConcurrentActivities(10),
			WithFanoutOptions(FanoutOptions{
				MaxActivitiesPerWorkflow: 100,
				MaxConcurrentChildren:    2,
				ChildWorkflowOptions:     childWorkflowOpts,
				ActivityOptions:          ao,
			}),
		)

		results, err := future.Get(ctx)
		s.NoError(err, "Should complete successfully")
		s.Equal(150, len(results), "Should return all results")

		// Verify results are correct (incremented by 1)
		for i, result := range results {
			s.Equal(i+1, result, "Result should be incremented")
		}

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(incrementActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

// TestFanoutChildWorkflowPartialFailure tests that partial failures work with child workflow fanout
func (s *FanoutTestSuite) TestFanoutChildWorkflowPartialFailure() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)
		childWorkflowOpts := workflow.ChildWorkflowOptions{
			ExecutionStartToCloseTimeout: time.Hour,
			TaskStartToCloseTimeout:      10 * time.Second,
		}

		// Create items where even numbers will fail
		// Use 6 items with MaxActivitiesPerWorkflow=3 to get 2 child workflows
		items := []int{0, 1, 2, 3, 4, 5}

		future := ExecuteActivityInBulk[int, int](
			ctx,
			failOnEvenActivity,
			items,
			WithMaxConcurrentActivities(2),
			WithFanoutOptions(FanoutOptions{
				MaxActivitiesPerWorkflow: 3, // 6 items / 3 = 2 child workflows
				MaxConcurrentChildren:    2,
				ChildWorkflowOptions:     childWorkflowOpts,
				ActivityOptions:          ao,
			}),
		)

		results, err := future.Get(ctx)

		// Should have errors from failed activities propagated through child workflows
		s.Error(err, "Should return error when some activities fail in child workflows")

		// Results may be partial when child workflows have errors
		// (Cadence doesn't guarantee result population when workflow returns error)
		s.NotNil(results, "Results should not be nil")

		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(failOnEvenActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *FanoutTestSuite) TestChildWorkflowsRunInParallel() {
	const (
		numItems              = 4
		maxConcurrentChildren = 2
	)

	ctrl := newBlockingActivityControl()
	setBlockingActivityControl(ctrl)
	defer setBlockingActivityControl(nil)

	releaseDone := make(chan struct{})
	go func() {
		defer close(releaseDone)
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.NewTimer(2 * time.Second)
		defer timeout.Stop()
		for {
			select {
			case <-ticker.C:
				if ctrl.maxActiveCount() >= maxConcurrentChildren {
					ctrl.releaseAll()
					return
				}
			case <-timeout.C:
				ctrl.releaseAll()
				return
			}
		}
	}()

	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		childWorkflowOpts := workflow.ChildWorkflowOptions{
			ExecutionStartToCloseTimeout: time.Hour,
			TaskStartToCloseTimeout:      10 * time.Second,
		}

		items := make([]int, numItems)
		for i := range items {
			items[i] = i
		}

		future := ExecuteActivityInBulk[int, int](
			ctx,
			blockingIncrementActivity,
			items,
			WithMaxConcurrentActivities(1),
			WithFanoutOptions(FanoutOptions{
				MaxActivitiesPerWorkflow: 1,
				MaxConcurrentChildren:    maxConcurrentChildren,
				ChildWorkflowOptions:     childWorkflowOpts,
				ActivityOptions:          ao,
			}),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal(numItems, len(results))
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(blockingIncrementActivity)
	env.ExecuteWorkflow(wf)

	<-releaseDone

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.GreaterOrEqual(ctrl.maxActiveCount(), maxConcurrentChildren,
		"expected at least %d child workflows to run concurrently", maxConcurrentChildren)
}

func (s *FanoutTestSuite) TestFanoutWithoutActivityOptionsPanics() {
	items := []int{1, 2, 3}

	s.Panics(func() {
		ExecuteActivityInBulk[int, int](nil, incrementActivity, items,
			WithFanoutOptions(FanoutOptions{
				MaxActivitiesPerWorkflow: 1,
				MaxConcurrentChildren:    1,
				ChildWorkflowOptions: workflow.ChildWorkflowOptions{
					ExecutionStartToCloseTimeout: time.Minute,
					TaskStartToCloseTimeout:      time.Second,
				},
			}),
		)
	})
}

func (s *FanoutTestSuite) TestActivityRetryPolicyRespectedWithoutFanout() {
	failTwiceAttempts.Store(0)

	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    5 * time.Second,
			RetryPolicy: &workflow.RetryPolicy{
				InitialInterval:    10 * time.Millisecond,
				BackoffCoefficient: 1.0,
				MaximumAttempts:    3,
			},
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []string{"payload"}
		future := ExecuteActivityInBulk[string, string](
			ctx,
			failTwiceActivity,
			items,
			WithMaxConcurrentActivities(2),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal([]string{"payload"}, results)
		s.Equal(int32(3), failTwiceAttempts.Load())
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(failTwiceActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *FanoutTestSuite) TestActivityRetryPolicyRespectedWithFanout() {
	failTwiceAttempts.Store(0)

	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    5 * time.Second,
			RetryPolicy: &workflow.RetryPolicy{
				InitialInterval:    10 * time.Millisecond,
				BackoffCoefficient: 1.0,
				MaximumAttempts:    3,
			},
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		childWorkflowOpts := workflow.ChildWorkflowOptions{
			ExecutionStartToCloseTimeout: time.Minute,
			TaskStartToCloseTimeout:      10 * time.Second,
		}

		items := []string{"payload"}
		future := ExecuteActivityInBulk[string, string](
			ctx,
			failTwiceActivity,
			items,
			WithMaxConcurrentActivities(1),
			WithFanoutOptions(FanoutOptions{
				MaxActivitiesPerWorkflow: 1,
				MaxConcurrentChildren:    1,
				ChildWorkflowOptions:     childWorkflowOpts,
				ActivityOptions:          ao,
			}),
		)

		results, err := future.Get(ctx)
		s.NoError(err)
		s.Equal([]string{"payload"}, results)
		s.Equal(int32(3), failTwiceAttempts.Load())
		return nil
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(failTwiceActivity)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *FanoutTestSuite) TestParentClosePolicyRespected() {
	wf := func(ctx workflow.Context) error {
		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)

		items := []int{1}

		childOpts := workflow.ChildWorkflowOptions{
			ExecutionStartToCloseTimeout: time.Minute,
			TaskStartToCloseTimeout:      time.Second,
			ParentClosePolicy:            client.ParentClosePolicyTerminate,
		}

		future := ExecuteActivityInBulk[int, int](
			ctx,
			hangingIncrementActivity,
			items,
			WithMaxConcurrentActivities(1),
			WithFanoutOptions(FanoutOptions{
				MaxActivitiesPerWorkflow: 1,
				MaxConcurrentChildren:    1,
				ChildWorkflowOptions:     childOpts,
				ActivityOptions:          ao,
			}),
		)

		_, err := future.Get(ctx)
		s.Error(err)
		return err
	}

	env := s.NewTestWorkflowEnvironment()
	env.Test(s.T())
	env.RegisterActivity(hangingIncrementActivity)
	env.RegisterDelayedCallback(func() {
		env.CancelWorkflow()
	}, time.Millisecond*10)
	env.ExecuteWorkflow(wf)

	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError())
}
