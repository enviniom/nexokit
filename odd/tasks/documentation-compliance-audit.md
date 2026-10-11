# Documentation compliance audit and remediation

## Objective
Audit current implementation against repository documentation, correct demonstrated violations, and leave verified work-unit commits on the remediation branch for human validation.

## Constraints and interpretation
- Updated user authorization: integrate documentation branch into main, create a remediation branch, and commit completed work units there. No push, publishing or PR creation.
- main fast-forwarded from 2f732d4 to 58ce598; working branch is fix/documentation-compliance. Initial working tree was clean.
- Parent may make recommended technical decisions; record all accepted project decisions for final review.
- Existing flat module layout is documented legacy, not a migration mandate. New generated modules must follow new-code rules.
- Do not implement optional roadmap features, change CI policy or weaken documentation to claim compliance.
- Do not read secrets, mutate real services/databases or perform destructive operations.
- Single writer; applicable behavioral fixes use observed RED/GREEN tests.
- RDD is off globally; no native review transaction requested.

## Findings and scope
1. Role-permission replacement accepts role ID without tenant context at its write boundary (`internal/modules/iam/roles/assign_permissions_to_role/repository.go`).
2. Module generator templates expose technical errors, use value validation receivers, map responses in handlers, accept numeric tenant IDs and emit legacy root layout for new code.
3. Several IAM handlers use tenant.NewRoot when tenant context is absent, violating fail-closed persistence policy.
4. Permission update maps JSON binding errors to 422 instead of 400.
5. Completed production inventory: 36/36 repositories, 22/22 tracked queries, 31/31 tracked handlers, 7/7 routes read. Core review completed 21/21; bounded production import scan checked 164 Go files and found zero sibling-module imports. Root-guarded company operations are intentional global policy, not tenant violations.

## Tasks
- [x] T1 (done): Complete source compliance inventory and report demonstrated findings with explicit coverage. Audit all production persistence, routes/handlers, imports and remaining documentation/runtime surfaces; distinguish current rules, legacy conventions and future guidance.
- [x] T2 (done): Scope role-permission mutations to trusted tenant context, with regression tests for same/cross tenant, root-global, selected-root and transaction failures as applicable. Verify focused IAM tests.
- [x] T5 (done): Reject absent tenant context in IAM handlers; assert no service calls and preserve explicit root/global/scoped contexts. Update affected HTTP integration wiring only when required.
- [x] T6 (done): Return 400 for malformed permission-update JSON and retain 422 for valid JSON with invalid fields; prove with handler regression tests.
- [x] T3 (done): Make generated modules conform to new-code layout, safe DTO mapping and normalization, trusted tenant scope and safe errors. Generate normal/tenant fixtures in temporary directories and prove compiled behavior with regression tests; update affected scaffold documentation accurately. Keep generator compatibility where safe; no unrelated module migration.
- [x] T7 (done): Reconcile generator OpenSpec with current accepted guide/output, mark obsolete context snapshot as historical, remove the intentionally dangling planned ADR link, and correct audit inventory counts. Verify doc/source consistency without changing behavior.
- [x] T4 (done): Run final independent checks, reconcile every finding and publish audit evidence/limitations for human validation. Full tests, vet, module guards, format, diff checks; record work-unit commits.

## Acceptance
- Each confirmed violation is corrected with observed applicable checks, or explicitly recorded as blocked.
- Generated DTOs belong to entity core, mappings expose only safe public fields, tenant operations cannot trust request company IDs, production errors do not expose technical text.
- Role association mutation cannot affect foreign-tenant or deleted roles.
- Final audit identifies checked and unchecked surfaces; green SQLite tests do not certify PostgreSQL/Redis.
- Changes are committed only on fix/documentation-compliance; no remote publication. User validates final result.

## Baseline evidence
Verifier: `go test -timeout 180s ./...`, `go test -short ./...`, `go vet ./...`, `make check-module-errors`, tracked Go `gofmt -l`, `git diff --check`: all pass. Integration uses SQLite; real PostgreSQL/migrations and Redis unverified. Working tree remained clean.

