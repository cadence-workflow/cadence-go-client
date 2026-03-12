# Timer to Histogram Migration Guide

## Summary

**Starting in v1.3.1, all latency metrics automatically emit BOTH timer and histogram formats.**

- Your existing dashboards continue to work (timers)
- New histogram metrics available (with `_ns` suffix)
- Migrate dashboards at your own pace
- No code changes needed

## What Changed

All latency metrics now dual-emit:
- **Timer**: `cadence-decision-poll-latency` (existing)
- **Histogram**: `cadence-decision-poll-latency_ns` (new)

### Affected Metrics (62 total)

**Worker Metrics (13):**
- DecisionPollLatency, DecisionScheduledToStartLatency, DecisionExecutionLatency, DecisionResponseLatency
- ActivityPollLatency, ActivityScheduledToStartLatency, ActivityExecutionLatency, ActivityResponseLatency, ActivityEndToEndLatency
- LocalActivityExecutionLatency, WorkflowEndToEndLatency, WorkflowGetHistoryLatency, ReplayLatency

**Service Call Metrics (49 operations):**
- All Cadence service calls emit `cadence-latency` under their operation scope
- Examples: StartWorkflowExecution, SignalWithStartWorkflowExecution, TerminateWorkflowExecution, etc.

## Impact

**Cardinality:** +62 histogram metrics (temporary during migration)
**Performance:** Minimal impact
**Compatibility:** 100% backward compatible

## Why Migrate?

1. **Better precision control** - Exponential buckets: fine detail at low values, coarse at high
2. **OTEL compatible** - OpenTelemetry exponential histogram specification
3. **Cardinality control** - Can reduce resolution with `subsetTo()`
4. **Query-time aggregation** - Downsample during queries

## Histogram Buckets

### Default1ms100s (80 buckets)
- **Range**: 1ms → ~15 minutes
- **Use for**: Most client-side metrics (API calls, decision/activity execution, polls)

### Low1ms100s (40 buckets)
- **Range**: 1ms → ~15 minutes
- **Use for**: High-cardinality metrics (per-activity-type, per-workflow-type)

### High1ms24h (112 buckets)
- **Range**: 1ms → ~3 days
- **Use for**: Long-running operations (workflow end-to-end, long activities, scheduled-to-start)

### Mid1ms24h (56 buckets)
- **Range**: 1ms → ~3 days
- **Use for**: Long-running operations with high cardinality

## Developer Guide (Future)

Currently, the client automatically dual-emits. In a future version when you want to use histogram-only APIs:

```go
// Simple recording
metrics.RecordHistogram(scope, metrics.DecisionPollLatency, latency, metrics.Default1ms100s)

// Stopwatch pattern
sw := metrics.StartHistogram(scope, metrics.ActivityExecutionLatency, metrics.Default1ms100s)
// ... do work ...
sw.Stop()
```

## Choosing the Right Histogram

| Metric Type | Cardinality | Recommended |
|-------------|-------------|-------------|
| Short operations (<15min) | Low | Default1ms100s |
| Short operations (<15min) | High | Low1ms100s |
| Long operations (hours/days) | Low | High1ms24h |
| Long operations (hours/days) | High | Mid1ms24h |

**High cardinality** = metrics with many tag combinations (e.g., tagged by activity_type or workflow_type)

## Migration Timeline

### Phase 1: Automatic Dual-Emit (Now)
- Upgrade to v1.3.1
- Both metrics automatically emit
- No code changes needed

### Phase 2: Migrate Dashboards (Your Pace)
1. Identify dashboards using timer metrics
2. Create new panels using histogram metrics (`_ns` suffix)
3. Compare side-by-side
4. Switch over when validated

### Phase 3: Timers Removed (Future v1.3.*)
- Timer emission removed in future major version
- Plenty of advance notice provided

## Testing

```bash
go test ./internal/common/metrics/... -v
```

## Questions?

- Review this guide
- Check the test files for examples
- Open a GitHub issue if you need help
