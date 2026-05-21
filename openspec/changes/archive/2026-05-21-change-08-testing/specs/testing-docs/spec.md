# Testing Documentation Specification

## Purpose

Define the developer guide for testing in NexoKit (`docs/testing.md`), providing conventions, patterns, and guidelines that future developers can follow to write, run, and maintain tests.

## Requirements

### Requirement: Testing developer guide

The system MUST provide `docs/testing.md` as a comprehensive testing guide covering: test structure, running tests, writing new tests, testing conventions, and troubleshooting.

#### Scenario: New developer understands test structure

- GIVEN a developer reads `docs/testing.md`
- WHEN they look for where to place tests
- THEN the document explains unit tests live alongside code (`*_test.go`)
- AND integration tests live under `tests/integration/`

#### Scenario: Developer knows how to run tests

- GIVEN a developer wants to run tests
- WHEN they consult the guide
- THEN the document lists all `make test-*` targets with descriptions
- AND explains when to use each target

### Requirement: Test pattern guidelines

The guide MUST document Go testing patterns used in NexoKit: table-driven tests, subtests with `t.Run`, `t.TempDir()` for filesystem tests, `testing.Short()` for skip gates, and the decision between same-package vs `*_test` package.

#### Scenario: Developer writes a table-driven test

- GIVEN a developer needs to test a function with multiple cases
- WHEN they follow the guide's pattern
- THEN they write a table-driven test with `t.Run` subtests
- AND the test covers happy path and edge cases

### Requirement: Testing conventions

The guide MUST document mandatory conventions: descriptive test names, explicit error checking, `t.Fatal` for setup errors, `t.Error` for assertion failures, small interfaces for mocking, and stdlib-first testing approach.

#### Scenario: Reviewer verifies test follows conventions

- GIVEN a PR contains new test code
- WHEN a reviewer checks against `docs/testing.md`
- THEN they can verify each convention requirement
- AND deviations are identifiable and discussable

### Requirement: Integration test guidelines

The guide MUST explain how to write integration tests: using test helpers, seeding fixtures, database cleanup with `t.Cleanup()`, and when to skip with `testing.Short()`.

#### Scenario: Developer adds a new integration test

- GIVEN a developer needs to test an endpoint end-to-end
- WHEN they follow the integration test guidelines
- THEN they use the provided helpers for database and auth setup
- AND the test cleans up after itself

### Requirement: Makefile and CI documentation

The guide MUST document each Makefile test target, the GitHub Actions CI workflow, and what each CI check validates.

#### Scenario: Developer understands CI failures

- GIVEN a CI check fails on a PR
- WHEN the developer consults the guide
- THEN they can identify which check failed (test, vet, or fmt)
- AND know the local command to reproduce and fix it

## Constraints

- The guide MUST be written in a scannable format (headings, tables, code examples).
- Code examples MUST use the same patterns already established in the codebase.
- The guide SHALL NOT prescribe testify or any third-party testing framework as mandatory.
