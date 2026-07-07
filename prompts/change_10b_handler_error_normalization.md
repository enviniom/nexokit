> Lee también `_context.md` y `docs/api-conventions.md` antes de implementar este change.

# Change 10b: Normalización de handlers para filtros, paginación y errores

## Objetivo

Estandarizar los handlers para que sigan las convenciones actuales de API en tres aspectos: filtros, paginación y manejo de errores.

Este change reduce lógica repetida en handlers, evita que cada módulo mantenga su propio `switch apperror.Status(err)` y asegura que todos los listados expongan metadata consistente de paginación y filtros.

## Contexto

`docs/api-conventions.md` define que los handlers deben delegar el mapeo de errores conocidos a:

```go
response.HandleError(c, err)
```

El handler de `users` ya sigue este patrón para errores de servicio. En cambio, otros handlers todavía usan switches locales o helpers propios.

También existen helpers en plataforma para normalizar listados con filtros y paginación:

```go
query.ListFromGin(c)
response.PaginatedWithFilters(c, status, message, data, pagination, filters)
```

Por lo tanto, la normalización de filtros y paginación no debería vivir dentro de handlers de módulos salvo que exista una necesidad específica no cubierta por `platform/query`.

## Alcance de este change

Implementar:

- Reemplazar `switch apperror.Status(err)` en handlers por `response.HandleError`.
- Eliminar wrappers locales de error cuando solo replican `HandleError`.
- Mantener wrappers locales únicamente si convierten errores de dominio en errores de validación por campo.
- Normalizar validaciones con `response.RespondIfInvalid` cuando aplique.
- Normalizar filtros y paginación de listados con `query.ListFromGin` y `response.PaginatedWithFilters`.
- Eliminar normalización local de filtros/paginación en handlers cuando ya esté cubierta por `platform/query`.
- Remover imports de `apperror` en handlers donde ya no sean necesarios.
- Evaluar si helpers directos de `response` como `Conflict`, `Forbidden`, `NotFound`, `InternalServerError`, etc. siguen siendo usados; eliminar solo si quedan realmente sin uso.

## Reglas

- Los errores devueltos por servicios deben manejarse con `response.HandleError(c, err)`.
- Un handler no debe tener `switch apperror.Status(err)`.
- Un helper local como `respondError` solo se permite si agrega valor específico del módulo.
- El caso válido de helper local es convertir un error de dominio a una respuesta de validación por campo, por ejemplo `slug` duplicado.
- `users` debe tomarse como referencia para errores de servicio.
- `companies` puede conservar una función local solo si sigue necesitando mapear `ErrDuplicateSlug` a error de campo.
- Los filtros y paginación de listados deben usar helpers de `platform/query` y `platform/response`, no lógica local dentro del módulo.
- Todo endpoint de listado debe devolver metadata de paginación y filtros de forma consistente.
- No cambiar contratos públicos de respuesta salvo para hacerlos consistentes con `docs/api-conventions.md`.

## Áreas a revisar

```txt
internal/modules/users/handler.go
internal/modules/companies/handler.go
internal/modules/roles/handler.go
internal/modules/permissions/handler.go
internal/modules/auth/handler.go
internal/platform/response/response.go
internal/platform/query/query.go
docs/api-conventions.md
```

## Normalización esperada por módulo

### Users

- Mantener `response.HandleError` para errores de servicio.
- Evaluar migrar validaciones manuales a `response.RespondIfInvalid` para quedar igual que `companies`.
- Mantener `query.ListFromGin` y `response.PaginatedWithFilters`.

### Companies

- Revisar `respondError`.
- Si solo existe para `ErrDuplicateSlug`, puede mantenerse como wrapper mínimo documentado.
- Revisar cualquier normalización local de filtros/paginación.
- Si la normalización local ya existe en `platform/query`, eliminar la duplicación y usar el helper común.
- Si `platform/query` no cubre un caso necesario, mover la capacidad faltante a `platform/query` en vez de dejarla en `companies`.

### Roles

- Reemplazar switches de `apperror.Status` por `response.HandleError`.
- Usar `query.ListFromGin` y `response.PaginatedWithFilters` para exponer filtros y paginación consistentes.
- Remover imports y helpers obsoletos.

### Permissions

- Reemplazar `writePermissionError` por `response.HandleError` si no aporta mapeos específicos.
- Usar `query.ListFromGin` y `response.PaginatedWithFilters` para exponer filtros y paginación consistentes.

### Auth

- Reemplazar `respondError` por `response.HandleError` si solo mapea 401/403/default.

## Preguntas de diseño a responder antes de implementar

1. ¿`response.HandleError` cubre todos los errores usados por los handlers actuales?
2. ¿Algún handler necesita un mapeo de dominio a error de campo que no debería vivir en `HandleError`?
3. ¿`platform/query` ya cubre toda la normalización de filtros y paginación que hoy hacen los handlers?
4. Si `companies` tiene normalización propia, ¿qué falta mover a `platform/query` para eliminarla?
5. ¿Hay helpers de `response` que sigan siendo parte útil del API pública para handlers, aunque algunos módulos usen `HandleError`?
6. ¿Conviene eliminar helpers directos no usados en este change o dejar su limpieza para otro change?

## Criterios de aceptación

Este change se considera completo cuando:

1. No quedan `switch apperror.Status(err)` en handlers.
2. Los errores de servicio usan `response.HandleError`.
3. Los wrappers locales de error se eliminaron salvo excepciones justificadas.
4. Si `companies.respondError` permanece, su razón está clara y cubierta por tests.
5. Los listados usan `query.ListFromGin` y `response.PaginatedWithFilters` cuando corresponde.
6. Todos los listados devuelven metadata consistente de filtros y paginación.
7. No hay lógica de normalización de filtros o paginación duplicada dentro de handlers.
8. Si faltaba soporte común para filtros/paginación, quedó en `platform/query` o `platform/response`, no en un módulo específico.
9. Los imports de `apperror` se eliminan de handlers que ya no los necesitan.
10. Helpers directos de `response` se eliminan solo si no tienen usos restantes ni valor como API pública.
11. `go test ./...` pasa.
12. `go build ./...` pasa.
