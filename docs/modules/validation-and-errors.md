# Validation and errors

Validation and error mapping are the contract between the module and the HTTP layer. This document covers DTO validation, the response envelope, status code mapping, the module `core/errors.go` pattern, and the expected-control-flow rule.

## Quick path

1. DTOs own their own `Validate()` and return `response.ValidationErrors` keyed by field.
2. Binding / invalid JSON = 400; DTO validation = 422; `AppError` is for business / app outcomes.
3. Reusable module errors live in `core/errors.go` and use `platform/apperror` helpers.
4. Services and repositories MUST NOT construct ad-hoc `apperror` values inline.
5. Handlers funnel business / app errors through `response.HandleError`.
6. Expected control flow uses explicit contracts like `(*Customer, bool, error)`, not `AppError`.

## API response envelope reminder

Use the platform response envelopes everywhere; do not invent new shapes.

| Envelope | Purpose |
|---|---|
| `response.APIResponse` | Standard success envelope. |
| `response.ErrorResponse` | Standard error envelope. |
| `response.ValidationErrorResponse` | Field-keyed validation envelope. |
| `response.PaginatedResponse` | Paginated list response. |
| `response.PaginationMeta` | Pagination metadata. |

See [`api-conventions.md`](../api-conventions.md) for the full envelope and DTO naming rules.

## DTO validation contract

Request DTOs that need validation MUST expose a `Validate()` method that returns `response.ValidationErrors`:

```go
func (r CreateProductRequest) Validate() response.ValidationErrors {
    errs := make(response.ValidationErrors)
    validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
    validator.Field(errs, "slug", r.Slug).Required().Apply(validator.ValidSlug())
    validator.Field(errs, "status", r.Status).Optional().Apply(validator.InList("active", "inactive"))
    return errs
}
```

Handlers call `response.RespondIfInvalid(c, req.Validate())` immediately after binding. Validation failures are returned as HTTP 422 with `errors` keyed by field name:

```json
{
  "success": false,
  "message": "Error de validación",
  "errors": {
    "email": ["es requerido"]
  }
}
```

| Rule | Why |
|---|---|
| Each request DTO owns its own `Validate()`. | Validation is part of the DTO contract, not the handler. |
| `Validate()` returns `response.ValidationErrors` keyed by field name. | The envelope is uniform across modules. |
| Handlers MUST call `RespondIfInvalid` right after binding. | The handler stays thin; the DTO owns the contract. |
| Do not surface validation as `AppError`. | `AppError` is for business / app outcomes, not field-level rules. |

## Status code mapping

| Outcome | Status | Mechanism |
|---|---|---|
| Request binding / invalid JSON. | 400 Bad Request. | Gin binding error. |
| DTO `Validate()` failure. | 422 Unprocessable Entity. | `response.RespondIfInvalid` with `response.ValidationErrors`. |
| Business / app outcome (not found, forbidden, conflict, insufficient stock, etc.). | `AppError`-based status (404, 403, 409, ...). | `response.HandleError(c, err)` reads the `AppError` code. |
| Unexpected internal error. | 500. | `response.HandleError` redacts the message to `messages.MsgInternalError`; the original error is logged by `ErrorLogger`. A `debug` field is added only when `Config.ExposeDebugErrors()` is true (local/development/test). |

Keep this contract visible in code review: 400 is for malformed input, 422 is for valid-shape-but-invalid-content, and `AppError` is for business outcomes.

## Error ownership

| Concern | Owner |
|---|---|
| HTTP status, public message, internal error chain, `Unwrap()`, helper constructors, platform HTTP-category `Code` constants. | `platform/apperror` (`AppError` + helpers). |
| Reusable, module-scoped application errors used across slices. | `internal/modules/<module>/core/errors.go`, built with `apperror` helpers and module-owned `Code` constants. |
| Ad-hoc, slice-scoped business error values. | Declared in `slices/<slice>/` only when not reused. |
| Centralized structured logging of handled errors and panics. | `middleware.ErrorLogger`. |
| Panic recovery and pushing panic errors into `c.Errors`. | `middleware.Recovery`. |

See [`docs/error-handling.md`](../error-handling.md) for the full platform contract, `AppError` shape, `Wrap` semantics, `ErrorLogger` flow, and release-mode redaction rules.

## `core/errors.go` pattern

A module's reusable errors live in a single file, declared with `apperror` helpers and module-owned `Code` constants:

