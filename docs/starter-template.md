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
| auth | Credentials, access/refresh tokens and session handling |
| companies | Company/domain administration and tenant resolution |
| IAM | Users, roles, permissions and authenticated-user lookup |
| onboarding | Root-controlled creation of a company and its initial identities |

Removing a module is more than removing its folder: remove its imports, container fields/construction, injected dependencies and `RegisterModules` mounting in `internal/app/container.go`. Review `SyncPermissions`, tenant resolvers, auth lookup, CLI root creation, seeds, migrations, fixtures and tests. Auth relies on IAM user data; current private routing relies on companies; CLI root creation relies on IAM. Choose an alternate identity/tenant policy before removing those capabilities.

For a fresh project, adjust initial schema and seeds coherently. For an existing deployed project, preserve migration history and add reviewed forward migrations rather than deleting already-applied files.

## Add your domain

```bash
go run ./cmd/nexokit make module widgets --crud --migration --tenant
```

`widgets` is an illustrative new domain, not a shipped module. The generator currently emits flat legacy layer files, not a ready-to-mount vertical-slice container. The generator’s flat module-root `dto.go`/`model.go` output also needs adaptation to the [entity DTO and mapping convention](architecture_and_vsa.md#4-models-dtos-and-mapping); keep module contracts in `core/contracts.go`. This is adopter work, not a migration already performed in this checkout. Adapt the scaffold into use-case packages, review code/migration and compilation, construct a container, wire routes into the app, choose guards, and test tenant isolation. Default simple operations to two layers; add services for real orchestration. See [architecture](architecture_and_vsa.md) and [CLI flags](references/runtime/cli.md).

Run focused tests, then relevant [quality checks](testing_and_quality.md#4-verification-commands). Build/deploy using the [runbook](deployment.md), not local development defaults.
