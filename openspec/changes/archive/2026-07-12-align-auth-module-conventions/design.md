# Design: Align Auth Module Conventions

## Technical Approach

Complete the behavior-preserving auth rehome with a universal repository persistence boundary. Keep repository signatures idiomatic `error`, translate every GORM failure through entity-specific unary mappers, map known outcomes to auth domain AppErrors, wrap unknown failures in module-owned internal AppErrors with their causes preserved, and define zero-row semantics for revocation. Routes, DTOs, and public APIs do not change.

## Architecture Decisions

| Decision | Alternatives considered | Rationale |
|---|---|---|
| Move each package intact to `auth/slices/{authenticate_user,rotate_token,revoke_token,view_session}` | Wrappers, aliases, partial moves | Real moves match module conventions and avoid prohibited shims; package names and exported APIs remain unchanged. |
| Add entity-specific `queries.MapUserError(err error) error` and `queries.MapRefreshTokenError(err error) error` | Generic `MapNotFound(err, mapped)`; repository-selected sentinels; repository-local checks | The documented module convention requires entity-specific helpers that accept only the persistence error. Each mapper owns the GORM-to-auth decision, so repositories cannot inject domain policy. |
| Keep repository interfaces typed as `error`; require module-owned `*apperror.AppError` values at runtime | Concrete `*apperror.AppError` signatures; raw persistence errors | Idiomatic interfaces stay decoupled from platform implementation while the runtime boundary remains strict and testable with `errors.As`. |
| Map unknown persistence failures to entity-specific internal auth AppErrors and preserve the cause | Return unknown errors unchanged; use a platform-global internal sentinel | Raw errors cannot escape, and module-owned codes identify the failed persistence entity while `Internal`/`Unwrap()` retains logging evidence. |
| Treat zero-row token revocation as invalid refresh token | Ignore `RowsAffected`; report success | Revoking a missing hash has domain meaning and must not silently succeed unless the use case explicitly defines idempotency. |
| Keep route/container fields and constructors stable | Rename exported composition API | Only import paths require change; stable wiring minimizes semantic and review risk. |
| Apply as one ordered refactor in one PR | Transitional compatibility layer | The repository builds atomically, and shims are explicitly out of scope. |

## Data Flow

