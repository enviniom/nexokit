# NexoKit CLI reference

Run the project-local CLI from the repository/release root: `go run ./cmd/nexokit <command>` during development, or `./bin/nexokit <command>` after `make build`. This is a small custom dispatcher, not Cobra and not a globally installed project generator.

## Commands and Make equivalents

| Command | Actual effect | Make equivalent |
|---|---|---|
| `help`, `--help`, `-h` | List top-level commands | — |
| `serve` | Bootstrap/start API; interrupt/SIGTERM shutdown | `make dev` |
| `config` | Print selected resolved configuration as JSON | — |
| `status` | Build version, database reachability, SQL file count and Goose version | — |
| `migrate up` | Goose Up: pending migrations | `make migrate-up` |
| `migrate down` | Goose Down: one latest migration, not an entire deployment batch | `make migrate-down` |
| `migrate status` | Goose status listing; not an execution dry run | `make migrate-status` |
| `migrate reset` | Goose Reset: all migrations down | `make migrate-reset` |
| `migrate create <name>` | New Goose SQL migration | `make migrate-create` (prompts for name) |
| `make migration <name>` | Same SQL migration generator | — |
| `make seed <name>` | Timestamped seed stub | — |
| `make module <name> [--crud] [--migration] [--tenant]` | Module scaffold | — |
| `seed` | Discover and execute seed functions | `make seed` |
| `create-root [--name ...] [--email ...] [--password ...] [--force]` | Persist initial root IAM user | `make create-root` (no flag forwarding) |

`make build` produces `bin/api` and `bin/nexokit`. Migration/seed paths are fixed relative to the working directory (`migrations/`, `seeds/`); there is no directory/config-file flag or migration dry-run flag. Except for module generation and root creation, commands define no custom flags. Top-level unknown commands fail; no arguments print usage and exit 1. `status` can report an unreachable database and still exit successfully—do not use it as a readiness gate.

## Module generation

Put the name **before** flags. Names must satisfy the generator's snake_case validation; migration/seed names accept lowercase letters, digits and underscores.

| Flag | Effect |
|---|---|
| `--crud` | Adds create/update/delete slices and routes to the default get/list scaffold |
| `--migration` | Generates corresponding migration |
| `--tenant` | Requires trusted company scope; only list allows explicit root-global context |

The generator writes formatted Go following the [new-code architecture](../../architecture_and_vsa.md):

```text
internal/modules/<name>/
├── container.go
├── routes.go
├── core/
│   ├── contracts.go
│   ├── errors.go
│   └── <name>/
│       ├── model.go
│       └── dtos.go
└── slices/
    ├── get_<name>/       # handler.go + repository.go
    ├── list_<name>/
    └── ...              # create/update/delete with --crud
```

Plain CRUD uses handler + repository, not a pass-through service. DTO pointer validation trims names before rules; response DTOs own `FromModel`, exposing public IDs, names and timestamps only. Entity `BeforeCreate` generates ULIDs using `platform/identity` and preserves valid supplied IDs. Handlers use `response.HandleError` for safe unexpected errors; repositories translate GORM not-found, duplicate-key and foreign-key sentinels into module errors. Enable GORM `TranslateError` in the application's database configuration for driver constraint errors to become those sentinels; otherwise unexpected driver errors still receive safe 500 responses.

Tenant handlers resolve `tenant.FromGin`; missing/invalid scope returns 400 before persistence. Create/get/update/delete require a selected company. Create assigns `company_id` from trusted scope; JSON `company_id` is ignored and is never returned. Updates cannot reassign ownership. Every repository read/write explicitly applies scope; only list permits a trusted root-global context established by deliberate middleware policy. Lists support normalized pagination (maximum 100), name search, creation-date filters and an explicit sorting allowlist (`name`, `created_at`, `updated_at`); there is no status field/filter in this scaffold.

Generation does **not** migrate existing modules or wire routes into the app. Construct the generated container and explicitly register it on a group with appropriate authentication, permissions and tenant middleware. Review the placeholder migration, schema and route policy before use; generated migration stubs are not a finished tenant schema. Existing target directories are refused rather than overwritten. See [adoption](../../starter-template.md).

## Root creation and password handling

Flags `--name`, `--email`, `--password` override `ROOT_USER_NAME`, `ROOT_USER_EMAIL`, `ROOT_USER_PASSWORD`; missing values are prompted. Put root flags before positional arguments (standard Go `flag` parsing). Name/email are trimmed. Validation requires a nonempty name, a parseable email, and a password at least eight characters long containing uppercase, lowercase and a digit.

Non-local/non-test environments require typing `yes`, unless `--force` skips confirmation. Use that override only for controlled automation. The prompt uses ordinary scanner input and **does not hide the password**. Prefer secure environment injection; avoid `--password` in shell history/process listings, and do not log credentials.

Persistence is implemented: the command connects to PostgreSQL and uses `password.Manager` plus `rootStorage`. It finds the IAM role with slug `root`, hashes the password, generates a public ID and inserts an active user linked to that role. Run migrations and required IAM seeds first; the command does not create the role. If any user already has that root role, it prints a skip message and returns success. The existence check and insert are separate operations, not a serialized/transactional concurrency guarantee—run root provisioning once, not concurrently.

Sources: `internal/cli/commands/createroot.go`, `root_storage.go`, `internal/cli/root/root.go`.

## Configuration display

`config.Load` optionally loads working-directory `.env`; exported environment values take precedence. The `config` command masks DB password and database URL with `***masked***`, and displays app, DB pool, CORS, logging, shutdown and cache driver fields. It **omits** auth, Redis, rate-limit and platform-domain settings; it is not a full configuration audit. Treat even masked output as operational information, not public diagnostics.

## Seeds

Files must use `package seeds` and export a zero-argument function ending in `Seed` that returns exactly `error`. Discovery parses Go files; no matching functions is a successful no-op. Execution generates a runner under `.tmp/`, invokes `go run`, then cleans it up. The runner needs the source tree, Go toolchain, writable working directory and database access; a standalone binary is insufficient. After module renaming, update the hardcoded seeds import in `internal/cli/commands/seed.go` too.

Review seeds before executing them against production; there is no general dry-run, transaction or idempotence guarantee.

## Destructive operations

`migrate down` and `migrate reset` execute Down SQL and may irreversibly remove data. The CLI has no production guard or confirmation for them. Never use reset as an ordinary deployment/rollback step; require approval and a tested backup/restore plan. Follow the [deployment runbook](../../deployment.md#rollback).
