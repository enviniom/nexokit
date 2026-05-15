# Platform Stubs Specification

## Purpose
Define the placeholder packages and helper files that establish folder structure and compilation compatibility for future changes.

## Requirements

### Requirement: Empty platform stubs

The system MUST create empty but compilable package files for:
- `internal/platform/query`
- `internal/platform/identity`
- `internal/platform/password`
- `internal/platform/token`

Each stub MUST contain a package declaration and a `// TODO: implement in future change` comment.

#### Scenario: Build succeeds with stubs

- GIVEN stub files exist in all four packages
- WHEN `go build ./...` runs
- THEN compilation succeeds with no errors

### Requirement: CLI stub

The system MUST create `internal/cli/` with a minimal package file and `cmd/nexokit/main.go` as a CLI entry point stub.

#### Scenario: CLI compiles

- GIVEN `cmd/nexokit/main.go` exists
- WHEN `go build ./cmd/nexokit` runs
- THEN it produces a binary

### Requirement: Modules directory structure

The system MUST create `internal/modules/` with example subdirectories (e.g., `auth/`, `users/`, `roles/`, `companies/`) containing minimal stub files.

#### Scenario: Module stubs compile

- GIVEN `internal/modules/auth/` contains `model.go`, `handler.go`, etc.
- WHEN `go build ./...` runs
- THEN all module packages compile

### Requirement: Test helper NewTestApp

The system MUST provide `tests/helpers/app.go` with `NewTestApp() *app.App` that bootstraps a test instance using test configuration. All comments MUST be in English.

#### Scenario: Integration test setup

- GIVEN `NewTestApp()` is called in a test
- WHEN it returns
- THEN the `App` has an in-memory or test-database configuration
- AND the server is not started automatically

## Constraints and Edge Cases

- Stub files MUST compile but MUST NOT contain production logic.
- `tests/helpers/` MUST be importable by integration tests without creating import cycles.
- English comments in `tests/helpers/` ensure consistency for international contributors.
