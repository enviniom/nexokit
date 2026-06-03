# IAM Module Specification

## Purpose

Unified module covering user CRUD, role CRUD, permission CRUD, auth resolution, permission resolution (cache-backed), permission sync, and role-permission assignment. Eliminates cross-module imports via partial local models.

## Requirements

### Requirement: IAM user endpoints

The system SHALL expose all existing user endpoints with identical routes, payloads, status codes, and tenant-scoping behavior:

| Method | Path | Permission |
|--------|------|------------|
| GET | `/api/v1/users` | `users.list` |
| POST | `/api/v1/users` | `users.create` |
| GET | `/api/v1/users/:id` | `users.view` |
| PUT | `/api/v1/users/:id` | `users.update` |
| DELETE | `/api/v1/users/:id` | `users.delete` |
| PATCH | `/api/v1/users/:id/password` | `users.update` |
| PATCH | `/api/v1/users/:id/role` | `users.change_role` |
| PATCH | `/api/v1/users/:id/status` | `users.update` |

Root users with `IsRootScope=true` see all users; tenant-scoped users see only their company's users. Response DTOs SHALL NOT include `password` or `password_hash`. DELETE SHALL return HTTP 204 with no body.

#### Scenario: Create user with company_id

- GIVEN valid user data with `company_id`
- WHEN `POST /api/v1/users` is called
- THEN response returns HTTP 201 with created user, password fields excluded

#### Scenario: Admin sees only own company's users

- GIVEN admin with `company_id = 1` and users in companies 1 and 2
- WHEN `GET /api/v1/users` is called with that admin's TenantContext
- THEN only users where `company_id = 1` are returned

#### Scenario: Cross-tenant update returns 404

- GIVEN admin with `company_id = 1` targeting a user with `company_id = 2`
- WHEN `PUT /api/v1/users/:id` is called
- THEN response returns HTTP 404

#### Scenario: Delete returns 204 with empty body

- GIVEN an existing user within the requester's tenant scope
- WHEN `DELETE /api/v1/users/:id` is called
- THEN response returns HTTP 204 with an empty body and the user is soft-deleted

### Requirement: IAM role endpoints

The system SHALL expose all existing role endpoints with identical routes, payloads, status codes, and behavior:

| Method | Path | Permission |
|--------|------|------------|
| GET | `/api/v1/roles` | `roles.list` |
| GET | `/api/v1/roles/select` | `roles.select` |
| GET | `/api/v1/roles/:id` | `roles.view` |
| POST | `/api/v1/roles` | `roles.create` |
| PUT | `/api/v1/roles/:id` | `roles.update` |
| DELETE | `/api/v1/roles/:id` | `roles.delete` |
| GET | `/api/v1/roles/:id/permissions` | `roles.view` |
| PUT | `/api/v1/roles/:id/permissions` | `roles.assign_permissions` |

System roles (`is_system: true`) SHALL NOT be editable or deletable (HTTP 403). Reserved slugs (`root`, `admin`, `user`) SHALL be rejected with HTTP 422. Role permission assignment SHALL invalidate cache for all role members.

#### Scenario: Create custom role

- GIVEN an authenticated user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with valid `name`, `slug`, and optional `description`
- THEN the response returns HTTP 201 with the created role (`is_system: false`)

#### Scenario: Delete blocked by assigned users

- GIVEN a custom role with one or more assigned users
- WHEN `DELETE /api/v1/roles/:id` is called
- THEN the response returns HTTP 422 with message "role has assigned users"

#### Scenario: Admin role permission assignment is forbidden

- GIVEN a tenant `admin` role with assigned permissions
- WHEN `PUT /api/v1/roles/:id/permissions` is called for that role
- THEN the response returns HTTP 403 and the role's permissions remain unchanged

#### Scenario: Cache invalidation on permission assignment

- GIVEN a role with cached permissions for its members
- WHEN `PUT /api/v1/roles/:id/permissions` succeeds
- THEN the permission cache for all members of that role is invalidated

### Requirement: IAM permission endpoints

The system SHALL expose all existing permission endpoints with identical routes, payloads, status codes, and behavior:

