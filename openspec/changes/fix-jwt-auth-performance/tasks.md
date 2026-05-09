## 1. JWT Auth Performance Fixes

- [ ] 1.1 Make `GetClaims` check `c.Get("claims")` before re-parsing JWT from header in `server/internal/api/sysManagement/claims_utils.go`
- [ ] 1.2 Change Casbin `Enforce` log level from Info to Debug in `server/internal/service/sysManagement/casbin.go:123`
- [ ] 1.3 Fix ServiceToken routes: `GET /delete` → `DELETE /delete`, `GET /update` → `PUT /update` in `server/internal/router/sysTool/service_token.go`
- [ ] 1.4 Cache JWT instance (package-level) to avoid `[]byte` re-allocation of signing key in `server/internal/pkg/jwt/jwt.go`

## 2. Verification

- [ ] 2.1 Verify with `lsp_diagnostics`, `go build`, and `go vet`
