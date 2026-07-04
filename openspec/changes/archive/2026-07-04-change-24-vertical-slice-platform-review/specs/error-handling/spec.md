# Delta for error-handling

## Purpose

Pin the module-owned `AppError` `Code` format and the module-level error test contract that change-23 left as out-of-scope. Pure refactor onto the existing `error-handling` contract: no public surface, route, payload, or HTTP status change.

## MODIFIED Requirements

### Requirement: Module-owned business `Code` format

Modules MUST declare business `Code` constants in the form `code:<snake_case>` (e.g. `code:user_not_found`, `code:company_slug_taken`). The `<snake_case>` segment MUST be unique across all modules and MUST use lowercase letters, digits, and underscores. Platform HTTP-category `Code` constants (`apperror.CodeNotFound`, `apperror.CodeConflict`, etc.) remain prefix-free and MUST NOT be reused as module codes.
(Previously: platform owned only HTTP-category `Code` constants; module-owned business `Code` format was not pinned.)

#### Scenario: Module code uses code: prefix

- GIVEN a module declares a business sentinel such as `ErrUserNotFound`
- WHEN the corresponding `Code` constant is inspected
- THEN it equals `code:user_not_found` (prefix `code:` + snake_case identifier)

#### Scenario: Code format is enforced by tests

- GIVEN any module's `core/errors.go`
- WHEN `core/errors_test.go` is executed
- THEN every declared `Code` constant MUST start with the literal `code:` and a snake_case suffix
- AND any constant that fails the rule MUST cause the test to fail

### Requirement: Module sentinel test coverage

Each module's `core/errors.go` MUST have a corresponding `core/errors_test.go` that pins the `apperror` kind, the module-owned `Code`, the `PublicMessage`, and the `HTTPStatus` for every declared sentinel. Repositories MUST have a test that maps `gorm.ErrRecordNotFound` to the matching module sentinel. Services MUST have a test that returns the matching module sentinel and wraps internal errors with `fmt.Errorf("...: %w", err)`.
(Previously: error-sentinel coverage was informal; no pinned contract for module tests.)

#### Scenario: Each declared sentinel is covered

- GIVEN a module's `core/errors.go` declares `ErrNotFound = apperror.NotFound(CodeUserNotFound, "user not found", nil)`
- WHEN `core/errors_test.go` is executed
- THEN the test asserts `apperror.Status(ErrNotFound) == 404`, `ErrNotFound.Code == "code:user_not_found"`, and `ErrNotFound.PublicMessage == "user not found"`

#### Scenario: Repository maps not-found to module error

- GIVEN a repository test seeds an empty database and calls `GetByPublicID("missing")`
- WHEN the query returns `gorm.ErrRecordNotFound`
- THEN the repository returns the module's `core.ErrNotFound` sentinel
- AND `apperror.Status(err) == 404`

#### Scenario: Service returns module sentinel, wraps internals

- GIVEN a service method encounters a database failure that is not `ErrRecordNotFound`
- WHEN the service returns
- THEN it MUST return a module sentinel from `core/errors.go` OR wrap the internal error with `fmt.Errorf("...: %w", err)` so the unwrap chain still reaches the original failure

### Requirement: Handler funnels business errors through HandleError

Handlers MUST route every business / app error to `response.HandleError(c, err)` exactly once. Handlers MUST NOT import `platform/apperror` for the purpose of remapping module sentinels. Handlers MUST NOT contain a per-handler `mapServiceError` switch that re-maps module `core.Err*` values to `apperror.Err*` values, because module sentinels are already `*AppError` instances and `response.HandleError` already handles them.
(Previously: handlers were allowed to import `apperror` and re-map sentinels; the platform contract did not forbid it.)

#### Scenario: Handler does not import apperror for remapping

- GIVEN a slice handler file `handler.go`
- WHEN its imports are inspected
- THEN `github.com/enviniom/nexokit/internal/platform/apperror` MUST NOT appear
- AND any `apperror.Err*` literal in the handler MUST cause a CI grep guard to fail

#### Scenario: Handler has no mapServiceError switch

- GIVEN any slice `handler.go` after change-24 is complete
- WHEN the file is grep-searched for `mapServiceError`
- THEN the identifier MUST NOT appear

#### Scenario: Business error reaches HandleError unchanged

- GIVEN a service returns `core.ErrUserNotFound`
- WHEN the handler calls `response.HandleError(c, err)`
- THEN the response status is 404, the envelope `message` is the sentinel's `PublicMessage`, and the `code` is the sentinel's `Code`
- AND the HTTP status, envelope, and payload are identical to the pre-change-24 behavior
