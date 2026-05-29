# Delta for Vertical Slice Modules

## ADDED Requirements

### Requirement: Auth module migration

The `auth` module SHALL be the second migration target to vertical slice architecture, following the `companies` module pattern. It SHALL have 4 slices: `authenticate_user`, `rotate_token`, `revoke_token`, `view_session`. Each slice SHALL co-locate handler, service, repository, and tests. The auth module SHALL eliminate all cross-module imports to `modules/users` and `modules/roles`, using local partial models in `core/` and reusable queries in `queries/`.

#### Scenario: Auth module has 4 slices

- GIVEN the auth module migration is complete
- WHEN the `internal/modules/auth/` directory is inspected
- THEN it contains `authenticate_user/`, `rotate_token/`, `revoke_token/`, `view_session/` subdirectories
- AND each contains handler, service, repository, and test files

#### Scenario: Auth module has no cross-module imports

- GIVEN the auth module migration is complete
- WHEN auth source files are inspected for imports
- THEN no file imports `modules/users` or `modules/roles`
- AND user data is accessed via local partial models in `auth/core/`

#### Scenario: Auth module uses queries/ package

- GIVEN the auth module migration is complete
- WHEN shared DB lookup logic is needed by multiple slices
- THEN it lives in `auth/queries/` with a corresponding `_test.go` per query file

## MODIFIED Requirements

### Requirement: Incremental migration pattern

Existing modules SHALL NOT be mass-migrated unless already undergoing substantial change. New or non-trivial modules SHALL use the vertical slice pattern from the start. Simple modules MAY remain flat. The `companies` module SHALL be the pilot migration target. The `auth` module SHALL be the second migration target, using `companies` as the reference pattern.
(Previously: Only companies module was the pilot migration target.)

#### Scenario: Companies module is migrated

- GIVEN the `companies` module currently uses flat files
- WHEN the migration is complete
- THEN `companies` uses vertical slice organization with 7 slices: `list_companies`, `view_company`, `update_company`, `delete_company`, `list_company_domains`, `create_company_domain`, `update_company_domain`
- AND it has no `create_company` slice

#### Scenario: Auth module is migrated

- GIVEN the `auth` module currently uses flat files
- WHEN the migration is complete
- THEN `auth` uses vertical slice organization with 4 slices: `authenticate_user`, `rotate_token`, `revoke_token`, `view_session`
- AND it has `core/` with local partial models and `queries/` with reusable queries

#### Scenario: Remaining modules unchanged

- GIVEN modules `users`, `roles`, `permissions`, `onboarding` exist with flat structure
- WHEN the auth migration is complete
- THEN these modules retain their existing flat file structure
- AND they continue to function without modification

#### Scenario: New modules use vertical slices

- GIVEN a new module is created after this pattern is established
- WHEN the module scaffold is generated
- THEN it uses vertical slice organization by default
