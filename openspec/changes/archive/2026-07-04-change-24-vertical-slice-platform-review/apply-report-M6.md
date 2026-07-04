# Apply Report — M6: Wire `apperror` Grep Guard into Makefile + CI

## Status

success

## Executive Summary

Added the `check-module-errors` Makefile target and wired it into the existing `lint` target and the CI pipeline as a new `module-errors-guard` job. The guard enforces the change-24 boundary contract: production module `*service.go`, `*repository.go`, and `*handler.go` files under `internal/modules/` must not import or reference `platform/apperror` directly, module services must not reference `gorm.io/gorm` directly, and legacy `mapServiceError` adapters are forbidden in non-test code. Tests are explicitly excluded from all three checks.

The guard passed against the current M0–M5 state: no violations remain in the targeted file sets. A negative test confirmed the target exits with a non-zero status and prints the offending file when a synthetic `apperror.` reference is introduced.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `Makefile` | Modified | Added `check-module-errors` to `.PHONY`; added the target with three empty-grep checks (`apperror.` in handler/service/repository, `gorm.` in service, `mapServiceError` anywhere in modules); wired it into `lint: vet check-module-errors`. |
| `.github/workflows/ci.yml` | Modified | Added a `module-errors-guard` job that checks out the repo, sets up Go, and runs `make check-module-errors`. |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M6.1–M6.4 as complete. |

## Deviations from Design

None — implementation matches the design and the locked scope.

## Issues Found

None.

## Verification

| Command | Outcome |
|---------|---------|
| `make check-module-errors` | PASS (zero exit, all three grep checks empty) |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |
| Negative test: append `// apperror.X` to `internal/modules/auth/view_session/service.go` and run `make check-module-errors` | FAIL with non-zero exit and clear file/line output; reverted immediately. |

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M6 — Wire `apperror` grep guard into Makefile + CI
- Estimated M6 diff: ~2 files changed; ~35 lines added (Makefile + CI), well under the 800-line budget.
- Boundary: Starts after M5; ends before M7. No docs, no runtime behavior change.

## Risks

- The guard uses filename suffixes (`*service.go`, `*repository.go`, `*handler.go`) rather than package roles, so any future file that happens to end with those suffixes but is not a service/repository/handler would still be scanned. This matches the spec and is the intended, low-cost enforcement.
- `gorm.` is only forbidden in `*service.go`, not in repositories or handlers, because repositories legitimately use GORM and handlers must remain persistence-agnostic.
- The `lint` target now depends on `check-module-errors`; local `make lint` will fail if the boundary is violated, which is the intended behavior.

## Next Recommended

verify — run the full verification suite and proceed to the verify phase before M7 (docs).
