## Why

The backend has 19 locations across 3 categories where multi-write operations lack database transaction safety. Without atomicity, partial failures can leave the database in an inconsistent state — e.g., an API record created without its corresponding permission, or a menu deleted but its permission association surviving. These are data corruption risks in production.

## What Changes

- **Wrap all service-layer multi-write operations in Gorm transactions** — 14 locations across `api.go`, `menu.go`, `button.go`, `role.go`, and `service_token.go` currently perform multiple DB writes without atomicity.
- **Replace manual `Begin`/`Commit`/`Rollback` with `Transaction(func)`** — 2 locations in `user_repository.go` use error-prone manual transaction management that can leak connections on early return or panic.
- **Wrap repository multi-write operations in Gorm transactions** — 3 locations in `dept_repository.go` and `role_repository.go` perform multiple writes without any transaction.
- **Add `WithTx(*gorm.DB)` propagation method** to affected repository interfaces so the service layer can pass a transaction handle through the repository abstraction.

## Capabilities

### New Capabilities
- `transaction-safety`: Database transaction safety for all multi-write operations across service and repository layers.

### Modified Capabilities
None — this is a correctness fix with no spec-level behavior changes.

## Impact

- **Files modified**:
  - `server/internal/service/sysManagement/api.go` (4 methods: Create, Delete, DeleteByIds, Update)
  - `server/internal/service/sysManagement/menu.go` (3 methods: Create, Update, Delete)
  - `server/internal/service/sysManagement/button.go` (3 methods: Create, Update, Delete)
  - `server/internal/service/sysManagement/role.go` (1 method: Delete)
  - `server/internal/service/sysTool/service_token.go` (3 methods: Create, Update, Delete)
  - `server/internal/model/sysManagement/user_repository.go` (2 methods: Create, Update)
  - `server/internal/model/sysManagement/dept_repository.go` (2 methods: Create, Update + updateChildrenPath)
  - `server/internal/model/sysManagement/role_repository.go` (1 method: UpdateRolePermission)
  - Repository interfaces for API, Permission, Menu, Button, ServiceToken (add `WithTx`)
- **No new dependencies** — uses Gorm's built-in `Transaction` function.
- **No API contract changes** — behavior is identical on success; only failure behavior changes (atomic rollback instead of partial writes).
- **No schema migrations needed**.
