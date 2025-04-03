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
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
)

// Workflow workflow decider
func Workflow(ctx workflow.Context, name string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("helloworld workflow started")
	var helloworldResult string
	v := workflow.GetVersion(ctx, "test-change", workflow.DefaultVersion, 1)
	if v == workflow.DefaultVersion {
		return errors.New("no default-version history")
	} else {
		err := workflow.ExecuteActivity(ctx, helloworldActivity, name).Get(ctx, &helloworldResult)
		if err != nil {
			logger.Error("First activity failed.", zap.Error(err))
			return err
		}
		err = workflow.ExecuteActivity(ctx, helloworldActivity, name).Get(ctx, &helloworldResult)
		if err != nil {
			logger.Error("Second activity failed.", zap.Error(err))
			return err
		}
	}

	err := workflow.ExecuteActivity(ctx, helloworldActivity, name).Get(ctx, &helloworldResult)
	if err != nil {
		logger.Error("Third activity failed.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Result", helloworldResult))

	return nil
}

// Workflow2 workflow decider
func Workflow2(ctx workflow.Context, name string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("helloworld workflow started")
	var helloworldResult string

	workflow.GetVersion(ctx, "test-change", workflow.DefaultVersion, 1)

	err := workflow.UpsertSearchAttributes(ctx, map[string]interface{}{"CustomKeywordField": "testkey"})
	if err != nil {
		logger.Error("upsert failed", zap.Error(err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, helloworldActivity, name).Get(ctx, &helloworldResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Result", helloworldResult))

	return nil
}

func helloworldActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("helloworld activity started")
	return "Hello " + name + "!", nil
}

func childWorkflowBug(ctx workflow.Context) error {
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ExecutionStartToCloseTimeout: 30 * time.Second,
		TaskStartToCloseTimeout:      30 * time.Second,
		WorkflowIDReusePolicy:        client.WorkflowIDReusePolicyTerminateIfRunning, // not relevant, just convenient
		Bugports: workflow.Bugports{
			// child_bug.json records a history against this workflow where the child workflow IS executed,
			// despite the canceled context.
			//
			// this bug has been fixed, and this flag un-does the fix.
			// if the flag is removed, just delete this test + that history.
			StartChildWorkflowsOnCanceledContext: true,
		},
	})
	ctx, cancel := workflow.WithCancel(ctx)
	cancel()
	return workflow.ExecuteChildWorkflow(ctx, "child").Get(ctx, nil)
}

func childWorkflow(ctx workflow.Context) error {
	_ = workflow.Sleep(ctx, 10*time.Second)
	return nil
}
