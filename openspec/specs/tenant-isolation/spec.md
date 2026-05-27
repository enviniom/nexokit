# Tenant Isolation Specification

## Purpose

TenantContext struct, GORM scope helpers, and tenant middleware that enforce per-company data isolation for private and public routes.

## Requirements

### Requirement: TenantContext struct

The system MUST define a `TenantContext` struct in `internal/platform/tenant/` with fields `CompanyID uint`, `CompanySlug string`, and `IsRootScope bool`. TenantContext MUST be immutable after creation — only tenant middleware SHALL create instances.

#### Scenario: Non-root user gets scoped context

- GIVEN an authenticated admin with `company_id = 5`
- WHEN tenant middleware resolves the context
- THEN `TenantContext{CompanyID: 5, IsRootScope: false}` is set

#### Scenario: Root user gets global scope

- GIVEN an authenticated root user with `company_id` null
- WHEN tenant middleware resolves the context
- THEN `TenantContext{IsRootScope: true, CompanyID: 0}` is set

#### Scenario: Root user scoped via header

- GIVEN an authenticated root user and header `X-Company-ID: 7`
- WHEN tenant middleware resolves the context
- THEN `TenantContext{CompanyID: 7, IsRootScope: false}` is set

#### Scenario: Invalid X-Company-ID for root

- GIVEN an authenticated root user and header `X-Company-ID: 99999` (non-existent)
- WHEN tenant middleware resolves the context
- THEN the response returns HTTP 400

### Requirement: GORM tenant scope helpers

The system MUST provide `WithCompany(db, companyID)` adding `WHERE company_id = ?`, and `ApplyTenantScope(db, TenantContext)` that applies `WithCompany` when `IsRootScope` is false or returns `db` unchanged when `IsRootScope` is true.

#### Scenario: Scoped query filters by company_id

- GIVEN `TenantContext{CompanyID: 3, IsRootScope: false}`
- WHEN `ApplyTenantScope(db, ctx)` is called
- THEN the query includes `WHERE company_id = 3`

#### Scenario: Root scope returns unfiltered query

- GIVEN `TenantContext{IsRootScope: true}`
- WHEN `ApplyTenantScope(db, ctx)` is called
- THEN the query has no company_id filter

### Requirement: Tenant middleware for private routes

The system MUST provide tenant middleware that resolves TenantContext from the authenticated user's `company_id`. For root users with `X-Company-ID` header, scope to that company; otherwise `IsRootScope=true`.

#### Scenario: Private route sets tenant from admin user

- GIVEN an authenticated admin with `company_id = 4`
- WHEN tenant middleware runs on a private route
- THEN Gin context has `TenantContext{CompanyID: 4, IsRootScope: false}`

#### Scenario: Private route root without header

- GIVEN an authenticated root user with no `X-Company-ID`
- THEN `TenantContext{IsRootScope: true}` is set

### Requirement: Tenant middleware for public routes

The system MUST provide tenant middleware for unauthenticated routes resolving company by: (1) Host header normalized and matched exactly against `company_domains.domain` where `status = active`, (2) `X-Tenant` header only in development. The system MUST NOT fallback from first subdomain to company slug for production tenant resolution. Results MUST be cached with a short TTL. When a matching domain has `redirect_to_primary = true`, the system MUST return a permanent redirect (308) to the company's active primary domain, preserving path and query string.

#### Scenario: Domain resolves to company

- GIVEN `Host: store.acme.com` and an active `company_domains` row with `domain = "store.acme.com"` for company A
- THEN `TenantContext.CompanyID` matches company A

#### Scenario: Subdomain does NOT resolve to company slug (production)

- GIVEN `Host: acme.app.nexokit.com` and a company with `slug = "acme"`
- AND no `company_domains` row exists for `acme.app.nexokit.com`
- THEN tenant resolution MUST NOT resolve to the company by slug
- AND the response returns HTTP 404

#### Scenario: Explicit www alias resolves tenant

- GIVEN active `company_domains` rows for both `acme.com` and `www.acme.com` belonging to company A
- WHEN `Host: www.acme.com`
- THEN `TenantContext.CompanyID` matches company A

#### Scenario: Redirect alias returns 308 to primary

- GIVEN company A has active primary domain `acme.com`
- AND active alias domain `www.acme.com` with `redirect_to_primary = true`
- WHEN `Host: www.acme.com`
- THEN the response MUST be HTTP 308 redirecting to `https://acme.com` with preserved path and query

#### Scenario: X-Tenant header in development only

- GIVEN `APP_ENV=development` and header `X-Tenant: acme`
- THEN `TenantContext.CompanySlug` is "acme"
- GIVEN `APP_ENV=production` and header `X-Tenant: acme`
- THEN the header is ignored

#### Scenario: No tenant resolved on public route

- GIVEN no Host/domain/subdomain/header matches any company
- THEN the response returns HTTP 404

### Requirement: Cross-tenant access protection

Non-root queries MUST filter by `company_id`. A non-root user MUST NOT read or modify another company's data.

#### Scenario: Admin cannot read cross-tenant data

- GIVEN admin with `company_id = 1`
- WHEN `GET /api/v1/users` is called
- THEN only users with `company_id = 1` are returned

#### Scenario: Admin cannot modify cross-tenant resource

- GIVEN admin with `company_id = 1` targeting a user with `company_id = 2`
- WHEN `PUT /api/v1/users/:id` is called
- THEN the response returns HTTP 404

### Requirement: Multitenant model pattern documentation

The system MUST include documentation in `internal/platform/tenant/` explaining how to add `company_id` to new models and apply `ApplyTenantScope`.

#### Scenario: Developer reads multitenant guide

- GIVEN a developer adding a new model
- WHEN they check `internal/platform/tenant/doc.go`
- THEN they find instructions for adding `company_id` and tenant-scoping queries