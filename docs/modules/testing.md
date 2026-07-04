# Module testing

A module's tests are how reviewers verify the rules in the other documents hold. This document covers the test layers, the DTO and error contract tests, the GORM partial model tests, and a practical heuristic for how much to test.

For project-wide test commands, integration setup, and CI reproduction, see [`../testing.md`](../testing.md).

## Quick path

1. Test the slice at three layers: handler, service, repository.
2. Test the DTO validation contract and module errors as part of the module's contract.
3. Test partial GORM model `TableName()` directly.
4. Scale test depth with slice complexity, not line count.

## Test layers

| Layer | What to test | Where the test lives |
|---|---|---|
| Handler | Success and mapped error responses. | `slices/<slice>/handler_test.go` (or `slices/<entity>/<slice>/handler_test.go`). |
| Service | Business rules and repository error propagation. | `slices/<slice>/service_test.go`. |
| Repository | Persistence behavior and error translation. | `slices/<slice>/repository_test.go`. |
| Reusable query | Query behavior, once, in `queries/`. | `queries/<query_name>_test.go`. |
| DTO validation | Field-keyed validation contract. | `core/dto_test.go` or alongside the DTO. |
| Module errors | Each declared error in `core/errors.go` maps to the expected `apperror` kind. | `core/errors_test.go`. |
| Partial GORM model | `TableName()` returns the real migration table name. | `core/<model>_test.go` or `queries/<model>_test.go`. |

## DTO validation tests

| Check | Why |
|---|---|
| Each `Validate()` method has table-driven tests covering the success path and every validation rule. | The DTO contract is part of the API; regressions break the API silently. |
| Validation errors are field-keyed. | The envelope contract is `response.ValidationErrors` keyed by field name. |
| Handlers convert `Validate()` results through `response.RespondIfInvalid`. | The handler should not have its own validation branching. |

## Module error tests

| Check | Why |
|---|---|
| Each declared error in `core/errors.go` maps to the expected `apperror` kind (`NotFound`, `Conflict`, `Forbidden`, ...). | Centralized errors drift if no test pins the kind. |
| The repository maps `gorm.ErrRecordNotFound` to the right `core` error. | The translation is part of the persistence-to-domain contract. |
| Services return the right `core` error and wrap internal errors with `fmt.Errorf("...: %w", err)`. | The error contract must be visible from the service tests. |
| The handler funnels business / app errors through `response.HandleError`. | Status code regressions are caught at the handler layer. |

## Partial GORM model table-name tests

| Check | Why |
|---|---|
| `TableName()` returns the real migration table name. | A wrong `TableName()` is invisible in SQLite `AutoMigrate` tests and breaks in production. |
| Test the function directly, not through `AutoMigrate`. | `AutoMigrate` can create the wrong table name in SQLite and pass. |
| Table names match the real Goose migrations. | Local models and migrations must agree. |

## CRUD-light vs business-heavy heuristic

Use this rule to decide how deep the test pyramid should go for a slice.

| Slice shape | Test focus |
|---|---|
| CRUD-light (a few fields, no policy, no branching). | Repository translation + handler error mapping. Service tests stay small and pin the happy path. |
| Business-heavy (policy, branching, idempotency, multi-step orchestration). | Full handler / service / repository / query coverage, plus a table of edge cases for the business rule. |

| Signal | What it implies |
|---|---|
| The service has no branching, just orchestrates a single repository call. | CRUD-light: focus on the translation and the response shape. |
| The service branches on `bool` from `(*T, bool, error)` or on a typed result. | Business-heavy: cover both branches and the resulting state changes. |
| The slice has an idempotent or "create-or-update" flow. | Business-heavy: cover both paths and the resulting persistence state. |
| The slice is a thin wrapper around a `queries/` file. | CRUD-light: lean on the `queries/` tests; the slice repository tests stay light. |
| The slice exposes a new module error. | Business-heavy: pin the error kind and the handler mapping. |

## Test checklist

- [ ] Handler tests cover success and mapped error responses.
- [ ] Service tests cover business rules and repository error propagation.
- [ ] Repository tests cover persistence behavior and error translation.
- [ ] Reusable query tests cover query behavior once in `queries/`.
- [ ] Repository wrapper tests stay light and point to query tests when full query behavior is already covered.
- [ ] DTO `Validate()` tests cover every rule and confirm field-keyed output.
- [ ] Module error tests pin the `apperror` kind for each declared error.
- [ ] Partial model table-name tests exist when using local GORM models for existing tables.
- [ ] Test depth matches the CRUD-light / business-heavy heuristic above.
- [ ] Tests follow the project-wide patterns in [`../testing.md`](../testing.md) (table-driven, `t.TempDir()`, `testing.Short()` gating for integration, helpers and fixtures, no mandatory third-party assertion framework).
