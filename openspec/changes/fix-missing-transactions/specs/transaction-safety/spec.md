## ADDED Requirements

### Requirement: Service-layer multi-write operations MUST be atomic
When a service method performs writes across multiple repository calls (e.g., create record + create associated permission), the system SHALL wrap all DB writes in a single Gorm transaction. If any write fails, ALL writes in the transaction MUST be rolled back.

#### Scenario: API creation permission write fails
- **WHEN** `ApiService.Create` creates an API record successfully
- **AND** the subsequent permission record creation fails
- **THEN** the API record MUST also be rolled back
- **AND** the method SHALL return an error

#### Scenario: API deletion permission write fails
- **WHEN** `ApiService.Delete` attempts to delete the permission and API records
- **AND** the permission deletion fails
- **THEN** the API deletion MUST still proceed (logged but non-fatal)
- **AND** the overall operation SHALL NOT leave partial state (permission deletion failure is logged, not rolled back)

#### Scenario: Bulk API deletion succeeds fully
- **WHEN** `ApiService.DeleteByIds` is called with valid IDs
- **THEN** ALL associated permissions SHALL be deleted
- **AND** ALL API records SHALL be deleted
- **AND** Casbin policies SHALL be cleaned up
- **AND** the operation SHALL be atomic for the Gorm writes

#### Scenario: API update with permission creation
- **WHEN** `ApiService.Update` updates an API record
- **AND** the corresponding permission does not exist and needs creation
- **THEN** both the API update and permission creation SHALL be atomic
- **AND** Casbin policy update SHALL occur after the transaction succeeds

#### Scenario: Menu creation permission write fails
- **WHEN** `MenuService.Create` creates a menu successfully
- **AND** the subsequent permission record creation fails
- **THEN** the menu record MUST also be rolled back

#### Scenario: Menu update with permission upsert
- **WHEN** `MenuService.Update` updates a menu record
- **AND** the corresponding permission is created or updated
- **THEN** both the menu update and permission upsert SHALL be atomic

#### Scenario: Menu deletion removes menu and permission atomically
- **WHEN** `MenuService.Delete` deletes a menu
- **THEN** both the menu record and its associated permission SHALL be removed atomically

#### Scenario: Button creation permission write fails
- **WHEN** `ButtonService.Create` creates a button record
- **AND** the subsequent permission creation via `global.TD27_DB.Create` fails
- **THEN** the button record MUST also be rolled back

#### Scenario: Button update atomicity
- **WHEN** `ButtonService.Update` updates a button record and its permission name
- **THEN** both the button update and permission name update SHALL be atomic

#### Scenario: Button deletion removes both records
- **WHEN** `ButtonService.Delete` deletes a button
- **THEN** both the permission record and the button record SHALL be removed atomically

#### Scenario: Role deletion cleans up all associations
- **WHEN** `RoleService.Delete` deletes a role
- **THEN** the role record, role-menu associations, and role-permission associations SHALL be deleted atomically

#### Scenario: Service token creation atomicity
- **WHEN** `ServiceTokenService.Create` creates a token and sets its permissions
- **THEN** the token record AND its permission associations SHALL be created atomically
- **AND** Casbin policy sync SHALL occur after the transaction succeeds

#### Scenario: Service token update atomicity
- **WHEN** `ServiceTokenService.Update` updates a token and its permissions
- **THEN** the token update AND permission association update SHALL be atomic
- **AND** Casbin policy sync SHALL occur after the transaction succeeds

#### Scenario: Service token deletion atomicity
- **WHEN** `ServiceTokenService.Delete` deletes a token
- **THEN** the token permission associations AND the token record SHALL be removed atomically
- **AND** Casbin policy cleanup SHALL occur before the transaction

### Requirement: Repository multi-write operations MUST use `Transaction(func)` pattern
Repository methods performing multiple DB writes SHALL use `e.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error { ... })` instead of manual `Begin`/`Commit`/`Rollback` or no transaction at all. The `Transaction` function SHALL auto-rollback on error or panic.

#### Scenario: User creation with role association
- **WHEN** `userEntity.Create` creates a user and associates roles
- **THEN** the user record AND role associations SHALL be created within a single `Transaction` closure
- **AND** on any error, ALL writes SHALL be auto-rolled back
- **AND** manual `tx.Begin()` / `tx.Rollback()` / `tx.Commit()` SHALL NOT be used

#### Scenario: User update with role replacement
- **WHEN** `userEntity.Update` updates user fields and replaces role associations
- **THEN** the user update, role clear, and role append SHALL be within a single `Transaction` closure
- **AND** on any error, ALL writes SHALL be auto-rolled back

#### Scenario: Department creation with path update
- **WHEN** `deptEntity.Create` creates a department and updates its materialized path
- **THEN** the department insert AND path update SHALL be within a single `Transaction` closure

#### Scenario: Department reparenting updates children paths atomically
- **WHEN** `deptEntity.Update` reparents a department (changes parent_id)
- **THEN** the department field update AND all recursive child path updates SHALL be within a single `Transaction` closure
- **AND** on any error during recursive child path updates, ALL path changes SHALL be rolled back

#### Scenario: Role permission batch insert atomicity
- **WHEN** `roleRepo.UpdateRolePermission` creates multiple role-permission associations
- **THEN** all `Create` calls SHALL be within a single `Transaction` closure
- **AND** on any error, NO role-permission records SHALL be persisted

### Requirement: Repository interfaces SHALL support transaction propagation
Repository interfaces used by the service layer for multi-write operations SHALL provide a `WithTx(*gorm.DB)` method that returns a new repository instance bound to the given transaction handle. This enables the service layer to pass a transaction through the repository abstraction.

#### Scenario: Service uses WithTx for transactional multi-write
- **WHEN** a service method starts a Gorm transaction
- **THEN** it SHALL call `repo.WithTx(tx)` to obtain a transactional repository instance
- **AND** all DB operations within the transaction closure SHALL use the transactional instance

### Requirement: Manual `Begin`/`Commit`/`Rollback` SHALL NOT be used
All existing manual transaction management (`tx.Begin()`, `tx.Commit()`, `tx.Rollback()`) in the codebase SHALL be replaced with the `Transaction(func(tx *gorm.DB) error)` pattern.

#### Scenario: No manual transaction calls remain
- **WHEN** the codebase is scanned for `\.Begin\(\)` and `\.Rollback\(\)` calls
- **THEN** no remaining instances SHALL exist in production code (test files may be excluded)
