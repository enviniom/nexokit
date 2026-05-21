# API conventions for modules

Use this guide when adding or reviewing NexoKit API modules. New modules should follow the same shape as the `users` and `companies` modules: explicit DTOs, field-keyed validation, shared list query helpers, tenant-aware repositories, soft deletes by default, and route-level permission guards.

## Quick path

1. Create the module files under `internal/modules/{name}/`.
2. Define request/response DTOs and `Validate()` methods in `dto.go`.
3. Parse list query params with `query.ListFromGin(c)` in handlers.
4. Apply tenant scope and GORM helpers in repositories.
5. Return responses through `platform/response` helpers.
6. Register routes from the app container under `/api/v1` with permission guards.

## Module file layout

Every business module SHOULD keep the same flat file layout:

| File | Responsibility |
|------|----------------|
| `model.go` | GORM model and persistence fields. |
| `dto.go` | Request and response DTOs plus validation methods. |
| `handler.go` | Gin HTTP handlers, binding, tenant extraction, and response rendering. |
| `service.go` | Business use cases and transaction-free orchestration. |
| `repository.go` | Database queries, tenant filtering, GORM helper composition. |
| `routes.go` | Module route registration and permission/role guards. |
| `validation.go` | Optional validation helpers when DTO validation outgrows `dto.go`. |

Place the files in `internal/modules/{name}/`. Expose a `Register` function from `routes.go`, then call it from `internal/app/container.go`. The server mounts the container under `/api/v1` in `internal/server/router.go`.

```go
func Register(v1 *gin.RouterGroup, handler *Handler, requirePermission func(string) gin.HandlerFunc) {
	products := v1.Group("/products")
	products.GET("", requirePermission("products.index"), handler.List)
	products.POST("", requirePermission("products.create"), handler.Create)
}
```

## DTOs and response envelopes

Use explicit DTO names so API contracts are visible at the module boundary:

| DTO kind | Naming convention | Example |
|----------|-------------------|---------|
| Create request | `Create{Name}Request` | `CreateUserRequest` |
| Update request | `Update{Name}Request` | `UpdateCompanyRequest` |
| Read response | `{Name}Response` | `UserResponse` |
| Action request | `{Action}{Name}Request` | `ChangePasswordRequest` |

Platform response envelopes are named explicitly and should be referenced in docs, tests, and reviews:

- `response.APIResponse`
- `response.ErrorResponse`
- `response.ValidationErrorResponse`
- `response.PaginatedResponse`
- `response.PaginationMeta`

Do not expose internal database IDs unless the module explicitly requires it. Prefer public IDs or slugs in routes and response DTOs. Do not expose `DeletedAt` in API DTOs.

## Validation and field errors

Request DTOs that need validation MUST expose:

```go
func (r CreateProductRequest) Validate() response.ValidationErrors {
	errs := make(response.ValidationErrors)
	validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
	validator.Field(errs, "slug", r.Slug).Required().Apply(validator.ValidSlug())
	validator.Field(errs, "status", r.Status).Optional().Apply(validator.InList("active", "inactive"))
	return errs
}
```

Handlers SHOULD call `validator.RespondIfInvalid(c, req.Validate())` immediately after binding. Validation failures are returned as HTTP 422 with `errors` keyed by field name:

```json
{
  "success": false,
  "message": "Error de validación",
  "errors": {
    "email": ["es requerido"]
  }
}
```

Use the shared rule chain for normal field checks:

- `Field(errs, name, value).Required()` for required values.
- `.Optional().Apply(rule)` when empty values should skip later rules.
- `ValidSlug()`, `ValidURL()`, and `InList(...)` for common string constraints.

## Error handling

Handlers SHOULD use centralized error mapping instead of local `switch apperror.Status(err)` blocks:

```go
item, err := h.service.GetByPublicID(c.Request.Context(), tc, id)
if err != nil {
	response.HandleError(c, err)
	return
}
```

`response.HandleError` maps known `apperror` sentinels to the standard API envelope, including `ErrUnauthorized` 401, `ErrForbidden` 403, `ErrNotFound` 404, `ErrConflict` 409, `ErrValidation` 422, and unknown errors as 500.

## Pagination, filters, search, and sorting

List handlers SHOULD parse query params once and pass the value through service and repository layers:

```go
params := query.ListFromGin(c)
items, total, err := h.service.List(c.Request.Context(), tc, params)
if err != nil {
	response.HandleError(c, err)
	return
}
response.PaginatedWithFilters(c, "Products retrieved", items, params.Pagination, params.Filters, params.Sort, params.Search, total)
```

