# Proposal: RBAC — Permissions, Role-Permissions, and Authorization Middleware

## Intent

NexoKit authenticates but cannot authorize — all authenticated users access all routes equally. This change introduces RBAC: permissions with UI-friendly metadata, role-permission associations, role-permission administration endpoints, and authorization middleware for fine-grained endpoint access control. The API must serve permission data in a shape a future UI can render without transformation.

## Scope

### In Scope
- Permission model with fields: `module`, `action`, `slug`, `name`, `description`, `is_system`, `display_order`
- Admin permission management: CRUD endpoints + seeds for base modules
- Role-permission administration API: `GET /api/v1/roles/:id/permissions` (grouped catalog with granted flags) and `PUT /api/v1/roles/:id/permissions` (slug replacement)
- Permission actions: `index`, `view`, `list`, `create`, `update`, `delete`, plus business actions (`users.change_role`, `roles.assign_permissions`)
- `RequirePermission` and `RequireRole` middleware with cache-backed lazy loading
- `authctx.User` extension with `Permissions []string`
- `/auth/me` response including permissions
- Root bypass + root seeded with all permissions
- Migration: `20260518XXXXXX_rbac.sql`

### Out of Scope
- Frontend UI (project is API-only; API shape supports future UI)
- Custom roles (future change)
- Multi-role per user (architecture stays single-role)
- Individual permission revocation/inactivation

## Capabilities

### New Capabilities
- `permissions`: Permission model (with `module`, `action`, `slug`, `name`, `description`, `is_system`, `display_order`), repository, service, handler, DTOs, CRUD routes, seeds for base modules and business actions
- `rbac-authorization`: RequirePermission and RequireRole middleware, PermissionResolver interface, cache-backed lazy loading, root bypass

### Modified Capabilities
- `auth`: Me endpoint MUST return user permissions alongside role
- `roles`: Role responses MUST include associated permission slugs; add `GET /:id/permissions` (grouped catalog with granted flags) and `PUT /:id/permissions` (replace permission slugs)
- `middleware-auth`: `authctx.User` struct gains `Permissions []string`; authorization middleware populates it

## Approach

Lazy loading with cache backing. Auth middleware resolves identity only. `RequirePermission` fetches and caches the user's permission set by `public_id` (5-min TTL, invalidation on role-permission mutations). Root role bypasses all checks. `RequireRole` checks `authctx.User.Role` slug.

`GET /roles/:id/permissions` returns the full permission catalog grouped by module, each permission annotated with a `granted` boolean for that role — zero N+1, UI-ready.

`PUT /roles/:id/permissions` accepts an array of permission slugs and replaces the role's assignments. System permissions cannot be removed from system roles.

Permission seeds use explicit actions (`index`, `view`, `list`, `create`, `update`, `delete`) and business actions (`users.change_role`, `roles.assign_permissions`) instead of ambiguous `read`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/permissions/` | New | Permission module (model, repo, service, handler, dto, routes, seeds) |
| `internal/middleware/authorization.go` | New | RequirePermission + RequireRole |
| `internal/modules/roles/handler.go`, `routes.go` | Modified | Add role-permission catalog + assignment endpoints |
| `internal/modules/roles/model.go` | Modified | Permissions has-many via RolePermissions |
| `internal/modules/auth/handler.go`, `dto.go` | Modified | Me includes permissions |
| `internal/platform/authctx/authctx.go` | Modified | Add `Permissions []string` |
| `internal/app/container.go` | Modified | Wire permission module + authz MW |
| `migrations/20260518XXXXXX_rbac.sql` | New | permissions + role_permissions tables |
| `seeds/permissions.go`, `seeds/role_permissions.go` | New | Base permission + assignment seeds |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Cache staleness | Medium | 5-min TTL + invalidation on mutations |
| Missing authctx.User call sites | Medium | Grep all `User{}` constructions pre-merge |
| GORM join table mismatch | Low | Explicit `joinTable` config matching migration |
| Route protection rollout gaps | Medium | Phase 1: critical routes only; audit remaining |
| Permission sprawl on seed additions | Low | Enumerate all slugs in seed file; CI check for duplicates |

## Rollback Plan

Revert migration, remove permissions module and authorization middleware, restore `authctx.User` to original struct, remove permission fields from auth/role DTOs, remove role-permission endpoints. Routes revert to auth-only protection.

## Dependencies

- Existing `Cache` interface (Redis/Noop) for permission caching

## Success Criteria

- [ ] `permissions` and `role_permissions` tables exist via migration
- [ ] Root has all permissions seeded; admin has admin permissions; user has basic
- [ ] Permission records contain `module`, `action`, `slug`, `name`, `description`, `is_system`, `display_order`
- [ ] `RequirePermission("users.create")` grants access to authorized users, 403 to others
- [ ] `RequireRole("root")` grants access by role slug
- [ ] Unauthenticated requests to protected routes return 401
- [ ] `/auth/me` returns role name and permission slugs
- [ ] `GET /roles/:id/permissions` returns grouped catalog with `granted` flags
- [ ] `PUT /roles/:id/permissions` replaces role's permission slugs; system roles reject removal of system permissions
- [ ] Auth and authorization middleware are separate