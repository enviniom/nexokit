# Platform Boundary Rules

## Purpose

Define what belongs in `internal/platform/*` (cross-application contracts) vs `internal/modules/*/core/*` (domain-specific language). Platform packages MUST NOT contain domain-specific messages, constants, or error sentinels.

## Requirements

### Requirement: Platform package classification

The system MUST classify each `platform` subpackage as either **generic** (cross-application utility) or **domain-restricted** (must not accept domain-specific content):

| Subpackage | Classification | Allowed Content |
|------------|---------------|-----------------|
| `platform/messages` | Generic | API response messages, validation rule messages, middleware messages — zero domain-specific messages |
| `platform/apperror` | Generic | Sentinel errors mapped to HTTP statuses — messages MUST be generic, not domain-specific |
| `platform/permissions` | Generic | `Action*` constants, `Format()`, `Humanize*()`, `Register()`, `ListRegistered()` — zero `Module*` constants |
| `platform/response` | Generic | Single API response contract source — `APIResponse`, `HandleError`, helpers |
| `platform/tenant` | Generic | Tenant context and scoping utilities |
| `platform/identity` | Generic | Public ID generation |
| `platform/token` | Generic | Token utilities |
| `platform/password` | Generic | Password hashing/validation |
| `platform/validator` | Generic | Validation rules |
| `platform/gormutil` | Generic | GORM helpers |
| `platform/query` | Generic | Query parsing |
| `platform/authctx` | Generic | Auth context utilities |
| `platform/cache` | Generic | Cache adapter (if present) |

#### Scenario: No domain messages in platform/messages

- GIVEN `platform/messages/messages.go`
- WHEN all message constants are reviewed
- THEN zero messages reference domain concepts (roles, users, companies, etc.)

#### Scenario: No module constants in platform/permissions

- GIVEN `platform/permissions/constants.go`
- WHEN all constants are reviewed
- THEN zero `Module*` constants exist — only `Action*` constants and utility functions remain

#### Scenario: Generic sentinel messages in platform/apperror

- GIVEN `platform/apperror/apperror.go`
- WHEN all sentinel error messages are reviewed
- THEN zero messages reference domain concepts — `ErrUnprocessable` has no role-specific message

### Requirement: Module-owned domain language

Each module MUST define its own domain-specific language in `modules/<name>/core/`:

| Artifact | Location | Example |
|----------|----------|---------|
| Module name constant | `core/constants.go` | `const ModuleUsers = "users"` |
| Domain error messages | `core/messages.go` | `MsgRoleHasAssignedUsers` |
| Domain error sentinels | `core/error.go` | `ErrRoleHasAssignedUsers` |

#### Scenario: Module defines its own name constant

- GIVEN a module `users`
- WHEN `modules/users/core/constants.go` exists
- THEN it defines `const ModuleUsers = "users"`

#### Scenario: Module owns its error sentinel

- GIVEN the roles module needs a 422 error for "role has assigned users"
- WHEN `modules/roles/core/error.go` exists
- THEN it defines `ErrRoleHasAssignedUsers` mapping to HTTP 422 with `MsgRoleHasAssignedUsers`

### Requirement: platform/response as single response contract

The system MUST use `platform/response` as the sole source of API response contracts. No module MAY define its own response envelope or `HandleError` equivalent.

#### Scenario: All handlers use platform/response

- GIVEN any HTTP handler in any module
- WHEN it writes a response
- THEN it uses `platform/response` helpers (`Success`, `Error`, `NoContent`, `HandleError`)

## Constraints and Edge Cases

- `Action*` constants in `platform/permissions` MUST remain — they are generic permission verbs used across all modules
- Module constant duplication across modules IS acceptable — it avoids coupling between unrelated modules
- Moving domain language to modules MUST NOT change HTTP response shapes, status codes, or JSON envelopes
