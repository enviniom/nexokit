# Auth Specification

## Purpose

Token lifecycle and authentication endpoints for the NexoKit API.

## Requirements

### Requirement: Login endpoint

The system MUST expose `POST /api/v1/auth/login`. It MUST accept `email` and `password`, validate the user is active, and return a standard DTO response containing a PASETO access token and an opaque refresh token. It MUST NOT reveal whether the email or password was incorrect.

#### Scenario: Successful login

- GIVEN an active user with email `u@example.com` and password `Secret1!`
- WHEN `POST /api/v1/auth/login` is called with matching credentials
- THEN the response returns HTTP 200, `success: true`, and `data` contains `access_token`, `refresh_token`, and a user object without `password` or `password_hash`

#### Scenario: Inactive user denied

- GIVEN an inactive user with email `u@example.com`
- WHEN `POST /api/v1/auth/login` is called with correct credentials
- THEN the response returns HTTP 401 with a generic authentication failure message

#### Scenario: Invalid credentials

- GIVEN no user matches the supplied email or password
- WHEN `POST /api/v1/auth/login` is called
- THEN the response returns HTTP 401 with a generic authentication failure message

### Requirement: Refresh endpoint

The system MUST expose `POST /api/v1/auth/refresh`. It MUST accept a valid refresh token, rotate the token pair (issue new access and refresh tokens), and revoke the old refresh token.

#### Scenario: Successful refresh

- GIVEN a valid refresh token
- WHEN `POST /api/v1/auth/refresh` is called with that token
- THEN the response returns HTTP 200 with a new `access_token` and `refresh_token`
- AND the previous refresh token is revoked

#### Scenario: Revoked refresh token

- GIVEN a refresh token that has been revoked
- WHEN `POST /api/v1/auth/refresh` is called with that token
- THEN the response returns HTTP 401 with an invalid token message

### Requirement: Logout endpoint

The system MUST expose `POST /api/v1/auth/logout`. It MUST revoke the provided refresh token.

#### Scenario: Successful logout

- GIVEN a valid refresh token
- WHEN `POST /api/v1/auth/logout` is called with that token
- THEN the response returns HTTP 200 and the refresh token is revoked

### Requirement: Me endpoint

The system MUST expose `GET /api/v1/auth/me`. It MUST return the authenticated user without `password` or `password_hash`. The response `data` object MUST include a `permissions` field containing an array of permission slug strings associated with the user's role.

#### Scenario: Get current user

- GIVEN an authenticated request with a valid access token
- WHEN `GET /api/v1/auth/me` is called
- THEN the response returns HTTP 200 and `data` contains the user object without password fields
- AND `data.permissions` contains the permission slugs assigned to the user's role

#### Scenario: Root user permissions

- GIVEN an authenticated root user
- WHEN `GET /api/v1/auth/me` is called
- THEN `data.permissions` contains all system permission slugs

#### Scenario: User with no permissions assigned

- GIVEN an authenticated user whose role has no permissions assigned
- WHEN `GET /api/v1/auth/me` is called
- THEN `data.permissions` contains an empty array

### Requirement: Token security

The system MUST issue PASETO v4.local access tokens with claims `sub`, `role`, `company_id`, `token_type`, `issued_at`, and `expires_at`. Refresh tokens MUST be opaque random strings stored only as a hash. Passwords MUST be hashed with argon2id.

#### Scenario: Access token claims

- GIVEN a successful login
- WHEN the access token is inspected
- THEN it contains the required claims and no sensitive data

#### Scenario: No password leakage

- GIVEN any auth endpoint response
- WHEN the response body is inspected
- THEN it does not contain `password`, `password_hash`, or `refresh_token` plaintext
