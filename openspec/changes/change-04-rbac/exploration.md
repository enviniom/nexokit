## Exploration: RBAC — Permissions, Role-Permissions, Middleware, and `/auth/me`

### Current State

NexoKit already has a working auth system built on PASETO v4.local access tokens + opaque refresh tokens. The authentication flow is:

1. `middleware.Auth()` validates the `Authorization: Bearer <token>` header, parses PASETO claims, resolves the user via `AuthUserLookup`, checks `IsActive`, and injects an `authctx.User` into the Gin context.
2. `authctx.User` carries: `ID`, `PublicID`, `Email`, `Name`, `Role` (role **name** string), `RoleID`, `CompanyID`, `IsActive`.
3. The PASETO token embeds `sub` (publicID), `role` (name), and `company_id` in its claims.
4. `auth.Handler.Me()` reads from `authctx.FromGin()` and returns a `users.UserResponse` containing `PublicID`, `Name`, `Email`, `IsActive`, `RoleID`, `RoleName`, `CompanyID`.
5. Protected routes use `protected.Use(c.authMW)` in `container.go` — a blanket auth check with **no permission or role discrimination**.

The roles module (`internal/modules/roles/`) provides full CRUD for `roles.Role` (ID, PublicID, Name, Slug, Description, IsSystem). Three system roles (`root`, `admin`, `user`) are seeded via `seeds/roles.go`. The users module (`internal/modules/users/`) has `User.RoleID` as a foreign key to `roles`. There is **no concept of permissions** anywhere yet.

The cache infrastructure (`internal/infra/cache/`) defines a `Cache` interface with `Get/Set/Delete/Close` using `context.Context` and `[]byte` values. Currently only `NoopCache` is wired in bootstrap; `RedisCache` (using `go-redis`) exists but is not enabled.

Route registration follows a flat module pattern: each module exposes `func Register(v1 *gin.RouterGroup, ...)`. The container wires all modules and mounts them via `RegisterModules(v1)`.

### Affected Areas

- `internal/modules/permissions/model.go` — **NEW**: Permission model (new module)
- `internal/modules/permissions/repository.go` — **NEW**: Permission persistence
- `internal/modules/permissions/service.go` — **NEW**: Permission business logic
- `internal/modules/permissions/handler.go` — **NEW**: Permission HTTP handlers (optional, for admin)
- `internal/modules/permissions/dto.go` — **NEW**: Permission DTOs
- `internal/modules/permissions/routes.go` — **NEW**: Permission route registration
- `internal/modules/roles/model.go` — **MODIFY**: Add `Permissions []Permission` has-many via `RolePermissions`
- `internal/modules/roles/repository.go` — **MODIFY**: Preload permissions where needed
- `internal/modules/roles/service.go` — **MODIFY**: Include permissions in role responses
- `internal/modules/roles/dto.go` — **MODIFY**: Add permissions to `RoleResponse`
- `internal/modules/auth/handler.go` — **MODIFY**: Me handler must include permissions
- `internal/modules/auth/service.go` — **MODIFY**: Login/refresh may need to return permissions info
- `internal/modules/auth/dto.go` — **MODIFY**: Me response DTO needs permissions
- `internal/middleware/auth.go` — **MODIFY**: Consider loading permissions into authctx
- `internal/middleware/authorization.go` — **NEW**: `RequirePermission` and `RequireRole` middleware
- `internal/platform/authctx/authctx.go` — **MODIFY**: Add `Permissions []string` or similar to `User`
- `internal/app/container.go` — **MODIFY**: Wire permission module and authorization middleware
- `internal/app/bootstrap.go` — **MODIFY**: If cache wiring changes for permission cache
- `migrations/YYYYMMDDHHMMSS_rbac.sql` — **NEW**: Create `permissions` and `role_permissions` tables
- `seeds/permissions.go` — **NEW**: Seed base permissions for all modules
- `seeds/role_permissions.go` — **NEW**: Seed role-permission assignments
- `seeds/roles.go` — **MODIFY**: Possibly unify with new seed files
- `tests/integration/` — **NEW**: Integration tests for RBAC middleware
- `internal/middleware/auth_test.go` — **MODIFY**: Update or add tests for auth + permission flow

### Approaches

#### 1. Eager Loading — AuthMiddleware loads permissions into context

**Description**: During `Auth()`, after resolving the user, also query and load the user's permissions into `authctx.User`. Every authenticated request pays the cost of a permissions query (with caching).

