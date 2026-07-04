# Tasks: Vertical-Slice & Platform/Shared Review (change-24)

> **Locked scope reminders.** No slice-folder migration. No public route, payload,
> HTTP status, DB migration, or business-feature change. Pure normalizers live in
> `internal/platform/shared/string`. GORM duplicate-key helper extends existing
> `internal/platform/gormutil`. `gormutil.IsUniqueConstraintError` MUST keep
> Postgres + SQLite/current behavior, including `nil`/generic/connection false
> cases. AppError public `Code` format is `code:<snake_case>`. Companies
> sentinels enumerated up-front (only refinements allowed to avoid duplicates).
> IAM `delete_role` inline `apperror.Wrap` moves to module-owned
> `ErrRoleHasAssignedUsers` sentinel. Auth tests pivot from
> `apperror.ErrUnauthorized` to module sentinels (HTTP 401 preserved). IAM
> `queries/normalize_slugs.go` is NOT moved. M3e audits `iam/internal/*/service.go`
> for GORM/apperror leaks. M6 grep guard targets handler/service/repository
> files, excluding tests. Duplicate `get_role_by_public_id_preloads.go` is
> deleted; preload behavior pinned by regression test.

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated changed lines (src + test) | ~3,300 net (+2,000 src / +1,300 test) across 12 work units |
| 800-line budget risk | **High** — every work unit under the budget, but cumulative single-PR diff is ~3,300 |
| 400-line budget risk | **High** — M2, M3a, M4 each estimated 400–540 net lines |
| Chained PRs recommended | **Yes** |
| Suggested split | 12 chained PRs (M0 → M1 → M2 → M3a → M3b → M3c → M3d → M3e → M4 → M5 → M6 → M7), each its own work-unit PR |
| Delivery strategy | `chained-pr` (user switched after workload forecast) |
| Chain strategy | `stacked-to-main` |

Decision needed before apply: **Yes**
Chained PRs recommended: **Yes**
Chain strategy: **stacked-to-main** — each work-unit PR merges to main in order
800-line budget risk: **High**
400-line budget risk: **High**

**Delivery decision.** The original `single-pr-default` strategy conflicted with
`review_budget_lines: 800` because the cumulative change is over budget. The
maintainer switched delivery to **chained PRs** with **stacked-to-main** strategy:
each work-unit PR merges to main in order before the next PR is prepared.

> Enumerated Companies sentinels (only refinements allowed):
> `code:company_not_found`,
> `code:company_domain_not_found`,
> `code:company_domain_duplicate`,
> `code:primary_domain_exists`,
> `code:company_domain_does_not_belong`.

## Phase 1: Foundation — Shared Helpers (M0, M1)

### M0 — Create `internal/platform/shared/string` package

- [x] M0.1 Create `internal/platform/shared/string/normalize_slug.go` exporting `func NormalizeSlug(s string) string` (returns `strings.ToLower(strings.TrimSpace(s))`).
- [x] M0.2 Create `internal/platform/shared/string/normalize_domain.go` exporting `func NormalizeDomain(s string) string` (`TrimSpace` → `ToLower` → `TrimSuffix(".")`).
- [x] M0.3 Create `internal/platform/shared/string/normalize_email.go` exporting `func NormalizeEmail(s string) string` (`TrimSpace` → `ToLower`).
- [x] M0.4 Create `internal/platform/shared/string/normalize_*_test.go` with one table-driven test per helper covering empty, whitespace, mixed case, trailing dot, control chars.

**Files likely touched.** `internal/platform/shared/string/{normalize_slug,normalize_domain,normalize_email,normalize_*_test}.go` (new).
**Acceptance criteria.** Helpers exist, compile, return documented values; no callers yet (additive).
**Tests.** `go test ./internal/platform/shared/string/...` — all rows pass.
**Verification commands.** `go vet ./... && go build ./... && go test ./internal/platform/shared/string/...`.
**Rollback boundary.** `git revert <M0-sha>` deletes the new package; nothing in the existing build depended on it.
**Review-size estimate.** ~120 src / ~200 test (~320 total).

### M1 — Extend `internal/platform/gormutil` with `IsUniqueConstraintError`

