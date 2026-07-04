# Error Handling Specification

## Purpose
Define typed application errors (`AppError`) and their mapping to HTTP status codes and standard API responses.

## Requirements

### Requirement: Typed sentinel errors

The system MUST define exported sentinel errors: `ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrConflict`, `ErrBadRequest`, `ErrValidation`, `ErrInternal`.

#### Scenario: ErrNotFound

- GIVEN a repository returns `ErrNotFound`
- WHEN the error propagates to the HTTP layer
- THEN the response status is 404
- AND the envelope message is "Resource not found"

#### Scenario: ErrInternal

- GIVEN an unhandled database error
- WHEN it is wrapped as `ErrInternal`
- THEN the response status is 500
- AND in `production` the message is generic ("Internal server error")

#### Scenario: ErrValidation

- GIVEN a validation failure occurs in a service
- WHEN `ErrValidation` is checked with `errors.Is(err, apperror.ErrValidation)`
- THEN it returns `true`

### Requirement: ErrValidation sentinel

The system MUST define `ErrValidation` as an exported sentinel error in `platform/apperror` mapped to HTTP 422 Unprocessable Entity.

#### Scenario: ErrValidation returns 422

- GIVEN `ErrValidation` is returned from a service
- WHEN `apperror.Status(err)` is called
- THEN it returns 422

#### Scenario: HandleError maps ErrValidation to 422

- GIVEN `ErrValidation` is passed to `HandleError`
- WHEN the handler processes the error
- THEN the response status is 422 and the message is the validation error message

### Requirement: Error wrapping

`apperror.Wrap(err error, message string, cause ...error) *AppError` MUST continue to exist. When `err` is an `*AppError` or matches a known sentinel via `errors.Is`, `Wrap` MUST inherit the wrapped error's `Code` and `HTTPStatus` so `apperror.Status(wrapped)` returns the sentinel/AppError status and `errors.Is(wrapped, sentinel)` returns `true`. Otherwise `Wrap` MUST default to `Code: CodeInternal, HTTPStatus: 500`. The variadic `cause` arguments MUST form the unwrap chain after `err`. (Previously: `Wrap` set `Err`/`Cause`; an earlier revision made it always 500, breaking `apperror.Wrap(apperror.ErrUnprocessable, ...)` in `modules/iam/roles/delete_role/handler.go`. It is now status-preserving.)

#### Scenario: Wrap preserves sentinel status and code

- GIVEN `apperror.Wrap(apperror.ErrUnprocessable, core.MsgRoleHasAssignedUsers)`
- WHEN `apperror.Status(err)` is called
- THEN it returns 422
- AND `errors.Is(err, apperror.ErrUnprocessable)` and `errors.Is(err, apperror.CodeUnprocessable)` both return `true`

#### Scenario: Wrap preserves 404

- GIVEN `apperror.Wrap(apperror.ErrNotFound, "user 123 not found")`
- WHEN `errors.Is(err, apperror.ErrNotFound)` and `apperror.Status(err)` are checked
- THEN both return `true` and `404` respectively
- AND the unwrap chain still reaches the inner error when one is supplied via `cause`

#### Scenario: Wrap unknown error defaults to 500

- GIVEN `apperror.Wrap(errors.New("random failure"), "ctx")` where the inner error matches no known sentinel
- WHEN `apperror.Status(err)` runs
- THEN it returns 500
- AND `errors.Is(err, apperror.ErrInternal)` returns `true`
- AND the unwrap chain still reaches `"random failure"`

### Requirement: HTTP status mapping

The system MUST map each sentinel to a fixed HTTP status:

| Sentinel | Status |
|----------|--------|
| ErrNotFound | 404 |
| ErrForbidden | 403 |
| ErrUnauthorized | 401 |
| ErrConflict | 409 |
| ErrBadRequest | 400 |
| ErrValidation | 422 |
| ErrTooManyRequests | 429 |
| ErrInternal | 500 |

#### Scenario: Unknown error

- GIVEN an error without a known sentinel
- WHEN the HTTP layer handles it
- THEN it defaults to 500 with generic message

### Requirement: ErrTooManyRequests sentinel

The system MUST define `ErrTooManyRequests` as an exported sentinel error in `platform/apperror` mapped to HTTP 429 Too Many Requests.

#### Scenario: ErrTooManyRequests returns 429

- GIVEN `ErrTooManyRequests` is returned from a handler
- WHEN `apperror.Status(err)` is called
- THEN it returns 429

#### Scenario: HandleError maps ErrTooManyRequests to 429

- GIVEN `ErrTooManyRequests` is passed to `HandleError`
- WHEN the handler processes the error
- THEN the response status is 429 and the message is `MsgTooManyRequests`

### Requirement: ErrUnprocessable generic sentinel

The system MUST define `ErrUnprocessable` as an exported sentinel error in `platform/apperror` mapped to HTTP 422 Unprocessable Entity. The sentinel MUST NOT carry a domain-specific message — it serves as a generic 422 category for modules that need unprocessable semantics without owning a dedicated sentinel.

#### Scenario: ErrUnprocessable returns 422

- GIVEN `ErrUnprocessable` is returned from any layer
- WHEN `apperror.Status(err)` is called
- THEN it returns 422

#### Scenario: ErrUnprocessable has no domain message

- GIVEN `ErrUnprocessable` sentinel definition
- WHEN its `Message` field is inspected
- THEN it is empty or generic — never references roles, users, or other domain concepts

