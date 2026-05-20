# Proposal: Multitenancy by company_id

## Intent

Enforce tenant data isolation by `company_id` so a single NexoKit instance safely serves multiple companies. Today, any authenticated user with list permissions sees ALL rows — zero tenant enforcement exists despite partial scaffolding (User.CompanyID, PASETO claim, authctx field).

## Scope

### In Scope

- `Company` model + `companies` table migration
- `TenantContext` struct (`CompanyID`, `CompanySlug`, `IsRootScope`)
- `tenant` platform package with `WithCompany()` and `ApplyTenantScope()` GORM scope helpers
- Tenant middleware for private routes (resolve from authenticated user)
- Tenant middleware for public routes (resolve from Host / domain / subdomain / X-Tenant header)
- Companies CRUD module (root-only create)
- `users` module refactored to tenant-scoped queries
- Module generator template updated from `ctx.Value` to `ApplyTenantScope`
- Integration tests proving cross-tenant isolation and root global access
- Documentation for adding tenant scope to new models

### Out of Scope

- Row-level security at the DB level (future hardening)
- Tenant-specific feature flags or configuration
- Company admin self-service (invite, billing) — deferred to a later change
- Public storefront API endpoints (only the resolution mechanism)

## Capabilities

### New Capabilities

- `tenant-isolation`: TenantContext struct, GORM scope helpers, tenant middleware for private and public routes. Core isolation enforcement.
- `companies-crud`: Company model, migration, CRUD endpoints (root-only create).

### Modified Capabilities

- `users`: Requirements change — all queries MUST filter by `company_id` for non-root users; company_id required for admin/user creation.
- `middleware-auth`: Requirement change — auth middleware MUST set TenantContext alongside User context when user is authenticated.

## Approach

TenantContext struct in `internal/platform/tenant/` with GORM scopes `WithCompany(db, id)` and `ApplyTenantScope(db, TenantContext)`. Middleware resolves tenant: **private routes** from `authctx.User.CompanyID` (root users get `IsRootScope=true`); **public routes** from Host header / domain / subdomain lookup with cache. Repositories receive `TenantContext` and use `ApplyTenantScope` for every query on tenant-scoped models. Root bypass: `IsRootScope=true` skips the filter. Companies module follows NexoKit flat module convention. Module generator template updated to use `ApplyTenantScope`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/platform/tenant/` | New | TenantContext, WithCompany, ApplyTenantScope, Gin context helpers |
| `internal/middleware/tenant.go` | New (replaces stub) | Tenant resolution for private + public routes |
| `internal/modules/companies/` | New (replaces stub) | Full CRUD module |
| `migrations/` | New | companies table |
| `internal/modules/users/` | Modified | Tenant-scoped repository, service, handler, dto |
| `internal/platform/authctx/authctx.go` | Modified | Add tenant context helpers or integration point |
| `internal/app/container.go` | Modified | Wire companies module + tenant middleware |
| `internal/server/router.go` | Modified | Private/public route groups with tenant middleware |
| `internal/cli/templates/module/repository.tmpl` | Modified | Use ApplyTenantScope pattern |
| `tests/integration/` | New | Cross-tenant isolation tests |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Cross-tenant data leak from missed ApplyTenantScope | Medium | Integration tests with two companies; change repository interface to require TenantContext |
| Root-scope bypass misuse in handlers | Low | TenantContext is read-only after middleware; only tenant middleware creates it |
| Public route DB hit per request for domain lookup | Medium | Cache company-by-domain with short TTL |
| Migration ordering (companies before FK) | Low | companies migration runs before any FK constraints |
| Existing module signature changes trigger compiler errors | Low | Compiler catches all call sites — safe refactoring |

## Rollback Plan

1. Revert the migration (drop companies table).
2. Revert middleware changes (tenant middleware returns to stub).
3. Remove `internal/platform/tenant/` package.
4. Remove `internal/modules/companies/` module files (restore `module.go` stub).
5. Revert `users` module to pre-tenant state.
6. All changes are additive except the `users` refactor — Git revert handles it.

## Dependencies

- Goose must be available for migration execution
- Existing auth middleware and authctx must be stable (they are — change-01 through change-04 are complete)

## Success Criteria

- [ ] companies table exists with correct schema
- [ ] Admin/user CANNOT read or modify data from another company
- [ ] Root CAN operate globally (no company_id filter)
- [ ] Root CAN scope to a specific company via header
- [ ] Tenant middleware works on both private and public routes
- [ ] ApplyTenantScope is a reusable GORM scope
- [ ] Cross-tenant isolation is verified by integration tests
- [ ] Module generator produces tenant-aware code