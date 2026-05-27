# Fresh Review: change-14-company-domains

## Verdict
Pass with non-blocking risks. The uncommitted diff implements the requested model/table, removes `companies.domain` / `companies.subdomain` from the source-of-truth path, keeps the consolidated starter migration, provisions onboarding domain rows transactionally, and changes public tenant resolution to exact active `company_domains.domain` matches. I found no blocking correctness issue.

## Blocking Issues
None found.

## Non-Blocking Issues / Risks
- `internal/modules/onboarding/service.go:74-122`: if the optional primary `domain` equals the generated technical domain (`<slug>.<platform-domain>`), both preflight availability checks can pass before either row exists; the primary insert then succeeds and the technical insert fails on the DB unique constraint. The transaction still rolls back, but the returned error may be a raw persistence error rather than the explicit duplicate-domain validation path.
- `internal/modules/companies/repository.go:104-117`: host resolution filters active `company_domains` rows and active primary rows, but does not explicitly filter the owning `companies.status`. This matches the written spec focus on domain status, but if inactive companies should be unreachable this needs a follow-up rule/test.
- `internal/middleware/tenant.go:146-153`: positive host resolutions are cached for five minutes, so domain deactivation or redirect flag changes may remain effective until cache expiry. This is consistent with the pre-existing middleware cache pattern, but it is operationally relevant for domain lifecycle changes.
- `migrations/20260101000000_init.sql:25-38`: `status` and `kind` are free-form strings with no SQL check constraints. The Go constants define supported values, but database-level enforcement is absent.

## Parent Follow-Up After Review
- Fixed the matching primary-domain/generated-technical-domain edge case with explicit `ErrDuplicateTechnicalDomain` validation and rollback test.
- Fixed host resolution to require the owning company to be active, with repository test coverage.
- Re-ran `go test ./internal/modules/onboarding ./internal/modules/companies ./internal/middleware && go test ./... && go build ./...`; all passed.

Remaining accepted non-blocking notes:
- Positive host resolutions are cached for five minutes, so domain status/redirect changes may wait for TTL until future domain-management invalidation exists.
- `status` and `kind` remain Go-enforced strings without SQL check constraints.

## Test Confidence
High. I verified the apply-progress evidence and ran `go test ./...` locally; it passed. The reported TDD evidence includes RED/GREEN/TRIANGULATE coverage for tenant host resolution, redirect behavior, onboarding domain provisioning, rollback, migration schema, and config loading.

Notable covered behavior:
- consolidated migration includes `company_domains` and removes `companies.domain` / `companies.subdomain`;
- onboarding creates primary and optional technical domain rows;
- duplicate domain scenarios roll back onboarding;
- production subdomain fallback is removed;
- inactive domains do not resolve;
- explicit alias redirect preserves path/query.

## Files to Review
- `migrations/20260101000000_init.sql`
- `internal/modules/companies/model.go`
- `internal/modules/companies/repository.go`
- `internal/modules/onboarding/dto.go`
- `internal/modules/onboarding/service.go`
- `internal/modules/onboarding/handler.go`
- `internal/middleware/tenant.go`
- `internal/platform/tenant/tenant.go`
- `internal/config/config.go`
- `internal/app/container.go`
- `internal/modules/companies/*_test.go`
- `internal/modules/onboarding/*_test.go`
- `internal/middleware/tenant_test.go`
- `internal/config/config_test.go`
