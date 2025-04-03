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
	"time"

	"go.uber.org/zap"

	"go.uber.org/cadence/workflow"
)

// greetingsWorkflowActivity demonstrates the case in replayer where if an activity name is changed in a workflow or if a new activity is introdced in a workflow then the shdower fails.
// There is also no other way to register the activity except for changing the json file so far.
// CDNC-2267 addresses a fix for this.
func greetingsWorkflowActivity(ctx workflow.Context) error {
	// Get Greeting.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	var greetResult int64
	err := workflow.ExecuteActivity(ctx, getGreetingActivitytest).Get(ctx, &greetResult)
	if err != nil {
		logger.Error("Get greeting failed.", zap.Error(err))
		return err
	}

	// Get Name.
	var nameResult string
	err = workflow.ExecuteActivity(ctx, getNameActivity).Get(ctx, &nameResult)
	if err != nil {
		logger.Error("Get name failed.", zap.Error(err))
		return err
	}

	// Say Greeting.
	var sayResult string
	err = workflow.ExecuteActivity(ctx, sayGreetingActivity, greetResult, nameResult).Get(ctx, &sayResult)
	if err != nil {
		logger.Error("Marshalling failed with error.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Result", sayResult))
	return nil
}

// Get Greeting Activity.
func getGreetingActivitytest() (string, error) {
	return "Hello", nil
}
