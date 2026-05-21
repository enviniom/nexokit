# Design: Resilience Infrastructure — Cache & Rate Limiting

## Technical Approach

Complete the existing cache/rate-limit scaffolding without changing module boundaries. Bootstrap will build cache and limiter dependencies from env config, keep `CACHE_DRIVER=none` as the safe default, and wire stricter IP-based limits only on auth login/refresh while preserving the current health/readiness behavior.

## Architecture Decisions

| Topic | Choice | Alternatives considered | Rationale |
|------|--------|-------------------------|-----------|
| Cache contract | Keep `Close()` and add `Exists(ctx,key)` | Replace `Close()` with `Exists()` | `Close()` is already used by `App.Stop()`; `Exists()` supports cache-aside checks without fetching values. |
| Redis rate limiting | Lua script with atomic increment + TTL | Separate `INCR` then `EXPIRE`; sorted-set sliding window | Separate commands race under concurrency; sorted sets are more precise but too complex for this change. |
| Default limiter | In-memory `x/time/rate` per IP with cleanup | Redis-only limiter | Local default preserves zero-Redis deployments and fits `RATE_LIMIT_ENABLED=false` rollback. |
| Redis startup | Cache/limiter factories fall back safely when disabled; Redis mode logs and falls back to no-op cache/local limiter on connection failure | Fail app startup | Resilience infra must not break existing apps that do not require Redis. |

## Data Flow

```text
request ─→ base middleware ─→ /api/v1/auth/login|refresh
                               │
                               └─→ RateLimitMiddleware
                                     ├─ local: IP bucket Allow()
                                     └─ redis: EVAL Lua increment+TTL
                                           allowed? continue : 429 envelope

Bootstrap ─→ cache factory ─→ Container/services + HealthDeps
          └→ limiter factory ─→ auth route middleware ─→ App.Stop closes cleanup/client
```

Redis Lua flow: compute key `rl:{scope}:{ip}`, `INCR`, if value is `1` then `EXPIRE window`, read `TTL`, return `{allowed, count, ttl}`. Middleware returns `response.TooManyRequests` when `allowed=false`; Redis errors fail open with a warning to avoid auth outage.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `go.mod`, `go.sum` | Modify | Add `github.com/redis/go-redis/v9` and promote `golang.org/x/time/rate` if needed. |
| `internal/infra/cache/cache.go` | Modify | Add `Exists(ctx,key) (bool,error)` to `Cache`. |
| `internal/infra/cache/noop.go`, `noop_test.go` | Modify | `Exists` returns `false,nil`; extend tests. |
| `internal/infra/cache/redis.go`, `redis_test.go` | Modify/Create | Implement Redis get/set/delete/exists/close and normalize cache misses to `nil,nil`. |
| `internal/config/config.go`, `config_test.go` | Modify | Add Redis config, default TTL, and rate-limit global/login/refresh fields. |
| `internal/middleware/rate_limit.go`, `rate_limit_test.go` | Modify/Create | Add limiter interfaces, local/Redis/noop implementations, middleware, IP key extraction, cleanup. |
| `internal/app/bootstrap.go`, `app.go`, `container.go` | Modify | Build cache/limiter from config, pass limiter into auth registration, close limiter on shutdown. |
| `internal/server/router.go`, tests | Modify | Preserve base middleware; no global limiter unless enabled/configured for global scope. |
| `internal/modules/auth/routes.go` | Modify | Accept login/refresh middleware funcs and apply before handlers. |
| `internal/platform/{apperror,response,messages}` | Modify | Add 429 sentinel, response helper, and message. |
| `.env.example`, `docker-compose.yml` | Modify | Document Redis/rate-limit env vars and add local Redis service + volume/healthcheck. |

## Interfaces / Contracts

```go
type Cache interface {
  Get(context.Context, string) ([]byte, error)
  Set(context.Context, string, []byte, time.Duration) error
  Delete(context.Context, string) error
  Exists(context.Context, string) (bool, error)
  Close() error
}

type Limiter interface {
  Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
  Close() error
}
```

Config defaults: cache driver `none`, Redis `localhost:6379`, DB `0`, default TTL `300s`; rate limiting enabled `true`, global disabled by limit `0`, login `5/min`, refresh `10/min`, cleanup interval `5m`.

## Health Readiness Compatibility

`/health/ready` currently probes cache with `Get(ctx,"health:probe")` only when cache is enabled. `RedisCache.Get` must treat `redis.Nil` as a healthy miss (`nil,nil`) so an absent probe key does not fail readiness; connection/timeouts still make readiness return 503.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | config defaults/invalid ints, noop/redis cache contract, local limiter cleanup, Lua result parsing, 429 helper | Table-driven tests; Redis tests behind short-skip or miniredis/test container if available. |
| HTTP | login/refresh allow then 429, disabled limiter passes, IP extraction | `httptest` with Gin test mode and fake limiter. |
| Lifecycle | `App.Stop` closes cache and limiter; cleanup goroutine stops | small fakes plus context timeout. |
| Integration | Redis cache and Redis limiter atomic behavior | Skippable when Redis is unavailable; run in normal local compose path. |

## Migration / Rollout

No data migration required. Roll out with defaults first (`CACHE_DRIVER=none`, local limiter). Enable Redis cache by setting `CACHE_DRIVER=redis`; enable distributed rate limiting by selecting Redis limiter config. Roll back via `RATE_LIMIT_ENABLED=false` or `CACHE_DRIVER=none`.

## Open Questions

None.
