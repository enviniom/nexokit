# Migrations Specification

## Purpose
Define the Goose migration system: file naming, directory layout, Makefile targets, and execution behavior.

## Requirements

### Requirement: Migration file format

The system MUST store migration files in `migrations/` using the format `YYYYMMDDHHMMSS_description.sql`.

#### Scenario: Create new migration

- GIVEN `make migrate-create NAME=add_users`
- WHEN the command executes
- THEN a file `migrations/YYYYMMDDHHMMSS_add_users.sql` is created
- AND it contains empty `-- +goose Up` and `-- +goose Down` sections

### Requirement: Makefile targets

The system MUST provide Makefile targets: `migrate-up`, `migrate-down`, `migrate-create`, `migrate-status`.

#### Scenario: Run pending migrations

- GIVEN pending migrations exist in `migrations/`
- WHEN `make migrate-up` executes
- THEN all pending migrations apply in timestamp order
- AND `goose_db_version` is updated

#### Scenario: No pending migrations

- GIVEN all migrations are already applied
- WHEN `make migrate-up` executes
- THEN it exits with code 0
- AND produces output indicating no changes

#### Scenario: Revert last migration

- GIVEN at least one migration is applied
- WHEN `make migrate-down` executes
- THEN exactly one most-recent migration is rolled back

#### Scenario: Check status

- GIVEN migrations exist
- WHEN `make migrate-status` executes
- THEN it lists applied and pending migrations

### Requirement: Migration failure handling

The system MUST fail fast when a migration contains invalid SQL.

#### Scenario: Malformed SQL

- GIVEN a migration file with invalid SQL
- WHEN `make migrate-up` executes
- THEN the command exits with a non-zero code
- AND no subsequent migrations are applied

## Constraints and Edge Cases

- Timestamps MUST be UTC and sortable lexicographically.
- Concurrent `make migrate-up` runs SHOULD be safe (Goose uses advisory locks).
- The `migrations/` directory MUST be committed to version control.
- Down migrations MUST be provided for every up migration.
