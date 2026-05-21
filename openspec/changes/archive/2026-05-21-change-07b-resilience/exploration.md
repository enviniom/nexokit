# Exploration: change-07b-resilience — Cache & Rate Limiting

## Current State

The codebase has **partial scaffolding** for both cache and rate limiting:

- **Cache interface** exists at `internal/infra/cache/cache.go` with `Get`, `Set`, `Delete`, `Close` methods.
- **NoopCache** is fully implemented and tested at `internal/infra/cache/noop.go`.
- **RedisCache** is a TODO stub (single comment line).
- **Rate limiter middleware** is a TODO stub (single comment line).
- **go-redis** is NOT in `go.mod` — no Redis client dependency exists yet.
- `CacheConfig` only has `Driver` field; no Redis connection params.
- `bootstrap.go` hardcodes `cache.NewNoop()` — no driver-based factory.
- `.env.example` has `CACHE_DRIVER=none` but no Redis env vars.
- `docker-compose.yml` only includes PostgreSQL.
- `response` package has no `TooManyRequests` helper.
- `apperror` has no `ErrTooManyRequests` sentinel.
- `messages` has no `MsgTooManyRequests` constant.

## Affected Areas

| File | Why Affected |
|------|-------------|
| `go.mod` | Add `github.com/redis/go-redis/v9` dependency |
| `internal/config/config.go` | Extend `CacheConfig` with Redis params; add `RateLimitConfig` |
| `internal/config/env.go` | No changes needed |
| `internal/infra/cache/redis.go` | Implement `RedisCache` (currently TODO stub) |
| `internal/infra/cache/cache.go` | Interface mismatch: prompt requests `Exists()` method, current interface has `Close()`. Decision needed. |
| `internal/infra/cache/noop.go` | Already complete. May need `Exists()` if interface changes. |
| `internal/infra/cache/noop_test.go` | Already complete. May need update if interface changes. |
| `internal/middleware/rate_limit.go` | Implement rate limiter (currently TODO stub) |
| `internal/app/bootstrap.go` | Replace hardcoded `NewNoop()` with driver-based cache factory |
| `internal/app/container.go` | Pass rate limiter config to middleware wiring |
| `internal/server/router.go` | Wire rate limit middleware into global/auth route groups |
| `internal/modules/auth/routes.go` | Apply rate limit middleware to login/refresh endpoints |
| `internal/platform/response/response.go` | Add `TooManyRequests` helper |
| `internal/platform/apperror/apperror.go` | Add `ErrTooManyRequests` sentinel |
| `internal/platform/messages/messages.go` | Add `MsgTooManyRequests` constant |
| `.env.example` | Add Redis and rate limit env vars |
| `docker-compose.yml` | Add Redis/Valkey service |

## Approaches

### 1. Cache Interface: Keep `Close()`, Add `Exists()`

- **Description**: Extend the existing `Cache` interface to include `Exists()` while keeping `Close()`. Both methods are useful — `Close()` for lifecycle, `Exists()` for cache-aside patterns.
- **Pros**: Non-breaking for existing `Close()` usage in `app.Stop()`; `Exists()` enables efficient cache checks without fetching data.
- **Cons**: Interface grows to 5 methods.
- **Effort**: Low

### 2. Cache Interface: Replace `Close()` with `Exists()`

- **Description**: Match the prompt spec exactly — swap `Close()` for `Exists()`. Move `Close()` to a separate `Closer` interface or call it directly on concrete types.
- **Pros**: Matches prompt spec exactly.
- **Cons**: Breaks `app.Stop()` which calls `a.Cache.Close()`. Requires refactoring shutdown logic.
- **Effort**: Medium

### 3. Rate Limiter: In-Memory Token Bucket (Default) + Redis (Optional)

- **Description**: Use `golang.org/x/time/rate` for in-memory limiting (single instance). Add Redis-backed limiter using `INCR` + `EXPIRE` for distributed scenarios.
- **Pros**: Zero dependencies for default case; `golang.org/x/time/rate` is already an indirect dependency via other packages. Simple, battle-tested algorithm.
- **Cons**: In-memory limiter doesn't share state across instances.
- **Effort**: Medium

### 4. Rate Limiter: Sliding Window Log (Redis-only)

- **Description**: Use Redis sorted sets for precise sliding window rate limiting.
- **Pros**: More accurate than token bucket; distributed by default.
- **Cons**: Requires Redis even for basic rate limiting; higher memory usage; more complex.
- **Effort**: High

## Recommendation

**Approach 1** for cache interface (add `Exists()`, keep `Close()`) — both methods serve different purposes and the prompt's interface is a reasonable superset.

**Approach 3** for rate limiter (in-memory default, Redis optional) — matches the prompt's requirement of "memoria local por defecto" and keeps the no-dependency default path clean.

### Implementation Strategy

1. **Add `go-redis/v9`** to `go.mod`
2. **Extend `Cache` interface** with `Exists()` method (keep `Close()`)
3. **Implement `RedisCache`** using `go-redis` with connection pooling
4. **Update `CacheConfig`** with `RedisAddr`, `RedisPassword`, `RedisDB`, `DefaultTTL`
5. **Add rate limit config** struct with global and per-endpoint settings
6. **Implement in-memory rate limiter** using `golang.org/x/time/rate` with per-IP buckets
7. **Implement Redis rate limiter** using `INCR` + `EXPIRE` for distributed mode
8. **Create rate limit middleware** that accepts a `Limiter` interface
9. **Wire rate limiting** to auth routes (login, refresh) with stricter limits
10. **Add `TooManyRequests`** response helper, `ErrTooManyRequests` sentinel, and message constant
11. **Update `bootstrap.go`** with driver-based cache factory
12. **Add Redis to `docker-compose.yml`** and env vars to `.env.example`

## Risks

- **Interface mismatch**: The prompt's `Cache` interface differs from the existing one (`Exists` vs `Close`). This needs explicit resolution.
- **go-redis version**: Need to select the correct major version (v9 is the current stable).
- **Rate limiter state cleanup**: In-memory limiter needs periodic cleanup of expired IP entries to prevent memory leaks.
- **Distributed rate limiting**: Redis-based rate limiting with `INCR`+`EXPIRE` has race conditions under high concurrency. Consider Lua scripts for atomicity.
- **Testing**: Redis integration tests need a real Redis instance or mock. The existing `NoopCache` test pattern should be extended.

## Ready for Proposal

**Yes** — sufficient codebase understanding exists to proceed to proposal phase. The main decision point is the cache interface (`Exists` + `Close` vs replace), which can be resolved during proposal.
