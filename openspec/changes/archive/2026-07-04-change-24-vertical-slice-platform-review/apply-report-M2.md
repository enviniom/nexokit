# Apply Report — M2

**Change**: change-24-vertical-slice-platform-review
**Work unit**: M2 — Migrate `onboarding` to `apperror` + shared helpers
**Mode**: Standard
**Date**: 2026-07-04

## Completed Tasks

- [x] M2.1 Rewrite `internal/modules/onboarding/core/error.go` to declare module-owned `Code` constants and build `Err*` sentinels via `apperror.Conflict(...)`.
- [x] M2.2 Add `internal/modules/onboarding/core/errors_test.go` (table-driven) pinning `Status`, `Code` (`code:` prefix), and `PublicMessage` for every sentinel.
- [x] M2.3 Add `internal/modules/onboarding/core/dto_test.go` covering each `Validate()` rule per spec.
- [x] M2.4 Add `internal/modules/onboarding/core/model_test.go` (per-model `TableName` direct unit test).
- [x] M2.5 Remove `gorm.io/gorm` and `platform/apperror` imports from `internal/modules/onboarding/onboard_company/service.go`; move GORM translation into the slice repository via `Repository.WithTx`; use `str.NormalizeSlug`, `str.NormalizeDomain`, `str.NormalizeEmail` instead of inlined copies.
- [x] M2.6 Delete `mapServiceError` in `onboard_company/handler.go`; call `response.HandleError(c, err)` directly; drop `platform/apperror` import.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/modules/onboarding/core/error.go` | Modified | Replaced plain `errors.New` sentinels with `apperror.Code` constants and `apperror.Validation(...)` sentinels (status 422) |
| `internal/modules/onboarding/core/errors_test.go` | Created | Table-driven test pinning status (422), code prefix, and public message for all five sentinels |
| `internal/modules/onboarding/core/dto_test.go` | Created | Table-driven coverage of every `OnboardCompanyRequest.Validate()` rule |
| `internal/modules/onboarding/core/model_test.go` | Created | Direct `TableName()` unit tests for all five onboarding models |
| `internal/modules/onboarding/onboard_company/service.go` | Modified | Removed `gorm.io/gorm` and `platform/apperror` imports; delegated transaction boundary to `Repository.WithTx`; used shared normalizers |
| `internal/modules/onboarding/onboard_company/repository.go` | Modified | Added internal `db`, `WithTx`, removed `*gorm.DB` from method signatures, translated unique violations to module conflict sentinels |
| `internal/modules/onboarding/onboard_company/handler.go` | Modified | Reintroduced thin `respondOnboardingError` mapping that returns 422 field-keyed `ValidationErrorResponse` for the five known onboarding sentinels; funnels all other errors through `response.HandleError` |
| `internal/modules/onboarding/container.go` | Modified | Updated `NewRepository(db)` and `NewService(repo, ...)` wiring |
| `internal/modules/onboarding/onboard_company/service_test.go` | Modified | Updated `newService` helper to match new repository/service constructors |
| `internal/modules/onboarding/onboard_company/repository_test.go` | Modified | Updated to use repository-bound methods and added unique-violation translation tests |
| `internal/modules/onboarding/onboard_company/handler_test.go` | Modified | Restored 422 field-keyed `ValidationErrorResponse` expectations for the five known onboarding sentinels |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M2 tasks complete; added corrective-fix note |
| `openspec/changes/change-24-vertical-slice-platform-review/apply-report-M2.md` | Modified | Recorded the M2 corrective fix and preserved public HTTP contract |

## Test Summary

- **Total tests written**: 3 new core test files + repository unique-violation cases + updated service/handler tests
- **Total tests passing**: All onboarding tests pass
- **Layers used**: Unit
- **Approval tests**: None

## Verification Commands

```bash
go test ./internal/modules/onboarding/...                            # PASS
go vet ./...                                                         # PASS
go build ./...                                                       # PASS
go test ./...                                                        # PASS (all packages)
grep -RE 'gorm\.|apperror\.' internal/modules/onboarding/ --include='*service.go' --include='*handler.go' | grep -v _test.go
# (no output)
```

## Corrective Fix Record

Fresh M2 verification failed because the public onboarding HTTP contract had drifted:

- **Expected**: `422 Unprocessable Entity` with `ValidationErrorResponse` carrying field-keyed errors for `slug`, `domain`, `generate_technical_domain`, and `admin_email`.
- **Observed**: `409 Conflict` with flat `ErrorResponse` via `response.HandleError`.
- **Root cause**: `apperror.Conflict(...)` sentinels plus direct `response.HandleError` routing removed the prior per-field mapping.
- **Fix applied**:
  1. Switched onboarding sentinels in `core/error.go` from `apperror.Conflict(...)` to `apperror.Validation(...)`, so the sentinels themselves carry the correct 422 status.
  2. Reintroduced `respondOnboardingError` in `onboard_company/handler.go` to map each known onboarding sentinel to the original field-keyed `ValidationErrorResponse`.
  3. Restored handler tests to assert 422 + field-keyed errors.
  4. Updated sentinel tests in `core/errors_test.go` to expect status 422.
- **Boundary improvements preserved**: `onboard_company/service.go` still has no `gorm.io/gorm` or `platform/apperror` imports; repository still owns persistence translation; shared `str.Normalize*` helpers remain in use.
- **Public contract status**: preserved. No route, payload, status, or envelope change remains.

## Deviations from Design

- `Repository.WithTx` passes a transaction-scoped `Repository` to the callback rather than a `*gorm.DB`, keeping GORM out of the service layer while preserving the existing transaction-per-onboarding semantics.
- The original M2 implementation used `apperror.Conflict(...)` sentinels and routed all service errors directly through `response.HandleError`, which drifted the public HTTP contract from `422 Unprocessable Entity` with field-keyed `ValidationErrorResponse` to `409 Conflict` with a flat `ErrorResponse`. This was caught during fresh M2 verification. The corrective fix switches the sentinels to `apperror.Validation(...)` and reintroduces a thin handler mapping (`respondOnboardingError`) that emits the original field-keyed 422 response for the five known onboarding sentinels. All other errors still flow through `response.HandleError`.

## Issues Found

- `validator.ValidationErrors` is a `map[string][]string` and does not implement `error`, so the service returns `fmt.Errorf("validation failed: %v", errs)` instead of a `%w` wrap. The handler already short-circuits on validation errors via `RespondIfInvalid`, so this path is not exercised in normal HTTP flows.

## Remaining Tasks

- [ ] M3a — Migrate `iam/core` sentinels
- [ ] M3b — Migrate `iam/users` slices
- [ ] M3c — Migrate `iam/roles` slices
- [ ] M3d — Migrate `iam/permissions` + delete duplicate query
- [ ] M3e — Migrate `iam/internal` resolver slices + audit
- [ ] M4 — Migrate `companies` to `apperror` + shared helpers
- [ ] M5 — Migrate `auth` to `apperror`
- [ ] M6 — Wire `apperror` grep guard into Makefile + CI
- [ ] M7 — Publish `docs/module-error-conventions.md`

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M2
- Boundary: M2 only; onboarding module migration, no IAM/companies/auth/gormutil changes
- Estimated review budget impact: ~260 src / ~180 test as forecast; actual diff is within the 800-line work-unit budget

## Status

M2 complete. Ready for verify or next apply batch (M3a).
