# Proposal: Custom Administerable Roles

## Intent

Enable CRUD operations for custom roles through existing but unwired handlers, with slug validation, assigned-user delete protection, system-role immutability, and permission seeds for new `roles.*` actions. The roles module already has 80% of the implementation — handlers, service methods, and DTOs exist.

## Scope

### In Scope
- Wire POST/PUT/DELETE routes with `requirePermission` guards
- Add `ValidSlug()` to `CreateRoleRequest` and `UpdateRoleRequest`
- Add `CountByRoleID` to users repository for efficient delete guard
- Add assigned-user check in `Delete` service method (returns 422)
- Seed `roles.create`, `roles.update`, `roles.delete` permissions
- Add error message for "role has assigned users"
- Tests for route wiring, slug validation, delete guard

### Out of Scope
- Separate `PATCH /users/:id/role` endpoint (deferred — breaking change, affects users module)
- Hard delete for custom roles (soft delete via GORM `deleted_at` follows project convention)
- Role hierarchy or permission inheritance

## Capabilities

### New Capabilities
- `role-crud`: Full CRUD for custom roles (POST, PUT, DELETE) with permission guards, slug validation, and assigned-user delete protection

### Modified Capabilities
- `roles`: Delta spec — upgrade from read-only to full CRUD; add slug uniqueness check; add assigned-user delete guard
- `permissions`: Delta spec — seed `roles.create`, `roles.update`, `roles.delete`

## Approach

Extend the existing roles module. Wire routes in `routes.go`, add `ValidSlug()` to DTOs, add `CountByRoleID` to users repository, check assigned users before delete in service, seed new permissions. Handlers already exist and need no changes.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/roles/routes.go` | Modified | Mount POST, PUT, DELETE with `requirePermission` |
| `internal/modules/roles/dto.go` | Modified | Add `ValidSlug()` to Create/Update requests |
| `internal/modules/roles/service.go` | Modified | Add slug uniqueness in Create; assigned-user check in Delete |
| `internal/modules/users/repository.go` | Modified | Add `CountByRoleID(roleID uint) (int64, error)` |
| `internal/platform/messages/messages.go` | Modified | Add `MsgRoleHasAssignedUsers` |
| `seeds/permissions.go` | Modified | Seed `roles.create`, `roles.update`, `roles.delete` |
| `seeds/role_permissions.go` | Modified | Assign new role permissions to admin role |
| `openspec/specs/roles/spec.md` | Delta | Upgrade from read-only to CRUD spec |
| `openspec/specs/permissions/spec.md` | Delta | Add new role permissions to seed table |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Slug collision with existing roles | Medium | Check slug uniqueness in Create and Update (against all roles, not just name) |
| Delete guard performance | Low | Use `CountByRoleID` (COUNT query) instead of `ListPublicIDsByRoleID` |
| Review budget | Low | Estimated ~200-300 lines (routes, validation, delete guard, seeds, tests). Within 400-line budget. |
| Soft delete semantics | Low | Consistent with BaseModel convention; `deleted_at` already in place |

## Rollback Plan

1. Remove POST/PUT/DELETE route registrations from `routes.go`
2. Revert DTO validation changes (remove `ValidSlug()` calls)
3. Remove new permission seeds (idempotent — safe to leave, but can be cleaned)
4. No data migration needed — custom roles remain in DB but become inaccessible via API

## Dependencies

- None (all required infrastructure exists: handlers, service methods, repository CRUD)

## Success Criteria

- [ ] `POST /api/v1/roles` creates custom roles with validated slugs, returns 409 on name/slug collision
- [ ] `PUT /api/v1/roles/:id` updates custom roles, returns 403 on system roles
- [ ] `DELETE /api/v1/roles/:id` returns 422 when role has assigned users, 403 on system roles
- [ ] `roles.create`, `roles.update`, `roles.delete` permissions are seeded and assigned to admin role
- [ ] All new routes require appropriate `roles.*` permissions
- [ ] Tests cover slug validation, delete guard, and system-role protection
