# Permission identities and RBAC guards

Authorization uses the authenticated `authctx.User` and route-declared guards. Permission slugs and error codes are distinct: `users.update` gates a route, while a module error code identifies an outcome. Standard error responses do not expose a machine code field.

## Guard behavior

| Guard | Effect |
|---|---|
| `RequirePermission(slug)` | Registers slug at route construction; 401 without actor; permits root, trusted tenant admin in its own non-global scope, or matching permission (`*` also matches); otherwise 403 |
| `RequireRole(slug)` | 401 without actor; root requires trusted `IsRoot`; other roles require exact role slug (fallback role field); otherwise 403 |

Root bypass applies to permission guards, **not** arbitrary role guards. Role guards for `root` protect onboarding and most company administration. IAM user/role routes and permission catalog reads use permission guards; catalog updates use a root role guard. Current IAM auth lookup loads role permissions into the actor; optional `AttachPermissions` is not mounted by the app container.

## Tenant administrator authority

Admin authority uses the exact database-hydrated `RoleSlug == "admin"`, never display names or the legacy `Role` field. It requires a resolved, non-global tenant context with a nonzero company ID equal to the actor's company ID. Missing, global or foreign scope cannot enable the bypass. Other roles still require assigned permissions; root retains its independent permission bypass.

Tenant admins can assign or remove grants on `user` and any other own-tenant role without explicit grants, including system roles and synchronized system permissions. Only target `root` and `admin` grants are immutable for every actor because their authority is implicit. Reserved `root`, `admin`, and `user` identities still retain API creation/rename/delete protections; identity reservation does not imply immutable grants. Assignment services validate target policy and catalog selection, then persist tenant-scoped changes; route middleware authorizes the actor. Company-view repositories respect root's explicit tenant selection and return not found for foreign companies; no-header root remains global on global-capable routes. Root-only company operations remain root-only.

Startup admin autogrants are retained for compatibility, **not** the source of tenant-admin authority.

## Global permission catalog

Permission catalog rows are global, unlike tenant-owned role assignments. `PUT /permissions/:id` requires a trusted authenticated root actor through route-declared `RequireRole(roles.RootRoleSlug)` middleware, regardless of assigned permissions or tenant-admin bypass. Non-root actors cannot mutate catalog rows even with `permissions.manage`; root may update non-system rows, while system rows remain immutable. Catalog list/view remain available through permission guards so tenant admins can inspect permissions when managing their own custom roles.

All application role constants (`RootRoleSlug`, `AdminRoleSlug`, `UserRoleSlug`, `ReservedSlugs`) and normalized reserved-identity policy live in `internal/platform/roles`. Consumers import that package directly; IAM and authctx define no role aliases. Middleware consumes the authenticated projection without importing IAM.

## Naming and new route path

Use `permissions.Format(module, action)` to construct `<module>.<action>` identities. Current examples include `companies.view`, `users.list`, `roles.select`. Central action constants include view/list/select/export/create/update/delete/manage and IAM assignment actions; do not invent a separate colon-delimited convention.

1. Define/reuse meaningful module and action constants.
2. Mount the prepared handler with a permission or exact-role guard.
3. Add route/auth tests for absent actor, allowed actor and denial.
4. Verify startup [permission synchronization](permissions-sync.md) and role assignments.

Use view/list/select/export for read capabilities and create/update/delete for mutations; HTTP method alone does not replace authorization. No admin frontend or permission-matrix component is shipped as part of this contract.

Permissions govern capability access; tenant selection separately limits data scope. A permission grant never replaces [repository isolation](../infrastructure/multitenancy.md#persistence-boundary).

Sources: `internal/middleware/authorization.go`, `internal/platform/permissions`, module `routes.go`, IAM `internal/resolve_auth_user`.
