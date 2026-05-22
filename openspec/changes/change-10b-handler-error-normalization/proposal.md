# Proposal: Normalize handler errors and list metadata

## Intent

Make API handlers consistently delegate known service errors and paginated-list response metadata to shared platform helpers. This removes duplicated `apperror.Status` switches and aligns list responses with `docs/api-conventions.md`.

## Scope

### In scope

- Replace handler-local service error switches/wrappers with `response.HandleError` where they do not add module-specific behavior.
- Keep module-specific wrappers only when they convert domain errors into field-level validation errors.
- Use `query.ListFromGin` and `response.PaginatedWithFilters` for paginated list endpoints in roles and permissions.
- Standardize request validation short-circuiting on `response.RespondIfInvalid`.
- Add or update tests for changed handler contracts and service signatures.

### Out of scope

- Removing direct `response.NotFound`, `response.Forbidden`, or similar helpers from valid non-service-error paths.
- Changing repository query semantics for search, sort, or filters.
- Removing public response helper functions solely because touched handlers no longer use them.
- Reworking users or companies handlers that already follow the convention.

## Affected areas

- `internal/modules/auth/handler.go`
- `internal/modules/permissions/handler.go`
- `internal/modules/permissions/service.go`
- `internal/modules/permissions/handler_test.go`
- `internal/modules/roles/handler.go`
- `internal/modules/roles/service.go`
- `internal/modules/roles/handler_test.go`
- `internal/modules/roles/service_test.go`

## Approach

1. Replace duplicate error mapping in handlers with `response.HandleError`.
2. Preserve `companies.respondError` because it maps duplicate slugs to a `slug` field validation response.
3. Parse list request params with `query.ListFromGin` in roles and permissions.
4. Return list responses through `response.PaginatedWithFilters(c, message, data, params, total)` so `meta.pagination` and `meta.filters` are consistent without unpacking `ListParams` in handlers.
5. Keep repository pagination calls unchanged by unpacking `params.Pagination.Page` and `params.Pagination.PerPage` in services.
6. Use `response.RespondIfInvalid(c, req.Validate())` consistently for request validation.

## Success criteria

- [ ] No handler contains `switch apperror.Status(err)`.
- [ ] Service errors in auth, roles, and permissions use `response.HandleError`.
- [ ] Local error wrappers are removed unless they add module-specific field mapping.
- [ ] Roles and permissions paginated list endpoints return pagination and filters metadata.
- [ ] Obsolete `apperror` imports are removed from handlers.
- [ ] Tests cover the changed contracts.
- [ ] Request validation uses `response.RespondIfInvalid` consistently in handlers.
- [ ] `go test ./...` passes.
- [ ] `go build ./...` passes.
