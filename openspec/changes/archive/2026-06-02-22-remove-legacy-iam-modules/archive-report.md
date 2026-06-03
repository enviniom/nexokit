# Archive Report: Remove Legacy IAM Modules

## Status

Archived successfully.

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| `iam-module` | Updated | Removed legacy preservation requirement; added no residual legacy references requirement. |
| `app-orchestration` | Updated | Replaced legacy-compile scenario with no-legacy-directories scenario. |
| `users` | Updated | Marked standalone users module spec as superseded by IAM. |
| `roles` | Updated | Marked standalone roles module spec as superseded by IAM. |
| `permissions` | Updated | Marked standalone permissions module spec as superseded by IAM. |
| `rbac-authorization` | Updated | Moved user/role/permission domain constants ownership to IAM. |
| `platform-boundary-rules` | Updated | Moved user/role/permission domain language ownership to IAM and removed legacy path references. |

## Verification

| Command | Result |
|---------|--------|
| `go list ./...` | Passed |
| `go build ./...` | Passed |
| `go test ./...` | Passed |

## Archive Notes

- Legacy `internal/modules/users/`, `internal/modules/roles/`, and `internal/modules/permissions/` Go files were removed.
- IAM is the sole source for user, role, and permission routes/models/contracts.
- During archive sync, IAM route module constants and `MsgRoleHasAssignedUsers` were added so canonical specs accurately match implementation.
