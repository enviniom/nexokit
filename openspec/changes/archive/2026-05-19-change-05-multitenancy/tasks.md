# Tasks: Multitenancy by company_id

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 900-1,400 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 Foundation → PR 2 Companies CRUD → PR 3 Users isolation → PR 4 Generator/docs/seeds |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Tenant primitives and middleware contract | PR 1 | Base for all scoped modules; unit tests included. |
| 2 | Root-owned company CRUD | PR 2 | Depends on PR 1; migration, routes, service tests included. |
| 3 | Tenant-scoped users behavior | PR 3 | Depends on PR 1/2; integration tests prove isolation. |
| 4 | Generated-module safety and docs | PR 4 | Depends on PR 1; golden tests and docs included. |

## Phase 1: Foundation / Tenant contract

- [x] 1.1 RED: add table-driven tests for `internal/platform/tenant/tenant.go` covering scoped, root, and zero-company `ApplyTenantScope`.
- [x] 1.2 Create `internal/platform/tenant/{tenant,doc}.go` with immutable `TenantContext`, Gin helpers, `WithCompany`, and `ApplyTenantScope`.
- [x] 1.3 RED: add `internal/middleware/tenant_test.go` cases for private root/global, root `X-Company-ID`, admin scope, public host/subdomain, dev-only `X-Tenant`, and 404/400 failures.
- [x] 1.4 Replace `internal/middleware/tenant.go` with private/public middleware and resolver interfaces; keep non-root missing company as 403.

## Phase 2: Companies CRUD

- [x] 2.1 Create `migrations/20260519000000_companies.sql` for companies plus nullable `users.company_id` FK/index; verify up/down behavior.
- [x] 2.2 RED: add company service/handler tests for root create, admin/user 403, duplicate slug 422, inactive filtering, and PublicID-only `:id` routes.
- [x] 2.3 Replace `internal/modules/companies/module.go` with flat files `{model,dto,repository,service,handler,routes,validation}.go` using response helpers and no exposed uint IDs.
- [x] 2.4 Wire companies repository/service/handler in `internal/app/container.go` and mount `/api/v1/companies` in `internal/server/router.go` with root-only create/update/delete guards.

## Phase 3: Users isolation

- [x] 3.1 RED: extend `internal/modules/users/*_test.go` for required `company_id`, root nullable company, scoped list/get/update/delete, password/status cross-tenant 404.
- [x] 3.2 Modify `internal/modules/users/{repository,service,handler,dto,model}.go` so tenant-scoped methods require `tenant.Context`; auth bootstrap lookups remain unscoped.
- [x] 3.3 Update `internal/middleware/auth.go` and `internal/platform/authctx/authctx.go` so authenticated requests also receive TenantContext, including root header override.
- [x] 3.4 Add `tests/integration` httptest coverage seeding two companies/users: admin isolation, cross-tenant 404, root global, root scoped.

## Phase 4: Generator, permissions, docs

- [x] 4.1 Update `internal/cli/templates/module/{repository,model}.tmpl` to use `tenant.Context` and `ApplyTenantScope`; refresh golden files and rerun without update mode.
- [x] 4.2 Align `seeds/role_permissions.go` with root-only company mutations or ensure routes enforce role plus permission.
- [x] 4.3 Create `docs/multitenancy.md` describing tenant model fields, repository scope rules, and review checklist for new modules.
- [x] 4.4 Run `go test ./...` and confirm all spec scenarios above are covered before verification.
