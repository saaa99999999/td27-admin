## Why

The backend performs N+1 database queries in 8 locations across user management, API management, service tokens, dashboard, audit logging, and department management. These scale linearly with data size — the dashboard alone issues 6 separate COUNT queries per page view. As the system grows, these patterns cause unnecessary database load and degrade API response times.

## What Changes

- **User Create/Update**: Replace per-role existence loop with batch `WHERE id IN (?)` queries
- **API DeleteByIds**: Batch permission deletion with `WHERE domain_id IN (?)` and single-pass Casbin policy cleanup
- **Service Token List**: Replace per-token COUNT loop with a subquery or batch COUNT
- **Dashboard Statistics**: Combine 6 individual COUNT queries into a single query or use caching
- **Async Logger saveBatch**: Batch-insert operation logs instead of one-at-a-time
- **Department updateChildrenPath**: Wrap recursive path updates in a single transaction
- All changes are internal implementation optimizations — no API contract or data model changes

## Capabilities

### New Capabilities
- `query-optimization`: Optimize N+1 query patterns across the backend to reduce database round-trips

### Modified Capabilities

*None — no existing specs change.*

## Impact

- **Backend**: 8 files across `service/sysManagement/`, `service/sysTool/`, `service/sysMonitor/`, `pkg/async/`, and `model/sysManagement/`
- **Dependencies**: Batch insert may need a repository method addition for operation logs
- **Performance**: Significant reduction in DB queries for paginated lists (service tokens), batch deletes (APIs), dashboard loads
- **No API/DB schema changes**: All changes are in service/repository layers
