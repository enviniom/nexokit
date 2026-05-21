# Tasks: Change 06 — Utilities

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 750-1000 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 platform utilities → PR 2 users/companies adoption → PR 3 docs |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Platform query, GORM, response, error, validator utilities | PR 1 | Base main; include unit/httptest/SQLite tests. |
| 2 | Adopt utilities in `users` and `companies` list/delete flows | PR 2 | Depends on PR 1; include handler/service/repository tests. |
| 3 | Publish API conventions | PR 3 | Depends on PR 2 examples; docs-only verification. |

## Phase 1: Query and GORM Utilities

- [x] 1.1 RED: add `internal/platform/query/query_test.go` for pagination defaults, `FilterParams` dates as `*time.Time`, `SortParams`, `SearchParams.Query`, and `ListFromGin`.
- [x] 1.2 GREEN: update `internal/platform/query/query.go` with `PaginationParams`/`Pagination`, `FilterParams`, `SortParams`, `SearchParams`, `ListParams`, and Gin parsers.
- [x] 1.3 RED: add `internal/platform/gormutil/gormutil_test.go` for `ApplySorting` allowlist/defaults, `ApplySearch(db, search, columns...)`, date range, status, pagination, and no-op cases.
- [x] 1.4 GREEN: create `internal/platform/gormutil/gormutil.go`; helpers must compose without `Unscoped()` so GORM soft deletes remain active.

## Phase 2: Responses, Errors, Validation

- [x] 2.1 RED/GREEN: update `internal/platform/response/{response.go,response_test.go}` for `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, `PaginationMeta`, and concrete `PaginatedWithFilters` params.
- [x] 2.2 RED/GREEN: verify validation errors are field-keyed and `RespondIfInvalid` calls `response.ValidationError` for 422 responses.
- [x] 2.3 RED/GREEN: update `internal/platform/apperror/{apperror.go,apperror_test.go}` and `response.HandleError` tests for `ErrValidation` 422 plus 401/403/404/409/500 mappings.
- [x] 2.4 RED/GREEN: update `internal/platform/messages/messages.go` and `internal/platform/validator/{rules.go,rules_test.go}` for `ValidSlug`, `ValidURL`, `InList`, Spanish messages, and optional-field skip.

## Phase 3: Reference Module Adoption

- [x] 3.1 RED: extend `internal/modules/users/*_test.go` for `page/per_page`, status/date filters, search, sorting, `meta.filters`, `HandleError`, and soft-delete delete behavior.
- [x] 3.2 GREEN: update `internal/modules/users/{dto.go,handler.go,service.go,repository.go}` to use shared query params, gormutil scopes, `PaginatedWithFilters`, tenant scope, and centralized errors.
- [x] 3.3 RED: extend `internal/modules/companies/*_test.go` for shared filters/search/sorting plus `include_inactive` remaining module-specific and soft delete defaults.
- [x] 3.4 GREEN: update `internal/modules/companies/{dto.go,handler.go,service.go,repository.go}` to compose generic helpers with company-specific inactive filtering.

## Phase 4: Docs and Verification

- [x] 4.1 Create `docs/api-conventions.md` covering module files, explicit response DTO names, validation, field errors, pagination, filters/search/sorting, soft deletes, tenant scope, permissions, and route registration.
- [x] 4.2 Run `go test ./...` and verify all spec scenarios pass without auth/middleware or schema changes.
