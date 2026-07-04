# Apply Report — M0

**Change**: change-24-vertical-slice-platform-review  
**Work unit**: M0 — Create `internal/platform/shared/string` package  
**Mode**: Strict TDD  
**Date**: 2026-07-04  

## Completed Tasks

- [x] M0.1 Create `internal/platform/shared/string/normalize_slug.go` exporting `func NormalizeSlug(s string) string` (returns `strings.ToLower(strings.TrimSpace(s))`).
- [x] M0.2 Create `internal/platform/shared/string/normalize_domain.go` exporting `func NormalizeDomain(s string) string` (`TrimSpace` → `ToLower` → `TrimSuffix(".")`).
- [x] M0.3 Create `internal/platform/shared/string/normalize_email.go` exporting `func NormalizeEmail(s string) string` (`TrimSpace` → `ToLower`).
- [x] M0.4 Create `internal/platform/shared/string/normalize_*_test.go` with one table-driven test per helper covering empty, whitespace, mixed case, trailing dot, control chars.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/platform/shared/string/normalize_slug.go` | Created | `NormalizeSlug` implementation |
| `internal/platform/shared/string/normalize_domain.go` | Created | `NormalizeDomain` implementation |
| `internal/platform/shared/string/normalize_email.go` | Created | `NormalizeEmail` implementation |
| `internal/platform/shared/string/normalize_slug_test.go` | Created | Table-driven tests for slug normalizer |
| `internal/platform/shared/string/normalize_domain_test.go` | Created | Table-driven tests for domain normalizer |
| `internal/platform/shared/string/normalize_email_test.go` | Created | Table-driven tests for email normalizer |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M0 tasks complete |

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| M0.1 | `internal/platform/shared/string/normalize_slug_test.go` | Unit | N/A (new) | Written | Passed | 8 cases | None needed |
| M0.2 | `internal/platform/shared/string/normalize_domain_test.go` | Unit | N/A (new) | Written | Passed | 9 cases | None needed |
| M0.3 | `internal/platform/shared/string/normalize_email_test.go` | Unit | N/A (new) | Written | Passed | 9 cases | None needed |

### Test Summary
- **Total tests written**: 26 (8 slug + 9 domain + 9 email)
- **Total tests passing**: 26
- **Layers used**: Unit (26)
- **Approval tests**: None — no refactoring tasks
- **Pure functions created**: 3

## Deviations from Design

One intentional naming choice:

- The package directory is `internal/platform/shared/string` as required, but the package clause is `package str` rather than `package string`. Using `package string` would make the default import identifier `string`, shadowing the predeclared `string` type at every call site. `package str` keeps the required import path while avoiding builtin shadowing.

No behavioral deviations.

## Issues Found

None.

## Remaining Tasks

- [ ] M1 — Extend `internal/platform/gormutil` with `IsUniqueConstraintError`
- [ ] M2 — Migrate `onboarding` to `apperror` + shared helpers
- [ ] M3a — Migrate `iam/core` sentinels
- [ ] M3b — Migrate `iam/users` slices
- [ ] M3c — Migrate `iam/roles` slices
- [ ] M3d — Migrate `iam/permissions` + delete duplicate query
- [ ] M3e — Migrate `iam/internal` resolver slices + audit
- [ ] M4 — Migrate `companies` to `apperror` + shared helpers
- [ ] M5 — Migrate `auth` to `apperror`; pivot tests
- [ ] M6 — Wire `apperror` grep guard into Makefile + CI
- [ ] M7 — Publish `docs/module-error-conventions.md`

## Workload / PR Boundary

- **Delivery strategy**: `chained-pr`
- **Chain strategy**: `stacked-to-main`
- **Current work unit**: M0
- **Boundary**: Additive shared helper package only; no callers updated, no modules touched
- **Estimated review budget impact**: ~120 src / ~200 test (~320 total), under 400-line budget

## Verification Commands

```bash
go test ./internal/platform/shared/string/... -v   # PASS (26/26)
go vet ./...                                       # PASS
go build ./...                                     # PASS
go test ./...                                      # PASS (all cached + new package)
```

## Status

4/4 M0 tasks complete. Ready for next batch (M1).
