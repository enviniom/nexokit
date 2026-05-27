# Delta for Company Domains

## MODIFIED Requirements

### Requirement: Company Domains Model

The system MUST store company-owned hostnames in a `company_domains` table. The system MUST NOT use `companies.domain` or `companies.subdomain` as tenant host sources of truth. Each company domain MUST belong to exactly one company. Each company domain MUST have a globally unique `domain` value. Each company domain MUST have a `status`. Each company domain MUST have a `kind`. The system MUST initially support domain statuses: `active`, `inactive`, and `pending_verification`. The system MUST initially support domain kinds: `primary`, `alias`, and `technical`. The system MUST support redirect behavior with `redirect_to_primary`. A company SHOULD have at most one active primary domain. A domain with `status != active` MUST NOT resolve a public tenant request.
(Previously: Same model contract; model moves to `internal/modules/companies/shared`, domain slices own their repositories)

#### Scenario: Domain model schema

- GIVEN the `company_domains` table exists
- WHEN a row is inserted
- THEN it MUST have `company_id`, `domain` (unique), `status`, `kind`, and `redirect_to_primary`

#### Scenario: Unsupported status rejected

- GIVEN a domain creation request
- WHEN `status` is not one of `active`, `inactive`, `pending_verification`
- THEN the request MUST fail validation

#### Scenario: Unsupported kind rejected

- GIVEN a domain creation request
- WHEN `kind` is not one of `primary`, `alias`, `technical`
- THEN the request MUST fail validation

### Requirement: Domain Lifecycle

The system SHOULD deactivate domains instead of soft-deleting them. The system MUST allow a deactivated domain to be reactivated for the same owning company. The system MUST keep `domain` globally unique across all statuses until an explicit future release/transfer operation exists. The system MUST NOT implicitly transfer an inactive domain to another company.
(Previously: Same lifecycle rules; implementation moves to `list_company_domains/`, `create_company_domain/`, `update_company_domain/` slices)

#### Scenario: Deactivate and reactivate domain

- GIVEN an active domain for company A
- WHEN root updates status to `inactive`
- THEN the domain remains owned by company A with `status = inactive`
- WHEN root later updates status to `active`
- THEN the domain resolves tenants again

#### Scenario: Inactive domain not transferable

- GIVEN an inactive domain owned by company A
- WHEN an attempt is made to assign it to company B
- THEN the system MUST reject the operation

### Requirement: Onboarding Domain Creation

The onboarding endpoint MUST continue accepting optional `domain` as a simple root-operator input. When onboarding receives `domain`, the system MUST create a `company_domains` row for the new company with: `domain = normalized input domain`, `kind = primary`, `status = active`, `redirect_to_primary = false`. The onboarding endpoint MUST NOT write `domain` to `companies`. The onboarding endpoint MUST NOT accept or write `subdomain` as a company field. The onboarding endpoint MUST accept a boolean option controlling whether a technical platform domain should be generated. When technical-domain generation is requested and a platform base domain is configured, onboarding MUST create a `company_domains` row with: `domain = <slug>.<platform-base-domain>`, `kind = technical`, `status = active`, `redirect_to_primary = false`. The onboarding process MUST create company, domain rows, tenant roles, role permissions, and initial admin user in a single transaction. If any domain uniqueness check or domain creation step fails, onboarding MUST rollback the entire transaction.
(Previously: Same onboarding contract; domain creation logic moves to `create_company_domain/` slice repository)

#### Scenario: Onboarding creates primary custom domain

- GIVEN a root operator submits onboarding with `slug = "dreammakers"` and `domain = "dreammakers.com.co"`
- WHEN onboarding succeeds
- THEN the system MUST create the company AND a `company_domains` row for `dreammakers.com.co` with `kind = primary`, `status = active`
- AND the company row MUST NOT contain `domain` or `subdomain` fields as host source of truth

#### Scenario: Onboarding creates optional technical domain

- GIVEN platform base domain is configured as `kilashop.com`
- AND a root operator submits onboarding with `slug = "dreammakers"` and technical-domain generation enabled
- WHEN onboarding succeeds
- THEN the system MUST create `dreammakers.kilashop.com` as a `company_domains` row with `kind = technical`, `status = active`

