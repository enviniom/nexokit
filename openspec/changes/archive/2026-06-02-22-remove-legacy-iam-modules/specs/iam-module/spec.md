# Delta for IAM Module

## ADDED Requirements

### Requirement: No residual legacy references

The IAM module and all production code SHALL contain zero import paths referencing `internal/modules/users/`, `internal/modules/roles/`, or `internal/modules/permissions/`. All user, role, and permission types SHALL be sourced exclusively from IAM's local models in `iam/core/model.go`.

#### Scenario: Zero legacy imports in production code

- GIVEN the full production codebase
- WHEN `go list ./...` is run
- THEN no package import path contains `internal/modules/users`, `internal/modules/roles`, or `internal/modules/permissions`

#### Scenario: Zero legacy imports in test infrastructure

- GIVEN all test files under `tests/`
- WHEN imports are reviewed
- THEN no test file imports `internal/modules/users`, `internal/modules/roles`, or `internal/modules/permissions`

## REMOVED Requirements

### Requirement: Legacy module preservation

(Reason: Legacy modules `internal/modules/users/`, `internal/modules/roles/`, and `internal/modules/permissions/` are deleted. Preservation is no longer applicable — IAM is the sole boundary for users, roles, and permissions.)
