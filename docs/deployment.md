# NexoKit Deployment Guide

This guide covers running NexoKit-based applications in production. If you are adapting the starter first, read [`starter-template.md`](starter-template.md). You need a server with Go 1.26+ (or a pre-built Linux binary), PostgreSQL 15+, and a reverse proxy for TLS termination. Redis/Valkey is optional (`CACHE_DRIVER=none` works without it). For local development, see the root [`README.md`](../README.md).

## Build

```bash
make build
```

This produces:

- `bin/api` — the HTTP server.
- `bin/nexokit` — the operations CLI.

For a reproducible build, pin `GOOS` and `GOARCH`:

```bash
GOOS=linux GOARCH=amd64 make build
```

## Environment configuration

Copy `.env.example` to `.env` and set production values:

```bash
cp .env.example .env
```

### Required settings

| Variable | Purpose |
|----------|---------|
| `APP_ENV` | `production` |
| `APP_PORT` | Port the API binds to (e.g., `8080`) |
| `APP_URL` | Public URL, e.g., `https://api.example.com` |
| `DATABASE_URL` **or** `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | PostgreSQL connection |

### Security-critical settings

| Variable | Guidance |
|----------|----------|
| `PASETO_KEY` | 32-byte symmetric key. Generate in production with a CSPRNG and rotate only with a logout window. |
| `ROOT_USER_PASSWORD` | Change from the example value before running `create-root`. |
| `DB_SSL_MODE` | Use `require`, `verify-ca`, or `verify-full` in production; avoid `disable`. |
| `CORS_ALLOWED_ORIGINS` | Set explicit origins instead of `*`. |

### Logging settings

| Variable | Purpose |
|----------|---------|
| `LOG_LEVEL` | `info` or `warn` in production |
| `LOG_FORMAT` | `json` recommended for log aggregation |
| `LOG_FILE` / `LOG_ERROR_FILE` / `LOG_GIN_FILE` | Application, error-only, and HTTP access log paths |
| `LOG_MAX_*` / `LOG_COMPRESS` | Log rotation settings |

### Optional cache and rate limiting

Set `CACHE_DRIVER=none` to run without Redis/Valkey, or configure `REDIS_*` and set `CACHE_DRIVER=redis`. Keep `RATE_LIMIT_ENABLED=true` in production.

## Database setup

### SSL

Require TLS for PostgreSQL connections in production. The connection string or `DB_SSL_MODE` should enforce encryption. If you use self-signed certificates, mount the CA file and set the connection parameters accordingly.

### Migrations

Run migrations before starting the API:

```bash
./bin/nexokit migrate up
```

For a dry-run view:

```bash
./bin/nexokit migrate status
```

### Danger Zone — destructive migration commands

The following commands delete or rollback data. They are forbidden by operational policy in production unless explicitly approved. The CLI binary does not block these commands; enforcement is the operator's responsibility. Use them only with a verified backup and explicit approval.

| Command | Risk | When to use |
|---------|------|-------------|
| `migrate down` | Rolls back the last migration batch; may drop columns/tables/indices | Emergency rollback only after backup verification |
| `migrate reset` | Rolls back **all** migrations; effectively empties the schema | Non-production environments or disaster recovery with signed approval |

```bash
# Emergency rollback only. Verify a fresh backup first.
./bin/nexokit migrate down

# Destructive. Never run in production unless approved as disaster recovery.
./bin/nexokit migrate reset
```

## Seeds and root user

Load any required seeds:

```bash
./bin/nexokit seed
```

> **Seeding needs Go on the host.** `nexokit seed` generates a temporary Go runner and executes it with `go run`, so the host performing seeding must have the Go toolchain. Safe alternatives:
> - Run `nexokit seed` from a build or admin environment that has Go and database access.
> - Keep Go installed on the production host that performs the one-time seed step.
> Do not pass database passwords on the command line; use the environment file or a secret manager.

Create the initial root user:

```bash
./bin/nexokit create-root
```

The command reads `ROOT_USER_NAME`, `ROOT_USER_EMAIL`, and `ROOT_USER_PASSWORD` from the environment. In non-local environments it asks for interactive confirmation unless `--force` is passed. Prefer interactive prompts or a secret manager; avoid `--password` on shared command lines.

## Running the API

```bash
./bin/api
```

The server binds to `APP_PORT` and writes structured logs to the configured files.

### Reverse proxy / TLS

Do not expose `bin/api` directly to the internet. Place it behind a reverse proxy (nginx, Caddy, Traefik, or a cloud load balancer) that:

- Terminates TLS.
- Sets `X-Forwarded-For` and `X-Forwarded-Proto` if the app consumes them.
- Forwards traffic to `http://localhost:APP_PORT`.

### Process supervision

Use systemd, Docker, Kubernetes, or a process manager to keep the binary running. Point the service at the binary, `.env` file, and log directory.

## Logging and monitoring

Log files rotate based on `.env` values. Forward logs to your aggregation stack (e.g., Loki, Datadog, CloudWatch, Splunk) from:

- `LOG_FILE` — structured application logs.
- `LOG_ERROR_FILE` — ERROR-level and above.
- `LOG_GIN_FILE` — HTTP access logs.

Health endpoints are available unversioned:

```text
GET /health
GET /health/live
GET /health/ready
```

### Alert thresholds

Use these thresholds as starting points; tune them to your traffic baseline.

| Signal | Investigate | Emergency | All-hands |
|--------|-------------|-----------|-----------|
| HTTP error rate (5xx / total) | > 1% | > 2% | > 5% |
| p95 latency | above project-defined budget | > 2× budget | > 5× budget |
| p99 latency | above project-defined budget | > 2× budget | > 5× budget |
| Readiness probe failures | 2 consecutive | 5 consecutive | 10 consecutive |

**Rollback vs. fix-forward criteria:**
- Roll back when error rate breaches emergency threshold and the cause is unknown or the last deployment changed the schema or auth path.
- Fix-forward when the failure is narrow, well-understood, and a patch can ship faster than a rollback without data loss.
- Always verify the database is backed up before running `migrate down` or `migrate reset`.

## Operational checklist

- [ ] `.env` is not committed; secrets are injected by the host or secret manager.
- [ ] `APP_ENV=production` and a fresh 32-byte `PASETO_KEY` are set.
- [ ] PostgreSQL SSL is enabled and `CORS_ALLOWED_ORIGINS` is restricted.
- [ ] Migrations run before the API starts; seeds and root user run once after migrations.
- [ ] A reverse proxy terminates TLS; logs are rotated and forwarded.
- [ ] Health checks are monitored and PostgreSQL is backed up.

## Rollback

If a deployment fails:

1. Stop the API service.
2. If the schema changed, roll back the migration only after verifying a fresh backup: `./bin/nexokit migrate down` (see [Danger Zone](#danger-zone--destructive-migration-commands)).
3. Restore the previous binary and environment.
4. Restart the service.

For a documentation-only rollback, revert the docs PR with `git revert`.
