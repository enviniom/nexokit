# Tasks: Testing Strategy, CI & Developer Guide

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1200 (across 4 PRs) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (CI infra) → PR 2 (helpers) → PR 3 (integration) → PR 4 (docs) |
| Delivery strategy | force-chained (auto-chain) |
| Chain strategy | feature-branch-chain |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Makefile targets + GitHub Actions CI | PR 1 `feat/testing-1-ci-infra` | Targets main; ~100 lines |
| 2 | Test helpers (DB, auth, fixtures) | PR 2 `feat/testing-2-helpers` | Targets PR 1 branch; ~300 lines |
| 3 | Integration test suites (auth, users, tenant, rbac, health) | PR 3 `feat/testing-3-integration` | Targets PR 2 branch; ~600 lines |
| 4 | `docs/testing.md` developer guide | PR 4 `feat/testing-4-docs` | Targets PR 3 branch; ~200 lines |

## Phase 1: CI Infrastructure (PR 1 — targets `main`)

- [x] 1.1 Add `.PHONY` entries and targets `test-unit`, `test-integration`, `test-coverage` to `Makefile` (extend existing `.PHONY` line; keep `test` backward-compatible)
- [x] 1.2 Create `.github/workflows/ci.yml` with jobs: `test` (`go test ./...`), `vet` (`go vet ./...`), `fmt-check` (`gofmt -l .`), triggered on push to `main` and all `pull_request` events
- [x] 1.3 Verify `make test-unit` skips integration tests (uses `-short` flag)
- [x] 1.4 Verify `make test-coverage` produces `coverage.out` and prints summary to stdout

## Phase 2: Test Helpers & Fixtures (PR 2 — targets PR 1 branch)

- [ ] 2.1 Create `tests/helpers/database.go` with `NewSQLiteDB(t *testing.T, models ...any) *gorm.DB` — opens SQLite `:memory:`, auto-migrates models, registers `t.Cleanup` to close connection
- [ ] 2.2 Create `tests/helpers/auth.go` with `SeedUser(t, db, opts)`, `CreateTestToken(t, db, user)`, `AuthenticatedRequest(t, method, path, body, actor) *http.Request` helpers
- [ ] 2.3 Create `tests/helpers/fixtures.go` with factory functions: `SeedRole`, `SeedCompany`, `SeedUserWithRole`, `SeedPermission` — deterministic, relationship-aware
- [ ] 2.4 Create `tests/fixtures/` directory with Go fixture builder files (not static JSON/YAML)
- [ ] 2.5 Write table-driven unit tests for each helper in `tests/helpers/*_test.go` — verify DB isolation, token generation, fixture determinism

## Phase 3: Integration Test Suites (PR 3 — targets PR 2 branch)

- [ ] 3.1 Create `tests/integration/auth_test.go` — login success (200), invalid credentials (401), inactive user, valid refresh token, revoked refresh token, logout revocation; use `testing.Short()` skip gate
- [ ] 3.2 Create `tests/integration/users_test.go` — CRUD endpoints with authenticated requests; list isolation, create, update, delete; seed fixtures via helpers
- [ ] 3.3 Create `tests/integration/tenant_test.go` — admin cannot access other company data (403/empty), root global access, root scoped access; two-company setup
- [ ] 3.4 Create `tests/integration/rbac_test.go` — user with permission accesses resource, user without permission gets 403, unauthenticated gets 401, root has all permissions
- [ ] 3.5 Create `tests/integration/health_test.go` — GET health endpoint returns 200 with healthy status
- [ ] 3.6 Ensure all integration tests use `NewSQLiteDB` helper, `t.Cleanup` for cleanup, and `gin.SetMode(gin.TestMode)`
- [ ] 3.7 Run `make test-integration` and verify all suites pass; run `make test-unit` and verify integration tests are skipped

## Phase 4: Documentation (PR 4 — targets PR 3 branch)

- [ ] 4.1 Create `docs/testing.md` with: Quick path (running tests), folder conventions (unit vs integration), Makefile targets table, Go testing patterns (table-driven, `t.Run`, `t.TempDir`, `testing.Short()`), integration test guidelines (helpers, fixtures, cleanup), SQLite/PostgreSQL limits, Redis skip behavior, coverage policy, conventions checklist
- [ ] 4.2 Include code examples matching existing codebase patterns (stdlib-only, no testify)
- [ ] 4.3 Document CI workflow: what each check validates, how to reproduce failures locally
