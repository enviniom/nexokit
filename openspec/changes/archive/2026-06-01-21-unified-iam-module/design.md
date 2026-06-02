# Design: Unified IAM Module

## Technical Approach

Create `internal/modules/iam/` as the single IAM/RBAC boundary and copy behavior from legacy `users`, `roles`, and `permissions` without deleting or modifying those modules. IAM uses the multi-entity vertical slice convention because users, roles, and permissions each expose more than three use cases. Public routes, payloads, status codes, cache keys, permission slugs, and middleware contracts remain unchanged; only root wiring moves from three legacy modules to one IAM container.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Module shape | Multi-entity vertical slice with `users/`, `roles/`, `permissions/`, plus non-HTTP `internal/` slices | Flat IAM slices; single huge container | Matches `_context.md` multi-entity rule and keeps 19 HTTP endpoints reviewable. |
| Model ownership | IAM defines local partial GORM models in `iam/core/model.go` | Import legacy models; move files destructively | Enforces zero cross-module imports while keeping migrations as schema source. |
| Query reuse | `iam/queries/` is demand-driven: one file per query and only for queries reused by more than one slice/repository; single-use queries stay inside the owning slice repository | Put all reads in a broad shared layer; put reusable queries in services/core | Keeps reuse explicit, avoids speculative abstractions, and keeps business rules out of `core`. |
| Coexistence | Legacy modules remain untouched and compilable, but app wiring mounts IAM only | Delete legacy now; dual-register routes | Supports review/rollback and avoids duplicate route behavior. |

## Data Flow

```txt
HTTP /api/v1/* ──→ iam.Register ──→ entity routes ──→ slice handler ──→ service ──→ repository ──→ iam/queries ──→ DB

Auth token ──→ middleware.Auth ──→ app.userLookup ──→ IAM.ResolveAuthUser ──→ internal/resolve_auth_user ──→ authctx.User

Authz ──→ AttachPermissions ──→ IAM.ResolvePermissions ──→ cache rbac:permissions:{publicID} ──→ DB fallback

Bootstrap ──→ app.SyncPermissions ──→ IAM.SyncPermissions ──→ internal/sync_permissions ──→ permissions + role_permissions
```

## Final File Structure

| Path | Action | Description |
|---|---|---|
| `internal/modules/iam/container.go` | Create | Root container exporting `Users`, `Roles`, `Permissions`, `Resolver`, `Syncer`, `AuthUserResolver`, `RoleResolver`. |
| `internal/modules/iam/routes.go` | Create | Delegates route registration to entity route files. |
| `internal/modules/iam/core/{model,dto,error,constants,contracts}.go` | Create | Local IAM models/DTOs/domain errors/constants/contracts. |
| `internal/modules/iam/queries/` | Create | Query reuse boundary with one-file-per-query convention; add files only when reuse is proven across slices. |
| `internal/modules/iam/users/{container,routes}.go` + slices | Create | `list_users`, `create_user`, `view_user`, `update_user`, `delete_user`, `change_user_password`, `assign_role_to_user`, `toggle_user_status`. |
| `internal/modules/iam/roles/{container,routes}.go` + slices | Create | `list_roles`, `list_selectable_roles`, `view_role`, `create_role`, `update_role`, `delete_role`, `view_role_permission_catalog`, `assign_permissions_to_role`. |
| `internal/modules/iam/permissions/{container,routes}.go` + slices | Create | `list_permissions`, `view_permission`, `update_permission`. |
| `internal/modules/iam/internal/*` | Create | `resolve_auth_user`, `resolve_user_permissions`, `sync_permissions`, `resolve_role_by_slug`, `list_all_permissions`. |
| `internal/app/container.go` | Modify | Replace legacy IAM fields/imports with `IAM *iam.Container`; delegate adapters to IAM. |
| `internal/modules/{users,roles,permissions}/` | Preserve | Legacy reference modules remain unmodified. |

## Interfaces / Contracts

`iam/core/contracts.go` exposes the stable app-facing contracts: `ResolveAuthUser(publicID string) (*authctx.User, error)`, `Resolve(publicID string) ([]string, error)`, `SyncPermissions(slugs []string) error`, `ResolveRoleBySlug(slug string) (*core.Role, error)`, and `ListAllPermissions() ([]core.Permission, error)`. DTO JSON fields must mirror legacy DTOs exactly, including current `role_id uint` and `company_id` behavior.

## Repository / Service / Handler Responsibilities

Handlers bind/validate with `platform/validator`, derive `tenant.TenantContext`/actor, and respond through `platform/response`. Services own parity business rules: tenant scoping, root protection, reserved slug checks, system-role/permission protection, cache invalidation, and selection normalization. Repositories perform slice-specific persistence and call `iam/queries` for reusable reads/writes; reusable queries get their own tests.

## App Wiring, Adapters, and Aliases

`app.NewContainer` calls `iam.NewContainer(db, cache, log)` and stores `IAM`. `RegisterModules` mounts companies/onboarding as today, then `iam.Register(globalProtected, c.IAM, tenantProtected, middleware.RequirePermission, middleware.RequireRole)`. Transitional app adapters keep middleware signatures: `userLookup.GetAuthUser` calls `c.IAM.ResolveAuthUser`, `roleResolverAdapter.GetBySlug` calls `c.IAM.ResolveRoleBySlug`, and `SyncPermissions` calls `c.IAM.SyncPermissions`. No legacy route is registered.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Slice unit | Every handler/service/repository preserves legacy behavior | Port legacy tests into IAM slices; add cache invalidation cases. |
| Query | Any query promoted to `iam/queries/` after proven reuse | One `_test.go` per promoted query file (or equivalent focused query tests). |
| Integration/contract | 19 endpoints, middleware auth/authz, bootstrap sync, legacy compile | httptest route tests plus `go build ./internal/modules/users/... ./internal/modules/roles/... ./internal/modules/permissions/...`. |
| Full suite | Regression | `go test ./...` and `go build ./...`. |

## Migration / Rollout

No DB migration required. Rollout is a container swap: add IAM, wire IAM, stop mounting legacy modules. Rollback is reverting `internal/app/container.go` route/wiring changes; IAM code can remain unused with no data loss.

## Open Questions

- [ ] Confirm whether current permissions routes intentionally use `permissions.manage` for list/view instead of auth-only spec wording; implementation should preserve current behavior unless a later task explicitly changes it.
