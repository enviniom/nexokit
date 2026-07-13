# Apply Progress: Align Auth Module Conventions

## Status

All 27 tasks are complete. Phases 1–3 were standard-mode work; Phases 4–5 used Strict TDD. This merges earlier work and the focused 5.6 remediation.

## Completed Tasks

- [x] 1.1–1.5 Mapper contract and repository migration
- [x] 2.1–2.4 Atomic slice rehome and wiring
- [x] 3.1–3.4 Behavior and structural verification
- [x] 4.1–4.6 Authorized mapper ownership correction
- [x] 5.1–5.8 Universal auth persistence boundary correction

## Phase 5 Inventory

| Repository / method | Persistence boundary | Result |
|---|---|---|
| `authenticate_user.GetByEmail` | user/role reads | User mapper |
| `authenticate_user.CreateRefreshToken` | refresh-token create | Refresh-token mapper |
| `rotate_token.GetByHash` / `CreateRefreshToken` / `Revoke` | read/create/update | Mapper; zero-row revoke is invalid token |
| `revoke_token.GetByHash` / `Revoke` | read/update | Mapper; zero-row revoke is invalid token |
| `view_session.BuildSession` | none | In-memory only; no persistence |

## Work Unit Evidence

| Work unit | Focused test command and result | Runtime harness/scenario and result | Rollback boundary |
|---|---|---|---|
| 1–3 prior work | `go test ./internal/modules/auth/...` — PASS (7 auth packages) | `go test ./... && go build ./...` — PASS | Slice moves, container wiring, mapper and core filename. |
| 4 mapper ownership | `go test ./internal/modules/auth/queries ./internal/modules/auth/slices/{authenticate_user,rotate_token,revoke_token}` — PASS (4 packages) | `go test ./... && go build ./...` — PASS | Unary mapper, three repositories, structural checks. |
| 5.6 remediation | `go test ./internal/modules/auth/queries -run 'Test(RepositoriesUseUnaryEntityMappers|AuthRepositoryFilesDiscoverNestedRepositories|RepositoryBoundaryGuardRejectsRawExposure|RepositoryBoundaryGuardRejectsNestedRawErrorVariable)' -count=1` — PASS | `go test ./internal/modules/auth/... && go test ./... && go build ./...` — PASS | `queries/map_errors_structure_test.go`; recursive AST guard only, no production behavior. |

## TDD Cycle Evidence — Phase 5

| Task | Test File | Layer | Safety Net | RED | GREEN | TRIANGULATE | REFACTOR |
|---|---|---|---|
| 5.1 | `design.md` | Structural | N/A | Inventory inspection | Confirmed | Four repository scopes | None needed |
| 5.2 | mapper/core tests | Unit | ✅ focused baseline | Constructors undefined | PASS | Known/unknown cases | Table-driven |
| 5.3 | mapper/core tests | Unit | ✅ focused baseline | Constructor RED | PASS | Both entities | Constructors |
| 5.4 | mapper tests | Unit | ✅ focused baseline | Unknown-cause RED | PASS | Nil/known/unknown | Nil-first |
| 5.5 | repository tests | Unit | ✅ focused baseline | Raw closed-DB/zero-row | PASS | Reads/writes/revokes | Result variables |
| 5.6 | `map_errors_structure_test.go` | Structural | ✅ queries PASS | FAIL: raw variable accepted | PASS: 4 guard tests | Nested discovery + raw variable | Recursive AST/data flow |
| 5.7 | auth/full/build | Verification | N/A | N/A | PASS | N/A | None needed |
| 5.8 | OpenSpec artifacts | Documentation | N/A | Inspection | Confirmed | N/A | None needed |

## Verification

- Focused remediation command above — PASS (4 structural tests).
- `go test ./internal/modules/auth/...` — PASS (7 packages).
- `go test ./...` — PASS.
- `go build ./...` — PASS.
- `git diff --check` — PASS.

## Delivery

- Mode: single PR, maintainer-approved `size:exception`; no commit created.
- Forecast/decision: final 1,400 changed-line budget and `size:exception` approved by maintainer.
- Current cumulative authored diff: 1,393 changed lines (within 1,400 after remediation).
- Deviations: None. `view_session` has no persistence and rotate-token transaction atomicity remains explicitly out of scope.
