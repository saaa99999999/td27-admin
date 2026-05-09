## Context

`GetClaims()` in `claims_utils.go:27` always re-parses the JWT from the `x-token` header, even though `JWTAuth` middleware already parses it and stores the result in `c.Set("claims", ...)` at `middleware/jwt.go:44`. The sibling function `GetUserInfo()` already checks context first — `GetClaims()` was never updated to match.

Casbin's `Enforce` logs every request at Info level (`casbin.go:123`), generating substantial log noise in production.

ServiceToken routes at `service_token.go:18-21` use `GET` for delete and update endpoints, which violates REST semantics and can cause unexpected caching/proxy behavior.

`NewJWT()` at `jwt/jwt.go:23` converts `global.TD27_CONFIG.JWT.SigningKey` to `[]byte` on every allocation. This string-to-bytes conversion runs on every request that parses a token.

## Goals / Non-Goals

**Goals:**
- Eliminate redundant JWT parsing from `GetClaims()`
- Reduce Casbin log volume by using Debug level for per-request enforcement checks
- Fix REST method correctness for ServiceToken mutation endpoints
- Cache the signing key to avoid per-request `[]byte` allocation

**Non-Goals:**
- No changes to JWT token format, storage, or validation logic
- No changes to Casbin authorization model or policy structure
- No API URL restructuring — only HTTP methods change

## Decisions

1. **GetClaims context-first** — Match `GetUserInfo` exactly: check `c.Get("claims")`, fall back to parsing header if absent. This is zero-risk since `JWTAuth` middleware always sets the claims key before reaching handlers.

2. **Info→Debug for Enforce** — The log line is labeled "Enforce debug" and contains per-request debug info. Downgrading to Debug level matches its semantic intent. Error/Warn paths remain at their appropriate levels.

3. **REST method correction** — `DELETE` for delete endpoint, `PUT` for update endpoint. The frontend may need updates to match. URLs remain identical.

4. **Package-level JWT cache** — Replace per-call `NewJWT()` with a package-level cached instance initialized once. The signing key is read from config at startup and never changes at runtime. Alternative considered: caching just the `[]byte` key, but caching the whole `JWT` struct is simpler and equally safe.

## Risks / Trade-offs

- [Low] Frontend must update HTTP methods for ServiceToken delete/update — existing frontend code expects GET. Check `web/src/api/sysTool/` for callers.
- [Low] JWT caching means a config hot-reload won't pick up signing key changes — this is acceptable since JWT key rotation already requires coordinated rollout.
- [None] GetClaims change is additive — checking context first only adds a fast path, never breaks existing behavior.
