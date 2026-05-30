# Tasks: Vertical Slice Migration — Onboarding Module

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1200-1500 (additions + deletions) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: Foundation (core + queries) → PR 2: Slice + wiring + cleanup |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Scaffold core/, queries/ with tests; no behavior change | PR 1 | Base = main; self-contained, verifiable via `go build ./...` |
| 2 | Implement onboard_company slice, wire container/routes, delete legacy | PR 2 | Base = main (stacked); depends on PR 1 types/queries being merged |

## Phase 1: Foundation — Core Types, Constants, DTOs, Errors

- [x] 1.1 Create `internal/modules/onboarding/core/model.go` with partial models (`OnboardingCompany`, `OnboardingCompanyDomain`, `OnboardingUser`, `OnboardingRole`, `OnboardingPermission`) each with `TableName()` overrides and local constants.
- [x] 1.2 Create `internal/modules/onboarding/core/dto.go` — move `OnboardCompanyRequest`, `OnboardCompanyResponse`, and `Validate()` from `dto.go`.
- [x] 1.3 Create `internal/modules/onboarding/core/error.go` — move all `Err*` variables from `service.go`.

## Phase 2: Queries — Reusable Data-Access Functions + Tests

- [x] 2.1 Create `internal/modules/onboarding/queries/check_slug_available.go` + `_test.go` — accepts `*gorm.DB`, returns error on duplicate.
- [x] 2.2 Create `internal/modules/onboarding/queries/check_domain_available.go` + `_test.go` — accepts `*gorm.DB`, domain string, duplicateErr; table-driven tests against in-memory SQLite.
- [x] 2.3 Create `internal/modules/onboarding/queries/check_email_available.go` + `_test.go` — accepts `*gorm.DB`, email string; table-driven tests.
- [x] 2.4 Create `internal/modules/onboarding/queries/list_system_permissions.go` + `_test.go` — returns `[]core.OnboardingPermission`; seed + assert test.
- [x] 2.5 Create `internal/modules/onboarding/queries/assign_permission_to_role.go` + `_test.go` — creates `role_permissions` join row via `tx.Table()`; verify row exists.

## Phase 3: Slice Implementation — onboard_company/

- [x] 3.1 Create `internal/modules/onboarding/onboard_company/repository.go` — struct with `*gorm.DB`, methods delegating to `queries/` for uniqueness checks and direct `tx.Create()` for company, domains, roles, user.
- [x] 3.2 Create `internal/modules/onboarding/onboard_company/repository_test.go` — SQLite in-memory tests verifying delegation/wiring; document heavy logic covered in `queries/`.
- [x] 3.3 Create `internal/modules/onboarding/onboard_company/service.go` — `Service` interface, `Onboard()` method with same transaction flow, `normalizeDomain()`, `WithPlatformDomain` option, injected `PasswordHasher` contract.
- [x] 3.4 Create `internal/modules/onboarding/onboard_company/service_test.go` — migrate all 10 scenarios from `service_test.go` using `core` partial models for AutoMigrate and assertions; table-driven where applicable.
- [x] 3.5 Create `internal/modules/onboarding/onboard_company/handler.go` — `Handler` struct, `Handle()` method, `respondError()` mapping errors to validation fields; uses `platform/response`.
- [x] 3.6 Create `internal/modules/onboarding/onboard_company/handler_test.go` — migrate 3 test groups from `handler_test.go` with fake service, table-driven conflict cases, `httptest` router setup.

## Phase 4: Wiring — Container, Routes, App Integration

- [x] 4.1 Create `internal/modules/onboarding/container.go` — `Container` struct with `OnboardCompany *onboard_company.Handler`, `NewContainer(db, hasher, opts...)` wiring repository → service → handler.
- [x] 4.2 Modify `internal/modules/onboarding/routes.go` — change `Register(v1, *Handler, ...)` to `Register(v1, *Container, ...)` routing to `c.OnboardCompany.Handle`.
- [x] 4.3 Modify `internal/app/container.go` — replace `onboardingHandler` field with `Onboarding *onboarding.Container`; call `onboarding.NewContainer(...)` instead of `NewService`+`NewHandler`; update `Register` call.
- [x] 4.4 Delete legacy files: `internal/modules/onboarding/handler.go`, `service.go`, `dto.go`, `handler_test.go`, `service_test.go`.
- [x] 4.5 Run `go build ./...` and `go test ./internal/modules/onboarding/...` — verify zero cross-module model imports (except `shared.BaseModel`, `users.PasswordHasher`).

## Phase 5: Verification — Full Suite

- [x] 5.1 Run `go test ./...` — all tests pass including integration tests.
- [x] 5.2 Run `go build ./...` — clean build with no unused imports.
- [x] 5.3 Verify `POST /api/v1/onboarding/companies` route still registered with `requireRole("root")` middleware.
