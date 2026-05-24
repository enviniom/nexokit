# Specifications: Company Onboarding, Auto-Admin Sync and Role Protections

## System Behavior & Constraints

### Onboarding Endpoint
- The system MUST expose an endpoint at `POST /api/v1/onboarding/companies`.
- The system MUST protect this endpoint so it is ONLY accessible by users with the system `root` role. Requests from any non-root user (including tenant admins) MUST fail with status `403 Forbidden` or `401 Unauthorized` as appropriate.
- The system MUST validate that required parameters (`name`, `slug`, `admin_name`, `admin_email`, `admin_password`) are provided and conform to minimum length/format rules.
- The system MUST ensure that company `slug`, `domain` (if provided), and `subdomain` (if provided) are globally unique.
- The system MUST ensure that the admin user's `email` is globally unique.
- The system MUST execute the entire onboarding process inside a single database transaction. If any step fails, the system MUST rollback the entire transaction cleanly.
- The onboarding process MUST create:
  1. The `Company` record.
  2. The `admin` role for this company (`slug = "admin"`, `company_id = company.ID`, `is_system = true`).
  3. The `user` role for this company (`slug = "user"`, `company_id = company.ID`, `is_system = true`).
  4. The initial `User` record with the `admin` role and `company_id = company.ID`.
- The system MUST map all currently registered system permissions to the tenant `admin` role in `role_permissions` join table.
- The system MUST map a predefined subset of base permissions to the tenant `user` role.
- The system MUST NOT create or modify the global system `root` role.
- The system MUST NOT allow direct company creation; the route `POST /api/v1/companies` MUST be deactivated, forcing all company signups to use the Onboarding flow.

### Automatic Permission Synchronization
- GIVEN that new system permissions are registered in code
- WHEN the application starts up and synchronizes permissions
- THEN the system MUST automatically assign the new permissions to all existing tenant `admin` roles in the database.

### Admin Role Permission Protections
- The system MUST NOT allow anyone (including `root`) to remove/revoke permissions from a tenant `admin` role.
- WHEN a request is made to assign permissions to an `admin` role
- THEN the system MUST verify that all currently assigned permissions are included in the new payload. If any assigned permission is missing, the request MUST fail with status `403 Forbidden`.

---

## Scenarios

### Scenario 1: Successful Company Onboarding by Root
- **GIVEN** a request made by an authenticated `root` user
- **AND** a valid payload with `name = "Acme Corp"`, `slug = "acme"`, `admin_name = "Jane Doe"`, `admin_email = "jane@acme.com"`, and `admin_password = "SecurePassword123"`
- **WHEN** the `POST /api/v1/onboarding/companies` endpoint is called
- **THEN** the system MUST return status `201 Created`
- **AND** a company with slug `acme` MUST exist in the database
- **AND** the company MUST have tenant roles `admin` and `user`
- **AND** the `admin` role MUST have all catalog permissions mapped to it
- **AND** the user `jane@acme.com` MUST exist with the tenant `admin` role
- **AND** the database transaction MUST commit successfully.

### Scenario 2: Non-Root Attempt to Onboard Fails
- **GIVEN** an authenticated user who does not have the `root` role (e.g. a tenant admin or regular user)
- **WHEN** they call `POST /api/v1/onboarding/companies`
- **THEN** the system MUST return status `403 Forbidden`
- **AND** no company or tenant roles MUST be created.

### Scenario 3: Onboarding Fails Due to Duplicate Slug (Rollback)
- **GIVEN** an authenticated `root` user
- **AND** an existing company in the database with slug `acme`
- **WHEN** a new onboarding request with slug `acme` is submitted
- **THEN** the system MUST return status `409 Conflict` (or validation error)
- **AND** the transaction MUST rollback, leaving no orphan records.

### Scenario 4: Onboarding Fails Due to Duplicate Email (Rollback)
- **GIVEN** an authenticated `root` user
- **AND** an existing user with email `jane@acme.com`
- **WHEN** a new onboarding request is submitted with `admin_email = "jane@acme.com"` and `slug = "new-company"`
- **THEN** the system MUST return status `409 Conflict`
- **AND** the transaction MUST rollback cleanly
- **AND** the company `new-company` MUST NOT exist in the database.

### Scenario 5: Admin Role Permissions Isolation
- **GIVEN** an onboarded company `acme` and its admin user `jane@acme.com`
- **WHEN** the admin user performs an operation (e.g. creating a user)
- **THEN** the system MUST isolate their operations to the `acme` tenant using GORM query scopes
- **AND** the admin user MUST NOT be able to view, edit, or interact with resources from other tenants.

### Scenario 6: Direct Company Creation is Blocked
- **GIVEN** any user (including `root`) trying to create a company directly
- **WHEN** they call `POST /api/v1/companies`
- **THEN** the system MUST return status `404 Not Found` (since the route is completely removed).

### Scenario 7: Revoking Admin Permissions is Forbidden
- **GIVEN** an existing tenant `admin` role with permissions `["users.view", "users.create"]`
- **WHEN** a request is made to assign permissions `["users.view"]` to this role (omitting `users.create`)
- **THEN** the system MUST return status `403 Forbidden`
- **AND** the role's permissions MUST remain unchanged.

### Scenario 8: Auto-Sync New Permissions
- **GIVEN** a new system permission `billing.invoice` registered in code
- **AND** an existing company `acme` with its `admin` role in the database
- **WHEN** the application starts up and triggers permission sync
- **THEN** the `admin` role for `acme` MUST be automatically assigned the `billing.invoice` permission.
