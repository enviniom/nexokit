# Architecture and vertical slices

Use one package per use case, with explicit dependencies. This guide owns layout, wiring and layer responsibilities; [the application map](system-map.md) describes what is currently composed.

## 1. Layout and current variations

Use this target shape for new modules:

```text
internal/modules/<name>/
├── container.go       # dependency construction
├── routes.go          # HTTP mounting and guards
├── core/
│   ├── contracts.go   # module-wide consumer contract declarations (for now)
│   ├── errors.go      # shared module errors (existing names may differ)
│   └── <entity>/
│       ├── model.go   # entity model
│       └── dtos.go    # all entity request/response DTOs and their mapping
├── queries/           # reusable persistence operations
└── slices/
    └── <use_case>/    # handler, repository, optional service and tests
```

Do not put new slice implementation files in the module root or import that root from a slice. Extract reused persistence queries into `queries/`, preferably one operation per file; non-persistence helpers, DTO mapping and business policy belong elsewhere.


Auth uses `core/user` and `core/refresh_token` directly, without flat core model/DTO aliases or slice aliases. `SessionView` belongs to the user entity; the core root retains module contracts and errors. IAM uses `core/user`, `core/role` and `core/permission`, each with `model.go` and `dtos.go`, without flat compatibility aliases. Its core root retains module contracts, errors and constants. Companies uses `core/company` and `core/company_domain` for models and all DTOs, with DTO-owned response mapping and root core errors/constants. Domain associations use a minimal company_domain-owned companies-table record to avoid an entity import cycle. Onboarding owns local bootstrap records in `core/{company,company_domain,user,role,permission}` and company workflow DTOs in `core/company/dtos.go`; its use case lives in `slices/onboard_company`, with contracts/errors/constants retained in parent core. The response is an orchestration result, not a persistence-model projection.

IAM groups use cases under `slices/users`, `slices/roles` and `slices/permissions`, with non-HTTP capabilities under `slices/internal`. Entity containers and routes stay with their groups. `slices/capabilities.go` assembles internal services and exposes root `core` contracts to the module container: Go permits imports of `slices/internal` only from within `slices`, not from the IAM root.

## 2. Choose the smallest useful layering

| Layer | Owns | Must not own |
|---|---|---|
| Handler | Bind/validate, authenticated and tenant context, call DTO mapping methods and write responses | SQL/GORM calls or business decisions |
| Optional service | Business policy, orchestration, multiple repositories or external capabilities | Gin, HTTP statuses or GORM |
| Repository | Persistence, query reuse, translating database errors into module errors | HTTP responses or use-case policy |

Start simple CRUD with handler + repository. Add a service when there is genuine orchestration or business policy; avoid pass-through services. Existing modules often have three layers—this preference does not require removing them.

## 3. Containers and routes

`container.go` constructs repositories, services and handlers; `routes.go` mounts those handlers and declares guards. Neither owns business policy. Entity grouping is useful when it improves navigation, not a mandatory extra directory for every use case.

Module `Register` functions receive prepared containers and guard factories from `internal/app/container.go`. Signatures differ by module; copy the actual nearby route shape rather than inventing a universal signature. Keep cross-module composition in the app container; see [boundaries](boundaries_and_cqrs.md#1-module-boundaries).

## 4. Models, DTOs and mapping

**All module request/response DTOs always belong in `core/<entity>/dtos.go`, alongside `core/<entity>/model.go`, including DTOs used by only one slice. Never define DTOs in slices.** Models do not depend on DTOs; model-to-response mapping is implemented by the response DTO in that same `dtos.go`, not by handlers or `queries/`.

Illustrative target code (not an existing Company API):

```go
// core/company/dtos.go (package company); Company is in model.go.
type CompanyResponse struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func (r *CompanyResponse) FromModel(model Company) {
    r.ID = model.PublicID // public string identifier, not the internal primary key
    r.Name = model.Name
}

// In a handler after obtaining model:
var dto company.CompanyResponse
dto.FromModel(model)
response.Success(c, dto)

// For a list, reuse the same method; do not reconstruct fields in the handler.
items := make([]company.CompanyResponse, len(models))
for i, model := range models {
    items[i].FromModel(model)
}
```

Adapt the illustrative model field types and response helper arguments to the actual API. Mapping exposes safe public contract fields, never private entity primary IDs, passwords or hashes. Existing IAM contracts retain numeric user `role_id`/`company_id` and numeric `created_by`/`updated_by` audit fields; these are not private primary IDs. Role responses use a public string company ID. DTOs perform no database calls or business orchestration. Request validation/sanitization uses pointer receivers; request-to-model conversion may live in `dtos.go` when applicable, but requests need not map from persistence models and bidirectional mapping is not required. [API validation and errors](api_validation_and_errors.md#4-dto-sanitization-and-validation) owns validation details.

IAM response mapping lives in the corresponding entity `dtos.go`, including role catalog/assignment projections in `core/role/dtos.go`. Callers retain permission selection and sorting; DTO mapping preserves existing optional-field and collection shapes.

For now, declare contracts centrally in module `core/contracts.go`, not entity or slice packages. Keep signatures small and consumer-defined, with constructor injection. Parent `core` contracts may import entity types; entity packages must not import parent `core`, so dependencies stay one-way. If an entity DTO needs a shared error, use an appropriate platform error type or return validation results for the caller to translate into module errors; do not introduce a parent import cycle.

## 5. Tenant-owned use cases

Resolve a trusted context before tenant-owned operations; pass it explicitly to repositories, which enforce `tenant.ApplyTenantScope`. Missing context is not permission to query globally. Global operations require intentional route policy. Generated tenant slices reject absent/zero or inconsistent contexts at both handler and repository boundaries. Only lists accept explicit trusted root-global context; create/get/update/delete require a selected company. Create assigns company ownership from that context, never JSON; update changes only editable fields, never company ownership. Generated entity models own a `BeforeCreate` hook calling `platform/identity.Generate` because shared `BaseModel` has no hook; explicitly supplied IDs must be valid ULIDs. The exact root/header rules and isolation checklist live in [multitenancy](references/infrastructure/multitenancy.md), not in each slice guide.

## 6. Review checklist

- [ ] One discoverable package per use case, with no slice-to-module-root cycle.
- [ ] Layer count justified; handlers/services do not query GORM.
- [ ] Containers construct dependencies; routes mount prepared handlers and guards.
- [ ] Persistence failures map to reusable module errors; queries contain persistence only.
- [ ] All DTOs live in entity `dtos.go` beside `model.go`; handlers call DTO-owned mapping, including lists.
- [ ] Contracts remain in module `core/contracts.go`; entity packages never import parent core.
- [ ] DTOs expose safe public fields only, with no database calls or orchestration.
- [ ] Tenant policy and public identifiers are explicit.
- [ ] Tests protect the contract using [Testing and quality](testing_and_quality.md).
