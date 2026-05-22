# Tasks: Custom Administerable Roles

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~250-350 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR — all tasks fit within 400-line budget |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation: error sentinel, message, CountByRoleID, slug validation | PR 1 | Base branch; tests included |
| 2 | Service: slug uniqueness + delete guard + handler 204/422 mapping | PR 1 | Depends on Unit 1 |
| 3 | Routes: wire POST/PUT/DELETE with permission guards | PR 1 | Depends on Unit 2 |
| 4 | Seeds: roles.create/update/delete + admin assignment | PR 1 | Independent |

## Phase 1: Foundation (TDD: RED → GREEN)

- [x] 1.1 **RED**: Add test in `apperror/apperror_test.go` for `ErrUnprocessable` mapping to HTTP 422 in `Status()`.
- [x] 1.2 **GREEN**: Add `ErrUnprocessable = &AppError{Message: messages.MsgRoleHasAssignedUsers}` to `apperror.go` and `Status()` case returning `http.StatusUnprocessableEntity`.
- [x] 1.3 **RED**: Add test in `messages/messages.go` verifying `MsgRoleHasAssignedUsers = "role has assigned users"` constant exists (or add constant directly — trivial, no separate test needed).
- [x] 1.4 **GREEN**: Add `MsgRoleHasAssignedUsers` constant to `messages.go`.
- [x] 1.5 **RED**: Add test in `users/repository_test.go` for `CountByRoleID` — create users with a role, assert count matches.
- [x] 1.6 **GREEN**: Add `CountByRoleID(roleID uint) (int64, error)` to `users.Repository` interface and `GormRepository` implementation using `COUNT(*) WHERE role_id = ?`.
- [x] 1.7 **RED**: Add slug format test cases to `dto_test.go` — `UPPER-CASE`, `-leading`, `trailing-`, valid `lower-case` for both `CreateRoleRequest` and `UpdateRoleRequest`.
- [x] 1.8 **GREEN**: Add `validator.ValidSlug()` call to `CreateRoleRequest.Validate()` and `UpdateRoleRequest.Validate()` in `dto.go`.

## Phase 2: Core Implementation (TDD: RED → GREEN)

- [x] 2.1 **RED**: Add service test in `service_test.go` — `Create` returns `ErrConflict` when slug already exists (extend `fakeRepository` with `GetBySlug` support).
- [x] 2.2 **GREEN**: Add slug uniqueness check in `service.Create()` — call `repo.GetBySlug(req.Slug)` before create, return `ErrConflict` if found.
- [x] 2.3 **RED**: Add service test — `Update` returns `ErrConflict` when changed slug already exists on another role.
- [x] 2.4 **GREEN**: Add changed-slug uniqueness check in `service.Update()` — if `role.Slug != req.Slug`, check `GetBySlug`, return `ErrConflict` if collision.
- [x] 2.5 **RED**: Add service test — `Delete` returns `ErrUnprocessable` when `CountByRoleID > 0` (extend `fakeRoleMemberRepository` with `CountByRoleID`).
- [x] 2.6 **GREEN**: Add `CountByRoleID` to `roleMemberRepository` interface; in `service.Delete()`, after system-role check, call `CountByRoleID(role.ID)`, return `ErrUnprocessable` if count > 0.
- [x] 2.7 **RED**: Add handler test in `handler_test.go` — `Delete` returns HTTP 204 on success (no body).
- [x] 2.8 **GREEN**: Change `handler.Delete()` to return `c.Status(http.StatusNoContent)` instead of `response.Success[any](...)` on successful delete.
- [x] 2.9 **RED**: Add handler test — `Delete` returns HTTP 422 when service returns `ErrUnprocessable`.
- [x] 2.10 **GREEN**: Add `case http.StatusUnprocessableEntity` in `handler.Delete()` error mapping, return `response.Error(c, http.StatusUnprocessableEntity, messages.MsgRoleHasAssignedUsers, nil)`.

## Phase 3: Route Wiring

- [x] 3.1 **RED**: Add route test cases in `routes_test.go` for `POST /roles` → `roles.create`, `PUT /roles/:id` → `roles.update`, `DELETE /roles/:id` → `roles.delete`.
- [x] 3.2 **GREEN**: Register `POST`, `PUT`, `DELETE` routes in `routes.go` with `requirePermission("roles.create")`, `requirePermission("roles.update")`, `requirePermission("roles.delete")` respectively, wiring to `handler.Create`, `handler.Update`, `handler.Delete`.

## Phase 4: Seeds

- [x] 4.1 **RED**: Add seed test in `seeds/permissions_test.go` — verify `roles.create`, `roles.update`, `roles.delete` exist after `seedPermissions()`.
- [x] 4.2 **GREEN**: Add `roles.create`, `roles.update`, `roles.delete` permissions to `systemPermissions()` in `seeds/permissions.go` with display orders between existing role permissions.
- [x] 4.3 **RED**: Add seed test — verify admin role receives `roles.create`, `roles.update`, `roles.delete` after `seedRolePermissions()`.
- [x] 4.4 **GREEN**: Add `"roles.create"`, `"roles.update"`, `"roles.delete"` to `adminPermissionSlugs()` in `seeds/role_permissions.go`.

## Phase 5: Verification

- [x] 5.1 Run `go test ./internal/platform/apperror ./internal/modules/roles ./internal/modules/users ./seeds -v` — all pass.
- [x] 5.2 Run `go test ./...` — full suite passes.
- [x] 5.3 Run `go build ./...` — no compilation errors.
