# Histogram Metrics - Dual Emission

## Summary

This fork includes histogram metrics that are **always dual-emitted** alongside timer metrics.

## What Changed

Starting with this version, all latency metrics emit **both** timer and histogram formats:

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

## New Files

- `internal/common/metrics/histograms.go` - Histogram bucket definitions and helpers
- `internal/common/metrics/histograms_test.go` - Histogram tests
- `internal/common/metrics/emit.go` - Dual-emit helper functions
- `internal/common/metrics/emit_test.go` - Dual-emit tests
- `internal/common/metrics/MIGRATION.md` - Migration guide for users

## Modified Files

- `internal/internal_task_pollers.go` - Updated to use dual-emit helpers
- `internal/internal_task_handlers.go` - Updated to use dual-emit helpers
- `internal/workflow_shadower_activities.go` - Updated to use dual-emit helpers
- `internal/common/metrics/service_wrapper.go` - Updated to use dual-emit helpers

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

## Usage

No code changes needed! Dual-emission happens automatically.

```go
// In your code, this still works:
scope.Timer(metrics.DecisionPollLatency).Record(latency)

// But now you also get:
// - cadence-decision-poll-latency (timer, existing)
// - cadence-decision-poll-latency_ns (histogram, new)
```

## Migrating Dashboards

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

- **`Default1ms100s`**: 1ms → ~15 min (80 buckets, scale=2)
  - Use for: Most client-side latency metrics

- **`Low1ms100s`**: 1ms → ~15 min (40 buckets, scale=1)
  - Use for: High-cardinality metrics

- **`High1ms24h`**: 1ms → ~3 days (112 buckets, scale=2)
  - Use for: Long-running operations

- **`Mid1ms24h`**: 1ms → ~3 days (56 buckets, scale=1)
  - Use for: Long-running operations with high cardinality

## Testing

All tests pass:
```bash
go test ./internal/common/metrics/... -v
```

## Documentation

See `internal/common/metrics/MIGRATION.md` for detailed migration guide.

## Future

In a future major version (v2.0+), timer emission may be removed, keeping only histograms. This gives everyone time to migrate dashboards and alerts at their own pace.
