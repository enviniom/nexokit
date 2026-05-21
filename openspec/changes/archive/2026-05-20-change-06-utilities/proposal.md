# Proposal: Change 06 — Utilities: DTOs, Validation, Pagination, Filters & GORM Helpers

## Intent

NexoKit modules duplicate pagination, filtering, sorting, and error-mapping logic. This change completes the platform utilities layer so every future module gets consistent, reusable building blocks instead of ad-hoc implementations.

## Scope

### In Scope
- Explicit base response DTO contracts: `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, `PaginationMeta`
- `FilterParams`, `PaginationParams`, `SortParams`, `SearchParams` structs and Gin parsers in `platform/query`
- New `platform/gormutil` package with `ApplyPagination`, `ApplySorting`, `ApplySearch`, `ApplyDateRange`, `ApplyStatusFilter`
- `PaginatedWithFilters` response helper in `platform/response`
- `HandleError(c, err)` centralized error→response mapper in `platform/response`
- `ErrValidation` sentinel (status 422) in `platform/apperror`
- Validation rules: `ValidSlug`, `ValidURL`, `InList` in `platform/validator`
- Validation responses by field only; `RespondIfInvalid` MUST call `response.ValidationError`
- Spanish message constants for new rules
- Update `users` and `companies` modules to adopt new helpers
- Convention documentation (module creation, DTOs, validation, pagination, filters, soft deletes, tenant scope, permissions, route registration)

### Out of Scope
- New business modules
- CLI or generator changes
- Auth/middleware changes
- Cache-related utilities

## Capabilities

### New Capabilities
- `query-parsing`: `PaginationParams`, `FilterParams`, `SortParams`, `SearchParams`, and Gin parsers for pagination, filters, sort, and search query parameters
- `gorm-helpers`: Pure-function GORM scopes for pagination, sorting, search, date range, and status filtering
- `api-conventions`: Documentation covering module creation, DTOs, validation, pagination, filters, soft deletes, tenant scope, permissions, and route registration patterns

### Modified Capabilities
- `api-response`: Explicit response DTO names, add `PaginatedWithFilters` (filter metadata in responses), and add `HandleError` (centralized error→response mapping)
- `request-validation`: Add `ValidSlug`, `ValidURL`, `InList` rules
- `error-handling`: Add `ErrValidation` sentinel error mapped to HTTP 422

## Approach

Extend `platform/query` with filter/sort/search parsing. Create `platform/gormutil` for GORM scope helpers (pure functions accepting `*gorm.DB` + params). Add `PaginatedWithFilters` alongside existing `Paginated` (backward-compatible delegation). Keep validation errors as `map[field][]messages` through `response.ValidationError` and `RespondIfInvalid`. Add `HandleError(c, err)` to `platform/response` as an explicit handler (not middleware). Add `ErrValidation` to `platform/apperror` with status 422. Extend validator with three new rules and matching Spanish messages. Update `users` and `companies` handlers to adopt the new packages. Document GORM soft delete conventions: base models use `gorm.DeletedAt`, normal queries exclude soft-deleted rows, delete endpoints soft-delete, and hard delete requires explicit documented exception.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/platform/query/query.go` | Modified | Add PaginationParams, FilterParams, SortParams, SearchParams, parsers |
| `internal/platform/gormutil/` | New | GORM scope helpers package |
| `internal/platform/response/response.go` | Modified | Add PaginatedWithFilters, HandleError |
| `internal/platform/apperror/apperror.go` | Modified | Add ErrValidation sentinel, Status 422 mapping |
| `internal/platform/validator/rules.go` | Modified | Add ValidSlug, ValidURL, InList |
| `internal/platform/messages/messages.go` | Modified | New Spanish message constants |
| `internal/modules/users/` | Modified | Adopt FilterParams, GORM helpers, HandleError |
| `internal/modules/companies/` | Modified | Adopt FilterParams, GORM helpers, HandleError |
| `docs/api-conventions.md` | New | Convention documentation, including response DTO names and soft deletes |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `Paginated` signature is breaking for callers | Low | Only 2 existing callers; update atomically; `PaginatedWithFilters` is additive |
| GORM helpers couple to `*gorm.DB` | Low | Pure-function scope pattern; testable with `gorm.DB` session mocks |
| Company-specific filters (`include_inactive`) leak into generic helpers | Medium | Generic helpers accept column names/config; domain-specific filters stay in module code |
| Search column configuration per module | Low | `ApplySearch` accepts variadic column names; no hardcoded columns |
| Soft delete expectations vary by module | Medium | Document default soft-delete behavior and require explicit opt-in for unscoped/hard-delete behavior |

## Rollback Plan

Revert module handler changes first (they're the consumers). Then remove `gormutil` package and `query` additions. `PaginatedWithFilters` and `HandleError` are additive — removing them is safe. `ErrValidation` removal requires reverting `Status()` mapping. No destructive schema changes.

## Dependencies

- Change 1 (platform-stubs, response, validator, apperror, query) must be complete
- GORM and Gin are already in the dependency graph

## Success Criteria

- [ ] `PaginatedWithFilters` includes filter metadata in response
- [ ] `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, and `PaginationMeta` are explicit and documented
- [ ] `HandleError` maps every sentinel to its HTTP status
- [ ] `ErrValidation` returns 422
- [ ] Validation errors are returned by field and `RespondIfInvalid` uses `response.ValidationError`
- [ ] `ValidSlug`, `ValidURL`, `InList` rules work with Spanish messages
- [ ] `PaginationParams`, `FilterParams`, `SortParams`, `SearchParams` parse from Gin context
- [ ] All 5 GORM helpers work as composable scopes
- [ ] `users` and `companies` list endpoints use shared helpers with no inline pagination
- [ ] Convention documentation exists, references running modules, and covers soft delete conventions
