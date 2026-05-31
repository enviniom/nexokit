# Delta for API Response

## MODIFIED Requirements

### Requirement: Explicit response DTO names

The system MUST expose and document these response DTOs: `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, and `PaginationMeta`. `ValidationErrorResponse.Errors` MUST use `validator.ValidationErrors` as its field type.

(Previously: `ValidationErrors` type sourced from `platform/response`; now sourced from `platform/validator`)

#### Scenario: Base response DTOs are available

- GIVEN a handler builds a success, error, validation, or paginated response
- WHEN it uses `platform/response`
- THEN the response contract maps to one of the explicit DTO names

#### Scenario: Validation errors remain field keyed

- GIVEN field validation fails for `email` and `password`
- WHEN a validation response is rendered
- THEN `errors` is an object keyed by field names with message arrays
