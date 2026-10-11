# Spec: module-generator

> Reconciled with accepted generator output in `fa2f893`, the [CLI reference](../../../docs/references/runtime/cli.md#module-generation) and [new-code architecture](../../../docs/architecture_and_vsa.md). This supersedes the mandatory flat seven-file layout inherited from change-02-cli (2026-05-16), not the applicable safety or behavioral requirements.
> Verified sources: `internal/cli/generator/module.go`, `internal/cli/templates/templates.go` and `internal/cli/templates/module/*.tmpl`.

## Requirements

| # | Requirement | Strength |
|---|-------------|----------|
| 1 | `make module <name>` MUST generate `internal/modules/<name>/` with entity `core/<name>/model.go` and `dtos.go`, module `core/contracts.go` and `core/errors.go`, `container.go`, `routes.go`, and get/list slices containing `handler.go` and `repository.go`. | MUST |
| 2 | `--crud` MUST add create/update/delete slices and routes. Plain CRUD MUST NOT introduce pass-through services. | MUST |
| 3 | The model MUST embed `shared.BaseModel` with internal `ID` (uint) and external `PublicID` (ULID), declare its module table name, and own a `BeforeCreate` hook that generates a ULID through `platform/identity` when absent and validates a supplied ID strictly. | MUST |
| 4 | Entity DTO validation MUST use pointer receivers and persist name trimming before required/minimum-length rules. Response DTOs MUST own `FromModel` mapping and expose only public ID, name and timestamps, never internal keys, company ownership or audit identities. | MUST |
| 5 | Handlers MUST return 400 for JSON binding failures, 422 for field validation failures, and use `response.HandleError` for operation errors without exposing unexpected technical text. Repositories MUST translate GORM not-found, duplicate-key and foreign-key sentinels into module errors while retaining diagnostic causes. Driver constraint translation requires application GORM `TranslateError`; unrecognized failures remain safe 500 responses. | MUST |
| 6 | Routes MUST expose `Register` compatible with `*gin.RouterGroup` and the generated container. Construction and registration with authentication, authorization and applicable tenant middleware MUST remain explicit adopter work; generation MUST NOT silently edit global route registration or existing files. | MUST |
| 7 | The generator SHOULD support `--migration` to create a timestamped Goose SQL stub in `migrations/`. This does not implement the entity/tenant schema or execute a migration; the adopter must complete and review it. | SHOULD |
| 8 | The generator SHOULD support `--tenant`. Tenant handlers MUST resolve trusted `tenant.FromGin` context and reject absent/invalid scope with 400 before persistence. Get/create/update/delete MUST require a selected company; only list MAY accept explicit trusted root-global scope established by middleware policy. | SHOULD / MUST |
| 9 | Tenant repositories MUST enforce the same scope policy at their boundary and explicitly apply tenant scope to reads/writes. Create MUST derive `company_id` solely from trusted context, ignoring JSON ownership. Update MUST use a scoped transaction, never Save/upsert or ownership reassignment. Missing/foreign/deleted records MUST NOT be mutated. | MUST |
| 10 | Lists MUST support normalized bounded pagination, name search, creation-date filters and a sorting allowlist of `name`, `created_at`, `updated_at`; the scaffold has no status field/filter. Successful delete MUST use `response.NoContent` (204, empty body). | MUST |
| 11 | Names MUST pass snake_case validation. If the target module directory already exists, generation MUST fail clearly rather than overwrite it. | MUST |

## Scenarios

### Scenario: Generating a basic module

- GIVEN a clean project state
- WHEN the developer runs `go run ./cmd/nexokit make module products`
- THEN entity model/DTO core, module contracts/errors, container/routes and `slices/get_products/` and `slices/list_products/` handler/repository pairs are created as formatted, compilable Go
- AND no mutation slices or pass-through services are emitted, and no application routes are automatically mounted

### Scenario: Generating CRUD

- GIVEN a clean project state
- WHEN the developer runs `go run ./cmd/nexokit make module products --crud`
- THEN create/update/delete handler/repository slices and POST/PUT/DELETE routes are added to the default GET routes
- AND create/update normalize and validate DTOs, responses use safe DTO mappings, unexpected errors remain safe, and successful delete returns 204 without an envelope

### Scenario: Generating a tenant module

- GIVEN a module generated with `--tenant` (and `--crud` for mutations)
- WHEN a handler receives absent or invalid trusted context
- THEN it returns 400 without calling persistence
- WHEN a selected company performs an operation
- THEN repository access is scoped to that company; create assigns its ownership from context and update cannot change it or insert on a miss
- WHEN explicit trusted root-global scope performs list
- THEN global listing is permitted, but get/create/update/delete still require selected-company scope

### Scenario: Generating a module with migration

- GIVEN a clean project state
- WHEN the developer runs `go run ./cmd/nexokit make module products --migration`
- THEN the module files and a timestamped Goose SQL stub are created
- AND the adopter must implement/review the schema and explicitly apply migrations; generation performs neither schema implementation nor migration execution

### Scenario: Generating a module that already exists

- GIVEN `internal/modules/products/` already exists
- WHEN the developer runs `go run ./cmd/nexokit make module products`
- THEN the command exits with a non-zero code and a clear error, preserving existing module files

## Out-of-Scope Boundaries

- Existing modules are not automatically migrated to this new-code layout.
- `nexokit new` (global interactive project scaffolding) is explicitly out of scope; it will be addressed in a future change.
- `permissions sync` (automatic introspection and registration of module permissions) is explicitly out of scope for module generation; current application startup synchronization is separate.
- The CLI is internal to the project (`cmd/nexokit`) and not intended for global installation in this version.
