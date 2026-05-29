# Exploration: Auth Module → Vertical Slice Architecture

## Current State

The `internal/modules/auth/` module uses a **flat legacy structure** with 9 files:
- `handler.go` — single handler with 4 methods: `Login`, `Refresh`, `Logout`, `Me`
- `service.go` — single service struct with 4 methods, depends on `users.Repository` (cross-module)
- `repository.go` — `RefreshRepository` interface + GORM implementation for `refresh_tokens` table
- `model.go` — `RefreshToken` struct with GORM relation to `users.User` (cross-module)
- `dto.go` — request/response DTOs; `LoginResponse` and `MeResponse` embed `users.UserResponse` (cross-module)
- `routes.go` — `Register()` mounts 4 endpoints: `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me`
- `handler_test.go` — 6 test cases using fake service
- `service_test.go` — 6 test cases with fakes for `users.Repository`, password verifier, token issuer, refresh generator, refresh repository
- `repository_test.go` — 1 integration flow test using real SQLite + `users.NewRepository(db)` + `roles.Role` + `users.User`

### Cross-module dependencies (must be eliminated)

| File | Imports | What it uses |
|------|---------|-------------|
| `handler.go` | `modules/users` | `users.UserResponse` in `Me` handler |
| `service.go` | `modules/users` | `users.Repository` (injected), `users.User` model |
| `model.go` | `modules/users` | `users.User` as GORM preload relation on `RefreshToken` |
| `dto.go` | `modules/users` | `users.UserResponse` embedded in `LoginResponse` and `MeResponse` |
| `handler_test.go` | `modules/users` | `users.UserResponse` in test assertions |
| `service_test.go` | `modules/users`, `modules/roles`, `shared` | `users.User`, `users.Repository`, `roles.Role`, `shared.BaseModel` |
| `repository_test.go` | `modules/users`, `modules/roles`, `shared` | `users.User`, `users.NewRepository()`, `roles.Role`, `shared.BaseModel` |

The service also uses `gorm.ErrRecordNotFound` directly (platform concern, acceptable).

### Endpoints and their responsibilities

1. **POST /auth/login** — verify email+password, issue access+refresh token pair, return tokens + user DTO
2. **POST /auth/refresh** — validate refresh token hash, check not revoked/expired, rotate token pair
3. **POST /auth/logout** — validate and revoke refresh token
4. **GET /auth/me** — read user from auth context middleware, return sanitized user + role + permissions

## Affected Areas

- `internal/modules/auth/` — entire module restructured
- `internal/app/container.go` — wiring changes: auth container replaces flat handler/service construction
- `internal/app/bootstrap.go` or `server/router.go` — route registration call may change signature (from `auth.Register(v1, handler, ...)` to `auth.Register(v1, container, ...)`)
- `internal/middleware/` — `userLookup` adapter in `container.go` uses `users.Repository`; remains at app level (not auth module concern)

## Proposed Slice Mapping

| Endpoint | Slice Name | Rationale |
|----------|-----------|-----------|
| POST /auth/login | `authenticate_user` | Business intention: authenticate a user with credentials |
| POST /auth/refresh | `rotate_token` | Business intention: rotate an existing token pair |
| POST /auth/logout | `revoke_token` | Business intention: revoke a refresh token |
| GET /auth/me | `view_session` | Business intention: view the current authenticated session/user |

## Target Structure

```
internal/modules/auth/
  container.go              # NEW: composition root — wires slices, registers routes
  routes.go                 # UPDATED: Register() receives *Container, dispatches to slice handlers
  model.go                  # KEPT: transverse models (RefreshToken) or moved to core/
  core/
    model.go                # AuthUser (partial local model for users table), RefreshToken, RoleSummary
    dto.go                  # LoginRequest, RefreshRequest, TokenPairResponse, LoginResponse, MeResponse
    error.go                # Module-specific errors (if any beyond platform apperror)
  queries/
    find_user_by_email.go + _test.go          # SELECT * FROM users WHERE email = ?
    find_user_by_id_with_role.go + _test.go   # SELECT with preload role for refresh/rotate flows
  authenticate_user/
    handler.go + handler_test.go
    service.go + service_test.go
    repository.go + repository_test.go
  rotate_token/
    handler.go + handler_test.go
    service.go + service_test.go
    repository.go + repository_test.go
  revoke_token/
    handler.go + handler_test.go
    service.go + service_test.go
    repository.go + repository_test.go
  view_session/
    handler.go + handler_test.go
    service.go + service_test.go
    repository.go + repository_test.go  # may be minimal/no-op if only reads context
```

