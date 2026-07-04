# GORM Helpers Specification

## Purpose
Define pure-function GORM scopes for pagination, sorting, search, date-range, and status filtering. Each helper accepts a `*gorm.DB` and relevant params, returning a `*gorm.DB` for composability.

## Requirements

### Requirement: ApplyPagination

The system MUST provide `ApplyPagination(db *gorm.DB, page, perPage int) *gorm.DB` that applies `OFFSET` and `LIMIT` based on 1-indexed page.

#### Scenario: Second page of 15 items

- GIVEN page=2, perPage=15
- WHEN `ApplyPagination(db, 2, 15)` is called
- THEN the query includes `OFFSET 15 LIMIT 15`

### Requirement: ApplySorting

The system MUST provide `ApplySorting(db *gorm.DB, sort SortParams) *gorm.DB` that applies `ORDER BY` only when `Sort` is non-empty.

#### Scenario: Sort by name ascending

- GIVEN `SortParams{Sort: "name", Order: "asc"}`
- WHEN `ApplySorting(db, params)` is called
- THEN the query includes `ORDER BY name asc`

#### Scenario: Empty sort is no-op

- GIVEN `SortParams{Sort: ""}`
- WHEN `ApplySorting(db, params)` is called
- THEN the query is unchanged

### Requirement: ApplySearch

The system MUST provide `ApplySearch(db *gorm.DB, search SearchParams, columns ...string) *gorm.DB` that applies `WHERE col1 LIKE ? OR col2 LIKE ?` when `Query` is non-empty.

#### Scenario: Search across name and email

- GIVEN `SearchParams{Query: "john"}` and columns `["name", "email"]`
- WHEN `ApplySearch(db, params, "name", "email")` is called
- THEN the query includes `WHERE name LIKE '%john%' OR email LIKE '%john%'`

#### Scenario: Empty query is no-op

- GIVEN `SearchParams{Query: ""}`
- WHEN `ApplySearch(db, params, "name")` is called
- THEN the query is unchanged

### Requirement: ApplyDateRange

The system MUST provide `ApplyDateRange(db *gorm.DB, filters FilterParams, column string) *gorm.DB` that applies `created_at >= ?` and/or `created_at <= ?` when date pointers are non-nil.

#### Scenario: Both from and to dates

- GIVEN `FilterParams{CreatedFrom: &from, CreatedTo: &to}` and column `"created_at"`
- WHEN `ApplyDateRange(db, filters, "created_at")` is called
- THEN the query includes `WHERE created_at >= ? AND created_at <= ?`

#### Scenario: Only from date

- GIVEN `FilterParams{CreatedFrom: &from, CreatedTo: nil}`
- WHEN `ApplyDateRange(db, filters, "created_at")` is called
- THEN the query includes `WHERE created_at >= ?` only

### Requirement: ApplyStatusFilter

The system MUST provide `ApplyStatusFilter(db *gorm.DB, filters FilterParams, column string) *gorm.DB` that applies `WHERE column = ?` when `Status` is non-empty.

#### Scenario: Active status filter

- GIVEN `FilterParams{Status: "active"}` and column `"status"`
- WHEN `ApplyStatusFilter(db, filters, "status")` is called
- THEN the query includes `WHERE status = 'active'`

#### Scenario: Empty status is no-op

- GIVEN `FilterParams{Status: ""}`
- WHEN `ApplyStatusFilter(db, filters, "status")` is called
- THEN the query is unchanged

### Requirement: IsUniqueConstraintError

`internal/platform/gormutil` MUST export `IsUniqueConstraintError(err error) bool`. The helper MUST return `true` when `err` matches `gorm.ErrDuplicatedKey` via `errors.Is`, OR when `err`'s lower-cased message indicates a unique-constraint violation across supported drivers. This includes Postgres-style duplicate/unique messages (`duplicate key`, `unique constraint`) and SQLite/current behavior messages such as `unique failed` or `UNIQUE constraint failed`. The helper MUST return `false` for other SQL/connection errors, non-unique constraint failures (for example foreign-key or check constraints), and for `nil`. The helper MUST NOT panic on `nil` input.

#### Scenario: Detects gorm.ErrDuplicatedKey

- GIVEN `err == gorm.ErrDuplicatedKey`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Detects Postgres unique violation message

- GIVEN an error whose message is `ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Detects SQLite unique constraint failure

- GIVEN an error whose message is `UNIQUE constraint failed: users.email`
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `true`

#### Scenario: Returns false for generic SQL error

- GIVEN a SQL error that does not indicate a unique-constraint violation
- WHEN `gormutil.IsUniqueConstraintError(err)` is called
- THEN it returns `false`

#### Scenario: Returns false for non-unique constraint failure

- GIVEN a SQL error indicating a foreign-key, check, or not-null constraint failure
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
- THEN a table-driven test covers at least: `gorm.ErrDuplicatedKey`, Postgres duplicate-key and unique-constraint messages, SQLite unique-constraint failure, a generic SQL error, a non-unique constraint failure, a connection error, and `nil`
- AND each row asserts the expected boolean result

### Requirement: Soft delete compatibility

GORM helpers MUST preserve GORM's default soft delete scope for models using `gorm.DeletedAt`; they MUST NOT call `Unscoped()` or include deleted rows implicitly.

#### Scenario: Helper composition excludes soft-deleted rows

- GIVEN a model embeds `BaseModel` with `DeletedAt`
- WHEN pagination, sorting, search, date, or status helpers are applied
- THEN normal GORM soft delete filtering remains active
