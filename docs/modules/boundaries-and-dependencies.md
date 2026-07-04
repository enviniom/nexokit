# Module boundaries and dependencies

Modules are independent units. They collaborate through interfaces, never through direct imports. This document covers the rules that keep that contract clean.

## Quick path

1. Modules MUST NOT import other modules directly.
2. Cross-module helpers live in `platform/shared` or another appropriate `platform/*` package.
3. A module exposes implementations of capabilities it owns and declares interfaces for capabilities it needs.
4. The app container injects those implementations wherever needed.
5. Module-internal types and behavior MUST NOT leak across module boundaries.

## No direct module-to-module imports

| Rule | Why |
|---|---|
| A module MUST NOT import another module's package. | Direct imports create compile-time coupling and make slices harder to review, refactor, or extract. |
| A module MUST NOT reach into another module's `core/`, `queries/`, or `slices/`. | These are implementation details; only the owning module should change them. |
| Cross-module collaboration happens through interfaces declared by the consumer and implemented by the provider. | Interfaces invert the dependency and let the app container wire the actual implementation. |

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
| Provider module | Exposes concrete implementations of capabilities the module owns. The owning module decides the implementation. |
| Consumer module | Declares the interface it depends on, named for the capability, not the provider. |
| App container | Wires the provider's concrete implementation into the consumer through constructor injection. |

This is the standard "depend on abstractions" pattern. The provider does not need to know who consumes it; the consumer does not need to import the provider.

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
| Two modules both copy a slug helper into their own `core/`. | Bad | Duplicated logic drifts over time. |
| A module reaches into another module's `slices/` to call a service. | Bad | Slices are implementation details; only the owning module owns them. |
| `platform/shared` exports a `slug` helper used by several modules. | Good | The helper is shared, stable, and platform-owned. |
| A module reaches into `platform/response` from inside `core/`. | Bad | `core/` is for domain language; response helpers belong at the handler boundary. |

## Boundaries checklist

- [ ] No module imports another module's package.
- [ ] No module reaches into another module's `core/`, `queries/`, or `slices/`.
- [ ] Shared helpers live in `platform/shared` or another appropriate `platform/*` package.
- [ ] Each module exposes implementations of the capabilities it owns.
- [ ] Each module declares the interfaces it needs in its own package.
- [ ] The app container wires provider implementations into consumer modules.
- [ ] `core/` does not import `platform/response` or GORM. The only allowed `platform/apperror` import in `core/` is for reusable module application errors in `core/errors.go`.
