# Apply Report — M3d: Migrate `iam/permissions` + delete duplicate query

## Status

success

## Executive Summary

Migrated the IAM permissions slice handlers to route business errors directly through `response.HandleError(c, err)` and removed all handler-level `mapServiceError` switches and `platform/apperror` imports. The `update_permission` repository now uses the shared `gormutil.IsUniqueConstraintError` helper instead of a local duplicate-key detection block.

Deleted the byte-identical duplicate `internal/modules/iam/queries/get_role_by_public_id_preloads.go`; the surviving caller in `assign_permissions_to_role/repository.go` now calls `queries.GetRoleByPublicID`, which already preloads `Company` and `Permissions`. Added regression coverage in `get_role_by_public_id_test.go` to pin the preload behavior for both associations plus the existing not-found path.

No service signatures, routes, payloads, HTTP statuses, DB schemas, or business behavior were changed.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/iam/permissions/update_permission/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `errors` and `apperror` imports. |
| `internal/modules/iam/permissions/view_permission/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `errors`, `core`, and `apperror` imports. |
| `internal/modules/iam/permissions/update_permission/repository.go` | Modified | Replaced local duplicate-key detection with `gormutil.IsUniqueConstraintError`; removed unused `strings` import. |
| `internal/modules/iam/queries/get_role_by_public_id_preloads.go` | Deleted | Byte-identical duplicate of `get_role_by_public_id.go`; removed. |
| `internal/modules/iam/queries/get_role_by_public_id_test.go` | Modified | Added regression cases for `Company` preload, `Permissions` preload, and preserved the existing not-found path. |
| `internal/modules/iam/roles/assign_permissions_to_role/handler.go` | Modified | Removed remaining `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `errors` and `apperror` imports. |
| `internal/modules/iam/roles/assign_permissions_to_role/repository.go` | Modified | Switched caller from deleted `GetRoleByPublicIDPreloads` to `GetRoleByPublicID`. |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M3d.1–M3d.5 as complete. |

## Deviations from Design

None — implementation matches the design and the locked scope.

## Issues Found

None.

## Verification

| Command | Outcome |
|---------|---------|
| `go test ./internal/modules/iam/permissions/...` | PASS |
| `go test ./internal/modules/iam/queries/...` | PASS |
| `go test ./internal/modules/iam/core/...` | PASS |
| `go test ./internal/modules/iam/...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |
| `grep -RE 'apperror\.' internal/modules/iam/permissions/ --include='*service.go' --include='*handler.go' \| grep -v _test.go` | empty |
| `grep -RE 'mapServiceError' internal/modules/iam/permissions/ \| grep -v _test.go` | empty |
| `grep -RE 'GetRoleByPublicIDPreloads' internal/modules/` | empty |
| `test -f internal/modules/iam/queries/get_role_by_public_id_preloads.go` | deleted |

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M3d
- Estimated M3d diff: ~6 files changed; ~167 changed lines (well under the 800-line budget)
- Boundary: Starts after M3c; ends before M3e.

## Risks

- The HTTP response `code` field now returns module-owned codes (e.g., `code:iam_resource_not_found`, `code:system_immutable`, `code:iam_resource_conflict`) instead of the previous platform HTTP-category codes. HTTP status, envelope shape, and public messages are preserved.
- `list_permissions/handler.go` already used `response.HandleError(c, err)` directly and required no changes.

## Next Recommended

apply-next — proceed to M3e (`iam/internal` resolver slices + audit) once the M3d chained PR is reviewed/merged.
