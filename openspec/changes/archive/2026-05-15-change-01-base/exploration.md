# Exploration: change-01-base — NexoKit Foundation

## Current State

The project is essentially empty: only `go.mod` (module `github.com/enviniom/nexokit`, Go 1.26.3) and a `docs/` folder with change specifications exist. No source code, no tests, no infrastructure. This change builds the entire foundational framework from zero.

---

## Full Scope Analysis

This change is **large and foundational**. It creates ~15+ new packages and touches every layer of the application:

### Core Infrastructure (must exist for anything else to work)
1. `internal/config/` — Typed environment config with `.env` support
2. `internal/infra/db/` — PostgreSQL connection + GORM setup
3. `internal/infra/logger/` — Structured logging with slog + lumberjack
4. `internal/infra/cache/` — Cache interface + Redis + Noop implementations
5. `internal/shared/model.go` — BaseModel, BaseModelSimple

### Platform Capabilities (leaf packages, no internal deps)
6. `internal/platform/response/` — Standard API response envelope, pagination, helpers
7. `internal/platform/apperror/` — Typed application errors (ErrNotFound, ErrForbidden, etc.)
8. `internal/platform/validator/` — Custom composable validator (Rule, FieldValidator, gin helper)
9. `internal/platform/query/` — Query param parsing (pagination, filters, sorting) — *structure only*
10. `internal/platform/identity/` — PublicID generation — *structure only*
11. `internal/platform/password/` — argon2id hashing — *structure only*
12. `internal/platform/token/` — PASETO token handling — *structure only*

### HTTP Layer
13. `internal/middleware/` — CORS, Request ID, Recovery, Logger middleware
14. `internal/server/` — HTTP server setup, router with `/api/v1/` prefix, route mounting convention

### Application Orchestration
15. `internal/app/app.go` — App type aggregating server, DB, logger
16. `internal/app/bootstrap.go` — Boot sequence: config → DB → logger → router → server
17. `internal/app/container.go` — Dependency graph (empty initially but must exist)

### Entry Points
18. `cmd/api/main.go` — Thin API entry point
19. `cmd/nexokit/main.go` — Thin CLI entry point (stub for change-02)

### DevOps & Tooling
20. `docker-compose.yml` — PostgreSQL for local dev
21. `.env.example` — Documented environment variables
22. `Makefile` — Build, test, migration commands
23. `migrations/` — Goose migration files (initial schema)
24. `README.md` — Getting started guide
25. `tests/helpers/`, `tests/fixtures/` — Test infrastructure scaffolding

---

## Critical Design Decisions

### 1. Package Structure (already decided in _context.md)
The folder layout is well-defined and designed to prevent circular imports:
- `platform/` → leaf packages (no internal imports)
- `infra/` → depends on `config/` only
- `middleware/` → depends on `platform/` + `infra/logger/`
- `server/` → depends on `middleware/` + `platform/` + `gin`
- `app/` → depends on everything above; orchestrates bootstrap
- `cmd/` → depends only on `app/`

**Decision**: Follow the structure exactly. No deviations.

### 2. Error Handling (`platform/apperror`)
Typed errors map to HTTP status codes. Each error carries a message and optionally a cause. This avoids scattered `c.Status()` calls in handlers.

Key types needed:
- `AppError` struct with `Code`, `Message`, `Cause`
- Sentinel errors: `ErrNotFound`, `ErrForbidden`, `ErrUnauthorized`, `ErrValidation`, `ErrInternal`
- Conversion helper: `ToHTTPStatus(err error) int`

### 3. Response Format (`platform/response`)
Standard JSON envelope. All handlers must use this. Never `gin.H`.

Key functions needed:
- `OK(c, data, message)`
- `Created(c, data, message)`
- `Paginated(c, data, pagination)`
- `Error(c, err)` — converts `apperror` to envelope
- `ValidationError(c, errs)` — uses `validator.ValidationErrors`
- `BadRequest(c, message)`
- `NotFound(c, message)`

### 4. Validator (`platform/validator`)
Custom composable rules, no reflection, no struct tags. Spec is complete with code samples.

Key insight: `validator/gin.go` has a **dependency cycle risk** with `platform/response`. The spec shows `RespondIfInvalid` using inline `gin.H`, with a note: *"Cuando `platform/response` esté implementado en el Change 1, este helper debe actualizarse"*.

**Resolution**: Implement `platform/response` BEFORE `validator/gin.go`. Then `gin.go` can import `platform/response` cleanly. `validator.go` and `rules.go` have zero dependencies and can be built first.

### 5. DB Initialization (`infra/db/`)
Two concerns in one package:
- `postgres.go` — Connection pool setup using GORM
- `migrations.go` — Goose wrapper for running SQL migrations

