# Redis limiter test isolation and main integration

## Objective
Fix repeat-run isolation of TestRedisLimiterIntegration, verify the full suite, and integrate fix/documentation-compliance into main locally.

## Authorization
User explicitly selected: fix Redis test on feature branch, commit correction, then merge feature branch into main. No push. Preserve all unrelated code and data. No Redis flush/config/service changes; delete only a uniquely owned test key in cleanup.

## Starting state
Branch fix/documentation-compliance, HEAD 11527e10b2dfcf37730b1d143e7da319c3c059f2, clean working tree. Main exists. Feature branch includes earlier documentation/generator/IAM commits in addition to auth; merge integrates that entire branch.

## Tasks
- [x] R1: Isolate Redis integration key and cleanup, preserving allow/deny/TTL assertions. Status: done; worker repeated5 and parent repeated2 real-Redis runs passed.
- [x] R2: Independently verify repeated real-Redis execution, full default suite, vet and formatting. Status: done; repeated10 realRedis tests and default full suite count2 passed.
- [ ] R3: Commit correction and merge feature into main locally; check final working tree and record delivery. Status: in progress; parent committing verified correction then fast-forwarding main.

## Evidence and checks
Prior read-only diagnosis: fixed key rl:login:127.0.0.1 uses 30s TTL and no cleanup. First run with absent key passes; repeated executions fail until natural expiry. Middleware code unchanged by auth. Observe focused RED before fix (existing repeat-run reproducer), then GREEN with count greater than one. Redis must be reachable for executed evidence, not just SKIP. Run full default go test ./... -count=1 after isolation fix, go vet ./..., diff/format checks. No artificial Redis reset.

## Scope
Production rate-limiter behavior unchanged. Allowed source edits: internal/middleware/rate_limit_test.go only. User branch changes and merge require parent execution; writer never stages/commits.

## R1 evidence
Baseline focused count2: first PASS, second FAIL on request1 (observed RED). Fix uses crypto/rand128-bit test prefix; exact-key deletion in bounded2s context runs before Close via LIFO cleanup and reports errors. Repeated5 tests PASS, no skips; middleware package and formatting/diff checks PASS. Parent repeated2 tests PASS against real Redis. No production edits, sleeps, shared-key deletes, flush/config/service changes. ASSESS unassessable due untracked task document with RDD off; conservative independent verifier required.

## Next step
Independent verifier passed focused realRedis count10 (no skips), default go test ./... -count=2, go vet ./..., git diff --check and clean gofmt scan. Only test file (22 additions/3 deletions) and this task document changed. Parent now commits and integrates via fast-forward, no push. Post-commit identities and final main/clean-tree evidence will be stored in the Engram mirror to avoid self-referential hashes.
