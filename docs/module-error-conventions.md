# Module error conventions

This document is the canonical reference for every reusable module sentinel in `internal/modules/<module>/core/error.go`. Use it to pick the right error, verify its HTTP status and public message, and keep the module vocabulary in sync with the code.

## Quick path

1. Add a reusable sentinel to `internal/modules/<module>/core/error.go` only.
2. Build it with a helper from `platform/apperror` so the HTTP status is explicit.
3. Record it in the table below in the same change.
4. Add or extend `core/errors_test.go` to pin `Status`, `Code`, and `PublicMessage`.

## Auth module

`internal/modules/auth/core/error.go`

| Sentinel | Code | HTTPStatus | PublicMessage | Usage / notes |
|----------|------|------------|---------------|---------------|
| `ErrInvalidCredentials` | `code:invalid_credentials` | 401 | `No autorizado` | Returned by `authenticate_user` service/repository for missing users, inactive users, or password mismatch. Intentionally preserves the legacy generic 401 message. |
| `ErrInvalidRefreshToken` | `code:invalid_refresh_token` | 401 | `No autorizado` | Returned by `revoke_token` and `rotate_token` service/repository for missing, expired, or revoked tokens. Also preserves the legacy generic 401 message. |

## IAM module

`internal/modules/iam/core/error.go`

| Sentinel | Code | HTTPStatus | PublicMessage | Usage / notes |
|----------|------|------------|---------------|---------------|
| `ErrNotFound` | `code:iam_resource_not_found` | 404 | `Recurso no encontrado` | Generic missing-resource sentinel for IAM lookups. Intentionally preserves the platform/public 404 message from the pre-Change-24 handler remapping. |
| `ErrConflict` | `code:iam_resource_conflict` | 409 | `El recurso ya existe` | Generic conflict sentinel when no more specific code exists. Intentionally preserves the platform/public 409 message from the pre-Change-24 handler remapping. |
| `ErrForbidden` | `code:iam_forbidden` | 403 | `Acceso denegado` | Generic forbidden sentinel for IAM actions. Intentionally preserves the platform/public 403 message from the pre-Change-24 handler remapping. |
| `ErrUnauthorized` | `code:iam_unauthorized` | 401 | `No autorizado` | Generic unauthorized sentinel for IAM flows. Intentionally preserves the platform/public 401 message from the pre-Change-24 handler remapping. |
| `ErrInvalidPermissionSlug` | `code:invalid_permission_slug` | 400 | `Solicitud inválida` | Returned when a permission slug fails format/business rules. Intentionally preserves the platform/public 400 message from the pre-Change-24 handler remapping. |
| `ErrRoleNameAlreadyExists` | `code:role_name_already_exists` | 422 | `role name already exists` | Returned when a role name collides with an existing role. |
| `ErrRoleSlugAlreadyExists` | `code:role_slug_already_exists` | 422 | `role slug already exists` | Returned when a role slug collides with an existing role. |
| `ErrReservedRoleIdentity` | `code:reserved_role_identity` | 422 | `reserved role identity` | Returned when a create/update targets a reserved role identity. |
| `ErrRoleHasAssignedUsers` | `code:role_has_assigned_users` | 422 | `El rol tiene usuarios asignados` | Returned by `delete_role` service when the role still has users. Uses the module-owned `MsgRoleHasAssignedUsers` constant and replaces the old inline `apperror.Wrap` in the handler. |
| `ErrRoleProtected` | `code:role_protected` | 403 | `role is protected` | Returned when an operation targets a protected role. |
| `ErrSystemImmutable` | `code:system_immutable` | 403 | `system resource is immutable` | Returned when an operation targets a system-owned immutable resource. |
| `ErrUserEmailAlreadyExists` | `code:user_email_already_exists` | 422 | `user email already exists` | Returned by `create_user`/`update_user` repositories when the email violates a unique constraint. |
| `ErrForbiddenRoleAssignment` | `code:forbidden_role_assignment` | 403 | `forbidden role assignment` | Returned when a role assignment is not allowed. |
| `ErrInvalidCompanyScope` | `code:invalid_company_scope` | 400 | `Solicitud inválida` | Returned when the company scope is invalid. Intentionally preserves the platform/public 400 message from the pre-Change-24 handler remapping. |
| `ErrForbiddenCompanyScope` | `code:forbidden_company_scope` | 403 | `forbidden company scope` | Returned when the company scope is forbidden. |

## Companies module

`internal/modules/companies/core/error.go`

