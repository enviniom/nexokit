# Tasks: Tenant-scoped roles with global root

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 550–650 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (migration + model + repository) → PR 2 (service + handler + DTO) → PR 3 (seeds + tests) |
| Delivery strategy | auto-chain |
| Chain strategy | feature-branch-chain |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

> **Note**: Estimate (550–650 lines) exceeds the user's 800-line stop budget? No — it is within the 800-line budget but above the standard 400-line review budget. Chained PRs are recommended to keep each review slice under 400 lines.

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Migration + model field + repository tenant context | PR 1 | Base = feature/change-10-tenant-roles branch; includes migration SQL review |
| 2 | Service tenant isolation + handler wiring + DTO company_id | PR 2 | Base = PR 1 branch; depends on PR 1; reserved slug guard expansion |
| 3 | Seed changes + all test updates + integration verification | PR 3 | Base = PR 2 branch; depends on PR 2; seed-only-root + test fakes rewrite |

## Phase 1: Migration + Model (Foundation)

- [x] 1.1 Create `migrations/20260520000000_tenant_roles.sql` adding nullable `company_id` to `roles`, FK to `companies(id)`, composite unique indexes `(slug, company_id)` and `(name, company_id)`, partial unique index `WHERE company_id IS NULL` for global uniqueness, and down migration restoring old indexes.
- [x] 1.2 Modify `internal/modules/roles/model.go` — add `CompanyID *uint` field with GORM tag; remove global `uniqueIndex` from `Name` and `Slug` (uniqueness now DB-managed via migration indexes).
- [x] 1.3 Add `ReservedSlugs` constant slice `[]string{"root", "admin", "user"}` to `internal/modules/roles/model.go` alongside existing `RootRoleSlug`, `AdminRoleSlug`, `UserRoleSlug` constants.

## Phase 2: Repository (Tenant-Aware Persistence)

- [x] 2.1 Update `internal/modules/roles/repository.go` interface — add `tenant.TenantContext` as first parameter to `List`, `Count`, `GetByPublicID`, `GetByName`, `GetBySlug`, `Delete`.
- [x] 2.2 Implement `List(tc, page, perPage)` — use `tenant.ApplyTenantScope(db, tc)` before query; preload permissions.
- [x] 2.3 Implement `Count(tc)` — use `tenant.ApplyTenantScope(db, tc)` before count.
- [x] 2.4 Implement `GetByPublicID(tc, publicID)` — add tenant scope to query; preload permissions.
- [x] 2.5 Implement `GetByName(tc, name)` and `GetBySlug(tc, slug)` — add tenant scope to queries; preload permissions.
- [x] 2.6 Implement `Delete(tc, publicID)` — add tenant scope to delete query.
- [x] 2.7 Add `ExistsByName(tc, name) (bool, error)` helper — scoped existence check for uniqueness validation.
- [x] 2.8 Add `ExistsBySlug(tc, slug) (bool, error)` helper — scoped existence check for uniqueness validation.

## Phase 3: Service (Business Logic + Tenant Isolation)

- [x] 3.1 Update `internal/modules/roles/service.go` interface — add `tenant.TenantContext` to `List`, `GetByPublicID`, `Create`, `Update`, `Delete` signatures.
- [x] 3.2 Replace `isRootRole(name, slug)` with `isReservedIdentity(name, slug)` checking against all reserved slugs (`root`, `admin`, `user`).
- [x] 3.3 Update `List(tc, page, perPage)` — pass tenant context to repository; map `CompanyID` in response DTOs.
- [x] 3.4 Update `GetByPublicID(tc, publicID)` — pass tenant context; return `ErrNotFound` if role exists but belongs to different tenant.
- [x] 3.5 Update `Create(tc, req)` — set `role.CompanyID` from `tc.CompanyID` when not root scope; use `ExistsByName`/`ExistsBySlug` for scoped uniqueness; reject reserved slugs via `isReservedIdentity`.
- [x] 3.6 Update `Update(tc, publicID, req)` — pass tenant context to lookup; prevent cross-tenant update; forbid reserved slug/name changes; use scoped uniqueness checks.
- [x] 3.7 Update `Delete(tc, publicID)` — pass tenant context; prevent cross-tenant delete; forbid reserved role deletion.
- [x] 3.8 Update `GetPermissionCatalog` and `AssignPermissions` — add tenant context to role lookup; verify tenant ownership before permission operations.

## Phase 4: Handler + DTO (HTTP Wiring)

- [x] 4.1 Modify `internal/modules/roles/handler.go` — add `tenantContext(c *gin.Context) tenant.TenantContext` method mirroring users handler pattern (`tenant.FromGin` fallback to `tenant.NewRoot()`).
- [x] 4.2 Update `List` handler — extract tenant context, pass to service; use `response.HandleError` pattern.
- [x] 4.3 Update `GetByPublicID`, `Create`, `Update`, `Delete` handlers — extract and pass tenant context to service calls.
- [x] 4.4 Update `GetPermissionCatalog` and `AssignPermissions` handlers — pass tenant context to service.
- [x] 4.5 Modify `internal/modules/roles/dto.go` — add `CompanyID *uint json:"company_id,omitempty"` to `RoleResponse`.

## Phase 5: Seeds (Data Changes)

- [x] 5.1 Modify `seeds/roles.go` — seed only root role (global, `company_id NULL`); remove admin and user from seed list.
- [x] 5.2 Modify `seeds/role_permissions.go` — remove `adminPermissionSlugs()` and `userPermissionSlugs()` assignments; keep only root with `allSystemPermissionSlugs()` (root still gets `"*"` via middleware, but seed rows are removed per design decision).
- [x] 5.3 Update `seeds/roles_test.go` — assert only 1 role (root) exists after seed; update idempotency test.
- [x] 5.4 Update `seeds/role_permissions_test.go` (if exists) — verify only root has permission rows, or no role_permissions rows needed for root.

## Phase 6: Testing (Verification)

- [x] 6.1 Update `internal/modules/roles/service_test.go` fake repository — add `tenant.TenantContext` to all interface methods; implement scoped filtering in fake.
- [x] 6.2 Add service tests: tenant-scoped list returns only roles for current company.
- [x] 6.3 Add service tests: root scope (`IsRootScope=true`) lists all roles globally.
- [x] 6.4 Add service tests: `Create` with tenant context sets `CompanyID`; root scope creates global role (`CompanyID=nil`).
- [x] 6.5 Add service tests: `isReservedIdentity` rejects `root`, `admin`, `user` by name and slug (case-insensitive).
- [x] 6.6 Add service tests: cross-tenant `GetByPublicID` returns `ErrNotFound`.
- [x] 6.7 Add service tests: cross-tenant `Update`/`Delete` returns `ErrNotFound` or `ErrForbidden`.
- [x] 6.8 Update `internal/modules/roles/handler_test.go` fake service — add `tenant.TenantContext` to all interface methods.
- [x] 6.9 Add handler tests: verify `tenant.SetGin` context is passed through to service calls using `tenant.NewScoped()` and `tenant.NewRoot()`.
- [x] 6.10 Run `go test ./...` — verify all existing and new tests pass.
- [x] 6.11 Run `go build ./...` — verify compilation.
