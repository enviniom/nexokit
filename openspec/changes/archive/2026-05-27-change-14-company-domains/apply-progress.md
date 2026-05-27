# Apply Progress: change-14-company-domains

## Status
Implemented and fresh-reviewed; pending human code review. No commit created.

## Completed Tasks
- Phase 1: schema/model foundation.
- Phase 2: companies API cleanup.
- Phase 3: onboarding domain provisioning.
- Phase 4: tenant host resolution and redirects.
- Phase 5.1-5.3: strict TDD verification, build verification, and workload check.
- Phase 6: root-only company domain administration endpoints.

Phase 5.4 fresh review was run by the parent session after apply. Two reviewer risks were addressed before handoff: matching primary/technical domain validation and inactive-company host resolution.

## Files Changed
- `.env.example`
- `internal/app/container.go`
- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/middleware/tenant.go`
- `internal/middleware/tenant_test.go`
- `internal/modules/companies/dto.go`
- `internal/modules/companies/migration_test.go`
- `internal/modules/companies/model.go`
- `internal/modules/companies/repository.go`
- `internal/modules/companies/repository_test.go`
- `internal/modules/companies/service.go`
- `internal/modules/onboarding/dto.go`
- `internal/modules/onboarding/handler.go`
- `internal/modules/onboarding/handler_test.go`
- `internal/modules/onboarding/service.go`
- `internal/modules/onboarding/service_test.go`
- `internal/platform/tenant/tenant.go`
- `migrations/20260101000000_init.sql`
- `openspec/changes/change-14-company-domains/tasks.md`
- `openspec/changes/change-14-company-domains/specs.md`
- `openspec/changes/change-14-company-domains/design.md`
- `openspec/changes/change-14-company-domains/domain-admin-apply-result.md`

## Test Commands Run
| Command | Result |
| --- | --- |
| `go test ./internal/middleware` | RED: failed before implementation because `HostResolution`/`ResolveHost` did not exist and fake resolver no longer matched the contract. |
| `go test ./internal/modules/onboarding ./internal/modules/companies` | RED: failed before implementation because `CompanyDomain`, `WithPlatformDomain`, and migration schema were missing. |
| `go test ./internal/middleware ./internal/modules/onboarding ./internal/modules/companies` | GREEN: passed after implementing model/schema/onboarding/resolver/middleware. |
| `go test ./internal/modules/companies` | GREEN/TRIANGULATE: passed after adding repository host-resolution edge tests. |
| `go test ./internal/modules/onboarding` | TRIANGULATE: passed after adding skip-technical and duplicate-technical rollback tests. |
| `go test ./...` | PASS. |
| `go build ./...` | PASS. |
| `go test ./internal/modules/onboarding ./internal/modules/companies ./internal/middleware && go test ./... && go build ./...` | PASS after fresh-review follow-up fixes. |
| `go test ./internal/modules/companies` | RED for domain admin before implementation: missing DTOs/service methods/routes, then GREEN after implementation. |
| `go test ./internal/modules/companies && go test ./... && go build ./...` | PASS after root company domain administration implementation. |
| `go test ./internal/modules/companies ./internal/modules/onboarding && go test ./... && go build ./...` | PASS after domain-admin fresh-review normalization fixes. |

## TDD Cycle Evidence
| Cycle | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- |
| Tenant host resolution + redirects | Added middleware tests requiring `ResolveHost`, exact company-domain resolution, no subdomain fallback, and 308 redirect to primary; focused test failed to compile because the contract did not exist. | Added `tenant.HostResolution`, resolver contract, repository `ResolveHost`, and middleware redirect handling. | Added repository test proving active alias resolves with primary metadata and inactive domains do not resolve. | Split public resolution into host lookup, dev `X-Tenant` fallback, redirect helper, and cache helpers. |
| Onboarding domain provisioning | Added service tests requiring `CompanyDomain`, `WithPlatformDomain`, `generate_technical_domain`, primary domain creation, and duplicate-domain rollback; focused tests failed to compile. | Added `CompanyDomain`, onboarding options, primary/technical domain creation inside transaction, and domain uniqueness checks. | Added skip-technical and duplicate-technical rollback tests. | Extracted `ensureDomainAvailable`, `createCompanyDomain`, and `normalizeDomain`. |
| Migration/model/API cleanup | Added migration tests for `company_domains` and absence of `companies.domain/subdomain`; test failed against old migration. | Updated consolidated Goose migration and removed company direct domain/subdomain DTO/model/service mapping. | Repository and full-suite tests validated DTO/model compile across modules. | Kept domain ownership in model/repository while company profile service only manages company profile fields. |
| Config | Added config assertions for `APP_PLATFORM_DOMAIN`. | Added `AppConfig.PlatformDomain`, loader support, container wiring, and `.env.example` entry. | Full suite verified default empty and custom platform-domain cases. | Passed platform domain via onboarding service option rather than reading environment inside service. |
| Root company domain admin | Added handler/service tests requiring nested root-only list/create/update domain endpoints; focused companies test failed to compile due to missing DTOs and service methods. | Added domain admin DTOs, service/repository methods, handlers, and routes. | Added service cases for duplicate domains, active primary conflicts, and cross-company update rejection; handler test verifies no DELETE route. | Kept domain admin in companies module while preserving separate company profile update behavior. |

## Deviations from Design
- `shared.BaseModel` still provides `deleted_at` on `CompanyDomain` through GORM model reuse and the consolidated migration includes `deleted_at` for consistency. Domain lifecycle behavior is status-based; delete flows were not added.
- Redirect uses `308 Permanent Redirect` and honors `X-Forwarded-Proto` before request TLS/URL scheme.
- If `redirect_to_primary = true` but no active primary domain exists, middleware serves the matched tenant without redirecting to avoid loops/failures.

## Fresh Review Follow-Up
- Added explicit onboarding validation for the edge case where the supplied primary `domain` equals the generated technical domain (`<slug>.<platform-domain>`), returning `ErrDuplicateTechnicalDomain` before inserts.
- Added repository filtering so host resolution requires both an active `company_domains` row and an active owning company.
- Added tests for matching primary/technical domain rollback and active domain on inactive company not resolving.
- Domain-admin fresh review found normalized-empty and trailing-dot canonical duplicate blockers; fixed by validating normalized domain values in companies DTOs and aligning onboarding normalization with admin normalization.
- Added tests for normalized-empty domain validation and trailing-dot duplicate rollback.

## Remaining Tasks
- Human review of migration consistency, transaction rollback safety, tenant resolution security, redirect loop behavior, stale DTO fields, and domain administration endpoint behavior.
- Future cache invalidation for host-resolution changes remains out of scope for this change.

## Workload / PR Boundary
- Original implementation diff excluding local Pi hygiene files was approximately 19 files, 464 insertions, 155 deletions.
- The domain-admin extension adds companies module DTO/service/repository/handler/route tests and OpenSpec updates, increasing review scope but staying within the already-approved current change.
- No chained PR split was applied by this subagent; parent/user can still split before commit if final diff review feels too large.
