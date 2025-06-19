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

	"go.uber.org/cadence/activity"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
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

// Gives a non-deterministic-error because the getGreetingActivitytest was not registered on the replayer.
func TestGreetingsWorkflowforActivity(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(greetingsWorkflowActivity, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.Error(t, err)
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
	replayer.RegisterActivityWithOptions(getNameActivity3, activity.RegisterOptions{Name: "main.getNameActivity", DisableAlreadyRegisteredCheck: true})
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow3, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}

// Fails because the expected signature was different from history.
func TestGreetingsWorkflow4(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterActivityWithOptions(getNameActivity4, activity.RegisterOptions{Name: "main.getNameActivity", DisableAlreadyRegisteredCheck: true})
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow4, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.Error(t, err)
}

// Panic with failed to register activity. This passes in cadence_samples because it's registered in Helper.
// To test it on cadence_samples change the https://github.com/uber-common/cadence-samples/blob/master/cmd/samples/recipes/greetings/greetings_workflow.go
// to include the extra return types in getNameActivity.
func TestGreetingsWorkflow2(t *testing.T) {

	t.Skip("Panic with failed to register activity. Here the activity returns incompatible arguments so the test should fail")
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterActivityWithOptions(getNameActivity2, activity.RegisterOptions{Name: "main.getNameActivity", DisableAlreadyRegisteredCheck: true})
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow2, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.Error(t, err)
}

// Ideally replayer doesn't concern itself with the change in the activity content until it matches the expected output type.
// History has recorded the output of banana activity instead. The replayer should have failed because we have not registered any
// activity here in the test.
// The replayer still runs whatever it found in the history and passes.
func TestExclusiveChoiceWorkflowWithUnregisteredActivity(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(exclusiveChoiceWorkflow, workflow.RegisterOptions{Name: "choice"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "choice.json")
	require.NoError(t, err)
}

// This test registers Cherry Activity as the activity but calls Apple activity in the workflow code. Infact, Cherry and Banana
// activities are not even a part of the workflow code in question.
// History has recorded the output of banana activity. Here, The workflow is not waiting for the activity so it doesn't notice
// that registered activity is different from executed activity.
// The replayer relies on whatever is recorded in the History so as long as the main activity name in the options matched partially
// it doesn't raise errors.
func TestExclusiveChoiceWorkflowWithDifferentActvityCombo(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(exclusiveChoiceWorkflow2, workflow.RegisterOptions{Name: "choice"})
	replayer.RegisterActivityWithOptions(getAppleOrderActivity, activity.RegisterOptions{Name: "main.getOrderActivity"})
	replayer.RegisterActivityWithOptions(orderAppleActivity, activity.RegisterOptions{Name: "testactivity"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "choice.json")
	require.NoError(t, err)
}

func TestBranchWorkflow(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(sampleBranchWorkflow, workflow.RegisterOptions{Name: "branch"})

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch.json")
	require.NoError(t, err)
}

// Fails with a non deterministic error because there was an additional unexpected branch. Decreasing the number of branches will
// also fail the test because the history expects the same number of branches executing the activity.
func TestBranchWorkflowWithExtraBranch(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(sampleBranchWorkflow2, workflow.RegisterOptions{Name: "branch"})

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch.json")
	assert.ErrorContains(t, err, "nondeterministic workflow")
}

// TestSequentialStepsWorkflow replays a history with 2 sequential activity calls and runs it against new version of the workflow code which only calls 1 activity.
// This should be considered as non-determinism error.
func TestSequentialStepsWorkflow(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(replayerHelloWorldWorkflow, workflow.RegisterOptions{Name: "fx.ReplayerHelloWorldWorkflow"})
	replayer.RegisterActivityWithOptions(replayerHelloWorldActivity, activity.RegisterOptions{Name: "replayerhello"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "sequential.json")
	assert.NoError(t, err)
}

func TestParallel(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(sampleParallelWorkflow, workflow.RegisterOptions{Name: "branch2"})

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch2.json")
	require.NoError(t, err)
}

// Should have failed since the first go routine has only one branch whereas the history has two branches.
// The replayer totally misses this change.
func TestParallel2(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflowWithOptions(sampleParallelWorkflow2, workflow.RegisterOptions{Name: "branch2"})

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch2.json")
	require.NoError(t, err)
}

// Runs a history which ends with WorkflowExecutionContinuedAsNew. Replay fails because of the additional checks done
// for continue as new case by replayWorkflowHistory().
// This should not have any error because it's a valid continue as new case.
func TestContinueAsNew(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(ContinueAsNewWorkflow, workflow.RegisterOptions{Name: "fx.SimpleSignalWorkflow"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "continue_as_new.json")
	assert.ErrorContains(t, err, "missing replay decision for WorkflowExecutionContinuedAsNew")
}

// TestSafeDeploymentVersionedWorkflow verifies that versioned workflows can be executed
// safely across different versions without causing non-deterministic errors.
// There are 2 workflow executions:
//
// * VersionedWorkflowFoo - which is the first version of the workflow which version of change id is DefaultVersion
//   - This workflow is supposed to execute FooActivity
//
// * VersionedWorkflowBar - which is the second version of the workflow which version of change id is 1
//   - This workflow is supposed to execute BarActivity
//
// There are 4 versions of the workflow:
//
// * VersionedWorkflowV1 - which supports only DefaultVersion and executes FooActivity
//   - This workflow is able to replay the history of only of VersionedWorkflowFoo
//
// * VersionedWorkflowV2 - which supports DefaultVersion and Version 1, and can execute FooActivity or BarActivity
//   - This workflow is able to replay the history of both of VersionedWorkflowFoo and VersionedWorkflowBar
//   - A first execution of this workflow will should execute FooActivity, because of usage workflow.ExecuteWithMinVersion(),
//     but the test can't check it due to Replay
//
// * VersionedWorkflowV3 - which supports DefaultVersion and Version 1, and can execute FooActivity or BarActivity
//   - This workflow is able to replay the history of both of VersionedWorkflowFoo and VersionedWorkflowBar
//   - A first execution of this workflow will should execute BarActivity, but the test can't check it due to Replay
//
// * VersionedWorkflowV4 - which supports Version 1, and can execute BarActivity
//   - This workflow is able to replay the history only of VersionedWorkflowBar
//
// So the test focusing workflows supports forward and backward compatibility of the workflows
func TestVersionedWorkflows(t *testing.T) {
	const (
		versionedWorkflowFooHistoryFile = "versioned_workflow_foo.json"
		versionedWorkflowBarHistoryFile = "versioned_workflow_bar.json"
	)

	t.Run("VersionedWorkflowV1", func(t *testing.T) {
		replayer := worker.NewWorkflowReplayer()
		replayer.RegisterWorkflowWithOptions(VersionedWorkflowV1, workflow.RegisterOptions{Name: versionedWorkflowName})
		replayer.RegisterActivityWithOptions(FooActivity, activity.RegisterOptions{Name: fooActivityName})

		t.Run("successfully replayed with VersionedWorkflowFoo", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowFooHistoryFile)
			require.NoError(t, err, "Failed to replay VersionedWorkflowFoo history")
		})

		t.Run("fail to replay with VersionedWorkflowBar", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowBarHistoryFile)
			require.Error(t, err, "Expected to fail replaying VersionedWorkflowBar history")
		})
	})

	t.Run("VersionedWorkflowV2", func(t *testing.T) {
		replayer := worker.NewWorkflowReplayer()
		replayer.RegisterWorkflowWithOptions(VersionedWorkflowV2, workflow.RegisterOptions{Name: versionedWorkflowName})
		replayer.RegisterActivityWithOptions(FooActivity, activity.RegisterOptions{Name: fooActivityName})
		replayer.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: barActivityName})

		t.Run("successfully replayed with VersionedWorkflowFoo", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowFooHistoryFile)
			require.NoError(t, err, "Failed to replay VersionedWorkflowFoo history")
		})

		t.Run("successfully replayed with VersionedWorkflowBar", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowBarHistoryFile)
			require.NoError(t, err, "Failed to replay VersionedWorkflowBar history")
		})
	})

	t.Run("VersionedWorkflowV3", func(t *testing.T) {
		replayer := worker.NewWorkflowReplayer()
		replayer.RegisterWorkflowWithOptions(VersionedWorkflowV3, workflow.RegisterOptions{Name: versionedWorkflowName})
		replayer.RegisterActivityWithOptions(FooActivity, activity.RegisterOptions{Name: fooActivityName})
		replayer.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: barActivityName})

		t.Run("successfully replayed with VersionedWorkflowFoo", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowFooHistoryFile)
			require.NoError(t, err, "Failed to replay VersionedWorkflowFoo history")
		})

		t.Run("successfully replayed with VersionedWorkflowBar", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowBarHistoryFile)
			require.NoError(t, err, "Failed to replay VersionedWorkflowBar history")
		})
	})

	t.Run("VersionedWorkflowV4", func(t *testing.T) {
		replayer := worker.NewWorkflowReplayer()
		replayer.RegisterWorkflowWithOptions(VersionedWorkflowV4, workflow.RegisterOptions{Name: versionedWorkflowName})
		replayer.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: barActivityName})

		t.Run("fail to replay with VersionedWorkflowFoo", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowFooHistoryFile)
			require.Error(t, err, "Expected to fail replaying VersionedWorkflowFoo history")
		})

		t.Run("successfully replayed with VersionedWorkflowBar", func(t *testing.T) {
			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), versionedWorkflowBarHistoryFile)
			require.NoError(t, err, "Failed to replay VersionedWorkflowBar history")
		})
	})
}
