## Context

The Go backend uses GORM as its ORM with PostgreSQL. All 8 N+1 patterns exist in the service and repository layers. No existing query batching utilities are shared — each module implements its own DB access pattern. The async logger already has batch size and flush interval constants defined (`batchSize=100`) but the `saveBatch` method ignores them.

## Goals / Non-Goals

**Goals:**
- Eliminate all 8 N+1 query patterns
- Reduce dashboard page-load DB queries from 6 to 1
- Use only GORM-native batching (no new dependencies)
- Preserve existing error handling semantics (partial failures logged, not fatal)

**Non-Goals:**
- No API or data model changes
- No new caching layer (dashboard caching deferred — 6→1 query is sufficient)
- No refactoring beyond the N+1 fixes

## Decisions

1. **Batch `WHERE id IN (?)` for role/permission lookups** — GORM natively supports slice parameters. Single query replaces loop with no transaction overhead. Alternatives (temp table, EXISTS subquery) add complexity without benefit.
2. **Subquery for service token API count** — GORM supports subqueries in SELECT via `Select("*, (SELECT count(*) FROM ...) as api_count")`. Alternative (preload) would require struct changes; subquery is minimal.
3. **Single query with UNION ALL for dashboard** — PostgreSQL can combine multiple COUNTs in one round-trip using `SELECT COUNT(*) FROM ... UNION ALL SELECT COUNT(*) FROM ...`. Alternative (GORM raw query) trades readability for performance — acceptable here since this is display-only.
4. **Batch insert via `db.Create(&slice)`** — GORM supports bulk insert by passing a slice. Need to add `BatchCreate` to operation log repository. Alternative (raw SQL INSERT) is less idiomatic.
5. **Transaction wrapper for recursive dept path update** — The `updateChildrenPath` is already recursive; wrapping in a transaction ensures atomicity. No change to the recursion itself — just add `Transaction` wrapper in the caller.

## Risks / Trade-offs

- **Casbin batch cleanup**: Casbin's `RemoveFilteredPolicy` only handles one policy at a time. No batch API exists in casbin. Acceptable — the N+1 is bounded by the number of APIs being deleted (not per-item in a list).
- **Dashboard UNION ALL query**: Returns multiple rows instead of columns. The Go code must scan them in order — fragile to field reordering. Mitigate with explicit ordering and named struct scan.
