# Exploration: Vertical Slice Modules for NexoKit

## Current State

NexoKit uses a **horizontal layered architecture** within each module. Every module under `internal/modules/` follows the same flat structure:

```
internal/modules/{module}/
├── model.go          # Domain models (GORM entities)
├── dto.go            # Request/response DTOs + validation
├── handler.go        # HTTP handlers (Gin)
├── service.go        # Business logic (interfaces + implementations)
├── repository.go     # Data access (GORM)
├── routes.go         # Route registration
├── contracts.go      # (roles only) shared constants
└── *_test.go         # Tests for each layer
```

**Wiring** happens in `internal/app/container.go`, which:
1. Imports all 6 modules directly
2. Creates repositories, services, and handlers for each module
3. Stores handlers (and some repos/services) as fields on `Container`
4. Calls each module's `Register()` function in `RegisterModules()`

**Cross-module dependencies** exist:
- `auth` → `users` (auth service uses `users.Repository`)
- `onboarding` → `companies`, `users`, `roles`, `permissions` (orchestrates multi-module transaction)
- `users` → `roles` (via `roleResolverAdapter` in container)
- `roles` → `permissions` (permission catalog)
- `companies` → none (cleanest candidate)

**Companies module** (10 files, ~1708 LOC) is the cleanest migration candidate:
- No cross-module imports
- 7 registered endpoints mapping to 7 use-case slices
- Single handler with 8 methods (1 unregistered: `Create`), single service with 8 methods, single repository with 12 methods
- Tests: handler_test.go, service_test.go, repository_test.go, migration_test.go

## Affected Areas

| File/Directory | Impact | Reason |
|---|---|---|
| `internal/modules/companies/` | **Restructure** | First migration target — split into use-case slices |
| `internal/modules/companies/shared/` | **New shared package** | Shared models, DTOs, constants, and contracts used by slices without importing root `companies` |
| `internal/modules/companies/routes.go` | **Move to root** | Cross-cutting route registration |
| `internal/modules/companies/container.go` | **New** | Module-level composition root (wiring) |
| `internal/modules/companies/list_companies/` | **New slice** | `GET /companies` — handler + service + repository + tests |
| `internal/modules/companies/view_company/` | **New slice** | `GET /companies/:id` — handler + service + repository + tests |
| `internal/modules/companies/update_company/` | **New slice** | `PUT /companies/:id` — handler + service + repository + tests |
| `internal/modules/companies/delete_company/` | **New slice** | `DELETE /companies/:id` — handler + service + repository + tests |
| `internal/modules/companies/list_company_domains/` | **New slice** | `GET /companies/:id/domains` — handler + service + repository + tests |
| `internal/modules/companies/create_company_domain/` | **New slice** | `POST /companies/:id/domains` — handler + service + repository + tests |
| `internal/modules/companies/update_company_domain/` | **New slice** | `PUT /companies/:id/domains/:domain_id` — handler + service + repository + tests |
| `internal/app/container.go` | **Simplify** | Calls module container instead of individual constructors |
| `internal/modules/auth/` | **Future migration** | Cross-module dep on users — harder |
| `internal/modules/users/` | **Future migration** | Tenant-aware, moderate complexity |
| `internal/modules/roles/` | **Future migration** | Most complex (2849 LOC), cache, permission catalog |
| `internal/modules/permissions/` | **Future migration** | Moderate complexity |
| `internal/modules/onboarding/` | **Future migration** | Cross-module orchestrator — special case |

## Approaches

### Approach 1: Use-Case Slices with Shared Models (Recommended)

Each use case gets its own sub-package containing handler, service, repository, and tests. Models, DTOs, constants, and shared contracts live in a dedicated `shared` subpackage so slices do not import the root `companies` package and create Go import cycles. Routes and container wiring stay at module root.

```
internal/modules/companies/
├── routes.go                   # Cross-cutting: all route registration
├── container.go                # Composition root: wires all slices
├── shared/
│   ├── model.go                # Shared: Company, CompanyDomain
│   ├── dto.go                  # Shared: request/response DTOs
│   └── contracts.go            # Shared constants/interfaces/contracts
├── list_companies/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── *_test.go
├── view_company/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── *_test.go
├── update_company/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── *_test.go
├── delete_company/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── *_test.go
├── list_company_domains/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── *_test.go
├── create_company_domain/
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── *_test.go
└── update_company_domain/
    ├── handler.go
    ├── service.go
    ├── repository.go
    └── *_test.go
```

- **Pros**: Clean separation, each slice independently testable, models not duplicated, avoids Go import cycles, incremental migration possible, root container only calls module container
- **Cons**: Repository methods split across slices may need shared interface or base repository, some code duplication in repository setup
- **Effort**: Medium for companies (pilot), Medium-High for full migration
- **Review budget risk**: High if done all at once — MUST be incremental per module

### Approach 2: CQRS-Inspired Slices (Commands and Queries)

Separate read and write paths entirely, with dedicated repositories per side.

```
internal/modules/companies/
├── model.go
├── dto.go
├── routes.go
├── container.go
├── commands/
│   ├── update_company/
│   ├── delete_company/
│   ├── create_company_domain/
│   └── update_company_domain/
└── queries/
    ├── list_companies/
    ├── view_company/
    └── list_company_domains/
```

