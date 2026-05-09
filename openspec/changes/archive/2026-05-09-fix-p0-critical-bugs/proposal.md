## Why

The backend has 6 critical P0 bugs that cause data corruption, runtime query failures, and broken behavior. These are the highest-priority issues identified in a comprehensive codebase audit — they block any further optimization work until fixed. Each bug has a clear, minimal fix with no side effects.

## What Changes

- **Fix inverted auto-migrate condition** — `main.go:49` currently runs `RegisterTables` when `disable-auto-migrate: true` (the default). Fix the boolean check.
- **Fix non-existent column in query** — `role_repository.go:119` queries `type = 'menu'` but the column is `domain`. Causes runtime SQL errors when loading menu permissions.
- **Fix department path corruption** — `dept_model.go:25-27` uses `string(rune(id))` which produces multi-byte UTF-8 for IDs > 127, breaking materialized path matching.
- **Fix SQL NULL comparison** — `pg_cache.go:63` uses `= null` instead of `IS NULL`, causing the UPSERT UPDATE branch to never match soft-deleted records.
- **Fix duplicate `DeleteRoleMenu` call** — `role.go:71` calls `DeleteRoleMenu` a second time instead of the intended permission cleanup function.
- **Fix hardcoded error message** — `login.go:157` returns "登出失败" (logout failed) even on successful logout.

## Capabilities

### New Capabilities
- `critical-bug-fixes`: Fixes for 6 P0 bugs in the backend server code affecting data integrity, query correctness, and API responses.

### Modified Capabilities
None — these are internal bug fixes with no spec-level behavior changes.

## Impact

- **Files modified**:
  - `server/cmd/server/main.go` (1 line — flip boolean)
  - `server/internal/model/sysManagement/role_repository.go` (1 line — column name)
  - `server/internal/model/sysManagement/dept_model.go` (1 line — string conversion)
  - `server/internal/service/sysManagement/role.go` (1 line — remove duplicate call)
  - `server/internal/pkg/cache/pg_cache.go` (1 line — null comparison)
  - `server/internal/api/sysManagement/login.go` (1 line — error message)
- **No new dependencies**. Each fix is a one-line change.
- **No API contract changes** — all fixes are internal.
- **No schema migrations needed**.
