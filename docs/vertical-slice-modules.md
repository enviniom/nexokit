# Vertical-slice module conventions

This is the default architecture contract for module slices in Nexokit. Apply these rules first; treat exceptions as explicit design decisions.

The goal is simple: a slice must be reviewable as one business use case, without hiding persistence or business behavior in entity-level shared layers.

## Quick path

1. Keep every module root uniform: `container.go`, `routes.go`, `core/`, `queries/`, and `slices/`.
2. Keep boundary responsibilities strict (`handler` → `service` → `repository`).
3. Put reusable cross-slice queries in `queries/` (one file per query).
4. Keep single-use persistence logic inside the slice repository.
5. Add handler, service, and repository tests for each non-trivial slice.

## Module folder shape

Every vertical-slice module SHOULD use this root shape:

```txt
internal/modules/<module>/
  container.go
  routes.go
  core/
  queries/
  slices/
```

The root tells reviewers where each kind of code lives:

| Path | Purpose |
|---|---|
| `container.go` | Module composition root and dependency wiring only. |
| `routes.go` | Module route registration only. |
| `core/` | Shared domain language for the module. |
| `queries/` | Reusable persistence queries only. |
| `slices/` | Business use-case slices and, when needed, entity groups. |

Flat single-entity modules place slices directly under `slices/`:

```txt
slices/
  list_products/
  view_product/
  create_product/
```

Multi-entity modules group by entity under `slices/`:

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

## Core rules (non-negotiable)

| Area | Rule |
|---|---|
| Slice ownership | Every slice owns `handler.go`, `service.go`, `repository.go`, and tests when applicable. |
| Shared repos | Do **not** create `shared/repository.go` or entity-level shared repositories. |
| `core/` scope | `core/` contains cross-slice domain models/DTOs/contracts/errors/constants and tiny pure domain helpers. It must not contain persistence, HTTP/API mapping, or complex behavior. |
| `queries/` scope | `queries/` uses one file per reusable persistence query and only when the query is used by more than one slice. Single-use queries stay inside the slice repository. |
| `slices/` scope | `slices/` contains all use-case slices. Entity folders may exist only inside `slices/`, not beside `core/` or `queries/`. |
| Repository boundary | Repository wraps reusable queries, owns slice-specific persistence, and maps DB/GORM/persistence errors into domain errors before returning. |
| Service boundary | Service contains business rules and must not import GORM or `platform/apperror`, not even for sentinels such as `gorm.ErrRecordNotFound`. |
| Handler boundary | Handler maps domain errors to HTTP/API responses and field-level validation errors. |

## Boundary responsibilities

| Layer | Owns | Must not own |
|---|---|---|
| Handler | Request binding, request validation short-circuiting, tenant/context extraction, response writing, domain-to-HTTP mapping. | Business decisions, GORM calls, persistence error inspection. |
| Service | Use-case orchestration and business rules. | GORM, SQL details, `platform/apperror`, HTTP status decisions. |
| Repository | Slice persistence, GORM/DB calls, reusable query wrappers, persistence-to-domain error translation. | HTTP/API mapping, field-level validation response shape, business policy. |
| Core | Shared domain language: models, DTOs, contracts, constants, domain errors, small pure helpers. | Repositories, response helpers, GORM queries, handlers, service orchestration. |

### Error mapping rule

Errors move inward as domain errors and outward as API responses:

```txt
DB/GORM error → repository → core/domain error → service → handler → HTTP/API response
```

Examples:

| Case | Repository returns | Handler response |
|---|---|---|
| Missing row | `core.ErrNotFound` | 404 via `apperror.ErrNotFound` |
| Duplicate field | domain error such as `core.ErrUserEmailAlreadyExists` | 422 with `errors.email` |
| Protected resource | domain error such as `core.ErrRoleProtected` | 403 |
| Unexpected DB failure | original/wrapped technical error | 500 via `response.HandleError` |

Do not return `platform/apperror` from services or repositories.

### Expected missing data

If missing data is an exceptional lookup failure, repositories SHOULD translate persistence not-found errors to a domain error:

```txt
gorm.ErrRecordNotFound → core.ErrNotFound
```

If missing data is expected control flow, prefer an explicit existence contract instead of making services compare errors:

```go
FindBySlug(slug string) (*core.IAMPermission, bool, error)
```

Use this shape when the service naturally needs to branch on "exists vs missing", such as idempotent sync operations.

