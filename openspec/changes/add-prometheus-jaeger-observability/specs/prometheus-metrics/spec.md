## ADDED Requirements

### Requirement: Prometheus Metrics Endpoint
The system SHALL expose a public `/metrics` endpoint on the same port as the API server, serving Prometheus-formatted metrics. The endpoint SHALL NOT require authentication.

#### Scenario: Metrics endpoint accessible
- **WHEN** GET request is sent to `/metrics`
- **THEN** HTTP 200 response is returned with Prometheus text format metrics
- **AND** no JWT token or authentication is required

### Requirement: Standard HTTP RED Metrics
The system SHALL collect and expose standard HTTP RED (Rate/Error/Duration) metrics for all API endpoints.

#### Scenario: Request metrics collected
- **WHEN** any API request is handled
- **THEN** `http_requests_total` counter is incremented with labels: `method` (HTTP method), `path` (registered route path, not raw URL), `status` (HTTP status code)
- **AND** `http_request_duration_seconds` histogram records request latency with labels: `method`, `path`
- **AND** `http_requests_in_flight` gauge tracks concurrent requests with label: `method`

#### Scenario: Excluded routes not measured
- **WHEN** request is sent to `/health`, `/metrics`, or any `/swagger/*` path
- **THEN** no HTTP metrics are recorded for the request

### Requirement: Go Runtime and Process Metrics
The system SHALL automatically expose standard Go runtime and process metrics via the `/metrics` endpoint.

#### Scenario: Runtime metrics exposed
- **WHEN** `/metrics` endpoint is queried
- **THEN** metrics include: `go_goroutines`, `go_memstats_*`, `go_gc_duration_seconds`, `process_cpu_seconds_total`, `process_resident_memory_bytes`, `process_open_fds`
