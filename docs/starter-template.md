# Adopt NexoKit as a starter

Fork or clone the repository into your namespace; it is a starting application, not a library dependency. Read [the application map](system-map.md) before changing composition.

## Quick path

Prerequisites: Go 1.26+ and the project dependencies available to Go. PostgreSQL is
needed for the application and provisioning commands, **not** for `init`; Docker
Compose is an optional local dependency runner. Redis/Valkey is optional when
`CACHE_DRIVER=none`.

1. From a fresh local clone/fork root, preview, review the sorted paths, then apply:

```bash
go run ./cmd/nexokit init
# Or use explicit flags for automation:
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api
# --dry-run is an optional explicit preview; neither preview form writes files.
go run ./cmd/nexokit init --name mi-api --module example.com/team/mi-api --apply
go build ./...
go test ./...
```

Bare `init` asks for application display name and GitHub username. It derives
`github.com/user-name/my-api` from `My API` and `user-name`, shows the plan, then
asks `Apply this plan? [y/N]`. Only `yes`/`y` (case insensitive) applies; No,
blank or EOF cancels without writes. No GitHub repository is created, no network
lookup is performed, and no credentials are requested. The repository slug is
lowercase ASCII, with runs of spaces/dots/underscores/hyphens collapsed to a
hyphen. Unsupported characters and invalid GitHub usernames are rejected.

2. Application identity changes in the root README heading and example `APP_NAME`;
   Go module imports, embedded generator templates, seed runner and generated-test
   source change consistently. The developer tool remains `nexokit` in
   `cmd/nexokit`; NexoKit attribution is retained. Names with spaces are supported:
   use `--name "My API"`. This is one-time independent application ownership, not
   downloading, updating or synchronizing an upstream project.
3. Existing `.env` is left untouched. For a new local setup only, copy the updated
   `.env.example` to `.env`, or inject environment values. Review URLs/domains and
   root-user defaults yourself; keep secrets out of version control. Configure
   PostgreSQL, a strong `PASETO_KEY`, and deliberate CORS/rate-limit settings. See
   [deployment configuration](deployment.md#configuration-and-secrets); local
   defaults are not production safeguards. `init` generates no production secrets
   and makes no production-readiness guarantee.
4. Review and deliberately execute local provisioning from the repository root:

```bash
docker compose up -d
make migrate-up
make seed
make create-root
make dev
```

5. Check `/health/live` and `/health/ready`, then exercise login and a protected route. Root user routes require `X-Company-ID`; see [request flow](references/runtime/request-flow.md).

These are commands for the adopter to run. `init` never starts a database, runs
migrations/seeds, creates root accounts, executes Git operations or changes
`go.sum`/dependencies. Local private adoption does not select public distribution
license terms.

## Automatic CI

GitHub Actions runs tests, build, vet, formatting and module error-boundary checks
on pushes and pull requests to all branches; manual dispatch remains available.
Hosted execution requires a push or pull request (or manual dispatch on GitHub)
and has not been run locally. You control whether and when to publish; local
checks do not prove a hosted run succeeded.

## Initialization safety

Flag calls never prompt: both `--name` and `--module` are required, even with
`--dry-run`. Preview is the flag default; only `--apply` writes (or explicit Yes
in bare interactive mode), and `--apply` plus `--dry-run` is an
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
retrying; no automatic repair or repeated-init/upgrade framework is provided.

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
