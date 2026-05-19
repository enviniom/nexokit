# Delta for Auth

## MODIFIED Requirements

### Requirement: Me endpoint

The system MUST expose `GET /api/v1/auth/me`. It MUST return the authenticated user without `password` or `password_hash`. The response `data` object MUST include a `permissions` field containing an array of permission slug strings associated with the user's role.
(Previously: Me endpoint returned user object without password fields but did not include permissions)

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