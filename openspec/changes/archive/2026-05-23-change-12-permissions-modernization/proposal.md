# Proposal: Permissions Modernization, Global Constants, and CRUD Refactoring

## Intent

To centralize and type module and action definitions via transversal global constants, eliminate raw string literal references in routers through a compilable format helper, globally replace the base `"index"` CRUD action with `"list"`, and implement an automatic in-memory self-discovery and synchronization (bootstrapping) mechanism at server start. Additionally, we will restrict the permissions API so it acts as read-only for structural changes (creation/deletion deactivated, update restricted to name/description only).

## Scope

### In Scope

- **Global Constants**: Define all modules (`users`, `roles`, `companies`, `settings`, `auth`, `permissions`) and predefined actions (`list`, `select`, `view`, `create`, `update`, `delete`, `manage`, `change_role`, `assign_permissions`) as constants in `internal/platform/permissions/constants.go`.
- **Formatting Helper**: Implement `permissions.Format(module, action string) string` to compile slugs like `"users.list"`.
- **Estandarización CRUD (`index` -> `list`)**: Remove `"index"` action globally across the codebase (routers, seeds, models, DTOs, tests, integration helper) and replace it with `"list"`.
- **Self-Discovery Registry**: Build a thread-safe registry (`permissions.Register(slug)` and `permissions.ListRegistered()`) in `internal/platform/permissions`.
- **Middleware Hook**: Call `permissions.Register(slug)` in `RequirePermission(slug)` middleware during router setup to discover all active route guards on startup.
- **Idempotent Bootstrapping Sync**:
  - Implement a `SyncPermissions(slugs []string) error` method in `permissions.Service` and `permissions.Repository`.
  - Expose this capability via a method `SyncPermissions() error` on `app.Container` that reads `permissions.ListRegistered()` and delegates sync to `permissions.Service`.
  - Trigger this synchronization inside `internal/app/bootstrap.go` during bootstrapping.
  - Sincronización automatically inserts or updates system permissions, humanising default values for new ones and preserving edits for existing ones.
- **Permissions API Hardening**:
  - Remove `POST /api/v1/permissions` and `DELETE /api/v1/permissions/:id` routes.
  - Modify `PUT /api/v1/permissions/:id` to strictly allow updating only `Name` and `Description`, explicitly rejecting modifications to structural fields (`Slug`, `Module`, `Action`).
- **Tests Validation**: Refactor all unit/integration tests to use new constants, `list` actions, and validated `PUT` behaviors.

### Out of Scope

- **Dynamic UI implementation**: Frontend components that consume the permissions list are outside the backend scope.
- **Custom database migration tables**: The database schema of `permissions` table remains the same.
- **Extending permissions beyond current modules**: Only existing modules/permissions will be modernized.

## Capabilities

### New Capabilities

- `self-discovery-bootstrapping`: The backend automatically registers and syncs all permissions required by API routes at server startup, rendering manual permission seeds obsolete.
- `permissions-constants`: Transversal, safe, and compile-checked constants for permission slugs and modules.

### Modified Capabilities

- `permissions-api`: The permissions management API is restricted to description edits; creation and deletion endpoints are eliminated.
- `global-crud`: The base collection query action is renamed from `"index"` to `"list"` across the entire project.

## Approach

1. **Transversal Constants & Registry**
   - Create `internal/platform/permissions/constants.go`.
   - Implement thread-safe registry with `sync.RWMutex` to record all unique slugs.
   - Implement `Format(module, action string) string`.

2. **Middleware Registration**
   - Import `github.com/enviniom/nexokit/internal/platform/permissions` inside `internal/middleware/authorization.go`.
   - Call `permissions.Register(slug)` inside `RequirePermission(slug string) gin.HandlerFunc`. Since this function is called during routes setup (instantiation of route guards), it automatically registers the slug during app startup.

