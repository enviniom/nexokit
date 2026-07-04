# Delta for platform-boundary-rules

## Purpose

Pin the layer-specific boundary rules the existing contract already implies but does not pin: services MUST NOT import GORM or `platform/apperror`; handlers MUST NOT remap sentinels; repositories MUST translate persistence errors; pure repeated helpers MUST live in `internal/platform/shared/string`; module test coverage MUST be enforced; a CI grep guard MUST forbid `apperror.` in non-test handler/service/repository files.

## MODIFIED Requirements

### Requirement: Service layer boundaries

A `service.go` file MUST NOT import `gorm.io/gorm` and MUST NOT import `github.com/enviniom/nexokit/internal/platform/apperror`. A service MUST return a reusable module error from `core/errors.go` OR a wrapped internal error (`fmt.Errorf("...: %w", err)`). A service MUST NOT construct ad-hoc `apperror` values inline.

#### Scenario: Service file has no GORM or apperror import

- GIVEN any `internal/modules/*/*/service.go`
- WHEN the file's imports are inspected
- THEN the import set MUST NOT contain `gorm.io/gorm` or `platform/apperror`

#### Scenario: Service returns module error or wrapped internal

- GIVEN a service encounters a persistence failure
- WHEN it returns
- THEN it returns a sentinel from `core/errors.go` OR `fmt.Errorf("...: %w", err)`
- AND it MUST NOT construct a new `apperror.NotFound(...)` / `apperror.Conflict(...)` value inline

### Requirement: Handler layer boundaries

A `handler.go` file MUST funnel every business / app error through `response.HandleError(c, err)`. A handler MUST NOT define a `mapServiceError(err) error` switch or import `platform/apperror` for the purpose of error remapping. Field-level validation errors are routed through `response.RespondIfInvalid` / `response.ValidationError`.

#### Scenario: No mapServiceError in any handler

- GIVEN the full module tree after change-24 is complete
- WHEN the tree is grep-searched for `mapServiceError` outside `_test.go`
- THEN the search returns no matches

### Requirement: Repository translates persistence errors

A repository MUST translate `gorm.ErrRecordNotFound` to the matching module sentinel before returning; it MUST NOT propagate `gorm.ErrRecordNotFound` to the service layer. A repository MUST detect unique-constraint violations via `gormutil.IsUniqueConstraintError` and map them to the matching module conflict sentinel.

#### Scenario: Repository maps not-found and unique violation

- GIVEN a `GetByPublicID` against an empty table OR a `Create` against a unique-indexed column
- WHEN GORM returns `gorm.ErrRecordNotFound` OR a unique-constraint violation
- THEN the repository returns the matching module sentinel
- AND the service layer never sees raw GORM errors

### Requirement: Shared pure helpers in platform/shared

Pure repeated helpers used by 2 or more modules MUST live in `internal/platform/shared/string`. Module-local helpers MUST stay in that module's package. The existing `internal/platform/gormutil` package continues to own GORM-specific helpers; pure GORM-agnostic helpers MUST NOT be duplicated there. `iam/queries/normalize_slugs.go` MUST stay in place because it is a single-use plural list de-dup helper, distinct from singular shared `NormalizeSlug`.

#### Scenario: Normalize helpers live in platform/shared/string

- GIVEN `NormalizeSlug`, `NormalizeDomain`, and `NormalizeEmail` are used by 2 or more modules
- WHEN the change is complete
- THEN the three helpers live in `internal/platform/shared/string/`
- AND no module defines its own duplicate
- AND `iam/queries/normalize_slugs.go` remains module-local

### Requirement: Module test coverage contract

Each module MUST own `core/errors_test.go` that pins the `Code`, `HTTPStatus`, and `PublicMessage` of every declared sentinel. Each module that owns request DTOs MUST own `core/dto_test.go` with table-driven coverage of every `Validate()` rule. Each module that defines a partial GORM model with a non-default table name MUST own a direct `TableName()` unit test.

#### Scenario: All required core test files exist

- GIVEN a module at `internal/modules/<name>/`
- WHEN the change is complete
- THEN `core/errors_test.go` exists
- AND `core/dto_test.go` exists if the module owns request DTOs
- AND a per-model `<model>_test.go` exists if the module defines a partial GORM model

### Requirement: CI grep guard for apperror in handlers, services, and repositories

A CI grep guard MUST fail the build when `apperror.` appears in any `*service.go`, `*repository.go`, or non-test `*handler.go` under `internal/modules/`. The guard is additive and does not relax the file-level import contract.

#### Scenario: Grep guard fails on apperror

- GIVEN a `*service.go`, `*repository.go`, or `*handler.go` is modified to import `apperror` for error construction or remapping
- WHEN the CI grep guard runs `grep -RE 'apperror\.' internal/modules/ --include='*service.go' --include='*repository.go' --include='*handler.go' | grep -v _test.go`
- THEN the guard returns a non-zero exit code
