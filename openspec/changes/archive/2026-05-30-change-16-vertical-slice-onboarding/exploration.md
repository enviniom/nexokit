# Exploration: Vertical Slice Migration — Onboarding Module

## Current State

The `internal/modules/onboarding/` module uses a **flat legacy structure** with 6 files:

| File | Lines | Role |
|------|-------|------|
| `handler.go` | 64 | Single HTTP handler (`OnboardCompany`) |
| `service.go` | 251 | Entire onboarding flow in one `Onboard()` method (transaction with 5+ entity creations) |
| `dto.go` | 36 | Request/Response DTOs + validation |
| `routes.go` | 11 | Single route registration |
| `handler_test.go` | 169 | HTTP-level tests with fake service |
| `service_test.go` | 389 | Integration tests with SQLite + real GORM |

**Single endpoint:** `POST /api/v1/onboarding/companies` (root-only)

### Cross-Module Dependencies (Rule Violations)

The service directly imports models and constants from **4 other modules**, violating rule #7 ("No importar repositories ni modelos GORM de otro módulo"):

| Import | What's Used | Violation |
|--------|-------------|-----------|
| `companies.Company` | Full GORM model (create, count) | Model import |
| `companies.CompanyDomain` | Full GORM model (create, count) | Model import |
| `companies.CompanyStatusActive`, `CompanyDomainKindPrimary`, `CompanyDomainKindTechnical`, `CompanyDomainStatusActive` | Status/kind constants | Constant import |
| `users.User` | Full GORM model (create, count) | Model import |
| `users.PasswordHasher` | Interface (injected) | Acceptable as contract |
| `roles.Role` | Full GORM model (create) | Model import |
| `roles.AdminRoleSlug`, `roles.UserRoleSlug` | Role slug constants | Constant import |
| `permissions.Permission` | Full GORM model (find all) | Model import |
| `shared.BaseModel` | Base model embedding | Acceptable (platform-level) |

### Transaction Flow

The `Onboard()` method runs a **single GORM transaction** that:
1. Checks slug uniqueness → creates `Company`
2. Checks/creates primary domain → creates `CompanyDomain`
3. Generates/creates technical domain → creates `CompanyDomain`
4. Checks admin email uniqueness → creates `User`
5. Creates admin + user `Role` records
6. Fetches all system `Permission` records → assigns to roles via `role_permissions` join table

## Endpoint → Proposed Slice Mapping

Only **one endpoint** exists, so only **one slice** is needed:

| Endpoint | Proposed Slice | Files |
|----------|---------------|-------|
| `POST /api/v1/onboarding/companies` | `onboard_company/` | handler, service, repository + tests |

## Proposed Target Structure

```
internal/modules/onboarding/
  container.go              ← NEW: composition root (wiring + route registration)
  routes.go                 ← MODIFIED: delegates to container handlers
  core/
    model.go                ← NEW: local partial models for Company, CompanyDomain, Role, Permission, User
    dto.go                  ← MOVED: request/response DTOs + validation
    error.go                ← NEW: module errors (moved from service.go)
  queries/
    check_slug_available.go         ← NEW: reusable slug uniqueness check
    check_domain_available.go       ← NEW: reusable domain uniqueness check
    check_email_available.go        ← NEW: reusable email uniqueness check
    list_system_permissions.go      ← NEW: fetch all system permissions
    assign_permission_to_role.go    ← NEW: create role_permission join row
  onboard_company/
    handler.go              ← NEW: HTTP handler (extracted from current handler.go)
    handler_test.go         ← NEW: HTTP tests
    service.go              ← NEW: orchestration logic (extracted from current service.go)
    service_test.go         ← NEW: service-level tests
    repository.go           ← NEW: data access (wraps queries/)
    repository_test.go      ← NEW: repository tests
```

### Local Partial Models (`core/model.go`)

Each model declares **only the fields the onboarding flow reads/writes**:

