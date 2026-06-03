# Delta for App Orchestration

## MODIFIED Requirements

### Requirement: RegisterModules mounts IAM only

The system SHALL mount IAM routes via a single `iam.Register(globalProtected, c.IAM, tenantProtected, middleware.RequirePermission, middleware.RequireRole)` call. Legacy module registrations for `users`, `roles`, and `permissions` SHALL NOT exist in `RegisterModules`. The legacy module directories `internal/modules/users/`, `internal/modules/roles/`, and `internal/modules/permissions/` SHALL NOT exist on disk.
(Previously: Legacy modules remained compilable but unreachable at runtime; now legacy directories are deleted entirely)

#### Scenario: IAM routes mounted

- GIVEN `RegisterModules` is called during bootstrap
- WHEN the router is inspected
- THEN all 19 IAM endpoints respond at their expected `/api/v1/*` paths

#### Scenario: Legacy routes not mounted

- GIVEN `RegisterModules` is called
- WHEN a legacy route path is requested
- THEN the response returns HTTP 404 (route not registered)

#### Scenario: No legacy module directories exist

- GIVEN the codebase after legacy removal
- WHEN `internal/modules/` is inspected
- THEN directories `users/`, `roles/`, and `permissions/` do NOT exist
