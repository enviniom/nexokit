# NexoKit CLI

The `nexokit` CLI is an internal developer tool shipped with every cloned project. It is **not** intended for global installation.

## Commands

| Command | Description |
|---------|-------------|
| `serve` | Start the HTTP server |
| `config` | Print resolved configuration (secrets masked) |
| `status` | Print app version, DB status, and migration count |
| `migrate up` | Apply pending database migrations |
| `migrate down` | Rollback the last batch of migrations |
| `migrate status` | Show current migration version |
| `migrate reset` | Rollback all migrations |
| `migrate create <name>` | Create a new timestamped migration file |
| `make module <name>` | Generate a module directory with CRUD files |
| `make migration <name>` | Create a migration file via Goose |
| `make seed <name>` | Create a seed file stub |
| `create-root` | Validate root user input and confirmation (persistence blocked until auth schema is wired) |
| `seed` | Discover and run seed files from `seeds/` |

## Flags

### `make module`

| Flag | Description |
|------|-------------|
| `--crud` | Include full CRUD handlers, service, and repository methods |
| `--migration` | Create a corresponding Goose migration file |
| `--tenant` | Add `company_id` scoping to repository queries |

### `create-root`

| Flag | Description |
|------|-------------|
| `--email` | Root user email (skips prompt) |
| `--password` | Root user password (skips prompt) |
| `--force` | Skip confirmation prompt in non-local environments |

## Out of Scope

- `nexokit new` — global interactive project scaffolding is planned for a future change.
- `permissions sync` — automatic permission discovery waits for the RBAC change.

## Seed Files

Seed files live in `seeds/` and must:
- Use `package seeds`
- Export a function named `*Seed` with signature `func() error`

The `seed` command discovers these functions dynamically and runs them in a temporary Go program.

## Root Creation Safety

`create-root` validates email format and password strength, requires confirmation in non-local environments, and is idempotent. Today it returns a clear error indicating that root creation is blocked until the auth schema and password hashing changes are wired. The CLI boundary (flags, prompts, validation, and confirmation) is stable and ready; the underlying storage persistence will be enabled by a future auth change.
