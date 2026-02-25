# Migrating from Timers to Histograms

This guide shows how to migrate existing timer metrics to exponential histograms in the Cadence Go client.

## Dual-Emission by Default

**Starting in v1.X.0, all latency metrics automatically emit BOTH timer and histogram formats.**

This gives you a smooth migration path:
- Your existing dashboards and alerts continue to work (using timers)
- New histogram metrics are available for migration (with `_ns` suffix)
- Migrate at your own pace - no hard deadline
- In a future release, timer emission may be removed (with advance notice)

## What Changed

When you upgrade to v1.X.0:
- **Timer metrics**: Continue to work as before (e.g., `cadence-decision-poll-latency`)
- **Histogram metrics**: Automatically added (e.g., `cadence-decision-poll-latency_ns`)
- **No code changes needed**: Dual-emission happens automatically

## Impact

- **+14 histogram metrics** added to your metrics (one per latency metric)
- **Temporary cardinality increase** during migration period
- **No action required** - existing dashboards continue working

## Why Migrate?

Histograms provide several advantages over timers:
1. **Better precision control**: Exponential buckets give fine detail at low values, coarse at high values
2. **OTEL compatibility**: Compatible with OpenTelemetry's exponential histogram specification
3. **Cardinality control**: Can reduce resolution with `subsetTo()` for high-cardinality metrics
4. **Query-time aggregation**: Can be downsampled during querying for better visualization

## Quick Start

### 1. Simple Duration Recording

**Before (Timer):**
```go
scope.Timer(metrics.DecisionPollLatency).Record(latency)
```

**After (Histogram):**
```go
metrics.RecordTimer(scope, metrics.DecisionPollLatency, latency, metrics.Default1ms100s)
```

### 2. Stopwatch Pattern

**Before (Timer):**
```go
sw := scope.Timer(metrics.ActivityExecutionLatency).Start()
// ... do work ...
sw.Stop()
```

**After (Histogram):**
```go
sw := metrics.StartTimer(scope, metrics.ActivityExecutionLatency, metrics.Default1ms100s)
// ... do work ...
sw.Stop()
```

### 3. Manual Timing

**Before (Timer):**
```go
start := time.Now()
// ... do work ...
scope.Timer(metrics.WorkflowEndToEndLatency).Record(time.Since(start))
```

**After (Histogram):**
```go
start := time.Now()
// ... do work ...
metrics.RecordTimer(scope, metrics.WorkflowEndToEndLatency, time.Since(start), metrics.High1ms24h)
```

## Choosing the Right Histogram

### Default1ms100s (80 buckets)
- **Range**: 1ms → ~15 minutes
- **Use for**: Most client-side latency metrics
- **Examples**:
  - Decision poll latency
  - Decision execution latency
  - Activity execution latency (short-running)
  - API call latency
  - Response latency

```go
metrics.RecordTimer(scope, metrics.DecisionPollLatency, latency, metrics.Default1ms100s)
```

### Low1ms100s (40 buckets)
- **Range**: 1ms → ~15 minutes
- **Use for**: High-cardinality metrics where cost is a concern
- **Examples**:
  - Per-activity-type latencies
  - Per-workflow-type latencies
  - Per-domain metrics

```go
// For metrics with many tag combinations
activityScope := scope.Tagged(map[string]string{"activity_type": activityType})
metrics.RecordTimer(activityScope, metrics.ActivityExecutionLatency, latency, metrics.Low1ms100s)
```

### High1ms24h (112 buckets)
- **Range**: 1ms → ~3 days
- **Use for**: Long-running operations
- **Examples**:
  - Workflow end-to-end latency
  - Long-running activity latency
  - Scheduled-to-start latency (can be hours/days)

```go
metrics.RecordTimer(scope, metrics.WorkflowEndToEndLatency, latency, metrics.High1ms24h)
```

### Mid1ms24h (56 buckets)
- **Range**: 1ms → ~3 days
- **Use for**: Long-running operations with high cardinality
- **Examples**:
  - Per-workflow-type end-to-end latency

```go
workflowScope := scope.Tagged(map[string]string{"workflow_type": workflowType})
metrics.RecordTimer(workflowScope, metrics.WorkflowEndToEndLatency, latency, metrics.Mid1ms24h)
```

## Complete Migration Examples

### Example 1: Decision Task Handler

**Before:**
```go
func (wth *workflowTaskHandlerImpl) ProcessWorkflowTask(task *workflowTask) error {
    scope := wth.metricsScope

    // Poll latency
    pollLatency := time.Since(task.pollTime)
    scope.Timer(metrics.DecisionPollLatency).Record(pollLatency)

    // Execution timing
    executionStart := time.Now()
    err := wth.executeDecision(task)
    executionLatency := time.Since(executionStart)
    scope.Timer(metrics.DecisionExecutionLatency).Record(executionLatency)

    return err
}
```

**After:**
```go
func (wth *workflowTaskHandlerImpl) ProcessWorkflowTask(task *workflowTask) error {
    scope := wth.metricsScope

    // Poll latency - use Default1ms100s for short operations
    pollLatency := time.Since(task.pollTime)
    metrics.RecordTimer(scope, metrics.DecisionPollLatency, pollLatency, metrics.Default1ms100s)

    // Execution timing - use Default1ms100s for typical decision execution
    executionStart := time.Now()
    err := wth.executeDecision(task)
    executionLatency := time.Since(executionStart)
    metrics.RecordTimer(scope, metrics.DecisionExecutionLatency, executionLatency, metrics.Default1ms100s)

    return err
}
```

