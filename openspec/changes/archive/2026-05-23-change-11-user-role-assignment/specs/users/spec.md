# Delta for Users

## MODIFIED Requirements

### Requirement: User CRUD endpoints

The system MUST expose `GET /api/v1/users`, `POST /api/v1/users`, `GET /api/v1/users/:id`, `PUT /api/v1/users/:id`, and `DELETE /api/v1/users/:id`. Responses MUST use the standard DTO envelope except successful DELETE responses, which MUST return HTTP 204 with no body. Response DTOs MUST NOT include `password` or `password_hash`. For non-root users, all queries MUST filter by `company_id` from TenantContext. Root users with `IsRootScope=true` see all users; root users scoped to a specific company see only that company's users. The `company_id` field MUST be required for admin/user roles; root MAY have null `company_id`.

`PUT /api/v1/users/:id` MUST be limited to general user fields and MUST require only `users.update`. It MUST NOT accept `role_id` as a meaningful update field and MUST NOT modify the target user's role even when `role_id` is present in the JSON body. Non-root general updates MUST NOT move a user to another company.
(Previously: `PUT /api/v1/users/:id` required both `users.update` and `users.change_role`, required `role_id`, and changed `role_id` in the general update flow.)

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
- WHEN `PUT /api/v1/users/:id` is called with general user fields
- THEN response returns HTTP 200 with updated user

#### Scenario: General update ignores role_id

- GIVEN an existing user with role `user`
- AND the requester has `users.update` permission but not `users.change_role`
- WHEN `PUT /api/v1/users/:id` is called with valid general fields and a `role_id` field in the JSON body
- THEN response returns HTTP 200
- AND the user's role remains `user`

#### Scenario: General update does not require role_id

- GIVEN an existing user within the requester's tenant scope
- AND the requester has `users.update` permission
- WHEN `PUT /api/v1/users/:id` is called without `role_id`
- THEN response returns HTTP 200 with updated general fields

#### Scenario: General update cannot move user to another company

- GIVEN admin with `company_id = 1` targeting a user in company 1
- WHEN `PUT /api/v1/users/:id` includes `company_id` for company 2
- THEN the response returns HTTP 403 or HTTP 422
- AND the user's `company_id` remains unchanged

#### Scenario: Delete user within tenant scope

- GIVEN an existing user within the requester's tenant scope
- WHEN `DELETE /api/v1/users/:id` is called
- THEN response returns HTTP 204 with an empty body
- AND the user is soft-deleted

#### Scenario: Cross-tenant update returns 404

- GIVEN admin with `company_id = 1` targeting a user with `company_id = 2`
- WHEN `PUT /api/v1/users/:id` is called
- THEN response returns HTTP 404

## ADDED Requirements

### Requirement: Dedicated user role assignment endpoint

The system MUST expose `PATCH /api/v1/users/:id/role` to change a user's role. The endpoint MUST require `users.change_role`. The request body MUST contain `role_id` as the target role PublicID string. The endpoint MUST NOT accept internal numeric role IDs as its public contract.

The system MUST reject attempts to assign the `root` role through this endpoint. The system MUST reject self role changes through this endpoint. The system MUST enforce tenant scope for both the target user and target role: non-root users MUST only assign roles belonging to their own company, and MUST NOT assign roles from another company. Root bypass remains global for authorization, but root still MUST NOT use this endpoint to assign the `root` role to any user.

The system MUST explicitly verify company compatibility between the target user and target role. When both records have `company_id`, the values MUST match before assignment. This invariant MUST be enforced even when the requester is root in global scope and both records are otherwise visible.

After a successful role change, the system MUST invalidate cached permissions for the target user.

#### Scenario: Change user role with permission

- GIVEN an admin with `users.change_role` permission
- AND a target user in the same company
- AND an assignable role in the same company
- WHEN `PATCH /api/v1/users/:id/role` is called with the role PublicID
- THEN response returns HTTP 200 with the updated user
- AND the target user's role is changed
- AND the target user's permission cache is invalidated

#### Scenario: Role change without permission is forbidden

- GIVEN a requester without `users.change_role` permission
- WHEN `PATCH /api/v1/users/:id/role` is called
- THEN response returns HTTP 403
- AND the target user's role is unchanged

#### Scenario: Root role cannot be assigned

- GIVEN the `root` role exists
- WHEN `PATCH /api/v1/users/:id/role` is called with the root role PublicID
- THEN response returns HTTP 403 or HTTP 422
- AND the target user's role is unchanged

#### Scenario: Cross-company role cannot be assigned by non-root

- GIVEN admin with `company_id = 1`
- AND a target user in company 1
- AND a role belonging to company 2
- WHEN `PATCH /api/v1/users/:id/role` is called with the company 2 role PublicID
- THEN response returns HTTP 404
- AND the target user's role is unchanged

#### Scenario: Root global cannot assign role across companies

- GIVEN root is operating in global scope
- AND a target user in company 1
- AND a non-root role belonging to company 2
- WHEN `PATCH /api/v1/users/:id/role` is called with the company 2 role PublicID
- THEN response returns HTTP 403 or HTTP 422
- AND the target user's role is unchanged

#### Scenario: User cannot change own role

- GIVEN a user with `users.change_role` permission
- WHEN the user calls `PATCH /api/v1/users/:id/role` targeting their own PublicID
- THEN response returns HTTP 403
- AND the user's role is unchanged

#### Scenario: Missing role_id is invalid

- GIVEN a requester with `users.change_role` permission
- WHEN `PATCH /api/v1/users/:id/role` is called without `role_id`
- THEN response returns HTTP 422

#### Scenario: Unknown role returns not found

- GIVEN a requester with `users.change_role` permission
- WHEN `PATCH /api/v1/users/:id/role` is called with an unknown role PublicID
- THEN response returns HTTP 404
- AND the target user's role is unchanged
