# Design: change-01-base — NexoKit foundation

## Technical Approach

Build the project bottom-up from leaf packages to orchestration so every layer compiles without circular imports. The 11 spec domains map to four layers: typed config/shared/platform leaves, infrastructure adapters, HTTP server/middleware, and app bootstrap/dev tooling.

## Package Dependency Graph

```txt
cmd/api ─┐
cmd/nexokit → internal/cli
         └→ app → config, infra/{db,logger,cache}, server, *gorm.DB, slog
server → middleware, platform/response, gin, net/http
middleware → config, platform/response, slog, gin
infra/db → config, gorm/postgres, goose via *sql.DB
infra/logger → config, slog, lumberjack
infra/cache → context only (Redis file is stubbed, not wired)
platform/validator/gin → platform/response, gin
platform/{response,apperror,validator core,query,identity,password,token}, shared → leaf packages
modules/* → stubs only; future modules may import platform/shared, not app/infra/server
```

**Rationale**: `app` is the composition root. Modules never import `infra` or each other, preventing cycles and preserving testable boundaries.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|---|---|---|---|
| Config | `config.Config` with nested typed structs | global env reads, string map | Fail-fast validation and no loose strings. |
| DB schema | GORM for runtime queries; Goose SQL migrations only | `AutoMigrate` | Goose gives auditable SQL; GORM remains query layer. No `AutoMigrate` ever. |
| Errors | sentinel-backed `AppError` | ad-hoc HTTP errors | `errors.Is` works and HTTP mapping is centralized. |
| Responses | `platform/response` helpers only | `gin.H` | One envelope and null semantics everywhere. |
| Validator | composable functions, no tags/reflection | go-playground/validator | Predictable errors and field-level composition. |
| Cache | `Cache` interface + `NoopCache`; Redis stub | wire Redis now | Change-01 has no cache use case; avoid speculative runtime dependency. |
| Modules | stubs + registration convention | sample module | Avoid dead business logic before auth/users changes. |

## Data Flow

```txt
main → app.Bootstrap → LoadConfig → logger.New → db.Connect
     → app.NewContainer(empty) → server.NewRouter → server.New
request → RequestID → Logger → Recovery → CORS → route → response helpers
shutdown signal → app.Stop(ctx) → http.Server.Shutdown → sqlDB.Close
```

## Interfaces / Contracts

