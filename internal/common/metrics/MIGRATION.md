# Timer to Histogram Migration Guide

This guide covers the migration from timer metrics to exponential histograms in the Cadence Go client.

## Summary

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

### Affected Metrics (14 total):
1. `DecisionPollLatency` → also emits `DecisionPollLatency_ns`
2. `DecisionScheduledToStartLatency` → also emits `DecisionScheduledToStartLatency_ns`
3. `DecisionExecutionLatency` → also emits `DecisionExecutionLatency_ns`
4. `DecisionResponseLatency` → also emits `DecisionResponseLatency_ns`
5. `WorkflowGetHistoryLatency` → also emits `WorkflowGetHistoryLatency_ns`
6. `ActivityPollLatency` → also emits `ActivityPollLatency_ns`
7. `ActivityScheduledToStartLatency` → also emits `ActivityScheduledToStartLatency_ns`
8. `ActivityExecutionLatency` → also emits `ActivityExecutionLatency_ns`
9. `ActivityResponseLatency` → also emits `ActivityResponseLatency_ns`
10. `ActivityEndToEndLatency` → also emits `ActivityEndToEndLatency_ns`
11. `LocalActivityExecutionLatency` → also emits `LocalActivityExecutionLatency_ns`
12. `WorkflowEndToEndLatency` → also emits `WorkflowEndToEndLatency_ns`
13. `ReplayLatency` → also emits `ReplayLatency_ns`
14. `CadenceLatency` → also emits `CadenceLatency_ns`

## Impact

### Cardinality
- **+14 histogram metrics** (temporary increase during migration)
- Histograms have `_ns` suffix to distinguish from timers
- Eventually timers will be removed, returning to original cardinality

### Performance
- Minimal impact - histogram recording is fast
- Exponential buckets are efficient

### Compatibility
- **100% backward compatible** - existing dashboards continue working
- Timers still emit exactly as before
- New histograms are additive
- **No action required** - existing dashboards continue working

## Why Migrate?

Histograms provide several advantages over timers:
1. **Better precision control**: Exponential buckets give fine detail at low values, coarse at high values
2. **OTEL compatibility**: Compatible with OpenTelemetry's exponential histogram specification
3. **Cardinality control**: Can reduce resolution with `subsetTo()` for high-cardinality metrics
4. **Query-time aggregation**: Can be downsampled during querying for better visualization

## Migrating Your Dashboards

### Prometheus/Grafana Example:

**Before (Timer):**
```promql
histogram_quantile(0.99,
  rate(cadence_decision_poll_latency_bucket[5m])
)
```

**After (Histogram):**
```promql
histogram_quantile(0.99,
  rate(cadence_decision_poll_latency_ns_bucket[5m])
)
```

### M3/Statsd Example:

**Before (Timer):**
```
cadence.decision-poll-latency.p99
```

**After (Histogram):**
```
cadence.decision-poll-latency_ns.p99
```

## Histogram Buckets

Pre-defined histogram configurations:

### Default1ms100s (80 buckets, scale=2)
- **Range**: 1ms → ~15 minutes
- **Use for**: Most client-side latency metrics
- **Examples**: Decision poll latency, Decision execution latency, Activity execution latency (short-running), API call latency, Response latency

```go
metrics.RecordTimer(scope, metrics.DecisionPollLatency, latency, metrics.Default1ms100s)
```

### Low1ms100s (40 buckets, scale=1)
- **Range**: 1ms → ~15 minutes
- **Use for**: High-cardinality metrics where cost is a concern
- **Examples**: Per-activity-type latencies, Per-workflow-type latencies, Per-domain metrics

```go
// For metrics with many tag combinations
activityScope := scope.Tagged(map[string]string{"activity_type": activityType})
metrics.RecordTimer(activityScope, metrics.ActivityExecutionLatency, latency, metrics.Low1ms100s)
```

