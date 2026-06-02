# Vertical-slice module conventions

This is the default architecture contract for module slices in Nexokit. Apply these rules first; treat exceptions as explicit design decisions.

The goal is simple: a slice must be reviewable as one business use case, without hiding persistence or business behavior in entity-level shared layers.

## Quick path

1. Create one directory per use-case slice.
2. Keep boundary responsibilities strict (`handler` → `service` → `repository`).
3. Put reusable cross-slice queries in `queries/` (one file per query).
4. Keep single-use persistence logic inside the slice repository.
5. Add handler, service, and repository tests for each non-trivial slice.

## Core rules (non-negotiable)

| Area | Rule |
|---|---|
| Slice ownership | Every slice owns `handler.go`, `service.go`, `repository.go`, and tests when applicable. |
| Shared repos | Do **not** create `shared/repository.go` or entity-level shared repositories. |
| `core/` scope | `core/` contains cross-slice domain models/DTOs/contracts/errors/constants and tiny pure domain helpers. It must not contain persistence, HTTP/API mapping, or complex behavior. |
| `queries/` scope | `queries/` uses one file per reusable persistence query and only when the query is used by more than one slice. Single-use queries stay inside the slice repository. |
| Repository boundary | Repository wraps reusable queries, owns slice-specific persistence, and maps DB/GORM/persistence errors into domain errors before returning. |
| Service boundary | Service contains business rules and must not import GORM or `platform/apperror`. |
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

## Multi-entity grouping heuristic

Use this rule exactly:

> If a module contains more than one entity **AND** each entity has more than 3 use-cases, group slices into entity-specific subdirectories (e.g., `categories/`, `product_types/`). Otherwise, slices can reside flat under module root.

### IAM example

IAM is a multi-entity module, so grouping by entity is expected (`roles/`, `permissions/`, `users/`) with slices under each group.

## Slice checklist

- [ ] `handler.go` exists and only maps API concerns.
- [ ] `service.go` exists and only enforces business rules.
- [ ] `repository.go` exists and owns persistence translation.
- [ ] Tests cover handler/service/repository behavior for the slice.
- [ ] Any reusable query extracted to `queries/<query_name>.go` with dedicated tests.

## Query extraction checklist

- [ ] Query is reused by more than one slice.
- [ ] Query has its own file in `queries/`.
- [ ] Query has dedicated tests in `queries/`.
- [ ] Slice repository wraps the query and still translates persistence errors to domain errors.
- [ ] Optional parameters such as `excludeID` live in the reusable query when they generalize repeated create/update checks.
- [ ] Query files do not contain response mappers or business policy.

## Mapper and helper placement

| Item | Placement |
|---|---|
| Single-use response mapper | Keep local to the slice repository/service that needs it. |
| Reused persistence query | Extract to `queries/<query_name>.go` with tests. |
| Pure domain helper used across slices | Put in `core/` with small table-driven tests if behavior can regress. |
| Response/API mapper | Keep in handler or slice-local helper near the handler. Do not put it in `core/`. |
| Catalog/response assembly used by one slice | Keep local to that slice; do not move to `queries/`. |

Example: `core.IsReservedRoleIdentity(name, slug)` belongs in `core` because it is pure IAM domain logic used by multiple role slices and request flows. It is not a query and does not belong in `queries/`.

## Slice test checklist

- [ ] Handler tests cover success and mapped error responses.
- [ ] Service tests cover business rules and repository error propagation.
- [ ] Repository tests cover persistence behavior and error translation.
- [ ] Reusable query tests cover query behavior once in `queries/`.
- [ ] Repository wrapper tests stay light and point to query tests when full query behavior is already covered.
