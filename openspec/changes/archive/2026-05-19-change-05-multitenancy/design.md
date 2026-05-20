# Design: Multitenancy by company_id

## Technical Approach

Implement tenant isolation as an explicit request context plus repository contract: middleware resolves a `tenant.Context`, handlers pass it to tenant-scoped services, and repositories apply `tenant.ApplyTenantScope(db, tc)` to every tenant-owned query. Outcome: non-root users are confined to their company, root can remain global or select one company, and public routes can resolve a company before auth-only business code exists.

## Architecture Decisions

| Decision | Choice | Alternatives / Tradeoff | Rationale |
|---|---|---|---|
| Tenant carrier | `internal/platform/tenant.Context{CompanyID *uint, CompanySlug string, IsRootScope bool}` with Gin helpers | Raw `context.Value` is smaller but hidden; pre-scoped `*gorm.DB` risks scope leakage | Type-safe, reviewable, and matches prompt. |
| External company identity | API/header route values use company `public_id` or `slug`; middleware maps to internal `uint` | Header named `X-Company-ID` could expose DB IDs | Preserves NexoKit rule: internal uint IDs are never exposed. |
| Enforcement point | Repositories require `tenant.Context` on tenant-scoped methods | Handler-only filtering is easy to bypass | Compiler-visible contracts reduce missed scopes. |
| Company CRUD auth | `POST /companies` root-only; update/delete root-only unless later spec says otherwise | Current admin seed has `companies.create/delete` | Security wins; seeds must align or route guard must use role/permission together. |

## Data Flow

Private routes:
```txt
Bearer token -> Auth -> authctx.User -> PrivateTenant -> AttachPermissions -> Handler
Handler -> Service(tenant.Context) -> Repository -> ApplyTenantScope -> GORM
```

Public routes:
```txt
Host/X-Tenant(dev) -> PublicTenant -> CompanyResolver -> tenant.Context -> public handler
```

Resolution rules: private non-root MUST use authenticated `CompanyID`; missing company is `403`. Private root uses global scope by default, or selected company from `X-Company-ID` public_id/slug. Public routes resolve by exact domain, then subdomain, then development-only `X-Tenant`.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/platform/tenant/tenant.go` | Create | Context type, Gin helpers, `WithCompany`, `ApplyTenantScope`. |
| `internal/middleware/tenant.go` | Replace | Private/Public tenant middleware and resolver interfaces. |
| `internal/modules/companies/{model,dto,repository,service,handler,routes,validation}.go` | Create | Flat CRUD module using response helpers and public IDs. |
| `migrations/20260519000000_companies.sql` | Create | `companies` table and `users.company_id` FK/index. |
| `internal/modules/users/{repository,service,handler,dto,model}.go` | Modify | Add tenant parameters, `company_id` index/validation, scoped CRUD. |
| `internal/app/container.go`, `internal/server/router.go` | Modify | Wire companies, tenant middleware, public/private groups. |
| `internal/cli/templates/module/{repository,model}.tmpl` | Modify | Generated tenant modules use `tenant.Context`, not `ctx.Value`. |
| `seeds/role_permissions.go` | Modify | Remove company create/delete from admin or guard root-only explicitly. |
| `docs/multitenancy.md` | Create | How to add tenant-scoped models safely. |

## Interfaces / Contracts

```go
type CompanyResolver interface {
    FindByPublicIDOrSlug(value string) (*companies.Company, error)
    FindByHost(host string) (*companies.Company, error)
}
func ApplyTenantScope(db *gorm.DB, tc tenant.Context) *gorm.DB
```

Tenant-scoped repository methods add `tc tenant.Context`; auth lookup methods (`GetByEmail`, `GetAuthUser`) remain unscoped because they bootstrap auth.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Tenant scopes and middleware resolution | Table-driven tests with GORM dry-run / `httptest`. |
| Unit | Company/user service rules | Table-driven service tests: root-only create, non-root company required. |
| Integration | Cross-tenant list/get/update/delete isolation, root global/scoped mode | Seed two companies/users via test DB and hit Gin routes with `httptest`. |
| CLI golden | Tenant module templates | Update golden files and rerun without update mode. |

## Migration / Rollout

Create `companies` first, then add nullable `users.company_id` FK/index. Existing root remains `NULL`; existing non-root users require manual company assignment before production rollout. Add optional dev seed company only if tests need stable fixtures. Cache public host lookups later unless profiling proves needed now.

## Work-Unit Boundaries

1. Foundation: migration, tenant package, middleware tests.
2. Companies CRUD: module, routes, root-only guards.
3. Users isolation: scoped contracts, DTO validation, integration tests.
4. Generator/docs/seeds: templates, docs, permission alignment.

## Open Questions

- [ ] Confirm whether admin should retain `companies.update` for its own company; create/delete are designed root-only.
