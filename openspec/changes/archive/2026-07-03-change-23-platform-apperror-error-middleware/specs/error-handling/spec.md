# Delta for error-handling

## Purpose

Reshape `AppError` to a code-first struct with HTTP helpers, encode regression-safe redaction, and preserve `errors.Is`/`errors.As` for existing sentinels. Module-owned `Code` migration is Change 18 (out of scope).

## ADDED Requirements

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

## MODIFIED Requirements

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
