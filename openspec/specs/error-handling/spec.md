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

The system MUST allow wrapping a sentinel with a custom message while preserving unwrap.

#### Scenario: Wrap with context

- GIVEN `apperror.Wrap(ErrNotFound, "user 123 not found")`
- WHEN `errors.Is(err, ErrNotFound)` is checked
- THEN it returns `true`

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

## Constraints and Edge Cases

- Sensitive details (stack traces, SQL errors) MUST NOT leak in production.
- `ErrInternal` SHOULD be logged with full details at ERROR level.
- Wrapped errors MUST be compatible with Go 1.13+ error chains.
