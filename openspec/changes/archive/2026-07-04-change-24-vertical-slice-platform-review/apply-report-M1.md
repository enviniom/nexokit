# Apply Report — M1

**Change**: change-24-vertical-slice-platform-review
**Work unit**: M1 — Extend `internal/platform/gormutil` with `IsUniqueConstraintError`
**Mode**: Strict TDD
**Date**: 2026-07-04

## Completed Tasks

- [x] M1.1 Add `internal/platform/gormutil/unique.go` exporting `IsUniqueConstraintError(err error) bool`
- [x] M1.2 Add `internal/platform/gormutil/unique_test.go` with table-driven coverage
- [x] M1.3 Leave `internal/modules/iam/queries/normalize_slugs.go` untouched

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `internal/platform/gormutil/unique.go` | Created | Exported `IsUniqueConstraintError` matching Postgres + SQLite/current behavior; nil safe |
| `internal/platform/gormutil/unique_test.go` | Created | Table-driven test covering all spec scenarios |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M1 tasks complete |

## TDD Cycle Evidence

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|------|-----------|-------|------------|-----|-------|-------------|----------|
| M1.1/M1.2 | `internal/platform/gormutil/unique_test.go` | Unit | N/A (new file) | Written (compile error: undefined `IsUniqueConstraintError`) | ✅ Passed (9/9 rows) | ✅ 9 cases covering gorm sentinel, Postgres, SQLite, generic SQL, connection, nil | ➖ None needed |

## Test Summary

- **Total tests written**: 1 table-driven test with 9 sub-cases
- **Total tests passing**: 9/9
- **Layers used**: Unit (1)
- **Approval tests**: None — no refactoring tasks
- **Pure functions created**: 1 (`IsUniqueConstraintError`)

## Verification Commands

```bash
go test ./internal/platform/gormutil/... -run Unique -v   # PASS (9/9)
go vet ./...                                              # PASS
go build ./...                                            # PASS
go test ./...                                             # PASS (all packages)
```

## Deviations from Design

None — implementation matches design and spec.

## Issues Found

None.

## Remaining Tasks

- [ ] M2 — Migrate `onboarding` to `apperror` + shared helpers
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
- Current work unit: M1
- Boundary: M1 only; no module callers updated, no module files touched
- Estimated review budget impact: ~60 src / ~120 test (~180 total), within 800-line budget

## Status

M1 complete. Ready for verify or next apply batch (M2).
