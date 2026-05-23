# Design: User role assignment separation

## Technical Approach

Keep user management and role catalog responsibilities separate:

- The users module owns changing a target user's assigned role through `PATCH /api/v1/users/:id/role`.
- The roles module owns listing roles for selects through `GET /api/v1/roles/select`.
- Authorization remains route-level via `RequirePermission`.
- Tenant isolation is enforced at service/repository boundaries with `tenant.TenantContext`.
- Root retains authorization bypass but not assignment bypass for the reserved `root` role.

## Architecture Decisions

| Decision | Alternatives considered | Rationale |
|---|---|---|
| Use `GET /api/v1/roles/select` | `GET /api/v1/users/assignable-roles` | Roles for select are role catalog data; keeping it in roles avoids making users own role listing. |
| Add `roles.select` | Reuse `roles.index` or `users.change_role` | UI select data is narrower than full role listing and separate from actual mutation. |
| Use role PublicID in `PATCH /users/:id/role` body | Internal numeric ID for consistency with legacy DTOs | Public APIs must not expose internal IDs; new endpoint should establish the correct contract. |
| Ignore `role_id` in general update instead of treating it as mutation | Reject requests containing `role_id` | Acceptance requires the role not change when `role_id` is present. Ignoring preserves compatibility while removing the sensitive effect. |
| Forbid all self role changes | Only forbid changes that increase permissions | Simpler, safer, and avoids a role-permission diff engine outside this change scope. |
| Invalidate target user permission cache directly from users service | Wait for cache TTL | Security-sensitive role changes should take effect immediately. |
| Explicitly compare target user company and target role company | Rely only on tenant-scoped lookup | Root/global contexts can see multiple companies, so lookup visibility alone does not prove assignability. |
| Use a roles-owned contract summary for users service | Return `roles.Role` GORM model directly | Base architecture forbids importing another module's repository/model internals; a small contract preserves module ownership. |
| Tighten `company_id` in general update | Leave as-is because adjacent | Existing update can move company IDs by request body; tightening prevents a nearby tenant escalation while touching the same service path. |

## Data Flow

### General user update

```text
PUT /api/v1/users/:id
  └─ RequirePermission("users.update")
     └─ users.Handler.Update
        └─ Validate UpdateUserRequest(name,email,company_id?)
           └─ users.Service.Update(tc,targetID,actorID,req)
              ├─ repo.GetByPublicID(tc,targetID)
              ├─ root self-edit guard when target is root
              ├─ ignore role_id because DTO has no RoleID field
              ├─ enforce company_id cannot move outside tenant scope
              └─ repo.Update(user)
```

### Dedicated role change

```text
PATCH /api/v1/users/:id/role
  └─ RequirePermission("users.change_role")
     └─ users.Handler.ChangeRole
        └─ Validate ChangeUserRoleRequest(role_id public id)
           └─ users.Service.ChangeRole(tc,targetID,actorID,req)
              ├─ reject actorID == targetID
              ├─ targetUser := usersRepo.GetByPublicID(tc,targetID)
              ├─ role := roleAssignmentReader.FindAssignableByPublicID(tc,req.RoleID)
              ├─ reject role.Slug == "root"
              ├─ reject targetUser.CompanyID != role.CompanyID when both are non-nil
              ├─ targetUser.RoleID = role.InternalID
              ├─ usersRepo.Update(targetUser)
              └─ cache.Delete("rbac:permissions:" + targetUser.PublicID)
```

### Roles select

