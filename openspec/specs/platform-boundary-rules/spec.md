# Platform Boundary Rules

## Purpose

Define what belongs in `internal/platform/*` (cross-application contracts) vs `internal/modules/*/core/*` (domain-specific language). Platform packages MUST NOT contain domain-specific messages, constants, or error sentinels.

## Requirements

### Requirement: Platform package classification

The system MUST classify each `platform` subpackage as either **generic** (cross-application utility) or **domain-restricted** (must not accept domain-specific content):

| Subpackage | Classification | Allowed Content |
|------------|---------------|-----------------|
| `platform/messages` | Generic | API response messages, validation rule messages, middleware messages — zero domain-specific messages |
| `platform/apperror` | Generic | Sentinel errors mapped to HTTP statuses — messages MUST be generic, not domain-specific |
| `platform/permissions` | Generic | `Action*` constants, `Format()`, `Humanize*()`, `Register()`, `ListRegistered()` — zero `Module*` constants |
| `platform/response` | Generic | Single API response contract source — `APIResponse`, `HandleError`, helpers; imports `validator` for `ValidationErrors` |
| `platform/tenant` | Generic | Tenant context and scoping utilities |
| `platform/identity` | Generic | Public ID generation |
| `platform/token` | Generic | Token utilities |
| `platform/password` | Generic | Password hashing/validation |
| `platform/validator` | Generic | Validation primitives: `ValidationErrors`, `FieldValidator`, `Rule` functions — owns the error accumulator |
| `platform/gormutil` | Generic | GORM helpers |
| `platform/query` | Generic | Query parsing |
| `platform/authctx` | Generic | Auth context utilities |
| `platform/cache` | Generic | Cache adapter (if present) |

#### Scenario: No domain messages in platform/messages

- GIVEN `platform/messages/messages.go`
- WHEN all message constants are reviewed
- THEN zero messages reference domain concepts (roles, users, companies, etc.)

#### Scenario: No module constants in platform/permissions

- GIVEN `platform/permissions/constants.go`
- WHEN all constants are reviewed
- THEN zero `Module*` constants exist — only `Action*` constants and utility functions remain

#### Scenario: Generic sentinel messages in platform/apperror

- GIVEN `platform/apperror/apperror.go`
- WHEN all sentinel error messages are reviewed
- THEN zero messages reference domain concepts — `ErrUnprocessable` has no role-specific message

### Requirement: Module-owned domain language

Each module MUST define its own domain-specific language in `modules/<name>/core/`:

| Artifact | Location | Example |
|----------|----------|---------|
| Module name constant | `core/constants.go` | `const ModuleIAM = "iam"` |
| Domain error messages | `core/messages.go` | `MsgRoleHasAssignedUsers` |
| Domain error sentinels | `core/error.go` | `ErrRoleHasAssignedUsers` |

After legacy removal, the IAM module is the sole owner of user, role, and permission domain language. Legacy module paths (`modules/users/core/`, `modules/roles/core/`, `modules/permissions/core/`) no longer exist.
(Previously: Examples referenced `modules/users/core/constants.go` and `modules/roles/core/error.go`; those directories are deleted — IAM owns all user/role/permission domain language)

#### Scenario: IAM module owns domain language

- GIVEN the IAM module at `modules/iam/core/`
- WHEN domain constants, messages, and error sentinels are reviewed
- THEN they are defined within the IAM module's core package

#### Scenario: Module owns its error sentinel

- GIVEN the IAM module needs a 422 error for "role has assigned users"
- WHEN `modules/iam/core/error.go` exists
- THEN it defines `ErrRoleHasAssignedUsers` mapping to HTTP 422 with `MsgRoleHasAssignedUsers`

#### Scenario: No legacy module domain language references

- GIVEN the full codebase after legacy removal
- WHEN imports and references are reviewed
- THEN zero paths reference `modules/users/core/`, `modules/roles/core/`, or `modules/permissions/core/`

### Requirement: platform/response as single response contract

The system MUST use `platform/response` as the sole source of API response contracts. No module MAY define its own response envelope or `HandleError` equivalent.

#### Scenario: All handlers use platform/response

- GIVEN any HTTP handler in any module
- WHEN it writes a response
- THEN it uses `platform/response` helpers (`Success`, `Error`, `NoContent`, `HandleError`)

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

## Constraints and Edge Cases

- `Action*` constants in `platform/permissions` MUST remain — they are generic permission verbs used across all modules
- Module constant duplication across modules IS acceptable — it avoids coupling between unrelated modules
- Moving domain language to modules MUST NOT change HTTP response shapes, status codes, or JSON envelopes
