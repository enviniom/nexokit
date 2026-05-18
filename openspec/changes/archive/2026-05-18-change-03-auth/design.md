# Design: Authentication, Users, Roles & Root

## Technical Approach

Implement the auth stack bottom-up: platform primitives (`identity`, `password`, `token`), schema/seeds, `roles` and `users`, then `auth` endpoints and middleware. Keep the current flat DI style: `Container` owns repositories/services/handlers and registers module routes under `/api/v1`.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| PASETO library | Use a v4.local-capable package behind `internal/platform/token`; do not expose package types. | `github.com/o1egl/paseto` from proposal. | `o1egl/paseto` documents v1/v2 only, so it cannot satisfy v4.local. Wrapper keeps replacement cheap if dependency approval changes. |
| Refresh tokens | Generate 32+ random bytes, return opaque token once, store `sha256(token)` with `revoked_at`, `expires_at`, `replaced_by_hash`. | Store raw token; use JWT-like refresh. | Spec requires opaque hashed refresh and revocation/rotation. Random high entropy makes SHA-256 lookup safe and fast. |
| User roles | `users.role_id -> roles.id`, preload role for DTOs/claims. | Join table. | Spec requires single-role users; FK is simpler and enforces it. |
| Middleware boundary | Middleware depends on token parser + small user lookup interface, not modules. | Import `users` service from middleware. | Preserves no cross-module imports and avoids cycles. |
| CLI root | `create-root` loads config/db, constructs GORM root storage + argon2id hasher. | Bootstrap full app; keep nil storage. | Meets real-storage requirement without starting HTTP/cache/router. |

## Data Flow

Login/refresh:

    Handler -> AuthService -> UserRepository -> PasswordVerifier
       |             |-> TokenManager issues PASETO access
       |             `-> RefreshTokenRepository stores hash
       `-> response.Success/Created DTO

Protected request:

    Bearer token -> middleware.Auth -> TokenManager.Parse
       -> AuthUserLookup by sub -> reject inactive -> context user -> handler

## File Changes

| File | Action | Description |
|---|---|---|
| `go.mod`/`go.sum` | Modify | Add PASETO v4.local dependency and use existing `x/crypto` for argon2id. |
| `internal/config/config.go`, `.env.example` | Modify | Add `AuthConfig`: PASETO key, access TTL, refresh TTL. |
| `internal/platform/identity/identity.go` | Modify | ULID/public ID generator for models. |
| `internal/platform/password/password.go` | Modify | Argon2id hash/verify with encoded params. |
| `internal/platform/token/token.go` | Modify | PASETO v4.local access token issue/parse and opaque refresh generation/hash. |
| `migrations/20260516000000_auth.sql` | Create | Create `roles`, `users`, `refresh_tokens` with indexes/constraints. |
| `seeds/roles.go` | Create | Idempotently seed `root`, `admin`, `user` as system roles. |
| `internal/modules/roles/*` | Create/Modify | Model, repository, service, handler, DTO, routes for list/get. |
| `internal/modules/users/*` | Create/Modify | Model, repository, service, handler, DTO, routes for CRUD/password/status. |
| `internal/modules/auth/*` | Create/Modify | Login, refresh, logout, me service/handler/routes. |
| `internal/middleware/auth.go` | Modify | Bearer validation, active user lookup, safe context injection. |
| `internal/app/container.go` | Modify | Wire repositories/services/handlers and protected groups. |
| `internal/cli/commands/createroot.go`, `internal/cli/root/*` | Modify/Create | Read root env vars, wire DB storage + argon2id hasher, keep interactive fallback. |

## Interfaces / Contracts

```go
type AccessClaims struct { Sub, Role, TokenType string; CompanyID *uint; IssuedAt, ExpiresAt time.Time }
type AuthContextUser struct { ID uint; PublicID, Email, Name, Role string; CompanyID *uint; IsActive bool }
```

Routes: `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me`; protected `GET /roles`, `GET /roles/:id`; protected users CRUD plus `PATCH /users/:id/password` and `PATCH /users/:id/status`.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | argon2id verify, token claims/expiry, refresh hashing/rotation, root validation | Table-driven tests with deterministic clocks/random readers. |
| Handler/middleware | DTO envelopes, 401/403/404 paths, no password leakage | `httptest` + small fake services/lookups. |
| Repository/CLI | migrations, role seed idempotency, root create/re-run | Skippable integration tests; use test DB config and `testing.Short()`. |

## Migration / Rollout

Apply Goose migration, run role seed, then run `nexokit create-root` with `ROOT_USER_NAME`, `ROOT_USER_EMAIL`, `ROOT_USER_PASSWORD`. Rollback drops `refresh_tokens`, `users`, `roles` in reverse order.

## Open Questions

- [ ] Confirm dependency/license approval for the selected v4.local PASETO package before apply.
