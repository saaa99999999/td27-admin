## 1. Add Redis Dependency

- [ ] 1.1 Add `github.com/redis/go-redis/v9` to `server/go.mod` via `go get`
- [ ] 1.2 Run `go mod tidy` to ensure clean dependency graph

## 2. Create Redis Configuration

- [ ] 2.1 Create `server/configs/redis.go` with `Redis` struct (enabled, host, port, password, db, pool-size)
- [ ] 2.2 Add `Redis Redis` field to `Server` struct in `server/configs/config.go`
- [ ] 2.3 Add `redis:` section to `server/configs/config.yaml` with defaults (enabled: true, host: localhost, port: 6379, db: 0, pool-size: 10)
- [ ] 2.4 Add Redis config validation to `core/viper.go` `validateConfig()` function

## 3. Create Redis Client Initialization

- [ ] 3.1 Create `server/internal/core/redis.go` — `func Redis() *redis.Client` initializing go-redis client from config
- [ ] 3.2 Add `TD27_REDIS *redis.Client` global variable to `server/internal/global/global.go`
- [ ] 3.3 Wire Redis initialization into `server/cmd/server/main.go` after config/logger, before DB (with graceful nil handling when disabled)

## 4. Implement Cache Interface and RedisCache Adapter

- [ ] 4.1 Define `Cache` interface in `server/internal/pkg/cache/cache.go` with methods: Get, Set, Del, Exists, TTL, ListKeysByPrefix, DelByUsername
- [ ] 4.2 Ensure existing `PGCache` implements the new `Cache` interface (add any missing methods)
- [ ] 4.3 Create `server/internal/pkg/cache/redis_cache.go` with `RedisCache` struct implementing `Cache` interface via go-redis
- [ ] 4.4 Implement `NewCache()` factory function in `server/internal/pkg/cache/cache.go` — returns `RedisCache` when Redis enabled, `PGCache` otherwise

## 5. Migrate JwtService to RedisCache

- [ ] 5.1 Change `JwtService.cache` field type from `*PGCache` to `Cache` interface
- [ ] 5.2 Update `NewJwtService()` to use `cache.NewCache()` factory
- [ ] 5.3 Verify all JwtService methods work with RedisCache (Get, Set, Del, ListKeysByPrefix, TTL)

## 6. Migrate RBAC PermissionCache to RedisCache

- [ ] 6.1 Change `PermissionCache.cache` field type from `*PGCache` to `Cache` interface
- [ ] 6.2 Update `NewPermissionCache()` to use `cache.NewCache()` factory
- [ ] 6.3 Implement `ClearAllPermissions()` method in `RedisCache` using Redis SCAN + DEL
- [ ] 6.4 Add role-permission cache eviction on role update

## 7. Add Dashboard Stats Caching

- [ ] 7.1 Add `RedisCache` field to `DashboardService`
- [ ] 7.2 In `GetStatistics()`, check `td27:dashboard:stats` key first; on miss, execute queries and cache result for 30s
- [ ] 7.3 Return cached stats on subsequent requests within TTL

## 8. Add Dictionary Data Caching

- [ ] 8.1 Add `RedisCache` field to `DictService` and `DictDetailService`
- [ ] 8.2 Cache `DictService.List()` result at `td27:dict:list`
- [ ] 8.3 Cache `DictDetailService.Flat(dictId)` result at `td27:dict:{dictId}:flat`
- [ ] 8.4 Evict corresponding key on dict/detail Create, Update, Delete

## 9. Add Department Tree Caching

- [ ] 9.1 Add `RedisCache` field to `DeptService`
- [ ] 9.2 Cache `DeptService.List()` result at `td27:dept:tree`
- [ ] 9.3 Cache `DeptService.GetElTreeDepts()` result at `td27:dept:eltree`
- [ ] 9.4 Evict department cache keys on dept Create, Update, Delete

## 10. Add Menu Tree Caching

- [ ] 10.1 Add `RedisCache` field to `MenuService`
- [ ] 10.2 Cache `MenuService.List(roleIDs)` result at `td27:menu:role:{roleID}` per role
- [ ] 10.3 Cache `MenuService.ElTree(roleID)` result at `td27:menu:eltree:{roleID}`
- [ ] 10.4 Evict all menu cache keys on menu Create, Update, Delete, and on role-permission binding changes

## 11. Add Button Permission Caching

- [ ] 11.1 Add `RedisCache` field to `ButtonService`
- [ ] 11.2 Cache `GetPageButtons(pagePath, roleIDs)` result at `td27:button:page:{pagePath}:{roleIDsHash}`
- [ ] 11.3 Cache `GetUserButtons(roleIDs)` result at `td27:button:user:{roleIDsHash}`
- [ ] 11.4 Evict button cache on button Create, Update, Delete, and on role-permission binding changes

## 12. Add Service Token Caching

- [ ] 12.1 Add `RedisCache` field to `ServiceTokenService`
- [ ] 12.2 In `AuthenticateToken()`, check `td27:service_token:{tokenHash}` first; on miss, validate from DB and cache result for 15 minutes
- [ ] 12.3 Evict token cache on token status/expiry update or deletion

## 13. Update Docker Compose

- [ ] 13.1 Add `redis` service to `docker-compose/compose.yml` (health check, network, port 6379)
- [ ] 13.2 Add `depends_on: redis` to server service
- [ ] 13.3 Update `server/configs/config.yaml` Redis host to `redis` (Docker service name)

## 14. Verify

- [ ] 14.1 Build backend: `cd server && make build`
- [ ] 14.2 Run backend tests: `cd server && make test`
- [ ] 14.3 Run frontend lint: `cd web && pnpm lint`
- [ ] 14.4 Verify PGCache fallback works by setting `redis.enabled: false` and confirming all operations succeed
- [ ] 14.5 Verify Redis caching works by setting `redis.enabled: true` with a running Redis instance
