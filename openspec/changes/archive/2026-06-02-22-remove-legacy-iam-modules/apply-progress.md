# Apply Progress: Remove Legacy IAM Modules

## Current Batch

- **Work unit**: PR 1 — migrate production code references to IAM types
- **Delivery strategy**: chained PRs
- **Chain strategy**: feature-branch-chain
- **Mode**: Standard implementation with targeted verification

## Completed Tasks

- [x] 1.1 Replace `roles.Role`, `roles.RootRoleSlug`, and `users.User` imports in `internal/cli/commands/root_storage.go` with `iamcore.IAMRole`, `iamcore.RootRoleSlug`, and `iamcore.IAMUser`
- [x] 1.2 Update type references and struct literals in `internal/cli/commands/root_storage.go` to use IAM model field names
- [x] 1.3 Migrate `internal/cli/commands/root_storage_test.go` assertions and fixtures from legacy types to IAM types
- [x] 1.4 Migrate `internal/cli/commands/integration_test.go` from legacy types to IAM types
- [x] 1.5 Replace `roles.Role` and `roles.RootRoleSlug` imports in `seeds/roles.go` with `iamcore.IAMRole` and `iamcore.RootRoleSlug`
- [x] 1.6 Update seed struct literals and slug comparisons in `seeds/roles.go` to use IAM constants
- [x] 1.7 Migrate `seeds/roles_test.go` assertions from legacy types to IAM types
- [x] 1.8 Verify: `go build ./internal/cli/... ./seeds/...` and `go test ./internal/cli/... ./seeds/...` pass

## Files Changed

| File | Action | Summary |
|------|--------|---------|
| `internal/cli/commands/root_storage.go` | Modified | Replaced legacy user/role models and constants with IAM core equivalents. |
| `internal/cli/commands/root_storage_test.go` | Modified | Migrated in-memory migrations, fixtures, and assertions to IAM core models. |
| `internal/cli/commands/integration_test.go` | Modified | Migrated integration setup from legacy user/role models to IAM core models. |
| `seeds/roles.go` | Modified | Replaced legacy role model and root slug constant with IAM core equivalents. |
| `seeds/roles_test.go` | Modified | Migrated seed test migrations, counts, and lookups to IAM core role model. |
| `openspec/changes/22-remove-legacy-iam-modules/tasks.md` | Modified | Marked Phase 1 tasks complete. |

## Verification

| Command | Result |
|---------|--------|
| `go build ./internal/cli/... ./seeds/...` | Passed |
| `go test ./internal/cli/... ./seeds/...` | Passed |

## Deviations

None — implementation stayed within Phase 1 / PR 1 scope.

## Remaining Work

All implementation and archive-sync tasks are complete.

## Batch 2

- **Work unit**: PR 2 — migrate test infrastructure to IAM types
- **Delivery strategy**: chained PRs
- **Chain strategy**: feature-branch-chain
- **Mode**: Standard implementation with targeted verification

## Batch 2 Completed Tasks

- [x] 2.1 Update `tests/fixtures/factories.go` — replace `users.User`, `roles.Role`, `permissions.Permission` return types with `iamcore.IAMUser`, `iamcore.IAMRole`, `iamcore.IAMPermission`; replace `permissions.ActionView` with `platformPerms.ActionView`
- [x] 2.2 Update `tests/helpers/auth.go` — migrate `Actor` struct, `SeedUser`, and `CreateTestToken` to use IAM types
- [x] 2.3 Update `tests/helpers/fixtures.go` — migrate seed helper return types to IAM types
- [x] 2.4 Update `tests/helpers/database_test.go` — replace `roles.Role` in `AutoMigrate` with `iamcore.IAMRole`
- [x] 2.5 Update `tests/helpers/fixtures_test.go` — replace all legacy type assertions with IAM equivalents
- [x] 2.6 Update `tests/helpers/auth_test.go` — replace `users.User` and `roles.Role` with IAM types
- [x] 2.7 Update `tests/integration/auth_test.go` — replace `roles.Role` and `users.User` in `AutoMigrate` and setup with IAM types
- [x] 2.8 Update `tests/integration/users_test.go` — replace legacy users module wiring with IAM users container and IAM DTOs
- [x] 2.9 Update `tests/integration/tenant_test.go` — replace legacy users module wiring with IAM users container and IAM DTOs
- [x] 2.10 Update `tests/integration/users_isolation_test.go` — replace legacy users module wiring with IAM users container and IAM DTOs
- [x] 2.11 Remove `tests/integration/role_resolver_adapter_test.go` because Phase 2 integration tests no longer use the legacy role repository adapter
- [x] 2.12 Verify: `go test ./tests/...` passes with zero failures

## Batch 2 Files Changed

