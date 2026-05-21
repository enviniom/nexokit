# Cache Adapter Specification

## Purpose

Define the cache interface contract, Redis/Valkey implementation, NoopCache fallback, and driver-based factory for optional caching in NexoKit.

## Requirements

### Requirement: Cache interface contract

The system MUST define a `Cache` interface with methods: `Get(ctx, key) (value, error)`, `Set(ctx, key, value, ttl) error`, `Delete(ctx, key) error`, `Exists(ctx, key) (bool, error)`, and `Close() error`.

#### Scenario: Interface satisfaction — RedisCache

- GIVEN `RedisCache` implements all five methods
- WHEN compiled with `go build ./...`
- THEN no interface mismatch errors occur

#### Scenario: Interface satisfaction — NoopCache

- GIVEN `NoopCache` implements all five methods
- WHEN compiled with `go build ./...`
- THEN no interface mismatch errors occur

#### Scenario: Exists returns true for set key

- GIVEN a key was set with `Set(ctx, "user:1", data, ttl)`
- WHEN `Exists(ctx, "user:1")` is called
- THEN it returns `(true, nil)`

#### Scenario: Exists returns false for missing key

- GIVEN no value was set for key `"missing:key"`
- WHEN `Exists(ctx, "missing:key")` is called
- THEN it returns `(false, nil)`

### Requirement: RedisCache implementation

The system MUST provide a `RedisCache` backed by `go-redis/v9` that serializes values as JSON and respects TTL on `Set`.

#### Scenario: Get retrieves set value

- GIVEN `Set(ctx, "k", "v", 60s)` succeeded
- WHEN `Get(ctx, "k")` is called
- THEN it returns the original value with no error

#### Scenario: Get returns ErrCacheMiss for absent key

- GIVEN no value exists for key `"absent"`
- WHEN `Get(ctx, "absent")` is called
- THEN it returns an error matching `ErrCacheMiss`

#### Scenario: Set with TTL expires value

- GIVEN `Set(ctx, "k", "v", 1s)` succeeded
- WHEN 2 seconds pass and `Get(ctx, "k")` is called
- THEN it returns `ErrCacheMiss`

#### Scenario: Delete removes key

- GIVEN `Set(ctx, "k", "v", 60s)` succeeded
- WHEN `Delete(ctx, "k")` is called
- THEN `Get(ctx, "k")` returns `ErrCacheMiss`

#### Scenario: Close releases connection

- GIVEN a connected `RedisCache`
- WHEN `Close()` is called
- THEN subsequent `Get`/`Set` calls return an error

### Requirement: NoopCache implementation

The system MUST provide a `NoopCache` that implements the full `Cache` interface as a no-op: `Get` returns `ErrCacheMiss`, `Set`/`Delete`/`Exists` return `nil`/`false`/`nil`, and `Close` returns `nil`.

#### Scenario: Get always returns cache miss

- GIVEN a `NoopCache` instance
- WHEN `Get(ctx, "any-key")` is called
- THEN it returns `("", ErrCacheMiss)`

#### Scenario: Set is a silent no-op

- GIVEN a `NoopCache` instance
- WHEN `Set(ctx, "k", "v", 60s)` is called
- THEN it returns `nil` with no side effects

#### Scenario: Exists always returns false

- GIVEN a `NoopCache` instance
- WHEN `Exists(ctx, "any-key")` is called
- THEN it returns `(false, nil)`

### Requirement: Driver-based cache factory

The system MUST provide a factory that selects the cache implementation based on `CACHE_DRIVER`: `"redis"` creates `RedisCache`, `"none"` returns `NoopCache`.

#### Scenario: Redis driver creates RedisCache

- GIVEN `CACHE_DRIVER=redis` and Redis is reachable
- WHEN the factory is invoked
- THEN it returns a `*RedisCache` instance

#### Scenario: None driver returns NoopCache

- GIVEN `CACHE_DRIVER=none`
- WHEN the factory is invoked
- THEN it returns a `*NoopCache` instance

#### Scenario: Redis connection failure falls back to NoopCache

- GIVEN `CACHE_DRIVER=redis` but Redis is unreachable
- WHEN the factory is invoked
- THEN it returns a `*NoopCache` and logs a warning

## Constraints and Edge Cases

- `Get` values MUST be JSON-serialized; deserialization errors MUST be returned as errors.
- TTL MUST be clamped to a minimum of 1 second.
- `Close()` MUST be safe to call multiple times (idempotent).
- The factory MUST NOT block indefinitely on Redis connection — it MUST use a configurable dial timeout.
