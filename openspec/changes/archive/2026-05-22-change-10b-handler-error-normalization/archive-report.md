# Archive report: Handler error and list metadata normalization

## Status

Archived.

## Commits

- `82666f1 refactor(api): normalize handler responses`
- `afb8af8 docs(sdd): record handler normalization verification`

## Summary

- Normalized service error handling in auth, roles, and permissions handlers with `response.HandleError`.
- Standardized validation short-circuiting with `response.RespondIfInvalid`.
- Migrated roles and permissions paginated list responses to `query.ListFromGin` and `response.PaginatedWithFilters`.
- Updated `response.PaginatedWithFilters` to accept `query.ListParams` directly.
- Added permissions handler tests and updated roles/response tests.
- Removed tracked `.atl/*` files from the Git index; local `.atl/` remains ignored.

## Verification

- `go test ./internal/platform/response/... ./internal/modules/...` — PASS
- `go test ./...` — PASS
- `go build ./...` — PASS

## Notes

- `companies.respondError` remains intentionally because it maps `ErrDuplicateSlug` to a field-level `slug` validation response before falling through to `response.HandleError`.
- This change had no canonical spec delta to sync; it normalized implementation and API conventions already documented in `docs/api-conventions.md`.
