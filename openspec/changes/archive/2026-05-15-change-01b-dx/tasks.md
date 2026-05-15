# Tasks: change-01b-dx Developer Experience

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~160 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Migrate middleware to centralized constants | PR 1 | Single PR; all changes are mechanical and behavior-preserving |
| 2 | Add pre-commit script, Make targets, and README docs | PR 1 | Included in same PR; no cross-unit dependencies |

## Phase 1: Message Constants

- [ ] 1.1 Add middleware constants block to `internal/platform/messages/messages.go` after existing middleware messages (`MsgHTTPRequest`, `CtxRequestID`, `HeaderRequestID`, `CORSAllowedMethods`, `CORSAllowedHeaders`, `CORSMaxAge`).
- [ ] 1.2 Update `internal/middleware/logger.go`: import `messages` package; replace `"http request"` with `messages.MsgHTTPRequest`; replace `"request_id"` with `messages.CtxRequestID`.
- [ ] 1.3 Update `internal/middleware/request_id.go`: import `messages` package; remove private `requestIDHeader` const; replace `requestIDHeader` with `messages.HeaderRequestID`; replace `"request_id"` with `messages.CtxRequestID`.
- [ ] 1.4 Update `internal/middleware/cors.go`: import `messages` package; replace methods string with `messages.CORSAllowedMethods`; replace headers string with `messages.CORSAllowedHeaders`; replace max-age string with `messages.CORSMaxAge`.
- [ ] 1.5 Update `internal/middleware/logger_test.go`: import `messages` package; replace `"http request"` assertion with `messages.MsgHTTPRequest`.
- [ ] 1.6 Verify: `go vet ./...` exits 0 and `go test ./...` passes.

## Phase 2: Pre-commit Script

- [ ] 2.1 Create `scripts/` directory at repo root.
- [ ] 2.2 Create `scripts/pre-commit.sh` with color helpers (`pass`, `fail`, `warn`) and five self-contained checks in order: `check_binaries`, `check_file_size`, `check_env_parity`, `check_go_vet`, `check_go_fmt`.
- [ ] 2.3 Ensure `check_binaries` blocks (red ✗) on any staged binary file.
- [ ] 2.4 Ensure `check_file_size` warns (yellow ⚠) for staged files >1MB but does not block.
- [ ] 2.5 Ensure `check_env_parity` silently skips when `.env` or `.env.example` is missing; otherwise warns (yellow ⚠) on key differences using `sed`/`awk` key extraction.
- [ ] 2.6 Ensure `check_go_vet` blocks (red ✗) on `go vet ./...` failures.
- [ ] 2.7 Ensure `check_go_fmt` blocks (red ✗) on `gofmt -l` hits and prints "run make fmt".
- [ ] 2.8 Make `scripts/pre-commit.sh` executable (`chmod +x`).

## Phase 3: Makefile Targets

- [ ] 3.1 Add `install-hooks` target: copy `scripts/pre-commit.sh` to `.git/hooks/pre-commit` and `chmod +x`.
- [ ] 3.2 Add `uninstall-hooks` target: `rm -f .git/hooks/pre-commit`.
- [ ] 3.3 Add `check-env` target: run `bash scripts/pre-commit.sh --check-env-only`.
- [ ] 3.4 Append `install-hooks`, `uninstall-hooks`, `check-env` to `.PHONY` line.
- [ ] 3.5 Verify: `make install-hooks` creates executable `.git/hooks/pre-commit`; `make uninstall-hooks` removes it; `make check-env` runs without error.

## Phase 4: README Documentation

- [ ] 4.1 Add rows for `make install-hooks`, `make uninstall-hooks`, and `make check-env` to the Commands table in `README.md`.
- [ ] 4.2 Insert `## Pre-commit Hooks` section immediately after `## Commands` and before `## Log Files`, documenting `make install-hooks`, what the hook checks, and bypass via `git commit --no-verify`.
