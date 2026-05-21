# Design: Change 06 — Utilities

## Technical Approach

Extend the existing platform utilities without changing module boundaries: `query` owns HTTP query parsing, new `gormutil` owns reusable GORM scopes, `response` owns explicit API DTO envelopes/error rendering, and `validator` owns field rules. `users` and `companies` become the reference consumers by replacing inline list/error logic with shared helpers while preserving GORM soft delete defaults.

## Architecture Decisions

| Decision | Choice | Alternative | Rationale |
|---|---|---|---|
| Query vs DB helpers | Keep parsers in `internal/platform/query`; create `internal/platform/gormutil`. | Put GORM helpers inside `query`. | Preserves focused platform subpackages and avoids coupling HTTP parsing to GORM. |
| Pagination metadata | Add `response.PaginatedWithFilters`; keep `Paginated` delegating with nil filters. | Change `Paginated` signature. | Keeps existing callers safe while enabling `meta.filters`. |
| Error mapping | Add explicit `response.HandleError(c, err)`. | Middleware or per-handler switches. | Keeps handlers explicit/testable and removes duplicated sentinel mapping. |
| Domain filters | Generic helpers accept columns/config; module-specific flags stay in module DTOs. | Encode `include_inactive` globally. | Prevents company business semantics from leaking into platform utilities. |
| Soft deletes | Document default `gorm.DeletedAt` behavior; helpers never call `Unscoped()`. | Add a generic include-deleted filter. | Keeps deleted records hidden unless a module explicitly documents an exception. |

## Data Flow

```txt
Gin handler ──→ query parsers ──→ module service ──→ repository
     │                  │                    │              │
     │                  └─ filters meta      │              └─ tenant.ApplyTenantScope + gormutil scopes
     └─ response.HandleError / PaginatedWithFilters ←───────┘
```

List endpoints parse `page`, `per_page`, `sort`, `order`, `search`, `status`, `created_from`, and `created_to`. Repositories apply tenant scope first, then status/date/search/sort/pagination scopes. Count uses the filtered query before pagination. Normal queries keep GORM's soft delete scope active.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/platform/query/query.go` | Modify | Add `PaginationParams`, `FilterParams`, `SortParams`, `SearchParams`, combined list params, Gin parsers, defaults. |
| `internal/platform/gormutil/gormutil.go` | Create | Add `ApplyPagination`, `ApplySorting`, `ApplySearch`, `ApplyDateRange`, `ApplyStatusFilter`. |
| `internal/platform/response/response.go` | Modify | Make `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, `PaginationMeta` explicit; add `PaginatedWithFilters` and `HandleError`. |
| `internal/platform/apperror/apperror.go` | Modify | Add `ErrValidation` mapped to 422. |
| `internal/platform/validator/rules.go` | Modify | Add `ValidSlug`, `ValidURL`, `InList`. |
| `internal/platform/messages/messages.go` | Modify | Add Spanish validation messages. |
| `internal/modules/users/{dto,handler,service,repository}.go` | Modify | Introduce list params, search/sort/filter support, shared pagination/error handling. |
| `internal/modules/companies/{dto,handler,service,repository}.go` | Modify | Compose generic filters with `include_inactive`; use shared response/error helpers. |
| `docs/api-conventions.md` | Create | Document module, DTO, validation, response, pagination, filters/search/sorting, soft deletes, tenant, permissions, routes. |

## Interfaces / Contracts

```go
type FilterParams struct { Status string; CreatedFrom, CreatedTo *time.Time }
type PaginationParams struct { Page, PerPage int }
type SortParams struct { Sort, Order string }
type SearchParams struct { Query string }
type ListParams struct { Pagination PaginationParams; Filters FilterParams; Sort SortParams; Search SearchParams }

func ApplySearch(db *gorm.DB, search SearchParams, columns ...string) *gorm.DB
func PaginatedWithFilters[T any](c *gin.Context, msg string, data T, pagination query.PaginationParams, filters query.FilterParams, sort query.SortParams, search query.SearchParams, total int64)
func HandleError(c *gin.Context, err error)
func RespondIfInvalid(c *gin.Context, errs response.ValidationErrors) bool // calls response.ValidationError
```

Response DTOs: `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, `PaginationMeta`.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | query parsing, validator rules, app error status, response DTOs/meta/error mapping | Table-driven tests; Gin `httptest` for response helpers. |
| Repository | GORM scopes, filtered counts, search/sort/date/status composition, soft delete defaults | SQLite-backed GORM tests with deterministic fixtures. |
| Handler/service | users and companies list params, `meta.filters`, centralized errors | Existing fake services plus `httptest`; preserve tenant-scope assertions. |
| Docs | convention doc examples remain discoverable | Lightweight docs test if links/required sections are already enforced. |

## Migration / Rollout

No database migration required. Existing `BaseModel`/`BaseModelSimple` already include `gorm.DeletedAt`. Roll out helpers first, then migrate `users` and `companies` atomically so the reference modules compile and tests pass together.

## Open Questions

None.
