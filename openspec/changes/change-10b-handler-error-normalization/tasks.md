# Tasks: Handler error and list metadata normalization

## Review workload forecast

| Field | Value |
|---|---|
| Estimated changed lines | ~180–260 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single review unit |

Decision needed before apply: Yes — maintainer reviews this task plan before accepting the already-produced diff.

> Note: These tasks were reconstructed after an implementation was mistakenly applied before the maintainer review checkpoint. Use this file as the review checklist before deciding whether to keep, adjust, or revert the diff.

## Phase 1: Error handling normalization

- [x] 1.1 Update `internal/modules/auth/handler.go` to remove `respondError` and use `response.HandleError` in `Login`, `Refresh`, and `Logout`.
- [x] 1.2 Update `internal/modules/permissions/handler.go` to remove `writePermissionError` and use `response.HandleError` for service errors.
- [x] 1.3 Update `internal/modules/roles/handler.go` to replace the `Delete` `apperror.Status` switch with `response.HandleError`.
- [x] 1.4 Confirm no handlers retain `switch apperror.Status(err)` or obsolete `apperror` imports.

## Phase 2: List metadata normalization

- [x] 2.1 Update `roles.List` handler to use `query.ListFromGin` and `response.PaginatedWithFilters`.
- [x] 2.2 Update `permissions.ListPaginated` handler to use `query.ListFromGin` and `response.PaginatedWithFilters`.
- [x] 2.3 Update roles service `List` signature to accept `query.ListParams` while preserving repository pagination behavior.
- [x] 2.4 Update permissions service `List` signature to accept `query.ListParams` while preserving repository pagination behavior.
- [x] 2.5 Update `response.PaginatedWithFilters` to accept `query.ListParams` instead of decomposed pagination/filter/sort/search arguments, and update all call sites.

## Phase 3: Validation normalization

- [x] 3.1 Standardize auth handler validation on `response.RespondIfInvalid`.
- [x] 3.2 Standardize users handler validation on `response.RespondIfInvalid`.
- [x] 3.3 Standardize roles handler validation on `response.RespondIfInvalid`.
- [x] 3.4 Standardize permissions handler validation on `response.RespondIfInvalid`.

## Phase 4: Tests

- [x] 4.1 Update roles handler fake service and assert `meta.filters` exists in list responses.
- [x] 4.2 Update roles service tests to pass `query.ListParams`.
- [x] 4.3 Add focused permissions handler tests for list metadata and shared error handling paths.
- [x] 4.4 Update response helper tests for the `query.ListParams` signature.

## Phase 5: Verification

- [x] 5.1 Run narrow touched-package tests.
- [x] 5.2 Run `go test ./...`.
- [x] 5.3 Run `go build ./...`.
- [x] 5.4 Run fresh diff review before commit.

## Acceptance checklist

- [x] No `switch apperror.Status(err)` in handlers.
- [x] Service errors use `response.HandleError`.
- [x] Local wrappers removed except justified module-specific mappings.
- [x] `companies.respondError` remains only for `ErrDuplicateSlug` field mapping.
- [x] Roles and permissions paginated lists include consistent filters metadata.
- [x] `PaginatedWithFilters` call sites pass `query.ListParams` directly.
- [x] Request validation uses `response.RespondIfInvalid` consistently.
- [x] No unrelated `.atl/*` files are included in the review unit.
