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
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.uber.org/cadence/workflow"
)

/**
 * This sample workflow executes multiple branches in parallel using workflow.Go() method.
 */

// sampleParallelWorkflow workflow decider
func sampleParallelWorkflow(ctx workflow.Context) error {
	waitChannel := workflow.NewChannel(ctx)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch1.1").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		err = workflow.ExecuteActivity(ctx, sampleActivity, "branch1.2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	// wait for both of the coroutinue to complete.
	var errMsg string
	for i := 0; i != 2; i++ {
		waitChannel.Receive(ctx, &errMsg)
		if errMsg != "" {
			err := errors.New(errMsg)
			logger.Error("Coroutine failed", zap.Error(err))
			return err
		}
	}

	logger.Info("Workflow completed.")
	return nil
}

// Removing one branch from coroutine 1 made no difference. Here we remove one branch ie branch1.2from the coroutine 1. `
func sampleParallelWorkflow2(ctx workflow.Context) error {
	waitChannel := workflow.NewChannel(ctx)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch1.1").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	// wait for both of the coroutinue to complete.
	var errMsg string
	for i := 0; i != 2; i++ {
		waitChannel.Receive(ctx, &errMsg)
		if errMsg != "" {
			err := errors.New(errMsg)
			logger.Error("Coroutine failed", zap.Error(err))
			return err
		}
	}

	logger.Info("Workflow completed.")
	return nil
}

func sampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}
