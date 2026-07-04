> Lee también `_context.md` antes de implementar este change.

---
### ROL
Senior Architect & Go Specialist, foco en diseño de errores, observabilidad y consistencia HTTP.

### CONTEXTO DE ENTRADA

### OBJETIVO DEL CAMBIO
Rediseñar los errores de aplicación de la plataforma como **wrappers de error de Go** con la siguiente forma canónica:


```go
type Code string
Platform define infraestructura:
type AppError struct {
    Code          Code
    HTTPStatus    int
    PublicMessage string
    Internal      error
}
```
Y los módulos definen sus propios codes:

```go
const CodeInsufficientStock apperror.Code = "sales.insufficient_stock"
```

o si realmente pertenece a catálogo:

```go
const CodeInsufficientStock apperror.Code = "catalog.insufficient_stock"
```

La plataforma puede mantener helpers por HTTP:

```go
apperror.Conflict(code, publicMsg, internal)
```

pero NO debería centralizar códigos de negocio.

Y exponer helpers ergonómicos:

```go
apperror.NotFound(publicMsg string, internal error) *AppError
apperror.BadRequest(publicMsg string, internal error) *AppError
apperror.Forbidden(publicMsg string, internal error) *AppError
apperror.Conflict(publicMsg string, internal error) *AppError
apperror.Internal(publicMsg string, internal error) *AppError
```

Reglas del rediseño:

- `apperror.Status(err)` y `apperror.PublicMessage(err)` se reescriben para leer `Code` / `HTTPStatus` / `PublicMessage` desde `AppError`.
- `response.HandleError(c, err)` consume el `HTTPStatus` y `PublicMessage` para producir la respuesta unificada actual (`success`, `message`, `data`, `meta`, `errors`). No reemplazar el envelope de la API por otro shape.
- El middleware de errores (Gin) registra el `Internal` con `slog` (nivel `error`) incluyendo contexto de request: `request_id`, `method`, `path`, `tenant_id`, `actor_id`, `latency_ms`, y la traza completa (`Internal` y todos sus `Unwrap()`).
- En **producción**, la respuesta al cliente sólo expone `PublicMessage` (nunca el `Internal`).
- En **desarrollo**, se puede añadir un campo `debug` con el `Internal.Error()` para no perder señal.

Este change **define infraestructura**. No migra módulos; eso es Change 18.

### INSTRUCCIONES PASO A PASO

#### 1. Rediseño del paquete `platform/apperror`
- Reescribir `api/internal/platform/apperror/apperror.go` con la struct `AppError` indicada y los helpers `NotFound`, `BadRequest`, `Forbidden`, `Conflict`, `Internal`.
- Mantener compatibilidad con `errors.Is` y `errors.As` para los sentinels actuales (`ErrNotFound`, `ErrForbidden`, `ErrConflict`, `ErrInsufficientStock`, etc.) — preferentemente modelándolos como `AppError{Code: "NOT_FOUND", ...}` y verificando por `Code` además de por puntero.
- Añadir tests unitarios:
  - `errors.Is` entre un `AppError` envuelto y un sentinel sigue funcionando.
  - `errors.As` recupera el `*AppError` desde un `fmt.Errorf("...: %w", err)`.
  - `Status` mapea correctamente cada `Code`.
  - `PublicMessage` oculta `Internal` en producción y lo expone en dev.

#### 2. `platform/response`
- Reescribir `response.HandleError` para que:
  - Si el error es `*AppError`, use `HTTPStatus` y `PublicMessage`.
  - Si es un error genérico, devuelva 500 con mensaje genérico (nunca el texto crudo).
  - Mantenga intacto el manejo actual de validación: los DTOs devuelven `response.ValidationErrors` y los handlers responden con `response.ValidationError(c, errs)`.
- Mantener el envelope JSON unificado existente:

  ```json
  { "success": false, "message": "Resource not found", "data": null, "meta": null, "errors": null }
  ```

#### 3. Middleware de errores y logging
- En `api/internal/app/middleware` (o donde estén los middlewares globales), añadir `ErrorLogger` que:
  - Envolva `c.Next()`.
  - Recoja errores escritos con `c.Error(err)` o panic recuperado.
  - Si el error es `*AppError`, loguea `Internal` (si existe) con `level=error` y contexto de request; si no, lo trata como 500.
  - Añada un `request_id` por request (middleware separado o aquí mismo) y lo propague al `slog` default y al header `X-Request-Id`.
- Configurar el logger (`slog`) con handler JSON en producción y texto en dev.

#### 4. Documentación
- Actualizar [docs/modules/validation-and-errors.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/validation-and-errors.md) con la nueva tabla "Error de `core` → `Code` → `HTTPStatus` → ejemplo de uso" y la regla de mapping de envelope. Reflejar cualquier cambio transversal en [docs/modules.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules.md) si afecta al "Read this when..." / "Core rules at a glance".
- Crear `docs/error-handling.md` con:
  - Shape del `AppError`.
  - Tabla `Code → HTTPStatus` canónica.
  - Reglas de `PublicMessage` vs `Internal`.
  - Ejemplos de uso desde `core/errors.go`, service, repository y handler.
  - Política de logging (qué se loguea, qué se expone, qué queda en `errors`).

#### 5. Compatibilidad y migración
- Marcar el paquete con un `CHANGELOG` interno (comentario al inicio del archivo) que indique que los módulos aún no se migraron; eso se hará en Change 18.
- Garantizar que `apperror.Status` y `apperror.PublicMessage` siguen funcionando con código existente (los handlers actuales que hacen `if errors.Is(err, apperror.ErrNotFound) { ... }` no deben romperse).

### CRITERIOS DE VERIFICACIÓN
- Test unitario: `apperror.NotFound("foo", errors.New("db boom")).Internal.Error() == "db boom"` y `Status == 404` y `PublicMessage == "foo"`.
- Test del middleware: un handler que devuelve `apperror.Internal("boom", errors.New("kaboom"))` produce status 500 con el envelope unificado y el log contiene `"kaboom"` + `request_id`.
- Test de validación: un handler que recibe `response.ValidationErrors{"email": []string{"invalid"}}` responde con 422 y `errors.email[0] == "invalid"` mediante `response.ValidationError`.
- En producción, el campo `debug` (si se incluye) está ausente y `message` no contiene texto del `Internal`.

### ENTREGABLES DEL CAMBIO
- Conventional commit sugerido: `feat(platform): redesign AppError with code/status/public/internal, helpers, error middleware and request logging`.
- Archivos a tocar: `api/internal/platform/apperror/**`, `api/internal/platform/response/**`, `api/internal/app/middleware/**`, [docs/modules/validation-and-errors.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/validation-and-errors.md) (actualizar tabla de mapping), [docs/modules.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules.md) (si aplica), `docs/error-handling.md` (nuevo).
- Outputs SDD: `explore.md`, `proposal.md`, `specs.md` (con escenarios de mapeo por code), `design.md`, `tasks.md`, y reportes de `apply`/`verify`.

### NO-OBJETIVOS
- No migrar módulos al nuevo patrón. Eso es Change 18.
- No tocar mensajes de negocio (i18n) más allá de los `PublicMessage` que ya existen en `platform/messages`.
- No introducir un sistema de códigos versionados por API; los `Code` son internos y estables, no contractuales.
