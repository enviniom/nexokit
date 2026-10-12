# Request flow

This reference owns middleware ordering and route-scope policy. [Multitenancy](../infrastructure/multitenancy.md) owns tenant-context and repository isolation rules; [the application map](../../system-map.md) locates the composition sources.

## Global middleware

Keep global middleware in `internal/server/router.go` in this order:

```text
RequestID → DebugErrors → GinLogger → Logger → ErrorLogger → Recovery → CORS
```

Gin unwinds middleware after the handler in reverse order. Recovery attaches panic errors before ErrorLogger's post-handler logic runs. DebugErrors uses configuration environment, not Gin mode; [API validation and errors](../../api_validation_and_errors.md#6-response-safety-and-logging) owns error/debug policy.

Health routes are unversioned. `Container.RegisterModules` mounts business routes below `/api/v1`.

## Route policy

| Surface | Middleware before handler |
|---|---|
| `POST /auth/login` | Global → login limiter |
| `POST /auth/refresh` | Global → refresh limiter |
| `POST /auth/logout` | Global; request refresh token drives revocation, no bearer middleware |
| `GET /auth/me` | Global → Auth |
| `/companies`, `/roles`, `/permissions`, `/onboarding/companies` | Global → Auth → AllowRootGlobalScope → route guard |
| `/users` and user subroutes | Global → Auth → RequireTenantScope → permission guard |

Paths in this table are relative to `/api/v1`. Onboarding requires root and is not a public signup endpoint. Companies use root guards except view, which uses a permission guard; IAM uses permission guards in its entity registrars.

## Authenticated context

`middleware.Auth` validates bearer PASETO, resolves a fresh sanitized user through the injected IAM resolver, rejects inactive/missing users and stores `authctx.User`. IAM's resolver supplies role permissions with that user, so permission guards consume the authenticated context without a separate permission-attachment step.

Tenant middleware runs after auth. Root users may select a company by public ID or slug via `X-Company-ID`. For `/users`, root must select a company (missing header is 400); on global-capable groups root may omit it. Non-root users are scoped to their authenticated company; their selection header is ignored. See the [complete decision table](../infrastructure/multitenancy.md#private-resolution).

## Login and persistence

Login has no trusted actor yet and uses an unscoped credentials lookup. Its service checks user activity and password, then issues access/refresh tokens. Future requests re-resolve the actor from storage rather than trusting token claims alone.

Private tenant-owned handlers read `tenant.FromGin(c)`, then pass the context to persistence. Repositories add company filtering; a foreign-company public-ID lookup should appear absent rather than reveal another tenant's record.

## Optional public tenant capability

For a public tenant surface, explicitly register `middleware.PublicTenant` with `companies.Resolver.ResolveHost` and test host-based isolation. Do not substitute public host resolution for private authenticated selection. Exact domain matching, development header fallback, redirects and operational caveats belong in [domains](../infrastructure/domains.md).

## Review path

Read router → app registration → module routes → container → use-case handler/service/repository. Verify actual auth, scope and guard order; a new handler alone does not make a route secure. For throttling scope and proxy caveats, see [rate limiting](../infrastructure/rate-limiting.md).
