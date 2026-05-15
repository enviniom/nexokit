# Request Validation Specification

## Purpose
Define the composable validator: `ValidationErrors`, `FieldValidator`, reusable `Rule` functions, and Gin integration.

## Requirements

### Requirement: ValidationErrors accumulator

The system MUST provide `ValidationErrors map[string][]string` with `Add(field, message)` and `HasErrors() bool`.

#### Scenario: Accumulate multiple errors

- GIVEN an empty `ValidationErrors`
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

The system MUST provide `RespondIfInvalid(c *gin.Context, errs ValidationErrors) bool` that writes a 422 response if errors exist.

#### Scenario: Validation fails

- GIVEN `errs.HasErrors()` is `true`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it writes status 422 with the standard error envelope
- AND returns `true`

#### Scenario: Validation passes

- GIVEN `errs.HasErrors()` is `false`
- WHEN `RespondIfInvalid(c, errs)` is called
- THEN it returns `false` and writes nothing

## Constraints and Edge Cases

- `MinLength` and `MaxLength` MUST count runes, not bytes.
- `Matches` MUST compile the regex once per rule creation, not per validation.
- The validator MUST NOT use struct tags or reflection.
- Error messages are in Spanish ("es requerido", "debe ser un email válido") per NexoKit convention.
