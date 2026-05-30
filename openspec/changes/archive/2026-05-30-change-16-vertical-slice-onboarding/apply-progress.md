# Apply Progress: change-16-vertical-slice-onboarding

## Status
Implemented cumulative onboarding migration for stacked-to-main delivery: PR1 foundation (Phase 1 + Phase 2) and remaining tasks (Phase 3 + Phase 4 + Phase 5). No commit created.

## Completed Tasks
- [x] 1.1 Create `internal/modules/onboarding/core/model.go` with partial models and local constants.
- [x] 1.2 Create `internal/modules/onboarding/core/dto.go` with onboarding request/response and validation.
- [x] 1.3 Create `internal/modules/onboarding/core/error.go` with onboarding domain errors.
- [x] 2.1 Create `queries/check_slug_available.go` + `_test.go`.
- [x] 2.2 Create `queries/check_domain_available.go` + `_test.go`.
- [x] 2.3 Create `queries/check_email_available.go` + `_test.go`.
- [x] 2.4 Create `queries/list_system_permissions.go` + `_test.go`.
- [x] 2.5 Create `queries/assign_permission_to_role.go` + `_test.go`.
- [x] 3.1 Create `internal/modules/onboarding/onboard_company/repository.go`.
- [x] 3.2 Create `internal/modules/onboarding/onboard_company/repository_test.go`.
- [x] 3.3 Create `internal/modules/onboarding/onboard_company/service.go`.
- [x] 3.4 Create `internal/modules/onboarding/onboard_company/service_test.go` (10 migrated scenarios).
- [x] 3.5 Create `internal/modules/onboarding/onboard_company/handler.go`.
- [x] 3.6 Create `internal/modules/onboarding/onboard_company/handler_test.go` (3 migrated groups).
- [x] 4.1 Create `internal/modules/onboarding/container.go`.
- [x] 4.2 Modify `internal/modules/onboarding/routes.go` to register via module container.
- [x] 4.3 Modify `internal/app/container.go` to wire onboarding module container only.
- [x] 4.4 Delete legacy onboarding root files (`handler.go`, `service.go`, `dto.go`, `handler_test.go`, `service_test.go`).
- [x] 4.5 Run module build/tests and verify no forbidden cross-module model imports.
- [x] 5.1 Run `go test ./...`.
- [x] 5.2 Run `go build ./...`.
- [x] 5.3 Verify `POST /api/v1/onboarding/companies` route remains protected by `requireRole("root")`.

## Files Changed
- `internal/modules/onboarding/core/model.go`
- `internal/modules/onboarding/core/dto.go`
- `internal/modules/onboarding/core/error.go`
- `internal/modules/onboarding/queries/check_slug_available.go`
- `internal/modules/onboarding/queries/check_slug_available_test.go`
- `internal/modules/onboarding/queries/check_domain_available.go`
- `internal/modules/onboarding/queries/check_domain_available_test.go`
- `internal/modules/onboarding/queries/check_email_available.go`
- `internal/modules/onboarding/queries/check_email_available_test.go`
- `internal/modules/onboarding/queries/list_system_permissions.go`
- `internal/modules/onboarding/queries/list_system_permissions_test.go`
- `internal/modules/onboarding/queries/assign_permission_to_role.go`
- `internal/modules/onboarding/queries/assign_permission_to_role_test.go`
- `internal/modules/onboarding/queries/test_helpers_test.go`
- `internal/modules/onboarding/onboard_company/repository.go`
- `internal/modules/onboarding/onboard_company/repository_test.go`
- `internal/modules/onboarding/onboard_company/service.go`
- `internal/modules/onboarding/onboard_company/service_test.go`
- `internal/modules/onboarding/onboard_company/handler.go`
- `internal/modules/onboarding/onboard_company/handler_test.go`
- `internal/modules/onboarding/container.go`
- `internal/modules/onboarding/routes.go`
- `internal/app/container.go`
- `internal/modules/onboarding/handler.go` (deleted)
- `internal/modules/onboarding/service.go` (deleted)
- `internal/modules/onboarding/dto.go` (deleted)
- `internal/modules/onboarding/handler_test.go` (deleted)
- `internal/modules/onboarding/service_test.go` (deleted)
- `openspec/changes/change-16-vertical-slice-onboarding/tasks.md`

## Test Commands Run
| Command | Result |
| --- | --- |
| `go test ./internal/modules/onboarding/queries` | PASS |
| `go test ./internal/modules/onboarding/...` | PASS |
| `go build ./...` | PASS |
| `go test ./...` | PASS |

## Deviations from Design
None — implementation matches design for all scoped phases.

## Remaining Tasks
None.

## Maintainer Review Adjustments
- Replaced onboarding dependency on `users.PasswordHasher` with consumer-owned `core.PasswordHasher` contract in `internal/modules/onboarding/core/contracts.go`.
- Added module-level `onboarding.Config` and rewired app container to call `onboarding.NewContainer(db, onboarding.Config{PasswordHasher: passwordManager, PlatformDomain: cfg.App.PlatformDomain})`.
- Removed root container import of onboarding slice package (`onboard_company`).
- Simplified onboarding slice service constructor to receive explicit `platformDomain` argument instead of option leakage across module boundaries.
- Renamed `internal/modules/onboarding/core/enums.go` to `internal/modules/onboarding/core/constants.go` preserving vocabulary constants.

## Workload / PR Boundary
- Mode: stacked PR slice
- Current work unit: PR 2 (slice + wiring + cleanup + verification)
- Boundary: starts from existing PR1 foundation state and completes remaining Phase 3/4/5 scope in working tree.
- Estimated review budget impact: expected within configured 800-line review budget target for current user-requested full remaining scope.
