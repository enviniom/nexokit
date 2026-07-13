# Tasks: Align Auth Module Conventions

## Review Workload Forecast

| Field | Value |
|---|---|
| Original completed scope | 350–550 (renames detected as moves) |
| Previous completed scope | 440–700 |
| Corrective phase increment | 160–260 |
| Cumulative estimate | 1,150–1,400 |
| 1,400-line budget risk | Final checkpoint required |
| Chained PRs recommended | No |
| Suggested split | One authorized PR; corrective persistence boundary is work unit 4 |
| Delivery strategy | single-pr / maintainer-approved `size:exception` / 1,400-line budget |
| Chain strategy | size-exception |

Decision needed before apply: No — maintainer explicitly approved a single PR with `size:exception` and a 1,400 changed-line budget.
Chained PRs recommended: No
Chain strategy: size-exception
1,400-line budget risk: Final line-count checkpoint is required before verification.

### Suggested Work Units

| Unit | Goal | Likely PR | Focused test command | Runtime harness | Rollback boundary |
|---|---|---|---|---|---|
| 1 | Mapper and repositories | PR 1, commit 1 | `go test ./internal/modules/auth/...` | same command | Mapper and repositories |
| 2 | Rehome and wiring | PR 1, commit 2 | `go test ./internal/modules/auth/...` | `go test ./... && go build ./...` | Slices, container, error filename |
| 3 | Correct mapper ownership | PR 1, correction commit | `go test ./internal/modules/auth/queries ./internal/modules/auth/slices/{authenticate_user,rotate_token,revoke_token}` | `go test ./... && go build ./...` | Mapper, three repositories, structural guards |
| 4 | Enforce universal auth persistence boundary | PR 1, corrective commit | `go test ./internal/modules/auth/queries ./internal/modules/auth/slices/...` | `go test ./... && go build ./...` | Auth core errors, mappers, all auth repositories, guards, and aligned artifacts |

## Phase 1: Mapper Contract and Repository Migration

- [x] 1.1 RED: establish the original mapper baseline for nil, direct/wrapped `gorm.ErrRecordNotFound`, and unrelated errors in `queries/map_errors_test.go` (unknown-error behavior is superseded by Phase 5).
- [x] 1.2 GREEN: add `queries.MapNotFound(err, mapped error)`.
- [x] 1.3 RED: test missing user/role map to `core.ErrInvalidCredentials` in `authenticate_user/repository_test.go`.
- [x] 1.4 GREEN: replace inline GORM translation in `authenticate_user/repository.go` with `queries.MapNotFound`.
- [x] 1.5 RED/GREEN: route missing refresh tokens in rotate/revoke repository tests through the mapper to `core.ErrInvalidRefreshToken`.

## Phase 2: Atomic Slice Rehome and Wiring

- [x] 2.1 Move all auth slice production/tests under `internal/modules/auth/slices/*/`; retain packages and APIs.
- [x] 2.2 Update `internal/modules/auth/container.go` imports; preserve wiring.
- [x] 2.3 Rename `core/error.go` to `core/errors.go`; retain sentinel values, codes, and messages.
- [x] 2.4 Run auth tests and `TestNewContainer_WiresAllSlices` after the import rewrite.

## Phase 3: Behavior and Structural Verification

- [x] 3.1 Prove login 200/token-user, validation 422, and generic-401 behavior with auth route tests.
- [x] 3.2 Verify refresh/logout/me paths, payloads, statuses, middleware, and no shim route.
- [x] 3.3 Guard: only `queries/map_errors.go` maps not-found; old slices/imports and `core/error.go` are absent.
- [x] 3.4 Run `go test ./...` and `go build ./...`; record work-unit results.

## Phase 4: Authorized Mapper Ownership Correction

- [x] 4.1 RED: add the prior unary-mapper cases for nil and direct/wrapped `gorm.ErrRecordNotFound`; its unknown-error identity expectation is explicitly superseded by Phase 5.
- [x] 4.2 GREEN: replace/remove `MapNotFound(err, mapped)` in `internal/modules/auth/queries/map_errors.go`; implement unary `MapUserError(err)` and `MapRefreshTokenError(err)`.
- [x] 4.3 Migrate `slices/authenticate_user/repository.go`, `slices/rotate_token/repository.go`, and `slices/revoke_token/repository.go` to pass only `err` to their entity-specific mapper.
- [x] 4.4 Add `internal/modules/auth/queries/map_errors_structure_test.go` guards rejecting mapper signatures with caller-selected sentinels and repository calls to `MapNotFound` or any two-argument mapper.
- [x] 4.5 Run focused mapper/repository tests, then `go test ./...` and `go build ./...`; record results in `apply-progress.md`.
- [x] 4.6 Verify `docs/modules/queries-and-persistence.md`, `specs/auth/spec.md`, and `design.md` describe unary entity mappers and no remaining generic mapper.

## Phase 5: Universal Auth Persistence Boundary Correction

- [x] 5.1 Inventory every `internal/modules/auth/slices/**/repository.go` method, every auth GORM operation, every `.Error` path, and every domain-significant `RowsAffected` outcome; keep the design inventory exhaustive as files evolve.
- [x] 5.2 RED: extend mapper, core-error, and repository tests to assert `errors.As(..., *apperror.AppError)`, expected status/code, `errors.Is` cause preservation, and no raw GORM/SQL/driver error leak for reads, creates, updates, and revokes.
- [x] 5.3 GREEN: add module-owned internal user and refresh-token persistence codes/constructors in `core/errors.go`, preserving unknown causes through `Internal`/`Unwrap()` while repository interfaces remain typed as `error` and do not import `apperror`.
- [x] 5.4 GREEN: update `MapUserError` and `MapRefreshTokenError` so nil stays nil, known outcomes map specifically, and every unknown persistence failure maps to the correct module-owned internal AppError.
- [x] 5.5 Migrate every audited repository call site, including both `CreateRefreshToken` writes and both revoke updates, so every GORM `.Error` passes through the entity mapper; map zero-row revokes to invalid refresh token.
- [x] 5.6 Replace the narrow structural checks with a generic guard that recursively scans all auth repository files/interfaces/methods and rejects concrete `apperror` exposure, direct raw `.Error` returns, and unmapped persistence paths without hard-coded method names. Completed after focused phase-contract remediation.
- [x] 5.7 Run focused mapper/core/repository tests, then `go test ./internal/modules/auth/...`, `go test ./...`, and `go build ./...`; record exact results and confirm the single PR remains at or below the authorized 1,400 authored changed-line budget.
- [x] 5.8 Reconcile `docs/modules.md`, both module guides, auth delta spec, design inventory, tasks, and apply evidence so the universal boundary, zero-row semantics, test contract, one-PR delivery, and forecast agree.
