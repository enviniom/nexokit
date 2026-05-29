# Design: Migrate Auth Module to Vertical Slice Architecture

## Technical Approach

Refactor only `internal/modules/auth/` from one flat handler/service/repository to four endpoint-aligned slices. Preserve routes, JSON contracts, PASETO/refresh-token behavior, DB schema, and app-level `userLookup`. This design implements the auth and vertical-slice delta specs plus `proposal.md`, `exploration.md`, and project vertical-slice rules.

Current → target:

```txt
auth/{handler,service,repository,model,dto,routes}.go
  → auth/{container,routes}.go
  → auth/core/{model,dto,error}.go
  → auth/queries/{find_user_by_email,find_refresh_token_by_hash_with_user}.go
  → auth/{authenticate_user,rotate_token,revoke_token,view_session}/...
```

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Slice names | `authenticate_user`, `rotate_token`, `revoke_token`, `view_session` | HTTP/mechanical names like `login` or `get_me` | Matches business intention naming from project rules. |
| Shared data access | Put reused lookups in `queries/`; slice repositories delegate | Duplicate GORM in each slice | `rotate_token` and `revoke_token` share refresh-token lookup; rules require reusable queries outside `core/`. |
| Cross-module data | Local partial models in `core/` for `users`, `roles`, `refresh_tokens` | Import `users.User`, `roles.Role`, or `users.Repository` | Auth must be self-contained; migrations remain schema source. |
| Wiring | `auth.NewContainer(db, verifier, issuer, refreshManager, refreshTTL)` and `auth.Register(v1, container, ...)` | Root app wires each slice | Root container should know module container only, not slices. |

## Data Flow

```txt
POST /auth/login   → authenticate_user.Handler → Service → Repository → queries.FindUserByEmail → DB
POST /auth/refresh → rotate_token.Handler       → Service → Repository → queries.FindRefreshTokenByHashWithUser → DB
POST /auth/logout  → revoke_token.Handler       → Service → Repository → queries.FindRefreshTokenByHashWithUser → DB
GET /auth/me       → view_session.Handler       → authctx.FromGin → response
```

All handlers continue using `platform/response`, `platform/messages`, and local DTO validation.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/auth/core/model.go` | Create | `AuthUser`, `AuthRole`, `RefreshToken`; table names map to existing tables. |
| `internal/modules/auth/core/dto.go` | Create | `LoginRequest`, `RefreshRequest`, `TokenPairResponse`, `LoginResponse`, `MeResponse`, `AuthUserResponse`. |
| `internal/modules/auth/core/error.go` | Create | Module constants/errors if needed; otherwise minimal. |
| `internal/modules/auth/queries/*.go` | Create | `FindUserByEmail`, `FindRefreshTokenByHashWithUser`, with query tests. |
| `internal/modules/auth/{authenticate_user,rotate_token,revoke_token,view_session}/` | Create | Each slice owns `handler.go`, `service.go`, `repository.go`, and tests. |
| `internal/modules/auth/container.go` | Create | Composition root only; constructs slice repositories/services/handlers. |
| `internal/modules/auth/routes.go` | Modify | Dispatch current endpoints to container slice handlers. |
| `internal/app/container.go` | Modify | Replace `auth.NewService/NewHandler` with `auth.NewContainer`; keep app-level `userLookup`. |
| Legacy auth flat files/tests | Delete/Move | Remove after equivalent slice/core coverage exists. |

## Interfaces / Contracts

Endpoint → slice map:

| Endpoint | Slice | Contract |
|---|---|---|
| `POST /auth/login` | `authenticate_user` | Credentials in, access+opaque refresh+sanitized user out. |
| `POST /auth/refresh` | `rotate_token` | Refresh token in, rotated pair out, old hash revoked with replacement. |
| `POST /auth/logout` | `revoke_token` | Refresh token in, usable hash revoked. |
| `GET /auth/me` | `view_session` | Context user plus role slug/permissions out. |

`core.AuthUser` includes only fields auth reads: internal ID, public ID, name, email, password hash, active flag, role ID, company ID, timestamps/audit fields, and `Role AuthRole`. No password hash appears in responses.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Slice handlers validation, status mapping, no password leaks | `httptest`, fake services, table-driven cases where repeated. |
| Unit | Services: login failures, inactive user, rotation, revoked/expired token, logout | Small fakes for slice repositories/token/password boundaries. |
| Integration | `queries/` and repositories against SQLite | `AutoMigrate` local partial models only; no `users`/`roles` imports. |
| Module route | Existing `/auth/*` behavior and middleware placement | Gin router tests for route registration and unchanged endpoint surface. |

Run narrow package first: `go test ./internal/modules/auth/...`, then `go test ./...`.

## Migration / Rollout

No data migration required. Sequence: create `core/` partial models and `queries/`; migrate `authenticate_user`; migrate `rotate_token` and `revoke_token`; migrate `view_session`; switch `container.go`/`routes.go`/`internal/app/container.go`; delete flat files after tests pass. Rollback is a git revert; no schema/API changes.

Risks and mitigations: partial GORM relations may not preload roles correctly, so prove them with query integration tests. Review size may exceed 800 lines, so task planning should split reviewable chained PRs. Accidental behavior drift is mitigated by porting all existing cases plus route behavior tests before deleting legacy files.

## Open Questions

None blocking.

## Non-Goals

- No migration of modules outside `internal/modules/auth/`.
- No behavior, API response, route, auth middleware, token, password, or DB schema change.
- No change to app-level `userLookup` dependency on `users.Repository`.
