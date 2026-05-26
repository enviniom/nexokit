# SDD Tasks: Company Onboarding, Auto-Admin Sync and Role Protections

## Phase 1: System Permissions Auto-Sync Enhancement

- [x] **1.1 Repository Extension for Auto-Assign**
  - Implemented `AutoAssignToAdmins(permissionID uint) error` in `permissions.Repository` using GORM `Exec`:
    `INSERT INTO role_permissions (role_id, permission_id) SELECT id, ? FROM roles WHERE slug = 'admin' ON CONFLICT DO NOTHING;`

- [x] **1.2 Service Integration**
  - Modified `SyncPermissions` in `permissions.Service` (`internal/modules/permissions/service.go`) to automatically call `AutoAssignToAdmins(newPerm.ID)` immediately after a new system permission is successfully created.

---

## Phase 2: Roles Protection (Admin Permission Lock)

- [x] **2.1 Service Revocation Lock**
  - Implemented admin-role protection in `AssignPermissions` in `internal/modules/roles/service.go`.
  - Reconciled implementation note: the landed code uses a stricter lock than the original task wording and rejects direct permission assignment changes for roles with slug `admin`, preventing revocation by construction.

---

## Phase 3: Onboarding Module Core Implementation

- [x] **3.1 Onboarding Package DTOs**
  - Created `internal/modules/onboarding/dto.go` defining `OnboardCompanyRequest` validation rules and response DTO structure.

- [x] **3.2 Transactional Service Implementation**
  - Created `internal/modules/onboarding/service.go`.
  - Wrapped the onboarding process inside GORM's `db.Transaction` callback to enforce full rollback:
    1. Parse and validate inputs.
    2. Check system-wide uniqueness of `slug`, `domain`, and `subdomain`.
    3. Check user email uniqueness.
    4. Save Company.
    5. Save tenant `admin` and `user` roles (`company_id = company.ID`, `is_system = true`).
    6. Fetch all system permissions from the DB.
    7. Map all permissions to the tenant `admin` role.
    8. Map base permissions (`users.view`, `roles.view`) to the tenant `user` role.
    9. Hash the initial admin's password and save the admin user (`company_id = company.ID`).

- [x] **3.3 Handler & Routing**
  - Created `internal/modules/onboarding/handler.go` with `POST /api/v1/onboarding/companies`.
  - Created `internal/modules/onboarding/routes.go` to mount onboarding routes protected with `RequireRole("root")`.

---

## Phase 4: Dependency Injection & Routing Upgrades

- [x] **4.1 Container Wiring**
  - Modified `internal/app/container.go` to instantiate and configure `onboarding.Service` and `onboarding.Handler`, mounting them on the global protected router.

- [x] **4.2 Remove Direct Company Creation Route**
  - Modified `internal/modules/companies/routes.go` to remove `companies.POST("", requireRole("root"), handler.Create)`.
  - Updated `internal/modules/companies/handler_test.go` to verify direct company creation returns `404` for all roles.

---

## Phase 5: Verification & Tests

- [x] **5.1 Onboarding Integration & Unit Tests**
  - Added `internal/modules/onboarding/service_test.go` and `internal/modules/onboarding/handler_test.go` covering validation/routing, success path, duplicate slug rollback, duplicate email rollback, role permission assignments, and root-only endpoint access.

- [x] **5.2 Admin Permission Revocation Tests**
  - Added/updated tests in `internal/modules/roles/service_test.go` verifying attempts to modify tenant `admin` role permissions are blocked with `403` semantics (`apperror.ErrForbidden`).

- [x] **5.3 Auto-Sync Permission Tests**
  - Added tests in `internal/modules/permissions/service_test.go` verifying permission synchronization calls automatic assignment for newly created permissions.

- [x] **5.4 Project Execution check**
  - Ran `go test ./...` successfully on 2026-05-26.
