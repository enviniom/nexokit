# Exploration: change-08-testing — Testing Strategy & CI for NexoKit

## Current State

NexoKit already has a **substantial testing foundation** — far more than a bare starter. The codebase has **59 `_test.go` files** across all major areas, with all tests passing in short mode. Key findings:

### What Already Exists

| Area | Test Files | Coverage | Notes |
|------|-----------|----------|-------|
| `platform/response` | 1 | 92.7% | Comprehensive: success, error, paginated, validation, HandleError table-driven |
| `platform/apperror` | 1 | 90.2% | Wrap, Status, PublicMessage, sentinels — well covered |
| `platform/query` | 1 | 100% | Pagination, filters, sort, search from Gin context |
| `platform/validator` | 2 | 100% | Rules + validator core — fully covered |
| `platform/token` | 1 | 91.3% | PASETO issue/parse, expiry, tampering, refresh tokens |
| `platform/password` | 1 | 74.4% | Hash/verify table-driven, unique hashes |
| `platform/identity` | 1 | — | ULID generation, uniqueness, sortability |
| `platform/tenant` | 1 | 91.7% | ApplyTenantScope with GORM dry-run, Gin context round-trip |
| `middleware/auth` | 1 | — | Fake parser/lookup, valid token, expired, inactive user |
| `middleware/authorization` | 1 | — | AttachPermissions, RequirePermission, RequireRole with table-driven |
| `middleware/tenant` | 1 | — | AllowRootGlobalScope, RequireTenantScope, PublicTenant — comprehensive |
| `middleware/rate_limit` | 1 | — | LocalLimiter, RedisLimiter integration, HTTP middleware with spy |
| `middleware/logger/recovery/cors/request_id` | 4 | — | Basic middleware tests exist |
| `infra/cache/noop` | 1 | — | NoopCache operations |
| `infra/cache/redis` | 1 | 11.5% | Full Redis integration test with table-driven cases, `testing.Short()` skip |
| `modules/auth/handler` | 1 | 68.9% | Login, refresh, logout, me with fake service |
| `modules/auth/service` | 1 | — | Service-level tests exist |
| `modules/auth/repository` | 1 | — | Repository tests exist |
| `modules/users` | handler, service, repository, routes, dto | 77.5% | Full CRUD test coverage |
| `modules/companies` | handler, service, repository, migration | 68.8% | Full coverage |
| `modules/roles` | handler, service, routes, dto | 70.9% | Good coverage |
| `modules/permissions` | service, routes | 29.8% | Lower coverage — mostly routes |
| `server` | 2 | 96.2% | Health endpoints, router, start/stop |
| `config` | 2 | 85.0% | Config loading and env parsing |
| `cli/*` | 7 | 58-95% | Commands, generator, root, templates — golden file tests |
| `tests/integration` | 1 | — | Users isolation with SQLite in-memory, httptest |
| `tests/cli` | 1 | — | Golden file test for module generator |
| `tests/docs` | 1 | — | Multitenancy guide content verification |
| `seeds` | 2 | 59.3% | Permissions and roles seed tests |

### What's Missing (per the prompt's acceptance criteria)

1. **Makefile test targets**: Only `make test` exists (`go test ./...`). Missing:
   - `make test-unit`
   - `make test-integration`
   - `make test-coverage`

2. **GitHub Actions CI**: No `.github/workflows/` directory exists at all.

3. **Test helpers**: Only `tests/helpers/app.go` exists (full bootstrap). Missing:
   - `tests/helpers/database.go` — test DB setup helper
   - `tests/helpers/auth.go` — token/user helpers for integration tests
   - `tests/helpers/fixtures.go` — factory/fixture helpers

4. **Fixtures directory**: `tests/fixtures/` does not exist.

5. **Integration tests**: Only `tests/integration/users_isolation_test.go` exists. Missing:
   - `auth_test.go` — login, refresh, logout with real DB
   - `users_test.go` — CRUD with real DB
   - `tenant_test.go` — tenant isolation scenarios
   - `rbac_test.go` — permission-based access over HTTP
   - `health_test.go` — health endpoints with real dependencies

6. **Docker Compose for test infrastructure**: `docker-compose.yml` exists with PostgreSQL + Valkey, but no test-specific compose or test database configuration.

7. **Documentation**: No testing guide exists under `docs/`.

### Patterns Already Established

