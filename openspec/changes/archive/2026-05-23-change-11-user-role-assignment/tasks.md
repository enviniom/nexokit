# Tasks: User role assignment separation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 500–650 |
| User review budget | 700 changed lines |
| Budget risk | Medium |
| Chained PRs recommended | Optional |
| Suggested split if needed | PR 1 users role-change endpoint → PR 2 roles select + permission seed → PR 3 tests/cleanup |
| Delivery strategy | single PR acceptable under current 700-line budget; pause if forecast rises above 700 |

Decision needed before apply: Yes — human should review generated proposal/spec/design/tasks before implementation.
Chained PRs recommended: Optional
Chain strategy: single-pr-default-with-auto-forecast
700-line budget risk: Medium

## Phase 1: Baseline RED tests for current unsafe behavior

- [x] 1.1 Add/update users route test proving `PUT /api/v1/users/:id` requires only `users.update` and not `users.change_role`.
- [x] 1.2 Add handler/integration binding test proving general `PUT /users/:id` with `role_id` in the JSON body does not mutate `RoleID`; add service test proving `Update` has no role mutation path.
- [x] 1.3 Add users service/handler test proving `PATCH /api/v1/users/:id/role` requires a non-empty role PublicID.
- [x] 1.4 Add authorization/route test proving callers without `users.change_role` receive 403 on `PATCH /users/:id/role`.
- [x] 1.5 Add roles route/service test proving `GET /api/v1/roles/select` requires `roles.select`.
- [x] 1.6 Run targeted tests and record RED evidence before implementation.

## Phase 2: User update DTO and route separation

- [x] 2.1 Modify `internal/modules/users/dto.go` — remove `RoleID` from `UpdateUserRequest` and remove `role_id` validation from `Validate()`.
- [x] 2.2 Add `ChangeUserRoleRequest` with `RoleID string json:"role_id"` and validation requiring `role_id`.
- [x] 2.3 Modify `internal/modules/users/routes.go` — change `PUT /:id` to only require `users.update`.
- [x] 2.4 Add `PATCH /:id/role` requiring `users.change_role`, registered near other `/:id` user operations.
- [x] 2.5 Update users handler fakes/tests to compile with the new DTO shape.

## Phase 3: Users service role-change logic

- [x] 3.1 Add a roles-owned assignment reader contract/summary in `roles/contracts.go` for tenant-scoped role lookup by PublicID; implement it in `roles/repository.go` (`GormRepository`).
- [x] 3.2 Implement **Functional Options** (`ServiceOption`) in `users.NewService` to inject optional collaborators: `WithRoleReader(r roles.AssignmentRoleReader)` and `WithCache(c cache.Cache)`. This prevents breaking other constructors and test fixtures.
- [x] 3.3 Modify `userService.Update()` — remove `user.RoleID = req.RoleID` and preserve the existing role for all general updates.
- [x] 3.4 Tighten non-root `company_id` handling in `Update()` — reject attempts to move the user outside the current tenant company; do not let request body choose an arbitrary company.
- [x] 3.5 Implement `ChangeRole(tc, targetPublicID, actorPublicID, req)`.
- [x] 3.6 In `ChangeRole`, reject empty actor or actor targeting their own PublicID with `ErrForbidden`.
- [x] 3.7 In `ChangeRole`, load target user through `usersRepo.GetByPublicID(tc, targetPublicID)` and map tenant misses to `ErrNotFound`.
- [x] 3.8 In `ChangeRole`, load target role through the roles-owned assignment reader using `req.RoleID` as role PublicID.
- [x] 3.9 In `ChangeRole`, reject `root` role assignment regardless of actor.
- [x] 3.10 In `ChangeRole`, explicitly reject target user / target role company mismatch when both have `company_id`; this must also cover root/global contexts.
- [x] 3.11 In `ChangeRole`, update only `targetUser.RoleID`, persist through users repo, reload response, and invalidate target permission cache.

## Phase 4: Users handler wiring

- [x] 4.1 Add `Handler.ChangeRole(c *gin.Context)` in `internal/modules/users/handler.go`.
- [x] 4.2 Bind JSON into `ChangeUserRoleRequest`, run validation, and respond with validation errors using existing project response helpers.
- [x] 4.3 Extract tenant context and actor PublicID using the same helpers as existing update/password/status handlers.
- [x] 4.4 Call `service.ChangeRole` and return standard success response with updated user DTO.
- [x] 4.5 Update handler tests for success, validation error, forbidden self-change, unknown role, and root role rejection.

