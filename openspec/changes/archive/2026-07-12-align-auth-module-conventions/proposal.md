# Proposal: Align Auth Module Conventions

## Intent

Align `internal/modules/auth` with the current module conventions without changing observable auth behavior. The module is already stable; the work removes layout drift, centralizes GORM-to-domain error translation, and brings file naming in line with the documented pattern.

## Scope

### In Scope
- Move `authenticate_user`, `rotate_token`, `revoke_token`, and `view_session` to `internal/modules/auth/slices/` using real directory moves.
- Add `internal/modules/auth/queries/map_errors.go` and route all GORM not-found translation through it.
- Rename `internal/modules/auth/core/error.go` to `internal/modules/auth/core/errors.go`; update wiring, imports, and affected tests.

### Out of Scope
- No shims, aliases, or compatibility wrappers.
- No changes to routes, payloads, HTTP methods, or auth semantics.
- No expansion to other modules.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
None.

## Approach

Perform a mechanical rehome: move the four slice packages, update container/route imports, introduce a shared query error mapper, and rename the core error file. Keep the current 401/422 behavior intact by preserving existing domain error codes and response flow.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/auth/slices/*` | New/Modified | Auth slice packages move under `slices/`. |
| `internal/modules/auth/queries/map_errors.go` | New | Central GORM→domain error translation. |
| `internal/modules/auth/core/errors.go` | Renamed | Canonical module error sentinel file. |
| `internal/modules/auth/container.go` | Modified | Update imports/wiring for moved slices. |
| `internal/modules/auth/routes.go`, tests | Modified | Fix paths/imports and keep behavior coverage. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Missed import/path rewrite breaks build | Med | Move packages mechanically, then run focused auth tests. |
| Error mapping changes observable 401/422 behavior | Med | Preserve existing domain sentinels and map only GORM not-found cases. |
| Rename noise obscures real changes | Low | Keep the change rename-only wherever possible. |

## Rollback Plan

Revert the directory moves, restore `core/error.go`, remove `queries/map_errors.go`, and re-inline the prior repository error translation. This returns auth to the previous layout with no user-facing change.

## Dependencies

- None.

## Success Criteria

- [ ] Auth routes, payloads, and 401/422 outcomes remain unchanged.
- [ ] Auth slices live under `internal/modules/auth/slices/` and builds/tests pass.
- [ ] GORM not-found translation is centralized in `queries/map_errors.go`.
