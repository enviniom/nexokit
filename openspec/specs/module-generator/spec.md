# Spec: module-generator

> Source of truth for module generator requirements.
> Merged from change-02-cli (2026-05-16).

## Requirements

| # | Requirement | Strength |
|---|-------------|----------|
| 1 | `make module <name>` MUST generate a module directory under `internal/modules/<name>/` with files: `handler.go`, `service.go`, `repository.go`, `dto.go`, `model.go`, `routes.go`, `validation.go`. | MUST |
| 2 | The generated model MUST embed `BaseModel` with internal `ID` (uint) and external `PublicID` (ULID). | MUST |
| 3 | The generated routes MUST expose a `Register` function compatible with `gin.RouterGroup`. | MUST |
| 4 | The generator SHOULD support a `--migration` flag to create a corresponding Goose migration in `migrations/`. | SHOULD |
| 5 | The generator SHOULD support a `--tenant` flag to include `company_id` scope in the repository and queries. | SHOULD |
| 6 | The generator MUST NOT modify existing files or global route registration silently. | MUST |
| 7 | If the target module directory already exists, the generator MUST fail with a clear error. | MUST |

## Scenarios

### Scenario: Generating a basic module

- GIVEN a clean project state
- WHEN the developer runs `go run ./cmd/nexokit make module products`
- THEN `internal/modules/products/` is created with all 7 required files and compiles successfully

### Scenario: Generating a module with migration

- GIVEN a clean project state
- WHEN the developer runs `go run ./cmd/nexokit make module products --migration`
- THEN the module files are created AND a timestamped migration file is added to `migrations/`

### Scenario: Generating a module that already exists

- GIVEN `internal/modules/products/` already exists
- WHEN the developer runs `go run ./cmd/nexokit make module products`
- THEN the command exits with a non-zero code and prints an error message

## Out-of-Scope Boundaries

- `nexokit new` (global interactive project scaffolding) is explicitly out of scope; it will be addressed in a future change.
- `permissions sync` (automatic introspection and registration of module permissions) is explicitly out of scope; it will be addressed after the RBAC change.
- The CLI is internal to the project (`cmd/nexokit`) and not intended for global installation in this version.
