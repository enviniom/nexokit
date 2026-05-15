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

The system MUST expose `GET /health` returning HTTP 200 with the standard envelope and `data.status = "ok"`.

#### Scenario: Healthy API

- GIVEN the server is running
- WHEN `GET /health` is requested
- THEN the response is 200
- AND the body is `{ success: true, message: "API is healthy", data: { status: "ok" }, meta: null, errors: null }`

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

## Constraints and Edge Cases

- `GET /health` MUST NOT require authentication or CORS origin validation.
- The router MUST reject requests to undefined paths with 404 and the standard envelope.
- Gin mode (`debug` vs `release`) MUST align with `APP_ENV`.
