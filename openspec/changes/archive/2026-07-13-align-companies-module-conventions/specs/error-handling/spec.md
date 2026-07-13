# Delta for Error Handling

## ADDED Requirements

### Requirement: Companies repository errors are module-owned AppErrors

The companies module MUST rename `internal/modules/companies/core/error.go` to `internal/modules/companies/core/errors.go`. Repository interfaces MUST continue returning `error`, but every non-nil persistence failure returned by a companies repository MUST be a module-owned `*apperror.AppError` created through an entity-specific unary mapper in `internal/modules/companies/queries/map_errors.go`. Unknown persistence failures MUST preserve the original cause in the AppError unwrap chain. Raw GORM, SQL, or driver errors MUST NOT cross the repository boundary.
(Previously: repositories mixed raw GORM, inline not-found handling, and unique-constraint checks.)

#### Scenario: Unknown failure preserves cause

- GIVEN a companies repository hits an unexpected database failure
- WHEN the error is translated
- THEN `errors.As` finds a module-owned `*apperror.AppError`
- AND `errors.Is` still reaches the original cause
- AND the repository does not return the raw failure

#### Scenario: Known persistence outcomes stay typed

- GIVEN a repository query does not find a company or domain
- WHEN the mapper runs
- THEN it returns the matching module not-found AppError
- AND `errors.Is` matches that sentinel

#### Scenario: Canonical error file exists

- GIVEN the companies module core package
- WHEN the directory is inspected
- THEN `core/errors.go` exists
- AND `core/error.go` does not exist
