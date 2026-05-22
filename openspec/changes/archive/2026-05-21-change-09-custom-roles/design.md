# Design: Custom Administerable Roles

## Technical Approach

Complete the existing roles module instead of introducing a new module. Handlers and service CRUD methods already exist; implementation should wire mutation routes, tighten DTO/service validation, add an efficient user-count dependency, and extend idempotent seeds. User role-change separation remains out of scope.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Route wiring | Add POST/PUT/DELETE in `internal/modules/roles/routes.go` with `requirePermission("roles.create|update|delete")` | New role admin router | Existing modules register all CRUD routes in one `Register` function. |
| Delete guard | Reuse `WithRoleMembers(usersRepo)` and extend its interface with `CountByRoleID(uint) (int64, error)` | Query users from roles repo; list user IDs | The app container already injects `usersRepo`; COUNT avoids loading IDs just to guard delete. |
| Assigned-user error | Return a 422 app error/message and map it explicitly in `Handler.Delete` | Return 409 or generic validation | Spec requires 422 and user-facing “role has assigned users”. |
| System-role protection | Keep service-level guards in `Update`, `Delete`, and existing `AssignPermissions` | Middleware-only guard | Protection is business logic and must hold outside HTTP tests too. |
| Seeding | Add `roles.create/update/delete` to `systemPermissions()` and `adminPermissionSlugs()` | Runtime-created permissions | Seeds are already the idempotent source for system permissions. |

## Data Flow

```text
HTTP request → roles route permission guard → Handler DTO validation
  → roleService → roles.Repository + users.Repository.CountByRoleID
  → GORM soft delete / create / update → response envelope or 204
```

Delete flow: load role by public ID → reject missing/system role → `CountByRoleID(role.ID)` → if count > 0 return 422 → `repo.Delete(publicID)` soft-deletes.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/roles/routes.go` | Modify | Register `POST /roles`, `PUT /roles/:id`, `DELETE /roles/:id` with role CRUD permissions. |
| `internal/modules/roles/dto.go` | Modify | Apply `validator.ValidSlug()` to create/update slug validation. |
| `internal/modules/roles/service.go` | Modify | Check slug uniqueness in `Create` and changed-slug uniqueness in `Update`; extend `roleMemberRepository`; guard delete with `CountByRoleID`; return 422 error for assigned users. |
| `internal/modules/roles/handler.go` | Modify | Map assigned-user delete error to HTTP 422; return 204 No Content on successful delete. |
| `internal/modules/users/repository.go` | Modify | Add `CountByRoleID(roleID uint) (int64, error)` to interface and GORM repository. |
| `internal/platform/messages/messages.go` | Modify | Add `MsgRoleHasAssignedUsers = "role has assigned users"`. |
| `seeds/permissions.go` | Modify | Add role create/update/delete permissions with display orders before `assign_permissions`. |
| `seeds/role_permissions.go` | Modify | Add new role management permissions to admin role. |
| Tests listed below | Modify | Extend existing table-driven unit tests. |

## Interfaces / Contracts

```go
type roleMemberRepository interface {
    ListPublicIDsByRoleID(roleID uint) ([]string, error)
    CountByRoleID(roleID uint) (int64, error)
}
```

Slug validation contract: `slug` is required and must pass `validator.ValidSlug()` (`^[a-z0-9]+(?:-[a-z0-9]+)*$`). Delete success contract should match spec: HTTP 204 with no response body.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| DTO unit | Invalid uppercase/leading-hyphen slugs for create/update | Extend `roles/dto_test.go` table cases. |
| Service unit | Create/update slug collisions; delete blocks assigned users; system-role delete still 403 | Extend fake repositories in `roles/service_test.go`. |
| Repository unit | `CountByRoleID` counts active users assigned to a role | Add case in `users/repository_test.go` using SQLite + `t.TempDir` not needed. |
| Route unit | POST/PUT/DELETE permission guards | Extend `roles/routes_test.go` table. |
| Handler unit | Delete returns 204 on success and 422 for assigned users | Extend `roles/handler_test.go`. |
| Seeds unit | `roles.create/update/delete` exist and admin receives them | Extend `seeds/permissions_test.go`. |

Run narrow package tests first (`go test ./internal/modules/roles ./internal/modules/users ./seeds`), then `go test ./...`.

## Migration / Rollout

No schema migration required. Existing deployments must rerun idempotent seeds so new permissions and admin role links are inserted. Rollback removes route wiring and leaves harmless seeded permission rows.

## Open Questions

- None.
