# Nexokit Architecture Guide

This guide defines how Nexokit code is organized, how dependencies move between layers, and where each kind of responsibility belongs. Use it as the review contract for new API code.

The architecture optimizes for three outcomes:

1. **Clear ownership** — each module owns its use cases and API projections.
2. **Low coupling** — modules do not import each other directly.
3. **Replaceable infrastructure** — domain and use-case code do not depend on GORM, Gin, or startup wiring.

---

## Architectural decision

Nexokit uses a modular vertical-slice architecture with shared platform services and canonical domain models.

```text
cmd/ → app/ → modules/ → platform/ → domain/
                  │
                  └── modules do not import other modules
```

The only place where multiple modules are wired together is `internal/app/container.go`.

---

## Folder map

```text
api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── container.go
│   ├── domain/
│   ├── platform/
│   │   ├── apperror/
│   │   ├── contracts/
│   │   ├── database/
│   │   ├── logger/
│   │   ├── response/
│   │   └── tenant/
│   └── modules/
│       └── <module>/
│           ├── container.go
│           ├── routes.go
│           ├── core/
│           ├── queries/
│           └── slices/
└── migrations/
```

---

## Import rules

| Package | May import | Must not import |
|---|---|---|
| `cmd/api` | `internal/app` | modules, platform, domain directly |
| `app` | modules, platform | business logic packages from slices |
| `modules/<m>/routes.go` | Gin, module container handlers | repositories, services construction |
| `modules/<m>/container.go` | own slices, platform contracts, infrastructure dependencies | business logic |
| `modules/<m>/slices/<usecase>` | own `core`, own `queries`, `domain`, `platform` as needed | another module |
| `modules/<m>/core` | `domain`, `platform/apperror` | Gin, GORM, another module |
| `modules/<m>/queries` | GORM, `domain`, own `core` | HTTP concerns, another module |
| `platform` | `domain` when needed | modules |
| `domain` | standard library and pure value libraries | `internal/*`, Gin, GORM |

**Non-negotiable rule:** modules never import other modules. If one module needs another module's capability, use `platform/contracts` and wire the implementation in `app/container.go`.

---

## Layer responsibilities

### `domain/`

Canonical business entities shared by modules.

Use `domain` for:

- Plain Go structs representing core business entities.
- Pure value objects when they are not tied to transport or persistence.
- Cross-module concepts that must have one canonical meaning.

Do not put here:

- GORM tags or persistence records.
- Gin/request/response DTOs.
- Validation tags.
- Module-specific projections.
- Infrastructure imports.

Example:

```go
package domain

import "github.com/shopspring/decimal"

type Product struct {
    ID          uint
    Name        string
    Slug        string
    SKU         string
    Price       decimal.Decimal
    CategoryID  uint
    Description string
    ImageURL    string
    IsActive    bool
}
```

---

### `platform/`

Shared application infrastructure used by modules.

| Package | Responsibility |
|---|---|
| `platform/apperror` | Sentinel errors, app error classification, safe public messages, status resolution. |
| `platform/response` | HTTP response envelope and response writing helpers. |
| `platform/contracts` | Cross-module interfaces and neutral input/output types. |
| `platform/database` | Database connection and setup. |
| `platform/tenant` | Tenant extraction and tenant context helpers. |
| `platform/logger` | Logging adapter and request-aware logging helpers. |

Platform may be shared, but it is not a place for business policy. If logic belongs to a use case, keep it inside a module slice.

---

### `modules/<module>/core/`

The module's shared language.

Use `core` for:

- Module DTOs.
- Module domain errors.
- Constants and small pure helpers reused by multiple slices.
- Pure mappers from `domain.X` to the module's own DTOs.

Do not put here:

- HTTP handlers.
- GORM queries.
- Service orchestration.
- Repositories.
- DTOs imported from another module.

Rule of thumb: if code is shared by multiple slices in the same module and has no infrastructure dependency, it probably belongs in `core`.