#### Scenario: Onboarding skips technical domain

- GIVEN platform base domain is configured
- AND a root operator submits onboarding with technical-domain generation disabled
- WHEN onboarding succeeds
- THEN the system MUST NOT create a technical domain automatically

#### Scenario: Duplicate domain rolls back onboarding

- GIVEN an existing `company_domains` row for `dreammakers.com.co`
- WHEN a root operator submits onboarding with `domain = "dreammakers.com.co"`
- THEN onboarding MUST return a conflict or validation error for `domain`
- AND the transaction MUST rollback — no company, roles, role permissions, users, or domain rows from the failed request MUST remain

### Requirement: Tenant Host Resolution

Public tenant resolution MUST normalize the request host before lookup. Public tenant resolution MUST match the normalized host exactly against `company_domains.domain`. Public tenant resolution MUST only match rows with `status = active`. Public tenant resolution MUST return the company referenced by the matching company domain row. Public tenant resolution MUST NOT infer `www`/apex equivalence from string rules. Public tenant resolution MUST NOT fallback from first subdomain to company slug for production tenant resolution. A technical platform hostname MUST resolve only when it exists as an active `company_domains` row.
(Previously: Same resolution rules; unchanged by slice reorganization)

#### Scenario: Exact host resolves tenant

- GIVEN company `dreammakers` owns an active company domain `dreammakers.com.co`
- WHEN a public request arrives with `Host: dreammakers.com.co`
- THEN tenant resolution MUST resolve the request to the `dreammakers` company

#### Scenario: `www` requires explicit domain row

- GIVEN company `dreammakers` owns active domain `dreammakers.com.co`
- AND no row exists for `www.dreammakers.com.co`
- WHEN a public request arrives with `Host: www.dreammakers.com.co`
- THEN the system MUST NOT infer ownership from `dreammakers.com.co`
- AND tenant resolution MUST fail as not found

#### Scenario: Explicit `www` alias resolves tenant

- GIVEN company `dreammakers` owns active domains `dreammakers.com.co` and `www.dreammakers.com.co`
- WHEN a public request arrives with `Host: www.dreammakers.com.co`
- THEN tenant resolution MUST resolve the request to the `dreammakers` company

#### Scenario: Inactive domain does not resolve

- GIVEN company `dreammakers` owns domain `old-dreammakers.com.co` with `status = inactive`
- WHEN a public request arrives with `Host: old-dreammakers.com.co`
- THEN tenant resolution MUST NOT resolve the request to `dreammakers`

### Requirement: Redirect Behavior

If an active matching company domain has `redirect_to_primary = false`, the system MUST serve the request for the resolved company without redirecting. If an active matching company domain has `redirect_to_primary = true`, the system MUST redirect to the same company's active primary domain. Redirect behavior MUST preserve the original path and query string. Redirect behavior MUST NOT be inferred from domain spelling such as a `www.` prefix. If a redirect-enabled domain belongs to a company without an active primary domain, the system MUST NOT enter a redirect loop and SHOULD serve or fail deterministically according to implementation design. The active primary domain itself MUST NOT redirect to itself.
(Previously: Same redirect rules; unchanged by slice reorganization)

#### Scenario: Redirect alias redirects to primary

- GIVEN company `dreammakers` owns active primary domain `dreammakers.com.co`
- AND owns active alias domain `www.dreammakers.com.co` with `redirect_to_primary = true`
- WHEN a public request arrives for `https://www.dreammakers.com.co/products?tag=sale`
- THEN the system MUST return a permanent redirect to `https://dreammakers.com.co/products?tag=sale`

#### Scenario: No redirect when matched host equals primary

- GIVEN company `dreammakers` owns active primary domain `dreammakers.com.co`
- WHEN a public request arrives with `Host: dreammakers.com.co`
- THEN the system MUST serve the request without redirect

### Requirement: Companies API Surface

Company response DTOs MUST NOT expose `domain` or `subdomain` as direct company fields. Company detail responses SHOULD include the company's `domains` collection so root administration screens can inspect a company and its domains with one detail request. Company list responses SHOULD remain lean and MUST NOT include domains by default. Company update DTOs MUST NOT accept `domain` or `subdomain` as direct company fields. Domain management MUST be modeled separately from company profile management.
(Previously: Same DTO contract; unchanged by slice reorganization)

