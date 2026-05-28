# Tasks: Vertical Slice Modules

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 800–1200 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: core + slices + container (no deletion) → PR 2: wire app + delete old files → PR 3: tests + verification |
| Delivery strategy | ask-on-risk |
| Chain strategy | feature-branch-chain |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Create `core/`, 7 slices, and `container.go`; old files still compile | PR 1 | base: feature/change-15-vertical-slice-modules; includes slice unit tests; zero behavioral change |
| 2 | Switch app/container.go to companies.NewContainer; delete flat files | PR 2 | base: PR 1 branch; routes + middleware adaptation |
| 3 | Redistribute tests to slices; full verification | PR 3 | base: PR 2 branch; go test ./... must pass |

## Phase 1: Foundation / Container

- [x] 1.1 Create `internal/modules/companies/core/` and move shared models, DTOs, constants, and contracts there.
- [x] 1.2 Create `internal/modules/companies/container.go` with `NewContainer(db)` wiring placeholder slice handlers.
- [x] 1.3 Add `Resolver()` method to `companies.Container` returning `middleware.CompanyResolver` via `companies/core` contracts.
- [x] 1.4 Add `RegisterRoutes(group, mw)` method to `companies.Container` delegating to `routes.go`.

## Phase 2: Slice Implementation — Company CRUD

- [x] 2.1 Create `internal/modules/companies/list_companies/` with handler, service, repository, and table-driven tests for `GET /companies` list/filter behavior.
- [x] 2.2 Create `internal/modules/companies/view_company/` with handler, service, repository, and tests for `GET /companies/:id` (GetByPublicID, includes domains collection).
- [x] 2.3 Create `internal/modules/companies/update_company/` with handler, service, repository, and tests for `PUT /companies/:id` (name, status; preserves domains).
- [x] 2.4 Create `internal/modules/companies/delete_company/` with handler, service, repository, and tests for `DELETE /companies/:id` (soft delete, 204 response).

## Phase 3: Slice Implementation — Domain Administration

- [x] 3.1 Create `internal/modules/companies/list_company_domains/` with handler, service, repository, and tests for `GET /companies/:id/domains`.
- [x] 3.2 Create `internal/modules/companies/create_company_domain/` with handler, service, repository, and tests for `POST /companies/:id/domains` (validation: status/kind enums, uniqueness, active-primary constraint).
- [x] 3.3 Create `internal/modules/companies/update_company_domain/` with handler, service, repository, and tests for `PUT /companies/:id/domains/:domain_id` (ownership check, status change, redirect flag).

## Phase 4: Wiring and Integration

- [x] 4.1 Update `internal/modules/companies/container.go` to wire all 7 slices with real handlers, services, and repositories.
- [x] 4.2 Update `internal/modules/companies/routes.go` to register handlers from `*Container`; intentionally omit `POST /companies` and `DELETE /companies/:id/domains/:id`.
- [x] 4.3 Update `internal/app/container.go`: replace `companiesHandler`/`companiesRepo` fields with `Companies *companies.Container`; middleware uses `c.Companies.Resolver()`.
- [x] 4.4 Delete `internal/modules/companies/{model,dto,handler,service,repository}.go` after core moves and new structure build pass.

## Phase 5: Testing and Verification

- [x] 5.1 Move handler tests to corresponding slices; run `go test ./internal/modules/companies/...`.
- [x] 5.2 Move service tests to corresponding slices; verify table-driven cases cover spec scenarios (list excludes inactive, update preserves domains).
- [x] 5.3 Move repository tests to corresponding slices; verify list filters, soft delete, domain uniqueness, resolver behavior.
- [x] 5.4 Run `go test ./...` — all tests must pass; verify `POST /api/v1/companies` returns 404, `DELETE /api/v1/companies/:id/domains/:id` returns 404.
- [x] 5.5 Verify `go build ./...` succeeds with no unused import errors after old file deletion.
- [x] 5.6 Verify no slice imports root `internal/modules/companies`; slices import `internal/modules/companies/core` for shared models/DTOs/contracts.

## Phase 6: Review Refinement — Core vs Queries

- [x] 6.1 Preserve `internal/modules/companies/core` ownership for shared models, DTOs/contracts, enums/constants, errors, and shared non-query values.
- [x] 6.2 Create `internal/modules/companies/queries/` and move repeated repository query methods (company lookup by public ID, domain lookup by domain, active primary count) into one-query-per-file units.
- [x] 6.3 Add one `_test.go` file per query file in `queries/`.
- [x] 6.4 Update slice repositories to delegate duplicated query logic to `queries` while keeping endpoint-specific behavior inside slice repositories.
- [x] 6.5 Ensure every slice source file (`handler.go`, `service.go`, `repository.go`) has a corresponding `_test.go` file; thin repository wrappers include comments clarifying query behavior coverage in `queries` tests and wrapper delegation intent.
