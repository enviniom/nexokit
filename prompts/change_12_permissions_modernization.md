> Lee también `_context.md` antes de implementar este change.

# Change 12: Modernización de Permisos, Constantes Globales y Refactorización CRUD

## Objetivo

Centralizar y tipar la definición de módulos y acciones mediante constantes globales en un archivo transversal (`internal/platform/permissions/constants.go`), eliminar las referencias hardcodeadas de strings crudos en el registro de rutas mediante una función de formateo compilable, refactorizar la acción base de `index` a `list`, e implementar un sistema de **Autodescubrimiento y Sincronización Automática (Bootstrapping)** al arrancar el servidor para poblar los permisos en la base de datos de manera transparente. Asimismo, ajustar el API de permisos para restringir su edición únicamente a campos descriptivos y desactivar la creación/eliminación manual externa, permitiendo al frontend renderizar de forma dinámica y limpia la matriz de permisos a partir del listado real.

## Contexto

Actualmente, los permisos del sistema se definen y validan en los routers (`routes.go`) y en las semillas (`seeds/permissions.go`) utilizando literales de string como `"users.index"` o `"roles.view"`. Esto introduce fragilidad ante errores tipográficos accidentales y dificulta la extensión del sistema de permisos de forma estructurada. 

Además, existe una mezcla conceptual entre la acción `"index"` y `"list"`. Estandarizar el CRUD básico en torno a acciones bien definidas (`list, select, view, create, update, delete`) permite mayor consistencia. 

Para administrar eficientemente los permisos por roles desde el frontend sin crear inconsistencias, el backend debe ser la **Única Fuente de Verdad** de los permisos reales. Al iniciar el servidor web, el sistema debe registrar automáticamente todos los permisos requeridos por los handlers de forma dinámica y sincronizarlos con la base de datos. De esta forma, el listado de permisos (`GET /api/v1/permissions`) siempre expondrá la lista exacta de permisos soportados por el binario ejecutándose, haciendo innecesaria la creación de semillas manuales.

El frontend consumirá este listado y podrá agrupar los permisos dinámicamente por el campo `module` de cada objeto de permiso, renderizando en pantalla únicamente las acciones que de verdad existen y están soportadas para cada módulo en particular. Esto evita renderizar cuadrículas cartesianas "vacías" o checkboxes fantasma (como mostrar la acción `change_role` bajo el módulo de compañías).

Debido a que el código gobierna estructuralmente qué permisos existen (pues cada permiso requiere un handler que lo consuma), **permitir la creación o eliminación manual de permisos desde el API es obsoleto y propenso a errores**. Por lo tanto, el API de permisos se cerrará para creación/eliminación, y su edición (`PUT /permissions/:id`) se limitará estrictamente a los campos descriptivos y de visualización (`Name`, `Description`), protegiendo la integridad del sistema.

## Alcance de este change

Implementar:

- **Estructura Transversal de Constantes**: Crear el archivo `internal/platform/permissions/constants.go` con las constantes globales de todos los módulos (`ModuleUsers`, `ModuleRoles`, `ModuleCompanies`, `ModuleSettings`, `ModuleAuth`, `ModulePermissions`) y acciones CRUD predefinidas.
- **Estandarización CRUD (`index` -> `list`)**: Eliminar por completo el uso de `"index"` y reemplazarlo globalmente por `"list"` en todas las rutas, semillas, modelos, payloads y pruebas unitarias/integración de todo el monolito.
- **Registro Seguro y Tipado**: Implementar la función de formato dinámico `permissions.Format(module, action string) string` que ensamble y devuelva de manera segura la cadena del permiso (ej. `"users.list"`).
- **Refactorización de Enrutadores**: Modificar todos los archivos de rutas (`routes.go`) de todos los módulos para utilizar `permissions.Format` con las constantes globales correspondientes en lugar de strings literales crudos.
- **Autodescubrimiento en Memoria (Self-Discovery)**: Implementar en `internal/platform/permissions` un registro seguro en memoria (`permissions.Register(slug)` y `permissions.ListRegistered()`). Modificar el middleware `RequirePermission(slug)` para que registre automáticamente el permiso utilizado durante la carga de las rutas en el arranque.
- **Sincronización Automática (Bootstrapping)**: Implementar una rutina `SyncPermissions(db *gorm.DB) error` que se ejecute al arrancar la API. Esta leerá los permisos acumulados en memoria e insertará/actualizará de manera idempotente (Upsert) cada permiso en la base de datos (extrayendo módulo y acción para poblar los campos).
- **Restricción del CRUD del API de Permisos**:
  - Desactivar o eliminar por completo los endpoints de creación (`POST /api/v1/permissions`) y de eliminación (`DELETE /api/v1/permissions/:id`) del API, delegando la gestión de su existencia al autodescubrimiento.
  - Modificar el handler y el servicio de actualización (`PUT /api/v1/permissions/:id`) para que **únicamente** permita modificar el nombre legible (`Name`) y la descripción (`Description`). Debe rechazar explícitamente cualquier intento de alterar campos estructurales (`Slug`, `Module`, `Action`) que pertenecen a la definición de código del backend.

## Reglas

- Las constantes globales de módulos y acciones del sistema deben residir obligatoriamente en `internal/platform/permissions/constants.go` para ser accesibles de manera transversal sin generar acoplamiento circular.
- Queda estrictamente prohibido el uso de strings crudos en los routers para definir permisos; todos deben ser ensamblados a través del formateador compilable `permissions.Format(permissions.ModuleX, permissions.ActionY)`.
- La acción `"index"` desaparece por completo del monolito. La acción de consulta de colecciones se unifica bajo `"list"`.
- Al iniciar el servidor web, se debe ejecutar la rutina `SyncPermissions` para actualizar la base de datos de manera transparente y automática antes de escuchar peticiones HTTP.
- Queda prohibida la creación y eliminación manual de permisos desde el API; toda la existencia y depuración estructural de los permisos es gobernada de forma automática por la rutina de autodescubrimiento y sincronización al arrancar el servidor.
- La edición de un permiso vía `PUT` únicamente es válida para los campos `Name` y `Description`. Intentar modificar `Slug`, `Module` o `Action` debe ser rechazado con un error de validación o forbidden.

## Criterios de aceptación

Este change se considera completo cuando:

1. Se crea el archivo `internal/platform/permissions/constants.go` centralizando los módulos y acciones del sistema.
2. Todas las llamadas a `requirePermission` en los archivos `routes.go` de todos los módulos utilizan `permissions.Format(...)` con constantes globales.
3. Se refactorizan todas las ocurrencias y referencias de la acción `"index"` a `"list"` en base de datos, routers, semillas, DTOs y tests, corriendo de forma exitosa.
4. Al arrancar el servidor, todos los permisos requeridos por los handlers/enrutadores se sincronizan e insertan automáticamente en la base de datos de manera idempotente (Upsert), poblando los campos de módulo y acción de forma correcta.
5. Intentar llamar a `POST /permissions` o `DELETE /permissions/:id` devuelve un error de ruta no soportada / método no permitido.
6. La actualización de un permiso mediante `PUT /permissions/:id` solo modifica `Name` y `Description`, y rechaza cualquier modificación de `Slug`, `Module` o `Action` con un error controlado.
7. Toda la suite de pruebas del proyecto (`go test ./...`) compila y pasa exitosamente (`GREEN`) con las nuevas firmas e integraciones.
