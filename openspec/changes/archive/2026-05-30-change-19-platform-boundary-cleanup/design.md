# Design: Reinforce platform boundary — no functional change

## Technical Approach

Move domain vocabulary out of `internal/platform/*` while keeping platform as the source for generic API contracts/utilities. The refactor is deliberately mechanical: replace domain constants/messages at their use sites, keep `platform/response.HandleError` unchanged, and preserve the same HTTP status/message/body behavior for existing endpoints.

## Architecture Decisions

| Topic | Choice | Alternatives considered | Rationale |
|-------|--------|--------------------------|-----------|
| Role delete 422 | Add `roles/core.ErrRoleHasAssignedUsers = apperror.Wrap(apperror.ErrUnprocessable, MsgRoleHasAssignedUsers)` and return it from delete | Keep role message in `apperror.ErrUnprocessable`; make every module use generic 422 directly | The roles module owns the domain message while `errors.Is(err, apperror.ErrUnprocessable)` and status 422 still work. |
| Generic 422 sentinel | Change `apperror.ErrUnprocessable` to a generic/empty-message 422 sentinel | Remove sentinel entirely | Existing tests and callers can continue matching a generic category without domain leakage. |
| Permission module names | Define module-name constants in owning module `core/constants.go`; use them in route registration | Shared `internal/shared` constants; keep constants in platform | Local constants avoid module-to-module coupling and follow the context rule that duplication is acceptable to preserve autonomy. |
| Platform response | Do not change `platform/response` | Module-specific error handlers/envelopes | Specs require `platform/response` as the sole API response contract; no response shape changes are allowed. |

## Data Flow

Role delete error path remains behaviorally identical:

```txt
roles service ── returns roles/core.ErrRoleHasAssignedUsers
      │
      ▼
platform/response.HandleError ──→ apperror.Status = 422
      │                           apperror.PublicMessage = "El rol tiene usuarios asignados"
      ▼
standard API error envelope
```

Permission route slug construction remains generic:

```txt
module/core.ModuleX + platform/permissions.ActionY ──→ platform/permissions.Format()
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/platform/messages/messages.go` | Modify | Remove `MsgRoleHasAssignedUsers`; keep generic API, validation, middleware messages. |
| `internal/platform/apperror/apperror.go` | Modify | Make `ErrUnprocessable` generic while preserving 422 mapping. |
| `internal/platform/permissions/constants.go` | Modify | Remove all `Module*` constants; keep `Action*`, `Format`, `Humanize*`, display order. |
| `internal/modules/roles/core/messages.go` | Create | Own `MsgRoleHasAssignedUsers`. |
| `internal/modules/roles/core/error.go` | Create | Define `ErrRoleHasAssignedUsers` wrapping generic 422 with role message. |
| `internal/modules/roles/service.go` | Modify | Return `core.ErrRoleHasAssignedUsers` when deleting a role with assigned users. |
| `internal/modules/users/core/constants.go` | Create | Define `ModuleUsers = "users"` for flat legacy module route use. |
| `internal/modules/roles/core/constants.go` | Create | Define `ModuleRoles = "roles"`. |
| `internal/modules/companies/core/constants.go` | Modify/Create | Define `ModuleCompanies = "companies"`. |
| `internal/modules/permissions/core/constants.go` | Create | Define `ModulePermissions = "permissions"`. |
| `internal/modules/auth/core/constants.go` | Create | Define `ModuleAuth = "auth"` for ownership of existing platform constant. |
| `internal/modules/{users,roles,companies,permissions}/routes.go` | Modify | Import module `core` and pass module-local constants into `platform/permissions.Format`. |
| `internal/platform/permissions/permissions_test.go` | Modify | Use literal module strings or package-local test constants because platform no longer exports `Module*`. |
| `internal/platform/apperror/apperror_test.go`, `internal/modules/roles/*_test.go` | Modify | Assert generic 422 remains and role-specific sentinel preserves status/message. |

## Interfaces / Contracts

```go
// internal/modules/roles/core/error.go
var ErrRoleHasAssignedUsers = apperror.Wrap(apperror.ErrUnprocessable, MsgRoleHasAssignedUsers)
```

No public HTTP contract changes: status remains 422 and response envelope still comes from `platform/response`.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|--------------|----------|
| Unit | `ErrUnprocessable` maps to 422 without domain message | Update `platform/apperror` tests. |
| Unit | Role delete sentinel matches generic 422 and exposes same public message | Update roles service/handler tests. |
| Unit | Permission `Format` remains generic after removing module constants | Update permissions tests with string inputs. |
| Integration/build | Import rewiring has no cycles or behavior drift | Run `go test ./...`, `go vet ./...`, `go build ./...`. |

## Migration / Rollout

No migration required. Pure source refactor; rollback is reverting the commit.

## Open Questions

- [ ] `ModuleSettings` currently exists in platform but no `internal/modules/settings` owner exists. Design removes it from platform with no replacement unless a settings module is later introduced.
