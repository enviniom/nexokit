# Apply Report — M4: Migrate `companies`

## Status

success

## Executive Summary

Migrated the `companies` module onto the module-owned error, persistence-boundary, and shared-helper contracts introduced by change-24. Rewrote `internal/modules/companies/core/error.go` with `code:<snake_case>` sentinels, added `core/errors_test.go`, `core/dto_test.go`, and `core/model_test.go`, and removed `gorm.io/gorm` and `platform/apperror` imports from all seven companies service files. Slice repositories now translate `gorm.ErrRecordNotFound` to module sentinels and unique-constraint violations to `core.ErrDuplicateCompanyDomain` / `core.ErrDuplicateCompanySlug`. Services use `str.NormalizeSlug` and `str.NormalizeDomain` from `internal/platform/shared/string` instead of inlined normalizers.

The public HTTP contract is preserved: duplicate domain/slug and active-primary-domain conflicts still return field-keyed 422 `ValidationErrorResponse` via handler-level mapping, while not-found / ownership failures return 404 through `response.HandleError`. The duplicate-domain and duplicate-slug sentinels are modeled as 409 `Conflict` per the M4 acceptance criteria; the handlers map them to the original 422 field-keyed response.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/companies/core/error.go` | Rewrote | Declared module-owned `Code*` constants and `apperror` sentinels for company/domain not-found, duplicate domain/slug, active primary domain, and cross-company domain ownership. |
| `internal/modules/companies/core/dto.go` | Modified | Replaced local `normalizeCompanyDomain` with `str.NormalizeDomain`; removed unused `strings` import. |
| `internal/modules/companies/core/model.go` | Modified | Added explicit `TableName()` methods for `Company` and `CompanyDomain`. |
| `internal/modules/companies/core/errors_test.go` | Created | Table-driven coverage of status, code format, public message, and uniqueness for every sentinel. |
| `internal/modules/companies/core/dto_test.go` | Created | Table-driven coverage of `Validate()` rules for all four request DTOs. |
| `internal/modules/companies/core/model_test.go` | Created | Direct `TableName()` assertions for both partial GORM models. |
| `internal/modules/companies/view_company/service.go` | Modified | Removed `gorm.io/gorm` and `platform/apperror` imports; service now propagates repository-translated errors unchanged. |
| `internal/modules/companies/view_company/repository.go` | Modified | Translates `gorm.ErrRecordNotFound` to `core.ErrCompanyNotFound`. |
| `internal/modules/companies/view_company/repository_test.go` | Modified | Added not-found regression test. |
| `internal/modules/companies/view_company/service_test.go` | Modified | Added not-found test asserting `core.ErrCompanyNotFound`. |
| `internal/modules/companies/update_company/service.go` | Modified | Removed `gorm`/`apperror` imports; uses `core.ErrCompanyNotFound`, `core.ErrDuplicateCompanySlug`, and `str.NormalizeSlug`. |
| `internal/modules/companies/update_company/handler.go` | Modified | Maps `core.ErrDuplicateCompanySlug` to field-keyed 422 (contract preserved). |
| `internal/modules/companies/update_company/repository.go` | Modified | Translates not-found to `core.ErrCompanyNotFound` and unique violation on slug to `core.ErrDuplicateCompanySlug`. |
| `internal/modules/companies/update_company/repository_test.go` | Modified | Added not-found and unique-violation regression tests. |
| `internal/modules/companies/update_company/service_test.go` | Modified | Updated fakes and assertions to use module sentinels. |
| `internal/modules/companies/delete_company/service.go` | Modified | Removed `gorm`/`apperror` imports; propagates repository errors. |
| `internal/modules/companies/delete_company/repository.go` | Modified | Translates not-found to `core.ErrCompanyNotFound`. |
| `internal/modules/companies/delete_company/repository_test.go` | Modified | Updated soft-delete assertion to expect `core.ErrCompanyNotFound`; added explicit not-found test. |
| `internal/modules/companies/delete_company/service_test.go` | Modified | Added not-found test. |
| `internal/modules/companies/list_company_domains/service.go` | Modified | Removed `gorm`/`apperror` imports; propagates repository errors. |
| `internal/modules/companies/list_company_domains/repository.go` | Modified | Translates not-found to `core.ErrCompanyNotFound`. |
| `internal/modules/companies/list_company_domains/repository_test.go` | Modified | Added not-found regression test. |
| `internal/modules/companies/list_company_domains/service_test.go` | Modified | Added not-found test. |
| `internal/modules/companies/create_company_domain/service.go` | Modified | Removed `gorm`/`apperror` imports; uses `core.ErrCompanyNotFound`, `core.ErrCompanyDomainNotFound`, `core.ErrDuplicateCompanyDomain`, and `str.NormalizeDomain`. |
| `internal/modules/companies/create_company_domain/repository.go` | Modified | Translates company not-found and domain not-found to module sentinels; unique violation on create returns `core.ErrDuplicateCompanyDomain`. |
| `internal/modules/companies/create_company_domain/repository_test.go` | Modified | Added not-found and unique-violation regression tests. |
| `internal/modules/companies/create_company_domain/service_test.go` | Modified | Updated fakes and assertions to use module sentinels; added company-not-found test. |
| `internal/modules/companies/update_company_domain/service.go` | Modified | Removed `gorm`/`apperror` imports; uses module sentinels and `str.NormalizeDomain`. |
| `internal/modules/companies/update_company_domain/repository.go` | Modified | Translates company/domain not-found and unique violation to module sentinels. |
| `internal/modules/companies/update_company_domain/repository_test.go` | Modified | Added domain-not-found and unique-violation regression tests. |
| `internal/modules/companies/update_company_domain/service_test.go` | Modified | Updated fakes and assertions; added company-not-found test. |
| `internal/modules/companies/update_company/handler_test.go` | Modified (corrective) | Added HTTP-level 422 contract test for duplicate slug field-keyed response. |
| `internal/modules/companies/create_company_domain/handler_test.go` | Modified (corrective) | Added HTTP-level 422 contract tests for duplicate domain and active primary domain field-keyed responses. |
| `internal/modules/companies/update_company_domain/handler_test.go` | Modified (corrective) | Added HTTP-level 422 contract tests for duplicate domain / active primary domain and 404 contract test for cross-company domain ownership. |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M4.1–M4.5 as complete. |

## Deviations from Design

None — implementation matches the design and the locked scope.

A refinement was required for the `update_company` duplicate-slug case: the enumerated sentinel list did not include a slug-duplicate code, and onboarding already owns `code:duplicate_company_slug`. Companies therefore uses the refined code `code:company_slug_duplicate` to avoid cross-module code duplication while preserving the existing field-keyed 422 response.

`ErrDuplicateCompanyDomain` and `ErrDuplicateCompanySlug` are declared as `apperror.Conflict` (409) to satisfy the M4 acceptance criterion `apperror.Status(core.ErrDuplicateCompanyDomain) == 409`. The create/update handlers still map those sentinels to field-keyed 422 `ValidationErrorResponse`, so the public envelope remains unchanged.

## Corrective Test Hardening

M4 verify flagged that the duplicate-slug / duplicate-domain / active-primary-domain 422 paths were only covered by unit tests on repositories and services, with no HTTP-level test pinning the field-keyed `ValidationErrorResponse` envelope. Added handler contract tests mirroring the M2 onboarding pattern:

- `internal/modules/companies/update_company/handler_test.go` — `TestHandler_Handle_DuplicateSlugValidationError` asserts 422 + field `slug` + `messages.MsgConflict`.
- `internal/modules/companies/create_company_domain/handler_test.go` — `TestHandler_Handle_ValidationErrors` asserts 422 for duplicate domain (field `domain`) and active primary domain (field `kind`).
- `internal/modules/companies/update_company_domain/handler_test.go` — `TestHandler_Handle_ValidationErrors` asserts 422 for duplicate domain (field `domain`) and active primary domain (field `kind`); `TestHandler_Handle_DomainDoesNotBelong` asserts the cross-company ownership case returns 404 `messages.MsgNotFound`.

No production code was changed; the public contract is preserved.

## Issues Found

None.

## Verification

| Command | Outcome |
|---------|---------|
| `go test ./internal/modules/companies/...` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |
| `grep -RE 'gorm\.\|apperror\.' internal/modules/companies/ --include='*service.go' --include='*handler.go' \| grep -v _test.go` | empty |
| `grep -RE 'mapServiceError' internal/modules/companies/ \| grep -v _test.go` | empty |

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M4 — Migrate `companies`
- Estimated M4 diff: ~28 files changed; ~390 insertions / ~101 deletions (~491 changed lines, under the 800-line budget)
- Boundary: Starts after M3e; ends before M5. No auth, no slice-folder migration, no CI/docs work.

## Risks

- `companies/resolver.go` still imports `gorm.io/gorm` and returns raw GORM errors; it was intentionally left out of the M4 scope because the verification grep guard targets only `*service.go` and `*handler.go` files, and the resolver is not a slice service/handler. A future change may want to migrate it for consistency.
- `ErrDuplicateCompanyDomain` and `ErrDuplicateCompanySlug` are declared as `apperror.Conflict` (409) per the M4 acceptance criteria, but the create/update handlers map them to field-keyed 422 responses to preserve the public envelope. Any future caller that routes those sentinels through `response.HandleError` without the handler mapping will return 409 instead of 422.

## Next Recommended

apply-next — proceed to M5 (`auth` migration) once the M4 chained PR is reviewed/merged.
