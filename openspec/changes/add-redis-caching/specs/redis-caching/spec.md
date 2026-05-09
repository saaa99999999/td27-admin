## ADDED Requirements

### Requirement: Redis Configuration
The system SHALL support Redis connection configuration via `config.yaml` under a `redis` section. Redis SHALL be optional — when disabled, all cache operations fall back to PGCache.

#### Scenario: Redis enabled
- **WHEN** `redis.enabled` is set to `true` in config.yaml
- **THEN** the system SHALL connect to Redis at the configured `host:port`
- **AND** the system SHALL use `go-redis/v9` for all cache operations
- **AND** authentication SHALL use the configured `password` (empty string for no auth)
- **AND** connection pool SHALL be configured via `pool-size`

#### Scenario: Redis disabled
- **WHEN** `redis.enabled` is set to `false` in config.yaml
- **THEN** the system SHALL NOT attempt to connect to Redis
- **AND** all cache operations SHALL fall back to PGCache (PostgreSQL-backed)
- **AND** no additional error or warning SHALL be raised

### Requirement: Cache Interface Abstraction
The system SHALL provide a `Cache` interface in `server/internal/pkg/cache/` that both `PGCache` and `RedisCache` implement. Consumers SHALL depend on the interface rather than the concrete type.

#### Scenario: Interface defined
- **WHEN** a consumer requires caching
- **THEN** it SHALL reference the `Cache` interface
- **AND** the concrete implementation (PGCache or RedisCache) SHALL be injected at construction time

### Requirement: JWT Token Caching via Redis
The `JwtService` SHALL use RedisCache for storing and validating JWT tokens when Redis is enabled.

#### Scenario: Token added with Redis
- **WHEN** a new JWT token is created for a user
- **THEN** the token SHALL be stored in Redis with key `td27:jwt:*`
- **AND** the token SHALL have a TTL equal to the configured JWT expiry
- **AND** multi-login token limits SHALL be enforced using Redis SCAN

#### Scenario: Token validated with Redis
- **WHEN** a request with a JWT token is received
- **THEN** the token SHALL be validated against Redis (not PostgreSQL)
- **AND** an invalid or expired token SHALL return false

### Requirement: RBAC Permission Caching via Redis
The `PermissionCache` in `server/internal/pkg/rbac/` SHALL use RedisCache for storing and retrieving user and role permissions when Redis is enabled.

#### Scenario: Permissions cached in Redis
- **WHEN** user permissions are loaded from the database
- **THEN** they SHALL be cached in Redis with key prefix `td27:rbac:user:perm:*` with 30-minute TTL
- **AND** subsequent permission checks SHALL read from Redis first

#### Scenario: Permissions invalidated
- **WHEN** a role or user permission is updated
- **THEN** the corresponding key SHALL be deleted from Redis
- **AND** next permission check SHALL reload from the database

### Requirement: Dashboard Statistics Caching
The dashboard statistics endpoint SHALL cache its results in Redis with a 30-second TTL.

#### Scenario: Dashboard stats cached
- **WHEN** the dashboard statistics endpoint is called
- **THEN** the system SHALL check Redis for cached stats at `td27:dashboard:stats`
- **AND** if cache misses, SHALL execute the 6 COUNT queries and cache the result for 30 seconds

### Requirement: Dictionary Data Caching
Dictionary and dictionary detail data SHALL be cached in Redis with a 1-hour TTL.

#### Scenario: Dictionary data cached
- **WHEN** dictionary data is queried
- **THEN** the result SHALL be cached in Redis at `td27:dict:{dictID}`
- **AND** subsequent reads SHALL return cached data
- **AND** when a dictionary entry is created, updated, or deleted, the corresponding Redis key SHALL be evicted

### Requirement: Department Tree Caching
The department tree data SHALL be cached in Redis with a 1-hour TTL.

#### Scenario: Department tree cached
- **WHEN** the department list is queried
- **THEN** the result SHALL be cached in Redis at `td27:dept:tree`
- **AND** when a department is created, updated, or deleted, the Redis key SHALL be evicted

### Requirement: Per-Role Menu Tree Caching
The per-role menu tree SHALL be cached in Redis with a 1-hour TTL.

#### Scenario: Menu tree cached per role
- **WHEN** a user's menu tree is generated based on their roles
- **THEN** the result SHALL be cached in Redis at `td27:menu:role:{roleID}`
- **AND** when a menu or role-permission binding is updated, all menu cache keys SHALL be evicted

### Requirement: Button Permission Caching
Button permissions per role/page SHALL be cached in Redis with a 1-hour TTL.

#### Scenario: Button permissions cached
- **WHEN** button permissions are queried for a role
- **THEN** the result SHALL be cached in Redis at `td27:button:page:{pagePath}` and `td27:button:user:{roleIDsHash}`
- **AND** subsequent checks SHALL read from Redis
- **AND** when a button or role-permission binding is updated, affected keys SHALL be evicted

### Requirement: Service Token Caching
Service token authentication results SHALL be cached in Redis with a 15-minute TTL.

#### Scenario: Service token validated via Redis
- **WHEN** a service token is authenticated
- **THEN** the validation result (token ID + status) SHALL be cached in Redis at `td27:service_token:{tokenHash}`
- **AND** subsequent authentication attempts for the same token SHALL read from Redis
- **AND** if the token is disabled or expired, the cache SHALL be evicted
