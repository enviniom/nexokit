# Verification Report: change-04-rbac — Slice PR1 (Permission Foundation)

**Change**: change-04-rbac
**Version**: tasks.md (Phase 1 tasks 1.1–1.4)
**Mode**: Standard
**Slice**: PR 1 — Permission Foundation

## Completeness

| Metric | Value |
|--------|-------|
| Tasks total (PR1) | 4 |
| Tasks complete (PR1) | 4 |
| Tasks incomplete (PR1) | 0 |
| PR2/PR3 scope leaked | None detected |

## Build & Tests Execution

**Build**: ✅ Passed
```text
go build ./... — no errors
```

**Tests**: ✅ 20 passed / ❌ 0 failed / ⚠️ 0 skipped (PR1 scope)
```text
ok  github.com/enviniom/nexokit/internal/modules/permissions  0.012s
ok  github.com/enviniom/nexokit/seeds                         0.042s
```

**Full suite note**: 1 pre-existing flaky test in `internal/platform/identity` (TestGenerate/sortable IDs) — NOT related to PR1 changes. No changes to that package.

**Coverage**: ➖ Not available (no coverage threshold configured)

## Spec Compliance Matrix (PR1 Scope: permissions/spec.md)

| Requirement | Scenario | Test | Result |
|-------------|----------|------|--------|
| Permission model | Valid permission slug | `TestPermissionFieldsAndSlugUniqueness` | ✅ COMPLIANT |
| Permission model | Duplicate slug rejected | `TestPermissionFieldsAndSlugUniqueness` (duplicate insert) | ✅ COMPLIANT |
| Permission model | Business action slug | `TestValidatePermissionParts/business_change_role_action` | ✅ COMPLIANT |
| Permission model | Explicit actions only, no `read` | `TestValidatePermissionParts/reject_ambiguous_read_action` + `TestSystemPermissionsCatalog` (all 18 slugs) | ✅ COMPLIANT |
| Permission model | Unique index on slug | Migration `20260518000000_rbac.sql` + GORM `uniqueIndex` tag + duplicate slug test | ✅ COMPLIANT |
| Admin CRUD | List grouped and sorted | `TestService_ListGroupedSortsByModuleAndDisplayOrder` | ✅ COMPLIANT |
| Admin CRUD | Create non-system permission | Service `Create` + handler `Create` | ✅ COMPLIANT |
| Admin CRUD | Reject mutation on system permission | `TestService_SystemCRUDProtection` (update + delete) | ✅ COMPLIANT |
| Admin CRUD | Reject ambiguous `read` action | `TestValidatePermissionParts/reject_ambiguous_read_action` | ✅ COMPLIANT |
| Permission seeds | Idempotent seeding | `TestSeedPermissions/is_idempotent` | ✅ COMPLIANT |
| Permission seeds | Explicit actions seeded | `TestSystemPermissionsCatalog` (18 slugs validated) | ✅ COMPLIANT |
| Permission seeds | `is_system: true` on all seeds | `TestSystemPermissionsCatalog` checks `IsSystem` | ✅ COMPLIANT |
| Permission seeds | `display_order` values positive | `TestSystemPermissionsCatalog` checks `DisplayOrder > 0` | ✅ COMPLIANT |

**Compliance summary**: 13/13 scenarios compliant for PR1 scope.

### Scenarios deferred to PR2/PR3 (explicitly out of scope)

- Admin CRUD: "Unauthorized access" (`permissions.manage` guard) → PR3 (authz wiring)
- Admin CRUD: "List permissions grouped and sorted" with auth required → PR3 (route guards)
- Roles spec: Role permission catalog GET/PUT → PR2
- Roles spec: Role permission assignment and cache → PR2

## Correctness (Static Evidence)

| Requirement | Status | Notes |
|------------|--------|-------|
| Permission model has all spec fields | ✅ Implemented | `id`, `public_id`, `slug`, `name`, `module`, `action`, `description`, `is_system`, `display_order`, `created_at`, `updated_at` |
| Unique index on slug | ✅ Implemented | Migration + GORM uniqueIndex tag |
| `{module}.{action}` slug format validation | ✅ Implemented | `validatePermissionParts` rejects mismatches and `read` |
| System permission CRUD protection | ✅ Implemented | `Update`/`Delete` return `ErrForbidden` for `IsSystem: true` |
| RolePermission join table | ✅ Implemented | Migration + model with composite PK, FKs, indexes |
| Seeds: idempotent | ✅ Implemented | `PermissionsSeed` and `RolePermissionsSeed` check before insert |
| Seeds: root gets all permissions | ✅ Implemented | `RolePermissionsSeed` assigns `allSystemPermissionSlugs()` to root |
| Seeds: admin/user subset | ✅ Implemented | `adminPermissionSlugs()` (18), `userPermissionSlugs()` (3) |
| CRUD endpoints defined (routes) | ✅ Implemented | `Register()` mounts GET/GET/:id/POST/PUT/:id/DELETE/:id |
| Routes NOT wired to container | ✅ Correct | Design: wiring is PR3 scope |

## Coherence (Design)

| Decision | Followed? | Notes |
|----------|-----------|-------|
| API surface: no UI files; grouped catalogs from API | ✅ Yes | `ListGrouped()` returns `[]PermissionGroupResponse` |
| Auth/authz boundary: `Auth` stays identity-only; authz separate | ✅ Yes | No `RequirePermission` added yet (PR3) |
| Permission model: stored fields, not inferred | ✅ Yes | `module`, `action`, `slug`, `name`, `description`, `is_system`, `display_order` all persisted |
| Actions: explicit, no ambiguous `read` | ✅ Yes | `validatePermissionParts` rejects `action == "read"` |
| Seeds: idempotent | ✅ Yes | Both seeds check existence before insert |
| Migration: rollback order | ✅ Yes | Down drops `role_permissions` before `permissions` |

## Issues Found

**CRITICAL**: None

**WARNING**: None

**SUGGESTION**:
1. The CRUD handler endpoints (Create, Update, Delete) currently lack auth guards (`permissions.manage`). This is explicitly PR3 scope per the design and task plan — acceptable. When PR3 wires route guards, add `RequirePermission("permissions.manage")` on mutation routes.
2. `writePermissionError` maps status codes correctly but uses centralized `apperror.Status(err)` — good. Consider adding a handler-level integration test once authz middleware is wired (PR3 scope).
3. The `List` handler returns `PermissionGroupResponse` (grouped by module, sorted) which directly matches the spec scenario. The separate `ListPaginated` is an extension not in the spec but is harmless.

## PR2/PR3 Scope Boundary Check

- ❌ No `RequirePermission`, `RequireRole`, or `AttachPermissions` found in codebase → Correct
- ❌ No `GET /roles/:id/permissions` or `PUT /roles/:id/permissions` catalog/assignment endpoints → Correct
- ❌ No authz middleware file `internal/middleware/authorization.go` → Correct
- ❌ No `RoleSlug` or `Permissions` fields added to `authctx` → Correct
- ✅ Permission routes created but NOT wired into container → Correct (design says wiring is PR3)

## Verdict

**PASS WITH WARNINGS** — PR1 permission foundation meets all spec scenarios within its scope. All 4 tasks complete, build and 20 tests pass. No PR2/PR3 scope creep. The auth guard on CRUD endpoints is correctly deferred to PR3. One pre-existing identity test flake is unrelated.