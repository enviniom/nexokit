# Delta Spec: change-02-cli

## Domain: dev-tooling

### Requirements

| # | Requirement | Strength |
|---|-------------|----------|
| 1 | The Makefile MUST provide a `dev` target that runs the API in development mode. | MUST |
| 2 | The Makefile MUST provide `build`, `test`, `migrate-up`, `migrate-down`, `migrate-create`, `seed`, `create-root`, `lint`, `fmt` targets. | MUST |
| 3 | The Makefile SHOULD load environment variables from `.env` where applicable. | SHOULD |
| 4 | The Makefile targets MUST fail with a clear error message if required variables are missing. | MUST |

### Scenarios

#### Scenario: Developer runs API in dev mode

- GIVEN the project is cloned and `.env` is present
- WHEN the developer runs `make dev`
- THEN the API starts on the configured port

#### Scenario: Missing database URL during migration

- GIVEN `DATABASE_URL` is not set in `.env` or environment
- WHEN the developer runs `make migrate-up`
- THEN the command fails immediately with a clear error indicating the missing variable

---

## Domain: cli-commands

### Requirements

| # | Requirement | Strength |
|---|-------------|----------|
| 1 | The CLI MUST expose subcommands: `serve`, `create-root`, `migrate up`, `migrate down`, `migrate create`, `make module`, `make migration`, `make seed`, `status`, `config`. | MUST |
| 2 | `create-root` MUST be idempotent: if a root user already exists, it MUST skip creation and report success without error. | MUST |
| 3 | `migrate` commands MUST use Goose as the underlying migration engine. | MUST |
| 4 | `serve` MUST initialize the full application container before starting the HTTP server. | MUST |
| 5 | `config` MUST display the current resolved configuration without exposing secrets. | MUST |

### Scenarios

#### Scenario: Starting the API via CLI

- GIVEN a valid `.env` and database connection
- WHEN the developer runs `go run ./cmd/nexokit serve`
- THEN the HTTP server starts and listens on the configured port

#### Scenario: Creating root user

- GIVEN the database is migrated and no root user exists
- WHEN the developer runs `go run ./cmd/nexokit create-root`
- THEN a user with role `root` is created with a secure random password

#### Scenario: Idempotent root creation

- GIVEN a root user already exists in the database
- WHEN the developer runs `go run ./cmd/nexokit create-root` again
- THEN the command exits with code 0 and prints a message indicating the user already exists

#### Scenario: Running migrations

- GIVEN pending Goose migration files in `migrations/`
- WHEN the developer runs `go run ./cmd/nexokit migrate up`
- THEN all pending migrations are applied in order

#### Scenario: Creating a migration

- GIVEN a valid database connection
- WHEN the developer runs `go run ./cmd/nexokit migrate create create_products_table`
- THEN a new timestamped SQL file is created in `migrations/`

---

## Domain: module-generator

### Requirements

| # | Requirement | Strength |
|---|-------------|----------|
| 1 | `make module <name>` MUST generate a module directory under `internal/modules/<name>/` with files: `handler.go`, `service.go`, `repository.go`, `dto.go`, `model.go`, `routes.go`, `validation.go`. | MUST |
| 2 | The generated model MUST embed `BaseModel` with internal `ID` (uint) and external `PublicID` (ULID). | MUST |
| 3 | The generated routes MUST expose a `Register` function compatible with `gin.RouterGroup`. | MUST |
| 4 | The generator SHOULD support a `--migration` flag to create a corresponding Goose migration in `migrations/`. | SHOULD |
| 5 | The generator SHOULD support a `--tenant` flag to include `company_id` scope in the repository and queries. | SHOULD |
| 6 | The generator MUST NOT modify existing files or global route registration silently. | MUST |
| 7 | If the target module directory already exists, the generator MUST fail with a clear error. | MUST |

### Scenarios

#### Scenario: Generating a basic module

- GIVEN a clean project state
- WHEN the developer runs `go run ./cmd/nexokit make module products`
- THEN `internal/modules/products/` is created with all 7 required files and compiles successfully

#### Scenario: Generating a module with migration

- GIVEN a clean project state
- WHEN the developer runs `go run ./cmd/nexokit make module products --migration`
- THEN the module files are created AND a timestamped migration file is added to `migrations/`

#### Scenario: Generating a module that already exists

- GIVEN `internal/modules/products/` already exists
- WHEN the developer runs `go run ./cmd/nexokit make module products`
- THEN the command exits with a non-zero code and prints an error message

---

## Out-of-Scope Boundaries

- `nexokit new` (global interactive project scaffolding) is explicitly out of scope for this change; it will be addressed in a future change.
- `permissions sync` (automatic introspection and registration of module permissions) is explicitly out of scope for this change; it will be addressed after the RBAC change.
- The CLI is internal to the project (`cmd/nexokit`) and not intended for global installation in this version.
