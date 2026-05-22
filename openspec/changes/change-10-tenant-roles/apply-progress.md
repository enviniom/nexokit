# Apply Progress: change-10-tenant-roles

## Work Units

| Unit | Scope | Evidence status |
| --- | --- | --- |
| PR 1 | Migration, role model `company_id`, tenant-aware repository foundation | Completed before this fix pass; RED/GREEN evidence reconstructed from prior session summaries and committed history. |
| PR 2 | Service tenant isolation, handler tenant context wiring, DTO exposure | Completed before this fix pass; RED/GREEN evidence reconstructed from prior session summaries and committed history. |
| PR 3 | Seed changes and tenant role test updates | Completed before this fix pass; reviewer reported `go test ./...` and `go build ./...` passing. |
| Verification blocker fix | Reserved identity 422, public company ID DTO mapping, dead seed helper removal, documentation evidence | Live strict-TDD evidence recorded below. |

## TDD Cycle Evidence

| Cycle | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- |
| PR 1 foundation | Reconstructed: repository/model changes were driven by tenant-role tasks 1-2 and compile/test failures after signature changes. Exact live RED output was not preserved in this artifact. | Reconstructed: PR1 committed as `c8c8e86`; subsequent full build was fixed with app container adapter per recorded session summary. | Repository methods were updated consistently for list/count/get/delete/name/slug paths. | Foundation was isolated as a reviewable work-unit commit. |
| PR 2 service/handler wiring | Reconstructed: service/handler tests and consumers failed while repository signatures changed to require `tenant.TenantContext`. Exact live RED output was not preserved in this artifact. | Reconstructed: PR2 committed as `ad677ce`; session summary records full build/test success after tenant context wiring. | Handler pass-through tests covered scoped and root tenant contexts. | Service reserved identity helper was centralized as `isReservedIdentity`. |
| PR 3 seeds/tests | Reconstructed: seed and role tests were updated after seed behavior changed from root/admin/user to root-only. Exact live RED output was not preserved in this artifact. | Reviewer reported `git diff --check`, `go test ./seeds ./internal/modules/roles`, `go test ./...`, and `go build ./...` passing for PR3. | Added service tests for tenant-scoped list/get/create/update/delete behavior and handler tenant pass-through. | PR3 was committed separately as `079b49c`. |
| Verification blocker fix | Live RED: after updating tests first, `go test ./internal/modules/roles` failed with `unknown field Company in struct literal of type Role` and `mismatched types uint and string`, proving the public-company-ID DTO contract was not implemented. A later fresh review found one remaining pre-DB validation gap for reserved update requests. | Live GREEN: `go test ./internal/modules/roles ./seeds` passed after implementation fixes; final focused validation also passed with `go test -count=1 ./internal/modules/roles ./seeds`. | Full validation passed with `go test ./...` and `go build ./...`; final `git diff --check` passed. | Removed obsolete admin/user permission helper tests and functions; kept system-role forbidden behavior separate from reserved-identity validation; moved requested reserved identity validation before update lookup. |

## Review Workload Boundary

Initial planning forecasted 550-650 changed lines with `delivery_strategy: auto-chain` and a `feature-branch-chain` recommendation. The implementation was kept as chained work-unit commits on the same branch instead of separate branch PRs:

1. `c8c8e86 feat(roles): add tenant-scoped role foundation` — PR1-equivalent foundation slice.
2. `ad677ce feat(roles): wire tenant context through service and handlers` — PR2-equivalent service/handler slice.
3. `079b49c feat(roles): update seeds and tenant role tests` — PR3-equivalent seed/test slice.
4. Verification blocker fix — current uncommitted fix pass.

`size:exception`: accepted for the local branch shape only. Reviewer-facing delivery SHOULD still present the commits as chained/reviewable slices or explain that the branch contains multiple work-unit commits plus one Pi runtime config commit (`a67a385`) that is not part of the product change.

## Latest Validation

- `go test ./internal/modules/roles` — RED during test-first blocker fix, failed before implementation as expected.
- `go test ./internal/modules/roles ./seeds` — PASS.
- Fresh-review fix: requested reserved identity validation now happens before update lookup, with service coverage for repository-error lookup returning `ErrValidation` first.
- Final `go test -count=1 ./internal/modules/roles ./seeds` — PASS.
- Final `go test ./...` — PASS.
- Final `go build ./...` — PASS.
- Final `git diff --check` — PASS.
