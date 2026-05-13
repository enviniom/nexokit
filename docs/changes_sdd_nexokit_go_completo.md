# NexoKit — Changes SDD para framework starter API en Go

## Identidad del proyecto

### Nombre

```txt
NexoKit
```

### Repositorio sugerido

```txt
nexokit-go
```

### CLI futuro sugerido

```bash
nexokit new tienda-saas
nexokit make module products
nexokit make migration create_products_table
nexokit create-root
nexokit permissions sync
```

### Descripción corta

NexoKit es un framework starter modular en Go para construir APIs listas para SaaS con autenticación, RBAC, multitenancy, migraciones, cache, rate limiting, logging, testing y convenciones listas para producción.

### Descripción en inglés

NexoKit is a modular Go framework starter for building SaaS-ready APIs with authentication, RBAC, multitenancy, migrations, caching, rate limiting, logging, testing and production-ready conventions.

---

# Objetivo general

Crear un marco de trabajo inicial en Go que permita iniciar nuevos proyectos backend con el boilerplate y las funcionalidades comunes ya implementadas.

NexoKit debe servir como base para:

- APIs REST.
- Aplicaciones SaaS.
- Paneles administrativos.
- Backends multitenant.
- Productos propios.
- La futura tienda virtual SaaS.

La idea no es construir una API específica para una tienda, sino un framework starter reutilizable que ya incluya:

- Estructura de proyecto.
- Configuración por ambiente.
- Base de datos.
- ORM con GORM.
- Migraciones.
- Autenticación.
- Usuario root inicial.
- Sistema de roles.
- Un rol por usuario.
- RBAC.
- Multitenancy por `company_id`.
- Respuestas API estandarizadas.
- DTOs.
- Validaciones.
- Paginación.
- Filtros.
- Rate limit.
- Cache con Redis o Valkey.
- Logger.
- Log rotation.
- Manejo centralizado de errores.
- Middleware común.
- Health checks.
- Seeds iniciales.
- Testing.
- CI básico.
- Developer experience.
- Documentación mínima.

---

# Decisiones base iniciales

Para NexoKit se asumen estas decisiones iniciales:

```txt
Lenguaje: Go
Framework HTTP: Gin
ORM: GORM
Base de datos inicial: PostgreSQL
Cache: Redis o Valkey
Migraciones: golang-migrate
Autenticación: JWT access token + refresh token
Roles: un solo rol por usuario
Autorización: RBAC por permisos
Multitenancy: company_id
Deploy: binario Go + systemd o Docker simple
Testing: testing estándar de Go + httptest + testify opcional
```

## Decisiones sugeridas

Para acelerar el desarrollo inicial se recomienda:

```txt
Gin
GORM
PostgreSQL
Goose (atlas o golang-migrate)
JWT
bcrypt o argon2id
Redis/Valkey opcional
uint como ID principal
Soft deletes
Request ID
Logger estructurado
Lumberjack para rotación de logs
```

---

# Recomendación sobre los changes

Conviene separar NexoKit en varios changes SDD atómicos.

No conviene hacer un solo change gigante porque se mezclarían decisiones de arquitectura, auth, RBAC, multitenancy, cache, logging, testing y developer experience.

Changes recomendados:

1. Base del proyecto, configuración, GORM, migraciones y respuesta estándar.
2. Auth, usuario root, usuarios, roles y refresh tokens.
3. RBAC, permisos y autorización.
4. Multitenancy por `company_id`.
5. Utilidades API: DTOs, validaciones, paginación, filtros, errores y documentación.
6. Infraestructura transversal: logger, log rotator, cache, rate limit y health checks.
7. Testing, calidad y CI básico.
8. CLI y developer experience para nuevos módulos.

---

# Change 1: Base del proyecto, configuración, GORM, migraciones y respuesta estándar

## Objetivo

Crear la base técnica de NexoKit, dejando lista una estructura limpia, modular y reutilizable para futuros proyectos.

Este change debe enfocarse en la arquitectura inicial, configuración, conexión a base de datos, ORM, migraciones, respuesta API estándar y convenciones generales del proyecto.

## Contexto

NexoKit será usado como punto de partida para múltiples aplicaciones, incluyendo una plataforma SaaS de tiendas virtuales. Por eso debe ser suficientemente genérico, pero preparado para multitenancy, auth, RBAC y crecimiento modular.

