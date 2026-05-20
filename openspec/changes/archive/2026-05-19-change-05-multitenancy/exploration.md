## Exploration: Multitenancy by company_id

### Current State

NexoKit already has significant multitenancy **scaffolding** in place, but no working tenant isolation. Specifically:

- **User model** already has `CompanyID *uint` (nullable, matching the prompt's rule: root = nil, admin/user = required).
- **PASETO access token** already carries `company_id` as a claim (`AccessClaims.CompanyID *uint`).
- **authctx.User** already has `CompanyID *uint`, populated during auth from the DB user row.
- **Auth middleware** (`middleware/auth.go`) resolves user identity but does NOT resolve or inject tenant context.
- **Authorization middleware** (`middleware/authorization.go`) checks permissions via `RequirePermission` and role via `RequireRole`. Root gets wildcard `*`.
- **Companies module** exists at `internal/modules/companies/` but is a stub: `module.go` only has an empty `Register(v1 interface{})` function.
- **Tenant middleware** file exists at `internal/middleware/tenant.go` but contains only `// TODO: implement in change-05-multitenancy`.
- **Module generator** CLI template (`internal/cli/templates/module/repository.tmpl`) already has `{{- if .Tenant }}` branches that filter by `ctx.Value("company_id")`. However, this uses raw `context.Context` values rather than a structured `TenantContext`.
- **Seeds** already include `companies.*` permissions (`index`, `view`, `create`, `update`, `delete`).
- **No `companies` table** exists in migrations. The DB has `roles`, `users`, `refresh_tokens`, `permissions`, and `role_permissions`.
- **No GORM tenant scope helpers** exist (`WithCompany`, `ApplyTenantScope` are not implemented).
- **No TenantContext struct** exists.
- **Container wiring** (`app/container.go`) does NOT wire companies — no repository, service, or handler.
- **Route registration** does NOT mount `/api/v1/companies`.

The system currently has **zero tenant isolation enforcement**: any authenticated user with list permissions can see ALL rows across all companies.

### Affected Areas

- `internal/modules/companies/` — **EXPAND**: Full CRUD module (model, dto, repository, service, handler, routes, validation)
- `internal/middleware/tenant.go` — **IMPLEMENT**: Tenant resolution and context injection (private routes + public routes)
- `internal/platform/authctx/authctx.go` — **MODIFY**: Add tenant-related helpers (`TenantContext` extraction, `IsRootScope` checks)
- `internal/shared/model.go` — **POSSIBLY MODIFY**: Consider adding a `TenantModel` embedding or just document that tenant-scoped models add `CompanyID uint`
- `internal/modules/users/repository.go` — **MODIFY**: All queries need `company_id` filtering (except root-scope)
- `internal/modules/users/service.go` — **MODIFY**: Accept and use TenantContext for filtering
- `internal/modules/users/handler.go` — **MODIFY**: Extract tenant context and pass to service; enforce company_id rules on create/update
- `internal/modules/users/dto.go` — **MODIFY**: Enforce company_id required for admin/user creation; add company_id validation
- `internal/modules/users/model.go` — **REVIEW**: Already has `CompanyID *uint`; may need `gorm:"index"` tag
- `internal/app/container.go` — **MODIFY**: Wire companies module; inject tenant middleware into route chain
- `internal/server/router.go` — **REVIEW**: May need public vs private route groups for tenant resolution
- `internal/modules/roles/repository.go` — **MODIFY**: May need tenant-scoped queries (roles are global, but listing could be filtered)
- `internal/modules/permissions/repository.go` — **REVIEW**: Permissions are global; no tenant filtering needed
- `internal/modules/auth/handler.go` — **REVIEW**: `Me()` already returns `CompanyID`; no change needed
- `internal/modules/auth/service.go` — **REVIEW**: Already passes `user.CompanyID` to token issuer
- `internal/platform/token/token.go` — **REVIEW**: Already carries `CompanyID`; may need to include `company_slug` or `is_root_scope` for tenant context
- `migrations/YYYYMMDDHHMMSS_companies.sql` — **NEW**: Create `companies` table
- `seeds/companies.go` — **NEW**: Seed initial company or leave empty for root to create
- `internal/cli/templates/module/repository.tmpl` — **UPDATE**: Replace `ctx.Value("company_id")` with proper `TenantContext`-based filtering
- `internal/modules/auth/routes.go` — **REVIEW**: Login/refresh are public (no tenant); Me is auth-only
- `tests/` — **NEW**: Integration tests for tenant isolation, cross-tenant access denial, root global scope

### Approaches

#### 1. TenantContext struct with GORM scope helpers (recommended by spec)

**Description**: Create a `TenantContext` struct holding `CompanyID`, `CompanySlug`, and `IsRootScope`. Inject it via middleware into `gin.Context`. Provide `WithCompany(db, id)` and `ApplyTenantScope(db, ctx)` as GORM `Scopes` that conditionally add `WHERE company_id = ?`. Repositories accept `TenantContext` or a `*gorm.DB` pre-scoped.

- **Pros**: Matches the exact spec from the change prompt. Explicit, testable, and type-safe. Clean separation of tenant resolution (middleware) from data filtering (GORM scopes). Root-scope bypass is straightforward (`IsRootScope` = no filter).
- **Cons**: Requires changing all existing repository signatures to accept or use `TenantContext`. More invasive change across every tenant-scoped module.
- **Effort**: Medium-high

#### 2. Context value with GORM scopes (extend existing pattern)

**Description**: Store resolved `company_id` in `context.Context` via a typed key. Build GORM scope helpers that read from `context.Context`. Keep the pattern from the CLI template but replace raw string keys with a proper context key type.

- **Pros**: Minimal API surface change — repository methods already take `context.Context` (in generated templates) or can add it retroactively. Works with the existing template pattern.
- **Cons**: Untyped context values are less discoverable and more error-prone than a named struct. Mixing `gin.Context` with `context.Context` for tenant resolution can be confusing. Harder to carry `IsRootScope` alongside `company_id`.
- **Effort**: Medium

#### 3. Session-scoped *gorm.DB injection per request

**Description**: Instead of filtering per-query, resolve the tenant in middleware and create a scoped `*gorm.DB` (`db.Where("company_id = ?", companyID)`) that is injected into the Gin context. Repositories read this pre-scoped DB object.

- **Pros**: Zero per-query filtering logic inside repositories. The DB object itself enforces isolation. Simple repositories.
- **Cons**: Root-scope requires a different (unscoped) DB object. Two code paths in every handler that might need root access. GORM session mode carries state that can leak if not carefully managed. Harder to test — you need two DB instances per test.
- **Effort**: Medium

### Recommendation

**Approach 1 (TenantContext struct + GORM scope helpers)**.

Rationale:

1. **Spec alignment** — The change prompt explicitly requests `TenantContext` with `company_id`, `company_slug`, `is_root_scope` and GORM helpers `WithCompany(db, id)` and `ApplyTenantScope(db, ctx)`. Approach 1 delivers exactly this.

2. **Type safety** — A `TenantContext` struct is discoverable, testable, and impossible to misuse silently. Context values (`approach 2`) hide tenant data behind `context.Value(string)` which is Go's weakest pattern.

3. **Root access** — `IsRootScope` is a first-class citizen. When true, `ApplyTenantScope` simply returns the unmodified `*gorm.DB`. This is the cleanest way to handle the root-can-see-everything requirement.

4. **Testability** — `TenantContext` is a plain struct. Tests construct it directly without magic. GORM scopes are pure functions `func(db *gorm.DB) *gorm.DB` — trivial to unit test.

5. **Consistency with module generator** — The CLI template can be updated to use `ApplyTenantScope` instead of `ctx.Value("company_id")`, making the generated code consistent with hand-written code.

### Implementation outline

**New files**:
- `internal/platform/tenant/tenant.go` — `TenantContext` struct, `WithCompany()`, `ApplyTenantScope()`, helpers to get/set from `gin.Context`
- `internal/modules/companies/model.go` — `Company` model
- `internal/modules/companies/dto.go` — Create/Update/List DTOs + validation
- `internal/modules/companies/repository.go` — Company CRUD
- `internal/modules/companies/service.go` — Company business logic
- `internal/modules/companies/handler.go` — Company HTTP handlers
- `internal/modules/companies/routes.go` — Route registration
- `internal/modules/companies/validation.go` — Validation rules
- `internal/middleware/tenant.go` — Implement `TenantMiddleware` for private and public routes
- `migrations/YYYYMMDDHHMMSS_companies.sql` — Create `companies` table
- `seeds/companies.go` — (Optional) seed or leave creation to root

**Modified files**:
- `internal/app/container.go` — Wire companies module and tenant middleware
- `internal/server/router.go` — Private/public route groups with tenant middleware
- `internal/modules/users/repository.go` — Tenant-scoped queries using `ApplyTenantScope`
- `internal/modules/users/service.go` — Pass `TenantContext` through service layer
- `internal/modules/users/handler.go` — Extract `TenantContext` from Gin context
- `internal/modules/users/dto.go` — Company_id required for non-root users
- `internal/cli/templates/module/repository.tmpl` — Use `ApplyTenantScope`
- `internal/cli/templates/module/model.tmpl` — Use `CompanyID` consistently
- `seeds/permissions.go` — Already has `companies.*` permissions; no change needed

### Risks

- **Cross-tenant data leak**: The biggest risk. Every repository query on a tenant-scoped model MUST go through `ApplyTenantScope` or `WithCompany`. A single missed query exposes cross-tenant data. Mitigation: integration tests that seed two companies and verify admin A cannot read company B's data.
- **Root-scope bypass misuse**: `IsRootScope` must be set ONLY by the tenant middleware when the authenticated user is root AND no specific company override is provided. If any handler accidentally sets `IsRootScope = true`, data isolation breaks. Mitigation: `TenantContext` is read-only after middleware; only middleware creates it.
- **Public route tenant resolution**: Public routes (e.g., product catalogs for storefronts) resolve tenant from `Host` header, `domain`, or `subdomain`. This requires a company lookup by domain/subdomain in middleware, which adds a DB hit per request. Mitigation: cache company-by-domain lookups with short TTL.
- **Migration ordering**: The `companies` table must be created BEFORE `users` can reference it. Currently `users.company_id` is nullable, which is fine. But any new foreign key constraint requires careful migration.
- **Existing module refactoring**: Users, roles, and permissions repositories currently have no tenant awareness. Adding it requires changing signatures and all call sites. Risk: missing a call site. Mitigation: compiler-enforced — change the repository interface to require `TenantContext` parameter.
- **Module generator template**: The current template uses `ctx.Value("company_id")` with raw context keys. This must be updated to use the new `tenant` package, or generated modules will be inconsistent. Moderate effort but important for maintainability.

### Ready for Proposal

Yes. The exploration is complete. All critical patterns (auth middleware flow, authctx, PASETO token claims, RBAC middleware, existing CompanyID on User, module flat structure, container wiring, migration conventions, seeding conventions, CLI templates, GORM patterns, test patterns) are understood. The next phase should create the proposal for this change.