# Proposal: Validation Errors Ownership Boundary

## Intent

`ValidationErrors` (error accumulator) lives in `platform/response` but is a validation primitive. This forces `platform/validator` → `platform/response` dependency — lower-level imports higher-level HTTP layer. This change corrects that boundary.

## Scope

### In Scope
- Move `ValidationErrors` type, `Add()`, `HasErrors()` from `response` to `validator`
- Update `response` to import `validator` for the type
- Adjust imports across DTOs, handlers, tests (~54 refs, most already import both)
- Update 3 specs to reflect new boundary

### Out of Scope
- JSON response shape, HTTP status codes, `RespondIfInvalid` behavior
- `ValidationErrorResponse` structure or `APIResponse` envelope
- Validator rules, `FieldValidator`, or `Rule` functions
- Module-level validation logic

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `request-validation`: Clarify `ValidationErrors` originates from `validator`; `RespondIfInvalid` stays in `response`
- `api-response`: Clarify `ValidationErrorResponse` uses `validator.ValidationErrors`
- `platform-boundary-rules`: Update classification — `validator` owns validation primitives including `ValidationErrors`

## Approach

Move type definition from `response/response.go` to `validator/validator.go`. Keep HTTP concerns (`ValidationErrorResponse`, `ValidationError()`, `RespondIfInvalid()`) in `response`. Flips dependency: `response` imports `validator` — correct since response is higher-level HTTP layer consuming validation results.

No logic changes. Zero behavior change.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/platform/response/response.go` | Modified | Remove type; import `validator` |
| `internal/platform/validator/validator.go` | Modified | Add type + methods |
| `internal/platform/validator/validator_test.go` | Modified | Swap import |
| `internal/modules/*/core/dto.go` (7) | Modified | Update refs |
| `internal/modules/*/handler.go` (12) | Modified | Update refs |
| `tests/cli/testdata/golden/goldenmod/dto.go` | Modified | Update golden ref |
| `openspec/specs/request-validation/spec.md` | Modified | Clarify boundary |
| `openspec/specs/api-response/spec.md` | Modified | Clarify boundary |
| `openspec/specs/platform-boundary-rules/spec.md` | Modified | Update table |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Import churn (~20 files) | Medium | Most import both; `go build` catches all |
| Golden test breakage | Low | Update in same batch |
| Accidental behavior change | Low | Pure type move; tests verify |

## Rollback Plan

Revert git commit. Atomic change — all in single commit. If needed, temporarily add type alias in `response` pointing to `validator.ValidationErrors`, then remove.

## Dependencies

None.

## Success Criteria

- [ ] `go build ./...` passes
- [ ] `go test ./...` passes — no test logic changes
- [ ] `validator` has zero imports from `response`
- [ ] `response` imports `validator` for `ValidationErrors` only
- [ ] API JSON shape identical (golden tests pass)
- [ ] Specs reflect corrected boundary
