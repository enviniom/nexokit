# Backend modules

This is the practical guide for how backend modules in Nexokit are organized. The rule that defines the vertical-slice shape is owned by OpenSpec; this guide is the navigation entry point for the per-concern tutorials.

> **The vertical-slice rule is owned by OpenSpec.** See [`openspec/core_context.md` §2.2](../../openspec/core_context.md) for the canonical rule (heuristic, `core/` files, slice shape, container, routes) and [`openspec/specs/backend/vertical-slice-modules/spec.md`](../../openspec/specs/backend/vertical-slice-modules/spec.md) for the formal contract. The per-concern tutorials below stay aligned with that rule; if a tutorial ever disagrees with the spec, the spec wins.

The goal is reviewability: a subagent should be able to load only the document it needs and still apply the same rules the rest of the team uses.

## Quick path

1. Decide what you are working on: a whole module, a slice, a dependency, errors/validation, queries/persistence, or tests.
2. Open the matching document below and follow its checklist.
3. Keep `docs/api-conventions.md` open as the cross-cutting API reference.

## Read this when...

| If your task is... | Read this document |
|---|---|
| Create a new module | [`modules/module-structure.md`](modules/module-structure.md) |
| Create or change a slice | [`modules/vertical-slices.md`](modules/vertical-slices.md) |
| Review or wire module dependencies | [`modules/boundaries-and-dependencies.md`](modules/boundaries-and-dependencies.md) |
| Add or change DTOs, validation, error mapping, AppError usage | [`modules/validation-and-errors.md`](modules/validation-and-errors.md) and [`error-handling.md`](error-handling.md) |
| Add or change a reusable query, repository translation, partial GORM model | [`modules/queries-and-persistence.md`](modules/queries-and-persistence.md) |
| Write or review tests for a module | [`modules/testing.md`](modules/testing.md) |
| Need the old consolidated reference | [`vertical-slice-modules.md`](vertical-slice-modules.md) (compatibility stub) |

## Document map

| Document | Covers |
|---|---|
| `modules/module-structure.md` | Standard folder shape, `container.go` / `routes.go` / `core/` / `queries/` / `slices/` responsibilities, startup order. |
| `modules/vertical-slices.md` | Slice shape (`handler` / `service` / `repository`), single-entity vs multi-entity grouping, slice checklist, moving existing slices. |
| `modules/boundaries-and-dependencies.md` | No direct module-to-module imports, `platform/shared` rules, module-owned implementations and module-declared interfaces, app container wiring. |
| `modules/validation-and-errors.md` | DTO `Validate()` contract, response envelope, 400 / 422 / AppError mapping, `core/errors.go` with module-owned `Code` constants and `apperror` helpers, expected control flow with `(*Customer, bool, error)`. |
| `modules/queries-and-persistence.md` | When a query belongs in `queries/`, repository translation rules, GORM partial model `TableName()` rule, mapper placement. |
| `modules/testing.md` | Handler / service / repository / query / DTO / module-error / table-name tests, CRUD-light vs business-heavy criteria. |

## Core rules at a glance

These rules apply everywhere and are restated in detail in the dedicated documents. Do not skip them even if you only read one document.

| Rule | Where it lives |
|---|---|
| Module root is uniform: `container.go`, `routes.go`, `core/`, `queries/`, `slices/`. | [`module-structure.md`](modules/module-structure.md) |
| Slices own `handler.go`, `service.go`, `repository.go` and tests. | [`vertical-slices.md`](modules/vertical-slices.md) |
| Slices always live under `slices/`; multi-entity modules may group by entity under `slices/<entity>/`. | [`vertical-slices.md`](modules/vertical-slices.md) |
| Modules MUST NOT import other modules directly. | [`boundaries-and-dependencies.md`](modules/boundaries-and-dependencies.md) |
| Shared cross-module code belongs in `platform/shared` or the appropriate `platform/*` package. | [`boundaries-and-dependencies.md`](modules/boundaries-and-dependencies.md) |
| A module exposes implementations of capabilities it owns and declares interfaces for capabilities it needs. The app container injects implementations. | [`boundaries-and-dependencies.md`](modules/boundaries-and-dependencies.md) |
| DTOs own `Validate()` and return `response.ValidationErrors` keyed by field. | [`validation-and-errors.md`](modules/validation-and-errors.md) |
| Binding / invalid JSON = 400; DTO validation = 422; `AppError` is not used for field validation. | [`validation-and-errors.md`](modules/validation-and-errors.md) |
| Reusable module errors live in `core/errors.go`, use module-owned `Code` constants, and are built with `platform/apperror` helpers. | [`validation-and-errors.md`](modules/validation-and-errors.md) |
| Services and repositories MUST NOT create ad-hoc `apperror` values inline. | [`validation-and-errors.md`](modules/validation-and-errors.md) |
| Handlers route business / app errors through `response.HandleError`. | [`validation-and-errors.md`](modules/validation-and-errors.md) |
| Expected control flow uses explicit contracts like `(*Customer, bool, error)`, not `AppError`. | [`validation-and-errors.md`](modules/validation-and-errors.md) |
| `queries/` holds reusable persistence queries used by more than one slice; single-use persistence stays in the slice repository. | [`queries-and-persistence.md`](modules/queries-and-persistence.md) |
| Partial GORM models for tables the module does not own MUST implement `TableName()` when the struct name differs from the migration table name. | [`queries-and-persistence.md`](modules/queries-and-persistence.md) |
| Tests cover handler / service / repository behavior for non-trivial slices. | [`testing.md`](modules/testing.md) |

## Module startup checklist

Use this for a brand new module. Each item links to the document with the full rule.

- [ ] Create the module root: `container.go`, `routes.go`, `core/`, `queries/`, `slices/`. ([`module-structure.md`](modules/module-structure.md))
- [ ] Decide whether the module is single-entity or multi-entity and pick the slice grouping. ([`vertical-slices.md`](modules/vertical-slices.md))
- [ ] Put shared domain models, DTOs, contracts, constants, and `core/errors.go` in `core/`. ([`module-structure.md`](modules/module-structure.md), [`validation-and-errors.md`](modules/validation-and-errors.md))
- [ ] Register routes from `routes.go`; the app container calls `Register` and chooses the route group. ([`module-structure.md`](modules/module-structure.md))
- [ ] Wire dependencies in `container.go`; do not import other modules. ([`boundaries-and-dependencies.md`](modules/boundaries-and-dependencies.md))
- [ ] For each slice, add `handler.go`, `service.go`, `repository.go` and tests. ([`vertical-slices.md`](modules/vertical-slices.md), [`testing.md`](modules/testing.md))
- [ ] Promote repeated queries to `queries/<query_name>.go` with their own tests. ([`queries-and-persistence.md`](modules/queries-and-persistence.md))
- [ ] Add a `TableName()` test for every partial GORM model that targets a non-owned table. ([`queries-and-persistence.md`](modules/queries-and-persistence.md))

## Related cross-cutting docs

| Document | When to also read it |
|---|---|
| [`api-conventions.md`](api-conventions.md) | DTO naming, list query helpers, soft delete, tenant scope, response envelopes. |
| [`testing.md`](testing.md) | Project-wide test commands, integration setup, CI reproduction. |
| [`multitenancy.md`](multitenancy.md) | Tenant scope rules for repositories and handlers. |
| [`request-flow.md`](request-flow.md) | End-to-end HTTP request path through modules. |
