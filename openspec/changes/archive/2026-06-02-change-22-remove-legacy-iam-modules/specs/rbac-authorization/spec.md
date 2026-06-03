# Delta for RBAC Authorization

## MODIFIED Requirements

### Requirement: Module-owned name constants

Each module MUST define its own `Module<Name>` constant in `modules/<name>/core/constants.go`. Route definitions MUST reference the module-local constant, NOT `platform/permissions.Module*`. The `platform/permissions` package MUST NOT export `Module*` constants — it retains only `Action*` constants and utility functions. After legacy removal, the IAM module is the sole owner of user, role, and permission domain constants.
(Previously: Scenarios referenced `modules/users/core/constants.go` and `modules/roles/core/constants.go`; those directories no longer exist — IAM owns these constants)

#### Scenario: IAM module defines domain constants

- GIVEN `modules/iam/core/constants.go` exists
- WHEN routes reference the users, roles, or permissions module name
- THEN they use constants defined within the IAM module

#### Scenario: Platform permissions has no Module* constants

- GIVEN `platform/permissions/constants.go`
- WHEN the file is reviewed
- THEN it contains zero `Module*` constants — only `Action*` constants and utility functions

#### Scenario: No legacy module constant references

- GIVEN the full production codebase after legacy removal
- WHEN imports and constant references are reviewed
- THEN zero references to `modules/users/core`, `modules/roles/core`, or `modules/permissions/core` exist
