# Delta for Permissions

## MODIFIED Requirements

### Requirement: Permission seeds

The system MUST seed base permissions on startup or via seed command using explicit actions:

| Module | Actions | Business Actions |
|--------|---------|------------------|
| users | index, view, create, update, delete | change_role |
| roles | index, view, create, update, delete | assign_permissions |
| companies | index, view, create, update, delete | — |
| settings | view, update | — |
| auth | view | — |
| permissions | manage | — |

Slugs follow `{module}.{action}` (e.g., `roles.create`, `roles.update`, `roles.delete`). The operation MUST be idempotent. All seeded permissions MUST have `is_system: true` and explicit `display_order` values.
(Previously: `roles` module only seeded `index`, `view` actions; `create`, `update`, `delete` were not seeded)

#### Scenario: Idempotent seeding

- GIVEN permissions have already been seeded
- WHEN the seed command runs again
- THEN no duplicate permissions are created and the process exits successfully

#### Scenario: Role management permissions are seeded

- GIVEN the system seeds have run
- WHEN permissions are queried by module `roles`
- THEN `roles.create`, `roles.update`, and `roles.delete` are present alongside `roles.index`, `roles.view`, and `roles.assign_permissions`
