## Context

The TD27 Admin backend currently uses `PGCache` (PostgreSQL-backed) for all caching operations including JWT token storage, RBAC permissions, and user cache. Every cache read/write is a full SQL query — `SELECT ... FROM sys_tool_cache` or `INSERT/UPDATE`. This adds 1-5ms per cache operation on a local network and more under load. Since JWT validation happens on every authenticated request (and internal API calls do cache lookups for permissions), the latency compounds significantly.

The project already has Redis infrastructure available (README lists it as a tech stack, docker-compose/ directory exists), but no Redis instance is running and no Go Redis client is imported. The existing observability change (add-prometheus-jaeger) established patterns for adding new external dependencies — config struct in `server/configs/`, client init in `server/internal/core/`, global var in `server/internal/global/global.go`, and wiring in `server/cmd/server/main.go`. This design follows the same pattern.

## Goals / Non-Goals

**Goals:**
1. Implement a `RedisCache` adapter that implements the same operational interface as `PGCache` (Get, Set, Del, Exists, TTL, ListKeysByPrefix, DelByUsername)
2. Extract a `Cache` interface in `server/internal/pkg/cache/` that both `PGCache` and `RedisCache` implement, enabling consumers to be backend-agnostic
3. Migrate `JwtService` from PGCache to RedisCache (highest payoff — every request)
4. Migrate `PermissionCache` in `pkg/rbac/` from PGCache to RedisCache
5. Add Redis-based caching for dashboard statistics (6 COUNT queries, 30s TTL)
6. Add Redis-based caching for dictionary data (read-heavy, write-rare)
7. Add Redis-based caching for department tree (static data, infrequently changed)
8. Add Redis-based caching for per-role menu tree (read-heavy, changes only on role config)
9. Add Redis-based caching for button permissions (per-role/page, checked on every page load)
10. Add Redis-based caching for service tokens (third-party auth, every request)
11. Keep PGCache as a zero-dependency fallback when Redis is disabled
12. Add Redis service to docker-compose.yml
13. Zero breaking changes to existing API contracts, business logic, or data

**Non-Goals:**
1. Replace PGCache entirely — it remains as fallback and for backward compatibility
2. Remove the `sys_tool_cache` database table — it stays for PGCache fallback
3. Add Redis cluster/sentinel support — single-instance Redis for now, expandable later
4. Implement distributed cache invalidation across multiple server instances — out of scope
5. Change the existing `JwtService` API or `PermissionCache` API — only the backing store changes

## Decisions

### 1. go-redis vs redigo
**Decision**: Use `github.com/redis/go-redis/v9` (go-redis)
**Rationale**:
- Actively maintained, widely adopted, idiomatic Go API with full context support
- Built-in connection pooling, retry, and health-check
- Supports all Redis commands we need (GET, SET, DEL, EXISTS, TTL, KEYS/SCAN, EXPIRE) with type-safe methods
- `redigo` requires manual connection management and has a lower-level API
- go-redis is the community standard for new Go projects
**Alternative considered**: `redigo` — lower-level, requires more boilerplate for connection pooling, less idiomatic

### 2. Keep PGCache as Fallback
**Decision**: Keep PGCache as a configurable fallback. When `redis.enabled: false`, all cache operations fall back to PGCache.
**Rationale**:
- Zero-config dev experience: Redis is optional, not required
- No breaking changes: existing deployments without Redis continue to work
- Enables gradual rollout: can enable Redis per environment
- PGCache code already exists and is tested
- The `Cache` interface makes the backend transparent to consumers

### 3. TTL Strategy

| Cache Type | TTL | Rationale |
|---|---|---|
| JWT tokens | Same as JWT expiry (config.yaml `jwt.expires-time`, default 24h) | Token must remain valid until its natural expiry |
| RBAC permissions | 30 minutes (existing) | Permissions change infrequently, cache invalidation on role update |
| Dashboard stats | 30 seconds | Stale stats acceptable; avoids 6 COUNT queries per page load |
| Dictionary data | 1 hour | Rarely changes, admin-triggered cache clear on update |
| Department tree | 1 hour | Changes infrequently, cache clear on dept CRUD |
| Menu tree (per-role) | 1 hour | Changes on menu/role config, cache clear on menu/role update |
| Button permissions (per-role/page) | 1 hour | Changes on button/role config, cache clear on button/role update |
| Service tokens | 15 minutes | Token validity checked against DB on auth; cache accelerates repeated checks |

