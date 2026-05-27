# Apply Result: change-14-company-domains

## Status
Implemented and fresh-reviewed. No commit created.

## Summary
- Added `CompanyDomain` model with status, kind, and `redirect_to_primary` metadata.
- Removed `Domain`/`Subdomain` from direct company model/DTO/service mapping.
- Updated existing Goose consolidated migration `migrations/20260101000000_init.sql`; no new migration file was added.
- Updated onboarding to keep optional `domain`, remove `subdomain`, add `generate_technical_domain`, and create primary/technical domain rows transactionally.
- Added `APP_PLATFORM_DOMAIN` config and `.env.example` documentation.
- Updated public tenant resolution to exact active `company_domains.domain` matches only.
- Removed production subdomain-to-slug fallback.
- Implemented explicit 308 redirect behavior for `redirect_to_primary`, preserving path/query and honoring `X-Forwarded-Proto`.
- Updated `tasks.md` and wrote `apply-progress.md` with TDD evidence.
- Addressed fresh-review follow-up: matching primary/technical domain now fails with explicit validation, and active domains for inactive companies no longer resolve.

## Tests
- `go test ./internal/middleware` — RED first, then PASS.
- `go test ./internal/modules/onboarding ./internal/modules/companies` — RED first, then PASS.
- `go test ./internal/modules/companies` — PASS after repository edge tests.
- `go test ./internal/modules/onboarding` — PASS after onboarding triangulation tests.
- `go test ./...` — PASS.
- `go build ./...` — PASS.
- Parent follow-up verification after review fixes: `go test ./internal/modules/onboarding ./internal/modules/companies ./internal/middleware && go test ./... && go build ./...` — PASS.

## Changed Files
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
- `openspec/changes/change-14-company-domains/apply-progress.md`
- `openspec/changes/change-14-company-domains/apply-result.md`

## Risks / Review Notes
- `CompanyDomain` embeds `shared.BaseModel`, so `deleted_at` remains structurally present even though domain lifecycle is status-based.
- Redirect-enabled alias serves normally if no active primary domain exists; this avoids redirect loops but should be reviewed as intended behavior.
- No domain-management CRUD endpoints were added; this change establishes schema, onboarding, and resolution foundations.
- Fresh external review passed with non-blocking risks; two low-risk findings were fixed before handoff.

## Workload
Implementation diff excluding local Pi hygiene files is approximately 20 files, 549 insertions, 187 deletions, under the selected 700-line review budget.
