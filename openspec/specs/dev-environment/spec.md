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

The system MUST provide Makefile targets: `build`, `test`, `test-unit`, `test-integration`, `test-coverage`, `run`, `migrate-up`, `migrate-down`, `migrate-create`, `migrate-status`, `fmt`, `vet`, `install-hooks`, `uninstall-hooks`, `check-env`. All MUST appear in `.PHONY`.

- `test` runs `go test ./...` (all tests).
- `test-unit` runs `go test ./... -short` (skips integration tests).
- `test-integration` runs `go test ./tests/integration/...` (integration tests only).
- `test-coverage` runs `go test ./... -coverprofile=coverage.out` and prints coverage summary.

#### Scenario: Run tests

- GIVEN `make test` is executed
- WHEN tests complete
- THEN `go test ./...` has been invoked
- AND exit code reflects test results

#### Scenario: Run unit tests only

- GIVEN `make test-unit` is executed
- WHEN tests complete
- THEN `go test ./... -short` has been invoked
- AND integration tests guarded by `testing.Short()` are skipped

#### Scenario: Run integration tests

- GIVEN `make test-integration` is executed
- WHEN tests complete
- THEN `go test ./tests/integration/...` has been invoked
- AND only integration tests run

#### Scenario: Generate coverage report

- GIVEN `make test-coverage` is executed
- WHEN tests complete
- THEN `coverage.out` is generated in the project root
- AND a coverage summary is printed to stdout

#### Scenario: Create migration

- GIVEN `make migrate-create NAME=add_users`
- WHEN the command completes
- THEN a new file exists in `migrations/` with the correct timestamp format

### Requirement: Pre-commit hook

`scripts/pre-commit.sh` MUST be executable. It MUST block on binary files, `go vet` errors, or unformatted Go files, and fail fast. It SHOULD warn (non-blocking) for files >1MB or `.env`/`.env.example` key mismatches (ignoring comments and blanks). Output MUST use green ✓, red ✗, yellow ⚠.

#### Scenario: Binary file blocks commit

- GIVEN a staged binary file
- WHEN the pre-commit hook runs
- THEN it blocks with a red ✗ message
- AND the commit does not proceed

#### Scenario: Unformatted Go blocks commit

- GIVEN a staged Go file that fails `gofmt -l`
- WHEN the pre-commit hook runs
- THEN it blocks with a red ✗ containing "run make fmt"

#### Scenario: Warnings allow commit

- GIVEN a staged file >1MB and a `.env` key missing from `.env.example` (ignoring comments and blanks)
- WHEN the pre-commit hook runs
- THEN yellow ⚠ warnings are printed
- AND the commit proceeds

#### Scenario: Missing .env skips silently

- GIVEN `.env` does not exist
- WHEN the env check runs
- THEN no warning is emitted

### Requirement: install-hooks

`make install-hooks` MUST copy `scripts/pre-commit.sh` to `.git/hooks/pre-commit` and make it executable.

#### Scenario: Hook install and uninstall are reversible

- GIVEN `make install-hooks` has run
- WHEN `make uninstall-hooks` runs
- THEN `.git/hooks/pre-commit` does not exist

### Requirement: uninstall-hooks

`make uninstall-hooks` MUST remove `.git/hooks/pre-commit`.

### Requirement: check-env

`make check-env` MUST run the `.env`/`.env.example` parity check.

### Requirement: README with setup instructions

The system MUST provide a `README.md` with: project description, prerequisites, `.env` setup, `docker-compose up` step, `make migrate-up`, `make run`, and a "Pre-commit Hooks" section documenting setup, checks, and bypass via `git commit --no-verify`. The commands table MUST include `install-hooks`, `uninstall-hooks`, and `check-env`.

#### Scenario: New developer onboarding

- GIVEN a developer clones the repository
- WHEN they follow README instructions
- THEN the API starts locally without additional undocumented steps

## Constraints and Edge Cases

- `docker-compose.yml` MUST use a named volume so data persists across container restarts.
- `.env.example` MUST NOT contain real secrets or passwords.
- `Makefile` MUST fail with a clear message if required tools (`go`, `docker-compose`, `goose`) are missing.
