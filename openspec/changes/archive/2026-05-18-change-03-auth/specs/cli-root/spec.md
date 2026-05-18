# CLI Root Specification

## Purpose

Idempotent root-user creation via the internal CLI, wired to real storage and password hashing.

## Requirements

### Requirement: Real storage and hasher wiring

The `create-root` command MUST use the real `RootStorage` and `PasswordHasher` implementations. It MUST NOT return `ErrStorageNotWired` or leave unresolved TODOs in `internal/cli/root/root.go`.

#### Scenario: Command uses real dependencies

- GIVEN the application container is initialized with user/role storage and argon2id hasher
- WHEN `go run ./cmd/nexokit create-root` executes
- THEN it delegates to the concrete storage and hasher without stub errors

### Requirement: Idempotent root creation

The command MUST be idempotent: if a root user already exists, it MUST exit with code 0 and print a message indicating the user already exists.

#### Scenario: Create root when absent

- GIVEN no root user exists
- WHEN `go run ./cmd/nexokit create-root` executes
- THEN a root user is created with the `root` role and a secure password hash

#### Scenario: Idempotent re-run

- GIVEN a root user already exists
- WHEN `go run ./cmd/nexokit create-root` executes again
- THEN the command exits with code 0 and does not create a duplicate

### Requirement: Root credential input

The command MUST accept root credentials via environment variables (`ROOT_USER_NAME`, `ROOT_USER_EMAIL`, `ROOT_USER_PASSWORD`). If variables are absent, it MAY fall back to interactive input or fail gracefully. It MUST NOT use a hardcoded password.

#### Scenario: Create root via environment variables

- GIVEN `ROOT_USER_NAME`, `ROOT_USER_EMAIL`, and `ROOT_USER_PASSWORD` are set
- WHEN `go run ./cmd/nexokit create-root` executes
- THEN the root user is created with the provided values

#### Scenario: Missing credentials

- GIVEN no environment variables are set and interactive input is not available
- WHEN the command runs
- THEN it exits with a non-zero code and a clear error message
