## 1. Fix Auto-Migrate Logic Bug

- [x] 1.1 Invert condition in `server/cmd/server/main.go:49` — change `if global.TD27_CONFIG.System.DisableAutoMigrate` to `if !global.TD27_CONFIG.System.DisableAutoMigrate`

## 2. Fix Permission Query Column Name

- [x] 2.1 Change `type = 'menu'` to `domain = 'menu'` in `server/internal/model/sysManagement/role_repository.go:119`

## 3. Fix Department Path Corruption

- [x] 3.1 Add `strconv` import to `server/internal/model/sysManagement/dept_model.go`
- [x] 3.2 Replace `string(rune(d.ID))` with `strconv.FormatUint(uint64(d.ID), 10)` on lines 25 and 27

## 4. Fix SQL NULL Comparison

- [x] 4.1 Change `"deleted_at" = null` to `"deleted_at" IS NULL` in `server/internal/pkg/cache/pg_cache.go:63`

## 5. Fix Duplicate DeleteRoleMenu Call

- [x] 5.1 Replace the second `s.roleRepository.DeleteRoleMenu(s.ctx, id)` at `server/internal/service/sysManagement/role.go:71` with `global.TD27_DB.Where("role_id = ?", id).Delete(&modelSysManagement.RolePermissionModel{})`

## 6. Fix Logout Success Message

- [x] 6.1 Change `common.OkWithMessage("登出失败", c)` to `common.OkWithMessage("登出成功", c)` in `server/internal/api/sysManagement/login.go:157`

## 7. Verify

- [x] 7.1 Run `lsp_diagnostics` on all 6 changed files
- [x] 7.2 Run `go build ./cmd/server/` from `server/` to verify compilation
- [x] 7.3 Run `go vet ./...` from `server/` to verify no new issues
