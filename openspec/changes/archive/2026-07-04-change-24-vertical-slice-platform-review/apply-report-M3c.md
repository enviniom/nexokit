# Apply Report — M3c: Migrate `iam/roles` slices

## Status

success

## Executive Summary

Migrated the IAM roles slice handlers to route business errors directly through `response.HandleError(c, err)` and removed all handler-level `mapServiceError` switches and `platform/apperror` imports. The `delete_role` handler now relies on the module-owned `core.ErrRoleHasAssignedUsers` sentinel (added in M3a) instead of an inline `apperror.Wrap` call, preserving the existing 422 status and public message.

No service signatures, repositories, routes, payloads, DB schemas, or business behavior were changed.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/iam/roles/delete_role/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `errors`, `core`, and `apperror` imports. |
| `internal/modules/iam/roles/update_role/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `apperror` import; kept field-keyed `mapDomainErrorToValidation`. |
| `internal/modules/iam/roles/view_role/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `errors`, `core`, and `apperror` imports. |
| `internal/modules/iam/roles/view_role_permission_catalog/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `errors`, `core`, and `apperror` imports. |
| `internal/modules/iam/roles/assign_permissions_to_role/handler.go` | Modified | Removed `mapServiceError`; calls `response.HandleError(c, err)` directly; removed unused `errors` and `apperror` imports. |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M3c.1, M3c.2, and M3c.3 as complete. |

## Deviations from Design

None — implementation matches the design and the locked scope.

## Issues Found

None.

## Verification

| Command | Outcome |
|---------|---------|
| `go test ./internal/modules/iam/roles/...` | PASS |
| `go test ./internal/modules/iam/core/...` | PASS |
| `go test ./internal/modules/iam/...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |
| `grep -RE 'apperror\.|mapServiceError' internal/modules/iam/roles/ \| grep -v _test.go` | empty |
| `grep -RE 'apperror\.' internal/modules/iam/roles/ --include='*service.go' --include='*handler.go' \| grep -v _test.go` | empty |

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M3c
- Estimated M3c diff: ~5 handlers changed; ~81 changed lines (well under the 800-line budget)
- Boundary: Starts after M3b; ends before M3d.

## Risks

- The HTTP response `code` field now returns module-owned codes (e.g., `code:iam_resource_not_found`, `code:role_has_assigned_users`) instead of the previous platform HTTP-category codes. HTTP status, envelope shape, and public messages are preserved.
- No permissions or internal slices were touched, leaving M3d/M3e unchanged.

## Next Recommended

apply-next — proceed to M3d (`iam/permissions` + duplicate query deletion) once the M3c chained PR is reviewed/merged.
