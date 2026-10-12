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
- Newly declared permissions automatically grant capability to matching admin roles. Review that security consequence before adding guards.
- There is no standalone `permissions sync` CLI command; route guards, not manual catalogs or a frontend, are the discovery source.

Sources: `internal/platform/permissions/registry.go`, `internal/middleware/authorization.go`, `internal/app/bootstrap.go`, `container.go`, IAM sync service/repository.
