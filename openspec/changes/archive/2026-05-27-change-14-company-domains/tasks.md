# SDD Tasks: Company Domains for Multi-Domain Tenants

> Stop point: this plan is intentionally prepared through tasks only. Do not start apply until the user reviews and approves these artifacts.

## Review Workload Forecast

Estimated implementation risk: **medium-high**.

Expected touched areas:

- migration schema;
- companies model/DTO/repository/service tests;
- onboarding DTO/service/handler tests;
- tenant middleware and tests;
- app/container resolver contract if changed;
- OpenSpec canonical specs during archive.

Forecast: this change may exceed the 700 changed-line review budget if implemented in one pass with full redirect behavior and test coverage.

Recommended delivery strategy before apply:

1. Prefer one implementation branch/change, but keep work organized into reviewable phases.
2. If the RED/GREEN implementation crosses ~700 changed lines, split into chained PRs:
   - PR 1: schema/model + onboarding domain creation;
   - PR 2: tenant resolution + redirect behavior;
   - PR 3: DTO cleanup/spec sync if needed.

---

## Phase 1: Schema and Model Foundation

- [x] **1.1 Add `CompanyDomain` model**
  - Add constants for statuses: `active`, `inactive`, `pending_verification`.
  - Add constants for kinds: `primary`, `alias`, `technical`.
  - Add fields: `CompanyID`, `Domain`, `Status`, `Kind`, `RedirectToPrimary`.
  - Add relationship from `Company` to `[]CompanyDomain`.

- [x] **1.2 Remove direct host fields from `Company` model**
  - Remove `Domain` from `Company`.
  - Remove `Subdomain` from `Company`.
  - Ensure model still supports name, slug, and status behavior.

- [x] **1.3 Update existing Goose consolidated migration**
  - Modify `migrations/20260101000000_init.sql` directly; do not add a new migration file for this template project.
  - Remove `domain` and `subdomain` columns from the `companies` table definition.
  - Remove `idx_companies_domain` and `idx_companies_subdomain` from the existing migration.
  - Add `company_domains` table to the existing Goose migration.
  - Add unique constraint/index for `company_domains.domain` in the existing migration.
  - Add indexes for `company_id`, `status`, and `kind` in the existing migration.
  - Update the Goose `Down` section to drop `company_domains` consistently with the rest of the template schema.

- [x] **1.4 Update migration/model tests**
  - Add test expectations for `migrations/20260101000000_init.sql` containing `company_domains`.
  - Add test expectations that the existing migration no longer defines `companies.domain` or `companies.subdomain`.
  - Update AutoMigrate tests to include `CompanyDomain` where needed.

---

## Phase 2: Companies API Cleanup

- [x] **2.1 Update companies DTOs**
  - Remove `domain` and `subdomain` from company response DTOs.
  - Remove `domain` and `subdomain` from update DTOs.
  - Ensure create DTO compatibility is handled according to existing direct-create route removal.

- [x] **2.2 Update companies service mapping**
  - Stop reading/writing `Domain` and `Subdomain` on company create/update paths.
  - Keep company profile updates limited to profile/status fields.

- [x] **2.3 Update companies tests**
  - Adjust handler/service tests expecting company domain/subdomain fields.
  - Ensure domain ownership is not mutated by company profile update.

---

## Phase 3: Onboarding Domain Provisioning

- [x] **3.1 Update onboarding request DTO**
  - Keep optional `domain`.
  - Remove `subdomain`.
  - Add `generate_technical_domain` boolean.

- [x] **3.2 Add platform base-domain configuration access**
  - Add or reuse config for platform base domain, e.g. `APP_PLATFORM_DOMAIN`.
  - Define behavior when `generate_technical_domain = true` and config is missing.

- [x] **3.3 Create primary domain during onboarding**
  - Normalize `domain` input.
  - Check uniqueness against `company_domains.domain`.
  - Create `CompanyDomain{Kind: primary, Status: active, RedirectToPrimary: false}` inside the onboarding transaction.

- [x] **3.4 Create optional technical domain during onboarding**
  - Build `<slug>.<platform-base-domain>` when requested.
  - Check uniqueness against `company_domains.domain`.
  - Create `CompanyDomain{Kind: technical, Status: active, RedirectToPrimary: false}` inside the onboarding transaction.

- [x] **3.5 Update onboarding error mapping**
  - Keep duplicate custom domain mapped to `domain`.
  - Remove duplicate `subdomain` error mapping.
  - Add clear error for unavailable technical domain or missing platform base-domain config.

- [x] **3.6 Update onboarding tests**
  - Success path with custom primary domain.
  - Success path with generated technical domain.
  - Success path without technical domain when boolean is false.
  - Duplicate custom domain rolls back all writes.
  - Duplicate technical domain rolls back all writes.
  - Existing root-only and role/user provisioning tests still pass.

---

## Phase 4: Tenant Host Resolution and Redirects

