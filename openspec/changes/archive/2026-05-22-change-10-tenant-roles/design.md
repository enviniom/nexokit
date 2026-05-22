# Design: Tenant-scoped roles with global root

## Technical Approach

Extend the existing roles module to use the project’s `tenant.TenantContext` pattern already used by users. Keep routes under `globalProtected` with `AllowRootGlobalScope`: root can omit `X-Company-ID` for global reads, or provide it for tenant-scoped work; non-root users are always scoped by middleware. The service/repository boundary enforces tenant isolation, reserved slugs, and root immutability.

## Architecture Decisions

| Decision | Alternatives considered | Rationale |
|---|---|---|
| Pass `tenant.TenantContext` through roles handler/service/repository | Raw `*uint companyID`, route splitting | Matches `internal/modules/users/*`, carries root-global vs scoped intent, and avoids duplicating endpoints. |
| Root remains the only global seeded role (`company_id NULL`) | Keep global `admin`/`user`; create tenant roles here | Tenant onboarding is out of scope; global admin/user would violate SaaS isolation. |
| Enforce uniqueness by tenant in DB and service | Application-only checks | Partial/composite indexes protect concurrent writes; service keeps friendly conflict behavior. |
| Replace `isRootRole` with reserved slug/name guard for API writes | Keep root-only guard | Proposal requires API to reject `root`, `admin`, and `user` reserved identities. |
| Keep authorization middleware unchanged | Add root role permissions rows | `AttachPermissions` already grants root `"*"`; root `role_permissions` rows are redundant and should be removed from seeds. |

## Data Flow

```text
Request ─→ Auth ─→ AllowRootGlobalScope ─→ roles.Handler
                                      │         │
                                      └─ tenant.TenantContext
                                                │
roles.Service ── enforces reserved/root rules ──→ roles.Repository
                                                │
                                      tenant.ApplyTenantScope(db, tc)
```

## File Changes

| File | Action | Description |
|---|---|---|
| `migrations/20260520000000_tenant_roles.sql` | Create | Add nullable `roles.company_id`, FK to `companies`, remove global unique constraints, add partial global and tenant composite indexes; down restores old indexes. |
| `internal/modules/roles/model.go` | Modify | Add `CompanyID *uint`; change GORM indexes from global unique to tenant-aware DB-managed uniqueness. |
| `internal/modules/roles/repository.go` | Modify | Change `List`, `Count`, `GetByPublicID`, `Delete` to accept `tenant.TenantContext`; add scoped uniqueness helpers for name/slug. |
| `internal/modules/roles/service.go` | Modify | Change public methods to accept tenant context; set `CompanyID` on create; prevent scoped users from accessing other tenants; forbid reserved create/update/delete. |
| `internal/modules/roles/handler.go` | Modify | Extract `tenant.FromGin` like users handler; pass context to service; prefer `response.HandleError` while preserving `204` delete. |
| `internal/modules/roles/dto.go` | Modify | Add `CompanyID *uint json:"company_id,omitempty"` to response. Do not add request `company_id`; scope comes from tenant context/header. |
| `internal/app/container.go` | Modify | No route move expected; keep roles on `globalProtected`. Ensure `CompanySlug` is populated in `userLookup` if needed by tenant tests. |
| `seeds/roles.go` | Modify | Seed only root as system/global. |
| `seeds/role_permissions.go` | Modify | Remove root/admin/user default role assignments for this change; permission catalog remains. |
| `internal/modules/roles/*_test.go`, `seeds/*_test.go` | Modify | Update fakes and table tests for tenant scoping, reserved slugs, and seed behavior. |

## Interfaces / Contracts

```go
type Repository interface {
    List(tc tenant.TenantContext, page, perPage int) ([]Role, error)
    Count(tc tenant.TenantContext) (int64, error)
    GetByPublicID(tc tenant.TenantContext, publicID string) (*Role, error)
    GetByName(tc tenant.TenantContext, name string) (*Role, error)
    GetBySlug(tc tenant.TenantContext, slug string) (*Role, error)
}
```

Service contract mirrors this by adding `tenant.TenantContext` to list/get/create/update/delete and permission catalog/assignment lookups.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Reserved slug/name guard; tenant-scoped List/Get/Create/Update/Delete; root global behavior | Table-driven service tests with updated fakes. |
| Handler | Tenant context is passed from Gin; missing context falls back safely to root for isolated unit tests | httptest with `tenant.SetGin`. |
| Repository/migration | Index behavior: same slug across companies allowed, duplicate within same company rejected, one global root | SQLite/GORM where possible; migration SQL reviewed plus `go test ./...`. |
| Seeds | Only root role exists; no role_permissions rows required for root | Existing seed tests updated. |

## Migration / Rollout

Development migration only: add nullable `company_id`, leave existing root global, remove admin/user seeds. Production orphan role assignment remains out of scope and needs a separate rollout script before production use.

## Open Questions

- [ ] Should root create tenant roles by providing `X-Company-ID`, or should root be read-only for tenant roles until onboarding creates defaults?