| Method | Path | Permission |
|--------|------|------------|
| GET | `/api/v1/permissions` | `permissions.manage` |
| GET | `/api/v1/permissions/:id` | `permissions.manage` |
| PUT | `/api/v1/permissions/:id` | `permissions.manage` |

List SHALL return permissions grouped by `module`, sorted by `display_order`. System permissions (`is_system: true`) SHALL NOT be mutable (HTTP 403).

#### Scenario: List permissions grouped and sorted

- GIVEN multiple permissions exist across modules
- WHEN `GET /api/v1/permissions` is called with `permissions.manage`
- THEN the response returns HTTP 200 with `data` grouped by `module`, each group sorted by `display_order`

#### Scenario: Reject mutation on system permission

- GIVEN a system permission (`is_system: true`)
- WHEN a PUT is attempted on that permission
- THEN the response returns HTTP 403 with an appropriate error

### Requirement: Auth user resolution

The system SHALL provide an internal `ResolveAuthUser(publicID string) (*authctx.User, error)` contract that returns a user with role and permissions populated for auth middleware. The implementation SHALL use partial IAM models and SHALL NOT import legacy module packages.

#### Scenario: Resolve valid user

- GIVEN a valid user publicID exists
- WHEN `ResolveAuthUser` is called
- THEN a `*authctx.User` is returned with role slug and company_id populated

#### Scenario: Resolve non-existent user

- GIVEN a publicID that does not match any user
- WHEN `ResolveAuthUser` is called
- THEN an error is returned indicating user not found

### Requirement: Permission resolution with cache

The system SHALL provide an internal `ResolvePermissions(publicID string) ([]string, error)` contract that returns permission slugs for a user, using cache-first strategy with 5-minute TTL and database fallback. The contract SHALL match the existing `PermissionResolver` interface signature.

#### Scenario: Cache hit returns without DB query

- GIVEN a user whose permissions were cached within 5 minutes
- WHEN `ResolvePermissions` is called
- THEN permissions are returned from cache without a database query

#### Scenario: Cache miss loads from database

- GIVEN a user whose permissions are not cached
- WHEN `ResolvePermissions` is called
- THEN permissions are loaded from the database and stored in cache with 5-min TTL

### Requirement: Permission synchronization

The system SHALL provide an internal `SyncPermissions(slugs []string) error` contract that synchronizes registered permission slugs at bootstrap time. Newly created system permissions SHALL be automatically assigned to all existing tenant `admin` roles. The operation SHALL be idempotent.

#### Scenario: Newly synced permission assigned to tenant admins

- GIVEN an existing tenant `admin` role and a permission slug not yet in the database
- WHEN `SyncPermissions` is called with that slug
- THEN the permission is created and assigned to the tenant `admin` role

#### Scenario: Existing permission is not reassigned

- GIVEN a permission slug already exists in the database
- WHEN `SyncPermissions` runs again
- THEN no duplicate permission assignment is created

### Requirement: Zero cross-module imports

The IAM module SHALL NOT import any package under `internal/modules/users/`, `internal/modules/roles/`, `internal/modules/permissions/`, or `internal/modules/companies/`. All related entity data SHALL be accessed via partial local models defined in `iam/core/model.go`.

#### Scenario: IAM compiles without legacy imports

- GIVEN the IAM module source code
- WHEN `go list` is run on `internal/modules/iam/...`
- THEN no import path contains `internal/modules/users`, `internal/modules/roles`, `internal/modules/permissions`, or `internal/modules/companies`

### Requirement: No residual legacy references

The IAM module and all production code SHALL contain zero import paths referencing `internal/modules/users/`, `internal/modules/roles/`, or `internal/modules/permissions/`. All user, role, and permission types SHALL be sourced exclusively from IAM's local models in `iam/core/model.go`.

#### Scenario: Zero legacy imports in production code

- GIVEN the full production codebase
- WHEN `go list ./...` is run
- THEN no package import path contains `internal/modules/users`, `internal/modules/roles`, or `internal/modules/permissions`

#### Scenario: Zero legacy imports in test infrastructure

- GIVEN all test files under `tests/`
- WHEN imports are reviewed
- THEN no test file imports `internal/modules/users`, `internal/modules/roles`, or `internal/modules/permissions`