- [x] **4.1 Introduce host resolution result**
  - Add a resolver return type carrying company ref, matched domain, kind, redirect flag, and primary domain.
  - Keep or adapt `FindByHost` only if it can support redirect metadata cleanly.

- [x] **4.2 Query active `company_domains` by exact host**
  - Match normalized host exactly against `company_domains.domain`.
  - Only consider `status = active`.
  - Return owning company reference.

- [x] **4.3 Remove production subdomain slug fallback**
  - Remove first-subdomain-to-company-slug production resolution.
  - Keep dev-only `X-Tenant` fallback if existing behavior requires it.
  - Ensure technical domains resolve only through explicit `company_domains` rows.

- [x] **4.4 Implement redirect behavior**
  - If matched domain has `redirect_to_primary = true`, find active primary domain for same company.
  - If primary exists and differs from request host, return permanent redirect preserving path and query.
  - Avoid redirect loops when host already equals primary.
  - Define deterministic behavior when no active primary exists.

- [x] **4.5 Update tenant middleware tests**
  - Active domain resolves tenant.
  - Inactive/pending domains do not resolve.
  - `www` host does not resolve unless explicit row exists.
  - Explicit `www` alias resolves tenant.
  - Redirect-enabled alias redirects to primary preserving path/query.
  - Technical domain resolves only when explicit row exists.

---

## Phase 5: Integration, Verification, and Review Guard

- [x] **5.1 Run strict TDD test suite**
  - RED: capture failing tests before each behavior implementation where practical.
  - GREEN: run `go test ./...` after implementation slices.
  - TRIANGULATE: add at least one edge case for duplicate/redirect behavior.
  - REFACTOR: clean duplicated normalization/domain creation code.

- [x] **5.2 Run build verification**
  - Run `go build ./...`.

- [x] **5.3 Review workload check before finalizing apply**
  - Inspect diff size.
  - If changed lines exceed 700 or review burden is high, pause and propose chained PR split before continuing.

- [ ] **5.4 Fresh review before commit/PR**
  - Run a fresh-context reviewer against the diff.
  - Specifically ask reviewer to check: migration consistency, transaction rollback safety, tenant resolution security, redirect loop risk, and stale DTO fields.
  - Not executed by this apply subagent; parent/user review remains the next step.

---

## Phase 6: Root Company Domain Administration

- [x] **6.0 Embed domains in company detail**
  - Update `GET /api/v1/companies/:id` response mapping to include the company's domains.
  - Keep `GET /api/v1/companies` list responses lean without domains by default.
  - Preserve dedicated nested domain endpoints for domain lifecycle management.

- [x] **6.1 Extend specs and design**
  - Document root-only domain administration endpoints under companies.
  - Confirm `:id` follows existing company public ID route convention.
  - Keep domain lifecycle status-based and omit DELETE endpoint.

- [x] **6.2 Add domain administration DTOs**
  - Add `CompanyDomainResponse`.
  - Add `CreateCompanyDomainRequest`.
  - Add `UpdateCompanyDomainRequest`.
  - Validate `domain`, `kind`, `status`, and `redirect_to_primary` rules.

- [x] **6.3 Extend companies service/repository**
  - List domains for a company public ID.
  - Create a domain for a company public ID.
  - Update a domain by company public ID and domain public ID.
  - Enforce global domain uniqueness.
  - Enforce at most one active primary domain per company.
  - Reject cross-company domain updates.

- [x] **6.4 Add root-only routes and handlers**
  - Add `GET /api/v1/companies/:id/domains`.
  - Add `POST /api/v1/companies/:id/domains`.
  - Add `PUT /api/v1/companies/:id/domains/:domain_id`.
  - Do not add a DELETE route.

- [x] **6.5 Verify domain admin behavior**
  - Add RED tests before implementation for route/service behavior.
  - Run focused companies tests.
  - Run full `go test ./...`.
  - Run `go build ./...`.

---

## Acceptance Checklist

- [x] `companies.domain` and `companies.subdomain` are no longer source-of-truth fields.
- [x] `company_domains` stores all tenant-resolvable hostnames.
- [x] Onboarding creates primary domain from `domain` input.
- [x] Onboarding optionally creates technical platform domain via boolean flag.
- [x] `www` aliases require explicit rows.
- [x] Redirect behavior is explicit via `redirect_to_primary`.
- [x] Tenant resolution uses only active exact `company_domains.domain` matches.
- [x] Domain uniqueness is global.
- [x] Domains are deactivated/reactivated via status rather than normal deletion.
- [x] `GET /api/v1/companies/:id` includes the company's domains for detail/admin inspection.
- [x] Root can list/create/update domains for a company through nested companies endpoints.
- [x] No company-domain DELETE route exists.
- [x] Domain admin validates kind, status, redirect-to-primary, uniqueness, active primary count, and company ownership.
- [x] `go test ./...` passes.
- [x] `go build ./...` passes.
