# Timer to Histogram Migration Guide

## Summary

**The timer→histogram migration is complete. The default is now `EmitHistogramsOnly`.**

- Only histogram metrics emit by default
- Users who still need timers can opt in with `EmitTimersOnly` or `EmitBoth`
- See `EMIT_MODE.md` for configuration details

## What Changed

All latency metrics now emit as histograms by default:
- **Histogram**: `cadence-decision-poll-latency_ns` (default)
- **Timer**: `cadence-decision-poll-latency` (opt-in via EmitTimersOnly or EmitBoth)

### Affected Metrics (62 total)

**Worker Metrics (13):**
- DecisionPollLatency, DecisionScheduledToStartLatency, DecisionExecutionLatency, DecisionResponseLatency
- ActivityPollLatency, ActivityScheduledToStartLatency, ActivityExecutionLatency, ActivityResponseLatency, ActivityEndToEndLatency
- LocalActivityExecutionLatency, WorkflowEndToEndLatency, WorkflowGetHistoryLatency, ReplayLatency

**Service Call Metrics (49 operations):**
- All Cadence service calls emit `cadence-latency` under their operation scope
- Examples: StartWorkflowExecution, SignalWithStartWorkflowExecution, TerminateWorkflowExecution, etc.

## Impact

**Cardinality:** Reduced — only histogram metrics emit by default
**Performance:** Minimal impact
**Compatibility:** Users can opt back into timers via EmitTimersOnly or EmitBoth

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

## Developer Guide

Use histogram APIs directly:

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

### Phase 1: Automatic Dual-Emit (Completed)
- Both metrics emitted automatically
- No code changes needed

### Phase 2: Histogram-Only Default (Current)
- Default switched to `EmitHistogramsOnly`
- Users who haven't migrated dashboards can opt in to `EmitBoth` or `EmitTimersOnly`

### Phase 3: Timers Removed (Future)
- Timer emission support removed in future major version
- Plenty of advance notice provided

## Testing

```bash
go test ./internal/common/metrics/... -v
```

## Questions?

- Review this guide
- Check the test files for examples
- Open a GitHub issue if you need help
