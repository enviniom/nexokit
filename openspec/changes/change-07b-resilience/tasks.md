# Tasks: Resilience Infrastructure — Cache & Rate Limiting

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 650–800 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (foundation) → PR 2 (cache) → PR 3 (rate-limit + wiring) |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Add 429 error sentinel, response helper, message constant; extend config with Redis + rate-limit fields | PR 1 | Independent foundation; no runtime behavior change |
| 2 | Add `Exists` to Cache interface; implement RedisCache + tests; update NoopCache + tests; driver-based factory in bootstrap | PR 2 | Depends on PR 1 merged; cache-only scope |
| 3 | Implement Limiter interface, in-memory + Redis backends, middleware; wire into auth routes + bootstrap shutdown; update health checks, env, docker-compose | PR 3 | Depends on PR 1+2 merged; full integration |

## Phase 1: Foundation — Errors, Responses, Config (PR 1)

- [x] 1.1 Add `MsgTooManyRequests = "Demasiadas solicitudes. Intente nuevamente más tarde."` to `internal/platform/messages/messages.go`
- [x] 1.2 Add `ErrTooManyRequests` sentinel to `internal/platform/apperror/apperror.go` and map to 429 in `Status()`
- [x] 1.3 Add `TooManyRequests(c *gin.Context, message string)` helper to `internal/platform/response/response.go` returning 429 envelope
- [x] 1.4 Add `RedisConfig` struct (Host, Port, Password, DB, DialTimeout) and `RateLimitConfig` struct (Enabled, Driver, GlobalRPM, LoginRPM, RefreshRPM, WindowSeconds, CleanupIntervalMinutes) to `internal/config/config.go`
- [x] 1.5 Parse Redis env vars (`REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_DIAL_TIMEOUT_SECONDS`) in `config.Load()`
- [x] 1.6 Parse rate-limit env vars (`RATE_LIMIT_ENABLED`, `RATE_LIMIT_DRIVER`, `RATE_LIMIT_GLOBAL_RPM`, `RATE_LIMIT_LOGIN_RPM`, `RATE_LIMIT_REFRESH_RPM`, `RATE_LIMIT_WINDOW_SECONDS`, `RATE_LIMIT_CLEANUP_INTERVAL_MINUTES`) in `config.Load()`
- [x] 1.7 Add config unit tests: Redis defaults, rate-limit defaults, invalid int rejection for RPM fields and REDIS_PORT
- [x] 1.8 Update `.env.example` with all Redis and rate-limit variables and defaults

## Phase 2: Cache Infrastructure (PR 2)

- [ ] 2.1 Add `Exists(ctx context.Context, key string) (bool, error)` to `Cache` interface in `internal/infra/cache/cache.go`; define `ErrCacheMiss` sentinel
- [ ] 2.2 Add `Exists(ctx, key) (false, nil)` stub to `NoopCache` in `internal/infra/cache/noop.go`; fix `Get` to return `("", ErrCacheMiss)` per spec
- [ ] 2.3 Extend `TestNoopCache` in `internal/infra/cache/noop_test.go`: test `Exists` returns false, test `Get` returns ErrCacheMiss
- [ ] 2.4 Implement `RedisCache` in `internal/infra/cache/redis.go`: constructor with `go-redis/v9`, Get/Set/Delete/Exists/Close, JSON serialization, TTL clamp ≥1s, redis.Nil → ErrCacheMiss normalization, idempotent Close
- [ ] 2.5 Write `internal/infra/cache/redis_test.go`: table-driven tests for Get roundtrip, Get miss → ErrCacheMiss, TTL expiry, Delete, Exists true/false, Close idempotency, post-close error; skip real Redis tests via `testing.Short()`
- [ ] 2.6 Add driver-based cache factory in `internal/app/bootstrap.go`: switch on `CACHE_DRIVER` ("redis" → RedisCache with dial timeout, "none" → NoopCache); log warning and fall back to NoopCache on connection failure
- [ ] 2.7 Verify `go build ./...` passes with both cache implementations satisfying the interface

## Phase 3: Rate Limiting + Wiring (PR 3)

- [ ] 3.1 Define `Limiter` interface (`Allow(ctx, key) (bool, error)`, `Close() error`) and `NoopLimiter` in `internal/middleware/rate_limit.go`
- [ ] 3.2 Implement `LocalLimiter` using `golang.org/x/time/rate`: per-IP token bucket map, configurable rate/burst, `Allow()`, `Close()` stops cleanup goroutine
- [ ] 3.3 Implement periodic cleanup in `LocalLimiter`: configurable interval (default 5m), remove expired buckets, stop on `Close()`
- [ ] 3.4 Write `internal/middleware/rate_limit_test.go` — LocalLimiter: table-driven Allow under/over limit, token refill after wait, cleanup removes expired entries, Close stops goroutine
- [ ] 3.5 Implement `RedisLimiter`: Lua script for atomic INCR + conditional EXPIRE, key prefix `rl:{scope}:{ip}`, return `{allowed, count, ttl}`, fail on Redis error (caller decides)
- [ ] 3.6 Write RedisLimiter tests: Lua atomic increment, TTL on first request, exceeded limit returns false, connection error propagation; skip via `testing.Short()`
- [ ] 3.7 Implement `RateLimitMiddleware(limiter, enabled, scope, limit, window)` in `internal/middleware/rate_limit.go`: IP extraction (X-Forwarded-For first, fallback RemoteAddr stripped of port), call Allow, return `response.TooManyRequests` on false, pass-through when disabled
- [ ] 3.8 Write middleware HTTP tests via `httptest`: request within limit passes, over limit returns 429 with MsgTooManyRequests, disabled limiter bypasses, IP extraction from X-Forwarded-For and RemoteAddr
- [ ] 3.9 Add limiter factory in `internal/app/bootstrap.go`: select driver (memory/redis), fall back to LocalLimiter on Redis failure with warning
- [ ] 3.10 Update `App.Stop()` in `internal/app/app.go` to call `Limiter.Close()` after cache close
- [ ] 3.11 Update `Container` in `internal/app/container.go` to build login/refresh rate-limit middleware funcs from limiter + config
- [ ] 3.12 Update `auth.Register()` in `internal/modules/auth/routes.go` to accept login/refresh middleware funcs and apply before handlers
- [ ] 3.13 Update `readyHandler` in `internal/server/health.go`: RedisCache.Get normalizes redis.Nil as healthy miss (already handled by 2.4); no additional change needed if redis.Nil → nil,nil
- [ ] 3.14 Add Redis/Valkey service to `docker-compose.yml` with healthcheck and volume
- [ ] 3.15 Add `go-redis/v9` dependency: `go get github.com/redis/go-redis/v9` and `go mod tidy`
