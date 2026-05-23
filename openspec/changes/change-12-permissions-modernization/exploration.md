# Exploration: Permissions Modernization, Global Constants, and CRUD Refactoring

## Problem framing

Currently, the system’s permissions are defined and validated in routers (`routes.go`) and seeds (`seeds/permissions.go`) using raw string literals like `"users.index"`, `"roles.view"`, etc. This introduces fragility due to potential typos and makes extending permissions harder.

Additionally, the CRUD action `"index"` is mixed with `"list"`. We need to standardise on `"list"` and remove `"index"` entirely. 

Furthermore, having manual creations/deletions of permissions via public HTTP endpoints is obsolete and risky because the permissions existence is inherently tied to codebase handlers. The backend must be the **Single Source of Truth** for existing permissions, which should be automatically discovered and synchronised with the database during server startup (Bootstrapping).

Current behavior and codebase layout discovered:

- **Constants and Slugs**:
  - Raw strings like `"users.index"`, `"roles.view"`, `"companies.index"` are scattered in routing packages (`internal/modules/*/routes.go`).
  - Permissions are seeded via a static helper in `seeds/permissions.go` mapping to hardcoded slugs.
- **Action Inconsistency**:
  - The `"index"` action is used for list/collection routes, but many handlers use `.List` or `.ListSelect`.
  - The action `"index"` needs to be replaced globally with `"list"` in all routers, seeds, models, payloads, and unit/integration tests.
- **Self-Discovery and Registry**:
  - There is no transversal package or registry that collects what permissions are actually required by the mounted routes.
  - Middlewares do not automatically track which permission slugs they are guarding.
- **REST endpoints restrictions**:
  - `POST /api/v1/permissions` and `DELETE /api/v1/permissions/:id` currently exist and allow full manual curation of permissions.
  - `PUT /api/v1/permissions/:id` allows modifying the structural `Slug`, `Module`, and `Action` fields alongside `Name` and `Description`.

## Recommended change id

`change-12-permissions-modernization`

## Affected domains

1. **Platform Permissions (`internal/platform/permissions`)**
   - **NEW** transversal package.
   - Houses global constants for all modules (`users`, `roles`, `companies`, `settings`, `auth`, `permissions`) and CRUD actions (`list`, `select`, `view`, `create`, `update`, `delete`, `manage`, `change_role`, `assign_permissions`).
   - Implements `Format(module, action string) string` returning `"module.action"`.
   - Implements thread-safe in-memory registry: `Register(slug string)` and `ListRegistered() []string`.

2. **Middleware (`internal/middleware/authorization.go`)**
   - Import `internal/platform/permissions` and call `permissions.Register(slug)` inside `RequirePermission(slug string)` to capture required permissions during route initialization.

3. **Routing (`internal/modules/*/routes.go`)**
   - Update all `RequirePermission` calls to use `permissions.Format` with global constants.
   - Replace `"index"` with `permissions.ActionList` everywhere.

4. **Permissions Module (`internal/modules/permissions`)**
   - Implement `SyncPermissions(slugs []string) error` in its service and repository, doing GORM upserts based on dynamic in-memory slugs.
   - Remove `POST` and `DELETE` routes.
   - Modify `PUT /api/v1/permissions/:id` DTO/service to only accept updates to `Name` and `Description`, and explicitly reject edits to `Slug`, `Module`, and `Action`.

5. **Container & Bootstrap (`internal/app/*`)**
   - Update `Container` to expose the permissions sync capability.
   - Trigger the permissions bootstrapping synchronization in `internal/app/bootstrap.go` utilizing `permissions.ListRegistered()` collected after route initialization.

6. **Seeds and Tests (`seeds/*` and `tests/*`)**
   - Refactor all references of `"index"` to `"list"`.
   - Update seeds and test assertions to align with the new automatic synchronization and constants.

## Design decisions to carry forward

- **No Circular Imports and Modularity**: `internal/platform/permissions` will remain a lightweight transversal package holding only constants and the in-memory registry. It has zero dependencies on GORM or module business logic.
- **Single Responsibility**: The permissions module (`internal/modules/permissions`) owns the database representation and synchronization logic, which prevents GORM model duplication.
- **Humanisation and Idempotency**: `SyncPermissions` will generate readable default names and descriptions (e.g. `"List users"`, `"Allows listing users"`) when inserting new permissions. If they already exist, it will NOT overwrite custom `Name` or `Description` made by administrators via the `PUT` endpoint.
- **Route Deactivation**: Removing GIN router handlers for `POST` and `DELETE` makes them naturally return `404 Not Found`, which perfectly complies with the requirements.

## Risks

- **Seeding & Local Tests**: Tests that rely on a clean database might fail if they expect permissions to be seeded but the router hasn't run. `SyncPermissions` should be called during test environment setup to ensure tests see the fully synchronized permissions catalog.
- **Structural Integrity on Update**: If an administrator sends structural fields like `Slug` inside a `PUT /permissions/:id` payload, the system must validate and reject changes with a clear error rather than silently ignoring or mutating them.
