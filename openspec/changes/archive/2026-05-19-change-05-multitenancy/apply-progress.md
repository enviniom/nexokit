# Apply Progress: change-05-multitenancy

## Mode

- Strict TDD: active
- Test runner: `go test ./...`
- Delivery: chained PR slice
- Chain strategy: stacked-to-main
- Current work unit: PR 4 Generator, permissions, docs

## Completed Tasks

- [x] 1.1 RED: add table-driven tests for `internal/platform/tenant/tenant.go` covering scoped, root, and zero-company `ApplyTenantScope`.
- [x] 1.2 Create `internal/platform/tenant/{tenant,doc}.go` with immutable `TenantContext`, Gin helpers, `WithCompany`, and `ApplyTenantScope`.
- [x] 1.3 RED: add `internal/middleware/tenant_test.go` cases for private root/global, root `X-Company-ID`, admin scope, public host/subdomain, dev-only `X-Tenant`, and 404/400 failures.
- [x] 1.4 Replace `internal/middleware/tenant.go` with private/public middleware and resolver interfaces; keep non-root missing company as 403.
- [x] 2.1 Create `migrations/20260519000000_companies.sql` for companies plus nullable `users.company_id` FK/index; verify up/down behavior.
- [x] 2.2 RED: add company service/handler tests for root create, admin/user 403, duplicate slug 422, inactive filtering, and PublicID-only `:id` routes.
- [x] 2.3 Replace `internal/modules/companies/module.go` with flat files `{model,dto,repository,service,handler,routes,validation}.go` using response helpers and no exposed uint IDs.
- [x] 2.4 Wire companies repository/service/handler in `internal/app/container.go` and mount `/api/v1/companies` in `internal/server/router.go` with root-only create/update/delete guards.
- [x] 3.1 RED: extend `internal/modules/users/*_test.go` for required `company_id`, root nullable company, scoped list/get/update/delete, password/status cross-tenant 404.
- [x] 3.2 Modify `internal/modules/users/{repository,service,handler,dto,model}.go` so tenant-scoped methods require `tenant.Context`; auth bootstrap lookups remain unscoped.
- [x] 3.3 Update `internal/middleware/auth.go` and `internal/platform/authctx/authctx.go` so authenticated requests also receive TenantContext, including root header override.
- [x] 3.4 Add `tests/integration` httptest coverage seeding two companies/users: admin isolation, cross-tenant 404, root global, root scoped.
- [x] 4.1 Update `internal/cli/templates/module/{repository,model}.tmpl` to use `tenant.Context` and `ApplyTenantScope`; refresh golden files and rerun without update mode.
- [x] 4.2 Align `seeds/role_permissions.go` with root-only company mutations or ensure routes enforce role plus permission.
- [x] 4.3 Create `docs/multitenancy.md` describing tenant model fields, repository scope rules, and review checklist for new modules.
- [x] 4.4 Run `go test ./...` and confirm all spec scenarios above are covered before verification.

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `internal/platform/tenant/tenant_test.go` | Unit | N/A (new package) | ✅ `go test ./internal/platform/tenant -run Test` failed on undefined tenant contract | ✅ `go test ./internal/platform/tenant -run Test` passed | ✅ scoped, root, zero-company cases | ✅ `gofmt`, focused tests still passed |
| 1.2 | `internal/platform/tenant/tenant_test.go` | Unit | N/A (new files) | ✅ Covered by 1.1 RED before production files existed | ✅ `go test ./internal/platform/tenant -run Test` passed | ✅ Gin scoped/root context cases plus GORM scope cases | ✅ `gofmt`, focused tests still passed |
| 1.3 | `internal/middleware/tenant_test.go` | Unit/httptest | ✅ `go test ./internal/middleware` passed before replacing stub | ✅ `go test ./internal/middleware -run TestPrivateTenant` failed on undefined middleware contract | ✅ `go test ./internal/middleware -run 'TestPrivateTenant|TestPublicTenant'` passed | ✅ private root/global, root scoped, admin scoped, missing company 403, public host/subdomain/dev header/not found cases | ✅ `gofmt`, focused tests still passed |
| 1.4 | `internal/middleware/tenant_test.go` | Unit/httptest | ✅ `go test ./internal/middleware` passed before replacing stub | ✅ Covered by 1.3 RED before implementation | ✅ focused middleware tests passed | ✅ public/private success and failure paths exercised | ✅ `gofmt`, focused tests still passed |
| 2.1 | `internal/modules/companies/migration_test.go` | Unit/static migration | N/A (new migration) | ✅ `go test ./internal/modules/companies -run 'TestService\|TestHandler\|TestCompaniesMigration'` failed on missing migration and undefined company contract | ✅ focused companies tests passed after migration implementation | ✅ asserts companies table, PublicID/slug/status, users FK/index, rollback drop | ✅ `gofmt`, focused tests still passed |
| 2.2 | `internal/modules/companies/service_test.go`, `internal/modules/companies/handler_test.go` | Unit + httptest | ✅ `go test ./internal/modules/companies ./internal/app ./internal/server` passed before replacing companies stub | ✅ RED failed on undefined Company/Service/Handler DTO contract | ✅ focused companies tests passed | ✅ root create, admin/user 403, duplicate slug 422, inactive filtering, PublicID-only routes, not found | ✅ `gofmt`, focused and full tests passed |
| 2.3 | `internal/modules/companies/{model,dto,repository,service,handler,routes,validation}.go` | Unit + httptest | ✅ companies package baseline had no tests; app/server tests passed | ✅ Covered by 2.2 RED before flat files existed | ✅ focused companies tests passed | ✅ service and handler scenarios force real CRUD contracts | ✅ `gofmt`, full suite passed |
| 2.4 | `internal/app/container.go`, `internal/modules/companies/routes.go` | Routing/httptest | ✅ `go test ./internal/app ./internal/server` passed before wiring | ✅ Handler route tests failed before `Register` accepted guards and mounted routes | ✅ focused changed-package tests passed | ✅ root-only mutation guards with admin/user 403, route `:id` is PublicID | ✅ `gofmt`, full suite passed |
| 3.1 | `internal/modules/users/{dto,service,handler}_test.go` | Unit + httptest | ✅ `go test ./internal/modules/users ./internal/middleware ./internal/platform/authctx` passed before changes | ✅ focused RED failed on missing tenant signatures and `authctx.CompanySlug` | ✅ users/middleware focused tests passed | ✅ required company_id, root nullable company, scoped reads/writes, cross-tenant update/delete/password/status 404 | ✅ `gofmt`, focused tests still passed |
| 3.2 | `internal/modules/users/{repository,service,handler,dto}.go` | Unit + repository contract | ✅ users package baseline passed | ✅ Covered by 3.1 RED before scoped contracts existed | ✅ changed-package tests passed | ✅ `tenant.TenantContext` required on scoped methods; `GetAuthUser` remains unscoped for auth bootstrap | ✅ full suite passed after auth test fakes were updated |
| 3.3 | `internal/middleware/auth_test.go` | Middleware/httptest | ✅ middleware baseline passed | ✅ `TestAuthSetsTenantContext` failed before `authctx.CompanySlug` and auth tenant injection existed | ✅ middleware focused tests passed | ✅ admin scoped tenant, root global tenant, root numeric `X-Company-ID` scoped tenant | ✅ `gofmt`, full suite passed |
| 3.4 | `tests/integration/users_isolation_test.go` | Integration/httptest | N/A (new integration file) | ✅ integration test referenced seeded two-company user isolation before scoped production contracts were complete | ✅ `go test ./tests/integration` passed | ✅ admin isolation, cross-tenant update 404, root global, root scoped list | ✅ `gofmt`, full suite passed |
| 4.1 | `tests/cli/golden_test.go`, `tests/cli/testdata/golden/goldenmod/{repository,service,handler}.go` | CLI golden | ✅ `go test ./tests/cli` passed before template edits | ✅ added tenant repository assertions; focused test failed on missing tenant import/context/scope and `ctx.Value` usage | ✅ `UPDATE_GOLDEN=1 go test ./tests/cli -run TestGolden_ModuleWithAllFlags` then rerun without update passed | ✅ repository assertions plus golden diff cover import, signature, ApplyTenantScope, no `ctx.Value` | ✅ focused tests passed after golden refresh |
| 4.2 | `seeds/permissions_test.go` | Unit | ✅ `go test ./seeds` passed before seed edit | ✅ root-only company mutation test failed while admin had create/update/delete | ✅ focused seed tests passed after removing admin mutation permissions | ✅ create, update, delete mutation slugs all excluded while root retains all permissions | ✅ `gofmt`, focused tests passed |
| 4.3 | `tests/docs/multitenancy_test.go` | Docs contract | N/A (new docs test/file) | ✅ docs test failed while `docs/multitenancy.md` was missing | ✅ docs test passed after guide creation | ✅ required model fields, repository rules, cross-tenant 404, and review checklist headings/items | ✅ focused docs test passed |
| 4.4 | `go test ./...` | Full suite | ✅ PR4 focused suites passed | ✅ Covered by 4.1-4.3 RED cycles | ✅ `go test ./...` passed | ✅ full spec coverage includes prior PR1/2/3 scenarios plus PR4 generator/seeds/docs checks | ✅ no additional refactor needed |

