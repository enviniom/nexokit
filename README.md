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

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Start PostgreSQL:
   ```bash
   docker compose up -d
   ```

3. Run migrations:
   ```bash
   make migrate-up
   ```

4. Run the API:
   ```bash
   make run
   ```

5. Check health:
   ```bash
   curl http://localhost:8080/health
   ```

## Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the API binary |
| `make run` | Run the API in development mode |
| `make test` | Run all tests |
| `make migrate-up` | Apply pending migrations |
| `make migrate-down` | Revert last migration |
| `make migrate-create` | Create a new migration |
| `make migrate-status` | Show migration status |
| `make fmt` | Format all Go files |
| `make vet` | Run go vet |
| `make install-hooks` | Install Git pre-commit hook |
| `make uninstall-hooks` | Remove Git pre-commit hook |
| `make check-env` | Check `.env` / `.env.example` parity |

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
