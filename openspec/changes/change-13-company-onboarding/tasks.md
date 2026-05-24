# SDD Tasks: Company Onboarding, Auto-Admin Sync and Role Protections

## Phase 1: System Permissions Auto-Sync Enhancement

- [ ] **1.1 Repository Extension for Auto-Assign**
  - Implement `AutoAssignToAdmins(permissionID uint) error` in `permissions.Repository` using raw GORM SQL `Exec` execution:
    `INSERT INTO role_permissions (role_id, permission_id) SELECT id, ? FROM roles WHERE slug = 'admin' ON CONFLICT DO NOTHING;`

- [ ] **1.2 Service Integration**
  - Modify `SyncPermissions` in `permissions.Service` (`internal/modules/permissions/service.go`) to automatically call `AutoAssignToAdmins(newPerm.ID)` immediately after a new system permission is successfully created.

---

## Phase 2: Roles Protection (Admin Permission Lock)

- [ ] **2.1 Service Revocation Lock**
  - Modify `AssignPermissions` method in `internal/modules/roles/service.go`.
  - Add validation: if `role.Slug == roles.AdminRoleSlug`, loop through currently assigned `role.Permissions` and verify that each is present in the `selected` map.
  - If any existing permission is missing, return `apperror.ErrForbidden`.

---

## Phase 3: Onboarding Module Core Implementation

- [ ] **3.1 Onboarding Package DTOs**
  - Create `internal/modules/onboarding/dto.go` defining the `OnboardCompanyRequest` validation rules and response DTO structure.

- [ ] **3.2 Transactional Service Implementation**
  - Create `internal/modules/onboarding/service.go`.
  - Wrap the onboarding process inside GORM's `db.Transaction` callback to enforce full rollback:
    1. Parse and validate inputs.
    2. Check system-wide uniqueness of `slug`, `domain`, and `subdomain`.
    3. Check user email uniqueness.
    4. Save Company.
    5. Save tenant `admin` and `user` roles (`company_id = company.ID`, `is_system = true`).
    6. Fetch all system permissions from the DB.
    7. Map all permissions to the tenant `admin` role.
    8. Map base permissions (e.g. `users.view`, `roles.view`) to the tenant `user` role.
    9. Hash the initial admin's password and save the admin user (`company_id = company.ID`).

- [ ] **3.3 Handler & Routing**
  - Create `internal/modules/onboarding/handler.go` with `POST /api/v1/onboarding/companies`. Protect it using `RequireRole("root")`.
  - Create `internal/modules/onboarding/routes.go` to mount onboarding routes on the v1 global protected router group.

---

## Phase 4: Dependency Injection & Routing Upgrades

- [ ] **4.1 Container Wiring**
  - Modify `internal/app/container.go` to instantiate and configure `onboarding.Service` and `onboarding.Handler`, mounting them on the global protected router.

- [ ] **4.2 Remove Direct Company Creation Route**
  - Modify `internal/modules/companies/routes.go` to remove `companies.POST("", requireRole("root"), handler.Create)`.
  - Update `internal/modules/companies/handler_test.go` or related tests to verify direct company creation is no longer allowed.

---

## Phase 5: Verification & Tests

- [ ] **5.1 Onboarding Integration & Unit Tests**
  - Write `internal/modules/onboarding/service_test.go` and `internal/modules/onboarding/handler_test.go` (table-driven, testing validations, success path, GORM transaction rollback on duplicate email, duplicate company slug, and role permission assignments).
  - Verify that only `root` users can call the endpoint (returns `403` for non-root).
  - Verify isolation boundary of the newly created tenant admin role.

- [ ] **5.2 Admin Permission Revocation Tests**
  - Create tests in `internal/modules/roles/service_test.go` verifying that attempts to revoke permissions from the tenant `admin` role are blocked and return `403`.

- [ ] **5.3 Auto-Sync Permission Tests**
  - Create tests in `internal/modules/permissions/service_test.go` verifying that executing permission synchronization automatically maps newly created permissions to all existing tenant `admin` roles in the database.

- [ ] **5.4 Project Execution check**
  - Run `go test ./...` to verify all tests pass completely.
