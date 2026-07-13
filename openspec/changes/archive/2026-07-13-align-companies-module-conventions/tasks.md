# Tasks: Align Companies Module Conventions

## Review Workload Forecast

| Field | Value |
|---|---|
| Authored changed lines | 760–800 |
| 800-line budget risk | Medium |
| Chained PRs recommended | No (single-pr-default) |
| Suggested split | Four atomic work-unit commits in one ordered PR |
| Delivery strategy | single-pr-default |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: High
800-line budget risk: Medium

### Suggested Work Units

| Unit | Goal / commit | Focused test command | Runtime harness | Rollback boundary |
|---|---|---|---|---|
| 1 | `test(core): define mapping contract` | `go test ./internal/modules/companies/core ./internal/modules/companies/queries` | N/A: pure errors/AST | `core/errors.go`, mapper and tests |
| 2 | `refactor(companies): rehome slices` | `go test ./internal/modules/companies/...` | httptest routes/resolver | seven moved packages + container imports |
| 3 | `fix(companies): total persistence mapping` | `go test ./internal/modules/companies/... -run 'Repository|Count'` | sqlite-backed repository cases | repository adapters, query tests |
| 4 | `test(companies): enforce boundaries` | `go test ./internal/modules/companies/...` | `go test ./... && go build ./...` | guards, fixtures, SDD evidence |

## Phase 1: Baseline and Error Contract

- [x] 1.1 Record the current routes/payloads/statuses, aliases, resolver behavior, and exact inventory of 18 repository-method paths and 19 GORM operations in `openspec/changes/align-companies-module-conventions/design.md` evidence notes.
- [x] 1.2 RED: add table tests in `internal/modules/companies/core/errors_test.go` for six sentinel codes, constructors, `errors.As/Is`, and unknown-cause unwrap.
- [x] 1.3 GREEN: rename `core/error.go` to `core/errors.go`; add persistence codes and company/domain `apperror.Internal` constructors.
- [x] 1.4 RED/GREEN: add `internal/modules/companies/queries/map_errors_test.go` and `map_errors.go` unary nil, known, wrapped-known, unique, existing-AppError, and unknown cases.

## Phase 2: Layout and Persistence

- [x] 2.1 RED: extend `internal/modules/companies/queries/count_active_primary_domains_test.go` for create-ignore-exclude and update-self-exclude.
- [x] 2.2 GREEN: keep conditional exclusion in `queries/count_active_primary_domains.go`; make create pass `""` and update pass its public ID.
- [x] 2.3 Atomically move all files/tests from `companies/{list_companies,view_company,update_company,delete_company,list_company_domains,create_company_domain,update_company_domain}` to `companies/slices/`.
- [x] 2.4 Update `container.go` imports/wiring only; preserve `routes.go`, `resolver.go`, root `model.go`/`dto.go` aliases, `Resolver()`, and tenant resolver calls.
- [x] 2.5 RED/GREEN: migrate every 18 repository-method paths and 19 GORM operations in the seven `slices/*/repository.go` files: K/U mapper behavior, empty `Find` success, and `RowsAffected==0` company/domain not-found.

## Phase 3: Boundary Guards and Compatibility

- [x] 3.1 RED: create `queries/map_errors_structure_test.go` fixtures for direct, variable-held, nested, same selector in two functions, a same-function mapped-then-raw `result.Error` collision, and single-result Transaction-style errors.
- [x] 3.2 GREEN: implement a recursive function-scoped AST guard that discovers `slices/**/repository.go`, tracks occurrence identities and variable provenance, and rejects raw GORM, mapper bypass, or `apperror` interfaces.
- [x] 3.3 Add/adjust `routes_absence_test.go`, `resolver_test.go`, and handler tests to prove seven routes, `:id` PublicID, envelopes/statuses, aliases, and absent `POST /api/v1/companies`.

## Phase 4: Verification and Evidence

- [x] 4.1 Run focused mapper, repository/count, guard, and all-companies tests; record commands/results in this file as apply-progress evidence.
- [x] 4.2 Run `go test ./...` and `go build ./...`; verify structural absence of old slice paths, `core/error.go`, raw repository errors, and a create-company slice.
- [x] 4.3 Reconcile `proposal.md`, all four delta specs, `design.md`, and this task plan; keep public-contract and rollback statements coherent.

## Apply Progress Evidence

| Work unit | Focused test result | Runtime harness result | Rollback boundary |
|---|---|---|---|
| 1 — errors and mappers | `go test ./internal/modules/companies/core ./internal/modules/companies/queries` — PASS | N/A: pure error and AST behavior | `core/errors.go`, `queries/map_errors.go`, related tests |
| 2 — slice rehome | `go test ./internal/modules/companies/...` — PASS | Existing httptest route/resolver coverage — PASS in companies suite | seven `slices/*` packages and container imports |
| 3 — persistence mapping | `go test ./internal/modules/companies/... -run 'Repository|Count'` — PASS | SQLite-backed repository/query cases — PASS | seven repository adapters and repository/query tests |
| 4 — boundary guard and final checks | `go test ./internal/modules/companies/...` — PASS | `go test ./... && go build ./...` — PASS | structural guard, fixtures, task evidence |

## Focused Remediation Evidence

- RED: `go test ./internal/modules/companies/queries -run '^TestCompanyRepositoryGuardFixtures/same_function_mapped_selector_then_raw_selector$' -count=1` — FAIL before the guard change: `problems=[], want problem=true`.
- GREEN: `go test ./internal/modules/companies/queries -run 'TestCompanyRepositoryGuardFixtures|TestCompanyRepositoriesUseEntityMappers' -count=1` — PASS.
- Regression coverage: `go test ./internal/modules/companies/... -count=1` — PASS; `go test ./...` — PASS; `go build ./...` — PASS; `git diff --check` — PASS.
- Inventory correction: 18 repository-method paths; 19 GORM operations (the `list.List` method performs both `Count` and `Find`).

### Authored-Line Budget

Baseline: current `HEAD`; count tracked diff plus untracked files, excluding only pre-existing planning artifacts: `openspec/changes/align-companies-module-conventions/proposal.md`, `openspec/changes/align-companies-module-conventions/specs/**`, `openspec/changes/align-companies-module-conventions/design.md`, and `openspec/changes/align-companies-module-conventions/tasks.md`.

```bash
{ git diff --numstat -- . ':(exclude)openspec/changes/align-companies-module-conventions/proposal.md' ':(exclude)openspec/changes/align-companies-module-conventions/specs/**' ':(exclude)openspec/changes/align-companies-module-conventions/design.md' ':(exclude)openspec/changes/align-companies-module-conventions/tasks.md'; git ls-files --others --exclude-standard | while IFS= read -r path; do case "$path" in openspec/changes/align-companies-module-conventions/proposal.md|openspec/changes/align-companies-module-conventions/specs/*|openspec/changes/align-companies-module-conventions/design.md|openspec/changes/align-companies-module-conventions/tasks.md) ;; *) git diff --no-index --numstat /dev/null "$path" || test $? -eq 1 ;; esac; done; } | awk '{add += $1; del += $2} END {printf "authored_lines=%d (additions=%d + deletions=%d)\n", add + del, add, del}'
```

Formula/result: additions + deletions = 614 + 122 = **736 authored lines**, within the accepted 800-line exception.