```text
GET /api/v1/roles/select
  └─ RequirePermission("roles.select")
     └─ roles.Handler.ListSelect
        └─ roles.Service.ListSelect(tc)
           └─ rolesRepo.ListSelect(tc)
              ├─ tenant.ApplyTenantScope(...)
              └─ WHERE slug <> 'root'
```

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/users/dto.go` | Modify | Remove `RoleID` from `UpdateUserRequest`; remove required role validation; add `ChangeUserRoleRequest` with string `RoleID`. |
| `internal/modules/users/routes.go` | Modify | Change `PUT /:id` to only require `users.update`; add `PATCH /:id/role` requiring `users.change_role`. |
| `internal/modules/users/handler.go` | Modify | Add `ChangeRole`; parse request, validate, pass actor PublicID. |
| `internal/modules/users/service.go` | Modify | Remove role assignment from `Update`; add `ChangeRole`; add cache invalidation option; tighten company movement; enforce target user/role company match. |
| `internal/modules/users/service_test.go` | Modify | Cover update preserving role, dedicated role assignment, root/self/cross-tenant/root-global company guards, cache invalidation. |
| `internal/modules/users/handler_test.go` | Modify | Cover new handler and update DTO behavior. |
| `internal/modules/users/routes_test.go` | Modify/Add | Assert route permission changes. |
| `internal/modules/roles/dto.go` | Modify | Add `RoleSelectResponse`. |
| `internal/modules/roles/repository.go` | Modify | Add `ListSelect(tc)` or equivalent scoped query excluding `root`. |
| `internal/modules/roles/service.go` | Modify | Add `ListSelect(tc)`. |
| `internal/modules/roles/handler.go` | Modify | Add `ListSelect`. |
| `internal/modules/roles/routes.go` | Modify | Register `GET /select` before `GET /:id`. |
| `internal/modules/roles/*_test.go` | Modify | Cover roles select route/service/repository behavior. |
| `internal/modules/permissions/model.go` | Modify | Add `ActionSelect = "select"`. |
| `seeds/permissions.go` | Modify | Seed `roles.select` with system flag and display order. |
| `seeds/permissions_test.go` | Modify | Update expected count and assert `roles.select`. |
| `internal/app/container.go` | Modify | Inject role lookup/list and cache dependencies needed by users service if not already available. |

## Interface Contracts

```go
type ChangeUserRoleRequest struct {
    RoleID string `json:"role_id"`
}

// internal/platform/response/response.go
type SelectResponse struct {
    ID   string         `json:"id"`
    Name string         `json:"name"`
    Meta map[string]any `json:"meta,omitempty"`
}
```

Suggested service additions and constructor adjustments:

```go
// users.Service
ChangeRole(tc tenant.TenantContext, targetPublicID string, actorPublicID string, req ChangeUserRoleRequest) (*UserResponse, error)

// roles.Service
ListSelect(tc tenant.TenantContext) ([]response.SelectResponse, error)
```

To prevent breaking existing constructors in other modules and unit tests, `users.NewService` will use the **Functional Options pattern** (`ServiceOption`) to inject optional collaborators like the cache and the role reader contract:

```go
// internal/modules/users/service.go
type ServiceOption func(*userService)

func NewService(repo Repository, hasher PasswordHasher, resolver RoleResolver, opts ...ServiceOption) Service

func WithCache(c cache.Cache) ServiceOption
func WithRoleReader(r roles.AssignmentRoleReader) ServiceOption
```

Users service needs a roles-owned assignment reader contract. It MUST NOT depend on the roles GORM model or repository implementation.

```go
// internal/modules/roles/contracts.go
type AssignmentRoleSummary struct {
    InternalID uint
    PublicID   string
    Slug       string
    CompanyID  *uint
}

type AssignmentRoleReader interface {
    FindAssignableByPublicID(tc tenant.TenantContext, publicID string) (*AssignmentRoleSummary, error)
}
```

The roles module implements this contract using its repository/service internals. `internal/app/container.go` injects the contract into users service.

## Error Mapping

| Case | Error |
|---|---|
| Missing/empty `role_id` | 422 validation |
| Target user not visible in tenant scope | 404 |
| Target role not visible in tenant scope | 404 |
| Target user and role company mismatch | 403 or 422 |
| Assign `root` | 403 preferred, 422 acceptable by spec |
| Self role change | 403 |
| Missing `users.change_role` | 403 via middleware |
| Missing `roles.select` | 403 via middleware |
| Cross-company general update move | 403 or 422 |

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Route | `PUT /users/:id` permission no longer includes `users.change_role`; new `PATCH /users/:id/role`; `GET /roles/select` before `/:id` | Existing route test style or httptest router. |
| Handler | Request binding/validation for change role and roles select | Fake services, assert calls and error responses. |
| Users service | Update preserves role; role change success; no self-change; no root assignment; cross-tenant role not found; root/global company mismatch; cache invalidation | Table-driven tests with fake user repo, fake role assignment reader, fake cache. |
| Roles service/repository | List select options excludes root and applies tenant context | Fake repository and/or GORM tests. |
| Seeds | `roles.select` exists and seed remains idempotent | Existing seed tests updated. |
| Integration-ish | Unauthorized role change returns 403; general update with `role_id` does not mutate | httptest where current suite supports middleware. |

## Migration / Rollout

No database schema migration is expected. Permission seed changes are additive and idempotent. Existing clients sending `role_id` to `PUT /users/:id` will no longer change roles and should migrate to `PATCH /users/:id/role`.

## Resolved Design Decisions

- **Same-company Tenant Enforcement for `company_id`**: For non-root users, any attempt to change `company_id` to a different company during `PUT /users/:id` general update must be rejected with `ErrForbidden` or ignored (forced to `tc.CompanyID`). This blocks cross-company tenant movements and resolves adjacent tenant escalation risks.
- **No Pagination for Roles Select**: Roles select options do not need pagination because role-selection catalog lists are small per tenant (normally < 20 roles), making a single lightweight fetch the optimal UX and engineering trade-off.
