# Apply Report — M5: Migrate `auth`

## Status

success

## Executive Summary

Migrated the `auth` module onto the module-owned error and persistence-boundary contracts introduced by change-24. Rewrote `internal/modules/auth/core/error.go` with `code:<snake_case>` constants (`code:invalid_credentials`, `code:invalid_refresh_token`) and `apperror.Unauthorized(...)` sentinels, added `core/errors_test.go`, `core/dto_test.go`, and `core/model_test.go`, and removed `gorm.io/gorm` and `platform/apperror` imports from the three auth service files (`authenticate_user`, `revoke_token`, `rotate_token`).

Slice repositories now translate `gorm.ErrRecordNotFound` to the appropriate module sentinel:

- `authenticate_user/repository.go` maps missing user/role rows to `core.ErrInvalidCredentials`.
- `revoke_token/repository.go` and `rotate_token/repository.go` map missing refresh tokens to `core.ErrInvalidRefreshToken`.

Services return module sentinels directly (`core.ErrInvalidCredentials` / `core.ErrInvalidRefreshToken`) for inactive users, wrong passwords, revoked tokens, expired tokens, and inactive owners. Handlers already routed errors through `response.HandleError`, so no handler production code changes were required. Handler tests were updated to assert that module sentinels produce HTTP 401 through `response.HandleError`.

### M5 corrective fix — preserve auth public unauthorized message

Fresh M5 verification discovered that the 401 response body had drifted: the `message` field changed from the prior platform generic `messages.MsgUnauthorized` (`"No autorizado"`) to the new module-owned English strings `"invalid credentials"` / `"invalid refresh token"`. The corrective fix keeps the module-owned codes and boundary improvements but restores the original public message by setting both sentinels' `PublicMessage` to `messages.MsgUnauthorized`. Body-level assertions were added to all three auth handler tests to lock in the preserved envelope shape.

