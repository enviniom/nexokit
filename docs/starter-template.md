# Adopt NexoKit as a starter

Fork or clone the repository into your namespace; it is a starting application, not a library dependency. Read [the application map](system-map.md) before changing composition.

## Quick path

Prerequisites: Go 1.26.3+ (current `go.mod`) and the project dependencies available
to Go. Run `make init` immediately after cloning/forking; it prepares only local
files and needs no running PostgreSQL. Start PostgreSQL later, after reviewing the
initial migrations and seeds. The initializer provisions no services.
Redis/Valkey is optional: `CACHE_DRIVER=none` and the default in-memory rate limiter
need no Redis. Configuring Redis is a deliberate separate choice.

1. From a fresh local clone/fork root, run the single setup entrypoint:

```bash
make init
# Only after setup succeeds:
go build ./...
go test ./...
```

Interactive `init` asks for display name and GitHub username, deriving
`github.com/user-name/my-api` from `My API` and `user-name`. It creates no GitHub
repository and requests no GitHub credentials. It collects environment choices
and local configuration, shows a redacted plan with the configured database target
without contacting it, then requests one confirmation, default No. Declining
leaves files unchanged. Approval applies only identity/environment files. Init
never connects to the database, runs migrations or seeds, creates root, or asks
for root credentials. Its completion message lists the later manual commands.

2. Application identity changes in the root README heading and example `APP_NAME`;
   Go module imports, embedded generator templates, seed runner and generated-test
   source change consistently. The developer tool remains `nexokit` in
   `cmd/nexokit`; NexoKit attribution is retained. Names with spaces are supported:
   use `--name "My API"`. This is one-time independent application ownership, not
   downloading, updating or synchronizing an upstream project.
