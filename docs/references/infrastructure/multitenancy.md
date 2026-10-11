# Tenant context and isolation

This reference owns private tenant selection and repository scoping. The mechanisms live in `internal/middleware/tenant.go` and `internal/platform/tenant/tenant.go`.

## Context fields

`TenantContext` has exactly `CompanyID uint`, `CompanySlug string`, `IsRootScope bool`. `NewRoot()` sets global scope; `NewScoped(id, slug)` selects one company. There is no `CompanyPublicID` or `CanAccessGlobal` field. Public API identifiers are ULIDs, not UUIDs; internal company IDs are persistence keys.

## Private resolution

| Actor / route policy | Input | Result |
|---|---|---|
| Root, `AllowRootGlobalScope` | No `X-Company-ID` | Global root context |
| Root, `RequireTenantScope` | No header | 400: company scope required |
| Root, either policy | Resolvable company public ID or slug | Selected company context |
| Root, either policy | Invalid/unresolved header | 400 |
| Non-root | Authenticated company present | That company; selection header ignored |
| Non-root | No authenticated company | 403 |
| Any | Missing authenticated actor | 401 |

The app uses required scope for IAM users and root-global-capable scope for companies, roles, permissions and onboarding. `FindByPublicIDOrSlug` does not enforce active company status; do not assume root header selection is the same active-domain gate as host resolution.

## Persistence boundary

Tenant-owned models need `company_id` plus appropriate migration indexes. Handlers must reject absent context, then pass it to repository methods. Apply `tenant.ApplyTenantScope` before every tenant-owned read/write; it adds `WHERE company_id = ?` unless `IsRootScope` is true. A scoped root is filtered like any other selected tenant.

The helper does not authorize the actor or validate the company ID; trusted middleware/context construction is essential. Zero-value context is not global scope. Cross-tenant misses return 404 through the module not-found outcome, without disclosing whether the foreign-company record exists. Global tables deliberately omit this scope and need their own guard policy.

Current authentication checks user activity, but does not universally reject users based on company status; do not promise cascade session invalidation or an inactive-company login gate. Host-resolution limitations are documented separately in [domains](domains.md).

## Isolation checklist

- [ ] Tenant-owned models include `company_id` with appropriate migration indexes.
- [ ] Trusted tenant context is required before persistence.
- [ ] Reads, updates, deletes, password/role/status changes preserve scope.
- [ ] Tests cover own-company success, foreign-company miss, root-global and selected root separately where allowed.
- [ ] Auth bootstrap/global-table exceptions are explicit, not accidental unscoped queries.
- [ ] Raw SQL carries company and deletion constraints explicitly; see [persistence policy](../../boundaries_and_cqrs.md#5-reusable-gorm-helpers).
