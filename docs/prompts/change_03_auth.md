> Lee también `_context.md` antes de implementar este change.

# Change 3: Auth con PASETO, usuario root inicial, usuarios, roles y refresh tokens

## Objetivo

Implementar el sistema base de autenticación y gestión de usuarios de NexoKit.

Este change debe dejar listo:

- Login.
- PASETO access token.
- Refresh token opaco.
- Logout.
- Usuario root inicial.
- Usuarios.
- Roles.
- Un solo rol por usuario.

## Decisión importante

El sistema manejará un solo rol por usuario.

No se implementarán múltiples roles por usuario en la versión inicial.

Esto simplifica:

- Consultas.
- Validaciones.
- UI futura.
- Seguridad.
- Mantenimiento.

## Alcance de este change

Implementar:

- Tabla de usuarios.
- Tabla de roles.
- Tabla de refresh tokens.
- Seed del rol root.
- Seed del usuario root inicial.
- Login.
- Refresh token.
- Logout.
- Endpoint `me`.
- Middleware de autenticación.
- Hash seguro de contraseña con argon2id.
- Cambio básico de contraseña.
- Estado de usuario activo/inactivo.

## Tablas

```txt
users
roles
refresh_tokens
```

## Campos sugeridos

### roles

```txt
id
public_id
name
slug
description
is_system
created_at
updated_at
```

Roles iniciales sugeridos:

```txt
root
admin
user
```

Para SaaS puede extenderse después a:

```txt
seller
manager
support
```

### users

```txt
id
public_id
company_id nullable
role_id
name
email
password_hash
status
last_login_at nullable
created_at
updated_at
deleted_at nullable
created_by nullable
updated_by nullable
```

### refresh_tokens

```txt
id
user_id
token_hash
expires_at
revoked_at nullable
created_at
updated_at
```

## Usuario root inicial

NexoKit debe permitir crear un usuario root inicial de forma segura.

Este change debe resolver el TODO dejado por el CLI interno en `internal/cli/root/root.go`: cablear `RootStorage` y `PasswordHasher` reales para que `go run ./cmd/nexokit create-root` deje de retornar `ErrStorageNotWired` y pueda crear el usuario root de forma idempotente.

Opciones recomendadas:

### Opción A: seed por variables de entorno

```txt
ROOT_USER_NAME
ROOT_USER_EMAIL
ROOT_USER_PASSWORD
```

### Opción B: comando CLI

```txt
go run cmd/nexokit create-root
```

## Recomendación

Usar comando CLI combinado con variables de entorno.

Nunca dejar una contraseña root fija en el código.

## Endpoints

### Auth

```txt
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
GET  /api/v1/auth/me
```

### Users

```txt
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id
PUT    /api/v1/users/:id/password
```

### Roles

```txt
GET /api/v1/roles
GET /api/v1/roles/:id
```

Inicialmente los roles pueden ser de solo lectura desde API para evitar que se dañen roles del sistema.

## Login request

```json
{
  "email": "admin@example.com",
  "password": "secret"
}
```

## Login response

```json
{
  "success": true,
  "message": "Login exitoso",
  "data": {
    "access_token": "paseto",
    "refresh_token": "opaque-refresh-token",
    "user": {
      "id": "01HY7V8J3F8WQ9F6K2H4D1M5NP",
      "name": "Root User",
      "email": "root@example.com",
      "role": "root"
    }
  },
  "meta": null,
  "errors": null
}
```

## Seguridad

Incluir:

- Password hashing con argon2id.
- PASETO con clave configurable.
- Expiración corta para access token.
- Expiración más larga para refresh token.
- Refresh tokens guardados hasheados.
- Revocación de refresh token al logout.
- Validación de usuario activo.
- No revelar si falló email o contraseña individualmente.
- No devolver contraseñas ni hashes en respuestas.

## Decisión sobre tokens

NexoKit usará PASETO en lugar de JWT para los access tokens.

```txt
Access token: PASETO v4.local.
Refresh token: token opaco random, guardado únicamente como hash.
```

PASETO evita varias clases de errores comunes de JWT, especialmente confusiones con algoritmos, validaciones incompletas y diferencias entre token firmado y token cifrado.

El refresh token debe seguir siendo opaco porque necesita revocación, rotación y almacenamiento seguro del hash.

Claims mínimos sugeridos para el access token:

```txt
sub: public_id del usuario
role: slug del rol
company_id: public_id de la company, si aplica
token_type: access
issued_at
expires_at
```

No guardar datos sensibles dentro del token aunque se use PASETO local cifrado.

## Variables de entorno

```txt
PASETO_VERSION
PASETO_LOCAL_KEY
PASETO_ACCESS_TTL_MINUTES
REFRESH_TOKEN_TTL_HOURS

ROOT_USER_NAME
ROOT_USER_EMAIL
ROOT_USER_PASSWORD
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existen migraciones para users, roles y refresh_tokens.
2. Existen seeds para roles iniciales.
3. Se puede crear usuario root inicial sin contraseña quemada.
4. Un usuario puede iniciar sesión.
5. Un usuario recibe access token y refresh token.
6. Un usuario puede refrescar sesión.
7. Un usuario puede cerrar sesión.
8. El refresh token se guarda hasheado.
9. El logout revoca el refresh token.
10. El endpoint `/api/v1/auth/me` retorna el usuario autenticado.
11. El middleware de autenticación protege rutas.
12. Un usuario inactivo no puede iniciar sesión.
13. Todas las respuestas usan el DTO estándar.
14. Las contraseñas nunca se devuelven en respuestas.
15. El comando `go run ./cmd/nexokit create-root` usa el storage real de usuarios/roles y el hasher argon2id, no deja pendiente el TODO de `internal/cli/root/root.go`.
