# Dev Environment Specification

## Purpose
Define local development infrastructure: Docker Compose, example environment file, Makefile targets, and project README.

## Requirements

### Requirement: docker-compose.yml with PostgreSQL

The system MUST provide a `docker-compose.yml` that runs PostgreSQL with a configurable database name, user, and password.

#### Scenario: Start local database

- GIVEN `docker-compose up -d` is executed
- WHEN the container is healthy
- THEN PostgreSQL is reachable on `DB_HOST:DB_PORT` with the configured credentials

### Requirement: .env.example

The system MUST provide `.env.example` containing all required environment variables with placeholder or safe default values.

#### Scenario: Copy and configure

- GIVEN `cp .env.example .env`
- WHEN the user fills in secrets
- THEN the application can load a complete configuration

### Requirement: Makefile targets

The system MUST provide Makefile targets: `build`, `test`, `run`, `migrate-up`, `migrate-down`, `migrate-create`, `migrate-status`.

#### Scenario: Run tests

- GIVEN `make test` is executed
- WHEN tests complete
- THEN `go test ./...` has been invoked
- AND exit code reflects test results

#### Scenario: Create migration

- GIVEN `make migrate-create NAME=add_users`
- WHEN the command completes
- THEN a new file exists in `migrations/` with the correct timestamp format

### Requirement: README with setup instructions

The system MUST provide a `README.md` with: project description, prerequisites, `.env` setup, `docker-compose up` step, `make migrate-up`, and `make run`.

#### Scenario: New developer onboarding

- GIVEN a developer clones the repository
- WHEN they follow README instructions
- THEN the API starts locally without additional undocumented steps

## Constraints and Edge Cases

- `docker-compose.yml` MUST use a named volume so data persists across container restarts.
- `.env.example` MUST NOT contain real secrets or passwords.
- `Makefile` MUST fail with a clear message if required tools (`go`, `docker-compose`, `goose`) are missing.
