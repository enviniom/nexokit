# Proposal: Separate user role assignment from general user update

## Intent

Separate ordinary user profile edits from privilege-sensitive role changes. `PUT /api/v1/users/:id` should require only `users.update` and must not change `role_id`; role assignment moves to a dedicated endpoint guarded by `users.change_role`.

## Scope

### In Scope

- Remove role mutation from the general user update flow.
- Add `PATCH /api/v1/users/:id/role` for changing a user's role.
- Add `GET /api/v1/roles/select` for role-select/list data.
- Add seeded `roles.select` permission.
- Validate that `root` cannot be assigned through API.
- Exclude `root` from assignable roles select options.
- Enforce tenant-aware role assignability.
- Prevent self role changes through the dedicated endpoint.
- Invalidate the target user's permission cache after successful role change.
- Add tests covering permissions, errors, tenant boundaries, root exclusion, self-change prevention, and general update behavior.

### Out of Scope

- Redesigning user creation DTOs that still use numeric `role_id`.
- Full migration from internal numeric IDs to public IDs in all existing user responses.
- Company onboarding/default role creation.
- A generic role hierarchy/permission-diff escalation engine.

## Capabilities

### New Capabilities

- `user-role-assignment`: Dedicated user role assignment endpoint and roles select selection flow.

### Modified Capabilities

- `users`: General update no longer accepts or mutates `role_id`; it requires only `users.update`.
- `roles`: Adds a roles select endpoint under the roles module.
- `permissions`: Adds `roles.select` as a seeded system permission.
- `tenant-scoped-roles`: Clarifies tenant-aware role select visibility and assignment constraints.

## Approach

1. **Users update split**
   - Remove `RoleID` from `UpdateUserRequest` and its validation.
   - Remove `user.RoleID = req.RoleID` from `userService.Update()`.
   - Change `PUT /users/:id` route middleware to require only `users.update`.
   - Preserve root self-edit restrictions.
   - Tighten `company_id` handling so non-root updates cannot move a user outside the tenant context.

2. **Dedicated role change**
   - Add `ChangeUserRoleRequest` with `RoleID string json:"role_id"` where the value is the role PublicID.
   - Add `Handler.ChangeRole` and `Service.ChangeRole`.
   - Route: `PATCH /api/v1/users/:id/role`, guarded by `users.change_role`.
   - Forbid self role changes by comparing actor PublicID with target PublicID.
   - Look up the target user with tenant scope.
   - Look up the target role with tenant scope.
   - Reject `root` role assignment regardless of actor, including root.
   - Reject cross-company assignment with an explicit invariant: when both target user and role have `company_id`, they must match. This check is required even for root/global contexts where tenant-scoped lookups may see multiple companies.
   - Persist the role change and invalidate the target user's permission cache.

3. **Roles select**
   - Add route `GET /api/v1/roles/select`, guarded by `roles.select`.
   - Register before `GET /api/v1/roles/:id` to avoid route matching conflicts.
   - Return tenant-scoped roles excluding `root`.
   - Use a compact `RoleSelectResponse` suitable for selects (`id`, `name`, optional `slug`, optional `company_id`) conforming to standard flat `{id, name, ...}` option layout.

4. **Permissions**
   - Add `permissions.ActionSelect = "select"` or equivalent constant.
   - Seed `roles.select` as system permission.
   - Update permission seed tests and any expected counts.

## Affected Areas

- `internal/modules/users/dto.go` — remove `RoleID` from update DTO; add change-role DTO.
- `internal/modules/users/routes.go` — adjust `PUT` permission; add `PATCH /:id/role`.
- `internal/modules/users/handler.go` — add `ChangeRole` handler.
- `internal/modules/users/service.go` — remove role mutation from `Update`; add `ChangeRole`; invalidate cache.
- `internal/modules/roles/contracts.go` — expose a small role assignment reader contract/summary owned by roles, not the GORM model.
- `internal/app/container.go` — inject the roles contract implementation into users service.
- `internal/modules/roles/dto.go` — add role select response DTO.
- `internal/modules/roles/routes.go` — add `GET /select` before `GET /:id`.
- `internal/modules/roles/handler.go` — add role select handler.
- `internal/modules/roles/service.go` — add `ListSelect`.
- `internal/modules/roles/repository.go` — add query helper or reuse scoped list with root exclusion.
- `internal/modules/permissions/model.go` — add action constant for `select`.
- `seeds/permissions.go` — seed `roles.select`.
- Tests under `internal/modules/users`, `internal/modules/roles`, `seeds`.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Route conflict with `/roles/:id` | Medium | Register `/roles/select` before `/:id`. |
| Stale authz cache after role change | Medium | Inject cache into users service and delete `rbac:permissions:<target_public_id>`. |
| Tenant role bypass | Medium | Use `tenant.TenantContext` for lookups and explicitly verify target user company matches target role company, including root/global contexts. |
| Existing tests expect `role_id` in update | High | Update tests to assert role remains unchanged and route no longer needs `users.change_role`. |
| Numeric/public ID confusion | Medium | Keep existing create/update scope narrow; new change-role endpoint explicitly uses role PublicID. |
| Review size | Medium | Keep implementation localized; pause if forecast exceeds 700 changed lines. |

## Rollback Plan

Revert code changes and permission seed addition. If `roles.select` has already been seeded in a development database, it can remain harmless or be removed with a targeted down/cleanup script. No schema migration is expected for this change.

## Dependencies

- Existing RBAC permissions and middleware.
- Tenant-scoped roles from change 10.
- Permission cache key convention: `rbac:permissions:<user_public_id>`.

## Success Criteria

- [ ] `PUT /api/v1/users/:id` requires only `users.update` and does not mutate role even if `role_id` is present in JSON.
- [ ] `PATCH /api/v1/users/:id/role` is the only role-change endpoint.
- [ ] Role changes require `users.change_role`.
- [ ] Users without `users.change_role` receive 403 for role changes.
- [ ] `root` cannot be assigned.
- [ ] `root` is excluded from roles select.
- [ ] Non-root admins only see/assign roles from their company.
- [ ] Cross-company roles cannot be assigned.
- [ ] Users cannot change their own role through the endpoint.
- [ ] Tests cover update, role change, roles select, and expected errors.
