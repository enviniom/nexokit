# Delta for Server Bootstrap

## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Health check endpoint

The system MUST expose `GET /health` returning HTTP 200 with the standard envelope and `data.status = "ok"`. Additionally, the system MUST expose `GET /health/live` for liveness probes (returns 200 with `{"status": "alive"}`) and `GET /health/ready` for readiness probes (returns 200 or 503 based on dependency health).

(Previously: Only `GET /health` existed, returning 200 with standard envelope)

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
