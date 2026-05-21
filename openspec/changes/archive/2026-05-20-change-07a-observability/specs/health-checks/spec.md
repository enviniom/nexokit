# Health Checks Specification

## Purpose

Define liveness and readiness health check endpoints for container orchestration, load balancer probes, and dependency status aggregation.

## Requirements

### Requirement: Liveness endpoint

The system MUST expose `GET /health/live` returning HTTP 200 with `{"status": "alive"}`. This endpoint MUST NOT depend on any external service (DB, cache, etc.).

#### Scenario: Liveness returns alive

- GIVEN the server process is running
- WHEN `GET /health/live` is requested
- THEN the response status is 200
- AND the body contains `{"status": "alive"}`

#### Scenario: Liveness ignores dependency failures

- GIVEN the database is unreachable
- WHEN `GET /health/live` is requested
- THEN the response status is still 200
- AND the body contains `{"status": "alive"}`

### Requirement: Readiness endpoint

The system MUST expose `GET /health/ready` that verifies connectivity to all active dependencies (database, cache if driver != "none") and returns HTTP 200 when all are healthy or HTTP 503 when any dependency is unhealthy.

#### Scenario: All dependencies healthy

- GIVEN the database is connected and the cache driver is "none"
- WHEN `GET /health/ready` is requested
- THEN the response status is 200
- AND the body contains per-dependency status with all dependencies marked healthy

#### Scenario: Database unreachable

- GIVEN the database connection has failed or is closed
- WHEN `GET /health/ready` is requested
- THEN the response status is 503
- AND the body identifies the database dependency as unhealthy

#### Scenario: Cache unreachable (when active)

- GIVEN the cache driver is not "none" and the cache connection has failed
- WHEN `GET /health/ready` is requested
- THEN the response status is 503
- AND the body identifies the cache dependency as unhealthy

#### Scenario: Multiple dependencies unhealthy

- GIVEN both the database and cache are unreachable
- WHEN `GET /health/ready` is requested
- THEN the response status is 503
- AND the body reports both dependencies as unhealthy individually

### Requirement: Per-dependency status reporting

The `/health/ready` response MUST include a structured JSON object with each dependency's name and status (`"healthy"` or `"unhealthy"`), and optionally an error message for unhealthy dependencies.

#### Scenario: Ready response structure

- GIVEN all dependencies are healthy
- WHEN `GET /health/ready` succeeds
- THEN the response body is a JSON object with a `dependencies` field
- AND each dependency entry has `name`, `status`, and optionally `error` fields

#### Scenario: Unhealthy dependency includes error

- GIVEN the database ping fails with error "connection refused"
- WHEN `GET /health/ready` is requested
- THEN the database dependency entry has `status: "unhealthy"` and `error` containing the failure reason

### Requirement: Health endpoints are unauthenticated

The system MUST NOT require authentication, CORS origin validation, or any middleware other than RequestID and Logger for `/health`, `/health/live`, and `/health/ready` endpoints.

#### Scenario: No auth header required

- GIVEN no `Authorization` header is sent
- WHEN `GET /health/live` is requested
- THEN the response is 200 (not 401)

#### Scenario: No CORS validation

- GIVEN a request from any origin
- WHEN `GET /health/ready` is requested
- THEN the response is not blocked by CORS policy

### Requirement: Existing /health endpoint preserved

The system MUST maintain the existing `GET /health` endpoint behavior unchanged, returning HTTP 200 with the standard envelope and `data.status = "ok"`.

#### Scenario: Existing health endpoint unchanged

- GIVEN the server is running
- WHEN `GET /health` is requested
- THEN the response is 200
- AND the body matches `{ success: true, message: "API is healthy", data: { status: "ok" }, meta: null, errors: null }`

## Constraints and Edge Cases

- `/health/live` MUST complete without any I/O — it only confirms the process is alive.
- `/health/ready` MUST use a configurable timeout for dependency pings (default 5 seconds) to avoid blocking readiness probes indefinitely.
- The readiness response MUST NOT expose sensitive connection details (hostnames, credentials) in error messages.
- When cache driver is "none", the cache dependency MUST be reported as `"healthy"` without attempting a ping.
