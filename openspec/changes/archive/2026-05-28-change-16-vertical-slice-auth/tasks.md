# Tasks: Migrate Auth Module to Vertical Slice Architecture

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Est. changed lines | 900–1200 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (core+queries) → PR 2 (authenticate_user) → PR 3 (rotate+revoke) → PR 4 (view+wiring+cleanup) |
| Delivery strategy | ask-always |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | PR | Base |
|------|------|-----|------|
| 1 | core + queries | PR 1 | main |
| 2 | authenticate_user | PR 2 | PR 1 |
| 3 | rotate + revoke | PR 3 | PR 2 |
| 4 | view_session + wiring | PR 4 | PR 3 |

## Phase 1: Foundation

- [x] 1.1 Create `internal/modules/auth/core/model.go` — `AuthUser`, `AuthRole`, `RefreshToken` partial models
- [x] 1.2 Create `internal/modules/auth/core/dto.go` — request/response DTOs, `AuthUserResponse`
- [x] 1.3 Create `internal/modules/auth/core/error.go` — module error constants
- [x] 1.4 Create `internal/modules/auth/queries/find_user_by_email.go` + `_test.go` — SQLite test with local AutoMigrate
- [x] 1.5 Create `internal/modules/auth/queries/find_user_by_id_with_role.go` + `_test.go` — preload user+role test
- [x] 1.6 Verify `go test ./internal/modules/auth/queries/...` passes

## Phase 2: `authenticate_user` Slice

- [x] 2.1 Create `authenticate_user/repository.go` + `_test.go` — delegates to `queries.FindUserByEmail`
- [x] 2.2 Create `authenticate_user/service.go` + `_test.go` — login; cases: success, inactive, wrong creds, missing
- [x] 2.3 Create `authenticate_user/handler.go` + `_test.go` — POST /auth/login; cases: validation, 401, 200, no password leak
- [x] 2.4 Wire slice into `container.go`

## Phase 3: `rotate_token` and `revoke_token` Slices

- [x] 3.1 Create `rotate_token/repository.go` + `_test.go` — delegates to `queries.FindRefreshTokenByHashWithUser`
- [x] 3.2 Create `rotate_token/service.go` + `_test.go` — rotation; cases: success, revoked, expired
- [x] 3.3 Create `rotate_token/handler.go` + `_test.go` — POST /auth/refresh
- [x] 3.4 Create `revoke_token/repository.go` + `_test.go` — delegates to `queries.FindRefreshTokenByHashWithUser`
- [x] 3.5 Create `revoke_token/service.go` + `_test.go` — revocation; cases: success, already revoked, invalid
- [x] 3.6 Create `revoke_token/handler.go` + `_test.go` — POST /auth/logout
- [x] 3.7 Wire both slices into `container.go`

## Phase 4: `view_session` Slice

- [x] 4.1 Create `view_session/handler.go` + `_test.go` — reads `authctx.FromGin`; cases: authenticated, root perms, no perms
- [x] 4.2 Create `view_session/service.go` + `repository.go` — minimal/pass-through
- [x] 4.3 Wire slice into `container.go`

## Phase 5: Wiring

- [x] 5.1 Complete `internal/modules/auth/container.go` — `NewContainer(db, verifier, issuer, refreshManager, refreshTTL)`
- [x] 5.2 Update `internal/modules/auth/routes.go` — `Register(v1, container, ...)` dispatches to slice handlers
- [x] 5.3 Update `internal/app/container.go` — replace `auth.NewService/NewHandler` with `auth.NewContainer`; keep `userLookup`
- [x] 5.4 Run `go test ./internal/modules/auth/...` — all 13+ cases pass

## Phase 6: Cleanup and Verification

- [x] 6.1 Delete legacy: `handler.go`, `service.go`, `repository.go`, `model.go`, `dto.go` and tests
- [x] 6.2 Verify zero `modules/users` or `modules/roles` imports in auth source
- [x] 6.3 Run `go test ./...` — zero behavior change
- [x] 6.4 Verify all 4 endpoints respond identically

> **NO-APPLY**: Maintainer MUST review before `sdd-apply`.