The public HTTP contract is now fully preserved: invalid credentials and invalid refresh tokens still return HTTP 401 with `message: "No autorizado"`, `success: false`, `data: null`, `errors: null`, and no `debug` field in non-debug contexts. The `view_session` slice required no changes and has no leaked `apperror` import.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/auth/core/error.go` | Rewrote | Declared module-owned `CodeInvalidCredentials` / `CodeInvalidRefreshToken` constants and `apperror.Unauthorized(...)` sentinels. |
| `internal/modules/auth/core/error.go` | Modified (corrective fix) | Set `PublicMessage` for both sentinels to `messages.MsgUnauthorized` to preserve the original Spanish public 401 body. |
| `internal/modules/auth/core/errors_test.go` | Created | Table-driven coverage of status (401), `code:` prefix, public message, and uniqueness for every sentinel. |
| `internal/modules/auth/core/errors_test.go` | Modified (corrective fix) | Expect `messages.MsgUnauthorized` as the public message for both sentinels. |
| `internal/modules/auth/core/dto_test.go` | Created | Table-driven coverage of `Validate()` rules for `LoginRequest` and `RefreshRequest`. |
| `internal/modules/auth/core/model_test.go` | Created | Direct `TableName()` assertions for `AuthRole`, `AuthUser`, and `RefreshToken`. |
| `internal/modules/auth/authenticate_user/service.go` | Modified | Removed `gorm.io/gorm` and `platform/apperror` imports; returns `core.ErrInvalidCredentials` for inactive users and password mismatches. |
| `internal/modules/auth/authenticate_user/repository.go` | Modified | Translates `gorm.ErrRecordNotFound` to `core.ErrInvalidCredentials`. |
| `internal/modules/auth/authenticate_user/repository_test.go` | Modified | Added not-found regression test asserting `core.ErrInvalidCredentials`. |
| `internal/modules/auth/authenticate_user/service_test.go` | Modified | Updated fake repository and assertions to pivot from `apperror.ErrUnauthorized` to `core.ErrInvalidCredentials`. |
| `internal/modules/auth/authenticate_user/handler_test.go` | Modified | Updated unauthorized path to use `core.ErrInvalidCredentials`; added body-level envelope assertions. |
| `internal/modules/auth/authenticate_user/handler_test.go` | Modified (corrective fix) | Added `assertUnauthorizedEnvelope` helper verifying `success`, `message`, `data`, `errors`, and `debug` fields. |
| `internal/modules/auth/revoke_token/service.go` | Modified | Removed `gorm.io/gorm` and `platform/apperror` imports; returns `core.ErrInvalidRefreshToken` for revoked/expired tokens. |
| `internal/modules/auth/revoke_token/repository.go` | Modified | Translates `gorm.ErrRecordNotFound` to `core.ErrInvalidRefreshToken`. |
| `internal/modules/auth/revoke_token/repository_test.go` | Modified | Added not-found regression test asserting `core.ErrInvalidRefreshToken`. |
| `internal/modules/auth/revoke_token/service_test.go` | Modified | Updated fake repository and assertions to pivot from `apperror.ErrUnauthorized` to `core.ErrInvalidRefreshToken`. |
| `internal/modules/auth/revoke_token/handler_test.go` | Modified | Updated unauthorized path to use `core.ErrInvalidRefreshToken` and still assert HTTP 401. |
| `internal/modules/auth/revoke_token/handler_test.go` | Modified (corrective fix) | Added body-level envelope assertions via `assertUnauthorizedEnvelope`. |
| `internal/modules/auth/rotate_token/service.go` | Modified | Removed `gorm.io/gorm` and `platform/apperror` imports; returns `core.ErrInvalidRefreshToken` for revoked/expired/inactive-user cases. |
| `internal/modules/auth/rotate_token/repository.go` | Modified | Translates `gorm.ErrRecordNotFound` to `core.ErrInvalidRefreshToken`. |
| `internal/modules/auth/rotate_token/repository_test.go` | Modified | Added not-found regression test asserting `core.ErrInvalidRefreshToken`. |
| `internal/modules/auth/rotate_token/service_test.go` | Modified | Updated fake repository and assertions to pivot from `apperror.ErrUnauthorized` to `core.ErrInvalidRefreshToken`. |
| `internal/modules/auth/rotate_token/handler_test.go` | Modified | Updated unauthorized path to use `core.ErrInvalidRefreshToken` and still assert HTTP 401. |
| `internal/modules/auth/rotate_token/handler_test.go` | Modified (corrective fix) | Added body-level envelope assertions via `assertUnauthorizedEnvelope`. |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M5.1–M5.5 as complete. |

## Deviations from Design

None — implementation matches the design and the locked scope.

The auth module intentionally keeps a very small error vocabulary (only `ErrInvalidCredentials` and `ErrInvalidRefreshToken`) because all auth failures are deliberately indistinguishable to the client. Both sentinels are `apperror.Unauthorized` (401) to preserve the existing public HTTP contract.

The initial M5 implementation unintentionally changed the public `message` text from `"No autorizado"` to English module-owned strings. The corrective fix restores the original public message while keeping the module-owned codes, so the deviation was resolved before the work unit was accepted.

## Issues Found

1. **M5 verification failed on public 401 message drift.** The module-owned sentinels originally used English public messages (`"invalid credentials"`, `"invalid refresh token"`), changing the HTTP 401 response body compared to the previous platform generic `"No autorizado"`.
   - **Root cause:** New `apperror.Unauthorized(...)` sentinels specified a module-owned public message without checking the prior public payload.
   - **Resolution:** Set both sentinels' `PublicMessage` to `messages.MsgUnauthorized` and added body-level assertions in the three auth handler tests to prevent future drift.

## Verification

| Command | Outcome |
|---------|---------|
| `go test ./internal/modules/auth/...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |
| `grep -RE 'gorm\.\|apperror\.' internal/modules/auth/ --include='*service.go' --include='*handler.go' \| grep -v _test.go` | empty |
| `grep -RE 'mapServiceError' internal/modules/auth/ \| grep -v _test.go` | empty |

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M5 — Migrate `auth` (with M5 corrective fix included)
- Estimated M5 diff: ~19 files changed; ~240 insertions / ~56 deletions (~296 changed lines on tracked files; new core test files add ~160 lines), plus the corrective fix adds a small amount of test code, still well under the 800-line budget.
- Boundary: Starts after M4; ends before M6. No CI guard, no docs, no slice-folder migration.

## Risks

- `auth/container.go` still accepts `*gorm.DB` for repository construction; this is expected because repositories (not services) own the GORM dependency.
- The public error message for invalid credentials and invalid refresh tokens is now restored to the generic platform Spanish `"No autorizado"`. HTTP status remains 401 and the response envelope shape is unchanged.

## Next Recommended

verify — run the full M5 verification suite and proceed to the verify phase before M6.
