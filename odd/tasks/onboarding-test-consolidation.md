# Onboarding test consolidation

## Objective
Consolidate onboarding tests at behavior boundaries without losing relevant regression coverage.

## Scope and constraints
Move onboard_company/wiring_test.go into tests/integration/. Remove dto_test.go, errors_test.go and serialization_test.go after retaining necessary scenarios in handler_test.go or service_test.go. Leave model_test.go and production code unchanged. User explicitly authorized commits after verification. No publishing or parallel writes authorized. Current branch: refactor/companies-onboarding-verification.

## Plan
- [x] T1 (verified, delegated writer plus explicitly approved parent removals): Relocated integration test and consolidated relevant validation/error/response scenarios into handler tests. Production, model and service tests unchanged.
- [x] T2 (verified, delegated): Independently verified focused packages and reviewed coverage preservation.

## Acceptance criteria
The three specified auxiliary tests and slice wiring test are absent. An onboarding integration test resides under tests/integration/. Relevant request-validation, public error and response behaviors have clear remaining coverage. Existing model tests and production behavior remain unchanged.

## Verification
Worker: go test ./internal/modules/onboarding/slices/onboard_company; go test ./tests/integration; git diff --check.
Verifier: independently run the same commands and review changed tests against removed scenarios.
Test-first exception: test-only consolidation/relocation preserves existing behavior; baseline tests before edits and focused checks afterward are meaningful, artificial RED is not required.

## Routing and delivery
Multi-file write trigger requires one gentle-ai-worker. Verification requires gentle-ai-verify. One coherent test-maintenance work unit, forecast approximately 450–650 diff lines including removals and relocation. Commits authorized after verification; no PR/push authorized. Delivery-size decision deferred until PR preparation. Native review switch currently reported off by harness; no review started.

## Evidence and next step
Baseline passed: go test ./internal/modules/onboarding/slices/onboard_company (0.085s), go test ./tests/integration (0.498s). Writer cannot delete files; user explicitly selected parent-owned removals after writer consolidation and destination creation. Continued writer has edit surfaces limited to handler/service tests and tests/integration/onboarding_test.go. Worker completed handler coverage and created tests/integration/onboarding_test.go. Parent spot-checked destination and handler diff, then removed the four explicitly authorized originals. Worker final checks passed before removal (slice 0.085s, integration 0.529s; diff check clean). Native ASSESS unassessable due to intended-untracked declaration requirement; returned independent-verifier plan, treated as high. Independent post-removal checks all passed: go test -count=1 ./internal/modules/onboarding/slices/onboard_company (0.083s); go test -count=1 ./tests/integration (0.546s); go test -short -count=1 ./tests/integration (0.012s); git diff --check (clean). Explicit short-mode run confirmed onboarding test SKIP. Reviewer found no blockers or behavior regressions. HTTP tests retain seven validation scenarios, service non-invocation, five known wrapped-error responses, exact response fields and unchanged forwarding. Integration provisioning assertions preserved.

Intentional granularity loss: no direct sentinel code/prefix/English message checks or exact validator-message checks; these are not the HTTP workflow contract. Direct Validate invocation replaced by handler scenarios. git diff --check excludes untracked destination; verifier inspected it and Go tests compiled/executed it. Full repository suite, race detector and go vet were not run. Native assessment unavailable (untracked declaration), independent verification completed under high-risk fallback; RDD off, no native review started. Implementation and verification complete. User authorized commits. Work-unit evidence:
- 11972a6 — test(onboarding): move provisioning coverage to integration suite.
- 63e5df5 — test(onboarding): consolidate validation and response coverage in handler.
The consolidation commit has 428 diff lines, primarily removal of redundant tests, and remains one coherent work unit. No push or PR performed. Tracking evidence committed separately.
