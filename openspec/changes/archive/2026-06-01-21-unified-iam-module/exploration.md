# Exploration: Unified IAM Module (21-unified-iam-module)

## Current State

The codebase has three separate modules for IAM concerns:

1. **`internal/modules/users/`** — Flat module (not vertical slice). Handles user CRUD, password change, status toggle, and role change. Imports `roles` module directly for `Role` model and `RoleResolver` interface. Has its own repository, service, handler, and DTOs.

2. **`internal/modules/roles/`** — Flat module with CRUD, permission catalog, permission assignment, and role selection. Imports `permissions` module directly for `Permission` model and `companies` module for `Company` model. Has `AssignmentRoleReader` contract for cross-module role lookup.

3. **`internal/modules/permissions/`** — Already vertical-slice organized with `container.go`, slices (`list_permissions`, `view_permission`, `update_permission`, `resolve_permissions`, `sync_permissions`), `queries/`, and `core/`. Exports `Resolver`, `Syncer`, and `Catalog` contracts.

**Cross-module dependencies today:**
- `users` → `roles` (imports `roles.Role`, `roles.Repository` via `RoleResolver`, `roles.AssignmentRoleReader`)
- `roles` → `permissions` (imports `permissions.Permission` model)
- `roles` → `companies` (imports `companies.Company` model)
- `roles` → `users` (indirectly via `roleMemberRepository` interface, injected from `users.Repository`)
- `app/container.go` → all three modules, wires them together with adapters

**Auth/Authz flow:**
- `middleware.Auth` uses `userLookup` (wraps `users.Repository.GetAuthUser`) → returns `*authctx.User`
- `middleware.AttachPermissions` uses `permissionsContainer.Resolver` → resolves permission slugs
- `middleware.RequirePermission` uses `authctx.User.Permissions` for authorization
- `platform/permissions` registry tracks registered permission slugs at route-mount time

## Affected Areas

- `internal/modules/iam/` — **NEW** module to create
- `internal/app/container.go` — Replace `usersHandler`, `rolesHandler`, `permissionsContainer` with `iamContainer`; update `RegisterModules` to use IAM; update adapters (`userLookup`, `roleResolverAdapter`, `SyncPermissions`)
- `internal/middleware/auth.go` — `AuthUserLookup` contract stays compatible; implementation changes from `users.Repository` to IAM's equivalent
- `internal/middleware/authorization.go` — `PermissionResolver` contract stays compatible; implementation changes from `permissions.Resolver` to IAM's equivalent
- `internal/platform/authctx/authctx.go` — No changes; `User` struct is the shared contract
- `internal/platform/permissions/` — Registry and constants stay; IAM will use the same `Format()`, `Action*` constants
- `internal/modules/users/` — **LEGACY** preserved, no changes
- `internal/modules/roles/` — **LEGACY** preserved, no changes
- `internal/modules/permissions/` — **LEGACY** preserved, no changes

## Endpoint Mapping (current → IAM slices)

| Current Route | Current Module | IAM Slice | Notes |
|---|---|---|---|
| `GET /api/v1/users` | users.Handler.List | `list_users` | Tenant-scoped |
| `POST /api/v1/users` | users.Handler.Create | `create_user` | Root-scope logic |
| `GET /api/v1/users/:id` | users.Handler.GetByPublicID | `view_user` | Tenant-scoped |
| `PUT /api/v1/users/:id` | users.Handler.Update | `update_user` | Root protection |
| `DELETE /api/v1/users/:id` | users.Handler.Delete | `delete_user` | Soft-delete |
| `PATCH /api/v1/users/:id/password` | users.Handler.ChangePassword | `change_user_password` | Self-change rule |
| `PATCH /api/v1/users/:id/role` | users.Handler.ChangeRole | `assign_role_to_user` | Uses AssignmentRoleReader |
| `PATCH /api/v1/users/:id/status` | users.Handler.ToggleStatus | `toggle_user_status` | |
| `GET /api/v1/roles` | roles.Handler.List | `list_roles` | |
| `GET /api/v1/roles/select` | roles.Handler.ListSelect | `list_selectable_roles` | |
| `GET /api/v1/roles/:id` | roles.Handler.GetByPublicID | `view_role` | |
| `POST /api/v1/roles` | roles.Handler.Create | `create_role` | Reserved slug check |
| `PUT /api/v1/roles/:id` | roles.Handler.Update | `update_role` | System role protection |
| `DELETE /api/v1/roles/:id` | roles.Handler.Delete | `delete_role` | Assigned-user guard |
| `GET /api/v1/roles/:id/permissions` | roles.Handler.GetPermissionCatalog | `view_role_permission_catalog` | |
| `PUT /api/v1/roles/:id/permissions` | roles.Handler.AssignPermissions | `assign_permissions_to_role` | Cache invalidation |
| `GET /api/v1/permissions` | permissions.ListHandler.List | `list_permissions` | Grouped by module |
| `GET /api/v1/permissions/:id` | permissions.ViewHandler.GetByPublicID | `view_permission` | |
| `PUT /api/v1/permissions/:id` | permissions.UpdateHandler.Update | `update_permission` | System immutable |

