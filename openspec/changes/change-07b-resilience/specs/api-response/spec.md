# Delta for API Response

## ADDED Requirements

### Requirement: TooManyRequests response helper

The system MUST provide a `TooManyRequests(c *gin.Context, message string)` helper that returns HTTP 429 with the standard error envelope using `MsgTooManyRequests` as the default message.

#### Scenario: TooManyRequests returns standard envelope

- GIVEN a rate limit is exceeded
- WHEN `TooManyRequests(c, "")` is called
- THEN the response is 429 with `success: false`, `message: MsgTooManyRequests`, `data: null`, `errors: null`

#### Scenario: TooManyRequests with custom message

- GIVEN a rate limit is exceeded
- WHEN `TooManyRequests(c, "Please wait before retrying")` is called
- THEN the response is 429 with `message: "Please wait before retrying"`
