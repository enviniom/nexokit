# Middleware Auth Specification

## Purpose

PASETO access-token validation, user lookup, and request-context injection.

## Requirements

### Requirement: Token validation

The auth middleware MUST validate the PASETO access token from the `Authorization` header (`Bearer <token>`). It MUST reject missing, malformed, or expired tokens with HTTP 401 and a standard DTO error.

#### Scenario: Valid token

- GIVEN a request with a valid, non-expired PASETO access token
- WHEN the request reaches a protected route
- THEN the middleware allows the request to proceed

#### Scenario: Missing token

- GIVEN a request without an `Authorization` header
- WHEN the request reaches a protected route
- THEN the response returns HTTP 401 and `success: false`

#### Scenario: Expired token

- GIVEN a request with an expired PASETO access token
- WHEN the request reaches a protected route
- THEN the response returns HTTP 401 with an invalid or expired token message

### Requirement: User lookup and context injection

The middleware MUST extract the `sub` claim from the token, lookup the corresponding user by `public_id`, and inject the user into the request context. The injected user MUST NOT contain `password` or `password_hash`.

#### Scenario: User injected into context

- GIVEN a valid token for an active user
- WHEN the request reaches a protected handler
- THEN the handler can retrieve the full user object from the context without password fields

### Requirement: Inactive user rejection

The middleware MUST reject requests for users whose `status` is `inactive`.

#### Scenario: Inactive user blocked

- GIVEN a valid token belonging to an inactive user
- WHEN the request reaches a protected route
- THEN the response returns HTTP 401 or 403 with a message indicating the user is inactive

### Requirement: Protected route enforcement

Routes secured by the auth middleware MUST return a standard DTO error for any unauthenticated or unauthorized request.

#### Scenario: Unauthenticated access to protected route

- GIVEN a protected route (e.g., `GET /api/v1/users`)
- WHEN a request without a valid token is made
- THEN the response returns HTTP 401 and uses the standard error envelope
