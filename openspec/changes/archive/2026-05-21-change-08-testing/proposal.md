# Proposal: Testing Strategy, CI & Developer Guide for NexoKit

## Intent

NexoKit has 59 `_test.go` files with solid unit-test coverage but lacks CI automation, granular Makefile targets, integration test infrastructure, and a developer guide. This change closes the gap between "tests exist" and "tests are reliable, automatable, and maintainable by future developers."

## Scope

### In Scope
- Makefile targets: `test-unit`, `test-integration`, `test-coverage`
- GitHub Actions CI workflow (test, vet, fmt check on push/PR)
- Test helpers: `tests/helpers/database.go`, `auth.go`, `fixtures.go`
- `tests/fixtures/` directory with factory helpers
- Integration tests: auth, users CRUD, tenant isolation, RBAC, health
- `docs/testing.md` — developer guide with testing guidelines and conventions
- Docker Compose test infrastructure configuration

### Out of Scope
- Rewriting existing unit tests or changing test patterns
- Adding testify as a dependency (keep stdlib-only; testify optional per exploration)
- PostgreSQL-based integration tests (SQLite `:memory:` for now, per existing pattern)
- Testcontainers-go adoption
- Coverage threshold enforcement (starts at 0%, raised incrementally later)

## Capabilities

### New Capabilities
- `testing-ci`: GitHub Actions workflow, Makefile test targets, coverage reporting
- `testing-integration`: Integration test infrastructure, helpers, fixtures, and test suites
- `testing-docs`: Developer guide and testing conventions documentation

### Modified Capabilities
- `dev-environment`: Makefile gains `test-unit`, `test-integration`, `test-coverage` targets (extends existing `test` target)

## Approach

Incremental expansion building on established patterns: table-driven tests, fake interfaces for DI, `gin.SetMode(gin.TestMode)`, `httptest`, SQLite `:memory:` for integration, `testing.Short()` for skip gates. No new testing frameworks. Documentation is a first-class deliverable — `docs/testing.md` is a contract for maintainability, not decoration.

## Chained PR Strategy (force-chained, 800-line budget)

| PR | Branch | Deliverable | Est. Lines |
|----|--------|-------------|------------|
| 1 | `feat/testing-1-ci-infra` | Makefile targets + GitHub Actions CI | ~150 |
| 2 | `feat/testing-2-helpers` | Test helpers + fixtures infrastructure | ~250 |
| 3 | `feat/testing-3-integration` | Integration tests (auth, users, tenant, rbac, health) | ~500 |
| 4 | `feat/testing-4-docs` | `docs/testing.md` developer guide | ~200 |

Each PR targets the previous PR's branch (Feature Branch Chain). PR 1 targets `main`.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `Makefile` | Modified | Add test-unit, test-integration, test-coverage targets |
| `.github/workflows/ci.yml` | New | CI workflow: test, vet, fmt check |
| `tests/helpers/database.go` | New | Test DB setup helper (SQLite + GORM) |
| `tests/helpers/auth.go` | New | Token/user creation helpers for integration tests |
| `tests/helpers/fixtures.go` | New | Factory/fixture loading helpers |
| `tests/fixtures/` | New | Directory for test data factories |
| `tests/integration/auth_test.go` | New | Auth integration tests |
| `tests/integration/users_test.go` | New | Users CRUD integration tests |
| `tests/integration/tenant_test.go` | New | Tenant isolation integration tests |
| `tests/integration/rbac_test.go` | New | RBAC HTTP access integration tests |
| `tests/integration/health_test.go` | New | Health endpoint integration tests |
| `docs/testing.md` | New | Developer guide with testing guidelines |
| `docker-compose.yml` | Modified | Test-specific service configuration |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| SQLite vs PostgreSQL dialect differences | Medium | Use GORM-compatible types; flag PostgreSQL-specific features with `testing.Short()` skip |
| Test isolation with shared in-memory DB | Medium | Each integration test uses separate DB instance or cleans up via `t.Cleanup()` |
| CI flakiness from timing-dependent tests | Low | Avoid time-dependent assertions; use table-driven with explicit expected values |
| PR 3 (integration tests) exceeds 800-line budget | Medium | Split into two PRs if needed: auth+users first, tenant+rbac+health second |
| testify debate slows review | Low | Keep stdlib-only; document rationale in `docs/testing.md` |

## Rollback Plan

All changes are additive — no existing code is modified except Makefile (new targets appended) and `docker-compose.yml` (optional test service). To rollback:
1. Delete `.github/workflows/ci.yml`
2. Remove new Makefile targets (revert to previous `Makefile`)
3. Delete `tests/helpers/database.go`, `auth.go`, `fixtures.go`
4. Delete `tests/fixtures/` directory
5. Delete all new `tests/integration/*_test.go` files
6. Delete `docs/testing.md`
7. Revert `docker-compose.yml` changes

No database migrations or schema changes involved. Zero production impact.

## Dependencies

- None. Builds on existing test infrastructure (59 `_test.go` files already pass).
- GitHub repository must have Actions enabled (standard for GitHub repos).

## Success Criteria

- [ ] `make test-unit` runs all unit tests (package-level `_test.go` files)
- [ ] `make test-integration` runs integration tests (skippable with `-short`)
- [ ] `make test-coverage` produces coverage report
- [ ] GitHub Actions CI passes on push and PR
- [ ] All integration tests pass with SQLite `:memory:`
- [ ] `docs/testing.md` exists and covers: test structure, patterns, running tests, adding new tests, conventions
- [ ] No new dependencies added (stdlib-only testing)
- [ ] Each chained PR stays under 800-line review budget
