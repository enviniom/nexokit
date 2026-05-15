# App Orchestration Specification

## Purpose
Define the `App` type, bootstrap sequence, dependency container, and lifecycle methods.

## Requirements

### Requirement: App struct

The system MUST define an `App` struct that holds references to the server, database, logger, and configuration.

#### Scenario: Access dependencies

- GIVEN a successfully bootstrapped `App`
- WHEN `app.DB`, `app.Server`, `app.Logger`, or `app.Config` are accessed
- THEN they are non-nil and initialized

### Requirement: Bootstrap sequence

The system MUST enforce the bootstrap order: load config → initialize logger → connect database → build container → setup router → start server.

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

### Requirement: Dependency container

The system MUST provide a `Container` type that wires repositories, services, and handlers, and is built during bootstrap.

#### Scenario: Container wiring

- GIVEN bootstrap succeeds
- WHEN `app.Container` is inspected
- THEN it contains initialized dependencies for all registered modules

### Requirement: Start and Stop lifecycle

The system MUST expose `Start()` to run the server and `Stop(ctx)` to release resources.

#### Scenario: Start server

- GIVEN a bootstrapped `App`
- WHEN `Start()` is called
- THEN the HTTP server begins listening

#### Scenario: Stop server

- GIVEN a running `App`
- WHEN `Stop(ctx)` is called
- THEN the HTTP server shuts down gracefully
- AND the database connection is closed

## Constraints and Edge Cases

- `Bootstrap()` MUST be idempotent within a single process; calling it twice without cleanup SHOULD return a new `App` each time (for tests).
- Any bootstrap failure MUST return an error with the step name (e.g., "db connect: ...").
- `Container` MUST NOT import business modules directly; modules are registered by the caller.
