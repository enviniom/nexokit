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
    modules/        <- módulos de negocio en estructura plana
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
- Los módulos no se importan entre sí directamente. Usan contratos (interfaces pequeñas) inyectados desde `container.go`.

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

## Estructura modular: plana

Cada módulo usa un archivo por responsabilidad, sin subdirectorios por capa:

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
- Cada módulo expone una función `Register`:

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

- El router en `server/router.go` monta todos los módulos:

```go
v1 := r.Group("/api/v1")
users.Register(v1, container.Users.Handler)
companies.Register(v1, container.Companies.Handler)
```

---

## Comunicación entre módulos

Los módulos NO se importan directamente entre sí. El módulo dueño de los datos expone un contrato:

```go
// internal/modules/customers/contracts.go
type CustomerReader interface {
    FindSummaryByPublicID(ctx context.Context, publicID string) (CustomerSummary, error)
}
```

La dependencia se inyecta desde `internal/app/container.go`.

```txt
Permitido:   importar contratos (interfaces) de otro módulo
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
5. No importar repositories de otro módulo desde un módulo externo.
6. No usar `gin.H{}` para respuestas; usar `platform/response`.
7. No crear subdirectorios dentro de un módulo salvo decisión explícita.
8. No agregar lógica de negocio en `cmd/`, `infra/` ni `shared/`.
