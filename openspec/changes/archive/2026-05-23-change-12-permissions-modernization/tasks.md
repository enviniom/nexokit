# Tasks: Permissions Modernization, Global Constants, and CRUD Refactoring

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 500–600 |
| User review budget | 700 changed lines |
| Budget risk | Medium |
| Chained PRs recommended | Optional |
| Delivery strategy | auto-chain (split as needed; since we stop at tasks, single integration branch works well for planning) |
| Chain strategy | featured-branch-chain |

Decision needed before apply: Yes — human should review generated proposal/spec/design/tasks before implementation.
Chained PRs recommended: Optional
700-line budget risk: Medium

---

## Phase 1: Baseline RED tests for modernisations and restrictions

- [x] 1.1 Add/update permissions handler/service tests to assert that structural fields (`Slug`, `Module`, `Action`) cannot be modified via `PUT /api/v1/permissions/:id`.
- [x] 1.2 Add router/integration test asserting that `POST /api/v1/permissions` and `DELETE /api/v1/permissions/:id` return `404 Not Found` (no longer routed).
- [x] 1.3 Add platform permissions test proving that `permissions.Format` returns correct slugs and constants match.
- [x] 1.4 Add platform permissions test proving that `permissions.Register` and `ListRegistered` thread-safely record and return unique slugs.
- [x] 1.5 Add integration test in GIN router proving that initializing the server automatically runs bootstrapping and syncs permissions.
- [x] 1.6 Run targeted tests and record RED evidence.

## Phase 2: Transversal Constants and Discovery Registry

- [x] 2.1 Create `internal/platform/permissions/constants.go` containing:
  - Global module name constants (`ModuleUsers`, `ModuleRoles`, etc.).
  - Global action constants (`ActionList`, `ActionSelect`, etc.).
  - Formatter helper `Format(module, action string) string`.
- [x] 2.2 Create `internal/platform/permissions/registry.go` containing:
  - In-memory registry with `sync.RWMutex` (`Register`, `ListRegistered`).

## Phase 3: Module-level Sync Logic and Bootstrapping Integration

- [x] 3.1 Modify `internal/middleware/authorization.go` — import `github.com/enviniom/nexokit/internal/platform/permissions`.
- [x] 3.2 In `RequirePermission(slug string)`, call `permissions.Register(slug)` to intercept and record the permission slug during route mounting.
- [x] 3.3 Modify `internal/modules/permissions/service.go` — implement `SyncPermissions(slugs []string) error` with:
  - Dynamic module and action extraction from slugs.
  - Humanisation helper mapping modules/actions to readable default Name/Description.
  - display order assignment map based on actions.
  - Idempotency logic: checking `slug` existence in the repo, inserting missing items (generating `PublicID` via `identity.Generate()`, setting `IsSystem = true`), and updating existing system attributes while preserving custom `Name` and `Description`.
- [x] 3.4 Modify `internal/app/container.go` — expose `SyncPermissions() error` wrapper delegating `permissions.ListRegistered()` to the `permissionsService`.
- [x] 3.5 Modify `internal/app/bootstrap.go` — call `container.SyncPermissions()` right after container wiring.

## Phase 4: Global CRUD Rename (`index` -> `list`)

- [x] 4.1 In `internal/modules/permissions/model.go`, remove or deprecate `ActionIndex = "index"`, ensuring `ActionList = "list"` exists.
- [x] 4.2 Modify `internal/modules/companies/routes.go` — replace `"companies.index"` with `permissions.Format(permissions.ModuleCompanies, permissions.ActionList)`.
- [x] 4.3 Modify `internal/modules/roles/routes.go` — replace `"roles.index"` with `permissions.Format(permissions.ModuleRoles, permissions.ActionList)` and refactor other endpoints to use constants.
- [x] 4.4 Modify `internal/modules/users/routes.go` — replace `"users.index"` with `permissions.Format(permissions.ModuleUsers, permissions.ActionList)` and refactor other endpoints to use constants.
- [x] 4.5 Modify `seeds/permissions.go` — update base permissions definition to use `permissions.ActionList` and global constants.
- [x] 4.6 Update `tests/integration/rbac_test.go` and other tests in modules/seeds to change `"users.index"`, `"roles.index"`, and `"companies.index"` to `"users.list"`, `"roles.list"`, and `"companies.list"`.

## Phase 5: Permissions API Hardening

- [x] 5.1 Modify `internal/modules/permissions/routes.go` — remove GIN routes for `POST ""` and `DELETE "/:id"`.
- [x] 5.2 Modify `internal/modules/permissions/dto.go` — adjust DTO fields if necessary, or keep them but update handler validation.
- [x] 5.3 Modify `internal/modules/permissions/service.go`:
  - In `Update`, load permission by `publicID`.
  - Validate that `req.Slug != permission.Slug || req.Module != permission.Module || req.Action != permission.Action` evaluates to false. If any structural field is changed, reject with `apperror.ErrForbidden`.
  - Only assign `Name` and `Description` before committing the GORM update.
- [x] 5.4 Modify `internal/modules/permissions/handler.go` — clean up or deactivate unused creation and deletion methods.
- [x] 5.5 Clean up/disable old handler/service tests for Create/Delete or update them to verify they are gone/return expected errors.

## Phase 6: Full Integration and Verification

- [x] 6.1 Run test suite: `go test ./...`.
- [x] 6.2 Fix compilation errors due to `"index"` -> `"list"` renaming.
- [x] 6.3 Run build: `go build ./...` to verify all imports are clean and there are no circular dependencies.
- [x] 6.4 Triangulate: Verify `SyncPermissions` preserves custom Name/Description in database when seeded/bootstrap rerun.
- [x] 6.5 Run final verification check.

---

## Acceptance Mapping

| Acceptance criterion | Tasks |
|---|---|
| Creation of central `permissions/constants.go` transversal package | 2.1 |
| RequirePermission routes use `permissions.Format` with global constants | 3.1, 3.2, 4.2, 4.3, 4.4 |
| Refactor `"index"` to `"list"` globally | 4.1, 4.2, 4.3, 4.4, 4.5, 4.6 |
| Auto-sync & Upsert system permissions on server start | 3.3, 3.4, 3.5 |
| POST/DELETE permissions endpoints return route not supported (404) | 1.2, 5.1 |
| PUT permissions only updates Name/Description and rejects structural alterations | 1.1, 5.3 |
| Compile and GREEN test suite | 6.1, 6.2, 6.3 |
