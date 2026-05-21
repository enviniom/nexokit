# NexoKit Testing Guide

This guide is the source of truth for running, writing, and reviewing tests in NexoKit. It explains the project testing structure, the required conventions, and how to reproduce CI failures locally.

## Quick path

1. Run unit-focused checks first:

   ```bash
   make test-unit
   make vet
   make fmt
   ```

2. Run integration suites when your change touches handlers/services/repos/auth:

   ```bash
   make test-integration
   ```

3. Before opening a PR, run the same baseline as CI:

   ```bash
   go test ./...
   go vet ./...
   gofmt -l .
   ```

4. If needed, generate coverage:

   ```bash
   make test-coverage
   ```

## Test layout and naming

| Location | Purpose | Naming convention |
|---|---|---|
| `internal/**/**_test.go` | Unit tests close to code under test | `*_test.go` |
| `tests/helpers/*` | Shared setup helpers for integration tests | helper names by concern (`database.go`, `auth.go`, `fixtures.go`) |
| `tests/fixtures/*` | Reusable fixture builders (Go code, not static files) | builder-focused file names |
| `tests/integration/*_test.go` | End-to-end module behavior through HTTP / real wiring | `<domain>_test.go` |

Keep tests near the behavior they validate unless they are shared integration helpers.

## Makefile test targets

| Target | Command | When to use |
|---|---|---|
| `make test` | `go test ./...` | Full local verification across all packages/tests |
| `make test-unit` | `go test -short $(go list ./... \| grep -v '/tests/integration$')` | Fast feedback, unit-focused flow |
| `make test-integration` | `go test ./tests/integration/...` | Integration suites only |
| `make test-coverage` | `go test ./... -coverprofile=coverage.out` + `go tool cover -func=coverage.out` | Coverage report generation |

`test-unit` and integration `testing.Short()` gates work together to keep unit loops fast.

## Go testing patterns used in NexoKit

### 1) Table-driven tests with subtests (`t.Run`)

Use table-driven tests whenever the same behavior has multiple scenarios.

```go
func TestSomething(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "valid", input: "ok", wantErr: false},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// arrange / act / assert
		})
	}
}
```

### 2) `t.TempDir()` for filesystem tests

Do not write to real directories or rely on developer machine state.

```go
func TestWritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	// write/read assertions
}
```

### 3) `testing.Short()` for slow/external integration behavior

Integration tests must be skippable in short mode.

```go
func TestAuthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}
	// integration setup + assertions
}
```

### 4) Package choice (`package foo` vs `package foo_test`)

- Use same-package tests when you need internal/unexported behavior.
- Use `*_test` package when validating external API behavior and boundaries.

Prefer behavior-focused assertions over implementation-detail assertions.

## Integration test guidelines

### Required setup conventions

1. Use `helpers.NewSQLiteDB(t, models...)` for isolated test DBs.
2. Use helper fixtures/auth setup (`tests/helpers/auth.go`, `tests/helpers/fixtures.go`) instead of duplicating setup logic.
3. Use `gin.SetMode(gin.TestMode)` in HTTP integration suites.
4. Rely on helper `t.Cleanup()` and local test-scoped resources.

Example skeleton:

```go
func TestUsersCRUDIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration tests in short mode")
	}

	gin.SetMode(gin.TestMode)
	db := helpers.NewSQLiteDB(t, &users.User{}, &companies.Company{}, &roles.Role{})
	actor := helpers.SeedAuthActor(t, db, helpers.UserOptions{RoleSlug: roles.AdminRoleSlug})

	req := helpers.AuthenticatedRequest(t, http.MethodGet, "/api/users", nil, actor)
	// route invocation + assertions
}
```

### Fixtures and determinism

- Use fixture/helper builders to create related entities (user/company/role/permission).
- Keep fixture inputs explicit and deterministic.
- Avoid hidden cross-test dependencies.

## CI workflow and local reproduction

CI definition: `.github/workflows/ci.yml`

| CI check | What it validates | Local reproduction |
|---|---|---|
| `test` | Complete Go test suite | `go test ./...` |
| `vet` | Static analysis (`go vet`) | `go vet ./...` |
| `fmt-check` | Formatting compliance (`gofmt -l .` must be empty) | `gofmt -l .` then `make fmt` |

CI runs on:

- `push` to `main`
- every `pull_request` event

If CI fails:

1. Identify the failing job (`test`, `vet`, `fmt-check`).
2. Run the matching command locally.
3. Fix root cause and rerun until clean.

## Coverage policy

- Generate coverage with `make test-coverage`.
- Output file: `coverage.out` (Go standard format).
- Coverage threshold is intentionally not enforced yet (current baseline policy is 0%).
- Treat coverage as insight for risk-based improvements, not as a vanity metric.

## SQLite/PostgreSQL/Redis caveats

| Area | Current behavior | Caveat / guideline |
|---|---|---|
| Integration DB | SQLite `:memory:` via helpers | PostgreSQL-specific SQL/dialect behavior is not fully validated in these integration suites |
| External services (Redis/Valkey) | Optional/skippable in tests | Use short-mode and availability skip gates to avoid flaky CI/local runs |
| Test isolation | Per-test DB setup + cleanup | Avoid shared mutable state across tests |

When adding tests that rely on PostgreSQL-only features or mandatory Redis behavior, gate or scope them so default CI/local developer flow remains deterministic.

## Testing conventions checklist (for authors and reviewers)

- [ ] Test names are descriptive and scenario-based.
- [ ] Setup failures use `t.Fatal`/`t.Fatalf`.
- [ ] Assertion failures use `t.Error`/`t.Errorf` when multiple assertions should continue.
- [ ] Behavior is tested directly; implementation trivia is not.
- [ ] Mocks/fakes are kept small (interface boundaries only when needed).
- [ ] Integration tests use helpers/fixtures and `testing.Short()` gating.
- [ ] No mandatory third-party assertion framework is introduced (stdlib-first approach).
