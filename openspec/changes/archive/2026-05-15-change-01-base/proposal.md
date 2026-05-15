# Proposal: Base project, config, GORM, migrations, standard API response

## Intent

Build the foundational NexoKit framework from zero: typed config, PostgreSQL/GORM connection, Goose migrations, standard API response envelope, typed errors, custom validator, HTTP middleware, server bootstrap, and local dev environment. This is the base upon which all future changes depend.

## Scope

### In Scope
- Project structure (`cmd/`, `internal/`, `tests/`, `migrations/`)
- Config loading from `.env` with environment-specific behavior
- PostgreSQL connection pool via GORM
- Goose migration system (Makefile targets)
- Standard API response (`APIResponse`, pagination, helpers)
- Typed application errors (`AppError`, sentinel errors)
- Custom composable validator (Rule, FieldValidator, Gin helper)
- HTTP middleware: CORS, Request ID, Recovery, Logger
- Server bootstrap with `/api/v1/` prefix and route mounting convention
- `internal/app/` (App, Bootstrap, Container)
- `docker-compose.yml`, `.env.example`, Makefile, README
- Empty platform stubs (`query`, `identity`, `password`, `token`) with TODO comments
- Minimal `tests/helpers/NewTestApp()`

### Out of Scope
- CLI migration commands (change-02)
- Business modules (change-03+)
- Authentication/authorization implementation
- Integration tests (no modules to test)
- Cache Redis wiring (interface + Noop only)

## Capabilities

### New Capabilities
- `environment-config`: Typed struct loaded from `.env` with `joho/godotenv`
- `database-connection`: GORM PostgreSQL setup with single connection pool
- `migrations`: Goose SQL migrations via Makefile targets (`migrate-up`, `migrate-down`, `migrate-create`, `migrate-status`)
- `api-response`: Standard JSON envelope, pagination metadata, helper functions
- `error-handling`: Typed `AppError` with HTTP status mapping
- `request-validation`: Composable Rule/FieldValidator without reflection, Gin integration
- `http-middleware`: CORS, Request ID, Recovery, Logger middleware stack
- `server-bootstrap`: Gin server, `/api/v1/` router, module route mounting convention
- `app-orchestration`: App type, bootstrap sequence, dependency container

### Modified Capabilities
None

## Approach

Bottom-up implementation following dependency graph: platform leaves (`response` → `apperror` → `validator` core) → infra (`config` → `db` → `logger` → `cache`) → platform glue (`validator/gin.go` + stubs) → HTTP layer (`middleware` → `server`) → app orchestration → entry points → DevOps.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/api/` | New | Thin API entry point |
| `cmd/nexokit/` | New | CLI stub with TODO |
| `internal/app/` | New | App, bootstrap, container |
| `internal/config/` | New | Typed env config |
| `internal/infra/db/` | New | Postgres + Goose wrapper |
| `internal/infra/logger/` | New | slog + lumberjack |
| `internal/infra/cache/` | New | Cache interface + Noop |
| `internal/server/` | New | HTTP server + router |
| `internal/middleware/` | New | CORS, request ID, recovery, logger |
| `internal/platform/*` | New | Response, apperror, validator, stubs |
| `internal/shared/` | New | BaseModel, BaseModelSimple |
| `tests/helpers/` | New | Minimal NewTestApp() |
| `migrations/` | New | Initial Goose migration |
| `docker-compose.yml` | New | PostgreSQL for local dev |
| `Makefile` | New | Build, test, migration targets |
| `README.md` | New | Getting started guide |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Oversized PR (>400 lines) | High | Delivery exception accepted for foundational commit; single PR |
| Dependency order violations | Med | Strict bottom-up implementation order |
| GORM AutoMigrate conflict | Low | Document: never use AutoMigrate; Goose only |

## Rollback Plan

1. `git revert` the merge commit
2. `docker-compose down -v` if DB schema changes caused issues
3. Re-create `openspec/changes/change-01-base/` if needed

## Dependencies

None — this is the first change.

## Success Criteria

- [ ] Project compiles (`go build ./...`)
- [ ] API starts from `cmd/api/main.go`
- [ ] Config loads from environment variables
- [ ] `.env.example` exists and is documented
- [ ] `docker-compose.yml` runs PostgreSQL
- [ ] PostgreSQL connection works
- [ ] GORM is configured
- [ ] Goose migrations run via Makefile
- [ ] `/health` returns standard response
- [ ] All responses use standard envelope
- [ ] Errors are consistent
- [ ] CORS is configurable
- [ ] Request ID middleware active
- [ ] Recovery middleware active
- [ ] Routes use `/api/v1/` prefix
- [ ] Route registration convention documented
- [ ] `internal/app/container.go` exists
- [ ] Modular structure is in place
- [ ] README has run instructions
- [ ] Makefile has migration targets
