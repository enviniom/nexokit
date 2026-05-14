> Lee también `_context.md` antes de implementar este change.

# Change 6: Utilidades API: DTOs, validaciones, paginación, filtros y documentación de convenciones

## Objetivo

Crear las utilidades reutilizables que harán que todos los módulos futuros de NexoKit tengan un comportamiento consistente.

Este change refina y documenta las utilidades base iniciadas en el Change 1. El Change 1 define las estructuras mínimas para arrancar; el Change 6 las completa, agrega filtros, helpers GORM y documenta las convenciones para que cualquier desarrollador pueda crear módulos nuevos de forma consistente.

## Alcance de este change

Implementar:

- DTOs base refinados.
- Reglas de validación adicionales (el validador propio ya existe desde Change 1).
- Manejo de errores de validación por campo.
- Paginación.
- Filtros.
- Ordenamiento.
- Search básico.
- Soft delete conventions.
- Helpers GORM reutilizables.
- Documentación de convenciones.

## DTOs base

Completar y documentar:

```txt
APIResponse
ErrorResponse
ValidationErrorResponse
PaginatedResponse
PaginationMeta
```

## Validación

NexoKit usa el validador propio definido en `internal/platform/validator/` (Change 1).

En el Change 6 se completa con:

- Reglas adicionales según necesidades de los módulos base (`ValidSlug`, `ValidURL`, `InList`, etc.).
- Documentación de cómo agregar reglas nuevas.
- Confirmación de que `RespondIfInvalid` usa `response.ValidationError` (integración definitiva con `platform/response`).

Los errores de validación deben retornarse siempre por campo, nunca como string genérico.

## Paginación

Parámetros estándar:

```txt
page
per_page
sort
order
search
```

Ejemplo:

```txt
GET /api/v1/users?page=1&per_page=15&sort=created_at&order=desc&search=jhon
```

Respuesta:

```json
{
  "success": true,
  "message": "Datos obtenidos correctamente",
  "data": [],
  "meta": {
    "pagination": {
      "page": 1,
      "per_page": 15,
      "total": 100,
      "total_pages": 7
    },
    "filters": {
      "search": "jhon",
      "sort": "created_at",
      "order": "desc"
    }
  },
  "errors": null
}
```

## Filtros

Crear estructura reusable:

```txt
FilterParams
PaginationParams
SortParams
```

Soportar filtros simples:

```txt
status
created_from
created_to
search
```

Cada módulo puede extender filtros propios.

## GORM helpers

Crear helpers para:

```txt
ApplyPagination
ApplySorting
ApplySearch
ApplyDateRange
ApplyStatusFilter
```

## Manejo de errores

Completar errores de aplicación iniciados en Change 1:

```txt
ErrNotFound
ErrUnauthorized
ErrForbidden
ErrValidation
ErrConflict
ErrInternal
```

El handler debe convertir errores a respuestas HTTP consistentes.

## Módulo de referencia

No se crea un módulo `examples` descartable. Los módulos `users` y `companies`, ya implementados en changes anteriores, sirven como referencia oficial de cómo construir un módulo completo con:

- Listado paginado.
- Filtros.
- Crear, editar, eliminar.
- Validaciones.
- Respuesta estándar.
- Tenant scope.

## Documentación

Agregar documentación sobre:

- Cómo crear un módulo nuevo.
- Cómo crear DTOs.
- Cómo validar requests.
- Cómo responder errores.
- Cómo usar paginación.
- Cómo usar filtros.
- Cómo aplicar tenant scope.
- Cómo proteger rutas con permisos.
- Cómo registrar rutas de un módulo nuevo.

## Criterios de aceptación

Este change se considera completo cuando:

1. Existen DTOs base de respuesta completos y documentados.
2. Existen DTOs de paginación.
3. Existe parser de query params para paginación.
4. Existe parser de filtros base.
5. Existe helper GORM para paginación.
6. Existe helper GORM para sorting.
7. Existe helper GORM para search.
8. Los errores se manejan de forma centralizada.
9. Las validaciones retornan errores por campo.
10. Los módulos `users` y `companies` sirven como referencia de endpoint paginado con filtros.
11. Existe documentación de convenciones del framework.
12. Todos los endpoints base usan la respuesta estándar.
