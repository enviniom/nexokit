# Exploration: Platform AppError Redesign and Centralized Error Logging

Change: `change-23-platform-apperror-error-middleware`
Source doc alias: `handled-api-error-logging`
Scope: This change defines infrastructure only. No module migration (Change 18).

## Current State

### `platform/apperror` (current shape)

`internal/platform/apperror/apperror.go` (112 lines) defines a single struct and a flat list of sentinels:

```go
type AppError struct {
    Err     error
    Message string
    Cause   error
}
```

- `Error()` returns `Err.Error()` if set, else `Message`.
- `Unwrap()` returns `Cause` only.
- `Is(target)` compares by pointer for `*AppError` and falls through to `errors.Is(ae.Err, target)`. Two sentinels compare equal only if they share the same pointer or both have no `Err` and the same `Message`.
- Sentinels: `ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrConflict`, `ErrBadRequest`, `ErrTooManyRequests`, `ErrValidation`, `ErrUnprocessable`, `ErrInternal`. Each is `&AppError{Message: messages.MsgXxx}`.
- `Wrap(err, message, cause...)` is the only constructor that returns a `*AppError` (used 3 times outside the package, including in `modules/iam/roles/delete_role/handler.go`).
- `Status(err)` maps by `errors.Is` against each sentinel; unknown errors default to 500.
- `PublicMessage(err, env)` returns `ae.Message` if set, else `MsgInternalError` when `env == "production"`, else `err.Error()`.

`apperror_test.go` (115 lines) covers `Wrap`, `Status` (table of 11 sentinels), `PublicMessage` (dev/prod/nil), and the sentinel message contract.

### Sentinel usage blast radius

- 58 call sites for apperror sentinels/helpers across 5 modules: `auth`, `companies`, `iam`, `onboarding`. Major call sites include `apperror.ErrNotFound` (12+), `apperror.ErrUnauthorized` (8+), `apperror.Wrap(apperror.ErrUnprocessable, ...)` (1, in roles delete handler).
- Modules already declare their own domain sentinels as plain `errors.New` in `internal/modules/iam/core/error.go` (15 entries) and translate them to `apperror.*` in handler-side `mapServiceError` (see `iam/roles/delete_role/handler.go:28-39`).

### `platform/response` (current shape)

`internal/platform/response/response.go` (239 lines):

