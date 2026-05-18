# Exploration: change-03-auth — Authentication, Users, Roles & Root

## Current State

The NexoKit foundation (change-01) and CLI tooling (change-02) are in place. Auth-related code exists only as empty stubs:

- `internal/middleware/auth.go` — 3 lines, TODO for change-03
- `internal/modules/auth/module.go` — stub `Register` function
- `internal/modules/users/module.go` — stub `Register` function
- `internal/modules/roles/module.go` — stub `Register` function
- `internal/platform/password/password.go` — stub (TODO: change-02, still empty)
- `internal/platform/token/token.go` — stub (TODO: change-02, still empty)
- `internal/platform/identity/identity.go` — stub (TODO: change-02, still empty)

The CLI root creator (`internal/cli/root/root.go`) defines clean boundary interfaces (`RootStorage`, `PasswordHasher`) and validation logic, but `commands/createroot.go` passes `nil, nil`, so `create-root` always returns `ErrStorageNotWired`.

Database has zero tables. The only migration (`20260101000000_init.sql`) is empty. No `seeds/` directory exists. Config lacks auth/PASETO env vars. `go.mod` lacks a PASETO library (`golang.org/x/crypto` is already present indirectly for argon2).

---

## Affected Areas

| Path | Impact |
|------|--------|
| `internal/config/config.go` | Add `AuthConfig`, `RootConfig`, PASETO/TTL fields |
| `.env.example` | Document new env vars |
| `go.mod` / `go.sum` | Add PASETO library (`github.com/o1egl/paseto`) |
| `migrations/` | Add Goose migrations for `roles`, `users`, `refresh_tokens` |
| `seeds/` | New package with role seeds + root user seed |
| `internal/platform/identity/` | Implement PublicID generation (ULID) |
| `internal/platform/password/` | Implement argon2id hasher + verifier |
| `internal/platform/token/` | Implement PASETO v4.local builder + parser |
| `internal/modules/roles/` | `model.go`, `repository.go`, `service.go`, `handler.go`, `dto.go`, `routes.go`, `validation.go` |
| `internal/modules/users/` | `model.go`, `repository.go`, `service.go`, `handler.go`, `dto.go`, `routes.go`, `validation.go` |
| `internal/modules/auth/` | `model.go`, `repository.go`, `service.go`, `handler.go`, `dto.go`, `routes.go`, `validation.go` |
| `internal/middleware/auth.go` | `AuthMiddleware`, `RequirePermission`, `RequireRole`, context user injection |
| `internal/app/container.go` | Wire auth, users, roles repos/services/handlers |
| `internal/app/bootstrap.go` | Wire new container fields if needed |
| `internal/cli/root/root.go` | No structural change; boundaries finally satisfied by real implementations |
| `internal/cli/commands/createroot.go` | Wire real storage + hasher (needs DB bootstrap) |
| `tests/` | Add module tests, middleware tests, integration tests |

---

## Approaches

### 1. PASETO Library

| Library | Pros | Cons | Effort |
|---------|------|------|--------|
| `github.com/o1egl/paseto` | Most popular Go impl, v4.local + v4.public, well documented | Slightly larger API surface | Low |
| `github.com/aidantwoods/go-paseto` | Clean API, maintained | Less community usage | Low |

**Recommendation**: `o1egl/paseto`. It is the de-facto standard in Go, has v4.local support, and ample examples.

### 2. Root Creation Strategy

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| A. CLI only (interactive) | No env var leakage | Manual step on every deploy | Low |
| B. Env vars only | Good for IaC | Password in env (acceptable if injected by secrets manager) | Low |
| C. CLI + env vars (recommended by prompt) | Flexible: env for automation, interactive fallback | Slightly more code to handle both | Low |

**Recommendation**: C — matches the prompt exactly. Read `ROOT_USER_*` from env first; if missing, prompt interactively.

### 3. Refresh Token Storage

| Approach | Pros | Cons | Effort |
|----------|------|------|--------|
| A. PostgreSQL (proposed) | Single source of truth, revocable, transactional with user ops | Slightly more DB load | Low |
| B. Redis | Faster revocation lookup | Another dependency, complexity | Medium |

**Recommendation**: A — the prompt explicitly requires `refresh_tokens` table in PostgreSQL. No need for Redis now.

