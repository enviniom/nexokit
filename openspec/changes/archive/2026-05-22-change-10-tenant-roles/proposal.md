# Proposal: Tenant-scoped roles with global root

## Intent

Scope roles to companies for SaaS multitenancy while keeping `root` as a global system role. Current RBAC is fully global — each company needs its own `admin`, `user`, and custom roles.

## Scope

### In Scope
- `company_id` nullable on `roles` with proper unique indexes
- Tenant-scoped queries for non-root users
- Seed only `root` globally; remove `admin`/`user` seeds
- Remove root `role_permissions` seed (middleware handles bypass)
- Reserved slugs (`root`, `admin`, `user`) protected from API creation
- Root protected from edit/delete via API
- DTO includes `company_id`; isolation tests

### Out of Scope
- Company onboarding (creates `admin`/`user` per tenant)
- Production data migration tooling
- Authorization middleware changes

## Capabilities

### New Capabilities
- `tenant-scoped-roles`: Company-scoped roles, global root, reserved slug protection, tenant-isolated queries

### Modified Capabilities
- `roles`: Seed only root, composite unique constraints, `company_id` in response
- `rbac-authorization`: No changes

## Approach

Single endpoint with tenant context filtering. Handler extracts tenant from Gin context; service/repository accept optional `companyID`. Root sees all; non-root scoped to company.

1. **Migration**: `company_id` column, partial unique index (`WHERE company_id IS NULL`), composite indexes `(slug, company_id)`, `(name, company_id)`
2. **Model**: `CompanyID *uint`
3. **Repository**: Optional `companyID` on `List`, `Count`, `GetByPublicID`
4. **Service**: `isReservedSlug()` replaces `isRootRole()`; tenant-aware queries
5. **Handler**: Extract `companyID` from `authctx.User`
6. **Seeds**: Only root; no root permission assignments
7. **DTO**: `CompanyID *string` (omitempty)

## Affected Areas

- `migrations/` — New migration for `company_id` and indexes
- `internal/modules/roles/model.go` — Add `CompanyID`
- `internal/modules/roles/repository.go` — Tenant-scoped queries
- `internal/modules/roles/service.go` — `isReservedSlug()`, tenant filtering
- `internal/modules/roles/handler.go` — Extract tenant context
- `internal/modules/roles/dto.go` — Add `CompanyID` to response
- `seeds/roles.go` — Seed only root
- `seeds/role_permissions.go` — Remove root assignments
- `openspec/specs/roles/spec.md` — Delta spec

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Production roles need migration | Medium | Assign orphans to default company |
| Companies without default roles | Medium | Onboarding change must follow |

## Rollback Plan

`goose down` removes `company_id`, restores original indexes. Re-run original seeds. Revert code via git.

## Dependencies

- change-04-rbac (merged)
- Company onboarding change (future)

## Success Criteria

- [ ] `roles.company_id` exists; null only for root
- [ ] Seed creates only root, no root role_permissions
- [ ] Root bypass works without role_permissions
- [ ] API rejects creation/edit/delete of root
- [ ] API rejects reserved slugs (root, admin, user)
- [ ] Non-root lists only own company roles (404 for others)
- [ ] Tests cover root global, tenant roles, isolation