```go
// Company — only fields written during onboarding
type OnboardingCompany struct {
    shared.BaseModel
    Name   string `gorm:"not null"`
    Slug   string `gorm:"type:varchar(120);uniqueIndex;not null"`
    Status string `gorm:"type:varchar(20);not null;default:'active'"`
}
func (OnboardingCompany) TableName() string { return "companies" }

// CompanyDomain — only fields written during onboarding
type OnboardingCompanyDomain struct {
    shared.BaseModel
    CompanyID         uint   `gorm:"not null;index"`
    Domain            string `gorm:"type:varchar(255);uniqueIndex;not null"`
    Status            string `gorm:"type:varchar(40);not null;default:'active'"`
    Kind              string `gorm:"type:varchar(40);not null;index"`
    RedirectToPrimary bool   `gorm:"not null;default:false"`
}
func (OnboardingCompanyDomain) TableName() string { return "company_domains" }

// OnboardingRole — only fields written during onboarding
type OnboardingRole struct {
    shared.BaseModel
    Name        string `gorm:"not null"`
    Slug        string `gorm:"not null"`
    CompanyID   *uint
    Description string
    IsSystem    bool `gorm:"not null;default:false"`
}
func (OnboardingRole) TableName() string { return "roles" }

// OnboardingPermission — only fields read during onboarding
type OnboardingPermission struct {
    shared.BaseModel
    Slug string `gorm:"type:varchar(120);uniqueIndex;not null"`
}
func (OnboardingPermission) TableName() string { return "permissions" }

// OnboardingUser — only fields written during onboarding
type OnboardingUser struct {
    shared.BaseModel
    Name         string `gorm:"not null"`
    Email        string `gorm:"uniqueIndex;not null"`
    PasswordHash string `gorm:"not null"`
    RoleID       uint   `gorm:"not null"`
    CompanyID    *uint
    IsActive     bool `gorm:"not null;default:true"`
}
func (OnboardingUser) TableName() string { return "users" }
```

### Constants

Role slugs and status/kind constants are **duplicated locally** in `core/model.go` or `core/error.go` rather than imported:

```go
const (
    RoleSlugAdmin = "admin"
    RoleSlugUser  = "user"
)

const (
    DomainStatusActive = "active"
    DomainKindPrimary  = "primary"
    DomainKindTechnical = "technical"
    CompanyStatusActive = "active"
)
```

## Dependencies Toward Other Modules — Elimination Plan

| Current Dependency | Elimination Strategy |
|-------------------|---------------------|
| `companies.Company` | → `OnboardingCompany` local partial model |
| `companies.CompanyDomain` | → `OnboardingCompanyDomain` local partial model |
| `companies.CompanyStatusActive` etc. | → Local constants in `core/` |
| `users.User` | → `OnboardingUser` local partial model |
| `users.PasswordHasher` | **Keep as injected interface** — this is a contract, not a model. The interface is defined in `users/service.go` and onboarding receives it via DI. This is acceptable per `_context.md`: "contrato pequeño inyectado desde app/container.go". |
| `roles.Role` | → `OnboardingRole` local partial model |
| `roles.AdminRoleSlug`, `roles.UserRoleSlug` | → Local constants in `core/` |
| `permissions.Permission` | → `OnboardingPermission` local partial model (only needs `ID` and `Slug`) |
| `shared.BaseModel` | **Keep** — this is the platform base model, not a module model |

## Duplicate/Reusable Queries — `queries/` Candidates

The current service has these inline data-access patterns that are candidates for `queries/`:

| Inline Code | Proposed Query File | Reusable By |
|-------------|-------------------|-------------|
| `tx.Model(&Company{}).Where("slug = ?", slug).Count(&count)` | `queries/check_slug_available.go` | Future company-related slices |
| `tx.Model(&CompanyDomain{}).Where("domain = ?", domain).Count(&count)` | `queries/check_domain_available.go` | companies module already has similar query |
| `tx.Model(&User{}).Where("email = ?", email).Count(&count)` | `queries/check_email_available.go` | users module has similar check |
| `tx.Find(&systemPermissions)` | `queries/list_system_permissions.go` | Any flow needing all system permissions |
| `tx.Table("role_permissions").Create(...)` | `queries/assign_permission_to_role.go` | Role-permission linking logic |

