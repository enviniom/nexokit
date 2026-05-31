# Apply Progress: 20-validation-errors-boundary

- Mode: Strict TDD
- Delivery: single PR/work unit (budget risk low; no size exception)

## Completed Tasks

- [x] 1.1 Move `ValidationErrors` type, `Add()`, `HasErrors()` to `validator`
- [x] 1.2 Remove `response` dependency from `validator.FieldValidator`/`Field`
- [x] 1.3 Update `response` to consume `validator.ValidationErrors`
- [x] 1.4 Update `validator` tests to local `ValidationErrors`
- [x] 1.5 Update `response` tests to `validator.ValidationErrors`
- [x] 2.1 Update module DTO validation return/make types to `validator.ValidationErrors`
- [x] 2.2 Update handlers that manually build validation maps to `validator.ValidationErrors`
- [x] 3.1 Update golden DTO expected output to `validator.ValidationErrors`
- [x] 3.2 Build verification: `go build ./...`
- [x] 3.3 Test verification: `go test ./...`
- [x] 3.4 Dependency direction check: `validator` has zero imports from `response`
- [x] 3.5 Golden/API shape verification: `tests/cli` pass; response tests pass unchanged envelope

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| 1.1 | `internal/platform/validator/validator_test.go` | Unit | ✅ `go test ./internal/platform/validator ./internal/platform/response ./internal/modules/... ./tests/cli/...` | ⚠️ Existing tests were adapted for moved type ownership (refactor/approval style) | ✅ `go test ./internal/platform/validator -run TestValidationErrors` | ➖ Structural move (type ownership refactor) | ✅ Minimal cleanup with unchanged behavior |
| 1.2 | `internal/platform/validator/validator_test.go` | Unit | ✅ same baseline run | ⚠️ Approval-style refactor; no new behavior | ✅ `go test ./internal/platform/validator -run TestValidationErrors` | ➖ Structural | ✅ import boundary cleaned |
| 1.3 | `internal/platform/response/response_test.go` | Unit | ✅ same baseline run | ⚠️ Approval-style refactor; no new behavior | ✅ `go test ./internal/platform/response -run TestRespondIfInvalid` | ➖ Structural | ✅ type boundary only |
| 1.4 | `internal/platform/validator/validator_test.go` | Unit | ✅ same baseline run | ⚠️ Structural adaptation | ✅ package tests passing | ➖ Structural | ➖ None needed |
| 1.5 | `internal/platform/response/response_test.go` | Unit | ✅ same baseline run | ⚠️ Structural adaptation | ✅ package tests passing | ➖ Structural | ➖ None needed |
| 2.1 | module DTO files | Unit/Compile | ✅ same baseline run | ⚠️ Structural adaptation | ✅ `go build ./...` | ➖ Structural | ➖ None needed |
| 2.2 | module handlers | Unit/Compile | ✅ same baseline run | ⚠️ Structural adaptation | ✅ `go build ./...` | ➖ Structural | ➖ None needed |
| 3.1-3.5 | golden + full suite | Integration/Contract | ✅ same baseline run | ✅ Golden mismatch detected and fixed via template alignment | ✅ `go test ./...` | ✅ golden path + full suite pass | ✅ kept JSON envelope unchanged |

## Test Summary

- Total tests written: 0 new (change is ownership refactor; existing tests used as approval tests)
- Total tests passing: full suite passing (`go test ./...`)
- Layers used: Unit, Integration/Contract
- Approval tests (refactoring): existing platform/response/validator/module tests
- Pure functions created: 0

## Notes

- During verification, `tests/cli` surfaced a golden mismatch because generator template `internal/cli/templates/module/dto.tmpl` still emitted `response.ValidationErrors`. Template was updated to `validator.ValidationErrors` to keep generated code and golden fixtures aligned.
