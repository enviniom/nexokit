# Proposal: Resilience Infrastructure — Cache & Rate Limiting

## Intent

Add optional Redis/Valkey caching and distributed rate limiting to NexoKit, enabling cache-aside patterns and protecting sensitive endpoints (login, refresh) from abuse. Projects without cache needs remain unaffected via `CACHE_DRIVER=none`.

## Scope

### In Scope
- Cache adapter interface with `Get`, `Set`, `Delete`, `Exists`, `Close`
- `RedisCache` implementation using `go-redis/v9`
- `NoopCache` (already exists, add `Exists`)
- Rate limiter interface with in-memory (default) and Redis (distributed) backends
- Rate limit middleware applied to login/refresh endpoints
- `ErrTooManyRequests` sentinel, `TooManyRequests` response helper, message constant
- Redis env vars (`REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`, `CACHE_DEFAULT_TTL_SECONDS`)
- Rate limit env vars (`RATE_LIMIT_ENABLED`, `RATE_LIMIT_REQUESTS`, `RATE_LIMIT_WINDOW_SECONDS`, `LOGIN_RATE_LIMIT_*`)
- Redis service in `docker-compose.yml`

### Out of Scope
- Cache warming / invalidation strategies
- Rate limit UI or admin dashboard
- Per-user rate limiting (IP-based only for this change)
- Redis cluster / sentinel support

## Capabilities

### New Capabilities
- `cache-adapter`: Cache interface contract, RedisCache, NoopCache implementations, driver-based factory
- `rate-limiting`: Rate limiter interface, in-memory and Redis backends, rate limit middleware

### Modified Capabilities
- `http-middleware`: Add rate limit middleware to the stack order
- `error-handling`: Add `ErrTooManyRequests` sentinel mapped to HTTP 429
- `api-response`: Add `TooManyRequests` response helper
- `environment-config`: Add Redis and rate limit configuration fields
- `app-orchestration`: Cache lifecycle (`Close`) and driver-based factory in bootstrap
- `server-bootstrap`: Wire rate limit middleware into route groups

## Approach

**Cache interface**: Extend existing `Cache` interface to add `Exists()` while keeping `Close()`. `Close()` is needed for lifecycle shutdown in `app.Stop()`. `Exists()` enables efficient cache-aside checks without fetching data.

**Redis client**: Use `go-redis/v9` with connection pooling. Factory in `bootstrap.go` selects driver (`redis` vs `none`).

**Rate limiting**: In-memory default using `golang.org/x/time/rate` with per-IP token buckets and periodic cleanup. Redis backend uses a Lua script for atomic `INCR` + `EXPIRE` to avoid race conditions under high concurrency.

**Middleware**: `RateLimitMiddleware(limiter, config)` accepts a `Limiter` interface. Applied globally (optional) and per-route for login/refresh with stricter limits.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/infra/cache/cache.go` | Modified | Add `Exists()` method to interface (keep `Close()`) |
| `internal/infra/cache/redis.go` | New | `RedisCache` implementation with `go-redis/v9` |
| `internal/infra/cache/noop.go` | Modified | Add `Exists()` stub |
| `internal/config/config.go` | Modified | Add `CacheConfig` fields + `RateLimitConfig` struct |
| `internal/app/bootstrap.go` | Modified | Driver-based cache factory, pass rate limit config |
| `internal/middleware/rate_limit.go` | Modified | Replace TODO with full rate limiter + middleware |
| `internal/platform/apperror/apperror.go` | Modified | Add `ErrTooManyRequests` sentinel |
| `internal/platform/response/response.go` | Modified | Add `TooManyRequests` helper |
| `internal/platform/messages/messages.go` | Modified | Add `MsgTooManyRequests` constant |
| `internal/server/router.go` | Modified | Wire rate limit middleware |
| `internal/modules/auth/routes.go` | Modified | Apply rate limit to login/refresh |
| `.env.example` | Modified | Add Redis + rate limit variables |
| `docker-compose.yml` | Modified | Add Redis/Valkey service |
| `go.mod` | Modified | Add `github.com/redis/go-redis/v9` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Redis connection failure at startup | Medium | Factory returns `NoopCache` with warning; does not crash |
| In-memory rate limiter memory leak | Medium | Periodic cleanup goroutine with configurable interval |
| Redis rate limit race conditions | Low | Lua script ensures atomic INCR+EXPIRE in single operation |
| Interface growth (5 methods) | Low | Acceptable tradeoff; `Close` and `Exists` serve distinct purposes |

## Rollback Plan

1. Set `CACHE_DRIVER=none` to disable Redis cache (uses `NoopCache`)
2. Set `RATE_LIMIT_ENABLED=false` to disable rate limiting entirely
3. Remove Redis service from `docker-compose.yml`
4. Revert `go.mod` dependency if needed

## Dependencies

- `github.com/redis/go-redis/v9` (new)
- `golang.org/x/time/rate` (already indirect dependency)

## Success Criteria

- [ ] `RedisCache` implements full `Cache` interface with connection pooling
- [ ] `NoopCache` implements full `Cache` interface (including `Exists`)
- [ ] Rate limiter returns HTTP 429 when limits exceeded
- [ ] Login/refresh endpoints have stricter rate limits than global
- [ ] `CACHE_DRIVER=none` uses `NoopCache` without Redis dependency
- [ ] Redis Lua script handles atomic distributed rate limiting
- [ ] `docker-compose.yml` includes Redis for local development
- [ ] All existing tests pass; new cache/rate-limit tests added
