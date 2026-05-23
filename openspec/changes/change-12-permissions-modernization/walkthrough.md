# Walkthrough: Permissions Modernization, Global Constants, and CRUD Refactoring

We have successfully implemented **Change 12: Permissions Modernization, Global Constants, and CRUD Refactoring**. The entire application has been modernized, the permissions API has been hardened, and all automated unit and integration tests are compiled and pass with **100% green status**.

---

## Changes Implemented

### 1. Transversal Platform Package (`internal/platform/permissions`)
* **`constants.go`**: Centralizes global modules, actions, and a type-safe formatter helper:
  - `ModuleUsers`, `ModuleRoles`, `ModuleCompanies`, `ModulePermissions`, etc.
  - `ActionList` (replacing the old `index`), `ActionView`, `ActionCreate`, `ActionUpdate`, `ActionDelete`, `ActionManage`, etc.
  - `Format(module, action string) string` helper that formats module and action into `"module.action"` safely.
* **`registry.go`**: Implements a thread-safe, in-memory registry using `sync.RWMutex` to record all active permissions during route definitions.
* **`permissions_test.go`**: Validates formatter accuracy and registry thread-safety under concurrent load.

### 2. Auto-Discovery Middleware Hook
* **`internal/middleware/authorization.go`**: Updated `RequirePermission(slug)` to call `permissions.Register(slug)` automatically as routes are mounted during application startup.

### 3. Dynamic Bootstrapping Sync
* **`internal/modules/permissions/service.go`**: Implemented `SyncPermissions(slugs []string) error` with GORM upsert logic:
  - Automatically parses the `module` and `action` from slugs.
  - Formulates readable, default system-level names and descriptions (e.g. `users.list` becomes `"List users"`).
  - Assigns an action-based display order.
  - Performs idempotent upsert matching by `slug` — inserting missing permissions (with `identity.Generate()` for `PublicID` and `IsSystem = true`) and updating structural fields while preserving customized `Name` or `Description` manual edits.
* **`internal/app/container.go`**: Exposed a `SyncPermissions()` method linking registered platform slugs to the permissions service.
* **`internal/app/bootstrap.go`**: Wired `container.SyncPermissions()` right after container creation to automatically synchronize the DB with routes at startup.

### 4. Global CRUD Refactor (`index` -> `list`)
* **`internal/modules/permissions/model.go`**: Removed `ActionIndex = "index"`.
* **Routing Configuration Refactoring**: Updated company, role, and user routes to use global constants and `permissions.ActionList` instead of `"*.index"`.
* **Database Seeds & Integration Tests**: Modified `seeds/permissions.go`, `tests/integration/rbac_test.go`, and test fixtures across the application to consistently query `"*.list"`.

### 5. Hardened Permissions API
* **`internal/modules/permissions/routes.go`**: Completely removed GIN endpoints for `POST ""` (creation) and `DELETE "/:id"` (deletion), naturally making the server return `404 Not Found`.
* **`internal/modules/permissions/service.go`**: Hardened the `Update` method to explicitly check if `Slug`, `Module`, or `Action` are being mutated. If so, it returns `apperror.ErrForbidden` (HTTP 403) to prevent unauthorized tampering of system structures.
* **`internal/modules/permissions/dto.go`**: Adjusted the update validation request DTO to focus solely on non-structural updates (`Name`, `Description`).

---

## Verification Results

### Automated Test Execution
We executed the entire suite of automated unit, module, and integration tests (`go test ./...`) inside the workspace:

```bash
$ go test ./...
ok      github.com/enviniom/nexokit/internal/middleware (cached)
ok      github.com/enviniom/nexokit/internal/modules/auth       (cached)
ok      github.com/enviniom/nexokit/internal/modules/companies  (cached)
ok      github.com/enviniom/nexokit/internal/modules/permissions        (cached)
ok      github.com/enviniom/nexokit/internal/modules/roles      (cached)
ok      github.com/enviniom/nexokit/internal/modules/users      (cached)
ok      github.com/enviniom/nexokit/internal/platform/permissions       0.003s
ok      github.com/enviniom/nexokit/tests/integration   (cached)
# All tests passed successfully!
```

### Key Security Safeguards Verified
1. **Rejection of Structural Changes**: Asserted that attempting to modify a permission's `Slug`, `Module`, or `Action` via `PUT /api/v1/permissions/:id` returns a strict `403 Forbidden` error.
2. **Endpoint Deactivation**: Asserted that calling `POST /api/v1/permissions` or `DELETE /api/v1/permissions/:id` returns a strict `404 Not Found` error.
3. **Upsert Idempotency**: Verified that restarting the server synchronizes new guards into GORM without altering pre-existing custom `Name` or `Description` attributes.
