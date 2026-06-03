# Tasks: Remove Legacy IAM Modules

> **STOP**: Do NOT start apply/implementation until the maintainer has reviewed the proposal, specs, design, and these tasks. All artifacts are planning outputs — no source code should be modified yet.

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~5,300–8,500 (deletions dominate: 68 files) |
| 800-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 (see work units below) |
| Delivery strategy | ask-always |
| Chain strategy | pending — maintainer must choose before apply |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
800-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Migrate production code references to IAM types | PR 1 | ~50–85 lines changed; base = main |
| 2 | Migrate test infrastructure to IAM types | PR 2 | ~180–305 lines changed; base = PR 1 branch or main (depends on chain strategy) |
| 3 | Delete legacy directories, verify, update specs | PR 3 | ~5,000–8,000 deletions; low cognitive load despite line count |

## Phase 1: Migrate Production Code References

- [x] 1.1 Replace `roles.Role`, `roles.RootRoleSlug`, and `users.User` imports in `internal/cli/commands/root_storage.go` with `iamcore.IAMRole`, `iamcore.RootRoleSlug`, and `iamcore.IAMUser`
- [x] 1.2 Update type references and struct literals in `internal/cli/commands/root_storage.go` to use IAM model field names
- [x] 1.3 Migrate `internal/cli/commands/root_storage_test.go` assertions and fixtures from legacy types to IAM types
- [x] 1.4 Migrate `internal/cli/commands/integration_test.go` from legacy types to IAM types
- [x] 1.5 Replace `roles.Role` and `roles.RootRoleSlug` imports in `seeds/roles.go` with `iamcore.IAMRole` and `iamcore.RootRoleSlug`
- [x] 1.6 Update seed struct literals and slug comparisons in `seeds/roles.go` to use IAM constants
- [x] 1.7 Migrate `seeds/roles_test.go` assertions from legacy types to IAM types
- [x] 1.8 Verify: `go build ./internal/cli/... ./seeds/...` and `go test ./internal/cli/... ./seeds/...` pass

## Phase 2: Migrate Test Infrastructure

- [x] 2.1 Update `tests/fixtures/factories.go` — replace `users.User`, `roles.Role`, `permissions.Permission` return types with `iamcore.IAMUser`, `iamcore.IAMRole`, `iamcore.IAMPermission`; replace `permissions.ActionView` with `platformPerms.ActionView`
- [x] 2.2 Update `tests/helpers/auth.go` — migrate `Actor` struct, `SeedUser`, and `CreateTestToken` to use IAM types
- [x] 2.3 Update `tests/helpers/fixtures.go` — migrate seed helper return types to IAM types
- [x] 2.4 Update `tests/helpers/database_test.go` — replace `roles.Role` in `AutoMigrate` with `iamcore.IAMRole`
- [x] 2.5 Update `tests/helpers/fixtures_test.go` — replace all legacy type assertions with IAM equivalents
- [x] 2.6 Update `tests/helpers/auth_test.go` — replace `users.User` and `roles.Role` with IAM types
- [x] 2.7 Update `tests/integration/auth_test.go` — replace `roles.Role` and `users.User` in `AutoMigrate` and setup with IAM types
- [x] 2.8 Update `tests/integration/users_test.go` — replace legacy types with IAM types
- [x] 2.9 Update `tests/integration/tenant_test.go` — replace legacy types with IAM types
- [x] 2.10 Update `tests/integration/users_isolation_test.go` — replace legacy types with IAM types
- [x] 2.11 **Special**: Rewrite or remove `tests/integration/role_resolver_adapter_test.go` — `roles.Repository` has no IAM equivalent; rewrite against `iamcore.RoleBySlugResolver` or delete if coverage exists in `internal/app/container_test.go`
- [x] 2.12 Verify: `go test ./tests/...` passes with zero failures

## Phase 3: Delete Legacy Modules and Verify

- [x] 3.1 Delete `internal/modules/users/` directory (12 files)
- [x] 3.2 Delete `internal/modules/roles/` directory (14 files)
- [x] 3.3 Delete `internal/modules/permissions/` directory (42 files)
- [x] 3.4 Verify: `go list ./...` shows zero packages under `internal/modules/users`, `internal/modules/roles`, or `internal/modules/permissions`
- [x] 3.5 Verify: `go build ./...` succeeds with zero errors
- [x] 3.6 Verify: `go test ./...` passes with zero failures
- [x] 3.7 Targeted check: confirm public users/roles/permissions routes respond correctly via IAM (integration test assertions from Phase 2 cover this)
- [x] 3.8 Targeted check: confirm auth/login/session user resolution works via IAM (auth integration tests from Phase 2 cover this)
- [x] 3.9 Targeted check: confirm seeds/bootstrap/sync permissions operate independently from legacy (seed tests from Phase 1 cover this)

## Phase 4: Update OpenSpec Specs (at Archive)

- [x] 4.1 Update `openspec/specs/iam-module/spec.md` — remove "Legacy module preservation" requirement, add "No residual legacy references" requirement
- [x] 4.2 Update `openspec/specs/app-orchestration/spec.md` — remove "Legacy modules still compile" scenario, add "No legacy module directories exist" scenario
- [x] 4.3 Mark `openspec/specs/users/spec.md` as superseded by `iam-module`
- [x] 4.4 Mark `openspec/specs/roles/spec.md` as superseded by `iam-module`
- [x] 4.5 Mark `openspec/specs/permissions/spec.md` as superseded by `iam-module`
- [x] 4.6 Update `openspec/specs/rbac-authorization/spec.md` — remove legacy file path references, add IAM-only constant scenarios
- [x] 4.7 Update `openspec/specs/platform-boundary-rules/spec.md` — remove legacy file path references, add IAM-only domain language scenarios