### High1ms24h (112 buckets, scale=2)
- **Range**: 1ms → ~3 days
- **Use for**: Long-running operations
- **Examples**: Workflow end-to-end latency, Long-running activity latency, Scheduled-to-start latency (can be hours/days)

```go
metrics.RecordTimer(scope, metrics.WorkflowEndToEndLatency, latency, metrics.High1ms24h)
```

### Mid1ms24h (56 buckets, scale=1)
- **Range**: 1ms → ~3 days
- **Use for**: Long-running operations with high cardinality
- **Examples**: Per-workflow-type end-to-end latency

```go
workflowScope := scope.Tagged(map[string]string{"workflow_type": workflowType})
metrics.RecordTimer(workflowScope, metrics.WorkflowEndToEndLatency, latency, metrics.Mid1ms24h)
```

## Developer Guide: Using Histogram APIs

### 1. Simple Duration Recording

**Before (Timer):**
```go
scope.Timer(metrics.DecisionPollLatency).Record(latency)
```

**After (Histogram only):**
```go
metrics.RecordTimer(scope, metrics.DecisionPollLatency, latency, metrics.Default1ms100s)
```

**Note**: In v1.X.0, both timer and histogram are emitted automatically. This example shows the histogram-only API for future use.

### 2. Stopwatch Pattern

**Before (Timer):**
```go
sw := scope.Timer(metrics.ActivityExecutionLatency).Start()
// ... do work ...
sw.Stop()
```

**After (Histogram only):**
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

**After (Histogram only):**
```go
start := time.Now()
// ... do work ...
metrics.RecordTimer(scope, metrics.WorkflowEndToEndLatency, time.Since(start), metrics.High1ms24h)
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

If you prefer to use histograms directly:

```go
// Record a duration
scope.Histogram(metrics.DecisionPollLatency+"_ns", metrics.Default1ms100s.Buckets()).
    RecordDuration(latency)

// Use stopwatch
sw := scope.Histogram(metrics.ActivityExecutionLatency+"_ns", metrics.Low1ms100s.Buckets()).Start()
sw.Stop()
```

**Important**: Always add the `_ns` suffix to histogram names to differentiate them from timers.

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

The histogram implementation includes comprehensive tests:

```bash
# Run all histogram tests
cd internal/common/metrics
go test -v -run TestHistogram

# See actual bucket values
go test -v -run TestHistogramValues

# Run all metrics tests
go test -v ./...
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

## Migration Timeline

### Phase 1: Automatic Dual-Emit (v1.X.0)
- ✅ Upgrade to v1.X.0
- ✅ Both timer and histogram metrics automatically emit
- ✅ No code changes needed
- ✅ Existing dashboards continue working

### Phase 2: Dashboard Migration (Your Pace)
1. Identify dashboards using timer metrics
2. Create new dashboards using histogram metrics (`_ns` suffix)
3. Compare side-by-side to validate accuracy
4. Switch over when confident

### Phase 3: Future - Timers Removed (v2.0+)
- Timer emission will be removed in a future major version
- Migrate dashboards before upgrading to v2.0+
- Plenty of advance notice will be provided

## Implementation Details

The following files were added/modified for histogram support:

### New Files:
- `internal/common/metrics/histograms.go` - Histogram bucket definitions and helpers
- `internal/common/metrics/histograms_test.go` - Histogram tests
- `internal/common/metrics/emit.go` - Dual-emit helper functions
- `internal/common/metrics/emit_test.go` - Dual-emit tests

### Modified Files:
- `internal/internal_task_pollers.go` - Updated to use dual-emit helpers
- `internal/internal_task_handlers.go` - Updated to use dual-emit helpers
- `internal/workflow_shadower_activities.go` - Updated to use dual-emit helpers
- `internal/common/metrics/service_wrapper.go` - Updated to use dual-emit helpers

## Questions or Issues?

For questions about the migration:
1. Check this guide first
2. Review the code examples above
3. Look at the test files for more examples
4. Open an issue on GitHub if you need help
