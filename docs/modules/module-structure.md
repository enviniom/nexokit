# Module structure

A backend module is the unit of vertical organization under `internal/modules/<module>/`. This document covers the standard shape, what each root file and folder is for, and the order to create things in a brand new module.

## Quick path

1. Create the root: `container.go`, `routes.go`, `core/`, `queries/`, `slices/`.
2. Add shared domain language to `core/` and reusable errors to `core/errors.go`.
3. Register routes from `routes.go`; the app container wires dependencies in `container.go`.
4. Add slices under `slices/`. See [`vertical-slices.md`](vertical-slices.md).
5. Promote repeated queries to `queries/`. See [`queries-and-persistence.md`](queries-and-persistence.md).

## Standard module shape

Every backend module MUST use this root shape:

```txt
internal/modules/<module>/
  container.go
  routes.go
  core/
  queries/
  slices/
```

The root is the only place these file and folder names appear. Reviewers can locate code from this shape alone.

## Root file and folder responsibilities

| Path | Owns | MUST NOT own |
|---|---|---|
| `container.go` | Module composition root: instantiates services, repositories, handlers, and the slice `Queries` used by the module. | HTTP route definitions, request binding, business rules, persistence queries. |
| `routes.go` | A `Register(v1 *gin.RouterGroup, ...)` function that mounts the module's HTTP routes and applies permission / role guards. | Business logic, repository wiring, `container.go` style composition. |
| `core/` | Shared cross-slice domain language: models, DTOs, contracts, constants, `core/errors.go` with module-level `apperror` values, small pure domain helpers. | Persistence, response helpers, GORM, handler logic, complex orchestration, HTTP mapping. |
| `queries/` | Reusable persistence queries, one file per query, used by more than one slice. | Single-use queries (those stay in the slice repository), business policy, response mappers, HTTP mapping. |
| `slices/` | Business use-case slices. Each slice has `handler.go`, `service.go`, `repository.go` and tests. Multi-entity modules may group slices by entity under `slices/<entity>/`. | Module-wide wiring, top-level composition, route registration. |

## What each root file looks like

### `container.go`

- Module composition root and dependency wiring only.
- Constructs services, repositories, and handlers.
- Exposes the values that `routes.go` and the app container need.
- Does not register routes; routes are registered in `routes.go`.

### `routes.go`

- Module route registration only.
- Exposes a `Register` function that the app container calls.
- Applies permission / role guards to routes.
- Does not construct services or repositories; those are passed in.

### `core/`

- Shared cross-slice domain language.
- Contains:
  - Domain models used by more than one slice.
  - DTOs that are shared across slices.
  - Contracts and constants.
  - `core/errors.go` with reusable `apperror`-backed module errors.
  - Tiny pure domain helpers (no I/O, no GORM, no HTTP).
- MUST NOT contain:
  - Repositories or persistence queries.
  - Response helpers, handler logic, or HTTP mapping.
  - Complex behavior, transactions, or orchestration.

### `queries/`

- Reusable persistence queries.
- One file per query, named after the query (e.g. `find_user_by_email.go`).
- Each file has dedicated tests in the same package.
- MUST NOT contain:
  - Response mappers.
  - Business policy.
  - Non-persistence helpers.
  - Single-use queries used by only one slice.

### `slices/`

- Business use-case slices.
- Flat for single-entity modules:

  ```txt
  slices/
    list_products/
    view_product/
    create_product/
  ```

- Grouped by entity for multi-entity modules:

  ```txt
  slices/
    users/
      list_users/
      create_user/
    roles/
      list_roles/
      create_role/
    permissions/
      list_permissions/
  ```

- In both shapes, every slice lives under `slices/`. The module root MUST NOT host slice files directly.

## Startup order for a new module

1. Create the root files and folders.
2. Define shared domain models and DTOs in `core/`.
3. Define reusable module errors in `core/errors.go`.
4. Add `Register` in `routes.go` with placeholder handlers.
5. Wire the composition root in `container.go`.
6. Add slices under `slices/`, one per use case. See [`vertical-slices.md`](vertical-slices.md).
7. Promote repeated queries to `queries/` as they appear. See [`queries-and-persistence.md`](queries-and-persistence.md).

## Module startup checklist

- [ ] Root contains `container.go`, `routes.go`, `core/`, `queries/`, `slices/`.
- [ ] `container.go` only does composition and wiring.
- [ ] `routes.go` only registers routes via a `Register` function.
- [ ] `core/` only holds shared cross-slice domain language and `core/errors.go`.
- [ ] `queries/` only holds reusable persistence queries.
- [ ] `slices/` only holds slices (flat or grouped by entity).
- [ ] No slice files live at the module root.
- [ ] No direct imports of other modules. See [`boundaries-and-dependencies.md`](boundaries-and-dependencies.md).
- [ ] Each non-trivial slice has handler, service, repository, and tests. See [`testing.md`](testing.md).
