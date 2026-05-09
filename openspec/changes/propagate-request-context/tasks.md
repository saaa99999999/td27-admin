## 1. Remove ctx field from service structs

- [ ] 1.1 `sysManagement/UserService` — remove `ctx context.Context` field
- [ ] 1.2 `sysManagement/RoleService` — remove `ctx context.Context` field
- [ ] 1.3 `sysManagement/MenuService` — remove `ctx context.Context` field
- [ ] 1.4 `sysManagement/ApiService` — remove `ctx context.Context` field
- [ ] 1.5 `sysManagement/ButtonService` — remove `ctx context.Context` field
- [ ] 1.6 `sysManagement/DeptService` — remove `ctx context.Context` field
- [ ] 1.7 `sysManagement/DictService` — remove `ctx context.Context` field
- [ ] 1.8 `sysManagement/DictDetailService` — remove `ctx context.Context` field
- [ ] 1.9 `sysManagement/RolePermissionService` — remove `ctx context.Context` field
- [ ] 1.10 `sysMonitor/OperationLogService` — remove `ctx context.Context` field
- [ ] 1.11 `sysTool/CronService` — remove `ctx context.Context` field
- [ ] 1.12 `sysTool/ServiceTokenService` — remove `ctx context.Context` field
- [ ] 1.13 `sysTool/FileService` — remove `ctx context.Context` field

## 2. Add ctx param to all service methods

- [ ] 2.1 `sysManagement/user.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.2 `sysManagement/role.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.3 `sysManagement/menu.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.4 `sysManagement/api.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.5 `sysManagement/button.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.6 `sysManagement/dept.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.7 `sysManagement/dict.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.8 `sysManagement/dict_detail.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.9 `sysManagement/role_permission.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.10 `sysMonitor/operation_log.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.11 `sysTool/cron.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.12 `sysTool/service_token.go` — add `ctx context.Context` first param to all public methods
- [ ] 2.13 `sysTool/file.go` — add `ctx context.Context` first param to all public methods

## 3. Fix JwtService context usage

- [ ] 3.1 `JwtService.AddToken` — change signature: add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 3.2 `JwtService.ValidateToken` — change signature: add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 3.3 `JwtService.RemoveToken` — change signature: add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 3.4 `JwtService.RemoveAllTokens` — change signature: add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 3.5 `JwtService.GetUserActiveSessions` — change signature: add `ctx context.Context` first param, remove both `context.Background()` calls
- [ ] 3.6 `JwtService.GetCachedUser` — change signature: add `ctx context.Context` first param, remove `context.Background()` call

## 4. Fix PermissionCache context usage

- [ ] 4.1 `PermissionCache.CacheUserPermissions` — add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 4.2 `PermissionCache.GetUserPermissions` — add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 4.3 `PermissionCache.ClearUserPermissions` — add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 4.4 `PermissionCache.CacheRolePermissions` — add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 4.5 `PermissionCache.GetRolePermissions` — add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 4.6 `PermissionCache.ClearRolePermissions` — add `ctx context.Context` first param, remove `context.Background()` call
- [ ] 4.7 `PermissionCache.CheckPermissionWithCache` — add `ctx context.Context` first param, pass through to `GetUserPermissions`

## 5. Update API handlers to pass c.Request.Context()

- [ ] 5.1 `api/sysManagement/user.go` — pass `c.Request.Context()` to all `userService` calls
- [ ] 5.2 `api/sysManagement/role.go` — pass `c.Request.Context()` to all `roleService` calls
- [ ] 5.3 `api/sysManagement/menu.go` — pass `c.Request.Context()` to all `menuService` calls
- [ ] 5.4 `api/sysManagement/api.go` — pass `c.Request.Context()` to all `apiService` calls
- [ ] 5.5 `api/sysManagement/button.go` — pass `c.Request.Context()` to all `service` calls
- [ ] 5.6 `api/sysManagement/dept.go` — pass `c.Request.Context()` to all `deptService` calls
- [ ] 5.7 `api/sysManagement/dict.go` — pass `c.Request.Context()` to all `dictService` calls
- [ ] 5.8 `api/sysManagement/dict_detail.go` — pass `c.Request.Context()` to all `dictDetailService` calls
- [ ] 5.9 `api/sysManagement/role_permission.go` — pass `c.Request.Context()` to `rolePermissionService.Rebuild`
- [ ] 5.10 `api/sysManagement/login.go` — pass `c.Request.Context()` to `logRegService.Login`, `jwtService.AddToken`, and local `jwtService.RemoveToken`
- [ ] 5.11 `api/sysMonitor/operation_log.go` — pass `c.Request.Context()` to all `operationLogService` calls
- [ ] 5.12 `api/sysMonitor/dashboard.go` — pass `c.Request.Context()` to all `dashboardService` calls
- [ ] 5.13 `api/sysTool/cron.go` — pass `c.Request.Context()` to all `cronService` calls
- [ ] 5.14 `api/sysTool/service_token.go` — pass `c.Request.Context()` to all `service` calls
- [ ] 5.15 `api/sysTool/file.go` — pass `c.Request.Context()` to all `fileService` calls
- [ ] 5.16 `middleware/jwt.go` — pass `c.Request.Context()` to `jwtService.ValidateToken`, `jwtService.RemoveToken`, `jwtService.AddToken`, and `serviceTokenService.AuthenticateToken`

## 6. Fix server.go timeout values

- [ ] 6.1 Change `ReadTimeout` from `120 * time.Second` to `30 * time.Second`
- [ ] 6.2 Change `WriteTimeout` from `120 * time.Second` to `60 * time.Second`
- [ ] 6.3 Add `IdleTimeout: 120 * time.Second`
- [ ] 6.4 Add `ReadHeaderTimeout: 20 * time.Second`

## 7. Verify

- [ ] 7.1 Run `go build ./...` — ensure no compilation errors
- [ ] 7.2 Run `grep -rn 'context\.Background()' server/internal/service/ server/internal/pkg/rbac/` — verify zero results
- [ ] 7.3 Run `make test` — ensure all tests pass
- [ ] 7.4 Run `make lint` — ensure no linting issues