## Phase 5: Roles select endpoint

- [x] 5.1 Add standard transversal `SelectResponse` DTO to `internal/platform/response/response.go` with flat `id` and `name` strings, and optional `meta` map.
- [x] 5.2 Modify `internal/modules/roles/repository.go` — add `ListSelect(tc tenant.TenantContext) ([]Role, error)` or equivalent query helper.
- [x] 5.3 Implement roles select query using `tenant.ApplyTenantScope` and `WHERE slug <> 'root'`; preload company to populate company PublicID.
- [x] 5.4 Modify `internal/modules/roles/service.go` — add `ListSelect(tc)` mapping roles to standard `response.SelectResponse` DTOs (populating slug and company_id in meta).
- [x] 5.5 Modify `internal/modules/roles/handler.go` — add `ListSelect` handler using tenant context and standard response envelope.
- [x] 5.6 Modify `internal/modules/roles/routes.go` — register `GET /select` with `roles.select` before `GET /:id`.
- [x] 5.7 Add roles tests proving root is excluded, tenant scope is applied, root global behavior still excludes root, and route order does not treat `select` as `:id`.

## Phase 6: Permission seed update

- [x] 6.1 Modify `internal/modules/permissions/model.go` — add `ActionSelect = "select"`.
- [x] 6.2 Modify `seeds/permissions.go` — add seeded system permission `roles.select` with a stable display order near role business actions.
- [x] 6.3 Update `seeds/permissions_test.go` expected permission count and assert `roles.select` has module `roles`, action `select`, `is_system = true`.
- [x] 6.4 Run seed tests and record GREEN evidence.

## Phase 7: Full behavior tests and triangulation

- [x] 7.1 Add handler/integration test: general update with a different `role_id` in JSON leaves `RoleID` unchanged; service-level assertion should verify there is no role mutation path.
- [x] 7.2 Add service test: successful role change updates role and invalidates cache.
- [x] 7.3 Add service test: assigning root returns forbidden/validation and leaves role unchanged.
- [x] 7.4 Add service test: target role from another company is not found for non-root tenant scope and leaves role unchanged.
- [x] 7.5 Add service test: root/global cannot assign a role from company 2 to a user in company 1.
- [x] 7.6 Add service test: user cannot change their own role.
- [x] 7.7 Add handler/route test: missing `users.change_role` yields 403 for `PATCH /users/:id/role`.
- [x] 7.8 Add handler/route test: missing `roles.select` yields 403 for `GET /roles/select`.
- [x] 7.9 Add test: `PUT /users/:id` without `role_id` succeeds.
- [x] 7.10 Add test: `PUT /users/:id` with `role_id` does not change role.

## Phase 8: Verification and cleanup

- [x] 8.1 Run targeted package tests for users, roles, permissions, and seeds.
- [x] 8.2 Run strict TDD full test command: `go test ./...`.
- [x] 8.3 Run build command: `go build ./...` if apply reaches verify stage.
- [x] 8.4 Review diff size; if changed lines exceed 700, stop and propose chained PR split before continuing.
- [x] 8.5 Update SDD apply progress with RED/GREEN/TRIANGULATE/REFACTOR evidence.

## Acceptance Mapping

| Acceptance criterion | Tasks |
|---|---|
| `PUT /users/:id` does not change role with `role_id` in body | 2.1, 3.3, 7.1, 7.10 |
| Role changes only via `PATCH /users/:id/role` | 2.4, 3.5, 4.1 |
| Role change requires `users.change_role` | 2.4, 1.4, 7.7 |
| Caller without `users.change_role` receives 403 | 1.4, 7.7 |
| Cannot assign root | 3.9, 7.3 |
| Root excluded from roles select | 5.3, 5.7 |
| Admin only assigns same-company roles | 3.8, 3.10, 5.3, 7.4 |
| Cannot assign role from another company | 3.8, 3.10, 7.4, 7.5 |
| Cannot self-escalate role | 3.6, 7.6 |
| Tests cover update, role change, roles select, errors | Phases 1, 5, 7, 8 |
