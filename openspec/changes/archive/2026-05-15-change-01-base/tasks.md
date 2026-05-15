# Tasks: NexoKit Foundation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~2,000 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | Single size:exception PR |
| Delivery strategy | single-pr |
| Chain strategy | size-exception |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: size-exception
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Platform leaves + tests | PR 1 | response, apperror, validator, shared, stubs |
| 2 | Infrastructure + tests | PR 1 | config, logger, db, cache |
| 3 | HTTP layer + tests | PR 1 | middleware, server, router, health |
| 4 | App orchestration + entrypoints | PR 1 | app, container, cmd, modules, cli |
| 5 | Dev environment + docs | PR 1 | docker, Makefile, README, test helpers |

## Phase 1: Platform Leaves

- [ ] 1.1 `internal/shared/model.go` — BaseModel, BaseModelSimple. Deps: none. AC: compiles; fields match design.
- [ ] 1.2 `internal/platform/response/` — response.go + tests. Deps: none. AC: envelope helpers compile; table tests pass.
- [ ] 1.3 `internal/platform/apperror/` — apperror.go + tests. Deps: none. AC: sentinels, Wrap, Status, PublicMessage work.
- [ ] 1.4 `internal/platform/validator/` — validator.go, gin.go + tests. Deps: 1.2. AC: composable rules, rune counting, 422 helper pass.
- [ ] 1.5 `internal/platform/{query,identity,password,token}.go` — stubs with TODO. Deps: none. AC: compile.

## Phase 2: Infrastructure

- [ ] 2.1 `internal/config/` — config.go, env.go + tests. Deps: none. AC: typed structs, fail-fast validation, table tests pass.
- [ ] 2.2 `internal/infra/logger/` — logger.go. Deps: 2.1. AC: returns JSON slog.Logger.
- [ ] 2.3 `internal/infra/db/` — postgres.go, migrations.go. Deps: 2.1. AC: GORM connect, pool, Goose stubs.
- [ ] 2.4 `internal/infra/cache/` — cache.go (interface), noop.go, redis.go stub. Deps: none. AC: NoopCache satisfies Cache; compile.

## Phase 3: HTTP Layer

- [ ] 3.1 `internal/middleware/` — request_id.go, logger.go, recovery.go, cors.go + tests. Deps: 2.1, 2.2, 1.2. AC: chain order correct; request_id, recovery, CORS scenarios pass.
- [ ] 3.2 `internal/server/router.go` — gin.Engine, health check, v1 group. Deps: 1.2, 3.1. AC: GET /health returns 200 JSON.
- [ ] 3.3 `internal/server/server.go` + tests — http.Server wrapper. Deps: 3.2. AC: Start/Stop; health and 404 envelope tests pass.

## Phase 4: App Orchestration

- [ ] 4.1 `internal/app/app.go` — App struct, Start, Stop. Deps: 2.1, 2.2, 2.3, 2.4. AC: Stop closes DB, cache, server.
- [ ] 4.2 `internal/app/bootstrap.go` — init order. Deps: 4.1, 2.4. AC: returns wired *App.
- [ ] 4.3 `internal/app/container.go` — Container, RegisterModules. Deps: 4.1. AC: empty container compiles.
- [ ] 4.4 `cmd/api/main.go` — signal handler, Bootstrap, Start, Stop. Deps: 4.2. AC: `go build` succeeds.
- [ ] 4.5 `cmd/nexokit/main.go`, `internal/cli/cli.go` — CLI stub. Deps: none. AC: compile.
- [ ] 4.6 `internal/modules/{auth,users,roles,companies}/module.go` — stubs with Register. Deps: 4.3. AC: compile.

## Phase 5: Dev Environment

- [ ] 5.1 `docker-compose.yml` — Postgres service. Deps: none. AC: `docker compose up -d` boots Postgres.
- [ ] 5.2 `.env.example` — all required vars. Deps: 2.1. AC: matches Load() requirements.
- [ ] 5.3 `Makefile` — migrate-up/down/create/status, build, run. Deps: 2.3. AC: targets run without error.
- [ ] 5.4 `README.md` — setup, env, run, migrate. Deps: 5.3. AC: new dev can bootstrap from README.
- [ ] 5.5 `tests/helpers/app.go` — NewTestApp(). Deps: 4.2. AC: returns bootstrapped test App with cleanup.
