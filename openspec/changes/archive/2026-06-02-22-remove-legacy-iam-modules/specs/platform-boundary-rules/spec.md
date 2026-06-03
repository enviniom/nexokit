# Delta for Platform Boundary Rules

## MODIFIED Requirements

### Requirement: Module-owned domain language

Each module MUST define its own domain-specific language in `modules/<name>/core/`:

| Artifact | Location | Example |
|----------|----------|---------|
| Module name constant | `core/constants.go` | `const ModuleIAM = "iam"` |
| Domain error messages | `core/messages.go` | `MsgRoleHasAssignedUsers` |
| Domain error sentinels | `core/error.go` | `ErrRoleHasAssignedUsers` |

After legacy removal, the IAM module is the sole owner of user, role, and permission domain language. Legacy module paths (`modules/users/core/`, `modules/roles/core/`, `modules/permissions/core/`) no longer exist.
(Previously: Examples referenced `modules/users/core/constants.go` and `modules/roles/core/error.go`; those directories are deleted — IAM owns all user/role/permission domain language)

#### Scenario: IAM module owns domain language

- GIVEN the IAM module at `modules/iam/core/`
- WHEN domain constants, messages, and error sentinels are reviewed
- THEN they are defined within the IAM module's core package

#### Scenario: Module owns its error sentinel

- GIVEN the IAM module needs a 422 error for "role has assigned users"
- WHEN `modules/iam/core/error.go` exists
- THEN it defines `ErrRoleHasAssignedUsers` mapping to HTTP 422 with `MsgRoleHasAssignedUsers`

#### Scenario: No legacy module domain language references

- GIVEN the full codebase after legacy removal
- WHEN imports and references are reviewed
- THEN zero paths reference `modules/users/core/`, `modules/roles/core/`, or `modules/permissions/core/`
