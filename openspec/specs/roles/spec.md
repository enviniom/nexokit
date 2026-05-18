# Roles Specification

## Purpose

Read-only role management and system seeding.

## Requirements

### Requirement: Read-only role API

The system MUST expose `GET /api/v1/roles` and `GET /api/v1/roles/:id`. It MUST NOT expose mutation endpoints for roles. Responses MUST use the standard DTO envelope.

#### Scenario: List roles

- GIVEN the system seeds have run
- WHEN `GET /api/v1/roles` is called
- THEN the response returns HTTP 200 and `data` contains the roles `root`, `admin`, and `user`

#### Scenario: Get role by ID

- GIVEN a seeded role exists
- WHEN `GET /api/v1/roles/:id` is called with a valid ID
- THEN the response returns HTTP 200 and `data` contains the role

### Requirement: Role seeds

The system MUST seed the roles `root`, `admin`, and `user` on startup or via a seed command. The operation MUST be idempotent.

#### Scenario: Idempotent seeding

- GIVEN the roles have already been seeded
- WHEN the seed command runs again
- THEN no duplicate roles are created and the process exits successfully

### Requirement: System role protection

The system MUST mark seeded roles as system-level (`is_system: true`) to distinguish them from future custom roles.

#### Scenario: System flag present

- GIVEN the seeded roles exist
- WHEN a role is retrieved
- THEN `is_system` is `true`