- **Pros**: Clear read/write boundary, easier to optimize queries independently
- **Cons**: Over-engineering for current complexity, more packages to navigate, model sharing becomes harder
- **Effort**: High
- **Review budget risk**: High

### Approach 3: Feature Folders with Layer Sub-folders

Group by feature first, then layer within each feature.

```
internal/modules/companies/
├── model.go
├── dto.go
├── routes.go
├── container.go
├── company/
│   ├── handler.go        # list_companies, view_company, update_company, delete_company
│   ├── service.go
│   ├── repository.go
│   └── *_test.go
└── domain/
    ├── handler.go        # list_company_domains, create_company_domain, update_company_domain
    ├── service.go
    ├── repository.go
    └── *_test.go
```

- **Pros**: Fewer packages, less fragmentation, easier to navigate for small modules
- **Cons**: Company slice would be very large (all CRUD in one), doesn't truly separate use cases
- **Effort**: Low-Medium
- **Review budget risk**: Medium

## Recommendation

**Approach 1** (Use-Case Slices with Shared Models) is the best fit given the constraints:

1. **Models stay shared** inside `internal/modules/companies/shared` — no duplication and no slice imports from root `companies`
2. **Each use case is independently testable** — handler, service, repository co-located
3. **Module container is the composition root** — root container calls only module containers
4. **Incremental migration** — companies first, then others one at a time
5. **Routes stay at root** — cross-cutting concern, registers all slices

**Architecture correction after apply:** earlier wording said shared models/DTOs stay at module root. In Go, that creates an import cycle because root `companies` must import slice packages for routing/container wiring while slices would also import root `companies` for shared types. The corrected contract is: root `companies` contains only route/module wiring; shared models/DTOs/contracts live in `companies/shared`. The module boundary is unchanged because `shared` remains under `internal/modules/companies/`.

For the **companies module specifically**, the slice breakdown maps 1:1 to registered endpoints (7 slices, no `create_company` since there is no public `POST /companies` route):

| Slice | Endpoint | Handler Method |
|-------|----------|----------------|
| `list_companies/` | `GET /companies` | `List` |
| `view_company/` | `GET /companies/:id` | `GetByPublicID` |
| `update_company/` | `PUT /companies/:id` | `Update` |
| `delete_company/` | `DELETE /companies/:id` | `Delete` |
| `list_company_domains/` | `GET /companies/:id/domains` | `ListDomains` |
| `create_company_domain/` | `POST /companies/:id/domains` | `CreateDomain` |
| `update_company_domain/` | `PUT /companies/:id/domains/:domain_id` | `UpdateDomain` |

**Note:** The `Create` handler method exists in code but has no registered route — `create_company` is intentionally excluded from this change.

**Key design decisions for the pilot:**
- Repository interfaces and shared DTO/model contracts stay in `internal/modules/companies/shared` so slices can reference them without importing root `companies`
- Each slice has its own concrete repository implementation that embeds or references a shared DB connection
- DTOs stay in `internal/modules/companies/shared` since they're shared between handlers and external consumers without forcing slice imports from root `companies`
- The module `container.go` wires all slices and exposes only the handler(s) needed for route registration

## Risks

1. **Repository method fragmentation** — The current `GormRepository` has 12 methods. Splitting them across 7 slices means either duplicating the `*gorm.DB` setup or creating a shared base. Recommendation: each slice gets its own thin repository struct with only the methods it needs, all sharing the same `*gorm.DB`.

2. **Cross-module dependencies in other modules** — Auth depends on users, onboarding depends on 4 modules. These will need interface-based decoupling or a different slice strategy. The pilot (companies) avoids this entirely.

3. **Test migration overhead** — 21 test files across 6 modules. Each test file must be moved to its corresponding slice. Table-driven tests may need restructuring.

4. **Container complexity** — The module `container.go` will need to wire multiple slices. If not careful, it becomes a service locator. Must remain a pure composition root: instantiate → wire → expose.

5. **Review budget** — Migrating even just the companies module will likely exceed 400 lines. The 800-line budget from the preflight helps, but chained PRs should be considered:
   - PR 1: New slice structure + container (no deletion of old files)
   - PR 2: Wire new structure, delete old files
   - PR 3: Test updates and verification

6. **Import path changes** — All imports referencing `internal/modules/companies` types will need updating to `internal/modules/companies/shared` where they refer to shared models/DTOs/contracts. The `app/container.go` and any middleware that uses companies repository contracts (like `AllowRootGlobalScope`) will need adaptation.

## Ready for Proposal

**Yes.** The exploration is complete with sufficient detail to proceed to proposal, spec, design, and tasks phases. The companies module is a clean pilot candidate with no cross-module dependencies. The recommended approach (use-case slices with shared models/DTOs/contracts in `companies/shared`) satisfies all stated constraints.

The orchestrator should tell the user:
- Exploration complete — companies module identified as pilot candidate
- 3 approaches evaluated; use-case slices with shared models recommended
- 6 risks identified, most notably repository fragmentation and review budget
- Ready to proceed to proposal phase for review before any implementation
