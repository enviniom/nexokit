# Spec: dev-tooling

> Source of truth for developer tooling requirements.
> Merged from change-02-cli (2026-05-16).

## Requirements

| # | Requirement | Strength |
|---|-------------|----------|
| 1 | The Makefile MUST provide a `dev` target that runs the API in development mode. | MUST |
| 2 | The Makefile MUST provide `build`, `test`, `migrate-up`, `migrate-down`, `migrate-create`, `seed`, `create-root`, `lint`, `fmt` targets. | MUST |
| 3 | The Makefile SHOULD load environment variables from `.env` where applicable. | SHOULD |
| 4 | The Makefile targets MUST fail with a clear error message if required variables are missing. | MUST |

## Scenarios

### Scenario: Developer runs API in dev mode

- GIVEN the project is cloned and `.env` is present
- WHEN the developer runs `make dev`
- THEN the API starts on the configured port

### Scenario: Missing database URL during migration

- GIVEN `DATABASE_URL` is not set in `.env` or environment
- WHEN the developer runs `make migrate-up`
- THEN the command fails immediately with a clear error indicating the missing variable
