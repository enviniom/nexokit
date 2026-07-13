# Delta for Vertical Slice Modules

## ADDED Requirements

### Requirement: Companies module uses slice-root rehome

The companies module MUST keep its seven use cases under `internal/modules/companies/slices/`: `list_companies`, `view_company`, `update_company`, `delete_company`, `list_company_domains`, `create_company_domain`, and `update_company_domain`. The module root MUST remain wiring-only, with container, routes, resolver, and compatibility aliases only. Public HTTP behavior MUST NOT change because of the rehome.
(Previously: the companies module still used flat root slice packages.)

#### Scenario: Companies is rehomed

- GIVEN the companies module tree is inspected
- WHEN its directories are listed
- THEN the seven slices are under `slices/`
- AND the root contains wiring only

#### Scenario: Other modules stay unchanged

- GIVEN the codebase after the move
- WHEN non-companies modules are inspected
- THEN their layouts are unchanged
- AND no new `create_company` slice appears
