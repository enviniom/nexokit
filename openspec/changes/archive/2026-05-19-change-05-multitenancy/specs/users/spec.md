# Delta for Users

## MODIFIED Requirements

### Requirement: User CRUD endpoints

The system MUST expose CRUD endpoints for users. All responses MUST use the standard DTO envelope and MUST NOT include `password` or `password_hash`. For non-root users, all queries MUST filter by `company_id` from TenantContext. Root users with `IsRootScope=true` see all users; root users scoped to a specific company see only that company's users. The `company_id` field MUST be required for admin/user roles; root MAY have null `company_id`.
(Previously: Endpoints returned all users regardless of tenant; no company_id filtering or validation.)

#### Scenario: Create user with company_id

- GIVEN valid user data with `company_id`
- WHEN `POST /api/v1/users` is called
- THEN response returns HTTP 201 with created user, password fields excluded

#### Scenario: Create admin/user without company_id rejected

- GIVEN user data with admin or user role but no `company_id`
- WHEN `POST /api/v1/users` is called
- THEN response returns HTTP 422 with validation error on `company_id`

#### Scenario: Create root with nullable company_id

- GIVEN user data with root role and `company_id` omitted
- WHEN `POST /api/v1/users` is called
- THEN response returns HTTP 201 with `company_id` null

#### Scenario: Admin sees only own company's users

- GIVEN admin with `company_id = 1` and users in companies 1 and 2
- WHEN `GET /api/v1/users` is called with that admin's TenantContext
- THEN only users where `company_id = 1` are returned

#### Scenario: Root sees all users in global mode

- GIVEN a root user with `IsRootScope = true`
- WHEN `GET /api/v1/users` is called
- THEN users from all companies are returned

#### Scenario: Root scoped to one company

- GIVEN a root user scoped to `company_id = 3`
- WHEN `GET /api/v1/users` is called
- THEN only users where `company_id = 3` are returned

#### Scenario: Update user within tenant scope

- GIVEN an existing user within the requester's tenant scope
- WHEN `PUT /api/v1/users/:id` is called
- THEN response returns HTTP 200 with updated user

#### Scenario: Delete user within tenant scope

- GIVEN an existing user within the requester's tenant scope
- WHEN `DELETE /api/v1/users/:id` is called
- THEN response returns HTTP 200, user is soft-deleted

#### Scenario: Cross-tenant update returns 404

- GIVEN admin with `company_id = 1` targeting a user with `company_id = 2`
- WHEN `PUT /api/v1/users/:id` is called
- THEN response returns HTTP 404

### Requirement: Password change endpoint

The system MUST expose `PUT /api/v1/users/:id/password` requiring current password verification. The endpoint MUST verify the requester's `company_id` matches the target's, or the requester is root.
(Previously: No tenant scope check on password change.)

#### Scenario: Successful password change

- GIVEN a user changing their own password with valid current password
- WHEN `PUT /api/v1/users/:id/password` is called
- THEN response returns HTTP 200, hash is updated

#### Scenario: Wrong current password

- WHEN called with an incorrect `current_password`
- THEN response returns HTTP 422

#### Scenario: Cross-tenant password change blocked

- GIVEN admin with `company_id = 1` targeting a user with `company_id = 2`
- WHEN requesting `PUT /api/v1/users/:id/password`
- THEN response returns HTTP 404

### Requirement: User status toggle

The system MUST support toggling `status` between `active` and `inactive`. Non-root users MUST NOT change status of users in a different company.
(Previously: No tenant scope check on status toggle.)

#### Scenario: Deactivate same-company user

- GIVEN an active user in the same company as the requester
- WHEN `PUT /api/v1/users/:id` sets `status` to `inactive`
- THEN the user becomes inactive

#### Scenario: Reactivate same-company user

- GIVEN an inactive user in same company as requester
- WHEN `PUT /api/v1/users/:id` sets `status` to `active`
- THEN the user can authenticate again

#### Scenario: Cross-tenant status toggle blocked

- GIVEN admin with `company_id = 1` targeting a user with `company_id = 2`
- WHEN requesting status change
- THEN response returns HTTP 404