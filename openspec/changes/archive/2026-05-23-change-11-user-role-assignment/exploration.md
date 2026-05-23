# Exploration: User role assignment separation

## Problem framing

`PUT /api/v1/users/:id` currently mixes general user edits with role changes. That forces ordinary user edits to require both `users.update` and `users.change_role`, and makes a broad update endpoint capable of privilege-sensitive role mutation.

Current behavior discovered:

- `internal/modules/users/dto.go`
  - `UpdateUserRequest` includes `RoleID uint json:"role_id"`.
  - `UpdateUserRequest.Validate()` requires `role_id`.
- `internal/modules/users/service.go`
  - `Update()` assigns `user.RoleID = req.RoleID` for non-root users.
  - It blocks non-root promotion to root, but role mutation still lives in the general update path.
- `internal/modules/users/routes.go`
  - `PUT /users/:id` requires both `users.update` and `users.change_role`.
- `internal/modules/roles/*`
  - Roles are tenant-scoped through `tenant.TenantContext`.
  - `root` is global (`company_id NULL`) and protected.
  - There is no assignable-role endpoint.
- `seeds/permissions.go`
  - `users.change_role` already exists.
  - `roles.assignable` does not exist.

## Recommended change id

`change-11-user-role-assignment`

## Affected domains

1. Users
   - Remove role mutation from general update.
   - Add `PATCH /api/v1/users/:id/role`.
   - Add a role-change DTO using role PublicID.
   - Add service method for role assignment with tenant, root, and self-change guards.

2. Roles
   - Add assignable roles list endpoint.
   - Recommended route: `GET /api/v1/roles/assignable`.
   - Register it before `GET /roles/:id` to avoid route conflicts.

3. Permissions
   - Add seeded `roles.assignable` system permission.
   - Keep existing `users.change_role` for actual role mutation.

4. Tenant-scoped roles / tenant isolation
   - Assignable roles must be filtered by tenant context.
   - Non-root users must only see and assign roles from their company.
   - `root` must be excluded even for root actors.

## Design decisions to carry forward

- Use `GET /api/v1/roles/assignable`, not `GET /api/v1/users/assignable-roles`, because the roles module owns role catalog data.
- Add `roles.assignable` instead of reusing `roles.index`; assignment select data is a narrower capability than full role listing.
- `PATCH /api/v1/users/:id/role` accepts `role_id` as the role PublicID string.
- `PUT /api/v1/users/:id` must not process `role_id`; if clients send it, it is ignored by the update DTO and does not mutate the user role.
- Forbid any self role change through the dedicated endpoint.
- Treat `company_id` mutation in general update as adjacent security risk; this change should tighten it by preventing non-root update from moving a user to another company.
- Changing a user's role must invalidate that user's permission cache.

## Risks

- Route conflict risk if `/roles/assignable` is registered after `/roles/:id`.
- API ID mismatch: existing user DTOs expose internal numeric `role_id`; new endpoint expects a public role ID.
- Seed drift: adding `roles.assignable` requires tests to expect the extra permission.
- Authorization regression: removing `users.change_role` from `PUT /users/:id` requires route/handler tests to change.
- Tenant escalation: role lookup for assignment must use tenant scope and still exclude root.
- Cache risk: stale permission cache after role change would preserve old permissions until TTL.
