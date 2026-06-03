# Delta for Permissions

## REMOVED Requirements

### Requirement: Permission model

(Reason: Superseded by `iam-module` — IAM defines `IAMPermission` in `iam/core/model.go` with identical schema and constraints)

### Requirement: Admin CRUD endpoints

(Reason: Superseded by `iam-module` — "IAM permission endpoints" covers identical routes, grouping, and system permission protection)

### Requirement: Automatic admin permission synchronization

(Reason: Superseded by `iam-module` — "Permission synchronization" covers identical idempotent sync and admin auto-assignment)

### Requirement: Permission seeds

(Reason: Superseded by `iam-module` and `platform/permissions` — seeding is owned by IAM sync and platform registration)
