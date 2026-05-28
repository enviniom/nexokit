# NexoKit — Contexto base (va en todos los prompts)

Este bloque debe pegarse al inicio de cada prompt de implementación. Contiene las decisiones que NO pueden contradecirse en ningún change.

---

## Identidad

NexoKit es un framework starter modular y opinionado en Go para construir APIs listas para SaaS.

Repositorio: `nexokit-go`

---

## Stack fijo

```txt
Lenguaje:       Go
Framework HTTP: Gin
ORM:            GORM
Base de datos:  PostgreSQL
Migraciones:    Goose
Auth:           PASETO v4.local (access) + refresh token opaco
Passwords:      argon2id
Cache:          Redis o Valkey (go-redis), opcional con CACHE_DRIVER=none
Logger:         slog (estándar Go 1.21+)
Log rotation:   lumberjack
Testing:        testing estándar + httptest + testify opcional
```

---

## Estructura de carpetas

```txt
nexokit-go/
  cmd/
    api/          <- punto de entrada de la API, sin lógica de negocio
    nexokit/      <- punto de entrada del CLI interno
  internal/
    app/
      app.go        <- tipo App con dependencias principales
      bootstrap.go  <- orden de arranque
      container.go  <- grafo de dependencias (repositorios, servicios, handlers)
    config/         <- struct Config tipada, carga desde .env
    infra/
      db/           <- conexión PostgreSQL + GORM + helpers de migraciones
      cache/        <- interfaz Cache, RedisCache, NoopCache
      logger/       <- slog + lumberjack
    server/
      server.go     <- servidor HTTP
      router.go     <- router principal, monta módulos con /api/v1/
    middleware/     <- auth, tenant, request_id, recovery, rate_limit, logger
    platform/
      response/     <- APIResponse, errores, paginación, helpers
      apperror/     <- errores tipados: ErrNotFound, ErrForbidden, etc.
      query/        <- parser de paginación, filtros, sorting
      identity/     <- generación de PublicID (ULID / UUIDv4)
      password/     <- hash argon2id y verificación
      token/        <- generación y validación de PASETO
      validator/    <- validador propio (Rule, FieldValidator, reglas base)
    modules/        <- módulos de negocio; módulos nuevos/no triviales usan vertical slice
    cli/            <- comandos, generadores y templates del CLI
    shared/
      model.go      <- BaseModel, BaseModelSimple
  migrations/       <- SQL con Goose, formato YYYYMMDDHHMMSS_descripcion.sql
  scripts/          <- automatizaciones para Makefile y CI
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

---

## Reglas de carpetas que no pueden romperse

- `infra/` conecta con el mundo externo (DB, cache, logger a disco). Los módulos no lo importan directamente; lo usa `app/`.
- `platform/` tiene subpaquetes pequeños y enfocados. No un paquete único gigante.
- `shared/` solo tiene `BaseModel`, `BaseModelSimple` y tipos base sin reglas de negocio.
- `cmd/*` nunca contiene lógica de negocio.
- Los módulos deben ser autocontenidos: no importan repositories ni modelos de otros módulos.
- Si un módulo necesita datos de una tabla relacionada, define su propio modelo local parcial con solo los campos que lee/escribe.

---

## Estrategia de IDs duales

```txt
ID interno:  uint autoincremental
             → primary key, foreign keys, joins
             → NUNCA exponer en API

PublicID:    string externo
             → ULID (char 26) para entidades normales
             → UUIDv4 / nanoid para recursos sensibles (tokens, invitaciones)
             → se expone como "id" en JSON
```

### BaseModel

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

`BaseModelSimple` es igual pero sin `CreatedBy` / `UpdatedBy`, para modelos que no necesiten auditoría.

---

## Estructura modular: vertical slice

Los módulos nuevos o no triviales usan vertical slice. Un endpoint existente equivale a un caso de uso y vive en su propia carpeta.

```txt
internal/modules/companies/
  container.go              <- composition root del módulo: wiring + registro de rutas
  routes.go                 <- rutas del módulo
  model.go                  <- aliases/compatibilidad si son necesarios
  dto.go                    <- aliases/compatibilidad si son necesarios
  resolver.go               <- comportamiento transversal del módulo, si aplica
  resolver_test.go
  routes_absence_test.go    <- tests transversales de superficie HTTP, si aplica
  core/
    model.go                <- modelos locales del módulo
    dto.go                  <- DTOs/contratos del módulo
    error.go                <- errores del módulo
  queries/
    get_company_by_public_id.go
    get_company_by_public_id_test.go
  view_company/
    handler.go
    handler_test.go
    service.go
    service_test.go
    repository.go
    repository_test.go
```

### Reglas de vertical slice

- Un endpoint existente = un caso de uso = un slice.
- Nombrar slices por intención de negocio: `view_company`, no `get_company`.
- No crear slices para endpoints que no existen.
- Cada slice debe tener `handler.go`, `service.go`, `repository.go` y su `_test.go` correspondiente.
- Si un `repository.go` solo wrappea una query reutilizable, igual debe tener `repository_test.go`; el test puede documentar que la lógica fuerte está cubierta en `queries/` y validar delegación/wiring.
- La raíz del módulo solo contiene archivos transversales o de compatibilidad.
- `container.go` del módulo es solo composition root: wiring y route registration. No tiene lógica de negocio ni funciona como service locator.
- El root container en `internal/app/container.go` solo llama al container del módulo; no conoce cada slice.

### `core/`

`core/` contiene elementos compartidos del módulo que no tienen lógica testeable directa:

- modelos locales del módulo;
- DTOs y contratos;
- enums/constants;
- errores;
- valores compartidos del módulo.

Los modelos en `core/` sirven para leer/escribir los campos que el módulo necesita. NO son la fuente del esquema completo de la base de datos; la fuente real del schema son las migraciones.

### `queries/`

`queries/` contiene lógica reusable de acceso a datos que se repite entre repositories de slices:

- una query por archivo cuando sea práctico;
- cada query debe tener su propio `_test.go`;
- los repositories de slices pueden delegar en `queries/`;
- `queries/` no contiene handlers, services, rutas ni casos de uso completos.

### Módulos simples o legacy

Un módulo simple existente puede permanecer plano si migrarlo no aporta valor inmediato. La migración a vertical slice debe hacerse módulo por módulo, normalmente cuando el módulo vaya a crecer o se toque de forma sustancial.

Estructura plana legacy aceptable para módulos simples:

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

---

## Convención de rutas

- Prefijo `/api/v1/` en todos los endpoints desde el inicio.
- Los parámetros de ruta usan `:id` que representa el `PublicID`.
- Cada módulo expone una función `Register`. En módulos vertical slice, `Register` recibe el container del módulo y registra handlers de slices:

```go
// internal/modules/companies/routes.go
func Register(v1 *gin.RouterGroup, c *Container, requireRole RoleMiddleware, requirePermission PermissionMiddleware) {
    g := v1.Group("/companies")
    g.GET("", requireRole("root"), c.ListCompanies.Handle)
    g.GET("/:id", requirePermission("companies.view"), c.ViewCompany.Handle)
}
```

En módulos planos legacy, `Register` puede seguir recibiendo un handler único:

```go
// internal/modules/users/routes.go
func Register(v1 *gin.RouterGroup, h *Handler) {
    g := v1.Group("/users")
    g.GET("", h.List)
    g.POST("", h.Create)
    g.GET("/:id", h.Get)
    g.PUT("/:id", h.Update)
    g.DELETE("/:id", h.Delete)
}
```

- El router/container raíz monta módulos, no slices individuales:

```go
v1 := r.Group("/api/v1")
users.Register(v1, container.Users.Handler)      // módulo plano legacy
companies.Register(v1, container.Companies, ...) // módulo vertical slice
```

---

## Comunicación entre módulos y autonomía

Los módulos NO se importan directamente entre sí. Tampoco deben usar repositories ni modelos GORM de otro módulo.

Preferencia actual: cada módulo debe ser autocontenido. Si necesita consultar una tabla que conceptualmente pertenece a otro módulo, define un modelo local parcial en su propio `core/` con los campos mínimos que necesita.

Ejemplo: `auth` no debe depender del repository de `users` para autenticar. Puede definir un modelo local parcial para la tabla `users` con email, password hash, estado y campos mínimos necesarios.

```go
// internal/modules/auth/core/model.go
type AuthUser struct {
    ID           uint   `gorm:"primaryKey"`
    PublicID     string `gorm:"column:public_id"`
    Email        string `gorm:"column:email"`
    PasswordHash string `gorm:"column:password_hash"`
    Status       string `gorm:"column:status"`
}

func (AuthUser) TableName() string { return "users" }
```

Cuando una dependencia externa sea inevitable, usar contratos pequeños inyectados desde `internal/app/container.go`, nunca repositories concretos:

```go
// internal/modules/customers/contracts.go
type CustomerReader interface {
    FindSummaryByPublicID(ctx context.Context, publicID string) (CustomerSummary, error)
}
```

La dependencia se inyecta desde `internal/app/container.go`.

```txt
Preferido:   modelo local parcial en core/ + repository propio del módulo
Permitido:   contrato pequeño inyectado desde app/container.go cuando haga falta coordinación
Prohibido:   importar repositories o modelos GORM de otro módulo
```

---

## Respuesta API estándar

```json
{
  "success": true | false,
  "message": "string",
  "data": {} | [] | null,
  "meta": null | { "pagination": {...}, "filters": {...} },
  "errors": null | { "field": ["mensaje"] }
}
```

Todos los handlers deben usar `platform/response` para construir respuestas. Nunca `gin.H` directo.

---

## Validador propio

NexoKit NO usa `go-playground/validator`. Usa un validador propio en `internal/platform/validator/`:

```go
// Patrón de uso en validation.go de cada módulo
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

    return errs
}
```

Y en el handler:

```go
errs := ValidateCreateUser(req)
if validator.RespondIfInvalid(c, errs) {
    return
}
```

---

## Roles y autorización

- Un usuario tiene UN SOLO rol.
- Roles base: `root`, `admin`, `user`.
- Autorización por permisos (`users.create`, `products.read`, etc.), no por nombre de rol.
- Middleware: `AuthMiddleware()`, `RequirePermission("slug")`, `RequireRole("slug")`.
- Root tiene todos los permisos.
- La autenticación y la autorización son middlewares separados.

---

## Multitenancy

- Aislamiento por `company_id`.
- Root puede operar globalmente (company_id nullable).
- Admin y user siempre tienen company_id.
- Todas las queries de módulos tenant deben filtrar por company_id.
- GORM helpers: `WithCompany(db, id)`, `ApplyTenantScope(db, ctx)`.

---

## Reglas que nunca deben violarse

1. No exponer `ID` interno (uint) en respuestas API.
2. No hardcodear contraseñas, keys o secrets en el código.
3. No devolver `password_hash` en respuestas.
4. No revelar si falló el email o la contraseña por separado en login.
5. No importar módulos entre sí salvo contratos pequeños y explícitos.
6. No usar `gin.H{}` para respuestas; usar `platform/response`.
7. No importar repositories ni modelos GORM de otro módulo.
8. No agregar lógica de negocio en `cmd/`, `infra/`, `shared/`, `core/` ni `container.go`.
9. No poner queries reutilizables en `core/`; deben vivir en `queries/` con tests.
10. No crear slices para endpoints que no existen.
