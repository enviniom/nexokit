# Documentation coverage ledger

Review replacement coverage here without navigating removed documents. This is an ownership ledger, not a quantified audit or proof that every code path conforms to the guidance.

## Core ownership

| Concern | Owning section |
|---|---|
| Slice layout and current directory variations | [Architecture §1](architecture_and_vsa.md#1-layout-and-current-variations) |
| Layer responsibilities and pragmatic layering | [Architecture §2](architecture_and_vsa.md#2-choose-the-smallest-useful-layering) |
| Containers vs route mounting | [Architecture §3](architecture_and_vsa.md#3-containers-and-routes) |
| Entity DTO ownership, safe mapping, public identifiers and centralized module contracts | [Architecture §4](architecture_and_vsa.md#4-models-dtos-and-mapping) |
| Sibling imports and consumer ports | [Boundaries §§1–2](boundaries_and_cqrs.md#1-module-boundaries) |
| Transactions and partial records | [Boundaries §§3–4](boundaries_and_cqrs.md#3-transaction-boundaries) |
| Safe persistence helpers and soft-delete exceptions | [Boundaries §5](boundaries_and_cqrs.md#5-reusable-gorm-helpers) |
| Read projections as future guidance | [Boundaries §6](boundaries_and_cqrs.md#6-cqrs-read-projections-design-guidance) |
| HTTP outcomes and envelopes | [API §§1–2](api_validation_and_errors.md#1-handler-outcomes) |
| List inputs, DTO validation and module errors | [API §§3–5](api_validation_and_errors.md#3-list-query-inputs) |
| Debug exposure, error logging and credential safety | [API §6](api_validation_and_errors.md#6-response-safety-and-logging) |
| Fakes, database fidelity, isolation and quality commands | [Testing](testing_and_quality.md) |

## Replacement coverage beyond architecture

| Need | Owner |
|---|---|
| Module rename, customization and first run | [Starter adoption](starter-template.md) |
| Entrypoints, bootstrap, platform and mounted modules | [Application map](system-map.md) |
| Build, secrets, migrations, TLS, supervision, probes and rollback | [Deployment](deployment.md) |
| Complete operations command/flag inventory | [CLI](references/runtime/cli.md) |
| Middleware and private route scope | [Request flow](references/runtime/request-flow.md) |
| Access/refresh lifecycle and known limits | [Tokens](references/security/auth-tokens.md) |
| Guard identity and startup sync | [RBAC](references/security/permissions-and-rbac.md), [sync](references/security/permissions-sync.md) |
| Tenant isolation and host/alias resolution | [Multitenancy](references/infrastructure/multitenancy.md), [domains](references/infrastructure/domains.md) |
| Optional cache and rate-limit drivers/fallback | [Cache](references/infrastructure/cache.md), [rate limiting](references/infrastructure/rate-limiting.md) |
| Future decision process | [ADR guide](adr_guide_and_template.md) |

## Corrections to inherited assumptions

Auth consumers import entity models and DTOs directly from `core/user` and `core/refresh_token`; no flat compatibility aliases remain. Mapping and validation regression tests live with those entities.

Specialized references remain because they own unique operational/security contracts. Removed downstream-only material is not retained as a historical runtime specification. ADR records are future artifacts, not evidence claimed by this ledger.
