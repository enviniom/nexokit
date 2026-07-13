# Vertical Slice Modules Specification

## Purpose

Defines the vertical slice architecture pattern within modules. Each use case co-locates its handler, service, repository, and tests. Root module packages keep route/container wiring only. Shared non-query concerns live in module-local `core/`; reusable DB query logic lives in `queries/` with one `_test.go` per query file. Module `container.go` acts as composition root. Root container delegates to module containers only.

## Requirements

### Requirement: Module root structure

The module root SHALL retain only cross-cutting wiring files such as `routes.go` and `container.go` (plus compatibility aliases, resolver, and route-absence tests when protecting module-level behavior). Shared models, DTOs/contracts, enums/constants, errors, and shared values with no direct testable logic SHALL live in a module-local `core/` package. The module root SHALL NOT contain `handler.go`, `service.go`, or `repository.go` files for individual use cases, and slices SHALL NOT import the root module package.

#### Scenario: Module root has only cross-cutting files

- GIVEN a module using vertical slices
- WHEN the module root directory is inspected
- THEN it contains route/container wiring and optional root-level compatibility/resolver files only
- AND it does NOT contain `handler.go`, `service.go`, or `repository.go`

#### Scenario: Core types are shared not duplicated

- GIVEN multiple use-case slices in the same module
- WHEN any slice needs a model type
- THEN it imports the model from the module-local `core` package
- AND no slice defines its own copy of the model

#### Scenario: Slice imports avoid root cycle

- GIVEN root `companies` imports slice packages for container and route wiring
- WHEN a slice needs shared DTOs, models, or contracts
- THEN it imports `internal/modules/companies/core`
- AND it does NOT import `internal/modules/companies`

### Requirement: Query reuse package

Reusable DB lookup/count/query logic duplicated across slice repositories SHALL be extracted to `queries/`. Each query SHOULD be focused, one file per query when practical, and each query file SHALL have a corresponding `_test.go` file.

#### Scenario: Query logic is extracted and tested

- GIVEN repeated lookup logic exists across slice repositories
- WHEN that logic is refactored
- THEN shared query functions live under `internal/modules/companies/queries`
- AND each query file has a matching `_test.go`

### Requirement: Use-case slice structure

Each use case SHALL be a sub-package containing its own `handler.go`, `service.go`, `repository.go`, and associated test files. Each slice's repository SHALL own only the data access methods it needs. Each slice SHALL NOT import sibling slices within the same module.

#### Scenario: Slice has all layers co-located

- GIVEN a use case `list_companies` within the companies module
- WHEN the `list_companies/` directory is inspected
- THEN it contains `handler.go`, `service.go`, `repository.go`, and test files
- AND all three layers reference each other via local imports

#### Scenario: Slice does not import sibling slices

- GIVEN two slices `list_companies/` and `view_company/` within the same module
- WHEN `list_companies/` source files are inspected
- THEN they do NOT import `view_company/` or any other sibling slice

#### Scenario: Slice repository owns only needed methods

- GIVEN a slice `view_company/` that only reads a single entity
- WHEN its `repository.go` is inspected
- THEN it defines only the query methods needed for that use case
- AND it does NOT define methods belonging to other slices

### Requirement: Module container as composition root

Each module SHALL have a `container.go` that instantiates all slice handlers and exposes them for route registration. The module container SHALL NOT contain business logic. The module container SHALL NOT act as a service locator — it instantiates, wires, and exposes only.

#### Scenario: Module container wires slices

- GIVEN a module with multiple use-case slices
- WHEN `NewContainer(db)` is called
- THEN it instantiates each slice's handler, service, and repository
- AND it returns a struct exposing the handlers needed for routing

#### Scenario: Module container has no business logic

- GIVEN a module `container.go`
- WHEN its source is inspected
- THEN it contains only constructor calls, dependency passing, and struct initialization
- AND it does NOT contain validation, transformation, or query logic

#### Scenario: Module container is not a service locator

- GIVEN a module `container.go`
- WHEN its source is inspected
- THEN it does NOT expose generic `GetService()` or `GetRepository()` methods
- AND it exposes only concrete handler references

### Requirement: Root container delegates to module containers

The root `container.go` SHALL call module-level `NewContainer(db)` functions instead of wiring individual repositories, services, and handlers per module. The root container SHALL NOT import individual slice packages from any module.

#### Scenario: Root container calls module container

