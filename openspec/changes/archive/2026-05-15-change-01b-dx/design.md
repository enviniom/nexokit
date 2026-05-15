# Design: change-01b-dx Developer Experience

## Technical Approach

Implement a behavior-preserving DX cleanup: add middleware-owned constants to `internal/platform/messages`, migrate middleware references, add a self-contained pre-commit script, expose hook/env helpers through Make, and document onboarding. Specs covered: HTTP Middleware and Dev Environment.

## Architecture Decisions

| Decision | Choice | Alternatives considered | Rationale |
|---|---|---|---|
| Middleware constants | Add one new `const ()` block in `internal/platform/messages/messages.go` after existing middleware messages. | Per-file private constants in middleware. | Keeps project-owned strings centralized without creating a mega-package. |
| HTTP header names | Use `messages.HeaderRequestID` only for the project request ID header; keep CORS response header names literal. | Constantize every standard HTTP header. | Spec allows standard header names as literals; only project-owned values move. |
| Hook warnings | Binary, vet, and fmt fail fast; size and env parity warn. | Block on every warning. | Matches spec: warnings should improve DX without blocking commits. |
| Env key parsing | Use `sed` + `awk -F=` in the script. | Go helper or external dependency. | Bash-only script stays portable and easy to install as a Git hook. |

## Data Flow

```text
git commit → .git/hooks/pre-commit
  → check_binaries → fail on binary
  → check_file_size → warn >1MB
  → check_env_parity → warn key drift
  → check_go_vet → fail on vet
  → check_go_fmt → fail on gofmt diff
```

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/platform/messages/messages.go` | Modify | Add middleware constants below existing middleware messages. |
| `internal/middleware/logger.go` | Modify | Import `messages`; replace `"http request"` and `"request_id"`. |
| `internal/middleware/request_id.go` | Modify | Import `messages`; remove private `requestIDHeader`; replace header/context literals. |
| `internal/middleware/cors.go` | Modify | Import `messages`; replace methods, allowed headers, and max-age values. |
| `internal/middleware/logger_test.go` | Modify | Import `messages`; assert against `messages.MsgHTTPRequest`. |
| `scripts/pre-commit.sh` | Create | Git hook script with binary, size, env, vet, and fmt checks. |
| `Makefile` | Modify | Add hook/env targets and update `.PHONY`. |
| `README.md` | Modify | Add command rows and Pre-commit Hooks section. |

## Interfaces / Contracts

### Message constants

Add this exact block to `messages.go`:

```go
const (
	// Middleware — Logger
	MsgHTTPRequest = "http request"
	CtxRequestID   = "request_id"

	// Middleware — Request ID
	HeaderRequestID = "X-Request-ID"

	// Middleware — CORS
	CORSAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	CORSAllowedHeaders = "Content-Type, Authorization, X-Request-ID"
	CORSMaxAge         = "86400"
)
```

Files importing `github.com/enviniom/nexokit/internal/platform/messages`: `logger.go`, `request_id.go`, `cors.go`, `logger_test.go`. References change to `messages.MsgHTTPRequest`, `messages.CtxRequestID`, `messages.HeaderRequestID`, `messages.CORSAllowedMethods`, `messages.CORSAllowedHeaders`, `messages.CORSMaxAge`.

### Pre-commit script architecture

Top-level layout:

```bash
FILE_SIZE_LIMIT_MB=1
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'

pass() { printf "${GREEN}✓ %s${NC}\n" "$1"; }
fail() { printf "${RED}✗ %s${NC}\n" "$1"; exit 1; }
warn() { printf "${YELLOW}⚠ %s${NC}\n" "$1"; }
```

Checks are self-contained functions: `check_binaries()`, `check_file_size()`, `check_env_parity()`, `check_go_vet()`, `check_go_fmt()`. Main execution order is exactly: binaries, file size, env parity, vet, fmt. Fail-fast happens by calling `fail` inside blocking checks.

`check_env_parity()` silently returns when `.env` or `.env.example` is absent. It extracts keys from each file with `sed '/^[[:space:]]*#/d;/^[[:space:]]*$/d' FILE | awk -F= '{print $1}' | sort`, compares keys, and emits `warn` on differences.

### Makefile targets

```make
.PHONY: build run test migrate-up migrate-down migrate-create migrate-status fmt vet install-hooks uninstall-hooks check-env

install-hooks:
	cp scripts/pre-commit.sh .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit

uninstall-hooks:
	rm -f .git/hooks/pre-commit

check-env:
	bash scripts/pre-commit.sh --check-env-only
```

### README placement and outline

Insert `## Pre-commit Hooks` immediately after `## Commands` and before `## Log Files`.

````markdown
## Pre-commit Hooks

Install the local Git hook:

```bash
make install-hooks
```

The hook checks for staged binary files, warns for files over 1MB, warns when `.env` and `.env.example` keys differ, runs `go vet ./...`, and blocks unformatted Go files with guidance to run `make fmt`.

Bypass only when necessary:

```bash
git commit --no-verify
```
````

Also add command rows for `make install-hooks`, `make uninstall-hooks`, and `make check-env`.

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Unit | Middleware still sets/logs same values | Existing middleware tests plus compile checks. |
| Script | Hook fail/warn behavior | Run script against staged binary, large file, env drift, unformatted Go. |
| Project | No regression | `go vet ./...`, `go test ./...`, `make check-env`. |

## Migration / Rollout

No migration required. Hook installation is opt-in via `make install-hooks`.

## Open Questions

None.
