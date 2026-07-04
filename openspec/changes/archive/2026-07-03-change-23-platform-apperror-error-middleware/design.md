# Design: Platform AppError Redesign and Centralized Error Logging

## Technical Approach

Replace sentinel/message-only `AppError` with a code-first contract, keep the response envelope, and route handled errors plus recovered panics through Gin `c.Errors` so `middleware.ErrorLogger` is the only error-log owner. This is infrastructure-only; validation responses and module migrations remain unchanged.

## Architecture Decisions

| Decision | Alternatives considered | Rationale |
|---|---|---|
| `AppError{Code, HTTPStatus, PublicMessage, Internal}` | Infer status from code; keep `Err/Message/Cause` | Status must be explicit and helper-owned; `PublicMessage` is client-safe; `Internal` is the log/unwrap source. |
| Code equality for `errors.Is` only when both codes are non-empty | Pointer-only matching; empty-code equality | Preserves sentinel compatibility without broad empty-code overmatching and still falls through to `Internal`. |
| `ErrorLogger` registered before `Recovery` | Register after `Recovery`; log in `Recovery` | Gin unwinds in reverse. `RequestID → Logger → ErrorLogger → Recovery → CORS → RateLimit` (when enabled) lets `Recovery` attach panic errors before `ErrorLogger` logs after `c.Next()`. |
| Inject separate error logger into router | Reuse app logger; create logger inside middleware | Router wiring stays testable and production uses `infra/logger.NewErrorLogger`. |

## Data Flow

```text
Handled error:
handler/service error → response.HandleError → c.Error(err) once → JSON envelope
  → Gin unwind → ErrorLogger logs code/status/public_message/internal_chain

Panic:
handler panic → Recovery defer → c.Error(panic AppError) → 500 envelope → Abort
  → Gin unwind → ErrorLogger logs exactly once
```

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/platform/apperror/apperror.go` | Modify | Add `type Code string`, HTTP-category code constants, new struct, sentinels, helpers, `New`, `Wrap`, `Status`, `PublicMessage`, `Is`, `Unwrap`. |
| `internal/platform/response/response.go` | Modify | `HandleError` calls `c.Error(err)` once, builds envelope from `AppError`, redacts unknowns, adds `Debug string 'json:"debug,omitempty"'` to `ErrorResponse` only. |
| `internal/middleware/error_logger.go` | Create | Post-`c.Next()` middleware that logs `c.Errors` with structured fields and unwrap chains. |
| `internal/middleware/recovery.go` | Modify | Stop logging; convert panic to an error, attach through `c.Error`, write 500, abort. |
| `internal/server/router.go` | Modify | Accept `errorLog *slog.Logger`; order: `RequestID`, Gin access logger, `Logger`, `ErrorLogger`, `Recovery`, `CORS`, `RateLimit` when enabled/applicable. |
| `internal/app/bootstrap.go` | Modify | Create `errorLog` via `logger.NewErrorLogger(cfg.Log)` and pass it to `server.NewRouter`. |
| Tests/docs | Modify/Create | Update platform/middleware/router tests; add `docs/error-handling.md`; adjust validation/errors docs. |

## Interfaces / Contracts

```go
type Code string

type AppError struct {
    Code          Code
    HTTPStatus    int
    PublicMessage string
    Internal      error
}

func New(code Code, status int, publicMsg string, internal error) *AppError
func NotFound(code Code, publicMsg string, internal error) *AppError
func BadRequest(code Code, publicMsg string, internal error) *AppError
func Forbidden(code Code, publicMsg string, internal error) *AppError
func Conflict(code Code, publicMsg string, internal error) *AppError
func Unauthorized(code Code, publicMsg string, internal error) *AppError
func TooManyRequests(code Code, publicMsg string, internal error) *AppError
func Validation(code Code, publicMsg string, internal error) *AppError
func Unprocessable(code Code, publicMsg string, internal error) *AppError
func Internal(code Code, publicMsg string, internal error) *AppError
func Wrap(err error, message string, cause ...error) *AppError
func Status(err error) int
func PublicMessage(err error, mode string) string
func (ae *AppError) Unwrap() error
func (ae *AppError) Error() string

func ErrorLogger(log *slog.Logger) gin.HandlerFunc
func NewRouter(cfg *config.Config, log, errorLog *slog.Logger, ginWriter io.Writer, healthDeps HealthDeps, registerModules func(*gin.RouterGroup)) *gin.Engine
```

`Status(err)` returns first `*AppError.HTTPStatus`, known sentinel status via `errors.Is`, 200 for nil, else 500. `PublicMessage(err, mode)` returns `ae.PublicMessage`; unknown errors always return `messages.MsgInternalError`; debug belongs to `response`, not `apperror`. `(*AppError).Unwrap()` returns `Internal`. `(*AppError).Error()` returns `Internal.Error()` when set, else `PublicMessage`.

`Is` order is explicit: pointer match → non-empty `Code` equality → `errors.Is(Internal, target)`.

`Wrap`: nil/unknown errors become `CodeInternal/500`; if `errors.As(err, &ae)` succeeds, preserve `ae.Code`/`HTTPStatus`; otherwise check known sentinels. For known sentinel/AppError wrapping, use passed `message` as returned `PublicMessage`. Unwrap must include `err` then variadic causes, preserving `errors.Is`/`errors.As`.

`response.ErrorResponse` owns optional debug: `Debug string 'json:"debug,omitempty"'`. Validation responses never include debug. `apperror.Validation(...)` is for service-layer/application errors and does not replace DTO `response.ValidationError` responses.

`Recovery` converts panics with `apperror.Internal(apperror.CodeInternal, messages.MsgInternalError, fmt.Errorf("panic: %v", r))`, attaches it via `c.Error`, writes 500, aborts, and does not log.

`ErrorLogger` extraction: `request_id` from `messages.CtxRequestID`; `tenant_id` from `tenant.FromGin(c).CompanySlug` or empty; `actor_id` from `authctx.PublicIDFromGin(c)`; `code/public_message` from `errors.As(*AppError)`; `internal_chain` from `AppError.Internal` or original error, walking `errors.Unwrap`.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Helpers, sentinels, `Status`, `PublicMessage`, `Is`, `As`, `Wrap` preservation | Table tests in `apperror_test.go`. |
| Unit | `HandleError` c.Error count, redaction, debug gate, validation unaffected | `httptest` Gin context; inspect `c.Errors` and JSON. |
| Middleware | Handled errors and panics log once with required fields | Buffer-backed `slog.JSONHandler`; router-level tests. |
| Integration | Bootstrap/router wiring and full suite | `go test ./...`. |

## Migration / Rollout

No data migration required. This is a compatible infrastructure rollout: old sentinels and `Wrap` remain; module-owned business codes come later. Rollback is reverting the touched platform/middleware/router/bootstrap/docs files.

## Open Questions

None.