- **Table-driven tests** are the dominant pattern (used in response, apperror, query, token, password, middleware, etc.)
- **Fake interfaces** for dependency injection in tests (fakeAccessParser, fakeAuthUserLookup, fakePermissionResolver, fakeCompanyResolver, fakeAuthService, fakeDB, fakeCache)
- **`gin.SetMode(gin.TestMode)`** used consistently
- **`httptest.NewRecorder()` + `httptest.NewRequest()`** for HTTP tests
- **`t.Helper()`** used in test helpers
- **`t.TempDir()`** used in CLI golden tests
- **`testing.Short()`** for skipping integration tests (Redis)
- **SQLite `:memory:`** for integration tests that need a real DB but not PostgreSQL
- **GORM DryRun** for testing SQL generation without hitting a database
- **No testify dependency** — pure stdlib `testing` package
- **Package-internal tests** (`package auth`) for accessing unexported functions
- **External package tests** (`package cli_test`, `package integration_test`) for black-box testing

## Affected Areas

- `Makefile` — add test-unit, test-integration, test-coverage targets
- `.github/workflows/` — create CI workflow (new directory)
- `tests/helpers/` — add database.go, auth.go, fixtures.go
- `tests/fixtures/` — create directory with factory helpers
- `tests/integration/` — add auth, users, tenant, rbac, health tests
- `docs/testing.md` — create testing guide (new file, per user requirement)
- `docker-compose.yml` — potentially add test-specific service or test DB config
- `go.mod` — may need testify if adopted (currently not present)

## Approaches

### 1. Incremental Expansion (Recommended)
Build on existing patterns: add Makefile targets, CI workflow, test helpers, fixtures directory, and integration tests following the established table-driven + fake interface patterns.

- **Pros**: Respects existing code conventions, minimal disruption, each piece independently valuable
- **Cons**: Requires careful coordination to not break existing tests
- **Effort**: Medium-High (many test files + CI + docs)

### 2. Full Rewrite of Test Infrastructure
Replace current approach with a more opinionated framework (testify, testcontainers-go, etc.)

- **Pros**: More powerful assertions, easier container management
- **Cons**: Breaks existing patterns, adds dependencies, contradicts "no heavy frameworks" principle from prompt
- **Effort**: High

### 3. Minimal Additions
Only add what's explicitly missing: Makefile targets + CI workflow + docs

- **Pros**: Fastest path to acceptance criteria
- **Cons**: Leaves integration test gaps, no helper infrastructure
- **Effort**: Low

## Recommendation

**Approach 1 — Incremental Expansion** is the clear choice. The codebase already has excellent testing patterns; we need to:

1. **Add Makefile targets** (`test-unit`, `test-integration`, `test-coverage`) — these are straightforward
2. **Create GitHub Actions CI** — basic workflow with `go test`, `go vet`, `go fmt` check
3. **Add test helpers** — database setup, auth helpers, fixtures in `tests/helpers/` and `tests/fixtures/`
4. **Add integration tests** — auth, users, tenant, rbac, health using SQLite in-memory (consistent with existing `users_isolation_test.go`)
5. **Create `docs/testing.md`** — comprehensive guide for future developers (per user requirement)

### Chained PR Strategy (force-chained, 800-line review budget)

Given the `force-chained` strategy and 800-line budget, this change should be split into at least 3-4 PRs:

1. **PR 1: Infrastructure** — Makefile targets + GitHub Actions CI + Docker Compose test config
2. **PR 2: Test Helpers & Fixtures** — `tests/helpers/database.go`, `auth.go`, `fixtures.go` + `tests/fixtures/`
3. **PR 3: Integration Tests** — auth, users, tenant, rbac, health integration tests
4. **PR 4: Documentation** — `docs/testing.md` guide

Each PR should include its own tests/docs and stay under the review budget.

## Risks

1. **Redis integration tests require running Redis** — the existing `redis_test.go` uses `testing.Short()` to skip. Integration tests that need Redis should follow the same pattern.
2. **SQLite vs PostgreSQL differences** — existing integration tests use SQLite `:memory:`. GORM dialect differences (e.g., JSON columns, specific PostgreSQL types) could cause tests to pass with SQLite but fail with PostgreSQL. Consider adding a PostgreSQL-based integration test path.
3. **Test isolation** — the existing `users_isolation_test.go` seeds data directly into the same in-memory DB. Multiple integration tests running in parallel could conflict if they share the same DB instance. Each test should use its own DB or clean up properly.
4. **Coverage threshold** — currently no coverage threshold is enforced. Setting one too high could block PRs; too low is meaningless. Recommend starting at 0% (as configured in openspec) and raising gradually.
5. **go.mod dependency** — the prompt suggests `testify` as optional. The codebase currently has zero testify usage. Adding it would be a new dependency. Recommendation: keep stdlib-only for now, add testify only if a specific test genuinely benefits.

## Ready for Proposal

**Yes.** The codebase has a strong existing testing foundation with clear patterns. The missing pieces are well-defined: Makefile targets, CI workflow, test helpers/fixtures, integration tests, and documentation. The exploration provides enough context for the spec and design phases to proceed.
