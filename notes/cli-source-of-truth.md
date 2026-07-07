# CLI Source of Truth

Canonical commands parsed from `Makefile` and `cmd/nexokit` (`internal/cli/commands/`).

## Makefile targets

`make dev`, `make build`, `make test`, `make test-unit`, `make test-integration`, `make test-coverage`, `make migrate-up`, `make migrate-down`, `make migrate-create`, `make migrate-status`, `make migrate-reset`, `make seed`, `make create-root`, `make fmt`, `make vet`, `make lint`, `make install-hooks`, `make check-env`.

## Direct `nexokit` commands

| Command | Usage |
|---------|-------|
| `serve` | `nexokit serve` |
| `config` | `nexokit config` |
| `status` | `nexokit status` |
| `migrate up/down/status/reset/create` | `nexokit migrate up`, `nexokit migrate create <name>` |
| `make module` | `nexokit make module <name> [--crud] [--migration] [--tenant]` |
| `make migration` | `nexokit make migration <name>` |
| `make seed` | `nexokit make seed <name>` |
| `create-root` | `nexokit create-root [--name] [--email] [--password] [--force]` |
| `seed` | `nexokit seed` |

## Notes

- `nexokit` is the internal developer CLI, not a global tool.
- Migration commands require `DATABASE_URL` or `DB_*` variables.
- `create-root` is idempotent.
