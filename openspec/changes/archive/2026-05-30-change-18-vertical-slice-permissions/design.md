# Design: Permissions Vertical Slice Migration

## Technical Approach

Migrate only `internal/modules/permissions/` from flat root files to a module container plus vertical slices. Preserve the three currently registered endpoints, the auth resolver, and startup sync behavior. Existing user-assisted slice files are the starting point. No unified IAM pivot; that remains future work.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Scope boundary | Keep all new implementation under `internal/modules/permissions/`; update only `internal/app/container.go` for wiring | Move users/roles/permissions into IAM | User explicitly scoped change 18 to permissions only; unified IAM is larger and riskier. |
| Slice mapping | HTTP slices only for registered routes; internal slices for resolver/sync | Create slices for `Create`, `Delete`, `ListPaginated` | Routes currently register only list/view/update. Internal non-HTTP methods still need isolated behavior because middleware/bootstrap call them. |
| Shared code | Put DTO/model/contracts in `core/`; reusable DB reads in `queries/` | Keep root `model.go`, `dto.go`, `repository.go` | Root should become composition/wiring only; queries are testable and reusable by slice repositories. |
| Role compatibility | Add temporary root compatibility aliases only if needed by existing roles import; do not change roles behavior | Refactor roles now to local partial permission model | Resolving `roles -> permissions` import is out of scope, but deleting root flat files must not break compilation. |

## Data Flow

HTTP:

    app.Container -> permissions.Container -> Register -> slice Handler -> Service -> Repository -> queries/GORM

Internal:

    middleware.AttachPermissions -> permissions.Container.Resolver -> resolve_permissions.Service -> cache/GORM
    app.Container.SyncPermissions -> permissions.Container.Syncer -> sync_permissions.Service -> GORM/platform permissions

## Endpoint / Method Slice Mapping

| Surface | Slice | Contract preserved |
|---|---|---|
| `GET /permissions` | `list_permissions` | grouped by module, sorted by module/display_order/slug |
| `GET /permissions/:id` | `view_permission` | lookup by `public_id`, `ErrNotFound` mapping |
| `PUT /permissions/:id` | `update_permission` | update name/description/display_order response |
| `Resolve(publicID)` | `resolve_permissions` | cache key `rbac:permissions:{publicID}`, 5 minute TTL, ordered slugs |
| `SyncPermissions(slugs)` | `sync_permissions` | idempotent create, system metadata, auto-assign admin roles |

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/permissions/container.go` | Create | Module composition root: instantiate slice repositories/services/handlers plus Resolver/Syncer/CatalogReader. |
| `internal/modules/permissions/routes.go` | Modify | `Register(v1, container, requirePermission)` maps the 3 registered endpoints to slice handlers. |
| `internal/modules/permissions/core/contracts.go` | Create | Define `Resolver`, `Syncer`, `PermissionCatalogReader`; no repository dependency exposed. |
| `internal/modules/permissions/core/{model,dto,enums}.go` | Update | Keep shared module vocabulary; add missing errors/constants only if needed. |
| `internal/modules/permissions/queries/*.go` | Create/Update | `GetByPublicID`, `GetBySlug`, `ListAll`, `ListSlugsByUserPublicID`, `AutoAssignToAdmins` as reusable query functions with tests. |
| `list_permissions/`, `view_permission/`, `update_permission/` | Update | Complete partial handlers/services/repositories, fix `view_permission` package name, add tests. |
| `resolve_permissions/`, `sync_permissions/` | Create | Internal non-HTTP slices with service/repository tests. |
| `internal/modules/permissions/{handler,service,repository,model,dto}.go` | Delete/replace | Remove flat implementation after slice wiring passes; keep only temporary alias/compat file if compilation requires it. |
| `internal/app/container.go` | Modify | Store `*permissions.Container`; pass resolver to middleware and syncer to `SyncPermissions`. |

## Interfaces / Contracts

```go
type Resolver interface { Resolve(publicID string) ([]string, error) }
type Syncer interface { SyncPermissions(slugs []string) error }
type PermissionCatalogReader interface { ListAll() ([]core.Permission, error) }
```

No permissions slice may import another module repository. Direct SQL/GORM queries against `permissions`, `role_permissions`, `users`, or `roles` are allowed only inside permissions repositories/queries when preserving existing behavior.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | grouping, not-found mapping, update conflict mapping, resolver cache behavior, sync slug parsing | Table-driven tests with fake repositories/cache. |
| Repository/query | ordering, public_id/slug lookup, user permission slugs, admin auto-assignment | GORM test DB helpers if available; otherwise focused integration-style package tests. |
| HTTP/routes | three registered routes guarded by `permissions.manage`; POST/DELETE remain 404 | `httptest` route tests updated for container signature. |
| Full suite | compile and behavior preservation | `go test ./internal/modules/permissions/...`, then `go test ./...`. |

## Migration / Rollout

No DB migration required. Roll out by wiring container while flat files still exist, migrate tests to slices, then delete flat files. Rollback is `git revert`; no persisted data or API contract changes.

## Risks

- Temporary root compatibility may be required because `roles/service.go` imports `permissions.Permission`; fully removing that coupling belongs to a future change.
- `view_permission` currently uses package name `view_permissions`; fix before build.
- Existing main permission spec lists CRUD endpoints not currently registered; this change preserves runtime behavior, not that broader desired spec.

## Open Questions

- None blocking.
