# Exploration: change-24-vertical-slice-platform-review

> **Scope of this artifact.** Audit-oriented exploration only. No code is modified. The output is a per-module violation table, a list of cross-module shared-helper candidates, and a module-by-module iteration plan (M1..MN) sized for a 800-line review budget per PR.

## TL;DR

The four production modules (`auth`, `iam`, `companies`, `onboarding`) have a strong vertical-slice skeleton (handler / service / repository / queries co-located per use case with tests), but they all miss the **error migration** required by the new `platform/apperror` redesign (change-23) and they have **boundary drift** in two specific places:

1. **Every `core/error.go`** still uses `errors.New(...)` instead of `apperror.*` sentinels with module-owned `Code` constants.
2. **Every `slices/<slice>/handler.go`** has a `mapServiceError(err)` switch that re-imports `platform/apperror` and converts `core.Err*` to `apperror.Err*` — violating the rule that handlers must funnel through `response.HandleError(c, err)` directly.
3. **Several `slices/<slice>/service.go`** import `gorm.io/gorm` and `apperror.Err*` directly (auth, onboarding, companies), instead of returning module errors from `core/errors.go`.

Shared helpers that belong in a new `platform/shared/string` (or similar) — duplicated across **3+ modules**:

| Helper | Duplicated in | Used in |
|---|---|---|
| `NormalizeSlug(s) → TrimSpace + ToLower` | iam/queries/normalize_slugs.go (slice variant), iam/core/role_identity.go, onboarding/onboard_company/service.go, companies/update_company/service.go, companies/core/dto.go | 4 modules |
| `NormalizeDomain(s) → TrimSpace + ToLower + TrimSuffix(".")` | onboarding/onboard_company/service.go, companies/create_company_domain/service.go, companies/update_company_domain/service.go | 2 modules |
| `NormalizeEmail(s) → TrimSpace + ToLower` | onboarding/onboard_company/service.go | 1 module (extract when needed by others) |
| `IsUniqueConstraintError(err)` | iam/users/create_user/repository.go, iam/users/update_user/repository.go, iam/permissions/update_permission/repository.go | 1 module (3 sites) — extract as `gormutil.IsUniqueConstraintError` |

> **Note on `platform/shared`.** The current tree has no `internal/platform/shared` directory. The plan creates it as part of M0 (or piggybacks on M1). It is the home for the pure helpers above and for future shared value objects.

> **Note on slice placement.** The current code (auth, iam, companies, onboarding) places slice directories at the **module root** (`<module>/<slice>/`). The canonical spec `openspec/specs/vertical-slice-modules/spec.md` explicitly endorses this layout. The newer `docs/modules/vertical-slices.md` and `docs/modules/module-structure.md` say slices MUST live under `<module>/slices/`. **Spec and docs disagree.** This audit treats the **docs** as authoritative for change-24 (per the prompt), but the disagreement is a **preflight risk** that must be resolved before MN is signed off.

---

## 1. Module inventory

Source of truth: `internal/modules/*` on `main`. `go build ./...` and `go vet ./...` are clean at the start of this audit.

| Module | Shape today | LOC (src) | Slices | Has `core/` | Has `queries/` | Slice placement |
|---|---|---|---|---|---|---|
| `auth` | flat at module root | ~1 769 | 4 (`authenticate_user`, `rotate_token`, `revoke_token`, `view_session`) | yes | yes | module root |
| `iam` | split into `users/`, `roles/`, `permissions/`, `internal/` sub-packages, each with its own `container.go` and `routes.go` | ~10 432 | 22 use-case slices across 3 entities + 5 internal (no-HTTP) slices | yes (shared) | yes | NOT under `slices/`; siblings of `iam/core` |
| `companies` | flat at module root | ~1 967 | 7 (`list_companies`, `view_company`, `update_company`, `delete_company`, `list_company_domains`, `create_company_domain`, `update_company_domain`) | yes | yes | module root |
| `onboarding` | flat at module root | ~1 116 | 1 (`onboard_company`) | yes | yes | module root |
| `iam/internal` (no-HTTP resolver slices) | n/a | n/a | 5 (`resolve_auth_user`, `resolve_role_by_slug`, `resolve_user_permissions`, `sync_permissions`, `list_all_permissions`) | shares `iam/core` | n/a | inside `iam/internal/`, not under `slices/` |

