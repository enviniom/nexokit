## Exploration: align-auth-module-conventions

### Current State
`internal/modules/auth` is functionally complete but still uses the pre-`docs/modules.md` layout: slice packages live directly under the module root (`authenticate_user/`, `rotate_token/`, `revoke_token/`, `view_session/`) instead of under `slices/`, and the module has `queries/` without the mandatory `queries/map_errors.go` central mapper.

The observable auth behavior is already stable: `/auth/login`, `/auth/refresh`, `/auth/logout`, and `/auth/me` keep the same routes, payloads, and 401/422 behavior. The current repositories inline the persistence-to-domain translation with `gorm.ErrRecordNotFound → core.ErrInvalidCredentials` / `core.ErrInvalidRefreshToken`.

### Affected Areas
- `internal/modules/auth/container.go` — imports and wires the four slice packages; import paths will change if slices move under `slices/`.
- `internal/app/container.go` — top-level app wiring imports the auth module container; likely unchanged behavior, but must be verified after path moves.
- `internal/modules/auth/core/error.go` — docs expect `core/errors.go`; current sentinel names are fine, file naming is not.
- `internal/modules/auth/queries/*.go` — existing reusable queries stay, but error translation must move into `queries/map_errors.go`.
- `internal/modules/auth/authenticate_user/*`, `rotate_token/*`, `revoke_token/*`, `view_session/*` — these packages should move under `slices/`.
- `internal/modules/auth/*_test.go` and `tests/integration/auth_test.go` — test imports and package paths need update/verification after the move.

### Approaches
1. **Mechanical rehome + centralized error mapping** — move the four slice directories under `slices/`, rename `core/error.go` to `core/errors.go`, add `queries/map_errors.go`, and replace inline not-found translation with shared helpers.
   - Pros: matches docs exactly, preserves runtime behavior, keeps the change reviewable as mostly filesystem moves.
   - Cons: touches many paths/imports; needs careful rename handling to preserve history.
   - Effort: Medium

2. **Phased alignment with temporary shims** — add new `slices/` and `queries/map_errors.go` while keeping old packages as compatibility wrappers for one cycle.
   - Pros: lower immediate import churn.
   - Cons: duplicates structure, violates the target shape longer, adds indirection without behavior value.
   - Effort: Medium

### Recommendation
Choose the mechanical rehome. Auth is already behavior-stable, so the right move is to align structure and persistence error mapping without changing public HTTP behavior. The current module is small enough that a direct rename/move is the cleanest path.

### Risks
- A missed import path rewrite could break app wiring or tests, especially after moving slice packages under `slices/`.
- `queries/map_errors.go` must preserve the current 401 semantics exactly; a wrong mapping would be an observable behavior regression.
- The repo-wide rename set is noisy even though most of it is mechanical; use actual moves, not copy/paste.
- Preliminary diff size should stay well under the 800-line review budget; estimate ~120-180 authored line changes, with most file moves being rename-only.

### Ready for Proposal
Yes — the scope is clear enough for `sdd-propose` once you want the change plan. The proposal should keep behavior unchanged and treat only the layout/error-mapping work as mandatory.
