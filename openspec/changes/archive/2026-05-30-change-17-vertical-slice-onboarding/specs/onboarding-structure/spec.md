# Delta for Onboarding Structure

## ADDED Requirements

### Requirement: Vertical slice organization for onboarding module

The onboarding module SHALL be reorganized from flat legacy structure to vertical slice architecture. The module root SHALL retain only `container.go` (composition root), `routes.go` (route wiring), and optional compatibility aliases. A single slice `onboard_company/` SHALL co-locate `handler.go`, `service.go`, `repository.go`, and corresponding `_test.go` files. Shared types (models, DTOs, errors, constants) SHALL live in `core/`. Reusable data-access logic SHALL live in `queries/` with one `_test.go` per query file.

#### Scenario: Module root has only cross-cutting files

- GIVEN the onboarding module after migration
- WHEN the module root directory is inspected
- THEN it contains `container.go`, `routes.go`, `core/`, `queries/`, and `onboard_company/`
- AND it does NOT contain `handler.go`, `service.go`, or `dto.go` at the root level

#### Scenario: Slice has all layers co-located

- GIVEN the `onboard_company/` slice directory
- WHEN its contents are inspected
- THEN it contains `handler.go`, `service.go`, `repository.go`, and their `_test.go` files

#### Scenario: Queries have matching test files

- GIVEN each file in `queries/`
- WHEN the directory is inspected
- THEN each query file has a corresponding `_test.go` file

### Requirement: Cross-module model import elimination

The onboarding module SHALL NOT import GORM models or constants from `companies`, `roles`, `users`, or `permissions` modules. Local partial models SHALL be defined in `core/model.go` with `TableName()` overrides targeting the correct database tables. Status, kind, and role-slug constants SHALL be duplicated locally in `core/`. The `users.PasswordHasher` interface MAY remain as an injected contract (acceptable per cross-module communication rules). `shared.BaseModel` MAY remain as an import (platform-level, not module-level).

#### Scenario: No cross-module model imports

- GIVEN the onboarding module source files after migration
- WHEN imports are inspected across all files
- THEN no file imports `internal/modules/companies`, `internal/modules/roles`, `internal/modules/users` (except `PasswordHasher` interface), or `internal/modules/permissions`
- AND `shared.BaseModel` and `users.PasswordHasher` are the only acceptable cross-module type references

#### Scenario: Local partial models target correct tables

- GIVEN each local partial model in `core/model.go`
- WHEN its `TableName()` method is inspected
- THEN it returns the correct database table name matching the original external model

### Requirement: Identical endpoint behavior after migration

The `POST /api/v1/onboarding/companies` endpoint SHALL return identical HTTP status codes, response bodies, and error behaviors after migration. The transactional flow (company → domains → roles → permissions → admin user) SHALL execute in a single database transaction. All validation error mappings (slug, domain, email conflicts) SHALL produce the same field-level error responses.

#### Scenario: Root onboards company — same response

- GIVEN an authenticated root user submits valid onboarding data
- WHEN `POST /api/v1/onboarding/companies` is called after migration
- THEN the response returns HTTP 201 with identical body structure (company_public_id, company_slug, admin_public_id, admin_email)

#### Scenario: Duplicate slug — same error

- GIVEN an existing company with slug `acme`
- WHEN root submits onboarding with slug `acme`
- THEN the response returns a validation error on the `slug` field (same as before migration)

#### Scenario: Transaction rollback on failure

- GIVEN any step in the onboarding transaction fails
- WHEN the transaction completes
- THEN no partial records remain in any affected table (companies, company_domains, roles, users, role_permissions)

### Requirement: Container wiring update

The root `internal/app/container.go` SHALL call `onboarding.NewContainer()` instead of `onboarding.NewService()` + `onboarding.NewHandler()`. The `onboarding.Register()` function SHALL accept the module container instead of a standalone handler. Route paths and HTTP methods SHALL remain unchanged.

#### Scenario: Root container uses module container

- GIVEN `internal/app/container.go` after migration
- WHEN onboarding wiring is inspected
- THEN it calls `onboarding.NewContainer(...)` and passes the container to `onboarding.Register(...)`

#### Scenario: Routes unchanged

- GIVEN the onboarding `routes.go` after migration
- WHEN route registration is inspected
- THEN `POST /api/v1/onboarding/companies` is registered with `requireRole("root")` middleware (same as before)
