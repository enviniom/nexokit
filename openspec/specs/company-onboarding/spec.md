# Company Onboarding Specification

## Purpose

Root-only tenant onboarding that provisions a company, tenant system roles, and the first tenant administrator in one transaction.

## Requirements

### Requirement: Root-only company onboarding endpoint

The system MUST expose `POST /api/v1/onboarding/companies`. The endpoint MUST be protected by the `root` role, not by a permission slug, so tenant administrators cannot gain onboarding access through permission synchronization. Non-root authenticated users MUST receive HTTP 403. Unauthenticated users MUST receive HTTP 401.

#### Scenario: Root onboards company

- GIVEN an authenticated root user
- WHEN `POST /api/v1/onboarding/companies` is called with valid company and admin data
- THEN the response returns HTTP 201
- AND the response contains the created company public ID, company slug, admin public ID, and admin email

#### Scenario: Non-root cannot onboard company

- GIVEN an authenticated tenant admin or regular user
- WHEN `POST /api/v1/onboarding/companies` is called
- THEN the response returns HTTP 403
- AND no onboarding records are created

### Requirement: Transactional tenant provisioning

The system MUST execute company onboarding in a single database transaction. The transaction MUST create the company, tenant `admin` role, tenant `user` role, and initial admin user. The tenant `admin` and `user` roles MUST have `company_id` set to the new company and `is_system = true`. The system MUST NOT create or modify the global `root` role during onboarding.

#### Scenario: Provisioned tenant roles and admin user

- GIVEN valid onboarding input from root
- WHEN onboarding succeeds
- THEN a company exists with the requested slug
- AND tenant roles with slugs `admin` and `user` exist for that company
- AND the initial admin user exists with that company ID and the tenant `admin` role

#### Scenario: Duplicate company slug rolls back

- GIVEN an existing company with slug `acme`
- WHEN root submits onboarding with slug `acme`
- THEN the request returns a validation/conflict error
- AND no admin user or tenant roles are created for the failed onboarding attempt

#### Scenario: Duplicate admin email rolls back

- GIVEN an existing user with email `jane@acme.com`
- WHEN root submits onboarding using `admin_email = "jane@acme.com"`
- THEN the request returns a validation/conflict error
- AND the requested company is not created

### Requirement: Onboarding role permission assignment

During onboarding, the system MUST assign all currently registered permissions to the new tenant `admin` role. The system MUST assign only the base user permission subset to the new tenant `user` role.

#### Scenario: Tenant admin receives all current permissions

- GIVEN registered permissions exist in the database
- WHEN a company is onboarded
- THEN the tenant `admin` role has all registered permissions assigned

#### Scenario: Tenant user receives base permissions only

- GIVEN registered permissions exist in the database
- WHEN a company is onboarded
- THEN the tenant `user` role has only the base permissions intended for normal tenant users

### Requirement: Direct company creation disabled

The system MUST NOT expose `POST /api/v1/companies`. Company creation MUST go through `POST /api/v1/onboarding/companies`.

#### Scenario: Direct company creation route is absent

- GIVEN any user, including root
- WHEN `POST /api/v1/companies` is called
- THEN the response returns HTTP 404
