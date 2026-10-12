# Integrate authentication into IAM

## Objective and final scope
IAM owns authentication without changing HTTP routes, credential checks, session sanitization, token rotation/revocation, middleware or rate limits. Authentication models live in `internal/modules/iam/core/auth` (package `user`) and `core/refresh_token`; shared contracts/errors and queries belong directly to IAM. Auth use cases remain under `internal/modules/iam/slices/auth`.

## Decisions and constraints
- Initial whole-module relocation under slices was rejected by the user; the user supplied the integrated IAM core/query structure. That structure was preserved and obsolete imports/aliases repaired.
- Preserve current token algorithms, DTOs, schema and public `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me` contracts.
- Do not restore `core/errors_test.go`: the user intentionally removed sentinel-constant tests in favor of handler-boundary contracts. Audit found all 15 IAM sentinels retain handler HTTP-status assertions; public messages and field-validation remapping are covered by handlers and the central renderer. Application codes are not serialized; internal code uniqueness is an optional uncovered invariant, not a delivery blocker.
- Narrow the relocated mapper guard to the four auth repositories, preserving generic discovery and exposure fixtures. Fail if no auth repository validates.
- Correct active documentation to match IAM's integrated core/query/auth-slice structure.
- User authorized commits and local main merge after passing checks; no push or remote publication.

## Completed work units
- [x] T1: Initial relocation and IAM/app ownership wiring with regression tests; superseded structure reconciled with the user's integration.
- [x] T2: Import repair against the user's tree and authorized auth-only structural guard correction; verified with focused, integration and full suites.
- [x] T3: Final coverage audit, documentation corrections, tests/lint, coherent integration commit and local fast-forward main merge.

T1/T2 outcomes and their tests/docs were delivered as one coherent integration work unit; a split would break imports/wiring. The roughly 400-line planning heuristic was advisory, not a reason to omit tests or artificially split relocation. Commit: `17a456a074acd6e1695f2a35e005ad177e05af9c` — `refactor(iam): integrate authentication into identity module` (55 files, 198 additions, 320 deletions with rename detection).

## Verification evidence
- Observed RED/GREEN for `TestRegisterOwnsAuthenticationRoutes`; asserts ownership and public versus protected/rate-limit middleware boundaries.
- Import repair: existing suite initially failed compilation; corrected imports compile repository-wide.
- Mapper guard: observed RED on 24 non-auth repositories after test relocation; GREEN after auth-only scope correction. All four auth repositories remain validated; exposure fixtures remain effective.
- Final independent `go test ./... -count=1`: PASS, including docs/helpers/integration.
- Final independent `make lint`: PASS (`go vet ./...` and module error guard).
- Changed/new Go formatting and `git diff --check`: PASS.
- Independent `go test ./tests/docs/... -count=1`: PASS; parent docs-test spot check: PASS.
- Updated active docs contain no obsolete `slices/auth/core`, `slices/auth/slices` or `internal/modules/auth` references.
- Native assessment was high/unassessable across intermediate candidates; independent verifier used throughout. Native RDD review skipped because the user-owned switch remains off.
- No failing required checks remain. No separate browser/server run: this is an API-preserving ownership/import migration with deterministic integration coverage.

## Delivery evidence
- Base: `e578f01f41530a5319e28d87f48ca78db4d8064b`.
- Integration commit created on `refactor/integrate-auth-into-iam`; fast-forward merge into local `main` observed at `17a456a074acd6e1695f2a35e005ad177e05af9c`.
- Staged state changed externally during earlier verification rounds; no rollback was attempted, and final delivery staged only the verified feature surfaces.
- Rollback boundary: revert the integration commit to restore independent auth ownership and the original tests/docs without unrelated changes.
- No push, PR, release, destructive reset or branch deletion.

## Next step
Commit this evidence checkpoint on the feature branch, fast-forward main to include it, and confirm clean final state. Remote publication remains the user's decision.
