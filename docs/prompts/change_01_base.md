> Lee también `_context.md` antes de implementar este change.

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
