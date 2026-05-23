# Delta for Tenant-Scoped Roles

## ADDED Requirements

### Requirement: Tenant-aware assignable roles

The system MUST apply tenant isolation to every role selected for user assignment. A non-root requester MUST only see and assign roles whose `company_id` matches the requester's TenantContext. A requester MUST NOT assign a role from another company to a user in their own company. The global `root` role MUST NOT be assignable and MUST NOT appear in assignable role lists.

Root authorization bypass does not remove assignment safety rules: root MAY administer tenant roles according to TenantContext, but root MUST NOT assign the `root` role through the API. If root operates in global scope and can see roles/users across companies, assignment MUST still require the target user's `company_id` to match the target role's `company_id` when both are set.

#### Scenario: Non-root assignable roles scoped to company

- GIVEN admin with `company_id = 1`
- AND roles exist for companies 1 and 2
- WHEN the admin requests assignable roles
- THEN only company 1 roles are returned

#### Scenario: Non-root cannot assign another company's role

- GIVEN admin with `company_id = 1`
- AND a target user in company 1
- AND a role belonging to company 2
- WHEN the admin attempts to assign the company 2 role
- THEN the system returns HTTP 404
- AND the target user's role remains unchanged

#### Scenario: Root remains unassignable

- GIVEN the global `root` role exists with `company_id = NULL`
- WHEN any requester asks for assignable roles or attempts a role assignment
- THEN `root` is excluded or rejected

#### Scenario: Root global cannot cross-assign tenant roles

- GIVEN root is operating in global scope
- AND a target user belongs to company 1
- AND a role belongs to company 2
- WHEN root attempts to assign the company 2 role to the company 1 user
- THEN the system returns HTTP 403 or HTTP 422
- AND the target user's role remains unchanged
