## Exploration: change-18-vertical-slice-permissions

### Current State

The permissions module is currently a **flat module** with all logic at the root level (`handler.go`, `service.go`, `repository.go`, `model.go`, `dto.go`). It exposes 3 registered HTTP endpoints and 2 non-HTTP service methods (`Resolve`, `SyncPermissions`).

**Registered endpoints:**
| Method | Path | Handler | Business Intent |
|--------|------|---------|-----------------|
| GET | `/permissions` | `handler.List` | List all permissions grouped by module |
| GET | `/permissions/:id` | `handler.GetByPublicID` | View a single permission |
| PUT | `/permissions/:id` | `handler.Update` | Update a non-system permission |

**Unregistered handler methods** (exist in `handler.go` but NOT wired in `routes.go`):
- `ListPaginated` — paginated list with filters
- `Create` — create non-system permission
- `Delete` — delete non-system permission

**Non-HTTP service methods:**
- `Resolve(publicID)` — used by `middleware.AttachPermissions`, cache-backed user permission resolution
- `SyncPermissions(slugs)` — called from `app.Container.SyncPermissions()` at bootstrap, syncs registered slugs to DB

**Partial vertical slice work already exists (13 untracked files):**
- `core/dto.go`, `core/enums.go`, `core/model.go` — shared elements extracted
- `list_permissions/` — handler, service, repository (ListGrouped only)
- `view_permission/` — handler, service, repository (GetByPublicID)
- `update_permission/` — handler, service, repository (Update)
- `queries/get_permission_by_public_id.go` — shared query

**Important:** These untracked files are from change-16 leftover work. They must be treated as **input**, not disposable. They are mostly correct but incomplete and have minor issues (see Risks).

### Affected Areas

- `internal/modules/permissions/` — primary target, full restructuring
- `internal/app/container.go` — wiring changes: from flat handler/service to container-based
- `internal/server/router.go` — no changes (mounts module via `permissions.Register`)
- `internal/modules/roles/service.go` — imports `permissions.Permission` type directly; needs local partial model
- `internal/modules/roles/repository.go` — `ReplacePermissions` uses raw SQL on `role_permissions` (acceptable, no change needed)
- `internal/middleware/auth.go` — uses `permissions.Service.Resolve()` via `AttachPermissions`; contract must be preserved

### Endpoint → Slice Mapping

| Registered Endpoint | Proposed Slice | Notes |
|---------------------|---------------|-------|
| `GET /permissions` | `list_permissions` | Already partially exists; needs tests |
| `GET /permissions/:id` | `view_permission` | Already partially exists; uses shared query |
| `PUT /permissions/:id` | `update_permission` | Already partially exists |
| *(unregistered)* `ListPaginated` | `list_permissions_paginated` | Only create if endpoint will be registered |
| *(unregistered)* `Create` | `create_permission` | Only create if endpoint will be registered |
| *(unregistered)* `Delete` | `delete_permission` | Only create if endpoint will be registered |
| *(non-HTTP)* `Resolve` | `resolve_permissions` | Not an HTTP slice; keep as module-level service or extract to its own slice for internal wiring |
| *(non-HTTP)* `SyncPermissions` | `sync_permissions` | Bootstrap-only; keep at module level or extract to its own slice |

**Decision:** Only create slices for the 3 **registered** endpoints (`list_permissions`, `view_permission`, `update_permission`). The unregistered handler methods (`ListPaginated`, `Create`, `Delete`) and non-HTTP methods (`Resolve`, `SyncPermissions`) must be preserved as module-level code or extracted to internal slices **only when** their endpoints are registered in a future change. Creating slices for unregistered endpoints violates the "no slices for endpoints that don't exist" rule.

### Cross-Module Dependencies and How to Remove Them

**1. `roles` → `permissions` (direct import)**
- `roles/service.go` imports `github.com/enviniom/nexokit/internal/modules/permissions` to use `permissions.Permission` type
- `roles/service.go` defines `permissionCatalogRepository` interface with `ListAll() ([]permissions.Permission, error)`
- `app/container.go` passes `permissionsRepo` directly to roles via `WithPermissionCatalog(permissionsRepo)`

**Resolution:** Define a local partial model in `permissions/core/model.go` (already done — `core.Permission` exists). The roles module should define its own consumer interface that returns `[]PermissionSummary` with only the fields it needs (`Slug`, `Name`, `Module`, `Action`, `Description`, `IsSystem`, `DisplayOrder`, `PublicID`, `ID`). The permissions module can expose this via a small contract in `core/contracts.go`.

**2. `permissions` → `users` + `roles` tables (in repository)**
- `ListSlugsByUserPublicID` joins `permissions`, `role_permissions`, and `users` tables
- `AutoAssignToAdmins` inserts into `role_permissions` selecting from `roles`

