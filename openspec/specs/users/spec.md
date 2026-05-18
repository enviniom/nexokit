# Users Specification

## Purpose

User management CRUD, password change, and active/inactive status.

## Requirements

### Requirement: User CRUD endpoints

The system MUST expose `GET /api/v1/users`, `POST /api/v1/users`, `GET /api/v1/users/:id`, `PUT /api/v1/users/:id`, and `DELETE /api/v1/users/:id`. All responses MUST use the standard DTO envelope and MUST NOT include `password` or `password_hash`.

#### Scenario: Create user

- GIVEN valid user data including `name`, `email`, `role_id`, and `password`
- WHEN `POST /api/v1/users` is called
- THEN the response returns HTTP 201, `success: true`, and `data` contains the created user without password fields
- AND the stored password is hashed with argon2id

#### Scenario: List users

- GIVEN multiple users exist
- WHEN `GET /api/v1/users` is called
- THEN the response returns HTTP 200 and no user object contains `password_hash`

#### Scenario: Update user

- GIVEN an existing user
- WHEN `PUT /api/v1/users/:id` is called with updated `name` or `email`
- THEN the response returns HTTP 200 and `data` contains the updated user without password fields

#### Scenario: Delete user

- GIVEN an existing user
- WHEN `DELETE /api/v1/users/:id` is called
- THEN the response returns HTTP 200 and the user is soft-deleted

### Requirement: Password change endpoint

The system MUST expose `PUT /api/v1/users/:id/password`. It MUST require the current password, verify it against the stored hash, and update to a new hash using argon2id.

#### Scenario: Successful password change

- GIVEN a user with current password `OldPass1!`
- WHEN `PUT /api/v1/users/:id/password` is called with `current_password: OldPass1!` and `new_password: NewPass1!`
- THEN the response returns HTTP 200 and the stored hash is updated

#### Scenario: Wrong current password

- GIVEN a user with current password `OldPass1!`
- WHEN the endpoint is called with `current_password: WrongPass1!`
- THEN the response returns HTTP 422 with a validation error message

### Requirement: User status toggle

The system MUST support toggling a user's `status` between `active` and `inactive`.

#### Scenario: Deactivate user

- GIVEN an active user
- WHEN `PUT /api/v1/users/:id` sets `status` to `inactive`
- THEN the user becomes inactive and can no longer authenticate

#### Scenario: Reactivate user

- GIVEN an inactive user
- WHEN `PUT /api/v1/users/:id` sets `status` to `active`
- THEN the user can authenticate again
