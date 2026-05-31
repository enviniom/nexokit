# Delta for RBAC Authorization

## ADDED Requirements

### Requirement: Module-owned name constants

Each module MUST define its own `Module<Name>` constant in `modules/<name>/core/constants.go`. Route definitions MUST reference the module-local constant, NOT `platform/permissions.Module*`. The `platform/permissions` package MUST NOT export `Module*` constants — it retains only `Action*` constants and utility functions.

#### Scenario: Users module defines ModuleUsers

- GIVEN `modules/users/core/constants.go` exists
- WHEN routes reference the users module name
- THEN they use the local `ModuleUsers` constant

#### Scenario: Roles module defines ModuleRoles

- GIVEN `modules/roles/core/constants.go` exists
- WHEN routes reference the roles module name
- THEN they use the local `ModuleRoles` constant

#### Scenario: Platform permissions has no Module* constants

- GIVEN `platform/permissions/constants.go`
- WHEN the file is reviewed
- THEN it contains zero `Module*` constants — only `Action*` constants and utility functions
