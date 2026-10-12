# Rate limiting

Keep limiter construction in bootstrap and route-scope selection in middleware registration. Apply login and refresh scopes explicitly; configuration alone does not attach middleware to routes.

## Drivers

| Driver | Behavior |
|---|---|
| `memory` (default) | Per-process token buckets; burst/average rate derive from limit and window; idle buckets cleaned periodically |
| `redis` | Shared fixed-window counter using atomic Lua `INCR`/`EXPIRE`; prefix `rl` |
| Redis startup failure / unsupported driver | Warning and local-limiter fallback |

`Limiter` exposes `Allow(ctx, key)` and `Close()`. Local cleanup defaults to five minutes. Local limit <= 0 means unlimited; Redis does not share that special behavior—zero limit denies requests.

## Scope and keys

Middleware keys are `<scope>:<client-ip>` (`login`, `refresh`); Redis adds its prefix. Client IP uses the first `X-Forwarded-For` value, otherwise RemoteAddr. It does not key on tenant, account or email. Overwrite untrusted forwarded headers and isolate the backend behind the proxy as specified in [deployment](../../deployment.md#tls-and-supervision).

## Operational caveats

Login/refresh registration does not establish a global router limit. Verify coverage for every surface requiring throttling.

Although configuration has login/refresh RPM fields (defaults 5/10), `RateLimitMiddleware` currently **ignores its limit and window arguments**. The shared limiter is built using `RATE_LIMIT_GLOBAL_RPM` (default 0) and `RATE_LIMIT_WINDOW_SECONDS` (default 60). Thus the default memory setup leaves those auth endpoints unlimited despite `RATE_LIMIT_ENABLED=true`; separate scopes do not establish separate quota values. Choose a positive global limit deliberately, and do not claim per-endpoint configured quotas work independently.

A denied request receives standard 429 JSON; no Retry-After header is added. Runtime limiter errors fail open (`c.Next()`); there is no later distributed-to-local failover. Local fallback is per instance, not a cluster-wide quota guarantee. `RATE_LIMIT_CLEANUP_INTERVAL_MINUTES` controls local bucket cleanup.

Test effective throttling through the actual wired routes before depending on it for abuse prevention. Endpoint-specific policy should not be inferred from configuration field names alone.
