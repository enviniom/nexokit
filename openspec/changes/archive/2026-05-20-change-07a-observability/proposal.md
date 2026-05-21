# Proposal: Observability Infrastructure — Health Endpoints

## Intent

NexoKit already has structured logging (slog + lumberjack), request logging middleware, recovery middleware, graceful shutdown, and full config for observability env vars. The only missing piece is extended health check endpoints (`/health/live`, `/health/ready`) needed for container orchestration and load balancer readiness probes.

## Scope

### In Scope
- Create `internal/server/health.go` with handlers for `/health`, `/health/live`, `/health/ready`
- `/health/live`: returns 200 with `{"status": "alive"}` — no dependencies
- `/health/ready`: pings DB and cache (if active), returns 200 or 503 with per-dependency status
- Wire health routes in `internal/server/router.go`
- Add tests for all three health endpoints (success + failure scenarios)
- Register health routes before auth middleware (no auth required)

### Out of Scope
- Redis cache implementation (blocked by change-07b) — readiness check works with NoopCache only
- Stack trace capture in recovery middleware (optional enhancement)
- SIGHUP log rotation signal support
- Metrics/tracing endpoints (future observability work)

## Capabilities

### New Capabilities
- `health-checks`: `/health/live` and `/health/ready` endpoints with dependency aggregation, per-dependency status reporting, and 503 on readiness failure

### Modified Capabilities
- `server-bootstrap`: Add requirement that health routes are registered before auth middleware; clarify `/health/live` and `/health/ready` behavior alongside existing `/health`

## Approach

Extract health handlers from inline router code into `internal/server/health.go`. The `/health/ready` handler accepts `*sql.DB` and `cache.Cache` as parameters to avoid server→infra coupling. Returns aggregated JSON with per-dependency status. Routes registered on the base Gin engine (before `/api/v1/` group and auth middleware).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/server/health.go` | New | Health handler functions for live/ready/health |
| `internal/server/router.go` | Modified | Register health routes, delegate to health.go |
| `internal/server/server_test.go` | Modified | Add tests for live + ready endpoints |
| `openspec/specs/health-checks/spec.md` | New | Spec for health endpoint behavior |
| `openspec/specs/server-bootstrap/spec.md` | Modified | Delta: health route registration order |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Cache readiness always passes (NoopCache) until change-07b | High | Document as known limitation; interface supports future Redis ping |
| Health endpoint coupling to infra packages | Low | Pass `*sql.DB` and `cache.Cache` as params, not direct imports |
| Auth middleware blocking health routes | Low | Register health routes on base engine before auth group |

## Rollback Plan

Revert the commit adding `health.go` and router changes. The existing `/health` endpoint remains untouched in router.go history. No database migrations or config changes to rollback.

## Dependencies

- change-07b (Redis cache) for full `/health/ready` cache verification — not blocking, but limits readiness to DB-only until implemented

## Success Criteria

- [ ] `GET /health` returns 200 with existing behavior unchanged
- [ ] `GET /health/live` returns 200 with `{"status": "alive"}`
- [ ] `GET /health/ready` returns 200 when DB is connected
- [ ] `GET /health/ready` returns 503 when DB is disconnected, with per-dependency status
- [ ] Health endpoints do not require authentication
- [ ] All health endpoint tests pass with `go test ./...`
- [ ] `go vet ./...` passes with zero issues