## Internal Methods → IAM Internal Slices

| Current Internal Method | Current Location | IAM Internal Slice | Notes |
|---|---|---|---|
| `Resolve(publicID)` | `resolve_permissions.Service` | `resolve_user_permissions` | Cache-backed, used by middleware |
| `SyncPermissions(slugs)` | `sync_permissions.Service` | `sync_permissions` | Bootstrap-time, auto-assign to admins |
| `GetAuthUser(publicID)` | `users.Repository` | `resolve_auth_user` | Used by auth middleware |
| `GetBySlug(slug)` | `roles.Repository` | `resolve_role_by_slug` | Used by roleResolverAdapter |
| `ListAll()` | `list_permissions.Repository` | `list_all_permissions` | Used as PermissionCatalogReader |

## Cross-Module Dependencies That Disappear Inside IAM

Today these cross-module dependencies exist:
1. **`users` → `roles`**: `users/model.go` imports `roles.Role`; `users/service.go` imports `roles.AssignmentRoleReader`, `roles.RootRoleSlug`
2. **`roles` → `permissions`**: `roles/model.go` imports `permissions.Permission`; `roles/service.go` imports `permissions/core.Permission`
3. **`roles` → `companies`**: `roles/model.go` imports `companies.Company`
4. **`roles` → `users`**: `app/container.go` injects `usersRepo` as `roleMemberRepository` into roles service

Inside IAM, all of these become **internal** — the IAM module defines its own partial models (`IAMUser`, `IAMRole`, `IAMPermission`, `IAMCompany`) with only the fields needed, eliminating all cross-module imports.

## Approaches

### Approach 1: Multi-entity vertical slice with entity sub-folders (RECOMMENDED)

Create `internal/modules/iam/` with entity sub-folders following the multi-entity pattern from `_context.md` Section 8:

```
internal/modules/iam/
  container.go
  routes.go
  core/
    model.go          # Partial models: IAMUser, IAMRole, IAMPermission, IAMCompany
    dto.go            # All DTOs for all three entities
    error.go          # IAM domain errors
    constants.go      # ModuleIAM, reserved slugs
    contracts.go      # Resolver, Syncer, AuthUserLookup interfaces
  queries/            # Shared queries across entities
  users/
    container.go
    routes.go
    list_users/
    view_user/
    create_user/
    update_user/
    delete_user/
    change_user_password/
    toggle_user_status/
  roles/
    container.go
    routes.go
    list_roles/
    view_role/
    create_role/
    update_role/
    delete_role/
    list_selectable_roles/
    view_role_permission_catalog/
    assign_permissions_to_role/
  permissions/
    container.go
    routes.go
    list_permissions/
    view_permission/
    update_permission/
  internal/
    resolve_user_permissions/
    sync_permissions/
    resolve_auth_user/
    resolve_role_by_slug/
    list_all_permissions/
```

- **Pros**: Follows existing multi-entity convention; clear boundaries; each entity has its own container/routes; easy to navigate; internal slices clearly separated from HTTP-facing ones
- **Cons**: More files; deeper nesting; requires three entity sub-containers
- **Effort**: High

### Approach 2: Flat vertical slice (all slices at top level)

All 19 slices live directly under `internal/modules/iam/` without entity sub-folders:

```
internal/modules/iam/
  container.go
  routes.go
  core/
  queries/
  list_users/
  view_user/
  create_user/
  ...
  resolve_user_permissions/
  sync_permissions/
  ...
```

- **Pros**: Simpler structure; fewer containers; flatter navigation
- **Cons**: 19 slices in one directory is hard to scan; violates multi-entity convention when entities have >3 use cases each
- **Effort**: Medium

### Approach 3: Single container with grouped slices (hybrid)

One container, one routes file, but slices organized by entity sub-folders without sub-containers:

```
internal/modules/iam/
  container.go          # Single container wires ALL slices
  routes.go             # Single routes file registers ALL handlers
  core/
  queries/
  users/
    list_users/
    view_user/
    ...
  roles/
    list_roles/
    ...
  permissions/
    list_permissions/
    ...
  internal/
    resolve_user_permissions/
    ...
```

- **Pros**: Single wiring point; no sub-containers; entity grouping for navigation
- **Cons**: Container grows large; routes file grows large; deviates from multi-entity pattern
- **Effort**: Medium

## Recommendation

**Approach 1** (multi-entity with sub-containers) is the recommended approach because:

