## ADDED Requirements

### Requirement: Batch role existence check

The system SHALL verify all role IDs exist in a single query when creating or updating a user.

#### Scenario: Create user with valid roles
- **WHEN** creating a user with role IDs `[1, 2, 3]`
- **THEN** the system SHALL execute one `SELECT` query with `WHERE id IN (1,2,3)` instead of three individual queries

#### Scenario: Update user with non-existent role
- **WHEN** updating a user with role ID that does not exist
- **THEN** the system SHALL return an error

### Requirement: Batch permission delete by domain IDs

The system SHALL delete permissions for multiple domain IDs in a single query.

#### Scenario: Delete permissions for multiple APIs
- **WHEN** deleting APIs with IDs `[1, 2, 5]`
- **THEN** the system SHALL execute one `DELETE` query with `WHERE domain_id IN (1,2,5)` instead of three individual queries

### Requirement: Batch Casbin policy cleanup

The system SHALL perform a single-pass batch Casbin policy removal for multiple APIs.

#### Scenario: Delete multiple APIs
- **WHEN** deleting 5 APIs
- **THEN** the system SHALL call `RemoveFilteredPolicy` for each API (Casbin API limitation) but SHALL NOT make additional DB calls per API beyond the Casbin enforcer

### Requirement: Subquery for token API count

The system SHALL include the API count for each service token as a correlated subquery instead of a per-token COUNT.

#### Scenario: List service tokens with counts
- **WHEN** listing 50 service tokens
- **THEN** the system SHALL execute 1 query with a subquery instead of 51 queries (1 list + 50 COUNTs)

### Requirement: Combined dashboard statistics

The system SHALL retrieve all dashboard statistics in at most 2 database round-trips.

#### Scenario: Load dashboard
- **WHEN** loading the dashboard page
- **THEN** the system SHALL execute at most 2 queries (not 6) to gather all statistics

### Requirement: Batch insert operation logs

The async logger SHALL insert buffered operation logs in a single batch query.

#### Scenario: Flush 100 logs
- **WHEN** the flush interval triggers with 100 buffered logs
- **THEN** the system SHALL execute 1 batch `INSERT` instead of 100 individual inserts

### Requirement: Atomic department path update

The system SHALL wrap recursive child path updates in a database transaction.

#### Scenario: Update department path
- **WHEN** updating a department's parent
- **THEN** all recursive child path updates SHALL execute within a single transaction
- **AND** a failure at any level SHALL roll back all path changes
