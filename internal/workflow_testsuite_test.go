// Copyright (c) 2017-2020 Uber Technologies Inc.
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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetWorkflowTimeout(t *testing.T) {
	t.Parallel()
	env := newTestWorkflowEnv(t)

	executionTimeout := 30 * time.Minute
	require.Equal(t, int32(1), env.impl.workflowInfo.ExecutionStartToCloseTimeoutSeconds)

	result := env.SetWorkflowTimeout(executionTimeout)

	require.Equal(t, env, result)
	require.Equal(t, int32(executionTimeout.Seconds()), result.impl.workflowInfo.ExecutionStartToCloseTimeoutSeconds)
	require.Equal(t, executionTimeout, result.impl.executionTimeout)
}

func TestSetMemoOnStart(t *testing.T) {
	t.Parallel()
	env := newTestWorkflowEnv(t)

	memo := map[string]interface{}{
		"key": make(chan int),
	}
	err := env.SetMemoOnStart(memo)
	require.Error(t, err)

	memo = map[string]interface{}{
		"memoKey": "memo",
	}
	require.Nil(t, env.impl.workflowInfo.Memo)
	err = env.SetMemoOnStart(memo)
	require.NoError(t, err)
	require.NotNil(t, env.impl.workflowInfo.Memo)
}

func TestSetSearchAttributesOnStart(t *testing.T) {
	t.Parallel()
	env := newTestWorkflowEnv(t)

	invalidSearchAttr := map[string]interface{}{
		"key": make(chan int),
	}
	err := env.SetSearchAttributesOnStart(invalidSearchAttr)
	require.Error(t, err)

	searchAttr := map[string]interface{}{
		"CustomIntField": 1,
	}
	err = env.SetSearchAttributesOnStart(searchAttr)
	require.NoError(t, err)
	require.NotNil(t, env.impl.workflowInfo.SearchAttributes)
}

func TestUnregisteredActivity(t *testing.T) {
	t.Parallel()
	env := newTestWorkflowEnv(t)
	workflow := func(ctx Context) error {
		ctx = WithActivityOptions(ctx, ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
		})
		return ExecuteActivity(ctx, "unregistered").Get(ctx, nil)
	}
	env.RegisterWorkflow(workflow)
	env.ExecuteWorkflow(workflow)
	require.Error(t, env.GetWorkflowError())
	ee := env.GetWorkflowError()
	require.NotNil(t, ee)
	require.True(t, strings.HasPrefix(ee.Error(), "unable to find activityType=unregistered"), ee.Error())
}

func TestNoExplicitRegistrationRequired(t *testing.T) {
	env := newTestWorkflowEnv(t)
	activity := func(ctx context.Context, arg string) (string, error) { return arg + " World!", nil }
	env.RegisterActivity(activity)
	env.ExecuteWorkflow(func(ctx Context, arg1 string) (string, error) {
		ctx = WithActivityOptions(ctx, ActivityOptions{
			ScheduleToCloseTimeout: time.Hour,
			StartToCloseTimeout:    time.Hour,
			ScheduleToStartTimeout: time.Hour,
		})
		var result string
		err := ExecuteActivity(ctx, activity, arg1).Get(ctx, &result)
		if err != nil {
			return "", err
		}
		return result, nil
	}, "Hello")
	require.NoError(t, env.GetWorkflowError())
	var result string
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", result)
}

func TestWorkflowReturnNil(t *testing.T) {
	env := newTestWorkflowEnv(t)

	var isExecuted bool
	testWF := func(ctx Context) error {
		isExecuted = true
		return nil
	}
	env.ExecuteWorkflow(testWF)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.True(t, isExecuted)

	var r struct{}
	err := env.GetWorkflowResult(&r)
	require.NoError(t, err)
}