```go
package core

import "github.com/enviniom/nexokit/internal/platform/apperror"

const (
    CodeUserNotFound    apperror.Code = "user_not_found"
    CodeEmailAlreadyTaken apperror.Code = "email_already_taken"
    CodeCannotDeleteSelf  apperror.Code = "cannot_delete_self"
)

var (
    ErrUserNotFound      = apperror.NotFound(CodeUserNotFound, "user not found", nil)
    ErrEmailAlreadyTaken = apperror.Conflict(CodeEmailAlreadyTaken, "email already taken", nil)
    ErrCannotDeleteSelf  = apperror.Forbidden(CodeCannotDeleteSelf, "cannot delete your own account", nil)
)

func UserNotFound(err error) error {
    return apperror.NotFound(CodeUserNotFound, "user not found", err)
}
```

| Rule | Why |
|---|---|
| Reusable errors are declared in `core/errors.go` only. | The module's error vocabulary is centralized and reviewable. |
| Modules own their business `apperror.Code` constants; platform only owns HTTP-category codes. | Business semantics belong to modules; platform stays infrastructure. |
| Construct error values with `apperror` helpers. | Helpers set the right HTTP status, public message, and `Unwrap()` chain. |
| Do not construct ad-hoc `apperror` values inline inside `service.go` or `repository.go`. | Inline declarations hide the module's error contract and drift. |
| Ad-hoc, slice-scoped errors stay in a slice-local `errors.go` only when not reused. | They are not part of the module's shared vocabulary. |

## Layer rules for errors

| Layer | Rule |
|---|---|
| Service | Returns reusable errors from `core/errors.go` or wraps internal errors with `fmt.Errorf("...: %w", err)`. MUST NOT construct ad-hoc `apperror` values. |
| Repository | Maps DB / GORM errors to domain errors before returning. MUST NOT construct ad-hoc `apperror` values. |
| Handler | Calls `response.HandleError(c, err)` for business / app errors. MUST NOT inspect `apperror` codes manually. |

## Error mapping path

Errors move inward as domain errors and outward as API responses:

```txt
DB / GORM error → repository → core / domain error → service → handler → HTTP / API response
```

| Case | Layer returns | Handler response |
|---|---|---|
| Missing row | Module error from `core/errors.go`, built with `apperror.NotFound(...)`. | 404 via `response.HandleError`. |
| Field-level duplicate caught during DTO / form validation. | `response.ValidationErrors{"email": []string{"already exists"}}` from `dto.Validate()` or handler validation. | 422 with `errors.email`. |
| Business conflict. | Module error from `core/errors.go`, built with `apperror.Conflict(...)`. | 409 via `response.HandleError`. |
| Protected resource. | Module error from `core/errors.go`, built with `apperror.Forbidden(...)`. | 403 via `response.HandleError`. |
| Invalid DTO. | `response.ValidationErrors{"email": []string{"invalid"}}` from `dto.Validate()`. | 422 with field-keyed error map. |
| Malformed JSON body. | Binding error from Gin. | 400 Bad Request. |
| Unexpected DB failure. | Wrapped technical error. | 500 via `response.HandleError`; internal error logged by middleware. |
| Expected "not present" (idempotent). | `(*Customer, bool, error)` with the bool as the existence signal. | Branch on bool, no error mapping. |

## Expected control flow

If missing data is an exceptional lookup failure, repositories SHOULD translate persistence not-found errors to a domain error:

```txt
gorm.ErrRecordNotFound → core.ErrNotFound
```

If missing data is expected control flow, prefer an explicit existence contract instead of making services compare errors:

```go
GetCustomerByEmail(ctx context.Context, email string) (*core.Customer, bool, error)
```

| Rule | Why |
|---|---|
| Use `(*T, bool, error)` (or a typed result) when the service naturally needs to branch on "exists vs missing". | A sentinel error is not a control flow signal; the contract makes the branch explicit. |
| Do not represent expected control flow as `AppError`. | `AppError` is for exceptional business outcomes, not for branching on "is it there". |
| Use the `verify_otp` `GetCustomer` / customer-creation path as the canonical example. | It must distinguish "first-time customer" from "returning customer" before deciding whether to create a new record. |

## Validation and errors checklist

- [ ] Each request DTO has its own `Validate()` returning `response.ValidationErrors`.
- [ ] Handlers call `response.RespondIfInvalid` right after binding.
- [ ] Binding / invalid JSON returns 400; DTO validation returns 422.
- [ ] No `AppError` is constructed inline in services or repositories.
- [ ] Reusable module errors live in `core/errors.go` and use `apperror` helpers.
- [ ] Handlers route business / app errors through `response.HandleError`.
- [ ] Expected control flow uses `(*T, bool, error)` or a typed result, not `AppError`.
- [ ] Repository maps `gorm.ErrRecordNotFound` to a module error when missing data is exceptional.