- `APIResponse[T]`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse[T]` — the standard envelope.
- `HandleError(c, err)` is one call: `Error(c, apperror.Status(err), apperror.PublicMessage(err, gin.Mode()), nil)`. No logging, no context capture.
- `ValidationError(c, errs)` and `RespondIfInvalid(c, errs)` handle the 422/field-keyed path separately. DTOs return `validator.ValidationErrors` and handlers call `RespondIfInvalid` after binding.
- `respondIfInvalid` is the only other place that writes an error envelope from a non-handler-controlled error.

### Existing middlewares

- `internal/middleware/request_id.go` — `RequestID()` reads `X-Request-ID` or generates a UUID, sets it on the Gin context (`messages.CtxRequestID`) and on the response header.
- `internal/middleware/logger.go` — `Logger(log)` records a single `slog.Info` per request: `method`, `path`, `status`, `latency`, `client_ip`, `body_size`, `request_id`. Does **not** read `gin.Context.Errors`.
- `internal/middleware/recovery.go` — `Recovery(log)` recovers panics, logs `slog.Error` with `request_id` + `error` + `path`, and writes a 500 via `response.InternalServerError`.
- Router order (`internal/server/router.go:22-27`): `RequestID → gin.LoggerWithWriter → Logger → Recovery → CORS`.

No middleware iterates `gin.Context.Errors`. A grep for `c.Error(` across the repo returns 0 matches.

### Logger setup

`internal/infra/logger/logger.go` already supports:

- `New(cfg.LogConfig)` returns a `*slog.Logger` writing to `LOG_FILE` (lumberjack-rotated) or stdout.
- `NewErrorLogger(cfg.LogConfig)` returns a separate `*slog.Logger` writing to `LOG_ERROR_FILE` at `slog.LevelError`.
- `newHandlerWithOptions` returns `slog.NewTextHandler` when `LOG_FORMAT=text`, else `slog.NewJSONHandler` (JSON is the default).
- `parseLevel` maps `LOG_LEVEL` to `slog.Level{Debug,Info,Warn,Error}`.

Environment mode is read from `cfg.App.Env` (`internal/config/env.go`): `"local" | "development" | "production" | "test"`. Gin mode is set independently by `server.NewRouter` from `cfg.IsProduction()` / `cfg.IsTest()`.

### Latent bug confirmed during exploration

`response.HandleError` calls `apperror.PublicMessage(err, gin.Mode())` (line 204). `gin.Mode()` returns `"debug" | "release" | "test"`, but `PublicMessage` redacts only when `env == "production"`. The redaction branch is **dead code** along the normal request path. Today, an unknown error always returns `err.Error()` to the client, even in production. The existing `TestHandleError` test asserts this behavior (test mode leaks the raw string `"database is down"`). The new `AppError` design MUST fix this: either encode redaction inside `AppError` itself (the safer route — `PublicMessage` takes only `err` and looks at `gin.Mode()` or accepts an env flag explicitly from `response`), or pass `cfg.IsProduction()` from `HandleError`.

### Docs to update

- `docs/modules/validation-and-errors.md` already documents the new helpers in the example (`apperror.NotFound`, `apperror.Conflict`, `apperror.Forbidden`, `apperror.Internal`) even though they do not exist yet. The table is therefore aspirational and the design must catch up to it.
- `docs/modules.md` references `validation-and-errors.md` in the "Read this when..." and "Core rules at a glance" tables; no structural change needed unless the error owner table changes.
- `docs/error-handling.md` does not exist yet; the prompt requires creating it.
- `openspec/specs/error-handling/spec.md` defines the current sentinel contract. Adding a `Code` and `HTTPStatus` to `AppError` changes the matching semantics, so this spec needs an update (probably `MODIFIED Requirements`).
- `openspec/specs/http-middleware/spec.md` documents the middleware order. The new `ErrorLogger` middleware should be added in the same chain position as `Recovery` and the order updated.

## Affected Areas

| Area | Impact | Why |
|---|---|---|
| `internal/platform/apperror/apperror.go` | Modified | Reshape `AppError`, add `Code`/`HTTPStatus`/`PublicMessage`/`Internal`, add new helpers, preserve sentinel identity for `errors.Is` |
| `internal/platform/apperror/apperror_test.go` | Modified | Cover new helpers, `errors.Is` chain, `Status` mapping, `PublicMessage` redaction |
| `internal/platform/response/response.go` | Modified | `HandleError` uses new `AppError` fields, attaches error to `gin.Context.Errors` (or stashes `*AppError`) for the new logger middleware |
| `internal/platform/response/response_test.go` | Modified | Update `TestHandleError` for the new contract, add a test that the error reaches the log |
| `internal/middleware/error_logger.go` (new) | New | Iterate `c.Errors` after `c.Next()`, log `Internal` with `request_id`, `method`, `path`, `status`, `latency_ms`, `tenant_id`, `actor_id`, and the full unwrap chain |
| `internal/middleware/recovery.go` | Modified | Push the panic value into `c.Errors` so the new middleware logs it consistently; keep its own log line for backward compat or remove and let `ErrorLogger` own it |
| `internal/server/router.go` | Modified | Register `ErrorLogger` after `Recovery` so panics and handled errors share one log path |
| `docs/error-handling.md` | New | Shape of `AppError`, `Code → HTTPStatus` table, `PublicMessage` vs `Internal` rules, examples from core/service/repo/handler, logging policy |
| `docs/modules/validation-and-errors.md` | Modified | Update the error mapping table to reference `Code` instead of sentinel names where useful; ensure the example is consistent with the new helpers |
| `openspec/specs/error-handling/spec.md` | Modified | Add `Code`/`HTTPStatus` requirements, update sentinel matching requirements |
| `openspec/specs/http-middleware/spec.md` | Modified | Add `ErrorLogger` requirement, update middleware order |
| `internal/middleware/request_id.go` | Unchanged | Already provides the `request_id` context key the new logger relies on |
| `internal/infra/logger/logger.go` | Unchanged | JSON in production / text in dev is already supported by `LOG_FORMAT` |
| Module handlers/services | Unchanged | Migration to the new `code`-first helpers is Change 18; the current sentinels keep working through `errors.Is` |

## Approaches

### 1. Recommended: `code`-first helpers + `c.Error()` plumbing for the new middleware (Maintainer-confirmed direction)

Reshape `AppError` to:

```go
type Code string

type AppError struct {
    Code          Code
    HTTPStatus    int
    PublicMessage string
    Internal      error
}
```

Add the HTTP helpers as `NotFound(code, publicMsg, internal)`, `BadRequest(code, publicMsg, internal)`, `Forbidden(code, publicMsg, internal)`, `Conflict(code, publicMsg, internal)`, `Internal(code, publicMsg, internal)`. Platform exports only one canonical "code" per HTTP category by convention (e.g. `apperror.CodeNotFound = "NOT_FOUND"`, `apperror.CodeConflict = "CONFLICT"`) used as fallback; modules define their own `const CodeXxx apperror.Code = "module.xxx"`.

Sentinels (`ErrNotFound`, etc.) become `&AppError{Code: CodeNotFound, HTTPStatus: 404, PublicMessage: messages.MsgNotFound}` and keep matching through `errors.Is` by also defining `Is(target)` to fall through to `Code` equality:

```go
func (ae *AppError) Is(target error) bool {
    if t, ok := target.(*AppError); ok {
        if ae == t { return true }
        if t.Code != "" && ae.Code == t.Code { return true }
    }
    if ae.Internal != nil { return errors.Is(ae.Internal, target) }
    return false
}
```

`response.HandleError` becomes:

```go
func HandleError(c *gin.Context, err error) {
    if err == nil { return }
    var ae *AppError
    if errors.As(err, &ae) {
        // attach to c.Errors so the new middleware logs it
        _ = c.Error(err)
        body := response.ErrorResponse{...}
        if gin.Mode() != "release" {
            body.Debug = ae.Internal // optional dev field
        }
        c.JSON(ae.HTTPStatus, body)
        return
    }
    // unknown → 500 with generic public message
    _ = c.Error(err)
    c.JSON(http.StatusInternalServerError, response.ErrorResponse{
        Success: false,
        Message: messages.MsgInternalError,
    })
}
```

New `middleware.ErrorLogger(log)` runs after `c.Next()`, iterates `c.Errors`, and logs each with structured fields: `request_id`, `method`, `path`, `status`, `latency_ms`, `tenant_id`, `actor_id` (from `authctx.FromGin`), plus the full unwrap chain of `*AppError.Internal` (using `errors.As` to extract the inner error chain).

- Pros: Aligns with the maintainer decision. Keeps `errors.Is(err, apperror.ErrNotFound)` working. Fixes the latent info-disclosure bug. Centralizes logging for handled errors and panics under one path. `Internal` is the single source of truth for the error to log. `Code` makes it cheap to filter logs by business domain.
- Cons: Breaks any code that constructed `&AppError{Err: ..., Message: ..., Cause: ...}` directly (only the apperror package itself does). The `Is` method needs to be careful: a nil-target comparison could over-match — the design must guard against `*AppError.Code == ""` matching all unset codes.
- Effort: Medium

### 2. `c.Error()` only, keep `AppError` field shape unchanged

Add the new `ErrorLogger` middleware but keep `AppError` exactly as today. `HandleError` only changes to call `c.Error(err)`. Logging reads `c.Errors` and reflects on `*AppError` for `Message` and `Cause`.

- Pros: Smallest diff. No backward-compat concerns. Easier to revert.
- Cons: Doesn't deliver the prompt's stated goal of `code/status/public/internal` design. No fix for the latent `PublicMessage` bug. No `Code` to filter logs by domain.
- Effort: Low

### 3. Inject logger directly into `HandleError`; skip the middleware

`HandleError` takes a `*slog.Logger` and logs the error inline before writing the response.

- Pros: No new middleware. No `c.Errors` plumbing.
- Cons: Panics and middleware-originated errors are not logged here. Every call site (33 across modules) has to thread the logger. Handlers today don't have direct access to the logger — would require a container refactor.
- Effort: Medium (with broad call-site churn)

### 4. External slog middleware only (no platform apperror changes)

Add a middleware that does its own error inspection (e.g. on `c.Errors` or on a custom context key) and logs; leave `AppError` and `HandleError` alone.

- Pros: Decoupled. Reusable.
- Cons: Doesn't satisfy the prompt's "redesign `AppError`" requirement. Doesn't fix the `PublicMessage` bug. Two systems: handlers call `HandleError` for response, middleware logs separately with no shared schema.
- Effort: Low–Medium

### 5. Async / background log queue

A heavier observability path: buffer errors, push to a queue, ship to a backend.

- Pros: Decouples request latency from log IO.
- Cons: Out of scope for this change. Lumberjack already handles async rotation. The prompt and `handled-api-error-logging` source doc both call for direct slog, not a queue.
- Effort: High

## Recommendation

**Approach 1.** It satisfies the prompt's full redesign, the maintainer decision (code-first helpers), and the secondary `handled-api-error-logging` objective (centralized logging) in a single coherent change. It also fixes the latent `PublicMessage` info-disclosure bug by encoding the redaction in `AppError` itself. Migration of modules is explicitly deferred to Change 18, so the change stays within the 400-line review budget by limiting itself to platform + middleware + docs + specs.

Key design decisions to lock in during `sdd-propose` / `sdd-design`:

1. **Helper signature is `(code, publicMsg, internal)`.** The platform exports `Code` constants for the HTTP categories (`CodeNotFound`, `CodeConflict`, `CodeForbidden`, `CodeBadRequest`, `CodeInternal`, `CodeUnauthorized`, `CodeTooManyRequests`, `CodeValidation`, `CodeUnprocessable`). Modules declare their own `Code` constants in `core/error.go` or `core/constants.go` and pass them through.
2. **`errors.Is` continues to work for all existing sentinels.** The `AppError.Is` method falls through to `Code` equality so `errors.Is(err, apperror.ErrNotFound)` and `errors.Is(err, apperror.CodeNotFound)` both work.
3. **`PublicMessage` redaction moves into `AppError`.** The function takes only `err` (and optionally an `env` hint). The new `HandleError` calls it with the current gin mode and the new logic returns `MsgInternalError` for unknown errors regardless of mode (closes the latent bug).
4. **The new `ErrorLogger` middleware is the single logging path for both handled errors and panics.** `Recovery` continues to recover and writes the 500 response, but pushes the panic value into `c.Errors` so `ErrorLogger` produces a single, consistent log line per request. `Recovery` may keep its own log line for panic context, or the design may deprecate it — design decision for `sdd-design`.
5. **The error envelope is unchanged.** The prompt is explicit. No new fields in the public response shape. A `debug` field may be added in non-release mode for dev ergonomics; this is opt-in and not part of the contract.
6. **No API contract change.** `errors.Is(err, apperror.ErrNotFound)`, `apperror.Status(err)`, `apperror.PublicMessage(err, env)` keep the signatures they have today (positional adjustments for the new helpers are internal to platform).
7. **Module migration is out of scope.** Confirmed by the prompt and the maintainer decision. Change 18 will introduce module `core/error.go` patterns that wrap the new helpers.

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| `AppError` field rename breaks the package and its callers | Low | Only `apperror.go` itself constructs `&AppError{...}` literals (verified by grep). The single external literal is `apperror.Wrap`, which lives in the same package. |
| `errors.Is(err, apperror.ErrNotFound)` regression across 58 call sites | Medium | Preserve sentinel identity via `Is` falling through to `Code` equality. Add an explicit `TestErrorsIs` table in `apperror_test.go` that covers all current sentinels. |
| `PublicMessage` redaction change inadvertently hides legitimate client-visible messages | Low | New `PublicMessage` returns `ae.PublicMessage` whenever `*AppError` is matched; only unknown errors are redacted. Add a `TestPublicMessage_AppErrorUsesPublicMessage` test. |
| Logging duplicates when `Recovery` and `ErrorLogger` both log panics | Medium | Pick one owner in `sdd-design` (likely `ErrorLogger`); have `Recovery` call `c.Error(fmt.Errorf("panic: %v", r))` so the panic reaches the same log path. |
| Latency budget for per-request error logging | Low | slog is sync write to lumberjack. Lumberjack rotates asynchronously. No backpressure risk for the expected error rate. |
| Middleware order change affects existing test fixtures | Low | Update `recovery_test.go` to assert presence/absence of the duplicate log line. `logger_test.go` is unaffected — it only asserts the request log line. |
| `c.Error(err)` interaction with `c.Abort()` from `Response`/`InternalServerError` | Low | The standard pattern is `c.Error(err)` first, then write the response. The middleware reads `c.Errors` regardless of whether `c.IsAborted()` is true. |
| `tenant_id` / `actor_id` not yet on the gin context when the logger runs | Low | Both come from `authctx.FromGin(c)` and `tenant.FromGin(c)`, which already populate the context before request handling. The middleware reads defensively (empty string when missing). |
| Docs drift between `docs/modules/validation-and-errors.md` and the new code | Low | The example in that doc already references the new helpers. Once the helpers ship, the doc becomes correct as-is except for the mapping table that should now reference `Code`. |
| Spec drift in `openspec/specs/error-handling/spec.md` and `http-middleware/spec.md` | Low | Plan a `MODIFIED Requirements` delta for the error-handling spec (new `Code`/`HTTPStatus`/`PublicMessage` requirements) and an `ADDED Requirements` for the `ErrorLogger` middleware in the http-middleware spec. |
| The `*AppError` redefinition is technically a type change | Low | It is a struct change inside a single package. The single direct consumer outside is `apperror.Wrap`; everywhere else uses sentinels. |

## Unknowns / Open Questions for the Propose Phase

- Should `Recovery` keep its own `slog.Error` line, or should the new `ErrorLogger` middleware own ALL error logging (including panics)? The design should pick one.
- Should the optional `debug` field in non-release mode live on the standard `ErrorResponse` (additive) or on a parallel struct? The prompt suggests "in development, you can add a `debug` field" — it should be opt-in and behind gin mode.
- Should `ErrorLogger` use `internal/infra/logger.NewErrorLogger` (writes to `logs/error.log` at `slog.LevelError`) directly, or accept an injected `*slog.Logger` so tests can capture output? Recommended: accept an injected logger (matches `Recovery` and `Logger` patterns) but the app composition root wires in the error logger.
- Does the `c.Error(err)` path need to dedupe when the same error is added multiple times by chained calls? Gin's `Error` deduplicates by `*errorString` pointer identity; `*AppError` will dedupe correctly only if the same pointer is used twice. Acceptable for now.
- Should the new helpers accept `code` as a `Code` type or `string`? Maintainer decision says `Code`. Sticking with the typed alias.
- For backward compatibility, should `Wrap(err, message, cause...)` continue to exist? Yes — the prompt requires "compatibility with `errors.Is` and `errors.As`" and the existing handler `iam/roles/delete_role/handler.go` uses it. Keep `Wrap`, model it as `&AppError{Code: CodeInternal, HTTPStatus: 500, PublicMessage: message, Internal: err, ...}`.

## Ready for Proposal

**Yes.** The investigation is complete, the maintainer decision is captured, the latent `PublicMessage` bug is documented, and a clear single approach is recommended. The change fits the 400-line review budget (estimated ~300 changed lines including the new middleware, tests, and doc updates) and the delivery strategy is `single-pr-default` per the orchestrator's session preflight. The orchestrator should proceed to `sdd-propose` to lock in scope and the (code, publicMsg, internal) helper signatures, then `sdd-spec` for the new `Code → HTTPStatus` mapping scenarios, then `sdd-design`, then `sdd-tasks` with a per-platform-unit forecast.
