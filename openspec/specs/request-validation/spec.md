# Request Validation Specification

## Purpose
Define the composable validator: `ValidationErrors`, `FieldValidator`, reusable `Rule` functions, and Gin integration.

## Requirements

### Requirement: ValidationErrors accumulator

The system MUST provide `ValidationErrors map[string][]string` with `Add(field, message)` and `HasErrors() bool` defined in `platform/validator`.

#### Scenario: Accumulate multiple errors

- GIVEN an empty `validator.ValidationErrors`
- WHEN `Add("email", "is required")` and `Add("email", "is invalid")` are called
- THEN `errs["email"]` contains both messages in order

### Requirement: FieldValidator chain

The system MUST provide `Field(errs, field, value) *FieldValidator` with chainable `Required()`, `Optional()`, and `Apply(rule Rule)`.

#### Scenario: Required empty field

- GIVEN `Field(errs, "name", "").Required()`
- WHEN the chain completes
- THEN `errs["name"]` contains `"es requerido"`

#### Scenario: Optional skips rules

- GIVEN `Field(errs, "bio", "").Optional().Apply(MinLength(10))`
- WHEN the chain completes
- THEN `errs.HasErrors()` is `false`

### Requirement: Reusable rules

The system MUST provide rules: `MinLength(n)`, `MaxLength(n)`, `ValidEmail()`, `HasUppercase()`, `HasDigit()`, `HasSpecialChar()`, `MinWords(n)`, `NoNumbers()`, `Matches(pattern)`.

#### Scenario: MinLength with runes

- GIVEN the string `"añ"` (2 runes)
- WHEN `MinLength(3)` is applied
- THEN it returns an error
- AND `"año"` (3 runes) passes

#### Scenario: ValidEmail

- GIVEN `"not-an-email"`
- WHEN `ValidEmail()` is applied
- THEN it returns `"debe ser un email válido"`

### Requirement: Gin helper

The system MUST provide `RespondIfInvalid(c *gin.Context, errs validator.ValidationErrors) bool` in `platform/response` that writes a 422 response if errors exist.

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

#### Scenario: RespondIfInvalid writes field errors

- GIVEN `validator.ValidationErrors{"email": {"es requerido"}}`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it returns true and writes a 422 `ValidationErrorResponse`
- AND `errors.email` contains the validation messages

#### Scenario: Empty validation errors do not write

- GIVEN an empty `validator.ValidationErrors`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it returns false and writes no response

### Requirement: ValidSlug rule

The system MUST provide a `ValidSlug()` rule that validates a value matches `^[a-z0-9]+(?:-[a-z0-9]+)*$` (lowercase alphanumeric with hyphens, no leading/trailing hyphens).

#### Scenario: Valid slug passes

- GIVEN `"my-awesome-slug"`
- WHEN `ValidSlug()` is applied
- THEN it returns empty string (no error)

#### Scenario: Invalid slug fails

- GIVEN `"My Slug!"`
- WHEN `ValidSlug()` is applied
- THEN it returns the Spanish message "debe ser un slug válido"

### Requirement: ValidURL rule

The system MUST provide a `ValidURL()` rule that validates a value parses as a valid URL with a scheme and host.

#### Scenario: Valid URL passes

- GIVEN `"https://example.com/path"`
- WHEN `ValidURL()` is applied
- THEN it returns empty string

#### Scenario: Invalid URL fails

- GIVEN `"not-a-url"`
- WHEN `ValidURL()` is applied
- THEN it returns the Spanish message "debe ser una URL válida"

### Requirement: InList rule

The system MUST provide an `InList(values ...string)` rule that validates a value is one of the allowed values.

#### Scenario: Value in list passes

- GIVEN `"active"` and `InList("active", "inactive", "suspended")`
- WHEN the rule is applied
- THEN it returns empty string

#### Scenario: Value not in list fails

- GIVEN `"pending"` and `InList("active", "inactive")`
- WHEN the rule is applied
- THEN it returns the Spanish message "debe ser uno de: active, inactive"

#### Scenario: Optional field with InList skips

- GIVEN `Field(errs, "status", "").Optional().Apply(InList("active", "inactive"))`
- WHEN the chain completes
- THEN no error is added

## Constraints and Edge Cases

- `MinLength` and `MaxLength` MUST count runes, not bytes.
- `Matches` MUST compile the regex once per rule creation, not per validation.
- The validator MUST NOT use struct tags or reflection.
- Error messages are in Spanish ("es requerido", "debe ser un email válido") per NexoKit convention.
