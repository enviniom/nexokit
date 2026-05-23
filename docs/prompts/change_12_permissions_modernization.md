> Lee también `_context.md` antes de implementar este change.

# Change 12: Modernización de Permisos, Constantes Globales y Refactorización CRUD

## Objetivo

Centralizar y tipar la definición de módulos y acciones mediante constantes globales en un archivo transversal (`internal/platform/permissions/constants.go`), eliminar las referencias hardcodeadas de strings crudos en el registro de rutas mediante una función de formateo compilable, refactorizar la acción base de `index` a `list`, y proveer un endpoint dinámico para poblar las opciones de creación de permisos desde el API.

## Contexto

Actualmente, los permisos del sistema se definen y validan en los routers (`routes.go`) y en las semillas (`seeds/permissions.go`) utilizando literales de string como `"users.index"` o `"roles.view"`. Esto introduce fragilidad ante errores tipográficos accidentales y dificulta la extensión del sistema de permisos de forma estructurada. 

Además, existe una mezcla conceptual entre la acción `"index"` y `"list"`. Estandarizar el CRUD básico en torno a acciones bien definidas (`list, select, view, create, update, delete`) permite mayor consistencia. Al requerir que un permiso sea creado a nivel de API/Base de datos, se debe validar obligatoriamente contra este catálogo estructurado de constantes para mantener el control estricto y evitar la inserción de permisos huérfanos o inválidos.

## Alcance de este change

Implementar:

- **Estructura Transversal de Constantes**: Crear el archivo `internal/platform/permissions/constants.go` con las constantes globales de todos los módulos (`ModuleUsers`, `ModuleRoles`, `ModuleCompanies`, `ModuleSettings`, `ModuleAuth`, `ModulePermissions`) y acciones CRUD predefinidas.
- **Estandarización CRUD (`index` -> `list`)**: Eliminar por completo el uso de `"index"` y reemplazarlo globalmente por `"list"` en todas las rutas, semillas, modelos, payloads y pruebas unitarias/integración de todo el monolito.
- **Registro Seguro y Tipado**: Implementar la función de formato dinámico `permissions.Format(module, action string) string` que ensamble y devuelva de manera segura la cadena del permiso (ej. `"users.list"`).
- **Refactorización de Enrutadores**: Modificar todos los archivos de rutas (`routes.go`) de todos los módulos para utilizar `permissions.Format` con las constantes globales correspondientes en lugar de strings literales crudos.
- **Endpoint de Opciones**: Crear un endpoint público/administrativo `GET /api/v1/permissions/options` que devuelva por separado los catálogos de módulos y acciones definidos en el código.
- **Validación del API al Crear Permisos**: Validar que cualquier creación de permisos a nivel de repositorio/servicio obligue a que el módulo y la acción suministrados formen parte del catálogo de constantes registradas del sistema.

## Reglas

- Las constantes globales de módulos y acciones del sistema deben residir obligatoriamente en `internal/platform/permissions/constants.go` para ser accesibles de manera transversal sin generar acoplamiento circular.
- Queda estrictamente prohibido el uso de strings crudos en los routers para definir permisos; todos deben ser ensamblados a través del formateador compilable `permissions.Format(permissions.ModuleX, permissions.ActionY)`.
- La acción `"index"` desaparece por completo del monolito. La acción de consulta de colecciones se unifica bajo `"list"`.
- El catálogo de acciones CRUD básicas obligatorias queda definido como: `list, select, view, create, update, delete` (además de acciones específicas permitidas como `change_role` o `assign_permissions`).
- Toda creación o inserción de un nuevo permiso en base de datos debe fallar con un error de validación si el módulo o la acción del permiso no coinciden con las constantes registradas en la plataforma.

## Endpoints

```txt
GET /api/v1/permissions/options
```

## DTO esperado

```json
GET /api/v1/permissions/options

Response format (200 OK):
{
  "success": true,
  "message": "Operación exitosa",
  "data": {
    "modules": [
      "users",
      "roles",
      "companies",
      "settings",
      "auth",
      "permissions"
    ],
    "actions": [
      "list",
      "select",
      "view",
      "create",
      "update",
      "delete",
      "change_role",
      "assign_permissions"
    ]
  }
}
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Se crea el archivo `internal/platform/permissions/constants.go` centralizando los módulos y acciones del sistema.
2. Todas las llamadas a `requirePermission` en los archivos `routes.go` de todos los módulos utilizan `permissions.Format(...)` con constantes globales.
3. Se refactorizan todas las ocurrencias y referencias de la acción `"index"` a `"list"` en base de datos, routers, semillas, DTOs y tests, corriendo de forma exitosa.
4. El endpoint `GET /api/v1/permissions/options` devuelve la estructura esperada listando de forma separada los módulos y acciones definidos globalmente.
5. Se incluye lógica de validación de negocio en el servicio de creación de permisos para rechazar cualquier módulo o acción que no pertenezca al catálogo tipado.
6. Toda la suite de pruebas del proyecto (`go test ./...`) compila y pasa exitosamente (`GREEN`) con las nuevas firmas e integraciones.
