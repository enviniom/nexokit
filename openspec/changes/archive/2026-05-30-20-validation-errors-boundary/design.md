# Design: Validation Errors Ownership Boundary

## Technical Approach

Move `ValidationErrors` from `internal/platform/response` to `internal/platform/validator` and update consumers to use `validator.ValidationErrors`. `response` keeps the HTTP boundary: `ValidationErrorResponse`, `ValidationError`, and `RespondIfInvalid`. This maps directly to the request-validation, api-response, and platform-boundary-rules deltas while preserving status codes, messages, and JSON shape.

## Dependency Direction

Current:

    modules/dto ──→ validator ──→ response ──→ gin/http
         │                              ▲
         └──────────────────────────────┘

Target:

    modules/dto ──→ validator
         │             ▲
         │             │
         └────→ response ──→ gin/http

DTOs and conflict handlers create validation errors; response only renders them.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|----------|--------|--------------------------|-----------|
| Type owner | `validator` owns `ValidationErrors`, `Add`, `HasErrors` | Keep in `response`; create `platform/valerr`; duplicate types with adapter | The type is a validation accumulator. A one-type package is fragmentation; duplicate types add conversions without value. |
| HTTP boundary | `response` imports `validator` | Move response helpers into validator; use raw `map[string][]string` | Gin/status/envelope code belongs in `response`; methods keep current caller ergonomics. |
| Compatibility | No permanent `response.ValidationErrors` alias | Permanent alias | Final API must make the boundary visible. Alias is only a rollback bridge if needed. |

## Data Flow

1. DTO creates `errs := make(validator.ValidationErrors)`.
2. DTO calls `validator.Field(errs, ...).Required().Apply(...)`.
3. Handler calls `response.RespondIfInvalid(c, req.Validate())`.
4. `response` writes the same 422 `ValidationErrorResponse` with field-keyed arrays.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/platform/validator/validator.go` | Modify | Add `ValidationErrors`; remove `response` import; update `FieldValidator.errs` and `Field`. |
| `internal/platform/response/response.go` | Modify | Import `validator`; type `ValidationErrorResponse.Errors`, `ValidationError` conversion, and `RespondIfInvalid` with `validator.ValidationErrors`. |
| `internal/platform/validator/validator_test.go` | Modify | Construct local `ValidationErrors`; keep behavior assertions. |
| `internal/platform/response/response_test.go` | Modify | Use `validator.ValidationErrors` in helper tests. |
| `internal/modules/**/core/dto.go`, `internal/modules/users/dto.go`, `internal/modules/roles/dto.go` | Modify | Return/create `validator.ValidationErrors`; validation logic unchanged. |
| `internal/modules/**/handler.go` | Modify | Manual validation-error construction uses `validator.ValidationErrors`; rendering stays `response.ValidationError`. |
| `tests/cli/testdata/golden/goldenmod/dto.go` | Modify | Update golden expected imports/types. |

## Interfaces / Contracts

```go
// package validator
type ValidationErrors map[string][]string
func (ve ValidationErrors) Add(field, message string)
func (ve ValidationErrors) HasErrors() bool
func Field(errs ValidationErrors, field, value string) *FieldValidator

// package response
func RespondIfInvalid(c *gin.Context, errs validator.ValidationErrors) bool
func ValidationError(c *gin.Context, errs any)
```

Compatibility constraints: keep `422`, `messages.MsgValidationError`, field names, message arrays/order, envelope keys, and `ValidationError` support for `map[string][]string` unchanged.

## Migration / Rollout

- Single atomic refactor: move type, update imports/usages, run gofmt.
- No DB migration, feature flag, or API versioning.
- Archive phase syncs spec language into `openspec/specs/{request-validation,api-response,platform-boundary-rules}/spec.md` after acceptance.

Rollback: revert the commit. If consumers need a temporary bridge, add `type ValidationErrors = validator.ValidationErrors` in `response`, then remove it after migration.

## Testing / Verification Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `Add`, `HasErrors`, `Field`, skip behavior | Existing `internal/platform/validator` tests after import swap. |
| Unit | `RespondIfInvalid`, `ValidationError` envelope | Existing `response` tests; confirm 422 and unchanged body. |
| Module | DTO return types and handlers | Existing module tests compile and pass. |
| Contract | Golden code and package graph | `go test ./...`, `go build ./...`; verify `validator` has no `response` import. |

## Open Questions

- None.
