# Review packet: Handler error and list metadata normalization

## Review status

Status: ready for maintainer review before deciding whether to keep the current diff.

This packet was reconstructed because implementation happened before the intended artifact/task checkpoint. Review this alongside the working-tree diff.

## What changed in the working tree

| File | Review focus |
|---|---|
| `internal/modules/auth/handler.go` | `respondError` removed; login/refresh/logout delegate to `response.HandleError`. |
| `internal/modules/permissions/handler.go` | `writePermissionError` removed; service errors use `response.HandleError`; paginated list includes filters metadata. |
| `internal/modules/permissions/service.go` | `List` now accepts `query.ListParams` and passes pagination values to the repo. |
| `internal/modules/permissions/handler_test.go` | New handler coverage for list metadata and error mapping paths. |
| `internal/modules/roles/handler.go` | `Delete` error switch removed; list includes filters metadata. |
| `internal/modules/roles/service.go` | `List` now accepts `query.ListParams` and passes pagination values to the repo. |
| `internal/modules/roles/handler_test.go` | Fake signature updated; list test asserts `filters` metadata. |
| `internal/modules/roles/service_test.go` | `List` call sites updated to pass `query.ListParams`. |
| `internal/modules/users/handler.go` | Request validation standardized on `response.RespondIfInvalid`. |
| `internal/platform/response/response.go` | `PaginatedWithFilters` now receives `query.ListParams` directly. |
| `internal/platform/response/response_test.go` | Helper tests updated for the `query.ListParams` signature. |

## Intentional non-changes

- `internal/modules/users/handler.go` already follows the list/error conventions.
- `internal/modules/companies/handler.go` keeps `respondError` because duplicate slug needs a field-level `slug` validation error.
- `internal/platform/response/response.go` keeps helper functions; deleting them is out of scope.
- `.atl/skill-registry.md` and `.atl/.skill-registry.cache.json` were dirty before this change and should not be included.

## Verification evidence

Reported by the implementation agent:

- `go test ./...` — passed
- `go build ./...` — passed

Fresh review result:

- PASS: no `switch apperror.Status(err)` in handlers.
- PASS: no obsolete `apperror` imports in handlers.
- PASS: roles and permissions paginated lists use shared query/response helpers.

## Maintainer review checklist

- [ ] Confirm additive `meta.filters` on roles/permissions list responses is acceptable.
- [ ] Confirm `query.ListParams` in services is acceptable even though only pagination is currently applied to repository queries.
- [ ] Confirm `PaginatedWithFilters` should receive `query.ListParams` directly.
- [ ] Confirm `response.RespondIfInvalid` is the standard validation pattern for handlers.
- [ ] Confirm permissions handler tests are enough for this change.
- [ ] Confirm `.atl/*` files stay out of the commit.
