# Archive Report: change-01b-dx

**Change**: change-01b-dx  
**Title**: Developer Experience — Centralize Constants, Pre-Commit Hook, Makefile, README  
**Archived**: 2026-05-15  
**Verification Outcome**: APPROVED

---

## Artifact Observation IDs (Engram Traceability)

| Artifact | Observation ID |
|----------|---------------|
| Proposal | #411 |
| Spec | #412 |
| Design | #413 |
| Tasks | #415 |
| Apply Progress | #416 |
| Verify Report | #417 |

---

## What Was Implemented

A behavior-preserving developer-experience cleanup across the nexokit middleware layer and local development tooling.

1. **Message Constants Centralization**
   - Added middleware constants (`MsgHTTPRequest`, `CtxRequestID`, `HeaderRequestID`, `CORSAllowedMethods`, `CORSAllowedHeaders`, `CORSMaxAge`) to `internal/platform/messages/messages.go`.
   - Migrated `internal/middleware/logger.go`, `request_id.go`, `cors.go`, and `logger_test.go` to use the new constants instead of string literals.

2. **Pre-commit Hook**
   - Created `scripts/pre-commit.sh` with five self-contained checks in order:
     - `check_binaries` — blocks on staged binary files (fail-fast, red ✗)
     - `check_file_size` — warns for files >1MB (yellow ⚠, non-blocking)
     - `check_env_parity` — warns when `.env` and `.env.example` keys differ (ignores comments/blanks), silently skips if files missing
     - `check_go_vet` — blocks on `go vet ./...` failures (red ✗)
     - `check_go_fmt` — blocks on unformatted Go files with "run make fmt" guidance (red ✗)

3. **Makefile Targets**
   - Added `install-hooks`, `uninstall-hooks`, and `check-env` targets.
   - Updated `.PHONY` to include all targets.

4. **README Documentation**
   - Added new command rows (`install-hooks`, `uninstall-hooks`, `check-env`) to the Commands table.
   - Inserted a **Pre-commit Hooks** section with setup instructions, checks table, and bypass guidance (`git commit --no-verify`).

---

## Files Changed

| File | Action |
|------|--------|
| `internal/platform/messages/messages.go` | Modified — added middleware constants block |
| `internal/middleware/logger.go` | Modified — migrated to `messages` constants |
| `internal/middleware/request_id.go` | Modified — migrated to `messages` constants |
| `internal/middleware/cors.go` | Modified — migrated to `messages` constants |
| `internal/middleware/logger_test.go` | Modified — assertion updated to use constant |
| `internal/middleware/recovery.go` | Modified — post-verification fix: migrated `"request_id"` to `messages.CtxRequestID` |
| `scripts/pre-commit.sh` | Created — executable pre-commit hook |
| `Makefile` | Modified — added hook/env targets and updated `.PHONY` |
| `README.md` | Modified — added commands and Pre-commit Hooks section |

---

## Decisions Made

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Middleware constants location | `internal/platform/messages/messages.go` | Keeps project-owned strings centralized without creating a mega-package. |
| HTTP header names | Only `HeaderRequestID` constantized; standard CORS headers kept as literals | Spec allows standard HTTP header names as literals; only project-owned values move. |
| Hook severity | Binary/vet/fmt fail fast; size/env parity warn only | Matches spec: warnings improve DX without blocking commits. |
| Env key parsing | `sed` + `awk -F=` in bash | Bash-only script stays portable and easy to install as a Git hook. |

---

## Verification Outcome

**APPROVED**

- **Build**: `go build ./...` passed
- **Vet**: `go vet ./...` passed
- **Tests**: `go test ./...` passed (all packages)
- **Spec compliance**: 12/12 scenarios compliant after post-verification fixes

---

## Post-Verification Fixes Applied

The initial verification report flagged one **CRITICAL** issue and one **SUGGESTION**. Both were resolved before final approval.

1. **`recovery.go` CtxRequestID migration**
   - **Issue**: `internal/middleware/recovery.go` was missed during the constant migration and still used the literal string `"request_id"` on lines 15 and 18.
   - **Fix**: Replaced `c.Get("request_id")` with `c.Get(messages.CtxRequestID)` and `slog.String("request_id", ridStr)` with `slog.String(messages.CtxRequestID, ridStr)`.

2. **`logger_test.go` error message updated**
   - **Issue**: The test failure message read `"expected log output to contain 'http request'"` while the assertion used `messages.MsgHTTPRequest`, creating a mismatch that could confuse future refactors.
   - **Fix**: Updated the error message text to reference the constant name (`MsgHTTPRequest`) instead of the literal string.

---

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| `http-middleware` | Updated | Added "Message constants" requirement and "Compile and vet pass after constant migration" scenario |
| `dev-environment` | Updated | Added Pre-commit hook, install-hooks, uninstall-hooks, check-env requirements and scenarios; expanded Makefile targets and README requirements |

---

## Archive Contents

- `proposal.md` ✅
- `spec.md` ✅
- `design.md` ✅
- `tasks.md` ✅ (16/16 tasks complete)
- `verify-report.md` ✅
- `archive-report.md` ✅ (this file)

---

## Source of Truth Updated

The following specs now reflect the new behavior:
- `openspec/specs/http-middleware/spec.md`
- `openspec/specs/dev-environment/spec.md`

---

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived. Ready for the next change.
