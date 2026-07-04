# Design: Vertical-Slice and Platform/Shared Review

## Technical Approach

Change-24 is a behavior-preserving refactor that moves the existing modules onto the documented error, persistence-boundary, shared-helper, and test contracts. Keep current module-root slice placement, routes, payloads, HTTP statuses, migrations, and business behavior unchanged.

## Architecture Decisions

| Area | Decision | Rationale |
|---|---|---|
| Error vocabulary | Each module keeps reusable sentinels in `internal/modules/<module>/core/error.go` with `Code* apperror.Code = "code:<snake_case>"` constants and `apperror.*` constructors. | Centralizes business errors and lets `response.HandleError` preserve status/message/code without handler switches; handlers must not construct inline `apperror.Wrap` values for reusable business cases. |
| GORM boundary | Repositories translate `gorm.ErrRecordNotFound` and unique violations; services return `core.Err*` or `fmt.Errorf("...: %w", err)`. | Services stay persistence-free; DB details stop leaking upward. |
| Shared helpers | Pure normalizers live in `internal/platform/shared/string`; GORM duplicate-key detection extends `internal/platform/gormutil`. | Pure and DB-specific helpers stay separated and reusable. |
| IAM duplicate query | Delete `internal/modules/iam/queries/get_role_by_public_id_preloads.go`; keep `GetRoleByPublicID` as the single preload-capable query. | Removes dead duplicate code while pinning preload behavior by regression test. |
| Slice folders | Defer moving slices under `slices/`. | The locked scope forbids this migration; moving files would inflate review size and mix architecture cleanup with filesystem churn. |

## Data Flow

```txt
DB/GORM error -> repository translation -> core AppError sentinel -> service -> handler -> response.HandleError -> API envelope
```

Missing-row behavior that is exceptional is translated in repositories. Expected existence checks should keep explicit `(*T, bool, error)`-style contracts when already appropriate; do not force `AppError` into control-flow branches.

## File / Package Architecture

| File or package | Action | Design |
|---|---|---|
| `internal/platform/shared/string/normalize_{slug,domain,email}.go` | Create | Export `NormalizeSlug`, `NormalizeDomain`, `NormalizeEmail`; table-driven tests in the same package. Do not move `iam/queries/normalize_slugs.go`: it is a plural list de-dup helper, not the singular shared normalizer. |
| `internal/platform/gormutil/unique.go` | Create/extend | Add `IsUniqueConstraintError(err error) bool`; detect `gorm.ErrDuplicatedKey`, Postgres duplicate/unique messages, and SQLite-style unique/constraint failure messages; return false for nil/unrelated errors. |
| `internal/modules/{auth,iam,companies,onboarding}/core/error.go` | Modify | Add unique `Code*` constants in `code:<snake_case>` format and construct sentinels with `apperror.NotFound`, `Conflict`, `Forbidden`, `Unauthorized`, `BadRequest`, or `Validation` as applicable. IAM must include `ErrRoleHasAssignedUsers = apperror.Unprocessable(CodeRoleHasAssignedUsers, core.MsgRoleHasAssignedUsers, nil)` instead of inline handler wrapping. Companies must plan sentinels for `code:company_not_found`, `code:company_domain_not_found`, `code:company_domain_duplicate`, `code:primary_domain_exists`, and `code:company_domain_does_not_belong`; implementation may only refine names to avoid duplicate codes. |
| Module handlers | Modify | Remove `mapServiceError`; call `response.HandleError(c, err)` directly for service errors. |
| Module services | Modify | Remove `gorm.io/gorm` and `platform/apperror` imports; use shared normalizers where relevant. Onboarding transaction ownership may stay in service only if GORM is hidden behind a repository transaction method; otherwise sub-slice it carefully. |
| Module repositories | Modify | Translate `gorm.ErrRecordNotFound` to matching `core.Err*`; use `gormutil.IsUniqueConstraintError` for duplicate-key conflicts. |
| `internal/modules/iam/queries/get_role_by_public_id_preloads.go` | Delete | Surviving query keeps `Company` and `Permissions` preloads. |
| `docs/module-error-conventions.md` and `docs/modules/validation-and-errors.md` | Create/modify | Publish and cross-link the module error table. |

## Interfaces / Contracts

```go
func NormalizeSlug(s string) string   // strings.ToLower(strings.TrimSpace(s))
func NormalizeDomain(s string) string // lower, trim space, trim trailing "."
func NormalizeEmail(s string) string  // lower, trim space
func IsUniqueConstraintError(err error) bool
```

Handler contract: bind -> `RespondIfInvalid` -> service -> `response.HandleError(c, err)`. No handler/service/repository constructs ad-hoc `apperror` values outside `core/error.go`.

## Testing Strategy

| Target | Tests |
|---|---|
| Platform helpers | Table-driven unit tests for normalization and duplicate-key detection: `gorm.ErrDuplicatedKey`, Postgres duplicate-key and unique-constraint messages, SQLite unique-constraint failure, generic SQL error, connection error, and nil. |
| Module errors | `core/errors_test.go` per module pins status, code format, uniqueness, and public message. |
| DTO/model contracts | `core/dto_test.go` for validation; direct `TableName()` tests for partial GORM models. |
| Repositories | Not-found and unique-violation translation tests per touched slice. |
| Handlers/services | Existing behavior tests updated to expect unchanged HTTP status/envelope and module sentinels. Auth service tests must pivot from `errors.Is(err, apperror.ErrUnauthorized)` to `core.ErrInvalidCredentials`, `core.ErrInvalidRefreshToken`, etc.; HTTP 401 behavior remains unchanged. |
| IAM query | Regression test in `internal/modules/iam/queries/get_role_by_public_id_test.go` for `Company`, `Permissions`, and not-found behavior. |

## Implementation Sequence

Single-PR-default is likely unrealistic for the full change within the 800-line review budget. Implement as reviewable work units: M0 shared helpers; M1 helper adoption in companies/onboarding; M2 onboarding error/boundary migration; M3a IAM core; M3b IAM users; M3c IAM roles; M3d IAM permissions plus duplicate query deletion; M3e IAM internal slices plus audit of `iam/internal/*/service.go` for GORM/apperror leaks; M4 companies; M5 auth; M6 CI grep guard over handler/service/repository files excluding tests; M7 docs. IAM must remain sub-sliced.

## Rollback and Verification

Rollback is `git revert` per work unit; no data migration or feature flag is required. Verify each slice with `go test ./...`, `go build ./...`, `go vet ./...`, grep guards for `apperror.` in services/repositories/handlers, no `gorm.` in services, no `mapServiceError`, deleted duplicate IAM query, and unchanged public API behavior.

## Open Questions

None.