**Note:** The `companies` module already has `queries/get_company_domain_by_domain.go`. The onboarding `check_domain_available` is conceptually similar but operates within a transaction context. During migration, we should check if the existing companies query can be reused or if onboarding needs its own variant.

## Risks

### High
1. **Transaction boundary complexity** — The current `Onboard()` method runs everything in a single GORM transaction. In vertical slice, the repository receives the transaction handle and delegates to `queries/`. This pattern must be consistent to avoid partial commits.
2. **Test migration scope** — `service_test.go` (389 lines) uses `AutoMigrate` with external module models. Tests must be rewritten to use local partial models. This is the largest single-file migration risk.

### Medium
3. **PasswordHasher interface location** — Currently defined in `users/service.go`. If we duplicate it in onboarding's `core/`, we risk drift. Better to keep it as an injected contract from `users` (acceptable per rules).
4. **Role-permission join table** — Uses raw `tx.Table("role_permissions")` without a model. The `queries/assign_permission_to_role.go` must replicate this exactly.
5. **Container wiring changes** — `internal/app/container.go` currently creates `onboarding.Handler` directly. Must change to call `onboarding.NewContainer(db, passwordManager, ...)` and pass the container to `Register`.

### Low
6. **Constant duplication** — Status/kind constants duplicated between `companies/core/` and `onboarding/core/`. If companies changes a constant value, onboarding won't know. Mitigation: these are database-level constants unlikely to change.
7. **`routes.go` signature change** — Current `Register(v1, handler, requireRole)` takes a `*Handler`. Must change to `Register(v1, container, requireRole)` to match vertical slice pattern.

## Migration Plan

### Phase 1: Scaffold (no behavior change)
1. Create directory structure: `core/`, `queries/`, `onboard_company/`
2. Create `core/model.go` with local partial models
3. Create `core/error.go` with module errors
4. Move `dto.go` content to `core/dto.go`

### Phase 2: Queries
1. Create each query function in `queries/` with tests
2. Each query accepts `*gorm.DB` (works with both direct DB and transaction)

### Phase 3: Slice Implementation
1. Create `onboard_company/repository.go` — wraps queries, accepts `*gorm.DB`
2. Create `onboard_company/service.go` — orchestration logic (same transaction flow)
3. Create `onboard_company/handler.go` — HTTP handler (same behavior)
4. Create all `_test.go` files

### Phase 4: Wiring
1. Create `container.go` — composition root
2. Update `routes.go` — delegate to container
3. Update `internal/app/container.go` — call `onboarding.NewContainer()`
4. Delete old `handler.go`, `service.go`, `dto.go`
5. Update tests to use new structure

## Scope Confirmation

**Only `internal/modules/onboarding/` is migrated in this change.** No other modules are converted to vertical slice.

**Root app/container wiring changes ARE included** (Phase 4, step 3) because the onboarding module's public API changes from `*Handler` to `*Container`. This is a minimal change in `internal/app/container.go`:
- Replace `onboarding.NewService(...)` + `onboarding.NewHandler(...)` with `onboarding.NewContainer(...)`
- Replace `c.onboardingHandler` with `c.Onboarding` (container reference)
- Update `onboarding.Register(...)` call to pass container

**No other module code is modified.**

## Ready for Proposal

**Yes.** The exploration is complete with:
- Clear current state analysis
- Explicit cross-module dependency elimination plan
- Local partial model definitions
- Query extraction candidates
- Risk assessment with mitigation strategies
- Phased migration plan
- Confirmed scope (onboarding only + minimal container wiring)

The orchestrator should proceed to `sdd-propose` to formalize intent, scope, and approach.