Slice totals: **31 handlers**, **36 services**, **36 repositories**, **21 query files**, **124 test files** for **~15 284 LOC** in `internal/modules/`.

---

## 2. Per-module violation table

Severity scale: **Critical** (breaks the apperror contract and is the explicit ask of change-24), **High** (boundary leak that the spec forbids), **Medium** (test gap or duplication that is a follow-up), **Low** (cosmetic / refactor only).

### 2.1 `auth` module

| # | File | Violation | Severity | Fix direction |
|---|---|---|---|---|
| 1 | `internal/modules/auth/core/error.go` | Errors declared as plain `errors.New(...)`; no `apperror` helpers, no module-owned `Code` constants. | Critical | Rewrite as `apperror.Unauthorized(CodeInvalidCredentials, "invalid credentials", nil)` sentinels. |
| 2 | `internal/modules/auth/authenticate_user/service.go` (lines 8, 10, 36, 41, 44) | Imports `apperror` and `gorm.io/gorm`; returns `apperror.ErrUnauthorized` directly; uses `errors.Is(err, gorm.ErrRecordNotFound)` to map to apperror inside the service. | Critical | Service must return `core.ErrInvalidCredentials`; repository must do the `gorm.ErrRecordNotFound → core.ErrInvalidCredentials` translation. |
| 3 | `internal/modules/auth/revoke_token/service.go`, `auth/rotate_token/service.go` | Same pattern as #2 (`apperror.ErrUnauthorized`, GORM import). | Critical | Same as #2. |
| 4 | `internal/modules/auth/authenticate_user/handler.go`, `revoke_token/handler.go`, `rotate_token/handler.go` | Import `apperror`; some do `apperror.ErrUnauthorized` re-mapping. | High | Funnel via `response.HandleError(c, err)`; the handler must not import `apperror` at all (only `core` + `response`). |
| 5 | `internal/modules/auth/core/model_test.go` exists; `core/error_test.go` and `core/dto_test.go` are missing. | High (testing.md requires module-error and DTO-validation tests for every module). | Add `core/error_test.go` (table-driven, one row per sentinel) and `core/dto_test.go` (table-driven over each `Validate()` rule). |
| 6 | Slices live at module root instead of under `auth/slices/`. | Low (spec is permissive; docs disagree — flagged separately). | Defer to final close-out change (M5+). |
| 7 | Slices are well-tested (handler + service + repository tests). | n/a | Keep. |

### 2.2 `iam` module

This is the **largest** module (~10.4k LOC, 22 use-case slices + 5 internal slices, 3 entity sub-containers). The doc asks for the same audit and the same migration, so the work breaks down naturally by entity.

