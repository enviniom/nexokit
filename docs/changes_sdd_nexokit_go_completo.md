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

### CLI sugerido

```bash
nexokit new tienda-saas
nexokit make module products
nexokit make migration create_products_table
nexokit create-root
```

### Descripción corta

NexoKit es un framework starter modular y opinionado en Go para construir APIs listas para SaaS con autenticación, RBAC, multitenancy, migraciones, cache, rate limiting, logging, testing, CLI de generación de módulos y convenciones listas para producción.

### Descripción en inglés

NexoKit is an opinionated modular Go framework starter for building SaaS-ready APIs with authentication, RBAC, multitenancy, migrations, caching, rate limiting, logging, module-generation CLI and production-ready conventions.

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
- Autenticación con PASETO.
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
- Developer experience y CLI para generación de módulos.
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
Migraciones: Goose
Autenticación: PASETO access token + refresh token opaco
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
Goose
PASETO
argon2id para passwords
Redis/Valkey opcional
uint como ID interno y PublicID como ID externo
Soft deletes
Auditoría básica: created_by / updated_by en modelos importantes (nullable)
Request ID
Logger estructurado con slog
Lumberjack para rotación de logs
Versionado de API desde el inicio: /api/v1/
```

## Estrategia de IDs

NexoKit usará una estrategia de IDs duales para equilibrar rendimiento interno, seguridad y experiencia de API.

### Regla base

```txt
ID interno: uint autoincremental, usado para primary keys, foreign keys, joins y lógica interna.
PublicID externo: string único, usado en API, URLs, logs públicos, eventos y referencias externas.
```

Los endpoints públicos y privados de API deben recibir y devolver `public_id`, no el `id` interno de base de datos.

En respuestas JSON, el `PublicID` puede exponerse como `id` para mantener una API limpia, mientras que el `ID` interno debe omitirse.

### BaseModel sugerido

```go
type BaseModel struct {
	ID        uint           `gorm:"primaryKey" json:"-"`
	PublicID  string         `gorm:"type:char(26);uniqueIndex;not null" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *uint          `gorm:"index" json:"-"`
	UpdatedBy *uint          `gorm:"index" json:"-"`
}
```

`CreatedBy` y `UpdatedBy` son nullable y opcionales. Los modelos que no requieran auditoría pueden omitirlos embebiendo un `BaseModelSimple` sin esos campos. La decisión de cuál usar queda en cada módulo.

### Formato del PublicID

Por defecto, NexoKit puede usar ULID (`char(26)`) para la mayoría de entidades porque es compacto, ordenable y cómodo para APIs.

Pero ULID codifica tiempo. Por eso no debe usarse cuando sea importante evitar inferir fecha, orden aproximado o secuencia de creación.

Usar según el caso:

```txt
uint interno:
- Todas las tablas principales.
- Relaciones internas.
- Joins y consultas eficientes.

ULID:
- Entidades públicas normales donde el orden temporal aproximado no sea sensible.
- users, companies, roles, permissions, products, orders, etc., si el proyecto acepta esa exposición temporal.

UUIDv4, nanoid o token random:
- Recursos sensibles.
- Invitaciones.
- Password reset.
- Verificación de email.
- Tokens públicos.
- Cualquier entidad donde no se quiera inferir tiempo, orden o volumen.
```

### Convención de rutas

Las rutas deben usar prefijo de versión y nombrarse como si trabajaran con `id`, pero ese `id` representa el `PublicID` externo:

```txt
GET /api/v1/users/:id
GET /api/v1/companies/:id
```

El versionado `/api/v1/` se define desde el primer endpoint. Agregar versión después cuando ya existen clientes consumiendo la API es un refactor costoso.

Internamente, los repositorios deben resolver `PublicID -> ID` cuando necesiten operar con relaciones o claves internas.

## Estrategia de producto: template primero, CLI después

NexoKit no debe empezar como un CLI complejo antes de tener una plantilla validada.

La estrategia recomendada es incremental:

```txt
v0.1: template clonable funcional + CLI interno mínimo.
v0.2: CLI interno para tareas del proyecto y generación de módulos.
v1.0: CLI instalable que genera proyectos nuevos con opciones interactivas.
```

### v0.1: template clonable funcional

El primer objetivo debe ser que un proyecto pueda iniciarse con:

```bash
git clone <repo> my-api
cd my-api
cp .env.example .env
docker compose up -d   # levanta PostgreSQL y Redis localmente
make dev
```

Esta versión debe traer una API base ejecutable, estructura modular, configuración, migraciones, health check, respuesta estándar y convenciones iniciales.

### v0.2: CLI interno

Una vez validada la plantilla, NexoKit debe incluir un CLI dentro del propio repositorio:

```bash
go run ./cmd/nexokit make module products
go run ./cmd/nexokit make migration create_products_table
go run ./cmd/nexokit create-root
```

Este CLI automatiza tareas repetitivas dentro de un proyecto ya creado.

### v1.0: CLI instalable

Cuando la plantilla esté estable, el CLI podrá instalarse globalmente:

```bash
go install github.com/<owner>/nexokit-go/cmd/nexokit@latest
nexokit new my-api
```

El comando `new` podrá preguntar qué incluir inicialmente:

```txt
Tipo de proyecto: API REST
Auth: sí/no
Multitenancy: sí/no
Rotación de logs: sí/no
Cache: none/redis
Docker Compose: sí/no
```

La plantilla debe diseñarse desde el inicio como si mañana fuera generada por CLI, evitando valores hardcodeados difíciles de reemplazar.

---

# Recomendación sobre los changes

Conviene separar NexoKit en varios changes SDD atómicos.

No conviene hacer un solo change gigante porque se mezclarían decisiones de arquitectura, auth, RBAC, multitenancy, cache, logging, testing y developer experience.

Changes recomendados:

1. Base del proyecto, configuración, GORM, migraciones y respuesta estándar.
2. CLI interno mínimo y developer experience para nuevos módulos.
3. Auth con PASETO, usuario root, usuarios, roles y refresh tokens.
4. RBAC, permisos y autorización.
5. Multitenancy por `company_id`.
6. Utilidades API: DTOs, validaciones, paginación, filtros y documentación de convenciones.
7a. Infraestructura de observabilidad: logger, log rotator, health checks y graceful shutdown.
7b. Infraestructura de resiliencia: cache y rate limit.
8. Testing, calidad y CI básico.

El Change 7 se divide en dos para evitar que una decisión sobre el logger bloquee el trabajo de cache o rate limit. Son responsabilidades independientes.

---

# Change 1: Base del proyecto, configuración, GORM, migraciones y respuesta estándar

> Prompt de implementación: [docs/prompts/change_01_base.md](prompts/change_01_base.md)

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
- Migraciones con Goose.
- Variables de entorno.
- Configuración centralizada.
- Docker Compose para desarrollo local.

## Alcance de este change

Implementar:

- Estructura base del proyecto.
- Carga de configuración desde `.env`.
- Configuración por ambiente: local, development, production, test.
- Conexión a PostgreSQL.
- Inicialización de GORM.
- Sistema de migraciones con Goose.
- Health check básico.
- Respuesta API estándar (`APIResponse`, errores, paginación).
- Manejo centralizado de errores (`AppError`, errores tipados).
- Logger inicial simple.
- Middleware base.
- CORS configurable.
- Request ID.
- Recovery middleware.
- Versionado de API `/api/v1/`.
- Convención de registro de rutas por módulo.
- Estructura para módulos.
- `docker-compose.yml` para desarrollo local con PostgreSQL.
- Documentación inicial del proyecto.

## Estructura de carpetas

```txt
nexokit-go/
  cmd/
    api/
      main.go
    nexokit/
      main.go
  internal/
    app/
      app.go
      bootstrap.go
      container.go
    config/
      config.go
      env.go
    infra/
      db/
        postgres.go
        migrations.go
      cache/
        redis.go
        noop.go
        cache.go
      logger/
        logger.go
        rotator.go
    server/
      server.go
      router.go
    middleware/
      auth.go
      tenant.go
      request_id.go
      recovery.go
      rate_limit.go
      logger.go
    platform/
      response/
      apperror/
      query/
      identity/
      password/
      token/
      validator/
    modules/
      auth/
      users/
      roles/
      companies/
    cli/
      commands/
      generator/
      templates/
    shared/
      model.go
  migrations/
  scripts/
  tests/
    integration/
    helpers/
    fixtures/
  docs/
  docker-compose.yml
  .env.example
  Makefile
  go.mod
  README.md
```

### Propósito de cada carpeta

| Carpeta | Propósito |
|---------|-----------|
| `cmd/api/` | Punto de entrada de la API. Debe ser delgado: cargar configuración, construir la app y arrancar el servidor. No debe contener lógica de negocio. |
| `cmd/nexokit/` | Punto de entrada del CLI de NexoKit. En v0.1/v0.2 sirve como CLI interno; en v1.0 podrá instalarse globalmente. |
| `internal/app/app.go` | Tipo `App` que agrupa el servidor, DB, logger y dependencias principales. Punto de arranque de la aplicación. |
| `internal/app/bootstrap.go` | Orquesta el orden de arranque: carga config, conecta DB, inicializa logger, monta router y levanta servidor. |
| `internal/app/container.go` | Construye y cablea todas las dependencias (repositorios, servicios, handlers). Separa el grafo de dependencias de la lógica de arranque. |
| `internal/config/` | Lectura, validación y normalización de configuración por ambiente. Expone una struct `Config` tipada, no strings sueltos. |
| `internal/infra/db/` | Conexión a PostgreSQL con GORM y helpers de migraciones. No debe contener queries de negocio. |
| `internal/infra/cache/` | Cliente Redis/Valkey, implementación `NoopCache` y la interfaz `Cache`. Aísla la dependencia externa del resto de la app. |
| `internal/infra/logger/` | Inicialización del logger estructurado (`slog`) y configuración de rotación de logs con `lumberjack`. |
| `internal/server/` | Configuración del servidor HTTP, router principal con versionado `/api/v1/` y montaje de rutas de módulos. |
| `internal/middleware/` | Middlewares HTTP concretos: auth, tenant, request ID, recovery, rate limit y logging de requests. |
| `internal/platform/response/` | Tipos y helpers para construir respuestas API estándar: `APIResponse`, errores, paginación y metadata. |
| `internal/platform/apperror/` | Errores tipados de aplicación (`ErrNotFound`, `ErrForbidden`, etc.) y conversión a respuestas HTTP. |
| `internal/platform/query/` | Parser y normalización de query params reutilizables: paginación, filtros, ordenamiento y búsqueda. |
| `internal/platform/identity/` | Generación de `PublicID`: ULID para entidades generales, UUIDv4/nanoid para recursos sensibles. |
| `internal/platform/password/` | Hash seguro de contraseñas con argon2id y verificación. |
| `internal/platform/token/` | Generación y validación de PASETO tokens. |
| `internal/platform/validator/` | Validador propio basado en reglas componibles (`Rule`, `FieldValidator`). No usa reflection ni tags. Las reglas se componen en funciones de validación por DTO. Incluye reglas base reutilizables y helper de integración con Gin. |
| `internal/modules/` | Dominio de negocio dividido por módulos. Cada módulo contiene handler, service, repository, DTOs, modelo, rutas y validaciones en estructura plana. |
| `internal/cli/` | Implementación interna del CLI: comandos, generadores y templates usados por `cmd/nexokit`. |
| `internal/shared/` | Código común mínimo y sin reglas de negocio. Contiene `BaseModel`, `BaseModelSimple` y tipos base realmente universales. Debe evitar convertirse en un cajón de sastre. |
| `migrations/` | Archivos SQL de migraciones usados por Goose. Nombrados con timestamp: `YYYYMMDDHHMMSS_descripcion.sql`. |
| `scripts/` | Automatizaciones auxiliares usadas por Makefile o CI (lint, build, seed, etc.). |
| `tests/integration/` | Integration tests que requieren DB o servicios externos. |
| `tests/helpers/` | Helpers reutilizables para tests: setup de app de prueba, autenticación de test, conexión a DB de test. |
| `tests/fixtures/` | Datos iniciales para tests: usuarios, roles, companies de prueba. |
| `docs/` | Documentación técnica, guías y SDD del proyecto. |

### Reglas de diseño de carpetas

- `infra/` agrupa lo que conecta con el mundo externo: DB, cache, logger a disco. Sus subpaquetes son importados por `app/` durante el bootstrap, no por los módulos directamente.
- `platform/` agrupa capacidades transversales de la aplicación que no son infraestructura de red. Sus subpaquetes son pequeños y enfocados.
- `shared/` debe ser mínimo. Si algo contiene reglas de negocio, pertenece a un módulo.
- Los módulos no deben importarse entre sí directamente salvo una decisión explícita; la coordinación debe pasar por contratos claros (ver sección de comunicación entre módulos).
- `cmd/*` no debe contener lógica de negocio.
- El CLI debe reutilizar templates y generadores, no strings hardcodeados dispersos.

### Convención de registro de rutas por módulo

Cada módulo debe exponer una función `Register` que recibe el grupo de rutas versionado y las dependencias necesarias:

```go
// internal/modules/users/routes.go
func Register(v1 *gin.RouterGroup, h *Handler) {
    users := v1.Group("/users")
    users.GET("", h.List)
    users.POST("", h.Create)
    users.GET("/:id", h.Get)
    users.PUT("/:id", h.Update)
    users.DELETE("/:id", h.Delete)
}
```

El router principal en `internal/server/router.go` monta todos los módulos:

```go
v1 := r.Group("/api/v1")
users.Register(v1, container.Users.Handler)
companies.Register(v1, container.Companies.Handler)
```

Esta convención debe estar documentada como estándar del framework. El CLI la usará para generar el archivo `routes.go` de cada módulo nuevo.

### Comunicación entre módulos

La regla general es evitar que un módulo consuma directamente el repository, modelo GORM o detalles internos de otro módulo.

Por ejemplo, si el módulo `sales` necesita datos del cliente, no debería importar `modules/customers/repository.go` ni manipular el modelo GORM `Customer` directamente.

En su lugar, el módulo dueño de los datos debe exponer un contrato pequeño orientado al caso de uso:

```go
// internal/modules/customers/contracts.go
type CustomerReader interface {
	FindSummaryByPublicID(ctx context.Context, publicID string) (CustomerSummary, error)
}

type CustomerSummary struct {
	ID       uint
	PublicID string
	Name     string
	Email    string
	Status   string
}
```

Luego `sales` depende de esa interfaz, no del repository ni del modelo completo:

```go
type Service struct {
	customers customers.CustomerReader
}
```

La implementación concreta puede vivir en `customers.Service` o en un adapter del módulo `customers`, y se inyecta desde `internal/app/container.go` durante el bootstrap.

Reglas prácticas:

```txt
Permitido:
- Importar contratos pequeños de otro módulo.
- Usar DTOs de lectura diseñados para integración interna.
- Inyectar dependencias desde app/container.

Evitar:
- Importar repositories de otro módulo.
- Importar modelos GORM de otro módulo para escribir queries externas.
- Hacer joins entre tablas de módulos distintos desde cualquier lugar sin una decisión explícita.
- Saltarse el service del módulo dueño de la regla de negocio.
```

Si un caso de uso requiere coordinar varios módulos, debe vivir en un servicio de aplicación/orquestación y no mezclar las reglas internas de los módulos.

## Estructura modular: plana

Cada módulo debe seguir esta estructura plana, con un archivo por responsabilidad dentro del mismo directorio:

```txt
internal/modules/users/
  handler.go
  service.go
  repository.go
  dto.go
  model.go
  routes.go
  validation.go
```

Esta estructura es la convención oficial de NexoKit. No se usan subdirectorios por capa dentro de un módulo. Si un módulo crece tanto que justifica subdirectorios, esa decisión se toma explícitamente para ese módulo puntual.

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

SHUTDOWN_TIMEOUT_SECONDS
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

## Validador

NexoKit usa un validador propio basado en reglas componibles en lugar de `go-playground/validator`.

### Por qué no go-playground/validator

`go-playground/validator` usa tags en structs y reflection. Funciona para casos simples, pero tiene limitaciones reales en un framework starter:

- Los errores que produce son difíciles de formatear consistentemente por campo sin una capa de adaptación.
- Las reglas condicionales (`Required_if`, `Required_unless`) se vuelven ilegibles en tags.
- No permite lógica de validación que dependa de múltiples campos sin custom validators registrados globalmente.
- En un starter que otros van a extender, inyectar un validador global con reflection es un punto de fricción.

El validador propio de NexoKit es más código inicial, pero es predecible, testeable sin setup y fácil de extender.

### Diseño

El validador vive en `internal/platform/validator/` y tiene tres archivos:

```txt
internal/platform/validator/
  validator.go    <- ValidationErrors, FieldValidator, Field()
  rules.go        <- reglas reutilizables (Rule)
  gin.go          <- helper de integración con Gin
```

### validator.go

```go
package validator

// ValidationErrors acumula errores por campo.
type ValidationErrors map[string][]string

func (ve ValidationErrors) Add(field, message string) {
    ve[field] = append(ve[field], message)
}

func (ve ValidationErrors) HasErrors() bool {
    return len(ve) > 0
}

// Rule es una función que recibe el valor del campo y retorna un mensaje de error o "".
type Rule func(value string) string

// FieldValidator encadena reglas sobre un campo.
type FieldValidator struct {
    field string
    value string
    skip  bool
    errs  ValidationErrors
}

// Field inicia la cadena de validación para un campo.
func Field(errs ValidationErrors, field, value string) *FieldValidator {
    return &FieldValidator{
        field: field,
        value: value,
        errs:  errs,
    }
}

// Required falla si el campo está vacío. Activa skip para no acumular errores sin sentido.
func (fv *FieldValidator) Required() *FieldValidator {
    if fv.value == "" {
        fv.errs.Add(fv.field, "es requerido")
        fv.skip = true
    }
    return fv
}

// Optional: si el campo está vacío, omite el resto de reglas.
func (fv *FieldValidator) Optional() *FieldValidator {
    if fv.value == "" {
        fv.skip = true
    }
    return fv
}

// Apply aplica una regla. Si skip está activo, no hace nada.
func (fv *FieldValidator) Apply(rule Rule) *FieldValidator {
    if fv.skip {
        return fv
    }
    if msg := rule(fv.value); msg != "" {
        fv.errs.Add(fv.field, msg)
    }
    return fv
}
```

### rules.go

```go
package validator

import (
    "fmt"
    "net/mail"
    "regexp"
    "strings"
    "unicode"
)

func MinLength(n int) Rule {
    return func(v string) string {
        if len([]rune(v)) < n {
            return fmt.Sprintf("debe tener al menos %d caracteres", n)
        }
        return ""
    }
}

func MaxLength(n int) Rule {
    return func(v string) string {
        if len([]rune(v)) > n {
            return fmt.Sprintf("no puede superar %d caracteres", n)
        }
        return ""
    }
}

func ValidEmail() Rule {
    return func(v string) string {
        _, err := mail.ParseAddress(v)
        if err != nil {
            return "debe ser un email válido"
        }
        return ""
    }
}

func HasUppercase() Rule {
    return func(v string) string {
        for _, ch := range v {
            if unicode.IsUpper(ch) {
                return ""
            }
        }
        return "debe contener al menos una mayúscula"
    }
}

func HasDigit() Rule {
    return func(v string) string {
        for _, ch := range v {
            if unicode.IsDigit(ch) {
                return ""
            }
        }
        return "debe contener al menos un número"
    }
}

func HasSpecialChar() Rule {
    return func(v string) string {
        for _, ch := range v {
            if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
                return ""
            }
        }
        return "debe contener al menos un carácter especial"
    }
}

func MinWords(n int) Rule {
    return func(v string) string {
        if len(strings.Fields(v)) < n {
            return fmt.Sprintf("debe contener al menos %d palabras", n)
        }
        return ""
    }
}

func NoNumbers() Rule {
    return func(v string) string {
        for _, ch := range v {
            if unicode.IsDigit(ch) {
                return "no debe contener números"
            }
        }
        return ""
    }
}

func Matches(pattern string) Rule {
    re := regexp.MustCompile(pattern)
    return func(v string) string {
        if !re.MatchString(v) {
            return "formato inválido"
        }
        return ""
    }
}
```

> `MinLength` y `MaxLength` usan `[]rune` en lugar de `len(v)` para manejar correctamente caracteres multibyte (tildes, ñ, etc.).

### gin.go

```go
package validator

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

// RespondIfInvalid escribe la respuesta de error de validación y retorna true si hay errores.
// Uso típico en handlers:
//
//   errs := ValidateCreateUser(req)
//   if validator.RespondIfInvalid(c, errs) {
//       return
//   }
func RespondIfInvalid(c *gin.Context, errs ValidationErrors) bool {
    if !errs.HasErrors() {
        return false
    }
    c.JSON(http.StatusUnprocessableEntity, gin.H{
        "success": false,
        "message": "Error de validación",
        "data":    nil,
        "meta":    nil,
        "errors":  errs,
    })
    return true
}
```

> Cuando `platform/response` esté implementado en el Change 1, este helper debe actualizarse para usar `response.ValidationError(c, errs)` en lugar del `gin.H` inline.

### Patrón de uso en handlers

Cada módulo define sus propias funciones de validación en `validation.go`. No hay validaciones globales registradas.

```go
// internal/modules/users/validation.go
package users

import "github.com/<owner>/nexokit-go/internal/platform/validator"

func ValidateCreateUser(req CreateUserRequest) validator.ValidationErrors {
    errs := make(validator.ValidationErrors)

    validator.Field(errs, "email", req.Email).
        Required().
        Apply(validator.ValidEmail())

    validator.Field(errs, "password", req.Password).
        Required().
        Apply(validator.MinLength(8)).
        Apply(validator.MaxLength(64)).
        Apply(validator.HasUppercase()).
        Apply(validator.HasDigit()).
        Apply(validator.HasSpecialChar())

    validator.Field(errs, "name", req.Name).
        Required().
        Apply(validator.MinWords(2)).
        Apply(validator.NoNumbers())

    return errs
}

func ValidateUpdateUser(req UpdateUserRequest) validator.ValidationErrors {
    errs := make(validator.ValidationErrors)

    validator.Field(errs, "name", req.Name).
        Optional().
        Apply(validator.MinWords(2)).
        Apply(validator.NoNumbers())

    validator.Field(errs, "bio", req.Bio).
        Optional().
        Apply(validator.MaxLength(300))

    return errs
}
```

Y en el handler:

```go
func (h *Handler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // error de binding (JSON malformado), no de validación
        response.BadRequest(c, "request inválido")
        return
    }

    errs := ValidateCreateUser(req)
    if validator.RespondIfInvalid(c, errs) {
        return
    }

    // continúa con lógica de negocio
}
```

### Tests del validador

Los tests viven junto al código en `internal/platform/validator/`:

```txt
internal/platform/validator/
  validator.go
  validator_test.go
  rules.go
  rules_test.go
  gin.go
```

Los tests del validador son pura lógica, sin servidor ni DB:

```go
func TestRequired(t *testing.T) {
    errs := make(ValidationErrors)
    Field(errs, "email", "").Required()

    if !errs.HasErrors() {
        t.Fatal("expected error for empty required field")
    }
    if errs["email"][0] != "es requerido" {
        t.Errorf("unexpected message: %s", errs["email"][0])
    }
}

func TestOptional_SkipsRules(t *testing.T) {
    errs := make(ValidationErrors)
    Field(errs, "bio", "").Optional().Apply(MinLength(10))

    if errs.HasErrors() {
        t.Fatal("expected no errors for empty optional field")
    }
}

func TestMinLength_Rune(t *testing.T) {
    rule := MinLength(3)
    // "añ" son 2 caracteres rune, no 3 bytes
    if msg := rule("añ"); msg == "" {
        t.Error("expected error for string shorter than 3 runes")
    }
    if msg := rule("año"); msg != "" {
        t.Errorf("unexpected error: %s", msg)
    }
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

NexoKit usa Goose porque es simple, cómodo y suficiente para una plantilla productiva.

Los archivos de migración viven en `migrations/` con el formato:

```txt
YYYYMMDDHHMMSS_descripcion.sql
```

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
5. Existe `docker-compose.yml` funcional con PostgreSQL.
6. Existe conexión funcional a PostgreSQL.
7. GORM queda configurado.
8. Existe sistema de migraciones con Goose.
9. Existe endpoint `/health` con respuesta estándar.
10. Todas las respuestas usan el formato estándar.
11. Los errores se responden de forma consistente.
12. Existe CORS configurable.
13. Existe Request ID.
14. Existe recovery middleware.
15. Las rutas usan prefijo `/api/v1/`.
16. Existe convención documentada de cómo registrar rutas por módulo.
17. Existe `internal/app/container.go` para el grafo de dependencias.
18. Existe una estructura modular base.
19. Existe README inicial con instrucciones para correr el proyecto.
20. Existen scripts o comandos Makefile para migraciones.

---

# Change 2: CLI interno y developer experience para nuevos módulos

> Prompt de implementación: [docs/prompts/change_02_cli.md](prompts/change_02_cli.md)

## Objetivo

Agregar herramientas de developer experience para que NexoKit sea realmente útil al trabajar dentro de proyectos creados desde la plantilla y al crear módulos repetibles.

Este change no debe tratarse como accesorio tardío. Pero en la primera versión el foco no será un CLI global tipo Laravel/Angular, sino un CLI interno mínimo que valide las convenciones de la plantilla clonable.

La estrategia es:

```txt
Primero: template clonable y ejecutable.
Después: CLI interno para tareas repetitivas.
Más adelante: CLI instalable con `nexokit new` interactivo.
```

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

### Opción B: CLI separado dentro del mismo repo

```txt
cmd/nexokit/main.go
```

Ejemplo:

```txt
go run ./cmd/nexokit make module products
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

- Usar Makefile para comandos de desarrollo frecuentes.
- Crear CLI mínimo dentro del proyecto (`cmd/nexokit`).
- Incluir desde temprano `serve`, `create-root`, `migrate` y `make module`.
- Evitar que la generación de módulos quede como mejora tardía, porque define la arquitectura real del framework.
- Dejar `nexokit new` interactivo para una versión posterior, cuando la plantilla ya esté validada.

Comandos esperados para la primera versión como CLI interno:

```bash
go run ./cmd/nexokit make module products
go run ./cmd/nexokit make migration create_products_table
go run ./cmd/nexokit create-root
```

Comandos esperados para una versión posterior instalable:

```bash
nexokit new app-name
nexokit new app-name --auth --tenant --cache=redis --log-rotation=false
```

## Generador de módulos

El generador de módulos debe crear una estructura plana consistente con la convención de NexoKit:

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

Y debe poder generar opcionalmente:

```txt
migrations/YYYYMMDDHHMMSS_create_products_table.sql
```

### Contrato del generador de módulos

El comando base será:

```bash
nexokit make module products
```

Flags sugeridos:

```bash
nexokit make module products --crud --migration --tenant
```

El flag `--permissions` queda fuera de la primera versión. La sincronización de permisos es una operación compleja que requiere introspección del sistema y se implementará cuando el CLI esté más maduro.

El generador debe poder crear:

```txt
Modelo con BaseModel.
DTOs de create, update, response y filtros.
Repository con búsqueda por PublicID.
Service con reglas de negocio mínimas.
Handler HTTP.
Routes del módulo con función Register.
Validaciones.
Migración SQL.
Scope de tenant si se usa --tenant.
```

El CLI no debe modificar lógica existente de forma silenciosa. Si necesita registrar rutas en archivos globales, debe hacerlo de forma explícita, idempotente y documentada.

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
10. Existe generador básico de módulos.
11. El generador puede crear módulo CRUD con migración opcional.
12. El generador usa `BaseModel` con `ID` interno y `PublicID` externo.
13. El generador produce la función `Register` de rutas estándar.
14. El documento explica que `nexokit new` global y `permissions sync` quedan para versiones posteriores.

---

# Change 3: Auth con PASETO, usuario root inicial, usuarios, roles y refresh tokens

> Prompt de implementación: [docs/prompts/change_03_auth.md](prompts/change_03_auth.md)

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

---

# Change 4: RBAC, permisos y autorización

> Prompt de implementación: [docs/prompts/change_04_rbac.md](prompts/change_04_rbac.md)

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

---

# Change 5: Multitenancy por company_id

> Prompt de implementación: [docs/prompts/change_05_multitenancy.md](prompts/change_05_multitenancy.md)

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
public_id
name
slug
domain nullable
subdomain nullable
status
created_at
updated_at
deleted_at nullable
created_by nullable
updated_by nullable
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
func WithCompany(db *gorm.DB, companyID uint) *gorm.DB
func ApplyTenantScope(db *gorm.DB, ctx TenantContext) *gorm.DB
```

Regla:

- Si el usuario es root y está en modo global, puede consultar todo.
- Si el usuario no es root, todas las consultas deben filtrar por `company_id`.

## Endpoints de companies

```txt
GET    /api/v1/companies
POST   /api/v1/companies
GET    /api/v1/companies/:id
PUT    /api/v1/companies/:id
DELETE /api/v1/companies/:id
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

# Change 6: Utilidades API: DTOs, validaciones, paginación, filtros y documentación de convenciones

> Prompt de implementación: [docs/prompts/change_06_utilities.md](prompts/change_06_utilities.md)

## Objetivo

Crear las utilidades reutilizables que harán que todos los módulos futuros de NexoKit tengan un comportamiento consistente.

Este change refina y documenta las utilidades base iniciadas en el Change 1. El Change 1 define las estructuras mínimas para arrancar; el Change 6 las completa, agrega filtros, helpers GORM y documenta las convenciones para que cualquier desarrollador pueda crear módulos nuevos de forma consistente.

## Alcance de este change

Implementar:

- DTOs base refinados.
- Validación de requests con go-playground/validator.
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

---

# Change 7a: Infraestructura de observabilidad: logger, log rotator, health checks y graceful shutdown

> Prompt de implementación: [docs/prompts/change_07a_observability.md](prompts/change_07a_observability.md)

## Objetivo

Implementar las funcionalidades de observabilidad y ciclo de vida del servidor que deben estar listas en NexoKit para cualquier proyecto.

Este change se separa del Change 7b (cache y rate limit) para evitar que decisiones sobre el logger bloqueen el trabajo de infraestructura de resiliencia. Son responsabilidades independientes.

## Alcance de este change

Implementar:

- Logger estructurado con `slog`.
- Middleware de request logging.
- Log rotator con `lumberjack`.
- Configuración de archivo de logs.
- Health check de base de datos.
- Health checks extendidos (`/health/live`, `/health/ready`).
- Graceful shutdown.
- Recovery middleware refinado.

## Logger

NexoKit usa `slog` (estándar de Go desde 1.21) para minimizar dependencias externas.

Si en un proyecto concreto se necesita máximo rendimiento o estructura avanzada, se puede reemplazar por `zap`, pero `slog` es el default del framework.

## Log rotation

Usar `lumberjack` para que funcione igual en cualquier entorno sin depender del sistema operativo.

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
- Cerrar cache si está activa.
- Respetar timeout.

Variables:

```txt
SHUTDOWN_TIMEOUT_SECONDS
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe logger estructurado con `slog`.
2. Los requests se registran con método, path, status y duración.
3. Los errores se registran con contexto.
4. Existe rotación de logs con `lumberjack`.
5. Los logs se guardan en archivo si está configurado.
6. Los logs pueden imprimirse en consola en desarrollo.
7. `/health/live` funciona.
8. `/health/ready` valida DB y cache.
9. Existe graceful shutdown.
10. Existe recovery middleware refinado.

---

# Change 7b: Infraestructura de resiliencia: cache y rate limit

> Prompt de implementación: [docs/prompts/change_07b_resilience.md](prompts/change_07b_resilience.md)

## Objetivo

Implementar cache y rate limiting como capacidades de resiliencia opcionales de NexoKit.

## Alcance de este change

Implementar:

- Cache adapter con interfaz.
- Cliente Redis/Valkey.
- `NoopCache` para proyectos sin cache.
- Rate limiter por IP.
- Rate limiter para endpoints sensibles (login, refresh).

## Cache

Soportar Redis o Valkey mediante `go-redis`, compatible con ambos.

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

Implementación:

- Memoria local por defecto (suficiente para instancia única).
- Redis si está configurado (para rate limit distribuido).

Variables:

```txt
RATE_LIMIT_ENABLED
RATE_LIMIT_REQUESTS
RATE_LIMIT_WINDOW_SECONDS
LOGIN_RATE_LIMIT_REQUESTS
LOGIN_RATE_LIMIT_WINDOW_SECONDS
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe cliente Redis/Valkey configurable.
2. Existe interfaz `Cache`.
3. Existe `RedisCache`.
4. Existe `NoopCache` si cache está deshabilitada.
5. Existe rate limit global opcional.
6. Existe rate limit específico para login.
7. Los endpoints rate-limited responden 429 cuando corresponde.
8. El `docker-compose.yml` incluye Redis para desarrollo local.

---

# Change 8: Testing, calidad y CI básico

> Prompt de implementación: [docs/prompts/change_08_testing.md](prompts/change_08_testing.md)

## Objetivo

Agregar una estrategia de testing y calidad para NexoKit, asegurando que las funcionalidades base del framework puedan verificarse automáticamente y que los proyectos creados con NexoKit tengan una base confiable de pruebas.

## Principio clave

En Go, los unit tests deben vivir junto al código que prueban, usando archivos `_test.go` dentro del mismo paquete.

La carpeta `/tests` debe reservarse principalmente para:

- Integration tests.
- Test helpers.
- Fixtures.
- Setup de base de datos de prueba.

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
- `testify` opcional para assertions cuando aporta claridad.
- PostgreSQL de test usando Docker Compose.
- Redis/Valkey de test usando Docker Compose.
- GitHub Actions para CI.

## Estructura recomendada

### Unit tests junto al código

```txt
internal/
  platform/response/
    response.go
    response_test.go

  modules/auth/
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

  platform/query/
    pagination.go
    pagination_test.go
    filters.go
    filters_test.go

  infra/cache/
    noop.go
    noop_test.go
    redis.go
    redis_test.go
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
PASETO generation/parsing
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

# Orden recomendado de implementación

## Orden recomendado para tu caso

Como NexoKit se usará para potenciar el desarrollo de la tienda SaaS, el orden recomendado es:

1. Change 1: Base técnica.
2. Change 2: CLI interno y developer experience mínimo.
3. Change 3: Auth con PASETO, root, usuarios y roles.
4. Change 5: Multitenancy.
5. Change 4: RBAC.
6. Change 6: DTOs, paginación, filtros y documentación.
7. Change 7a: Logger, rotación, health checks y graceful shutdown.
8. Change 7b: Cache y rate limit.
9. Change 8: Testing, calidad y CI básico.

La razón es que el CLI define temprano las convenciones del framework y evita escribir módulos a mano con estructuras inconsistentes. Después, para la tienda SaaS necesitas `company_id` muy pronto. RBAC puede quedar inmediatamente después.

## Orden ideal si se quiere máxima limpieza conceptual

1. Change 1: Base técnica.
2. Change 2: CLI interno y developer experience mínimo.
3. Change 3: Auth con PASETO, root, usuarios y roles.
4. Change 4: RBAC.
5. Change 5: Multitenancy.
6. Change 6: DTOs, paginación, filtros y documentación.
7. Change 7a: Logger, rotación, health checks y graceful shutdown.
8. Change 7b: Cache y rate limit.
9. Change 8: Testing, calidad y CI básico.

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
nexokit permissions sync (queda para cuando el CLI esté más maduro)
nexokit new interactivo (queda para v1.0)
```

Algunas de estas cosas pueden ser módulos opcionales después.

---

# Cosas que sí conviene adicionar desde el inicio

## IDs duales

Usar `uint` como ID interno y `PublicID` como identificador externo.

Reglas:

```txt
ID interno:
- uint.
- Primary key.
- Foreign keys.
- Nunca exponer por API.

PublicID externo:
- ULID por defecto.
- UUIDv4, nanoid o token random para recursos sensibles.
- Se expone como `id` en JSON.
```

## Auditoría básica

Incluir `created_by` y `updated_by` como nullable en `BaseModel` desde el inicio. Agregarlos después de que existan tablas creadas es un refactor costoso.

Los módulos que no necesiten auditoría pueden usar `BaseModelSimple` sin esos campos.

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

## Versionado de API

Usar `/api/v1/` desde el primer endpoint. Agregar versión después cuando ya existen clientes consumiendo la API es un refactor costoso.

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
- Migraciones con Goose
- Variables de entorno
- Respuesta API estandarizada
- Estructura modular plana
- Estrategia de IDs duales: `uint` interno + `PublicID` externo
- Auditoría básica: `created_by` / `updated_by` nullable en BaseModel
- Versionado de API: /api/v1/ desde el inicio
- Convención de registro de rutas: función Register por módulo
- CLI temprano para comandos base y generación de módulos
- Docker Compose para desarrollo local
- Preparado para auth, RBAC y multitenancy en changes posteriores

Necesito que analices este change y generes:

1. Recomendación técnica final.
2. Estructura de carpetas.
3. Convenciones del proyecto.
4. Configuración por ambiente.
5. Inicialización del servidor HTTP.
6. Conexión a PostgreSQL con GORM.
7. Sistema de migraciones con Goose.
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

1. ¿Confirmamos `uint` como ID interno y `PublicID` como ID externo en las tablas principales?
2. ¿Confirmamos Gin definitivamente?
3. ¿Confirmamos Goose para migraciones?
4. ¿Confirmamos PostgreSQL como única base soportada inicialmente?
5. ¿Redis/Valkey será opcional mediante `CACHE_DRIVER=none`?
6. ¿Confirmamos argon2id para passwords?
7. ¿Los permisos serán administrables desde API o solo seeds iniciales?
8. ¿Qué flags mínimos debe soportar `nexokit make module` en la primera versión?
9. ¿Quieres Docker Compose para desarrollo local desde el inicio? (recomendado: sí)
10. ¿Quieres usar `testify` o solo `testing` estándar?
11. ¿Quieres agregar GitHub Actions desde la primera versión?
