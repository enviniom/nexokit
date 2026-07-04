# Tasks: Platform AppError Redesign and Centralized Error Logging

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines | ~700-1100 across platform/middleware/response/router/bootstrap/docs |
| 800-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 platform → PR 2 middleware → PR 3 wiring+docs |
| Delivery strategy | single-pr-default |
| Chain strategy | none — maintainer-approved `size:exception` |

Decision needed before apply: No — maintainer approved `size:exception`; applied as single PR.
Chained PRs recommended: Yes
Chain strategy: none
400-line budget risk: High (resolved by size:exception)

### Suggested Work Units

| Unit | Goal | Likely PR | Base |
|---|---|---|---|
| 1 | AppError contract, helpers, `Wrap` preservation, response redaction, new `docs/error-handling.md` | PR 1 | `main` |
| 2 | `ErrorLogger` middleware + `Recovery` refactor + router reorder | PR 2 | PR 1 branch |
| 3 | Bootstrap wiring + update `docs/modules/validation-and-errors.md` + `docs/modules.md` | PR 3 | PR 2 branch |

## Phase 1: Platform AppError Contract (PR 1, TDD)

- [x] 1.1 RED — extend `internal/platform/apperror/apperror_test.go`: `Code` constants, helper status, `Is` Code-equality + `Internal` fall-through, `errors.As` recovery, `Wrap` status preservation (incl. `apperror.Wrap(apperror.ErrUnprocessable, ...)` → 422), `Unwrap` chain.
- [x] 1.2 GREEN — rewrite `internal/platform/apperror/apperror.go`: `type Code string`; HTTP-category constants; struct `{Code, HTTPStatus, PublicMessage, Internal}`; sentinels; helpers `NotFound`/`BadRequest`/`Forbidden`/`Conflict`/`Unauthorized`/`TooManyRequests`/`Validation`/`Unprocessable`/`Internal`; `New`; status-preserving `Wrap`; `Is`/`Unwrap`/`Error`; `Status`; `PublicMessage`.
- [x] 1.3 REFACTOR — `Is` order (pointer → non-empty `Code` → `errors.Is(Internal, target)`); `go test ./...` keeps the 58 module call sites green.
- [x] 1.4 RED — `internal/platform/response/response_test.go`: redaction, release unknown → `MsgInternalError`, debug omitted in release / present otherwise, `c.Error` once, validation path unchanged.
- [x] 1.5 GREEN — `internal/platform/response/response.go`: `HandleError` calls `c.Error(err)` once, envelope from `*AppError`, add `Debug string \`json:"debug,omitempty"\`` to `ErrorResponse`, gate debug via the request context's `debug_errors` flag set by `middleware.DebugErrors` from `Config.ExposeDebugErrors()`.

## Phase 2: ErrorLogger + Recovery (PR 2, TDD)

- [x] 2.1 RED — `internal/middleware/error_logger_test.go`: one `slog.Error` per `c.Errors` entry, required fields, `request_id`/`tenant_id`/`actor_id`, missing context → `""`, panic in `internal_chain`.
- [x] 2.2 GREEN — `internal/middleware/error_logger.go`: `ErrorLogger(log *slog.Logger) gin.HandlerFunc`; post-`c.Next()`; iterates `c.Errors`; record with `request_id`/`method`/`path`/`status`/`latency_ms`/`tenant_id`/`actor_id`/`code`/`public_message`/`internal_chain` (walk `errors.Unwrap`).
- [x] 2.3 RED — update `internal/middleware/recovery_test.go`: 500 status, `c.Errors` populated, no `slog.Error` line emitted.
- [x] 2.4 GREEN — `internal/middleware/recovery.go`: drop `slog.Error`; panic → `apperror.Internal(apperror.CodeInternal, messages.MsgInternalError, fmt.Errorf("panic: %v", r))`; `c.Error`; `response.InternalServerError`; abort.
- [x] 2.5 `internal/server/router.go`: accept `errorLog *slog.Logger`; order `RequestID → gin.LoggerWithWriter → Logger → ErrorLogger → Recovery → CORS` (+ `RateLimit` when enabled); update `router_test.go` signature.

## Phase 3: Bootstrap Wiring + Docs (PR 3)

- [x] 3.1 `internal/app/bootstrap.go`: build `errorLog` via `logger.NewErrorLogger(cfg.Log)`; pass to `server.NewRouter`; bubble error.
- [x] 3.2 Create `docs/error-handling.md`: AppError shape, helpers, `Wrap`, `ErrorLogger` flow, release redaction, debug gate, `c.Errors` ownership.
- [x] 3.3 Update `docs/modules/validation-and-errors.md`: new helpers, redaction, split of `response.ValidationError` (422) from `AppError`.
- [x] 3.4 Update `docs/modules.md`: link `docs/error-handling.md` from the error section.

## Phase 4: Verification

- [x] 4.1 `go test ./...` — 58 sentinel call sites still match.
- [x] 4.2 `go build ./...` clean.
- [x] 4.3 Router test: panicking handler → exactly one `slog.Error`; release-mode response has no `debug` and no `Internal` text.

## Phase 5: Debug gate source-of-truth adjustment

- [x] 5.1 Add `Config.ExposeDebugErrors() bool` in `internal/config/env.go` returning true for local/development/test and false for production.
- [x] 5.2 Add `middleware.DebugErrors(enabled bool) gin.HandlerFunc` in `internal/middleware/debug_errors.go` that stores the flag under `messages.CtxDebugErrors`.
- [x] 5.3 Update `internal/server/router.go` to register `DebugErrors(cfg.ExposeDebugErrors())` immediately after `RequestID`.
- [x] 5.4 Update `internal/platform/response/response.go` `HandleError` to read the `debug_errors` context flag instead of `gin.Mode()`.
- [x] 5.5 Update tests: `internal/config/env_test.go` for `ExposeDebugErrors`; `internal/platform/response/response_test.go` for config-driven debug and Gin-mode independence; `internal/middleware/debug_errors_test.go` for the new middleware; `internal/server/server_test.go` router panic test for config-vs-gin misalignment.
- [x] 5.6 Update `docs/error-handling.md`, `docs/modules/validation-and-errors.md`, `docs/nexokit-architecture.md`, and `openspec/changes/change-23-platform-apperror-error-middleware/specs/error-handling/spec.md` to document that `AppConfig.Env` controls debug exposure, not `gin.Mode()`.
