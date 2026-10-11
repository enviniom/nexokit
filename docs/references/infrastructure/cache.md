# Cache adapters and readiness

`internal/infra/cache` supplies `Get`, `Set`, `Delete`, `Exists`, `Close`. There are Redis and no-op implementations; **no in-memory cache adapter** is present.

## Drivers and fallback

| Configuration | Bootstrap behavior |
|---|---|
| `CACHE_DRIVER=none` (default) | No-op cache; reads return `ErrCacheMiss`, writes/deletes succeed without storage |
| `CACHE_DRIVER=redis` | Connect/ping Redis using `REDIS_HOST/PORT/PASSWORD/DB` and dial timeout |
| Redis connection failure or unsupported driver | Warning and no-op fallback |

Redis `Set` JSON-marshals the byte slice (thus encoding it as JSON/Base64), clamps TTL to at least one second, and writes it. `Get` unmarshals that representation; missing keys return `ErrCacheMiss`, transport/decoding failures remain errors. It is not a raw-byte optimization or a Redis Cluster-specific adapter. `Close` is idempotent; operations after close fail.

## Health caveat

Bootstrap considers cache enabled whenever configured driver is not `none`, even if it fell back to no-op. Readiness calls `Get("health:probe")`; the handler currently treats **any error except `context.Canceled` as unhealthy**, including `ErrCacheMiss`. Therefore a healthy empty Redis cache or enabled no-op fallback can return 503. Do not infer cache availability from the API's ability to start, or claim misses are accepted by readiness.

Configure optional caching deliberately and monitor errors; the code does not transparently switch a running Redis adapter to no-op on later failures. Caller-specific cache policy belongs with the consuming use case, not an invented global tag/invalidation system.

Sources: `internal/infra/cache/{cache,noop,redis}.go`, `internal/app/bootstrap.go`, `internal/server/health.go`. Probe rollout guidance lives in [deployment](../../deployment.md#health-and-observation).
