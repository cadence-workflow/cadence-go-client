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
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
)

type replayerSampleMessage struct {
	Message string
}

// replayerHelloWorldWorkflow is a sample workflow that runs replayerHelloWorldActivity.
// In the previous version it was running the activity twice, sequentially.
// History of a past execution is in sequential.json
// Corresponding unit test covers the scenario that new workflow's history records is subset of previous version's history.
//
//	v1: wf started -> call activity -> call activity -> wf complete
//	v2:  wf started -> call activity -> wf complete
//
// The v2 clearly has determinism issues and should be considered as non-determism error for replay tests.
func replayerHelloWorldWorkflow(ctx workflow.Context, inputMsg *replayerSampleMessage) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	logger := workflow.GetLogger(ctx)
	logger.Info("executing replayerHelloWorldWorkflow")
	ctx = workflow.WithActivityOptions(ctx, ao)

	count := 1
	for i := 0; i < count; i++ {
		var greeting string
		err := workflow.ExecuteActivity(ctx, replayerHelloWorldActivity, inputMsg).Get(ctx, &greeting)
		if err != nil {
			logger.Error("replayerHelloWorldActivity is broken ", zap.Error(err))
			return err
		}

		logger.Sugar().Infof("replayerHelloWorldWorkflow is greeting to you %dth time -> ", i, greeting)
	}

	return nil
}

// replayerHelloWorldActivity takes a sample message input and return a greeting to the caller.
func replayerHelloWorldActivity(ctx context.Context, inputMsg *replayerSampleMessage) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("executing replayerHelloWorldActivity")

	return fmt.Sprintf("Hello, %s!", inputMsg.Message), nil
}
