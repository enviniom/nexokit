# Delta for Dev Environment

## MODIFIED Requirements

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

(Previously: Only `test` target existed; now includes `test-unit`, `test-integration`, `test-coverage`)
