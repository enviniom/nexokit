# Domain Admin Apply Result: change-14-company-domains

## Status
Implemented. No commit created.

## Summary
Extended the already-applied `change-14-company-domains` with root-only company domain administration endpoints under the companies module.

Implemented endpoints:

- `GET /api/v1/companies/:id/domains`
- `POST /api/v1/companies/:id/domains`
- `PUT /api/v1/companies/:id/domains/:domain_id`

No `DELETE` endpoint was added. Domain lifecycle remains status-based: `active`, `inactive`, `pending_verification`.

The existing companies `:id` route convention is used: `:id` is the company public ID, not an internal numeric ID.

## Changed Files

Code:

- `internal/modules/companies/dto.go`
- `internal/modules/companies/handler.go`
- `internal/modules/companies/handler_test.go`
- `internal/modules/companies/repository.go`
- `internal/modules/companies/routes.go`
- `internal/modules/companies/service.go`
- `internal/modules/companies/service_test.go`

OpenSpec/apply artifacts:

- `openspec/changes/change-14-company-domains/specs.md`
- `openspec/changes/change-14-company-domains/design.md`
- `openspec/changes/change-14-company-domains/tasks.md`
- `openspec/changes/change-14-company-domains/apply-progress.md`
- `openspec/changes/change-14-company-domains/domain-admin-apply-result.md`

## Behavior Implemented

- Root-only nested company domain admin routes.
- List multiple domains for a company.
- Create company domains with `domain`, `kind`, `status`, and `redirect_to_primary`.
- Update company domains with the same fields.
- No delete behavior.
- Domain normalization before persistence.
- Global domain uniqueness validation.
- Supported kind validation: `primary`, `alias`, `technical`.
- Supported status validation: `active`, `inactive`, `pending_verification`.
- At most one active primary domain per company.
- `redirect_to_primary=true` rejected for primary domains.
- Update rejects domains that do not belong to the specified company.
- Tenant resolver semantics from the prior change remain unchanged.

## Strict TDD Evidence

| Cycle | RED | GREEN | TRIANGULATE | REFACTOR |
| --- | --- | --- | --- | --- |
| Root domain admin routes | Added handler tests for nested list/create/update, root-only access, public ID route params, and absent DELETE route. `go test ./internal/modules/companies` failed because DTOs/service methods/routes did not exist. | Added DTOs, handlers, routes, and service interface methods. | Handler test verifies non-root receives 403 and DELETE route returns 404. | Kept domain admin as separate nested endpoints instead of overloading company update. |
| Domain admin service rules | Added service tests for list, create, duplicate domain, second active primary, and cross-company update rejection. | Added repository contract methods and service business rules. | Added ownership check and active primary count exclusion behavior for updates. | Extracted domain normalization, domain availability, active-primary validation, and response mapping helpers. |

## Tests Run

- `go test ./internal/modules/companies` — PASS
- `go test ./...` — PASS
- `go build ./...` — PASS

Combined verification command executed:

```bash
go test ./internal/modules/companies && go test ./... && go build ./...
```

## Risks / Review Notes

- This extension adds review volume to the already-large change 14 diff; parent/user may still split before commit if desired.
- Host resolution cache invalidation remains out of scope. Domain admin changes may take up to the existing public tenant cache TTL to affect cached positive host resolutions.
- `status` and `kind` remain Go-enforced string values; no SQL CHECK constraints were added.
- Domain-management list/create/update is root-only and implemented in the companies module; no tenant admin self-service domain management was added.

## Skill Resolution

none
