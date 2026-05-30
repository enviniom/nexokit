# Proposal: Vertical Slice Migration — Onboarding Module

## Intent

Refactor `internal/modules/onboarding/` from flat legacy to vertical slice, eliminating cross-module model imports (rule #7 violation). No functional changes.

## Scope

### In Scope
- One slice: `onboard_company/` (handler, service, repository + tests)
- `core/` with local partial models for Company, CompanyDomain, User, Role, Permission
- `queries/` with reusable data-access functions + tests
- `container.go` as composition root; `routes.go` delegates to container
- `internal/app/container.go` calls `onboarding.NewContainer()`
- Delete old `handler.go`, `service.go`, `dto.go`

### Out of Scope
- No new endpoints or slices
- No other module migrations
- Root container does NOT wire individual slices

## Capabilities

### New Capabilities
None

### Modified Capabilities
None

> Pure architectural refactor. `company-onboarding` spec unchanged. `vertical-slice-modules` spec already covers the target pattern.

## Approach

1. Scaffold `core/`, `queries/`, `onboard_company/`
2. Define partial GORM models in `core/model.go` with `TableName()` overrides
3. Duplicate status/kind/role-slug constants locally
4. Extract 5 query functions into `queries/` with tests
5. Implement `onboard_company/` slice (same transaction flow)
6. Create `container.go`, update `routes.go` and `internal/app/container.go`
7. Delete legacy files, verify tests

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/modules/onboarding/` | Restructured | Flat → vertical slice |
| `internal/app/container.go` | Modified | `NewService`+`NewHandler` → `NewContainer` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Transaction boundary breakage | Medium | Queries accept `*gorm.DB` (works with tx or direct) |
| Test migration (service_test.go: 389 lines) | Medium | Rewrite with local partial models; preserve scenarios |
| PasswordHasher interface drift | Low | Keep as injected contract from `users` |

## Rollback Plan

`git revert` restores all deleted files. No DB migration needed. No feature flag.

## Dependencies

- `users.PasswordHasher` interface must remain stable

## Success Criteria

- [ ] `POST /api/v1/onboarding/companies` returns identical responses
- [ ] Zero cross-module model imports (except `shared.BaseModel`, `users.PasswordHasher`)
- [ ] All tests pass
- [ ] `go build ./...` and `go test ./...` pass
