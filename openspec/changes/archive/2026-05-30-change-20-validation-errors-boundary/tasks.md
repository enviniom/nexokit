# Tasks: Validation Errors Ownership Boundary

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~150-200 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-always |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Move type + update platform packages | PR 1 | validator.go, response.go, both test files |
| 2 | Update module DTOs and handlers | PR 1 | 7 DTOs + ~17 handlers, pure import swap |
| 3 | Update golden + verify build/tests | PR 1 | golden dto.go + full test suite |

## Phase 1: Foundation — Move Type to Validator

- [x] 1.1 Move `ValidationErrors` type, `Add()`, `HasErrors()` from `internal/platform/response/response.go` to `internal/platform/validator/validator.go`
- [x] 1.2 Update `internal/platform/validator/validator.go`: replace `response.ValidationErrors` with local `ValidationErrors` in `FieldValidator.errs` and `Field()` param; remove `response` import
- [x] 1.3 Update `internal/platform/response/response.go`: add `validator` import; change `ValidationErrorResponse.Errors` type to `validator.ValidationErrors`; update `ValidationError()` and `RespondIfInvalid()` to use `validator.ValidationErrors`
- [x] 1.4 Update `internal/platform/validator/validator_test.go`: replace `response` import with local `ValidationErrors` in all 8 test functions
- [x] 1.5 Update `internal/platform/response/response_test.go`: add `validator` import; change 2 `make(ValidationErrors)` to `make(validator.ValidationErrors)`

## Phase 2: Module Import Updates

- [x] 2.1 Update 7 DTO files: replace `response.ValidationErrors` with `validator.ValidationErrors` in return types and `make()` calls
  - `internal/modules/auth/core/dto.go`, `internal/modules/users/dto.go`, `internal/modules/roles/dto.go`
  - `internal/modules/companies/core/dto.go`, `internal/modules/companies/dto.go`
  - `internal/modules/permissions/core/dto.go`, `internal/modules/onboarding/core/dto.go`
- [x] 2.2 Update ~17 handler files: replace `make(response.ValidationErrors)` with `make(validator.ValidationErrors)` (handlers keep `response` import for `RespondIfInvalid`, `Success`, etc.)
  - All `internal/modules/*/handler.go` and `internal/modules/*/*/handler.go` files with `response.ValidationErrors` usage

## Phase 3: Golden Test + Verification

- [x] 3.1 Update `tests/cli/testdata/golden/goldenmod/dto.go`: change `response.ValidationErrors` to `validator.ValidationErrors` in golden expected output
- [x] 3.2 Run `go build ./...` — verify zero compilation errors
- [x] 3.3 Run `go test ./...` — verify all tests pass with no behavior changes
- [x] 3.4 Verify `validator` package has zero imports from `response` (dependency direction confirmed)
- [x] 3.5 Verify API JSON shape unchanged — golden tests pass identically