- GIVEN the root `container.go` in `internal/app/`
- WHEN wiring the companies module
- THEN it calls `companies.NewContainer(db)`
- AND it does NOT instantiate individual companies handlers, services, or repositories

#### Scenario: Root container does not know slices

- GIVEN the root `container.go`
- WHEN its imports are inspected
- THEN it imports only the module root package (e.g., `internal/modules/companies`)
- AND it does NOT import any slice sub-packages (e.g., `internal/modules/companies/list_companies`)

### Requirement: Incremental migration pattern

Existing modules SHALL NOT be mass-migrated unless already undergoing substantial change. New or non-trivial modules SHALL use the vertical slice pattern from the start. Simple modules MAY remain flat. The `companies` module SHALL be the pilot migration target. The `auth` module SHALL be the second migration target, using `companies` as the reference pattern.

#### Scenario: Companies module is migrated

- GIVEN the `companies` module currently uses flat files
- WHEN the migration is complete
- THEN `companies` uses vertical slice organization with 7 slices: `list_companies`, `view_company`, `update_company`, `delete_company`, `list_company_domains`, `create_company_domain`, `update_company_domain`
- AND it has no `create_company` slice

#### Scenario: Auth module is migrated

- GIVEN the `auth` module currently uses flat files
- WHEN the migration is complete
- THEN `auth` uses vertical slice organization with 4 slices: `authenticate_user`, `rotate_token`, `revoke_token`, `view_session`
- AND it has `core/` with local partial models and `queries/` with reusable queries

#### Scenario: Remaining modules unchanged

- GIVEN modules `users`, `roles`, `permissions`, `onboarding` exist with flat structure
- WHEN the auth migration is complete
- THEN these modules retain their existing flat file structure
- AND they continue to function without modification

#### Scenario: New modules use vertical slices

- GIVEN a new module is created after this pattern is established
- WHEN the module scaffold is generated
- THEN it uses vertical slice organization by default

### Requirement: Auth module migration

The `auth` module SHALL be the second migration target to vertical slice architecture, following the `companies` module pattern. It SHALL have 4 slices: `authenticate_user`, `rotate_token`, `revoke_token`, `view_session`. Each slice SHALL co-locate handler, service, repository, and tests. The auth module SHALL eliminate all cross-module imports to `modules/users` and `modules/roles`, using local partial models in `core/` and reusable queries in `queries/`.

#### Scenario: Auth module has 4 slices

- GIVEN the auth module migration is complete
- WHEN the `internal/modules/auth/` directory is inspected
- THEN it contains `authenticate_user/`, `rotate_token/`, `revoke_token/`, `view_session/` subdirectories
- AND each contains handler, service, repository, and test files

#### Scenario: Auth module has no cross-module imports

- GIVEN the auth module migration is complete
- WHEN auth source files are inspected for imports
- THEN no file imports `modules/users` or `modules/roles`
- AND user data is accessed via local partial models in `auth/core/`

#### Scenario: Auth module uses queries/ package

- GIVEN the auth module migration is complete
- WHEN shared DB lookup logic is needed by multiple slices
- THEN it lives in `auth/queries/` with a corresponding `_test.go` per query file

### Requirement: Routes stay at module root

Route registration SHALL remain in `routes.go` at the module root. The `routes.go` file SHALL register handlers exposed by the module container. Route paths and HTTP methods SHALL NOT change as a result of the slice reorganization.

#### Scenario: Routes register slice handlers

- GIVEN a module `routes.go`
- WHEN route registration executes
- THEN it calls handlers exposed by the module container
- AND the same URL paths and HTTP methods are registered as before

### Requirement: Companies module uses slice-root rehome

The companies module MUST keep its seven use cases under `internal/modules/companies/slices/`: `list_companies`, `view_company`, `update_company`, `delete_company`, `list_company_domains`, `create_company_domain`, and `update_company_domain`. The module root MUST remain wiring-only, with container, routes, resolver, and compatibility aliases only. Public HTTP behavior MUST NOT change because of the rehome.
(Previously: the companies module still used flat root slice packages.)

#### Scenario: Companies is rehomed

- GIVEN the companies module tree is inspected
- WHEN its directories are listed
- THEN the seven slices are under `slices/`
- AND the root contains wiring only

#### Scenario: Other modules stay unchanged

- GIVEN the codebase after the move
- WHEN non-companies modules are inspected
- THEN their layouts are unchanged
- AND no new `create_company` slice appears
