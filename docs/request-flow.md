# Request flow in NexoKit

This guide explains how an HTTP request moves through NexoKit, how authentication and tenant resolution interact, and which paths are different for login and public tenant routes.

## Quick map

| Request type | Main path | Tenant source | Expected scope |
|--------------|-----------|---------------|----------------|
| Login | `POST /api/v1/auth/login` | None during login | Auth bootstrap only |
| Private as root, no tenant header | Auth → root global tenant → permissions | Authenticated user | Global root scope |
| Private as root, with `X-Company-ID` | Auth → selected tenant → permissions | Header resolved by company public ID or slug | One company |
| Private as admin/user | Auth → user company tenant → permissions | Authenticated user `company_id` | One company |
| Public tenant route | Public tenant middleware → handler | Host, domain, subdomain, or dev header | One company |

## Global middleware

Every request starts in `internal/server/router.go` with the Gin engine middleware chain:

```txt
RequestID → GinLogger → AppLogger → Recovery → CORS
```

The router then mounts versioned modules under `/api/v1` through `Container.RegisterModules`.

```go
v1 := r.Group("/api/v1")
registerModules(v1)
```

Today `RegisterModules` receives one router group. If NexoKit adds `/api/v2`, the container should not hardcode version-specific assumptions inside modules. A clean v2 shape would be:

```go
v1 := r.Group("/api/v1")
container.RegisterV1(v1)

v2 := r.Group("/api/v2")
container.RegisterV2(v2)
```

or a version-aware registrar:

```go
container.RegisterModules(server.APIVersion{Group: v1, Version: "v1"})
```

That keeps v1 and v2 routes independent when signatures, DTOs, or middleware order diverge.

## Private authenticated routes

Private module routes are mounted behind this chain:

```txt
Auth → PrivateTenant → AttachPermissions → module route guards → handler
```

### 1. `Auth`

`internal/middleware/auth.go` validates the `Authorization: Bearer <token>` header.

It then:

1. Parses the PASETO access token.
2. Loads a sanitized user through `userLookup.GetAuthUser`.
3. Rejects inactive users.
4. Stores the user in `authctx`.
5. Stores an initial tenant context when it can.

For root users, `Auth` can set global root scope or a numeric `X-Company-ID` scope. The later `PrivateTenant` middleware is still the stronger tenant resolver because it can resolve public IDs or slugs through the companies repository.

### 2. `PrivateTenant`

`internal/middleware/tenant.go` resolves the final tenant context for private routes.

Decision table:

| Actor | Input | Result |
|-------|-------|--------|
| Root | no `X-Company-ID` | `tenant.NewRoot()` global scope |
| Root | `X-Company-ID: <public_id-or-slug>` | `tenant.NewScoped(company.ID, company.Slug)` |
| Admin/user | has `company_id` | `tenant.NewScoped(user.CompanyID, "")` |
| Admin/user | missing `company_id` | `403 Forbidden` |

### 3. `AttachPermissions` and route guards

`authMW` means authentication: “who is the user?”

`authzMW` means authorization enrichment: “which permissions does the user have?”

After `authzMW`, modules can use route guards like:

```go
requirePermission("users.index")
requireRole("root")
```

## Login flow

Login is special because there is no authenticated user yet.

```txt
POST /api/v1/auth/login
Global middleware → auth handler → auth service → user lookup by credentials → token issue
```

Tenant middleware does not run before login. The service validates credentials and returns tokens. The access token includes user identity data, and future requests resolve tenant scope from the authenticated user plus optional root override.

Important consequence: login must remain an unscoped bootstrap lookup. Tenant isolation starts after authentication, when the request has a trusted user context.

## Public tenant routes

Public tenant routes should use `PublicTenant` before their handlers.

Resolution order:

1. Normalize `Host`.
2. Try company `domain` match.
3. Try first subdomain as company slug.
4. In development only, try `X-Tenant`.

Examples:

| Request | Resolution |
|---------|------------|
| `Host: shop.acme.com` | first tries `domain = shop.acme.com`, then slug `shop` if domain is missing |
| `Host: acme.localhost` | may resolve slug `acme` depending on host shape |
| `X-Tenant: acme` in development | resolves by public ID or slug |

If no company is resolved, the middleware returns `404 Not Found`.

## Tenant context and repository filtering

Tenant-owned handlers read the tenant context from Gin and pass it to services/repositories.

```go
tc, ok := tenant.FromGin(c)
```

Repositories apply the scope with:

```go
db = tenant.ApplyTenantScope(db, tc)
```

Rules:

- `IsRootScope == true`: no `company_id` filter is applied.
- `IsRootScope == false`: `WHERE company_id = ?` is applied.
- Cross-tenant private reads and writes should look like `404 Not Found`, not “forbidden but exists”.

## Example flows

### Admin lists users

```txt
GET /api/v1/users
Authorization: Bearer <admin-token>
```

1. Auth loads admin user with `company_id = 10`.
2. PrivateTenant creates `TenantContext{CompanyID: 10}`.
3. Permissions are attached.
4. `users.Handler.List` passes tenant context to service.
5. Repository applies `WHERE company_id = 10`.

### Root lists all users

```txt
GET /api/v1/users
Authorization: Bearer <root-token>
```

1. Auth loads root.
2. PrivateTenant sees no company header.
3. Tenant context is root-global.
4. Repository does not add `company_id` filter.

### Root lists one company’s users

```txt
GET /api/v1/users
Authorization: Bearer <root-token>
X-Company-ID: acme
```

1. Auth loads root.
2. PrivateTenant resolves `acme` to an internal company ID.
3. Tenant context is scoped to Acme.
4. Repository adds `WHERE company_id = <acme-id>`.

### Admin tries to edit another company’s user

```txt
PUT /api/v1/users/usr_globex
Authorization: Bearer <acme-admin-token>
```

1. Auth loads Acme admin.
2. PrivateTenant scopes request to Acme.
3. Repository searches `public_id = usr_globex AND company_id = acme`.
4. No row is found.
5. API returns `404 Not Found`.

## Current design notes

- Tenant primitives live in `internal/platform/tenant` because handlers, middleware, repositories, and generated modules all need the same concept.
- Tenant HTTP resolution lives in `internal/middleware/tenant.go` because it depends on headers, host, Gin, and request lifecycle.
- Auth and tenant are related but separate: auth identifies the actor; tenant defines the data boundary.

## Review checklist for new tenant-owned modules

- [ ] Model has `company_id` and an index.
- [ ] Handler reads tenant context from Gin.
- [ ] Service/repository methods accept tenant context for tenant-owned reads and writes.
- [ ] Repository calls `ApplyTenantScope` before query execution.
- [ ] Cross-tenant reads and writes return `404`.
- [ ] Root global and root scoped behavior are tested separately.
- [ ] Public routes, if any, use `PublicTenant`.
