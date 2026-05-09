## 1. Fix Service Layer — api.go (4 methods)

- [ ] 1.1 Add `WithTx(*gorm.DB)` method to `apiEntity` and `permissionEntity` structs, add to `APIRepository` and `PermissionRepository` interfaces
- [ ] 1.2 Wrap `ApiService.Create` Gorm writes in `Transaction(func(tx) { apiRepo.WithTx(tx).Create(...); permRepo.WithTx(tx).Create(...) })`
- [ ] 1.3 Wrap `ApiService.Delete` Gorm writes in `Transaction(func(tx) { permRepo.WithTx(tx).DeleteByDomainID(...); apiRepo.WithTx(tx).Delete(...) })`, keep Casbin outside
- [ ] 1.4 Wrap `ApiService.DeleteByIds` Gorm writes in `Transaction(func(tx) { ... })`, keep Casbin outside
- [ ] 1.5 Wrap `ApiService.Update` Gorm writes in `Transaction(func(tx) { apiRepo.WithTx(tx).Update(...); permRepo.WithTx(tx).FindByDomainID/Create/Update(...) })`, keep Casbin outside

## 2. Fix Service Layer — menu.go (3 methods)

- [ ] 2.1 Add `WithTx(*gorm.DB)` to `menuEntity` struct and `MenuRepository` interface
- [ ] 2.2 Wrap `MenuService.Create` in `Transaction(func(tx) { menuRepo.WithTx(tx).Create(...); permRepo.WithTx(tx).Create(...) })`
- [ ] 2.3 Wrap `MenuService.Update` in `Transaction(func(tx) { menuRepo.WithTx(tx).Update(...); permRepo.WithTx(tx).FindByDomainID/Create/Update(...) })`
- [ ] 2.4 Wrap `MenuService.Delete` in `Transaction(func(tx) { permRepo.WithTx(tx).DeleteByDomainID(...); menuRepo.WithTx(tx).Delete(...) })`

## 3. Fix Service Layer — button.go (3 methods)

- [ ] 3.1 Add `WithTx(*gorm.DB)` to `buttonEntity` struct and `ButtonRepository` interface
- [ ] 3.2 Wrap `ButtonService.Create` in `Transaction(func(tx) { buttonRepo.WithTx(tx).Create(...); tx.Create(permission) })`
- [ ] 3.3 Wrap `ButtonService.Update` in `Transaction(func(tx) { buttonRepo.WithTx(tx).Update(...); tx.Model(...).Update(...) })`
- [ ] 3.4 Wrap `ButtonService.Delete` in `Transaction(func(tx) { tx.Where(...).Delete(...); buttonRepo.WithTx(tx).Delete(...) })`

## 4. Fix Service Layer — role.go (1 method)

- [ ] 4.1 Add `WithTx(*gorm.DB)` to `roleRepo` struct and `RoleRepository` interface
- [ ] 4.2 Wrap `RoleService.Delete` Gorm writes in `Transaction(func(tx) { roleRepo.WithTx(tx).Delete(...); roleRepo.WithTx(tx).DeleteRoleMenu(...); tx.Where(...).Delete(&RolePermissionModel{}) })`, keep Casbin outside

## 5. Fix Service Layer — service_token.go (3 methods)

- [ ] 5.1 Refactor `serviceTokenRepo` from `global.TD27_DB` to `conn *gorm.DB` field, update constructor
- [ ] 5.2 Add `WithTx(*gorm.DB)` to `serviceTokenRepo` struct and `ServiceTokenRepository` interface
- [ ] 5.3 Wrap `ServiceTokenService.Create` Gorm writes in `Transaction(func(tx) { repo.WithTx(tx).Create(...); repo.WithTx(tx).SetTokenPermissions(...) })`, keep Casbin outside
- [ ] 5.4 Wrap `ServiceTokenService.Update` Gorm writes in `Transaction(func(tx) { repo.WithTx(tx).Update(...); repo.WithTx(tx).SetTokenPermissions(...) })`, keep Casbin outside
- [ ] 5.5 Wrap `ServiceTokenService.Delete` Gorm writes in `Transaction(func(tx) { repo.WithTx(tx).DeleteTokenPermissions(...); repo.WithTx(tx).Delete(...) })`, keep Casbin outside

## 6. Fix Repository — user_repository.go (2 methods)

- [ ] 6.1 Replace `tx.Begin()`/`tx.Commit()`/`tx.Rollback()` in `Create` with `e.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error { ... })`
- [ ] 6.2 Replace manual transaction management in `Update` with `Transaction(func(tx *gorm.DB) error { ... })`
- [ ] 6.3 Verify the refactored methods still return the same types (`*UserModel, error`)

## 7. Fix Repository — dept_repository.go (2 entry points)

- [ ] 7.1 Wrap `deptEntity.Create` multi-write (insert + path update) in `e.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error { ... })`
- [ ] 7.2 Wrap `deptEntity.Update` + `updateChildrenPath` multi-write in `Transaction(func(tx *gorm.DB) error { ... })`, making `updateChildrenPath` accept and pass `tx`
- [ ] 7.3 Remove `string(rune(child.ID))` path corruption in `updateChildrenPath` (use `strconv.FormatUint` if not already fixed)

## 8. Fix Repository — role_repository.go (1 method)

- [ ] 8.1 Wrap `roleRepo.UpdateRolePermission` permission creation loop in `e.conn.WithContext(ctx).Transaction(func(tx *gorm.DB) error { ... })`

## 9. Verify

- [ ] 9.1 Run `lsp_diagnostics` on all modified files
- [ ] 9.2 Run `go build ./cmd/server/` from `server/` to verify compilation
- [ ] 9.3 Run `go vet ./...` from `server/` to verify no new issues
- [ ] 9.4 Run `make test` from `server/` to verify existing tests pass
