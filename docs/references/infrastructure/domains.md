# Company domains and optional host resolution

Companies own domain records and host resolution as tenant infrastructure. Register public host middleware explicitly on public tenant surfaces. Private header selection follows [multitenancy](multitenancy.md), not host matching.

## Domain resolution

`companies.Resolver.ResolveHost` joins `company_domains` to `companies`, matches an exact normalized host and requires both statuses to be `active`. `PublicTenant` lowercases/trims Host and strips its port before lookup. There is **no first-subdomain-as-slug fallback**.

When host resolution fails, only `APP_ENV=development` permits `X-Tenant` fallback by public ID or slug. This fallback is not enabled for the separate `local` environment string. Private company-ID/slug lookup and this fallback do not apply the same active-status filter as host lookup.

Company domains carry primary/alias/technical kind and optional `RedirectToPrimary`. If an alias requests redirect and an active primary domain is found, middleware emits 308, preserving request URI/path/query. Scheme comes from the first `X-Forwarded-Proto` value, TLS or URL, then defaults to HTTP. Only trusted proxies should supply forwarded headers; see [deployment](../../deployment.md#tls-and-supervision).

## Administration policy

Keep company domain administration root-guarded under `/api/v1/companies/:id/domains`. Onboarding may create an explicit primary domain and an optional technical domain derived from `APP_PLATFORM_DOMAIN`. Domain records do not provision DNS, verify domain ownership or issue TLS certificates; operators manage DNS, certificates and proxy routing separately.

## Operational caveats

- Ordinary GORM domain queries filter domain soft deletion. The resolver's join does **not explicitly filter `companies.deleted_at`**; do not claim a comprehensive deleted-company host rejection guarantee from it.
- Company status is checked for host matching, not universally for credential login or private header lookup.
- `PublicTenant` caches successful host/header lookups for five minutes in closure-local maps. No invalidation or synchronization is present; do not assume immediate domain revocation or concurrent-map safety.

Before mounting public routes, address/test these isolation and concurrency concerns for the intended deployment. Middleware registration alone does not establish safe public host resolution.

Sources: `internal/modules/companies/resolver.go`, `internal/modules/companies/core/company_domain/model.go`, `internal/modules/companies/routes.go`; `internal/middleware/tenant.go`; `internal/modules/onboarding/slices/onboard_company/service.go` and `repository.go`.
