# API Response Specification

## Purpose
Define the standard JSON envelope used by API responses, including success, error, and paginated variants. Successful no-content operations are the explicit exception and return HTTP 204 without a JSON body.

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

The system MUST render absent `data` as `null` (not omitted) and absent `meta` as `null`. When `PaginatedWithFilters` is used, `meta.filters` MUST always be present (not null), containing at minimum the sort, order, and search defaults.

#### Scenario: Empty success response

- GIVEN no data to return but the operation still returns a success envelope
- WHEN `response.Success(c, "Deleted", nil)` is called
- THEN `data` is `null` in JSON

#### Scenario: PaginatedWithFilters always includes filters

- GIVEN a paginated request with no explicit filters
- WHEN `PaginatedWithFilters` is called
- THEN `meta.filters` is present with default sort, order, and empty search

### Requirement: NoContent response helper

The system MUST provide `NoContent(c *gin.Context)` for successful operations that return HTTP 204 and no response body. This helper MUST NOT emit the standard JSON envelope because HTTP 204 responses cannot carry content.

#### Scenario: Successful no-content response

- GIVEN a delete operation succeeds and has no representation to return
- WHEN `response.NoContent(c)` is called
- THEN the response status is HTTP 204
- AND the response body is empty

### Requirement: Explicit response DTO names

The system MUST expose and document these response DTOs: `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, and `PaginationMeta`. `ValidationErrorResponse.Errors` MUST use `validator.ValidationErrors` as its field type.

#### Scenario: Base response DTOs are available

- GIVEN a handler builds a success, error, validation, or paginated response
- WHEN it uses `platform/response`
- THEN the response contract maps to one of the explicit DTO names

#### Scenario: Validation errors remain field keyed

- GIVEN field validation fails for `email` and `password`
- WHEN a validation response is rendered
- THEN `errors` is an object keyed by field names with message arrays

### Requirement: PaginatedWithFilters response helper

The system MUST provide `PaginatedWithFilters[T any](c *gin.Context, message string, data T, pagination query.PaginationParams, filters query.FilterParams, sort query.SortParams, search query.SearchParams, total int64)` that returns a 200 response with `meta.pagination` AND `meta.filters` containing filter metadata.

#### Scenario: Paginated list with filters and search

- GIVEN a list of 50 items with page=1, per_page=10, status=active, sort=name, order=asc, search=jhon
- WHEN `PaginatedWithFilters` is called
- THEN `meta.pagination` contains page, per_page, total, total_pages
- AND `meta.filters` contains `{"status":"active","sort":"name","order":"asc","search":"jhon"}`

#### Scenario: PaginatedWithFilters with empty filters

- GIVEN no filter or search params
- WHEN `PaginatedWithFilters` is called with empty FilterParams, SortParams, SearchParams
- THEN `meta.filters` contains default sort/order values and empty status/search

### Requirement: HandleError centralized error mapping

The system MUST provide `HandleError(c *gin.Context, err error)` that maps any `AppError` sentinel to its HTTP status and standard response, replacing manual `switch apperror.Status(err)` blocks in handlers.

#### Scenario: ErrNotFound maps to 404

- GIVEN `apperror.ErrNotFound` is passed
- WHEN `HandleError(c, err)` is called
- THEN response status is 404 and message is "Recurso no encontrado"

#### Scenario: ErrValidation maps to 422

- GIVEN `apperror.ErrValidation` is passed
- WHEN `HandleError(c, err)` is called
- THEN response status is 422 and message is "Error de validación"

#### Scenario: Unknown error defaults to 500

- GIVEN an error not matching any sentinel
- WHEN `HandleError(c, err)` is called
- THEN response status is 500 and message is "Error interno del servidor"

#### Scenario: Nil error is no-op

- GIVEN `nil` is passed
- WHEN `HandleError(c, nil)` is called
- THEN no response is written

### Requirement: TooManyRequests response helper

The system MUST provide a `TooManyRequests(c *gin.Context, message string)` helper that returns HTTP 429 with the standard error envelope using `MsgTooManyRequests` as the default message.

#### Scenario: TooManyRequests returns standard envelope

- GIVEN a rate limit is exceeded
- WHEN `TooManyRequests(c, "")` is called
- THEN the response is 429 with `success: false`, `message: MsgTooManyRequests`, `data: null`, `errors: null`

#### Scenario: TooManyRequests with custom message

- GIVEN a rate limit is exceeded
- WHEN `TooManyRequests(c, "Please wait before retrying")` is called
- THEN the response is 429 with `message: "Please wait before retrying"`

## Constraints and Edge Cases

- All handlers MUST use `platform/response`; direct `gin.H` is prohibited.
- Content-Type MUST be `application/json; charset=utf-8` for JSON envelope responses.
- HTTP 204 responses produced by `NoContent` MUST have an empty body and no JSON content type requirement.
- `errors` field MUST be a map of field names to string arrays.
