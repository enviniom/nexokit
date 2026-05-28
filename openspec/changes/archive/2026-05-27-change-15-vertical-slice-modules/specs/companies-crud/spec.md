# Delta for Companies CRUD

## MODIFIED Requirements

### Requirement: Company CRUD endpoints

The system MUST expose `GET /api/v1/companies`, `GET /api/v1/companies/:id`, `PUT /api/v1/companies/:id`, and `DELETE /api/v1/companies/:id`. The system MUST NOT expose `POST /api/v1/companies`; new companies MUST be created through `POST /api/v1/onboarding/companies`. Responses MUST use the standard DTO envelope except successful DELETE responses, which MUST return HTTP 204 with no body. The `:id` parameter MUST reference the `PublicID`, never the internal `ID`. Each endpoint maps to one use-case slice: `list_companies`, `view_company`, `update_company`, `delete_company`.
(Previously: Same endpoint contract, implemented via flat handler/service/repository files)

#### Scenario: List companies

- GIVEN multiple companies exist
- WHEN `GET /api/v1/companies` is called by a root user
- THEN the response returns HTTP 200 with all companies

#### Scenario: View company

- GIVEN an existing company with `PublicID = "01HXYZ"`
- WHEN `GET /api/v1/companies/01HXYZ` is called
- THEN the response returns HTTP 200 with the company data

#### Scenario: Update company

- GIVEN an existing company
- WHEN `PUT /api/v1/companies/:id` is called with updated `name` or `status`
- THEN the response returns HTTP 200 with the updated company
- AND domain ownership MUST remain unchanged (domains are managed through dedicated endpoints)

#### Scenario: Delete company

- GIVEN an existing company
- WHEN `DELETE /api/v1/companies/:id` is called by a root user
- THEN the response returns HTTP 204 with an empty body
- AND the company is soft-deleted

### Requirement: Company model and migration

The system MUST define a `Company` model with fields: `ID uint` (primaryKey), `PublicID string` (char(26), uniqueIndex), `Name string` (not null), `Slug string` (uniqueIndex, not null), `Status string` (not null), `CreatedAt time.Time`, `UpdatedAt time.Time`, `DeletedAt gorm.DeletedAt` (index), `CreatedBy *uint` (index), `UpdatedBy *uint` (index). The model MUST have a `Domains` relationship to `[]CompanyDomain`. The system MUST NOT include `Domain` or `Subdomain` fields on the `Company` model — tenant hostname ownership is managed exclusively through the `company_domains` table. A Goose migration MUST create the `companies` table. Core model/DTO contracts MUST live in `internal/modules/companies/core` so endpoint slices can import them without importing root `companies`.
(Previously: Model defined in flat root `model.go`; unchanged behavior, moved to module-local `core` package to avoid import cycles)

#### Scenario: Migration creates companies table

- GIVEN the Goose migration for companies
- WHEN `make migrate-up` executes
- THEN the `companies` table exists with all columns and indexes matching the model

#### Scenario: Rollback drops companies table

- GIVEN the companies migration is applied
- WHEN `make migrate-down` executes
- THEN the `companies` table no longer exists

### Requirement: Direct company creation disabled

Company creation MUST be restricted to the onboarding flow. The system MUST NOT expose `POST /api/v1/companies` for any role, including `root`. No `create_company` slice exists.
(Previously: Same constraint; route absence unchanged)

#### Scenario: Direct create route is absent

- GIVEN any authenticated user, including root
- WHEN `POST /api/v1/companies` is called
- THEN the response returns HTTP 404

### Requirement: Company slug uniqueness

The `slug` field on companies MUST be unique across all companies, including soft-deleted ones.
(Previously: Same constraint; validation logic moves to onboarding slice)

#### Scenario: Duplicate slug rejected

- GIVEN a company with `slug = "acme"` exists
- WHEN `POST /api/v1/onboarding/companies` is called with `slug = "acme"`
- THEN the response returns HTTP 422 with a validation error on slug

#### Scenario: Slug available after permanent delete

- GIVEN a company with `slug = "acme"` was permanently deleted (hard delete)
- WHEN `POST /api/v1/onboarding/companies` is called with `slug = "acme"`
- THEN the response returns HTTP 201 (slug is available again)

### Requirement: Company status

Companies MUST have a `status` field supporting at least `active` and `inactive` values.
(Previously: Same constraint; status field unchanged)

#### Scenario: Deactivate company

- GIVEN an active company
- WHEN `PUT /api/v1/companies/:id` sets `status` to `inactive`
- THEN the company becomes inactive

#### Scenario: List excludes inactive companies by default

- GIVEN one active and one inactive company
- WHEN `GET /api/v1/companies` is called
- THEN the response includes only the active company unless a filter parameter requests inactive
