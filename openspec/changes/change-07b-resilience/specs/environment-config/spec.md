# Delta for Environment Config

## ADDED Requirements

### Requirement: Redis configuration fields

The system MUST expose typed fields for Redis connection: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_DIAL_TIMEOUT_SECONDS`.

#### Scenario: Redis config populated from env

- GIVEN `.env` contains `REDIS_HOST=localhost`, `REDIS_PORT=6379`, `REDIS_DB=0`
- WHEN `LoadConfig()` executes
- THEN `Config.Redis` is populated with those values

#### Scenario: Redis password is optional

- GIVEN `REDIS_PASSWORD` is not set
- WHEN `LoadConfig()` executes
- THEN `Config.Redis.Password` is empty string (no error)

### Requirement: Rate limit configuration fields

The system MUST expose typed fields for rate limiting: `RATE_LIMIT_ENABLED` (bool), `RATE_LIMIT_DRIVER` (string: `memory` or `redis`), `RATE_LIMIT_GLOBAL_RPM` (int), `RATE_LIMIT_LOGIN_RPM` (int), `RATE_LIMIT_REFRESH_RPM` (int), `RATE_LIMIT_WINDOW_SECONDS` (int), `RATE_LIMIT_CLEANUP_INTERVAL_MINUTES` (int).

#### Scenario: Rate limit config populated from env

- GIVEN `.env` contains `RATE_LIMIT_ENABLED=true`, `RATE_LIMIT_GLOBAL_RPM=100`, `RATE_LIMIT_LOGIN_RPM=10`
- WHEN `LoadConfig()` executes
- THEN `Config.RateLimit` is populated with those values

#### Scenario: Rate limit defaults when unset

- GIVEN no `RATE_LIMIT_*` variables are set
- WHEN `LoadConfig()` executes
- THEN `Config.RateLimit.Enabled` is `false` and RPM fields use sensible defaults

### Requirement: CACHE_DRIVER configuration

The system MUST support `CACHE_DRIVER` with values `"redis"` or `"none"` (default: `"none"`).

#### Scenario: Default cache driver is none

- GIVEN `CACHE_DRIVER` is not set
- WHEN `LoadConfig()` executes
- THEN `Config.Cache.Driver` equals `"none"`

## Constraints and Edge Cases

- `RATE_LIMIT_GLOBAL_RPM`, `RATE_LIMIT_LOGIN_RPM`, `RATE_LIMIT_REFRESH_RPM` MUST be positive integers; invalid values MUST fail fast.
- `REDIS_PORT` MUST be parsed as `int`; invalid values MUST fail fast.
- `RATE_LIMIT_WINDOW_SECONDS` MUST default to 60 when unset.
