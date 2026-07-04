# Delta for gorm-helpers

## Purpose

Extend `internal/platform/gormutil` with `IsUniqueConstraintError` so repositories can detect unique-constraint violations without re-implementing substring matching or importing `gorm.io/gorm` from service code. The helper lives next to the existing GORM scopes, NOT in `internal/platform/shared/`.

## ADDED Requirements

### Requirement: IsUniqueConstraintError

`internal/platform/gormutil` MUST export `IsUniqueConstraintError(err error) bool`. The helper MUST return `true` when `err` matches `gorm.ErrDuplicatedKey` via `errors.Is`, OR when `err`'s lower-cased message indicates a unique-constraint violation across supported drivers. This includes Postgres-style duplicate/unique messages (`duplicate key`, `unique constraint`) and SQLite/current behavior messages such as `unique failed`, `constraint failed`, or `UNIQUE constraint failed`. The helper MUST return `false` for other SQL/connection errors and for `nil`. The helper MUST NOT panic on `nil` input.
(Previously: each repository defined its own `isUniqueConstraintError` helper using `strings.ToLower(err.Error())` and substring matching; the pattern was duplicated in three IAM slice repositories.)

#### Scenario: Detects gorm.ErrDuplicatedKey

- GIVEN `err == gorm.ErrDuplicatedKey`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Detects Postgres unique violation message

- GIVEN an error whose message is `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Detects Postgres unique constraint message

- GIVEN an error whose message contains `unique constraint`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Detects SQLite unique constraint failure

- GIVEN an error whose message is `UNIQUE constraint failed: users.email`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Detects SQLite generic constraint failure form

- GIVEN an error whose lower-cased message contains `constraint failed` or `unique failed`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Returns false for generic SQL error

- GIVEN a SQL error that does not indicate a unique-constraint violation
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `false`

#### Scenario: Returns false for unrelated errors

- GIVEN an arbitrary error such as `errors.New("connection refused")`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `false`

#### Scenario: Returns false for nil

- GIVEN `err == nil`
- WHEN `gormutil.IsUniqueConstraintError(nil)` is called
- THEN it returns `false`
- AND it does NOT panic

#### Scenario: Has table-driven unit test

- GIVEN the `gormutil` package
- WHEN the test suite runs
- THEN a table-driven test covers at least: `gorm.ErrDuplicatedKey`, Postgres duplicate-key and unique-constraint messages, SQLite unique-constraint failure, a generic SQL error, a connection error, and `nil`
- AND each row asserts the expected boolean result
