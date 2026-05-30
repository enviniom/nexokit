# Proposal: Migrate Permissions Module to Vertical Slice Architecture

## Intent

Refactor `internal/modules/permissions/` from flat layout to vertical slice. Partial work from change-16 (13 untracked files) provides the foundation — must be completed, not discarded.

## Scope

### In Scope
- Complete 3 existing partial slices: `list_permissions`, `view_permission`, `update_permission`
- Create 2 internal non-HTTP slices: `resolve_permissions` (auth middleware), `sync_permissions` (bootstrap)
- Create `container.go` (composition root), update `routes.go`
- Extract shared queries to `queries/`
- Create `core/contracts.go` for `PermissionCatalogReader` interface
- Delete root flat files: `handler.go`, `service.go`, `repository.go`, `model.go`, `dto.go`
- Update `app/container.go` wiring
- Add tests to all slices

### Out of Scope
- Unified IAM module (users + roles + permissions) — separate future change
- Migrating roles or users modules
- Resolving `roles → permissions` direct import
- Registering unregistered handlers (`Create`, `Delete`, `ListPaginated`)
- Any API or behavioral changes

## Capabilities

### New Capabilities
None — structural refactor only.

### Modified Capabilities
None — `openspec/specs/permissions/spec.md` requirements unchanged.

## Approach

Full migration in one change. Existing change-16 partial files cover 3 of 5 slices. Remaining: container, route wiring, 2 internal slices, query extraction, tests, cleanup, app wiring update.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/permissions/` | Restructured | Flat → vertical slice (container, core/, queries/, 5 slices) |
| `internal/app/container.go` | Modified | Wiring: flat injection → module container |
| `internal/modules/roles/service.go` | Unchanged | Continues importing permissions types |
| `internal/middleware/auth.go` | Unchanged | `PermissionResolver` contract preserved |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Auth flow breakage (Resolve/SyncPermissions) | Medium | Preserve exact signatures; test middleware integration |
| Duplicate types during migration | Low | Delete root files only after slices reference core/ + tests pass |
| Package name inconsistency (`view_permissions` vs `view_permission`) | Low | Fix to singular before building |
| Review load (~600+ lines) | Medium | Tasks split for chained PRs; guard at tasks phase |

## Rollback Plan

`git revert` the merge commit. All flat files preserved in git. No DB or API changes.

## Dependencies

- Existing untracked partial slice files from change-16
- `go test ./...` must pass after each task unit

## Success Criteria

- [ ] All 3 registered endpoints respond identically (status codes, bodies, errors)
- [ ] `Resolve()` works in auth middleware flow
- [ ] `SyncPermissions()` runs at bootstrap without errors
- [ ] `go test ./internal/modules/permissions/...` passes
- [ ] `go test ./...` passes
- [ ] No root flat files remain
- [ ] All slices have handler/service/repository + test files