#### Scenario: Company detail includes domains

- GIVEN a root user
- AND a company with public ID `01HCOMPANY` with multiple domains
- WHEN root calls `GET /api/v1/companies/01HCOMPANY`
- THEN the system SHOULD return the company detail with a `domains` collection
- AND the response MUST NOT expose `domain` or `subdomain` as direct company fields

#### Scenario: Company list excludes domains

- GIVEN multiple companies with domains
- WHEN root calls `GET /api/v1/companies`
- THEN the response MUST NOT include domains by default

#### Scenario: Company update does not manage domains

- GIVEN an existing company
- WHEN a company update request is submitted
- THEN the request MUST NOT accept `domain` or `subdomain` as company profile fields
- AND domain ownership MUST remain unchanged by company profile update

### Requirement: Root Company Domain Administration

The system MUST expose root-only company domain administration under the companies module. The system MUST expose `GET /api/v1/companies/:id/domains` where `:id` is the existing company public ID route parameter convention. The system MUST expose `POST /api/v1/companies/:id/domains` for root-created company domains. The system MUST expose `PUT /api/v1/companies/:id/domains/:domain_id` for root updates to an existing company domain. The system MUST NOT expose a DELETE company-domain endpoint. The list endpoint MUST return all non-deleted domains for the specified company. Create and update payloads MUST include `domain`, `kind`, `status`, and `redirect_to_primary`. Create and update MUST normalize domain values before persistence. Create and update MUST enforce global domain uniqueness. Create and update MUST reject unsupported `kind` values. Create and update MUST reject unsupported `status` values. Create and update MUST enforce at most one active primary domain per company. Create and update MUST reject `redirect_to_primary = true` when `kind = primary`. Update MUST verify `:domain_id` belongs to the company identified by `:id`. Domain deactivation/reactivation MUST be performed by changing `status` through update. Domain administration MUST preserve tenant resolver semantics: only active exact domains for active companies resolve public requests.
(Previously: Same API surface; implementation moves to `list_company_domains/`, `create_company_domain/`, `update_company_domain/` slices, which import shared DTO/model contracts from `companies/shared`)

#### Scenario: Root lists company domains

- GIVEN a root user
- AND a company with public ID `01HCOMPANY` with multiple domains
- WHEN root calls `GET /api/v1/companies/01HCOMPANY/domains`
- THEN the system MUST return the company's domains

#### Scenario: Root creates company domain

- GIVEN a root user
- AND a company with public ID `01HCOMPANY`
- WHEN root calls `POST /api/v1/companies/01HCOMPANY/domains` with `domain = "www.acme.com"`, `kind = "alias"`, `status = "active"`, and `redirect_to_primary = true`
- THEN the system MUST create the domain for that company
- AND future tenant resolution MAY resolve or redirect that host according to the stored metadata

#### Scenario: Root updates domain status instead of deleting

- GIVEN a root user
- AND company `01HCOMPANY` owns domain `01HDOMAIN`
- WHEN root calls `PUT /api/v1/companies/01HCOMPANY/domains/01HDOMAIN` with `status = "inactive"`
- THEN the system MUST update the domain status
- AND the system MUST NOT delete the domain row

#### Scenario: Non-root cannot administer domains

- GIVEN an authenticated non-root user
- WHEN the user calls any company-domain administration endpoint
- THEN the system MUST return `403 Forbidden`

#### Scenario: Company domain delete route is absent

- GIVEN any user
- WHEN the user calls `DELETE /api/v1/companies/01HCOMPANY/domains/01HDOMAIN`
- THEN the system MUST return `404 Not Found` because no delete route exists

#### Scenario: Active primary uniqueness is enforced

- GIVEN company `01HCOMPANY` already has an active primary domain
- WHEN root creates or updates another domain for the same company with `kind = "primary"` and `status = "active"`
- THEN the request MUST fail with a validation/conflict response

#### Scenario: Cross-company domain update is rejected

- GIVEN domain `01HDOMAIN` belongs to company A
- WHEN root calls `PUT /api/v1/companies/{companyB}/domains/01HDOMAIN`
- THEN the system MUST reject the update as not found
