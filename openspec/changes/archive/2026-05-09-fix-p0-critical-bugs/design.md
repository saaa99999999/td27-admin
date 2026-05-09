## Context

A comprehensive backend codebase audit identified 6 P0 bugs. Each is a single-line fix with zero architectural complexity. No new dependencies, no schema changes, no API contract changes. All bugs are in the Go backend (`server/` directory).

The current codebase state is "disciplined" — consistent patterns, proper logging, structured architecture. These bugs are isolated mistakes in an otherwise well-structured codebase.

## Goals / Non-Goals

**Goals:**
- Fix all 6 P0 bugs with minimal, safe one-line changes
- Maintain backward compatibility — zero behavioral changes beyond fixing broken behavior
- Verify each fix compiles and diagnostics are clean

**Non-Goals:**
- No refactoring or improvement beyond the bug fix
- No test additions (existing test coverage is adequate for regression)
- No config or schema changes

## Decisions

### Bug 1: Inverted auto-migrate condition (`main.go:49`)
- **Fix**: Change `if global.TD27_CONFIG.System.DisableAutoMigrate` → `if !global.TD27_CONFIG.System.DisableAutoMigrate`
- **Rationale**: Config key is `disable-auto-migrate: true` (default). When it's `true`, migration should be SKIPPED, not run. This is a simple boolean negation fix.
- **Alternative considered**: Rename the config key — rejected because it would break existing deployments.

### Bug 2: Non-existent column `type` in query (`role_repository.go:119`)
- **Fix**: Change `type = 'menu'` → `domain = 'menu'`
- **Rationale**: `PermissionModel` uses field `Domain` (column: `domain`) to store the permission type. The old `type` column was from a previous schema version and doesn't exist.
- **Risk**: Low — the column `domain` exists and is populated for all records.

### Bug 3: Department path corruption via `rune()` (`dept_model.go:25-27`)
- **Fix**: Change `string(rune(d.ID))` → `strconv.FormatUint(uint64(d.ID), 10)`
- **Rationale**: `string(rune(uint))` converts the integer to a Unicode code point — for IDs > 127, this produces multi-byte UTF-8 characters (e.g., ID=200 → `È`). The materialized path pattern (`/1/200/`) relies on ASCII-safe path segments for `LIKE` matching.
- **New import needed**: `strconv` (stdlib, no new dependency)

### Bug 4: SQL NULL comparison (`pg_cache.go:63`)
- **Fix**: Change `"deleted_at" = null` → `"deleted_at" IS NULL`
- **Rationale**: SQL uses `IS NULL` / `IS NOT NULL` for null comparisons. `= null` always evaluates to false. This means the UPSERT's UPDATE branch never matches soft-deleted records, always falling through to INSERT and potentially causing duplicate key violations.

### Bug 5: Duplicate `DeleteRoleMenu` call (`role.go:71`)
- **Fix**: Replace the second `s.roleRepository.DeleteRoleMenu(s.ctx, id)` with `global.TD27_DB.Where("role_id = ?", id).Delete(&modelSysManagement.RolePermissionModel{})`
- **Rationale**: `DeleteRoleMenu` only deletes menu-type role-permission records (filtered by `domain = 'menu'`). The second call's intent is to delete ALL role-permission associations for the deleted role. A direct delete on `RolePermissionModel` unconditionally cleans up all associations.
- **Safety**: Since the role is being deleted at line 58, all role-permission records are orphaned regardless of domain type.

### Bug 6: Hardcoded "登出失败" message (`login.go:157`)
- **Fix**: Change `common.OkWithMessage("登出失败", c)` → `common.OkWithMessage("登出成功", c)`
- **Rationale**: The function is in the success path (after `c.Next()` in the middleware chain). The message literally says "logout failed" while the operation succeeded.
- **Note**: This is a display-level bug — only affects the message shown to the user, not any logic or state.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Bug 1: Enabling auto-migrate could cause issues if schema has manual changes | The feature is disabled by default. The fix only corrects the boolean semantics — users who explicitly set `disable-auto-migrate: false` will now correctly get migration. |
| Bug 3: Dept paths with existing corrupted data need migration | No migration needed. The fix only affects new path generation. Existing corrupted paths will remain but can be rebuilt on next department update (path is regenerated on save). |
| Bug 5: Direct DB delete bypasses repository abstraction | Acceptable trade-off for a minimal fix. A follow-up can add a proper `DeleteRolePermissions` method to the repository interface. |
| All changes are untested | Each fix is a one-line change with obvious correctness. `lsp_diagnostics` and `go build` will verify compilation. |
