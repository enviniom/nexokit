# Testing and quality

Test observable behavior at the narrowest useful boundary. Put new mapper, validation and serialization tests in use-case slices or `tests/integration`, not core (including entity packages), containers or routes. Use the shared helpers for database and HTTP integration setup; existing tests need not be removed solely for their location.

## 1. Test responsibilities

| Test | Protects | Dependency strategy |
|---|---|---|
| Slice `dto_test.go` / `validation_test.go` | DTO validation, applicable normalization, safe mapping and serialization | Entity fixtures; no database required |
| `handler_test.go` | Binding, validation, status/envelope, auth and tenant context | Small repository/service fake; `httptest` and Gin test mode |
| `service_test.go` | Business policy, orchestration and failure propagation | Hand-written port/repository fakes |
| `repository_test.go` | SQL, table names, filtering and persistence-error mapping | Real test database |
| `queries/*_test.go` | Reused queries, positive/empty cases and branching behavior | Real test database |
| `tests/integration/*_test.go` | HTTP flows through real app wiring | `tests/helpers` application/database fixtures |

Use plain `t.Fatalf` for broken setup, `t.Errorf` for continuing assertions and `t.Run` for scenarios. Prefer small fakes to introducing an assertion/mock dependency.

## 2. Database fidelity

`tests/helpers.NewSQLiteDB(t, models...)` opens `:memory:`, auto-migrates supplied models and registers cleanup. Current default integration helpers also use SQLite. `make test-integration` does **not** imply PostgreSQL execution or production Goose migration validation.

SQLite is fast feedback, not proof of PostgreSQL dialect/schema compatibility. Test PostgreSQL-specific queries and migrations explicitly when changing them. Partial records need table-name assertions as specified in [Boundaries and persistence](boundaries_and_cqrs.md#4-local-persistence-records-and-table-names).

## 3. Cases worth protecting

- Handlers: malformed JSON vs invalid fields; no persistence call on validation failure; safe errors.
- DTOs: required field keys, persisted pointer-receiver normalization, invalid public IDs/slugs and length limits; table-driven cases. Test the [DTO-owned mapping rule](architecture_and_vsa.md#4-models-dtos-and-mapping) from slices or integration: public contract fields included; private entity primary IDs, passwords and hashes absent from serialized responses. Preserve IAM's numeric user role/company and audit fields, `omitempty`, and nil/empty list and catalog shapes.
- Responses: verify single and list payloads match the same DTO mapping contract; review that handlers call the mapping method rather than duplicate field construction. Request-to-model tests apply only when that conversion exists.
- Repositories: same-tenant success, cross-tenant miss, soft deletion, root/global vs selected scope, sorting allowlist and count accuracy.
- Reusable queries: test branching logic and meaningful positive/negative results, not a ceremonial test for each file.
- Security: absent/invalid authentication, role/permission denial, token revocation/expiry and route-specific tenant requirements.
- File generation: isolate writes with `t.TempDir()`; never write test artifacts to real project/user paths.

Current IAM mapping/serialization examples live in `users/view_user/dto_test.go`, `roles/view_role/dto_test.go` and `permissions/list_permissions/dto_test.go` under `internal/modules/iam/slices/`; validation examples live in `users/create_user/validation_test.go` and `roles/create_role/validation_test.go`. Existing root `core/errors_test.go` remains; no new core/container/routes tests are required.

Support `testing.Short()` skips for genuinely external or slow integration work. Do not label an in-memory suite as external merely because it lives under `tests/integration`.

## 4. Verification commands

Run focused package tests first; use broader commands when the change warrants them:

| Command | Actual scope |
|---|---|
| `go test ./internal/modules/auth/...` | Example focused module run |
| `make test` | `go test ./...` |
| `make test-unit` | `go test -short` excluding `/tests/integration` |
| `make test-integration` | `go test ./tests/integration/...` |
| `make test-coverage` | Writes `coverage.out`, then prints function coverage |
| `make vet` | `go vet ./...` |
| `make lint` | Vet plus `check-module-errors` |
| `gofmt -l .` | Lists formatting drift without rewriting files |

`check-module-errors` checks ad-hoc `apperror` use in module layer files, GORM leaks in services and legacy `mapServiceError` adapters. It is not a comprehensive architecture or sibling-import checker. Coverage helps identify risk; do not invent a percentage gate that the Makefile does not enforce.

For documentation-only work, verify links, examples against source and `git diff --check`; there is no meaningful behavioral RED/GREEN for prose edits.
