# Delta for Company Domains

## ADDED Requirements

### Requirement: Company domain persistence uses reusable queries and exhaustive guards

The system MUST keep company lookup and counting logic in reusable query helpers and MUST use them from all companies repositories. `CountActivePrimaryDomains` MUST ignore the exclude argument for create flows and MUST honor it for update flows. A zero `RowsAffected` result on update or delete MUST map to the matching entity-specific not-found error. Structural tests MUST recursively discover all seven companies `repository.go` files and fail if any repository path leaks raw `.Error`, raw GORM values, or bypasses the entity-specific mapper, including helper-scoped methods and single-result queries.
(Previously: query reuse existed, but flow-aware count behavior and recursive boundary guards were not comprehensive.)

#### Scenario: Create flow ignores exclude

- GIVEN `create_company_domain` checks active primary domains
- WHEN `CountActivePrimaryDomains` is called with a current domain public ID
- THEN the exclude value is ignored
- AND the count reflects all active primary domains

#### Scenario: Update flow honors exclude

- GIVEN `update_company_domain` edits an existing primary domain
- WHEN `CountActivePrimaryDomains` is called with that domain public ID
- THEN the current domain is excluded from the count
- AND the update may proceed when no other active primary exists

#### Scenario: Zero rows map to not-found

- GIVEN a delete or update affects zero rows
- WHEN the repository translates the result
- THEN it returns the matching company or company-domain not-found error
- AND it does not return a generic GORM error

#### Scenario: Guards discover every repository recursively

- GIVEN the companies module repository tree
- WHEN the structural guard runs
- THEN it inspects all seven repository files recursively
- AND it fails on raw GORM, raw `.Error`, or mapper bypass
