# SDD Explore: Company Onboarding with Roles and Initial Permissions

## Codebase Analysis & Discoveries

### 1. Multi-Tenant Architecture & Context Isolation
- **Tenant Context (`TenantContext`)**: Scopes requests using `CompanyID` and `CompanySlug`. The system has a system-wide `root` scope (`IsRootScope: true`).
- **GORM Tenant Query Filter (`ApplyTenantScope`)**: Automatically appends `WHERE company_id = ?` to queries when `IsRootScope` is false.
- **Root Bypass**: The `root` role bypasses permissions check at the middleware level (`isRoot` check in `RequirePermission` or `RequireRole`). Non-root requests go through normal RBAC validation.

### 2. Current State of Database Seeds & Migrations
- **Seeded Roles**: Only the global `root` role is seeded (`company_id = NULL` and `is_system: true`).
- **Tenant Roles**: `admin` and `user` are tenant-level roles (they must be created with a valid `company_id` referencing the company). Global seeds no longer register them.
- **Permissions**: Defined globally in the `permissions` table, with no `company_id`. They represent static access keys (e.g. `users.create`, `roles.update`, etc.).
- **Permissions Join**: Done in `role_permissions` join table mapping `role_id` to `permission_id`.

### 3. Requirements for Company Onboarding Flow

1. **Input Fields**:
   - `name` (Company Name)
   - `slug` (Company Slug - unique identifier used in tenant resolution)
   - `domain` (Nullable, unique domain)
   - `subdomain` (Nullable, unique subdomain)
   - `admin_name` (Name of the initial Admin user)
   - `admin_email` (Unique email of the initial Admin user)
   - `admin_password` (Optional password depending on invitation scheme; since we want to create them ready-to-login, we will allow direct creation with password or generate a random one if omitted).

2. **Action Sequence (within one single DB Transaction)**:
   - Validate input parameters.
   - Check unique constraint duplicates for company slug, domain, and subdomain across the system.
   - Check email uniqueness for the admin user.
   - Create `Company` record.
   - Create `Role` records for the new company:
     * **`admin` role**: `Name: "Admin", Slug: "admin", CompanyID: company.ID, IsSystem: true`.
     * **`user` role**: `Name: "User", Slug: "user", CompanyID: company.ID, IsSystem: true`.
   - Resolve all registered permissions in the system.
   - Map **all available permissions** to the newly created `admin` role (giving it super-admin capabilities inside its tenant scope).
   - Map a standard subset of permissions (e.g., `users.view`, `roles.view`) to the newly created `user` role.
   - Create the initial `User` record:
     * Name: `admin_name`, Email: `admin_email`, Password Hash, RoleID: `admin.ID` (newly created tenant admin role), CompanyID: `company.ID`.
   - **Rollback**: If any of the steps fail, cleanly rollback the database transaction to prevent leaving orphan companies or partially configured roles.

3. **System Permission Catalog**:
   - The onboarding needs a source of truth for available permissions.
   - Current system registers permissions via `permissions.Register(slug)` inside middleware files (like `RequirePermission`), and syncs them at startup using `c.SyncPermissions()` in `internal/app/container.go` which reads all registered permissions in `internal/platform/permissions/permissions.go` memory map and persists them to the DB.
   - We need endpoints to:
     * List all permissions: `GET /api/v1/permissions/catalog`
     * List unassigned permissions: `GET /api/v1/permissions/unassigned` (permissions not mapped to any role in the system).

## Design Options & Choices

### 1. Permission Source of Truth
We will leverage the existing memory-based registry in `internal/platform/permissions/permissions.go` where middlewares automatically register permissions. At bootstrap, these are synchronized into the database. Thus, the database `permissions` table serves as the runtime source of truth.

### 2. Transaction Management
To ensure a reliable rollback, GORM's `.Transaction(...)` will wrap the entire onboarding workflow. If any error occurs (duplicate email, validation failure, index crash), GORM automatically executes a database rollback.

### 3. First Admin Strategy
We will implement a direct-provisioning strategy: the admin's name, email, and password are provided in the onboarding payload. This allows the company to be immediately active and operational.

## Architectural Trade-Offs

- **Standalone Module vs Extending Companies**: We will introduce a dedicated `onboarding` module rather than cluttering the `companies` module. Onboarding is a high-level orchestration concerns that touches `companies`, `roles`, `users`, and `permissions` simultaneously. Placing it in its own module keeps our modular architecture clean and follows single-responsibility principles.
