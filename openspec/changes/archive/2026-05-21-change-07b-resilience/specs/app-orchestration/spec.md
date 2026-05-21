# Delta for App Orchestration

## MODIFIED Requirements

### Requirement: App struct

The system MUST define an `App` struct that holds references to the server, database, logger, configuration, and cache.

(Previously: server, database, logger, and configuration)

#### Scenario: Access dependencies

- GIVEN a successfully bootstrapped `App`
- WHEN `app.DB`, `app.Server`, `app.Logger`, `app.Config`, or `app.Cache` are accessed
- THEN they are non-nil and initialized

### Requirement: Bootstrap sequence

The system MUST enforce the bootstrap order: load config → initialize logger → connect database → initialize cache (driver-based factory) → build container → setup router → start server.

(Previously: load config → initialize logger → connect database → build container → setup router → start server)

#### Scenario: Cache initialized during bootstrap

- GIVEN `CACHE_DRIVER=redis` and Redis is reachable
- WHEN `Bootstrap()` reaches the cache step
- THEN a `RedisCache` is created and assigned to `app.Cache`

#### Scenario: Cache fallback on connection failure

- GIVEN `CACHE_DRIVER=redis` but Redis is unreachable
- WHEN `Bootstrap()` reaches the cache step
- THEN a `NoopCache` is assigned to `app.Cache` and a warning is logged
- AND bootstrap continues (does not fail)

### Requirement: Start and Stop lifecycle

The system MUST expose `Start()` to run the server and `Stop(ctx)` to release resources including the cache connection.

(Previously: Stop closed HTTP server and database connection)

#### Scenario: Stop server closes cache

- GIVEN a running `App` with an active `RedisCache`
- WHEN `Stop(ctx)` is called
- THEN the HTTP server shuts down gracefully
- AND the database connection is closed
- AND `Cache.Close()` is called

#### Scenario: Stop with NoopCache is safe

- GIVEN a running `App` with `NoopCache`
- WHEN `Stop(ctx)` is called
- THEN `Cache.Close()` returns `nil` without error
