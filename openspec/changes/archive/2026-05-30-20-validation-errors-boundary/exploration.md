# Exploration: Validation Errors Ownership Boundary

## Current State

`ValidationErrors` is a `map[string][]string` type defined in `platform/response/response.go` (line 75) with `Add()` and `HasErrors()` methods. It is the accumulator used by the custom validator to collect field-level validation errors.

**Current dependency direction:** `platform/validator` → `platform/response`

The full validation flow is:

1. **DTO** (e.g. `modules/auth/core/dto.go`) creates `errs := make(response.ValidationErrors)` and calls `validator.Field(errs, ...)`
2. **Validator** (`platform/validator/validator.go`) uses `response.ValidationErrors` as the field type in `FieldValidator` and receives it in `Field()`
3. **Handler** calls `response.RespondIfInvalid(c, req.Validate())` which lives in `platform/response`
4. **Response** (`platform/response/response.go`) owns `ValidationErrors`, `ValidationErrorResponse`, `ValidationError()`, and `RespondIfInvalid()`

**Files involved (54 usages of `response.ValidationErrors`):**
- `internal/platform/response/response.go` — type definition + `ValidationErrorResponse` + `RespondIfInvalid`
- `internal/platform/validator/validator.go` — imports `response` for `ValidationErrors` type
- `internal/platform/validator/validator_test.go` — imports `response` for test setup
- 7 module DTO files (`auth`, `users`, `companies`, `roles`, `permissions`, `onboarding`, golden test) — import both `response` and `validator`
- 12 handler files — import `response` for `RespondIfInvalid`

## Affected Areas

- `internal/platform/response/response.go` — defines `ValidationErrors`, `ValidationErrorResponse`, `ValidationError()`, `RespondIfInvalid()`
- `internal/platform/validator/validator.go` — imports `response` solely for `ValidationErrors` type
- `internal/platform/validator/validator_test.go` — imports `response` for test setup
- `internal/modules/*/core/dto.go` — all DTOs import both packages
- `internal/modules/*/handler.go` — handlers use `RespondIfInvalid`
- `openspec/specs/request-validation/spec.md` — spec treats validator + response as one unit
- `openspec/specs/api-response/spec.md` — spec references `ValidationErrors` as part of response contract
- `openspec/specs/platform-boundary-rules/spec.md` — classifies both packages as generic

## Approaches

### 1. Move `ValidationErrors` to `platform/validator` (Recommended)

Move the type definition (`ValidationErrors`, `Add()`, `HasErrors()`) from `response` to `validator`. Keep `ValidationErrorResponse`, `ValidationError()`, and `RespondIfInvalid()` in `response` — they are HTTP concerns.

- **Pros:**
  - Conceptually correct: validator owns the error accumulator, response owns the HTTP envelope
  - Fixes dependency inversion: `validator` no longer depends on `response`
  - `response` imports `validator` for the type — lower-level package has no upper-level dependency
  - Module DTOs already import both packages; no new imports needed
  - No API contract change — JSON shape, HTTP status, and behavior are identical
- **Cons:**
  - `response` now imports `validator` (reversed direction) — but this is correct since response is the higher-level HTTP layer
  - 54 call sites need import adjustment (minor — most already import both)
  - Tests in `validator_test.go` need import swap
- **Effort:** Low (type move + import updates, no logic change)

### 2. Keep status quo (`ValidationErrors` in `platform/response`)

No changes. Document that `ValidationErrors` lives in `response` by design because it is part of the API response contract.

- **Pros:**
  - Zero code changes
  - Single import for the full validation flow (`response` has everything)
  - Existing specs remain valid
- **Cons:**
  - Conceptual misalignment: `ValidationErrors` is a validation data structure, not an HTTP concern
  - `validator` depends on `response` — dependency inversion (validator should be the lower-level package)
  - Violates the platform boundary principle from `_context.md`: "Si cambia el contrato global de la API, va en `platform`. Si cambia el lenguaje de un módulo, va en el módulo." — `ValidationErrors` is not an API contract, it's a validation primitive
- **Effort:** None

### 3. Extract `ValidationErrors` to a new `platform/valerr` sub-package

Create a minimal package with just the type. Both `validator` and `response` import it.

- **Pros:**
  - Zero coupling between `validator` and `response`
  - Clear single responsibility: `valerr` is just the data type
- **Cons:**
  - New package for a single type + 2 methods — over-engineering
  - All consumers add a third import
  - Violates `platform/*` guideline: "debe tener subpaquetes enfocados; evitar un paquete gigante" — adding micro-packages fragments the platform
- **Effort:** Medium (new package, import updates everywhere, spec updates)

### 4. Define `ValidationErrors` in both packages with adapter

Validator has its own type, response has its own type, conversion at the boundary.

- **Pros:**
  - Complete separation
- **Cons:**
  - Duplication of identical type definition
  - Conversion overhead at every handler
  - Violates DRY, adds maintenance burden
  - No practical benefit — the types are structurally identical
- **Effort:** Medium-High (duplication + adapters + tests)

## Recommendation

**Approach 1: Move `ValidationErrors` to `platform/validator`.**

This is the cleanest option with the least friction. The type is fundamentally a validation concern — it accumulates field errors during validation. The HTTP envelope (`ValidationErrorResponse`) and the response helper (`RespondIfInvalid`) correctly remain in `response` because they are HTTP-layer concerns.

The dependency flip (`response` → `validator`) is correct: `response` is the higher-level HTTP formatting layer that consumes validation results, and `validator` is the lower-level primitive that produces them.

**Migration scope:**
1. Move `ValidationErrors` type, `Add()`, `HasErrors()` from `response/response.go` to `validator/validator.go`
2. Update `response` to import `validator` for the type (used in `ValidationErrorResponse`, `ValidationError()`, `RespondIfInvalid()`)
3. Update `validator_test.go` import
4. Update module DTOs — most already import both, so only files that import `response` solely for `ValidationErrors` need `validator` added
5. Update specs to reflect the new boundary

## Risks

- **Import churn:** 54 references across ~20 files need review. Most already import both packages, so actual changes are fewer than the raw count suggests.
- **Spec drift:** `request-validation/spec.md` and `api-response/spec.md` both reference `ValidationErrors` and will need boundary clarification.
- **Golden test data:** `tests/cli/testdata/golden/goldenmod/dto.go` references `response.ValidationErrors` and must be updated to keep CLI generation tests passing.

## Ready for Proposal

**Yes.** The analysis is complete with a clear recommended approach. The change is a pure refactor with zero behavior change. The orchestrator should proceed to `sdd-propose` to define scope, then `sdd-spec`, `sdd-design`, and `sdd-tasks`.