### Example 2: Activity Handler with High Cardinality

**Before:**
```go
func (ath *activityTaskHandlerImpl) Execute(task *activityTask) error {
    // Create scope with activity type tag
    activityScope := ath.metricsScope.Tagged(map[string]string{
        "activity_type": task.activityType,
    })

    sw := activityScope.Timer(metrics.ActivityExecutionLatency).Start()
    defer sw.Stop()

    return ath.executeActivity(task)
}
```

**After:**
```go
func (ath *activityTaskHandlerImpl) Execute(task *activityTask) error {
    // Create scope with activity type tag
    activityScope := ath.metricsScope.Tagged(map[string]string{
        "activity_type": task.activityType,
    })

    // Use Low1ms100s because we're tagging by activity type (high cardinality)
    sw := metrics.StartTimer(activityScope, metrics.ActivityExecutionLatency, metrics.Low1ms100s)
    defer sw.Stop()

    return ath.executeActivity(task)
}
```

### Example 3: Workflow End-to-End Latency

**Before:**
```go
func (w *workflowClient) recordWorkflowCompleted(execution *WorkflowExecution, closeTime time.Time) {
    endToEndLatency := closeTime.Sub(execution.StartTime)
    w.metricsScope.Timer(metrics.WorkflowEndToEndLatency).Record(endToEndLatency)
}
```

**After:**
```go
func (w *workflowClient) recordWorkflowCompleted(execution *WorkflowExecution, closeTime time.Time) {
    endToEndLatency := closeTime.Sub(execution.StartTime)
    // Use High1ms24h for long-running workflows that can take hours/days
    metrics.RecordTimer(w.metricsScope, metrics.WorkflowEndToEndLatency, endToEndLatency, metrics.High1ms24h)
}
```

## Direct Usage (Without Helpers)

If you prefer to use histograms directly without the helper functions:

```go
// Record a duration
scope.Histogram(metrics.DecisionPollLatency+"_ns", metrics.Default1ms100s.Buckets()).
    RecordDuration(latency)

// Use stopwatch
sw := scope.Histogram(metrics.ActivityExecutionLatency+"_ns", metrics.Low1ms100s.Buckets()).Start()
sw.Stop()
```

**Important**: Always add the `_ns` suffix to histogram names to differentiate them from timers.

## Migration Strategy

### Phase 1: Dual Emit (Recommended)
Emit both timer and histogram to compare:

```go
// Old timer (keep temporarily)
scope.Timer(metrics.DecisionPollLatency).Record(latency)

// New histogram (add)
metrics.RecordTimer(scope, metrics.DecisionPollLatency, latency, metrics.Default1ms100s)
```

### Phase 2: Monitor & Validate
1. Verify histogram data looks correct in your metrics backend
2. Compare timer and histogram percentiles (p50, p99, etc.)
3. Monitor cardinality and cost

### Phase 3: Remove Timers
Once validated, remove the timer calls:

```go
// Remove this line
// scope.Timer(metrics.DecisionPollLatency).Record(latency)

// Keep this
metrics.RecordTimer(scope, metrics.DecisionPollLatency, latency, metrics.Default1ms100s)
```

## Naming Convention

All histogram metrics MUST have a `_ns` suffix:
- Good: `cadence-decision-poll-latency_ns`
- Bad: `cadence-decision-poll-latency` (conflicts with timer name)

The helper functions (`RecordTimer`, `StartTimer`) automatically add this suffix.

## Cardinality Considerations

High-cardinality metrics (many unique tag combinations) should use lower-resolution histograms:

| Cardinality | Recommended Histogram |
|-------------|----------------------|
| Low (no tags or 1-2 tag values) | Default1ms100s (80 buckets) |
| Medium (5-10 tag combinations) | Low1ms100s (40 buckets) |
| High (100+ tag combinations) | Low1ms100s or consider removing tags |

For very high cardinality, consider:
1. Sampling the metric
2. Using Low1ms100s
3. Creating a custom histogram with even fewer buckets
4. Removing some tags

## Testing

The histogram implementation includes comprehensive tests. To run them:

```bash
cd internal/common/metrics
go test -v -run TestHistogram
```

To see the actual bucket values:

```bash
go test -v -run TestHistogramValues
```

## Advanced: Creating Custom Histograms

If the predefined histograms don't meet your needs:

```go
// For very fast operations (microseconds to milliseconds)
customFastOps := makeSubsettableHistogram(
    2,                    // scale (0-3, higher = more buckets)
    100*time.Microsecond, // start
    10*time.Millisecond,  // end (actual will exceed by 2x)
    40,                   // length (must be divisible by 2^scale)
)

// For very slow operations (minutes to hours)
customSlowOps := makeSubsettableHistogram(
    1,              // scale (lower = fewer buckets)
    time.Minute,    // start
    12*time.Hour,   // end
    40,             // length
)
```

**Note**: Only create custom histograms if the predefined ones truly don't fit. Consistency is valuable.
