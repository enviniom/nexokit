# Error Handling Specification

## Purpose
Define typed application errors (`AppError`) and their mapping to HTTP status codes and standard API responses.

## Requirements

### Requirement: Typed sentinel errors

The system MUST define exported sentinel errors: `ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrConflict`, `ErrBadRequest`, `ErrInternal`.

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
| ErrInternal | 500 |

#### Scenario: Unknown error

- GIVEN an error without a known sentinel
- WHEN the HTTP layer handles it
- THEN it defaults to 500 with generic message

## Constraints and Edge Cases

- Sensitive details (stack traces, SQL errors) MUST NOT leak in production.
- `ErrInternal` SHOULD be logged with full details at ERROR level.
- Wrapped errors MUST be compatible with Go 1.13+ error chains.
