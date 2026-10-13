# Tenant administrator access

## Objective
Root retains global authority, admin has full authority within its tenant without permission grants, and other roles are permission-controlled. Route access belongs in middleware; handlers transport trusted context; repositories enforce tenant data scope.

## Current user-approved policy
- Root must respect an explicitly selected `X-Company-ID`, including company view. Without a selection, existing root-global routes may remain global.
- Remove `RootGlobalCompanyView` and its route/container wiring; do not replace it with another endpoint-specific exception.
- Company identity filtering remains `companies.id = TenantContext.CompanyID`, not `ApplyTenantScope`, whose column is `company_id`.
- Reserved `root`, `admin`, `user` identities prevent API creation/renaming/deletion according to existing rules. Reservation is NOT permission-assignment immutability.
- Permission assignments of target roles `root` and `admin` cannot be changed by any actor: their authority is implicit.
- Admin may assign/remove permissions of `user` and all other roles in its tenant, including `IsSystem=true` roles and system permission grants. Cross-tenant role mutation stays forbidden.
- Remove admin-specific system-role restrictions and the later system-permission-removal restriction from assignment. Do not change global permission catalog mutation: root-only route middleware remains.
- Remove assignment's actor argument if no business rule still needs it; authorization is already route middleware-owned.
- Platform roles remains the sole application role constant/reserved-identity owner. Do not alter unrelated role creation/update/delete protections.

## Constraints
- User originally prohibited Git mutations, then explicitly authorized commits after final verification and completion of starter preparation. Stage/commit only scoped verified work-unit files; preserve unrelated user changes. No reset/restore/stash/push/PR/recovery.
- User is actively reviewing/staging files; concurrent user index changes are expected and must not be attributed to agents or reverted.
- Starter initialization is authorized as the next separately tracked feature after this verified permissions work-unit commit. sqlc and an upstream update mechanism remain out of scope.
- No `.env` reads, dependencies, migrations, or unrelated formatting/refactoring.
- Preserve existing changes, including staged content; source edits remain unstaged unless the user stages them.

## Tasks
- [x] T1 — Implement and verify admin tenant authority and the clarified root/role policy. **Verified and committed in `a662622` (`fix(auth): grant tenant admins full scoped authority`).**
- [x] T2 — Commit the verified permissions work unit after final verification. **Completed: scoped staging synchronized with tested worktree; commit `a662622`.**

## Acceptance and checks
- Admin with zero grants can manage its tenant, including `user` and custom/system role permissions.
- Adding/removing synchronized system permissions to/from `user` is allowed; root/admin target assignments remain unchanged on denial.
- Ordinary actors without grants are denied by middleware; foreign roles cannot be read/mutated by admin.
- Root selecting A cannot view company B; selecting B can view B; no-header root-global behavior remains where already permitted.
- No endpoint-specific root-scope override or assignment actor/permission authorization duplication remains.
- Existing immutable identity and global catalog route guards remain separate from assignment policy.
- Test-first: observe new policy failures before source changes, focused GREEN, then independent full tests/build/vet/module guard and structural readback.

## Prior observed evidence
- Initial RED: admin without grants got 403; foreign company view got 200; secondary assignment service gate denied admin.
- Prior focused/full tests, builds, vet and `make check-module-errors` passed after bounded corrections.
- Architecture correction passed: roles centralized in platform, catalog access moved to route role middleware, company ownership moved into repository query, duplicated service permission gate removed.
- Prior implementation included root company-view override and role-system restrictions; user explicitly rejected these, so prior GREEN does not establish acceptance of current policy.
- RDD switch is off. ASSESS repeatedly returned unassessable because untracked files require explicit declaration; followed high-risk-equivalent independent verification plan without starting review.
- Added integration tests exercise DB hydration and production registrars, not bearer-token transport.
- Baseline `gofmt -l .` debt in unchanged `internal/modules/onboarding/slices/onboard_company/handler_test.go` remains unrelated.
- IAM `core/constants.go` and `core/role_identity.go` were left package-only after relocation; cleanup is not part of this policy task.

## Git review clarification
User confirmed staging company view handler/test while reviewing; earlier index-check discrepancy is resolved as intentional user activity. No recovery needed or authorized. Parent/agents issued no Git mutation commands. Latest read-only snapshot: branch `feat/tenant-admin-access`, HEAD `761a41d6344eb5215347661b0befd40ad1aeb3ec`, index SHA256 `903058a22dbb86b6a56c8120b2c50054326a49888043a197bf00c5f3b6bd0d5c`. This is an observation, not a freeze of user activity.

## Current implementation evidence
- Writer removed `RootGlobalCompanyView`, companyViewScope registration plumbing, assignment actor argument, admin/system restriction, and RemovesSystemPermission interface/implementation/fakes.
- Only normalized root/admin target identities reject grant mutation; user/custom/system role grants are editable with tenant-scoped persistence.
- RED before source edits: role suite rejected editable user/system targets and removal of system permissions; integration returned 403 instead of 200 for system assignment and 200 instead of 404 for root-selected foreign company.
- Focused GREEN: platform roles/middleware, companies packages, IAM role slices, fresh integration regression run, gofmt of changed files, and diff check all passed. An intermediate fixture collision was isolated and checks rerun successfully.
- Parent read assignment service, confirmed no remaining override/removal-helper symbols and reran assignment tests successfully.
- ASSESS now classified the aggregate current changes as high risk (56 paths, 987 diff lines), RDD off, requiring independent verification; no native review started. Native continuation reported verbatim: `gentle-ai review status --cwd=/home/enviniom/Proyectos/Go/nexokit --contract=gentle-ai.review-integration/v2 --next-transition=true`.
- Writer and parent read-only before/after snapshots match branch/HEAD and index SHA256 `903058a22dbb86b6a56c8120b2c50054326a49888043a197bf00c5f3b6bd0d5c`. No Git mutation commands executed; user retains staging control.

## Final independent verification
All required focused packages, fresh verbose integration filter, `go test ./...`, `go build ./...`, `go vet ./...`, `make check-module-errors`, and diff checks passed. Fresh integration: five top-level tests PASS, zero failures/skips. Readback confirmed override plumbing and obsolete helper/actor removed, selected-root tenant respected, user/system grant removal persisted, protected/foreign joins unchanged and ordinary no-grant actor denied. No candidate-caused blockers. Baseline onboarding format debt and package-only IAM files remain separate. Added integration still excludes bearer transport. Index observation hashes used by verifier were not comparable to index-file checksums; no attribution/recovery needed. Scoped staging must reflect tested worktree, including removal of obsolete override hunks staged in three files.

## Next step
Permissions work unit completed and committed as `a662622` after independent verification. Commit changed 56 scoped files (+860/-139 including tests/docs/tracking); much of the path count is mechanical role-constant relocation. Continue starter bootstrap as a separate tracked feature. No push/PR authorized.
