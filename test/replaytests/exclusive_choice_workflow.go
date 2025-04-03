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

	"go.uber.org/zap"

	"go.uber.org/cadence/workflow"
)

/**
 * This sample workflow Execute one of many code paths based on the result of an activity.
 */

const (
	orderChoiceApple  = "apple"
	orderChoiceBanana = "banana"
	orderChoiceCherry = "cherry"
)

// exclusiveChoiceWorkflow Workflow Decider. This workflow executes Cherry order.
func exclusiveChoiceWorkflow(ctx workflow.Context) error {
	// Get order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var orderChoice string
	err := workflow.ExecuteActivity(ctx, getOrderActivity).Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := workflow.GetLogger(ctx)

	// choose next activity based on order result
	switch orderChoice {
	case orderChoiceBanana:
		workflow.ExecuteActivity(ctx, orderBananaActivity, orderChoice)
	case orderChoiceCherry:
		workflow.ExecuteActivity(ctx, orderCherryActivity, orderChoice)
	default:
		logger.Error("Unexpected order", zap.String("Choice", orderChoice))
	}

	logger.Info("Workflow completed.")
	return nil
}

// This workflow explicitly executes Apple Activity received from the getorderActivity.
func exclusiveChoiceWorkflow2(ctx workflow.Context) error {
	// Get order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var orderChoice string
	err := workflow.ExecuteActivity(ctx, getAppleOrderActivity).Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := workflow.GetLogger(ctx)

	// choose next activity based on order result. It's apple in this case.
	switch orderChoice {
	case orderChoiceApple:
		workflow.ExecuteActivity(ctx, orderAppleActivity, orderChoice)
	default:
		logger.Error("Unexpected order", zap.String("Choice", orderChoice))
	}

	logger.Info("Workflow completed.")
	return nil
}

func getOrderActivity() (string, error) {
	fmt.Printf("Order is for Cherry")
	return "cherry", nil
}

func getAppleOrderActivity() (string, error) {
	fmt.Printf("Order is for Apple")
	return "apple", nil
}

func orderAppleActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func orderBananaActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func orderCherryActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}
