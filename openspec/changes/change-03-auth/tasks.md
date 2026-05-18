# Tasks: Authentication, Users, Roles & Root

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1,200–1,500 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Platform + schema + seeds + roles module | PR 1 | Base branch main; includes migrations and config |
| 2 | Users module + CLI root wiring | PR 2 | Targets main; depends on PR 1 schema |
| 3 | Auth module + middleware + container wiring | PR 3 | Targets main; depends on PR 2 users |

## Phase 1: Foundation

- [x] 1.1 Add `AuthConfig` to `internal/config/config.go` and `.env.example` (PASETO key, access TTL, refresh TTL)
- [x] 1.2 RED: Write failing tests for `password.Hash` and `password.Verify`
- [x] 1.3 GREEN: Implement `internal/platform/password/password.go` with argon2id
- [x] 1.4 RED: Write failing tests for `token.IssueAccess` and `token.ParseAccess`
- [x] 1.5 GREEN: Implement `internal/platform/token/token.go` with PASETO v4.local and claims
- [x] 1.6 Implement `token.GenerateRefresh` and `token.HashRefresh` (opaque SHA-256)
- [x] 1.7 Implement `internal/platform/identity/identity.go` ULID/public ID generator
- [x] 1.8 Create `migrations/20260516000000_auth.sql` for `roles`, `users`, `refresh_tokens`
- [x] 1.9 Create `seeds/roles.go` to idempotently seed `root`, `admin`, `user` as system roles
- [x] 1.10 Create `internal/modules/roles/{model,dto,repository,service,handler,routes}.go`
- [x] 1.11 PR1 Correction: roles CRUD (create, update, delete), paginated list, unique name, system-role protection, BaseModel inheritance, slug/description fields, migration and seed alignment, CLI template pagination update

## Phase 2: Users & CLI Root

- [x] 2.1 Create `internal/modules/users/{model,dto,repository,service,handler,routes}.go` with CRUD
- [x] 2.2 Implement `PATCH /users/:id/password` requiring current password and argon2id rehash
- [x] 2.3 Implement `PATCH /users/:id/status` active/inactive toggle
- [x] 2.4 Wire `internal/cli/commands/createroot.go` to real `RootStorage` and `PasswordHasher`
- [x] 2.5 Update `create-root` to read `ROOT_USER_NAME`, `ROOT_USER_EMAIL`, `ROOT_USER_PASSWORD`
- [x] 2.6 PR2 Correction: DTO `Validate()` methods using `internal/platform/validator`; handlers bind without binding tags and return `ValidationError`; roles/users DTOs and CLI templates aligned.
- [x] 2.7 PR2 Correction: Root user business rules enforced in users service — API cannot create/modify root role users; root edits restricted to self (actorPublicID boundary wired, pending PR3 auth context); root has no company; CLI root creation preserved.

## Phase 3: Auth & Middleware

- [x] 3.1 Create `internal/modules/auth/{service,handler,routes}.go` for login, refresh, logout, me
- [x] 3.2 Implement login: verify password, issue PASETO access + opaque refresh, store hash
- [x] 3.3 Implement refresh: validate refresh hash, rotate pair, revoke old hash
- [x] 3.4 Implement logout: revoke provided refresh token
- [x] 3.5 Implement `GET /auth/me` returning authenticated user without password/hash
- [x] 3.6 Implement `internal/middleware/auth.go`: Bearer validation, user lookup, active check, context injection
- [x] 3.7 Wire all modules into `internal/app/container.go` and mount protected groups in router

## Phase 4: Testing & Verification

- [x] 4.1 Unit tests for identity, password, token with table-driven cases
- [x] 4.2 Handler tests for roles, users, auth using `httptest` and fake services
- [x] 4.3 Middleware tests: valid/missing/expired token, inactive user rejection
- [x] 4.4 Integration tests for migration, seed idempotency, root create/re-run, and auth DB refresh-token flows; skip with `testing.Short()` when external infrastructure is required
- [x] 4.5 Verify no password or `password_hash` leaks in any JSON response

## Phase 5: Cleanup

- [x] 5.1 Remove TODO comments from platform files and module stubs relevant to auth/users/roles
- [x] 5.2 Update `.env.example` docs for new auth variables
