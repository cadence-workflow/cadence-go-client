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
	BazActivityName = "BazActivity"

	// VersionedWorkflowName is the name of the versioned workflow.
	VersionedWorkflowName = "VersionedWorkflow"
)

var activityOptions = workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Minute,
	StartToCloseTimeout:    time.Minute,
	HeartbeatTimeout:       time.Second * 20,
}

// VersionWorkflowVersion is an enum representing the version of the VersionedWorkflow
type VersionWorkflowVersion int

const (
	VersionWorkflowVersionV1 VersionWorkflowVersion = iota + 1
	VersionWorkflowVersionV2
	VersionWorkflowVersionV3
	VersionWorkflowVersionV4
	VersionWorkflowVersionV5
	VersionWorkflowVersionV6
)

// MaxVersionWorkflowVersion is the maximum version of the VersionedWorkflow.
// Update this constant when adding new versions to the workflow.
const MaxVersionWorkflowVersion = VersionWorkflowVersionV6

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

// VersionedWorkflowV5 is the fifth version of the workflow. It supports Version 1 and 2.
// All workflows started by this version will have the change ID set to Version 1.
// It supports workflow executions started by VersionedWorkflowV3, VersionedWorkflowV4,
// VersionedWorkflowV5, VersionedWorkflowV6
func VersionedWorkflowV5(ctx workflow.Context, _ string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	var err error

	version := workflow.GetVersion(ctx, TestChangeID, 1, 2, workflow.ExecuteWithVersion(1))
	if version == 1 {
		err = workflow.ExecuteActivity(ctx, BarActivityName, "data").Get(ctx, &result)
	} else {
		err = workflow.ExecuteActivity(ctx, BazActivityName, "data").Get(ctx, &result)
	}
	if err != nil {
		return "", err
	}

	return result, nil
}

// VersionedWorkflowV6 is the sixth version of the workflow. It supports Version 1 and 2.
// All workflows started by this version will have the change ID set to Version 2.
// It supports workflow executions started by VersionedWorkflowV3, VersionedWorkflowV4,
// VersionedWorkflowV5, VersionedWorkflowV6
func VersionedWorkflowV6(ctx workflow.Context, _ string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	var err error

	version := workflow.GetVersion(ctx, TestChangeID, 1, 2)
	if version == 1 {
		err = workflow.ExecuteActivity(ctx, BarActivityName, "data").Get(ctx, &result)
	} else {
		err = workflow.ExecuteActivity(ctx, BazActivityName, "data").Get(ctx, &result)
	}
	if err != nil {
		return "", err
	}

	return result, nil
}

// FooActivity returns "foo" as a result of the activity execution.
func FooActivity(_ context.Context, _ string) (string, error) {
	return "foo", nil
}

// BarActivity returns "bar" as a result of the activity execution.
func BarActivity(_ context.Context, _ string) (string, error) {
	return "bar", nil
}

// BazActivity returns "baz" as a result of the activity execution.
func BazActivity(_ context.Context, _ string) (string, error) {
	return "baz", nil
}

func SetupWorkerForVersionedWorkflow(version VersionWorkflowVersion, w worker.Registry) {
	switch version {
	case VersionWorkflowVersionV1:
		SetupWorkerForVersionedWorkflowV1(w)
	case VersionWorkflowVersionV2:
		SetupWorkerForVersionedWorkflowV2(w)
	case VersionWorkflowVersionV3:
		SetupWorkerForVersionedWorkflowV3(w)
	case VersionWorkflowVersionV4:
		SetupWorkerForVersionedWorkflowV4(w)
	case VersionWorkflowVersionV5:
		SetupWorkerForVersionedWorkflowV5(w)
	case VersionWorkflowVersionV6:
		SetupWorkerForVersionedWorkflowV6(w)
	default:
		panic("unsupported version for versioned workflow")
	}
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

// SetupWorkerForVersionedWorkflowV5 registers VersionedWorkflowV6, BarActivity and BazActivity
func SetupWorkerForVersionedWorkflowV5(w worker.Registry) {
	w.RegisterWorkflowWithOptions(VersionedWorkflowV5, workflow.RegisterOptions{Name: VersionedWorkflowName})
	w.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: BarActivityName})
	w.RegisterActivityWithOptions(BazActivity, activity.RegisterOptions{Name: BazActivityName})
}

// SetupWorkerForVersionedWorkflowV6 registers VersionedWorkflowV6, BarActivity and BazActivity
func SetupWorkerForVersionedWorkflowV6(w worker.Registry) {
	w.RegisterWorkflowWithOptions(VersionedWorkflowV6, workflow.RegisterOptions{Name: VersionedWorkflowName})
	w.RegisterActivityWithOptions(BarActivity, activity.RegisterOptions{Name: BarActivityName})
	w.RegisterActivityWithOptions(BazActivity, activity.RegisterOptions{Name: BazActivityName})
}
