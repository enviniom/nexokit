# Design: Align Companies Module Conventions

## Technical Approach

Mechanically rehome all seven slices, then make repository persistence translation total without changing HTTP or composition contracts. Repositories keep `error` signatures but return module-owned `*apperror.AppError` values for every non-nil GORM outcome.

## Architecture Decisions

| Choice | Alternative | Rationale |
|---|---|---|
| Real moves; no shims | Wrappers | Matches the vertical-slice convention and avoids duplicate packages. |
| Unary `MapCompanyError(error) error` and `MapCompanyDomainError(error) error` | Generic/repository-selected mapper | Entity policy remains centralized and auditable. |
| Function-scoped recursive AST dataflow guard | Auth guard’s file-wide selector keys | Prevents equal `result.Error` selectors in different functions from falsely satisfying each other. |
| One ordered PR, ≤800 authored lines | Split PRs | Scope is cohesive; checkpoint before repository migration protects the budget. |

## Package Moves and Wiring

| From `companies/` | To `companies/slices/` |
|---|---|
| `list_companies` | `list_companies` |
| `view_company` | `view_company` |
| `update_company` | `update_company` |
| `delete_company` | `delete_company` |
| `list_company_domains` | `list_company_domains` |
| `create_company_domain` | `create_company_domain` |
| `update_company_domain` | `update_company_domain` |

Move each complete package including tests; only `container.go` imports change to `/slices/...`. Preserve `Container` fields, `NewContainer`, `Resolver()`, `RegisterRoutes`, all seven routes/middleware/statuses/payloads, root aliases in `model.go`/`dto.go`, and `Resolver.FindByPublicIDOrSlug`/`ResolveHost` used by app tenant middleware.

## Error Contracts

`core/error.go` becomes `core/errors.go`. Preserve six sentinels/codes (`company_not_found`, `company_domain_not_found`, `company_domain_duplicate`, `primary_domain_exists`, `company_domain_does_not_belong`, `company_slug_duplicate`) and their `NotFound`/`Conflict`/`Validation` constructors. Add `CodeCompanyPersistence`, `CodeCompanyDomainPersistence` and `CompanyPersistenceError(cause)`, `CompanyDomainPersistenceError(cause)` using `apperror.Internal`; causes remain unwrap-able.

Both mappers return nil for nil. Company/domain not-found maps to the matching sentinel; unique violations map to duplicate slug/domain; an existing module AppError remains known; every unknown error becomes the corresponding 500 persistence AppError.

## Exhaustive Repository/GORM Inventory

K/U means known/unknown mapping above.

| Method | GORM operation | Mapper; outcome | RowsAffected |
|---|---|---|---|
| `list.List` | `Count` | Company; U | N/A |
| `list.List` | `Find` | Company; U | empty success |
| `view.GetByPublicID` | preload + `First` | Company; K/U | not-found via error |
| `update.GetByPublicID` | query `First` | Company; K/U | not-found via error |
| `update.GetBySlugIncludingDeleted` | `Unscoped.First` | Company; K/U | not-found via error |
| `update.Update` | scoped `Updates` | Company; K/U | 0 → company not-found |
| `delete.GetByPublicID` | query `First` | Company; K/U | not-found via error |
| `delete.Delete` | `Delete` | Company; K/U | 0 → company not-found |
| `domains.List.GetByPublicID` | query `First` | Company; K/U | not-found via error |
| `domains.ListDomains` | `Find` | Domain; U | empty success |
| `domains.Create.GetByPublicID` | query `First` | Company; K/U | not-found via error |
| `domains.Create.GetDomainByDomain` | query `First` | Domain; K/U | not-found via error |
| `domains.Create.CountActivePrimaryDomains` | `Count` | Domain; U | N/A |
| `domains.Create.CreateDomain` | `Create` | Domain; K/U | N/A |
| `domains.Update.GetByPublicID` | query `First` | Company; K/U | not-found via error |
| `domains.Update.GetDomainByPublicID` | `First` | Domain; K/U | not-found via error |
| `domains.Update.GetDomainByDomain` | query `First` | Domain; K/U | not-found via error |
| `domains.Update.CountActivePrimaryDomains` | `Count` | Domain; U | N/A |
| `domains.Update.UpdateDomain` | scoped `Updates` | Domain; K/U | 0 → domain not-found |

Reusable queries remain raw internal helpers; repository adapters map their errors. `CountActivePrimaryDomains` keeps `public_id <> ?` only when exclusion is non-empty: create passes `""`; update passes the current `CompanyDomain.PublicID`.

## Structural Guard and Tests

Recursively discover `slices/**/repository.go`. For each `FuncDecl`, create independent raw/mapped variable and selector-position state; recursively inspect nested blocks. Detect direct `.Error`, `.Error` held in variables, multi-result errors, and single-result `err := persistenceCall()`. A mapper satisfies only the exact AST expression/variable in that function. RED fixtures include direct, variable-held, single-result, nested repositories, and two functions sharing `result.Error` where only one maps. Also reject `apperror` in repository interfaces.

Add mapper/core table tests; repository tests for every row, wrapped known/unknown causes, unique constraints, and zero rows; count tests for create/no exclusion and update/self exclusion; container, aliases, resolver, routes, full `go test ./...`, and `go build ./...`.

## Threat Matrix

N/A — no routing behavior, shell, subprocess, VCS/PR automation, executable classification, or process-integration boundary changes.

## Rollout, Forecast, Rollback

Sequence: RED tests/guard → errors/mappers → seven moves+wiring → repository migration → focused/full verification. Forecast: 650–800 authored changed lines, medium budget risk; stop and re-estimate before migration if tests/moves exceed 400. No migration or flag. Roll back the single PR atomically: restore paths/imports and `error.go`, remove mappers/guards, and restore prior repository operations; no data rollback.

## Open Questions

None.

## Apply Evidence Notes

- Baseline and post-move inspection confirm the seven existing route registrations, `:id` PublicID semantics, response envelopes, and statuses remain owned by `routes.go` and unchanged.
- `Container`, `Resolver()`, root model/DTO aliases, and tenant calls to `FindByPublicIDOrSlug` and `ResolveHost` remain root-level compatibility contracts.
- The 18 repository-method paths and 19 GORM operations listed above were reconciled against the seven moved repositories; `Find` remains an empty-success operation, while zero-row scoped writes map through the corresponding entity mapper.
