# Design: Testing Strategy, CI & Developer Guide

## Technical Approach

Expand the existing Go testing foundation without changing runtime architecture or adding test frameworks. Unit tests stay beside code; `/tests` remains for integration, helpers, fixtures, and database setup. New CI/Makefile entry points orchestrate the current suite plus isolated integration tests. `docs/testing.md` becomes the developer-facing contract for running, writing, and reviewing tests.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Testing API | stdlib `testing`, `httptest`, table-driven tests | Add `testify` | Existing 59 test files are stdlib-only; no new dependency or assertion DSL is needed. |
| Integration DB | SQLite `:memory:` per test/helper call | PostgreSQL-only tests, Testcontainers | Matches current repository tests and keeps CI fast. PostgreSQL-specific behavior is documented as a limit and deferred. |
| External Redis/Valkey tests | Skip in `testing.Short()` and `t.Skipf` when unavailable | Require Redis in every CI run | Existing Redis cache/rate-limit tests already use skip gates; CI must remain reliable without services. |
| Coverage policy | Generate reports, threshold remains 0% | Enforce global threshold now | Current coverage is uneven by package; enforcement should be raised incrementally after baseline stabilization. |
| Chained rollout | Feature Branch Chain, 4 PRs | Single PR or stacked-to-main | User selected `force-chained`; each slice remains under the 800-line review budget. |

## Data Flow

```text
developer/CI ──→ Makefile targets ──→ go test packages
                         │
                         ├─ unit: package-local *_test.go
                         ├─ integration: tests/integration + tests/helpers
                         └─ coverage: coverprofile + report

integration test ──→ helper DB/router/auth/fixtures ──→ module handlers/services/repos
                         │
                         └─ isolated SQLite :memory: DB per test
```

## File Changes

| File | Action | Description |
|---|---|---|
| `Makefile` | Modify | Add `.PHONY` and targets: `test-unit`, `test-integration`, `test-coverage`; keep `test` as the full suite. |
| `.github/workflows/ci.yml` | Create | Run checkout, Go setup, fmt check, vet, and tests on push/PR. |
| `tests/helpers/database.go` | Create | Open SQLite `:memory:`, auto-migrate requested models, register `t.Cleanup`. |
| `tests/helpers/auth.go` | Create | Build authenticated requests/contexts/tokens for integration scenarios without duplicating setup. |
| `tests/helpers/fixtures.go` | Create | Shared factories for roles, companies, users, permissions, and refresh tokens. |
| `tests/fixtures/*.go` | Create | Reusable fixture builders, not static mutable data. |
| `tests/integration/{auth,users,tenant,rbac,health}_test.go` | Create | Cover real HTTP/module paths for auth, users CRUD, tenant isolation, RBAC, and health. |
| `docker-compose.yml` | Modify | Add clearly named test DB/Redis configuration only if needed for optional external-service runs. |
| `docs/testing.md` | Create | Quick path, commands, folder conventions, patterns, SQLite/Postgres limits, Redis skips, coverage policy. |

## Interfaces / Contracts

Helpers should expose small testing contracts, for example:

```go
func NewSQLiteDB(t *testing.T, models ...any) *gorm.DB
func SeedUser(t *testing.T, db *gorm.DB, opts UserOptions) users.User
func AuthenticatedRequest(t *testing.T, method, path string, body io.Reader, actor Actor) *http.Request
```

No production interfaces change.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Existing packages, helpers, fixtures | Table-driven stdlib tests beside code or helper package. |
| Integration | Auth, users CRUD, tenant, RBAC, health | `httptest`, Gin test mode, isolated SQLite DB per test; avoid `t.Parallel` when sharing DB/router state. |
| External service | Redis cache/rate-limit behavior | Keep `testing.Short()` skips and skip when Redis/Valkey is unavailable. |
| CI | Formatting, vet, unit/integration suite | GitHub Actions with deterministic `go test ./...`; optional service tests remain local/manual. |

## Migration / Rollout

No migration required. Feature Branch Chain:
1. `feat/testing-1-ci-infra`: Makefile + CI.
2. `feat/testing-2-helpers`: DB/auth/fixture helpers.
3. `feat/testing-3-integration`: integration suites; split if it nears 800 changed lines.
4. `feat/testing-4-docs`: `docs/testing.md`.

## Open Questions

None.
