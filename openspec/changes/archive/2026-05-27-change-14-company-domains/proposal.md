# Proposal: Company Domains for Multi-Domain Tenants

## Goal Description
Replace single-host company ownership (`companies.domain` and `companies.subdomain`) with a first-class `company_domains` model. The platform must support multiple hostnames per company, exact host resolution, optional technical platform domains, and redirect-capable aliases without embedding hidden string heuristics in tenant middleware.

This change keeps onboarding simple for root operators while moving domain ownership to a proper one-to-many table.

---

## User Review Decisions Captured

> [!IMPORTANT]
> **Onboarding Input Compatibility**
> `domain` remains an onboarding input. When supplied, onboarding creates a primary `company_domains` row instead of writing `companies.domain`.

> [!IMPORTANT]
> **No Company Subdomain Field**
> `companies.subdomain` is removed as a source of truth. Technical subdomains are represented as explicit `company_domains` rows.

> [!NOTE]
> **Explicit Aliases, No Magic `www` Logic**
> `www.example.com` must be stored as its own domain row if it should resolve. Tenant resolution performs exact host matching against `company_domains.domain`.

> [!NOTE]
> **Redirect Behavior Included**
> Domain rows can opt into redirecting to the company's active primary domain. This enables `www → apex`, legacy brand redirects, or canonical SEO behavior without requiring immediate redirect use for every alias.

---

## Proposed Changes

### 1. Database Schema: `company_domains`

Because this repository is a starter template using Goose with a consolidated initial migration, modify `migrations/20260101000000_init.sql` directly instead of adding a new migration file.

Add a `company_domains` table:

- `id INTEGER PRIMARY KEY SERIAL`
- `public_id CHAR(26) UNIQUE NOT NULL`
- `company_id INTEGER NOT NULL REFERENCES companies(id)`
- `domain VARCHAR(255) UNIQUE NOT NULL`
- `status VARCHAR(40) NOT NULL`
- `kind VARCHAR(40) NOT NULL`
- `redirect_to_primary BOOLEAN NOT NULL DEFAULT false`

Remove from `companies`:

- `domain`
- `subdomain`

Initial domain statuses:

- `active`
- `inactive`
- `pending_verification`

Initial domain kinds:

- `primary`
- `alias`
- `technical`

### 2. Domain Lifecycle Semantics

Domains are not normally soft-deleted. They are deactivated/reactivated through `status`.

- Active domains resolve tenants.
- Inactive or pending domains do not resolve public tenants.
- `domain` is globally unique so an inactive domain still remains owned by its company until an explicit future transfer/release operation exists.

### 3. Onboarding Changes

Update company onboarding so it:

1. Keeps accepting optional `domain` as a simple input.
2. Creates the company without domain/subdomain fields.
3. Creates a `company_domains` row with `kind = 'primary'` and `status = 'active'` when `domain` is provided.
4. Accepts a boolean flag that controls whether a technical domain should be generated.
5. If technical-domain generation is requested and platform base-domain config exists, creates `slug.<platform-domain>` with `kind = 'technical'`.
6. Performs all domain creation inside the existing onboarding transaction.

### 4. Tenant Resolution Changes

Update tenant public-host resolution to use only active `company_domains` rows:

```text
Host header → normalize host → query company_domains.domain = host AND status = active → company_id
```

Remove fallback resolution from first subdomain to company slug. The platform technical host works only when represented as an explicit `company_domains` row.

### 5. Redirect Behavior

When an active domain row has `redirect_to_primary = true`, public host middleware or a dedicated redirect middleware resolves the tenant and responds with a permanent redirect to the company's active primary domain, preserving path and query string.

Example:

```text
GET https://www.example.com/products?tag=sale
→ 308 Location: https://example.com/products?tag=sale
```

Redirect behavior must not be inferred from `www` or string patterns; it must come from stored domain metadata.

### 6. Companies API Surface

Update companies model/DTO/spec behavior so `domain` and `subdomain` are no longer direct company fields. Company domain management may be exposed later as dedicated endpoints, but this change focuses on schema, onboarding, and tenant resolution foundations.

---

## Verification Plan

### Automated Tests

Run:

```bash
go test ./...
go build ./...
```

Add/update tests for:

- onboarding creates primary domain row from `domain` input;
- onboarding optionally creates technical domain row when requested;
- duplicate domain conflicts rollback onboarding;
- tenant resolution matches active company domains exactly;
- `www` host resolves only when stored explicitly;
- inactive/pending domains do not resolve;
- redirect-enabled alias redirects to active primary domain preserving path/query;
- company DTOs no longer expose/update direct `domain`/`subdomain` fields.

### Manual Verification

- Onboard a company with `domain = dreammakers.com.co` and technical-domain generation enabled.
- Add an alias `www.dreammakers.com.co` with redirect enabled.
- Verify `dreammakers.com.co` serves the tenant.
- Verify `www.dreammakers.com.co` redirects to `dreammakers.com.co`.
- Verify `dreammakers.kilashop.com` serves the tenant if created as a technical domain.

---

## Rollback Plan

If implementation proves too disruptive, retain `companies.domain` and `companies.subdomain` columns temporarily while adding `company_domains`, backfill rows, and switch tenant resolution behind tests first. Only remove the company columns after onboarding and tenant resolution pass verification.
