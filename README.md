# NexoKit

NexoKit is a modular Go starter framework for building SaaS APIs.

## Prerequisites

- Go 1.26+
- Docker & Docker Compose (for local PostgreSQL)
- Make
- [Goose](https://github.com/pressly/goose) for migrations:
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

## Quick Start

Follow these steps to set up the application with a local PostgreSQL instance:

### 1. Environment Configuration
Copy the example environment template to create your active `.env` file:
```bash
cp .env.example .env
```
*(Optionally open `.env` and verify your Postgres credentials match your running local container or docker-compose setup).*

### 2. Launch Local Database
Start the pre-configured PostgreSQL container in the background:
```bash
docker compose up -d
```

### 3. Initialize Schema & Database Seeders
Apply database migrations to structure the schema and seed the initial system permissions:
```bash
# Run GORM schema migrations
make migrate-up

# Seed base system permissions required for RBAC
make seed
```

### 4. Create Initial Global Root User
Provision your initial super-administrator user (`root`). This account will read the values configured under `ROOT_USER_NAME`, `ROOT_USER_EMAIL`, and `ROOT_USER_PASSWORD` inside your `.env`:
```bash
make create-root
```

### 5. Fire Up the Server 🚀
Start the API server in active development mode:
```bash
make dev
```
The server will boot on port `8080` (or the one defined under `APP_PORT` in your `.env`).

## Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the API and CLI binaries |
| `make dev` | Run the API server in development mode |
| `make test` | Run all tests |
| `make test-unit` | Run fast unit tests only |
| `make test-integration` | Run database integration tests |
| `make migrate-up` | Apply pending database migrations |
| `make migrate-down` | Revert the last applied database migration |
| `make migrate-create` | Create a new SQL migration template |
| `make migrate-status` | Display the status of database migrations |
| `make migrate-reset` | Rollback all database migrations |
| `make seed` | Load system permission catalog seeds into DB |
| `make create-root` | Provision the initial global root user from .env credentials |
| `make fmt` | Format all Go source files |
| `make vet` | Run go vet analysis |
| `make install-hooks` | Install local Git pre-commit hooks |
| `make uninstall-hooks` | Remove Git pre-commit hooks |
| `make check-env` | Check `.env` and `.env.example` key alignment |

## Pre-commit Hooks

Install the local Git hook:

```bash
make install-hooks
```

The hook performs the following checks on staged files:

| Check | Behavior |
|-------|----------|
| Binary files | Blocks the commit |
| File size > 1MB | Warns but allows the commit |
| `.env` / `.env.example` key drift | Warns but allows the commit |
| `go vet ./...` | Blocks the commit on errors |
| Unformatted Go files | Blocks the commit; run `make fmt` to fix |

Bypass the hook when necessary:

```bash
git commit --no-verify
```

## Log Files

The application writes to three separate log files under `logs/`:

| File | Purpose |
|------|---------|
| `gin.log` | HTTP access logs written by Gin's built-in logger |
| `app.log` | Structured application logs (all levels) from slog |
| `error.log` | Structured error logs (ERROR level and above only) from slog |

Configure paths via `LOG_GIN_FILE`, `LOG_FILE`, and `LOG_ERROR_FILE` in your `.env`.

## Project Structure

```
cmd/api/              - API entrypoint
cmd/nexokit/          - CLI entrypoint
internal/app/         - Application bootstrap and container
internal/config/      - Typed configuration
internal/infra/       - Database, cache, logger adapters
internal/server/      - HTTP server and router
internal/middleware/  - HTTP middleware chain
internal/platform/    - Cross-cutting concerns (response, errors, validator)
internal/modules/     - Business modules (stubs in change-01)
internal/shared/      - BaseModel types
migrations/           - Goose SQL migrations
tests/                - Integration tests and helpers
```

## Conventions

- All API responses use the standard envelope (`success`, `message`, `data`, `meta`, `errors`).
- No `gin.H` inline in handlers; use `platform/response` helpers.
- Modules register routes via a `Register` function.
- GORM is for runtime queries only; Goose handles schema migrations.
