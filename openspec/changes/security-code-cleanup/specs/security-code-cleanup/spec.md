## ADDED Requirements

### Requirement: Bcrypt Password Hashing
The system SHALL use bcrypt for all password hashing operations instead of MD5. The bcrypt cost factor SHALL be the default (10).

#### Scenario: Login verifies password with bcrypt
- **WHEN** a user submits a login request with correct password
- **THEN** the system compares the password using bcrypt.CompareHashAndPassword
- **AND** authentication succeeds if the hash matches

#### Scenario: New user password hashed with bcrypt
- **WHEN** an admin creates a new user
- **THEN** the user's password is stored as a bcrypt hash

#### Scenario: Password change hashed with bcrypt
- **WHEN** a user changes their password
- **THEN** the new password is stored as a bcrypt hash

#### Scenario: Old MD5 hash still works
- **WHEN** a user has an MD5 hash in the database and logs in with correct password
- **THEN** the system first verifies against bcrypt
- **AND** if that fails, verifies against the old MD5 hash
- **AND** on success, re-hashes the password with bcrypt and updates the stored hash

#### Scenario: MD5 helper is removed
- **WHEN** the codebase is built after the change
- **THEN** `pkg/md5.go` no longer exists
- **AND** no code references `pkg.MD5V`

### Requirement: File Download Path Traversal Prevention
The file download endpoint SHALL reject any download request where the resolved file path escapes the configured upload directory.

#### Scenario: Normal file download succeeds
- **WHEN** a valid filename is requested via `/file/download?name=report.csv`
- **AND** the file exists in the upload directory
- **THEN** the file is served with HTTP 200

#### Scenario: Path traversal attempt rejected
- **WHEN** a filename containing `../` or `..\\` is requested
- **AND** the resolved path falls outside the upload directory
- **THEN** the request is rejected with HTTP 400

#### Scenario: Absolute path attempt rejected
- **WHEN** a filename starting with `/` is requested
- **THEN** the request is rejected with HTTP 400

### Requirement: Data Permission Cache Without Goroutine Leak
The data permission cache SHALL NOT use a background goroutine with `time.Sleep` for entry expiry.

#### Scenario: Cache entry expires lazily
- **WHEN** a cached data permission entry's TTL (5 minutes) has elapsed
- **THEN** the next read returns nil
- **AND** the system fetches fresh data from the database

#### Scenario: No background goroutines for cache expiry
- **WHEN** the data permission service is instantiated and used
- **THEN** no goroutines are created solely for cache entry expiry

### Requirement: Goroutine Closure Safety
All goroutines in the service layer SHALL capture loop or outer variables correctly by shadowing.

#### Scenario: Role deletion goroutine uses correct err
- **WHEN** a role is deleted and the Casbin reload goroutine runs
- **THEN** the goroutine captures the local err variable at the time of spawning
- **AND** does not read a reassigned outer `err`

### Requirement: CORS Whitelist O(1) Lookup
The CORS whitelist check SHALL use a map for constant-time origin lookup instead of linear slice scan.

#### Scenario: Allowed origin passes CORS check
- **WHEN** a request comes from an origin in the CORS whitelist
- **THEN** CORS headers are set in the response

#### Scenario: Disallowed origin in strict mode is rejected
- **WHEN** a request comes from an origin NOT in the whitelist
- **AND** cors mode is `strict-whitelist`
- **THEN** the request is rejected with HTTP 403

### Requirement: Operation Log Body Size Limit
The operation log middleware SHALL limit buffered request body to 10KB and SHALL use `url.Values` for GET query parameters instead of JSON marshaling.

#### Scenario: Normal request body captured
- **WHEN** a POST/PUT/PATCH request body is 10KB or smaller
- **THEN** the full body is captured in the operation log

#### Scenario: Large request body truncated
- **WHEN** a POST/PUT/PATCH request body exceeds 10KB
- **THEN** only the first 10KB are captured in the operation log
- **AND** a suffix `... [truncated]` is appended

#### Scenario: GET query params captured as string
- **WHEN** a GET request is made with query parameters
- **THEN** the query parameters are captured as a URL-encoded string
- **AND** no JSON marshaling is performed

### Requirement: Service Layer Database Access via Repository
Services SHALL access the database through constructor-injected `*gorm.DB` instances or repository interfaces, not via `global.TD27_DB`.

#### Scenario: Login service uses injected DB
- **WHEN** `LogRegService.Login` is called
- **THEN** it queries the database through its own injected `*gorm.DB` or repository field
- **AND** does not reference `global.TD27_DB`

#### Scenario: Button service uses injected DB
- **WHEN** `ButtonService` methods query the database
- **THEN** they use the service's injected `*gorm.DB` or repository field
- **AND** do not reference `global.TD27_DB`

#### Scenario: Dashboard service uses injected DB
- **WHEN** `DashboardService` methods query the database
- **THEN** they use the service's injected `*gorm.DB` or repository field
- **AND** do not reference `global.TD27_DB`

### Requirement: Dead Code Removal
The codebase SHALL NOT contain dead code: commented-out logic blocks, unused types, unreferenced functions, or unreachable statements.

#### Scenario: Dead code locations removed
- **WHEN** the codebase is built with `go build ./...`
- **THEN** the build succeeds
- **AND** `go vet ./...` passes without errors
- **AND** the removed locations are confirmed absent:
  - `role.go:64` — commented `global.TD27_DB.Model(&roleModel).Association("Menus").Clear()`
  - `router.go:33-35` — commented CORS middleware lines
  - `permission_dto.go:3-6` — commented `ListPermissionReq` struct
  - `permission_model.go:72-74` — commented `ToCasbinRule` method
  - `operation_log_repository.go:31-33` — commented result line
  - `role_repository.go:107-108` — commented DeleteRoleMenu call
  - `server.go:28` — commented `global.TD27_LOG.Info` line