```text
HTTP route -> unchanged Container field -> moved handler/service/repository
                                              |
                                              v
                                      auth/queries query
                                              |
                GORM result -> entity mapper -> auth core AppError through error
                     |                              |
                     +-- RowsAffected semantics     +-- unknown cause retained
```

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/auth/{authenticate_user,rotate_token,revoke_token,view_session}/**` | Move | Move all production and test files to matching `internal/modules/auth/slices/*` directories; retain package declarations. |
| `internal/modules/auth/container.go` | Modify | Import moved packages from `auth/slices/*`; preserve constructor order, dependencies, fields, and handlers. |
| `internal/modules/auth/container_test.go` | Modify | Continue proving all four handlers are wired. |
| `internal/modules/auth/queries/map_errors.go` | Modify | Make user and refresh-token mappers total: nil stays nil, known outcomes map specifically, and unknown failures become module-owned internal AppErrors with preserved causes. |
| `internal/modules/auth/queries/map_errors_test.go` | Modify | Assert `errors.As`, status/code, known mappings, cause preservation, and no raw persistence leak. |
| `internal/modules/auth/slices/authenticate_user/repository.go` | Modify | Pass user lookup and role preload errors only to `queries.MapUserError`; do not select a sentinel in the repository. |
| `internal/modules/auth/slices/{authenticate_user,rotate_token,revoke_token}/repository.go` | Modify | Pass every read/write/update `.Error` through the entity mapper and map domain-significant zero-row updates. |
| `internal/modules/auth/core/errors.go` | Modify | Add module-owned internal user and refresh-token persistence codes/constructors that retain the original cause. |
| `internal/modules/auth/queries/map_errors_structure_test.go` | Modify | Discover all auth repositories and guard interfaces, direct `.Error` returns, and mapper coverage without a fixed method list. |
| `internal/modules/auth/core/error.go` -> `errors.go` | Rename | Filename-only normalization; sentinel values/codes/messages remain byte-for-byte unchanged. |
| `internal/modules/auth/routes.go`, `routes_test.go` | Verify | No import or behavior change expected; retain route/middleware contract coverage. |

## Interfaces / Contracts

```go
// MapUserError translates user-entity persistence failures into auth errors.
func MapUserError(err error) error

// MapRefreshTokenError translates refresh-token persistence failures into auth errors.
func MapRefreshTokenError(err error) error
```

Both helpers return `nil` for `nil` and use `errors.Is` for known outcomes. `MapUserError` maps not-found to `core.ErrInvalidCredentials`; `MapRefreshTokenError` maps not-found to `core.ErrInvalidRefreshToken`. Any unknown error becomes the corresponding module-owned internal persistence AppError with HTTP 500 and the original error in `Internal`. Repository interfaces still return `error` and do not import `apperror`; tests inspect concrete values with `errors.As`. Validation-driven 422 responses are unaffected.

## Exhaustive Auth Persistence Inventory

| Repository / method | Current operation | Intended mapper | `RowsAffected` behavior |
|---|---|---|---|
| `authenticate_user.GetByEmail` | `FindUserByEmail`: `Where(email).First(AuthUser)` | `MapUserError` for query `.Error` | `First` represents missing data as `ErrRecordNotFound`; no separate check. |
| `authenticate_user.GetByEmail` | `First(AuthRole, roleID)` | `MapUserError` for role `.Error` | `First` not-found maps to invalid credentials; no separate check. |
| `authenticate_user.CreateRefreshToken` | `Create(RefreshToken)` | `MapRefreshTokenError` for create `.Error` | Not applicable; successful create does not require an affected-row domain branch. |
| `rotate_token.GetByHash` | `FindRefreshTokenByHashWithUser`: `Preload(User.Role).Where(hash).First(RefreshToken)` | `MapRefreshTokenError` for query `.Error` | `First` not-found maps to invalid refresh token; no separate check. |
| `rotate_token.CreateRefreshToken` | `Create(RefreshToken)` | `MapRefreshTokenError` for create `.Error` | Not applicable; successful create does not require an affected-row domain branch. |
| `rotate_token.Revoke` | `Model(RefreshToken).Where(hash).Updates(revoked_at,replaced_by_hash)` | `MapRefreshTokenError` for update `.Error` | Zero rows maps to invalid refresh token. |
| `revoke_token.GetByHash` | `FindRefreshTokenByHashWithUser`: `Preload(User.Role).Where(hash).First(RefreshToken)` | `MapRefreshTokenError` for query `.Error` | `First` not-found maps to invalid refresh token; no separate check. |
| `revoke_token.Revoke` | `Model(RefreshToken).Where(hash).Updates(revoked_at)` | `MapRefreshTokenError` for update `.Error` | Zero rows maps to invalid refresh token. |
| `view_session.BuildSession` | No GORM operation; maps trusted `authctx.User` in memory | None | Not applicable. |

`queries.FindUserByIDWithRole` also executes `Preload(Role).First(AuthUser)` and currently has no production caller. It remains a raw reusable query by layer design: any repository that adopts it MUST pass its returned error through `MapUserError` before crossing the repository boundary.

## Testing Strategy

| Layer | What to test | Approach |
|---|---|---|
| Unit | Entity-specific mapper contracts and module-owned internal errors | Add RED table tests for nil, known direct/wrapped outcomes, unknown causes, `errors.As`, status/code, `errors.Is`, and no raw leak. |
| Repository | Every inventory row with persistence behavior | Cover reads, both creates, both revokes, and zero-row semantics; assert every non-nil persistence result is a module-owned AppError. |
| Structural | Universal repository boundary | Recursively discover all `auth/slices/**/repository.go` files; reject `apperror` in interfaces, direct raw `.Error` returns, and unmapped GORM result errors without hard-coded method names. |
| Integration | Composition and route contract | Run `TestNewContainer_WiresAllSlices` and route tests; confirm login/refresh/logout/me paths and middleware placement. |
| Verification | Import completeness and semantics | Run `go test ./internal/modules/auth/...`, `go test ./...`, then `go build ./...`; ensure no old `internal/modules/auth/{slice}` imports or directories remain. |

## Threat Matrix

N/A — no routing behavior, shell, subprocess, VCS/PR automation, executable-file classification, or process-integration boundary changes.

## Migration / Rollout

No data migration or feature flag is required. The corrective sequence is: inventory all repositories; add RED mapper/repository/structural tests; add module-owned internal persistence errors; make both mappers total; migrate every inventory call site; enforce zero-row revoke semantics; run focused and full tests/build; reconcile all artifacts. Routes and other modules remain untouched. Deliver as one maintainer-approved `size:exception` PR within the authorized 1,200-line budget and commit by coherent work unit.

Rollback the single PR: restore original directories/imports, restore `core/error.go`, remove mapper/tests, and re-inline the three repositories' previous GORM checks. No persisted data or external consumer rollback is needed.

## Open Questions

None. The zero-row review concludes that refresh-token revocation targets a required token and therefore maps zero affected rows to `core.ErrInvalidRefreshToken`; changing revocation to idempotent behavior would require a separate spec decision.
