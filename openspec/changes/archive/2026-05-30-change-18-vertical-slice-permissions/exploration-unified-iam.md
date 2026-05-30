## Exploration: Unify users + roles + permissions into one IAM module

### Current State

Today the project has **three separate modules** under `internal/modules/`:

| Module | Files | Style | Cross-module imports |
|--------|-------|-------|---------------------|
| `users` | flat (11 files) | legacy flat | imports `roles` (model + service) |
| `roles` | flat (11 files) | legacy flat | imports `permissions` (model), `companies` (model) |
| `permissions` | hybrid flat + partial slices | mid-migration | none (only `platform` + `shared`) |

**Key coupling observed in code:**

1. **`users/model.go`** — `import "github.com/enviniom/nexokit/internal/modules/roles"` — embeds `roles.Role` directly and calls `roles.RootRoleSlug`.
2. **`users/service.go`** — `import "github.com/enviniom/nexokit/internal/modules/roles"` — uses `RoleResolver` interface returning `*roles.Role`, and `roles.AssignmentRoleReader`.
3. **`roles/model.go`** — `import "github.com/enviniom/nexokit/internal/modules/permissions"` — embeds `[]permissions.Permission` via GORM many2many.
4. **`roles/service.go`** — `import "github.com/enviniom/nexokit/internal/modules/permissions"` — uses `permissions.Permission` in catalog building and system-permission checks.
5. **`app/container.go`** — wires all three together: passes `permissionsRepo` to roles, `rolesRepo` to users, `usersRepo` to auth middleware.

**Database reality (single migration):**
- `users.role_id` → FK to `roles.id` (ON DELETE RESTRICT)
- `roles` ↔ `permissions` via `role_permissions` join table
- All three tables are in the same migration file, created together

**Domain reality (from `_context.md`):**
- "Un usuario tiene un solo rol"
- "Roles base: root, admin, user"
- "Autorización por permisos, no por nombre de rol"
- "Root tiene todos los permisos"

These three entities form a **single RBAC bounded context**: users cannot exist without roles, roles are meaningless without permissions, and permissions only matter when assigned to roles consumed by users.

### Affected Areas

| Area | Impact if unified |
|------|-------------------|
| `internal/modules/users/` | Would become a sub-folder or entity group inside unified module |
| `internal/modules/roles/` | Would become a sub-folder or entity group inside unified module |
| `internal/modules/permissions/` | Would become a sub-folder or entity group inside unified module |
| `internal/app/container.go` | Wiring simplified: one module container instead of three, no cross-module adapter types |
| `internal/middleware/authorization.go` | `PermissionResolver` interface unchanged; only import path changes |
| `internal/platform/permissions/` | Constants/registry stay in platform (cross-app concern) — no change |
| `internal/modules/auth/` | Uses `users.Repository` via adapter; import path changes only |
| API routes | URL paths (`/api/v1/users`, `/api/v1/roles`, `/api/v1/permissions`) stay identical |
| Existing tests | All tests move with their code; test contracts unchanged |

### Approaches

#### 1. Unified module (`internal/modules/iam/`) with entity sub-folders

Create one module `iam` (or `access`, `identity_access`) using the multi-entity vertical slice pattern already defined in `_context.md` §8:

```
internal/modules/iam/
  container.go
  routes.go
  core/                    ← shared IAM vocabulary (errors, common DTOs)
  users/
    container.go
    routes.go
    create_user/
    view_user/
    update_user/
    delete_user/
    change_password/
    change_role/
    toggle_status/
  roles/
    container.go
    routes.go
    create_role/
    view_role/
    update_role/
    delete_role/
    assign_permissions/
    permission_catalog/
    list_select/
  permissions/
    container.go
    routes.go
    list_permissions/
    view_permission/
    update_permission/
  resolve_permissions/       ← internal, non-HTTP (auth middleware)
  sync_permissions/          ← internal, bootstrap
  queries/                   ← shared queries across entities
```

- **Pros:**
  - Eliminates ALL cross-module imports between users/roles/permissions — they share one module boundary
  - `core/` can hold shared IAM vocabulary (role slugs, permission format helpers) without polluting `platform/`
  - `app/container.go` wiring becomes simpler: one container, no adapter types like `roleResolverAdapter`
  - Matches the real bounded context: RBAC is one cohesive domain
  - Multi-entity vertical slice pattern is already documented and supported in `_context.md`
  - No API-breaking changes: routes, DTOs, and behavior stay identical
  - Future IAM features (e.g., teams, groups, SSO) have a natural home

- **Cons:**
  - **Large migration**: ~1200+ lines across three modules to restructure in one change
  - Loses module autonomy: all three entities deploy together (but they already do — same DB, same API)
  - Review cognitive load is high even with chained PRs
  - Name bikeshed: `iam` vs `access` vs `identity_access` — needs a decision
  - Existing change-18 scope would need to be absorbed or deferred

- **Effort:** **High** (3 modules × vertical slice migration + container + wiring)

#### 2. Keep separate modules, eliminate cross-module imports via contracts

Keep three modules. Each defines local partial models and consumer contracts for what it needs from the others:

