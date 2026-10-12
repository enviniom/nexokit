# Deploy and recover NexoKit

Deploy a built API with PostgreSQL, a controlled environment and a TLS reverse proxy. Run schema changes explicitly; API bootstrap does not migrate the database. Adapt the starter first using [adoption](starter-template.md).

## Build and release layout

From the repository root, with the Go version declared in `go.mod`:

```bash
make build
# For a Linux target, where your build environment supports cross-compilation:
GOOS=linux GOARCH=amd64 make build
```

Ship `bin/api`, `bin/nexokit` and `migrations/` in a versioned release directory. Run the CLI with that directory as its working directory. Keep a previous binary/configuration snapshot and record schema version before rollout. Seeds additionally need source and Go; see [CLI](references/runtime/cli.md#seeds).

## Configuration and secrets

`internal/config/config.go` reads environment variables and optionally working-directory `.env`. Inject production secrets through a secret manager or restricted service environment file. Never commit `.env`, copy example credentials into production or put passwords in command arguments.

| Settings | Operator requirement |
|---|---|
| `APP_ENV`, `APP_PORT`, `APP_URL` | `production`, chosen listen port, actual public URL |
| `APP_PLATFORM_DOMAIN` | Base domain for onboarding's optional generated technical domain |
| `DATABASE_URL` or `DB_HOST/PORT/NAME/USER/PASSWORD` | Intended database; URL takes precedence |
| `DB_SSL_MODE` / URL TLS parameters | Require verified TLS appropriate to the deployment; local default is `disable` |
| `PASETO_KEY` | Strong random secret, consistently injected across instances; plan session disruption for rotation |
| `ACCESS_TTL_MINUTES`, `REFRESH_TTL_DAYS` | Explicit session lifetime policy (defaults 15 minutes / 7 days) |
| `ROOT_USER_NAME/EMAIL/PASSWORD` | Secure initial provisioning inputs; remove persistent provisioning credentials when no longer needed |
| `CORS_ALLOWED_ORIGINS` | Explicit trusted origins, not default `*` |
| `SHUTDOWN_TIMEOUT_SECONDS` | Grace period (default 30); supervisor grace must accommodate it |
| `LOG_LEVEL/FORMAT`, `LOG_FILE/ERROR_FILE/GIN_FILE`, `LOG_MAX_SIZE/BACKUPS/AGE`, `LOG_COMPRESS` | Writable restricted log directory, rotation and aggregation |
| `CACHE_DRIVER`, `REDIS_HOST/PORT/PASSWORD/DB`, `REDIS_DIAL_TIMEOUT_SECONDS` | `none` or `redis`; understand [cache/readiness limits](references/infrastructure/cache.md) |
| `RATE_LIMIT_ENABLED/DRIVER/GLOBAL_RPM/LOGIN_RPM/REFRESH_RPM/WINDOW_SECONDS/CLEANUP_INTERVAL_MINUTES` | Review [effective quota limits](references/infrastructure/rate-limiting.md) before relying on defaults |

**Operational security caveat:** Configuration loading **does not enforce** secret strength, mandatory production TLS/CORS or a nonempty PASETO key. The token manager hashes whatever key string is supplied. These are operator requirements, not boot-time guarantees. `nexokit config` is a masked partial display, not a full security check.

## Migrate and provision

1. Verify destination DB and a recent **tested/restorable** backup. Review both Up and Down SQL and schema compatibility with old/new binaries.
2. Inspect and apply migrations from the release root:

```bash
./bin/nexokit migrate status
./bin/nexokit migrate up
```

`status` lists migration state; it is not a dry run of SQL. Run migrations once through a controlled admin job before starting incompatible binaries.

3. For first installation only, run reviewed required seeds and provision root:

```bash
./bin/nexokit seed
./bin/nexokit create-root
```

Seeding needs source, Go, writable `.tmp/` and DB access; run it from a suitable admin/build environment if production ships binaries only. Root creation requires the seeded root role; normal non-local execution asks for confirmation. Avoid concurrent root creation and avoid `--password` on shared command lines. The interactive password prompt is not hidden.

**Destructive warning:** `migrate down` reverts one latest migration; `migrate reset` executes all Down migrations. Neither command has a production safety gate. Require explicit operational approval, backup verification and a data-recovery plan; never run reset as routine recovery.

## TLS and supervision

Run `./bin/api` under systemd, a container supervisor or an equivalent service manager. Set the working directory, environment, restricted service user, restart policy, resource limits and writable log paths. The API listens on `:APP_PORT`; firewall the backend so only the proxy/internal network can reach it.

Terminate TLS at a trusted reverse proxy/load balancer. Preserve intended Host and overwrite client-supplied `X-Forwarded-For` / `X-Forwarded-Proto`; do not merely append untrusted values. **Operational security caveat:** The limiter reads the first forwarded IP directly, and optional public-domain redirects trust forwarded scheme. Enforce proxy trust at the network/proxy layer. Restrict CORS, protect admin endpoints and keep dependency services private.

On SIGTERM/interrupt the API attempts graceful shutdown with the configured timeout. Monitor shutdown failures; `App.Stop` returns on the first component failure and does not guarantee every later resource was closed.

## Health and observation

| Endpoint | Meaning |
|---|---|
| `GET /health` | Basic success envelope |
| `GET /health/live` | Process responds; does not probe dependencies |
| `GET /health/ready` | DB ping and optional cache read, five-second request timeout; 200 or 503 |

**Operational readiness caveat:** Readiness exposes dependency error strings; keep it on internal monitoring paths rather than public diagnostics. Cache misses currently count as readiness errors, including Redis-to-no-op fallback with cache configured enabled. Do not promise successful readiness on healthy empty caches.

After rollout, check probes **and** a real authenticated/tenant-scoped request; watch error rate, latency and database/log signals against your baseline. Route traffic only when the checks fit your actual cache configuration. The CLI `status` exit code is not a health guarantee. Forward rotating application, error and access logs to the chosen monitoring system.

## Rollback

1. Stop or drain the failed release; preserve logs and identify whether data/schema changed.
2. Prefer the previous binary only if compatible with the current schema. Do not automatically reverse migrations just because code failed.
3. If schema reversal is necessary, review Down SQL, approval and tested backup restoration first. `migrate down` handles one migration; multiple release migrations require an explicit plan. Restore from backup when Down SQL cannot recover removed data.
4. Restore the known-good binary and compatible environment, restart, verify probes and authenticated tenant behavior, then restore traffic.

Prefer fix-forward when safer than destructive schema reversal. Never confuse restoring binaries with restoring deleted data.