| File | Action | Summary |
|------|--------|---------|
| `tests/fixtures/factories.go` | Modified | Factory types now use IAM core models and platform permission action constants. |
| `tests/helpers/auth.go` | Modified | Auth helpers now seed and issue tokens for IAM users/roles. |
| `tests/helpers/fixtures.go` | Modified | Seed helpers now return IAM core user/role/permission models. |
| `tests/helpers/database_test.go` | Modified | Database isolation test now migrates/counts IAM roles. |
| `tests/helpers/fixtures_test.go` | Modified | Fixture tests now assert IAM models and dot-format permission slugs. |
| `tests/helpers/auth_test.go` | Modified | Auth helper tests now use IAM core models. |
| `tests/integration/auth_test.go` | Modified | Auth integration setup now uses IAM core users/roles. |
| `tests/integration/users_test.go` | Modified | Users CRUD integration now exercises IAM users routes/container. |
| `tests/integration/tenant_test.go` | Modified | Tenant integration now exercises IAM users routes/container. |
| `tests/integration/users_isolation_test.go` | Modified | Users isolation test now exercises IAM users routes/container. |
| `tests/integration/role_resolver_adapter_test.go` | Deleted | Legacy adapter removed after IAM users integration no longer needs it. |
| `openspec/changes/22-remove-legacy-iam-modules/tasks.md` | Modified | Marked Phase 2 tasks complete. |

## Batch 2 Verification

| Command | Result |
|---------|--------|
| `go test ./tests/...` | Passed |
| `go build ./internal/cli/... ./seeds/... && go test ./internal/cli/... ./seeds/... ./tests/...` | Passed |
| Search for test imports of `internal/modules/users`, `internal/modules/roles`, `internal/modules/permissions` | Passed — no imports found |

## Batch 2 Deviations

- `tests/integration/users_test.go`, `tenant_test.go`, and `users_isolation_test.go` were migrated from legacy users service wiring to the IAM users container, because keeping legacy handlers/services would preserve the legacy dependency this change is removing.

## Batch 3

- **Work unit**: PR 3 — delete legacy users/roles/permissions directories and verify
- **Delivery strategy**: chained PRs
- **Chain strategy**: feature-branch-chain
- **Mode**: Standard implementation with full verification

## Batch 3 Completed Tasks

- [x] 3.1 Delete `internal/modules/users/` directory (12 files)
- [x] 3.2 Delete `internal/modules/roles/` directory (14 files)
- [x] 3.3 Delete `internal/modules/permissions/` directory (42 files)
- [x] 3.4 Verify: `go list ./...` shows zero packages under `internal/modules/users`, `internal/modules/roles`, or `internal/modules/permissions`
- [x] 3.5 Verify: `go build ./...` succeeds with zero errors
- [x] 3.6 Verify: `go test ./...` passes with zero failures
- [x] 3.7 Targeted check: confirm public users/roles/permissions routes respond correctly via IAM (integration test assertions from Phase 2 cover this)
- [x] 3.8 Targeted check: confirm auth/login/session user resolution works via IAM (auth integration tests from Phase 2 cover this)
- [x] 3.9 Targeted check: confirm seeds/bootstrap/sync permissions operate independently from legacy (seed tests from Phase 1 cover this)

## Batch 3 Files Changed

| File | Action | Summary |
|------|--------|---------|
| `internal/modules/users/` | Deleted | Removed legacy users module after all external imports were migrated. |
| `internal/modules/roles/` | Deleted | Removed legacy roles module after all external imports were migrated. |
| `internal/modules/permissions/` | Deleted | Removed legacy permissions module after all external imports were migrated. |
| `openspec/changes/22-remove-legacy-iam-modules/tasks.md` | Modified | Marked Phase 3 tasks complete. |

## Batch 3 Verification

| Command | Result |
|---------|--------|
| Search for Go imports of `internal/modules/users`, `internal/modules/roles`, `internal/modules/permissions` | Passed — no imports found outside deleted files before deletion; no imports found after deletion |
| Glob for `internal/modules/users/**/*.go`, `internal/modules/roles/**/*.go`, `internal/modules/permissions/**/*.go` | Passed — no files found |
| `go list ./...` | Passed — no legacy module packages listed |
| `go build ./...` | Passed |
| `go test ./...` | Passed |

## Batch 3 Deviations

None — implementation stayed within Phase 3 / PR 3 scope. Phase 4 archive-time spec updates remain pending.

## Batch 4

- **Work unit**: Archive — sync delta specs into canonical OpenSpec specs
- **Mode**: OpenSpec archive

## Batch 4 Completed Tasks

- [x] 4.1 Update `openspec/specs/iam-module/spec.md` — remove "Legacy module preservation" requirement, add "No residual legacy references" requirement
- [x] 4.2 Update `openspec/specs/app-orchestration/spec.md` — remove "Legacy modules still compile" scenario, add "No legacy module directories exist" scenario
- [x] 4.3 Mark `openspec/specs/users/spec.md` as superseded by `iam-module`
- [x] 4.4 Mark `openspec/specs/roles/spec.md` as superseded by `iam-module`
- [x] 4.5 Mark `openspec/specs/permissions/spec.md` as superseded by `iam-module`
- [x] 4.6 Update `openspec/specs/rbac-authorization/spec.md` — remove legacy file path references, add IAM-only constant scenarios
- [x] 4.7 Update `openspec/specs/platform-boundary-rules/spec.md` — remove legacy file path references, add IAM-only domain language scenarios

## Batch 4 Verification

| Command | Result |
|---------|--------|
| `go list ./... && go build ./... && go test ./...` | Passed after archive conformance fix |

## Batch 4 Notes

- Archive sync found that delta specs expected IAM-owned module constants and a role-assigned-users message constant. Code was updated before archive to avoid making canonical specs contradict implementation.
