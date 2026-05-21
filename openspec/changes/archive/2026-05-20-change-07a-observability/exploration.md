# Exploration: change-07a-observability

## Current State

### Logger (`internal/infra/logger/logger.go`)
- **Already implemented** with `slog` + `lumberjack`.
- Supports JSON and text formats.
- Three outputs: app log (`logs/app.log`), error log (`logs/error.log`), Gin writer (`logs/gin.log`).
- Log rotation via `lumberjack` with configurable `MaxSize`, `MaxBackups`, `MaxAge`, `Compress`.
- Level parsing: debug, info (default), warn, error.
- **Gap**: No `trace` level support (not critical for slog).

### Request Logger Middleware (`internal/middleware/logger.go`)
- **Already implemented**.
- Logs: method, path, status, latency, client_ip, body_size, request_id.
- Uses structured `slog` fields.
- Integrates with `messages` package for constants.
- Has test coverage (`logger_test.go`).

### Recovery Middleware (`internal/middleware/recovery.go`)
- **Already implemented**.
- Catches panics, logs with request_id + path, returns structured 500.
- Has test coverage (`recovery_test.go`).
- **Gap**: No stack trace capture (could be added for debugging).

### Health Check (`internal/server/router.go`)
- **Partially implemented**: only `GET /health` exists.
- Returns `{"success": true, "message": "API is healthy", "data": {"status": "ok"}}`.
- **Missing**: `/health/live` and `/health/ready` endpoints.
- **Missing**: DB connectivity check in `/health/ready`.
- **Missing**: Cache connectivity check in `/health/ready` (when cache is active).

### Graceful Shutdown (`cmd/api/main.go`, `internal/app/app.go`, `internal/server/server.go`)
- **Already implemented**.
- `cmd/api/main.go`: signal handling (SIGINT, SIGTERM), timeout context, calls `app.Stop()`.
- `app.Stop()`: stops HTTP server → closes DB → closes cache, with error aggregation.
- `server.Stop()`: calls `httpServer.Shutdown(ctx)`.
- Config: `SHUTDOWN_TIMEOUT_SECONDS` (default 30).
- **Gap**: No explicit SIGHUP support for log rotation signal.

### Config (`internal/config/config.go`)
- **Already has** all logging env vars: `LOG_LEVEL`, `LOG_FORMAT`, `LOG_FILE`, `LOG_MAX_SIZE`, `LOG_MAX_BACKUPS`, `LOG_MAX_AGE`, `LOG_COMPRESS`, `LOG_GIN_FILE`, `LOG_ERROR_FILE`.
- **Already has** `SHUTDOWN_TIMEOUT_SECONDS`.
- **Already has** `CacheConfig.Driver` (`none` default).
- **Matches** the prompt's required variables exactly (prompt uses `LOG_MAX_SIZE_MB` but code uses `LOG_MAX_SIZE` — minor naming difference, same semantics).

### Cache (`internal/infra/cache/`)
- `Cache` interface defined: Get, Set, Delete, Close.
- `NoopCache` fully implemented.
- `RedisCache` is a **TODO stub** — not yet wired.
- Bootstrap always uses `NewNoop()` — Redis driver selection not yet implemented.

### Existing Tests
- `server_test.go`: tests `/health` endpoint, 404, server start/stop.
- `logger_test.go`: tests request logging middleware.
- `recovery_test.go`: tests panic recovery.
- All use `httptest` + standard `testing`.

## Affected Areas

| File | Why Affected |
|------|-------------|
| `internal/server/router.go` | Add `/health/live` and `/health/ready` endpoints |
| `internal/app/app.go` | `Stop()` already exists — may need minor refinement for error ordering |
| `internal/app/bootstrap.go` | Wire Redis cache when `CACHE_DRIVER != none` (out of scope for 07a, but readiness check needs it) |
| `internal/infra/cache/redis.go` | TODO stub — needs implementation for `/health/ready` cache check |
| `internal/config/config.go` | Already complete — no changes needed |
| `internal/middleware/recovery.go` | Already complete — optional: add stack trace |
| `internal/infra/logger/logger.go` | Already complete — no changes needed |
| `internal/middleware/logger.go` | Already complete — no changes needed |
| `cmd/api/main.go` | Already complete — no changes needed |
| `.env.example` | Already has all required vars |
| `internal/server/server_test.go` | Add tests for `/health/live` and `/health/ready` |

## Approaches

### 1. Minimal additions (recommended)
- Add `/health/live` → simple 200 returning "alive" status.
- Add `/health/ready` → check DB ping + cache ping (if active), return aggregated status.
- Create a `health` package or inline handlers in router.
- Add tests for both endpoints.

**Pros**: Small surface area, clear scope, no new dependencies.
**Cons**: Inline handlers in router.go could grow; better to extract if more checks are added later.
**Effort**: Low

### 2. Dedicated health package
- Create `internal/infra/health/` with `Checker` interface.
- Implement `DBChecker`, `CacheChecker`, `LivenessChecker`.
- Compose into a `CompositeChecker` for `/health/ready`.
- Router delegates to checker.

**Pros**: Extensible, testable, follows existing infra pattern.
**Cons**: More files for a small change, may be over-engineering for 3 endpoints.
**Effort**: Medium

### 3. Health as middleware
- Health checks as middleware that runs before routing.
- Not recommended — health endpoints should be explicit routes, not middleware.

**Pros**: N/A
**Cons**: Wrong abstraction, harder to test individually.
**Effort**: Medium

## Recommendation

**Approach 1 (Minimal)** with a small extraction: create `internal/server/health.go` with handler functions for `/health`, `/health/live`, and `/health/ready`. This keeps router.go clean while avoiding over-engineering. The `/health/ready` handler needs access to DB and Cache, so it should accept them as parameters.

Key implementation details:
- `/health/live`: returns 200 with `{"status": "alive"}` — no dependencies needed.
- `/health/ready`: pings DB (`sqlDB.Ping()`), pings cache if driver != "none" (call `cache.Get` with a probe key or add `Ping` to interface). Returns 200 if all healthy, 503 if any dependency fails.
- `/health` (existing): keep as-is or alias to `/health/live`.

## Risks

1. **Cache readiness check blocked by Redis TODO**: `RedisCache` is a stub. The `/health/ready` endpoint can only check the NoopCache (always succeeds) until change-07b implements Redis. This should be documented as a known limitation.
2. **Health endpoint coupling**: `/health/ready` needs DB + Cache access. If wired through `App`, it creates a dependency from `server` package to `infra` packages. Better to pass the DB `*sql.DB` (via `db.DB()`) and `cache.Cache` interface to the health handlers.
3. **Error aggregation**: If both DB and cache fail, the response should indicate which specific dependency failed, not just a generic "unhealthy".
4. **No existing health test for readiness**: Current `server_test.go` only tests basic `/health`. New tests needed for live/ready scenarios including failure cases.

## Ready for Proposal

**Yes** — the exploration is complete. The codebase already has ~80% of the observability infrastructure (logger, middleware, graceful shutdown, config). The remaining work is:
1. Add `/health/live` and `/health/ready` endpoints (new).
2. Wire Redis cache for readiness check (blocked by change-07b, but interface can be prepared).
3. Add tests for health endpoints.

The scope is well-defined and low-risk.
