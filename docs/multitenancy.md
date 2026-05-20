# Multitenancy review guide

NexoKit isolates tenant-owned data by passing an explicit tenant context from middleware to handlers, services, and repositories. New modules that store company data must include `company_id` and must scope repository queries with `tenant.ApplyTenantScope`.

## Quick path

1. Add `CompanyID uint` to tenant-owned models and migrations.
2. Read `tenant.TenantContext` from Gin in handlers, then pass it through service methods.
3. Apply `tenant.ApplyTenantScope(db, tc)` in every tenant-owned repository query.
4. Add same-company and cross-company tests before implementing behavior.

## Tenant model fields

| Field | Meaning | Review note |
|-------|---------|-------------|
| `CompanyID` | Internal company database ID used for GORM filtering. | Keep it internal; do not expose raw uint IDs in routes. |
| `CompanySlug` | Human-readable company slug resolved by tenant middleware when available. | Useful for logs and public tenant resolution, not for authorization. |
| `IsRootScope` | `true` when root is operating globally. | Global root queries are intentionally unfiltered. |

Tenant-owned models should include:

```go
CompanyID uint `gorm:"index;not null" json:"company_id"`
```

Root-owned or platform-global records can omit `company_id`, but that must be deliberate and documented in the module design.

## Repository scope rules

Repository methods are the enforcement boundary. Handlers and services can validate requests, but repositories must still apply the tenant filter so missed handler checks do not leak data.

```go
func (r *WidgetRepository) FindByPublicID(ctx context.Context, tc tenant.TenantContext, publicID string) (*Widget, error) {
	var widget Widget
	q := r.db.WithContext(ctx).Where("public_id = ?", publicID)
	q = tenant.ApplyTenantScope(q, tc)
	if err := q.First(&widget).Error; err != nil {
		return nil, err
	}
	return &widget, nil
}
```

Rules to keep:

- Root with `IsRootScope == true` can query globally.
- Root scoped to a company uses the same filter path as non-root users.
- Non-root reads, updates, deletes, password changes, and status changes must be filtered by `company_id`.
- Cross-tenant misses return 404 so other tenants' records appear non-existent.
- Auth bootstrap lookups may remain unscoped only when they are required to build the authenticated context.

## Review checklist for new modules

- [ ] Tenant-owned models include `company_id` and a migration index.
- [ ] Handlers read `tenant.FromGin(c)` before tenant-owned read/update/delete/list operations.
- [ ] Services accept `tenant.TenantContext` for tenant-owned operations.
- [ ] Repositories call `tenant.ApplyTenantScope` before executing tenant-owned queries.
- [ ] Routes use public IDs or slugs; internal uint IDs are never exposed.
- [ ] Tests cover same-company success, root global access, root scoped access, and cross-tenant 404 behavior.
- [ ] Generated modules from `nexokit make module --tenant` are compared against golden files after template changes.

## Next step

For reusable helpers and package-level notes, see `internal/platform/tenant/doc.go`.
