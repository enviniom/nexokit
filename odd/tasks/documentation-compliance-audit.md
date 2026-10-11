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
- [ ] T2 (in_progress): Scope role-permission mutations to trusted tenant context, with regression tests for same/cross tenant, root-global, selected-root and transaction failures as applicable. Verify focused IAM tests.
- [ ] T5 (pending): Reject absent tenant context in IAM handlers; assert no service calls and preserve explicit root/global/scoped contexts. Update affected HTTP integration wiring only when required.
- [ ] T6 (pending): Return 400 for malformed permission-update JSON and retain 422 for valid JSON with invalid fields; prove with handler regression tests.
- [ ] T3 (pending): Make generated modules conform to new-code layout, safe DTO mapping and normalization, trusted tenant scope and safe errors. Generate normal/tenant fixtures in temporary directories and prove compiled behavior with regression tests; update affected scaffold documentation accurately. Keep generator compatibility where safe; no unrelated module migration.
- [ ] T4 (pending): Run final independent checks, reconcile every finding and publish audit evidence/limitations for human validation. Full tests, vet, module guards, format, diff checks; record work-unit commits.

## Acceptance
- Each confirmed violation is corrected with observed applicable checks, or explicitly recorded as blocked.
- Generated DTOs belong to entity core, mappings expose only safe public fields, tenant operations cannot trust request company IDs, production errors do not expose technical text.
- Role association mutation cannot affect foreign-tenant or deleted roles.
- Final audit identifies checked and unchecked surfaces; green SQLite tests do not certify PostgreSQL/Redis.
- Changes are committed only on fix/documentation-compliance; no remote publication. User validates final result.

## Baseline evidence
Verifier: `go test -timeout 180s ./...`, `go test -short ./...`, `go vet ./...`, `make check-module-errors`, tracked Go `gofmt -l`, `git diff --check`: all pass. Integration uses SQLite; real PostgreSQL/migrations and Redis unverified. Working tree remained clean.

## Progress and next step
T2 in progress; remaining T1 source inventory runs read-only concurrently. T2 edits are restricted to the role-assignment slice and affected integration fixture. T1 complete; T2 writer active on fix/documentation-compliance. Additional confirmed defects added as T5/T6; generator is T3.
