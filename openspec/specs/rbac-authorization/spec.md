# RBAC Authorization Specification

## Purpose

RequirePermission and RequireRole middleware, PermissionResolver interface, cache-backed lazy loading, and root bypass for endpoint-level authorization.

## Requirements

### Requirement: RequirePermission middleware

The system SHALL provide a `RequirePermission(slug)` middleware that checks whether the authenticated user holds the specified permission slug. If the user's role is `root`, the request MUST be allowed regardless of the permission slug. Otherwise, the middleware SHALL resolve the user's permission set via the `PermissionResolver` (now delegated to IAM's internal resolver) and verify the slug is present. Unauthorized requests SHALL return HTTP 403 with the standard DTO error envelope. The interface contract and behavior SHALL remain unchanged — only the underlying implementation source changes from `permissions.Resolver` to IAM's equivalent.
(Previously: PermissionResolver backed by `permissions.Container.Resolver`; now backed by IAM's internal resolver)

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

The system SHALL provide a `RequireRole(slug)` middleware that checks whether the authenticated user's role from `authctx.User` matches the given slug. IAM SHALL populate that authenticated user via `ResolveAuthUser`; the app-facing `ResolveRoleBySlug` adapter remains available for bootstrap and compatibility contracts. Unauthorized requests SHALL return HTTP 403.
(Previously: Role resolution backed by `roles.Repository.GetBySlug`; now backed by IAM's internal resolver)

#### Scenario: User has matching role

- GIVEN an authenticated user with role `admin`
- WHEN `RequireRole("admin")` guards a route
- THEN the request proceeds to the handler

#### Scenario: User has different role

- GIVEN an authenticated user with role `user`
- WHEN `RequireRole("admin")` guards a route
- THEN the response returns HTTP 403

### Requirement: PermissionResolver interface

The system SHALL define a `PermissionResolver` interface with method `Resolve(public_id string) ([]string, error)`. The implementation SHALL delegate to IAM's internal `ResolvePermissions` contract. It SHALL load permissions by the user's `public_id`, first checking the cache and falling back to the database. The cache TTL SHALL be 5 minutes. Mutations on role-permission assignments SHALL invalidate the cache for affected role members. The interface signature and cache behavior SHALL remain unchanged.
(Previously: Implementation lived in `permissions.Container.Resolver`; now delegated to IAM)

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

### Requirement: AuthUserLookup interface

The system SHALL define an `AuthUserLookup` interface with method `GetAuthUser(public_id string) (*authctx.User, error)`. The implementation SHALL delegate to IAM's internal `ResolveAuthUser` contract. The interface contract, return type, and error behavior SHALL remain unchanged.
(Previously: Implementation wrapped `users.Repository.GetAuthUser`; now delegated to IAM)

#### Scenario: Auth middleware resolves user via IAM

- GIVEN a valid JWT with user publicID
- WHEN `middleware.Auth` calls `AuthUserLookup.GetAuthUser`
- THEN the user is resolved via IAM's internal resolver and attached to request context

#### Scenario: Auth middleware rejects unknown user

- GIVEN a JWT with a non-existent user publicID
- WHEN `middleware.Auth` calls `AuthUserLookup.GetAuthUser`
- THEN an error is returned and the request is rejected with HTTP 401

### Requirement: Adapter delegation to IAM

The system SHALL update middleware adapters in `internal/app/container.go` so that `userLookup` delegates to `c.IAM.ResolveAuthUser`, `roleResolverAdapter` delegates to `c.IAM.ResolveRoleBySlug`, and `SyncPermissions` delegates to `c.IAM.SyncPermissions`. Adapter method signatures SHALL remain compatible with existing middleware contracts.

#### Scenario: Auth adapter uses IAM

- GIVEN the app container is built with IAM
- WHEN `userLookup.GetAuthUser` is called
- THEN the call is forwarded to `c.IAM.ResolveAuthUser`

#### Scenario: SyncPermissions uses IAM

- GIVEN the app container is built with IAM
- WHEN `SyncPermissions` is called during bootstrap
- THEN the call is forwarded to `c.IAM.SyncPermissions`

### Requirement: Module-owned name constants

Each module MUST define its own `Module<Name>` constant in `modules/<name>/core/constants.go`. Route definitions MUST reference the module-local constant, NOT `platform/permissions.Module*`. The `platform/permissions` package MUST NOT export `Module*` constants — it retains only `Action*` constants and utility functions. After legacy removal, the IAM module is the sole owner of user, role, and permission domain constants.
(Previously: Scenarios referenced `modules/users/core/constants.go` and `modules/roles/core/constants.go`; those directories no longer exist — IAM owns these constants)

#### Scenario: IAM module defines domain constants

- GIVEN `modules/iam/core/constants.go` exists
- WHEN routes reference the users, roles, or permissions module name
- THEN they use constants defined within the IAM module

#### Scenario: Platform permissions has no Module* constants

- GIVEN `platform/permissions/constants.go`
- WHEN the file is reviewed
- THEN it contains zero `Module*` constants — only `Action*` constants and utility functions

#### Scenario: No legacy module constant references

- GIVEN the full production codebase after legacy removal
- WHEN imports and constant references are reviewed
- THEN zero references to `modules/users/core`, `modules/roles/core`, or `modules/permissions/core` exist
