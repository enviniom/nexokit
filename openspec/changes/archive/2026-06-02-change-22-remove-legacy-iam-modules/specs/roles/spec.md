# Delta for Roles

## REMOVED Requirements

### Requirement: Role CRUD API

(Reason: Superseded by `iam-module` — "IAM role endpoints" covers identical routes, payloads, and validation)

### Requirement: Delete guard for assigned users

(Reason: Superseded by `iam-module` — "IAM role endpoints" includes delete-blocked-by-assigned-users scenario)

### Requirement: Role API

(Reason: Superseded by `iam-module` — "IAM role endpoints" covers identical GET and mutation endpoints)

### Requirement: Role permission catalog endpoint

(Reason: Superseded by `iam-module` — "IAM role endpoints" includes `GET /api/v1/roles/:id/permissions`)

### Requirement: Role permission assignment endpoint

(Reason: Superseded by `iam-module` — "IAM role endpoints" includes `PUT /api/v1/roles/:id/permissions` with cache invalidation)

### Requirement: Role seeds

(Reason: Superseded by `iam-module` — seeding behavior is owned by IAM and platform permissions sync)

### Requirement: System role protection

(Reason: Superseded by `iam-module` — "IAM role endpoints" enforces system role and root role protection)

### Requirement: Reserved slug validation

(Reason: Superseded by `iam-module` — "IAM role endpoints" rejects reserved slugs with HTTP 422)
