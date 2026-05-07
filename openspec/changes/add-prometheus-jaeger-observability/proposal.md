## Why

The TD27 Admin backend currently lacks observability capabilities beyond structured logging and operation audit logs. There is no visibility into request performance, error rates, Go runtime health, or end-to-end request flows. This makes debugging production issues slow, prevents proactive performance optimization, and eliminates the ability to correlate log entries with request traces. Adding Prometheus metrics and Jaeger distributed tracing solves these problems with minimal implementation effort and zero breaking changes to existing functionality.

## What Changes

- New Prometheus `/metrics` endpoint exposing standard RED (Rate/Error/Duration) HTTP metrics
- Auto-collected Go runtime and process metrics (goroutine count, memory usage, CPU time, etc.)
- OpenTelemetry-based distributed tracing with direct export to Jaeger (no collector required for local/development use)
- Automatic trace ID/span ID injection into all existing `slog` log records for log-trace correlation
- GORM query tracing to track DB query performance within request traces
- Context propagation pattern that works with existing service/repository architecture
- No breaking changes to existing APIs, logging infrastructure, or business logic

## Capabilities

### New Capabilities
- `prometheus-metrics`: Prometheus metrics endpoint with standard HTTP and runtime metrics
- `jaeger-distributed-tracing`: OpenTelemetry-based distributed tracing with Jaeger export, trace ID log injection, and GORM instrumentation

### Modified Capabilities
- None: No existing specification requirements are changed, this is purely additive functionality

## Impact

- **Affected code**: `server/configs/`, `server/internal/initialize/`, `server/internal/middleware/`, `server/internal/core/logger.go`, `server/internal/global/global.go`, `server/cmd/server/main.go`
- **Dependencies added**: ~7 OpenTelemetry packages + 1 Prometheus client package (~200KB total, no transitive dependency conflicts)
- **APIs added**: New public `/metrics` endpoint (no auth required, standard for Prometheus scraping)
- **Breaking changes**: None, all existing functionality remains 100% compatible
- **External systems required**: Jaeger instance (optional, traces are only exported if configured in `config.yaml`)