- `users` defines `RoleSummary` locally instead of importing `roles.Role`
- `roles` defines `PermissionSummary` locally instead of importing `permissions.Permission`
- `app/container.go` wires implementations via interfaces

- **Pros:**
  - Smaller individual changes: each module migrates independently
  - Preserves module autonomy (project convention)
  - Change-18 stays focused on permissions only
  - Lower review cognitive load per PR
  - Follows existing `_context.md` §5 pattern ("definir modelo local parcial")

- **Cons:**
  - **Duplication**: role/permission partial models repeated across modules
  - The three modules ARE one bounded context — artificial separation creates friction
  - `app/container.go` needs adapter types (`roleResolverAdapter`, `userLookup`) that would disappear in a unified module
  - Every time IAM vocabulary changes (new role slug, new permission field), multiple modules may need updates
  - The "partial model" workaround exists precisely because the separation is artificial

- **Effort:** **Medium** per module (but three separate migrations)

#### 3. Hybrid: unified module but migrate incrementally

Create `internal/modules/iam/` and migrate one entity at a time, keeping old modules temporarily:

- Phase 1: Move `permissions` → `iam/permissions/` (absorbs change-18)
- Phase 2: Move `roles` → `iam/roles/`
- Phase 3: Move `users` → `iam/users/`
- Phase 4: Delete old module folders, finalize container

- **Pros:**
  - Each phase is reviewable and reversible
  - Change-18 becomes the first phase of a larger effort
  - Reduces risk of breaking auth flow

- **Cons:**
  - **Temporary messy state**: old and new modules coexist
  - More total work (migration + cleanup phases)
  - `app/container.go` needs conditional wiring during transition
  - Risk of forgetting to clean up old modules

- **Effort:** **High** total, but **Medium** per phase

### Comparison Matrix

| Criterion | Option 1: Unified | Option 2: Separate + contracts | Option 3: Hybrid incremental |
|-----------|-------------------|-------------------------------|------------------------------|
| Bounded context accuracy | ✅ Best | ❌ Artificial split | ✅ Eventually best |
| Cross-module imports | ✅ Zero | ⚠️ Eliminated via workaround | ✅ Zero (eventually) |
| Review load per PR | ❌ High | ✅ Low-Medium | ✅ Medium |
| Total effort | High | Medium × 3 | High |
| API compatibility | ✅ Unchanged | ✅ Unchanged | ✅ Unchanged |
| Test impact | Move with code | Move with code | Move with code |
| Alignment with `_context.md` | ✅ Multi-entity pattern | ✅ Partial model pattern | ⚠️ Not explicitly covered |
| Change-18 impact | Absorbed into larger change | Unaffected | Becomes phase 1 |

### Recommendation

**Option 1 (Unified IAM module) is architecturally correct but should NOT be change-18.**

Here's why:

1. **The bounded context argument is strong.** Users, roles, and permissions ARE one RBAC context. The cross-module imports prove the current separation is artificial. The multi-entity vertical slice pattern in `_context.md` was designed exactly for this scenario.

2. **But the migration cost is too high for change-18.** Change-18 was scoped as "migrate permissions only." Expanding to three modules triples the work, review load, and risk. This deserves its own change with proper planning.

3. **Recommended path:**
   - **Keep change-18 as-is**: migrate `permissions` to vertical slice, eliminate the `roles → permissions` import by having roles define a local `PermissionSummary` partial model. This is the right scope for this change.
   - **Create a future change** (e.g., `change-XX-unified-iam-module`) that unifies all three modules under `iam/` with entity sub-folders. This change should:
     - Start after change-18 is complete (permissions already in vertical slice)
     - Migrate roles and users to vertical slice as part of the unification
     - Use chained PRs with the multi-entity structure
     - Be planned as a coordinated effort, not a quick refactor

4. **If the user insists on unifying NOW**, recommend Option 3 (hybrid incremental) starting with permissions as phase 1, which naturally absorbs change-18.

### Risks

1. **Auth flow breakage**: Any mistake in wiring the permission resolver or user lookup breaks login for the entire app. Testing is critical.
2. **Route registration complexity**: Multi-entity modules need nested containers and route delegation. If done wrong, routes won't mount.
3. **GORM many2many**: The `role_permissions` join table is owned by both roles and permissions. In a unified module this is fine; in separate modules it requires careful partial model handling.
4. **Cached permission invalidation**: `rbac:permissions:{publicID}` cache keys are used across modules. Unification doesn't change this, but migration must preserve the cache invalidation flow.
5. **Existing untracked work**: `permissions/` already has partial vertical slice files from change-16. Any unification must preserve and integrate this work, not discard it.

### Ready for Proposal

**Yes, but NOT as change-18.** The orchestrator should:

1. Tell the user: "Tu intuición es correcta — users, roles y permissions forman un solo bounded context de RBAC. Pero unificar los tres módulos en un solo change es demasiado riesgo y trabajo para el change-18 actual."
2. Recommend: Keep change-18 focused on permissions vertical slice only, with roles defining a local partial model for permissions.
3. Propose a **new future change** for the full IAM unification, which would be a larger coordinated effort with chained PRs.
4. If the user wants to pivot change-18 to the unified approach, recommend Option 3 (incremental) starting with permissions as phase 1.