### Requirement: AppError canonical shape

`AppError` MUST be `struct { Code Code; HTTPStatus int; PublicMessage string; Internal error }` where `Code` is a typed string alias. `Code` is the `errors.Is` identity; `HTTPStatus` is set by helpers; `PublicMessage` is the only client-visible text; `Internal` is the unwrap chain.

#### Scenario: Helper sets Code and HTTPStatus

- GIVEN `apperror.NotFound(code, "foo", errors.New("db"))`
- WHEN the helper returns
- THEN `HTTPStatus == 404`, `Code == code`, `PublicMessage == "foo"`, `Internal.Error() == "db"`

#### Scenario: Internal is the unwrap source

- GIVEN an `AppError` whose `Internal` chains `fmt.Errorf("...: %w", originalErr)`
- WHEN `errors.Unwrap` iterates
- THEN the chain reaches `originalErr`

### Requirement: Platform HTTP-category codes

`apperror` MUST export `Code` constants for every HTTP category it owns (`CodeNotFound`, `CodeBadRequest`, `CodeForbidden`, `CodeConflict`, `CodeUnauthorized`, `CodeTooManyRequests`, `CodeValidation`, `CodeUnprocessable`, `CodeInternal`). Modules MUST own their own business `Code` constants.

#### Scenario: Platform code as fallback

- GIVEN a module wants a generic HTTP-category error without a domain code
- WHEN it calls `apperror.NotFound(apperror.CodeNotFound, "user not found", nil)`
- THEN `Code == apperror.CodeNotFound`

### Requirement: HTTP helper constructors

`apperror` MUST export helpers `(code Code, publicMsg string, internal error) *AppError`: `NotFound`, `BadRequest`, `Forbidden`, `Conflict`, `Unauthorized`, `TooManyRequests`, `Validation`, `Unprocessable`, `Internal`. A low-level `New(code Code, status int, publicMsg string, internal error) *AppError` MUST exist for uncommon statuses.

#### Scenario: Named helper sets status

- GIVEN a module calls `apperror.Conflict(apperror.CodeConflict, "taken", nil)`
- WHEN the helper returns
- THEN `HTTPStatus == 409` and `Code == apperror.CodeConflict`

### Requirement: PublicMessage redaction

`response.HandleError` MUST emit `ae.PublicMessage` when the error matches `*AppError`. Non-`AppError` errors MUST produce `messages.MsgInternalError` regardless of gin mode. The envelope MAY include `debug` only when the request context's `debug_errors` flag is true, which is derived from `AppConfig.Env` via `Config.ExposeDebugErrors()` and stored by `middleware.DebugErrors`.

#### Scenario: Known AppError uses its PublicMessage

- GIVEN `apperror.NotFound(code, "user not found", errors.New("db boom"))` is returned
- WHEN the envelope is built in release mode
- THEN `message == "user not found"` and the client never sees `"db boom"`

#### Scenario: Unknown error redacted when debug is disabled

- GIVEN `errors.New("database is down")` reaches the handler
- WHEN the request context has `debug_errors == false`
- THEN `message == messages.MsgInternalError` and the response has no `debug` field

### Requirement: errors.Is preserved by Code equality

`AppError.Is(target error) bool` MUST match by pointer identity OR by `Code` equality when both sides have non-empty `Code`. With `Internal` set, the comparison MUST fall through to `errors.Is(Internal, target)`.

#### Scenario: Code equality match

- GIVEN an `AppError{Code: apperror.CodeNotFound, ...}` is built via `apperror.NotFound`
- WHEN `errors.Is(err, apperror.ErrNotFound)` is checked
- THEN it returns `true` because the codes match

#### Scenario: errors.As recovers AppError from wrapped chain

- GIVEN `fmt.Errorf("layer: %w", apperror.NotFound(code, "foo", nil))` is built
- WHEN `errors.As(err, &ae)` is checked
- THEN `ae.PublicMessage == "foo"`

### Requirement: Validation response flow is separate

DTOs return `response.ValidationErrors`; handlers call `response.ValidationError(c, errs)` which writes HTTP 422 with field-keyed `errors`. This path MUST NOT route through `AppError` or `HandleError`.

#### Scenario: ValidationErrors writes 422

- GIVEN a DTO returns `response.ValidationErrors{"email": []string{"invalid"}}`
- WHEN the handler calls `response.ValidationError(c, errs)`
- THEN the response status is 422 and `errors.email[0] == "invalid"`

### Requirement: Module-owned business `Code` format

Modules MUST declare business `Code` constants in the form `code:<snake_case>` (e.g. `code:user_not_found`, `code:company_slug_taken`). The `<snake_case>` segment MUST be unique across all modules and MUST use lowercase letters, digits, and underscores. Platform HTTP-category `Code` constants (`apperror.CodeNotFound`, `apperror.CodeConflict`, etc.) remain prefix-free and MUST NOT be reused as module codes.

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

## Constraints and Edge Cases

- `PublicMessage` is the only client-visible text; `Internal` error details MUST NOT be exposed to clients.
- Debug exposure is config-driven via `Config.ExposeDebugErrors()` / the `messages.CtxDebugErrors` request-context flag.
- `errors.Is` by `Code` equality preserves sentinel-match compatibility; empty-`Code` values do not overmatch.
- Validation responses (`response.ValidationError`) MUST NOT route through `AppError` or `HandleError`.
- Wrapped errors MUST be compatible with Go 1.13+ error chains.
