## Context

The TD27 Admin backend already has a production-grade structured logging implementation using `log/slog` with per-level log files, HTTP access logging, and DB-stored operation audit logs. The current implementation lacks:
1. Metrics collection for monitoring system health and performance
2. Distributed tracing for debugging end-to-end request flows
3. Correlation between log entries and request traces

This design implements both capabilities following existing codebase patterns, with zero breaking changes to existing functionality.

## Goals / Non-Goals

**Goals:**
1. Implement Prometheus `/metrics` endpoint with standard HTTP and runtime metrics
2. Implement OpenTelemetry distributed tracing with direct Jaeger export (no OTel collector required for development)
3. Inject trace IDs into all existing `slog` log records automatically
4. Instrument GORM queries to create child spans in request traces
5. Follow all existing codebase patterns for config, initialization, and middleware
6. Zero breaking changes to existing business logic, APIs, or logging behavior

**Non-Goals:**
1. Deploy and manage a production observability stack (Grafana, OTel Collector, etc.) - this is out of scope for this change
2. Replace existing logging infrastructure - the current `slog` implementation remains untouched
3. Add frontend observability - this change focuses solely on backend capabilities
4. Modify existing service/repository method signatures for context propagation in one go - this will be done incrementally

## Decisions

### 1. Prometheus Integration Approach
**Decision**: Use official `prometheus/client_golang` directly with custom middleware, no third-party Gin prometheus libraries
**Rationale**:
- Full control over metric definitions, labels, and excluded routes
- Follows existing middleware patterns already used in the codebase (access log, operation log)
- Avoids unnecessary third-party dependencies
- Matches the project's pattern of minimal external dependencies where possible
**Alternative considered**: Using `Depado/ginprom` or `gin-contrib/openmetrics` libraries - rejected because they add unnecessary abstraction and don't follow the project's existing middleware patterns.

### 2. Tracing Implementation Approach
**Decision**: Use OpenTelemetry with direct OTLP gRPC export to Jaeger
**Rationale**:
- Jaeger >= 1.35 supports native OTLP ingestion, eliminating the need for a separate OTel collector for development use
- OpenTelemetry is the industry standard for distributed tracing, with broad ecosystem support
- Direct export simplifies development setup while maintaining compatibility with production collector deployments
- `otelgin` and `otelgorm` official instrumentation libraries provide out-of-the-box support for Gin and GORM
**Alternative considered**: Using deprecated Jaeger exporter - rejected because it is no longer maintained and lacks modern OTel features.

### 3. Trace ID Injection
**Decision**: Inject trace IDs directly into existing `multiHandler` in `core/logger.go`
**Rationale**:
- Zero-touch approach: all existing `global.TD27_LOG` calls automatically get trace IDs without any modifications
- Works across all parts of the application: HTTP requests, cron jobs, async operations, etc.
- Doesn't require wrapping the logger or changing any existing logging patterns
- Aligns with the existing pattern of injecting static attributes (service, env) into all log records
**Alternative considered**: Using `otelslog` bridge - rejected because it would require replacing the existing logger setup, which is unnecessary when we can add trace ID injection to the existing `multiHandler` with 5 lines of code.

### 4. Context Propagation
**Decision**: Incremental refactor of service methods to accept `context.Context` as first parameter
**Rationale**:
- No breaking changes: refactor can be done per service module without affecting existing functionality
- Follows Go standard library patterns for context propagation
- Allows GORM query tracing to work properly by passing the request context to DB calls
- Existing repository layer already accepts context parameters, so only service layer changes are needed
**Alternative considered**: Using Go's `context.WithValue` with a global context - rejected because it doesn't support per-request context propagation.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Metrics cardinality explosion from raw URL paths | Use `c.FullPath()` instead of `c.Request.URL.Path` for metric labels, which returns the registered route pattern (e.g., `/user/:id` instead of `/user/123`) |
| Tracing overhead in production | Tracing is disabled by default, configurable sampling rate will be added in production deployments |
| Performance impact of trace ID injection | The trace ID lookup is a simple context access and string conversion, with negligible overhead |
| Breaking changes during context propagation refactor | Refactor done incrementally per module, with full backward compatibility maintained at all times |
| Prometheus endpoint exposed publicly | Endpoint can be moved to an internal port or protected by network policies in production deployments |

## Migration Plan

1. Deploy the change with tracing disabled by default
2. Verify Prometheus metrics are being collected correctly
3. Enable tracing in development/staging environments to validate trace export
4. Incrementally refactor service methods to accept context parameters to enable GORM query tracing
5. For production, add an OTel collector between the application and Jaeger for sampling, batching, and backpressure handling

## Open Questions
- What sampling rate should be used for production tracing? (Default to 10% for now, configurable via config)
- Should the `/metrics` endpoint be protected by authentication in production? (Default to public for Prometheus scraping, can be changed later)
