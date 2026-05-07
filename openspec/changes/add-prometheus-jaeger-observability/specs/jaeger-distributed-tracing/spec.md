## ADDED Requirements

### Requirement: OpenTelemetry Distributed Tracing
The system SHALL support OpenTelemetry distributed tracing with direct OTLP gRPC export to Jaeger. Tracing SHALL be disabled by default and configurable via `config.yaml`.

#### Scenario: Tracing disabled by default
- **WHEN** `otel.enabled` is set to `false` in config.yaml
- **THEN** no traces are generated or exported
- **AND** no additional overhead is added to requests

#### Scenario: Tracing enabled
- **WHEN** `otel.enabled` is set to `true` in config.yaml
- **THEN** all HTTP requests are automatically traced with spans
- **AND** traces are exported to the configured Jaeger OTLP endpoint

### Requirement: Trace ID Injection into Logs
Every `slog` log record in the system SHALL include `trace_id` and `span_id` attributes when a valid trace context is present in the request context.

#### Scenario: Trace IDs in logs
- **WHEN** a request is being traced and a `global.TD27_LOG` call is made with the request context
- **THEN** the log record automatically includes `trace_id` and `span_id` fields matching the current trace
- **AND** no changes to existing logging calls are required

### Requirement: GORM Query Tracing
All GORM database queries SHALL automatically create child spans within the parent request trace when tracing is enabled.

#### Scenario: DB queries traced
- **WHEN** a GORM query is executed with a context that contains a valid trace span
- **THEN** a child span is created for the query
- **AND** span includes attributes: `db.query.text` (SQL query), `db.rows_affected` (number of rows modified/returned), `db.operation` (query type)

### Requirement: W3C Trace Context Propagation
The system SHALL propagate W3C Trace Context headers across requests to support distributed tracing across services.

#### Scenario: Trace context propagated
- **WHEN** an incoming request includes a valid `traceparent` header
- **THEN** the request uses the provided trace ID as the parent trace instead of generating a new one
