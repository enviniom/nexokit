# Delta for Error Handling

## ADDED Requirements

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

### Requirement: HTTP status mapping for 429

The system MUST include `ErrTooManyRequests` in the sentinel-to-status mapping table.

| Sentinel | Status |
|----------|--------|
| ErrTooManyRequests | 429 |

#### Scenario: Unknown error still defaults to 500

- GIVEN an error without a known sentinel
- WHEN the HTTP layer handles it
- THEN it defaults to 500 with generic message
