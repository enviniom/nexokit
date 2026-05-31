# Archive Report: 20-validation-errors-boundary

**Change**: 20-validation-errors-boundary
**Archived to**: `openspec/changes/archive/2026-05-30-20-validation-errors-boundary/`
**Date**: 2026-05-30
**Mode**: openspec

## Summary

Moved `ValidationErrors` type, `Add()`, and `HasErrors()` from `platform/response` to `platform/validator`, correcting the dependency direction. `response` now imports `validator` for the type — higher-level HTTP layer correctly depends on lower-level validation primitives. Zero behavior change, pure structural refactor.

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| request-validation | Updated | 3 requirements modified: ValidationErrors accumulator (now in validator), Gin helper (param type validator.ValidationErrors), Field-keyed responses (references updated) |
| api-response | Updated | 1 requirement modified: Explicit response DTO names (ValidationErrorResponse.Errors uses validator.ValidationErrors) |
| platform-boundary-rules | Updated | 1 requirement modified: Platform package classification table (validator now owns ValidationErrors, response imports validator) |

## Archive Contents

- proposal.md ✅
- exploration.md ✅
- specs/ ✅ (3 delta specs)
- design.md ✅
- tasks.md ✅ (15/15 tasks complete)
- apply-progress.md ✅
- verify-report.md ✅ (PASS — 0 CRITICAL, 0 WARNING)

## Verification Status

- Build: ✅ `go build ./...` passes
- Tests: ✅ `go test ./...` all 52 packages pass
- Coverage: validator 100%, response 95%
- Dependency direction: validator has zero imports from response ✅
- API JSON shape: unchanged (golden tests pass) ✅

## Source of Truth Updated

The following main specs now reflect the corrected boundary:
- `openspec/specs/request-validation/spec.md`
- `openspec/specs/api-response/spec.md`
- `openspec/specs/platform-boundary-rules/spec.md`

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived.
Ready for the next change.
