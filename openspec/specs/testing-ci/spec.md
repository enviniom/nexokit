# Testing CI Specification

## Purpose

Define automated testing infrastructure: Makefile test targets, GitHub Actions CI workflow, and coverage reporting for NexoKit.

## Requirements

### Requirement: Makefile test targets

The system MUST provide granular Makefile targets: `test-unit`, `test-integration`, `test-coverage`, in addition to the existing `test` target. All MUST appear in `.PHONY`.

| Target | Behavior |
|--------|----------|
| `test` | Runs `go test ./...` (all tests) |
| `test-unit` | Runs `go test ./... -short` (skips integration tests) |
| `test-integration` | Runs `go test ./tests/integration/...` (integration tests only) |
| `test-coverage` | Runs `go test ./... -coverprofile=coverage.out` and prints summary |

#### Scenario: Run unit tests only

- GIVEN the project has unit tests using `testing.Short()` gates
- WHEN `make test-unit` is executed
- THEN only unit tests run (integration tests are skipped)
- AND exit code reflects test results

#### Scenario: Run integration tests

- GIVEN integration tests exist under `tests/integration/`
- WHEN `make test-integration` is executed
- THEN only integration tests run
- AND tests requiring external services skip gracefully if unavailable

#### Scenario: Generate coverage report

- GIVEN the project has testable packages
- WHEN `make test-coverage` is executed
- THEN `coverage.out` is generated in the project root
- AND a coverage summary is printed to stdout

### Requirement: GitHub Actions CI workflow

The system MUST provide a GitHub Actions workflow at `.github/workflows/ci.yml` that runs on `push` to `main` and on all `pull_request` events.

The workflow MUST execute:
1. `go test ./...` — all tests
2. `go vet ./...` — static analysis
3. `go fmt` check — formatting verification

#### Scenario: CI runs on push to main

- GIVEN a commit is pushed to `main`
- WHEN the CI workflow triggers
- THEN all three checks (test, vet, fmt) execute
- AND the workflow fails if any check fails

#### Scenario: CI runs on pull request

- GIVEN a pull request is opened or updated
- WHEN the CI workflow triggers
- THEN all three checks execute against the PR branch
- AND results are visible in the PR checks

#### Scenario: Fmt check detects unformatted files

- GIVEN a Go file fails `gofmt -d`
- WHEN the CI workflow runs the fmt check
- THEN the workflow fails with a message indicating unformatted files
- AND the developer can run `make fmt` to fix

### Requirement: Coverage output format

The system MUST produce a human-readable coverage summary when `make test-coverage` runs. The `coverage.out` file MUST be in Go's standard coverage profile format.

#### Scenario: Coverage summary is readable

- GIVEN `make test-coverage` completes successfully
- WHEN the output is inspected
- THEN each package shows its coverage percentage
- AND a total coverage percentage is displayed

## Constraints

- No new testing dependencies SHALL be introduced (stdlib-only).
- The `test` target MUST remain backward-compatible with existing usage.
- CI workflow MUST use the latest stable Go version matching `go.mod`.