### 4. Dependency & Implementation Order

To keep the build green at every step:

```
Phase 1 — Platform leaves (no DB needed)
  ├── platform/identity (ULID generation)
  ├── platform/password (argon2id wrapper)
  └── platform/token (PASETO wrapper)

Phase 2 — Config + Migrations
  ├── config/config.go add AuthConfig, RootConfig
  ├── .env.example updates
  └── migrations for roles, users, refresh_tokens

Phase 3 — Seeds
  └── seeds/roles_seed.go, seeds/root_seed.go

Phase 4 — Modules (each: model → repo → service → handler → dto → routes → validation)
  ├── roles (read-only API, no mutation endpoints)
  ├── users (CRUD + password change)
  └── auth (login, refresh, logout, me)

Phase 5 — Middleware & Wiring
  ├── middleware/auth.go
  ├── app/container.go (full DI graph)
  └── server/router.go (register modules)

Phase 6 — CLI
  └── commands/createroot.go (bootstrap DB, wire real deps)
```

---

## Critical Design Decisions

### 1. One Role Per User
The prompt mandates `role_id` on `users` table (not a join table). This simplifies queries and validation. Enforce at DB level with `NOT NULL` and a foreign key.

### 2. Root User Is a Regular User
Root is simply a user with `role_id` pointing to the `root` role. `RootStorage` boundary in `cli/root/root.go` should be implemented by the `users` module repository, not a separate table.

### 3. Refresh Token Hashing
Store only `SHA-256` hash of the opaque refresh token. The raw token is returned once at creation/refresh and never stored plaintext. Using argon2id for refresh token hashing is overkill; a fast cryptographic hash (SHA-256) is standard practice for opaque token storage.

### 4. Auth Middleware Design
Use PASETO v4.local decryption in middleware. Extract `sub` (user public_id) and `role` (role slug). Look up the user from DB to ensure they still exist and are active. Inject a lightweight `AuthUser` struct into `gin.Context` for downstream handlers.

### 5. CLI `create-root` Needs DB
`commands/createroot.go` currently has no DB access. It must bootstrap at least `config.Load()` + `db.Connect()` to get a real `*gorm.DB`. Reuse the same bootstrap pattern as the API but lighter.

### 6. Module Contracts
The `users` module must expose a contract for the `auth` module (e.g. `UserReader` interface) so auth can look up users by email without importing the users repository directly. Similarly, `roles` must expose a contract for `roles` lookup.

---

## Risks

| # | Risk | Mitigation |
|---|------|------------|
| 1 | **New dependency** (`paseto`) may have breaking API changes or compatibility issues with Go 1.26 | Pin to a stable release; review release notes before `go get` |
| 2 | **Container explosion** — `container.go` will grow from 0 to ~6+ fields | Group related deps in small structs (e.g. `AuthDeps`, `UserDeps`) or keep flat but well-commented |
| 3 | **Circular imports** if middleware imports module models directly | Middleware must depend only on small interfaces (e.g. `UserReader`, `TokenValidator`) injected from container |
| 4 | **Seed command discovery** expects package `seeds` and functions ending in `Seed` | Ensure seed files use `package seeds` and `func RolesSeed() error`, `func RootSeed() error` |
| 5 | **Password validation divergence** — `cli/root/root.go` has its own validation rules (`hasMixedCaseAndDigit`) while the prompt suggests `platform/validator` rules | Reuse `platform/validator` rules in root validation, or keep CLI validation stricter. Align with prompt requirements (min 8, uppercase, lowercase, digit, special char) |
| 6 | **Migration ordering** — `users` depends on `roles`; `refresh_tokens` depends on `users` | Create migration files with correct timestamp order: roles → users → refresh_tokens |
| 7 | **Test data isolation** — auth tests need real DB users and tokens | Use transactions for DB isolation; Truncate tables in test teardown |

---

## Ready for Proposal

**Yes.**

All ambiguities are resolvable:
- PASETO library: `o1egl/paseto`
- Root creation: CLI + env vars
- Role-user relationship: single `role_id` on users
- Refresh tokens: PostgreSQL table with hashed tokens

No blockers. The change is well-scoped by the prompt. Proceed to **propose**.
