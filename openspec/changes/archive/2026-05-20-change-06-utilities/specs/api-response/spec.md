# Delta for api-response

## ADDED Requirements

### Requirement: Explicit response DTO names

The system MUST expose and document these response DTOs: `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, and `PaginationMeta`.

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

## MODIFIED Requirements

### Requirement: Null semantics

The system MUST render absent `data` as `null` (not omitted) and absent `meta` as `null`. When `PaginatedWithFilters` is used, `meta.filters` MUST always be present (not null), containing at minimum the sort, order, and search defaults.

(Previously: `meta` was either `null` for non-paginated or contained only `pagination`)

#### Scenario: Empty success response

- GIVEN no data to return
- WHEN `response.Success(c, "Deleted", nil)` is called
- THEN `data` is `null` in JSON

#### Scenario: PaginatedWithFilters always includes filters

- GIVEN a paginated request with no explicit filters
- WHEN `PaginatedWithFilters` is called
- THEN `meta.filters` is present with default sort, order, and empty search
