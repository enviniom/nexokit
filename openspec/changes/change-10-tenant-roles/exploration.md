# Exploration: Tenant-scoped roles with global root

## Current State

The system currently has a **global RBAC model** where all roles (`root`, `admin`, `user`) exist at the system level with no tenant isolation:

- **`roles` table**: has `id`, `public_id`, `name` (unique), `slug` (unique), `description`, `is_system`, timestamps. **No `company_id` column**.
- **Seeds**: `RolesSeed()` creates `root`, `admin`, `user` all with `is_system: true`. `RolePermissionsSeed()` assigns all system permissions to `root`, admin-level to `admin`, basic to `user`.
- **Authorization middleware**: `AttachPermissions` checks `user.IsRoot` and grants `"*"` permission marker. `RequirePermission` allows `"*"` or matching slug. Root bypass is handled at middleware level, not via `role_permissions` rows.
- **Roles API**: Currently mounted on `globalProtected` group (`AllowRootGlobalScope`), meaning all roles are visible globally. No tenant scoping on list/get.
- **Service protection**: `isRootRole()` blocks create/update/delete of roles matching "root" by name or slug. `IsSystem` flag blocks mutations on seeded roles.
- **Users**: `users.company_id` is nullable. Users have `role_id` FK to `roles(id)`.

### Key files affected
- `internal/modules/roles/model.go` — Role struct needs `CompanyID *uint`
- `internal/modules/roles/repository.go` — All queries need tenant scope
- `internal/modules/roles/service.go` — List/Get/Create need tenant awareness, reserved slug validation
- `internal/modules/roles/handler.go` — List/Get/Create need to extract tenant context from request
- `internal/modules/roles/dto.go` — Response needs `company_id` field
- `internal/modules/roles/routes.go` — May need route splitting (tenant-scoped vs global)
- `migrations/20260516000000_auth.sql` — Needs new migration to add `company_id` to `roles`
- `seeds/roles.go` — Must only seed `root` global; remove `admin`/`user` seeds
- `seeds/role_permissions.go` — Must remove root permission assignments from seed
- `internal/app/container.go` — Roles route mounting may need adjustment
- `internal/middleware/authorization.go` — Already handles root bypass correctly via `"*"` marker

## Affected Areas

- `internal/modules/roles/model.go` — Add `CompanyID *uint` field to Role struct
- `internal/modules/roles/repository.go` — Add `company_id` to all queries; new `ListByCompanyID` method
- `internal/modules/roles/service.go` — Tenant scoping in List/Create; reserved slug validation expansion
- `internal/modules/roles/handler.go` — Extract tenant context; pass to service
- `internal/modules/roles/dto.go` — Add `CompanyID` to RoleResponse
- `internal/modules/roles/routes.go` — Split global (root-only) vs tenant-scoped routes
- `migrations/` — New migration: `ALTER TABLE roles ADD COLUMN company_id INTEGER`
- `seeds/roles.go` — Remove admin/user seeds; keep only root
- `seeds/role_permissions.go` — Remove root permission assignments
- `internal/app/container.go` — Roles route group may need `RequireTenantScope` for non-root
- `internal/modules/roles/service_test.go` — Tests need tenant context
- `internal/modules/roles/handler_test.go` — Tests need tenant context
- `openspec/specs/roles/spec.md` — Update main spec for tenant-scoped behavior

## Approaches

### Approach 1: Single endpoint with tenant context filtering (Recommended)

Keep all routes on the same `/api/v1/roles` group. The handler extracts tenant context from the request (already set by middleware) and passes it to the service. Root users operate globally; non-root users are automatically scoped to their company.

- **Pros**: Simpler routing, single code path, middleware already sets tenant context
- **Cons**: Need to handle "root global list" vs "tenant-scoped list" in same endpoint
- **Effort**: Medium
- **Details**: 
  - `List` checks if user is root → list all; otherwise → filter by `user.CompanyID`
  - `Create` sets `company_id` from `user.CompanyID` (root cannot create roles via this endpoint, or creates for a specific company via `X-Company-ID` header)
  - `Get` checks ownership: root can get any role; non-root can only get roles from their company

