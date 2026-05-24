# Proposal: Company Onboarding with Auto-Admin Sync and Role Protections

## Goal Description
Create a highly robust tenant onboarding process. When a new company is registered, it will be automatically provisioned with its own scoped `admin` and `user` roles. The tenant `admin` role will be assigned all currently registered system permissions in a single database transaction.

To ensure maximum security, scalability, and ease of maintenance in a multi-tenant SaaS environment, we will enforce three core platform rules:
1. **Root-Only Onboarding Protection**: Company onboarding is restricted strictly to the global system operator (`root` role). By protecting the onboarding endpoint purely through the `RequireRole("root")` middleware instead of a permission slug, we completely prevent tenant administrators from gaining access to this capability.
2. **Automatic Permission Synchronization for Admins**: Whenever a new system permission is registered in code and synced at startup, the system will automatically assign it to all existing tenant `admin` roles in the database. This guarantees that administrators automatically gain access to new system features.
3. **Admin Permission Lock**: The system will protect the `admin` role from having any of its permissions revoked. Any attempt to update the permissions of an `admin` role that omits currently assigned permissions will be blocked at the service layer.

---

## User Review Required

> [!IMPORTANT]
> **Transactional Rollback Guard**:
> Onboarding involves creating a company, seeding roles, mapping permissions, and creating a user. A database transaction ensures that if any failure occurs (e.g. email already exists, subdomain duplicate), no partial state is written to the database.

> [!NOTE]
> **Root-Only Safeguard**:
> The `POST /api/v1/onboarding/companies` endpoint is placed under global root protection using the existing `RequireRole("root")` middleware. Only the platform operator can bootstrap new tenants.

---

## Proposed Changes

### 1. New Module: Onboarding

#### [NEW] [dto.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/onboarding/dto.go)
- Define `OnboardCompanyRequest` carrying company parameters (`name`, `slug`, `domain`, `subdomain`) and admin user parameters (`admin_name`, `admin_email`, `admin_password`).
- Implement `Validate()` for input parameters (unique constraint validation will be performed by the service layer).

#### [NEW] [service.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/onboarding/service.go)
- Create `Service` interface and implement onboarding logic.
- Run onboarding in a GORM transaction:
  1. Create company.
  2. Create `admin` and `user` tenant roles (`company_id = company.ID`, `is_system = true`).
  3. Load all permissions from the database.
  4. Associate all permissions with the new `admin` role.
  5. Associate a standard base subset of permissions (e.g. `users.view`, `roles.view`) with the `user` role.
  6. Create the first user, assign them to the `admin` role, and set their `company_id = company.ID`.
- Handle rollbacks and return appropriate errors (e.g. `ErrDuplicateSlug`, `ErrDuplicateEmail`).

#### [NEW] [handler.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/onboarding/handler.go)
- Implement `POST /api/v1/onboarding/companies` endpoint.
- Accept payloads and execute the service onboarding logic.

#### [NEW] [routes.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/onboarding/routes.go)
- Mount onboarding handler. This endpoint will be protected using `RequireRole("root")`.

---

### 2. Permissions Sync Extension (Admins Auto-Assignment)

#### [MODIFY] [repository.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/permissions/repository.go)
- Implement `AutoAssignToAdmins(permissionID uint) error` which executes a database-level `INSERT INTO role_permissions ... SELECT id, ? FROM roles WHERE slug = 'admin' ON CONFLICT DO NOTHING;` statement.

#### [MODIFY] [service.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/permissions/service.go)
- Update `SyncPermissions` to invoke the repository's `AutoAssignToAdmins` method whenever a new system permission is synchronized.

---

### 3. Roles Protection (Admin Permission Lock)

#### [MODIFY] [service.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/roles/service.go)
- Modify `AssignPermissions` method: if the target role's slug is `admin`, verify that all currently assigned permissions are present in the incoming payload. If any assigned permission is missing (revoked), reject the request with a `Forbidden` error.

---

### 4. Companies Module Updates

#### [MODIFY] [routes.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/modules/companies/routes.go)
- Remove `POST /companies` route handler setup. This enforces that companies can ONLY be registered through the Onboarding process.
- Keep the `GET`, `PUT`, `DELETE` routes for management by `root` (e.g. updating company settings, deleting, listing) but direct creation is deprecated/removed.

---

### 5. Dependency Injection

#### [MODIFY] [container.go](file:///home/enviniom/Proyectos/Go/nexokit/internal/app/container.go)
- Wire up the new `onboarding` module dependencies and mount the routes.

---

## Verification Plan

### Automated Tests
- Run `go test ./...` to verify no regressions in existing modules.
- Create unit and integration tests inside `internal/modules/onboarding/service_test.go` and `internal/modules/onboarding/handler_test.go` covering:
  - Happy path company onboarding by root user.
  - Verification that non-root users cannot access the onboarding endpoint (returns `403`).
  - Rollback on email conflict.
  - Rollback on company slug/subdomain conflict.
  - Verification that the generated `admin` role contains all permissions.
- Create unit/integration tests in `internal/modules/roles/service_test.go` verifying that revoking permissions from the `admin` role is forbidden.
- Create unit/integration tests in `internal/modules/permissions/service_test.go` verifying that new permissions are automatically assigned to existing `admin` roles.

### Manual Verification
- Execute manual POST request to `/api/v1/onboarding/companies` as root to ensure the company, roles, and admin are created correctly.
- Verify that trying to update the `admin` role to remove permissions fails with `403`.
