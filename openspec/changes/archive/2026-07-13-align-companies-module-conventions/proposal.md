# Proposal: Align Companies Module Conventions

## Intent

Align the `companies` module with the documented vertical-slice and error-convention patterns without changing the public HTTP contract. This removes layout drift, fixes persistence translation bugs, and makes repository boundaries consistent and auditable.

## Scope

### In Scope
- Real-move the 7 slices under `internal/modules/companies/slices/`.
- Preserve routes, payloads, statuses, container/resolver APIs, and public compatibility aliases.
- Rename `core/error.go` to `core/errors.go`, add `queries/map_errors.go`, and normalize all company persistence errors.
- Fix `CountActivePrimaryDomains` and map zero-row update/delete outcomes to the correct not-found error.
- Audit every companies repository/GORM path and add exhaustive tests plus structural guards.

### Out of Scope
- Onboarding integration changes.
- Other modules.
- New company features or API shape changes.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `vertical-slice-modules`: companies now live under `slices/` with root wiring only.
- `companies-crud`: update/delete persistence outcomes remain stable while zero-row writes map to company not-found.
- `company-domains`: domain repositories use entity-specific error mapping and fix primary-domain counting behavior.
- `error-handling`: company repositories still return `error`, but the concrete value is module-owned `*apperror.AppError`.

## Approach

Move the slice packages mechanically, then update imports/wiring and keep compatibility aliases intact. Centralize repository translation in `queries/map_errors.go` with unary entity-specific mappers. Normalize all GORM `.Error` and `RowsAffected == 0` paths to module-owned AppErrors, preserving causes for unexpected failures.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/companies/slices/*` | Modified | Rehome all 7 slices |
| `internal/modules/companies/core/error.go` | Removed | Renamed to `core/errors.go` |
| `internal/modules/companies/queries/map_errors.go` | New | Unary entity-specific persistence mappers |
| `internal/modules/companies/**/repository.go` | Modified | Audit all GORM paths and zero-row handling |
| `internal/modules/companies/**/_test.go` | Modified | Add mapping and structural guard coverage |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Import/path churn breaks wiring | Med | Mechanical moves, preserve aliases, verify routes/resolver tests |
| Error mapping misses a repository path | Med | Exhaustive repo audit plus structural guards |
| Domain-count bug changes write behavior | Low | Targeted tests for create vs update semantics |

## Rollback Plan

Revert the package moves and filename rename as one unit, restore prior imports, and keep the public routes/aliases unchanged. If error translation regresses, temporarily restore the prior repository mapping while retaining the public API.

## Dependencies

- `platform/apperror` for module-owned `AppError` construction.

## Success Criteria

- [ ] All 7 companies slices exist under `internal/modules/companies/slices/`.
- [ ] Existing routes/statuses/payloads and compatibility aliases remain unchanged.
- [ ] Every companies repository/GORM path returns module-owned `*apperror.AppError` on non-nil persistence errors.
- [ ] Tests cover the `CountActivePrimaryDomains` fix and zero-row not-found mapping.
