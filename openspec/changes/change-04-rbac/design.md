# Design: RBAC Permissions and Authorization

## Technical Approach

Add API-only RBAC beside the existing PASETO identity flow. `middleware.Auth` stays responsible for token validation and sanitized `authctx.User`; authorization remains explicitly declared on endpoints with `RequirePermission("module.action")` / `RequireRole("role-slug")`. Permissions are DB-backed, cache-resolved by user `public_id`, and exposed through grouped API DTOs so a future UI can render modules and granted flags without client-side reshaping.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| API surface | No UI files; serve grouped permission catalogs from the API | Add UI-specific endpoints later | Keeps this change API-only while preserving future UI flexibility. |
| Auth/authz boundary | Keep `Auth` identity-only; add `AttachPermissions`, `RequirePermission`, `RequireRole` in `internal/middleware/authorization.go` | Make `Auth` reject on permission failure | Preserves authentication semantics and keeps authorization visible at route declaration sites. |
| Permission model | Store `module`, `action`, `slug`, `name`, `description`, `is_system`, `display_order` | Infer module/action from slug only | Enables validation, sorting, grouping, and stable API contracts. |
| Actions | Seed explicit actions (`index`, `view`, `list`, `create`, `update`, `delete`) plus business actions (`users.change_role`, `roles.assign_permissions`, `permissions.manage` if kept by spec) | Generic `read`/`write` | Removes ambiguous permissions and matches endpoint intent. |
| Cache invalidation | On role-permission replacement, query affected role members and `Delete(rbac:permissions:{public_id})` per member | Wildcard cache delete | Existing cache interface only supports key deletion; per-user invalidation is explicit and testable. |

## Data Flow

```text
Request -> Auth(token,user lookup) -> authctx.User{Role, RoleSlug}
        -> AttachPermissions(resolver) -> Permissions []string
        -> RequirePermission("users.index") -> handler

PUT /roles/:id/permissions -> validate slugs -> replace join rows
                            -> list role member public_ids
                            -> delete rbac:permissions:{public_id}
                            -> return grouped catalog
```

`/api/v1/auth/me` uses `Auth -> AttachPermissions -> Handler.Me`; no authorization decision beyond authentication.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/permissions/{model,repository,service,dto,handler,routes}.go` | Create | Permission CRUD, grouped catalog DTOs, resolver, slug validation, and system-permission protection. |
| `internal/middleware/authorization.go` | Create | `AttachPermissions`, `RequirePermission`, `RequireRole`, resolver interface, root bypass. |
| `internal/platform/authctx/authctx.go` | Modify | Add `RoleSlug string` and `Permissions []string`. |
| `internal/app/container.go` | Modify | Wire permission module, resolver, authz middleware, role-permission service, and route guards. |
| `internal/modules/roles/{model,repository,service,dto,handler,routes}.go` | Modify | Add permission association/preloads/slug DTOs and role permission catalog/assignment endpoints. Remove role mutation route exposure. |
| `internal/modules/users/repository.go` | Modify | Add role-member lookup for cache invalidation. |
| `internal/modules/auth/{handler,dto,routes}.go` | Modify | `/auth/me` returns permission slugs from context. |
| `internal/modules/users/routes.go` | Modify | Keep endpoint-level `RequirePermission` declarations. |
| `migrations/20260518000000_rbac.sql` | Create | `permissions`, `role_permissions`, indexes, FKs, rollback order. |
| `seeds/{permissions,role_permissions}.go` | Create | Idempotent base permissions and root/admin/user assignments. |

## Interfaces / Contracts

```go
type PermissionResolver interface { Resolve(publicID string) ([]string, error) }
```

`GET /api/v1/roles/:id/permissions` returns the full catalog grouped by module. Each permission includes `slug`, `name`, `description`, `action`, `is_system`, `display_order`, and `granted`.

`PUT /api/v1/roles/:id/permissions` accepts `{"permissions":["users.index","roles.assign_permissions"]}` and replaces assignments exactly. It requires `roles.assign_permissions`, rejects removal of system permissions from system roles, invalidates caches for affected members, and returns the updated grouped catalog.

Route guards remain declared at routes: users use `users.index/view/create/update/delete` plus `users.change_role`; roles list/get use `roles.index/view`; role assignment uses `roles.assign_permissions`; permission CRUD follows explicit permission-module actions or the spec-retained `permissions.manage` business action.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Slug validation, grouping order, system protection, resolver cache hit/miss | Table-driven tests with fake repos/cache. |
| Middleware | 401, 403, root bypass, role slug matching, resolver failure | `httptest` Gin routes and fake resolver. |
| Integration | GORM joins/preloads, replacement semantics, cache invalidation, seed idempotency | SQLite where viable; PostgreSQL migration review for SQL details. |

## Migration / Rollout

Add Goose migration after `20260516000000_auth.sql`. Run migrations before seeds. Rollout is additive except role mutation routes become unmounted. Rollback drops `role_permissions` before `permissions`.

## Open Questions

None.
