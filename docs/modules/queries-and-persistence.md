# Queries and persistence

`queries/` is the home for reusable persistence queries. Slice repositories own slice-specific persistence. This document covers when to extract a query, how repositories translate persistence errors, the GORM partial model `TableName()` rule, and where mappers belong.

## Quick path

1. A query belongs in `queries/` only when more than one slice uses it.
2. Slice repositories wrap `queries/` files and translate persistence errors to domain errors.
3. Partial GORM models for non-owned tables MUST implement `TableName()` when the struct name differs from the migration table name.
4. Mappers and helpers have explicit placement rules; do not put them in the wrong place.

## Query extraction rules

| Rule | Why |
|---|---|
| Put a query in `queries/<query_name>.go` only when more than one slice uses it. | Single-use queries are easier to read inside the slice that owns them. |
| One file per query, named after the query (e.g. `find_user_by_email.go`). | File name = query name. Reviewers can locate the query in seconds. |
| Each `queries/` file has its own tests in the same package. | The reusable query has a single, dedicated test surface. |
| The slice repository wraps the query and still translates persistence errors to domain errors. | The repository remains the persistence-to-domain boundary. |
| Optional parameters such as `excludeID` live in the reusable query when they generalize repeated create / update checks. | The query stays general and the slice stays thin. |
| `queries/` files MUST NOT contain response mappers or business policy. | The query layer is persistence only. |
| `queries/` files MUST NOT contain non-persistence helpers. | If it does not query storage, it probably does not belong in `queries/`. |

## Query extraction checklist

- [ ] Query is reused by more than one slice.
- [ ] Query has its own file in `queries/`.
- [ ] Query has dedicated tests in `queries/`.
- [ ] Slice repository wraps the query and still translates persistence errors to domain errors.
- [ ] Optional parameters such as `excludeID` live in the reusable query when they generalize repeated checks.
- [ ] Query files do not contain response mappers or business policy.
- [ ] Query files do not contain non-persistence helpers.

## Repository persistence translation

| Rule | Why |
|---|---|
| Repository wraps reusable queries and owns slice-specific persistence. | The repository is the only place that knows both the query and the slice's needs. |
| Repository MUST map DB / GORM / persistence errors into domain errors before returning using the helpers in `queries/map_errors.go`. | Services and controllers stay free of GORM error inspection and logic. |
| Every module with persistence MUST have a `queries/` package, even if it has no shared query files. | It serves as the single source of truth for the database-to-domain error translation via `map_errors.go`. |
| `queries/map_errors.go` MUST define entity-specific mapping helpers (e.g., `MapCategoryError(err error) error`) taking only the GORM error as a parameter. | Firm signatures stay clean and translation logic is centralized per entity. |
| Repositories MUST NOT do ad-hoc DB/GORM error string checking or direct error comparisons inline. | Eliminates variance in how constraints, unique keys, and "not found" outcomes are mapped across slices. |
| Services MUST NOT import GORM. | Services are pure business rules; persistence is the repository's job. |
| When a slice does not need a reusable query, the repository contains the query inline. | `queries/` is for reuse, not for storage of every query. |


## GORM partial model `TableName()` rule

When a module defines local / partial GORM models for tables it does not own, the model MUST map to the real migration table name explicitly:

```go
type IAMUser struct { ... }

func (IAMUser) TableName() string { return "users" }
```

| Rule | Why |
|---|---|
| Every partial model has a `TableName()` override when its name differs from the migration table name. | GORM would otherwise infer the wrong table name and break queries in production. |
| Table names MUST match the real Goose migration tables. | Local models and migrations must agree. |
| A direct unit test covers model table names. | `AutoMigrate` tests can pass in SQLite even when the production table name is wrong. |
| Do not rely on `AutoMigrate` tests alone. | They can create the wrong table name in SQLite and hide production failures. |

### Table-name checklist

- [ ] Every partial model has a `TableName()` override when needed.
- [ ] Table names match the real Goose migration tables.
- [ ] A direct unit test covers model table names.

## Mapper and helper placement

| Item | Placement |
|---|---|
| Single-use response mapper. | Keep local to the slice repository / service that needs it. |
| Reused persistence query. | Extract to `queries/<query_name>.go` with tests. |
| Pure domain helper used across slices. | Put in `core/` with small table-driven tests if behavior can regress. |
| Response / API mapper. | Keep in the handler or a slice-local helper near the handler. Do not put it in `core/`. |
| Entity-level response mapper reused by several slices. | Put in `slices/<entity>/mapper.go` or `slices/<entity>/presenter.go`. Keep it dumb: no business rules, no DB, no HTTP. |
| Catalog / response assembly used by one slice. | Keep local to that slice; do not move to `queries/`. |

### Placement decision rule

| Question | If yes | If no |
|---|---|---|
| Does it query storage? | `queries/` if reused, otherwise slice repository. | Do not put it in `queries/`. |
| Is it pure domain language? | `core/`. | Keep checking. |
| Does it shape API / UI response data? | Handler / slice mapper or `slices/<entity>/mapper.go`. | Keep checking. |
| Is it used by only one slice? | Keep it local to that slice. | Consider entity-level mapper or `core/` based on responsibility. |

### Examples

- `core.IsReservedRoleIdentity(name, slug)` belongs in `core/` because it is pure IAM domain logic used by multiple role slices and request flows. It is not a query and does not belong in `queries/`.
- `BuildRolePermissionCatalog` belongs near the role-permission slice or an entity-level mapper if reused by role slices. It does not belong in `queries/` because it assembles response data and does not query storage.

## Persistence checklist

- [ ] Every module with persistence contains a `queries/` package and `queries/map_errors.go` file.
- [ ] `queries/map_errors.go` exposes entity-specific error mappers taking only the GORM error as a parameter.
- [ ] `queries/` contains reusable persistence queries, one file per query (if reused).
- [ ] Each `queries/` file has dedicated tests.
- [ ] Slice repository wraps reusable queries and translates persistence errors to domain errors using the helpers in `queries/map_errors.go`.
- [ ] No GORM imports in services.
- [ ] Partial GORM models for non-owned tables implement `TableName()` when names differ.
- [ ] Table names match the real Goose migration tables and have direct unit tests.
- [ ] Mappers and helpers are placed by the decision rule above.