## Stack

- Go.
- Gin como framework HTTP.
- GORM como ORM.
- PostgreSQL como base de datos inicial.
- Migraciones con Goose o golang-migrate.
- Variables de entorno.
- Configuración centralizada.

## Alcance de este change

Implementar:

- Estructura base del proyecto.
- Carga de configuración desde `.env`.
- Configuración por ambiente: local, development, production, test.
- Conexión a PostgreSQL.
- Inicialización de GORM.
- Sistema de migraciones.
- Health check básico.
- Respuesta API estándar.
- Manejo centralizado de errores.
- Logger inicial simple.
- Middleware base.
- CORS configurable.
- Request ID.
- Recovery middleware básico.
- Estructura para módulos.
- Documentación inicial del proyecto.

## Estructura sugerida

```txt
nexokit-go/
  cmd/
    api/
      main.go
  internal/
    app/
    config/
    database/
    server/
    middleware/
    response/
    errors/
    logger/
    validator/
    modules/
  migrations/
  scripts/
  docs/
  .env.example
  go.mod
  README.md
```

## Estructura modular sugerida

Cada módulo futuro debería seguir esta forma:

```txt
internal/modules/example/
  handler.go
  service.go
  repository.go
  dto.go
  model.go
  routes.go
  validation.go
```

También puede aceptarse una estructura más separada:

```txt
internal/modules/example/
  handlers/
  services/
  repositories/
  dtos/
  models/
  routes/
```

NexoKit debe escoger una de las dos y justificarla.

Para simplicidad inicial, se recomienda la primera estructura, con archivos por responsabilidad dentro del módulo.

## Variables de entorno iniciales

```txt
APP_NAME
APP_ENV
APP_PORT
APP_URL

DB_HOST
DB_PORT
DB_NAME
DB_USER
DB_PASSWORD
DB_SSL_MODE

CORS_ALLOWED_ORIGINS
```

Opcionalmente permitir:

```txt
DATABASE_URL
```

## Respuesta API estándar

Todas las respuestas deben usar un formato común.

### Éxito

```json
{
  "success": true,
  "message": "Operación exitosa",
  "data": {},
  "meta": null,
  "errors": null
}
```

### Error

```json
{
  "success": false,
  "message": "Error de validación",
  "data": null,
  "meta": null,
  "errors": {
    "field": ["mensaje de error"]
  }
}
```

