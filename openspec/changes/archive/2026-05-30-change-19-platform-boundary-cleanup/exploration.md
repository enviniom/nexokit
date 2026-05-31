## Exploration: 19-platform-boundary-cleanup

### Current State

`internal/platform/` contains 12 subpackages. After auditing every file, the classification is:

#### ✅ Generic (correctly in platform)
| Package | Reason |
|---------|--------|
| `response` | Standard API envelope, pagination, validation response types — cross-application contract |
| `apperror` | Generic error categories (NotFound, Forbidden, Unauthorized, etc.) with HTTP status mapping |
| `validator` | Generic validation rules (MinLength, ValidEmail, ValidSlug, etc.) and FieldValidator chain |
| `query` | Generic query parameter parsing (pagination, filters, sorting, search) |
| `tenant` | Request-scoped tenant context and GORM scope helpers — cross-application |
| `authctx` | Authenticated user context extraction — cross-application |
| `password` | Argon2id hashing primitives — technical utility |
| `token` | PASETO access/refresh token management — technical utility |
| `identity` | ULID generation — technical utility |
| `gormutil` | Generic GORM helpers (pagination, sorting, search, date range) |

#### ⚠️ Domain-specific elements mixed in platform
| Element | Location | Domain | Issue |
|---------|----------|--------|-------|
| `MsgRoleHasAssignedUsers` | `messages/messages.go` | roles | Role-specific error message in global messages |
| `ErrUnprocessable` → `MsgRoleHasAssignedUsers` | `apperror/apperror.go` | roles | Generic error category carrying domain-specific message |
| `ModuleUsers`, `ModuleRoles`, `ModuleCompanies`, `ModuleSettings`, `ModuleAuth`, `ModulePermissions` | `permissions/constants.go` | cross-module | Module name constants for permission slugs — borderline; used by middleware authorization and all module routes |
| `Action*` constants + `Format`, `HumanizeName`, `HumanizeDescription`, `DefaultDisplayOrder` | `permissions/constants.go` | permissions | Permission action constants and formatting helpers — used by middleware + module routes |
| `permissions.Registry` (Register/ListRegistered) | `permissions/registry.go` | permissions | In-memory permission slug registry — used by middleware authorization |

#### ℹ️ Messages classification detail
`messages/messages.go` contains:
- **Generic API messages** (OK to stay): `MsgSuccess`, `MsgCreated`, `MsgDeleted`, `MsgHealthy`, `MsgValidationError`, `MsgInternalError`, `MsgNotFound`, `MsgUnauthorized`, `MsgForbidden`, `MsgConflict`, `MsgBadRequest`, `MsgTooManyRequests`
- **Generic validation messages** (OK to stay): `MsgRequired`, `MsgMinLength`, `MsgMaxLength`, `MsgValidEmail`, `MsgHasUppercase`, `MsgHasDigit`, `MsgHasSpecialChar`, `MsgMinWords`, `MsgNoNumbers`, `MsgInvalidFormat`, `MsgValidSlug`, `MsgValidURL`, `MsgInList`
- **Generic middleware messages** (OK to stay): `MsgPanicRecovered`, `MsgPanicLog`, `MsgHTTPRequest`, `CtxRequestID`, `HeaderRequestID`, CORS constants
- **Domain-specific** (move): `MsgRoleHasAssignedUsers` → `internal/modules/roles/core/`

### Affected Areas

- `internal/platform/messages/messages.go` — remove `MsgRoleHasAssignedUsers`
- `internal/platform/apperror/apperror.go` — `ErrUnprocessable` needs domain-specific message handling
- `internal/platform/permissions/constants.go` — module/action constants and helpers
- `internal/platform/permissions/registry.go` — permission registry
- `internal/middleware/authorization.go` — imports `platform/permissions` for action constants and registry
- `internal/modules/*/routes.go` — all 5 modules import `platform/permissions` as `platformPerms`
- `internal/modules/roles/service.go` — uses `apperror.ErrUnprocessable`
- `internal/modules/roles/core/error.go` — needs new `ErrHasAssignedUsers` sentinel
- `internal/modules/roles/core/messages.go` — new file for `MsgRoleHasAssignedUsers`

