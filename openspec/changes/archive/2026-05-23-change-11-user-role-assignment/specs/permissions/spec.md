# Delta for Permissions

## MODIFIED Requirements

### Requirement: Permission seeds

The system MUST seed base permissions on startup or via seed command using explicit actions:

| Module | Actions | Business Actions |
|--------|---------|------------------|
| users | index, view, create, update, delete | change_role |
| roles | index, view, create, update, delete | assign_permissions, assignable |
| companies | index, view, create, update, delete | — |
| settings | view, update | — |
| auth | view | — |
| permissions | manage | — |

Slugs follow `{module}.{action}` (e.g., `users.change_role`, `roles.assignable`, `roles.assign_permissions`). The operation MUST be idempotent. All seeded permissions MUST have `is_system: true` and explicit `display_order` values.
(Previously: roles business actions included `assign_permissions` but not `assignable`.)

#### Scenario: Idempotent seeding

- GIVEN permissions have already been seeded
- WHEN the seed command runs again
- THEN no duplicate permissions are created and the process exits successfully

#### Scenario: Roles assignable permission seeded

- GIVEN the permission seed command has run
- WHEN permission `roles.assignable` is queried
- THEN it exists with module `roles`, action `assignable`, and `is_system = true`
