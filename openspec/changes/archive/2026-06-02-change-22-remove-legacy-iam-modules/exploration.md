## Exploration: Remove Legacy IAM Modules

### Current State

The IAM module (`internal/modules/iam/`) is fully operational as the single boundary for users, roles, and permissions. The app container (`internal/app/container.go`) already wires IAM exclusively — no legacy module imports exist in production wiring. IAM defines its own local GORM models (`IAMUser`, `IAMRole`, `IAMPermission`) in `iam/core/model.go` that map to the same database tables (`users`, `roles`, `permissions`) with identical schema tags.

Legacy modules (`users/`, `roles/`, `permissions/`) still compile independently but their routes are **not mounted**. They exist as dead code with ~68 Go files across three directories.

However, **16 external files** still import legacy types for test fixtures, seeds, CLI commands, and integration tests.

### Affected Areas

**Production code (must migrate):**
- `internal/cli/commands/root_storage.go` — Uses `roles.Role`, `roles.RootRoleSlug`, `users.User` for root user creation
- `seeds/roles.go` — Uses `roles.Role`, `roles.RootRoleSlug` for role seeding

**Test infrastructure (must migrate):**
- `tests/fixtures/factories.go` — Factory functions use `users.User`, `roles.Role`, `permissions.Permission`, `permissions.ActionView`
- `tests/helpers/fixtures.go` — Seed helpers return legacy types
- `tests/helpers/auth.go` — `Actor` struct, `SeedUser`, `CreateTestToken` use legacy types
- `tests/helpers/database_test.go` — Uses `roles.Role` for AutoMigrate
- `tests/helpers/fixtures_test.go` — Uses all three legacy types
- `tests/helpers/auth_test.go` — Uses `users.User`, `roles.Role`

**Integration tests (must migrate):**
- `tests/integration/auth_test.go` — Uses `roles.Role`, `users.User` for AutoMigrate and setup
- `tests/integration/users_test.go` — Uses `roles.Role`, `users.User`
- `tests/integration/tenant_test.go` — Uses `roles.Role`, `users.User`
- `tests/integration/users_isolation_test.go` — Uses `roles.Role`, `users.User`
- `tests/integration/role_resolver_adapter_test.go` — Uses `roles.Repository`, `roles.Role`

**CLI/seeds tests (must migrate):**
- `internal/cli/commands/root_storage_test.go` — Uses `roles.Role`, `users.User`
- `internal/cli/commands/integration_test.go` — Uses `roles.Role`, `users.User`
- `seeds/roles_test.go` — Uses `roles.Role`

**Legacy modules (to delete):**
- `internal/modules/users/` — 12 Go files
- `internal/modules/roles/` — 14 Go files
- `internal/modules/permissions/` — 42 Go files

**OpenSpec specs (to update):**
- `openspec/specs/iam-module/spec.md` — Remove "Legacy module preservation" requirement (lines 173-181)
- `openspec/specs/app-orchestration/spec.md` — Remove "Legacy modules still compile" scenario (lines 109-113)
- `openspec/specs/users/spec.md` — Mark as superseded by IAM
- `openspec/specs/roles/spec.md` — Mark as superseded by IAM
- `openspec/specs/permissions/spec.md` — Mark as superseded by IAM
- `openspec/specs/rbac-authorization/spec.md` — Update legacy file path references
- `openspec/specs/platform-boundary-rules/spec.md` — Update legacy file path references

### Approaches

1. **Migrate-then-delete** — Migrate all 16 external consumers to IAM types first, then delete legacy directories in one step.
   - Pros: Clean separation of concerns; each migration can be verified independently; safe rollback at each step
   - Cons: More intermediate commits; slightly more total effort
   - Effort: Medium

2. **Big-bang delete with fix-up** — Delete all three legacy directories at once, then fix all compile errors.
   - Pros: Single decisive action; compiler reveals all broken references
   - Cons: Large diff; harder to review; if tests break it's harder to isolate cause; risky if something unexpected depends on legacy
   - Effort: Low-Medium

### Recommendation

**Approach 1 (Migrate-then-delete)** is recommended, organized as three task phases:

1. **Phase A — Migrate production code**: Update `root_storage.go` and `seeds/roles.go` to use `iamcore.IAMRole`, `iamcore.IAMUser`, `iamcore.RootRoleSlug`. These are mechanical swaps since IAM models have identical GORM tags and table names.

2. **Phase B — Migrate test infrastructure**: Update `tests/fixtures/factories.go`, `tests/helpers/*.go`, and all integration tests to use IAM types. Replace `permissions.ActionView` etc. with `platform/permissions.ActionView` (already used by IAM routes).

3. **Phase C — Delete and verify**: Remove `internal/modules/users/`, `internal/modules/roles/`, `internal/modules/permissions/`. Run `go build ./...`, `go test ./...`, `go list ./...`. Update OpenSpec specs.

Key type mapping:

| Legacy Type | IAM Replacement |
|---|---|
| `users.User` | `iamcore.IAMUser` |
| `roles.Role` | `iamcore.IAMRole` |
| `permissions.Permission` | `iamcore.IAMPermission` |
| `roles.RootRoleSlug` | `iamcore.RootRoleSlug` |
| `roles.AdminRoleSlug` | `iamcore.AdminRoleSlug` |
| `roles.UserRoleSlug` | `iamcore.UserRoleSlug` |
| `permissions.ActionView` | `platformPerms.ActionView` |
| `roles.Repository` | Remove or replace with IAM resolver |
| `roles.AssignmentRoleReader` | Already unused outside legacy |

### Risks

- **GORM AutoMigrate compatibility**: IAM models must produce identical table schemas. Verified — `IAMUser`, `IAMRole`, `IAMPermission` have matching GORM tags and `TableName()` overrides.
- **Test fixture breakage**: 16 files need type migration. Mechanical but error-prone if any test asserts on legacy-specific fields/methods (e.g., `User.IsRoot()`). Must audit each usage.
- **`role_resolver_adapter_test.go`**: Uses `roles.Repository` interface which has no IAM equivalent. This test adapter may need removal or rewrite against `iamcore.RoleBySlugResolver`.
- **OpenSpec spec drift**: Multiple specs reference legacy paths. Must update or they become misleading. Low risk to runtime but high risk to spec accuracy.
- **Line count**: ~68 files deleted + ~16 files modified. Deletions will dominate the diff but the modified files need careful review. Chained PR strategy recommended if total changed lines exceed review budget.

### Ready for Proposal

**Yes.** The codebase is well-understood, IAM is fully operational, and the migration path is clear. The orchestrator should proceed to `sdd-propose` with this exploration as input. The proposal should define the three-phase approach and note that test infrastructure migration is the largest work unit.