### Approaches

1. **Minimal: Move only `MsgRoleHasAssignedUsers`**
   - Move the single domain-specific message to `roles/core/`
   - Redefine `ErrUnprocessable` in `apperror` as a generic category without a fixed message, or keep it but have roles module define its own `ErrRoleHasAssignedUsers` wrapping `apperror.ErrUnprocessable`
   - Pros: Smallest change, lowest risk, proves the pattern
   - Cons: Doesn't address `permissions` package ambiguity
   - Effort: Low

2. **Moderate: Move domain messages + clarify `permissions` boundary**
   - Approach 1 + move `permissions` package:
     - Keep `permissions.Format()` and `permissions.Register/ListRegistered()` in platform (generic utilities)
     - Move module name constants (`ModuleUsers`, etc.) to each module's `core/constants.go` or to a shared `internal/shared/` constants file
     - Keep action constants (`ActionList`, etc.) in platform (they're generic CRUD verbs)
   - Pros: Cleaner boundary, `permissions` becomes clearly generic utility
   - Cons: More files touched, need to update all module routes
   - Effort: Medium

3. **Full: Strict domain isolation**
   - Approach 2 + each module defines its own error sentinels in `core/error.go` instead of using `apperror.Err*` directly
   - `apperror` becomes purely a category/mapper (types + HTTP status mapping), modules map their own errors to categories
   - Pros: Cleanest architecture, modules fully own their error language
   - Cons: Significant refactor across all modules, high risk of subtle behavior changes
   - Effort: High

### Recommendation

**Approach 2 (Moderate)** — it achieves the stated goal without overreaching:

1. Move `MsgRoleHasAssignedUsers` from `platform/messages` to `modules/roles/core/`
2. Create `modules/roles/core/error.go` with `ErrRoleHasAssignedUsers` that maps to 422
3. Keep `apperror.ErrUnprocessable` as a generic 422 category sentinel (remove domain-specific message)
4. Clarify `platform/permissions`:
   - Keep `Action*` constants (generic CRUD verbs)
   - Keep `Format()`, `HumanizeName()`, `HumanizeDescription()`, `DefaultDisplayOrder()` (generic utilities)
   - Keep `Register()`/`ListRegistered()` (generic registry)
   - Move `Module*` constants to each owning module's `core/constants.go` (or create a single `internal/shared/module_names.go` if duplication is undesirable)
5. Document the platform boundary rules in the change's design/tasks

### Risks

- **`ErrUnprocessable` semantic change**: Currently `ErrUnprocessable` carries `MsgRoleHasAssignedUsers`. Removing the message means `PublicMessage()` returns the underlying error's message. Need to ensure roles module handles this correctly.
- **Module constants duplication**: Moving `ModuleUsers`, `ModuleRoles`, etc. to each module means duplication. This is acceptable per context rules (section 5: "repetir modelos parciales entre módulos está permitido si evita acoplamiento"), but needs explicit documentation.
- **Middleware dependency on `permissions`**: `internal/middleware/authorization.go` imports `platform/permissions` for `ActionView`, `Format()`, and the registry. This is correct — middleware is cross-application and permission checking is generic.

### Circular Import Risk

**None identified.** The proposed moves are:
- `platform/messages` → `modules/roles/core/` (one direction, no cycle)
- `platform/permissions` constants → `modules/*/core/` or `shared/` (one direction, no cycle)

Modules already import from platform; moving things *out* of platform into modules or shared cannot create cycles.

### Ready for Proposal

**Yes.** The scope is well-defined, incremental, and preserves API contracts. The orchestrator should proceed to `sdd-propose` with Approach 2 as the recommended path.
