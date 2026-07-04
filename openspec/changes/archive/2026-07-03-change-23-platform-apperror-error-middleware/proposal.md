# Proposal: Platform AppError Redesign and Centralized Error Logging

## Intent

Reshape `platform/apperror` to a code-first `AppError` with HTTP helpers, fix the latent `PublicMessage` redaction bug (it compares `gin.Mode()` against `"production"`, leaking unknown errors in production), and centralize logging of handled errors and panics in a new `ErrorLogger` middleware.

## Scope

### In Scope

- `AppError{Code, HTTPStatus, PublicMessage, Internal}`; HTTP helpers `(code Code, publicMsg string, internal error) *AppError`: `NotFound`, `BadRequest`, `Forbidden`, `Conflict`, `Internal`; lower-level `New(code, status, publicMsg, internal)`. Platform exports HTTP-category `Code` constants; modules own business codes. `AppError.Is` falls through to `Code` equality so existing sentinels keep matching.
- `PublicMessage` redaction encoded in `AppError`; `HandleError` redacts unknown errors in any mode. `response.HandleError` uses `HTTPStatus` + `PublicMessage` and attaches the error to `gin.Context.Errors`.
- New `middleware.ErrorLogger` (injected from `infra/logger.NewErrorLogger`) iterates `c.Errors` after `c.Next()` and emits one structured log per error: `request_id`, `method`, `path`, `status`, `latency_ms`, `tenant_id`, `actor_id`, `code`, `public_message`, `internal_chain`. `Recovery` stops logging and pushes panic into `c.Errors`; router order: `RequestID → Logger → ErrorLogger → Recovery → CORS → RateLimit` where rate limit is enabled/applicable. Gin unwinds middleware in reverse registration order, so `ErrorLogger` must be registered before `Recovery` to run after `Recovery` attaches panics.
- Optional `debug` field only when `gin.Mode() != "release"`. `docs/error-handling.md` (new); `docs/modules/validation-and-errors.md`, `openspec/specs/error-handling/spec.md`, `openspec/specs/http-middleware/spec.md` updated.

### Out of Scope

- Module migration to module-owned `Code` (Change 18); envelope shape or `messages.Msg*` constants; async log queue; validation path and module handlers, services, repositories.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `error-handling`: `AppError` fields change; `Is` extended; new helpers; `Wrap` retained.
- `http-middleware`: new `ErrorLogger` requirement; order updated; `Recovery` no longer logs.

## Approach

Code-first `AppError`; status is set by HTTP helpers, not inferred. `Code` filters logs, `PublicMessage` is the only client-visible text, `Internal` is the single source of truth for logging. `HandleError` writes the response and pushes the error into `c.Errors`; `ErrorLogger` runs after `c.Next()` and emits one log per error. `Recovery` becomes a thin panic catcher. Composition root wires `infra/logger.NewErrorLogger` into the router.

## Affected Areas

`internal/platform/apperror`, `internal/platform/response` (+ tests); new `internal/middleware/error_logger.go` + test; `internal/middleware/recovery.go` + test, `internal/server/router.go`, `internal/app/container.go`; new `docs/error-handling.md`; updated `docs/modules/validation-and-errors.md`, `openspec/specs/error-handling/spec.md`, `openspec/specs/http-middleware/spec.md`.

## Risks

- `errors.Is` regression across 58 call sites (Medium) → `Is` falls through to `Code`; new table test covers every sentinel.
- Duplicate logs (Recovery + ErrorLogger) for the same panic (Medium) → `Recovery` drops its own log; pushes panic into `c.Errors`.
- `tenant_id`/`actor_id` empty when logger runs (Low) → defensive reads; empty string on miss.
- `debug` field leaks to production (Low) → guarded by `gin.Mode() != "release"`; explicit release-mode test.

## Rollback Plan

Revert commits touching `platform/apperror`, `platform/response`, `middleware/error_logger.go`, `middleware/recovery.go`, `server/router.go`, `app/container.go`. Old `AppError{Err, Message, Cause}` and old `Recovery` log are restored. Module call sites untouched.

## Dependencies

- `internal/infra/logger.NewErrorLogger` (`logs/error.log` at `slog.LevelError`).
- `internal/middleware/request_id.go` (`request_id` context key).
- `internal/platform/authctx`, `internal/platform/tenant` (`FromGin` helpers).

## Success Criteria

- [ ] `errors.Is(err, apperror.ErrNotFound)` and `errors.Is(err, apperror.CodeNotFound)` both return `true` for wrapped errors.
- [ ] `apperror.Internal("boom", errors.New("kaboom"))` → 500 with unified envelope; log contains `"kaboom"` and `request_id`.
- [ ] `apperror.NotFound(code, "foo", errors.New("db boom"))` → `PublicMessage == "foo"`, `HTTPStatus == 404`; production never exposes `"db boom"`.
- [ ] Panicking handler produces exactly one `ErrorLogger` log line; release mode has no `debug` field and no `Internal` text.
- [ ] `go test ./...` passes; new tests cover helpers, `errors.Is`/`As`, `HandleError`, `ErrorLogger`, updated `Recovery`.
