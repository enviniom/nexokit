# Delta for Request Validation

## MODIFIED Requirements

### Requirement: ValidationErrors accumulator

The system MUST provide `ValidationErrors map[string][]string` with `Add(field, message)` and `HasErrors() bool` defined in `platform/validator`.

(Previously: Type defined in `platform/response`; now owned by `platform/validator`)

#### Scenario: Accumulate multiple errors

- GIVEN an empty `validator.ValidationErrors`
- WHEN `Add("email", "is required")` and `Add("email", "is invalid")` are called
- THEN `errs["email"]` contains both messages in order

### Requirement: Gin helper

The system MUST provide `RespondIfInvalid(c *gin.Context, errs validator.ValidationErrors) bool` in `platform/response` that writes a 422 response if errors exist.

(Previously: Parameter type was `response.ValidationErrors`; now `validator.ValidationErrors`)

#### Scenario: Validation fails

- GIVEN `errs.HasErrors()` is `true`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it writes status 422 with the standard error envelope
- AND returns `true`

#### Scenario: Validation passes

- GIVEN `errs.HasErrors()` is `false`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it returns `false` and writes nothing

### Requirement: Field-keyed validation responses

The system MUST return validation failures as field-keyed errors and `RespondIfInvalid` MUST render them through `response.ValidationError`.

(Previously: Referenced `response.ValidationErrors`; now `validator.ValidationErrors`)

#### Scenario: RespondIfInvalid writes field errors

- GIVEN `validator.ValidationErrors{"email": {"es requerido"}}`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it returns true and writes a 422 `ValidationErrorResponse`
- AND `errors.email` contains the validation messages

#### Scenario: Empty validation errors do not write

- GIVEN an empty `validator.ValidationErrors`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it returns false and writes no response
