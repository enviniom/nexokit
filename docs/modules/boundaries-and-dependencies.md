# Module boundaries and dependencies

Modules are independent units. They collaborate through interfaces, never through direct imports. This document covers the rules that keep that contract clean.

## Quick path

1. Modules MUST NOT import other modules directly.
2. Cross-module helpers live in `platform/shared` or another appropriate `platform/*` package.
3. A module declares interfaces for capabilities it needs; it may also expose stable capabilities for app/middleware use.
4. The app container injects those implementations wherever needed.
5. Module-internal types and behavior MUST NOT leak across module boundaries.

## No direct module-to-module imports

| Rule | Why |
|---|---|
| A module MUST NOT import another module's package. | Direct imports create compile-time coupling and make slices harder to review, refactor, or extract. |
| A module MUST NOT reach into another module's `core/`, `queries/`, or `slices/`. | These are implementation details; only the owning module should change them. |
| Cross-module collaboration happens through interfaces, not concrete module imports. | Interfaces invert the dependency and let the app container wire the actual implementation. |

## Where shared code lives

| Kind of code | Placement |
|---|---|
| Pure helpers reused across modules (e.g. slug rules, formatters, shared domain primitives). | `platform/shared` (or another appropriate `platform/*` package). |
| Cross-cutting infra (HTTP response envelope, error helpers, tenant helpers). | The matching `platform/*` package (`platform/response`, `platform/apperror`, `platform/tenant`, etc.). |
| Module-internal helpers, models, DTOs, errors. | The owning module's `core/`. |
| Slice-local helpers, mappers, response assemblers used by one slice. | The owning slice directory. |

`platform/shared` is the default home for cross-module helpers and small interfaces that are not HTTP-, DB-, or response-specific. Pick a more specific `platform/*` package when one exists (e.g. response helpers go in `platform/response`, not `platform/shared`).

## Interface ownership rule

| Direction | Rule |
|---|---|
| Consumer module | Declares the interface it depends on, named for the capability, not the provider. This is the default for module-to-module collaboration. |
| Provider module | Exposes concrete implementations of capabilities it owns. It may expose a stable interface only when the capability is consumed by app/middleware/platform code rather than by another business module. |
| App container | Wires the provider's concrete implementation into the consumer through constructor injection. |

This is the standard "depend on abstractions" pattern. The provider does not need to know who consumes it; the consumer does not need to import the provider.

### Consumer-owned port example

Use this when one business module needs a capability that belongs elsewhere:

```go
// internal/platform/role/ref.go
type Ref struct {
    ID   uint
    Slug string
}

// internal/modules/onboarding/core/contracts.go
type RoleLookup interface {
    ResolveRoleBySlug(slug string) (role.Ref, error)
}
```

The return type must not be an IAM model. If both sides need a non-primitive value, place the small shared reference type in `internal/platform/<concern>` or `internal/shared`.

```go
// internal/app/container.go
type iamRoleLookup struct {
    resolver iamcore.RoleBySlugResolver
}

func (l iamRoleLookup) ResolveRoleBySlug(slug string) (role.Ref, error) {
    iamRole, err := l.resolver.ResolveRoleBySlug(slug)
    if err != nil {
        return role.Ref{}, err
    }
    return role.Ref{ID: iamRole.ID, Slug: iamRole.Slug}, nil
}

onboarding.NewContainer(db, onboarding.Config{
    RoleLookup: iamRoleLookup{resolver: iamContainer.RoleResolver},
})
```

`onboarding` imports only its own contract and shared/platform types. `iam` provides the source capability, and the app container adapts IAM's internal role model into the shared reference type. Only `internal/app/container.go` imports both modules.

### Provider-exposed capability exception

Use this when a module exposes a stable capability for app, middleware, or platform wiring, not for another business module to import directly:

```go
// internal/modules/iam/core/contracts.go
type AuthUserResolver interface {
    ResolveAuthUser(publicID string) (*authctx.User, error)
}
```

`authctx.User` lives in `internal/platform/authctx`, so middleware can use an auth context type without importing IAM internals. The app container adapts IAM to middleware:

```go
// internal/app/container.go
authMW := middleware.Auth(tokenManager, userLookup{
    resolver: iamContainer.AuthUserResolver,
})
```

This is allowed because middleware is not a business module and must not import `internal/modules/iam` either. Do not use provider-owned contracts as a shortcut for module-to-module imports.

## App container wiring

| Concern | Owner |
|---|---|
| Choosing the concrete implementation of a cross-module capability. | The app container, in `internal/app/container.go`. |
| Constructing a module with its dependencies. | The app container calls into the module's `container.go`. |
| Registering a module's routes. | The app container calls into the module's `routes.go`. |
| Module-internal wiring. | The module's own `container.go`. |

The app container is the only place that knows the full graph of concrete types. Modules only know their own types and the interfaces they consume.

## Good vs bad dependencies

| Pattern | Verdict | Why |
|---|---|---|
| `users` module imports `iam` module to call `iam.GetRoleBySlug`. | Bad | Direct module-to-module import; `users` is coupled to `iam` internals. |
| `users` module declares a `RoleLookup` interface in its own package; the app container injects the `iam` implementation. | Good | `users` depends on its own abstraction; `iam` stays replaceable. |
| `iam` exposes `AuthUserResolver` returning `platform/authctx.User`; the app container adapts it into auth middleware. | Good exception | Middleware is not a module, and the shared auth context type avoids leaking IAM internals. |
| Two modules both copy a slug helper into their own `core/`. | Bad | Duplicated logic drifts over time. |
| A module reaches into another module's `slices/` to call a service. | Bad | Slices are implementation details; only the owning module owns them. |
| `platform/shared` exports a `slug` helper used by several modules. | Good | The helper is shared, stable, and platform-owned. |
| A module reaches into `platform/response` from inside `core/`. | Bad | `core/` is for domain language; response helpers belong at the handler boundary. |

## Boundaries checklist

- [ ] No module imports another module's package.
- [ ] No module reaches into another module's `core/`, `queries/`, or `slices/`.
- [ ] Shared helpers live in `platform/shared` or another appropriate `platform/*` package.
- [ ] Each module declares the interfaces it needs in its own package by default.
- [ ] Provider-owned interfaces are limited to stable app/middleware/platform capabilities.
- [ ] Cross-boundary contracts use primitives or shared/platform reference types, never another module's internal models.
- [ ] The app container wires provider implementations into consumer modules.
- [ ] `core/` does not import `platform/response` or GORM. The only allowed `platform/apperror` import in `core/` is for reusable module application errors in `core/errors.go`.
