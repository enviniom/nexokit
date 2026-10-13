# Startup permission discovery and sync

Route-declared permission guards populate a thread-safe in-memory registry. Bootstrap synchronizes the registry **after route construction and before serving**. This reference owns that lifecycle; [RBAC](permissions-and-rbac.md) owns guard behavior.

## Flow

```text
RequirePermission("users.list")
  → platform/permissions.Register
  → Container.SyncPermissions
  → IAM Syncer.SyncPermissions
  → permission insertion + assignment to admin roles
```

`internal/modules/iam/slices/internal/sync_permissions/service.go` processes slugs containing a dot, skipping malformed entries and existing permission records. New entries receive a public ID, humanized name/description, module/action, display order and `IsSystem=true`. The repository inserts a corresponding `role_permissions` row for every role with slug `admin`, using `ON CONFLICT DO NOTHING`.

## Operational consequences

- Sync needs migrated tables and database write privileges at API bootstrap; failure aborts startup.
- Existing entries are left unchanged; this is not a reconciliation job that refreshes labels, removes stale permissions or repairs existing admin assignments.
- Creation and admin assignment are separate calls, not a transaction. If assignment fails after insertion, a later sync skips the existing permission; do not promise automatic recovery from that partial failure.
- Concurrent instance startup is not a serialized migration/provisioning guarantee; inspect errors and role assignments during rollout.
- New permission entries still receive admin assignment rows for compatibility. Those rows are not the source of admin authority: trusted admins bypass permission guards only within their own resolved tenant scope. Sync does not update or remove stale entries; see [RBAC](permissions-and-rbac.md). Review route scope and root-only guards before adding capabilities.
- Runtime global catalog mutation uses root-only route middleware, even when an admin has a synchronized grant; system permission rows remain immutable through service business rules. Catalog reads and tenant-owned role assignments retain their separate authorization policies. Synchronized system grants may be added or removed from `user` and other own-tenant roles, including system roles. Target `root`/`admin` assignments remain immutable for all actors because their authority is implicit; reserved identity protections are separate.
- There is no standalone `permissions sync` CLI command; route guards, not manual catalogs or a frontend, are the discovery source.

Sources: `internal/platform/permissions/registry.go`, `internal/middleware/authorization.go`, `internal/app/bootstrap.go`, `container.go`, IAM sync service/repository.
