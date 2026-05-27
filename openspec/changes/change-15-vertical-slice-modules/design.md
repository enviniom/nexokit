# Design: Vertical Slice Modules

Companies becomes the pilot vertical-slice module. Existing modules stay flat. API paths, DTO envelopes, model schema, authorization intent, tenant resolution, and domain behavior remain unchanged.

## Technical Approach

Replace `internal/modules/companies/{handler,service,repository}.go` with one package per existing public endpoint. Shared models/DTOs/contracts move to `internal/modules/companies/shared` so root `companies` can import slices for wiring while slices import shared types without a Go import cycle. There is no `create_company` slice because there is no public `POST /api/v1/companies`; company creation remains owned by onboarding. The company detail endpoint becomes `view_company`, matching `companies:view` permission semantics already used by `routes.go` for `GET /companies/:id`.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Slice boundary | One existing endpoint = one use case/slice | CRUD groups; one `domains/` slice | Matches route surface and keeps review/test ownership obvious. |
| Company detail name | `view_company` | `get_company` | Aligns code vocabulary with View permission semantics, not persistence naming. |
| No create slice | Omit `create_company` entirely | Keep unrouted create package | Avoids dead API surface; onboarding remains the creation boundary. |
| Shared types | `companies/shared` holds models, DTOs/contracts, enums/constants, errors, and shared values that are not query logic | Keep shared files in root `companies`; duplicate per slice | Avoids Go import cycles while preserving the companies module boundary. |
| Query reuse | Repeated repository query methods move to `companies/queries` (one query per file + one test file per query file) | Keep duplicated methods in each slice repo | Removes repeated SQL/GORM logic while keeping endpoint-aligned slice repos. |
| Repositories | Narrow per-slice repos over shared `*gorm.DB` and `queries` helpers; shared contracts stay in `companies/shared` | God repository; duplicated models | Keeps slice ownership clear while deduplicating repeated query logic. |
| Container | `app.NewContainer` calls `companies.NewContainer(db)` only | Root imports slice packages | Keeps root orchestration module-level and prevents service-locator behavior. |

## Data Flow

```text
app.NewContainer ──→ companies.NewContainer(db)
                         ├─ list_companies
                         ├─ view_company  (GET /companies/:id, companies:view)
                         ├─ update_company
                         ├─ delete_company
                         ├─ list_company_domains
                         ├─ create_company_domain
                         ├─ update_company_domain
                         └─ resolver repo for tenant middleware

RegisterModules ──→ Companies.Register(container)
HTTP route ──→ slice handler ──→ slice service ──→ slice repo ──→ GORM
Tenant middleware ──→ container resolver ─────────────────────→ GORM
```

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/companies/shared/model.go` | Create | Move `Company`, `CompanyDomain`, constants, and shared model contracts here. |
| `internal/modules/companies/shared/dto.go` | Create | Move shared request/response DTOs, validators, normalization, mappers, shared errors here. |
| `internal/modules/companies/shared/error.go` | Create | Shared companies errors (`duplicate domain`, `active primary exists`, `domain ownership`). |
| `internal/modules/companies/shared/contracts.go` | Create | Shared interfaces/contracts needed by slices or middleware. |
| `internal/modules/companies/queries/*.go` | Create | Query helper package for repeated repository lookup/count logic; one file per query with matching test file. |
| `internal/modules/companies/routes.go` | Modify | Register unchanged routes from `*Container`; `GET /:id` uses `ViewCompany`; omit `POST /companies` and domain DELETE. |
| `internal/modules/companies/container.go` | Create | Wires all seven slices plus resolver; exposes concrete handlers; no business logic. |
| `internal/modules/companies/list_companies/` | Create | `GET /companies`. |
| `internal/modules/companies/view_company/` | Create | `GET /companies/:id`; includes domains and honors View permission semantics. |
| `internal/modules/companies/update_company/` | Create | `PUT /companies/:id`; does not manage domains. |
| `internal/modules/companies/delete_company/` | Create | `DELETE /companies/:id`; soft delete with 204 response. |
| `internal/modules/companies/list_company_domains/` | Create | `GET /companies/:id/domains`. |
| `internal/modules/companies/create_company_domain/` | Create | `POST /companies/:id/domains`. |
| `internal/modules/companies/update_company_domain/` | Create | `PUT /companies/:id/domains/:domain_id`. |
| `internal/modules/companies/{model,dto,handler,service,repository}.go` | Delete/Move | Shared model/DTO contents move to `shared`; flat layers removed after route/test parity. |
| `internal/app/container.go` | Modify | Replace companies handler/repo fields with `Companies *companies.Container`; middleware uses the container resolver. |

## Interfaces / Contracts

`companies.Container` exposes handlers by slice name and a resolver implementing `middleware.CompanyResolver`. Slices MUST import shared types/contracts from `internal/modules/companies/shared`, never root `companies`, and MUST NOT import sibling slices. Repositories MAY import `internal/modules/companies/queries` for duplicated query methods. Root container MUST NOT import `internal/modules/companies/*` subpackages except the root `companies` package.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Each slice handler/service, validation, error mapping, View permission route | Table-driven Go tests with local fakes; assert status/body and public ID parameters. |
| Repository | Per-slice query/update behavior, soft delete, domain uniqueness, active primary, resolver | Move existing SQLite/GORM tests beside owning slice; keep deterministic in-memory DB. |
| Integration | Root container delegation, route surface unchanged, missing `POST /companies`, missing domain DELETE | `httptest` plus `go test ./...`; no external DB dependency. |

## Migration / Rollout

1. Add module `container.go`, `shared/` contracts/models/DTOs, resolver, and seven slice packages while old flat files still compile.
2. Switch `internal/app/container.go` and `routes.go` to the module container.
3. Move tests beside owning slices, verify `go test ./...`, then delete flat layer files.
4. Rollback by reverting app wiring/routes to old handler and restoring flat files; no database rollback required.

## Open Questions

- [ ] None.
