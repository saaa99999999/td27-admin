## Why

The TD27 Admin backend has accumulated security and code quality issues: MD5 password hashing (weak), an unvalidated file download path, two goroutine bugs (leak + closure capture), a linear-scan CORS whitelist, full request body buffering in operation logs, direct `global.TD27_DB` access in service layers, and ~7 locations of dead code. These need fixing to meet production security standards and improve maintainability.

## What Changes

- Replace MD5 password hashing with bcrypt across all 4 call sites (login, user create, password change, password verify)
- Add path traversal sanitization to file download endpoint
- Fix goroutine leak in data_permission.go: use TTL-based cache instead of `go func() { time.Sleep; delete }()` pattern
- Fix goroutine closure bug in role.go: shadow `err` variable in goroutine
- Optimize CORS whitelist check: replace linear slice scan with map lookup
- Fix operation log memory buffering: truncate large request bodies, use `url.Values` for GET params instead of `json.Marshal`
- Replace direct `global.TD27_DB` access in login.go, button.go, dashboard.go with repository-pattern injection
- Remove ~7 dead code locations (commented blocks, unused types, dead functions)
- Update/remove `pkg/md5.go` and its test file

## Capabilities

### New Capabilities
- `security-code-cleanup`: Security hardening, goroutine safety fixes, performance optimizations, and dead code removal across the backend

### Modified Capabilities
- None: no spec-level requirements change, only implementation improvements

## Impact

- **Security**: MD5→bcrypt eliminates password hash cracking risk; path traversal sanitization closes directory escape vector
- **Reliability**: Goroutine leak fixed; closure bug fixed; operation log no longer buffers unbounded request bodies
- **Performance**: CORS check O(n)→O(1); no more per-request goroutine for cache expiry; reduced memory pressure from body buffering
- **Affected code**: `server/internal/pkg/md5.go`, `server/internal/pkg/md5_test.go`, `server/internal/api/sysTool/file.go`, `server/internal/middleware/cors.go`, `server/internal/middleware/operation_log.go`, `server/internal/middleware/data_permission.go`, `server/internal/service/sysManagement/` (login.go, role.go, button.go, data_permission.go), `server/internal/service/sysMonitor/dashboard.go`, `server/internal/model/sysManagement/` (permission_dto.go, permission_model.go, role_repository.go), `server/internal/model/sysMonitor/operation_log_repository.go`, `server/internal/initialize/router.go`, `server/internal/initialize/server.go`
- **Dependencies added**: `golang.org/x/crypto` (bcrypt) — already an indirect dep, just needs explicit import
- **Breaking changes**: None — all API contracts remain identical
