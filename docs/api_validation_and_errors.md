# API validation, responses and errors

Bind, sanitize, validate, execute, then map the result. This guide owns HTTP envelopes, list conventions, module errors and safe error logging.

## 1. Handler outcomes

| Outcome | Response path |
|---|---|
| Malformed JSON / binding failure | `response.BadRequest` → 400 |
| DTO validation failure | `response.RespondIfInvalid` → 422 with field-keyed arrays |
| Expected module error | `response.HandleError` → declared status/public message |
| Unexpected failure / recovered panic | Safe 500; structured error logging |
| Success | `Success` (200), `Created` (201), `NoContent` (204 without body) |

Return immediately after writing a failure. Keep HTTP decisions in handlers and platform response helpers, not services/repositories.

## 2. Envelopes

`platform/response` uses `success`, `message`, `data`, `meta`, `errors`. Unused fields are normally `null`; 204 has no envelope. Health/live and health/ready deliberately use their own probe payloads.

Paginated responses nest metadata under `meta.pagination` (`page`, `per_page`, `total`, `total_pages`) and `meta.filters` (`status`, `created_from`, `created_to`, `sort`, `order`, `search`). Do not assume `total_records` or flat pagination metadata.

Validation `errors` is a map of field names to string arrays. Public entity references should use ULID public IDs for new APIs; see [current DTO contracts](architecture_and_vsa.md#4-models-dtos-and-mapping). Existing IAM numeric user `role_id`/`company_id` and numeric audit fields remain public contract fields, not private entity primary IDs.

## 3. List query inputs

`platform/query.ListFromGin` parses:

| Parameter | Normalization |
|---|---|
| `page` | Default 1; minimum 1 |
| `per_page` | Default 20; maximum 100 |
| `sort` | Column name; default `created_at` |
| `order` | `asc` or `desc`; invalid values become `desc` |
| `search` | Search string |
| `status` | Status filter |
| `created_from`, `created_to` | `YYYY-MM-DD`; invalid dates are ignored |

Use `?sort=name&order=asc`, **not** `sort=name:asc`. Repositories choose supported fields and filters; parsing a parameter does not mean every endpoint implements it. SQL helper safeguards and date-boundary limitations live in [Boundaries and persistence](boundaries_and_cqrs.md#5-reusable-gorm-helpers).

GORM soft-deleted rows are hidden by default when using the appropriate model. There is no universal `include_deleted=true` API; any exceptional exposure needs authorization and tests.

## 4. DTO sanitization and validation

Define request DTOs at the [entity DTO location](architecture_and_vsa.md#4-models-dtos-and-mapping), never in slices. Response DTOs own safe model mapping; handlers instantiate them and call that method for individual and list results.

New request DTOs should implement `(r *Request) Validate() validator.ValidationErrors` so normalization persists before rule evaluation. `response.ValidationErrors` is **not** a type in this checkout. Existing value receivers remain in some DTOs; do not infer that all input is already sanitized.

```go
// In core/<entity>/dtos.go; validator owns the error map type.
func (r *Request) Validate() validator.ValidationErrors {
    r.Name = strings.TrimSpace(r.Name)
    errs := validator.ValidationErrors{}
    validator.Field(errs, "name", r.Name).Required().Apply(validator.MinLength(2))
    return errs
}
```

Sanitize before validating to reject whitespace-only required strings. Use platform normalization helpers for email/domain/slug where applicable. Handlers call `response.RespondIfInvalid(c, req.Validate())` and stop on true.

## 5. Module error identity

Declare reusable application errors in module `core/errors.go` (some existing modules use `error.go`). `platform/apperror.AppError` carries `Code`, `HTTPStatus`, `PublicMessage`, `Internal`. Entity packages must not import parent core for these errors; preserve the [one-way dependency rule](architecture_and_vsa.md#4-models-dtos-and-mapping). The module owns the business code; the platform owns generic sentinels and transport helpers, not a global business-code catalog.

Handlers/services/repositories must use module errors instead of importing `platform/apperror` to build ad-hoc errors. Repositories translate known GORM failures; unexpected technical errors travel outward. Use `%w`, `errors.Is` and `errors.As` so identity and diagnostics survive wrapping. Expected lookup branches can use `(*Entity, bool, error)` rather than treating normal absence as exceptional.

Error codes are available for inspection/logging; the current standard error JSON does **not** expose a `code` field.

## 6. Response safety and logging

`response.HandleError` attaches the error to Gin and maps its status/public message. Internal debug details are exposed only when `DebugErrors` is enabled by `Config.ExposeDebugErrors()` (`local`, `development`, `test`), not by Gin mode. Production-like contexts never include the debug string. Public messages must themselves be safe.

| Outcome | `middleware.ErrorLogger` policy |
|---|---|
| Any response with attached `c.Error(err)` | ERROR per attached error entry |
| Recovered panic | Recovery attaches error; ERROR |
| Response without attached errors, including a manually written 5xx | No ErrorLogger entry |

`HandleError` attaches even expected application errors, so those 4xx responses produce ERROR, not WARN. Error logs include request/actor/tenant identifiers, method/path, latency, `status`, `code`, `public_message` and `internal_chain`. Multiple attached errors yield multiple entries; there is no universal one-line-per-request guarantee. [Request flow](references/runtime/request-flow.md#global-middleware) owns middleware order.

Never put plaintext passwords, tokens, secret keys or sensitive payloads into public messages, error strings or log fields. The middleware does not magically redact arbitrary diagnostics; readiness may return dependency error text, so restrict probe exposure as described in [deployment](deployment.md).

## 7. Review checklist

- [ ] Binding 400 and validation 422 are distinct and short-circuit execution.
- [ ] DTO normalization is tested and survives the validation call.
- [ ] Envelope and nested pagination metadata match source.
- [ ] Known persistence failures become reusable module errors.
- [ ] Debug details and sensitive diagnostics are not exposed in production.
- [ ] Tenant isolation and explicit sorting allowlists are tested.
