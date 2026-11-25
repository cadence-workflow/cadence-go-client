# Metric Emission Mode Configuration

This document describes how to configure metric emission behavior for the timer→histogram migration.

## Overview

The cadence-go-client supports three metric emission modes:

1. **EmitTimersOnly** - Only timer metrics (legacy OSS behavior)
2. **EmitBoth** - Both timer and histogram metrics (for migration)
3. **EmitHistogramsOnly** (default) - Only histogram metrics (post-migration)

**Default**: `EmitHistogramsOnly` - The migration from timers to histograms is complete. Users who still need timers can opt in via configuration.

## Configuration Methods

### Method 1: WorkerOptions.FeatureFlags (Recommended - Code)

Configure the emission mode when creating a worker in code:

```go
import (
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/internal/common/metrics"
)

func main() {
	// Create worker with feature flags
	w := worker.New(
		serviceClient,
		domain,
		taskList,
		worker.Options{
			FeatureFlags: internal.FeatureFlags{
				// Option A: Use only timers (legacy behavior)
				MetricEmitMode: metrics.EmitTimersOnly,

				// Option B: Use both (dual-emit for migration)
				// MetricEmitMode: metrics.EmitBoth,

				// Option C: Use only histograms (default, no need to set)
				// MetricEmitMode: metrics.EmitHistogramsOnly,
			},
		},
	)
}
```

### Method 2: YAML Configuration (For cadencefx/monorepo)

If using cadencefx or YAML-based worker configuration:

```yaml
cadence:
  workers:
    - domain: my-domain
      task_list: my-task-list
      options:
        max_concurrent_activity_execution_size: 100
        worker_activities_per_second: 5
        # ... other options ...

        # Feature flags for controlling worker behavior
        feature_flags:
          # Options: "timers_only", "both", "histograms_only"
          # Default: "histograms_only" (post-migration)
          metric_emit_mode: "timers_only"  # or "both" or "histograms_only"
```

**Note**: YAML configuration support requires cadencefx integration. See your internal documentation for details.

### Method 3: Global SetEmitMode() Function

Set the mode globally during application initialization:

```go
import (
	"go.uber.org/cadence/internal/common/metrics"
)

func init() {
	// Set globally for all workers
	metrics.SetEmitMode(metrics.EmitTimersOnly)
}
```

**Note**: WorkerOptions.FeatureFlags.MetricEmitMode (Method 1) takes precedence and will override the global setting set by SetEmitMode(). Each worker initializes its emit mode from its own FeatureFlags.MetricEmitMode configuration.

## Use Cases

### OSS Users Who Want Only Timers

If you want to keep only the legacy timer metrics (no histograms):

```go
worker.Options{
	FeatureFlags: internal.FeatureFlags{
		MetricEmitMode: metrics.EmitTimersOnly,
	},
}

## Migration Path

### Phase 1: Dual Emission (Completed)

The `EmitBoth` mode was the previous default during migration.

- Both timer and histogram metrics were emitted
- New dashboards/alerts were created using histogram metrics (`_ns` suffix)
- Histogram metrics were validated against timer metrics

### Phase 2: Histogram-Only (Current Default)

The default is now `EmitHistogramsOnly`:

- Only histogram metrics emit
- Remove old timer-based dashboards/alerts
- Migration complete!

## Affected Metrics

All 62 latency metrics are controlled by this setting:

**Worker Metrics (13):**
- DecisionPollLatency, DecisionScheduledToStartLatency, DecisionExecutionLatency, DecisionResponseLatency
- ActivityPollLatency, ActivityScheduledToStartLatency, ActivityExecutionLatency, ActivityResponseLatency, ActivityEndToEndLatency
- LocalActivityExecutionLatency, WorkflowEndToEndLatency, WorkflowGetHistoryLatency, ReplayLatency

**Service Call Metrics (49 operations):**
- All Cadence service calls emit `cadence-latency` under their operation scope

## Benefits

✅ **Consistent across OSS and monorepo** - Same mechanism for everyone
✅ **Safe default** - Histogram-only reduces metric cardinality post-migration
✅ **Per-worker configuration** - Different workers can have different modes
✅ **Clear and explicit** - Configuration visible in WorkerOptions
✅ **Simple and lightweight** - No runtime overhead


## Questions?

- Review `MIGRATION.md` for general migration guidance
- Check test files for usage examples
- Open a GitHub issue if you need help
