# Design: Vertical Slice Migration — Onboarding Module

## Technical Approach

Refactor `internal/modules/onboarding/` from flat legacy files into one vertical slice, because the module has exactly one endpoint: `POST /api/v1/onboarding/companies` → `onboard_company/`. Behavior, route, response envelope, validation rules, transaction semantics, and root-only authorization stay unchanged. No delta specs exist for this change; the design follows `proposal.md`, `exploration.md`, and `docs/prompts/_context.md` vertical-slice rules.

## Architecture Decisions

| Option | Tradeoff | Decision |
|---|---|---|
| One slice per existing endpoint | Avoids fake structure; no new endpoint slices | Use only `onboard_company/` for `POST /onboarding/companies`. |
| Local partial models vs importing other module models | Duplicates minimal schema but preserves module autonomy | Add onboarding-owned partial GORM models in `core/model.go`; keep only `shared.BaseModel` and `users.PasswordHasher` external. |
| Query functions accept `*gorm.DB` | Slightly lower abstraction, but works with root DB and transactions | Repositories delegate to `queries/*` using the passed tx/db handle. |
| Root container wires module vs slices | Keeps app root from knowing slice internals | `internal/app/container.go` calls `onboarding.NewContainer(...)`; routes receive `*onboarding.Container`. |

## Data Flow

```txt
POST /api/v1/onboarding/companies
  → onboarding.Register(..., *Container, RequireRole)
  → onboard_company.Handler.Handle
  → onboard_company.Service.Onboard
  → db.Transaction(tx)
  → onboard_company.Repository(tx)
  → queries + direct creates using core partial models
```

Inside the transaction: normalize slug/domain/email → check slug/domain/email uniqueness → create company → optional primary/technical domains → create admin/user roles → list system permissions → assign role_permissions → hash password → create admin user → return existing response DTO.

## File Changes

| File | Action | Description |
|---|---|---|
| `internal/modules/onboarding/container.go` | Create | Module composition root; constructs repository/service/handler and exposes `OnboardCompany *onboard_company.Handler`. |
| `internal/modules/onboarding/routes.go` | Modify | `Register(v1, c *Container, requireRole)` and route `POST /companies` to `c.OnboardCompany.Handle`. |
| `internal/modules/onboarding/core/model.go` | Create | Partial models and local constants. |
| `internal/modules/onboarding/core/dto.go` | Create | Move request/response DTOs and validation. |
| `internal/modules/onboarding/core/error.go` | Create | Move duplicate/missing-domain errors. |
| `internal/modules/onboarding/queries/*.go` | Create | Reusable data-access helpers with tests. |
| `internal/modules/onboarding/onboard_company/{handler,service,repository}.go` | Create | Slice implementation. |
| `internal/modules/onboarding/{handler,service,dto}.go` | Delete | Replaced by core + slice files. |
| `internal/app/container.go` | Modify | Replace `onboardingHandler` with `Onboarding *onboarding.Container`; call only module container. |

Final layout:

```txt
internal/modules/onboarding/
  container.go routes.go
  core/{model.go,dto.go,error.go}
  queries/{check_slug_available.go,check_domain_available.go,check_email_available.go,list_system_permissions.go,assign_permission_to_role.go}
  onboard_company/{handler.go,handler_test.go,service.go,service_test.go,repository.go,repository_test.go}
```

## Interfaces / Contracts

`core/model.go` defines local partial models only for fields onboarding reads/writes: `OnboardingCompany`, `OnboardingCompanyDomain`, `OnboardingUser`, `OnboardingRole`, `OnboardingPermission`, each with `TableName()` for `companies`, `company_domains`, `users`, `roles`, `permissions`. Constants are duplicated locally: `CompanyStatusActive`, `DomainStatusActive`, `DomainKindPrimary`, `DomainKindTechnical`, `RoleSlugAdmin`, `RoleSlugUser`.

Reusable query files and delegates:

| Query | Used by |
|---|---|
| `CheckSlugAvailable(db, slug) error` | `onboard_company.Repository.EnsureCompanySlugAvailable` |
| `CheckDomainAvailable(db, domain, duplicateErr) error` | primary + technical domain checks |
| `CheckEmailAvailable(db, email) error` | admin email check |
| `ListSystemPermissions(db) ([]core.OnboardingPermission, error)` | role permission assignment |
| `AssignPermissionToRole(db, roleID, permissionID) error` | admin/user role permission joins |

## Testing Strategy

| Layer | What to Test | Approach |
|---|---|---|
| Handler | 201 success, root-only route, validation/conflict mapping | `httptest`, fake service, table-driven conflicts. |
| Service | Full transaction flow, normalization, rollback, password hashing, permission subsets | SQLite in-memory with `core` models; behavior assertions, not implementation trivia. |
| Repository | Delegation/wiring to queries and creates inside tx | Small SQLite tests; document query behavior covered in `queries/`. |
| Queries | Uniqueness checks, permission listing, role_permissions insert | Table-driven tests against in-memory DB. |

Run narrow packages first, then `go test ./...` and `go build ./...`.

## Migration / Rollout

No DB migration required. Implement scaffold → queries → slice → module container/routes → app container. Preserve the existing endpoint and do not touch other modules except `internal/app/container.go`. Rollback is `git revert`.

## Open Questions

None.
