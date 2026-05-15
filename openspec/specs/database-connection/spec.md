# Database Connection Specification

## Purpose
Define PostgreSQL connection behavior via GORM, including pool configuration and explicit prohibition of AutoMigrate.

## Requirements

### Requirement: Open PostgreSQL connection via GORM

The system MUST open a PostgreSQL connection using GORM with the driver `gorm.io/driver/postgres`.

#### Scenario: Valid credentials

- GIVEN a `Config` with valid database credentials
- WHEN `db.Connect(cfg)` is called
- THEN it returns a `*gorm.DB` and no error

#### Scenario: Invalid credentials

- GIVEN a `Config` with invalid `DB_PASSWORD`
- WHEN `db.Connect(cfg)` is called
- THEN it returns a non-nil error
- AND the returned `*gorm.DB` is nil

### Requirement: Configure connection pool

The system MUST configure the underlying `sql.DB` pool: `SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime`.

#### Scenario: Pool defaults

- GIVEN a successful connection
- WHEN the pool is inspected
- THEN `MaxOpenConns` is greater than zero
- AND `MaxIdleConns` is greater than zero

### Requirement: Prohibit AutoMigrate

The system MUST NOT call `AutoMigrate` anywhere in the codebase. Schema changes MUST go through Goose migrations only.

#### Scenario: Code search for AutoMigrate

- GIVEN the codebase is searched for `AutoMigrate`
- WHEN `grep -r AutoMigrate` runs
- THEN zero occurrences are found outside documentation

## Constraints and Edge Cases

- Connection timeout MUST be reasonable (default DSN behavior).
- On startup failure, the error message MUST include which config field failed (host, port, etc.) without leaking the password.
- `gorm.DB` instance MUST be shared across the application via dependency injection.