- [x] M1.1 Add `internal/platform/gormutil/unique.go` exporting `func IsUniqueConstraintError(err error) bool` that returns `true` for `errors.Is(err, gorm.ErrDuplicatedKey)` OR when the lower-cased message contains `duplicate key`, `unique constraint`, `unique failed`, or `constraint failed`. Returns `false` for `nil`, generic SQL errors, and connection errors. Never panics on `nil`.
- [x] M1.2 Add `internal/platform/gormutil/unique_test.go` table-driven test covering: `gorm.ErrDuplicatedKey`, Postgres `duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505)`, `unique constraint` substring, `UNIQUE constraint failed: users.email`, `constraint failed` lowercase, generic SQL error, `errors.New("connection refused")`, and `nil`.
- [x] M1.3 Leave `internal/modules/iam/queries/normalize_slugs.go` untouched (single-use plural de-dup helper, not the singular shared normalizer).

**Files likely touched.** `internal/platform/gormutil/unique.go`, `internal/platform/gormutil/unique_test.go`.
**Acceptance criteria.** `IsUniqueConstraintError` matches the locked scenario set (Postgres + SQLite/current, nil safe, generic/connection false).
**Tests.** Table-driven rows assert expected booleans per spec `gorm-helpers/spec.md`.
**Verification commands.** `go test ./internal/platform/gormutil/... -run Unique -v`.
**Rollback boundary.** `git revert <M1-sha>` removes the helper and its test; downstream consumers do not exist yet.
**Review-size estimate.** ~60 src / ~120 test (~180 total).

## Phase 2: Onboarding Module (M2) — Pattern Validation

### M2 — Migrate `onboarding` to `apperror` + shared helpers

- [x] M2.1 Rewrite `internal/modules/onboarding/core/error.go` to declare module-owned `Code` constants `code:duplicate_company_slug`, `code:duplicate_company_domain`, `code:duplicate_technical_domain`, `code:missing_platform_domain`, `code:duplicate_admin_email`, and build `Err*` sentinels via `apperror.Validation(...)`.
- [x] M2.2 Add `internal/modules/onboarding/core/errors_test.go` (table-driven) pinning `Status` (`422`), `Code` (`code:` prefix), and `PublicMessage` for every sentinel.
- [x] M2.3 Add `internal/modules/onboarding/core/dto_test.go` covering each `Validate()` rule per spec.
- [x] M2.4 Add `internal/modules/onboarding/core/model_test.go` (per-model `TableName` direct unit test).
- [x] M2.5 Remove `gorm.io/gorm` and `platform/apperror` imports from `internal/modules/onboarding/onboard_company/service.go`; move GORM translation into the slice repository (use `WithTx`/transaction method on the repository); use `string.NormalizeSlug`, `string.NormalizeDomain`, `string.NormalizeEmail` instead of the inlined copies.
- [x] M2.6 Replace the legacy `mapServiceError` in `onboard_company/handler.go` with a thin `respondOnboardingError` mapping that returns 422 field-keyed `ValidationErrorResponse` for the five known onboarding sentinels; funnel all other errors through `response.HandleError`; drop `platform/apperror` import.

> **M2 corrective fix note.** Fresh M2 verification found a public HTTP contract drift: the first implementation returned `409 Conflict` with a flat `ErrorResponse` because it combined `apperror.Conflict(...)` sentinels with direct `response.HandleError` routing. The corrective fix keeps the boundary improvements but restores the original contract by (1) switching sentinels to `apperror.Validation(...)` and (2) reintroducing a handler-side mapping to the original field-keyed 422 response. The drift is not accepted; the contract is preserved.