- **Pros**: Permissions always available in context; `RequirePermission` middleware is stateless and fast (just checks the context); simpler to reason about.
- **Cons**: Every authenticated request loads permissions even when no permission check is needed; auth middleware couples to permission module.
- **Effort**: Medium

#### 2. Lazy Loading — RequirePermission fetches on demand

**Description**: Auth middleware only resolves identity. `RequirePermission("users.create")` fetches and checks permissions at the point of need, optionally caching results.

- **Pros**: No unnecessary queries when endpoints don't check permissions; clean separation between auth and authorization; middleware is self-contained.
- **Cons**: Each permission check may trigger a DB/cache query (mitigated by caching); slightly more complex middleware.
- **Effort**: Medium

#### 3. Token-Embedded Claims — Store permission slugs in PASETO token

**Description**: Include permission slugs in the PASETO access token claims.

- **Pros**: Zero DB queries for permission checks; stateless authorization.
- **Cons**: Token size bloats (root gets all permissions); permissions are stale until token refresh; Makes token creation coupled to permission assignment; revocation impossible without token rotation.
- **Effort**: Low (but architecturally wrong for this system)

### Recommendation

**Approach 2 (Lazy Loading) with cache backing and root bypass.**

Rationale:

1. **Separation of concerns** — The project specification explicitly states "auth and authorization must be separate." Eager loading in auth middleware violates this by coupling auth to permission resolution. Lazy loading keeps auth middleware purely about identity verification.

2. **Performance** — With the existing `Cache` interface, `RequirePermission` can cache the user's full permission set by `user_public_id` with a short TTL (e.g., 5 minutes). This gives O(1) lookups after first check without coupling auth to permissions.

3. **Root bypass** — `RequirePermission` should check: if role is `root`, allow immediately. Additionally, the seed should explicitly assign all permissions to root for auditability. This gives both runtime performance and data-level completeness.

4. **Context enrichment** — After `RequirePermission` loads permissions, store them in Gin context so handlers can inspect them too. `authctx.User` should grow a `Permissions []string` field for the `/auth/me` endpoint.

5. **`RequireRole`** — Implement as a simpler middleware that checks the role slug from `authctx.User.Role`. Useful for endpoint protection where any user with a given role can access, regardless of specific permission granularity.

**Implementation structure**:

- New `permissions` module (`internal/modules/permissions/`) following flat module convention.
- New `role_permissions` join table (no separate module — just a GORM many-to-many on the `roles` model).
- `RequirePermission` middleware in `internal/middleware/authorization.go` — uses a `PermissionResolver` interface injected from container.
- `RequireRole` middleware in the same file — checks `authctx.User` role against slug.
- Permission cache uses the `Cache` interface (Redis in production, Noop for tests where middleware tests use direct mocks).
- Seeds: `seeds/permissions.go` for base CRUD-per-module slugs, `seeds/role_permissions.go` for role-permission assignments.
- Migration: `migrations/20260518XXXXXX_rbac.sql` creating `permissions` and `role_permissions` tables.

### Risks

- **Permission cache staleness**: If permissions change, cached permissions may be stale for up to TTL seconds. Mitigation: short TTL (5 min) and cache invalidation on permission/role-permission mutations.
- **Root bypass consistency**: If someone adds a new permission but forgets to seed it to root, root won't have it unless the `RequirePermission` middleware has an explicit root bypass. Mitigation: seed root with all permissions AND add root-role bypass in middleware.
- **Many-to-many GORM conventions**: GORM has specific patterns for join tables. Must use explicit `joinTable` configuration to match the `role_permissions` schema from the migration. Risk: GORM auto-creates differently-named join table if not configured.
- **authctx.User backward compatibility**: Adding `Permissions []string` to `authctx.User` affects all existing code that constructs this struct. The `container.go` `userLookup` and auth service must be updated. Risk: missing call sites.
- **Route protection rollout**: All existing protected routes currently only require authentication. After RBAC, they need explicit `RequirePermission` calls. This is a large manual change. Risk: missing endpoints or over-restricting.
- **N+1 queries on `/auth/me`**: Loading user + role + permissions requires joins. Must use GORM `Preload` to avoid N+1.

### Ready for Proposal

Yes. The exploration is complete. All critical patterns (auth middleware, authctx, container wiring, module structure, seeding conventions, migration convention, test patterns) are understood. The next phase should create the proposal for this change.