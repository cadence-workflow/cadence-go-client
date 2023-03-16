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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap/zaptest"
)

func TestReplayWorkflowHistoryFromFile(t *testing.T) {
	for _, testFile := range []string{"basic.json", "basic_new.json", "version.json", "version_new.json"} {
		t.Run("replay_"+strings.Split(testFile, ".")[0], func(t *testing.T) {
			replayer := worker.NewWorkflowReplayer()
			replayer.RegisterWorkflow(Workflow)
			replayer.RegisterWorkflow(Workflow2)

			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), testFile)
			require.NoError(t, err)
		})
	}
}

func TestReplayChildWorkflowBugBackport(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(childWorkflow, workflow.RegisterOptions{Name: "child"})
	replayer.RegisterWorkflowWithOptions(childWorkflowBug, workflow.RegisterOptions{Name: "parent"})

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "child_bug.json")
	require.NoError(t, err)
}

func TestGreetingsWorkflow(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}

// Should have failed but passed. Maybe, because the result recorded in history still matches the return type of the workflow.
func TestGreetingsWorkflow3(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	activity.RegisterWithOptions(getNameActivity3, activity.RegisterOptions{Name: "main.getNameActivity", DisableAlreadyRegisteredCheck: true})
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow3, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}

// Fails because the expected signature was different from history.
func TestGreetingsWorkflow4(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	activity.RegisterWithOptions(getNameActivity4, activity.RegisterOptions{Name: "main.getNameActivity", DisableAlreadyRegisteredCheck: true})
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow4, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}

// Fails with a non deterministic error. This passes in cadence_samples because it's registered in Helper.
// To test it on cadence_samples change the https://github.com/uber-common/cadence-samples/blob/master/cmd/samples/recipes/greetings/greetings_workflow.go
// to include the extra return types in getNameActivity.
func TestGreetingsWorkflow2(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	activity.RegisterWithOptions(getNameActivity2, activity.RegisterOptions{Name: "main.getNameActivity", DisableAlreadyRegisteredCheck: true})
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow2, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}
