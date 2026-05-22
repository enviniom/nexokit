# Exploration: Handler error and list metadata normalization

## Outcome

Normalize HTTP handlers so service errors, paginated lists, and filter metadata follow the shared API conventions instead of module-local branches.

## Inputs reviewed

- `docs/prompts/change_10b_handler_error_normalization.md`
- `docs/prompts/_context.md`
- `docs/api-conventions.md`
- `internal/modules/users/handler.go`
- `internal/modules/companies/handler.go`
- `internal/modules/roles/handler.go`
- `internal/modules/permissions/handler.go`
- `internal/modules/auth/handler.go`
- `internal/platform/response/response.go`
- `internal/platform/query/query.go`

## Findings

| Area | Finding | Decision |
|---|---|---|
| Error mapping | `response.HandleError` delegates to the shared `apperror.Status`/public message behavior and covers the local handler switches in `auth`, `roles`, and `permissions`. | Replace duplicated local switches/wrappers with `response.HandleError`. |
| Companies duplicate slug | `companies.respondError` maps `ErrDuplicateSlug` to a field-level `slug` validation error before falling through to `HandleError`. | Keep it; this is the allowed module-specific exception. |
| Users list | Already uses `query.ListFromGin` and `response.PaginatedWithFilters`. | No change required. |
| Companies list | Already uses common query/response helpers and only adds module-specific `include_inactive`. | No change required. |
| Roles list | Uses `query.PaginationFromGin` + `response.Paginated`, so it lacks filters metadata. | Migrate to `query.ListFromGin` + `response.PaginatedWithFilters`. |
| Permissions paginated list | Uses `query.PaginationFromGin` + `response.Paginated`, so it lacks filters metadata. | Migrate to `query.ListFromGin` + `response.PaginatedWithFilters`. |
| Validation responses | `companies` uses `response.RespondIfInvalid`; auth, users, roles, and permissions used the expanded `errs := req.Validate()` form. | Standardize request validation on `response.RespondIfInvalid` for concise shared behavior. |

## Scope map

| File | Needed work |
|---|---|
| `internal/modules/auth/handler.go` | Remove `respondError`; use `response.HandleError` in `Login`, `Refresh`, and `Logout`. |
| `internal/modules/permissions/handler.go` | Remove `writePermissionError`; use `response.HandleError`; migrate `ListPaginated` to list params and filters metadata. |
| `internal/modules/permissions/service.go` | Change `List` to accept `query.ListParams` and pass pagination to the repository. |
| `internal/modules/roles/handler.go` | Replace `Delete` error switch; migrate list response to filters metadata. |
| `internal/modules/roles/service.go` | Change `List` to accept `query.ListParams` and pass pagination to the repository. |
| `internal/modules/roles/handler_test.go` | Update fake service signature and assert filters metadata. |
| `internal/modules/roles/service_test.go` | Update `List` call sites to pass `query.ListParams`. |
| `internal/modules/permissions/handler_test.go` | Add focused coverage for changed permissions handler behavior. |
| `internal/modules/users/handler.go` | Standardize request validation on `response.RespondIfInvalid`. |
| `internal/platform/response/response.go` | Simplify `PaginatedWithFilters` to accept `query.ListParams` instead of decomposed query fields. |

## Risks

- `roles` and `permissions` list responses gain additive `meta.filters` data.
- Service `List` signatures change, requiring test doubles and direct call sites to be updated.
- `query.ListParams` currently passes only pagination deeper into services; search/sort/filter parsing is exposed in HTTP metadata but not applied to repository queries by this change.
