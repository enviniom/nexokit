# Verification Report: change-10-tenant-roles

Status: **PASS after verification blocker fix pass**

## Summary

The previous verification run failed on spec/TDD/workload blockers even though tests and build were green. This follow-up fix pass addressed the blockers in the approved scope and reran validation.

## Blocker Resolution

1. **Reserved identity HTTP contract** — fixed.
   - `Create` now returns `apperror.ErrValidation` for reserved `root`, `admin`, or `user` name/slug.
   - `Update` now returns `apperror.ErrValidation` when either the existing role identity or requested identity is reserved.
   - Handler tests now assert HTTP 422 for reserved identity create/update paths.

2. **Role DTO `company_id` contract** — fixed with the smallest safe mapping.
   - `RoleResponse.CompanyID` is now `*string`.
   - Roles preload `Company` in repository read paths.
   - DTO mapping emits `Company.PublicID` when available and omits `company_id` for global roles or non-preloaded company data.
   - The role model now declares the `Company` relationship through `CompanyID`.

3. **Seed helper cleanup** — fixed.
   - Removed obsolete `allSystemPermissionSlugs`, `adminPermissionSlugs`, and `userPermissionSlugs` helper code.
   - Removed tests that targeted the obsolete admin permission helper.

4. **Strict TDD evidence** — fixed.
   - Added `openspec/changes/change-10-tenant-roles/apply-progress.md` with reconstructed evidence for prior slices and live RED/GREEN/TRIANGULATE/REFACTOR evidence for this fix pass.

5. **Review workload boundary** — documented.
   - Added chained work-unit commit boundary and local-branch `size:exception` note in `apply-progress.md`.

## Validation Commands

- `go test ./internal/modules/roles` — failed during the test-first RED step before implementation, as expected.
- `go test ./internal/modules/roles ./seeds` — PASS.
- `go test ./...` — PASS.
- `go build ./...` — PASS.

## Remaining Notes

- The `size:exception` is limited to the current local branch shape. Reviewer-facing delivery should still present the product commits as chained/reviewable slices and keep the Pi runtime config commit separate from the tenant-role product change.
- Pi-lens created local `.pi-lens/` runtime cache files during analysis; they are not part of this change and were not modified intentionally.