### Approach 2: Split routes — global vs tenant-scoped

Create two route groups: one for root-global operations (`/api/v1/roles` with `AllowRootGlobalScope`) and one for tenant-scoped operations (with `RequireTenantScope`).

- **Pros**: Clearer separation of concerns, explicit scoping
- **Cons**: More complex routing, potential duplicate handlers or service methods
- **Effort**: Medium-High
- **Details**: Would need to decide which operations go where; could cause confusion

### Approach 3: Repository-level tenant scope with context injection

Pass `context.Context` with tenant info through the service to repository. Repository applies scope automatically.

- **Pros**: Clean separation, repository handles scoping transparently
- **Cons**: Requires changing repository interface signatures; more invasive
- **Effort**: High

## Recommendation

**Approach 1** is the best fit because:

1. The tenant middleware already sets `TenantContext` in the Gin context
2. The `authctx.User` already has `CompanyID` and `IsRoot`
3. The existing pattern in the codebase (users module) uses the same approach
4. Minimal interface changes — service methods can accept optional `companyID` parameter

### Implementation details:

1. **Migration**: Add `company_id INTEGER` nullable to `roles` table with index
2. **Model**: Add `CompanyID *uint` to Role struct
3. **Repository**: 
   - `List(page, perPage, companyID *uint)` — filter by company if provided
   - `Count(companyID *uint)` — count by company if provided
   - `ListByCompanyID(companyID uint, page, perPage int)` — alternative: keep signature and use optional param
4. **Service**:
   - `List(page, perPage, companyID *uint)` — pass company filter
   - `Create(req, companyID *uint)` — set company on new role
   - Reserved slug check: expand `isRootRole` → `isReservedSlug(slug)` checking `root`, `admin`, `user`
5. **Handler**: Extract `companyID` from auth context or tenant context
6. **Seeds**: 
   - `RolesSeed()` only creates `root` with `company_id = NULL`, `is_system = true`
   - `RolePermissionsSeed()` does NOT assign permissions to root (bypass is middleware-level)
7. **Routes**: Keep on `globalProtected` group — root can manage globally, non-root sees only their company's roles
8. **DTO**: Add `CompanyID *string` (public ID of company, not uint) to `RoleResponse`

## Risks

1. **Existing data migration**: If there are existing roles in production, they need to be assigned to a company or marked as global. The migration should handle this gracefully.
2. **Unique constraint conflict**: `name` and `slug` are globally unique today. With tenant scoping, they should be unique **per company** (except for global roles). This requires changing unique indexes to composite: `(name, company_id)` and `(slug, company_id)` where `company_id` can be NULL (PostgreSQL handles NULL in unique indexes as distinct values).
3. **Seed order dependency**: Removing `admin`/`user` from seeds means company onboarding must create these roles. If onboarding doesn't exist yet, new companies will have no default roles.
4. **Root permission bypass**: The middleware already handles root bypass via `"*"` marker. The seed should NOT create `role_permissions` rows for root. Existing rows (if any) should be cleaned up or ignored.
5. **API response change**: Adding `company_id` to RoleResponse is a breaking change for clients that don't expect it (though it should be `omitempty`).
6. **Roles currently mounted on `globalProtected`**: This means non-root users can see ALL roles globally. After this change, the List/Get endpoints must filter by company. The route group is fine but the service/repository must enforce scoping.

## Ready for Proposal

**Yes** — the exploration is complete. The recommended approach is clear:
- Add `company_id` nullable to roles table
- Scope role queries by company for non-root users
- Seed only `root` globally; remove admin/user from seed
- Root bypass remains at middleware level (no `role_permissions` needed)
- Reserved slugs (`root`, `admin`, `user`) cannot be created via API

The orchestrator should proceed to **sdd-propose** to define the scope and approach formally, then **sdd-spec** for detailed requirements.
