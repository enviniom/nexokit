# Proposal: Migrate Auth Module to Vertical Slice Architecture

## Intent

Refactor `internal/modules/auth/` from flat legacy structure to vertical slices, eliminating cross-module dependencies on `users` and `roles`. Aligns with `vertical-slice-modules` spec; pure structural refactor — zero behavior change.

## Scope

### In Scope
- 4 slices: `authenticate_user`, `rotate_token`, `revoke_token`, `view_session`
- `core/` package: local partial models (`AuthUser`), DTOs, module errors
- `queries/` package: `find_user_by_email`, `find_user_by_id_with_role` with tests
- `container.go` composition root; updated `routes.go`
- App container wiring: `auth.NewContainer(db)` replaces flat construction
- Migrate all 13 test cases + 1 integration flow

### Out of Scope
- Any other module migration
- `internal/middleware/userLookup` adapter (stays at app level)
- Behavioral changes or new features

## Capabilities

### New Capabilities
- None — pure structural refactor.

### Modified Capabilities
- `vertical-slice-modules`: Auth becomes second migration target after companies.
- `auth`: Structural delta only; all requirements unchanged.

## Approach

Full vertical slice with `queries/` (matches companies module pattern):
1. Foundation: `core/` with local partial models eliminating cross-module imports
2. Queries: shared DB lookups in `queries/`
3. Slices: build in order — authenticate_user → rotate_token + revoke_token → view_session
4. Wiring: container.go, routes.go, app container update
5. Cleanup: remove old flat files, verify full test suite

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/auth/` | Modified | Flat → 4 slices + core/ + queries/ + container.go |
| `internal/app/container.go` | Modified | `auth.NewContainer(db)` replaces flat construction |
| `server/router.go` or `bootstrap.go` | Modified | Route registration signature change |
| `openspec/specs/vertical-slice-modules/spec.md` | Modified | Auth added as migration target |
| `openspec/specs/auth/spec.md` | Modified | Structural delta (behavior unchanged) |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Test migration: repo test uses real `users.Repository` + `roles.Role` | High | Replace with local partial model AutoMigrate; split per slice |
| Exceeds 400-line review budget | High | Chained PRs (strategy: ask-always, budget: 800 lines) |
| `Me` handler reads from authctx only (no service/repo) | Low | `view_session` has minimal/pass-through service |

## Rollback Plan

`git revert` the merge commit — all old flat files in git history. No database migrations introduced; zero data risk.

## Dependencies

- Companies module migration (complete — reference pattern)
- Chained PRs required (will exceed 400 lines)

## Success Criteria

- [ ] `go test ./internal/modules/auth/...` passes (13+ cases)
- [ ] `go test ./...` passes — zero behavior change
- [ ] Zero `modules/users` or `modules/roles` imports in auth source
- [ ] All 4 endpoints respond identically
- [ ] App container uses `auth.NewContainer(db)`
