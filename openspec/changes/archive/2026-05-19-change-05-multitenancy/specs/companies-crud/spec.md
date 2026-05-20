# Companies CRUD Specification

## Purpose

Company model, migration, and full CRUD endpoints with root-only create enforcement.

## Requirements

### Requirement: Company model and migration

The system MUST define a `Company` model with fields: `ID uint` (primaryKey), `PublicID string` (char(26), uniqueIndex), `Name string` (not null), `Slug string` (uniqueIndex, not null), `Domain *string` (nullable), `Subdomain *string` (nullable), `Status string` (not null), `CreatedAt time.Time`, `UpdatedAt time.Time`, `DeletedAt gorm.DeletedAt` (index), `CreatedBy *uint` (index), `UpdatedBy *uint` (index). A Goose migration MUST create the `companies` table.

#### Scenario: Migration creates companies table

- GIVEN the Goose migration for companies
- WHEN `make migrate-up` executes
- THEN the `companies` table exists with all columns and indexes matching the model

#### Scenario: Rollback drops companies table

- GIVEN the companies migration is applied
- WHEN `make migrate-down` executes
- THEN the `companies` table no longer exists

### Requirement: Company CRUD endpoints

The system MUST expose `GET /api/v1/companies`, `POST /api/v1/companies`, `GET /api/v1/companies/:id`, `PUT /api/v1/companies/:id`, and `DELETE /api/v1/companies/:id`. All responses MUST use the standard DTO envelope. The `:id` parameter MUST reference the `PublicID`, never the internal `ID`.

#### Scenario: List companies

- GIVEN multiple companies exist
- WHEN `GET /api/v1/companies` is called by a root user
- THEN the response returns HTTP 200 with all companies

#### Scenario: Create company

- GIVEN valid company data including `name` and `slug`
- WHEN `POST /api/v1/companies` is called by a root user
- THEN the response returns HTTP 201 and `data` contains the created company with a generated `PublicID`

#### Scenario: Get company

- GIVEN an existing company with `PublicID = "01HXYZ"`
- WHEN `GET /api/v1/companies/01HXYZ` is called
- THEN the response returns HTTP 200 with the company data

#### Scenario: Update company

- GIVEN an existing company
- WHEN `PUT /api/v1/companies/:id` is called with updated `name` or `domain`
- THEN the response returns HTTP 200 with the updated company

#### Scenario: Delete company

- GIVEN an existing company
- WHEN `DELETE /api/v1/companies/:id` is called by a root user
- THEN the response returns HTTP 200 and the company is soft-deleted

### Requirement: Root-only company creation

Creating companies MUST be restricted to the `root` role. Admin and user roles MUST receive a 403 response when attempting to create a company.

#### Scenario: Root creates company

- GIVEN an authenticated root user
- WHEN `POST /api/v1/companies` is called with valid data
- THEN the company is created and the response returns HTTP 201

#### Scenario: Admin cannot create company

- GIVEN an authenticated admin user
- WHEN `POST /api/v1/companies` is called
- THEN the response returns HTTP 403 with `success: false`

#### Scenario: User cannot create company

- GIVEN an authenticated regular user
- WHEN `POST /api/v1/companies` is called
- THEN the response returns HTTP 403 with `success: false`

### Requirement: Company slug uniqueness

The `slug` field on companies MUST be unique across all companies, including soft-deleted ones.

#### Scenario: Duplicate slug rejected

- GIVEN a company with `slug = "acme"` exists
- WHEN `POST /api/v1/companies` is called with `slug = "acme"`
- THEN the response returns HTTP 422 with a validation error on slug

#### Scenario: Slug available after permanent delete

- GIVEN a company with `slug = "acme"` was permanently deleted (hard delete)
- WHEN `POST /api/v1/companies` is called with `slug = "acme"`
- THEN the response returns HTTP 201 (slug is available again)

### Requirement: Company status

Companies MUST have a `status` field supporting at least `active` and `inactive` values.

#### Scenario: Deactivate company

- GIVEN an active company
- WHEN `PUT /api/v1/companies/:id` sets `status` to `inactive`
- THEN the company becomes inactive

#### Scenario: List excludes inactive companies by default

- GIVEN one active and one inactive company
- WHEN `GET /api/v1/companies` is called
- THEN the response includes only the active company unless a filter parameter requests inactive