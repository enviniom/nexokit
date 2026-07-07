> Lee también `_context.md` antes de implementar este change.

# Change 4: RBAC, permisos y autorización

## Objetivo

Implementar RBAC en NexoKit, usando un solo rol por usuario y permisos asociados a roles.

El sistema debe permitir proteger endpoints por permisos, no solamente por nombre de rol.

## Contexto

Aunque inicialmente cada usuario tendrá un solo rol, el sistema debe estar preparado para administrar permisos de forma flexible.

Ejemplo:

```txt
root -> todos los permisos
admin -> permisos administrativos de su tenant
user -> permisos básicos
```

Para la tienda SaaS después podrían existir:

```txt
seller -> products.read, orders.read, orders.update_status
```

## Alcance de este change

Implementar:

- Permisos.
- Relación rol-permisos.
- Seeds de permisos base.
- Middleware `RequirePermission`.
- Middleware `RequireRole` opcional.
- Helper para consultar permisos del usuario.
- Cache opcional de permisos.
- Protección de rutas usando permisos.

## Tablas

```txt
permissions
role_permissions
```

## Campos sugeridos

### permissions

```txt
id
public_id
name
slug
description
module
created_at
updated_at
```

Ejemplos:

```txt
users.read
users.create
users.update
users.delete

roles.read

companies.read
companies.create
companies.update
companies.delete

settings.read
settings.update
```

### role_permissions

```txt
role_id
permission_id
created_at
```

## Reglas

- Un usuario tiene un solo rol.
- Un rol puede tener muchos permisos.
- Un permiso puede pertenecer a muchos roles.
- Root debe tener todos los permisos.
- Los permisos se validan en middleware.
- La autorización debe estar separada de la autenticación.

## Middleware requerido

```txt
AuthMiddleware
RequireRole("root")
RequirePermission("users.create")
```

Ejemplo de uso:

```go
router.POST("/api/v1/users", AuthMiddleware(), RequirePermission("users.create"), handler.Create)
```

## Seeds iniciales

Crear permisos para módulos base:

```txt
users
roles
companies
settings
auth
```

Crear asignación inicial:

```txt
root -> todos los permisos
admin -> permisos de users/settings dentro de su company
user -> permisos mínimos
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existen tablas de permissions y role_permissions.
2. Existen seeds de permisos base.
3. El rol root tiene todos los permisos.
4. El rol admin tiene permisos administrativos básicos.
5. El middleware `RequirePermission` funciona.
6. El middleware `RequireRole` funciona si se decide mantener.
7. Una ruta puede protegerse por permiso.
8. Un usuario sin permiso recibe 403.
9. Un usuario no autenticado recibe 401.
10. La autenticación y la autorización están separadas.
11. Los permisos se pueden consultar desde el usuario autenticado.
12. La respuesta de `/api/v1/auth/me` incluye rol y permisos.
