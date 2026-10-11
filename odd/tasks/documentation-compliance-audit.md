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
5. Completed production inventory: 36/36 repositories, 23/23 queries, 32/32 handlers, 7/7 routes read. Core review completed 21/21; bounded production import scan checked 164 Go files and found zero sibling-module imports. Root-guarded company operations are intentional global policy, not tenant violations.

## Tasks
- [x] T1 (done): Complete source compliance inventory and report demonstrated findings with explicit coverage. Audit all production persistence, routes/handlers, imports and remaining documentation/runtime surfaces; distinguish current rules, legacy conventions and future guidance.
- [x] T2 (done): Scope role-permission mutations to trusted tenant context, with regression tests for same/cross tenant, root-global, selected-root and transaction failures as applicable. Verify focused IAM tests.
- [x] T5 (done): Reject absent tenant context in IAM handlers; assert no service calls and preserve explicit root/global/scoped contexts. Update affected HTTP integration wiring only when required.
- [x] T6 (done): Return 400 for malformed permission-update JSON and retain 422 for valid JSON with invalid fields; prove with handler regression tests.
- [ ] T3 (in_progress): Make generated modules conform to new-code layout, safe DTO mapping and normalization, trusted tenant scope and safe errors. Generate normal/tenant fixtures in temporary directories and prove compiled behavior with regression tests; update affected scaffold documentation accurately. Keep generator compatibility where safe; no unrelated module migration.
- [ ] T4 (pending): Run final independent checks, reconcile every finding and publish audit evidence/limitations for human validation. Full tests, vet, module guards, format, diff checks; record work-unit commits.

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
- Coverage: all 36 repositories, 23 queries, 32 handlers, 7 route files and 21 module core files read; all 164 production module Go files scanned for sibling imports (none).
- Scope is source-level compliance review, not formal security certification; PostgreSQL/Goose and Redis runtime remain unverified.
- D1: Preserve documented legacy layouts and payload conventions; enforce entity DTO/new-code conventions on generated modules instead of broad unrelated migration.
- D2: Require explicit trusted tenant context at handler and mutation boundaries. Missing handler context returns existing 400 contract; global scope remains explicit middleware policy.
- Middleware spot-check correction: AllowRootGlobalScope DOES establish NewRoot/NewScoped context on successful requests (`internal/middleware/tenant.go`). Thus IAM fallback defects are fail-open defense gaps, not evidence that current mounted requests bypass tenant middleware.
- D3: Preserve intentional root-only global company operations; do not expose them to non-root routes.
- D4: Keep manual CI trigger, future CQRS guidance and documented operator production policies unchanged; no unsupported feature expansion.
- D5: Binding failures return 400 and field validation 422; preserve response helpers and safe public messages.
- D6: Do not enable RDD (currently off), publish branches or mutate real services for validation.

## Verification evidence
- T2 work-unit commit: `e0c0f4c`.
- T2: observed RED for foreign/deleted/missing/selected-root foreign mutations; GREEN for own/global/selected scopes, scoped cache deletion and duplicate-insert rollback. Writer and independent verifier passed focused assignment/integration and full IAM tests; verifier results cached. No blocker. SQLite does not certify PostgreSQL locking/Redis behavior.

- T5 work-unit commit: `f848eb2`.
- T5: all 15 handler fallbacks removed (7 roles, 8 users). Observed RED using unreachable-service sentinel, GREEN for missing400 and explicit root/scoped contexts. Formatted30 files; independent fresh role/user/integration tests and diff checks PASS. 872 changed lines largely preserve and expand existing test tables; coherent single policy unit.

- T6: RED for malformed JSON/wrong field types returning422; GREEN400 binding,422 invalid fields, zero service calls. Fresh writer+independent permission suite, format and diff checks PASS.

## Progress and next step
T1/T2/T5/T6 complete. Generator T3 active next; final T4 pending.