3. If a regular `.env` exists, explicitly choose **keep** or **backup and rebuild**.
   Keep preserves its bytes and mode, using its exact database URL/components;
   validate and deliberately update incompatible configuration yourself, rather
   than expecting silent replacement. Rebuild retains a protected backup and
   creates fresh development configuration from the example (details below).
   Review CORS/rate limits deliberately; local defaults are not production
   safeguards. See [deployment configuration](deployment.md#configuration-and-secrets).
4. Adapt initial migrations, seeds and modules to your application before applying
   any schema. When ready, use an appropriate empty PostgreSQL database with
   migration permissions. The bundled Docker Compose service is an option:

```bash
docker compose up -d postgres
docker compose ps postgres
# Wait until PostgreSQL is healthy, then run explicitly:
make migrate-up
make seed
make create-root
make dev
```

With Compose defaults, configure `localhost:5432`, database/user/password `nexokit`
and `DB_SSL_MODE=disable` during init. These are local-development defaults, not
production credentials; overrides or existing volumes require their actual values.
The wizard's proposed database name/user may differ. Root creation comes after
seeds because it requires the root role. Prefer protected `ROOT_USER_*` process
inputs: the separate create-root command's interactive password prompt is not hidden.

Make delegates `.env` loading to the application, not Make or shell evaluation.
Use dotenv quoting for credentials containing spaces, quotes, `#` or `$`;
externally supplied environment values (including empty ones) take precedence.
The wrappers do not print environment values. Bare `make` still runs `make dev`;
`make init` delegates to the same interactive initializer shown above.

5. Check `/health/live` and `/health/ready`, then exercise login and a protected route. Root user routes require `X-Company-ID`; see [request flow](references/runtime/request-flow.md).

These are deliberate commands for the adopter to run. `init` only prepares local
identity/configuration; it never provisions or contacts a database, executes Git
operations or changes `go.sum`/dependencies. It is once-only, not a maintenance/upgrade command. Local
private adoption does not select public distribution license terms.

## Local-development questionnaire

Interactive `init` / `make init` creates an absent `.env`, or rebuilds one only
when explicitly selected. For new/rebuilt configuration it collects: `DB_HOST` (localhost), `DB_PORT` (5432), `DB_NAME` (app slug),
`DB_USER` (postgres), password (no shared default), `DB_SSL_MODE` (disable),
`APP_PORT` (8080), `APP_URL` (http://localhost with the selected port), then
`CORS_ALLOWED_ORIGINS` (http://localhost:3000). Enter selects a displayed default;
Enter at password intentionally chooses an empty local password. Password spaces
are preserved. SSL modes are disable/allow/prefer/require/verify-ca/verify-full;
ports must be 1–65535 and URLs/origins must be credential-free HTTP(S). CORS accepts
comma-separated explicit origins, not a wildcard default.

Real terminal password entry uses no-echo input; piped input is consumed without
CLI echo but is not protected from shell history or the input producer. Do not
pretype passwords: already buffered terminal input is refused. Cancellation is
checked before/after synchronous reads, not while waiting for input. A fresh
32-byte random key is encoded as 64 hex characters in memory; password/key are
never previewed. Invalid input or entropy failure writes nothing. One default-No
confirmation covers only local identity/environment changes. Root credentials
are not collected or required by init. The later `make create-root` command
validates its own name, email and password; never store its password in dotenv,
command arguments or committed demo scripts.

The initializer creates `.env` from its snapshotted, identity-updated example,
preserving unknown content and comments. Existing regular files require the keep
or rebuild choice; symlinks and nonregular destinations are refused. Rebuild
retains `.env.backup-<32 lowercase hex characters>` at mode `0600`; only that
narrow backup-name pattern is ignored by Git. New/rebuilt files use mode `0600`
and exclusive hard-link publication. Keep leaves bytes/mode unchanged. A
filesystem without hard-link support fails closed instead of overwriting.
`APP_ENV` is forced to `development` and `DATABASE_URL` is cleared so `DB_*` applies.
These development overrides apply to new/rebuilt configuration, not kept files.
The plan layer does not invent root credentials.

Environment creation participates in staged identity updates and rollback. Use a
quiescent local copy: this is not a cross-process or power-loss transaction.
Rollback failures report recovery paths, never environment values. If temporary
file removal fails, a protected `0600` staging file can remain at the reported
path; review and remove that file locally without displaying its contents.

## Automatic CI

GitHub Actions runs tests, build, vet, formatting and module error-boundary checks
on pushes and pull requests to all branches; manual dispatch remains available.
Hosted execution requires a push or pull request (or manual dispatch on GitHub)
and has not been run locally. You control whether and when to publish; local
checks do not prove a hosted run succeeded.

## Initialization safety

Flag calls never prompt: both `--name` and `--module` are required, even with
`--dry-run`. Preview is stdin-free/write-free and is the flag default:

```bash
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api --dry-run
```

`--apply` performs local identity/environment initialization without stdin or any
database connection. New/rebuilt configuration requires the selected local process
inputs; root credentials are neither required nor used. Applying with an existing
environment requires `--env-action keep` or `--env-action rebuild`. `--apply` plus `--dry-run` is an
error. Unknown/duplicate flags and positional arguments are refused. The current
root must retain the starter module, README heading, example APP_NAME and CLI
source identity anchors; partial or repeated initialization is refused.

Only known Go/template families under `cmd`, `internal`, `tests` and `seeds`, the
module directive, README, example and selected adoption/CLI docs are considered.
The CLI reports remaining known tool/provenance references instead of replacing
every occurrence of `nexokit`. Git data, local environment files, logs, vendor,
temporary/task artifacts and unrecognized files are excluded. Planned symlinks
and symlink ancestors are refused rather than followed.

Use a quiescent local copy: stop editors/watchers that write source during apply.
All candidate contents, types and modes are checked before writes; replacements
and rollback copies are staged with preserved modes. A failed commit restores
completed replacements and reports any rollback failure. This is not a
cross-process transaction or a power-loss recovery system. Review errors before
recovery; no automatic repair or repeated-init/upgrade framework is provided.

## Failure recovery and maintenance

Local configuration/identity validation failure before apply leaves the project
unadopted. Publication errors retain the file transaction's rollback/recovery
diagnostics and protected recovery paths. Inspect backups locally without printing
secrets; init is not a filesystem lock or power-loss recovery mechanism. After
successful adoption, the one-time guard refuses another init.

Database preparation is separate and manual. Review migrations and seeds first;
then execute migrations, seeds and root creation in that order. If one of those
commands fails, stop and inspect its result before running the next command.
Database changes may already have persisted: there is no universal transaction
across these separate commands or arbitrary seeds. Use the appropriate existing
command for deliberate recovery, not another init:

| Command | Deliberate recovery/maintenance action |
|---|---|
| `make migrate-status` / `nexokit migrate status` | Inspect migration state |
| `make migrate-up` / `nexokit migrate up` | Apply remaining migrations |
| `make seed` / `nexokit seed` | Execute reviewed seeds; consider prior side effects |
| `make create-root` / `nexokit create-root` | Create a missing root after seeds; refuses an existing root, never resets its password |

These are the explicit database-preparation steps after local init and schema
review, and also maintenance/recovery tools. They are not permission to rerun init.
Start `make dev` only after schema, seeds and root are ready.

## Choose capabilities deliberately

| Module | Keep for |
|---|---|
| companies | Company/domain administration and tenant resolution |
| IAM | Authentication, access/refresh tokens, sessions, users, roles, permissions and authenticated-user lookup |
| onboarding | Root-controlled creation of a company and its initial identities |

Removing a module is more than removing its folder: remove its imports, container fields/construction, injected dependencies and `RegisterModules` mounting in `internal/app/container.go`. Review `SyncPermissions`, tenant resolvers, auth lookup, CLI root creation, seeds, migrations, fixtures and tests. Authentication is internal to IAM under `slices/auth`, not an independently constructed or registered module. Its credential, session and token projections remain separate from IAM entity models. Private tenant resolution depends on companies; CLI root creation depends on IAM. Choose an alternate identity/tenant policy before removing those capabilities.

For a fresh project, adjust initial schema and seeds coherently. For an existing deployed project, preserve migration history and add reviewed forward migrations rather than deleting already-applied files.

## Add your domain

```bash
go run ./cmd/nexokit make module widgets --crud --migration --tenant
```

`widgets` is an illustrative new domain, not a shipped module. New scaffolds include `core/widgets/model.go`/`dtos.go`, module `core/contracts.go`/`errors.go`, handler + repository packages under `slices/<use_case>/`, and root `container.go`/`routes.go`. Without `--crud`, only get/list slices and routes are generated. The command scaffolds a new module; it does not transform other modules.

To adopt the generated module:

1. Define the model, validation and migration schema together, including tenant ownership/indexes for `--tenant`; generated SQL is a migration stub, not a deployable schema. The entity hook generates ULID public IDs; responses omit internal relationship/audit IDs.
2. Construct `widgets.NewContainer(db)` in the app composition root and explicitly call `widgets.Register(group, container)`. Generation never edits the app or automatically mounts routes. Choose authentication, permission guards and trusted tenant middleware before registration; configure GORM `TranslateError` to classify database constraint failures.
3. For tenant operations, select a company through trusted middleware, not JSON. Create/get/update/delete reject missing or global scope; lists alone permit explicitly trusted root-global scope. JSON `company_id` cannot select/reassign ownership and is not exposed in responses.
4. Test normalization, safe error responses, own/foreign-company reads and writes, soft deletion and root-global policy. Add a service only when business policy/orchestration warrants it; simple CRUD stays two layers.

See [architecture](architecture_and_vsa.md) and [CLI flags](references/runtime/cli.md).

Run focused tests, then relevant [quality checks](testing_and_quality.md#4-verification-commands). Build/deploy using the [runbook](deployment.md), not local development defaults.
