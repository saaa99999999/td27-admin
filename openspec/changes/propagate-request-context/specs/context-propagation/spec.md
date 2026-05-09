## ADDED Requirements

### Requirement: Service methods accept context as first parameter
Every exported method on service structs SHALL accept `ctx context.Context` as its first parameter. Callers SHALL pass the request context from the Gin handler layer.

#### Scenario: Service method signature
- **WHEN** a service method is called
- **THEN** its first parameter SHALL be `ctx context.Context`
- **AND** the method SHALL use this ctx for all database and cache operations within its scope

#### Scenario: Context passed by handler
- **WHEN** an API handler calls a service method
- **THEN** it SHALL pass `c.Request.Context()` as the first argument

### Requirement: No stored context in service structs
Service structs SHALL NOT store a `context.Context` field. The `ctx context.Context` field SHALL be removed from all service structs.

#### Scenario: No ctx field
- **WHEN** a service struct is instantiated
- **THEN** its struct definition SHALL NOT contain a `ctx` field of type `context.Context`

#### Scenario: Constructor without ctx
- **WHEN** a service constructor (`New*Service`) is called
- **THEN** it SHALL NOT set a `ctx` field

### Requirement: No context.Background() in service layer
The `context.Background()` function SHALL NOT be called in `server/internal/service/` or `server/internal/pkg/rbac/` except in test files.

#### Scenario: Elimination in JwtService
- **WHEN** `JwtService.AddToken`, `ValidateToken`, `RemoveToken`, `RemoveAllTokens`, `GetUserActiveSessions`, or `GetCachedUser` is called
- **THEN** the method SHALL use the `ctx` parameter passed by the caller
- **AND** SHALL NOT create `context.Background()` locally

#### Scenario: Elimination in PermissionCache
- **WHEN** `PermissionCache.CacheUserPermissions`, `GetUserPermissions`, `ClearUserPermissions`, `CacheRolePermissions`, `GetRolePermissions`, or `ClearRolePermissions` is called
- **THEN** the method SHALL use the `ctx` parameter passed by the caller
- **AND** SHALL NOT call `context.Background()` internally

### Requirement: Server timeouts tightened
The HTTP server configuration SHALL use production-appropriate timeout values.

#### Scenario: Read timeout
- **WHEN** the HTTP server starts
- **THEN** `ReadTimeout` SHALL be set to 30s

#### Scenario: Write timeout
- **WHEN** the HTTP server starts
- **THEN** `WriteTimeout` SHALL be set to 60s

#### Scenario: Idle timeout
- **WHEN** the HTTP server starts
- **THEN** `IdleTimeout` SHALL be set to 120s

#### Scenario: Read header timeout
- **WHEN** the HTTP server starts
- **THEN** `ReadHeaderTimeout` SHALL be set to 20s

### Requirement: Background/cron jobs unaffected
Cron job runners and background task code SHALL continue to use `context.Background()` or their own context as appropriate. These paths are outside the request lifecycle and are not in scope for this change.

#### Scenario: Cron job runner context
- **WHEN** a cron job is triggered via `pkg/cron/job.go`
- **THEN** it MAY continue to use `context.Background()` as its execution context

#### Scenario: Test file context
- **WHEN** a test file calls a service method
- **THEN** it MAY pass `context.Background()` or `context.TODO()` as the context argument
