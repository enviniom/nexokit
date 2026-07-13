# Apply Progress: Align Companies Module Conventions

## Status

- Artifact store: OpenSpec.
- Mode: **Standard**. `openspec/config.yaml` does not declare `strict_tdd: true`; its `rules.apply.tdd: true` setting does not activate Strict TDD under the orchestrator contract.
- Delivery: single PR with maintainer-approved `size:exception`, maximum 800 authored lines.
- Tasks: **15/15 complete**; task checkboxes remain the source of completion status in `tasks.md`.
- Inventory: **18 repository-method paths** and **19 GORM operations**. `list.List` performs separate `Count` and `Find` operations.

## Completed Work

- Rehomed all seven companies slices beneath `internal/modules/companies/slices` and updated only container/test imports.
- Renamed `core/error.go` to `core/errors.go`; added company and company-domain persistence AppError constructors and unary error mappers.
- Mapped all repository persistence operations, including zero-row updates/deletes and active-primary-domain counts; unknown causes remain unwrap-able.
- Added mapper/core tests, SQLite zero-row tests, and a recursive AST repository guard with direct, variable-held, nested, same-name, transaction-style, and same-function mapped-then-raw selector fixtures.

## Targeted TDD Evidence

This is targeted RED/GREEN evidence recorded for mapper and guard work. It is not a claim of Strict TDD compliance and does not invent per-task TDD cycles.

| Scope | RED evidence | GREEN evidence | Result |
|---|---|---|---|
| Repository guard, same-function selector collision | `go test ./internal/modules/companies/queries -run '^TestCompanyRepositoryGuardFixtures/same_function_mapped_selector_then_raw_selector$' -count=1` — FAIL before the guard change: `problems=[], want problem=true` | `go test ./internal/modules/companies/queries -run 'TestCompanyRepositoryGuardFixtures|TestCompanyRepositoriesUseEntityMappers' -count=1` — PASS | The guard became occurrence/provenance-aware; direct, variable-held, nested, separate-functions, mapped-variable, and single-result cases remained covered. |
| Mapper and persistence work | RED/GREEN work is recorded in completed tasks 1.2, 1.3, 1.4, 2.1, 2.2, and 2.5. | Focused and regression commands below passed. | Evidence is targeted to the recorded mapper/guard work, not a full Strict-TDD task-cycle matrix. |

## Work Unit Evidence

| Unit | Focused test command and exact result | Runtime harness command/scenario and exact result | Rollback boundary |
|---|---|---|---|
| 1 — errors and mappers | `go test ./internal/modules/companies/core ./internal/modules/companies/queries` — PASS | N/A: pure error and AST behavior | `core/errors.go`, `queries/map_errors.go`, related tests |
| 2 — slice rehome | `go test ./internal/modules/companies/...` — PASS | Existing httptest route/resolver coverage — PASS in companies suite | seven `slices/*` packages and container imports |
| 3 — persistence mapping | `go test ./internal/modules/companies/... -run 'Repository|Count'` — PASS | SQLite-backed repository/query cases — PASS | seven repository adapters and repository/query tests |
| 4 — boundary guard and final checks | `go test ./internal/modules/companies/queries -run 'TestCompanyRepositoryGuardFixtures|TestCompanyRepositoriesUseEntityMappers' -count=1` — PASS | `go test ./...` — PASS; `go build ./...` — PASS | `map_errors_structure_test.go` guard/fixtures and this evidence artifact |

## Regression and Final Checks

- `go test ./internal/modules/companies/... -count=1` — PASS.
- `go test ./...` — PASS.
- `go build ./...` — PASS.
- `git diff --check` — PASS.
- The guard fixture exposed a test-guard false negative, not a repository persistence leak; no production change was made for that focused remediation.

## Deviations from Design

None — implementation matches the design. This persistence correction replaces the inaccurate Strict TDD classification with Standard mode and retains only factual targeted RED/GREEN evidence.

## Authored-Line Budget

Baseline: current `HEAD`; count tracked diff plus untracked files, excluding only pre-existing planning artifacts: `openspec/changes/align-companies-module-conventions/proposal.md`, `openspec/changes/align-companies-module-conventions/specs/**`, `openspec/changes/align-companies-module-conventions/design.md`, and `openspec/changes/align-companies-module-conventions/tasks.md`.

```bash
{ git diff --numstat -- . ':(exclude)openspec/changes/align-companies-module-conventions/proposal.md' ':(exclude)openspec/changes/align-companies-module-conventions/specs/**' ':(exclude)openspec/changes/align-companies-module-conventions/design.md' ':(exclude)openspec/changes/align-companies-module-conventions/tasks.md'; git ls-files --others --exclude-standard | while IFS= read -r path; do case "$path" in openspec/changes/align-companies-module-conventions/proposal.md|openspec/changes/align-companies-module-conventions/specs/*|openspec/changes/align-companies-module-conventions/design.md|openspec/changes/align-companies-module-conventions/tasks.md) ;; *) git diff --no-index --numstat /dev/null "$path" || test $? -eq 1 ;; esac; done; } | awk '{add += $1; del += $2} END {printf "authored_lines=%d (additions=%d + deletions=%d)\n", add + del, add, del}'
```

Formula/result: additions + deletions = **675 + 122 = 797 authored lines**, within the accepted **800-line** exception.

## Persistence Scope

- This update changes evidence classification and persistence only.
- No task checkboxes, production code, tests, commands, commits, or review state were changed by this correction.
