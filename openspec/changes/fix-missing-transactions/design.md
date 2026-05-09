## Context

The codebase has 19 locations where multi-write DB operations lack atomicity. During a code audit, we classified them into 3 categories:

- **Category A (14 locations)**: Service layer calls multiple repository methods without wrapping them in a transaction. A failure in the second write leaves the first write committed — data is inconsistent.
- **Category B (2 locations)**: Repository layer uses `tx.Begin()` / `tx.Commit()` / `tx.Rollback()` manually. This pattern is error-prone: missing rollback on early return, panic not handled, and deferred cleanup is easy to forget.
- **Category C (3 locations)**: Repository layer performs multiple writes in sequence with no transaction at all.

The existing `service_token_repository.go:118`, `role_permission_repository.go:27`, and `dict_detail_repository.go:161` already use the correct `Transaction(func(tx *gorm.DB) error)` pattern, providing a proven template to follow.

Casbin operations (policy management) use their own storage and cannot participate in Gorm transactions. They are idempotent and logged on failure — acceptable to leave outside the transaction boundary.

## Goals / Non-Goals

**Goals:**
- Make all 19 multi-write locations atomic via Gorm's `Transaction(func(tx *gorm.DB) error)` pattern
- Replace all manual `Begin`/`Commit`/`Rollback` calls with the declarative `Transaction` wrapper
- Add `WithTx(*gorm.DB)` propagation to repository interfaces so the service layer can pass a transaction handle through the abstraction
- Maintain identical success-path behavior; only failure behavior changes (full rollback instead of partial writes)

**Non-Goals:**
- No refactoring beyond adding transaction safety
- No changes to single-write operations
- No new test coverage (existing tests suffice for regression)
- No schema or config changes
- No changes to Casbin operations (they stay outside DB transactions)

## Decisions

### Decision 1: Use Gorm's `Transaction(func(tx *gorm.DB) error)` pattern everywhere

The codebase already uses this pattern in 3 places (`dict_detail_repository.go`, `role_permission_repository.go`, `service_token_repository.go`). It auto-rollbacks on error or panic, which is safer than manual `Begin`/`Commit`/`Rollback`.

**Alternatives considered:**
- Manual `Begin`/`Commit`/`Rollback` with deferred rollback — rejected because it's error-prone and inconsistent with existing codebase patterns.
- `saga` pattern — overengineered for this scope; Gorm transactions are sufficient for single-service multi-write operations.

### Decision 2: Add `WithTx(*gorm.DB)` to repository interfaces for service-layer propagation

For Category A, the service holds an interface reference (e.g., `apiRepo modelSysManagement.APIRepository`). To use a transaction, it needs a way to create a transactional instance of the same interface. We add:

```go
type APIRepository interface {
    WithTx(*gorm.DB) APIRepository
    Create(ctx, req) (*ApiModel, error)
    // ... existing methods
}
```

The concrete struct copies itself with the tx as `conn`:

```go
func (e *apiEntity) WithTx(tx *gorm.DB) APIRepository {
    return &apiEntity{conn: tx}
}
```

The service then wraps multi-write in a transaction:

```go
err := global.TD27_DB.WithContext(s.ctx).Transaction(func(tx *gorm.DB) error {
    apiRepo := s.apiRepo.WithTx(tx)
    permRepo := s.permissionRepo.WithTx(tx)
    // ... all DB writes using apiRepo, permRepo
    return nil
})
```

**Affected interfaces**: `APIRepository`, `PermissionRepository`, `MenuRepository`, `ButtonRepository`.  
`ServiceTokenRepository` uses `global.TD27_DB` directly — refactor it to use `conn *gorm.DB` like the others, then add `WithTx`.

**Alternatives considered:**
- Pass `*gorm.DB` as an extra parameter to each repo method — rejected because it pollutes every method signature.
- Service calls `global.TD27_DB` directly inside the transaction closure, bypassing the repo — rejected because it breaks the repository abstraction.
- Keep `WithTx` unexported (struct method only) and cast in the service — rejected because it couples the service to concrete types.

### Decision 3: Leave Casbin operations outside transaction boundaries

Casbin policies are stored in the same PostgreSQL database (via Gorm adapter), but Casbin uses its own connection and caching layer. Wrapping Casbin writes in the same Gorm transaction as model writes would require Casbin to use the same `*gorm.DB` handle, which is not how the library is designed. Casbin operations are idempotent and failures are already logged (non-fatal). This is acceptable.

### Decision 4: Replace manual `Begin`/`Commit`/`Rollback` in `user_repository.go`

Current pattern:
```go
tx := e.conn.WithContext(ctx).Begin()
if err := tx.Create(...).Error; err != nil {
    tx.Rollback()
    return nil, err
}
tx.Commit()
```

Replace with:
```go
var userModel UserModel
err := e.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&userModel).Error; err != nil {
        return err
    }
    // ... more operations using tx
    return nil
})
```

This ensures auto-rollback on any error or panic. The pattern applies to both `Create` and `Update`.

### Decision 5: Wrap multi-write in `dept_repository.go` and `role_repository.go`

- `deptEntity.Create`: Insert dept + update path are two writes. Wrap in `Transaction(func(tx) { ... })`.
- `deptEntity.Update`: Update dept fields + recursive `updateChildrenPath` (which does N individual DB writes). Wrap the entire operation in `Transaction(func(tx) { ... })`, with `updateChildrenPath` accepting and passing `tx`.
- `roleRepo.UpdateRolePermission`: Loop creating role-permission records. Wrap in `Transaction(func(tx) { ... })`.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Adding `WithTx` to repository interfaces breaks the interface contract for anyone implementing it | All repositories in the codebase follow the `struct + conn *gorm.DB` pattern. Implementing `WithTx` is trivial — copy struct with new conn. |
| Transaction in service layer may be slower (connection held longer) | Transaction scope is limited to Gorm writes only (no external calls, no Casbin). Duration is sub-millisecond for the writes involved. |
| `service_token_repo.go` refactor from `global.TD27_DB` to `conn *gorm.DB` could introduce regression | The refactor is mechanical (`global.TD27_DB.WithContext` → `r.conn.WithContext`). Constructor already receives `conn` but ignores it — fix the constructor body. Test by running existing tests. |
| Recursive `updateChildrenPath` wrapped in a long transaction could hold the connection | Dept hierarchies are shallow (typically <5 levels). Each level does a single `Find` + `Update`. The total transaction time is negligible. |
| `Transaction` with `WithTx` creates a new struct on every call (alloc) | Allocation of a small struct is negligible. No measurable performance impact. |
