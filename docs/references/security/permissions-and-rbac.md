# Permission identities and RBAC guards

Authorization uses the authenticated `authctx.User` and route-declared guards. Permission slugs and error codes are distinct: `users.update` gates a route, while a module error code identifies an outcome. Standard error responses do not expose a machine code field.

## Guard behavior

| Guard | Effect |
|---|---|
| `RequirePermission(slug)` | Registers slug at route construction; 401 without actor; permits root or matching permission (`*` also matches); otherwise 403 |
| `RequireRole(slug)` | 401 without actor; requires exact role slug (fallback role field); otherwise 403 |

Root bypass applies to permission guards, **not** arbitrary role guards. Role guards for `root` protect onboarding and most company administration. IAM user/role/permission routes use permission guards. Current IAM auth lookup loads role permissions into the actor; optional `AttachPermissions` is not mounted by the app container.

## Naming and new route path

Use `permissions.Format(module, action)` to construct `<module>.<action>` identities. Current examples include `companies.view`, `users.list`, `roles.select`. Central action constants include view/list/select/export/create/update/delete/manage and IAM assignment actions; do not invent a separate colon-delimited convention.

1. Define/reuse meaningful module and action constants.
2. Mount the prepared handler with a permission or exact-role guard.
3. Add route/auth tests for absent actor, allowed actor and denial.
4. Verify startup [permission synchronization](permissions-sync.md) and role assignments.

Use view/list/select/export for read capabilities and create/update/delete for mutations; HTTP method alone does not replace authorization. No admin frontend or permission-matrix component is shipped as part of this contract.

Permissions govern capability access; tenant selection separately limits data scope. A permission grant never replaces [repository isolation](../infrastructure/multitenancy.md#persistence-boundary).

Sources: `internal/middleware/authorization.go`, `internal/platform/permissions`, module `routes.go`, IAM `internal/resolve_auth_user`.
