## Why

PGCache stores everything in PostgreSQL — every cache Get/Set/Del is a full database round-trip. JWT token validation, permission checks, and dashboard stats all hit the DB instead of in-memory cache. The project already lists Redis as part of its tech stack and docker-compose/ can include it, but it remains unused. This adds significant unnecessary latency on every authenticated request and prevents the app from scaling under load.

## What Changes

- New `RedisCache` adapter implementing the same interface as `PGCache`
- New Redis configuration section in `config.yaml` (host, port, password, db, enabled flag)
- Redis client initialization in server/init sequence (after config, before DB)
- Cache abstraction interface so consumers can be backend-agnostic
- Migrate JWT token storage from PGCache to RedisCache (every request hit)
- Migrate RBAC permission caching from PGCache to RedisCache
- Add Redis caching for dashboard statistics (30s TTL, 6 COUNT queries)
- Add Redis caching for dictionary data (read-heavy, write-rare)
- Add Redis caching for department tree (static, infrequently changed)
- Add Redis caching for per-role menu tree (read-heavy, changes only on role config)
- Add Redis caching for button permissions (per-role/page, checked on every page load)
- Add Redis caching for service token authentication (third-party auth, every request)
- Add Redis service to docker-compose.yml
- Zero breaking changes — PGCache remains as a fallback when Redis is disabled

## Capabilities

### New Capabilities
- `redis-caching`: Redis-based caching layer with PGCache fallback, covering JWT tokens, RBAC permissions, dashboard stats, dictionaries, departments, menus, buttons, and service tokens

### Modified Capabilities
- None: All existing functionality remains 100% compatible. PGCache is preserved as an optional fallback.

## Impact

- **Affected code**: `server/configs/` (new Redis config struct), `server/configs/config.yaml` (new redis section), `server/internal/core/` (Redis client init), `server/internal/global/` (Redis client global var), `server/internal/pkg/cache/` (new Redis adapter, new interface), `server/internal/service/sysManagement/jwt.go`, `server/internal/pkg/rbac/permission_cache.go`, `server/internal/service/sysMonitor/dashboard.go`, `server/internal/service/sysManagement/dict.go`, `server/internal/service/sysManagement/dict_detail.go`, `server/internal/service/sysManagement/dept.go`, `server/internal/service/sysManagement/menu.go`, `server/internal/service/sysManagement/button.go`, `server/internal/service/sysTool/service_token.go`, `server/cmd/server/main.go`
- **Dependencies added**: `github.com/redis/go-redis/v9` (go-redis)
- **External systems required**: Redis instance (optional — caching falls back to PGCache when Redis is disabled)
- **Breaking changes**: None — all existing functionality is preserved, PGCache remains as fallback
