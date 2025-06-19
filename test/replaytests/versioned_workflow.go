package replaytests

import (
	"go.uber.org/cadence/workflow"
	"time"
)

const (
	// testChangeID is a constant used to identify the version change in the workflow.
	testChangeID = "test-change"

	// fooActivityName and barActivityName are the names of the activities used in the workflows.
	fooActivityName = "FooActivity"
	barActivityName = "BarActivity"

	// versionedWorkflowName is the name of the versioned workflow.
	versionedWorkflowName = "VersionedWorkflow"
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
	err := workflow.ExecuteActivity(ctx, fooActivityName, "").Get(ctx, &result)
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

	version := workflow.GetVersion(ctx, testChangeID, workflow.DefaultVersion, 1, workflow.ExecuteWithMinVersion())
	if version == workflow.DefaultVersion {
		err = workflow.ExecuteActivity(ctx, fooActivityName, "").Get(ctx, &result)
	} else {
		err = workflow.ExecuteActivity(ctx, barActivityName, "").Get(ctx, &result)
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

	version := workflow.GetVersion(ctx, testChangeID, workflow.DefaultVersion, 1)
	if version == workflow.DefaultVersion {
		err = workflow.ExecuteActivity(ctx, fooActivityName, "").Get(ctx, &result)
	} else {
		err = workflow.ExecuteActivity(ctx, barActivityName, "").Get(ctx, &result)
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

	workflow.GetVersion(ctx, testChangeID, 1, 1)
	err := workflow.ExecuteActivity(ctx, barActivityName, "").Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// FooActivity returns "foo" as a result of the activity execution.
func FooActivity(ctx workflow.Context, _ string) (string, error) {
	return "foo", nil
}

// BarActivity returns "bar" as a result of the activity execution.
func BarActivity(ctx workflow.Context, _ string) (string, error) {
	return "bar", nil
}