---

### `modules/<module>/queries/`

Reusable persistence queries for the same module.

A query belongs here only when it is used by two or more slices in the same module. Single-use persistence logic stays inside the slice repository.

Rules:

- One query per file.
- Dedicated tests for reusable queries.
- GORM and SQL are allowed.
- No HTTP response mapping.
- No cross-module imports.
- Return `domain` types or the module's own `core` types.

---

### `modules/<module>/slices/`

Business use cases. Each slice owns one use case end to end.

Preferred shape for a non-trivial slice:

```text
slices/
└── create_product/
    ├── handler.go
    ├── service.go
    └── repository.go
```

Use flat slices for small modules:

```text
slices/
├── list_products/
├── view_product/
└── create_product/
```

Group by entity only when the module has multiple entities and each entity has several use cases:

```text
slices/
├── users/
│   ├── list_users/
│   └── create_user/
└── roles/
    ├── list_roles/
    └── assign_permission/
```

---

## Slice boundaries

### Handler

Owns:

- Request binding.
- Request validation short-circuit.
- Tenant/context extraction.
- Calling the service.
- Writing HTTP responses.
- Translating application errors to HTTP responses through `response.HandleError`.

Must not own:

- Business decisions.
- GORM calls.
- Persistence error inspection.
- Domain model construction from request data.

### Service

Owns:

- Use-case orchestration.
- Business rule validation.
- Domain model construction.
- Calling repository contracts.
- Calling cross-module capabilities through `platform/contracts` interfaces.

Must not own:

- GORM or SQL details.
- Gin or HTTP status decisions.
- `platform/apperror` imports.
- Request binding.

### Repository

Owns:

- Slice persistence.
- GORM calls.
- Local persistence records with GORM tags.
- Persistence record ↔ domain conversion.
- Mapping expected persistence errors to module `core` errors.

Must not own:

- HTTP/API mapping.
- Business policy.
- Domain model construction from request data.
- `platform/apperror` imports.

---

## Error model

Expected application errors are declared in `modules/<module>/core` by wrapping `platform/apperror` sentinels.

```go
var ErrProductNotFound = apperror.Wrap(
    apperror.ErrNotFound,
    "Producto no encontrado",
)
```

Error flow:

```text
repository
  maps expected DB errors to core.ErrXxx
    ↓
service
  applies business rules and propagates core errors
    ↓
handler
  logs unexpected causes and calls response.HandleError
```

Mapping rules:

| Situation | Repository returns | HTTP result |
|---|---|---|
| Missing row | `core.ErrXxxNotFound` | 404 |
| Duplicate unique field | `core.ErrXxxAlreadyExists` | 409 |
| Forbidden business operation | `core.ErrXxxProtected` | 403 |
| Invalid state transition | `core.ErrInvalidXxx` | 400 |
| Valid request, invalid domain state | `core.ErrXxxUnprocessable` | 422 |
| Unexpected database failure | wrapped unexpected error with infrastructure cause | 500 |

Important distinction:

- **Expected errors** are safe to show to clients and usually do not need logging.
- **Unexpected errors** must preserve the infrastructure cause for logs, but expose only a safe public message in release mode.

Services should not branch on GORM errors. If a service needs to know whether something exists, the repository exposes that directly:

```go
FindBySlug(ctx context.Context, slug string) (*domain.Product, bool, error)
```

---

## DTO and mapper rules

Each module declares its own DTOs. Shared domain models do not imply shared API shapes.

| Need | Location |
|---|---|
| Canonical entity | `domain/` |
| Module API response DTO | `modules/<module>/core/` |
| Mapper from domain to module DTO | `modules/<module>/core/` |
| Cross-module input/output type | `platform/contracts/` |
| DTO shared across modules | Do not create one by default |

Why: `catalog.ProductResponse` and `sales.CartProductSummary` may share fields today, but they evolve for different API clients. Reusing one DTO couples unrelated contracts and makes future changes harder.

