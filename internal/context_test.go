// Copyright (c) 2017-2021 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package internal

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestContextChildParentCancelRace(t *testing.T) {
	/*
		Testing previous race happened while child and parent cancelling at the same time
		While child is trying to remove itself from the parent, parent tries to iterate
		its children and cancel them at the same time.
	*/
	env := newTestWorkflowEnv(t)

	wf := func(ctx Context) error {
		parentCtx, parentCancel := WithCancel(ctx)
		defer parentCancel()

		type cancelerContext struct {
			ctx      Context
			canceler func()
		}

		children := []cancelerContext{}
		numChildren := 100

		for i := 0; i < numChildren; i++ {
			c, canceler := WithCancel(parentCtx)
			children = append(children, cancelerContext{
				ctx:      c,
				canceler: canceler,
			})
		}

		for i := 0; i < numChildren; i++ {
			go children[i].canceler()
			if i == numChildren/2 {
				go parentCancel()
			}
		}

		return nil
	}
	env.RegisterWorkflow(wf)
	env.ExecuteWorkflow(wf)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func TestContextConcurrentCancelRace(t *testing.T) {
	/*
		A race condition existed due to concurrently ending goroutines on shutdown (i.e. closing their chan without waiting
		on them to finish shutdown), which executed... quite a lot of non-concurrency-safe code in a concurrent way.  All
		decision-sensitive code is assumed to be run strictly sequentially.

		Context cancellation was one identified by a customer, and it's fairly easy to test.
		In principle this must be safe to do - contexts are supposed to be concurrency-safe.  Even if ours are not actually
		safe (for valid reasons), our execution model needs to ensure they *act* like it's safe.

		This intentionally avoids `newTestWorkflowEnv` because the dangling goroutines
		may shut down after the test finishes -> produce a log, which fails on t.Log races.
		Ideally the test suite would stop goroutines too, but that isn't reliable currently either.
		This can be changed if we switch to the server's race-safe zaptest wrapper.
	*/
	env := (&WorkflowTestSuite{}).NewTestWorkflowEnvironment()
	wf := func(ctx Context) error {
		ctx, cancel := WithCancel(ctx)
		racyCancel := func(ctx Context) {
			defer cancel() // defer is necessary as Sleep will never return due to Goexit
			_ = Sleep(ctx, time.Hour)
		}
		// start a handful to increase odds of a race being detected
		for i := 0; i < 10; i++ {
			Go(ctx, racyCancel)
		}

		_ = Sleep(ctx, time.Minute) // die early
		return nil
	}
	env.RegisterWorkflow(wf)
	env.ExecuteWorkflow(wf)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func TestContextAddChildCancelParentRace(t *testing.T) {
	/*
		It's apparently also possible to race on adding children while propagating the cancel to children.

		This intentionally avoids `newTestWorkflowEnv` because the dangling goroutines
		may shut down after the test finishes -> produce a log, which fails on t.Log races.
		Ideally the test suite would stop goroutines too, but that isn't reliable currently either.
		This can be changed if we switch to the server's race-safe zaptest wrapper.
	*/
	env := (&WorkflowTestSuite{}).NewTestWorkflowEnvironment()
	wf := func(ctx Context) error {
		ctx, cancel := WithCancel(ctx)
		racyCancel := func(ctx Context) {
			defer cancel() // defer is necessary as Sleep will never return due to Goexit
			defer func() {
				_, ccancel := WithCancel(ctx)
				cancel()
				ccancel()
			}()
			_ = Sleep(ctx, time.Hour)
		}
		// start a handful to increase odds of a race being detected
		for i := 0; i < 10; i++ {
			Go(ctx, racyCancel)
		}

		_ = Sleep(ctx, time.Minute) // die early
		return nil
	}
	env.RegisterWorkflow(wf)
	env.ExecuteWorkflow(wf)
	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func TestContextCancellationOrderDeterminism(t *testing.T) {
	/*
		Previously, child-contexts were stored in a map, preventing deterministic order when propagating cancellation.
		The order of branches being selected in this test was random, both for the first event and in following ones.

		In principle this should be fine, but it's possible for the effects of cancellation to trigger a selector's
		future-done callback, which currently records the *real-time*-first event as the branch to unblock, rather than
		doing something more safe by design (e.g. choosing based on state when the selector's goroutine is unblocked).

		Unfortunately, we cannot change the selector's behavior without introducing non-backwards-compatible changes to
		currently-working workflows.

		So the workaround for now is to maintain child-context order, so they are canceled in a consistent order.
		As this order was not controlled before, and Go does a pretty good job at randomizing map iteration order,
		converting non-determinism to determinism should be strictly no worse for backwards compatibility, and it
		fixes the issue for future executions.
	*/
	check := func(t *testing.T, separateStart, separateSelect bool) {
		env := newTestWorkflowEnv(t)
		act := func(ctx context.Context) error {
			return nil // will be mocked
		}
		wf := func(ctx Context) ([]int, error) {
			ctx, cancel := WithCancel(ctx)
			Go(ctx, func(ctx Context) {
				_ = Sleep(ctx, time.Minute)
				cancel()
			})

			// start some activities, which will not complete before the timeout cancels them
			ctx = WithActivityOptions(ctx, ActivityOptions{
				TaskList:               "",
				ScheduleToCloseTimeout: time.Hour,
				ScheduleToStartTimeout: time.Hour,
				StartToCloseTimeout:    time.Hour,
			})
			s := NewSelector(ctx)
			var result []int
			for i := 0; i < 10; i++ {
				i := i
				// need a child context, a future alone is not enough as it does not become a child
				cctx, ccancel := WithCancel(ctx)

				s.AddFuture(ExecuteActivity(cctx, act), func(f Future) {
					ccancel() // TODO: is this necessary to prevent leaks?  if it is, how can we make it not?
					err := f.Get(ctx, nil)
					if err == nil || !IsCanceledError(err) {
						// fail the test, this should not happen - activities must be canceled or it's not valid.
						t.Errorf("activity completion or failure for some reason other than cancel: %v", err)
					}
					result = append(result, i)
				})

				if separateStart {
					// yield so they are submitted one at a time, in case that matters
					_ = Sleep(ctx, time.Second)
				}
			}
			for i := 0; i < 10; i++ {
				if separateSelect {
					// yield so they are selected one at a time, in case that matters
					_ = Sleep(ctx, time.Second)
				}
				s.Select(ctx)
			}

			return result, nil
		}
		env.RegisterWorkflow(wf)
		env.RegisterActivity(act)

		// activities must not complete in time
		env.OnActivity(act, mock.Anything).After(5 * time.Minute).Return(nil)

		env.ExecuteWorkflow(wf)
		require.NoError(t, env.GetWorkflowError())
		var result []int
		require.NoError(t, env.GetWorkflowResult(&result))
		require.NotEmpty(t, result)
		assert.Equal(t, 0, result[0], "first activity to be created should be the first one canceled")
		assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, result[1:], "other activities should finish in a consistent (but undefined) order")
	}

	type variant struct {
		name           string
		separateStart  bool
		separateSelect bool
	}
	// all variants expose this behavior, but being a bit more exhaustive in the face
	// of decision-scheduling differences seems good.
	for _, test := range []variant{
		{"many in one decision", false, false},
		{"many started at once, selected slowly", false, true},
		{"started slowly, selected quickly", true, false},
		{"started and selected slowly", true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			check(t, test.separateStart, test.separateSelect)
		})
	}
}

func BenchmarkSliceMaintenance(b *testing.B) {
	// all essentially identical
	b.Run("append", func(b *testing.B) {
		data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		for i := 0; i < b.N; i++ {
			data = append(data[:5], data[6:]...)
			data = append(data, i) // keep the slice the same size for all iterations
		}
	})
	b.Run("copy", func(b *testing.B) {
		data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		for i := 0; i < b.N; i++ {
			copy(data[5:], data[6:])
			data = data[:9]        // trim to actual size, as the last value is now duplicated.  capacity is still 10.
			data = append(data, i) // keep the slice the same size for all iterations
		}
	})
	b.Run("copy explicit capacity", func(b *testing.B) {
		data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
		for i := 0; i < b.N; i++ {
			copy(data[5:], data[6:])
			data = data[:9:10]     // trim to actual size, as the last value is now duplicated.  explicitly reserve 10 cap.
			data = append(data, i) // keep the slice the same size for all iterations
		}
	})
}