| Sentinel | Code | HTTPStatus | PublicMessage | Usage / notes |
|----------|------|------------|---------------|---------------|
| `ErrCompanyNotFound` | `code:company_not_found` | 404 | `company not found` | Returned when a company lookup fails. |
| `ErrCompanyDomainNotFound` | `code:company_domain_not_found` | 404 | `company domain not found` | Returned when a company domain lookup fails. |
| `ErrDuplicateCompanyDomain` | `code:company_domain_duplicate` | 409 | `company domain already exists` | Returned by `create_company_domain`/`update_company_domain` repositories on unique violation. The handler maps this to a field-keyed 422 validation response to preserve the original public contract. |
| `ErrActivePrimaryDomainExists` | `code:primary_domain_exists` | 422 | `company already has an active primary domain` | Returned when a second active primary domain is requested. |
| `ErrCompanyDomainDoesNotBelong` | `code:company_domain_does_not_belong` | 404 | `company domain does not belong to company` | Returned when a domain does not belong to the requested company. |
| `ErrDuplicateCompanySlug` | `code:company_slug_duplicate` | 409 | `company slug already exists` | Returned by `create_company`/`update_company` repositories on unique violation. The handler maps this to a field-keyed 422 validation response to preserve the original public contract. |

## Onboarding module

`internal/modules/onboarding/core/error.go`

| Sentinel | Code | HTTPStatus | PublicMessage | Usage / notes |
|----------|------|------------|---------------|---------------|
| `ErrDuplicateCompanySlug` | `code:duplicate_company_slug` | 422 | `company slug already exists` | Returned when the requested company slug already exists. Mapped to a field-keyed 422 response in `onboard_company/handler.go`. |
| `ErrDuplicateCompanyDomain` | `code:duplicate_company_domain` | 422 | `company domain already exists` | Returned when the requested company domain already exists. Mapped to a field-keyed 422 response. |
| `ErrDuplicateTechnicalDomain` | `code:duplicate_technical_domain` | 422 | `company technical domain already exists` | Returned when the generated technical domain already exists. Mapped to a field-keyed 422 response. |
| `ErrMissingPlatformDomain` | `code:missing_platform_domain` | 422 | `platform domain is required to generate technical domain` | Returned when the platform domain config is missing. Mapped to a field-keyed 422 response. |
| `ErrDuplicateAdminEmail` | `code:duplicate_admin_email` | 422 | `admin email already exists` | Returned when the admin email already exists. Mapped to a field-keyed 422 response. |

## Conventions

| Convention | Rule |
|------------|------|
| Code format | `code:<snake_case>` (e.g. `code:user_not_found`). |
| HTTP status | Set by the `apperror` helper used to build the sentinel (`NotFound` → 404, `Conflict` → 409, `Validation`/`Unprocessable` → 422, `Unauthorized` → 401, `Forbidden` → 403, `BadRequest` → 400). |
| PublicMessage | Human-readable, lower-case, no trailing punctuation. Stable across versions because it is client-visible text. Some modules intentionally reuse an older public message to preserve an existing HTTP contract. |
| Reuse vs. ad-hoc | Reusable sentinels live in `core/error.go`. Slice-scoped ad-hoc errors stay in the slice and are not part of the module vocabulary. |
| Wrapping | Internal errors are wrapped with `fmt.Errorf("...: %w", err)`. The wrapped error inherits the sentinel's `Code` and `HTTPStatus` via `apperror.Is` and `apperror.Wrap`. |
| Test coverage | Every sentinel in this doc is covered by `core/errors_test.go`, which pins `apperror.Status`, `Code`, and `PublicMessage`. |

## Layer rules

| Layer | Rule |
|-------|------|
| `core/error.go` | Owns the module's reusable `apperror.Code` constants and `Err*` sentinels. |
| Service | Returns `core.Err*` or wraps internal errors with `fmt.Errorf`. Must not import `gorm.io/gorm` or `platform/apperror`. |
| Repository | Translates persistence errors (`gorm.ErrRecordNotFound`, unique violations) into `core.Err*` sentinels. Must not leak GORM or `apperror` to services. |
| Handler | Funnels business errors through `response.HandleError(c, err)`. Must not import `platform/apperror` for reusable cases. When the original public contract requires a field-keyed 422 envelope, a thin handler-side mapping is allowed. |

## Enforcement

The runtime contract is enforced by `make check-module-errors`:

- No `apperror.` imports in module `*service.go`, `*repository.go`, or `*handler.go`.
- No `gorm.` imports in module `*service.go`.
- No `mapServiceError` adapters in module non-test files.

The target is wired into `make lint` and runs in CI via the `module-errors-guard` job. Keep this doc updated whenever a sentinel is added, removed, or renamed.

## Review checklist

- [ ] New sentinel added to `internal/modules/<module>/core/error.go`.
- [ ] New sentinel added to this document.
- [ ] `core/errors_test.go` covers `apperror.Status`, `Code`, and `PublicMessage`.
- [ ] No `apperror.` or `gorm.` imports leak into module services/handlers.
- [ ] Handler mapping to field-keyed 422 responses is intentional and documented in the usage notes column.

## Related docs

- [`docs/modules/validation-and-errors.md`](modules/validation-and-errors.md) — DTO validation, response envelopes, and the `core/error.go` pattern.
- [`docs/error-handling.md`](error-handling.md) — Platform `AppError` shape, `Wrap` semantics, and release-mode redaction rules.
