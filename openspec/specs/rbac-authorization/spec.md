# RBAC Authorization Specification

## Purpose

RequirePermission and RequireRole middleware, PermissionResolver interface, cache-backed lazy loading, and root bypass for endpoint-level authorization.

## Requirements

### Requirement: RequirePermission middleware

The system MUST provide a `RequirePermission(slug)` middleware that checks whether the authenticated user holds the specified permission slug. If the user's role is `root`, the request MUST be allowed regardless of the permission slug. Otherwise, the middleware MUST resolve the user's permission set via the `PermissionResolver` and verify the slug is present. Unauthorized requests MUST return HTTP 403 with the standard DTO error envelope.

#### Scenario: User holds required permission

- GIVEN an authenticated user with permission `users.create`
- WHEN `RequirePermission("users.create")` guards a route
- THEN the request proceeds to the handler

#### Scenario: User lacks required permission

- GIVEN an authenticated user without permission `users.create`
- WHEN `RequirePermission("users.create")` guards a route
- THEN the response returns HTTP 403 with `success: false`

#### Scenario: Root bypasses permission check

- GIVEN an authenticated user with role `root`
- WHEN `RequirePermission("users.create")` guards a route
- THEN the request proceeds regardless of explicit permission assignment

#### Scenario: Unauthenticated request to permission-protected route

- GIVEN a request without a valid access token
- WHEN `RequirePermission("users.create")` guards a route
- THEN the response returns HTTP 401

### Requirement: RequireRole middleware

The system MUST provide a `RequireRole(slug)` middleware that checks whether the authenticated user's role matches the given slug. Unauthorized requests MUST return HTTP 403.

#### Scenario: User has matching role

- GIVEN an authenticated user with role `admin`
- WHEN `RequireRole("admin")` guards a route
- THEN the request proceeds to the handler

#### Scenario: User has different role

- GIVEN an authenticated user with role `user`
- WHEN `RequireRole("admin")` guards a route
- THEN the response returns HTTP 403

### Requirement: PermissionResolver interface

The system MUST define a `PermissionResolver` interface with method `Resolve(public_id string) ([]string, error)`. It MUST load permissions by the user's `public_id`, first checking the cache and falling back to the database. The cache TTL MUST be 5 minutes. Mutations on role-permission assignments MUST invalidate the cache for affected role members.

#### Scenario: Cache hit

- GIVEN a user whose permissions were cached within 5 minutes
- WHEN `Resolve` is called for that user
- THEN permissions are returned from cache without a database query

#### Scenario: Cache miss loads from database

- GIVEN a user whose permissions are not cached
- WHEN `Resolve` is called for that user
- THEN permissions are loaded from the database and stored in cache with 5-min TTL

#### Scenario: Cache invalidation on mutation

- GIVEN a role with cached permissions for its members
- WHEN a permission is assigned to or removed from that role
- THEN the cache entries for all members of that role are invalidated

### Requirement: Module-owned name constants

Each module MUST define its own `Module<Name>` constant in `modules/<name>/core/constants.go`. Route definitions MUST reference the module-local constant, NOT `platform/permissions.Module*`. The `platform/permissions` package MUST NOT export `Module*` constants — it retains only `Action*` constants and utility functions.

#### Scenario: Users module defines ModuleUsers

- GIVEN `modules/users/core/constants.go` exists
- WHEN routes reference the users module name
- THEN they use the local `ModuleUsers` constant

#### Scenario: Roles module defines ModuleRoles

- GIVEN `modules/roles/core/constants.go` exists
- WHEN routes reference the roles module name
- THEN they use the local `ModuleRoles` constant

#### Scenario: Platform permissions has no Module* constants

- GIVEN `platform/permissions/constants.go`
- WHEN the file is reviewed
- THEN it contains zero `Module*` constants — only `Action*` constants and utility functions