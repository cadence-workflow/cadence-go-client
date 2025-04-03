// Copyright (c) 2017 Uber Technologies, Inc.
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

package replaytests

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
)

/**
 * This sample workflow executes multiple branches in parallel. The number of branches is controlled by passed in parameter.
 */

// sampleBranchWorkflow workflow decider
func sampleBranchWorkflow(ctx workflow.Context) error {
	var futures []workflow.Future
	// starts activities in parallel
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i <= 3; i++ {
		activityInput := fmt.Sprintf("branch %d of 3", i)
		future := workflow.ExecuteActivity(ctx, sampleActivity, activityInput)
		futures = append(futures, future)
	}

	// wait until all futures are done
	for _, future := range futures {
		if err := future.Get(ctx, nil); err != nil {
			return err
		}
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")

	return nil
}

// SampleBranchWorkflow2 run a workflow with different number of branch.
// If the number of expected branches is changed the replayer should catch it as a non deterministic error.
func sampleBranchWorkflow2(ctx workflow.Context) error {
	var futures []workflow.Future
	// starts activities in parallel
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i <= 2; i++ {
		activityInput := fmt.Sprintf("branch %d of 4", i)
		future := workflow.ExecuteActivity(ctx, sampleActivity, activityInput)
		futures = append(futures, future)
	}

	// wait until all futures are done
	for _, future := range futures {
		if err := future.Get(ctx, nil); err != nil {
			return err
		}
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")

	return nil
}
