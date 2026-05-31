# Delta for Platform Boundary Rules

## MODIFIED Requirements

### Requirement: Platform package classification

The system MUST classify each `platform` subpackage as either **generic** (cross-application utility) or **domain-restricted** (must not accept domain-specific content):

| Subpackage | Classification | Allowed Content |
|------------|---------------|-----------------|
| `platform/messages` | Generic | API response messages, validation rule messages, middleware messages — zero domain-specific messages |
| `platform/apperror` | Generic | Sentinel errors mapped to HTTP statuses — messages MUST be generic, not domain-specific |
| `platform/permissions` | Generic | `Action*` constants, `Format()`, `Humanize*()`, `Register()`, `ListRegistered()` — zero `Module*` constants |
| `platform/response` | Generic | Single API response contract source — `APIResponse`, `HandleError`, helpers; imports `validator` for `ValidationErrors` |
| `platform/tenant` | Generic | Tenant context and scoping utilities |
| `platform/identity` | Generic | Public ID generation |
| `platform/token` | Generic | Token utilities |
| `platform/password` | Generic | Password hashing/validation |
| `platform/validator` | Generic | Validation primitives: `ValidationErrors`, `FieldValidator`, `Rule` functions — owns the error accumulator |
| `platform/gormutil` | Generic | GORM helpers |
| `platform/query` | Generic | Query parsing |
| `platform/authctx` | Generic | Auth context utilities |
| `platform/cache` | Generic | Cache adapter (if present) |

(Previously: `validator` classified as "Validation rules" only; now owns `ValidationErrors` and `FieldValidator`. `response` table entry updated to note it imports `validator`.)

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