## Audit evidence and project decisions
- T1 work-unit commit: `f08da92` — `docs(odd): record compliance audit and remediation plan`.
- Coverage: all 36 repositories, 22 tracked query files, 31 tracked handler files, 7 route files and 21 module core files read; all 164 production module Go files scanned for sibling imports (none).
- Scope is source-level compliance review, not formal security certification; PostgreSQL/Goose and Redis runtime remain unverified.
- D1: Preserve documented legacy layouts and payload conventions; enforce entity DTO/new-code conventions on generated modules instead of broad unrelated migration.
- D2: Require explicit trusted tenant context at handler and mutation boundaries. Missing handler context returns existing 400 contract; global scope remains explicit middleware policy.
- Middleware spot-check correction: AllowRootGlobalScope DOES establish NewRoot/NewScoped context on successful requests (`internal/middleware/tenant.go`). Thus IAM fallback defects are fail-open defense gaps, not evidence that current mounted requests bypass tenant middleware.
- D3: Preserve intentional root-only global company operations; do not expose them to non-root routes.
- D4: Keep manual CI trigger, future CQRS guidance and documented operator production policies unchanged; no unsupported feature expansion.
- D5: Binding failures return 400 and field validation 422; preserve response helpers and safe public messages.
- D6: Do not enable RDD (currently off), publish branches or mutate real services for validation.
- D7: BaseModel has no shared public-ID hook. Generated entity models will own BeforeCreate ULID generation using existing platform/identity instead of changing shared model behavior across existing modules. Preserve explicitly provided valid IDs if appropriate; tests verify ULID format and safe responses.
- D8: Generated plain CRUD uses handler+repository slices, not pass-through services; default scaffold supplies get/list, --crud adds mutations; contracts/errors remain module core and wiring explicit.
- D9: Generated tenant list allows explicit root-global scope; get/create/update/delete require selected scope. JSON company_id is ignored, inserts derive ownership from context, updates cannot reassign it and use scoped transactions instead of Save/upsert.
- D10: Safe response mapping exposes ULID and safe fields; pointer DTO validation trims before rules; module errors preserve diagnostic causes while unknown technical failures become safe500.
- D11: Generated table names match module names explicitly; migration stubs remain adopter-owned and generated routes must be wired with project guards.
- D12: Reconcile the old generator OpenSpec with current accepted architecture and tested output; preserve obsolete context.md body with a historical warning instead of treating it as current guidance. Planned ADR paths are literal until records exist. No unrelated historical archive repair.
- Golden update convention is UPDATE_GOLDEN=1 (not a -update flag); rerun without the variable after inspecting fixture changes.

## Verification evidence
- T2 work-unit commit: `e0c0f4c`.
- T2: observed RED for foreign/deleted/missing/selected-root foreign mutations; GREEN for own/global/selected scopes, scoped cache deletion and duplicate-insert rollback. Writer and independent verifier passed focused assignment/integration and full IAM tests; verifier results cached. No blocker. SQLite does not certify PostgreSQL locking/Redis behavior.

- T5 work-unit commit: `f848eb2`.
- T5: all 15 handler fallbacks removed (7 roles, 8 users). Observed RED using unreachable-service sentinel, GREEN for missing400 and explicit root/scoped contexts. Formatted30 files; independent fresh role/user/integration tests and diff checks PASS. 872 changed lines largely preserve and expand existing test tables; coherent single policy unit.

- T6 work-unit commit: `0ce7393`.
- T6: RED for malformed JSON/wrong field types returning422; GREEN400 binding,422 invalid fields, zero service calls. Fresh writer+independent permission suite, format and diff checks PASS.

- T3 work-unit commit: `fa2f893` (47 files, 1696 additions/679 deletions, including emitted-code snapshots and compiled fixture tests). Large but one coherent generator contract; review code/templates before generated snapshots.
- T3 writer: observed RED for missing nested output and nonpersistent trim; GREEN four compiled fixture variants and behavior regressions (safe JSON/errors, binding/validation, tenants, sort, soft deletion, error identity, rollback). CLI/docs/full tests, vet, guards and format PASS. Independent fresh CLI/docs/full tests, vet, module guard, all-Go formatting, tracked/untracked whitespace and explicit emitted-code fixture run PASS. Stale traceability-ledger sentence corrected with source-based prose readback and diff check; prose has no meaningful RED/GREEN. No behavior blockers.

## Final verification and limitations
- T7/T4 documentation reconciliation and verified closeout work-unit commit: `b3e3e93` — `docs: reconcile generator specification and close compliance audit`.
- T4 independent fresh `go test -count=1 -timeout 300s ./...`, `go vet ./...`, `go build ./...`, `make check-module-errors`, tracked-Go `gofmt -l`, `git diff main..HEAD --check` and worktree diff check: PASS at fa2f893. Four generated variants compiled and behavior tests executed.
- Zero IAM handler implicit-root fallbacks; production sibling-import scan:164 files, zero violations. Reproducible tracked counts:36 repositories,22 query files,31 handlers,7 routes,21 core files. Initial scout23/32 claims corrected; inventory is not security certification.
- T7 passive docs: writer and independent `go test -count=1 -timeout 180s ./tests/docs ./tests/cli`, source/layout comparison, local links/anchors and diff checks PASS. Historical context body preserved exactly. No source changed after full checks; no behavioral RED/GREEN applies to prose.
- Failed final checks: none. Redis-backed tests SKIPPED because no service was available (`TestRedisCache/*`, `TestRedisLimiterIntegration`); PostgreSQL/Goose runtime and actual Redis behavior UNVERIFIED. SQLite integration is not production database certification.
- GORM constraint sentinel translation requires TranslateError. Generated update RowsAffected assumptions are tested on SQLite; project-target PostgreSQL driver/locking not exercised, no MySQL portability claim.
- No new broken current-documentation links/anchors. Pre-existing archived OpenSpec verify-report link to starter-template.md remains outside this change. Harmless retired templates/snapshots and unused FileNames helper retained, not emitted as runtime code.

## Progress and next step
All seven tasks complete within the documented audit scope. Review commits on fix/documentation-compliance against main, especially generator source/templates before snapshots. main remains58ce598. No push/PR or remediation merge to main; final human validation pending.
