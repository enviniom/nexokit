# Apply Report — M7: Publish `docs/module-error-conventions.md`

## Status

success

## Executive Summary

Created the canonical module-error conventions document and cross-linked it from the validation-and-errors guide. The doc lists every reusable sentinel introduced in M2–M5 for `auth`, `iam`, `companies`, and `onboarding`, with `Code`, `HTTPStatus`, `PublicMessage`, and usage notes that explain intentional public-message preservation and field-keyed 422 handler mappings.

## Files Changed

| File | Action | What Was Done |
|------|--------|---------------|
| `docs/module-error-conventions.md` | Created | Canonical per-module sentinel table, conventions section, layer rules, enforcement notes, and review checklist. |
| `docs/modules/validation-and-errors.md` | Modified | Added one relative cross-link to `docs/module-error-conventions.md` in the `core/errors.go` pattern section. |
| `openspec/changes/change-24-vertical-slice-platform-review/tasks.md` | Modified | Marked M7.1–M7.3 as complete. |

## Deviations from Design

None — implementation matches the design and the module-error-conventions spec.

## Issues Found

None.

## Verification

| Command | Outcome |
|---------|---------|
| `make check-module-errors` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |
| `grep -RE 'module-error-conventions' docs/` | Found link in `docs/modules/validation-and-errors.md` and new doc. |
| Docs/link checker | No dedicated link checker exists; `tests/docs` only covers the multitenancy guide. |

## Workload / PR Boundary

- Mode: chained PR slice (stacked-to-main)
- Current work unit: M7 — Publish module error conventions doc
- Estimated M7 diff: ~2 files changed; ~150 lines added (one new doc + one link), well under the 800-line budget.
- Boundary: Starts after M6; ends the change-24 apply sequence. No runtime behavior change.

## Risks

- The doc is a manual table; future sentinel additions require authors to update it. The review checklist in the doc and the existing module-error tests reduce the chance of drift.
- No automated link checker is present, so the relative cross-link was verified by grep.

## Next Recommended

verify — all M0–M7 work units are implemented; proceed to the verify phase before archive.
