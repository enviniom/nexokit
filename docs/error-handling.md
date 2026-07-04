# Error handling

This document describes the platform-wide error contract: the `AppError` shape, HTTP helpers, wrapping semantics, the centralized `ErrorLogger` middleware, and how client responses are redacted in production.

## Quick path

1. Modules own their business `Code` constants; `platform/apperror` only owns generic HTTP-category codes.
2. Construct errors with helpers: `apperror.NotFound(code, publicMsg, internal)`, `Conflict`, `Internal`, etc.
3. Wrap existing errors with `apperror.Wrap(err, publicMsg, cause...)` to preserve sentinel status/code.
4. Handlers route business/app errors through `response.HandleError(c, err)`.
5. `response.HandleError` writes the envelope, attaches the error to `c.Errors` once, and lets `ErrorLogger` own the log line.
6. Unknown errors are always redacted to `messages.MsgInternalError`; the `debug` field is included only when the request context's `debug_errors` flag is true, which is derived from `AppConfig.Env`.

## `AppError` shape

```go
type Code string

type AppError struct {
    Code          Code
    HTTPStatus    int
    PublicMessage string
    Internal      error
}
```

| Field | Purpose |
|---|---|
| `Code` | Typed error identity. Used by `errors.Is` for code-equality matching and by `ErrorLogger` for filtering. |
| `HTTPStatus` | HTTP status code set by the helper constructor. `response.HandleError` writes this status. |
| `PublicMessage` | Client-safe text. The only message exposed to API consumers. |
| `Internal` | Underlying error chain for logs and `errors.Unwrap`. Never sent to clients. |

`Code` implements the `error` interface so callers can write:

```go
errors.Is(err, apperror.CodeNotFound)
```

## Platform HTTP-category codes

`platform/apperror` exports generic codes for HTTP categories:

```go
const (
    CodeNotFound        Code = "not_found"
    CodeBadRequest      Code = "bad_request"
    CodeForbidden       Code = "forbidden"
    CodeConflict        Code = "conflict"
    CodeUnauthorized    Code = "unauthorized"
    CodeTooManyRequests Code = "too_many_requests"
    CodeValidation      Code = "validation"
    CodeUnprocessable   Code = "unprocessable"
    CodeInternal        Code = "internal"
)
```

Modules MUST declare their own business codes and use them with the helpers:

```go
package core

import "github.com/enviniom/nexokit/internal/platform/apperror"

const (
    CodeUserNotFound    apperror.Code = "user_not_found"
    CodeEmailInUse      apperror.Code = "email_in_use"
    CodeCannotDeleteSelf apperror.Code = "cannot_delete_self"
)
```

## Helper constructors

Each helper sets the matching HTTP status:

```go
apperror.NotFound(code, "user not found", err)      // 404
apperror.BadRequest(code, "invalid input", err)     // 400
apperror.Forbidden(code, "access denied", err)      // 403
apperror.Conflict(code, "email already taken", err) // 409
apperror.Unauthorized(code, "missing token", err)   // 401
apperror.TooManyRequests(code, "rate limited", err) // 429
apperror.Validation(code, "validation failed", err) // 422
apperror.Unprocessable(code, "cannot process", err) // 422
apperror.Internal(code, "internal error", err)      // 500
apperror.New(code, status, publicMsg, err)          // uncommon status
```

The `internal` argument may be `nil` for sentinel-style errors.

## Sentinel compatibility

The existing sentinels remain usable:

```go
var (
    ErrNotFound        = &AppError{Code: CodeNotFound, HTTPStatus: 404, PublicMessage: messages.MsgNotFound}
    ErrForbidden       = &AppError{Code: CodeForbidden, HTTPStatus: 403, PublicMessage: messages.MsgForbidden}
    ErrUnauthorized    = &AppError{Code: CodeUnauthorized, HTTPStatus: 401, PublicMessage: messages.MsgUnauthorized}
    ErrConflict        = &AppError{Code: CodeConflict, HTTPStatus: 409, PublicMessage: messages.MsgConflict}
    ErrBadRequest      = &AppError{Code: CodeBadRequest, HTTPStatus: 400, PublicMessage: messages.MsgBadRequest}
    ErrTooManyRequests = &AppError{Code: CodeTooManyRequests, HTTPStatus: 429, PublicMessage: messages.MsgTooManyRequests}
    ErrValidation      = &AppError{Code: CodeValidation, HTTPStatus: 422, PublicMessage: messages.MsgValidationError}
    ErrUnprocessable   = &AppError{Code: CodeUnprocessable, HTTPStatus: 422, PublicMessage: ""}
    ErrInternal        = &AppError{Code: CodeInternal, HTTPStatus: 500, PublicMessage: messages.MsgInternalError}
)
```

