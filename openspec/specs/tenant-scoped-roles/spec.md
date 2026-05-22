# Tenant-Scoped Roles Specification

## Purpose

Company-scoped role management with global `root` role, reserved slug protection, and tenant-isolated role queries for SaaS multitenancy.

## Requirements

### Requirement: Company-scoped role model

The system MUST include a nullable `company_id` field on the `roles` table. When `company_id` IS NULL, the role is global (root only). When `company_id` IS NOT NULL, the role belongs to that specific company. The system MUST enforce composite unique constraints on `(slug, company_id)` and `(name, company_id)`. A partial unique index MUST allow only one global role per slug where `company_id IS NULL`.

#### Scenario: Global role has null company_id

- GIVEN the `root` role exists
- WHEN the role is queried
- THEN `company_id` IS NULL

#### Scenario: Tenant role has company_id set

- GIVEN a role created for company with `public_id = "comp-abc"`
- WHEN the role is queried
- THEN `company_id` matches that company's internal ID

#### Scenario: Duplicate slug allowed across companies

- GIVEN company A has a role with slug `manager`
- WHEN company B creates a role with slug `manager`
- THEN the creation succeeds (different `company_id`)

#### Scenario: Duplicate slug rejected within same company

- GIVEN company A has a role with slug `manager`
- WHEN company A attempts to create another role with slug `manager`
- THEN the system returns HTTP 409 conflict

### Requirement: Reserved slug protection

The system MUST NOT allow creation of roles with slugs `root`, `admin`, or `user` via the API. These slugs are reserved for system use. The check MUST be case-insensitive and apply to both `slug` and `name` fields. Reserved slug protection applies regardless of tenant context.

#### Scenario: Reserved slug rejected on create

- GIVEN a user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with slug `root`
- THEN the response returns HTTP 422 with a validation error

#### Scenario: Reserved name rejected on create

- GIVEN a user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with name `Admin`
- THEN the response returns HTTP 422 with a validation error

#### Scenario: Non-reserved slug allowed

- GIVEN a user with `roles.create` permission
- WHEN `POST /api/v1/roles` is called with slug `team-lead`
- THEN the creation proceeds normally

### Requirement: Tenant-isolated role queries

The system MUST scope role queries by `company_id` for non-root users. Root users (`IsRootScope: true`) MUST see all roles globally. Non-root users MUST only see roles belonging to their own company. `GET /api/v1/roles` and `GET /api/v1/roles/:id` MUST enforce this isolation.

#### Scenario: Non-root user lists only own company roles

- GIVEN an admin user with `company_id = 5`
- WHEN `GET /api/v1/roles` is called
- THEN only roles with `company_id = 5` are returned

#### Scenario: Root user lists all roles globally

- GIVEN an authenticated root user
- WHEN `GET /api/v1/roles` is called
- THEN all roles (global and tenant-scoped) are returned

#### Scenario: Non-root user cannot access another company's role

- GIVEN an admin user with `company_id = 1`
- WHEN `GET /api/v1/roles/:id` is called with a role belonging to company 2
- THEN the response returns HTTP 404

#### Scenario: Root user can access any role

- GIVEN an authenticated root user
- WHEN `GET /api/v1/roles/:id` is called with any valid role ID
- THEN the response returns HTTP 200 with the role

### Requirement: Root role protection via API

The system MUST NOT allow editing or deleting the `root` role via any API endpoint. Any attempt to `PUT` or `DELETE` the root role MUST return HTTP 403. This protection is in addition to the existing `is_system` flag protection.

#### Scenario: Root role cannot be edited

- GIVEN the root role exists
- WHEN `PUT /api/v1/roles/:root_id` is called
- THEN the response returns HTTP 403

#### Scenario: Root role cannot be deleted

- GIVEN the root role exists
- WHEN `DELETE /api/v1/roles/:root_id` is called
- THEN the response returns HTTP 403

### Requirement: Role DTO includes company context

The system MUST include `company_id` (as a public ID string, omitempty) in role API responses. When `company_id` IS NULL (global role), the field MUST be omitted from the response. When `company_id` IS NOT NULL, the field MUST contain the company's `public_id`.

#### Scenario: Tenant role response includes company_id

- GIVEN a role with `company_id = 5` and company public_id `comp-abc`
- WHEN the role is retrieved via API
- THEN the response includes `company_id: "comp-abc"`

#### Scenario: Global role response omits company_id

- GIVEN the root role with `company_id = NULL`
- WHEN the root role is retrieved via API
- THEN the response omits the `company_id` field

### Requirement: Seed only root role globally

The system MUST seed only the `root` role on startup or via seed command. The `admin` and `user` roles MUST NOT be seeded globally. The `root` role MUST have `company_id = NULL` and `is_system = true`. The system MUST NOT seed `role_permissions` entries for the `root` role (bypass is handled by middleware).

#### Scenario: Seed creates only root

- GIVEN a fresh database
- WHEN the seed command runs
- THEN only the `root` role exists in the database

#### Scenario: Root has no role_permissions rows

- GIVEN the seed has run
- WHEN the `role_permissions` table is queried for root's role_id
- THEN no rows exist for the root role

#### Scenario: Idempotent root seeding

- GIVEN the root role already exists
- WHEN the seed command runs again
- THEN no duplicate root role is created and the process exits successfully
