# Proposal: Vertical-Slice & Platform/Shared Review (Module-by-Module)

## Intent

Bring the four production modules (`auth`, `iam`, `companies`, `onboarding`) back onto the canonical error and shared-helper contract from `docs/modules/validation-and-errors.md` and `docs/modules/boundaries-and-dependencies.md`. The vertical-slice skeleton is in place, but every module drifts from the contract in the same three places: `core/errors.go` uses plain `errors.New(...)`, services import `gorm.io/gorm` and `apperror` directly, and handlers carry a `mapServiceError(err)` switch. Pure refactor — no new business features, no public surface change.

## Scope

### In Scope

- **AppError migration per module.** Rewrite `core/errors.go` in `auth`, `iam`, `companies`, `onboarding` to use `platform/apperror` helpers with module-owned `Code` constants in the format `code:<snake_case>` (e.g. `code:user_not_found`).
- **Handler simplification.** Delete every `mapServiceError(err)` switch in slice handlers; funnel business errors through `response.HandleError(c, err)` only.
- **Service/repository cleanup.** Remove `gorm.io/gorm` and `apperror` imports from every `service.go`; repositories translate `gorm.ErrRecordNotFound` to module errors at the data boundary.
- **Shared helper extraction.** Create focused package `internal/platform/shared/string` (`NormalizeSlug`, `NormalizeDomain`, `NormalizeEmail`) and extend existing `internal/platform/gormutil` with `IsUniqueConstraintError`, both with table-driven tests; replace duplicated sites in `iam`, `companies`, `onboarding`.
- **Duplicate IAM query removal.** Delete `internal/modules/iam/queries/get_role_by_public_id_preloads.go` (byte-identical to `get_role_by_public_id.go`); pin preload behavior with a regression test.
- **Module error-convention docs and tests.** Add `core/error_test.go`, `core/dto_test.go`, `core/model_test.go` (`TableName`) for modules that lack them; publish `docs/module-error-conventions.md` cross-linked from `docs/modules/validation-and-errors.md`.
- **CI grep guard.** `grep -RE 'apperror\.' internal/modules/ --include='*service.go' --include='*repository.go'` MUST be empty after M6.

### Out of Scope

- No public route, payload, or status-code change. HTTP codes are preserved by the `apperror` helpers.
- No slice-folder migration. Module-root slice placement stays as-is; deferred to a follow-up change.
- No DB migration, no business feature, no spec-level capability change.
- No `platform/shared/slug` (full text-to-slug) — no consumer today.

## Capabilities

### New Capabilities
None

### Modified Capabilities
None

> Pure refactor onto the existing `error-handling` and `platform-boundary-rules` contracts. Module-owned `Code` constants are already mandated by `platform-boundary-rules.spec.md`; no new spec-level requirements are introduced.

## Approach

Execute the staged `M0..M7` plan from `explore.md` §4 — one branch + one PR per step. Order: shared helpers (M0/M1) → onboarding (M2) → iam core + sub-areas (M3a..M3e) → companies (M4) → auth (M5) → CI grep guard (M6) → module error-convention doc (M7). M2 validates the pattern on the smallest end-to-end migration before touching iam.

## Affected Areas

| Area | Impact |
|------|--------|
| `internal/modules/{auth,iam,companies,onboarding}/core/errors.go` | Modified — sentinel swap to `apperror` + module-owned `Code` |
| `internal/modules/{auth,iam,companies,onboarding}/slices/**/{handler,service,repository}.go` | Modified — boundary cleanup; no `apperror`/`gorm` in `service.go`; no `mapServiceError` in `handler.go` |
| `internal/platform/shared/string/` | New — `NormalizeSlug`, `NormalizeDomain`, `NormalizeEmail` + tests |
| `internal/platform/gormutil/` | Modified — add `IsUniqueConstraintError` + tests |
| `internal/modules/iam/queries/get_role_by_public_id_preloads.go` | Removed — preload behavior pinned by regression test |
| `internal/modules/{auth,companies,onboarding}/core/{error,dto,model}_test.go` | New — table-driven coverage |
| `docs/module-error-conventions.md` | New — canonical `core.Error` → `Code` → `HTTPStatus` table |
| Makefile / CI | Modified — grep guard forbids `apperror.` in `service.go` / `repository.go` |

## Risks

- **Review PR exceeds 800-line budget** (High; esp. iam — ~10.4k LOC, 22+ slices). Mitigation: split iam migration into M3a (core) + M3b (users, 8) + M3c (roles, 8) + M3d (permissions, 3 + duplicate cleanup) + M3e (internal, 5); each PR stays under 800.
- **Behavior drift from sentinel swap** (Med). Mitigation: per-module `core/error_test.go` pins `Status`, `Code`, `PublicMessage`; HTTP status preserved by `apperror` helpers; HTTP envelope unchanged.
- **Substring `IsUniqueConstraintError` fragility** (Low). Mitigation: focused unit tests on Postgres unique violation + generic error path; helper is colocated with the existing `platform/gormutil`.
- **Spec-vs-docs disagreement on slice placement** (Low). Mitigation: this change does not move slices; the disagreement is documented and deferred to a follow-up change.
- **Import-graph breakage when `apperror` is removed from services** (Low). Mitigation: per-module isolation; `go build ./...` + `go test ./...` runs per PR; service-level tests still cover business outcomes via module sentinels.

## Rollback Plan

Each M-step is an independent branch. `git revert <merge-sha>` restores the module state. No DB migration, no shared state, no feature flag. Shared-helper extraction is additive; reverting restores the inlined helpers. The duplicate IAM query deletion is reversible by restoring the file from git history (the regression test stays).

## Dependencies

- `platform/apperror` (already merged): provides `NotFound`, `Conflict`, `Forbidden`, `BadRequest`, `Unauthorized`, `Validation`, `Unprocessable`, `Wrap` helpers and HTTP-status mapping.
- `platform/gormutil` (already exists): M0 extends it with `IsUniqueConstraintError`; GORM-specific helpers stay in `internal/platform/gormutil`, while pure string normalization lives in `internal/platform/shared/string`.

## Success Criteria

- [ ] `go vet ./...`, `go build ./...`, `go test ./...` clean after every M-step.
- [ ] No `apperror.` import in any `service.go` or `repository.go` after M6.
- [ ] No `gorm.` import in any `service.go` after M5.
- [ ] Every module has `core/error_test.go`, `core/dto_test.go`, `core/model_test.go` (where partial GORM models exist).
- [ ] `get_role_by_public_id_preloads.go` deleted; preload behavior covered by a regression test.
- [ ] `docs/module-error-conventions.md` published and cross-linked from `docs/modules/validation-and-errors.md`.
- [ ] No public route, payload, or status-code change; HTTP envelope unchanged.
- [ ] Module-root slice placement unchanged.
- [ ] Each M-step PR stays under 800 changed lines.
