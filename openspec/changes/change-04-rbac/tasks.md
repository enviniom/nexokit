# Tasks: RBAC Permissions and Authorization

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 800-1,100 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 permission foundation → PR 2 role assignment/cache → PR 3 guards/responses |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Permission model, migration, seeds, grouped catalog | PR 1 | Independent foundation; no frontend work. |
| 2 | Role-permission GET/PUT, replacement rules, cache invalidation | PR 2 | Depends on PR 1; stacked-to-main. |
| 3 | Auth context permissions, middleware, route guards, `/auth/me` | PR 3 | Depends on PR 2; includes route guard tests. |

## Phase 1: Permission Foundation

- [x] 1.1 RED: add table-driven tests for permission fields, slug uniqueness, explicit actions, grouped ordering, and system CRUD protection.
- [x] 1.2 Create `migrations/20260518000000_rbac.sql` for `permissions`/`role_permissions` with indexes, FKs, and rollback order.
- [x] 1.3 Create `internal/modules/permissions/{model,repository,service,dto,handler,routes}.go` with `module`, `action`, `slug`, `name`, `description`, `is_system`, `display_order`.
- [x] 1.4 Add idempotent `seeds/{permissions,role_permissions}.go` using `index/view/list/create/update/delete` plus business actions; root gets all permissions.

## Phase 2: Role Assignment API / Cache

- [x] 2.1 RED: add handler/service tests for `GET /api/v1/roles/:id/permissions` grouped catalog with `granted` flags and 404 role missing.
- [x] 2.2 RED: add tests for `PUT /api/v1/roles/:id/permissions` exact slug replacement, invalid slug rejection, unauthorized assignment, and system role protection.
- [x] 2.3 Modify `internal/modules/roles/{model,repository,service,dto,handler,routes}.go` for permission preloads, read-only role DTO slugs, catalog GET, and assignment PUT.
- [x] 2.4 Add resolver/cache logic in `internal/modules/permissions/service.go` and role-member lookup in `internal/modules/users/repository.go`; delete `rbac:permissions:{public_id}` on assignment success.

## Phase 3: Authorization Wiring

- [ ] 3.1 RED: add `internal/middleware/authorization_test.go` for 401, 403, root bypass, role match/mismatch, cache miss, and resolver failure.
- [ ] 3.2 Modify `internal/platform/authctx/authctx.go` with `RoleSlug string` and `Permissions []string`; update call sites.
- [ ] 3.3 Create `internal/middleware/authorization.go` with `PermissionResolver`, `AttachPermissions`, `RequirePermission`, and `RequireRole`.
- [ ] 3.4 Wire permissions, resolver, role assignment, and guards in `internal/app/container.go`; keep API-only scope.

## Phase 4: Route Guards / Verification

- [ ] 4.1 RED: add route tests for guarded users, roles, permissions, `/api/v1/auth/me` permission slugs, and standard 401/403 envelopes.
- [ ] 4.2 Modify `internal/modules/auth/{handler,dto,routes}.go` so `/auth/me` returns role data and permission slugs from context.
- [ ] 4.3 Apply route permissions in `internal/modules/users/routes.go`, `internal/modules/roles/routes.go`, and `internal/modules/permissions/routes.go` per design taxonomy.
- [ ] 4.4 Run `go test ./...` and `go build ./...`; fix failures without weakening RBAC scenarios.
