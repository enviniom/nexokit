# Apply Report — M3e: Migrate/audit `iam/internal` resolver slices

## Status

success

## Executive Summary

Audited the five IAM internal resolver slices (`list_all_permissions`, `resolve_auth_user`, `resolve_user_permissions`, `resolve_role_by_slug`, `sync_permissions`). The per-slice `service.go` files were already free of `platform/apperror` and `gorm.io/gorm` imports. The only GORM leak was in the internal wiring file `internal/modules/iam/internal/services.go`, whose constructors accepted `*gorm.DB` directly. Refactored those constructors to accept the slice `Repository` interfaces and moved repository instantiation into `internal/modules/iam/container.go`, which already owns the GORM dependency.

No `mapServiceError` functions exist in the internal package (there is no HTTP layer). Repositories that perform single-row lookups already translate `gorm.ErrRecordNotFound` to `core.ErrNotFound`; list queries that legitimately return empty slices continue to return raw DB errors, which services propagate unchanged.

Added regression coverage for error propagation and cache behavior across all five slices:

- `list_all_permissions`: repository error propagation.
- `resolve_auth_user`: generic repository error propagation.
- `resolve_role_by_slug`: generic repository error propagation.
- `resolve_user_permissions`: repository error propagation and cache-hit short-circuit.
- `sync_permissions`: `Create` and `AutoAssignToAdmins` error propagation.

No public routes, payloads, HTTP statuses, DB schemas, or business behavior were changed.

## Audit Findings

| Slice | `service.go` `gorm` import | `service.go` `apperror` import | `mapServiceError` | Action taken |
|---|---|---|---|---|
| `list_all_permissions` | None | None | None | Added error-propagation regression test. |
| `resolve_auth_user` | None | None | None | Added generic error-propagation regression test. |
| `resolve_user_permissions` | None | None | None | Added repo-error and cache-hit regression tests. |
| `resolve_role_by_slug` | None | None | None | Added generic error-propagation regression test. |
| `sync_permissions` | None | None | None | Added `Create`/`AutoAssignToAdmins` error-propagation tests. |
| `internal/services.go` | `*gorm.DB` parameter leak | None | N/A | Removed GORM import; constructors now accept repositories. |

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/iam/internal/services.go` | Modified | Removed `gorm.io/gorm` import; changed constructor signatures to accept slice `Repository` interfaces instead of `*gorm.DB`. |
| `internal/modules/iam/container.go` | Modified | Created repositories via `NewRepository(db)` and passed them into the refactored internal service constructors. |
| `internal/modules/iam/internal/list_all_permissions/service_test.go` | Modified | Added `err` field to fake repository and a repository-error propagation test. |
| `internal/modules/iam/internal/resolve_auth_user/service_test.go` | Modified | Added a generic repository-error propagation test. |
| `internal/modules/iam/internal/resolve_role_by_slug/service_test.go` | Modified | Added a generic repository-error propagation test. |
| `internal/modules/iam/internal/resolve_user_permissions/service_test.go` | Modified | Added repository `err` field, a repository-error propagation test, and a cache-hit short-circuit test. |
| `internal/modules/iam/internal/sync_permissions/service_test.go` | Modified | Added `Create` and `AutoAssignToAdmins` error-propagation tests. |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M3e.1–M3e.4 as complete. |

## Deviations from Design

None — implementation matches the design and the locked scope.

## Issues Found

None.

## Verification

| Command | Outcome |
|---------|---------|
| `go test ./internal/modules/iam/internal/...` | PASS |
| `go test ./internal/modules/iam/core/...` | PASS |
| `go test ./internal/modules/iam/...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |
| `grep -RE 'apperror\.|gorm\.' internal/modules/iam/internal/ --include='*service.go' \| grep -v _test.go` | empty |
| `grep -RE 'mapServiceError' internal/modules/iam/internal/ \| grep -v _test.go` | empty |

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M3e
- Estimated M3e diff: ~7 files changed; ~135 changed lines (well under the 800-line budget)
- Boundary: Starts after M3d; ends before M4.

## Risks

- The internal service wiring refactor changes constructor signatures, but the only caller (`container.go`) has been updated. No external callers are affected.
- No HTTP or public API behavior changed because the internal slices have no HTTP handlers.

## Next Recommended

apply-next — proceed to M4 (`companies` migration) once the M3e chained PR is reviewed/merged.
