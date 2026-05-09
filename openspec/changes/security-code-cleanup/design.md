## Context

The TD27 Admin backend currently uses MD5 for password hashing (via `server/internal/pkg/md5.go`), which is considered cryptographically weak for credential storage. The file download endpoint (`api/sysTool/file.go:101-116`) accepts a `name` query parameter and passes it directly to `fmt.Sprintf` + `os.Stat` without sanitization, allowing path traversal.

The goroutine in `data_permission.go:82-85` fires a background goroutine with `time.Sleep(5*time.Minute)` for cache expiry — this leaks if the service is stopped before the sleep completes. The goroutine in `role.go:76-80` captures the outer `err` variable, causing a data race where the goroutine reads `err` which may have been reassigned by the time it executes.

CORS whitelist (`middleware/cors.go:68-76`) uses a linear scan over a slice for every OPTIONS/preflight request. Operation log middleware (`middleware/operation_log.go:46-66`) buffers the entire request body in memory.

Three services (`login.go`, `button.go`, `dashboard.go`) access `global.TD27_DB` directly instead of using repository injection. Approximately 7 locations contain dead code (commented blocks, unused types, dead functions).

All changes are purely internal — no API contracts, response formats, or user-facing behavior changes.

## Goals / Non-Goals

**Goals:**
1. Replace MD5 with bcrypt for password hashing (all 4 call sites + remove md5.go)
2. Sanitize file download path to prevent directory traversal
3. Fix goroutine leak in data_permission.go by replacing `time.Sleep`-based cache expiry with a TTL-aware map
4. Fix goroutine closure bug in role.go by shadowing the `err` variable
5. Optimize CORS whitelist from slice linear scan to map lookup
6. Fix operation log body buffering: truncate large bodies, use `url.Values` for GET params
7. Replace direct `global.TD27_DB` access in services with constructor-injected `*gorm.DB`
8. Remove all dead code locations
9. All existing tests continue to pass; no API contract changes

**Non-Goals:**
1. Replace the entire auth system — only password hashing mechanism changes
2. Add new features or API endpoints
3. Change database schema
4. Refactor the whole service layer to use dependency injection — only the 3 worst offenders

## Decisions

### 1. Bcrypt password hashing
**Decision**: Use `golang.org/x/crypto/bcrypt` with default cost (10)
**Rationale**: Industry-standard replacement for MD5, already an indirect dependency via `golang.org/x/crypto`. Default cost of 10 balances security and latency (~100ms per hash on modern hardware). No configuration needed.
**Alternative considered**: Argon2id — more secure but requires a new dependency and more complex setup. Bcrypt is simpler and sufficient for this admin dashboard.

### 2. File download sanitization
**Decision**: Use `path/filepath.Clean` plus `strings.Contains` to verify the resolved path stays within the upload directory
**Rationale**: Simple, well-understood approach. `filepath.Clean` normalizes `../` sequences, then checking the result prefix ensures no directory escape.
**Alternative considered**: Rejecting specific patterns (`../`, `..\\`) — fragile, easy to bypass with encoding tricks. Whitelist-based approach is more robust.

### 3. TTL cache for data permissions
**Decision**: Replace `sync.Map` + goroutine-based expiry with a simple TTL wrapper using `time.Now()` timestamps per entry, reusing the existing `sync.Map` structure
**Rationale**: Minimal change — no new dependency, no goroutine lifecycle management, no context cancellation. Each cached entry records its expiry time; reads check and skip expired entries. A background goroutine is not needed because stale entries are lazily evicted on read.
**Alternative considered**: `go-cache` library — unnecessary dependency for this simple use case. Context-based cancellation — over-engineered for a 5-minute cache TTL.

### 4. CORS whitelist optimization
**Decision**: Build a `map[string]*configs.CORSWhitelist` once at middleware initialization
**Rationale**: O(1) lookup per request. The whitelist is read from config at startup and doesn't change at runtime, so building the map once is safe and cheap.
**Alternative considered**: Keeping slice + binary search — not worth the complexity since whitelists are small (<100 entries typically).

### 5. Operation log body buffering
**Decision**: Truncate request body at 10KB; replace `json.Marshal` of GET query params map with direct `url.Values` string capture
**Rationale**: Prevents memory exhaustion from large uploads. GET params are already in `url.Values` format — marshaling to JSON adds unnecessary CPU/memory overhead.
**Alternative considered**: Streaming body to disk — over-engineered for an admin audit log. 10KB limit handles normal CRUD payloads.

### 6. Direct global.TD27_DB replacement
**Decision**: Add `*gorm.DB` field to service structs, inject via constructor (matching existing pattern in `data_permission.go`, `user.go`, `dept.go` etc.)
**Rationale**: Consistent with the codebase's existing pattern. The `dashboard.go`, `login.go`, and `button.go` services are outliers that bypass their own repository layer.
**Alternative considered**: Full DI framework — too heavy for this project; the existing manual constructor injection works fine.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Bcrypt cost=10 adds ~100ms to login latency | Negligible for an admin dashboard with <1000 concurrent users |
| Existing passwords hashed with MD5 are still in DB | Migration path: on first successful login with MD5 hash, re-hash with bcrypt and update. Clear in requirements. |
| Map-based CORS whitelist may stale if config hot-reloaded | Config hot-reload currently not supported for CORS section; map is built once at middleware init |
| Operation log truncation loses request body data for large uploads | CSV uploads are the only large payloads — their content is not meaningful in audit logs (file reference is) |
| TTL cache reads may return stale data within the 5-minute window | Same window as before — no change in behavior |
| Removing dead code may break imports if symbols are referenced elsewhere | All 7 dead-code locations will be verified with `go build` before removal |

## Migration Plan

1. Deploy bcrypt changes first — existing users can still log in via MD5; `ModifyPasswd` handler re-hashes with bcrypt on password change
2. File download sanitization is additive — cannot break existing downloads
3. Goroutine and CORS fixes are purely internal — no deployment coordination needed
4. Dead code removal runs through `go build` + `go vet` verification
5. Rollback: revert the single commit covering this change; bcrypt hashes remain valid since bcrypt `Verify` works on both old and new systems

## Open Questions
- None — all decisions are captured above
