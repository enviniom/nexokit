# Proposal: Developer Experience — Centralize Constants, Pre-Commit Hook, Makefile, README

## Intent
Eliminate scattered magic strings in middleware, automate quality checks before commit, and improve developer tooling. This reduces cognitive load and prevents `.env` drift and formatting regressions.

## Scope

### In Scope
- Add missing constants to `internal/platform/messages/` and migrate middleware magic strings
- Create `scripts/pre-commit.sh` with `.env` parity check and file-size guard
- Add Makefile targets: `setup-hooks`, `check-env`, `lint`
- Update README with new commands and conventions

### Out of Scope
- Functional changes to CORS policy, logger output, or request ID generation
- GitHub Actions CI/CD pipelines
- README full rewrite or new tutorial sections

## Capabilities

### New Capabilities
None

### Modified Capabilities
None — pure refactor and DX tooling; no spec-level behavior changes.

## Approach

- **Messages**: Extend `internal/platform/messages/messages.go` with middleware constants (`MsgHTTPRequest`, `CtxRequestID`, `HeaderRequestID`). Replace literals in `middleware/logger.go`, `middleware/request_id.go`, and `middleware/logger_test.go`. For CORS, constantize project-owned values (methods list, allowed headers list, max-age value) in a new `cors.go` constants block; keep standard HTTP header names as string literals.
- **Pre-commit**: Create `scripts/pre-commit.sh`. Check `.env` vs `.env.example` parity (ignore comments and blank lines). Warn on files >1MB. Fail on `go fmt` or `go vet` errors. Install via `make setup-hooks`.
- **Makefile**: Add `setup-hooks` (installs pre-commit), `check-env` (runs parity script), `lint` (runs `go vet` then `go fmt` check).
- **README**: Add new Makefile targets to the Commands table and document pre-commit hook installation.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/platform/messages/messages.go` | Modified | Add middleware constants |
| `internal/middleware/logger.go` | Modified | Replace `"http request"` with constant |
| `internal/middleware/request_id.go` | Modified | Replace `"request_id"` and `"X-Request-ID"` with constants |
| `internal/middleware/cors.go` | Modified | Replace method/header/max-age strings with constants |
| `internal/middleware/logger_test.go` | Modified | Update assertion to use constant |
| `scripts/pre-commit.sh` | New | Git pre-commit hook script |
| `Makefile` | Modified | Add `setup-hooks`, `check-env`, `lint` targets |
| `README.md` | Modified | Document new targets and hook |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|-------------|
| Constant rename breaks import cycle | Low | Keep package name; only add constants |
| Pre-commit hook rejects valid large files | Low | Warn only at 1MB; do not block commit |
| `.env` parity false positives | Low | Ignore comments and blank lines in both files |

## Rollback Plan
- Revert constants: replace constants back with original string literals
- Remove `scripts/` directory
- Revert Makefile additions
- Restore README from git history

## Dependencies
None

## Success Criteria
- [ ] All middleware magic strings use `messages` constants
- [ ] `make setup-hooks` installs the pre-commit hook
- [ ] `make check-env` passes when `.env` mirrors `.env.example` (ignoring comments/blanks)
- [ ] `make lint` passes with zero errors
- [ ] Tests pass after constant migration
