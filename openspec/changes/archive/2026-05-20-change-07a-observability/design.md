# Design: Observability Infrastructure — Health Endpoints

## Technical Approach

Add small server-local health handlers in `internal/server/health.go` and wire them from `router.go` before the `/api/v1` group. Bootstrap will extract the underlying `*sql.DB` from GORM and pass it, plus the existing cache instance and `CACHE_DRIVER` state, into the router. This keeps health checks unversioned, unauthenticated, and outside business modules.

## Architecture Decisions

| Decision | Alternatives considered | Rationale |
|----------|--------------------------|-----------|
| Keep health in `internal/server` | New `internal/infra/health` package | Health is HTTP-edge behavior here; a new infra package adds ceremony for three endpoints. |
| Pass a `HealthDeps` struct into `NewRouter` | Let `server` import app/db packages, use globals | Explicit dependency injection follows current bootstrap style and avoids server→app coupling. |
| Use small local interfaces (`dbPinger`, `cacheGetter`) | Expand `cache.Cache` with `Ping`, concrete Redis checks | Avoids breaking existing cache fakes and keeps 07a compatible with the Redis stub until 07b. |
| Register routes on the base Gin engine | Register under `/api/v1`, or inside auth groups | Orchestrators/load balancers expect stable unauthenticated probes. Base engine avoids auth middleware. |

## Data Flow

    app.Bootstrap
        ├─ db.Connect() -> *gorm.DB -> sqlDB := database.DB()
        ├─ cache.NewNoop() (07a) -> cache dependency
        └─ server.NewRouter(cfg, log, ginWriter, HealthDeps, modules)

    GET /health/live  ──→ liveHandler  ──→ 200 {status: alive}
    GET /health/ready ──→ readyHandler ──→ DB PingContext
                                      └──→ cache Get probe only when CACHE_DRIVER != "none"
                                      └──→ 200 ready or 503 not_ready

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/server/health.go` | Create | Defines `HealthDeps`, response structs, `/health`, `/health/live`, `/health/ready` handlers, and dependency aggregation. |
| `internal/server/router.go` | Modify | Replace inline `/health`; register `/health`, `/health/live`, `/health/ready` before `/api/v1`; update `NewRouter` signature. |
| `internal/app/bootstrap.go` | Modify | Extract `sqlDB` from GORM, derive `CacheEnabled` from `cfg.Cache.Driver != "none"`, pass `HealthDeps`. |
| `internal/server/server_test.go` | Modify | Add table-driven health endpoint tests and dependency failure coverage. |

## Interfaces / Contracts

```go
type HealthDeps struct {
    DB           dbPinger
    Cache        cacheGetter
    CacheEnabled bool
}
type dbPinger interface { PingContext(context.Context) error }
type cacheGetter interface { Get(context.Context, string) ([]byte, error) }
```

`/health` keeps the current envelope: `success=true`, message `API is healthy`, `data.status="ok"`.
`/health/live` returns plain health JSON: `{"status":"alive"}`.
`/health/ready` returns status plus per-dependency records. DB failure or active-cache failure returns HTTP 503. If cache is disabled, report it as skipped, not failed. If DB is nil, readiness fails; if cache is enabled but nil, readiness fails.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|--------------|----------|
| Unit | Readiness aggregation | Table-driven tests with fake DB pinger/cache getter; success, DB fail, cache fail, cache disabled, nil deps. |
| HTTP | Route behavior/status/body | `httptest` against `NewRouter`; assert `/health` compatibility, `/health/live` 200, `/health/ready` 200/503. |
| Integration | Full suite regression | Run `go test ./...`; no external Redis required. |

## Migration / Rollout

No migration required. Rollout is additive, except `NewRouter` signature changes require same-commit updates to bootstrap and tests.

## Open Questions

- None.
