## Exploration: Change 06 — Utilities: DTOs, Validation, Pagination, Filters & GORM Helpers

### Current State

NexoKit already has foundational platform packages from Change 1:

- **`platform/response`** — Generic `APIResponse[T]`, `PaginationMeta`, `ValidationErrors`, and helpers (`Success`, `Created`, `Error`, `BadRequest`, `NotFound`, `Unauthorized`, `Forbidden`, `Conflict`, `InternalServerError`, `ValidationError`, `Paginated`, `RespondIfInvalid`). The `Paginated` function only emits `meta.pagination`; it does **not** include `meta.filters`.
- **`platform/apperror`** — `AppError` type, sentinels (`ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrConflict`, `ErrBadRequest`, `ErrInternal`), `Wrap`, `Status`, `PublicMessage`. No `ErrValidation` sentinel yet.
- **`platform/query`** — `Pagination` struct (Page, PerPage), `ParsePagination`, `PaginationFromGin`. No filter, sort, or search parsing.
- **`platform/validator`** — `Rule`, `FieldValidator`, chaining (`Required`, `Optional`, `Apply`), and 9 rules: `MinLength`, `MaxLength`, `ValidEmail`, `HasUppercase`, `HasDigit`, `HasSpecialChar`, `MinWords`, `NoNumbers`, `Matches`. Missing: `ValidSlug`, `ValidURL`, `InList`.
- **`platform/tenant`** — `TenantContext`, `WithCompany`, `ApplyTenantScope`, Gin helpers.
- **`platform/messages`** — All Spanish message constants.
- **`platform/identity`** — ULID `Generate()`.
- **`shared/model.go`** — `BaseModel` (with `DeletedAt`), `BaseModelSimple`.
- **Modules (users, companies)** — Each implements pagination inline (offset/limit calc, count query, ordering). Companies also has ad-hoc filter parsing (`include_inactive`, `status`) inside `ListCompaniesRequest` and its repository. No shared GORM helpers.

The handler error-mapping pattern is manual: each handler maps `apperror.Status(err)` to `response.*` calls individually. There is no centralized error→response handler.

### Affected Areas

- **`internal/platform/response/response.go`** — Needs `PaginatedWithFilters` or signature change to include filter metadata; `ErrValidation` integration.
- **`internal/platform/apperror/apperror.go`** — Needs `ErrValidation` sentinel and possibly a `HandleError` or middleware helper.
- **`internal/platform/query/query.go`** — Needs `FilterParams`, `SortParams`, `SearchParams`, `ParseFiltersFromGin`, and filter-parsing functions.
- **`internal/platform/validator/rules.go`** — Needs `ValidSlug`, `ValidURL`, `InList`, and any additional rules.
- **`internal/platform/messages/messages.go`** — New validation message constants.
- **New package: `internal/platform/gormutil/`** (or `internal/platform/query/` extended) — GORM helpers: `ApplyPagination`, `ApplySorting`, `ApplySearch`, `ApplyDateRange`, `ApplyStatusFilter`.
- **`internal/modules/users/`** — Must adopt `FilterParams`/`PaginationParams`, update `List`, add search/sort support, use GORM helpers.
- **`internal/modules/companies/`** — Must adopt `FilterParams`/`PaginationParams`, update `List`, use GORM helpers.
- **`internal/modules/companies/dto.go`** — `ListCompaniesRequest` must embed new filter/sort structs.
- **`internal/modules/users/handler.go`** — Must adopt centralized error handling and filter parsing.
- **`docs/`** — New conventions documentation file.

### Approaches

1. **Extend `platform/query` + new `platform/gormutil` package** — Add filter/sort/search parsing to `query`, create a separate `gormutil` package for GORM scope helpers, update `response.Paginated` to accept optional filters, add validation rules to existing `validator`, add `ErrValidation` to `apperror`, add centralized error handler.
   - Pros: Clean separation (query parsing is HTTP-agnostic, GORM helpers are DB-specific); minimal disruption to existing packages; each package stays focused.
   - Cons: Two new small packages to maintain; `Paginated` signature change is a breaking change requiring updates to callers.
   - Effort: Medium

2. **Consolidate everything into `platform/query`** — Add GORM helpers, filter structs, and sort structs into the existing `query` package; add `PaginatedWithFilters` to `response`; keep validator/apperror changes the same.
   - Pros: Single import for all query-related concerns; fewer packages to navigate.
   - Cons: `query` becomes too broad (HTTP parsing + GORM coupling in one package); violates the project's pattern of "small focused sub-packages" under `platform/`.
   - Effort: Medium

3. **Minimal: extend existing packages only, no new package** — Add GORM helpers as methods on `query.Pagination` and new structs in `query`; keep everything in existing packages; add `PaginatedWithFilters` as a separate function instead of changing `Paginated`.
   - Pros: Smallest diff; no structural changes; backward-compatible `Paginated`.
   - Cons: `query` package mixes parsing and GORM concerns; harder to test GORM helpers independently; doesn't scale well as more modules need helpers.
   - Effort: Low

### Recommendation

**Approach 1** — Extend `platform/query` for parsing + create `platform/gormutil` for GORM helpers. This aligns with the project convention ("`platform/` has small focused sub-packages") and keeps HTTP parsing decoupled from database operations. The `Paginated` function should gain an optional filters parameter via a `PaginatedWithFilters` function to preserve backward compatibility, then the old `Paginated` can delegate to it with `nil` filters. The centralized error handler should be a `response.HandleError(c, err)` function rather than middleware, keeping it testable and explicit.

### Risks

- **`Paginated` signature change** — Every existing call site (users, companies handlers) must be updated. Breaking change surface is small (2 modules) but must be done atomically.
- **GORM helper design** — Helpers must accept `*gorm.DB` and return `*gorm.DB` (scope pattern). If designed as structs or closures, testing is harder. Recommend pure functions (`func ApplyPagination(db *gorm.DB, p query.Pagination) *gorm.DB`).
- **Company-specific filter fields** — `include_inactive` is business-logic specific to companies. Generic helpers should not encode domain semantics. Modules should compose generic helpers with their own domain filters.
- **Pagination struct embedding** — Both `users` and `companies` use different patterns for pagination in DTOs. `companies.ListCompaniesRequest` embeds `query.Pagination` directly; `users.List` parses it in the handler. The new `FilterParams`/`SortParams` pattern must be consistent across both.
- **Search field design** — The search param needs to specify which database columns to search against. GORM `ApplySearch` must accept configurable column names, not hardcode them.

### Ready for Proposal

Yes — the scope is clear, the existing code is well-understood, and the affected packages are identified. The next phase should define the exact API surface for each new/extended package, the migration path for existing modules, and the documentation plan.