# Proposal: Remove Legacy IAM Modules

## Intent

The legacy modules `internal/modules/users/`, `internal/modules/roles/`, and `internal/modules/permissions/` are dead code. IAM (`internal/modules/iam/`) is the sole boundary for users, roles, and permissions. Legacy modules compile but their routes are not mounted and no production wiring references them. Keeping ~68 dead Go files creates confusion about the source of truth and burdens spec accuracy.

## Scope

### In Scope
- Migrate 16 external consumers (CLI, seeds, test infrastructure) from legacy types to IAM equivalents
- Delete `internal/modules/users/` (12 files), `internal/modules/roles/` (14 files), `internal/modules/permissions/` (42 files)
- Update OpenSpec specs that reference legacy preservation as active state
- Verify `go build ./...`, `go test ./...`, `go list ./...` pass clean

### Out of Scope
- Changes to public routes, payloads, or HTTP behavior
- Database schema or migration changes
- Additional IAM refactoring unrelated to legacy removal

## Capabilities

### New Capabilities
None

### Modified Capabilities
- `iam-module`: Remove "Legacy module preservation" requirement and associated scenario
- `app-orchestration`: Remove "Legacy modules still compile" scenario
- `users`: Mark spec as superseded by `iam-module`
- `roles`: Mark spec as superseded by `iam-module`
- `permissions`: Mark spec as superseded by `iam-module`
- `rbac-authorization`: Update legacy file path references
- `platform-boundary-rules`: Update legacy file path references

## Approach

**Migrate-then-delete** in three phases:

1. **Phase A — Migrate production code**: Update `internal/cli/commands/root_storage.go` and `seeds/roles.go` to use IAM types (`iamcore.IAMRole`, `iamcore.IAMUser`, `iamcore.RootRoleSlug`). Mechanical swaps — IAM models have identical GORM tags and table names.

2. **Phase B — Migrate test infrastructure**: Update `tests/fixtures/factories.go`, `tests/helpers/*.go`, and all integration tests to use IAM types. Replace `permissions.ActionView` with `platformPerms.ActionView`.

3. **Phase C — Delete and verify**: Remove all three legacy directories. Run full build/test/list. Update OpenSpec specs.

### Type Mapping

| Legacy Type | IAM Replacement |
|---|---|
| `users.User` | `iamcore.IAMUser` |
| `roles.Role` | `iamcore.IAMRole` |
| `permissions.Permission` | `iamcore.IAMPermission` |
| `roles.RootRoleSlug` | `iamcore.RootRoleSlug` |
| `permissions.ActionView` | `platformPerms.ActionView` |
| `roles.Repository` | Remove or replace with IAM resolver |

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/users/` | Removed | 12 files deleted |
| `internal/modules/roles/` | Removed | 14 files deleted |
| `internal/modules/permissions/` | Removed | 42 files deleted |
| `internal/cli/commands/root_storage.go` | Modified | Migrate to IAM types |
| `seeds/roles.go` | Modified | Migrate to IAM types |
| `tests/fixtures/factories.go` | Modified | Migrate factory types |
| `tests/helpers/*.go` | Modified | Migrate helper types |
| `tests/integration/*_test.go` | Modified | Migrate test types (5 files) |
| `openspec/specs/iam-module/spec.md` | Modified | Remove legacy preservation requirement |
| `openspec/specs/app-orchestration/spec.md` | Modified | Remove legacy compile scenario |
| `openspec/specs/users/spec.md` | Modified | Mark superseded |
| `openspec/specs/roles/spec.md` | Modified | Mark superseded |
| `openspec/specs/permissions/spec.md` | Modified | Mark superseded |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `role_resolver_adapter_test.go` uses `roles.Repository` with no IAM equivalent | Med | Rewrite against `iamcore.RoleBySlugResolver` or remove |
| Test assertions on legacy-specific methods (e.g., `User.IsRoot()`) | Low | Audit each usage during Phase B migration |
| Spec drift if OpenSpec updates missed | Low | Explicit checklist of 7 specs to update in Phase C |

## Rollback Plan

Each phase produces independent commits. If Phase A or B migration breaks tests, revert that phase's commit. If Phase C (deletion) causes unexpected failures, revert the deletion commit — legacy directories remain in git history and can be restored with `git revert`. No data or schema changes involved, so rollback is purely source-level.

## Dependencies

- None — IAM module is already fully operational

## Success Criteria

- [ ] `go build ./...` succeeds with zero errors
- [ ] `go test ./...` passes with zero failures
- [ ] `go list ./...` shows no references to `internal/modules/users`, `internal/modules/roles`, or `internal/modules/permissions`
- [ ] All public routes (users, roles, permissions) respond correctly via IAM
- [ ] Seeds and CLI commands function without legacy imports
- [ ] OpenSpec specs updated — no legacy preservation requirements remain active
