# Delta for Roles

## ADDED Requirements

### Requirement: Role CRUD API

The system MUST expose `POST /api/v1/roles`, `PUT /api/v1/roles/:id`, and `DELETE /api/v1/roles/:id` in addition to existing GET endpoints. POST and PUT MUST require `roles.create` and `roles.update` permissions respectively. DELETE MUST require `roles.delete` permission. All mutation endpoints MUST reject operations on system roles (`is_system: true`) with HTTP 403. Request bodies MUST include validated `slug` fields following the `ValidSlug()` rule (lowercase alphanumeric with hyphens, no leading/trailing hyphens).

#### Scenario: Create custom role

- GIVEN an authenticated user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with valid `name`, `slug`, and optional `description`
- THEN the response returns HTTP 201 with the created role (`is_system: false`)

#### Scenario: Slug format validation

- GIVEN a user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with slug `UPPER-CASE` or `-leading-hyphen`
- THEN the response returns HTTP 422 with a validation error

#### Scenario: Slug and name uniqueness

- GIVEN a role with slug `manager` already exists
- WHEN `POST /api/v1/roles` is called with slug `manager` or a duplicate name
- THEN the response returns HTTP 409 with a conflict error

#### Scenario: Update custom role

- GIVEN a custom role exists and user has `roles.update` permission
- WHEN `PUT /api/v1/roles/:id` is called with updated fields
- THEN the response returns HTTP 200 with the updated role

#### Scenario: Update rejects system role

- GIVEN a system role (`is_system: true`)
- WHEN `PUT /api/v1/roles/:id` is called
- THEN the response returns HTTP 403

#### Scenario: Delete custom role

- GIVEN a custom role with no assigned users and user has `roles.delete` permission
- WHEN `DELETE /api/v1/roles/:id` is called
- THEN the response returns HTTP 204 and the role is soft-deleted

#### Scenario: Delete rejects system role

- GIVEN a system role (`is_system: true`)
- WHEN `DELETE /api/v1/roles/:id` is called
- THEN the response returns HTTP 403

### Requirement: Delete guard for assigned users

The system MUST NOT allow deletion of a role that has users assigned to it. Before deleting, the system MUST check for assigned users using an efficient COUNT query. If any users are assigned, the system MUST return HTTP 422 with a message indicating the role has assigned users.

#### Scenario: Delete blocked by assigned users

- GIVEN a custom role with one or more assigned users
- WHEN `DELETE /api/v1/roles/:id` is called
- THEN the response returns HTTP 422 with message "role has assigned users"
- AND the role is NOT deleted

#### Scenario: Delete allowed after users reassigned

- GIVEN a custom role that previously had assigned users, now reassigned to another role
- WHEN `DELETE /api/v1/roles/:id` is called
- THEN the response returns HTTP 204 and the role is soft-deleted

## MODIFIED Requirements

### Requirement: Role API

The system MUST expose `GET /api/v1/roles`, `GET /api/v1/roles/:id`, `POST /api/v1/roles`, `PUT /api/v1/roles/:id`, and `DELETE /api/v1/roles/:id`. GET endpoints return all roles including custom roles. Mutation endpoints MUST enforce permission guards and system-role protection. Responses MUST use the standard DTO envelope. Each role response MUST include a `permissions` field containing an array of permission slug strings associated with that role.
(Previously: Read-only role API with GET endpoints only; mutation endpoints were not exposed)

#### Scenario: List roles includes custom roles

- GIVEN the system has seeded roles and custom roles created via API
- WHEN `GET /api/v1/roles` is called
- THEN the response returns HTTP 200 with `data` containing all roles (seeded and custom)
- AND each role object includes a `permissions` array

#### Scenario: Get role by ID

- GIVEN a role exists (seeded or custom) with assigned permissions
- WHEN `GET /api/v1/roles/:id` is called with a valid ID
- THEN the response returns HTTP 200 with the role and its `permissions` array

### Requirement: Role seeds

The system MUST seed the roles `root`, `admin`, and `user` on startup or via a seed command. The operation MUST be idempotent. The `root` role MUST be seeded with all system permissions. The `admin` role MUST be seeded with admin-level permissions including `roles.create`, `roles.update`, and `roles.delete`. The `user` role MUST be seeded with basic permissions (`users.index`, `users.view`, `auth.view`). Seeded roles MUST have `is_system: true`; custom roles created via API MUST have `is_system: false`.
(Previously: Seeds did not include `roles.*` permissions for admin role; no distinction between system and custom roles in seeding context)

#### Scenario: Idempotent seeding

- GIVEN the roles have already been seeded
- WHEN the seed command runs again
- THEN no duplicate roles are created and the process exits successfully

#### Scenario: Admin role has role management permissions

- GIVEN the system seeds have run
- WHEN the admin role is retrieved
- THEN its `permissions` array includes `roles.create`, `roles.update`, and `roles.delete`

### Requirement: System role protection

The system MUST mark seeded roles as system-level (`is_system: true`) to distinguish them from custom roles. System roles MUST NOT be editable via `PUT /api/v1/roles/:id` (returns 403). System roles MUST NOT be deletable via `DELETE /api/v1/roles/:id` (returns 403). System role system permissions MUST NOT be removable via `PUT /api/v1/roles/:id/permissions` (returns 403).
(Previously: Only marked roles as system-level; no explicit protection against edit/delete)

#### Scenario: System flag present

- GIVEN the seeded roles exist
- WHEN a role is retrieved
- THEN `is_system` is `true`

#### Scenario: System role cannot be edited

- GIVEN a system role (`is_system: true`)
- WHEN `PUT /api/v1/roles/:id` is called with any valid payload
- THEN the response returns HTTP 403 and the role is unchanged

#### Scenario: System role cannot be deleted

- GIVEN a system role (`is_system: true`)
- WHEN `DELETE /api/v1/roles/:id` is called
- THEN the response returns HTTP 403 and the role is not deleted
