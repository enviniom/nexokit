# Delta for error-handling

## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Typed sentinel errors

The system MUST define exported sentinel errors: `ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrConflict`, `ErrBadRequest`, `ErrValidation`, `ErrInternal`.

(Previously: Only `ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrConflict`, `ErrBadRequest`, `ErrInternal` existed; `ErrValidation` is new)

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
| ErrInternal | 500 |

(Previously: Map did not include `ErrValidation` → 422)

#### Scenario: Unknown error

- GIVEN an error without a known sentinel
- WHEN the HTTP layer handles it
- THEN it defaults to 500 with generic message