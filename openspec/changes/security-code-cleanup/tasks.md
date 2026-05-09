## 1. Security Fixes

- [ ] 1.1 Add bcrypt helper functions in a new `server/internal/pkg/crypto.go` (HashPassword, VerifyPassword, IsMD5Hash)
- [ ] 1.2 Update `login.go` — replace MD5 with bcrypt verification; add MD5 fallback + re-hash migration logic
- [ ] 1.3 Update `user_repository.go` — replace MD5V calls with bcrypt in Create (line 199), ModifyPasswd old password verify (line 294), and new password update (line 302)
- [ ] 1.4 Remove `server/internal/pkg/md5.go` and `server/internal/pkg/md5_test.go`; verify no remaining imports of `pkg.MD5V`
- [ ] 1.5 Add path traversal sanitization to `file.go:Download` — use `filepath.Clean` + prefix check against upload dir; reject attempts with `../`, absolute paths, or symlink escapes

## 2. Goroutine Safety & Performance

- [ ] 2.1 Fix `data_permission.go` goroutine leak — replace `go func() { time.Sleep; s.cache.Delete }` with TTL-aware cache: store expiry timestamp alongside value, check on read, skip expired entries
- [ ] 2.2 Fix `role.go` goroutine closure bug — shadow `err` by declaring `localErr := err` inside the `go func()` body
- [ ] 2.3 Optimize CORS whitelist — build `map[string]*configs.CORSWhitelist` once in `CorsByRules()`, replace `checkCors` linear scan with map lookup
- [ ] 2.4 Fix operation log memory buffering — limit `io.ReadAll` to 10KB; replace `json.Marshal(map)` on GET params with direct `url.Values` capture; add truncation indicator

## 3. Code Quality

- [ ] 3.1 Refactor `login.go` — add `*gorm.DB` field to `LogRegService`, inject via constructor, replace `global.TD27_DB` usage
- [ ] 3.2 Refactor `button.go` — add `*gorm.DB` field to `ButtonService`, inject via constructor, replace all `global.TD27_DB` usages
- [ ] 3.3 Refactor `dashboard.go` — add `*gorm.DB` field to `DashboardService`, inject via constructor, replace all `global.TD27_DB` usages
- [ ] 3.4 Remove dead code at: `role.go:64`, `router.go:33-35`, `permission_dto.go:3-6`, `permission_model.go:72-74`, `operation_log_repository.go:31-33`, `role_repository.go:107-108`, `server.go:28`

## 4. Verification

- [ ] 4.1 Run `lsp_diagnostics` on all modified files; fix any type errors or lint issues
- [ ] 4.2 Run `go build ./...` from `server/` directory — ensure compilation succeeds
- [ ] 4.3 Run `go vet ./...` from `server/` directory — ensure no vet warnings
- [ ] 4.4 Run `make test` — ensure all existing tests pass (update `md5_test.go` to test bcrypt functions)