3. **Modules Integration & Bootstrapping Sync**
   - Define `SyncPermissions(slugs []string) error` in `internal/modules/permissions/service.go`.
   - In `SyncPermissions`:
     - Parse `module` and `action` from slugs (e.g. `module.action`).
     - Humanise name (e.g. `"users.list"` -> `"List users"`) and description (e.g. `"Allows listing users"`) with display orders mapped per action.
     - Check database existence for each slug.
     - IF NOT FOUND: generate `PublicID` via `identity.Generate()`, set `IsSystem = true`, and create the permission.
     - IF FOUND: update structural fields (`Module`, `Action`, `IsSystem = true`) using GORM `Updates()` while preserving custom `Name` or `Description`.
   - Expose the sync capability via `Container.SyncPermissions()` in `internal/app/container.go`.
   - In `internal/app/bootstrap.go`, call `container.SyncPermissions()` right after container wiring.

4. **Global CRUD Rename (`index` -> `list`)**
   - Replace `ActionIndex = "index"` with `ActionList = "list"`.
   - Change `"index"` to `"list"` in all `routes.go` files (e.g., `users.GET("", ...)`).
   - Update `seeds/permissions.go` to use `permissions.ActionList`.
   - Rename tests and mock assertions from `"index"` to `"list"`.

5. **Permissions API Hardening**
   - Modify `internal/modules/permissions/routes.go`: Remove `POST` and `DELETE` registrations.
   - Modify `internal/modules/permissions/service.go`:
     - In `Update`, verify if `req.Slug != "" && req.Slug != permission.Slug`, `req.Module != "" && req.Module != permission.Module`, or `req.Action != "" && req.Action != permission.Action`. If any structural field differs, return `apperror.ErrForbidden`.
     - Remove `Create` and `Delete` service logic if they are completely obsolete. Redefine/update their tests to verify they are gone or return appropriate errors.

## Affected Areas

- `internal/platform/permissions/constants.go` — **NEW** transversal package.
- `internal/middleware/authorization.go` — import and call `permissions.Register(slug)`.
- `internal/app/container.go` — add `SyncPermissions` wrapper method delegating to `permissionsService`.
- `internal/app/bootstrap.go` — call `container.SyncPermissions()` during bootstrapping.
- `internal/modules/permissions/service.go` — implement `SyncPermissions(slugs)` and reject structural changes in `Update`.
- `internal/modules/permissions/repository.go` — add repo method for sync or batch upsert.
- `internal/modules/permissions/routes.go` — remove `POST` and `DELETE` routes, adjust `PUT`.
- `internal/modules/permissions/handler.go` — remove `Create` and `Delete` endpoints.
- `internal/modules/permissions/dto.go` — adjust DTOs.
- `internal/modules/companies/routes.go` — replace `"companies.index"` with `permissions.Format(...)`.
- `internal/modules/roles/routes.go` — rename `"roles.index"`.
- `internal/modules/users/routes.go` — rename `"users.index"`.
- `seeds/permissions.go` — refactor to use constants.
- Unit and integration tests under all modules, middleware, seeds, and integrations.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Test suite database synchronization | High | Make sure in test helpers (like `tests/helpers/app.go`) that the test router registration runs, which automatically populates memory and executes bootstrapping. If needed, manually call `SyncPermissions` during test DB seeding. |
| Loss of custom name/description edits | Medium | Ensure `SyncPermissions` does not overwrite existing permissions' `Name` and `Description` fields if they exist in the DB. |
| Breaking client APIs on PUT | Low | The client should only send `Name` and `Description` for PUT. If they send structural fields, check if they mismatch and reject with `403 Forbidden` or `422 Unprocessable Entity` validation errors. |

## Rollback Plan

Revert code changes to routers, seeds, and middleware. The database can be rolled back by restoring a snapshot or running a targeted cleanup migration.

## Success Criteria

- [ ] Global constants package `internal/platform/permissions` exists.
- [ ] No raw string literals are used in route definitions; all use `permissions.Format`.
- [ ] Action `"index"` is globally removed and replaced by `"list"`.
- [ ] Start server automatically populates the `permissions` table with all guarded slugs via `SyncPermissions`.
- [ ] Existing custom Name and Description in DB are preserved during startup upsert.
- [ ] `POST /api/v1/permissions` and `DELETE /api/v1/permissions/:id` return `404 Not Found` (due to route removal).
- [ ] `PUT /api/v1/permissions/:id` rejects structural modifications with 403 Forbidden / Validation Error.
- [ ] Entire test suite compiles and runs green.