Goose operates on raw `*sql.DB` (or `*pgx.Conn` / `*sql.Conn`), not GORM. GORM gives us `db.DB()` to get the underlying `*sql.DB`. The migration helper should use that.

**Decision**: Connect with GORM first, then extract `*sql.DB` for Goose. This keeps a single connection pool.

### 6. Configuration (`internal/config/`)
Typed struct, loaded from env via a library like `joho/godotenv` or manual `os.Getenv`. The spec doesn't mandate a specific env library.

**Decision**: Use `joho/godotenv` for `.env` file support + manual `os.Getenv` mapping to typed struct. This avoids heavy config frameworks and keeps it predictable. No reflection-based mappers.

---

## Dependency Graph & Implementation Order

To avoid circular imports and keep the project buildable at every step, the implementation order should be:

```
Phase 1 — Leaves (buildable, testable independently)
  ├── go.mod dependencies (gin, gorm, goose, godotenv, lumberjack, etc.)
  ├── internal/shared/model.go
  ├── internal/config/
  ├── internal/platform/response/
  ├── internal/platform/apperror/
  ├── internal/platform/validator/ (validator.go + rules.go)
  └── internal/platform/query/ (stubs only)

Phase 2 — Infrastructure (depends on config)
  ├── internal/infra/db/postgres.go
  ├── internal/infra/db/migrations.go
  ├── internal/infra/logger/
  └── internal/infra/cache/ (interface + noop + redis stub)

Phase 3 — Platform glue (depends on response + validator)
  ├── internal/platform/validator/gin.go (can now import response)
  └── internal/platform/identity/ (stub)
  └── internal/platform/password/ (stub)
  └── internal/platform/token/ (stub)

Phase 4 — HTTP layer (depends on platform + infra)
  ├── internal/middleware/ (request_id, recovery, cors, logger)
  └── internal/server/ (server.go + router.go)

Phase 5 — Application orchestration (depends on everything)
  ├── internal/app/app.go
  ├── internal/app/bootstrap.go
  └── internal/app/container.go

Phase 6 — Entry points (depends on app)
  ├── cmd/api/main.go
  └── cmd/nexokit/main.go (stub)

Phase 7 — DevOps & documentation
  ├── docker-compose.yml
  ├── .env.example
  ├── Makefile
  ├── migrations/ (initial goose migration file)
  ├── tests/ scaffolding
  └── README.md
```

### Why this order matters
- `platform/` packages are the foundation. If `response` isn't done first, every middleware and handler that wants to return JSON will be blocked.
- `validator/gin.go` needs `response`, so it comes after Phase 1.
- `infra/` needs `config` for connection strings.
- `middleware/logger.go` needs `infra/logger`.
- `server/` needs `middleware/` to mount them.
- `app/` is the capstone: it wires everything together.
- `cmd/api/main.go` is trivial once `app/` compiles.

---

## Risks

### Risk 1: Change is oversized for a single PR
This change likely produces 800–1200+ lines of new code across 25+ files. The 400-line review budget will be exceeded.

**Mitigation**: During `sdd-tasks`, split into chained PRs or work units. Suggested slices:
- **Slice A**: Platform foundation (`config`, `response`, `apperror`, `validator` core) + `shared/model`
- **Slice B**: Infrastructure (`infra/db`, `infra/logger`, `infra/cache`, migrations)
- **Slice C**: HTTP layer + App bootstrap (`middleware`, `server`, `app`, `cmd/api`)
- **Slice D**: DevOps (`docker-compose`, `Makefile`, `README`, `.env.example`, test scaffolding)

### Risk 2: Validator `gin.go` vs `response` dependency
The spec acknowledges this. If `gin.go` is implemented before `response`, it uses inline `gin.H` and needs a second pass.

**Mitigation**: Implement `response` first, then update `gin.go` immediately. No need for a second pass.

### Risk 3: Goose vs GORM migration strategy
Goose runs raw SQL. GORM runs struct-based auto-migrate. The project mandates Goose. If someone later runs `db.AutoMigrate()`, it will conflict with Goose-managed schema.

**Mitigation**: Document clearly in `README.md` and code comments: **Never use GORM AutoMigrate in production. All schema changes go through Goose migrations.**

### Risk 4: Cache interface premature abstraction
The cache package has Redis + Noop implementations. For change-01, there's no actual caching use case (no modules exist yet). Building a full Redis client is speculative.

**Mitigation**: Implement the `Cache` interface and `NoopCache` fully. Create `RedisCache` as a stub that compiles but isn't wired in `container.go` yet.

