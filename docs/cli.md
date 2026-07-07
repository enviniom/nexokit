# NexoKit CLI

The `nexokit` CLI is an internal developer tool shipped with every cloned project. It is **not** intended for global installation. The root [`Makefile`](../Makefile) is the documented happy path for daily development; this guide is the source of truth for direct `nexokit` usage.

## Command overview

| Command | Usage | Description |
|---------|-------|-------------|
| `serve` | `nexokit serve` | Start the HTTP server |
| `config` | `nexokit config` | Print resolved configuration (secrets masked) |
| `status` | `nexokit status` | Print app version, DB status, and migration count |
| `migrate up` | `nexokit migrate up` | Apply pending database migrations |
| `migrate down` | `nexokit migrate down` | Rollback the last batch of migrations |
| `migrate status` | `nexokit migrate status` | Show current migration status |
| `migrate reset` | `nexokit migrate reset` | Rollback all migrations |
| `migrate create <name>` | `nexokit migrate create <name>` | Create a new timestamped migration file in `migrations/` |
| `make module <name>` | `nexokit make module <name> [flags]` | Generate a module directory with CRUD files |
| `make migration <name>` | `nexokit make migration <name>` | Create a migration file via Goose |
| `make seed <name>` | `nexokit make seed <name>` | Create a seed file stub in `seeds/` |
| `create-root` | `nexokit create-root [flags]` | Create the root user (idempotent) |
| `seed` | `nexokit seed` | Discover and run seed files from `seeds/` |

## `serve`

Starts the HTTP server on the configured `APP_PORT` (default `8080`). The server bootstraps the full dependency graph, runs the Gin middleware chain, and mounts module routes under `/api/v1`.

```bash
nexokit serve
```

## `config`

Prints the resolved configuration as JSON, with secrets such as `DB_PASSWORD` and `DATABASE_URL` masked. Requires the same environment as the API (`DATABASE_URL` or `DB_*` variables are loaded but not required to be reachable).

```bash
nexokit config
```

Expected output scope:

- `app` — name, environment, public URL, port.
- `db` — host, port, database name, user, SSL mode, pool settings, masked credentials.
- `cors` — allowed origins.
- `log` — level, format, file paths, rotation settings.
- `shutdown` — graceful shutdown timeout.
- `cache` — driver (`redis`, `valkey`, or `none`).

Use this command to verify that environment variables and `.env` are being read as expected before starting the server.

## `status`

Prints a short health snapshot: build version, database reachability, current Goose migration version, and the number of `.sql` migration files in `migrations/`.

```bash
nexokit status
```

Expected output scope:

- `Version` — build version, or `dev` when built without `ldflags`.
- `Database` — `connected` or `unreachable (<error>)`.
- `Migration version` — current applied version reported by Goose, or `unknown` if the DB is unreachable.
- `Migrations` — count of `.sql` files in `migrations/`.

This is useful for pre-flight checks and liveness/readiness debugging. For ongoing monitoring and alert thresholds, see [`docs/deployment.md`](deployment.md#alert-thresholds).

## `migrate`

Migration commands require a database connection. Set `DATABASE_URL` or the `DB_*` variables from [`.env.example`](../.env.example) before running them.

```bash
nexokit migrate up
nexokit migrate down
nexokit migrate status
nexokit migrate reset
nexokit migrate create add_users_table
```

Migration names must be `snake_case`. Files are created in `migrations/` with a `YYYYMMDDHHMMSS_` prefix.

> **Danger Zone.** `migrate down` and `migrate reset` are destructive and forbidden by operational policy in production unless explicitly approved. The CLI binary does not block these commands; enforcement is the operator's responsibility. Use them only with a verified backup and explicit approval. `migrate reset` is intended for non-production environments or disaster recovery.

## `seed`

Seed files live in `seeds/` and must:

- Use `package seeds`.
- Export a function named `*Seed` with signature `func() error`.

The `seed` command discovers these functions dynamically and runs them in a temporary Go program.

> **Prerequisite.** `nexokit seed` requires the Go toolchain in the execution environment, or must be run from an admin/build environment with DB access.

```bash
nexokit seed
```

## `create-root`

Creates the initial root user from flags, environment variables, or interactive prompts. The command is idempotent: if a root user already exists, it prints a skip message.

```bash
nexokit create-root
nexokit create-root --email root@example.com --password '<GENERATED-PASSWORD>' --force
```

> **Do not pass real passwords on the command line.** Shell history, process listings, and audit logs can capture CLI arguments. Prefer `nexokit create-root` (interactive prompt) or inject `ROOT_USER_PASSWORD` through a secret manager / environment file.

### `create-root` flags

| Flag | Description |
|------|-------------|
| `--name` | Root user name (falls back to `ROOT_USER_NAME`) |
| `--email` | Root user email (falls back to `ROOT_USER_EMAIL`) |
| `--password` | Root user password (falls back to `ROOT_USER_PASSWORD`) |
| `--force` | Skip confirmation prompt in non-local environments |

Validation requires a non-empty name, a valid email, and a password with at least 8 characters containing uppercase, lowercase, and a digit.

## `nexokit make module`

Generates a new module under `internal/modules/<name>/` with a standard folder structure and optional CRUD files.

```bash
nexokit make module products
nexokit make module products --crud --migration --tenant
```

### `nexokit make module` flags

| Flag | Description |
|------|-------------|
| `--crud` | Include full CRUD handlers, service, and repository methods |
| `--migration` | Create a corresponding Goose migration file |
| `--tenant` | Add `company_id` scoping to repository queries |

## `nexokit make migration`

Creates a new Goose migration file without generating a module.

```bash
nexokit make migration create_orders_table
```

## `nexokit make seed`

Creates a new seed stub in `seeds/`.

```bash
nexokit make seed system_permissions
```

## Makefile equivalents

| Direct CLI | Makefile target |
|------------|-----------------|
| `nexokit serve` | `make dev` |
| `nexokit migrate up` | `make migrate-up` |
| `nexokit migrate down` | `make migrate-down` |
| `nexokit migrate status` | `make migrate-status` |
| `nexokit migrate reset` | `make migrate-reset` |
| `nexokit migrate create <name>` | `make migrate-create` |
| `nexokit seed` | `make seed` |
| `nexokit create-root` | `make create-root` |

## Out of scope

- `nexokit new` — global interactive project scaffolding is planned for a future change.
- `permissions sync` — automatic permission discovery waits for the RBAC change.