**Files likely touched.** `internal/modules/onboarding/core/{error,errors_test,dto_test,model_test}.go`, `internal/modules/onboarding/onboard_company/{service,handler,repository,service_test,handler_test}.go`.
**Acceptance criteria.** Service has no `gorm.`/`apperror.` imports; handler has no `mapServiceError` and preserves the original 422 field-keyed `ValidationErrorResponse` for the five known onboarding sentinels; `apperror.Status(core.ErrDuplicateCompanySlug) == 422`; HTTP envelope/status unchanged.
**Tests.** Module tests, `go test ./internal/modules/onboarding/...`, repository not-found + unique-violation tests.
**Verification commands.** `go vet ./... && go build ./... && go test ./internal/modules/onboarding/... && grep -RE 'gorm\.|apperror\.' internal/modules/onboarding/ --include='*service.go' --include='*handler.go' | grep -v _test.go` empty.
**Rollback boundary.** `git revert <M2-sha>`; no DB migration, no shared state.
**Review-size estimate.** ~260 src / ~180 test (~440 total).

## Phase 3: IAM Module (M3a–M3e)

### M3a — Migrate `iam/core` sentinels

- [x] M3a.1 Rewrite `internal/modules/iam/core/error.go` to declare 14 module-owned `Code` constants in `code:<snake_case>` format and construct `Err*` sentinels with `apperror.NotFound/Conflict/Forbidden/Unauthorized/BadRequest/Validation/Unprocessable`. Include `ErrRoleHasAssignedUsers = apperror.Unprocessable(CodeRoleHasAssignedUsers, core.MsgRoleHasAssignedUsers, nil)`.
- [x] M3a.2 Add `internal/modules/iam/core/errors_test.go` table-driven across all sentinels (Status, Code format `code:` prefix, PublicMessage).
- [x] M3a.3 Add `internal/modules/iam/core/dto_test.go` table-driven for each DTO `Validate()` rule.

**Files likely touched.** `internal/modules/iam/core/error.go`, `core/errors_test.go`, `core/dto_test.go`.
**Acceptance criteria.** `code:` prefix enforced; sentinels carry the right HTTP status; `ErrRoleHasAssignedUsers` is an `*AppError` with status 422 and `PublicMessage == MsgRoleHasAssignedUsers`.
**Tests.** Table-driven coverage; uniqueness of codes asserted (no duplicates).
**Verification commands.** `go test ./internal/modules/iam/core/... -v`.
**Rollback boundary.** `git revert <M3a-sha>`; sentinels are the only change in this step.
**Review-size estimate.** ~200 src / ~200 test (~400 total).

### M3b — Migrate `iam/users` slices (8)

- [x] M3b.1 For each of `iam/users/{create_user,update_user,delete_user,view_user,list_users,change_user_password,toggle_user_status,assign_role_to_user}/handler.go`: delete `mapServiceError`; call `response.HandleError(c, err)`; drop `platform/apperror` import.
- [x] M3b.2 Update `iam/users/create_user/repository.go` and `iam/users/update_user/repository.go` to call `gormutil.IsUniqueConstraintError` (M1) and translate to `core.ErrUserEmailAlreadyExists`.
- [x] M3b.3 Update slice services to return module sentinels only (`core.Err*` or `fmt.Errorf("...: %w", err)`); no `apperror.Wrap` inline.
- [x] M3b.4 Update each `service_test.go` to assert module sentinels via `errors.Is(err, core.Err*)`.

**Files likely touched.** 8 handlers + 8 services + 2 repositories + ~8 service tests + 2 repository tests.
**Acceptance criteria.** No `apperror.` in any `service.go`/`handler.go` under `iam/users/`; unique-violation test pins `core.ErrUserEmailAlreadyExists`; preload `Role`/`Company` preloads preserved.
**Tests.** Per-slice handler/service tests; not-found + unique-violation repo tests.
**Verification commands.** `go test ./internal/modules/iam/users/... && grep -RE 'apperror\.' internal/modules/iam/users/ --include='*service.go' --include='*handler.go' | grep -v _test.go` empty.
**Rollback boundary.** Per-PR revert; no DB schema touched.
**Review-size estimate.** ~180 src / ~120 test (~300 total).

### M3c — Migrate `iam/roles` slices (8)

- [x] M3c.1 Delete `mapServiceError` in each of the 8 role-slices' `handler.go`; funnel through `response.HandleError(c, err)`. Remove inline `apperror.Wrap(apperror.ErrUnprocessable, ...)` in `delete_role/handler.go` — the service already returns `core.ErrRoleHasAssignedUsers` after M3a.
- [x] M3c.2 Drop `platform/apperror` imports from role-slices handlers; no `apperror.Wrap` anywhere.
- [x] M3c.3 Update each role slice's `service_test.go`/`handler_test.go` to assert `core.Err*` sentinels (not `apperror.Err*`).

