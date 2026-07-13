```yaml
schema: gentle-ai.verify-result/v1
evidence_revision: sha256:eacd9d462c6eac7128add3a969c77305ff0a87a6945ef15b95c0000d697e368e
verdict: pass
blockers: 0
critical_findings: 0
requirements: 4/4
scenarios: 12/12
test_command: "go test ./internal/modules/companies/queries && go test ./internal/modules/companies && go test ./internal/modules/companies/... && go test ./..."
test_exit_code: 0
test_output_hash: sha256:130df5c0052e80c7f50ba1a5fa233c0cfb7dfb7468332bbc32c525b7e0c2ee93
build_command: "go build ./..."
build_exit_code: 0
build_output_hash: sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
```

# Verification Report

**Change**: Align Companies Module Conventions
**Mode**: Standard
**Tasks**: 15/15 complete

## Completeness

| Metric | Value |
|---|---:|
| Requirements | 4/4 |
| Scenarios | 12/12 |
| Tasks | 15/15 |

## Build & Tests Execution

| Command | Exit | Output hash |
|---|---:|---|
| `go test ./internal/modules/companies/queries` | 0 | `sha256:f2640f9b4e523d540dcb15d5b02f60ccf66d03c46ce97837ab1ee809d92a73bb` |
| `go test ./internal/modules/companies` | 0 | `sha256:5f33fff7a3a94768635f7c9c5f3d0e455d03324d169f7129faeb361d3a15220b` |
| `go test ./internal/modules/companies/...` | 0 | `sha256:aa4ad47404ea2b5a9b77321c6bd74c8827e9f238a09f1a25682bd9be5c5c4cec` |
| `go test ./...` | 0 | `sha256:48d20b2871c923a1b4c61dac24cc16162873814cbd258e1a5e44d74256b2f2d7` |
| `go build ./...` | 0 | `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `git diff --cached --check` | 0 | `sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| authored budget check | 0 | `authored_lines=739 (additions=646 + deletions=93)` |

## Spec Compliance Matrix

| Requirement | Scenario | Runtime evidence | Result |
|---|---|---|---|
| Vertical slice modules | Companies is rehomed | `go test ./internal/modules/companies/...`; `git status --short` shows only `slices/*` renames | COMPLIANT |
| Vertical slice modules | Other modules stay unchanged | `go test ./...` passed; diff scope stayed within companies + openspec artifacts | COMPLIANT |
| Companies CRUD | CRUD responses stay stable | `internal/modules/companies/routes_absence_test.go`, handler tests, `go test ./internal/modules/companies/...` | COMPLIANT |
| Companies CRUD | Direct create remains absent | `internal/modules/companies/routes_absence_test.go` | COMPLIANT |
| Companies CRUD | Compatibility aliases remain usable | `core/model.go`, `core/dto.go`, `resolver_test.go`, `go test ./...`, `go build ./...` | COMPLIANT |
| Company domains | Create flow ignores exclude | `internal/modules/companies/queries/count_active_primary_domains_test.go`, `slices/create_company_domain/service_test.go` | COMPLIANT |
| Company domains | Update flow honors exclude | `internal/modules/companies/queries/count_active_primary_domains_test.go`, `slices/update_company_domain/service_test.go` | COMPLIANT |
| Company domains | Zero rows map to not-found | `slices/delete_company/repository_test.go`, `slices/update_company/repository_test.go`, `slices/update_company_domain/repository_test.go` | COMPLIANT |
| Company domains | Guards discover every repository recursively | `internal/modules/companies/queries/map_errors_structure_test.go` | COMPLIANT |
| Error handling | Unknown failure preserves cause | `internal/modules/companies/queries/map_errors_test.go`, `core/errors_test.go` | COMPLIANT |
| Error handling | Known persistence outcomes stay typed | `internal/modules/companies/queries/map_errors_test.go` | COMPLIANT |
| Error handling | Canonical error file exists | `core/errors.go` exists; `core/error.go` removed via rename | COMPLIANT |

**Compliance summary**: 12/12 scenarios compliant

## Correctness

| Requirement | Status | Notes |
|---|---|---|
| Slice-root rehome | Implemented | Seven companies slices are under `internal/modules/companies/slices/` and root stays wiring-only. |
| Public HTTP contract | Implemented | Routes, `:id`, aliases, resolver, and absent create route preserved. |
| Persistence error boundary | Implemented | Company repositories map to module-owned `*apperror.AppError` values. |
| Domain count/guarding | Implemented | `CountActivePrimaryDomains` flow split and recursive guard fixtures pass. |

## Coherence

| Decision | Followed? | Notes |
|---|---|---|
| Real moves; no shims | Yes | Packages were rehomed under `slices/`. |
| Unary entity mappers | Yes | `MapCompanyError` and `MapCompanyDomainError` are unary and tested. |
| Function-scoped recursive guard | Yes | Same-function raw/mapped collision fixture passes. |
| Single PR, ≤800 authored lines | Yes | Current staged diff measures 739 authored lines. |

## Issues Found

**CRITICAL**: None

**WARNING**: Budget evidence drift: tasks.md says 736 authored lines, apply-progress.md says 797, and the current staged diff measures 739. All are still within the 800-line exception, but the recorded figures are not identical.

**SUGGESTION**: None

## Verdict

PASS
All 4 requirements and 12 scenarios are covered by passing runtime evidence; only the authored-line evidence needs reconciliation.
