> Lee también `_context.md` antes de implementar este change.

# Change 11: Cambio de rol de usuarios separado del update general

## Objetivo

Separar la edición general de usuarios del cambio de rol para evitar permisos demasiado amplios y prevenir escalamiento de privilegios.

`PUT /api/v1/users/:id` no debe permitir cambiar `role_id`. El cambio de rol debe hacerse por un endpoint explícito y protegido.

## Contexto

En RBAC, cambiar el rol de un usuario es una operación sensible. No debe mezclarse con editar nombre, email u otros datos generales.

Si el DTO general permite cambiar el rol, un endpoint común termina necesitando `users.change_role` aunque solo se quiera editar información básica.

## Alcance de este change

Implementar:

- Remover cambio de rol del update general de usuario.
- Endpoint dedicado para cambiar rol.
- Permiso específico `users.change_role`.
- Validaciones para no asignar root.
- Listado/select de roles asignables.
- Reglas tenant-aware para roles asignables.
- Tests de permisos, errores y prevención de escalamiento.

## Reglas

- Editar datos generales de usuario requiere `users.update`.
- Cambiar el rol de usuario requiere `users.change_role`.
- `PUT /api/v1/users/:id` no debe aceptar ni modificar `role_id`.
- El rol `root` no se puede asignar a nadie desde API.
- El rol `root` no debe aparecer en endpoints pensados para select/listado de roles asignables.
- Un admin solo puede asignar roles de su propia company.
- Un usuario no debe poder cambiarse el rol a sí mismo para escalar privilegios.
- Root conserva bypass global, pero aun así no debe usar este endpoint para asignar el rol root a usuarios comunes.

## Endpoints

```txt
PATCH /api/v1/users/:id/role
GET   /api/v1/users/assignable-roles
```

Alternativa aceptable si encaja mejor con la estructura actual:

```txt
GET /api/v1/roles/assignable
```

La decisión debe documentarse en diseño SDD antes de implementar.

## DTO esperado

```txt
PATCH /api/v1/users/:id/role

body:
{
  "role_id": "public_id_del_rol"
}
```

## Permisos sugeridos

```txt
users.update
users.change_role
roles.assignable
```

## Criterios de aceptación

Este change se considera completo cuando:

1. `PUT /api/v1/users/:id` no cambia el rol aunque venga `role_id` en el body.
2. Cambiar rol se hace solo por `PATCH /api/v1/users/:id/role`.
3. El endpoint de cambio de rol requiere `users.change_role`.
4. Un usuario sin `users.change_role` recibe 403 al intentar cambiar rol.
5. No se puede asignar el rol `root`.
6. El rol `root` no aparece en el endpoint de roles asignables.
7. Un admin solo puede asignar roles de su company.
8. No se puede asignar un rol de otra company.
9. No se puede usar el endpoint para auto-escalarse privilegios.
10. Los tests cubren update general, cambio de rol, select de roles y errores esperados.