**Files likely touched.** 8 handlers + 8 handler tests (no service signature change).
**Acceptance criteria.** `delete_role` HTTP 422 path uses the new `ErrRoleHasAssignedUsers` sentinel; no `apperror.` import in any role-slice handler/service; HTTP envelope unchanged.
**Tests.** Per-slice handler test; `delete_role` regression test asserts the 422 path.
**Verification commands.** `go test ./internal/modules/iam/roles/... && grep -RE 'apperror\.|mapServiceError' internal/modules/iam/roles/ | grep -v _test.go` empty.
**Rollback boundary.** Per-PR revert; no DB or business behavior change.
**Review-size estimate.** ~160 src / ~120 test (~280 total).

### M3d — Migrate `iam/permissions` (3) + delete duplicate query

- [x] M3d.1 Delete `mapServiceError` in `iam/permissions/{list_permissions,update_permission,view_permission}/handler.go`; funnel through `response.HandleError(c, err)`.
- [x] M3d.2 Update `iam/permissions/update_permission/repository.go` to use `gormutil.IsUniqueConstraintError`; translate to module conflict sentinel.
- [x] M3d.3 Delete `internal/modules/iam/queries/get_role_by_public_id_preloads.go` (byte-identical duplicate).
- [x] M3d.4 Extend `internal/modules/iam/queries/get_role_by_public_id_test.go` with regression cases: role with `CompanyID` and seeded company row → `result.Company.ID == 1`; role with two `RolePermission` rows linked to two `Permission` rows → `result.Permissions` has length 2; not-found path returns `gorm.ErrRecordNotFound`.
- [x] M3d.5 Verify no caller references `GetRoleByPublicIDPreloads` (search and remove).

**Files likely touched.** 3 permission handlers, `update_permission/repository.go`, `get_role_by_public_id_test.go`, deletion of `get_role_by_public_id_preloads.go`.
**Acceptance criteria.** Preload regression test green; duplicate query file removed; no dangling callers of the deleted function.
**Tests.** Preload regression (Company + Permissions) + not-found path; per-slice handler tests.
**Verification commands.** `go test ./internal/modules/iam/queries/... ./internal/modules/iam/permissions/... && ! test -f internal/modules/iam/queries/get_role_by_public_id_preloads.go && ! grep -RE 'GetRoleByPublicIDPreloads' internal/modules/`.
**Rollback boundary.** Restore the deleted file from git history; regression test stays.
**Review-size estimate.** ~120 src / ~90 test (~210 total).

### M3e — Migrate `iam/internal` resolver slices (5) + audit

- [x] M3e.1 Audit each `iam/internal/*/service.go` for `gorm.io/gorm` and `platform/apperror` imports. Report findings in the PR description before changes.
- [x] M3e.2 For each slice with a leak: remove the import, translate persistence errors in the repository, return module sentinels or `fmt.Errorf("...: %w", err)`.
- [x] M3e.3 Delete any `mapServiceError` in `iam/internal/*/` (none expected — no HTTP layer).
- [x] M3e.4 Add per-slice regression tests for the 5 internal resolvers where missing.

**Files likely touched.** Up to 5 internal slice service files + their tests.
**Acceptance criteria.** Audit report attached to PR; no `apperror.` in `iam/internal/*/service.go`; no `gorm.io/gorm` in `iam/internal/*/service.go`.
**Tests.** `go test ./internal/modules/iam/internal/...`.
**Verification commands.** `go test ./internal/modules/iam/internal/... && grep -RE 'apperror\.|gorm\.' internal/modules/iam/internal/ --include='*service.go' | grep -v _test.go` empty.
**Rollback boundary.** Per-PR revert; no DB or HTTP behavior.
**Review-size estimate.** ~100 src / ~60 test (~160 total).

## Phase 4: Companies Module (M4)

### M4 — Migrate `companies` to `apperror` + shared helpers

