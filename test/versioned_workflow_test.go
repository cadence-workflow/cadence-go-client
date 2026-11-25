package test

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

// VersionedWorkflowVersion is an enum representing the version of the VersionedWorkflow
type VersionedWorkflowVersion int

const (
	VersionedWorkflowVersionV1 VersionedWorkflowVersion = iota + 1
	VersionedWorkflowVersionV2
	VersionedWorkflowVersionV3
	VersionedWorkflowVersionV4
	VersionedWorkflowVersionV5
	VersionedWorkflowVersionV6

	// MaxVersionedWorkflowVersion is the maximum version of the VersionedWorkflow.
	// Update this constant when adding new versions to the workflow.
	MaxVersionedWorkflowVersion = VersionedWorkflowVersionV6
)

// VersionedWorkflowV1 is the first version of the workflow, supports only DefaultVersion.
// All workflows started by this version will have the change ID set to DefaultVersion.
func VersionedWorkflowV1(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	return workflow.ExecuteActivity(ctx, FooActivityName).Get(ctx, nil)
}

// VersionedWorkflowV2 is the second version of the workflow, supports DefaultVersion and 1
// All workflows started by this version will have the change ID set to DefaultVersion.
func VersionedWorkflowV2(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	version := workflow.GetVersion(ctx, TestChangeID, workflow.DefaultVersion, 1, workflow.ExecuteWithMinVersion())
	if version == workflow.DefaultVersion {
		return workflow.ExecuteActivity(ctx, FooActivityName).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
}

// VersionedWorkflowV3 is the third version of the workflow, supports DefaultVersion and 1
// All workflows started by this version will have the change ID set to 1.
func VersionedWorkflowV3(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	version := workflow.GetVersion(ctx, TestChangeID, workflow.DefaultVersion, 1)
	if version == workflow.DefaultVersion {
		return workflow.ExecuteActivity(ctx, FooActivityName).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
}

// VersionedWorkflowV4 is the fourth version of the workflow, supports only version 1
// All workflows started by this version will have the change ID set to 1.
func VersionedWorkflowV4(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	workflow.GetVersion(ctx, TestChangeID, 1, 1)
	return workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
}

// VersionedWorkflowV5 is the fifth version of the workflow, supports versions 1 and 2
// All workflows started by this version will have the change ID set to 1.
func VersionedWorkflowV5(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	version := workflow.GetVersion(ctx, TestChangeID, 1, 2, workflow.ExecuteWithVersion(1))
	if version == 1 {
		return workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, BazActivityName).Get(ctx, nil)
}

// VersionedWorkflowV6 is the sixth version of the workflow, supports versions 1 and 2
// All workflows started by this version will have the change ID set to 2.
func VersionedWorkflowV6(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	version := workflow.GetVersion(ctx, TestChangeID, 1, 2)
	if version == 1 {
		return workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
	}
	return workflow.ExecuteActivity(ctx, BazActivityName).Get(ctx, nil)
}

// FooActivity returns "foo" as a result of the activity execution.
func FooActivity(_ context.Context) (string, error) { return "foo", nil }

// BarActivity returns "bar" as a result of the activity execution.
func BarActivity(_ context.Context) (string, error) { return "bar", nil }

// BazActivity returns "baz" as a result of the activity execution.
func BazActivity(_ context.Context) (string, error) { return "baz", nil }

// SetupWorkerForVersionedWorkflow registers the versioned workflow and its activities
func SetupWorkerForVersionedWorkflow(version VersionedWorkflowVersion, w worker.Registry) {
	switch version {
	case VersionedWorkflowVersionV1:
		SetupWorkerForVersionedWorkflowV1(w)
	case VersionedWorkflowVersionV2:
		SetupWorkerForVersionedWorkflowV2(w)
	case VersionedWorkflowVersionV3:
		SetupWorkerForVersionedWorkflowV3(w)
	case VersionedWorkflowVersionV4:
		SetupWorkerForVersionedWorkflowV4(w)
	case VersionedWorkflowVersionV5:
		SetupWorkerForVersionedWorkflowV5(w)
	case VersionedWorkflowVersionV6:
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
