# App Orchestration Specification

## Purpose
Define the `App` type, bootstrap sequence, dependency container, and lifecycle methods.

## Requirements

### Requirement: App struct

The system MUST define an `App` struct that holds references to the server, database, logger, configuration, and cache.
(Previously: Unchanged by slice reorganization)

#### Scenario: Access dependencies

- GIVEN a successfully bootstrapped `App`
- WHEN `app.DB`, `app.Server`, `app.Logger`, `app.Config`, or `app.Cache` are accessed
- THEN they are non-nil and initialized

### Requirement: Bootstrap sequence

The system MUST enforce the bootstrap order: load config → initialize logger → connect database → initialize cache (driver-based factory) → build container → setup router → start server.
(Previously: Unchanged order; container build step delegates differently)

#### Scenario: Valid environment

- GIVEN all environment variables are valid
- WHEN `Bootstrap()` is called
- THEN each step executes in order
- AND the returned `*App` is ready to start

#### Scenario: Invalid database config

- GIVEN `DB_PASSWORD` is invalid
- WHEN `Bootstrap()` reaches the database step
- THEN it returns an error
- AND the server is never started

#### Scenario: Cache initialized during bootstrap

- GIVEN `CACHE_DRIVER=redis` and Redis is reachable
- WHEN `Bootstrap()` reaches the cache step
- THEN a `RedisCache` is created and assigned to `app.Cache`

#### Scenario: Cache fallback on connection failure

- GIVEN `CACHE_DRIVER=redis` but Redis is unreachable
- WHEN `Bootstrap()` reaches the cache step
- THEN a `NoopCache` is assigned to `app.Cache` and a warning is logged
- AND bootstrap continues (does not fail)

### Requirement: Dependency container

The system MUST provide a `Container` type that wires repositories, services, and handlers, and is built during bootstrap. The root container MUST delegate module wiring to module-level `NewContainer(db)` functions. The root container MUST NOT instantiate individual repositories, services, or handlers for modules using vertical slices.
(Previously: Root container wired all modules' layers directly; now delegates to module containers)

#### Scenario: Container wiring via module containers

- GIVEN bootstrap succeeds
- WHEN `app.Container` is inspected
- THEN it contains module-level containers (e.g., `CompaniesContainer`) returned by each module's `NewContainer(db)`
- AND it does NOT contain individual handler/service/repository fields for migrated modules

#### Scenario: Root container imports module root only

- GIVEN the companies module uses endpoint-aligned slices
- WHEN `internal/app/container.go` imports companies wiring
- THEN it imports the root `internal/modules/companies` package
- AND it does NOT import companies slice packages

#### Scenario: Module container is called by root

- GIVEN the root container is being built
- WHEN wiring the companies module
- THEN `companies.NewContainer(db)` is called
- AND the returned container is stored on the root container

### Requirement: Start and Stop lifecycle

The system MUST expose `Start()` to run the server and `Stop(ctx)` to release resources including the cache connection.
(Previously: Unchanged by slice reorganization)

#### Scenario: Start server

- GIVEN a bootstrapped `App`
- WHEN `Start()` is called
- THEN the HTTP server begins listening

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

## Constraints and Edge Cases

- `Bootstrap()` MUST be idempotent within a single process; calling it twice without cleanup SHOULD return a new `App` each time (for tests).
- Any bootstrap failure MUST return an error with the step name (e.g., "db connect: ...").
- `Container` MUST NOT import business modules directly; modules are registered by the caller.