## Multi-entity grouping heuristic

Use this rule exactly:

> If a module contains more than one entity **AND** each entity has more than 3 use-cases, group slices into entity-specific subdirectories (e.g., `categories/`, `product_types/`). Otherwise, slices can reside flat under module root.

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

## Slice checklist

- [ ] `handler.go` exists and only maps API concerns.
- [ ] `service.go` exists and only enforces business rules.
- [ ] `repository.go` exists and owns persistence translation.
- [ ] Slice lives under `slices/` directly or under `slices/<entity>/` for multi-entity modules.
- [ ] Tests cover handler/service/repository behavior for the slice.
- [ ] Any reusable query extracted to `queries/<query_name>.go` with dedicated tests.

## Query extraction checklist

- [ ] Query is reused by more than one slice.
- [ ] Query has its own file in `queries/`.
- [ ] Query has dedicated tests in `queries/`.
- [ ] Slice repository wraps the query and still translates persistence errors to domain errors.
- [ ] Optional parameters such as `excludeID` live in the reusable query when they generalize repeated create/update checks.
- [ ] Query files do not contain response mappers or business policy.
- [ ] Query files do not contain non-persistence helpers; if it does not query storage, it probably does not belong in `queries/`.

## GORM partial model rule

When a module defines local/partial GORM models for tables it does not own, the model MUST map to the real migration table name explicitly.

```go
type IAMUser struct { ... }

func (IAMUser) TableName() string { return "users" }
```

This is mandatory when the struct name differs from the table name. Do not rely on `AutoMigrate` tests alone: they can create the wrong table name in SQLite and hide production failures.

Checklist:

- [ ] Every partial model has a `TableName()` override when needed.
- [ ] Table names match the real Goose migration tables.
- [ ] A direct unit test covers model table names.

## Mapper and helper placement

| Item | Placement |
|---|---|
| Single-use response mapper | Keep local to the slice repository/service that needs it. |
| Reused persistence query | Extract to `queries/<query_name>.go` with tests. |
| Pure domain helper used across slices | Put in `core/` with small table-driven tests if behavior can regress. |
| Response/API mapper | Keep in handler or slice-local helper near the handler. Do not put it in `core/`. |
| Entity-level response mapper reused by several slices | Put in `slices/<entity>/mapper.go` or `slices/<entity>/presenter.go`. Keep it dumb: no business rules, no DB, no HTTP. |
| Catalog/response assembly used by one slice | Keep local to that slice; do not move to `queries/`. |

Use this decision rule:

| Question | If yes | If no |
|---|---|---|
| Does it query storage? | `queries/` if reused, otherwise slice repository. | Do not put it in `queries/`. |
| Is it pure domain language? | `core/`. | Keep checking. |
| Does it shape API/UI response data? | Handler/slice mapper or `slices/<entity>/mapper.go`. | Keep checking. |
| Is it used by only one slice? | Keep it local to that slice. | Consider entity-level mapper or `core/` based on responsibility. |

Example: `core.IsReservedRoleIdentity(name, slug)` belongs in `core` because it is pure IAM domain logic used by multiple role slices and request flows. It is not a query and does not belong in `queries/`.

Example: `BuildRolePermissionCatalog` belongs near the role-permission slice or an entity-level mapper if reused by role slices. It does not belong in `queries/` because it assembles response data and does not query storage.

## Moving existing slices

When migrating an existing module to the `slices/` shape, prefer filesystem moves over rewriting files:

```bash
mkdir -p internal/modules/<module>/slices
mv internal/modules/<module>/<slice> internal/modules/<module>/slices/<slice>
```

For multi-entity modules:

```bash
mkdir -p internal/modules/<module>/slices/<entity>
mv internal/modules/<module>/<entity>/<slice> internal/modules/<module>/slices/<entity>/<slice>
```

After moving, update imports and package references only. Do not rewrite slice internals unless a rule violation is found.

## Slice test checklist

- [ ] Handler tests cover success and mapped error responses.
- [ ] Service tests cover business rules and repository error propagation.
- [ ] Repository tests cover persistence behavior and error translation.
- [ ] Reusable query tests cover query behavior once in `queries/`.
- [ ] Repository wrapper tests stay light and point to query tests when full query behavior is already covered.
- [ ] Partial model table-name tests exist when using local GORM models for existing tables.
