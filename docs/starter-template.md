# Adopt NexoKit as a starter

Fork or clone the repository into your namespace; it is a starting application, not a library dependency. Read [the application map](system-map.md) before changing composition.

## Quick path

1. Rename `github.com/enviniom/nexokit` in `go.mod` and all project import strings to your module path. `go mod edit -module your/module/path` changes only `go.mod`; use your editor's project-wide replacement for imports, generator templates and the seed runner's hardcoded import. Search for remaining old-module occurrences before building.
2. Update project identity in the root README and environment example (`APP_NAME`, URLs/domain and root-user defaults). Copy the example to local `.env` or inject environment values; keep secrets out of version control.
3. Configure PostgreSQL, a strong `PASETO_KEY`, and deliberate CORS/rate-limit settings. See [deployment configuration](deployment.md#configuration-and-secrets); local defaults are not production safeguards.
4. Start local dependencies and initialize from the repository root:

```bash
docker compose up -d
make migrate-up
make seed
make create-root
make dev
```

5. Check `/health/live` and `/health/ready`, then exercise login and a protected route. Root user routes require `X-Company-ID`; see [request flow](references/runtime/request-flow.md).

These are commands for the adopter to run, not part of a documentation-only verification.

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