- [x] M4.1 Rewrite `internal/modules/companies/core/error.go` with the enumerated sentinels (refine names only to avoid duplicates across modules): `code:company_not_found`, `code:company_domain_not_found`, `code:company_domain_duplicate`, `code:primary_domain_exists`, `code:company_domain_does_not_belong`. Build via `apperror.NotFound/Conflict/...`.
- [x] M4.2 Add `core/errors_test.go`, `core/dto_test.go`, and (since companies has partial GORM models) `core/model_test.go` (per-model `TableName` test).
- [x] M4.3 Remove `apperror` and `gorm.io/gorm` imports from each of the 7 companies services (`list_companies`, `view_company`, `update_company`, `delete_company`, `list_company_domains`, `create_company_domain`, `update_company_domain`); move `gorm.ErrRecordNotFound → core.ErrCompanyNotFound` translation into the slice repositories; use `string.NormalizeDomain` / `string.NormalizeSlug` instead of inlined copies; drop the local `NormalizeDomain` helper from `core/dto.go`.
- [x] M4.4 Delete `mapServiceError` in any companies handler that still has one.
- [x] M4.5 Update each slice's `service_test.go` to assert module sentinels via `errors.Is(err, core.Err*)`.

**Files likely touched.** 1 core file rewrite, 2 new core test files, 7 service files, 7 service tests, ~7 repositories, ~7 handler tests.
**Acceptance criteria.** No `apperror.`/`gorm.` import in any `companies/*/service.go`; companies sentinels are unique across modules; `apperror.Status(core.ErrDuplicateCompanyDomain) == 409`; HTTP envelope unchanged.
**Tests.** Per-slice handler + service tests; not-found + unique-violation repo tests.
**Verification commands.** `go test ./internal/modules/companies/... && grep -RE 'apperror\.|gorm\.' internal/modules/companies/ --include='*service.go' | grep -v _test.go` empty.
**Rollback boundary.** Per-PR revert; no DB schema, no public contract.
**Review-size estimate.** ~320 src / ~220 test (~540 total).

## Phase 5: Auth Module (M5)

### M5 — Migrate `auth` to `apperror`; pivot tests

- [x] M5.1 Rewrite `internal/modules/auth/core/error.go` with module-owned `Code` constants `code:invalid_credentials`, `code:invalid_refresh_token` and `apperror.Unauthorized(...)` sentinels.
- [x] M5.2 Add `auth/core/errors_test.go`, `auth/core/dto_test.go`, and `auth/core/model_test.go` (auth has partial GORM models — `AuthUser`, `AuthRole`, `RefreshToken`).
- [x] M5.3 Remove `gorm.io/gorm` and `platform/apperror` imports from `auth/{authenticate_user,revoke_token,rotate_token}/service.go`; move `gorm.ErrRecordNotFound → core.ErrInvalidCredentials` into the slice repository; service returns `core.ErrInvalidCredentials` / `core.ErrInvalidRefreshToken` only.
- [x] M5.4 Update `auth/{authenticate_user,revoke_token,rotate_token}/service_test.go` to pivot from `errors.Is(err, apperror.ErrUnauthorized)` to `errors.Is(err, core.ErrInvalidCredentials)` / `core.ErrInvalidRefreshToken`; HTTP 401 behavior must remain identical.
- [x] M5.5 View-session slice needs no behavior change (no GORM in service); ensure no `apperror` import leaked.
- [x] M5.6 **Corrective fix:** Preserve the original public 401 message `messages.MsgUnauthorized` (`"No autorizado"`) for both `core.ErrInvalidCredentials` and `core.ErrInvalidRefreshToken`; update `auth/core/errors_test.go`; add body-level envelope assertions to `auth/{authenticate_user,revoke_token,rotate_token}/handler_test.go`; update `apply-report-M5.md`.

