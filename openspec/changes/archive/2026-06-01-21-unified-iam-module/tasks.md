# Tasks: Unified IAM Module

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 3000–4000 (new IAM module + tests + wiring) |
| 800-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 → PR 4 → PR 5 |
| Delivery strategy | ask-always |
| Chain strategy | stacked-to-main (resolved) |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
800-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | IAM core: models, DTOs, errors, constants, contracts, and demand-driven query boundary | PR 1 | Base branch; tests included |
| 2 | IAM user slices (8 endpoints) + user sub-container/routes | PR 2 | Targets PR 1 branch |
| 3 | IAM role slices (8 endpoints) + role sub-container/routes | PR 3 | Targets PR 2 branch |
| 4 | IAM permission slices (3 endpoints) + internal slices (5) + permission sub-container/routes | PR 4 | Targets PR 3 branch |
| 5 | App wiring: container swap, RegisterModules, middleware adapters, integration tests | PR 5 | Targets PR 4 branch; final integration |

## Phase 1: Foundation / Core IAM

- [x] 1.1 Create `internal/modules/iam/core/model.go` with partial models: IAMUser, IAMRole, IAMPermission, IAMCompany, IAMRolePermission
- [x] 1.2 Create `internal/modules/iam/core/dto.go` with all request/response DTOs mirroring legacy payloads exactly
- [x] 1.3 Create `internal/modules/iam/core/error.go` with IAM domain errors
- [x] 1.4 Create `internal/modules/iam/core/constants.go` with ModuleIAM, reserved slugs, system role constants
- [x] 1.5 Create `internal/modules/iam/core/contracts.go` with ResolveAuthUser, Resolve, SyncPermissions, ResolveRoleBySlug, ListAllPermissions interfaces
- [x] 1.6 Establish `internal/modules/iam/queries/` as a demand-driven reuse boundary (one file per query; only queries reused by multiple slices belong here)
- [x] 1.7 Update PR1 query tests/artifacts to match demand-driven rule (remove speculative shared-query tests until reusable queries are promoted)
- [x] 1.8 Create `internal/modules/iam/container.go` root container with Users, Roles, Permissions sub-container fields + exported contracts
- [x] 1.9 Create `internal/modules/iam/routes.go` delegating to entity route files

## Phase 2: User Slices

- [x] 2.1 Create `internal/modules/iam/users/container.go` and `internal/modules/iam/users/routes.go`
- [x] 2.2 Create `internal/modules/iam/users/list_users/` handler, service, repository + tests; tenant-scoped list
- [x] 2.3 Create `internal/modules/iam/users/create_user/` handler, service, repository + tests; root-scope logic
- [x] 2.4 Create `internal/modules/iam/users/view_user/` handler, service, repository + tests; tenant-scoped view
- [x] 2.5 Create `internal/modules/iam/users/update_user/` handler, service, repository + tests; root protection
- [x] 2.6 Create `internal/modules/iam/users/delete_user/` handler, service, repository + tests; soft-delete, HTTP 204
- [x] 2.7 Create `internal/modules/iam/users/change_user_password/` handler, service, repository + tests; self-change rule
- [x] 2.8 Create `internal/modules/iam/users/assign_role_to_user/` handler, service, repository + tests; uses AssignmentRoleReader
- [x] 2.9 Create `internal/modules/iam/users/toggle_user_status/` handler, service, repository + tests

## Phase 3: Role Slices

- [x] 3.1 Create `internal/modules/iam/roles/container.go` and `internal/modules/iam/roles/routes.go`
- [x] 3.2 Create `internal/modules/iam/roles/list_roles/` handler, service, repository + tests
- [x] 3.3 Create `internal/modules/iam/roles/list_selectable_roles/` handler, service, repository + tests; selection normalization
- [x] 3.4 Create `internal/modules/iam/roles/view_role/` handler, service, repository + tests
- [x] 3.5 Create `internal/modules/iam/roles/create_role/` handler, service, repository + tests; reserved slug check
- [x] 3.6 Create `internal/modules/iam/roles/update_role/` handler, service, repository + tests; system role protection
- [x] 3.7 Create `internal/modules/iam/roles/delete_role/` handler, service, repository + tests; assigned-user guard
- [x] 3.8 Create `internal/modules/iam/roles/view_role_permission_catalog/` handler, service, repository + tests
- [x] 3.9 Create `internal/modules/iam/roles/assign_permissions_to_role/` handler, service, repository + tests; cache invalidation for role members

## Phase 4: Permission Slices + Internal Slices

- [x] 4.1 Create `internal/modules/iam/permissions/container.go` and `internal/modules/iam/permissions/routes.go`
- [x] 4.2 Create `internal/modules/iam/permissions/list_permissions/` handler, service, repository + tests; grouped by module, sorted by display_order
- [x] 4.3 Create `internal/modules/iam/permissions/view_permission/` handler, service, repository + tests
- [x] 4.4 Create `internal/modules/iam/permissions/update_permission/` handler, service, repository + tests; system immutable
- [x] 4.5 Create `internal/modules/iam/internal/resolve_auth_user/` service, repository + tests; GetAuthUser for auth middleware
- [x] 4.6 Create `internal/modules/iam/internal/resolve_user_permissions/` service, repository + tests; cache-backed, 5-min TTL
- [x] 4.7 Create `internal/modules/iam/internal/sync_permissions/` service, repository + tests; idempotent, auto-assign to admin roles
- [x] 4.8 Create `internal/modules/iam/internal/resolve_role_by_slug/` service, repository + tests
- [x] 4.9 Create `internal/modules/iam/internal/list_all_permissions/` service, repository + tests; PermissionCatalogReader

## Phase 5: App Wiring + Integration

- [x] 5.1 Modify `internal/app/container.go`: replace usersHandler, rolesHandler, permissionsContainer with `IAM *iam.Container`
- [x] 5.2 Modify `internal/app/container.go`: update `RegisterModules` to call `iam.Register(globalProtected, c.IAM, tenantProtected, middleware.RequirePermission, middleware.RequireRole)`; remove legacy mounts
- [x] 5.3 Modify `internal/app/container.go`: update `userLookup` adapter to delegate to `c.IAM.ResolveAuthUser`
- [x] 5.4 Modify `internal/app/container.go`: update `roleResolverAdapter` to delegate to `c.IAM.ResolveRoleBySlug`
- [x] 5.5 Modify `internal/app/container.go`: update `SyncPermissions` to delegate to `c.IAM.SyncPermissions`
- [x] 5.6 Write integration test: verify all 19 IAM endpoints respond at expected `/api/v1/*` paths
- [x] 5.7 Write integration test: verify legacy routes return HTTP 404 (not mounted)
- [x] 5.8 Write integration test: verify middleware auth/authz work with IAM adapters
- [x] 5.9 Verify `go build ./internal/modules/users/... ./internal/modules/roles/... ./internal/modules/permissions/...` succeeds (legacy compile)
- [x] 5.10 Verify `go test ./...` passes; verify zero cross-module imports from IAM via `go list`

## Non-Goals (Explicit)

- NO deletion of legacy modules (`users/`, `roles/`, `permissions/`)
- NO destructive file moves from legacy modules
- NO route renames or API behavior changes
- NO migration file changes
- NO logic refactoring during copy — reproduce behavior exactly
