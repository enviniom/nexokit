> Lee también `_context.md` antes de implementar este change.

# Change 12: Onboarding de company con roles y permisos iniciales

## Objetivo

Crear un flujo de onboarding para registrar una company y dejarla lista para operar con roles `admin` y `user` propios del tenant.

Este flujo debe crear la company, sus roles base tenant y los permisos necesarios para esos roles.

## Contexto

Después de mover roles a tenant, el seed inicial ya no debe crear roles `admin` y `user` globales. Cada company necesita sus propios roles y asignaciones.

El onboarding debe ser el punto controlado donde se inicializa esa estructura.

## Alcance de este change

Implementar:

- Endpoint o comando de onboarding de company.
- Creación de company.
- Creación de roles `admin` y `user` para esa company.
- Creación o sincronización de permisos necesarios.
- Asignación de permisos iniciales a roles tenant.
- Validaciones de duplicados de company, dominio y subdominio.
- Tests de flujo completo y rollback lógico ante errores.

## Datos requeridos

El onboarding debe solicitar los datos mínimos para crear una company:

```txt
name
slug
domain nullable
subdomain nullable
admin_name
admin_email
admin_password opcional según estrategia de invitación
```

La estrategia exacta para crear/invitar al primer admin debe definirse en diseño SDD.

## Roles creados

Para cada company creada:

```txt
admin -> is_system = true, company_id = company.ID
user  -> is_system = true, company_id = company.ID
```

El rol `root` no se crea en este flujo y no pertenece a ninguna company.

## Permisos y catálogo de módulos

El onboarding necesita una fuente canónica de permisos disponibles para asignar a roles tenant.

Diseñar una estrategia para registrar módulos y acciones permitidas, por ejemplo:

```txt
module: users
actions: index, view, create, update, delete, change_role

module: roles
actions: index, view, create, update, delete, assign_permissions

module: companies
actions: index, view, create, update, delete
```

La aplicación debe poder detectar si un endpoint exige un permiso que no existe en el catálogo.

Opciones a evaluar en diseño:

1. Catálogo en código con sync a DB.
2. Catálogo en DB administrable desde API.
3. Enfoque híbrido: permisos declarados en código, sincronizados y visibles desde API.

Regla recomendada inicial:

- Los permisos protegidos por middleware deben existir en el catálogo.
- En desarrollo/test, si una ruta exige un permiso inexistente, debe fallar rápido.
- En producción, el sistema debe registrar el error y devolver 403/500 según decisión de diseño.
- Debe existir un reporte o endpoint administrativo para ver permisos no asignados a ningún rol.

## Endpoints sugeridos

```txt
POST /api/v1/onboarding/companies
GET  /api/v1/permissions/catalog
GET  /api/v1/permissions/unassigned
```

Los endpoints de catálogo pueden quedar como parte de este change o documentarse como subfase si el diseño indica que excede el tamaño recomendado.

## Criterios de aceptación

Este change se considera completo cuando:

1. Se puede crear una company por onboarding.
2. El onboarding crea roles `admin` y `user` asociados a la company.
3. El onboarding no crea ni modifica el rol `root`.
4. Los roles creados tienen permisos iniciales correctos.
5. El primer admin queda creado o invitado según la estrategia elegida.
6. No se puede repetir `slug`, `domain` o `subdomain` de company.
7. Si falla una parte crítica del onboarding, no queda una company parcialmente configurada.
8. Existe una fuente canónica para permisos disponibles.
9. Se puede detectar un permiso usado por un endpoint pero ausente del catálogo.
10. Se puede identificar un permiso existente que no está asignado a ningún rol.
11. Los tests cubren happy path, duplicados, rollback y permisos iniciales.
