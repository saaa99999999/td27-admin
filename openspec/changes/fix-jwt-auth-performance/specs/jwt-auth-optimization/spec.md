## ADDED Requirements

### Requirement: GetClaims SHALL check context before re-parsing JWT
The system SHALL optimize `GetClaims()` to check `c.Get("claims")` from the Gin context before falling back to re-parsing the JWT from the `x-token` header. This eliminates redundant JWT parsing on every request that already passed through `JWTAuth` middleware.

#### Scenario: Context has claims
- **WHEN** `GetClaims(c)` is called and `c.Get("claims")` returns valid claims
- **THEN** system returns the cached claims directly without re-parsing the JWT

#### Scenario: Context has no claims
- **WHEN** `GetClaims(c)` is called and `c.Get("claims")` does not exist
- **THEN** system falls back to parsing the `x-token` header JWT

### Requirement: Casbin enforce logging SHALL use Debug level
The system SHALL log per-request Casbin enforcement checks at Debug level instead of Info level to reduce log noise in production.

#### Scenario: Enforce call succeeds
- **WHEN** Casbin `Enforce()` is called on every request
- **THEN** the debug log is written at Debug level, not Info level

### Requirement: ServiceToken routes SHALL use correct REST methods
The system SHALL use standard REST HTTP methods for ServiceToken mutation endpoints: `DELETE` for delete, `PUT` for update.

#### Scenario: Delete service token
- **WHEN** client calls the delete endpoint
- **THEN** the HTTP method SHALL be `DELETE` (not `GET`)

#### Scenario: Update service token
- **WHEN** client calls the update endpoint
- **THEN** the HTTP method SHALL be `PUT` (not `GET`)

### Requirement: JWT instance SHALL be cached
The system SHALL cache the `JWT` struct to avoid re-converting the signing key string to `[]byte` on every `NewJWT()` call.

#### Scenario: JWT parsing or creation
- **WHEN** any component calls `NewJWT()` or uses the JWT instance for token operations
- **THEN** the signing key `[]byte` SHALL be allocated only once at initialization
