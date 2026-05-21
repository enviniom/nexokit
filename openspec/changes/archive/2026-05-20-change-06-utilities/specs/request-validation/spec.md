# Delta for request-validation

## ADDED Requirements

### Requirement: Field-keyed validation responses

The system MUST return validation failures as field-keyed errors and `RespondIfInvalid` MUST render them through `response.ValidationError`.

#### Scenario: RespondIfInvalid writes field errors

- GIVEN `response.ValidationErrors{"email": {"es requerido"}}`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it returns true and writes a 422 `ValidationErrorResponse`
- AND `errors.email` contains the validation messages

#### Scenario: Empty validation errors do not write

- GIVEN an empty `response.ValidationErrors`
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
