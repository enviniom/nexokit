# Application map

NexoKit composes four modules into a Go/Gin API. Use this map to locate a change before reading the [slice conventions](architecture_and_vsa.md).

## Entrypoints and startup

| Source | Responsibility |
|---|---|
| `cmd/api/main.go` | Bootstrap, HTTP start, interrupt/SIGTERM handling and timed graceful shutdown |
| `cmd/nexokit/main.go` | CLI dispatch through `internal/cli` and `commands.All()` |
| `internal/app/bootstrap.go` | Config → loggers → PostgreSQL → cache → limiter → container → router → permission sync → server |
| `internal/app/container.go` | Constructs modules and injects capabilities; mounts module routes |
| `internal/app/app.go` | Starts server; stops server, database, cache and limiter |

Bootstrap creates the server but does not run migrations or start listening. `Start()` listens; `Stop()` closes resources in order and returns on the first close failure. Configuration optionally loads `.env` from the working directory; existing environment values take precedence.

## Routing and modules

All module routes are under `/api/v1`; health endpoints are unversioned. IAM URLs do **not** have an `/iam` prefix.

| Module | Mounted paths | Responsibility / scope |
|---|---|---|
| auth | `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me` | Credentials, opaque refresh rotation/revocation and current session; only `/me` applies bearer auth |
| companies | `/companies`, `/companies/:id`, `/companies/:id/domains`, `/companies/:id/domains/:domain_id` | Company and domain administration; mostly root guards |
| IAM | `/users`, `/roles`, `/permissions` and entity subroutes | Users, roles and permission catalog in `slices/{users,roles,permissions}`; capabilities in `slices/internal` via `slices/capabilities.go`; users require selected tenant, roles/permissions allow root-global scope |
| onboarding | `POST /onboarding/companies` | Root-authenticated company onboarding, **not public self-signup** |

The app mounts `Auth` + `AllowRootGlobalScope` for companies, roles, permissions and onboarding; users use `Auth` + `RequireTenantScope`. Route-specific permission/role guards follow. See [request flow](references/runtime/request-flow.md) for the authoritative middleware sequence and root selection table.

## Platform and infrastructure

| Location | Capabilities |
|---|---|
| `internal/config` | Typed environment settings and debug-exposure policy |
| `internal/infra/db` | PostgreSQL connection/pool and Goose migrations |
| `internal/infra/logger` | Application/error/access outputs and rotation |
| `internal/infra/cache` | Redis and no-op adapters |
| `internal/middleware` | Auth, authorization, tenant resolution, CORS, rate limits, request IDs, recovery and logging |
| `internal/platform` | Application errors/responses, auth context, tenant primitives, passwords, PASETO, ULID identity, permission registry, query/validation/normalization helpers |
| `internal/shared` | Shared persistence base fields |
| `migrations/`, `seeds/` | Schema and explicit initialization functions |

Capabilities are not all mounted features: `PublicTenant` exists but the current app container does not attach it to public business routes. Rate limiting is attached to auth login/refresh, not as a global router middleware. [Deployment](deployment.md) explains the operational implications.

## Follow a change

1. Locate its module `routes.go`, then container and use-case package.
2. Read handler → optional service → repository → reused query.
3. Inspect entity `model.go`/`dtos.go` in Auth's `core/user` or `core/refresh_token` and IAM's `core/user`, `core/role` or `core/permission`, plus module-root contracts/errors/constants and the corresponding migration. Companies uses `core/company` and `core/company_domain`. Onboarding uses `core/{company,company_domain,user,role,permission}` local records and company-owned workflow DTOs, with its use case in `slices/onboard_company`; for entity ownership and applicable DTO mapping, follow [Architecture §4](architecture_and_vsa.md#4-models-dtos-and-mapping).
4. Follow injected capabilities back to the app container; do not add sibling-module imports.
5. Read adjacent tests and the relevant reference from the [navigation hub](README.md).