### Root-level retained files

- `container.go` — composition root: creates slice services/handlers, exposes handler references
- `routes.go` — `Register(v1, *Container, ...)` mounts slice handlers
- `model.go` — compatibility alias OR removed if all models move to `core/`
- `dto.go` — compatibility alias OR removed if all DTOs move to `core/`
- Test files at root level: only if testing transverse behavior (route registration, migration)

## Cross-Module Dependency Elimination Strategy

### 1. `users.Repository` → Local partial model + own repository

The auth module currently injects `users.Repository` (full interface with 10 methods) but only uses **one method**: `GetByEmail(email)`. For refresh/rotate flows, it also needs user data via `RefreshToken.User` preload.

**Strategy**: Define a local partial model in `auth/core/model.go`:

```go
type AuthUser struct {
    ID           uint   `gorm:"primaryKey"`
    PublicID     string `gorm:"column:public_id"`
    Email        string `gorm:"column:email"`
    PasswordHash string `gorm:"column:password_hash"`
    Status       string `gorm:"column:status"`       // or IsActive bool
    RoleName     string `gorm:"->;column:roles.name"` // joined via preload
    CompanyID    *uint  `gorm:"column:company_id"`
}

func (AuthUser) TableName() string { return "users" }
```

The `RefreshToken` model changes from:
```go
User users.User  // cross-module
```
to:
```go
User AuthUser  // local partial model
```

### 2. `users.UserResponse` → Local auth DTO

`LoginResponse` and `MeResponse` embed `users.UserResponse`. Replace with a local `AuthUserResponse` in `core/dto.go` containing only the fields auth returns:

```go
type AuthUserResponse struct {
    PublicID  string    `json:"public_id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    IsActive  bool      `json:"is_active"`
    RoleID    uint      `json:"role_id"`
    RoleName  string    `json:"role_name"`
    CompanyID *uint     `json:"company_id,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    CreatedBy *uint     `json:"created_by,omitempty"`
    UpdatedBy *uint     `json:"updated_by,omitempty"`
}
```

### 3. `roles.Role` in tests → Local minimal struct

Tests that create `roles.Role{}` should use a local minimal struct or inline the fields needed.

### 4. `shared.BaseModel` in tests → Inline or local alias

Tests using `shared.BaseModel{ID: 7, PublicID: "..."}` should inline the fields or define a local test helper.

## Reusable Queries for `queries/`

| Query File | Purpose | Used By |
|-----------|---------|---------|
| `find_user_by_email.go` | `SELECT * FROM users WHERE email = ?` | `authenticate_user` |
| `find_user_by_id_with_role.go` | `SELECT refresh_token JOIN users JOIN roles WHERE token_hash = ?` | `rotate_token`, `revoke_token` |

The `find_user_by_id_with_role` query is the complex one — it loads a refresh token with its associated user and role via GORM preloading. This is shared between `rotate_token` and `revoke_token` slices.

Each slice repository delegates to these queries. The repository_test.go for each slice validates the delegation wiring.

## Approaches

### Approach 1: Full vertical slice with `queries/` (Recommended)

Create 4 slices, extract shared queries into `queries/`, define local partial models in `core/`.

- **Pros**: Matches the established pattern (companies module), eliminates ALL cross-module imports, clean separation, each slice independently testable
- **Cons**: More files to create (~24 new files), requires careful test migration
- **Effort**: Medium-High

### Approach 2: Vertical slice without `queries/`

Each slice has its own repository with duplicated query logic. No `queries/` folder.

