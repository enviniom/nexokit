# Auth documentation compliance

## Objective and authorization
Complete auth integration coverage and entity DTO/model layout while preserving HTTP payloads and existing authentication policy.

- User approved tests and DTO migration only; email normalization and session hardening deferred.
- User then removed flat core DTO/model aliases and tests, migrated queries and some consumers, and requested completion through direct entity imports.
- User explicitly approved minimal import-shadowing corrections in authenticate_user/repository_test.go and service_test.go.
- During implementation no Git mutations were authorized; staged snapshots were preserved. User now explicitly authorizes staging and committing the completed auth work if final checks pass. Push, publishing and unrelated changes remain unauthorized.
- Branch: fix/documentation-compliance. Single writer.
- Engram mirror: project nexokit, topic odd/auth-documentation-compliance/tasks (observation 4054).

## Tasks
- [x] T1: Test /auth/me using real PASETO bearer middleware and IAM database permissions: missing/invalid tokens rejected; login-issued token accepted.
- [x] T2: Test expired refresh and user deactivation after login: unauthorized, persisted session count/hash/state unchanged.
- [x] T3: Complete direct core/user and core/refresh_token model/DTO imports, DTO-owned mapping, relocated tests and accurate current-state docs. No flat or slice compatibility aliases remain.

- [x] T4: Add observable HTTP contract coverage in slices/integration only: safe login/me JSON, optional fields, 400 malformed body and 422 validation envelope with zero service calls. Status: done; worker focused suites and parent login-handler spot check passed. No production changes.
- [x] T5: Verify current post-deletion state with full tests, vet, formatting, direct imports and staged-index preservation. Status: done; independent final checks all passed.

## Final state
- Root auth/core imports are retained only for module errors and contracts; removed model/DTO symbols have no remaining Go references repository-wide.
- Flat dto.go/model.go remain deleted. User subsequently removed all core tests plus container/routes tests. Do not recreate any of them. Preserve useful HTTP guarantees in slices/integration instead.
- SessionView belongs to core/user and consumers reference it directly.
- Queries and other user-authored migrations preserved. No IAM-owned models moved. HTTP payloads and email/session policies unchanged.
- Validate receivers are pointers: in-repository callers compile; external calls on non-addressable values may break.

## Current closure round
User authorizes final test additions including representative 400 responses: test our response envelope and early-return behavior, not Gin parser internals. Preserve all existing policy and user changes. Only slice handler tests and auth integration test may be edited. Working tree initially clean relative to user's staged snapshot. Current staged diff SHA256: c00c057eaa1f1fdaccaf3146dbe2dbc2ea2fd0bbe1f038c0e67d21473ed4f116; preserve it exactly. Prior full verification below predates the deletions and is historical, not proof of current state.

## Final HTTP coverage evidence
Worker added coverage to four slice handler tests and auth integration only (+315/-22): login/refresh/logout 400/422 envelopes and zero service calls, login optional company/audit IDs, /me nil/empty permissions and optional company, actual login/me exact safe JSON fields and timestamps. Baseline and focused/auth-integration-middleware suites passed; assertions passed first run for existing behavior (no artificial RED). Parent fresh login handler package passed. Staged SHA256 unchanged. ASSESS now succeeds: high risk over aggregate current changes, RDD off, independent verifier required and completed. Final `go test ./... -count=1` and `go vet ./...` passed; integration executed (0.684s), dedicated verbose auth integration passed 11/11 subtests. Both working-tree/staged diff checks passed. Format scan clean for 34 existing changed Go files; deleted paths excluded. Staged binary diff SHA256 unchanged before/after: c00c057eaa1f1fdaccaf3146dbe2dbc2ea2fd0bbe1f038c0e67d21473ed4f116. No core/container/routes tests recreated; no production edits this round. Minor non-blocking notes: rotate/revoke slice success tests assert status only, integration covers actual behavior; no additional work required for authorized closure. No native review initiated.

## Verification evidence
- Follow-up pre-edit suite observed compile RED from removed aliases and test package shadowing; required focused suite passed after fixes. Existing-behavior coverage did not manufacture RED. Initial FromModel migration had an undefined-method RED then GREEN; FromSession tests were added afterward.
- Independent final go test ./... -count=1: PASS, 75 tested packages, integration executed (0.740s).
- Independent go vet ./...: PASS.
- Independent git diff --check and git diff --cached --check: PASS.
- gofmt -l over existing changed/new Go files: clean (40 files); deleted paths excluded from scan.
- Parent fresh go test ./internal/modules/auth/core/user ./internal/modules/auth/core/refresh_token -count=1: PASS.
- Historical pre-deletion follow-up index checksum: 427a8867ac23e2776100b8af102af7f979c631eae30b58c0b5e361cb4e2bce01. Superseded by user's later staged snapshot; current round uses c00c057eaa1f1fdaccaf3146dbe2dbc2ea2fd0bbe1f038c0e67d21473ed4f116.
- Earlier fixture PublicID and formatting failures were corrected and reverified. No failing functional checks remain.
- Native RDD off; ASSESS unassessable due undeclared untracked scope; conservative independent verification applied, native review not started.

## Limitations and next step
No live-server or production-database verification; integration uses SQLite. External Go consumers were not verified. No Git mutations were made during implementation; staged snapshots were preserved. User has now unstaged the accumulated work and explicitly requests its commit. Final precommit auth/docs/integration, build, vet, format and diff checks passed. Default full suite initially failed existing TestRedisLimiterIntegration. User requested read-only diagnosis: fixed Redis key rl:login:127.0.0.1 persists30s without cleanup; absent-key isolated PASS, immediate rerun with residual counter FAIL, natural expiry PASS. No middleware changes; auth integration uses no-op limiter. Full `go test ./... -p 1 -count=1` passed with expired key. This baseline test-isolation issue remains a follow-up, not an auth regression. No fix/Redis configuration change performed. Auth ready under original explicit commit authorization; stage/commit only this work, no push. Commit identity to be recorded in Engram after successful commit. Commit identity and final handoff will be recorded in the Engram recovery mirror after success. Email normalization and session hardening remain deferred.

Implementation complete and previously independently verified. Current authorization: commit the completed auth work after verification. Baseline Redis test-isolation cause confirmed, full serialized suite passed; disclose default-run failures and leave Redis fix outside commit. No push. Provide an IAM read-only review prompt for a new session, then stop. Do not start IAM in this session.
