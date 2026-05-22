# Code Context

## Files Retrieved
1. `internal/platform/response/response.go` (lines 13-20, 80-110, 138-175) - defines response DTO envelope and helpers; currently has `Success`, `Created`, `Error`, etc., but no no-body helper.
2. `internal/platform/response/response_test.go` (lines 32-52, 234-245) - covers existing response helper status/body behavior; no 204/no-body test exists.
3. `internal/modules/roles/handler.go` (lines 3-10, 120-128) - `Delete` imports `net/http` only to call `c.AbortWithStatus(http.StatusNoContent)` directly.
4. `internal/modules/roles/handler_test.go` (lines 424-440) - successful role delete already expects HTTP 204.
5. `openspec/specs/roles/spec.md` (lines 44-71) - role API contract explicitly requires HTTP 204 for successful role delete.
6. `openspec/specs/api-response/spec.md` (lines 1-16, 35-43, 51-59) - response spec says standard JSON envelope for API responses and documents `Success(..., nil)` null-body semantics; it does not yet carve out 204 no-content operations.
7. `internal/cli/templates/module/handler.tmpl` (lines 124-128) - generated module delete template already calls `response.NoContent(c)`.
8. `tests/cli/testdata/golden/goldenmod/handler.go` (lines 112-118) - golden generated handler also expects `response.NoContent(c)`.
9. `internal/modules/users/handler.go` (lines 88-95, 97-113) and `internal/modules/users/handler_test.go` (lines 520-555) - user delete currently returns 200 envelope; change-password also returns 200 envelope with nil data.
10. `internal/modules/companies/handler.go` (lines 75-81) and `internal/modules/companies/handler_test.go` (lines 205-221) - company delete currently returns 200 envelope and tests expect 200.
11. `internal/modules/permissions/handler.go` (lines 88-94) and `internal/modules/permissions/handler_test.go` (lines 228-241) - permission delete currently returns 200 envelope and tests expect 200.
12. `internal/modules/auth/handler.go` (lines 63-78) and `internal/modules/auth/handler_test.go` (lines 119-129) - logout returns 200 envelope with nil data; likely not a 204 candidate because it accepts/validates a request and currently has success-message semantics.
13. `internal/platform/messages/messages.go` (lines 3-18) - includes `MsgSuccess` and `MsgDeleted`; 204 helper should not need a message because no response body is legal.

## Key Code

Current response helper shape:

```go
// internal/platform/response/response.go:80-89
func Success[T any](c *gin.Context, message string, data T) {
	c.JSON(http.StatusOK, APIResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    nil,
		Errors:  nil,
	})
}
```

Current role delete direct Gin status:

```go
// internal/modules/roles/handler.go:120-128
func (h *Handler) Delete(c *gin.Context) {
	publicID := c.Param("id")
	if err := h.service.Delete(h.tenantContext(c), publicID); err != nil {
		response.HandleError(c, err)
		return
	}
	c.AbortWithStatus(http.StatusNoContent)
}
```

Existing generated-code expectation:

```go
// internal/cli/templates/module/handler.tmpl:124-128
if err := h.service.Delete(c.Request.Context(), {{- if .Tenant }} tc, {{- end }} publicID); err != nil {
	response.NotFound(c, "{{.Struct}} not found")
	return
}
response.NoContent(c)
```

Recommended helper:

```go
// NoContent returns a 204 No Content response with no JSON envelope/body.
func NoContent(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}
```

Status/body semantics:
- HTTP status: always `204 No Content`.
- Body: empty; do not call `c.JSON`; do not include `success`, `message`, `data`, `meta`, or `errors`.
- Message argument: none. `MsgSuccess`/`MsgDeleted` cannot be represented in a compliant 204 response.
- Abort: use `AbortWithStatus` to preserve current `roles.Delete` behavior exactly. If the team prefers response helpers never abort, `c.Status(http.StatusNoContent)` is a viable variant, but it would not be behavior-identical to the current role handler.

## Architecture

`internal/platform/response` centralizes API rendering. Most handlers use `response.Success[any](c, messages.MsgSuccess, nil)` for successful operations that have no data but still return a standard 200 JSON envelope. `roles.Delete` is the exception: its domain spec requires a true 204 no-body response, so the handler bypasses `platform/response` and uses Gin directly.

There are two distinct semantics in the codebase:

1. **Empty success envelope**: `200 OK` with JSON body and `data: null`.
   - Current examples: `auth.Logout`, `users.Delete`, `users.ChangePassword`, `companies.Delete`, `permissions.Delete`.
   - Covered by `openspec/specs/api-response/spec.md` null semantics.

2. **No content**: `204 No Content` with no body.
   - Current example: `roles.Delete`.
   - Required by `openspec/specs/roles/spec.md`.
   - Already assumed by CLI module generation templates/golden output as `response.NoContent(c)`, even though the helper is absent from `response.go`.

Recommendation: add `response.NoContent(c *gin.Context)` and switch only `roles.Delete` to it initially. Do not automatically migrate every `Success[any](..., nil)` call. Some are not DELETEs (`auth.Logout`, `users.ChangePassword`), and existing tests/contracts for users/companies/permissions currently assert 200.

Affected implementation files if accepted:
- `internal/platform/response/response.go`: add `NoContent` helper near `Created`/`Success`.
- `internal/modules/roles/handler.go`: replace direct `c.AbortWithStatus(http.StatusNoContent)` with `response.NoContent(c)` and remove the now-unused `net/http` import.
- `openspec/specs/api-response/spec.md`: add an explicit exception/requirement for `NoContent` so 204 no-body responses do not conflict with the “standard JSON envelope used by every API response” wording.

Affected tests if accepted:
- `internal/platform/response/response_test.go`: add `TestNoContent` asserting status 204 and `w.Body.Len() == 0`.
- `internal/modules/roles/handler_test.go`: existing success case already asserts 204; optionally add `w.Body.Len() == 0` to lock semantics.
- No required changes to `internal/modules/users/handler_test.go`, `internal/modules/companies/handler_test.go`, or `internal/modules/permissions/handler_test.go` unless the team separately decides to change those APIs to 204.

Compatibility risks:
- **API contract risk if migrating broad deletes**: users/companies/permissions delete tests expect `200 OK` with envelope. Changing them to 204 would break clients that parse `success/message/data` from deletes.
- **Spec wording conflict**: api-response spec currently implies every API response has the JSON envelope, while role spec requires 204. Add a documented 204 exception before broader use.
- **Content-Type risk**: `NoContent` should not set JSON content type because no body is sent. Tests should check body emptiness, not JSON fields.
- **Abort semantics risk**: `AbortWithStatus` stops pending Gin handlers. This matches current `roles.Delete`; however, most response helpers do not abort. Keep helper use limited to terminal handler/middleware paths, or document it as terminal.
- **Generated-code risk**: CLI generated handler templates/golden already call `response.NoContent(c)`. Adding the helper improves consistency and prevents generated modules from depending on a missing symbol. Adjacent note: those same generated files also call `response.OK(c, ...)`, which is also absent in current `response.go`; that is outside this task but may be another generation/runtime mismatch.

## Start Here

Start with `internal/platform/response/response.go` because the missing abstraction belongs there, and the existing CLI templates already expect `response.NoContent(c)`. Then open `internal/modules/roles/handler.go` to replace the one direct Gin 204 response and remove `net/http`.
