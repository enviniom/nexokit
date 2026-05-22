# Exploration: Custom Administerable Roles

## Current State

The RBAC system (change-04) already implements a solid foundation:

- **Model**: `Role` uses `BaseModel` (audit fields `created_at`, `updated_at`, `created_by`, `updated_by`, soft delete via `deleted_at`) and already has `IsSystem bool`. All required fields from the prompt are present.
- **DTOs**: `CreateRoleRequest` and `UpdateRoleRequest` exist with basic required+minLength validation, but **do NOT apply `validator.ValidSlug()`** to the slug field.
- **Handlers**: `Create`, `Update`, `Delete` handlers already exist in `handler.go` with proper error mapping (409, 403, 404, 500).
- **Service**: `Create`, `Update`, `Delete` methods already implement system role protection (`IsSystem` check returns `ErrForbidden`). **Critical gap**: `Delete` does NOT check for assigned users before deletion.
- **Routes**: Only **read-only** routes are mounted (`GET /roles`, `GET /roles/:id`, `GET /roles/:id/permissions`, `PUT /roles/:id/permissions`). POST, PUT, DELETE handlers exist but are **not wired** in `routes.go`.
- **Repository**: Full CRUD exists (`Create`, `Update`, `Delete`, `GetByName`, `GetBySlug`). Missing: `CountByRoleID` or `HasAssignedUsers` for delete-guard.
- **Users module**: `ListPublicIDsByRoleID(roleID uint)` already exists in `users/repository.go` — this can be reused to check if a role has assigned users.
- **Permissions**: `roles.assign_permissions` exists and is used. `roles.create`, `roles.update`, `roles.delete`, `roles.index`, `roles.view` are **not yet seeded**.
- **User role-change policy**: `PUT /api/v1/users/:id` currently requires BOTH `users.update` AND `users.change_role` (double middleware). This is the overly restrictive behavior the prompt flags.

## Affected Areas

- `internal/modules/roles/routes.go` — mount POST, PUT, DELETE routes with `requirePermission` guards
- `internal/modules/roles/dto.go` — add `ValidSlug()` to `CreateRoleRequest` and `UpdateRoleRequest`
- `internal/modules/roles/service.go` — add assigned-user check in `Delete`; consider adding `CountByRoleID` to repository contract
- `internal/modules/roles/repository.go` — add `CountByRoleID(roleID uint) (int64, error)` method
- `internal/modules/roles/model.go` — no changes needed (already complete)
- `internal/modules/roles/handler.go` — no changes needed (handlers already exist)
- `internal/modules/users/repository.go` — `ListPublicIDsByRoleID` already exists; may add `CountByRoleID` for efficiency
- `internal/platform/messages/messages.go` — add message for "role has assigned users" (e.g., `MsgRoleHasAssignedUsers = "El rol tiene usuarios asignados"`)
- `internal/platform/apperror/apperror.go` — may need `ErrUnprocessableEntity` (422) for validation errors, though `ErrValidation` already maps to 422
- `internal/cli/seeds/` — seed new `roles.*` permissions
- `openspec/specs/roles/spec.md` — update from read-only to full CRUD spec
- `internal/modules/roles/service_test.go` — add tests for assigned-user delete guard, slug validation
- `internal/modules/roles/handler_test.go` — add tests for POST/PUT/DELETE endpoints
- `internal/modules/roles/routes_test.go` — add route mounting tests

## Approaches

### Approach 1: Extend existing roles module (Recommended)

Add CRUD routes, slug validation, and assigned-user delete guard to the existing roles module. Add `CountByRoleID` to users repository contract for efficient checking.

- **Pros**: Minimal surface change; handlers already exist; follows existing patterns; no new modules
- **Cons**: Roles module grows slightly; needs careful test coverage for delete guard
- **Effort**: Medium

### Approach 2: Separate custom-roles module

Create a new `internal/modules/custom-roles/` module that only handles custom role CRUD, delegating to roles module for shared logic.

- **Pros**: Clean separation of concerns; system roles completely isolated
- **Cons**: Duplicates much of the existing handler/service logic; violates "modules don't import each other" convention unless done via contracts; adds complexity
- **Effort**: High

### Approach 3: Single role-change endpoint for users

Create `PATCH /api/v1/users/:id/role` endpoint requiring only `users.change_role`, and remove `RoleID` from `PUT /api/v1/users/:id` DTO.

- **Pros**: Clean separation of concerns; minimal permission surface for general edits
- **Cons**: Breaking change for existing API consumers; requires migration of client code
- **Effort**: Low (but breaking)

## Recommendation

**Approach 1** for the custom roles CRUD, and **defer the user role-change separation** to a follow-up change (or include as a small isolated task). The existing roles module already has 80% of the implementation — handlers, service methods, and DTOs exist. The work is primarily:

1. Wire routes (5 lines)
2. Add slug validation (2 lines per DTO)
3. Add `CountByRoleID` to users repo (5 lines)
4. Add assigned-user check in `Delete` service method (8 lines)
5. Add message constant (1 line)
6. Seed new permissions (update seed file)
7. Tests for new behavior

The user role-change separation (`PATCH /users/:id/role`) is a **separation-of-concerns** improvement that doesn't block custom role CRUD. It should be scoped separately because it affects the users module and is a potential breaking change.

## Phase Boundary Recommendation

| Phase | Scope |
|-------|-------|
| **Proposal** | Scope confirmation, acceptance criteria mapping, risk assessment |
| **Spec** | Delta spec for roles CRUD, permission seeds, delete guard, slug validation |
| **Design** | Route wiring diagram, service flow for delete guard, DTO validation changes |
| **Tasks** | Granular implementation tasks grouped by work unit (routes, validation, delete guard, seeds, tests) |

**Suggested split**: The user role-change policy (`PATCH /users/:id/role`) should be a **separate change** or a clearly isolated task at the end, because it touches the users module and requires breaking-change consideration.

## Risks

1. **Delete with assigned users**: Current `Delete` has NO guard. If implemented without `CountByRoleID`, would need to call `ListPublicIDsByRoleID` and check length — inefficient. Adding `CountByRoleID` is cleaner.
2. **Slug collision with system roles**: Creating a role with slug `root`, `admin`, or `user` should be blocked. The current `Create` only checks name uniqueness, not slug uniqueness against system roles.
3. **Permission seeds**: New `roles.*` permissions must be added to the seed file. If seeds run on startup, existing databases need the new permissions.
4. **Breaking change for user role updates**: Separating `PUT /users/:id` from role changes is a breaking API change. Should be deferred or clearly communicated.
5. **Soft delete behavior**: `BaseModel` uses GORM soft delete. The `Delete` repository method uses `db.Delete(&Role{})` which triggers soft delete. Need to confirm this is the intended behavior (vs hard delete for custom roles).
6. **Review budget**: Estimated ~200-300 lines of changes (routes, validation, delete guard, seeds, tests). Within the 400-line budget. **400-line budget risk: Low**.

## Ready for Proposal

**Yes** — the codebase is well-prepared for this change. The roles module already has handlers, service methods, and DTOs for CRUD. The gaps are route wiring, slug validation, assigned-user delete guard, and permission seeds. All patterns are established and testable.