| # | File | Violation | Severity | Fix direction |
|---|---|---|---|---|
| 1 | `internal/modules/iam/core/error.go` (entire file) | 14 sentinels declared as `errors.New(...)`; no `apperror` helpers, no `Code` constants. | Critical | Rewrite as `apperror.NotFound(CodeUserNotFound, ...)` / `Conflict` / `Forbidden` / `BadRequest` sentinels with a module-owned `Code` block at the top of the file. |
| 2 | `internal/modules/iam/users/<slice>/handler.go` (8 files) | Each one has a `mapServiceError(err error) error` switch that imports `apperror` and converts `core.Err*` → `apperror.Err*`. | Critical | Delete the switch. Module errors are already `*apperror.AppError` after #1, so `response.HandleError(c, err)` works as-is. |
| 3 | `internal/modules/iam/users/assign_role_to_user/handler.go` (line 49), `delete_user/handler.go`, `view_user/handler.go`, `toggle_user_status/handler.go`, `update_user/handler.go`, `change_user_password/handler.go`, `create_user/handler.go` | Same as #2. | Critical | Same as #2. |
| 4 | `internal/modules/iam/roles/<slice>/handler.go` (8 files) | Same as #2/3. | Critical | Same. |
| 5 | `internal/modules/iam/permissions/<slice>/handler.go` (3 files) | Same as #2/3. | Critical | Same. |
| 6 | `internal/modules/iam/users/create_user/service.go` (and 7 other `users/<slice>/service.go`) | Service does NOT import GORM (good) and does NOT import `apperror` (good) — already correct. Returns `core.Err*` sentinels. | n/a | Keep pattern; update only the **type** of the returned errors after #1 (sentinel swap, no code-shape change). |
| 7 | `internal/modules/iam/users/create_user/repository.go` and `update_user/repository.go` (lines 86-92) and `permissions/update_permission/repository.go` (line 42) | Each implements its own `isUniqueConstraintError(err)` helper that lower-cases `err.Error()` and substring-matches. | Medium (duplication; works but fragile) | Extract to `platform/gormutil` (or a new `gormutil.IsUniqueConstraintError`) and reuse. |
| 8 | `internal/modules/iam/queries/get_role_by_public_id.go` and `get_role_by_public_id_preloads.go` | The two files are **byte-identical** except for the function name. | High (dead code) | Delete `get_role_by_public_id_preloads.go` (or merge into one and keep only the named one); add a test that covers preloads if not already covered. |
| 9 | `internal/modules/iam/queries/normalize_slugs.go` | A `NormalizeSlugs([]string) []string` helper (dedup + trim) used only by `iam/roles/assign_permissions_to_role/repository.go`. | Low | Stays in `iam/queries/` because it is **single-use**; tests live next to it. Promote only if a second module needs it. |
| 10 | `internal/modules/iam/core/role_identity.go` | `IsReservedRoleIdentity` does `strings.TrimSpace` on name and slug locally. | Low | Reuse the new shared `NormalizeSlug` from `platform/shared/string` (M1). |
| 11 | `internal/modules/iam/users/create_user/repository.go` (line 90), `update_user/repository.go` (line 90), `permissions/update_permission/repository.go` (line 42) | `strings.ToLower(err.Error())` for GORM error inspection. | Medium (same as #7) | Move to `gormutil.IsUniqueConstraintError`. |
| 12 | `internal/modules/iam/users/`, `roles/`, `permissions/`, `internal/` | Slice directories are siblings of `iam/core`, not under `iam/slices/<entity>/`. | Low (spec-vs-docs disagreement) | Out of scope for change-24 unless the orchestrator resolves the disagreement first. |
| 13 | `internal/modules/iam/core/dto_test.go` is missing; `core/error_test.go` is missing. | High (testing.md) | Add during M3 (users) / M3 (roles) / M3 (permissions) — the M-steps that touch each `core/error.go`. |

### 2.3 `companies` module

| # | File | Violation | Severity | Fix direction |
|---|---|---|---|---|
| 1 | `internal/modules/companies/core/error.go` (entire file) | Sentinels declared as `errors.New(...)`. | Critical | Rewrite with `apperror.Conflict` / `apperror.NotFound` / etc. + module-owned `Code` block. |
| 2 | `internal/modules/companies/create_company_domain/service.go` (line 8, line 23) and `update_company_domain/service.go`, `list_company_domains/service.go`, `view_company/service.go`, `update_company/service.go`, `delete_company/service.go` (10 services total) | Import `apperror` and `gorm.io/gorm`; map `gorm.ErrRecordNotFound → apperror.ErrNotFound` **inside the service**. | Critical | Move the GORM-error translation into the slice repository; service returns module errors only. |
| 3 | `internal/modules/companies/create_company_domain/service.go` (line 28) and `update_company_domain/service.go` (line 36) | `domain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(req.Domain)), ".")` duplicated verbatim. | Medium | Promote `NormalizeDomain(s) string` to `platform/shared/string` (M1). |
| 4 | `internal/modules/companies/update_company/service.go` (line 28) | `slug := strings.ToLower(strings.TrimSpace(req.Slug))` — third site of the same normalizer. | Medium | Promote `NormalizeSlug(s) string` to `platform/shared/string` (M1). |
| 5 | `internal/modules/companies/core/dto.go` (lines 119-120) | Helper `NormalizeDomain` defined inside DTO file. | Low (not a bug, but a duplicate) | Replace with shared helper. |
| 6 | `internal/modules/companies/core/model_test.go` is missing; `core/error_test.go` and `core/dto_test.go` are missing. | High (testing.md) | Add during M4. |

### 2.4 `onboarding` module

| # | File | Violation | Severity | Fix direction |
|---|---|---|---|---|
| 1 | `internal/modules/onboarding/core/error.go` (5 sentinels) | Plain `errors.New(...)`. | Critical | Rewrite with `apperror.Conflict` + module-owned `Code` block. |
| 2 | `internal/modules/onboarding/onboard_company/service.go` (lines 8-9, 29, 35, 41, 62, 137) | Imports `apperror` and `gorm.io/gorm`; uses `apperror.ErrValidation` directly; has duplicated `normalizeDomain` (line 137), `strings.ToLower(strings.TrimSpace(...))` for slug (line 35) and email (line 62). | Critical | Move GORM logic out of the service (or accept a `*gorm.DB` in the repository and call the service with `context.Context` only). Use the new shared `NormalizeDomain` / `NormalizeSlug` / `NormalizeEmail`. |
| 3 | `internal/modules/onboarding/core/model_test.go` is missing; `core/error_test.go` and `core/dto_test.go` are missing. | High (testing.md) | Add during M2. |

---

## 3. Cross-module shared-helper extraction plan

| Target location | Helper signature | Replaces | Used by (post-extraction) |
|---|---|---|---|
| `internal/platform/shared/string/normalize_slug.go` | `func NormalizeSlug(s string) string` — returns `strings.ToLower(strings.TrimSpace(s))`. | `iam/queries/normalize_slugs.go` (one-element path), `iam/core/role_identity.go`, `onboarding/onboard_company/service.go` line 35, `companies/update_company/service.go` line 28, `companies/core/dto.go` line 119 | 4 modules |
| `internal/platform/shared/string/normalize_domain.go` | `func NormalizeDomain(s string) string` — `strings.TrimSuffix(strings.ToLower(strings.TrimSpace(s)), ".")`. | `onboarding/onboard_company/service.go` line 137, `companies/create_company_domain/service.go` line 28, `companies/update_company_domain/service.go` line 36, `companies/core/dto.go` line 120 | 2 modules (3 sites) |
| `internal/platform/shared/string/normalize_email.go` | `func NormalizeEmail(s string) string` — `strings.ToLower(strings.TrimSpace(s))`. | `onboarding/onboard_company/service.go` line 62 | 1 module (extract anyway; trivially small and likely to grow) |
| `internal/platform/gormutil/unique.go` (extend existing `gormutil.go`) | `func IsUniqueConstraintError(err error) bool` | `iam/users/create_user/repository.go` lines 86-92, `iam/users/update_user/repository.go` lines 86-92, `iam/permissions/update_permission/repository.go` line 42 | 1 module (3 sites) |
| `internal/platform/shared/slug` (optional, future) | `func Slugify(s string) string` — full text-to-slug with diacritics strip + separator + collision. | none found today | 0 modules |

> **Why not introduce `platform/shared/slug` (full text-to-slug) right now?** The current codebase has **no** text-to-slug function (search for `slugify` / `Slugify` / `ToSlug` returned 0 hits in production code; `NormalizeSlugs` in `iam/queries/` is a *list* normalizer). The prompt mentions it as a future candidate, but there is no consumer today. Do not over-engineer M0; add it when the first consumer appears.

---

## 4. Iteration plan M1..MN

Each step is one PR, with its own reviewable change set. The 800-line review budget is the binding constraint. Numbers below are **gross** (src + test), rounded up.

| Step | Title | Touches | Approx net lines | Review risk |
|---|---|---|---|---|
| **M0** | Create `internal/platform/shared/{string,gormutil}` skeletons + tests. | new files | +120 src / +200 test | Low (additive; no callers yet) |
| **M1** | Wire `platform/shared/string` helpers into **companies** and **onboarding**; delete duplicate `normalizeDomain` / `NormalizeSlug` in DTO files and services. | companies (3 sites), onboarding (3 sites), platform/shared | +60 src / +120 test | Low–Medium |
| **M2** | Migrate `onboarding/core/error.go` to `apperror`; update `onboard_company` service/handler; add `core/error_test.go` + `core/dto_test.go` + `core/model_test.go` (TableName). | onboarding (~6 src + 3 test files) | +260 src / +180 test | Medium (transaction code path, validation rules) |
| **M3a** | Migrate `iam/core/error.go` to `apperror`; add `core/error_test.go` + `core/dto_test.go`. | iam/core | +200 src / +200 test | Medium (large sentinel table) |
| **M3b** | Migrate `iam/users/*` (8 slices): delete `mapServiceError` in each handler; switch to `response.HandleError`; extract `gormutil.IsUniqueConstraintError` and reuse. | iam/users (~16 src + 8 test updates) | +180 src / +120 test | Medium |
| **M3c** | Migrate `iam/roles/*` (8 slices): same as M3b. | iam/roles | +160 src / +120 test | Medium |
| **M3d** | Migrate `iam/permissions/*` (3 slices) + delete duplicate `get_role_by_public_id_preloads.go` (file-level dead-code removal). | iam/permissions, iam/queries | +120 src / +90 test | Low–Medium |
| **M3e** | Migrate `iam/internal/*` (5 no-HTTP resolver slices). | iam/internal | +100 src / +60 test | Low |
| **M4** | Migrate `companies/core/error.go` to `apperror`; remove `apperror`/`gorm` imports from each service; update each repository to translate `gorm.ErrRecordNotFound → core.Err*`; add `core/error_test.go` + `core/dto_test.go` + `core/model_test.go`. | companies (~14 src + 6 test files) | +320 src / +220 test | Medium–High (10 services, transaction paths in onboarding that point at company-domain lookups) |
| **M5** | Migrate `auth/core/error.go`; remove `apperror`/`gorm` imports from auth services; add `core/error_test.go` + `core/dto_test.go` + `core/model_test.go` (partial GORM models). | auth (~12 src + 4 test files) | +200 src / +160 test | Medium |
| **M6** | Add CI grep guard: `grep -RE 'apperror\.' internal/modules/ | grep -v _test.go` must show apperror references **only** in `core/errors.go` and `handler.go` is **forbidden**. Wire into Makefile / CI. | Makefile, CI config | +20 src | Low |
| **M7** | Publish `docs/module-error-conventions.md` (canonical table: `core.Error` → `Code` → `HTTPStatus` → example). Cross-link from `docs/modules/validation-and-errors.md`. | docs | +1 doc | Low |

> **Order rationale.** M0/M1 ship the shared helpers early so M2..M5 can consume them. M2 (onboarding) is the **smallest** end-to-end migration and validates the pattern; M3a (iam core) is the largest sentinel table and unlocks the rest of iam in M3b..M3e. M4/M5 use the pattern already proven. M6/M7 close the change.

> **Spec-vs-docs disagreement (re-stated).** This plan assumes the **docs** win (`slices/` subfolder). If the orchestrator wants to also fix the slice placement in M1..MN, expect each step to roughly **double in size** and to break the 800-line budget. Recommended resolution: park the slice-folder migration as a **separate** future change (`change-25-slice-folder-alignment` or similar) so change-24 stays review-sized.

---

## 5. Review-size and risk summary

| Item | Single-PR-default (this change) | If slice-folder migration is folded in |
|---|---|---|
| Total files touched across M1..MN | ~80–100 (additive + targeted edits) | ~150–200 (filesystem moves + import rewrites) |
| Approx net `+/-` lines | +1 700 src / +1 500 test (≈3 200 total) | +3 500 src / +1 500 test (≈5 000 total) |
| Review budget per PR (800 lines) | comfortable for each M-step | 4 of 7 steps will exceed the budget; needs chained PRs |
| Behavior change risk | **Low** — sentinel swap + handler simplify; HTTP codes are preserved by `apperror.NotFound(...)` etc. | **Medium** — filesystem moves can break import paths; CI must run between moves |
| Test impact | **Net positive** — every module gains `core/error_test.go` + `core/dto_test.go` + `core/model_test.go` where missing | Same plus import-path fixes in ~124 test files |
| Public API change | **None** — `Code` strings are new but not exposed in URL bodies; `PublicMessage` is preserved (uses module-owned `PublicMessage` from the new sentinel) | None |

---

## 6. Verification plan (binding for each M-step)

- `go vet ./...` clean.
- `go build ./...` clean.
- `go test ./...` clean.
- `grep -RE "apperror\." internal/modules/ | grep -v _test.go` returns hits **only** in `core/error.go` and **not** in any `service.go` or `repository.go`. (handlers may keep apperror until M6 deletes the last `mapServiceError` switch).
- `grep -RE 'gorm\.' internal/modules/ | grep _test.go` is not a check, but `grep -RE 'gorm\.' internal/modules/ --include="*service.go"` must be **empty** after M2..M5.
- `grep -RE "apperror\.Err" internal/modules/ | grep -v _test.go` is **empty** after M6.
- One table-driven test per `core/errors.go` covering each sentinel → `Status`, `Code`, `PublicMessage`.
- One `TestTableName` per partial GORM model in `core/model_test.go` (already present in iam; missing in auth, companies, onboarding).

---

## 7. Open questions for the orchestrator

1. **Spec vs docs slice placement** — does change-24 include moving slices under `<module>/slices/`, or is that deferred? The 800-line budget is comfortable either way; the risk profile is materially different.
2. **`platform/shared` package layout** — is the preferred shape `internal/platform/shared/string`, `internal/platform/shared/slug`, `internal/platform/shared/gormutil`, or a single flat `internal/platform/shared` package with `slug.go`, `string.go`, `gormutil.go`? Both are valid; the flat one is faster to merge.
3. **Reusable query naming** — `get_role_by_public_id.go` and `get_role_by_public_id_preloads.go` are byte-identical. Should M3d delete the `_preloads` variant outright (presuming callers always want preloads), or split into two distinct queries and pick the right one at the call site? Recommend outright delete with a regression test that pins the preloaded behavior.
4. **Module-owned `Code` string format** — does the project want `code_user_not_found`, `code:user_not_found`, or `user.not_found`? The current `apperror.CodeNotFound` is `"not_found"` (lowercase snake). Recommend `code:user_not_found` style for modules (collision-free, prefix-protected). The orchestrator should pick.

---

## 8. Ready for proposal

**Yes**, with the four open questions above answered. The exploration gives the orchestrator a module-by-module violation table, a concrete iteration plan, a verification contract per step, and an honest review-size estimate that respects the 800-line budget **only if the slice-folder migration is deferred to a follow-up change**.
