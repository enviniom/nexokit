# Delta for Server Bootstrap

## ADDED Requirements

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
