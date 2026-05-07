# Add Prometheus /metrics + Jaeger Distributed Tracing

## TL;DR
> **Objective**: Add Prometheus metrics endpoint with RED metrics and OpenTelemetry-based Jaeger distributed tracing with trace ID injection into existing slog logs
>
> **Deliverables**:
> - New `/metrics` endpoint with HTTP + Go runtime metrics
> - OTel middleware for HTTP request tracing → Jaeger UI
> - Trace ID/span ID injection into all existing slog log records
> - GORM query tracing (span per SQL query)
> - Config-driven enable/disable (tracing disabled by default)
>
> **Estimated Effort**: Medium (~4-6 hours)
> **Parallel Execution**: YES - 3 waves

## Context

### Original Request
"Add Prometheus /metrics and Jaeger distributed tracing to the TD27 Admin backend"

### Research Summary
**Current state**:
- ✅ Structured `slog` logging with per-level log files and rotation
- ✅ HTTP access logging with status-based routing
- ✅ Async DB-stored operation audit logs
- ❌ No metrics collection or Prometheus endpoint
- ❌ No distributed tracing or trace correlation
- ❌ No trace IDs in log records

### Key Technical Decisions
- Use `prometheus/client_golang` directly (no third-party Gin wrapper) — matches existing custom middleware pattern
- Use OpenTelemetry with direct OTLP gRPC export to Jaeger (no collector needed for dev, Jaeger >= 1.35 accepts OTLP natively)
- Inject trace IDs into existing `multiHandler` in core logger (5-line additive change) — zero touch on all existing `global.TD27_LOG` calls
- OTel middleware must be FIRST in Gin chain (before access log) so spans exist when access log fires
- Both features optional via config.yaml — disabled = zero overhead

## Work Objectives

### Core Objective
Implement production-grade metrics and distributed tracing with zero breaking changes

### Concrete Deliverables
- `server/internal/middleware/prometheus.go` — RED metrics middleware
- `server/internal/middleware/otel.go` — OTel Gin middleware wrapper
- `server/internal/initialize/otel.go` — Tracer provider init
- `server/configs/observability.go` — Config struct
- Modified: `go.mod`, `global.go`, `server.go`, `main.go`, `router.go`, `logger.go`, `viper.go`, `gorm.go`, `config.go`, `config.yaml`

### Definition of Done
- [ ] `/metrics` endpoint returns valid Prometheus metrics (test with `curl http://localhost:8888/metrics`)
- [ ] `/health`, `/metrics`, `/swagger/*` excluded from metrics collection
- [ ] Jaeger UI shows traces for API requests (test with login/user list/dashboard)
- [ ] All log lines include `trace_id` and `span_id` when tracing enabled
- [ ] GORM queries appear as child spans in Jaeger traces
- [ ] Disabling features in config.yaml removes all overhead

## Verification Strategy

> **No human test confirmation required** — all verification is agent-executed.

### QA Tooling
- **Backend metrics**: `curl http://localhost:8888/metrics` → scrape and parse
- **Traces verification**: Query Jaeger API at `http://localhost:16686/api/traces`
- **Log verification**: Read log file output, grep for `trace_id`
- **Build**: `go build ./...` must pass with zero errors

## Execution Strategy

### Wave 1 — Config & Dependencies (Parallel: 4 tasks)
| Task | Depends On |
|------|-----------|
| 1. Add go.mod dependencies | - |
| 2. Create observability.go config struct | - |
| 3. Add config.yaml observability section | 2 |
| 4. Wire config into Server struct + validation | 2, 3 |

### Wave 2 — Prometheus Metrics (Parallel: 1 task)
| Task | Depends On |
|------|-----------|
| 5. Create prometheus.go middleware + register in router.go | 1, 4 |

### Wave 3 — Jaeger Tracing (Parallel: 4 tasks)
| Task | Depends On |
|------|-----------|
| 6. Create otel.go initializer | 1, 4 |
| 7. Add TD27_TP global + tracer shutdown | 4 |
| 8. Wire InitTracerProvider in main.go + server.go | 6, 7 |
| 9. Create otel.go middleware + register in router.go | 7, 8 |

### Wave 4 — Log & DB Integration (Parallel: 2 tasks)
| Task | Depends On |
|------|-----------|
| 10. Add trace ID injection to multiHandler.Handle in logger.go | 8 |
| 11. Add OTel GORM plugin in gorm.go | 6, 8 |

### Wave FINAL — Verification
| Task | Depends On |
|------|-----------|
| F1. go mod tidy + build | All |
| F2. Verify /metrics endpoint | All |
| F3. Verify traces in Jaeger UI | All |
| F4. Verify trace IDs in logs | All |

---

## TODOs

