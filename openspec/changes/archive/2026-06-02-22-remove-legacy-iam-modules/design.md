# Design: Remove Legacy IAM Modules

## Technical Approach

Use a migrate-then-delete sequence. Runtime wiring already uses IAM only: `internal/app/container.go` builds `iam.NewContainer`, registers IAM routes through `iam.Register`, and delegates permission sync to `c.IAM.Syncer`. The work removes remaining compile-time dependencies on `internal/modules/users`, `internal/modules/roles`, and `internal/modules/permissions` from production helpers and tests, then deletes those legacy directories. Public routes, payloads, database tables, and migrations stay unchanged because IAM models map to the existing `users`, `roles`, `permissions`, and `role_permissions` tables.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Source of truth | Use `internal/modules/iam/core` models/constants for all remaining user, role, and permission references. | Keep compatibility aliases or leave legacy packages compiling. | The proposal and specs require IAM as the sole boundary; aliases would preserve dead code and keep imports ambiguous. |
| Sequencing | Migrate production and test references before deleting directories. | Big-bang delete and fix compiler errors afterward. | Keeps each step reviewable and makes failures attributable to a migration group. |
| Test adapter handling | Remove or rewrite `tests/integration/role_resolver_adapter_test.go` with IAM-facing tests; do not create a new IAM equivalent for legacy `roles.Repository`. | Add an adapter that mimics `users.RoleResolver`. | The adapter only supports legacy `users.NewService`; IAM already exposes `RoleBySlugResolver` and app-level adapter coverage exists in `internal/app/container_test.go`. |
| IAM scope | Preserve IAM internals unless a change is needed to remove a legacy reference or satisfy delta specs. | Opportunistic IAM route/service refactors. | The change is source cleanup; unrelated IAM refactors increase behavioral risk. |

## Data Flow

Runtime path after cleanup remains:

```text
Bootstrap -> app.NewContainer -> iam.NewContainer
          -> server.NewRouter -> Container.RegisterModules -> IAM routes
          -> Container.SyncPermissions -> IAM Syncer -> permissions table
```

Removal flow:

```text
legacy import consumers -> IAM/platform type swap -> tests compile
                    -> delete legacy dirs -> go list/test/build prove no residual refs
```

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/cli/commands/root_storage.go`, `internal/cli/commands/root_storage_test.go`, `internal/cli/commands/integration_test.go` | Modify | Replace `roles.Role`, `roles.RootRoleSlug`, and `users.User` with IAM models/constants. |
| `seeds/roles.go`, `seeds/roles_test.go` | Modify | Seed root role using `iamcore.IAMRole` and IAM role slugs. |
| `tests/fixtures/factories.go` | Modify | Return `iamcore.IAMUser`, `IAMRole`, `IAMPermission`; use `platformPerms.Action*`. |
| `tests/helpers/auth.go`, `tests/helpers/fixtures.go`, `tests/helpers/*_test.go` | Modify | Migrate helper return types, AutoMigrate models, actor structs, and token setup to IAM types. |
| `tests/integration/auth_test.go`, `users_test.go`, `tenant_test.go`, `users_isolation_test.go`, `role_resolver_adapter_test.go` | Modify/Delete | Move integration coverage to IAM handlers/services or remove legacy-only adapter coverage. |
| `internal/modules/users/`, `internal/modules/roles/`, `internal/modules/permissions/` | Delete | Remove dead legacy modules after all external imports are gone. |
| `openspec/specs/{iam-module,app-orchestration,users,roles,permissions,rbac-authorization,platform-boundary-rules}/spec.md` | Modify at archive | Merge existing deltas so active specs no longer require legacy preservation. |

## Interfaces / Contracts

No public HTTP or DB contract changes. Type mapping is mechanical: `users.User` -> `iamcore.IAMUser`, `roles.Role` -> `iamcore.IAMRole`, `permissions.Permission` -> `iamcore.IAMPermission`, `roles.RootRoleSlug/AdminRoleSlug/UserRoleSlug` -> IAM constants, and `permissions.Action*` -> `internal/platform/permissions.Action*`. `roles.Repository` has no replacement; migrate tests away from the legacy service boundary.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Static/compile | No packages or imports remain for deleted modules. | Run `go list ./...`; targeted search for `internal/modules/(users|roles|permissions)` in Go files. |
| Unit | CLI root storage, seed idempotency, fixtures/helpers, IAM syncer. | Run package tests for `internal/cli/commands`, `seeds`, `tests/helpers`, `tests/fixtures`, and `internal/modules/iam/internal/sync_permissions`. |
| Integration | Public users/roles/permissions routes, auth lookup/token setup, tenant isolation. | Run `go test ./tests/...` and keep route assertions in `internal/app/container_test.go`. |
| Full verification | Whole repository builds and tests. | Run `go test ./...` and `go build ./...`. |

Targeted checks must cover public route preservation, auth/root bootstrap, role seeds, app bootstrap permission sync, and absence of legacy directories.

## Migration / Rollout

No data migration required. Roll out as source-only phases: production reference migration, test migration, delete legacy directories plus verification. Rollback is a git revert of the failing phase.

## Open Questions

None.
