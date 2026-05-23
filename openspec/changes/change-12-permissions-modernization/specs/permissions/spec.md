# Delta for Permissions

## MODIFIED Requirements

### Requirement: CRUD Operations Renaming (Index to List)
The collection query action is renamed globally from `"index"` to `"list"` in all slugs, models, seeds, and API routing.

### Requirement: Self-Discovery and Bootstrapping
On application startup, the system MUST automatically discover all required permission slugs that protect API routes (via GIN route definition loading and a middleware hook) and synchronise them into the `permissions` database table.
- **Idempotency**: The bootstrapping process must be idempotent. If a permission already exists, it is not recreated.
- **Structural Fields**: The bootstrapping process automatically parses the slug into `Module` and `Action` fields, registers `IsSystem = true`, and generates default values for `Name` and `Description` if they are not already present.
- **Preservation of Manual Edits**: If a permission already exists in the database, its user-edited `Name` and `Description` fields MUST NOT be overwritten during startup bootstrapping.

### Requirement: Restricted API Capabilities
The public CRUD API for permissions is restricted as follows:
- **Creation Endpoint**: `POST /api/v1/permissions` is disabled (returns HTTP 404).
- **Deletion Endpoint**: `DELETE /api/v1/permissions/:id` is disabled (returns HTTP 404).
- **Update Endpoint**: `PUT /api/v1/permissions/:id` allows updating **ONLY** the `Name` and `Description` fields.
  - **Structural Protection**: Any attempt to modify structural fields (`Slug`, `Module`, `Action`) via PUT request MUST be rejected with a clear error (e.g. `403 Forbidden` or `422 Unprocessable Entity`).

---

### Scenario: Startup Bootstrapping Sync
- GIVEN the web server is starting up
- WHEN GIN routes are registered with `RequirePermission("users.list")` and `RequirePermission("roles.view")`
- THEN the system automatically discovers these slugs
- AND syncs them to the database, creating them with `IsSystem = true`, parsing their module/action, and humanising default Name/Description if missing.

### Scenario: Preservation of Custom Details
- GIVEN an existing permission `users.list` in the database with custom Name `"Listado de Usuarios"` and Description `"Permite ver todos los usuarios"`
- WHEN the web server starts up and performs bootstrapping
- THEN the permission's Name and Description remain `"Listado de Usuarios"` and `"Permite ver todos los usuarios"` (they are not overwritten by defaults).

### Scenario: API Creation and Deletion Disabled
- GIVEN the application is running
- WHEN a client sends `POST /api/v1/permissions` or `DELETE /api/v1/permissions/:id`
- THEN the server returns `404 Not Found`.

### Scenario: Structural Modifications Rejected on Update
- GIVEN an existing permission with Slug `"users.list"`, Module `"users"`, and Action `"list"`
- WHEN a client sends `PUT /api/v1/permissions/:id` with Slug `"users.admin"`, Module `"users"`, and Action `"admin"`
- THEN the server rejects the request with `403 Forbidden` (or validation error) and the database record is not updated.
