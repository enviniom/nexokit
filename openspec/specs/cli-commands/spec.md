# Spec: cli-commands

> Source of truth for internal CLI command requirements.
> Merged from change-02-cli (2026-05-16).

## Requirements

| # | Requirement | Strength |
|---|-------------|----------|
| 1 | The CLI MUST expose subcommands: `serve`, `create-root`, `migrate up`, `migrate down`, `migrate create`, `make module`, `make migration`, `make seed`, `status`, `config`. | MUST |
| 2 | `create-root` MUST be idempotent: if a root user already exists, it MUST skip creation and report success without error. | MUST |
| 3 | `migrate` commands MUST use Goose as the underlying migration engine. | MUST |
| 4 | `serve` MUST initialize the full application container before starting the HTTP server. | MUST |
| 5 | `config` MUST display the current resolved configuration without exposing secrets. | MUST |

## Scenarios

### Scenario: Starting the API via CLI

- GIVEN a valid `.env` and database connection
- WHEN the developer runs `go run ./cmd/nexokit serve`
- THEN the HTTP server starts and listens on the configured port

### Scenario: Creating root user

- GIVEN the database is migrated and no root user exists
- WHEN the developer runs `go run ./cmd/nexokit create-root`
- THEN a user with role `root` is created with a secure random password

### Scenario: Idempotent root creation

- GIVEN a root user already exists in the database
- WHEN the developer runs `go run ./cmd/nexokit create-root` again
- THEN the command exits with code 0 and prints a message indicating the user already exists

### Scenario: Running migrations

- GIVEN pending Goose migration files in `migrations/`
- WHEN the developer runs `go run ./cmd/nexokit migrate up`
- THEN all pending migrations are applied in order

### Scenario: Creating a migration

- GIVEN a valid database connection
- WHEN the developer runs `go run ./cmd/nexokit migrate create create_products_table`
- THEN a new timestamped SQL file is created in `migrations/`
