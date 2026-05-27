# Domain Admin Fresh Review: change-14-company-domains

## Verdict
Initial fresh review found blocking normalization issues in the domain admin extension. Parent follow-up fixed the blockers and re-ran verification successfully.

## Blocking Issues Found
- Whitespace-only or normalized-empty domains could pass DTO validation because validation used the raw string before service normalization.
- Domain normalization differed between onboarding and admin endpoints: admin stripped trailing dots, onboarding did not, allowing canonical duplicates like `acme.com.` and `acme.com` across flows.

## Fixes Applied
- `CreateCompanyDomainRequest.Validate` and `UpdateCompanyDomainRequest.Validate` now validate `normalizeCompanyDomain(r.Domain)`, rejecting whitespace-only or normalized-empty values before service persistence.
- Onboarding `normalizeDomain` now strips a trailing dot, matching admin domain normalization.
- Added service test coverage for normalized-empty domain validation.
- Added onboarding test coverage for trailing-dot duplicate detection and rollback.

## Verification
- `go test ./internal/modules/companies ./internal/modules/onboarding` — PASS
- `go test ./...` — PASS
- `go build ./...` — PASS

## Remaining Non-Blocking Notes
- One-active-primary-per-company remains service-enforced rather than database-enforced and could be race-prone under concurrent writes.
- Host resolution cache invalidation remains out of scope; domain admin changes may take up to the existing cache TTL to affect cached positive resolutions.
- `status` and `kind` remain Go-enforced strings without SQL CHECK constraints.

## Files to Review
- `internal/modules/companies/dto.go`
- `internal/modules/companies/service.go`
- `internal/modules/companies/service_test.go`
- `internal/modules/onboarding/service.go`
- `internal/modules/onboarding/service_test.go`