### Respuesta paginada

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
    }
  },
  "errors": null
}
```

## Health check

Endpoint:

```txt
GET /health
```

Debe responder:

```json
{
  "success": true,
  "message": "API is healthy",
  "data": {
    "status": "ok"
  },
  "meta": null,
  "errors": null
}
```

## Migraciones

Definir herramienta recomendada entre:

- Goose.
- golang-migrate.

Recomendación inicial: Goose, porque es simple, cómodo y suficiente para una plantilla productiva.

NexoKit debe incluir comandos para:

```txt
crear migración
ejecutar migraciones
revertir migraciones
consultar estado de migraciones
```

## Criterios de aceptación

Este change se considera completo cuando:

1. El proyecto compila correctamente.
2. La API inicia desde `cmd/api/main.go`.
3. La configuración se carga desde variables de entorno.
4. Existe `.env.example`.
5. Existe conexión funcional a PostgreSQL.
6. GORM queda configurado.
7. Existe sistema de migraciones.
8. Existe endpoint `/health`.
9. Todas las respuestas usan el formato estándar.
10. Los errores se responden de forma consistente.
11. Existe CORS configurable.
12. Existe Request ID.
13. Existe recovery middleware.
14. Existe una estructura modular base.
15. Existe README inicial con instrucciones para correr el proyecto.
16. Existen scripts o comandos para migraciones.

---

# Change 2: Auth, usuario root inicial, usuarios, roles y refresh tokens

## Objetivo

Implementar el sistema base de autenticación y gestión de usuarios de NexoKit.

Este change debe dejar listo:

- Login.
- JWT access token.
- Refresh token.
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
- Hash seguro de contraseña.
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

Opciones recomendadas:

### Opción A: seed por variables de entorno

```txt
ROOT_USER_NAME
ROOT_USER_EMAIL
ROOT_USER_PASSWORD
```

### Opción B: comando CLI

```txt
go run cmd/api/main.go create-root
```

### Opción C: seed solo en primera migración

Menos recomendable si la contraseña queda quemada.

## Recomendación

Usar comando CLI o seed basado en variables de entorno.

Nunca dejar una contraseña root fija en el código.

## Endpoints

### Auth

```txt
POST /api/auth/login
POST /api/auth/refresh
POST /api/auth/logout
GET /api/auth/me
```

### Users

```txt
GET /api/users
POST /api/users
GET /api/users/:id
PUT /api/users/:id
DELETE /api/users/:id
PUT /api/users/:id/password
```

### Roles

```txt
GET /api/roles
GET /api/roles/:id
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
    "access_token": "jwt",
    "refresh_token": "refresh",
    "user": {
      "id": "uuid",
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

- Password hashing con bcrypt o argon2id.
- JWT firmado con secreto configurable.
- Expiración corta para access token.
- Expiración más larga para refresh token.
- Refresh tokens guardados hasheados.
- Revocación de refresh token al logout.
- Validación de usuario activo.
- No revelar si falló email o contraseña individualmente.
- No devolver contraseñas ni hashes en respuestas.

## Variables de entorno

```txt
JWT_ACCESS_SECRET
JWT_REFRESH_SECRET
JWT_ACCESS_TTL_MINUTES
JWT_REFRESH_TTL_HOURS

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
10. El endpoint `/api/auth/me` retorna el usuario autenticado.
11. El middleware de autenticación protege rutas.
12. Un usuario inactivo no puede iniciar sesión.
13. Todas las respuestas usan el DTO estándar.
14. Las contraseñas nunca se devuelven en respuestas.

---

# Change 3: RBAC, permisos y autorización

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
router.POST("/api/users", AuthMiddleware(), RequirePermission("users.create"), handler.Create)
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
12. La respuesta de `/api/auth/me` incluye rol y permisos.

---

# Change 4: Multitenancy por company_id

## Objetivo

Implementar multitenancy por `company_id` como funcionalidad base de NexoKit.

El sistema debe permitir que una misma API y una misma base de datos sirvan a múltiples empresas, manteniendo aislamiento de datos.

## Contexto

NexoKit será usado para aplicaciones SaaS. En el caso de la tienda virtual, cada empresa tendrá sus propios productos, pedidos, clientes, usuarios y configuración.

Por eso NexoKit debe traer multitenancy como característica base, no como agregado posterior.

## Alcance de este change

Implementar:

- Modelo Company.
- Company ID en usuarios.
- Tenant context.
- Middleware de tenant.
- Repositorios filtrados por company_id.
- Helpers para aplicar scope por tenant.
- Protección contra acceso cruzado.
- Root con acceso global.
- Admin/user limitado a su company.
- Resolución de tenant por header, dominio o subdominio.

## Tabla companies

```txt
id
name
slug
domain nullable
subdomain nullable
status
created_at
updated_at
deleted_at nullable
```

## Ajuste en users

```txt
company_id nullable
```

Reglas:

- Root puede tener `company_id` nulo.
- Admin y user deben tener `company_id`.
- El sistema debe validar esto.

## Resolución de tenant

Soportar inicialmente:

```txt
X-Company-ID
Host header
Subdomain
Domain
```

Orden sugerido para APIs privadas:

1. Usuario autenticado.
2. Company del usuario.
3. Header solo si root o modo desarrollo.

Orden sugerido para APIs públicas:

1. Host header.
2. Domain configurado.
3. Subdomain configurado.
4. Header `X-Tenant` solo en desarrollo.

## Tenant context

Crear un contexto reusable:

```txt
TenantContext
- company_id
- company_slug
- is_root_scope
```

## GORM scopes

Crear helpers:

```go
func WithCompany(db *gorm.DB, companyID uuid.UUID) *gorm.DB
func ApplyTenantScope(db *gorm.DB, ctx TenantContext) *gorm.DB
```

Regla:

- Si el usuario es root y está en modo global, puede consultar todo.
- Si el usuario no es root, todas las consultas deben filtrar por `company_id`.

## Endpoints de companies

```txt
GET /api/companies
POST /api/companies
GET /api/companies/:id
PUT /api/companies/:id
DELETE /api/companies/:id
```

Solo root debe poder crear empresas.

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe tabla companies.
2. Los usuarios pueden estar asociados a company.
3. El root puede crear companies.
4. El admin no puede crear companies.
5. El tenant se carga en contexto.
6. Las consultas se filtran por company_id.
7. Un admin no puede leer datos de otra company.
8. Un admin no puede modificar datos de otra company.
9. El root puede operar en modo global.
10. El root puede operar en contexto de una company específica.
11. El middleware de tenant funciona en rutas privadas.
12. El middleware de tenant funciona en rutas públicas.
13. Existe helper o scope reusable de GORM para aplicar tenant.
14. Existe documentación de cómo crear nuevos modelos multitenant.

---

# Change 5: Utilidades API: DTOs, validaciones, paginación, filtros, errores y documentación

## Objetivo

Crear las utilidades reutilizables que harán que todos los módulos futuros de NexoKit tengan un comportamiento consistente.

Este change debe dejar listas las bases para crear endpoints limpios y repetibles.

## Alcance de este change

Implementar:

- DTOs base.
- Validación de requests.
- Manejo de errores de validación.
- Paginación.
- Filtros.
- Ordenamiento.
- Search básico.
- Soft delete conventions.
- Respuestas estandarizadas.
- Documentación de convenciones.
- Ejemplo de módulo CRUD.

## DTOs base

Crear estructuras para:

```txt
APIResponse
ErrorResponse
ValidationErrorResponse
PaginatedResponse
PaginationMeta
```

## Validación

Evaluar:

- go-playground/validator.
- Validaciones manuales por DTO.
- Sistema propio de reglas.

Recomendación práctica:

- Para velocidad inicial: usar `go-playground/validator`.
- Permitir validaciones manuales por DTO cuando la regla sea compleja.

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
GET /api/users?page=1&per_page=15&sort=created_at&order=desc&search=jhon
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

Crear errores de aplicación:

```txt
ErrNotFound
ErrUnauthorized
ErrForbidden
ErrValidation
ErrConflict
ErrInternal
```

El handler debe convertir errores a respuestas HTTP consistentes.

## Ejemplo CRUD

Crear un módulo ejemplo, por ejemplo:

```txt
internal/modules/examples
```

O usar `companies`/`users` como referencia.

El módulo debe mostrar:

- Listado paginado.
- Filtros.
- Crear.
- Editar.
- Eliminar.
- Validaciones.
- Respuesta estándar.
- Tenant scope si aplica.

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

## Criterios de aceptación

Este change se considera completo cuando:

1. Existen DTOs base de respuesta.
2. Existen DTOs de paginación.
3. Existe parser de query params para paginación.
4. Existe parser de filtros base.
5. Existe helper GORM para paginación.
6. Existe helper GORM para sorting.
7. Existe helper GORM para search.
8. Los errores se manejan de forma centralizada.
9. Las validaciones retornan errores por campo.
10. Existe ejemplo de endpoint paginado.
11. Existe ejemplo de endpoint con filtros.
12. Existe documentación de convenciones.
13. Todos los endpoints base usan la respuesta estándar.

---

# Change 6: Infraestructura transversal: logger, log rotator, cache, rate limit y health checks

## Objetivo

Agregar funcionalidades transversales de infraestructura que deben estar listas en NexoKit para cualquier proyecto.

Este change debe implementar:

- Logger estructurado.
- Rotación de logs.
- Cache Redis o Valkey.
- Rate limit.
- Health checks extendidos.
- Métricas básicas opcionales.

## Alcance de este change

Implementar:

- Logger estructurado.
- Middleware de request logging.
- Log rotator.
- Configuración de archivo de logs.
- Cache adapter.
- Redis/Valkey client.
- Rate limiter por IP.
- Rate limiter por endpoint sensible.
- Health check de base de datos.
- Health check de cache.
- Graceful shutdown.
- Recovery middleware.

## Logger

Evaluar:

- slog estándar de Go.
- zap.
- zerolog.

Recomendación inicial:

- `slog` si se quiere minimizar dependencias.
- `zap` si se quiere máximo rendimiento y estructura avanzada.

Para NexoKit se recomienda iniciar con `slog`, salvo que haya una razón clara para usar `zap`.

## Log rotation

Evaluar:

- lumberjack.
- logrotate del sistema operativo.

Recomendación para NexoKit:

- Usar `lumberjack` para que funcione igual en cualquier entorno.
- Permitir apagar logs a archivo en desarrollo si se desea.

Variables:

```txt
LOG_LEVEL
LOG_FORMAT
LOG_DIR
LOG_FILE
LOG_MAX_SIZE_MB
LOG_MAX_BACKUPS
LOG_MAX_AGE_DAYS
LOG_COMPRESS
```

## Cache

Soportar Redis o Valkey mediante cliente compatible.

Recomendación:

- Usar `go-redis`, compatible con Redis y Valkey.

Variables:

```txt
CACHE_DRIVER
REDIS_ADDR
REDIS_PASSWORD
REDIS_DB
CACHE_DEFAULT_TTL_SECONDS
```

El sistema debe permitir:

```txt
CACHE_DRIVER=none
```

para proyectos que no requieran cache.

## Cache adapter

Crear una interfaz:

```go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

Implementaciones:

```txt
RedisCache
NoopCache
```

## Rate limit

Implementar rate limit para:

- Login.
- Refresh token.
- Endpoints públicos sensibles.
- API general opcional.

Puede usarse:

- Memoria local para MVP/framework simple.
- Redis para rate limit distribuido.

Recomendación:

- Implementar interfaz.
- Dejar memoria local por defecto.
- Permitir Redis si está configurado.

Variables:

```txt
RATE_LIMIT_ENABLED
RATE_LIMIT_REQUESTS
RATE_LIMIT_WINDOW_SECONDS
LOGIN_RATE_LIMIT_REQUESTS
LOGIN_RATE_LIMIT_WINDOW_SECONDS
```

## Health checks

Endpoints:

```txt
GET /health
GET /health/live
GET /health/ready
```

`/health/live`:

- API está viva.

`/health/ready`:

- DB conectada.
- Cache conectada si está habilitada.

## Graceful shutdown

Implementar cierre controlado:

- Detener servidor HTTP.
- Cerrar conexión DB.
- Cerrar cache.
- Respetar timeout.

Variables:

```txt
SHUTDOWN_TIMEOUT_SECONDS
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe logger estructurado.
2. Los requests se registran con método, path, status y duración.
3. Los errores se registran con contexto.
4. Existe rotación de logs.
5. Los logs se guardan en archivo si está configurado.
6. Los logs pueden imprimirse en consola en desarrollo.
7. Existe cliente Redis/Valkey configurable.
8. Existe cache adapter.
9. Existe NoopCache si cache está deshabilitada.
10. Existe rate limit global opcional.
11. Existe rate limit para login.
12. Los endpoints rate-limited responden 429 cuando corresponde.
13. `/health/live` funciona.
14. `/health/ready` valida DB y cache.
15. Existe graceful shutdown.
16. Existe recovery middleware.

---

# Change 7: Testing, calidad y CI básico

## Objetivo

Agregar una estrategia de testing y calidad para NexoKit, asegurando que las funcionalidades base del framework puedan verificarse automáticamente y que los proyectos creados con NexoKit tengan una base confiable de pruebas.

## Principio clave

En Go, los unit tests deben vivir junto al código que prueban, usando archivos `_test.go` dentro del mismo paquete.

La carpeta `/tests` debe reservarse principalmente para:

- Integration tests.
- Test helpers.
- Fixtures.
- Setup de base de datos de prueba.
- Pruebas end-to-end internas si más adelante se requieren.

## Alcance

Implementar:

- Unit tests.
- Integration tests.
- Test database.
- Test helpers.
- Factories o fixtures.
- Mocks para servicios externos cuando aplique.
- Tests de middlewares.
- Tests de respuesta API estándar.
- Tests de auth.
- Tests de RBAC.
- Tests de tenant scope.
- Tests de paginación y filtros.
- Tests de cache adapter.
- Tests de rate limit.
- Coverage.
- Makefile con comandos de test.
- GitHub Actions básico.

## Herramientas sugeridas

- `testing` estándar de Go.
- `httptest` para endpoints HTTP.
- `testify` opcional para assertions.
- PostgreSQL de test usando Docker Compose.
- Redis/Valkey de test usando Docker Compose.
- GitHub Actions para CI.

## Estructura recomendada

### Unit tests junto al código

```txt
internal/
  response/
    response.go
    response_test.go

  auth/
    service.go
    service_test.go
    password.go
    password_test.go
    token.go
    token_test.go

  middleware/
    auth.go
    auth_test.go
    rbac.go
    rbac_test.go
    tenant.go
    tenant_test.go
    rate_limit.go
    rate_limit_test.go

  pagination/
    pagination.go
    pagination_test.go

  filters/
    filters.go
    filters_test.go

  cache/
    cache.go
    noop.go
    redis.go
    noop_test.go
    redis_test.go

  tenant/
    context.go
    scope.go
    scope_test.go
```

### Integration tests separados

```txt
tests/
  integration/
    auth_test.go
    users_test.go
    tenant_test.go
    rbac_test.go
    health_test.go

  helpers/
    app.go
    database.go
    auth.go
    fixtures.go

  fixtures/
    users.go
    companies.go
    roles.go
```

## Paquetes de testing

Usar el mismo paquete cuando se necesite acceder a funciones no exportadas:

```go
package auth
```

Usar paquete externo cuando se quiera probar como consumidor del paquete:

```go
package auth_test
```

## Buenas prácticas obligatorias

Los tests de NexoKit deben seguir estas reglas:

- Los unit tests deben estar junto al código probado.
- Los archivos deben terminar en `_test.go`.
- Los tests deben usar nombres descriptivos.
- Usar subtests con `t.Run`.
- Usar table-driven tests para casos repetitivos.
- Revisar errores explícitamente.
- Usar `t.Fatal` o `t.Fatalf` para errores de setup.
- Usar `t.Error` o `t.Errorf` para aserciones que permiten continuar.
- Evitar panics en código productivo.
- Preferir interfaces pequeñas para facilitar mocks.
- No abusar de frameworks pesados de testing.
- Usar `testing` estándar de Go como base.
- `testify` puede usarse para assertions si aporta claridad, pero no debe ocultar la lógica.

## Ejemplo de table-driven test

```go
func TestPagination_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		perPage  int
		wantPage int
		wantPer  int
	}{
		{
			name:     "default values",
			page:     0,
			perPage:  0,
			wantPage: 1,
			wantPer:  15,
		},
		{
			name:     "negative page becomes first page",
			page:     -1,
			perPage:  20,
			wantPage: 1,
			wantPer:  20,
		},
		{
			name:     "per page above max is capped",
			page:     1,
			perPage:  500,
			wantPage: 1,
			wantPer:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePagination(tt.page, tt.perPage)

			if got.Page != tt.wantPage {
				t.Errorf("expected page %d, got %d", tt.wantPage, got.Page)
			}

			if got.PerPage != tt.wantPer {
				t.Errorf("expected per_page %d, got %d", tt.wantPer, got.PerPage)
			}
		})
	}
}
```

## Testing por capas

### Unit tests

Para lógica pura y rápida:

```txt
response
errors
pagination
filters
validators
password hashing
JWT generation/parsing
RBAC permission checks
tenant context
cache interface noop
rate limit logic
```

### Handler tests

Usando `httptest`:

```txt
login endpoint
me endpoint
protected route without token
protected route with invalid token
protected route without permission
paginated list endpoint
```

### Integration tests

Para probar componentes conectados:

```txt
login real con DB de test
refresh token real con DB
CRUD users
tenant isolation con DB
RBAC sobre rutas HTTP
migraciones
repositories GORM
cache Redis/Valkey
```

## Casos mínimos de prueba

### Response

- Respuesta exitosa.
- Respuesta de error.
- Respuesta de validación.
- Respuesta paginada.

### Auth

- Login exitoso.
- Login con credenciales inválidas.
- Usuario inactivo no puede iniciar sesión.
- Refresh token válido.
- Refresh token revocado.
- Logout revoca refresh token.

### RBAC

- Usuario con permiso accede.
- Usuario sin permiso recibe 403.
- Usuario no autenticado recibe 401.
- Root tiene todos los permisos.

### Multitenancy

- Admin solo accede a datos de su company.
- Admin no puede acceder a datos de otra company.
- Root puede acceder globalmente.
- Tenant se resuelve correctamente desde contexto.

### Paginación y filtros

- Paginación por defecto.
- Cambio de `page`.
- Cambio de `per_page`.
- Sorting.
- Search.
- Filtro por estado.

### Cache

- Cache disabled usa NoopCache.
- Redis/Valkey set/get/delete.
- TTL funciona correctamente.

### Rate limit

- Request dentro del límite pasa.
- Request por encima del límite responde 429.
- Login tiene rate limit propio.

## Comandos esperados

```txt
make test
make test-unit
make test-integration
make test-coverage
```

## CI básico

Crear workflow de GitHub Actions que ejecute:

```txt
go test ./...
go vet ./...
go fmt check
```

Opcional:

```txt
golangci-lint
coverage report
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe estructura base de tests.
2. Los unit tests viven junto al código probado.
3. Los integration tests viven en `/tests/integration`.
4. Existen tests unitarios para response.
5. Existen tests para validaciones.
6. Existen tests para auth.
7. Existen tests para RBAC.
8. Existen tests para tenant scope.
9. Existen tests para paginación.
10. Existen tests para filtros.
11. Existen tests para cache.
12. Existen tests para rate limit.
13. Existe comando `make test`.
14. Existe comando `make test-unit`.
15. Existe comando `make test-integration`.
16. Existe comando `make test-coverage`.
17. Existe CI básico con GitHub Actions.
18. La documentación explica cómo correr tests.
19. Los tests pueden ejecutarse en ambiente local.
20. Los tests siguen buenas prácticas idiomáticas de Go.

