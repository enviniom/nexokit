# Proposal: Vertical Slice Modules

## Intent

Replace the current horizontal layered architecture (handler/service/repository as flat files per module) with vertical slices organized by use case. Each slice co-locates its handler, service, repository, and tests. This reduces cognitive load when working on a single feature, enables independent testing per use case, and prevents the "god file" problem as modules grow.

## Scope

### In Scope
- Restructure `companies` module as pilot: 7 use-case slices (one per existing endpoint)
- Module root retains only cross-cutting files: `routes.go`, `container.go`; shared models/DTOs/contracts move to `shared/`
- Module `container.go` becomes composition root — wires slices, root container calls only module containers
- Models/DTOs/contracts shared via `internal/modules/companies/shared` — no duplication per slice and no root import cycle
- Incremental migration pattern documented for future modules

### Out of Scope
- Migrating other modules (auth, users, roles, permissions, onboarding) — deferred to future changes
- CQRS separation (commands vs queries) — over-engineering for current complexity
- Changing external API contracts or endpoint paths
- Module generator tool updates — deferred
- `create_company` slice — no public `POST /companies` endpoint exists; company creation is handled by onboarding

## Capabilities

### New Capabilities
- `vertical-slice-modules`: Architecture pattern introducing use-case slice organization within modules. Each module has a `container.go` composition root. Shared models/DTOs/contracts live in a module-local `shared` package when needed, so root wiring can import slices while slices import shared types without cycles. Slices co-locate handler + service + repository + tests. New modules adopt this pattern going forward.

### Modified Capabilities
- `companies-crud`: Requirements unchanged, but implementation moves from flat files to use-case slices. Endpoints, validation, and behavior remain identical.
- `company-domains`: Requirements unchanged, but implementation moves into dedicated slices within companies module.
- `app-orchestration`: Container wiring changes — root `container.go` calls module containers instead of individual constructors. Modules register via their own `container.go`.

## Approach

**Approach 1 from exploration, corrected for Go imports**: Use-case slices with shared models in a `shared` subpackage. Mixed architecture accepted — companies pilot uses vertical slices; existing modules remain flat. New modules adopt vertical slices going forward.

```
internal/modules/companies/
├── routes.go             # Cross-cutting route registration
├── container.go          # Composition root: wires all slices
├── shared/               # Shared entities, DTOs, constants, contracts
├── list_companies/       # GET /companies
├── view_company/         # GET /companies/:id
├── update_company/       # PUT /companies/:id
├── delete_company/       # DELETE /companies/:id
├── list_company_domains/     # GET /companies/:id/domains
├── create_company_domain/    # POST /companies/:id/domains
└── update_company_domain/    # PUT /companies/:id/domains/:domain_id
```

- Each slice has its own thin repository struct with only the methods it needs, sharing the same `*gorm.DB`
- Repository interfaces and shared models/DTOs live in `companies/shared` so slices never import root `companies`
- Module `container.go` instantiates all slices, exposes handlers for route registration
- Root `container.go` simplifies: calls `companies.NewContainer(db)` instead of wiring each layer

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/companies/` | Restructured | Flat files → 7 use-case slices |
| `internal/modules/companies/shared/` | New | Shared models, DTOs, constants, and contracts for slices/middleware |
| `internal/modules/companies/container.go` | New | Module composition root |
| `internal/app/container.go` | Simplified | Calls module container, removes per-layer wiring for companies |
| `internal/middleware/` | Modified | `AllowRootGlobalScope` adapts to new companies repo interface location |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Repository method fragmentation across slices | Medium | Each slice owns only methods it needs; shared `*gorm.DB` via constructor |
| Go import cycle between root and slices | High | Root `companies` imports slices; slices import `companies/shared`, never root `companies` |
| Import path cascade in tests and middleware | High | Audit all `companies.` references before deletion; use build errors as checklist |
| Module container becomes service locator | Low | Enforce: instantiate → wire → expose only; no business logic |
| Review budget exceeded (companies ~1700 LOC) | High | Split into 2-3 chained PRs; 800-line budget from preflight |

## Rollback Plan

1. Keep old flat files alongside new slices during PR 1 (no deletion)
2. If issues arise after wiring: revert route registration to old handlers, delete new slice directories
3. Reverting the migration change restores the flat structure entirely
4. No database changes — models and migrations remain identical

## Dependencies

- SDD init completed (testing capabilities, registry, persistence)
- Exploration phase completed (3 approaches evaluated, companies identified as pilot)

## Success Criteria

- [ ] Companies module compiles and all existing tests pass after restructuring
- [ ] Root container calls `companies.NewContainer(db)` instead of wiring individual layers
- [ ] Each use-case slice has its own handler, service, repository, and tests
- [ ] No cross-module imports within companies slices
- [ ] API endpoints return identical responses (no behavioral change)
- [ ] Migration pattern documented for future module adoption
