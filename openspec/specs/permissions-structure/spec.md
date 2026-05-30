# Permissions Structure Specification

## Purpose

Defines the vertical slice architecture for the permissions module, replacing the flat layout with co-located use-case slices.

## Requirements

### Requirement: Module root contains only cross-cutting files

The module root `internal/modules/permissions/` SHALL retain only wiring files (`container.go`, `routes.go`), the `core/` package, the `queries/` package, slice sub-packages, and route-absence tests. The module root SHALL NOT contain `handler.go`, `service.go`, `repository.go`, `model.go`, or `dto.go`.

#### Scenario: Root flat files are removed

- GIVEN the migration is complete
- WHEN `internal/modules/permissions/` root directory is inspected
- THEN it does NOT contain `handler.go`, `service.go`, `repository.go`, `model.go`, or `dto.go`
- AND it DOES contain `container.go` and `routes.go`

#### Scenario: Core package exists with shared types

- GIVEN the migration is complete
- WHEN the directory is inspected
- THEN `core/model.go`, `core/dto.go`, `core/enums.go`, and `core/contracts.go` exist
- AND slices import from `permissions/core` not from `permissions` root

### Requirement: HTTP use-case slices registered by endpoint

The module SHALL have one vertical slice per registered HTTP endpoint: `list_permissions`, `view_permission`, `update_permission`. Each slice SHALL co-locate `handler.go`, `service.go`, `repository.go`, and `_test.go` files. Slices SHALL NOT import sibling slices.

#### Scenario: Three HTTP slices exist with all layers

- GIVEN the migration is complete
- WHEN each of `list_permissions/`, `view_permission/`, `update_permission/` is inspected
- THEN each directory contains handler, service, repository, and at least one test file
- AND the handler responds to the matching registered route

#### Scenario: No slice imports its siblings

- GIVEN slice `view_permission/repository.go`
- WHEN its imports are inspected
- THEN it does NOT import `list_permissions` or `update_permission` packages

### Requirement: Internal non-HTTP slices for middleware and bootstrap

The module SHALL provide `resolve_permissions/` (for auth middleware `Resolve` method) and `sync_permissions/` (for bootstrap `SyncPermissions` method) as internal slices. Each SHALL co-locate service, repository, and test files. These slices SHALL NOT register HTTP routes.

#### Scenario: ResolvePermissions slice exists

- GIVEN the migration is complete
- WHEN `resolve_permissions/` is inspected
- THEN it contains `service.go`, `repository.go`, and `_test.go`
- AND the service exposes a `Resolve(publicID string)` method compatible with `middleware/auth.go`

#### Scenario: SyncPermissions slice exists

- GIVEN the migration is complete
- WHEN `sync_permissions/` is inspected
- THEN it contains `service.go`, `repository.go`, and `_test.go`
- AND the service exposes a `SyncPermissions(slugs []string)` method callable from `app/container.go`

### Requirement: Module container as composition root

The module SHALL have a `container.go` that instantiates all slice handlers, services, and repositories. The container SHALL expose handlers needed for `routes.go` registration and non-HTTP methods for external callers. The container SHALL NOT contain business logic or act as a service locator.

#### Scenario: Container wires all slices

- GIVEN the module container is instantiated
- WHEN its output struct is inspected
- THEN it exposes all three HTTP handlers plus Resolve and SyncPermissions methods
- AND it constructs them by passing dependencies only (db, cache, logger)

### Requirement: Routes stay at module root

Route registration SHALL remain in `routes.go` at the module root. The `routes.go` file SHALL register handlers exposed by the module container (not the old flat handler). Route paths, HTTP methods, and middleware chains SHALL NOT change as a result of the reorganization.

#### Scenario: Identical endpoints after migration

- GIVEN the migration is complete and server starts
- WHEN the registered routes are inspected via `GET /api/v1/permissions` et al.
- THEN the same 3 registered endpoints respond identically (status codes, bodies, errors) before and after migration

#### Scenario: Unregistered endpoints preserved but unwired

- GIVEN the migration is complete
- WHEN source code is inspected
- THEN unregistered handler methods (`ListPaginated`, `Create`, `Delete`) still exist somewhere accessible for future wiring
- AND they are NOT registered in `routes.go`
