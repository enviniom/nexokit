# Delta for Roles

## ADDED Requirements

### Requirement: Reserved slug validation

The system MUST reject role creation and update requests that use reserved slugs: `root`, `admin`, `user`. The validation MUST be case-insensitive and apply to both `slug` and `name` fields. Reserved slug rejection MUST occur before any database operation and return HTTP 422.

#### Scenario: Reserved slug rejected on create

- GIVEN a user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with slug `admin`
- THEN the response returns HTTP 422 with a validation error

#### Scenario: Reserved name rejected on create

- GIVEN a user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with name `Root`
- THEN the response returns HTTP 422 with a validation error

#### Scenario: Reserved slug rejected on update

- GIVEN a custom role exists
- WHEN `PUT /api/v1/roles/:id` is called with slug `user`
- THEN the response returns HTTP 422 with a validation error

## MODIFIED Requirements

### Requirement: Role CRUD API

The system MUST expose `POST /api/v1/roles`, `PUT /api/v1/roles/:id`, and `DELETE /api/v1/roles/:id` in addition to existing GET endpoints. POST and PUT MUST require `roles.create` and `roles.update` permissions respectively. DELETE MUST require `roles.delete` permission. All mutation endpoints MUST reject operations on system roles (`is_system: true`) with HTTP 403. All mutation endpoints MUST reject operations using reserved slugs (`root`, `admin`, `user`) with HTTP 422. Request bodies MUST include validated `slug` fields following the `ValidSlug()` rule (lowercase alphanumeric with hyphens, no leading/trailing hyphens).
(Previously: Rejected only system roles and invalid slug formats; did not validate against reserved slugs)

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

### Requirement: Role seeds

The system MUST seed only the `root` role on startup or via seed command. The `root` role MUST have `company_id = NULL` and `is_system: true`. The `admin` and `user` roles MUST NOT be seeded (they will be created per-company during onboarding). The system MUST NOT seed `role_permissions` entries for the `root` role (root bypass is handled at middleware level via the `"*"` permission marker).
(Previously: Seeded root, admin, and user roles with role_permissions assignments for all three)

#### Scenario: Idempotent seeding

- GIVEN the root role has already been seeded
- WHEN the seed command runs again
- THEN no duplicate root role is created and the process exits successfully

#### Scenario: Root role has null company_id

- GIVEN the system seeds have run
- WHEN the root role is retrieved
- THEN `company_id` IS NULL

#### Scenario: Admin and user roles not seeded

- GIVEN the system seeds have run
- WHEN all roles are queried
- THEN only `root` exists (no `admin` or `user` roles)

#### Scenario: No role_permissions for root

- GIVEN the system seeds have run
- WHEN the `role_permissions` table is queried for the root role
- THEN no permission assignments exist for root

### Requirement: System role protection

The system MUST mark seeded roles as system-level (`is_system: true`) to distinguish them from custom roles. System roles MUST NOT be editable via `PUT /api/v1/roles/:id` (returns 403). System roles MUST NOT be deletable via `DELETE /api/v1/roles/:id` (returns 403). The `root` role MUST receive additional protection: it MUST NOT be editable or deletable via any API endpoint, regardless of `is_system` flag. System role system permissions MUST NOT be removable via `PUT /api/v1/roles/:id/permissions` (returns 403).
(Previously: Protected all system roles equally; root had no explicit additional protection)

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

#### Scenario: Root role explicitly protected from edit

- GIVEN the root role exists
- WHEN `PUT /api/v1/roles/:root_id` is called
- THEN the response returns HTTP 403 and the role is unchanged

#### Scenario: Root role explicitly protected from delete

- GIVEN the root role exists
- WHEN `DELETE /api/v1/roles/:root_id` is called
- THEN the response returns HTTP 403 and the role is not deleted
