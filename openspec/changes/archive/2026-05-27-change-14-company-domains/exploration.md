# Exploration: change-14-company-domains

## Status
Explored.

## Executive Summary
Replace direct `companies.domain` and `companies.subdomain` ownership with a first-class `company_domains` table so one company can own multiple hostnames: apex domain, `www` redirect host, and technical platform subdomain.

This affects tenant host resolution, company onboarding, companies CRUD DTOs/models, migrations, and tests. The key design constraint is to avoid implicit string heuristics: if `www.example.com` must resolve to `example.com`, that alias should be represented as a domain row or an explicit canonicalization rule.

## User Problem
A single `domain` column on `companies` limits each tenant to one exact host. In real SaaS deployments, customers commonly require multiple domains and redirects. Example for company ID 1:

- `dreammakers.com.co` — primary domain
- `www.dreammakers.com.co` — redirect/canonical alias
- `dreammakers.kilashop.com` — free technical subdomain

## Proposed Data Model
`company_domains`

- `id INTEGER PRIMARY KEY SERIAL`
- `public_id CHAR(26) UNIQUE NOT NULL`
- `company_id INTEGER NOT NULL REFERENCES companies(id)`
- `domain VARCHAR(...) UNIQUE NOT NULL`
- `status VARCHAR(...) NOT NULL`
- `is_primary BOOLEAN NOT NULL DEFAULT false`
- `deleted_at TIMESTAMP WITH TIME ZONE`

Remove or deprecate from `companies`:

- `domain`
- `subdomain`

## Impacted OpenSpec Areas
- `openspec/specs/companies-crud/spec.md`
  - Currently models `Company.Domain *string` and `Company.Subdomain *string`.
  - States company creation is only through onboarding, so onboarding behavior is in scope.
- `openspec/specs/company-onboarding/spec.md`
  - Must describe creation of company domain rows during onboarding.
- `openspec/specs/tenant-isolation/spec.md`
  - Tenant data remains scoped by `company_id`, but host-to-company resolution changes.
- `openspec/specs/migrations/spec.md`
  - Must cover schema migration and backfill/removal strategy.

## Impacted Code Areas
- `migrations/20260101000000_init.sql`
  - Creates `companies.domain` and `companies.subdomain`.
  - Adds indexes `idx_companies_domain` and `idx_companies_subdomain`.
- `internal/modules/companies/model.go`
  - Remove `Domain`, `Subdomain` from `Company`.
  - Add `CompanyDomain` model and relation.
- `internal/modules/companies/dto.go`
  - Company responses/update requests expose `domain`/`subdomain` today.
- `internal/modules/companies/service.go`
  - Create/update currently map direct company fields.
- `internal/modules/companies/repository.go`
  - `FindByHost` currently queries `companies.domain = ?`.
- `internal/middleware/tenant.go`
  - Public tenant resolution calls `FindByHost(host)` then falls back to subdomain/slug.
  - `www.*` is ignored as subdomain, so `www.dreammakers.com.co` will not resolve unless stored exactly.
- `internal/modules/onboarding/dto.go`
  - Onboarding accepts `Domain` and `Subdomain`.
- `internal/modules/onboarding/service.go`
  - Validates duplicate domain/subdomain against `companies`, normalizes them, then stores them on `Company`.
- `internal/modules/onboarding/handler.go`
  - Returns validation errors for `domain` and `subdomain`.
- `internal/modules/auth/handler.go`
  - Has TODO for tenant-aware login by company domains.

## Test Impact
- `internal/middleware/tenant_test.go`
  - Exact host domain resolution should use `company_domains`.
- `internal/modules/onboarding/service_test.go`
  - AutoMigrate and assertions need `CompanyDomain`.
- `internal/modules/companies/repository_test.go`
  - Sort allowlist test references `domain` as disallowed sort.
- `internal/modules/companies/service_test.go` and handler tests
  - Likely need DTO field updates.

## Migration Requirements
- Add `company_domains` table.
- Backfill existing `companies.domain` values into primary `company_domains` rows.
- Decide what happens to existing `companies.subdomain` values:
  - convert into `<subdomain>.<platform-domain>` rows;
  - keep slug/subdomain fallback temporarily;
  - or remove from public API and migrate only when a platform domain exists.
- Remove old indexes and columns when safe.
- Decide soft-delete uniqueness:
  - plain unique `domain` blocks reuse after soft delete;
  - partial unique `domain WHERE deleted_at IS NULL` allows reuse.

## Onboarding Impact
Current onboarding accepts one optional `domain` and one optional `subdomain`. With `company_domains`, onboarding should create company + domain rows transactionally with roles, permissions, and admin user.

Open decisions:

1. Should onboarding still accept `domain`?
2. If yes, should it create an active primary `company_domains` row?
3. What should replace `subdomain`?
4. Should `www.example.com` automatically map to `example.com`, or must both be stored?
5. Should soft-deleted domains be reusable?

## Tenant Resolution Impact
- `CompanyResolver.FindByHost(host)` should query active, non-deleted `company_domains.domain`.
- Host normalization already exists in middleware, but apex/`www` equivalence does not.
- Cache key is currently `host:<host>`; domain status/deletion changes may remain cached up to five minutes unless cache invalidation is added.

## Likely Requirements
- The system MUST support multiple domains per company.
- The system MUST resolve public tenant requests through active, non-deleted `company_domains` rows.
- The system MUST enforce domain uniqueness across active company domains.
- The system MUST support exactly one primary active domain per company, unless no domains exist.
- Onboarding MUST create a primary company domain when a custom domain is supplied.
- Company update APIs MUST NOT write `domain` or `subdomain` directly on `companies`.

## Risks
- Breaking public tenant resolution if `FindByHost` is not moved to `company_domains`.
- Data migration ambiguity for existing `domain` and `subdomain` values.
- Unique-domain semantics conflict with soft delete if plain unique index is used.
- Onboarding transaction complexity increases.
- `www`/apex canonicalization could become hidden string logic unless modeled explicitly.
- Auth login remains non-tenant-aware despite domain resolution TODO.

## Next Recommended Phase
Create proposal/spec for `change-14-company-domains`, after confirming the product decisions around onboarding domain inputs, subdomain behavior, `www` aliases, and soft-delete reuse.
