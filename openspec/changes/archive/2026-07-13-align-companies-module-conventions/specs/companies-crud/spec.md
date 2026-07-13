# Delta for Companies CRUD

## ADDED Requirements

### Requirement: Companies HTTP surface remains exact

The system MUST preserve the existing companies routes, HTTP methods, response envelopes, payload shapes, status codes, `:id` semantics, compatibility aliases in `model.go` and `dto.go`, and the module resolver contract. Root wiring MUST stay wiring-only. The system MUST NOT add `POST /api/v1/companies`; direct creation remains onboarding-only.
(Previously: public contract stability was implicit while the module layout drifted.)

#### Scenario: CRUD responses stay stable

- GIVEN existing list, view, update, and delete requests
- WHEN they are executed after the rehome
- THEN the same paths, methods, payloads, and statuses are returned
- AND `:id` still means `PublicID`

#### Scenario: Direct create remains absent

- GIVEN any authenticated user
- WHEN `POST /api/v1/companies` is called
- THEN the response is 404
- AND no create-company slice is introduced

#### Scenario: Compatibility aliases remain usable

- GIVEN code imports companies root DTO or model aliases, or middleware resolves companies
- WHEN the module is compiled
- THEN `model.go`, `dto.go`, and `Resolver()` remain available
- AND routing still delegates through the module container