### 4. Cache Warming Strategy
**Decision**: No aggressive cache warming on startup. Caches are populated on first access (lazy loading). Cache invalidation happens synchronously on write operations (menu update, role update, dept update, dict update, button update).
**Rationale**:
- Most caches are per-role or per-user — warming all possible combinations at startup is wasteful
- Dashboard stats are cheap to compute on first access (lazy with 30s TTL)
- Dictionary, department, and menu data are read by all users — first request after startup triggers a warm, subsequent requests are cached
- Service tokens are authenticated per-request — cache hit rate goes to ~100% after first access per token
- Cache invalidation on write keeps stale windows minimal

### 5. Key Namespace Convention
**Decision**: Redis key prefixes follow existing PGCache conventions with a `td27:` namespace prefix for Redis key scoping.
```
td27:jwt:user:{username}          # JWT tokens
td27:jwt:user_tokens:{username}:{tokenID}
td27:jwt:user_single_token:{username}
td27:jwt:user_cache:{userID}
td27:rbac:user:perm:{userID}
td27:rbac:role:perm:{roleID}
td27:dashboard:stats              # Dashboard statistics
td27:dict:{dictID}                # Dictionary data
td27:dept:tree                    # Department tree (raw list)
td27:menu:role:{roleID}           # Per-role menu tree
td27:button:page:{pagePath}       # Buttons by page path
td27:button:user:{roleIDs}        # Per-role button list
td27:service_token:{tokenHash}    # Service token validation
```

### 6. Cache Invalidation Strategy
**Decision**: Synchronous cache clear on write operations. Write-heavy operations (menu CRUD, role CRUD, dept CRUD, dict CRUD, button CRUD) evict affected cache keys. No distributed invalidation protocol needed for single-instance deployment.
- Role update → evict `td27:rbac:user:perm:*`, `td27:rbac:role:perm:*`, `td27:menu:role:{roleID}`, `td27:button:user:{roleIDs}`
- Menu update → evict all `td27:menu:role:*`
- Dept update → evict `td27:dept:tree`
- Dict update → evict `td27:dict:{dictID}`
- Button update → evict all `td27:button:*`
- Token invalidation → evict specific `td27:jwt:*` key

### 7. Redis Client Initialization Location
**Decision**: New `core/redis.go` following the pattern of `core/viper.go`. Global `TD27_REDIS` in `global/global.go`. Called from `main.go` after config init.
**Rationale**:
- Consistent with existing patterns (Viper → Logger → Tracer → Gorm)
- Config-driven: Redis host/port/db from `config.yaml`
- PgSQL-like: if Redis is disabled, `TD27_REDIS` stays nil and cache falls back to PGCache

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Redis becomes a single point of failure | Configurable fallback to PGCache; app degrades gracefully without Redis |
| Cache staleness after write | Synchronous cache eviction on all write operations |
| Memory pressure from unbounded cache growth | TTL-based expiry on all keys; Redis eviction policy `allkeys-lru` recommended |
| Connection pool exhaustion | go-redis default pool (10 connections per CPU) is sufficient for single-instance; configurable via `config.yaml` |
| Key collisions between environments | `td27:` namespace prefix scopes keys; production recommends separate Redis DB or instance |
| Race condition on cache warm | Tolerated — first request after invalidation may see stale data for <1s; use TTL as safety net |

## Migration Plan

1. Add `redis` config struct and `config.yaml` section
2. Add go-redis dependency
3. Create `core/redis.go` for Redis client initialization
4. Create `Cache` interface in `server/internal/pkg/cache/` with both PGCache and RedisCache implementations
5. Wire Redis init into `main.go` after config, before DB
6. Migrate `JwtService` to accept `Cache` interface instead of `*PGCache`
7. Migrate `PermissionCache` to use `Cache` interface
8. Add Redis caching layer to each consumer (dashboard, dict, dept, menu, button, service_token)
9. Add Redis service to `docker-compose/compose.yml`
10. Test with Redis enabled and disabled (fallback to PGCache)

## Open Questions

- Should Redis be enabled by default in development config? (Yes, since docker-compose will provide it, but PGCache fallback means devs without Redis won't break)
- What Redis eviction policy should be documented? (`allkeys-lru` recommended for cache use case)
- Should the `TTL` and `ListKeysByPrefix` methods be supported in RedisCache? (Yes — `TTL` maps to Redis TTL command, `ListKeysByPrefix` maps to Redis SCAN with pattern)
