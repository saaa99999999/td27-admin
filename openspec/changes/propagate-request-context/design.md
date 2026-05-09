## Context

The Gin framework attaches a `context.Context` to every request via `c.Request.Context()`. This context:
- Is cancelled when the client disconnects
- Carries deadline information from upstream (e.g. reverse proxy timeout)
- Can carry trace/span IDs for observability
- Supports `WithValue` propagation for request-scoped data

Currently every service struct stores `context.Background()` — a never-cancelled, empty context — making all DB queries and cache operations immune to cancellation. The stored ctx pattern also obscures which context a method actually uses.

The fix is straightforward: remove the stored `ctx` field and thread `ctx context.Context` as the first parameter of every public service method. Callers in the request path use `c.Request.Context()`; callers outside the request path supply their own.

## Goals / Non-Goals

**Goals:**
- Every Gin API handler passes `c.Request.Context()` to every service method it calls
- Every service method accepts `ctx context.Context` as its first parameter (instead of reading from struct field)
- The stored `ctx` field is removed from all service structs
- `context.Background()` is eliminated from JwtService (6 methods, 7 call sites) and `pkg/rbac/permission_cache.go` (6 methods)
- Server timeouts are tightened in `server.go`

**Non-Goals:**
- No changes to repository/model layer signatures (they already accept `ctx context.Context`)
- No changes to `pkg/cache.PGCache` methods (already accept ctx)
- No changes to background/cron job code paths (they correctly use their own context)
- No introduction of new tracing or timeout middleware (that is a separate change)
- No change to CasbinService (it wraps a singleton enforcer — no DB query cancellation benefit)
- No change to LogRegService.Login, DashboardService methods, DataPermissionService's non-ctx methods — they use `global.TD27_DB` directly rather than a stored ctx field

## Decisions

1. **`ctx context.Context` as first parameter, not stored field** — Go convention. Makes the context flow explicit at the call site. Callers can pass different contexts per call. Eliminates the risk of stale ctx or wrong-ctx bugs.

2. **Not using `gin.Context` in service layer** — Services should not depend on Gin. Accepting `context.Context` keeps the service layer framework-agnostic and testable with `context.Background()`.

3. **JwtService: thread ctx through all 6 public methods** — Even though JWT cache lookups are fast, they go through `PGCache` which makes DB queries with `db.WithContext(ctx)`. Without a real context, client disconnects during login/logout can't cancel those queries.

4. **Server timeout values** — 30s read timeout matches Nginx/cloud LB defaults. 120s write timeout accommodates large file uploads and slow responses (the previous 120s read timeout could interfere with SSE or long-polling). IdleTimeout=120s follows HTTP keep-alive best practices. ReadHeaderTimeout=20s prevents slow-header attacks.

5. **PermissionCache: ctx as first param** — Same rationale as services. These methods call `PGCache.Get/Set/Del` which all accept `ctx context.Context`. Passing `context.Background()` defeats cancellation and tracing.

## Risks / Trade-offs

- [Breaking change] All in-tree callers will be updated, but out-of-tree consumers of service methods (if any) will need to pass a context. Mitigation: this is an internal package, not a public library.
- [Effort] ~80 method signatures change across ~15 files. High line count but mechanical, low-risk changes. Each file change is formulaic: remove ctx field, add ctx param, update constructor.
- [Missed callers] Cron job triggers, scheduler, or test files may still pass context.Background(). Mitigation: task 7 explicitly verifies no remaining `context.Background()` in service layer.
- [Middleware impact] `middleware/jwt.go` creates `jwtService` and `serviceTokenService` as package-level singletons and calls them without `c.Request.Context()`. This requires adding ctx param passing in the middleware chain.
