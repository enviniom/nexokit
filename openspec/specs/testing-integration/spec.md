# Testing Integration Specification

## Purpose

Define integration test infrastructure for NexoKit: test helpers, fixtures, and integration test suites that verify connected components work together.

## Requirements

### Requirement: Test database helper

The system MUST provide `tests/helpers/database.go` that creates an isolated SQLite `:memory:` database with GORM for integration tests. Each call MUST return a fresh database instance.

#### Scenario: Fresh database per test

- GIVEN an integration test calls the database helper
- WHEN the helper returns a `*gorm.DB`
- THEN the database is empty and ready for seeding
- AND it does not share state with other tests

#### Scenario: Cleanup after test

- GIVEN an integration test uses the database helper
- WHEN the test completes (via `t.Cleanup()`)
- THEN the database connection is closed
- AND no residual state affects subsequent tests

### Requirement: Test auth helper

The system MUST provide `tests/helpers/auth.go` with functions to create test users and generate valid authentication tokens for integration tests.

#### Scenario: Create authenticated request context

- GIVEN an integration test needs an authenticated user
- WHEN the auth helper creates a user and token
- THEN a valid token is returned that passes auth middleware
- AND the user exists in the test database

### Requirement: Test fixtures helper

The system MUST provide `tests/helpers/fixtures.go` and a `tests/fixtures/` directory with factory functions for creating test data (users, companies, roles, permissions).

#### Scenario: Seed test data with fixtures

- GIVEN an integration test needs pre-populated data
- WHEN fixture factory functions are called
- THEN valid records are created in the test database
- AND relationships (user→company, user→role) are correctly established

### Requirement: Integration test structure

Integration tests MUST live under `tests/integration/` with files named `<domain>_test.go`. Each test file MUST use `testing.Short()` to allow skipping when running unit-only suites.

#### Scenario: Integration tests are skippable

- GIVEN `go test ./... -short` is executed
- WHEN integration tests check `testing.Short()`
- THEN integration tests are skipped
- AND the test suite completes quickly

#### Scenario: Integration tests run fully

- GIVEN `go test ./tests/integration/...` is executed without `-short`
- WHEN integration tests run
- THEN all integration test cases execute
- AND they use the SQLite `:memory:` database

### Requirement: Auth integration tests

The system MUST provide `tests/integration/auth_test.go` covering: successful login, invalid credentials, inactive user login, valid refresh token, revoked refresh token, and logout token revocation.

#### Scenario: Successful login flow

- GIVEN a valid user exists in the test database
- WHEN a login request is sent with correct credentials
- THEN the response contains a valid access token
- AND HTTP status is 200

#### Scenario: Login with invalid credentials

- GIVEN a user exists in the test database
- WHEN a login request is sent with wrong password
- THEN the response is an error
- AND HTTP status is 401

### Requirement: Tenant isolation integration tests

The system MUST provide `tests/integration/tenant_test.go` verifying that admins can only access their own company's data, cannot access other companies' data, and that root users have global access.

#### Scenario: Admin cannot access other company data

- GIVEN two companies exist with separate admin users
- WHEN company A's admin queries company B's resources
- THEN the request returns 403 or empty results
- AND no data leakage occurs

### Requirement: RBAC integration tests

The system MUST provide `tests/integration/rbac_test.go` verifying: users with permissions can access resources, users without permissions receive 403, unauthenticated users receive 401, and root users have all permissions.

#### Scenario: User without permission receives 403

- GIVEN a user exists without a specific permission
- WHEN the user requests a protected endpoint requiring that permission
- THEN the response is 403 Forbidden
- AND the request does not reach the handler logic

### Requirement: Health integration tests

The system MUST provide `tests/integration/health_test.go` verifying the health endpoint returns correct status and component health information.

#### Scenario: Health endpoint returns OK

- GIVEN the application is running
- WHEN a GET request is sent to the health endpoint
- THEN the response indicates healthy status
- AND HTTP status is 200

## Constraints

- Integration tests MUST use SQLite `:memory:` (no PostgreSQL dependency for now).
- Fixtures MUST be deterministic — same input produces same output.
- No testify dependency SHALL be required (stdlib-only assertions).