---

# Change 8: CLI y developer experience para nuevos módulos

## Objetivo

Agregar herramientas de developer experience para que NexoKit sea realmente útil al iniciar nuevos proyectos.

Este change puede ser opcional para la primera versión, pero es muy valioso para acelerar desarrollo futuro.

## Alcance

Implementar un CLI o comandos simples para:

- Crear usuario root.
- Ejecutar migraciones.
- Revertir migraciones.
- Crear migración.
- Crear módulo base.
- Crear seed.
- Ver configuración actual.
- Ver estado de la aplicación.

## Opciones

### Opción A: comandos dentro del binario principal

Ejemplo:

```txt
go run cmd/api/main.go serve
go run cmd/api/main.go create-root
go run cmd/api/main.go migrate up
go run cmd/api/main.go migrate down
go run cmd/api/main.go make module products
```

### Opción B: CLI separado

```txt
cmd/cli/main.go
```

Ejemplo:

```txt
go run cmd/cli/main.go make module products
```

### Opción C: Makefile + scripts

Ejemplo:

```txt
make dev
make migrate-up
make migrate-down
make make-module name=products
make create-root
```

## Recomendación

Para la primera versión:

- Usar Makefile + scripts.
- Crear comando interno para `create-root`.
- Dejar generación de módulos como fase opcional.