```go
type Config struct{ App AppConfig; DB DBConfig; CORS CORSConfig; Log LogConfig; Shutdown ShutdownConfig; Cache CacheConfig }
type AppConfig struct{ Name, Env, URL string; Port int }
type DBConfig struct{ Host, Name, User, Password, SSLMode, DatabaseURL string; Port, MaxOpenConns, MaxIdleConns int; ConnMaxLifetime time.Duration }

type App struct{ Config *config.Config; Logger *slog.Logger; DB *gorm.DB; Server *server.Server; Container *Container }
func Bootstrap(ctx context.Context) (*App,error); func (a *App) Start() error; func (a *App) Stop(ctx context.Context) error
type Container struct{}; func NewContainer(cfg *config.Config, db *gorm.DB, log *slog.Logger, cache cache.Cache) *Container; func (c *Container) RegisterModules(v1 *gin.RouterGroup)

type AppError struct{ Err error; Message string; Cause error }
var ErrNotFound, ErrForbidden, ErrUnauthorized, ErrConflict, ErrBadRequest, ErrInternal error
func Wrap(err error,msg string,cause ...error)*AppError; func Status(err error) int; func PublicMessage(err error, env string) string

type APIResponse struct{ Success bool `json:"success"`; Message string `json:"message"`; Data any `json:"data"`; Meta any `json:"meta"`; Errors any `json:"errors"` }
type PaginationMeta struct{ Page int `json:"page"`; PerPage int `json:"per_page"`; Total int64 `json:"total"`; TotalPages int `json:"total_pages"` }
func Success(c *gin.Context,msg string,data any); func Created(c *gin.Context,msg string,data any); func Error(c *gin.Context,status int,msg string,errs any); func FromError(c *gin.Context,err error,env string); func ValidationError(c *gin.Context,errs validator.ValidationErrors); func Paginated(c *gin.Context,msg string,data any,page,perPage int,total int64)

type Cache interface{ Get(context.Context,string)([]byte,error); Set(context.Context,string,[]byte,time.Duration) error; Delete(context.Context,string) error; Close() error }
type NoopCache struct{}

type BaseModel struct{ ID uint `gorm:"primaryKey" json:"-"`; PublicID string `gorm:"type:char(26);uniqueIndex;not null" json:"id"`; CreatedAt,UpdatedAt time.Time; DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`; CreatedBy,UpdatedBy *uint `gorm:"index" json:"-"` }
type BaseModelSimple struct{ ID uint `gorm:"primaryKey" json:"-"`; PublicID string `gorm:"type:char(26);uniqueIndex;not null" json:"id"`; CreatedAt time.Time `json:"created_at"`; UpdatedAt time.Time `json:"updated_at"`; DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` }
```

**Rationale**: Contracts are small and package-owned. `Container` is intentionally empty for change-01 because no business modules exist yet.

## Bootstrap, Router, Middleware, Shutdown

Initialization order in `internal/app/bootstrap.go`: `config.Load()` → set Gin mode → `logger.New()` with slog+lumberjack → `db.Connect()` → `cache.NewNoop()` → `NewContainer()` → `server.NewRouter(cfg, log, container.RegisterModules)` → `server.New(cfg, router)` → return `App`. `cmd/api/main.go` starts the app and listens for SIGINT/SIGTERM.

Router structure: root `gin.Engine` mounts `GET /health` before versioned business routes, then `v1 := root.Group("/api/v1")`; future modules call `users.Register(v1, container.Users.Handler)`. No modules are registered in change-01.

Domain coverage: `environment-config` owns `Config`; `database-connection` owns GORM pool; `migrations` owns Goose targets; `api-response` owns envelope; `error-handling` owns `AppError`; `request-validation` owns validator + Gin helper; `http-middleware` owns chain; `server-bootstrap` owns router/server/shutdown; `app-orchestration` owns `App`/`Container`; `dev-environment` owns Docker/Make/README; `platform-stubs` owns future package stubs.

Middleware chain order is `RequestID → Logger → Recovery → CORS`. This ensures all logs and panic responses contain the same request ID; CORS handles preflight near the route boundary.

Graceful shutdown: signal received → create timeout context from `SHUTDOWN_TIMEOUT_SECONDS` → `Server.Shutdown(ctx)` stops accepts and waits → close `sql.DB` extracted from GORM → `cache.Close()` → log completion. Timeout forces close through `http.Server` semantics.

Logger: `infra/logger.New(cfg)` creates a JSON `slog.Logger`; output is stdout locally and `lumberjack.Logger` when file logging is configured. Rationale: standard structured logging with bounded log files.

## File Changes

| File | Action | Description |
|---|---|---|
| `cmd/api/main.go`, `cmd/nexokit/main.go` | Create | Thin API and CLI stub entrypoints. |
| `internal/{app,config,infra,server,middleware,platform,shared,modules,cli}/...` | Create | Foundation packages and stubs described above. |
| `migrations/`, `.env.example`, `docker-compose.yml`, `Makefile`, `README.md`, `tests/helpers/app.go` | Create | Goose, local Postgres, commands, docs, test scaffolding. |

## Testing Strategy

| Layer | What | Approach |
|---|---|---|
| Unit | config parsing, response structs, app errors, validator rules, NoopCache | `go test ./...` with table tests. |
| HTTP | health, 404 envelope, middleware request ID/recovery/CORS | `httptest` against router. |
| Build/tooling | entrypoints, Makefile command shape | `go build ./...`; migration targets documented, DB-dependent runs manual/local. |

## Migration / Rollout

No production migration required. Change introduces `migrations/` and Makefile targets. Initial SQL may be empty except Goose sections because no business tables exist.

## Open Questions

None.
