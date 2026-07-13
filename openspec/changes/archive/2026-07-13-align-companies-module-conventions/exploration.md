# Exploration: align-companies-module-conventions

## Current State
`internal/modules/companies` is still a flat root module. Slice packages (`list_companies/`, `view_company/`, `update_company/`, `delete_company/`, `list_company_domains/`, `create_company_domain/`, `update_company_domain/`) live directly under the module root instead of `slices/`, even though the docs require `slices/` and the auth module is already the aligned reference.

The public API is mostly stable: `container.go` wires all handlers, `routes.go` mounts `/companies`, and root aliases in `model.go`/`dto.go` preserve older import paths. The big gap is persistence translation: repositories still return raw GORM errors or inline `ErrRecordNotFound`/unique checks, and there is no `queries/map_errors.go` or `core/errors.go` naming alignment.

## Concise Companies Inventory
- Root wiring: `container.go`, `routes.go`, `resolver.go`
- Compatibility aliases: `model.go`, `dto.go`
- Shared domain: `core/constants.go`, `core/model.go`, `core/dto.go`, `core/error.go`
- Use cases: `list_companies`, `view_company`, `update_company`, `delete_company`, `list_company_domains`, `create_company_domain`, `update_company_domain`
- Reusable queries: `queries/get_company_by_public_id.go`, `queries/get_company_domain_by_domain.go`, `queries/count_active_primary_domains.go`
- Tests: core DTO/model/error tests, query tests, per-slice handler/service/repository tests, `routes_absence_test.go`, `resolver_test.go`, `migration_test.go`

## Deviations Matrix
| Area | Current | Required |
|---|---|---|
| Slice layout | Flat root packages | `slices/` (flat here, because one module with 7 use cases) |
| Container | Wires slices directly; exposes `Resolver()` and `RegisterRoutes()` | Same behavior, but slice imports should move under `slices/` |
| Routes | `/companies` with current methods/statuses | Preserve paths/methods/statuses exactly |
| Compatibility aliases | Root `model.go`/`dto.go` aliases exist | Preserve aliases during move |
| Repository translation | Mixed raw GORM, inline not-found, inline unique checks | All persistence errors through entity-specific `queries/map_errors.go` |
| `core/errors.go` | `core/error.go` exists instead | Rename/normalize; module-owned AppErrors and constructors belong there |
| Reusable queries | Present, but no mapper file | Keep queries; add mapper file beside them |
| Guards/tests | Query tests exist; no repository-boundary structural guard | Add structural guard coverage for raw `.Error`, `apperror` leakage, and mapper use |

## Affected Areas
- `internal/modules/companies/container.go` — import paths and slice wiring need rehome under `slices/`.
- `internal/modules/companies/routes.go` — route registration should stay stable; verify no accidental path/method drift.
- `internal/modules/companies/resolver.go` — resolver is part of the middleware contract (`CompanyResolver`).
- `internal/modules/companies/model.go`, `dto.go` — compatibility aliases that should remain intact.
- `internal/modules/companies/core/error.go` — must become `core/errors.go` per docs.
- `internal/modules/companies/queries/*.go` — add `map_errors.go`; query files stay reusable and tested.
- `internal/modules/companies/*/repository.go` — every repository method currently needs translation cleanup.
- `internal/modules/companies/*_test.go` — add/adjust boundary guards and repository mapping tests.
- `internal/app/container.go`, `internal/middleware/tenant.go` — resolver injection and tenant resolution dependencies must continue to work.

## Approaches
1. **Mechanical rehome + centralized mapping** — move slice packages under `slices/`, rename `core/error.go` to `core/errors.go`, add `queries/map_errors.go`, and normalize repository translation.
   - Pros: matches docs; keeps public HTTP contract stable; reviewable.
   - Cons: many path/import edits; must be careful with current alias exports.
   - Effort: Medium

2. **Compatibility shims first** — add new `slices/` and mapper file while leaving old packages as wrappers for a cycle.
   - Pros: less immediate churn.
   - Cons: duplicates structure; delays compliance; adds indirection without value.
   - Effort: Medium

## Recommendation
Prefer the mechanical rehome. The module already has stable HTTP behavior and the auth migration shows the right pattern: preserve routes/payloads, move structure, and centralize persistence mapping. Do not introduce shims unless a rename risk becomes unmanageable.

## Open Product / Architecture Decisions
- Whether `CountActivePrimaryDomains` should continue to ignore the exclude argument in create flows (safe) and honor it in update flows (required).
- Whether delete/update should map zero-row `RowsAffected` to `ErrCompanyNotFound` or another module sentinel when the row disappears between lookup and mutation.
- Whether to add module-owned internal persistence AppErrors for companies, or keep business sentinels only and map unexpected errors at the repository edge.

## Workload Forecast
- Auth reference was ~120–180 authored lines for a similar alignment.
- Companies likely exceeds that because it has 7 slices, 3 reusable queries, root aliases, and one confirmed behavior bug (`CountActivePrimaryDomains` exclude path ignored in `update_company_domain`).
- Estimated authored lines: ~180–260.
- Recommendation: this can still fit one PR if kept mechanical and focused, but it is close enough to the review budget that splitting is defensible if mapper/error refactors expand.

## Risks
- Raw GORM errors still leak from multiple repositories.
- The update-domain primary-count bug can change behavior if not fixed deliberately.
- Route/status regressions are easy to introduce while moving packages.
- Root compatibility aliases must not disappear during the rehome.

## Ready for Proposal
Yes — the scope is clear enough for `sdd-propose` once you are ready to lock the change plan.
