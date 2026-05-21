# Rate Limiting Specification

## Purpose

Define rate limiter interface, in-memory and Redis backends, and middleware for protecting sensitive endpoints from abuse.

## Requirements

### Requirement: Rate limiter interface

The system MUST define a `RateLimiter` interface with methods: `Allow(ctx, key) (bool, error)` and `Close() error`.

#### Scenario: Allow returns true under limit

- GIVEN a rate limiter configured with 5 requests per 60s
- WHEN `Allow(ctx, "ip:1.2.3.4")` is called for the 3rd time within the window
- THEN it returns `(true, nil)`

#### Scenario: Allow returns false over limit

- GIVEN a rate limiter configured with 5 requests per 60s
- WHEN `Allow(ctx, "ip:1.2.3.4")` is called for the 6th time within the window
- THEN it returns `(false, nil)`

### Requirement: In-memory rate limiter (default)

The system MUST provide an in-memory rate limiter using `golang.org/x/time/rate` with per-IP token buckets and periodic cleanup of expired entries.

#### Scenario: Token bucket allows burst

- GIVEN a limiter with rate=5/sec, burst=5
- WHEN 5 requests arrive simultaneously for the same IP
- THEN all 5 are allowed

#### Scenario: Token bucket refills over time

- GIVEN a limiter with rate=1/sec, burst=1 that has exhausted its tokens
- WHEN 2 seconds pass
- THEN the next `Allow` call returns `true`

#### Scenario: Expired entries are cleaned up

- GIVEN an IP has not made requests for longer than the cleanup interval
- WHEN the cleanup goroutine runs
- THEN that IP's bucket is removed from memory

#### Scenario: Cleanup runs periodically

- GIVEN the in-memory limiter is initialized
- WHEN the application runs for longer than the cleanup interval
- THEN the cleanup goroutine has executed at least once

### Requirement: Redis rate limiter (distributed)

The system MUST provide a Redis-backed rate limiter that uses an atomic Lua script combining `INCR` and `EXPIRE` to avoid race conditions under concurrent access.

#### Scenario: Lua script increments atomically

- GIVEN a Redis rate limiter with limit=5, window=60s
- WHEN two concurrent processes call `Allow` for the same key simultaneously
- THEN the counter increments exactly once per call (no lost updates)

#### Scenario: Lua script sets TTL on first request

- GIVEN a key does not exist in Redis
- WHEN `Allow(ctx, "ip:1.2.3.4")` is called
- THEN the key is created with value 1 and TTL equal to the window

#### Scenario: Redis limiter returns false when exceeded

- GIVEN a Redis rate limiter with limit=3, window=60s
- WHEN `Allow` has been called 3 times for the same key within the window
- THEN the 4th call returns `(false, nil)`

#### Scenario: Redis connection failure returns safe default

- GIVEN Redis is unreachable
- WHEN `Allow(ctx, "ip:1.2.3.4")` is called on a Redis limiter
- THEN it returns an error (caller decides whether to allow or deny)

### Requirement: Rate limit middleware

The system MUST provide HTTP middleware that extracts the client IP, calls `RateLimiter.Allow`, and returns HTTP 429 when the limit is exceeded.

#### Scenario: Request within limit passes through

- GIVEN a rate limit middleware with limit=100/min
- WHEN a request arrives and `Allow` returns `true`
- THEN the request proceeds to the next handler

#### Scenario: Request over limit returns 429

- GIVEN a rate limit middleware with limit=5/min
- WHEN a request arrives and `Allow` returns `false`
- THEN the response is HTTP 429 with the standard error envelope
- AND the body uses `MsgTooManyRequests`

#### Scenario: IP extraction uses X-Forwarded-For

- GIVEN a request with `X-Forwarded-For: 10.0.0.1, 10.0.0.2`
- WHEN the middleware extracts the client IP
- THEN it uses the first IP (`10.0.0.1`) as the rate limit key

#### Scenario: IP extraction falls back to RemoteAddr

- GIVEN a request without `X-Forwarded-For`
- WHEN the middleware extracts the client IP
- THEN it uses `RemoteAddr` (stripped of port) as the rate limit key

### Requirement: Sensitive endpoint rate limiting

The system MUST apply stricter rate limits to login (`POST /api/v1/auth/login`) and refresh (`POST /api/v1/auth/refresh`) endpoints than the global rate limit.

#### Scenario: Login endpoint has stricter limit

- GIVEN global limit=100/min and login limit=10/min
- WHEN 11 login requests arrive from the same IP within 1 minute
- THEN the 11th request returns HTTP 429

#### Scenario: Refresh endpoint has stricter limit

- GIVEN global limit=100/min and refresh limit=20/min
- WHEN 21 refresh requests arrive from the same IP within 1 minute
- THEN the 21st request returns HTTP 429

#### Scenario: Global limit does not affect sensitive endpoint limit

- GIVEN an IP has used 90/100 global requests but 0/10 login requests
- WHEN a login request arrives
- THEN it is allowed (sensitive endpoint has its own counter)

### Requirement: Rate limit configuration

The system MUST support enabling/disabling rate limiting via `RATE_LIMIT_ENABLED` and configuring limits via `RATE_LIMIT_GLOBAL_RPM`, `RATE_LIMIT_LOGIN_RPM`, `RATE_LIMIT_REFRESH_RPM`, and `RATE_LIMIT_WINDOW_SECONDS`.

#### Scenario: Rate limiting disabled

- GIVEN `RATE_LIMIT_ENABLED=false`
- WHEN the middleware is applied
- THEN all requests pass through without rate limit checks

#### Scenario: Custom limits applied

- GIVEN `RATE_LIMIT_LOGIN_RPM=5` and `RATE_LIMIT_WINDOW_SECONDS=120`
- WHEN 6 login requests arrive within 120 seconds
- THEN the 6th request returns HTTP 429

## Constraints and Edge Cases

- The Lua script MUST use `INCR` + conditional `EXPIRE` (only set TTL when key is newly created) to avoid resetting TTL on every request.
- In-memory cleanup interval MUST be configurable (default: 5 minutes).
- Rate limit keys MUST be prefixed (e.g., `rl:global:ip:1.2.3.4`, `rl:login:ip:1.2.3.4`) to avoid collisions.
- `RateLimiter.Close()` MUST stop the cleanup goroutine for in-memory limiter.
