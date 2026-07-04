# Apply Report — M3b

**Change**: change-24-vertical-slice-platform-review
**Work unit**: M3b — Migrate `iam/users` slices
**Mode**: Standard
**Date**: 2026-07-04

## Completed Tasks

- [x] M3b.1 For each of `iam/users/{create_user,update_user,delete_user,view_user,list_users,change_user_password,toggle_user_status,assign_role_to_user}/handler.go`: delete `mapServiceError`; call `response.HandleError(c, err)`; drop `platform/apperror` import.
- [x] M3b.2 Update `iam/users/create_user/repository.go` and `iam/users/update_user/repository.go` to call `gormutil.IsUniqueConstraintError` (M1) and translate to `core.ErrUserEmailAlreadyExists`.
- [x] M3b.3 Update slice services to return module sentinels only (`core.Err*` or `fmt.Errorf("...: %w", err)`); no `apperror.Wrap` inline.
- [x] M3b.4 Update each `service_test.go` to assert module sentinels via `errors.Is(err, core.Err*)`.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/iam/core/error.go` | Modified | Added `CodeForbiddenCompanyScope` and `ErrForbiddenCompanyScope` sentinel (403) to preserve the `assign_role_to_user` cross-company forbidden contract |
| `internal/modules/iam/core/errors_test.go` | Modified | Added table-driven coverage and code-uniqueness check for `ErrForbiddenCompanyScope` |
| `internal/modules/iam/users/create_user/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)`; dropped `platform/apperror` import; kept `mapDomainErrorToValidation` for duplicate email 422 |
| `internal/modules/iam/users/update_user/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)`; dropped `platform/apperror` import; kept `mapDomainErrorToValidation` for duplicate email 422 |
| `internal/modules/iam/users/delete_user/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)`; dropped `platform/apperror` import |
| `internal/modules/iam/users/view_user/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)`; dropped `platform/apperror` import |
| `internal/modules/iam/users/change_user_password/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)`; dropped `platform/apperror` import |
| `internal/modules/iam/users/toggle_user_status/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)`; dropped `platform/apperror` import |
| `internal/modules/iam/users/assign_role_to_user/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)`; dropped `platform/apperror` import |
| `internal/modules/iam/users/assign_role_to_user/service.go` | Modified | Returns `core.ErrForbiddenCompanyScope` for cross-company role assignment mismatch instead of `core.ErrInvalidCompanyScope` |
| `internal/modules/iam/users/assign_role_to_user/service_test.go` | Modified | Updated company-scope mismatch case to assert `errors.Is(err, core.ErrForbiddenCompanyScope)` |
| `internal/modules/iam/users/assign_role_to_user/handler_test.go` | Modified | Updated handler test case to expect 403 for `core.ErrForbiddenCompanyScope` |
| `internal/modules/iam/users/create_user/repository.go` | Modified | Replaced local `isUniqueConstraintError` with `gormutil.IsUniqueConstraintError` |
| `internal/modules/iam/users/update_user/repository.go` | Modified | Replaced local `isUniqueConstraintError` with `gormutil.IsUniqueConstraintError` |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M3b tasks complete |

## Test Summary

- **Total tests written**: 0 new test files; existing tests updated for new sentinel
- **Total tests passing**: All IAM users and core tests pass
- **Layers used**: Unit
- **Approval tests**: None

## Verification Commands

```bash
go test ./internal/modules/iam/users/...            # PASS
go test ./internal/modules/iam/core/...             # PASS
go test ./internal/modules/iam/...                  # PASS
go vet ./...                                        # PASS
go build ./...                                      # PASS
go test ./...                                       # PASS
grep -RE 'apperror\.' internal/modules/iam/users/ --include='*service.go' --include='*handler.go' | grep -v _test.go
# (empty)
```

## Deviations from Design

None — implementation matches the design. To preserve the locked public HTTP contract, a new module-owned sentinel `core.ErrForbiddenCompanyScope` (403) was introduced for the `assign_role_to_user` cross-company scope mismatch. The existing `core.ErrInvalidCompanyScope` (400) remains in use for `create_user` and `update_user` root-scope validation. No handler, service, or repository now imports `platform/apperror`.

## Issues Found

- Pre-existing status drift: `core.ErrInvalidCompanyScope` was mapped to 400 in `create_user`/`update_user` but to 403 in `assign_role_to_user`. This was resolved by adding `core.ErrForbiddenCompanyScope` for the assign-role cross-company forbidden case, keeping the 403 response without changing the 400 contract elsewhere.

## Remaining Tasks

- [ ] M3c — Migrate `iam/roles` slices
- [ ] M3d — Migrate `iam/permissions` + delete duplicate query
- [ ] M3e — Migrate `iam/internal` resolver slices + audit
- [ ] M4 — Migrate `companies` to `apperror` + shared helpers
- [ ] M5 — Migrate `auth` to `apperror`
- [ ] M6 — Wire `apperror` grep guard into Makefile + CI
- [ ] M7 — Publish `docs/module-error-conventions.md`

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M3b
- Boundary: M3b only; IAM users slice handlers, repositories, and the assign-role forbidden sentinel correction
- Estimated review budget impact: ~58 src / ~153 deletion as actual diff for M3b files; within the 800-line work-unit budget

## Status

M3b complete. Ready for verify or next apply batch (M3c).
