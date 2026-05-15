# API Response Specification

## Purpose
Define the standard JSON envelope used by every API response, including success, error, and paginated variants.

## Requirements

### Requirement: Standard envelope structure

The system MUST provide an `APIResponse` struct with fields: `success` (bool), `message` (string), `data` (any), `meta` (any), `errors` (any).

#### Scenario: Success response

- GIVEN a handler returns data
- WHEN `response.Success(c, "OK", data)` is called
- THEN the JSON contains `success: true`, `message: "OK"`, `data` populated, `meta: null`, `errors: null`

#### Scenario: Error response

- GIVEN a validation failure
- WHEN `response.Error(c, http.StatusUnprocessableEntity, "Validation failed", errs)` is called
- THEN the JSON contains `success: false`, `message: "Validation failed"`, `errors` populated, `data: null`

### Requirement: Pagination metadata

The system MUST provide a `Paginated` helper that injects `meta.pagination` with `page`, `per_page`, `total`, `total_pages`.

#### Scenario: Paginated list

- GIVEN a list of 100 items with page=2, per_page=15
- WHEN `response.Paginated(c, "List", items, page, per_page, total)` is called
- THEN `meta.pagination.total_pages` equals 7
- AND `meta.pagination.page` equals 2

### Requirement: Null semantics

The system MUST render absent `data` as `null` (not omitted) and absent `meta` as `null`.

#### Scenario: Empty success response

- GIVEN no data to return
- WHEN `response.Success(c, "Deleted", nil)` is called
- THEN `data` is `null` in JSON

## Constraints and Edge Cases

- All handlers MUST use `platform/response`; direct `gin.H` is prohibited.
- Content-Type MUST be `application/json; charset=utf-8`.
- `errors` field MUST be a map of field names to string arrays.
