# Proposal: Reinforce platform boundary — no functional change

## Intent

`internal/platform/*` must contain only cross-application contracts and utilities. Currently, domain-specific elements leak into platform: a role-specific error message (`MsgRoleHasAssignedUsers`) sits in `platform/messages`, and module name constants (`ModuleUsers`, `ModuleRoles`, etc.) live in `platform/permissions`. This change moves domain language to owning modules and clarifies the platform boundary, with zero functional or API contract changes.

## Scope

### In Scope
- Move `MsgRoleHasAssignedUsers` from `platform/messages` to `modules/roles/core/`
- Create `ErrRoleHasAssignedUsers` sentinel in `modules/roles/core/error.go` (maps to HTTP 422)
- Redefine `apperror.ErrUnprocessable` as a generic 422 category without a domain-specific message
- Move `Module*` constants (`ModuleUsers`, `ModuleRoles`, `ModuleCompanies`, `ModuleSettings`, `ModuleAuth`, `ModulePermissions`) from `platform/permissions` to each owning module's `core/constants.go`
- Document platform boundary rules as part of this change

### Out of Scope
- Changes to handler logic, service behavior, or business rules
- Changes to HTTP response shape, status codes, or JSON envelope
- Rewriting module error handling beyond the single `MsgRoleHasAssignedUsers` case
- `Action*` constants, `Format()`, `Humanize*()`, `Registry` — these stay in platform (generic permission utilities)

## Capabilities

### New Capabilities
- `platform-boundary-rules`: Documented rules for what belongs in `platform` vs modules, including the generic-vs-domain classification of each platform subpackage

### Modified Capabilities
- `error-handling`: `ErrUnprocessable` becomes a generic 422 category sentinel (no domain-specific message); roles module owns its own 422 error sentinel
- `rbac-authorization`: Module name constants move out of `platform/permissions`; each module defines its own `Module<Name>` constant in `core/constants.go`

## Approach

**Approach 2 (Moderate)** from exploration:

1. `platform/messages`: Remove `MsgRoleHasAssignedUsers`; all other messages stay (generic API, validation, middleware)
2. `modules/roles/core/`: New `messages.go` with `MsgRoleHasAssignedUsers`; new `error.go` with `ErrRoleHasAssignedUsers` wrapping 422
3. `platform/apperror`: `ErrUnprocessable` keeps HTTP 422 mapping but loses the role-specific message; becomes a generic category sentinel
4. `platform/permissions`: Keep `Action*` constants, `Format()`, `HumanizeName()`, `HumanizeDescription()`, `DefaultDisplayOrder()`, `Register()`, `ListRegistered()` — all generic
5. `modules/*/core/constants.go`: Each module defines its own `Module<Name>` constant (e.g., `const ModuleUsers = "users"` in users module)
6. Update all import sites: middleware, module routes, roles service

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/platform/messages/messages.go` | Modified | Remove `MsgRoleHasAssignedUsers` |
| `internal/platform/apperror/apperror.go` | Modified | `ErrUnprocessable` becomes generic 422 sentinel |
| `internal/platform/permissions/constants.go` | Modified | Remove `Module*` constants |
| `internal/modules/roles/core/messages.go` | New | `MsgRoleHasAssignedUsers` |
| `internal/modules/roles/core/error.go` | New | `ErrRoleHasAssignedUsers` (422) |
| `internal/modules/roles/service.go` | Modified | Use new module-level error sentinel |
| `internal/modules/*/core/constants.go` | New/Modified | Each module defines its `Module<Name>` constant |
| `internal/modules/*/routes.go` | Modified | Import module-local constant instead of platform |
| `internal/middleware/authorization.go` | Unchanged | Still imports `platform/permissions` for `Action*` and registry |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `ErrUnprocessable` semantic change breaks roles delete behavior | Medium | Roles module must define `ErrRoleHasAssignedUsers` that maps to 422 and uses `MsgRoleHasAssignedUsers`; verify HTTP 422 response unchanged |
| Module constant duplication across 6 modules | Low | Explicitly acceptable per context rules (section 5: partial repetition avoids coupling) |
| Import path updates missed in some route files | Low | `go vet` and `go build` will catch; full module scan before commit |

## Rollback Plan

1. Revert the git commit — all changes are pure refactors with no data or migration impact
2. `MsgRoleHasAssignedUsers` returns to `platform/messages`
3. `Module*` constants return to `platform/permissions/constants.go`
4. `ErrUnprocessable` in `apperror` restores its original `Message: messages.MsgRoleHasAssignedUsers`
5. No database or config changes to revert

## Dependencies

- None — pure code reorganization, no external prerequisites

## Success Criteria

- [ ] `platform/messages` contains zero domain-specific messages
- [ ] `platform/permissions/constants.go` contains zero `Module*` constants
- [ ] `apperror.ErrUnprocessable` has no domain-specific message
- [ ] `modules/roles/core/error.go` exists with `ErrRoleHasAssignedUsers` mapping to 422
- [ ] Each module (`users`, `roles`, `companies`, `settings`, `auth`, `permissions`) defines its own `Module<Name>` constant in `core/constants.go`
- [ ] `go build ./...` and `go vet ./...` pass with zero errors
- [ ] All existing tests pass without modification to test expectations
- [ ] HTTP response shape and status codes remain identical (verified by existing integration tests)
