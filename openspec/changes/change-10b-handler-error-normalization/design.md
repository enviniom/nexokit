# Design: Shared handler error and list response conventions

## Technical approach

Use the platform packages as the single HTTP convention boundary:

- `response.HandleError(c, err)` handles known service errors.
- `query.ListFromGin(c)` parses standard list request parameters.
- `response.PaginatedWithFilters(c, message, data, params, total)` serializes list metadata consistently from `query.ListParams`.
- `response.RespondIfInvalid(c, req.Validate())` centralizes the validation short-circuit pattern.

Handlers stay thin: bind/validate input, call the service, and delegate response formatting to platform helpers.

## Decisions

| Decision | Alternatives considered | Rationale |
|---|---|---|
| Replace local error switches with `response.HandleError` | Keep per-handler `apperror.Status` switches | Shared mapping already exists; local switches duplicate behavior and drift from API conventions. |
| Keep `companies.respondError` | Move duplicate slug mapping into `HandleError` | Duplicate slug is module-specific and should remain a field-level `slug` response, not a global error rule. |
| Services accept `query.ListParams` | Keep handlers unpacking `page`/`per_page`; push params to repository | Passing `ListParams` keeps handler/service contracts ready for standard list metadata while avoiding repository behavior changes in this change. |
| `PaginatedWithFilters` accepts `query.ListParams` | Keep decomposed pagination/filter/sort/search parameters | The helper belongs with the query abstraction; unpacking fields at every call site repeats knowledge and makes handlers noisier. |
| Standardize on `RespondIfInvalid` | Keep explicit `errs := req.Validate()` blocks | The helper expresses the convention directly and prevents drift in validation response handling. |
| Keep direct response helpers available | Delete helpers not used by touched files | Helper cleanup is broader API maintenance and not required for this normalization. |

## Response flow

```text
Request ─→ Handler ─→ Service
              │          │
              │          └─ returns domain/platform error
              │
              ├─ response.HandleError(c, err)
              └─ response.PaginatedWithFilters(..., params, total)
```

## List contract

For roles and permissions paginated endpoints, the response metadata should include:

```json
{
  "meta": {
    "pagination": { "page": 1, "per_page": 10, "total": 0, "total_pages": 0 },
    "filters": {}
  }
}
```

The change is additive for API consumers that already consumed `meta.pagination`.

## Testing strategy

| Layer | What to verify |
|---|---|
| Handler | Error statuses still map through shared `HandleError`; paginated responses include `filters`. |
| Service | `ListParams.Pagination` is correctly unpacked into existing repository calls. |
| Regression | Existing auth and roles behavior remains unchanged except additive list metadata. |
