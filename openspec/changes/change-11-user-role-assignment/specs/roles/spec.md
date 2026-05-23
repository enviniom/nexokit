# Delta for Roles

## ADDED Requirements

### Requirement: Assignable roles endpoint

The system MUST expose `GET /api/v1/roles/assignable` to return roles that may be assigned to users through the dedicated role assignment flow. The endpoint MUST require `roles.assignable`. It MUST be registered before `GET /api/v1/roles/:id` so the static `assignable` segment is not interpreted as a role PublicID.

The endpoint MUST return role select/list data using role PublicIDs, not internal numeric IDs. The `root` role MUST NOT appear in the response, even for root actors. For non-root users, the endpoint MUST return only roles belonging to the requester's company. Root users in global scope MAY see tenant roles across companies for administrative selection, but MUST still not see `root`. Cross-company visibility for root global selection MUST NOT imply cross-company assignment permission; assignment MUST still verify target user and role company compatibility.

#### Scenario: List assignable roles for same company

- GIVEN admin with `company_id = 1` and `roles.assignable` permission
- AND company 1 has roles `admin` and `manager`
- WHEN `GET /api/v1/roles/assignable` is called
- THEN response returns HTTP 200
- AND `data` contains only roles from company 1
- AND each role uses its PublicID as `id`

#### Scenario: Root role excluded from assignable roles

- GIVEN the global `root` role exists
- WHEN `GET /api/v1/roles/assignable` is called by any authenticated requester with `roles.assignable`
- THEN response returns HTTP 200
- AND no item has slug `root`

#### Scenario: Cross-company roles excluded for non-root

- GIVEN admin with `company_id = 1`
- AND company 2 has role `manager`
- WHEN `GET /api/v1/roles/assignable` is called
- THEN the company 2 role is not present in `data`

#### Scenario: Assignable roles require permission

- GIVEN a requester without `roles.assignable`
- WHEN `GET /api/v1/roles/assignable` is called
- THEN response returns HTTP 403
