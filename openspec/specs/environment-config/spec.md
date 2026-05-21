# Environment Config Specification

## Purpose
Define how NexoKit loads, validates, and exposes typed configuration from environment variables.

## Requirements

### Requirement: Load from .env file

The system MUST read environment variables from a `.env` file in the project root using `joho/godotenv`.

#### Scenario: Happy path — all variables present

- GIVEN a `.env` file with all required variables
- WHEN the application starts
- THEN the `Config` struct is fully populated with correct types

#### Scenario: Missing .env file

- GIVEN no `.env` file exists
- WHEN the application starts
- THEN it MUST fall back to actual environment variables
- AND fail only if required variables are still missing

### Requirement: Required variables

The system MUST expose typed fields for: `APP_NAME`, `APP_ENV`, `APP_PORT`, `APP_URL`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_SSL_MODE`, `CORS_ALLOWED_ORIGINS`, `SHUTDOWN_TIMEOUT_SECONDS`.

#### Scenario: Missing required DB_PASSWORD

- GIVEN `.env` lacks `DB_PASSWORD`
- WHEN bootstrap calls `LoadConfig()`
- THEN it MUST return an error immediately
- AND the application MUST exit before starting the server

### Requirement: Optional DATABASE_URL override

The system MAY accept `DATABASE_URL` as a single connection string. When `DATABASE_URL` is set, individual `DB_*` fields become optional.

#### Scenario: DATABASE_URL provided

- GIVEN `DATABASE_URL=postgres://user:pass@host/db`
- WHEN `LoadConfig()` executes
- THEN it MUST accept the URL and not require `DB_HOST`, `DB_PORT`, etc.

### Requirement: Default APP_ENV

The system SHOULD default `APP_ENV` to `development` when unset.

#### Scenario: Empty APP_ENV

- GIVEN `.env` does not define `APP_ENV`
- WHEN `LoadConfig()` executes
- THEN `Config.AppEnv` equals `development`

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

- `APP_PORT` MUST be parsed as `int`; invalid values MUST fail fast.
- `SHUTDOWN_TIMEOUT_SECONDS` MUST be parsed as `int`; invalid values MUST fail fast.
- Empty string for a required field MUST be treated as missing.
- `Config` MUST be immutable after load; no runtime mutation.
- `RATE_LIMIT_GLOBAL_RPM`, `RATE_LIMIT_LOGIN_RPM`, `RATE_LIMIT_REFRESH_RPM` MUST be positive integers; invalid values MUST fail fast.
- `REDIS_PORT` MUST be parsed as `int`; invalid values MUST fail fast.
- `RATE_LIMIT_WINDOW_SECONDS` MUST default to 60 when unset.
