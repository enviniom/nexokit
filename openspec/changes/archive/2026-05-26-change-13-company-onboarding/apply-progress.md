# Apply Progress: Company Onboarding

## Status

Applied before closure; reconciled during SDD closeout on 2026-05-26.

## Implementation Evidence

Primary implementation commit:

```text
807480f feat(onboarding): implement transactional company onboarding, auto-sync permissions, and admin lock
```

Implemented areas:

- `internal/modules/onboarding/` — new onboarding DTOs, handler, routes, transactional service, and tests.
- `internal/app/container.go` — onboarding service/handler wiring and route registration.
- `internal/modules/companies/routes.go` — direct `POST /api/v1/companies` route removed.
- `internal/modules/companies/handler_test.go` — direct company creation route returns 404.
- `internal/modules/permissions/repository.go` — `AutoAssignToAdmins` persistence hook.
- `internal/modules/permissions/service.go` — permission sync auto-assigns newly created permissions to admin roles.
- `internal/modules/permissions/service_test.go` — auto-sync verification.
- `internal/modules/roles/service.go` — admin role permission assignment lock.
- `internal/modules/roles/service_test.go` — admin role permission protection verification.

## Reconciliation Notes

- The original admin lock task described blocking revocation by checking omitted existing permissions. The landed implementation is stricter: direct permission assignment changes for `admin` roles are rejected entirely. Canonical specs were synchronized to describe the stricter current behavior.
- The change-specific tasks were unchecked even though code had landed; they were marked complete during closeout based on code/test evidence.
