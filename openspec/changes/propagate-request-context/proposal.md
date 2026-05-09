## Why

Gin provides `c.Request.Context()` with built-in cancellation on client disconnect and support for context propagation (timeouts, tracing). Currently zero API handlers pass this context to services — all 17 service constructors hardcode `context.Background()` and 7 JWT methods call `context.Background()` directly. Client disconnects leave DB queries running to completion; trace context is invisible across layers.

## What Changes

- Remove stored `ctx` field from all 14 service structs that currently use it
- Change every service method to accept `ctx context.Context` as its first parameter
- Add `ctx context.Context` parameter to JwtService methods that currently create `context.Background()` locally
- Add `ctx context.Context` parameter to all 6 `PermissionCache` methods
- Update all 15 API handlers to pass `c.Request.Context()` when calling service methods
- Tighten server timeouts in `server.go`: ReadTimeout from 120s → 30s, WriteTimeout from 120s → 60s, add IdleTimeout=120s and ReadHeaderTimeout=20s
- **BREAKING**: All service method signatures change — callers outside the request cycle (cron jobs, tests, seeds) must supply their own context

## Capabilities

### New Capabilities
- `context-propagation`: Request-scoped context from Gin handlers through service and repository layers, enabling client-disconnect cancellation, timeout propagation, and trace context continuity

### Modified Capabilities
<!-- No existing specs are modified — this is a pure implementation-level refactor -->

## Impact

- 17 service constructors across 3 packages (`sysManagement`, `sysMonitor`, `sysTool`) — remove stored `ctx`
- ~80 service method signatures — add `ctx context.Context` first param
- 6 JwtService methods — remove `context.Background()`, accept ctx param
- 6 PermissionCache methods — add `ctx context.Context` param
- 15 API handler files — pass `c.Request.Context()` to every service call
- 1 middleware file (`middleware/jwt.go`) — pass `c.Request.Context()` to JWT service calls
- Cron job runner in `pkg/cron/job.go` — uses `context.Background()` which is acceptable for background jobs
- `server.go` — update ReadTimeout, WriteTimeout, add IdleTimeout, add ReadHeaderTimeout