1. It follows the established convention in `_context.md` Section 8 for multi-entity modules
2. Each entity (users, roles, permissions) has more than 3 use cases, triggering the sub-folder rule
3. Internal slices (`resolve_user_permissions`, `sync_permissions`, etc.) don't need HTTP handlers and can live in an `internal/` sub-folder
4. The root container explicitly exports `Users`, `Roles`, `Permissions` sub-containers — clean DI surface
5. The root `routes.go` delegates to entity `routes.go` files — clean HTTP surface

The `internal/` sub-folder is a new convention needed here: slices that serve internal contracts (middleware, bootstrap) rather than HTTP endpoints. These still follow the slice pattern (service + repository) but have no handler.

## Container Wiring Change

**Current `internal/app/container.go`:**
```go
type Container struct {
    rolesHandler         *roles.Handler
    usersHandler         *users.Handler
    Companies            *companies.Container
    permissionsContainer *permissions.Container
    authContainer        *auth.Container
    Onboarding           *onboarding.Container
    authMW               gin.HandlerFunc
    authzMW              gin.HandlerFunc
    ...
}
```

**Target `internal/app/container.go`:**
```go
type Container struct {
    IAM                  *iam.Container        // NEW: replaces usersHandler, rolesHandler, permissionsContainer
    Companies            *companies.Container
    authContainer        *auth.Container
    Onboarding           *onboarding.Container
    authMW               gin.HandlerFunc
    authzMW              gin.HandlerFunc
    ...
}
```

The `RegisterModules` method changes from:
```go
roles.Register(globalProtected, c.rolesHandler, middleware.RequirePermission)
permissions.Register(globalProtected, c.permissionsContainer, middleware.RequirePermission)
users.Register(tenantProtected, c.usersHandler, middleware.RequirePermission)
```

To:
```go
iam.Register(globalProtected, c.IAM, tenantProtected, middleware.RequirePermission, middleware.RequireRole)
```

The adapters (`userLookup`, `roleResolverAdapter`) and `SyncPermissions` method delegate to IAM's exported contracts instead of individual module repositories.

## Legacy Coexistence Strategy

1. **No deletion**: `users/`, `roles/`, `permissions/` directories remain untouched
2. **No import changes in legacy**: Legacy modules continue to compile and work independently
3. **Container swap only**: `app/container.go` stops wiring legacy modules and wires IAM instead
4. **Dead code is acceptable**: Legacy modules become unreachable at runtime but compile fine — this is intentional for review/contrast
5. **Future cleanup**: A subsequent change (`22-legacy-iam-cleanup`) will delete the legacy modules after validation

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Partial model field mismatch | IAM models might miss fields that GORM needs for joins | Define partial models with exact fields used; test with real DB queries |
| Cache key collision | IAM uses same `rbac:permissions:{publicID}` cache keys as legacy | Intentional — same keys ensure seamless transition |
| Permission registry double-registration | If legacy routes are accidentally still mounted, `RequirePermission` re-registers slugs harmlessly (registry is idempotent) | Ensure `RegisterModules` only mounts IAM, not legacy |
| Container wiring complexity | IAM container has many collaborators (DB, cache, logger) | Follow existing container pattern; wire in phases |
| Test coverage gap | New module needs full test coverage | Each slice must have handler_test, service_test, repository_test |
| Migration scope creep | Temptation to "fix" things while copying | Strict rule: reproduce existing behavior exactly, no improvements |

## Migration Plan

1. **Create IAM module skeleton** — container.go, routes.go, core/ with partial models
2. **Implement user slices** — copy/adapt logic from users module, write tests
3. **Implement role slices** — copy/adapt logic from roles module, write tests
4. **Implement permission slices** — copy/adapt logic from permissions module, write tests
5. **Implement internal slices** — resolve_user_permissions, sync_permissions, resolve_auth_user, resolve_role_by_slug
6. **Wire IAM in app/container.go** — replace legacy wiring with IAM container
7. **Update RegisterModules** — mount IAM routes, remove legacy mounts
8. **Run full test suite** — `go test ./...` must pass
9. **Verify behavior parity** — manual API testing against all endpoints

## Rollback Plan

1. Revert `app/container.go` to wire legacy modules instead of IAM
2. Revert `RegisterModules` to mount legacy routes
3. IAM module code remains in tree but unused — no data loss
4. No migration files to rollback (no schema changes in this change)

## Ready for Proposal

**Yes.** The exploration is complete with:
- Full endpoint-to-slice mapping (19 slices identified)
- Internal method mapping (5 internal slices identified)
- Cross-module dependency analysis (4 dependencies eliminated)
- Container wiring change specified
- Legacy coexistence strategy defined
- Risks, migration plan, and rollback plan documented
- Recommended approach: multi-entity vertical slice with entity sub-folders

The orchestrator should proceed to: **proposal → specs → design → tasks**.