### Risk 5: `cmd/nexokit` stub may confuse
The CLI is change-02. A stub `cmd/nexokit/main.go` might create the impression that the CLI is ready.

**Mitigation**: Add a clear `// TODO: change-02` comment and a `fmt.Println("NexoKit CLI — not yet implemented")` so it's obviously a placeholder.

---

## Ambiguities Requiring Resolution

### Ambiguity 1: Platform packages that are "future stubs"
The folder structure lists `platform/query`, `platform/identity`, `platform/password`, `platform/token`. The change scope doesn't explicitly say to implement them. Should they:
- (a) Be empty `.go` files with package declarations?
- (b) Be fully implemented now even though nothing uses them?
- (c) Be created as directories with `.gitkeep`?

**Recommendation**: (a) — empty package files with clear `// TODO: change-0X` comments. This prevents "package not found" errors when later changes import them, and keeps the directory structure intact.

### Ambiguity 2: `internal/modules/` in change-01
The spec says "Estructura para módulos" (structure for modules) in the scope. Does this mean create empty module directories, or create a sample module?

**Recommendation**: Create empty module directories only. No sample module. The first real module comes in change-03 (auth/users). Creating a sample module would add dead code.

### Ambiguity 3: Migration commands via Makefile vs CLI
The spec lists under "Migraciones": "NexoKit debe incluir comandos para: crear migración, ejecutar migraciones, revertir migraciones, consultar estado". It also lists under acceptance criteria: "Existen scripts o comandos Makefile para migraciones."

Does this mean:
- (a) Makefile targets only (e.g., `make migrate-up`, `make migrate-down`)?
- (b) Both Makefile targets AND CLI subcommands in `cmd/nexokit`?

**Recommendation**: (a) for change-01. The CLI is change-02. Makefile targets wrapping `goose` binary or `go run` are sufficient. The CLI will add its own migration commands later.

### Ambiguity 4: `SHUTDOWN_TIMEOUT_SECONDS` but no graceful shutdown spec
The env var is listed, but the acceptance criteria don't explicitly mention graceful shutdown. Should `app.go` implement `Server.Shutdown(ctx)` with timeout?

**Recommendation**: Yes, implement basic graceful shutdown in `app.go`. It's ~10 lines with `signal.NotifyContext` and `http.Server.Shutdown`, and it's standard for production APIs. Mark it clearly as "basic graceful shutdown".

### Ambiguity 5: `tests/` directory scope
The spec shows `tests/integration/`, `tests/helpers/`, `tests/fixtures/`. Should change-01 create actual test helper code, or just empty directories?

**Recommendation**: Create `tests/helpers/` with a minimal `NewTestApp()` helper and `tests/fixtures/` with an empty file. Don't write integration tests yet — there are no modules to integrate. But having the helper ready prevents rework in change-03.

### Ambiguity 6: `platform/query` scope
The spec says "Parser y normalización de query params reutilizables: paginación, filtros, ordenamiento y búsqueda". For change-01, there are no endpoints that use pagination except possibly `/health` (which doesn't paginate). Should `query` be fully implemented?

**Recommendation**: Implement only pagination structs (`Page`, `PerPage`, `Offset`, `Limit`, `PaginationMeta`) and helpers. Filters/sorting/search can be stubs with TODO comments. Pagination is needed for the response package anyway.

---

## Approaches

### Approach A: Big Bang — implement everything in one branch
- **Pros**: Single coherent review, no intermediate states, fastest to "it works".
- **Cons**: Massive PR (1000+ lines), high review load, harder to debug if build fails.
- **Effort**: High

### Approach B: Phased Slices — implement in dependency order, merge incrementally
- **Pros**: Reviewable chunks, each slice compiles and passes tests, easier rollback.
- **Cons**: More branch management, temporary scaffolding needed between slices.
- **Effort**: Medium

### Recommendation
**Approach B (Phased Slices)**. Given the 400-line review budget, this is the only responsible path. The dependency graph above naturally suggests 4 slices that can each be a PR.

---

## Ready for Proposal

**Yes**, with the following clarifications passed to the user/orchestrator:

1. **Confirm stub policy**: Should future-only platform packages (`query` filters, `identity`, `password`, `token`) be empty package files with TODO comments?
2. **Confirm test scaffolding depth**: Should `tests/helpers/` include a working `NewTestApp()` helper, or just empty directories?
3. **Confirm Makefile-only for migrations in change-01**: The CLI (change-02) will add its own migration commands. Is Makefile sufficient for now?
4. **Delivery strategy**: This change exceeds the 400-line budget. Shall we proceed with chained PRs (slices A→D) or accept an exception?

If the user confirms these four points, exploration is complete and the change is ready for the **propose** phase.
