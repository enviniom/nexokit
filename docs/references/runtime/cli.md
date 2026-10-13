# NexoKit CLI reference

Run the project-local CLI from the repository/release root: `go run ./cmd/nexokit <command>` during development, or `./bin/nexokit <command>` after `make build`. This is a small custom dispatcher, not Cobra and not a globally installed project generator.

## Commands and Make equivalents

| Command | Effect | Make equivalent |
|---|---|---|
| `help`, `--help`, `-h` | List top-level commands | — |
| `init` | Ask display name and GitHub username; derive module, preview, confirm (default No) | — |
| `init --name <name> --module <module> [--dry-run \| --apply]` | One-time local identity preview; deliberate apply | — |
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

Make wrappers leave `.env` parsing to the application's dotenv loader, preserving
quoted credentials rather than interpreting them as Make or shell syntax.
Inherited environment values, including explicitly empty values, override dotenv;
the wrappers do not print environment values. Bare `make` runs `dev`, and
`make init` forwards to `go run ./cmd/nexokit init`. Migration configuration
validation and errors come from the CLI, not a separate Make environment guard.

`make build` produces `bin/api` and `bin/nexokit`. Migration/seed paths are fixed relative to the working directory (`migrations/`, `seeds/`); there is no directory/config-file flag or migration dry-run flag. Except for initialization, module generation and root creation, commands define no custom flags. Top-level unknown commands fail; no arguments print usage and exit 1. `status` can report an unreachable database and still exit successfully—do not use it as a readiness gate.

## One-time local initialization

From a fresh starter's current repository root:

```bash
go run ./cmd/nexokit init
# Explicit flags for automation:
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api --dry-run
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api --apply
```

Bare `init` asks for application display name and GitHub username, not a module.
`My API` and `user-name` produce `github.com/user-name/my-api`; the trimmed display
name is preserved. The lowercase ASCII slug collapses spaces/dots/underscores/
hyphens and trims separators. GitHub usernames must be 1–39 ASCII alphanumeric
or hyphen characters, without leading/trailing/consecutive hyphens. No GitHub
API, credentials or remote repository creation is involved. After preview,
`Apply this plan? [y/N]` accepts trimmed case-insensitive `yes`/`y`; No, blank or
EOF cancels successfully. Invalid confirmation fails without writes.

Flag calls never read stdin; both identity flags are required, including with
`--dry-run`. Both preview forms write nothing. Review the sorted changed paths and remaining
known tool/provenance references before applying. `--name` and `--module` are
required; unknown/duplicate flags, extra arguments and contradictory preview/apply
flags fail. Go module paths with a domain and normal import components (including
`/v2` or later) are accepted; traversal, control characters and injection are not.
Display names permit letters/digits, spaces, dots, hyphens and underscores.

Initialization rewrites the module, project imports (including embedded templates
and generated-test source), README heading and example APP_NAME. `cmd/nexokit`
and developer-tool attribution remain unchanged. It is a one-time local adoption,
not a downloader or update contract; invalid, partial, repeated or symlinked
starter identities fail safely. Candidates are preflighted and staged before
commit; write errors trigger rollback with explicit failure reporting. Stop
concurrent source writers during apply; this is not a cross-process transaction.

Existing `.env`, dependencies and `go.sum` are untouched. After apply, configure
PostgreSQL, PASETO, CORS and rate limits, then run `go build ./...` and
`go test ./...`. Run migrations, seeds and root-account creation deliberately after
configuration; none run automatically. Local defaults are not production security
guarantees, and no production secrets are generated. See [adoption](../../starter-template.md).

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

Generation is a scaffolding operation, separate from app composition and schema deployment. Construct the generated container and explicitly register it on a group with appropriate authentication, permissions and tenant middleware. Define and review the migration schema and route policy before use; generated SQL is a stub, not a deployable tenant schema. Existing target directories are refused rather than overwritten. See [adoption](../../starter-template.md).

## Root creation and password handling

Flags `--name`, `--email`, `--password` override `ROOT_USER_NAME`, `ROOT_USER_EMAIL`, `ROOT_USER_PASSWORD`; missing values are prompted. Put root flags before positional arguments (standard Go `flag` parsing). Name/email are trimmed. Validation requires a nonempty name, a parseable email, and a password at least eight characters long containing uppercase, lowercase and a digit.

Non-local/non-test environments require typing `yes`, unless `--force` skips confirmation. Use that override only for controlled automation. The prompt uses ordinary scanner input and **does not hide the password**. Prefer secure environment injection; avoid `--password` in shell history/process listings, and do not log credentials.

The command connects to PostgreSQL and uses `password.Manager` plus `rootStorage`. It finds the IAM role with slug `root`, hashes the password, generates a public ID and inserts an active user linked to that role. Run migrations and required IAM seeds first; the command does not create the role. If any user already has that root role, it prints a skip message and returns success. The existence check and insert are separate operations, not a serialized/transactional concurrency guarantee—run root provisioning once, not concurrently.

Sources: `internal/cli/commands/createroot.go`, `root_storage.go`, `internal/cli/root/root.go`.

## Configuration display

`config.Load` optionally loads working-directory `.env`; exported environment values take precedence. The `config` command masks DB password and database URL with `***masked***`, and displays app, DB pool, CORS, logging, shutdown and cache driver fields. It **omits** auth, Redis, rate-limit and platform-domain settings; it is not a full configuration audit. Treat even masked output as operational information, not public diagnostics.

## Seeds

Files must use `package seeds` and export a zero-argument function ending in `Seed` that returns exactly `error`. Discovery parses Go files; no matching functions is a successful no-op. Execution generates a runner under `.tmp/`, invokes `go run`, then cleans it up. The runner needs the source tree, Go toolchain, writable working directory and database access; a standalone binary is insufficient. The one-time `init --apply` rewrites the seed runner's project import together with other source imports.

Review seeds before executing them against production; there is no general dry-run, transaction or idempotence guarantee.

## Destructive operations

`migrate down` and `migrate reset` execute Down SQL and may irreversibly remove data. The CLI has no production guard or confirmation for them. Never use reset as an ordinary deployment/rollback step; require approval and a tested backup/restore plan. Follow the [deployment runbook](../../deployment.md#rollback).