## Test Summary

- Total tests written/extended: 18+ top-level tests/checks across tenant, middleware, companies, users service/handler/DTO, auth middleware, integration isolation, CLI golden templates, seed permissions, and docs contract checks.
- Total tests passing: focused PR4 tests and full suite (`go test ./...`).
- Layers used: Unit, static migration check, httptest middleware/handler tests, sqlite-backed integration httptest, CLI golden, and docs contract tests.
- Approval tests: none — behavior changes were specified by the multitenancy delta.
- Pure functions created: tenant scope helpers, middleware host/subdomain resolution helpers, company slug/status/list normalization helpers, auth tenant derivation helper.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/platform/tenant/doc.go` | Created | Added package guidance for multitenant model and repository scope usage. |
| `internal/platform/tenant/tenant.go` | Created | Added `TenantContext`, `CompanyRef`, Gin helpers, `WithCompany`, and `ApplyTenantScope`. |
| `internal/platform/tenant/tenant_test.go` | Created | Added table-driven GORM dry-run and Gin context tests. |
| `internal/middleware/tenant.go` | Replaced | Implemented private/public tenant middleware, resolver interface, root override, host/subdomain/dev header resolution, and short TTL cache. |
| `internal/middleware/tenant_test.go` | Created | Added private/public middleware behavior tests using `httptest`. |
| `migrations/20260519000000_companies.sql` | Created | Added companies table, company lookup indexes, nullable users company FK/index, and down migration. |
| `internal/modules/companies/model.go` | Created | Added Company model with PublicID, slug, domain/subdomain, status, audit fields. |
| `internal/modules/companies/dto.go` | Created | Added public DTOs and list/create/update request contracts without internal uint IDs. |
| `internal/modules/companies/validation.go` | Created | Added request validation for required names/slugs and active/inactive status. |
| `internal/modules/companies/repository.go` | Created | Added GORM repository plus tenant middleware resolver methods. |
| `internal/modules/companies/service.go` | Created | Added CRUD service, generated PublicID, duplicate slug rejection including soft-deleted rows, inactive list filtering. |
| `internal/modules/companies/handler.go` | Created | Added response-helper based HTTP handlers with duplicate slug 422 mapping. |
| `internal/modules/companies/routes.go` | Created | Added `/companies` routes with root-only create/update/delete guards. |
| `internal/modules/companies/{service,handler,migration}_test.go` | Created | Added strict-TDD service, handler, route guard, PublicID route, inactive filter, and migration coverage. |
| `internal/app/container.go` | Modified | Wired companies repository/service/handler and inserted private tenant middleware into protected routes. |
| `internal/platform/authctx/authctx.go` | Modified | Added `CompanySlug` to authenticated user context for TenantContext enrichment. |
| `internal/middleware/auth.go` | Modified | Sets TenantContext during auth for non-root users, root global, and root numeric `X-Company-ID` scope. |
| `internal/middleware/auth_test.go` | Modified | Added auth TenantContext coverage. |
| `internal/modules/users/dto.go` | Modified | Requires `company_id` for non-root role IDs while allowing root nullable company. |
| `internal/modules/users/repository.go` | Modified | Added tenant-scoped list/count/get/delete contracts plus unscoped `GetAuthUser`. |
| `internal/modules/users/service.go` | Modified | Passes tenant context through reads/writes/password/status and preserves unscoped email uniqueness checks. |
| `internal/modules/users/handler.go` | Modified | Reads TenantContext from Gin and passes it to tenant-scoped service methods. |
| `internal/modules/users/{dto,service,handler,routes}_test.go` | Modified | Added/updated strict-TDD coverage for company validation, scoped service behavior, and handler tenant propagation. |
| `internal/modules/auth/service_test.go` | Modified | Updated auth service fake repository to satisfy the tenant-scoped users repository contract. |
| `tests/integration/users_isolation_test.go` | Created | Added sqlite-backed httptest coverage for two-company users isolation and root global/scoped modes. |
| `internal/cli/templates/module/repository.tmpl` | Modified | Replaced tenant `ctx.Value("company_id")` filters with explicit `tenant.TenantContext` and `tenant.ApplyTenantScope`. |
| `internal/cli/templates/module/service.tmpl` | Modified | Passes tenant context through generated tenant service read/update/delete/list methods. |
| `internal/cli/templates/module/handler.tmpl` | Modified | Reads `tenant.FromGin(c)` for generated tenant read/update/delete/list handlers before calling services. |
| `tests/cli/golden_test.go` | Modified | Added assertions proving generated tenant repositories use `tenant.ApplyTenantScope` and never `ctx.Value("company_id")`. |
| `tests/cli/testdata/golden/goldenmod/{repository,service,handler}.go` | Modified | Refreshed generated tenant golden files after template changes. |
| `seeds/permissions_test.go` | Modified | Added regression coverage excluding root-only company mutations from admin permissions. |
| `seeds/role_permissions.go` | Modified | Removed `companies.create`, `companies.update`, and `companies.delete` from admin role seed permissions. |
| `docs/multitenancy.md` | Created | Added tenant model fields, repository scope rules, and review checklist for new modules. |
| `tests/docs/multitenancy_test.go` | Created | Added docs contract coverage for required multitenancy guide sections. |
| `openspec/changes/change-05-multitenancy/tasks.md` | Updated | Marked tasks 4.1-4.4 complete. |
| `openspec/changes/change-05-multitenancy/apply-progress.md` | Updated | Merged PR 1/2/3 progress with PR 4 Generator/docs/seeds progress and TDD evidence. |

## Deviations from Design

- Minor implementation detail: `internal/server/router.go` already delegates module mounting through `Container.RegisterModules`; `/api/v1/companies` was mounted through the container rather than editing router.go directly.
- Minor implementation detail: `internal/middleware/auth.go` can only parse numeric root `X-Company-ID` values because auth middleware has no company resolver dependency; the existing `PrivateTenant` middleware still resolves public_id/slug headers immediately after auth in the protected chain.

## Issues Found

- `TenantContext` immutability is conventional rather than compiler-enforced because the approved spec/design require exported struct fields.
- `authctx.User.CompanySlug` is supported, but `internal/app/userLookup` currently cannot populate it without introducing a cross-module companies dependency; protected routes still get slug enrichment from `PrivateTenant` when root scopes via resolver, while non-root auth-derived tenant scope may have an empty slug.
- Existing uncommitted PR 1/2/3 files are still present in the working tree; this slice was stacked on top without creating commits per user constraint.
- PR4 incident audit found no obvious pre-existing PR4 partial files (`docs/multitenancy.md`, template/golden changes, seed alignment) before this run; work continued without discarding any prior changes.

## Remaining Tasks

- None for `change-05-multitenancy` apply; ready for verification.

## Verification

- `go test ./internal/middleware` — passed baseline before replacing tenant stub.
- `go test ./internal/platform/tenant -run Test` — failed RED before tenant package implementation.
- `go test ./internal/platform/tenant -run Test` — passed after tenant package implementation.
- `go test ./internal/middleware -run TestPrivateTenant` — failed RED before middleware implementation.
- `go test ./internal/middleware -run 'TestPrivateTenant|TestPublicTenant'` — passed after middleware implementation.
- `go test ./internal/platform/tenant ./internal/middleware -run 'TestApplyTenantScope|TestGinTenantContext|TestPrivateTenant|TestPublicTenant'` — passed.
- `go test ./...` — passed.
- `go test ./internal/modules/companies ./internal/app ./internal/server` — passed baseline before replacing/wiring companies stub.
- `go test ./internal/modules/companies -run 'TestService|TestHandler|TestCompaniesMigration'` — failed RED on undefined company contract and missing migration.
- `go test ./internal/modules/companies -run 'TestService|TestHandler|TestCompaniesMigration'` — passed after companies CRUD implementation.
- `go test ./internal/modules/companies ./internal/app ./internal/server` — passed after wiring.
- `go test ./...` — passed after PR 2 slice.
- `go test ./internal/modules/users ./internal/middleware ./internal/platform/authctx` — passed baseline before users isolation changes.
- `go test ./internal/modules/users ./internal/middleware -run 'Test(CreateUserRequest|UpdateUserRequest|Service_Tenant|Service_Create|Handler_PassesTenant|AuthSetsTenant)'` — failed RED on missing tenant-scoped signatures / `authctx.CompanySlug`.
- `go test ./internal/modules/users ./internal/middleware -run 'Test(CreateUserRequest|UpdateUserRequest|Service_Tenant|Service_Create|Handler_PassesTenant|AuthSetsTenant)'` — passed after users/auth implementation.
- `go test ./internal/modules/users ./internal/middleware ./tests/integration -run 'Test(CreateUserRequest|UpdateUserRequest|Service_Tenant|Service_Create|Service_Update|Service_Delete|Service_ChangePassword|Service_ToggleStatus|Handler_PassesTenant|AuthSetsTenant|UsersIsolation)'` — passed.
- `go test ./internal/modules/users ./internal/middleware ./internal/app ./internal/server ./tests/integration` — passed.
- `go test ./...` — passed after PR 3 slice.
- `go test ./tests/cli` — passed baseline before PR4 template changes.
- `go test ./tests/cli -run TestGolden_ModuleWithAllFlags` — failed RED after adding tenant repository assertions, before template implementation.
- `UPDATE_GOLDEN=1 go test ./tests/cli -run TestGolden_ModuleWithAllFlags && go test ./tests/cli -run TestGolden_ModuleWithAllFlags` — passed after template and golden updates.
- `go test ./seeds` — passed baseline before seed alignment.
- `go test ./seeds -run TestAdminPermissionSlugsExcludeRootOnlyCompanyMutations` — failed RED before removing admin company mutation permissions.
- `go test ./seeds -run 'Test(AdminPermissionSlugsExcludeRootOnlyCompanyMutations|SeedRolePermissions)'` — passed after seed alignment.
- `go test ./tests/docs -run TestMultitenancyGuideCoversTenantModelRepositoryScopeAndReviewChecklist` — failed RED before `docs/multitenancy.md` existed.
- `go test ./tests/docs -run TestMultitenancyGuideCoversTenantModelRepositoryScopeAndReviewChecklist` — passed after docs creation.
- `go test ./tests/cli ./seeds ./tests/docs` — passed focused PR4 verification.
- `go test ./...` — passed after PR4 slice.

## Workload / PR Boundary

- Mode: chained PR slice.
- Current work unit: PR 4 Generator, permissions, docs.
- Boundary: starts from completed PR 1/2/3 tenant primitives, companies CRUD, and users isolation; ends with generated tenant module safety, root-only company mutation seed alignment, multitenancy guide, and full-suite verification.
- Review budget impact: PR4 adds a focused template/docs/seeds slice; cumulative working tree still contains prior uncommitted PR1/2/3 changes by user constraint.

## Status

16/16 tasks complete. Ready for sdd-verify. No commits created.