- **Pros**: Simpler initial structure, fewer files
- **Cons**: Duplicated query code between `rotate_token` and `revoke_token`, violates the "reusable queries in queries/" rule from _context.md
- **Effort**: Medium

### Approach 3: Hybrid — keep some files at root during transition

Keep `service.go` and `repository.go` at root during migration, only split handlers into slices first.

- **Pros**: Smaller initial PR, easier review
- **Cons**: Incomplete vertical slice, violates the target architecture, creates technical debt
- **Effort**: Low (but creates more work later)

## Recommendation

**Approach 1** — full vertical slice with `queries/`. The companies module already proves this pattern works. The auth module has exactly 4 endpoints mapping cleanly to 4 slices. The `queries/` folder is warranted because `rotate_token` and `revoke_token` share the same refresh-token-with-user lookup.

## Risks

1. **Test migration complexity**: `repository_test.go` currently runs a full login→refresh→logout flow using real `users.Repository` and `roles.Role`. This test must be split or rewritten to use local partial models. The SQLite AutoMigrate call currently includes `roles.Role{}` and `users.User{}` — these must be replaced.
2. **`Me` handler has no service/repository**: Currently `Me` reads directly from `authctx` middleware context with no service or repository call. The `view_session` slice will have a handler + handler_test but its service and repository may be minimal or pass-through. This is acceptable — the slice still encapsulates the use case.
3. **App container wiring change**: `internal/app/container.go` line 64 (`auth.NewService(...)`) and line 65 (`auth.NewHandler(...)`) must change to `auth.NewContainer(db)` with the new interface. Route registration on line 92 also changes signature.
4. **`userLookup` in app container**: The `userLookup` struct (line 107-136) still uses `users.Repository` for the auth middleware. This is at the **app level**, not the auth module level, so it stays as-is. The auth module itself becomes self-contained.
5. **Line budget**: This refactor will likely exceed the 400-line review budget. Chained PRs are recommended.

## Migration Plan (Phased)

### Phase 1: Foundation
- Create `core/model.go`, `core/dto.go`, `core/error.go` with local partial models and DTOs
- Create `queries/` with shared query functions and tests
- Create `container.go` (composition root skeleton)

### Phase 2: `authenticate_user` slice
- Create handler, service, repository, and tests
- Wire into container
- Verify `POST /auth/login` behavior unchanged

### Phase 3: `rotate_token` and `revoke_token` slices
- Create both slices (share `queries/` logic)
- Wire into container
- Verify `POST /auth/refresh` and `POST /auth/logout` behavior unchanged

### Phase 4: `view_session` slice
- Create handler (reads from authctx), minimal service/repository
- Wire into container
- Verify `GET /auth/me` behavior unchanged

### Phase 5: Cleanup
- Update `routes.go` to use container + slice handlers
- Update `internal/app/container.go` wiring
- Remove old flat files (`handler.go`, `service.go`, `repository.go`, `model.go`, `dto.go`) or convert to compatibility aliases
- Run full test suite

## Existing Tests

### Current auth test files
| File | Lines | Test Count | Description |
|------|-------|-----------|-------------|
| `handler_test.go` | 167 | 6 cases | Login success, validation error, unauthorized, refresh, logout, me |
| `service_test.go` | 223 | 6 cases | Login with active user, login failures (missing/wrong/inactive), refresh rotation, refresh revoked, logout |
| `repository_test.go` | 111 | 1 flow | Full login→refresh→logout integration flow with SQLite |

### Test commands
- **Auth module**: `go test ./internal/modules/auth/...` — currently passes (cached)
- **Full suite**: `go test ./...` — expected to pass after migration with zero behavior change

### Test migration notes
- Handler tests: mostly portable — use fake service interfaces that will map 1:1 to slice service interfaces
- Service tests: fakes for `users.Repository` must be replaced with fakes for local query interfaces
- Repository integration test: must replace `users.NewRepository(db)` and `roles.Role{}` with local partial model AutoMigrate

## Open Questions

None blocking. The scope is well-defined: 4 endpoints → 4 slices, eliminate cross-module imports to `users` and `roles`, preserve behavior.
