# Using NexoKit as a Starter Template

NexoKit is designed to be cloned or forked as the starting point for a new SaaS API project, not installed as a library.

## Adopt the repo

1. **Fork or clone** the repository into your own namespace.
2. **Rename the Go module** from `github.com/enviniom/nexokit` to your module path:

   ```bash
   find . -type f -name '*.go' -exec sed -i 's|github.com/enviniom/nexokit|your/module/path|g' {} +
   go mod edit -module your/module/path
   ```

3. **Update `README.md`**, `docs/README.md`, and `docs/starter-template.md` to describe your project.
4. **Replace example values** in `.env.example`:
   - `APP_NAME`
   - `APP_URL`
   - default database credentials
   - `ROOT_USER_EMAIL`

## Configure your environment

```bash
cp .env.example .env
# edit .env with your database, secrets, and rate-limit settings
```

## Choose your modules

The starter ships with four modules:

| Module | Keep if you need... |
|--------|---------------------|
| `auth` | Login, tokens, sessions |
| `companies` | Multi-company / tenant support |
| `iam` | Users, roles, permissions |
| `onboarding` | Public company sign-up flow |

Remove modules you do not need by deleting their folder under `internal/modules/` and removing the corresponding wiring from `internal/app/container.go`.

## Run the project

```bash
docker compose up -d
make migrate-up
make seed
make create-root
make dev
```

## Build your domain

1. Add migrations to `migrations/`.
2. Generate a module scaffold with `nexokit make module <name> --crud --migration --tenant`.
3. Implement slice `handler.go`, `service.go`, and `repository.go`.
4. Wire the new module in `internal/app/container.go`.
5. Add tests and run `make test`.

## Deploy

See [`deployment.md`](deployment.md) for the production runbook.