Supported shared query params:

| Query param | Parsed into | Notes |
|-------------|-------------|-------|
| `page` | `PaginationParams.Page` | 1-indexed; invalid values normalize to defaults. |
| `per_page` | `PaginationParams.PerPage` | Bounded by the platform maximum. |
| `status` | `FilterParams.Status` | Module decides which DB column or boolean mapping applies. |
| `created_from` | `FilterParams.CreatedFrom` | Date format: `YYYY-MM-DD`; invalid dates are ignored. |
| `created_to` | `FilterParams.CreatedTo` | Date format: `YYYY-MM-DD`; invalid dates are ignored. |
| `sort` | `SortParams.Sort` | Repositories MUST restrict sortable columns with an allowlist. |
| `order` | `SortParams.Order` | Only `asc` and `desc`; invalid values default to `desc`. |
| `search` | `SearchParams.Query` | Repository chooses searchable columns. |

Repositories SHOULD compose GORM helpers after tenant scope and before pagination:

```go
db := tenant.ApplyTenantScope(r.db.Model(&Product{}), tc)
db = gormutil.ApplyStatusFilter(db, params.Filters, "status")
db = gormutil.ApplyDateRange(db, params.Filters, "created_at")
db = gormutil.ApplySearch(db, params.Search, "name", "slug")
db = gormutil.ApplySorting(db, params.Sort, "created_at", "name")

var total int64
if err := db.Count(&total).Error; err != nil { return nil, 0, err }
db = gormutil.ApplyPagination(db, params.Pagination.Page, params.Pagination.PerPage)
```

Keep module-specific semantics in the module. For example, `companies` keeps `include_inactive` outside the generic helpers, and `users` maps `status=active|inactive` to its `is_active` column.

## Soft deletes

Models embedding `BaseModel` or `BaseModelSimple` use `gorm.DeletedAt`. Normal GORM queries exclude soft-deleted rows automatically.

Rules:

- Delete endpoints SHOULD call GORM delete APIs and rely on soft delete behavior.
- Shared GORM helpers MUST NOT call `Unscoped()`.
- List/read endpoints SHOULD NOT return soft-deleted rows.
- Hard deletes and unscoped reads require an explicit, documented module exception and tests.
- API DTOs SHOULD omit `DeletedAt`.

## Tenant scope

Tenant-owned handlers read the tenant context from Gin and pass it down:

```go
tc, ok := tenant.FromGin(c)
if !ok {
	response.HandleError(c, apperror.ErrForbidden)
	return
}
```

Repositories are the enforcement boundary and MUST call `tenant.ApplyTenantScope(db, tc)` for tenant-owned reads, updates, deletes, and list counts before executing the query. Cross-tenant misses should return `404 Not Found`, not information about a forbidden existing record.

Root-global endpoints may intentionally run without a company filter only when `tc.IsRootScope` is true. Root-scoped requests and non-root requests use the same company filter path.

## Permissions and route registration

Register module routes through `routes.go` and inject guards instead of importing middleware directly into every handler. Use permission slugs for capability checks and role guards only for role-specific platform operations.

```go
products.GET("", requirePermission("products.index"), handler.List)
products.GET("/:id", requirePermission("products.view"), handler.GetByPublicID)
products.POST("", requirePermission("products.create"), handler.Create)
products.PUT("/:id", requirePermission("products.update"), handler.Update)
products.DELETE("/:id", requirePermission("products.delete"), handler.Delete)
```

The app container chooses the route group before calling `Register`. Tenant-owned modules belong on a tenant-protected group; platform-global modules can use the global protected group.

## Review checklist

- [ ] Module includes the expected files under `internal/modules/{name}/`.
- [ ] DTOs use explicit request/response names and do not expose `DeletedAt`.
- [ ] DTO validation returns `response.ValidationErrors` and handlers call `RespondIfInvalid`.
- [ ] Validation responses are field-keyed and returned as 422.
- [ ] List endpoints use `query.ListFromGin`, GORM helpers, and `response.PaginatedWithFilters`.
- [ ] Repositories allowlist sortable columns and choose searchable columns explicitly.
- [ ] Normal reads preserve GORM soft delete scope; any unscoped/hard delete behavior is documented.
- [ ] Tenant-owned repositories call `tenant.ApplyTenantScope` before query execution.
- [ ] Routes are registered from `routes.go` with `RequirePermission`/role guards.
