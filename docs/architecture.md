# NexoKit Architecture

This document is the canonical architecture reference for NexoKit. For historical versions, see [`archive/README.md`](archive/README.md).

## Design goals

1. **Clear ownership** — every module owns its use cases, DTOs, and persistence rules.
2. **Low coupling** — modules do not import each other directly.
3. **Replaceable infrastructure** — business code does not depend on GORM, Gin, or startup wiring details.

## Dependency direction

```text
cmd/ → internal/app/ → internal/modules/ → internal/platform/ → internal/shared/
                           │
                           └── modules do not import other modules
```

The only place multiple modules are wired together is `internal/app/container.go`.

## Entrypoints

| Path | Responsibility |
|------|----------------|
| `cmd/api/main.go` | Production API server entrypoint. Handles signals and graceful shutdown. |
| `cmd/nexokit/main.go` | Internal developer CLI entrypoint. Delegates to `internal/cli/commands`. |

## Application bootstrap

`internal/app/` owns composition and startup:

- `bootstrap.go` — loads configuration, connects to the database/cache, builds the container, and starts the server.
- `container.go` — the single composition root. It constructs modules in dependency order, wires cross-module contracts, and exposes `RegisterModules`.
- `app.go` — the running application (`Start`, `Stop`, health checks).

## HTTP routing

`internal/server/router.go` creates the Gin engine and mounts the global middleware chain:

```text
RequestID → DebugErrors → GinLogger → AppLogger → ErrorLogger → Recovery → CORS
```

Versioned module routes are mounted under `/api/v1` through the app container:

```text
/api/v1/auth/*           → auth module (login, refresh, revoke, session)
/api/v1/companies/*      → companies module (CRUD + domains)
/api/v1/iam/*            → IAM module (users, roles, permissions)
/api/v1/onboarding/*     → onboarding module (company sign-up)
```

Auth, tenant, and permission middleware are applied by the app container when each module is registered.

## Modules

NexoKit currently ships four business modules:

| Module | Purpose | Key slices |
|--------|---------|------------|
| `auth` | Login, token refresh/revoke, session view | `authenticate_user`, `rotate_token`, `revoke_token`, `view_session` |
| `companies` | Company and company-domain CRUD | `list_companies`, `view_company`, `create_company_domain`, `update_company`, etc. |
| `iam` | Users, roles, permissions | `users/*`, `roles/*`, `permissions/*`, internal sync/resolver helpers |
| `onboarding` | Company self-onboarding | `onboard_company` |

### Module structure

Modules follow the vertical-slice layout documented in [`modules.md`](modules.md): a `container.go` composition root, `routes.go`, module-local `core/`, reusable `queries/`, and one slice folder per use case containing `handler.go`, `service.go`, and `repository.go`.

### Slice boundaries

| Layer | Owns | Must not own |
|-------|------|--------------|
| `handler.go` | Request binding, validation short-circuit, tenant/context extraction, response writing | Business decisions, GORM calls |
| `service.go` | Use-case orchestration, business rules, domain model construction | GORM/SQL, HTTP status decisions, `platform/apperror` |
| `repository.go` | Persistence records (GORM tags), GORM calls, DB→domain error mapping | Business policy, DTO construction, `platform/apperror` |

Modules use `internal/shared.BaseModel` for common fields. Module-local `core/` structs hold DTOs and partial persistence records when needed.

## Platform

`internal/platform/` holds cross-application utilities such as `apperror`, `response`, `tenant`, `password`, `token`, `identity`, `validator`, and `query`. Modules receive these through constructor injection and never instantiate infrastructure themselves.

## Cross-module collaboration

When one module needs a capability owned by another, the consumer declares a small local port instead of importing the provider module:

1. The consuming module declares the interface it needs in `internal/modules/<consumer>/core/contracts.go` or in the consuming slice.
2. The owning module implements that shape and exposes the concrete implementation from its `container.go`.
3. The consuming module receives the interface as a constructor parameter.
4. `internal/app/container.go` wires the implementation into the consumer.

If the contract needs a non-primitive value shared across boundaries, use a small type in `internal/platform/<concern>` or `internal/shared`; never return another module's internal model.

There is one allowed exception: a module may expose a stable capability for app, middleware, or platform wiring. Example: `internal/modules/iam/core/contracts.go` defines `AuthUserResolver`, returning `platform/authctx.User`, and `internal/app/container.go` adapts it into auth middleware. Middleware is not a business module, and it must not import IAM internals either. For module-to-module collaboration, prefer consumer-owned ports.

Example of the default pattern: `internal/modules/onboarding/core/contracts.go` declares `PasswordHasher`, implemented by `internal/platform/password` and injected by the app container.

This keeps modules independent while allowing composition.

## Request flow

A private request passes through global middleware, then auth, tenant, and permission middleware before reaching a module handler, service, and repository. For auth/tenant resolution tables, see [`request-flow.md`](request-flow.md).

## Error flow

Expected errors are declared in module `core/` by wrapping `platform/apperror` sentinels. They travel outward as `repository → service → handler → response.HandleError`, where `ErrorLogger` owns the structured log line. Unexpected errors become `500` responses; debug details are exposed only when `Config.ExposeDebugErrors()` is true.

## Migrations and seeds

- `migrations/` owns the database schema. Use Goose files named `YYYYMMDDHHMMSS_description.sql`.
- `seeds/` contains Go files in `package seeds` exporting `*Seed() error` functions, run by `nexokit seed`.
- `cmd/nexokit` also provides `migrate`, `create-root`, and `make module/migration/seed` generation commands.

See [`cli.md`](cli.md) for CLI details and [`deployment.md`](deployment.md) for production runbook.

## Quick review checklist

- Modules do not import other modules; module-to-module needs use consumer-owned ports and app-container wiring.
- Handlers bind input/write output; services own business rules; repositories own GORM.
- Migrations in `migrations/` are the schema source of truth.