- [ ] 1. Dependencies — Add Prometheus and OTel to go.mod

  **What to do**:
  - Add to `server/go.mod` require block:
    - `github.com/prometheus/client_golang v1.20.0`
    - `go.opentelemetry.io/otel v1.43.0`
    - `go.opentelemetry.io/otel/sdk v1.43.0`
    - `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0`
    - `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.68.0`
    - `go.opentelemetry.io/contrib/instrumentation/gorm.io/gorm v0.68.0`
    - `google.golang.org/grpc v1.71.0`
    - `google.golang.org/grpc/credentials/insecure v1.71.0`
  - Run `go mod tidy`

  **QA Scenarios**:
  ```
  Scenario: Dependencies resolve
    Tool: Bash
    Steps:
      1. cd server && go mod tidy
    Expected Result: No errors, all modules downloaded
    Failure Indicators: Permission errors, 404s, timeout
  ```

  **Parallelization**: Wave 1, group with 2, no dependencies
  **Blocks**: Wave 2, Wave 3

- [ ] 2. Config struct — Create observability.go

  **What to do**:
  - Create `server/configs/observability.go`:
  ```go
  package configs

  type Observability struct {
      Prometheus Prometheus `mapstructure:"prometheus" json:"prometheus" yaml:"prometheus"`
      Otel       Otel       `mapstructure:"otel" json:"otel" yaml:"otel"`
  }

  type Prometheus struct {
      Enabled     bool   `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
      MetricsPath string `mapstructure:"metrics-path" json:"metrics-path" yaml:"metrics-path"`
  }

  type Otel struct {
      Enabled      bool    `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
      Endpoint     string  `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
      ServiceName  string  `mapstructure:"service-name" json:"service-name" yaml:"service-name"`
      SamplingRate float64 `mapstructure:"sampling-rate" json:"sampling-rate" yaml:"sampling-rate"`
  }
  ```

  **Parallelization**: Wave 1, group with 1
  **Blocks**: 3, 4

- [ ] 3. Config YAML — Add observability section

  **What to do**:
  - Append to `server/configs/config.yaml`:
  ```yaml
  observability:
    prometheus:
      enabled: true
      metrics-path: /metrics
    otel:
      enabled: false
      endpoint: localhost:4317
      service-name: td27-admin
      sampling-rate: 1.0
  ```

  **Parallelization**: Wave 1, blocked by 2
  **Blocks**: 4

- [ ] 4. Root struct + validation — Wire observability into server

  **What to do**:
  - In `server/configs/config.go`: Add `Observability Observability` field to `Server` struct
  - In `server/internal/core/viper.go`: Add validation in `validateConfig()`:
    - If Prometheus.Enabled && MetricsPath == "" → error
    - If Otel.Enabled && Endpoint == "" → error
    - If Otel.Enabled && ServiceName == "" → error
    - If SamplingRate < 0 || > 1 → error

  **Parallelization**: Wave 1, blocked by 2, 3
  **Blocks**: Wave 2, Wave 3, Wave 4

- [ ] 5. Prometheus middleware + endpoint

  **What to do**:
  - Create `server/internal/middleware/prometheus.go` with:
    - `http_requests_total` counter (method, path, status labels)
    - `http_request_duration_seconds` histogram (method, path, 15 buckets)
    - `http_requests_in_flight` gauge (method label)
    - Excluded routes: `/health`, `/metrics`, `/swagger/*`
    - Use `c.FullPath()` (not raw URL) to prevent cardinality explosion
  - In `server/internal/initialize/router.go`:
    - Add import `"github.com/prometheus/client_golang/prometheus/promhttp"`
    - Register PrometheusMiddleware after GinLogger/GinRecovery
    - Register `/metrics` endpoint on root router: `r.GET(metricsPath, gin.WrapH(promhttp.Handler()))`
    - Both conditional on `config.Observability.Prometheus.Enabled`

  **Parallelization**: Wave 2, blocked by 1, 4
  **Blocks**: F1-F4

- [ ] 6. OTel tracer provider initializer

  **What to do**:
  - Create `server/internal/initialize/otel.go`:
    - `InitTracerProvider()` function
    - Create gRPC connection to OTelCfg.Endpoint with `credentials/insecure`
    - Create `otlptracegrpc` exporter
    - Create resource with `ServiceName` and `DeploymentEnvironment` attributes
    - Create `sdktrace.TracerProvider` with batcher and sampler
    - Set global `otel.SetTracerProvider(tp)` and `otel.SetTextMapPropagator(TraceContext + Baggage)`
  - If not enabled, log info and return nil (no error)
  - Use `attribute.String("service.name", ...)` and `attribute.String("deployment.environment", ...)`  instead of semconv imports to avoid dependency version issues

  **Parallelization**: Wave 3, blocked by 1, 4
  **Blocks**: 8, 9, 11

- [ ] 7. Global var + graceful shutdown

  **What to do**:
  - In `server/internal/global/global.go`:
    - Add import `"go.opentelemetry.io/otel/sdk/trace"`
    - Add `TD27_TP *trace.TracerProvider` to var block
  - In `server/internal/initialize/server.go`:
    - In `RunServer`, after `async.GetAsyncLogger().Stop()`: check `global.TD27_TP != nil` and call `Shutdown(ctx)`

  **Parallelization**: Wave 3, blocked by 4
  **Blocks**: 8

- [ ] 8. Wire tracer into main.go

  **What to do**:
  - In `server/cmd/server/main.go`:
    - Add `import ("context", "time")` to imports
    - After `global.TD27_LOG = core.Logger()`:
    ```
    tp, err := initialize.InitTracerProvider()
    if err != nil { log error but continue }
    if tp != nil {
        defer func() { ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second); defer cancel(); tp.Shutdown(ctx) }()
        global.TD27_TP = tp
    }
    ```

  **Parallelization**: Wave 3, blocked by 6, 7
  **Blocks**: 9, 10

- [ ] 9. OTel middleware + register in router

  **What to do**:
  - Create `server/internal/middleware/otel.go`:
    - `OTelMiddleware()` wraps `otelgin.Middleware(serviceName, otelgin.WithFilter(func(r *http.Request) bool { skip health/metrics/swagger }))`
    - If OTel disabled, return no-op middleware (`func(c *gin.Context) { c.Next() }`)
  - In `router.go`:
    - Add import `"server/internal/middleware"`
    - Register `middleware.OTelMiddleware()` as **FIRST** middleware (before GinLogger), conditional on config

  **Parallelization**: Wave 3, blocked by 7, 8
  **Blocks**: F2, F3

- [ ] 10. Trace ID injection into slog

  **What to do**:
  - In `server/internal/core/logger.go`:
    - Add import `"go.opentelemetry.io/otel/trace"`
    - In `multiHandler.Handle()`, at the very start before static attrs injection:
    ```
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        sc := span.SpanContext()
        r2 := r.Clone()
        r2.AddAttrs(slog.String("trace_id", sc.TraceID().String()), slog.String("span_id", sc.SpanID().String()))
        r = r2
    }
    ```

  **Parallelization**: Wave 4, blocked by 8
  **Blocks**: F3, F4

- [ ] 11. GORM OTel plugin

  **What to do**:
  - In `server/internal/initialize/gorm.go`:
    - Add import `"go.opentelemetry.io/contrib/instrumentation/gorm.io/gorm/otelgorm"`
    - In `Gorm()`, after `gorm.Open` success and before config setup:
    ```
    if global.TD27_CONFIG.Observability.Otel.Enabled {
        if err := db.Use(otelgorm.NewPlugin()); err != nil {
            global.TD27_LOG.Error("failed to register OTel GORM plugin", "error", err)
        }
    }
    ```
  - Note: If `go mod tidy` has issues with this dependency, make it optional — skip if download fails

  **Parallelization**: Wave 4, blocked by 6, 8
  **Blocks**: F3

---

## Final Verification Wave

- [ ] F1. **Build Verification** — `oracle`
  - Run `go mod tidy` and `go build ./...`
  - Both must pass with zero errors
  - Verify no new lint issues

- [ ] F2. **Metrics Verification** — `oracle`
  - Start Jaeger: `docker run -d --name jaeger -e COLLECTOR_OTLP_ENABLED=true -p 16686:16686 -p 4317:4317 jaegertracing/all-in-one:latest`
  - Start server: `make run`
  - `curl http://localhost:8888/metrics` → must return text with `td27_http_requests_total`, `go_goroutines`, etc.
  - Must NOT have `/health` or `/swagger/` in metric labels

- [ ] F3. **Traces Verification** — `oracle`
  - Enable tracing in config.yaml: `otel.enabled: true`
  - Make API requests (login, list users, visit dashboard)
  - Query Jaeger API: `curl http://localhost:16686/api/traces?service=td27-admin` → must return traces
  - Log output must contain `trace_id` and `span_id` attributes

- [ ] F4. **Rollback Verification** — `oracle`
  - Set `otel.enabled: false` and `prometheus.enabled: false` in config
  - Restart server, verify no errors, all existing functionality intact

## Success Criteria

- `/metrics` returns valid Prometheus output
- Jaeger UI shows complete traces with DB query spans
- All log lines have `trace_id`/`span_id` when tracing enabled
- Disabling observability in config = zero impact
- All existing APIs, auth, and business logic unaffected
