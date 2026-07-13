# Delta for Auth

## ADDED Requirements

### Requirement: Slice-aligned auth layout

The auth module MUST keep the four use cases under `internal/modules/auth/slices/`: `authenticate_user`, `rotate_token`, `revoke_token`, and `view_session`. The module root MUST remain wiring-only, and shared module errors MUST live in `internal/modules/auth/core/`.
(Previously: auth used a flatter layout with slice/package drift.)

#### Scenario: Auth tree uses slices

- GIVEN the auth module is inspected after the change
- WHEN its package layout is listed
- THEN the four use cases are under `slices/`
- AND the module root contains wiring only

#### Scenario: Layout does not expand scope

- GIVEN the auth module is migrated
- WHEN other modules are inspected
- THEN no other module is moved to `slices/`
- AND their existing layouts remain unchanged

### Requirement: Universal auth repository persistence boundary

Every auth repository interface method MUST retain an idiomatic `error` return and MUST NOT import or expose `platform/apperror`. Every non-nil persistence error returned by any auth repository MUST be a module-owned `*apperror.AppError` returned through `error`; no raw GORM, SQL, or driver error MAY cross the repository boundary. Entity-specific unary mappers in `internal/modules/auth/queries/map_errors.go` MUST own all translation. Known persistence outcomes MUST map to their specific auth domain AppErrors. Unknown failures MUST map to module-owned internal AppErrors that preserve the original cause through `Internal`/`Unwrap()`. Repositories MUST pass every GORM `.Error` through the correct mapper and MUST explicitly map zero `RowsAffected` when it means the target did not exist.
(Previously: only selected reads were mapped, unknown errors remained unchanged, writes returned raw `.Error`, and updates ignored zero affected rows.)

#### Scenario: Authentication lookup maps to invalid credentials

- GIVEN `authenticate_user` receives a GORM not-found error while loading a user or role
- WHEN the repository translates the persistence error
- THEN the result is `core.ErrInvalidCredentials`
- AND the repository does not pass a mapped sentinel into the mapper

#### Scenario: Refresh-token lookup maps to invalid refresh token

- GIVEN `rotate_token` or `revoke_token` receives a GORM not-found error while loading a refresh token
- WHEN the repository translates the persistence error
- THEN the result is `core.ErrInvalidRefreshToken`
- AND the repository does not pass a mapped sentinel into the mapper

#### Scenario: Unknown persistence failure is internal and preserves cause

- GIVEN any auth read, create, update, revoke, or other persistence operation receives an unknown persistence error
- WHEN the repository translates the error through the entity-specific mapper
- THEN `errors.As` finds a module-owned `*apperror.AppError`
- AND it has the module's internal persistence code and HTTP 500 status
- AND `errors.Is` reaches the original persistence cause
- AND the repository does not return the raw cause as its boundary value

#### Scenario: Every auth GORM error path is mapped

- GIVEN any auth repository executes a GORM read, create, update, revoke, delete, or transaction operation
- WHEN the operation exposes a non-nil `.Error`
- THEN the repository passes that error to the mapper for the affected entity
- AND no repository method returns `.Error` directly

#### Scenario: Zero-row update has domain meaning

- GIVEN a refresh-token revoke operation completes without a GORM error but affects zero rows
- WHEN the repository evaluates the result
- THEN it maps the missing token to `core.ErrInvalidRefreshToken`
- AND a successful revoke requires at least one affected row

#### Scenario: Repository interfaces remain idiomatic

- GIVEN every `Repository` interface under `internal/modules/auth/slices`
- WHEN its imports and method signatures are inspected
- THEN persistence failures are returned as `error`
- AND the interface neither imports nor exposes `platform/apperror`

#### Scenario: Tests enforce the universal boundary

- GIVEN mapper and repository error tests plus structural guards
- WHEN all auth repositories are discovered recursively
- THEN tests assert `errors.As` to `*apperror.AppError`, status/code, cause preservation, and no raw persistence leak
- AND structural guards scan every auth repository file and method without a fixed list of only selected call sites

### Requirement: Canonical auth error filename

The auth module MUST rename `internal/modules/auth/core/error.go` to `internal/modules/auth/core/errors.go`. No compatibility alias, re-export, or wrapper file MAY remain.
(Previously: the canonical shared error file name was singular.)

#### Scenario: Canonical file exists

- GIVEN the auth module root is inspected
- WHEN the core directory is listed
- THEN `core/errors.go` exists
- AND `core/error.go` does not exist

### Requirement: Auth surface remains exact

The system MUST preserve the existing HTTP methods, routes, response envelope, payload shape, and status codes for `POST /api/v1/auth/login`, `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, and `GET /api/v1/auth/me`. Successful responses MUST remain unchanged, invalid credentials MUST still return 401, and validation failures MUST still return 422.
(Previously: route and status preservation was an intent, not a tested contract.)

#### Scenario: Login success remains identical

- GIVEN an active user with valid credentials
- WHEN `POST /api/v1/auth/login` is called
- THEN the response remains HTTP 200
- AND the payload still includes the same token fields and user shape

#### Scenario: Validation and auth failures keep status codes

- GIVEN an invalid login body or invalid credentials
- WHEN `POST /api/v1/auth/login` is called
- THEN invalid input returns HTTP 422
- AND bad credentials return HTTP 401 with the same generic auth failure behavior

#### Scenario: Other auth endpoints stay stable

- GIVEN valid auth tokens or sessions
- WHEN `POST /api/v1/auth/refresh`, `POST /api/v1/auth/logout`, or `GET /api/v1/auth/me` is called
- THEN each endpoint preserves its existing path, method, payload shape, and status codes
- AND no shim endpoint is introduced
