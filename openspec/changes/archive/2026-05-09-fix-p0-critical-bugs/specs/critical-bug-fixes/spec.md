## ADDED Requirements

### Requirement: Backend server auto-migrate respects disable flag
The system SHALL correctly interpret the `disable-auto-migrate` configuration flag. When set to `true`, the server MUST skip auto-migration during startup. When set to `false`, the server MUST run auto-migration.

#### Scenario: Auto-migrate skipped when disabled
- **WHEN** `system.disable-auto-migrate` is `true` in config.yaml
- **THEN** the server MUST NOT call `RegisterTables()` during startup
- **AND** the server MUST proceed directly to cron initialization

#### Scenario: Auto-migrate runs when enabled
- **WHEN** `system.disable-auto-migrate` is `false` in config.yaml
- **THEN** the server MUST call `RegisterTables()` during startup

### Requirement: Permission queries use correct column name
The system SHALL use the correct column name `domain` (not `type`) when querying the unified permissions table.

#### Scenario: Menu permissions query succeeds
- **WHEN** a role with menu permissions exists and the menu permission association is loaded
- **THEN** the query MUST use `domain = 'menu'` instead of `type = 'menu'`
- **AND** the query MUST return the correct permission records without SQL error

### Requirement: Department materialized paths are ASCII-safe
The system SHALL generate department materialized paths using decimal string representation, not Unicode rune conversion.

#### Scenario: High-ID department path is valid
- **WHEN** a department with ID > 127 is created
- **THEN** its `GetFullPath()` MUST return an ASCII-safe path segment (e.g., `/200/` not `/È/`)
- **AND** the path MUST be usable in SQL `LIKE` queries

### Requirement: SQL NULL comparisons use IS NULL syntax
The cache UPSERT operation SHALL use correct SQL `IS NULL` syntax for null comparisons.

#### Scenario: Soft-deleted cache record is updated
- **WHEN** a cache key exists as a soft-deleted record
- **THEN** the UPDATE branch MUST match it using `"deleted_at" IS NULL`
- **AND** the operation MUST update the existing record instead of attempting an insert

### Requirement: Role deletion cleans up all permission associations
When a role is deleted, the system SHALL remove ALL role-permission associations for that role, not just menu-type permissions.

#### Scenario: Role with API and button permissions is deleted
- **WHEN** a role with API, menu, and button permissions is deleted
- **THEN** ALL `role_permission` records for that role MUST be removed
- **AND** no orphaned permission associations SHALL remain

### Requirement: Logout API returns correct success message
The logout endpoint SHALL return a success message indicating successful logout, not a failure message.

#### Scenario: User logs out successfully
- **WHEN** a valid JWT token is provided to the logout endpoint
- **THEN** the response MUST indicate successful logout
- **AND** the message MUST NOT say "登出失败" (logout failed)
