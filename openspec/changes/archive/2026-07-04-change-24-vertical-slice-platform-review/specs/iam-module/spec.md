# Delta for iam-module

## Purpose

Pin the contract that reusable queries are unique and preload behavior is covered by a regression test. After this delta, the IAM module MUST NOT ship byte-identical query files that differ only by name, and the preload behavior of `GetRoleByPublicID` MUST be pinned by a test.

## MODIFIED Requirements

### Requirement: Reusable query uniqueness

The IAM module MUST NOT ship two reusable query files whose bodies are byte-identical except for the function name. When a query has both a "preload" and a "non-preload" variant, the module MUST ship exactly one variant and document the preload decision in the function's godoc. Callers that need different preload sets MUST use a single query plus a clearly-named `...With` helper, not a duplicate file.
(Previously: `internal/modules/iam/queries/get_role_by_public_id.go` and `get_role_by_public_id_preloads.go` were byte-identical except for the function name; one was dead code.)

#### Scenario: Only one role-by-public-id query exists

- GIVEN the IAM module at `internal/modules/iam/queries/`
- WHEN the change is complete
- THEN exactly one file defines the role-by-public-id query
- AND the file is named `get_role_by_public_id.go`
- AND no file named `get_role_by_public_id_preloads.go` exists

#### Scenario: Callers use the single query

- GIVEN the deleted duplicate query is no longer present
- WHEN all callers under `internal/modules/iam/` are inspected
- THEN they import and call the surviving `GetRoleByPublicID` function
- AND no caller references the deleted `GetRoleByPublicIDPreloads` function

### Requirement: Preload regression test for GetRoleByPublicID

The query `GetRoleByPublicID(db, tc, publicID)` MUST preload the `Company` and `Permissions` associations when a matching role is found. A regression test under `internal/modules/iam/queries/get_role_by_public_id_test.go` MUST seed a role with associated company and permission rows, call the query, and assert that the returned `*core.IAMRole` carries the preloaded `Company` and `Permissions` values.
(Previously: the preload behavior was documented in the function body but not pinned by a test; deleting the duplicate file would have made this gap visible.)

#### Scenario: Preloads Company association

- GIVEN a role with `CompanyID = 1` and a company row with `id = 1`
- WHEN `GetRoleByPublicID(db, root, rolePublicID)` is called
- THEN `result.Company.ID == 1` and the company is loaded in a single query

#### Scenario: Preloads Permissions association

- GIVEN a role with two `RolePermission` rows linked to two `Permission` rows
- WHEN `GetRoleByPublicID(db, root, rolePublicID)` is called
- THEN `result.Permissions` has length 2 and contains the seeded permission slugs

#### Scenario: Not-found path is preserved

- GIVEN a non-existent publicID
- WHEN `GetRoleByPublicID(db, root, publicID)` is called
- THEN it returns `gorm.ErrRecordNotFound`
- AND callers (repositories) translate it to the matching module sentinel
