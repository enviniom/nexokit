# Permissions Specification

## Purpose

Permission model and admin CRUD for fine-grained access control slugs with UI-friendly metadata.

## Requirements

### Requirement: Permission model

The system MUST define a `Permission` model with fields `id`, `public_id`, `slug` (unique), `name`, `module`, `action`, `description`, `is_system`, `display_order`, `created_at`, `updated_at`. Slugs MUST follow `{module}.{action}` (e.g., `users.create`). The `action` field MUST use explicit actions (`index`, `view`, `list`, `create`, `update`, `delete`) or business actions (e.g., `change_role`, `assign_permissions`) — the system MUST NOT use ambiguous terms like `read`. The `display_order` field MUST control UI rendering order. The model MUST have a unique index on `slug`.

#### Scenario: Valid permission slug

- GIVEN a permission with slug `users.create` and action `create`
- WHEN the permission is persisted
- THEN it is stored and retrievable by slug

#### Scenario: Duplicate slug rejected

- GIVEN a permission with slug `users.create` already exists
- WHEN another permission with the same slug is created
- THEN the system rejects the creation with a unique constraint error

#### Scenario: Business action slug

- GIVEN a permission with slug `users.change_role`, module `users`, action `change_role`
- WHEN the permission is persisted
- THEN it is stored with the correct module and action values

### Requirement: Admin CRUD endpoints

The system MUST expose admin-only endpoints:

| Method | Path | Action |
|--------|------|--------|
| GET | `/api/v1/permissions` | List all permissions |
| GET | `/api/v1/permissions/:id` | Get permission by public_id |
| POST | `/api/v1/permissions` | Create a non-system permission |
| PUT | `/api/v1/permissions/:id` | Update a non-system permission |
| DELETE | `/api/v1/permissions/:id` | Delete a non-system permission |

Endpoints MUST use the standard DTO envelope except successful DELETE responses, which MUST return HTTP 204 with no body. List endpoints MUST return permissions grouped by `module`, sorted by `display_order` within each module. Mutation endpoints MUST require `permissions.manage`. The system MUST NOT allow creation, update, or deletion of system permissions (`is_system: true`).

#### Scenario: List permissions grouped and sorted

- GIVEN multiple permissions exist across modules
- WHEN `GET /api/v1/permissions` is called with valid auth
- THEN the response returns HTTP 200 with `data` grouped by `module`, each group sorted by `display_order`

#### Scenario: Create non-system permission

- GIVEN a user with `permissions.manage` permission
- WHEN `POST /api/v1/permissions` is called with valid `slug`, `name`, `module`, `action`
- THEN the response returns HTTP 201 with the created permission

#### Scenario: Delete non-system permission

- GIVEN a non-system permission exists and the user has `permissions.manage`
- WHEN `DELETE /api/v1/permissions/:id` is called
- THEN the response returns HTTP 204 with an empty body
- AND the permission is deleted

#### Scenario: Reject mutation on system permission

- GIVEN a system permission (`is_system: true`)
- WHEN a PUT or DELETE is attempted on that permission
- THEN the response returns HTTP 403 with an appropriate error

#### Scenario: Unauthorized access

- GIVEN a user without `permissions.manage` permission
- WHEN a mutation endpoint is called
- THEN the response returns HTTP 403

### Requirement: Permission seeds

The system MUST seed base permissions on startup or via seed command using explicit actions:

| Module | Actions | Business Actions |
|--------|---------|------------------|
| users | index, view, create, update, delete | change_role |
| roles | index, view, create, update, delete | assign_permissions |
| companies | index, view, create, update, delete | — |
| settings | view, update | — |
| auth | view | — |
| permissions | manage | — |

Slugs follow `{module}.{action}` (e.g., `users.change_role`, `roles.assign_permissions`). The operation MUST be idempotent. All seeded permissions MUST have `is_system: true` and explicit `display_order` values.

#### Scenario: Idempotent seeding

- GIVEN permissions have already been seeded
- WHEN the seed command runs again
- THEN no duplicate permissions are created and the process exits successfully