`errors.Is` works by pointer identity, then non-empty `Code` equality, then the `Internal` chain. Both of these return `true`:

```go
errors.Is(err, apperror.ErrNotFound)
errors.Is(err, apperror.CodeNotFound)
```

## Wrapping errors

`apperror.Wrap` preserves the status and code of known `*AppError` values or matching sentinels; otherwise it defaults to `CodeInternal/500`.

```go
wrapped := apperror.Wrap(apperror.ErrNotFound, "user 123 not found", dbErr)

apperror.Status(wrapped)                 // 404
errors.Is(wrapped, apperror.ErrNotFound) // true
errors.Is(wrapped, dbErr)                // true (via Internal chain)
```

Multiple causes are appended to the unwrap chain after `err`:

```go
wrapped := apperror.Wrap(inner, "request failed", cause1, cause2)
errors.Is(wrapped, inner)  // true
errors.Is(wrapped, cause1) // true
errors.Is(wrapped, cause2) // true
```

## Response mapping

Handlers use `response.HandleError`:

```go
if err := service.CreateUser(ctx, req); err != nil {
    response.HandleError(c, err)
    return
}
```

`HandleError`:

1. Attaches `err` to `c.Errors` exactly once.
2. Writes the status from `apperror.Status(err)`.
3. Writes `ae.PublicMessage` when `err` is (or wraps) an `*AppError`.
4. Redacts unknown errors to `messages.MsgInternalError` regardless of Gin mode.
5. Adds a `debug` field to the envelope only when the request context carries the `debug_errors` flag set by `middleware.DebugErrors`.

For `*AppError`, `debug` contains `Internal.Error()`. For unknown errors, `debug` contains `err.Error()`.

## Centralized error logging

`ErrorLogger` is registered before `Recovery` so Gin's reverse-unwind runs `ErrorLogger` after `Recovery` has pushed panic errors into `c.Errors`. `DebugErrors` runs early to store the config-derived flag:

```text
RequestID → DebugErrors → Logger → ErrorLogger → Recovery → CORS → RateLimit (per-route)
```

`ErrorLogger` iterates `c.Errors` after `c.Next()` and emits one `slog.Error` record per error with these fields:

| Field | Source |
|---|---|
| `request_id` | `messages.CtxRequestID` |
| `method` | HTTP method |
| `path` | Request path |
| `status` | Response status |
| `latency_ms` | Request latency |
| `tenant_id` | `tenant.FromGin(c).CompanySlug` or `""` |
| `actor_id` | `authctx.PublicIDFromGin(c)` or `""` |
| `code` | `AppError.Code` or `""` |
| `public_message` | `AppError.PublicMessage` or `""` |
| `internal_chain` | `AppError.Internal.Error()` or `err.Error()` |

`Recovery` no longer logs. It recovers panics, converts them into an `*AppError` with code `internal`, pushes it via `c.Error`, writes a 500 envelope, and aborts.

## Validation flow stays separate

DTO field-level validation uses `response.ValidationErrors` and `response.ValidationError(c, errs)`, which returns HTTP 422 with field-keyed errors. This path does NOT use `AppError` or `HandleError`.

`apperror.Validation(...)` and `apperror.Unprocessable(...)` are for service-layer/application outcomes, not for DTO field validation.

## Redaction rules

The `debug` field is controlled by `Config.ExposeDebugErrors()`, not by `gin.Mode()`:

| Error kind | Debug disabled (production) | Debug enabled (local/development/test) |
|---|---|---|
| `*AppError` | `PublicMessage`, no `debug` | `PublicMessage` + `debug` from `Internal` |
| Unknown error | `messages.MsgInternalError`, no `debug` | `messages.MsgInternalError` + `debug` from `err.Error()` |

## Checklist

- [ ] Module business codes are declared in `core/errors.go`.
- [ ] Errors are built with `apperror` helpers, not ad-hoc `AppError` literals.
- [ ] `apperror.Wrap` is used when a sentinel status/code must be preserved.
- [ ] Handlers route app errors through `response.HandleError`.
- [ ] DTO validation uses `response.ValidationError`, not `AppError`.
- [ ] `ErrorLogger` is the only middleware emitting structured error logs.
