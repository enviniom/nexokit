# Module Error Conventions Specification

## Purpose

Define a canonical documentation surface for module-level error conventions. The doc lists each module's `core.Err*` sentinels alongside their `Code`, `HTTPStatus`, and a usage example. The doc is cross-linked from `docs/modules/validation-and-errors.md` so reviewers can find the module's error vocabulary in one place.

## Requirements

### Requirement: Canonical module-error-conventions doc

The project MUST publish `docs/module-error-conventions.md`. The doc MUST contain, for every production module (`auth`, `iam`, `companies`, `onboarding`), a table of `core.Err*` → `Code` (`code:<snake_case>` format) → `HTTPStatus` → `PublicMessage` → one example call site. The doc MUST be updated whenever a module adds, removes, or renames a sentinel.
(Previously: no single canonical doc listed every module's error vocabulary; the contract was implicit in `core/errors.go` and scattered across the per-module README and the validation-and-errors guide.)

#### Scenario: Doc lists every module's sentinels

- GIVEN a production module adds a new sentinel to its `core/errors.go`
- WHEN the change is merged
- THEN `docs/module-error-conventions.md` is updated in the same change
- AND the new sentinel appears in the module's table with the correct `Code`, `HTTPStatus`, and `PublicMessage`

#### Scenario: Doc is cross-linked from validation-and-errors

- GIVEN `docs/modules/validation-and-errors.md` describes the `core/errors.go` pattern
- WHEN the cross-link is in place
- THEN the validation-and-errors doc contains a link to `docs/module-error-conventions.md`
- AND a reviewer reading the validation-and-errors doc can navigate to the full per-module error table in one click

### Requirement: Conventions for module-owned errors

The doc MUST encode the conventions the platform layer relies on:

| Convention | Rule |
|---|---|
| Code format | `code:<snake_case>` (e.g. `code:user_not_found`). |
| HTTP status | Set by the `apperror` helper used in the sentinel declaration; the doc MUST record the status the helper produces. |
| PublicMessage | Human-readable, lower-case, no trailing punctuation. Stable across versions because it is the client-visible text. |
| Reuse vs. ad-hoc | Reusable sentinels live in `core/errors.go`; slice-scoped ad-hoc errors stay in the slice. |
| Wrapping | Internal errors are wrapped with `fmt.Errorf("...: %w", err)`; the wrapping error inherits the sentinel's `Code` and `HTTPStatus` via `apperror.Wrap` or the `AppError.Is` chain. |
| Test coverage | Every sentinel in the table MUST be covered by `core/errors_test.go`; the test pin is `apperror.Status`, `Code`, and `PublicMessage`. |

#### Scenario: Conventions block lists every rule

- GIVEN `docs/module-error-conventions.md` is published
- WHEN a reviewer opens the "Conventions" section
- THEN the section lists at least the six rows in the table above
- AND every row has a one-line rationale

#### Scenario: Conventions are enforced by the module tests

- GIVEN a sentinel is added to `core/errors.go` without an entry in `docs/module-error-conventions.md`
- WHEN the change is reviewed
- THEN the review checklist flags the missing doc entry
- AND the change is not merged until the doc is updated
