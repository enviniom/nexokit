> Lee también `_context.md` antes de implementar este change.

# Change 9: Roles custom administrables

## Objetivo

Permitir administrar roles custom desde el API sin comprometer los roles base del sistema.

Este change extiende el RBAC ya implementado: los permisos y la relación rol-permisos existen, pero los roles `root`, `admin` y `user` siguen protegidos como roles del sistema.

## Contexto

En el RBAC inicial los roles base son de solo lectura para evitar que una configuración incorrecta rompa el acceso al sistema.

Para aplicaciones SaaS reales, puede ser necesario crear roles adicionales como:

```txt
seller
support
billing
warehouse
```

Estos roles custom deben poder recibir permisos mediante los endpoints de asignación de permisos existentes.

## Alcance de este change

Implementar:

- CRUD de roles custom.
- Protección de roles system.
- Validaciones de nombre y slug.
- Endpoints de creación, edición y eliminación de roles no-system.
- Integración con permisos existentes.
- Tests de protección y errores.

## Reglas

- Los roles `root`, `admin` y `user` son `is_system = true`.
- Un rol system no se puede editar ni borrar desde API.
- Un rol custom se puede crear, editar y borrar.
- No se puede borrar un rol que tenga usuarios asignados.
- No se puede crear un rol con slug duplicado.
- El slug debe ser estable y usable como identificador de negocio.
- La asignación de permisos al rol debe seguir usando `PUT /api/v1/roles/:id/permissions`.

## TODO relacionado: cambio de rol de usuarios

Separar la política de actualización general de usuario de la política de cambio de rol.

En el RBAC inicial `PUT /api/v1/users/:id` puede requerir `users.update` y `users.change_role` porque el DTO actual permite cambiar el rol dentro del update general.

Este comportamiento es seguro pero demasiado restrictivo: obliga a tener `users.change_role` incluso cuando solo se edita nombre, email u otros campos no relacionados al rol.

En este change o en uno posterior, definir una de estas estrategias:

```txt
PATCH /api/v1/users/:id/role -> requiere users.change_role
PUT   /api/v1/users/:id      -> requiere users.update y no permite cambiar role_id
```

Regla esperada:

- Editar datos generales de un usuario requiere `users.update`.
- Cambiar el rol de un usuario requiere `users.change_role`.
- Un usuario no debe poder escalar privilegios cambiándose a un rol superior.
- Root conserva bypass según las reglas globales de RBAC.

## Endpoints

```txt
GET    /api/v1/roles
POST   /api/v1/roles
GET    /api/v1/roles/:id
PUT    /api/v1/roles/:id
DELETE /api/v1/roles/:id
GET    /api/v1/roles/:id/permissions
PUT    /api/v1/roles/:id/permissions
```

## Permisos sugeridos

```txt
roles.index
roles.view
roles.create
roles.update
roles.delete
roles.assign_permissions
```

## Campos

La tabla `roles` ya existe. Confirmar que tenga o agregar si falta:

```txt
id
public_id
name
slug
description
is_system
created_at
updated_at
deleted_at nullable
created_by nullable
updated_by nullable
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Se puede crear un rol custom.
2. Se puede editar un rol custom.
3. Se puede eliminar un rol custom sin usuarios asignados.
4. No se puede editar un rol system.
5. No se puede eliminar un rol system.
6. No se puede eliminar un rol con usuarios asignados.
7. No se puede crear un rol con slug duplicado.
8. Los endpoints usan permisos `roles.*`.
9. La asignación de permisos sigue funcionando para roles custom.
10. Los roles custom aparecen en `GET /api/v1/roles` con sus permisos.
