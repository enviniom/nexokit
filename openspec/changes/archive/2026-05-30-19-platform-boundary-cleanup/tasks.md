# Tasks: Reinforce Platform Boundary — No Functional Change

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~350–450 (18 files, mostly small edits + 4 new files) |
| 400-line budget risk | Medium |
| 800-line review budget | Within budget |
| Chained PRs recommended | No |
| Suggested split | Single PR (focused refactor, each phase builds on previous) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Medium

## Phase 1: Foundation — New Core Files and Platform Cleanup

- [x] 1.1 Create `internal/modules/roles/core/constants.go` with `const ModuleRoles = "roles"`.
- [x] 1.2 Create `internal/modules/users/core/constants.go` with `const ModuleUsers = "users"`.
- [x] 1.3 Create `internal/modules/companies/core/constants.go` with `const ModuleCompanies = "companies"`.
- [x] 1.4 Create `internal/modules/permissions/core/constants.go` with `const ModulePermissions = "permissions"`.
- [x] 1.5 Create `internal/modules/auth/core/constants.go` with `const ModuleAuth = "auth"`.
- [x] 1.6 Remove `Module*` constants from `internal/platform/permissions/constants.go` (keep `Action*`, `Format`, `Humanize*`, `DefaultDisplayOrder`).
- [x] 1.7 Remove `MsgRoleHasAssignedUsers` from `internal/platform/messages/messages.go`.
- [x] 1.8 Change `ErrUnprocessable` in `internal/platform/apperror/apperror.go` to `&AppError{Message: ""}` (generic 422 sentinel).

## Phase 2: Domain Language — Roles Module Ownership

- [x] 2.1 Create `internal/modules/roles/core/messages.go` with `const MsgRoleHasAssignedUsers = "El rol tiene usuarios asignados"`.
- [x] 2.2 Create `internal/modules/roles/core/error.go` with `var ErrRoleHasAssignedUsers = apperror.Wrap(apperror.ErrUnprocessable, MsgRoleHasAssignedUsers)`.
- [x] 2.3 Update `internal/modules/roles/service.go` line 250: replace `apperror.ErrUnprocessable` with `core.ErrRoleHasAssignedUsers` (import `github.com/enviniom/nexokit/internal/modules/roles/core`).

## Phase 3: Integration — Route Rewiring

- [x] 3.1 Update `internal/modules/users/routes.go`: import module `core`, replace `platformPerms.ModuleUsers` with `core.ModuleUsers`.
- [x] 3.2 Update `internal/modules/roles/routes.go`: import module `core`, replace `platformPerms.ModuleRoles` with `core.ModuleRoles`.
- [x] 3.3 Update `internal/modules/companies/routes.go`: import module `core`, replace `platformPerms.ModuleCompanies` with `core.ModuleCompanies`.
- [x] 3.4 Update `internal/modules/permissions/routes.go`: import module `core`, replace `platformPerms.ModulePermissions` with `core.ModulePermissions`.
- [x] 3.5 Verify `internal/modules/auth/routes.go` needs no change (no `Module*` references).
- [x] 3.6 Verify `internal/middleware/authorization.go` unchanged (imports `platform/permissions` for `Action*` and registry only).

## Phase 4: Testing — Verify Behavior Preserved

- [x] 4.1 Update `internal/platform/apperror/apperror_test.go`: add test asserting `ErrUnprocessable.Message` is empty and `Status(ErrUnprocessable) == 422`.
- [x] 4.2 Update `internal/platform/permissions/permissions_test.go` `TestFormat`: replace `ModuleUsers`, `ModuleRoles`, `ModulePermissions` with literal strings `"users"`, `"roles"`, `"permissions"`.
- [x] 4.3 Update `internal/modules/roles/service_test.go` line 664: assert `errors.Is(err, core.ErrRoleHasAssignedUsers)` AND `errors.Is(err, apperror.ErrUnprocessable)` (both must hold).
- [x] 4.4 Run `go build ./...`, `go vet ./...`, `go test ./...` — all must pass with zero errors.
- [x] 4.5 Verify HTTP 422 response for role-delete-with-users is unchanged (same status, same message text).
