## Archive Report

**Change**: change-14-company-domains
**Archived to**: `openspec/changes/archive/2026-05-27-change-14-company-domains/`
**Mode**: openspec
**Date**: 2026-05-27

### Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| company-domains | Created | New main spec with 7 requirements, 20 scenarios (company domains model, lifecycle, onboarding, tenant resolution, redirect, companies API surface, root domain admin) |
| companies-crud | Updated | Removed `Domain *string` and `Subdomain *string` from Company model; added `Domains` relationship; updated update scenario to reflect domain management via dedicated endpoints |
| company-onboarding | Updated | Added domain creation to transactional provisioning; added duplicate domain rollback scenario |
| tenant-isolation | Updated | Replaced subdomain-to-slug fallback with exact `company_domains.domain` matching; added redirect alias scenario; removed production subdomain resolution |

### Archive Contents

- proposal.md ✅
- specs.md ✅
- design.md ✅
- tasks.md ✅ (22/23 tasks complete; 5.4 is human review, intentionally deferred)
- verify-report.md ✅ (PASS WITH WARNINGS)
- exploration.md ✅
- apply-progress.md ✅
- apply-result.md ✅
- domain-admin-apply-result.md ✅
- domain-admin-review.md ✅
- review-apply.md ✅
- verify-result.md ✅

### Verification Summary

- **Verdict**: PASS WITH WARNINGS
- **Build**: ✅ Passed
- **Tests**: ✅ All passed, 0 failed
- **Spec Compliance**: 37/37 scenarios compliant
- **TDD Compliance**: 6/6 checks passed
- **Warnings**: Companies package coverage at 53.9% (informational — domain admin tested via service layer); dead `CreateCompanyRequest` handler noted for future cleanup
- **Critical Issues**: None

### Source of Truth Updated

The following specs now reflect the new behavior:
- `openspec/specs/company-domains/spec.md` — NEW: full company domains specification
- `openspec/specs/companies-crud/spec.md` — UPDATED: Company model no longer has Domain/Subdomain fields
- `openspec/specs/company-onboarding/spec.md` — UPDATED: Onboarding creates company_domains rows
- `openspec/specs/tenant-isolation/spec.md` — UPDATED: Tenant resolution uses company_domains exact match

### Code File Modifications

**No code files were modified during this archive phase.** Only OpenSpec specification files were created or updated.

### SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
