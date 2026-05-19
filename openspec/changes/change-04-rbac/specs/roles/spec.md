# Delta for Roles

## ADDED Requirements

### Requirement: Role permission catalog endpoint

The system MUST expose `GET /api/v1/roles/:id/permissions`. It MUST return the full permission catalog grouped by `module`, with each permission annotated with a `granted` boolean indicating whether the role holds that permission. The response shape MUST be UI-friendly: a client can render the grouped catalog without transformation. This endpoint MUST require authentication.

#### Scenario: Catalog with granted flags

- GIVEN a role `admin` with permissions `users.index`, `users.view`, `roles.index`
- WHEN `GET /api/v1/roles/:admin_id/permissions` is called
- THEN the response returns HTTP 200 with `data` containing modules
- AND each module lists its permissions with `granted: true` for assigned and `granted: false` for unassigned

#### Scenario: Role not found

- GIVEN no role with the given ID
- WHEN `GET /api/v1/roles/:id/permissions` is called
- THEN the response returns HTTP 404

### Requirement: Role permission assignment endpoint

The system MUST expose `PUT /api/v1/roles/:id/permissions`. It MUST accept a request body containing an array of permission slugs and replace the role's assignments with exactly those slugs. It MUST require `roles.assign_permissions` permission. The system MUST NOT allow removal of system permissions (`is_system: true`) from system roles (`is_system: true`). After successful assignment, it MUST invalidate the permission cache for all members of that role.

#### Scenario: Replace role permissions

- GIVEN an admin user with `roles.assign_permissions` permission
- WHEN `PUT /api/v1/roles/:id/permissions` is called with `["users.index", "users.view"]`
- THEN the role's assignments are replaced with exactly those slugs
- AND the response returns HTTP 200 with the updated grouped catalog
- AND the permission cache for all members of that role is invalidated

#### Scenario: System role system permission protection

- GIVEN a system role (`is_system: true`) with system permission `users.index`
- WHEN `PUT /api/v1/roles/:id/permissions` is called without `users.index`
- THEN the response returns HTTP 403 and system permissions remain assigned

#### Scenario: Unauthorized assignment

- GIVEN a user without `roles.assign_permissions` permission
- WHEN `PUT /api/v1/roles/:id/permissions` is called
- THEN the response returns HTTP 403

## MODIFIED Requirements

### Requirement: Read-only role API

The system MUST expose `GET /api/v1/roles` and `GET /api/v1/roles/:id`. It MUST NOT expose mutation endpoints for roles. Responses MUST use the standard DTO envelope. Each role response MUST include a `permissions` field containing an array of permission slug strings associated with that role.
(Previously: Role API responses did not include associated permission slugs)

#### Scenario: List roles

- GIVEN the system seeds have run
- WHEN `GET /api/v1/roles` is called
- THEN the response returns HTTP 200 and `data` contains the roles `root`, `admin`, and `user`
- AND each role object includes a `permissions` array with the associated permission slugs

#### Scenario: Get role by ID

- GIVEN a seeded role exists with assigned permissions
- WHEN `GET /api/v1/roles/:id` is called with a valid ID
- THEN the response returns HTTP 200 and `data` contains the role with its `permissions` array

### Requirement: Role seeds

The system MUST seed the roles `root`, `admin`, and `user` on startup or via a seed command. The operation MUST be idempotent. The `root` role MUST be seeded with all system permissions. The `admin` role MUST be seeded with admin-level permissions. The `user` role MUST be seeded with basic permissions (`users.index`, `users.view`, `auth.view`).
(Previously: Role seeds did not include permission assignments)

#### Scenario: Idempotent seeding

- GIVEN the roles have already been seeded
- WHEN the seed command runs again
- THEN no duplicate roles are created and the process exits successfully

#### Scenario: Root role has all permissions

- GIVEN the system seeds have run
- WHEN the root role is retrieved
- THEN its `permissions` array contains all system permission slugs