**Files likely touched.** `auth/core/error.go`, 3 new core test files, 3 services, 3 service tests, 3 repositories.
**Acceptance criteria.** `apperror.Status(core.ErrInvalidCredentials) == 401`; tests assert module sentinels and HTTP 401 path; no `apperror.`/`gorm.` in `auth/*/service.go`.
**Tests.** Updated auth service tests + new core tests.
**Verification commands.** `go test ./internal/modules/auth/... && grep -RE 'apperror\.|gorm\.' internal/modules/auth/ --include='*service.go' | grep -v _test.go` empty.
**Rollback boundary.** Per-PR revert; no DB or HTTP contract.
**Review-size estimate.** ~200 src / ~160 test (~360 total).

## Phase 6: CI Grep Guard (M6)

### M6 — Wire `apperror` grep guard into Makefile + CI

- [x] M6.1 Add a Makefile target `check-module-errors` running:
  `grep -RE 'apperror\.' internal/modules/ --include='*service.go' --include='*repository.go' --include='*handler.go' | grep -v _test.go`
  that MUST return empty (zero exit on empty). Add a complementary guard for `gorm.` in `*service.go`.
- [x] M6.2 Add a `grep -RE 'mapServiceError' internal/modules/ | grep -v _test.go` empty guard.
- [x] M6.3 Wire `check-module-errors` into `.github/workflows/ci.yml` as a new `module-errors-guard` job that runs on every PR.
- [x] M6.4 Document the guard in `Makefile` comments and the `module-errors-guard` job's step name.

**Files likely touched.** `Makefile` (new target), `.github/workflows/ci.yml` (new job).
**Acceptance criteria.** All three grep commands empty after M2–M5 lands; CI fails when any handler/service/repository (non-test) under `internal/modules/` re-imports `apperror.` or re-introduces `mapServiceError`.
**Tests.** CI run on this PR proves the guard fires.
**Verification commands.** `make check-module-errors && grep -RE 'apperror\.' internal/modules/ --include='*service.go' --include='*repository.go' --include='*handler.go' | grep -v _test.go` empty.
**Rollback boundary.** `git revert <M6-sha>` disables the new job; nothing else changes.
**Review-size estimate.** ~25 lines (Makefile + CI).

## Phase 7: Documentation (M7)

### M7 — Publish `docs/module-error-conventions.md`

- [x] M7.1 Create `docs/module-error-conventions.md` with: per-module table `core.Err*` → `Code` (code: format) → `HTTPStatus` → `PublicMessage` → example call site, for `auth`, `iam`, `companies`, `onboarding`.
- [x] M7.2 Add a "Conventions" section listing the six rules (Code format, HTTP status, PublicMessage, reuse vs ad-hoc, wrapping, test coverage) per spec `module-error-conventions/spec.md`.
- [x] M7.3 Cross-link from `docs/modules/validation-and-errors.md` (single relative link in the relevant section).

**Files likely touched.** `docs/module-error-conventions.md` (new), `docs/modules/validation-and-errors.md` (one added link).
**Acceptance criteria.** Doc covers every sentinel introduced in M2–M5; validation-and-errors has a one-click link to the new doc; review checklist flags missing entries.
**Tests.** Manual review; CI grep guard already covers the runtime contract.
**Verification commands.** `grep -RE 'module-error-conventions' docs/` shows both files.
**Rollback boundary.** `git revert <M7-sha>`; pure docs.
**Review-size estimate.** ~80 lines (mostly the table).

## Implementation Order

M0 → M1 → M2 → M3a → M3b → M3c → M3d → M3e → M4 → M5 → M6 → M7.
M0/M1 are additive so M2 can be the smallest end-to-end migration that proves the
pattern. M3a unlocks the rest of iam (sentinels are prerequisites for M3b–M3e).
M4/M5 use the proven pattern. M6 closes the loop with CI, M7 with docs.

## Apply Gate (binding for the orchestrator)

`delivery_strategy: single-pr-default` + `review_budget_lines: 800` makes the
**cumulative** ~3,300-line change over budget. **Apply is BLOCKED until the
maintainer either:**

1. Approves `size:exception` to keep change-24 as a single PR, **or**
2. Switches `delivery_strategy` to `chained-pr` and confirms the chain strategy
   (`stacked-to-main` / `feature-branch-chain`).

No slice-folder migration is included in this change. No public route, payload,
HTTP status, DB migration, or business feature is changed.
