## 1. User Service — Batch Role Existence Check

- [ ] 1.1 Fix `user.go Create`: Replace per-role `FindOne` loop with a single `WHERE id IN (?)` query to check all role IDs at once
- [ ] 1.2 Fix `user.go Update`: Same batch replacement for role existence check

## 2. API Service — Batch Delete Operations

- [ ] 2.1 Fix `api.go DeleteByIds`: Replace per-ID `DeleteByDomainID` loop with `WHERE domain_id IN (?)` batch delete
- [ ] 2.2 Fix `api.go DeleteByIds`: Remove the `RemoveResourcePolicy` per-API loop if Casbin supports batch; otherwise document as bounded N+1

## 3. Service Token — Eliminate Per-Token COUNT

- [ ] 3.1 Fix `service_token.go List`: Replace per-token `getTokenAPICount` loop with a correlated subquery or batch prefetch

## 4. Dashboard — Reduce 6 COUNT Queries

- [ ] 4.1 Fix `dashboard.go GetStatistics`: Combine 6 individual COUNT queries into a single query using UNION ALL or a raw query with multiple subqueries

## 5. Async Logger — Batch Insert

- [ ] 5.1 Add `BatchCreate` method to `operation_log_repository.go` supporting slice insert
- [ ] 5.2 Fix `async_logger.go saveBatch`: Call `BatchCreate` instead of looping `Create`

## 6. Department — Transactional Path Update

- [ ] 6.1 Fix `dept_repository.go updateChildrenPath`: Wrap the recursive update call in a `Transaction` at the caller level

## 7. Verify

- [ ] 7.1 Run `lsp_diagnostics`, `go build ./...`, and `go vet ./...` to verify no regressions
- [ ] 7.2 Run `make test` to verify existing tests pass