Para una versión más madura:

```bash
nexokit new app-name
nexokit make module products
nexokit make migration create_products_table
nexokit create-root
nexokit permissions sync
```

## Generador de módulos

Si se implementa, debe generar:

```txt
internal/modules/products/
  handler.go
  service.go
  repository.go
  dto.go
  model.go
  routes.go
  validation.go
```

Y opcionalmente:

```txt
migrations/YYYYMMDDHHMMSS_create_products_table.sql
```

## Makefile sugerido

```txt
make dev
make build
make test
make test-unit
make test-integration
make test-coverage
make migrate-up
make migrate-down
make migrate-create name=create_users_table
make seed
make create-root
make lint
make fmt
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe Makefile funcional.
2. Existe comando para correr la API en desarrollo.
3. Existe comando para compilar.
4. Existe comando para correr tests.
5. Existe comando para crear migraciones.
6. Existe comando para ejecutar migraciones.
7. Existe comando para revertir migraciones.
8. Existe comando seguro para crear root.
9. Existe documentación de comandos.
10. Opcionalmente existe generador básico de módulos.

---

# Orden recomendado de implementación

## Orden recomendado para tu caso

Como NexoKit se usará para potenciar el desarrollo de la tienda SaaS, el orden recomendado es:

1. Change 1: Base técnica.
2. Change 2: Auth, root, usuarios y roles.
3. Change 4: Multitenancy.
4. Change 3: RBAC.
5. Change 5: DTOs, paginación, filtros y errores.
6. Change 6: Logger, cache, rate limit y health checks.
7. Change 7: Testing, calidad y CI básico.
8. Change 8: CLI y developer experience.

La razón es que para la tienda SaaS necesitas `company_id` muy pronto. RBAC puede quedar inmediatamente después.

## Orden ideal si se quiere máxima limpieza conceptual

1. Change 1: Base técnica.
2. Change 2: Auth, root, usuarios y roles.
3. Change 3: RBAC.
4. Change 4: Multitenancy.
5. Change 5: DTOs, paginación, filtros y errores.
6. Change 6: Logger, cache, rate limit y health checks.
7. Change 7: Testing, calidad y CI básico.
8. Change 8: CLI y developer experience.

---

# Qué dejar fuera de la primera versión de NexoKit

Para no hacer el framework demasiado pesado, dejar fuera inicialmente:

```txt
GraphQL
gRPC
WebSockets
CQRS
Event sourcing
Microservicios
Kubernetes
OpenTelemetry avanzado
Prometheus/Grafana obligatorio
Multi-role por usuario
ABAC complejo
Permisos por recurso individual
Auditoría avanzada
2FA
SSO
OAuth con Google/Facebook
Colas con RabbitMQ/Kafka
Sistema de notificaciones
Emails transaccionales
File storage S3
```

Algunas de estas cosas pueden ser módulos opcionales después.

---

# Cosas que sí conviene adicionar desde el inicio

## UUIDs

Usar UUID como ID principal en vez de enteros autoincrementales.

## Soft deletes

Usar `deleted_at` en entidades importantes.

## Request ID

Middleware que agregue:

```txt
X-Request-ID
```

Esto ayuda mucho para logs y debugging.

## Recovery middleware

Para evitar que un panic tumbe la API.

## Audit fields básicos

En modelos importantes, considerar:

```txt
created_by
updated_by
```

Puede ser opcional, pero es útil para proyectos SaaS.

## Estado estándar

Muchos modelos deberían tener:

```txt
status
```

Ejemplo:

```txt
active
inactive
```

## Seeds

Seeds claros para:

```txt
roles
permissions
root user
```

## Testing mínimo desde temprano

NexoKit debe incluir pruebas para:

```txt
response
auth
RBAC
tenant scope
pagination
filters
rate limit
cache
```

---

# Prompt corto para iniciar el Change 1

```md
Quiero iniciar el Change 1 de NexoKit, un framework starter modular en Go.

