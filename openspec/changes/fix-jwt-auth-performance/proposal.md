## Why

Every API request performs 2-3x unnecessary JWT parsing, adding ~50µs overhead per request. Casbin logs at Info level on every API call, flooding log files with debug noise. ServiceToken routes use GET for mutation operations, violating REST conventions. These are low-hanging performance and correctness issues.

## What Changes

- **Fix #1**: Make `GetClaims()` check `c.Get("claims")` before re-parsing JWT from header (match `GetUserInfo` pattern)
- **Fix #2**: Change Casbin `Enforce` debug logging from Info to Debug level
- **Fix #3**: Fix ServiceToken routes: `GET /delete` → `DELETE /delete`, `GET /update` → `PUT /update`
- **Fix #4**: Cache JWT instance to avoid `[]byte` re-allocation of signing key on every `NewJWT()` call

## Capabilities

### New Capabilities
- `jwt-auth-optimization`: Eliminate redundant JWT parsing, fix Casbin log level, correct REST semantics for ServiceToken routes, and cache JWT instance

### Modified Capabilities

None — no spec-level behavior changes, only performance/correctness improvements.

## Impact

- 4 files touched: `claims_utils.go`, `casbin.go` (service), `service_token.go` (router), `jwt/jwt.go`
- No schema changes, no new dependencies
- No API contract changes (URLs stay the same, only HTTP methods change for Delete/Update)
