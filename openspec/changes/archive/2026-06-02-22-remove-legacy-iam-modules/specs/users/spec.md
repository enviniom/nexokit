# Delta for Users

## REMOVED Requirements

### Requirement: User CRUD endpoints

(Reason: Superseded by `iam-module` — "IAM user endpoints" covers identical routes, payloads, status codes, and tenant-scoping behavior)

### Requirement: Password change endpoint

(Reason: Superseded by `iam-module` — "IAM user endpoints" includes `PATCH /api/v1/users/:id/password` with identical behavior)

### Requirement: User status toggle

(Reason: Superseded by `iam-module` — "IAM user endpoints" includes `PATCH /api/v1/users/:id/status` with identical behavior)