---

## Cross-module capabilities

When module A needs behavior owned by module B:

1. Declare an interface in `platform/contracts`.
2. Module B implements the interface internally.
3. Module B exposes the implementation through its `container.go`.
4. Module A receives the interface through its constructor.
5. `app/container.go` wires B into A.

Example:

```text
platform/contracts  declares DiscountEngine
campaigns           implements DiscountEngine
sales               receives contracts.DiscountEngine
app/container.go    wires campaigns → sales
```

This keeps both modules independent while still allowing collaboration.

---

## Composition and routing

### Module `container.go`

The module composition root constructs slice repositories, services, and handlers.

Rules:

- Wiring only.
- No business logic.
- Receives infrastructure dependencies from the outside.
- Exposes cross-module contract implementations when needed.

### Module `routes.go`

Route registration only.

Rules:

- No handler construction.
- No middleware implementation.
- No business decisions.

### `app/container.go`

The application composition root knows about all modules.

Rules:

- Construct modules in dependency order.
- Wire cross-module contracts.
- Register all module routes.
- No HTTP handler logic.
- No business logic.

If this file becomes difficult to scan, consider introducing Wire or another dependency-injection generator.

---

## Persistence model rules

Domain structs are not GORM models. Repositories define local persistence records.

```go
type productRecord struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"column:name;not null"`
    Slug string `gorm:"column:slug;uniqueIndex"`
}

func (productRecord) TableName() string { return "products" }
```

Rules:

- Persistence records stay in repositories or query files.
- Records do not leave the persistence boundary.
- Convert records to `domain` before returning to services.
- Add `TableName()` when the struct name does not exactly map to the database table.
- Table names must match Goose migrations.

---

## Review checklist

Use this checklist when reviewing a new slice or module.

### Imports

- [ ] `domain` imports no internal package.
- [ ] `platform` imports no module.
- [ ] Modules do not import other modules.
- [ ] `cmd/api` imports only `internal/app` from internal application code.

### Slice boundaries

- [ ] Handler binds input and writes output only.
- [ ] Service owns business rules and domain construction.
- [ ] Repository owns GORM and persistence records.
- [ ] GORM errors do not leak into services for expected cases.

### DTOs and contracts

- [ ] Module DTOs live in that module's `core` package.
- [ ] Cross-module collaboration uses `platform/contracts`.
- [ ] No module imports another module's DTOs.

### Errors

- [ ] Module errors are declared in `core`.
- [ ] Expected persistence errors are mapped to `core.ErrXxx`.
- [ ] Unexpected errors preserve their cause for logging.
- [ ] Public API messages are safe in release mode.

### Persistence

- [ ] Domain structs have no GORM tags.
- [ ] Persistence records stay local to repositories or queries.
- [ ] Partial models define `TableName()` when needed.
- [ ] Table names match migrations.

---

## Quick reference

```text
domain/                       canonical entities; no Gin, no GORM
platform/apperror             application error classification
platform/contracts            cross-module interfaces
platform/response             API response envelope and HTTP helpers
platform/database             database setup
platform/logger               logging adapter
platform/tenant               tenant context helpers

modules/<m>/core              DTOs, module errors, pure mappers, constants
modules/<m>/queries           reusable persistence queries for the same module
modules/<m>/slices/<usecase>  handler + service + repository per use case
modules/<m>/container.go      module wiring only
modules/<m>/routes.go         route registration only

app/container.go              full dependency graph and cross-module wiring
cmd/api/main.go               process entry point
migrations/                   database schema changes
```

---

## What changed from the original architecture reference

- Reduced long code samples to the minimum needed to explain the rule.
- Moved review-critical rules into tables and checklists.
- Clarified the difference between expected and unexpected persistence errors.
- Made `response.HandleError` the HTTP response boundary, while keeping `apperror` as classification.
- Kept the same architectural intent: modular slices, no module-to-module imports, domain without GORM, and composition in `app/container.go`.
