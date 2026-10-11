# Boundaries, persistence and CQRS guidance

Keep modules independently understandable. This guide owns dependency direction and persistence helper policy; [Architecture](architecture_and_vsa.md) owns layer responsibilities.

## 1. Module boundaries

**New direct imports between sibling `internal/modules/<A>` and `<B>` are forbidden.** A production-source import scan of this base found no sibling-module imports; do not carry downstream exception lists into NexoKit.

Consumers define small capability signatures; for now, declare these interfaces centrally in their module’s `core/contracts.go`, not in entity or slice packages, following [Architecture](architecture_and_vsa.md#4-models-dtos-and-mapping). Providers satisfy them; `internal/app/container.go` injects implementations. Shared boundary values belong in a focused `internal/platform/*` package or `internal/shared`, not another module's internal model. Composition-root imports of modules are intentional and are not sibling-module violations.

Current examples: auth receives a password verifier and token manager; onboarding receives a password hasher. The app adapts IAM's `AuthUserResolver` into auth middleware, returning `platform/authctx.User`. Middleware need not import IAM to use that capability.

## 2. Consumer ports

In module `core/contracts.go`, declare only the methods the consumer needs; accept interfaces in constructors and fake them in tests. Avoid exporting a provider's entire container as a dependency. Keep infrastructure construction at startup and module composition roots.

```go
// Illustrative declaration in module core/contracts.go (package core).
type PasswordHasher interface {
    HashPassword(password string) (string, error)
}
```

## 3. Transaction boundaries

A repository owns database transactions and must check begin, operation and commit failures; rollback on failure. Pass an explicit capability when multiple persistence operations must share atomicity. Do not pass raw database handles through handlers or business services.

This checkout has **no `gormutil.WithTx`/`GetDB` context transaction helper**. Do not call those imagined APIs or assume injected services automatically share a transaction. Introduce and test a transaction contract before relying on cross-capability atomicity.

## 4. Local persistence records and table names

Auth and onboarding use local records to access existing schema without importing another module. A partial record whose name differs from the migration table must implement `TableName()` and have a direct assertion against the migration-defined table name. SQLite `AutoMigrate` can otherwise create a fictitious table and hide a PostgreSQL failure.

Keep projections minimal. A local record is not permission to mutate another capability's tables from arbitrary new modules; coordinate business writes through the owning capability. Existing auth/onboarding persistence responsibilities are deliberate bootstrap workflows, not a blanket exception.

## 5. Reusable GORM helpers

`internal/platform/gormutil/gormutil.go` provides pagination, sorting, search, status and date filters:

| Helper | Actual behavior / policy |
|---|---|
| `ApplyPagination` | One-based offset/limit; normalize and cap inputs with `platform/query` first |
| `ApplySorting` | Validates column syntax and supplied allowlist; **always supply an explicit allowlist** (empty list permits any syntactically valid column) |
| `ApplySearch` | Parameterized `%query%` with `LIKE`, not a guaranteed PostgreSQL case-insensitive `ILIKE` |
| `ApplyDateRange` | `>=`/`<=` on supplied timestamps; date-only upper bounds are midnight, not end-of-day |
| `ApplyStatusFilter` | Parameterized equality on a safe column |

Shared helpers must not bypass soft deletes with `Unscoped()`. Hard deletes or unscoped reads need an explicit, tested use-case justification. Raw SQL must explicitly carry tenant and deletion constraints; GORM model scoping does not apply automatically to raw statements.

## 6. CQRS read projections (design guidance)

For future read aggregation, query into module-owned entity read DTOs at the [canonical DTO location](architecture_and_vsa.md#4-models-dtos-and-mapping), rather than import foreign GORM models with large relation graphs. A read projection may join existing tables but must enforce authorization, tenant scope and soft-delete rules explicitly. Mutations remain with the owning capability.

This is a reusable pattern, **not a claim that this base includes a dedicated reporting or BFF module**. Introduce a projection only for a concrete use case and test the SQL against its real schema.

## 7. Review checklist

- [ ] No new sibling-module imports or leaked module-internal types.
- [ ] Consumer defines the small port in module `core/contracts.go`; app container injects its implementation.
- [ ] Transaction atomicity is explicit and tested, not assumed from context.
- [ ] Partial records have correct `TableName()` and tests.
- [ ] Sorting uses an allowlist; raw reads include tenant/deletion predicates.
- [ ] Read projections do not become an alternate write path.
