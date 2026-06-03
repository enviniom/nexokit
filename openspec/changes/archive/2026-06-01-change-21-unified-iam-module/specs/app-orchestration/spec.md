# Delta for App Orchestration

## ADDED Requirements

### Requirement: IAM container wiring

The system SHALL replace `usersHandler`, `rolesHandler`, and `permissionsContainer` fields on the root `Container` with a single `IAM *iam.Container` field. The root container SHALL build the IAM container via `iam.NewContainer(db, cache, logger)` during bootstrap.

#### Scenario: IAM container present after bootstrap

- GIVEN bootstrap completes successfully
- WHEN `app.Container.IAM` is inspected
- THEN it is non-nil and contains `Users`, `Roles`, and `Permissions` sub-containers

#### Scenario: Legacy handler fields removed

- GIVEN the root container is built
- WHEN `app/container.go` is reviewed
- THEN it contains zero references to `usersHandler`, `rolesHandler`, or `permissionsContainer`

## MODIFIED Requirements

### Requirement: Dependency container

The system SHALL provide a `Container` type that wires repositories, services, and handlers, and is built during bootstrap. The root container SHALL delegate module wiring to module-level `NewContainer(db)` functions. The root container SHALL replace `usersHandler`, `rolesHandler`, and `permissionsContainer` with a single `IAM *iam.Container` field. The root container MUST NOT instantiate individual repositories, services, or handlers for modules using vertical slices.
(Previously: Root container held separate `usersHandler`, `rolesHandler`, and `permissionsContainer` fields wired from three distinct modules)

#### Scenario: Container wiring via module containers

- GIVEN bootstrap succeeds
- WHEN `app.Container` is inspected
- THEN it contains module-level containers (e.g., `CompaniesContainer`, `IAM`) returned by each module's `NewContainer(db)`
- AND it does NOT contain individual handler/service/repository fields for migrated modules

#### Scenario: Root container imports module root only

- GIVEN the IAM module uses multi-entity vertical slices
- WHEN `internal/app/container.go` imports IAM wiring
- THEN it imports the root `internal/modules/iam` package
- AND it does NOT import IAM entity or slice packages

#### Scenario: Module container is called by root

- GIVEN the root container is being built
- WHEN wiring the IAM module
- THEN `iam.NewContainer(db, cache, logger)` is called
- AND the returned container is stored as `c.IAM` on the root container

### Requirement: RegisterModules mounts IAM only

The system SHALL mount IAM routes via a single `iam.Register(globalProtected, c.IAM, tenantProtected, middleware.RequirePermission, middleware.RequireRole)` call. Legacy module registrations for `users`, `roles`, and `permissions` SHALL be removed from `RegisterModules`. Legacy modules SHALL remain compilable but unreachable at runtime.

#### Scenario: IAM routes mounted

- GIVEN `RegisterModules` is called during bootstrap
- WHEN the router is inspected
- THEN all 19 IAM endpoints respond at their expected `/api/v1/*` paths

#### Scenario: Legacy routes not mounted

- GIVEN `RegisterModules` is called
- WHEN a legacy route path is requested
- THEN the response returns HTTP 404 (route not registered)

#### Scenario: Legacy modules still compile

- GIVEN the app container wires IAM instead of legacy modules
- WHEN `go build ./internal/modules/users/... ./internal/modules/roles/... ./internal/modules/permissions/...` is run
- THEN all three legacy modules compile successfully
