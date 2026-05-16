# Design: Internal CLI and Developer Experience

## Technical Approach

Implement `cmd/nexokit` as a thin entry point that delegates all behavior to `internal/cli`. Use standard-library parsing (`flag` plus small command dispatcher) to keep the template dependency-light. The CLI wraps existing app/bootstrap, config, and Goose helpers, and the Makefile exposes stable shortcuts for day-to-day development.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| CLI boundary | `cmd/nexokit/main.go` only calls `cli.Execute(os.Args[1:], stdio)` | Put command logic in `cmd/`; merge with `cmd/api` | Preserves the project rule that `cmd/*` has no business logic and avoids coupling API serving to tooling. |
| CLI package shape | `internal/cli` with `commands`, `generator`, `templates`, `root`, `stdio` helpers | One large `internal/cli` package | Keeps command orchestration, filesystem generation, and root safety separately testable. |
| Parser | Standard `flag`-based dispatcher | Cobra | Current command surface is small; avoiding Cobra prevents a new framework dependency before the CLI contract stabilizes. |
| Generator templates | `text/template` files embedded with `embed.FS` | String concatenation; external template files only | Embedded templates are testable, versioned with code, and do not require runtime filesystem assumptions. |
| Root creation | Safe command wrapper with explicit env/input validation and idempotency checks | Blind insert seed | Root user creation touches credentials; it must refuse unsafe or ambiguous input. |

## Data Flow

```txt
Makefile ──→ go run ./cmd/nexokit <command>
                │
                ▼
          internal/cli dispatcher
           ├─ serve ──→ app.Bootstrap(ctx).Start()
           ├─ migrate ──→ config.Load → db.Connect → infra/db goose helpers
           ├─ create-root ──→ root service adapter → DB transaction
           └─ make module ──→ generator → templates → files + optional migration
```

## File Changes

| File | Action | Description |
|---|---:|---|
| `cmd/nexokit/main.go` | Modify | Replace placeholder with thin CLI entry point and exit-code handling. |
| `internal/cli/cli.go` | Create | Dispatcher, command registry, usage text, shared execution contract. |
| `internal/cli/commands/*.go` | Create | `serve`, `config`, `status`, `migrate`, `make`, `create-root`, `seed`. |
| `internal/cli/generator/*.go` | Create | Module name validation, filesystem writes, idempotency checks, template execution. |
| `internal/cli/templates/module/*.tmpl` | Create | Templates for flat module files. |
| `internal/cli/root/*.go` | Create | Root-user creation command/service boundary; adapters can evolve when auth schema lands. |
| `Makefile` | Modify | Add `dev`, `build`, `test-unit`, `test-integration`, `test-coverage`, CLI-backed migration, seed, root, lint targets. |
| `docs/cli.md` or README section | Create/Modify | Document internal CLI commands and non-goals. |
| `tests/fixtures/cli/golden/` | Create | Golden expected files for generator tests. |

## Interfaces / Contracts

```go
type Command interface { Name() string; Run(context.Context, []string) error }
type Stdio struct { In io.Reader; Out, Err io.Writer }
type ModuleOptions struct { Name string; CRUD, Migration, Tenant bool }
```

Generated modules use flat files: `handler.go`, `service.go`, `repository.go`, `dto.go`, `model.go`, `routes.go`, `validation.go`. CRUD generation uses `shared.BaseModel`, exposes `PublicID` as JSON `id`, includes `Register(v1 *gin.RouterGroup, h *Handler)`, and tenant-aware repository filters only when `--tenant` is set.

## Migration / Root Strategy

`migrate up|down|status|reset` uses existing `internal/infra/db` Goose helpers after `config.Load` and `db.Connect`. `make migration <name>` creates timestamped SQL in `migrations/` with validated snake_case names. Makefile targets call the CLI instead of duplicating Goose commands.

`create-root` requires explicit email/password input via flags or prompt, refuses empty/default credentials, requires confirmation outside local/test envs, runs in a transaction, checks whether root already exists, and never logs the password. Because auth/user schema is not present yet, implement the boundary so storage logic can be completed by the auth change without changing CLI UX.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Dispatcher, command validation, module name normalization | Table-driven Go tests. |
| Golden | Generated module with `--crud`, `--migration`, `--tenant` | Compare against `tests/fixtures/cli/golden`; update intentionally only. |
| Integration | Migration commands and root safety | DB-backed tests when integration env is available; otherwise skip with clear reason. |

## Non-Goals

- No global installable `nexokit new`; this CLI is internal to a cloned project.
- No `permissions sync`; permission discovery/sync waits for a mature RBAC model.

## Open Questions

- None blocking. Root storage implementation depends on the later auth schema, but the CLI command contract can be stabilized now.
