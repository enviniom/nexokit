# NexoKit documentation

NexoKit is a reusable Go API base with auth, companies, IAM and onboarding modules. Examples describe guidance unless explicitly identified as current behavior.

## Start here

| Goal | Reading path |
|---|---|
| Adopt the starter | [Starter quick path](starter-template.md) → [Application map](system-map.md) |
| Deploy or recover | [Deployment runbook](deployment.md) → [CLI](references/runtime/cli.md) |
| Add or fix a use case | [Architecture](architecture_and_vsa.md) → [API contracts](api_validation_and_errors.md) → [Testing](testing_and_quality.md) |
| Integrate capabilities | [Boundaries and CQRS](boundaries_and_cqrs.md) → [Multitenancy](references/infrastructure/multitenancy.md) |
| Review a design | [ADR guide](adr_guide_and_template.md) → [Coverage ledger](normative_traceability_matrix.md) |
| Work on security | [Request flow](references/runtime/request-flow.md) → [Tokens](references/security/auth-tokens.md) → [RBAC](references/security/permissions-and-rbac.md) |

## Four daily guides

1. [Architecture and VSA](architecture_and_vsa.md): layout, layer responsibilities, containers, route registration and DTO placement.
2. [Boundaries and CQRS](boundaries_and_cqrs.md): consumer ports, persistence helpers, transaction guidance and local read projections.
3. [API validation and errors](api_validation_and_errors.md): envelopes, list inputs, sanitization, domain errors and safe logging.
4. [Testing and quality](testing_and_quality.md): fakes, database tests, isolation and verification targets.

[Architecture §4](architecture_and_vsa.md#4-models-dtos-and-mapping) is the single owner of entity DTO placement/mapping and the temporary centralized contract convention; it distinguishes current Auth/IAM entity packages from the legacy flat Companies/onboarding layout. These guides own the cross-cutting rules. References link back instead of redefining those rules; present source and future design guidance must not be confused.

## Specialized references

| Area | Reference | Unique detail |
|---|---|---|
| Runtime | [Request flow](references/runtime/request-flow.md) | Middleware order and actual private route scopes |
| Runtime | [CLI](references/runtime/cli.md) | Commands, flags, root creation and seed execution |
| Security | [Auth tokens](references/security/auth-tokens.md) | Claims, refresh lifecycle and current safety limits |
| Security | [Permissions and RBAC](references/security/permissions-and-rbac.md) | Guards and permission identities |
| Security | [Permission sync](references/security/permissions-sync.md) | Startup discovery, persistence and admin assignment |
| Infrastructure | [Multitenancy](references/infrastructure/multitenancy.md) | Context fields, root selection and repository isolation |
| Infrastructure | [Domains](references/infrastructure/domains.md) | Host resolution and optional public middleware |
| Infrastructure | [Cache](references/infrastructure/cache.md) | Redis/no-op behavior and readiness limitations |
| Infrastructure | [Rate limiting](references/infrastructure/rate-limiting.md) | Keys, drivers, fallback and effective quota limitations |

Keep these references separate: their runtime and security details are not daily architectural rules. Significant future decisions belong in `docs/adr/`; the [ADR guide](adr_guide_and_template.md) defines the process without claiming records already exist.