El objetivo es crear una base reutilizable para futuros proyectos backend y SaaS. NexoKit será usado después para construir una tienda virtual SaaS multitenant, pero no debe estar acoplado a la lógica de tienda.

Stack preferido:
- Go
- Gin
- GORM
- PostgreSQL
- Migraciones con Goose o golang-migrate
- Variables de entorno
- Respuesta API estandarizada
- Estructura modular
- Preparado para auth, RBAC y multitenancy en changes posteriores

Necesito que analices este change y generes:

1. Recomendación técnica final.
2. Estructura de carpetas.
3. Convenciones del proyecto.
4. Configuración por ambiente.
5. Inicialización del servidor HTTP.
6. Conexión a PostgreSQL con GORM.
7. Sistema de migraciones recomendado.
8. Estructura de respuesta API estándar.
9. Manejo centralizado de errores.
10. Middleware base.
11. Health check.
12. CORS configurable.
13. Request ID.
14. Recovery middleware.
15. README inicial esperado.
16. Criterios de aceptación.
17. Lista de tareas implementables en orden.

No implementes todavía. Primero genera el plan técnico detallado del Change 1.
```

---

# Preguntas pendientes para afinar antes de implementar

Estas preguntas no bloquean los changes, pero sí conviene responderlas antes de programar:

1. ¿Confirmamos UUID como ID principal en todas las tablas?
2. ¿Confirmamos Gin definitivamente?
3. ¿Confirmamos Goose para migraciones?
4. ¿Confirmamos PostgreSQL como única base soportada inicialmente?
5. ¿Redis/Valkey será opcional mediante `CACHE_DRIVER=none`?
6. ¿Prefieres bcrypt o argon2id para passwords?
7. ¿Los permisos serán administrables desde API o solo seeds iniciales?
8. ¿Quieres incluir auditoría básica con `created_by` y `updated_by`?
9. ¿Quieres que el generador de módulos sea parte de la primera versión o mejora posterior?
10. ¿Quieres Docker Compose para desarrollo local desde el inicio?
11. ¿Quieres usar `testify` o solo `testing` estándar?
12. ¿Quieres agregar GitHub Actions desde la primera versión?
