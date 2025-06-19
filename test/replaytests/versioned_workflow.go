package replaytests

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

const (
	// TestChangeID is a constant used to identify the version change in the workflow.
	TestChangeID = "test-change"

	// FooActivityName and BarActivityName are the names of the activities used in the workflows.
	FooActivityName = "FooActivity"
	BarActivityName = "BarActivity"

	// VersionedWorkflowName is the name of the versioned workflow.
	VersionedWorkflowName = "VersionedWorkflow"
)

var activityOptions = workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Minute,
	StartToCloseTimeout:    time.Minute,
	HeartbeatTimeout:       time.Second * 20,
}

// VersionedWorkflowV1 is the first version of the workflow, and it supports only DefaultVersion.
// It supports workflow executions started by this version VersionedWorkflowV1
// and VersionedWorkflowV2, as all of them will have the change ID set to DefaultVersion.
func VersionedWorkflowV1(ctx workflow.Context, _ string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	err := workflow.ExecuteActivity(ctx, FooActivityName, "data").Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// VersionedWorkflowV2 is the second version of the workflow. It supports DefaultVersion and Version 1.
// All workflows started by this version will have the change ID set to DefaultVersion.
// It supports workflow executions started by VersionedWorkflowV1 and VersionedWorkflowV2,
// as all of them will have the change ID set to DefaultVersion.
// It also supports workflow executions started by VersionedWorkflowV3 and VersionedWorkflowV4
// because the code supports execution of Version 1 of the workflow.
func VersionedWorkflowV2(ctx workflow.Context, _ string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	var err error

	version := workflow.GetVersion(ctx, TestChangeID, workflow.DefaultVersion, 1, workflow.ExecuteWithMinVersion())
	if version == workflow.DefaultVersion {
		err = workflow.ExecuteActivity(ctx, FooActivityName, "data").Get(ctx, &result)
	} else {
		err = workflow.ExecuteActivity(ctx, BarActivityName, "data").Get(ctx, &result)
	}
	if err != nil {
		return "", err
	}

	return result, nil
}

// VersionedWorkflowV3 is the third version of the workflow. It supports DefaultVersion and Version 1 as well.
// However, all workflows started by this version will have the change ID set to Version 1.
// It supports workflow executions started by VersionedWorkflowV1 and VersionedWorkflowV2,
// as all of them will have the change ID set to DefaultVersion, and it supports them.
// It also supports workflow executions started by VersionedWorkflowV3 and VersionedWorkflowV4,
// because the code supports execution of Version 1 of the workflow.
func VersionedWorkflowV3(ctx workflow.Context, _ string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	var err error

	version := workflow.GetVersion(ctx, TestChangeID, workflow.DefaultVersion, 1)
	if version == workflow.DefaultVersion {
		err = workflow.ExecuteActivity(ctx, FooActivityName, "data").Get(ctx, &result)
	} else {
		err = workflow.ExecuteActivity(ctx, BarActivityName, "data").Get(ctx, &result)
	}
	if err != nil {
		return "", err
	}

	return result, nil
}

// VersionedWorkflowV4 is the fourth version of the workflow. It supports only Version 1.
// All workflows started by this version will have the change ID set to Version 1.
// It supports workflow executions started by VersionedWorkflowV3 and VersionedWorkflowV4,
// as all of them will have the change ID set to Version 1.
func VersionedWorkflowV4(ctx workflow.Context, _ string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string

	workflow.GetVersion(ctx, TestChangeID, 1, 1)
	err := workflow.ExecuteActivity(ctx, BarActivityName, "data").Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// FooActivity returns "foo" as a result of the activity execution.
func FooActivity(ctx context.Context, _ string) (string, error) {
	return "foo", nil
}

// BarActivity returns "bar" as a result of the activity execution.
func BarActivity(ctx context.Context, _ string) (string, error) {
	return "bar", nil
}

// SetupWorkerForVersionedWorkflowV1 registers VersionedWorkflowV1 and FooActivity
func SetupWorkerForVersionedWorkflowV1(w worker.Registry) {
	w.RegisterWorkflowWithOptions(VersionedWorkflowV1, workflow.RegisterOptions{Name: VersionedWorkflowName})
	w.RegisterActivityWithOptions(FooActivity, activity.RegisterOptions{Name: FooActivityName})
}

// SetupWorkerForVersionedWorkflowV2 registers VersionedWorkflowV2, FooActivity, and BarActivity
func SetupWorkerForVersionedWorkflowV2(w worker.Registry) {
	w.RegisterWorkflowWithOptions(VersionedWorkflowV2, workflow.RegisterOptions{Name: VersionedWorkflowName})
	w.RegisterActivityWithOptions(FooActivity, activity.RegisterOptions{Name: FooActivityName})
	w.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: BarActivityName})
}

// SetupWorkerForVersionedWorkflowV3 registers VersionedWorkflowV3, FooActivity, and BarActivity
func SetupWorkerForVersionedWorkflowV3(w worker.Registry) {
	w.RegisterWorkflowWithOptions(VersionedWorkflowV3, workflow.RegisterOptions{Name: VersionedWorkflowName})
	w.RegisterActivityWithOptions(FooActivity, activity.RegisterOptions{Name: FooActivityName})
	w.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: BarActivityName})
}

// SetupWorkerForVersionedWorkflowV4 registers VersionedWorkflowV4 and BarActivity
func SetupWorkerForVersionedWorkflowV4(w worker.Registry) {
	w.RegisterWorkflowWithOptions(VersionedWorkflowV4, workflow.RegisterOptions{Name: VersionedWorkflowName})
	w.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: BarActivityName})
}
