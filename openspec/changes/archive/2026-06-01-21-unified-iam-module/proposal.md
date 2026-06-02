# Proposal: Unified IAM Module

## Intent

Eliminate cross-module dependency chains (`users` → `roles` → `permissions` → `companies`) by consolidating into a single `internal/modules/iam/` module. Legacy modules preserved for review.

## Scope

### In Scope
- Create `internal/modules/iam/` with multi-entity vertical slice structure
- Wire IAM in `internal/app/container.go`, replacing three separate module refs
- Update `RegisterModules` to mount IAM routes; remove legacy mounts
- Update middleware adapters to delegate to IAM
- Preserve all public routes, payloads, status codes, authz contracts

### Out of Scope
- Delete legacy modules (deferred to `22-legacy-iam-cleanup`)
- Rename routes or change API behavior
- Modify migrations
- Refactor logic during copy

## Capabilities

### New Capabilities
- `iam-module`: Unified module covering user CRUD, role CRUD, permission CRUD, auth resolution, permission resolution (cache-backed), permission sync, role-permission assignment. Eliminates cross-module imports via partial local models.

### Modified Capabilities
- `app-orchestration`: Container replaces `usersHandler`, `rolesHandler`, `permissionsContainer` with `IAM *iam.Container`. `RegisterModules` mounts IAM only.
- `rbac-authorization`: `AuthUserLookup` and `PermissionResolver` adapters delegate to IAM. Interface contracts unchanged.

## Approach

Multi-entity vertical slice (per `_context.md` §8). Entity sub-folders: `users/`, `roles/`, `permissions/`, each with own container/routes/slices. `internal/` for non-HTTP slices (`resolve_user_permissions`, `sync_permissions`, `resolve_auth_user`, `resolve_role_by_slug`, `list_all_permissions`). Partial GORM models eliminate cross-module imports.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/iam/` | New | Full module: container, routes, core/, queries/, entity sub-folders, internal/ |
| `internal/app/container.go` | Modified | Replace 3 refs with `IAM`; update adapters |
| `internal/server/routes.go` | Modified | Mount IAM, remove legacy |
| `internal/middleware/auth.go` | Modified | Delegate to IAM |
| `internal/middleware/authorization.go` | Modified | Delegate to IAM |
| Legacy modules | Unchanged | Preserved, unreachable at runtime |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Partial model field mismatch | Medium | Test with real DB queries |
| Double route registration | Low | Mount IAM only |
| Test coverage gap | Medium | Per-slice tests required |
| Scope creep | Low | Reproduce behavior exactly |

## Rollback Plan

Revert `app/container.go` to wire legacy modules. IAM code remains unused — zero data loss. No migrations to rollback.

## Dependencies

- Exploration complete
- No new external libraries

## Success Criteria

- [ ] `go test ./...` passes
- [ ] All 19 endpoints respond identically to legacy
- [ ] Middleware auth/authz work with IAM adapters
- [ ] Legacy modules compile but not wired
- [ ] Zero cross-module imports from IAM
- [ ] `go build ./...` succeeds
