# Proposal: Authentication, Users, Roles & Root

## Intent

Add authentication and user management with PASETO v4.local tokens, opaque hashed refresh tokens, single-role-per-user, and an idempotent root-user CLI.

## Scope

### In Scope
- PASETO access tokens and opaque refresh tokens (hashed, revocable).
- Auth endpoints: login, refresh, logout, me.
- Users CRUD, password change, active/inactive status.
- Roles read-only API with seeds (`root`, `admin`, `user`).
- Root creation via CLI (`create-root`) with env or interactive input.
- Auth middleware, DB migrations, and seeds.

### Out of Scope
- Permission-based authorization beyond role checks.
- Multiple roles per user, OAuth, SSO, MFA, Redis revocation.

## Capabilities

### New
- `auth`: Token lifecycle (issue, parse, refresh, revoke) and auth endpoints.
- `users`: CRUD, password change, status toggle.
- `roles`: Read-only list/get and system seeds.
- `cli-root`: Idempotent root-user creation command.
- `middleware-auth`: PASETO validation, user lookup, context injection.

### Modified
- None.

## Approach

Build platform leaves (identity, password, token) first. Add config, migrations, and seeds. Implement modules bottom-up (roles → users → auth). Wire in `container.go`/`router.go`, then connect CLI root to real storage/hasher.

## Affected Areas

| Area | Impact |
|------|--------|
| `internal/platform/{identity,password,token}` | New wrappers |
| `internal/modules/{roles,users,auth}` | New modules |
| `internal/middleware/auth.go` | New middleware |
| `internal/app/container.go` | Wire deps |
| `internal/cli/commands/createroot.go` | Wire storage/hasher |
| `migrations/` | New tables |
| `seeds/` | New seeds |
| `.env.example` | New vars |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `paseto` API changes | Low | Pin stable release |
| Circular imports | Med | Middleware uses injected interfaces only |
| Container bloat | Low | Flat, grouped fields |

## Rollback Plan

1. Revert `go.mod` to remove `paseto`.  
2. Reverse Goose migrations for new tables.  
3. Revert `container.go`, `router.go`, and middleware.

## Dependencies

- `github.com/o1egl/paseto`  
- Goose migrations applied before seeds/API boot

## Success Criteria

- [ ] Migrations create `roles`, `users`, `refresh_tokens`.  
- [ ] Role seeds and idempotent root creation work.  
- [ ] Login returns PASETO access + opaque refresh token.  
- [ ] Refresh rotates token pair; logout revokes refresh.  
- [ ] Middleware rejects inactive users and invalid tokens.  
- [ ] Passwords hashed with argon2id; no leaks in responses.