**Resolution:** These are acceptable. The permissions module owns the `permissions` table and the `role_permissions` join table (it's a many-to-many where permissions is one side). Raw SQL queries against these tables within the permissions module's own repository do NOT violate the "no cross-module repository" rule — they query the database directly, not another module's repository. The `_context.md` says "no importar repositories ni modelos GORM de otro módulo" — direct SQL on shared tables is fine.

### Duplicated Queries Candidates for `queries/`

| Query | Used By | Candidate |
|-------|---------|-----------|
| `GetByPublicID` | `view_permission`, `update_permission`, root repository | ✅ Already extracted to `queries/get_permission_by_public_id.go` |
| `GetBySlug` | root repository (Create, SyncPermissions) | Create `queries/get_permission_by_slug.go` |
| `ListAll ordered` | `list_permissions`, root repository (ListGrouped) | Create `queries/list_all_permissions.go` |
| `List paginated ordered` | root repository (List) | Create `queries/list_permissions_paginated.go` |

### Proposed Final Structure

```
internal/modules/permissions/
  container.go              (NEW — composition root)
  routes.go                 (MODIFY — use container handlers)
  core/
    model.go                (EXISTS — Permission, RolePermission)
    dto.go                  (EXISTS — Request/Response DTOs)
    enums.go                (EXISTS — Action constants)
    contracts.go            (NEW — PermissionCatalogReader interface for roles to consume)
    error.go                (NEW — module errors if needed)
  queries/
    get_permission_by_public_id.go  (EXISTS)
    get_permission_by_slug.go       (NEW)
    list_all_permissions.go         (NEW)
    list_permissions_paginated.go   (NEW)
  list_permissions/
    handler.go              (EXISTS — needs tests)
    handler_test.go         (NEW)
    service.go              (EXISTS — needs tests)
    service_test.go         (NEW)
    repository.go           (EXISTS — needs tests)
    repository_test.go      (NEW)
  view_permission/
    handler.go              (EXISTS — needs tests)
    handler_test.go         (NEW)
    service.go              (EXISTS — needs tests)
    service_test.go         (NEW)
    repository.go           (EXISTS — needs tests)
    repository_test.go      (NEW)
  update_permission/
    handler.go              (EXISTS — needs tests)
    handler_test.go         (NEW)
    service.go              (EXISTS — needs tests)
    service_test.go         (NEW)
    repository.go           (EXISTS — needs tests)
    repository_test.go      (NEW)
  resolve_permissions/      (NEW — internal, non-HTTP)
    service.go
    service_test.go
    repository.go
    repository_test.go
  sync_permissions/         (NEW — internal, bootstrap)
    service.go
    service_test.go
    repository.go
    repository_test.go
  handler.go                (DELETE — replaced by slice handlers)
  service.go                (DELETE — logic distributed to slices)
  repository.go             (DELETE — logic distributed to slices)
  model.go                  (DELETE — moved to core/model.go)
  dto.go                    (DELETE — moved to core/dto.go)
  handler_test.go           (MODIFY or DELETE — tests move to slices)
  service_test.go           (MODIFY or DELETE — tests move to slices)
  routes_test.go            (KEEP — validates route wiring)
```

### Approaches

1. **Full migration in one change** — Migrate all 3 registered endpoints to slices + extract internal services + create container + update wiring
   - Pros: Clean end state, single PR to review
   - Cons: Large diff (~600+ lines), high review cognitive load
   - Effort: **High**

2. **Phased migration** — First migrate the 3 registered endpoints to slices, keep `Resolve`/`SyncPermissions` at module level temporarily
   - Pros: Smaller diff, safer, preserves behavior exactly
   - Cons: Temporary hybrid state, needs follow-up change
   - Effort: **Medium**

### Recommendation

**Approach 1 (Full migration)** with careful task splitting. The existing partial work already covers 3 of the slices, so the remaining work is:
- Create `container.go` and update `routes.go`
- Add tests to existing slices
- Create `resolve_permissions` and `sync_permissions` internal slices
- Extract remaining queries
- Clean up root-level files
- Update `app/container.go` wiring
- Create `core/contracts.go` for roles to consume

The total estimated diff is manageable if tasks are split properly. The review workload forecast should be evaluated at the tasks phase.

### Risks

1. **Package name inconsistency**: `view_permission/` folder has `package view_permissions` (plural) — must be fixed to `package view_permission` (singular) to match folder name and Go conventions.
2. **Duplicate types during migration**: `core/dto.go` and `dto.go` both define the same types. The root files must be deleted ONLY after slices reference `core/` types and all tests pass.
3. **`Resolve` and `SyncPermissions` are not HTTP endpoints**: These are called by middleware and bootstrap respectively. They need special handling — either kept as module-level functions or extracted to internal slices that the container wires without HTTP registration.
4. **Roles module cross-import**: `roles/service.go` directly imports `permissions.Permission`. This should be addressed in a **separate change** (not this one) by having roles define its own partial model. For this change, the import can remain since we're only restructuring permissions internally.
5. **Unregistered endpoints**: `Create`, `Delete`, `ListPaginated` exist in the flat handler but are not wired. They must be preserved somewhere (likely module-level or future slices) so they're not lost.
6. **Test preservation**: The existing `handler_test.go` and `service_test.go` contain valuable test coverage. Tests must be migrated to their corresponding slice test files, not discarded.
7. **`ListSlugsByUserPublicID` joins 3 tables**: This query is complex and used by the auth middleware. It should be extracted to its own repository in `resolve_permissions/` with thorough tests.

### Ready for Proposal

**Yes.** The exploration is complete. The orchestrator should proceed to `sdd-propose` with:
- Scope: Migrate only `internal/modules/permissions/` to vertical slice
- Preserve all existing behavior (3 registered endpoints + Resolve + SyncPermissions)
- Create container.go, update routes.go, wire in app/container.go
- Existing untracked files are the starting point, not disposable
- Cross-module roles→permissions import is OUT OF SCOPE for this change
