# Vertical slices

A slice is a single business use case packaged in one directory. Slices are the unit of review: a reviewer should be able to read a slice top to bottom and understand one use case without jumping across the module.

This document covers the slice shape, the single-entity vs multi-entity grouping heuristic, and how to move existing slices into the `slices/` layout.

## Quick path

1. One slice = one directory = one use case.
2. Slice owns `handler.go`, `service.go`, `repository.go`, and tests.
3. Single-entity modules place slices directly under `slices/`.
4. Multi-entity modules with more than 3 use-cases per entity group under `slices/<entity>/`.

## Slice shape

```txt
slices/
  <slice_name>/
    handler.go
    service.go
    repository.go
    handler_test.go      (when the slice is non-trivial)
    service_test.go
    repository_test.go
```

Or, for multi-entity modules:

```txt
slices/
  <entity>/
    <slice_name>/
      handler.go
      service.go
      repository.go
      *_test.go
```

## Layer responsibilities

| Layer | Owns | MUST NOT own |
|---|---|---|
| Handler | Request binding, request validation short-circuiting, tenant / context extraction, response writing, domain-to-HTTP mapping. | Business decisions, GORM calls, persistence error inspection. |
| Service | Use-case orchestration and business rules. | GORM, SQL details, `platform/apperror`, HTTP status decisions. |
| Repository | Slice persistence, GORM / DB calls, wrappers around reusable `queries/` files, persistence-to-domain error translation. | HTTP / API mapping, field-level validation response shape, business policy. |

See [`module-structure.md`](module-structure.md) for the parent / module responsibilities and [`validation-and-errors.md`](validation-and-errors.md) for error mapping details.

## Slice ownership

| Area | Rule |
|---|---|
| Slice files | Every slice owns `handler.go`, `service.go`, `repository.go`, and tests when applicable. |
| Shared repos | Do NOT create `shared/repository.go` or entity-level shared repositories. |
| Reusable queries | Repeated queries live in `queries/`; the slice repository wraps them and translates errors. |
| Slice folder | Slice folder lives under `slices/` directly, or under `slices/<entity>/` for multi-entity modules. |
| Module root | The module root MUST NOT host slice files directly. |

## Single-entity vs multi-entity grouping

Use this rule exactly:

> If a module contains more than one entity **AND** each entity has more than 3 use-cases, group slices into entity-specific subdirectories (e.g. `users/`, `roles/`, `permissions/`) under `slices/`. Otherwise, slices can reside flat under `slices/`.

| Module shape | Layout |
|---|---|
| Single-entity module, or multi-entity with at most 3 use-cases per entity | Flat: `slices/<slice>/`. |
| Multi-entity module with more than 3 use-cases per entity | Grouped: `slices/<entity>/<slice>/`. |

### IAM example

IAM is a multi-entity module, so grouping by entity is expected under `slices/`:

```txt
internal/modules/iam/
  container.go
  routes.go
  core/
  queries/
  slices/
    users/
    roles/
    permissions/
```

## Moving existing slices

When migrating an existing module to the `slices/` shape, prefer filesystem moves over rewriting files.

For a flat module:

```bash
mkdir -p internal/modules/<module>/slices
mv internal/modules/<module>/<slice> internal/modules/<module>/slices/<slice>
```

For a multi-entity module:

```bash
mkdir -p internal/modules/<module>/slices/<entity>
mv internal/modules/<module>/<entity>/<slice> internal/modules/<module>/slices/<entity>/<slice>
```

After moving, update imports and package references only. Do not rewrite slice internals unless a rule violation is found.

## Slice checklist

- [ ] `handler.go` exists and only maps API concerns.
- [ ] `service.go` exists and only enforces business rules.
- [ ] `repository.go` exists and owns persistence translation.
- [ ] Slice lives under `slices/` directly or under `slices/<entity>/` for multi-entity modules.
- [ ] Tests cover handler / service / repository behavior for the slice.
- [ ] Any reusable query is extracted to `queries/<query_name>.go` with dedicated tests.
- [ ] Service does not import GORM.
- [ ] Repository translates persistence errors to domain errors before returning.
- [ ] Handler routes business / app errors through `response.HandleError`.
- [ ] No `apperror` values are constructed inline in services or repositories.
