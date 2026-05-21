# Delta for Health Checks

## MODIFIED Requirements

### Requirement: Readiness endpoint

The system MUST expose `GET /health/ready` that verifies connectivity to all active dependencies (database, cache if driver != "none") and returns HTTP 200 when all are healthy or HTTP 503 when any dependency is unhealthy.

(Previously: verified database and cache if driver != "none"; now also reports rate limiter status when active)

#### Scenario: All dependencies healthy

- GIVEN the database is connected and the cache driver is "none"
- WHEN `GET /health/ready` is requested
- THEN the response status is 200
- AND the body contains per-dependency status with all dependencies marked healthy

#### Scenario: Cache unreachable (when active)

- GIVEN the cache driver is not "none" and the cache connection has failed
- WHEN `GET /health/ready` is requested
- THEN the response status is 503
- AND the body identifies the cache dependency as unhealthy

#### Scenario: Redis cache healthy

- GIVEN the cache driver is "redis" and Redis responds to ping
- WHEN `GET /health/ready` is requested
- THEN the response status is 200
- AND the cache dependency is marked healthy
