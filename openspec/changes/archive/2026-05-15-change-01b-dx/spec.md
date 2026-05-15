# Delta Specs for change-01b-dx

## HTTP Middleware

### ADDED Requirements

| Requirement | Rule |
|-------------|------|
| Message constants | Middleware MUST use `internal/platform/messages/` constants for project-owned string values. Standard HTTP header names MAY remain literals. |

#### Scenario: Compile and vet pass after constant migration

- GIVEN `messages.go` exports `MsgHTTPRequest`, `CtxRequestID`, `HeaderRequestID`, and CORS value constants
- AND middleware files and `logger_test.go` reference those constants
- WHEN `go vet ./...` runs
- THEN exit code is 0

## Dev Environment

### ADDED Requirements

| Requirement | Rule |
|-------------|------|
| Pre-commit hook | `scripts/pre-commit.sh` MUST be executable. It MUST block on binary files, `go vet` errors, or unformatted Go files, and fail fast. It SHOULD warn (non-blocking) for files >1MB or `.env`/`.env.example` key mismatches (ignoring comments and blanks). Output MUST use green ✓, red ✗, yellow ⚠. |

#### Scenario: Binary file blocks commit

- GIVEN a staged binary file
- WHEN the pre-commit hook runs
- THEN it blocks with a red ✗ message
- AND the commit does not proceed

#### Scenario: Unformatted Go blocks commit

- GIVEN a staged Go file that fails `gofmt -l`
- WHEN the pre-commit hook runs
- THEN it blocks with a red ✗ containing "run make fmt"

#### Scenario: Warnings allow commit

- GIVEN a staged file >1MB and a `.env` key missing from `.env.example` (ignoring comments and blanks)
- WHEN the pre-commit hook runs
- THEN yellow ⚠ warnings are printed
- AND the commit proceeds

#### Scenario: Missing .env skips silently

- GIVEN `.env` does not exist
- WHEN the env check runs
- THEN no warning is emitted

### ADDED Requirements (Makefile)

| Requirement | Rule |
|-------------|------|
| install-hooks | `make install-hooks` MUST copy `scripts/pre-commit.sh` to `.git/hooks/pre-commit` and make it executable. |
| uninstall-hooks | `make uninstall-hooks` MUST remove `.git/hooks/pre-commit`. |
| check-env | `make check-env` MUST run the `.env`/`.env.example` parity check. |

#### Scenario: Hook install and uninstall are reversible

- GIVEN `make install-hooks` has run
- WHEN `make uninstall-hooks` runs
- THEN `.git/hooks/pre-commit` does not exist

### MODIFIED Requirements

#### Requirement: Makefile targets

(Previously: `build`, `test`, `run`, `migrate-*`, `fmt`, `vet` only)

The system MUST provide targets: `build`, `test`, `run`, `migrate-up`, `migrate-down`, `migrate-create`, `migrate-status`, `fmt`, `vet`, `install-hooks`, `uninstall-hooks`, `check-env`. All MUST appear in `.PHONY`.

#### Scenario: Run tests

- GIVEN `make test` is executed
- WHEN tests complete
- THEN `go test ./...` has been invoked

#### Scenario: Create migration

- GIVEN `make migrate-create NAME=add_users`
- WHEN the command completes
- THEN a new file exists in `migrations/` with the correct timestamp format

#### Requirement: README with setup instructions

(Previously: omitted pre-commit hooks and new Makefile targets)

The system MUST provide a `README.md` with: project description, prerequisites, `.env` setup, `docker-compose up`, `make migrate-up`, `make run`, and a "Pre-commit Hooks" section documenting setup, checks, and bypass via `git commit --no-verify`. The commands table MUST include `install-hooks`, `uninstall-hooks`, and `check-env`.

#### Scenario: New developer onboarding

- GIVEN a developer clones the repository
- WHEN they follow README instructions
- THEN the API starts locally without additional undocumented steps
