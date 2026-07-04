# Apply Report — M3a

**Change**: change-24-vertical-slice-platform-review
**Work unit**: M3a — Migrate `iam/core` sentinels
**Mode**: Standard
**Date**: 2026-07-04

## Completed Tasks

- [x] M3a.1 Rewrite `internal/modules/iam/core/error.go` to declare 14 module-owned `Code` constants in `code:<snake_case>` format and construct `Err*` sentinels with `apperror.NotFound/Conflict/Forbidden/Unauthorized/BadRequest/Unprocessable`.
- [x] M3a.2 Add `internal/modules/iam/core/errors_test.go` table-driven across all sentinels (Status, Code format `code:` prefix, PublicMessage).
- [x] M3a.3 Add `internal/modules/iam/core/dto_test.go` table-driven for each DTO `Validate()` rule.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/iam/core/error.go` | Modified | Replaced plain `errors.New` sentinels with module-owned `apperror.Code` constants and `apperror.*` constructors |
| `internal/modules/iam/core/errors_test.go` | Created | Table-driven test pinning HTTP status, `code:` prefix, and public message for all 14 IAM sentinels plus code-uniqueness guard |
| `internal/modules/iam/core/dto_test.go` | Created | Table-driven coverage of every `Validate()` rule in the IAM core DTOs |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M3a tasks complete |

## Test Summary

- **Total tests written**: 2 new core test files (errors + DTO validation)
- **Total tests passing**: All IAM core tests pass
- **Layers used**: Unit
- **Approval tests**: None

## Verification Commands

```bash
go test ./internal/modules/iam/core/... -v    # PASS
go test ./internal/modules/iam/...            # PASS
go vet ./...                                  # PASS
go build ./...                                # PASS
go test ./...                                 # PASS
```

## Deviations from Design

None — implementation matches the design for module-owned sentinels. `ErrRoleHasAssignedUsers` is declared as `apperror.Unprocessable(...)` using the existing `core.MsgRoleHasAssignedUsers` public message, matching the M3a requirement.

## Issues Found

- `ErrInvalidCompanyScope` currently maps to `400 BadRequest` in `create_user`/`update_user` handlers but to `403 Forbidden` in `assign_role_to_user`. The M3a sentinel is `BadRequest` (400) to preserve the primary create/update user contract. When `mapServiceError` is removed in M3b, the `assign_role_to_user` handler test will need to be aligned to the single sentinel status or the slice will need a slice-local error for the forbidden case.
- No handler, service, or repository files were edited for M3a; existing `mapServiceError` switches remain intact for M3b/M3c/M3d.

## Remaining Tasks

- [ ] M3b — Migrate `iam/users` slices
- [ ] M3c — Migrate `iam/roles` slices
- [ ] M3d — Migrate `iam/permissions` + delete duplicate query
- [ ] M3e — Migrate `iam/internal` resolver slices + audit
- [ ] M4 — Migrate `companies` to `apperror` + shared helpers
- [ ] M5 — Migrate `auth` to `apperror`
- [ ] M6 — Wire `apperror` grep guard into Makefile + CI
- [ ] M7 — Publish `docs/module-error-conventions.md`

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M3a
- Boundary: M3a only; IAM core sentinels and core tests, no handler/service/repository migrations
- Estimated review budget impact: ~90 src / ~470 test as actual; within the 800-line work-unit budget

## Status

M3a complete. Ready for verify or next apply batch (M3b).
