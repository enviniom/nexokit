# Tasks: Migrate Permissions Module to Vertical Slice Architecture

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 600–800 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (foundation+queries) → PR 2 (HTTP slices) → PR 3 (internal+wiring+cleanup) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | PR | Notes |
|------|------|-----|-------|
| 1 | `core/`, `queries/`, contracts | PR 1 | Base branch; standalone |
| 2 | 3 HTTP slices + TDD tests | PR 2 | Targets PR 1 branch |
| 3 | Internal slices, container, routes, app wiring, cleanup | PR 3 | Targets PR 2 branch |

## Phase 1: Foundation

- [x] 1.1 Create `core/contracts.go`: `Resolver`, `Syncer`, `PermissionCatalogReader` interfaces; test interface satisfaction.
- [x] 1.2 Create `core/error.go`: `ErrNotFound`, `ErrConflict`, `ErrSystemImmutable` sentinels; test `errors.Is`.
- [x] 1.3 Fix `view_permission/handler.go` package from `view_permissions` to `view_permission`; verify build.
- [x] 1.4 Create `queries/get_permission_by_slug.go`; test with GORM test DB.
- [x] 1.5 Create `queries/list_all_permissions.go` with module/display_order/slug sort; test order.
- [x] 1.6 Create `queries/list_permissions_paginated.go`; test pagination.

## Phase 2: HTTP Slices — TDD

- [x] 2.1 RED `list_permissions/handler_test.go`: grouped-by-module response; GREEN: handler.
- [x] 2.2 RED `list_permissions/service_test.go`: fake repo `ListGrouped()`; GREEN: service.
- [x] 2.3 RED `list_permissions/repository_test.go`: GORM ordered grouped query; GREEN: repo.
- [x] 2.4 RED `view_permission/handler_test.go`: 200/success + 404 missing; GREEN: handler.
- [x] 2.5 RED `view_permission/service_test.go`: not-found → `core.ErrNotFound`; GREEN: service.
- [x] 2.6 RED `view_permission/repository_test.go`: uses shared query; GREEN: repo.
- [x] 2.7 RED `update_permission/handler_test.go`: 200 valid, 404 missing, 409 system conflict; GREEN: handler.
- [x] 2.8 RED `update_permission/service_test.go`: conflict mapping + update; GREEN: service.
- [x] 2.9 RED `update_permission/repository_test.go`: update query + system-flag; GREEN: repo.

## Phase 3: Internal Slices

- [x] 3.1 RED `resolve_permissions/service_test.go`: fake repo+cache, `Resolve()` ordered slugs, 5 min cache; GREEN: service.
- [x] 3.2 RED `resolve_permissions/repository_test.go`: 3-table join `ListSlugsByUserPublicID`; GREEN: repo.
- [x] 3.3 RED `sync_permissions/service_test.go`: idempotent `SyncPermissions()`, admin auto-assign; GREEN: service.
- [x] 3.4 RED `sync_permissions/repository_test.go`: slug upsert + `AutoAssignToAdmins`; GREEN: repo.

## Phase 4: Wiring

- [x] 4.1 Create `container.go`: instantiate all 5 slices, expose handlers+Resolver+Syncer+CatalogReader; test output struct.
- [x] 4.2 Modify `routes.go`: `Register` accepts `*Container`; map 3 endpoints; preserve middleware.
- [x] 4.3 RED `routes_test.go`: 3 routes correct status, POST/DELETE 404; GREEN: routes.
- [x] 4.4 Modify `internal/app/container.go`: replace flat repo/service/handler with `*permissions.Container`; pass Resolver to `AttachPermissions`, Syncer to `SyncPermissions`.

## Phase 5: Cleanup

- [x] 5.1 Delete `handler.go`, `service.go`, `repository.go`, `model.go`, `dto.go`; verify `go build ./...`.
- [x] 5.2 Remove root `handler_test.go`, `service_test.go` (migrated to slices).
- [x] 5.3 Run `go test ./internal/modules/permissions/...` — all pass.
- [x] 5.4 Run `go test ./...` — full suite passes.
