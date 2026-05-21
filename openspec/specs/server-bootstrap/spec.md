# Server Bootstrap Specification

## Purpose
Define HTTP server initialization, router setup with `/api/v1/` versioning, health check, module registration convention, and graceful shutdown.

## Requirements

### Requirement: Gin server initialization

The system MUST create a Gin engine, bind to `APP_PORT`, and start an HTTP server.

#### Scenario: Start on configured port

- GIVEN `APP_PORT=8080`
- WHEN the server starts
- THEN it listens on `:8080`

### Requirement: API versioning prefix

The system MUST mount all module routes under `/api/v1/`.

#### Scenario: Module route registration

- GIVEN a module registers `/users` on the `v1` group
- WHEN a request is made to `/api/v1/users`
- THEN it reaches the module handler

### Requirement: Health check endpoint

The system MUST expose `GET /health` returning HTTP 200 with the standard envelope and `data.status = "ok"`. Additionally, the system MUST expose `GET /health/live` for liveness probes (returns 200 with `{"status": "alive"}`) and `GET /health/ready` for readiness probes (returns 200 or 503 based on dependency health).

#### Scenario: Healthy API

- GIVEN the server is running
- WHEN `GET /health` is requested
- THEN the response is 200
- AND the body is `{ success: true, message: "API is healthy", data: { status: "ok" }, meta: null, errors: null }`

#### Scenario: Liveness probe succeeds

- GIVEN the server process is running
- WHEN `GET /health/live` is requested
- THEN the response is 200
- AND the body contains `{"status": "alive"}`

#### Scenario: Readiness probe with healthy dependencies

- GIVEN the database is connected and no active cache driver
- WHEN `GET /health/ready` is requested
- THEN the response is 200
- AND the body reports all dependencies as healthy

#### Scenario: Readiness probe with unhealthy dependency

- GIVEN the database connection has failed
- WHEN `GET /health/ready` is requested
- THEN the response is 503
- AND the body identifies the database as unhealthy

### Requirement: Module registration convention

The system MUST support a convention where each module exposes `Register(v1 *gin.RouterGroup, deps ...)`.

#### Scenario: Register users module

- GIVEN `internal/modules/users` has `Register(v1, handler)`
- WHEN `server/router.go` calls `users.Register(v1, container.Users.Handler)`
- THEN the routes are mounted under `/api/v1/users`

### Requirement: Graceful shutdown

The system MUST shutdown gracefully on `SIGTERM` or `SIGINT`, waiting up to `SHUTDOWN_TIMEOUT_SECONDS` for active connections.

#### Scenario: SIGTERM received

- GIVEN the server is handling requests
- WHEN `SIGTERM` is sent
- THEN the server stops accepting new connections
- AND waits for in-flight requests up to the timeout
- AND exits cleanly

#### Scenario: Shutdown timeout exceeded

- GIVEN a long-running request exceeds `SHUTDOWN_TIMEOUT_SECONDS`
- WHEN shutdown initiates
- THEN the server force-closes remaining connections after timeout

### Requirement: Health route registration order

The system MUST register `/health`, `/health/live`, and `/health/ready` routes on the base Gin engine BEFORE the `/api/v1/` router group and its auth middleware are applied.

#### Scenario: Health routes bypass auth

- GIVEN auth middleware is applied to the `/api/v1/` group
- WHEN `GET /health/live` is requested without credentials
- THEN the response is 200 (not 401)

#### Scenario: Health routes bypass CORS validation

- GIVEN CORS middleware validates origins on the `/api/v1/` group
- WHEN `GET /health/ready` is requested from any origin
- THEN the response is not blocked by CORS policy

### Requirement: Rate limit middleware wiring

The system MUST wire rate limit middleware to sensitive auth endpoints (`/api/v1/auth/login`, `/api/v1/auth/refresh`) during router setup when `RATE_LIMIT_ENABLED=true`.

#### Scenario: Rate limit applied to login

- GIVEN `RATE_LIMIT_ENABLED=true`
- WHEN the router is built
- THEN the login route has rate limit middleware applied

#### Scenario: Rate limit not applied when disabled

- GIVEN `RATE_LIMIT_ENABLED=false`
- WHEN the router is built
- THEN no rate limit middleware is applied to any route

#### Scenario: Rate limiter initialized from config driver

- GIVEN `RATE_LIMIT_DRIVER=redis` and Redis is reachable
- WHEN the router is built
- THEN a Redis-backed rate limiter is used

#### Scenario: Rate limiter falls back to in-memory

- GIVEN `RATE_LIMIT_DRIVER=redis` but Redis is unreachable
- WHEN the router is built
- THEN an in-memory rate limiter is used with a warning logged

## Constraints and Edge Cases

- `GET /health` MUST NOT require authentication or CORS origin validation.
- The router MUST reject requests to undefined paths with 404 and the standard envelope.
- Gin mode (`debug` vs `release`) MUST align with `APP_ENV`.
