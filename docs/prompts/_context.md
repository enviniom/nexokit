# NexoKit — Contexto base obligatorio

Este archivo contiene las decisiones que **no pueden contradecirse** en ningún change SDD/OpenSpec. Debe leerse antes de proponer, diseñar o implementar.

## 1) Identidad y stack fijo

NexoKit es un starter modular y opinionado en Go para construir APIs SaaS listas para producción.

| Área | Decisión |
| --- | --- |
| Lenguaje | Go |
| HTTP | Gin |
| ORM | GORM |
| Base de datos | PostgreSQL |
| Migraciones | Goose (`migrations/`, formato `YYYYMMDDHHMMSS_descripcion.sql`) |
| Auth | PASETO v4.local para access token + refresh token opaco |
| Passwords | argon2id |
| Cache | Redis o Valkey con `go-redis`; opcional con `CACHE_DRIVER=none` |
| Logging | `slog` + `lumberjack` |
| Testing | `testing` estándar + `httptest`; `testify` opcional |

## 2) Estructura base del repo

| Ruta | Responsabilidad |
| --- | --- |
| `cmd/api/` | Entry point de API; sin lógica de negocio. |
| `cmd/nexokit/` | Entry point del CLI interno. |
| `internal/app/` | Bootstrap y grafo de dependencias. El root container monta módulos, no slices. |
| `internal/config/` | Config tipada desde `.env`. |
| `internal/infra/` | Integraciones externas: DB, cache, logger a disco. Los módulos no importan `infra`. |
| `internal/server/` | Servidor HTTP y router raíz. |
| `internal/middleware/` | Auth, tenant, request ID, recovery, rate limit, logger. |
| `internal/platform/` | Contratos/utilidades transversales de aplicación. Ver sección 3. |
| `internal/modules/` | Módulos de negocio; módulos nuevos/no triviales usan vertical slice. |
| `internal/shared/` | Mínimo: `BaseModel`, `BaseModelSimple` y tipos base sin reglas de negocio. |
| `migrations/` | Fuente real del schema de base de datos. |
| `tests/` | Helpers, fixtures e integración. |

## 3) Frontera estricta de `platform`

`platform` existe para contratos/utilidades **cross-application**, no para lenguaje de dominio.

Permitido en `platform`:

- `response`: contrato JSON estándar y helpers de respuesta.
- `apperror`: categorías/errores transversales.
- `validator`: validador propio y reglas comunes.
- `query`: paginación, filtros y sorting.
- `tenant`, `authctx`: contexto transversal de request.
- `password`, `token`, `identity`: primitivas técnicas compartidas.
- mensajes genéricos de API, validación y middleware.

No va en `platform`:

- errores, mensajes o constantes propios de un dominio;
- reglas de negocio;
- repositories, modelos GORM de módulos o helpers específicos de un módulo.

Regla guía:

> Si cambia el contrato global de la API, va en `platform`. Si cambia el lenguaje de un módulo, va en el módulo.

## 4) Reglas por carpeta que no se rompen

- `cmd/*`, `infra/*`, `shared/*`, `core/*` y `container.go` no contienen lógica de negocio.
- `infra/*` conecta con el mundo externo; lo usa `app`, no los módulos.
- `shared/*` se mantiene chico; no es un cajón de utilidades.
- `platform/*` debe tener subpaquetes enfocados; evitar un paquete gigante.
- Los modelos Go de módulos no son la fuente del schema real; el schema real son las migraciones.
- Los handlers construyen respuestas con `platform/response`; prohibido usar `gin.H` directo para respuestas estándar.

## 5) Autonomía entre módulos

- Un módulo **no importa** otro módulo.
- Un módulo no usa repositories ni modelos GORM de otro módulo.
- Si necesita leer/escribir una tabla relacionada con otro módulo, define un modelo local parcial en `core/model.go` con solo los campos necesarios.
- Repetir modelos parciales entre módulos está permitido si evita acoplamiento.
- Si necesita una capacidad externa, el módulo declara un contrato consumidor pequeño en `core/contracts.go` y `internal/app/container.go` inyecta la implementación.
- El root container solo llama al container del módulo; no importa slices internos.
- El container del módulo puede importar slices de su módulo y debe ser composition root puro: wiring y registro, sin negocio ni service locator.

## 6) Lenguaje de dominio dentro del módulo

En módulos vertical slice, `core/` centraliza elementos compartidos sin lógica testeable directa:

```txt
core/
  model.go       <- modelos locales/parciales del módulo
  dto.go         <- DTOs y contratos de payload/respuesta
  error.go       <- errores de negocio del módulo
  constants.go   <- constantes del dominio
  contracts.go   <- interfaces consumidoras pequeñas
```

Go no tiene enums reales. Usar `constants.go`; si se necesita type-safety, usar constantes tipadas:

```go
type DomainKind string

const DomainKindPrimary DomainKind = "primary"
```

## 7) IDs duales

- `ID uint`: PK/FK/joins internos; **nunca** se expone en API.
- `PublicID string`: identificador externo expuesto como `id`.
  - ULID de 26 chars para entidades normales.
  - UUIDv4/nanoid para recursos sensibles como tokens o invitaciones.
- `shared.BaseModel` incluye `ID`, `PublicID`, timestamps, soft delete y auditoría.
- `shared.BaseModelSimple` omite auditoría cuando no aplica.

## 8) Vertical slice

Los módulos nuevos o no triviales usan vertical slice. La migración se hace módulo por módulo.

```txt
internal/modules/<module>/
  container.go
  routes.go
  core/
  queries/
  <business_intent_slice>/
    handler.go
    handler_test.go
    service.go
    service_test.go
    repository.go
    repository_test.go
```

Reglas:

- Un endpoint existente = un caso de uso = un slice.
- Nombrar slices por intención de negocio: `view_company`, no `get_company`.
- No crear slices para endpoints que no existen.
- Cada slice tiene `handler`, `service`, `repository` y tests propios.
- La raíz del módulo conserva solo archivos transversales: `container.go`, `routes.go`, compatibilidad necesaria y tests transversales.
- `queries/` contiene queries reutilizables, una por archivo cuando sea práctico, con `_test.go` propio.
- Los repositories de slices pueden delegar en `queries/`.
- Si un repository solo wrappea una query, igual debe tener `repository_test.go` para validar delegación/wiring.

Módulos simples legacy pueden permanecer planos si migrarlos no aporta valor inmediato.

### Módulos multi-entidad

Si un módulo maneja más de una entidad y cada entidad tiene más de 3 casos de uso, organizar slices por carpeta de entidad:

```txt
internal/modules/<module>/
  container.go
  routes.go
  core/
  queries/
  <entity_a>/
    container.go
    routes.go
    <business_intent_slice>/
      handler.go
      handler_test.go
      service.go
      service_test.go
      repository.go
      repository_test.go
  <entity_b>/
    container.go
    routes.go
    <business_intent_slice>/
```

Reglas:

- Las carpetas de entidad no son módulos nuevos; siguen dentro de la frontera del módulo padre.
- El container raíz del módulo exporta explícitamente los containers de entidad:

```go
type Container struct {
    Products   *products.Container
    Categories *categories.Container
}
```

- `internal/app/container.go` solo conoce el container raíz del módulo.
- `routes.go` raíz del módulo delega en los `routes.go` de cada entidad usando esos campos explícitos.
- No usar mapas de handlers ni service locator para ocultar containers internos.

## 9) Rutas y montaje HTTP

- Todos los endpoints usan prefijo `/api/v1/` desde el router raíz.
- Parámetros `:id` representan `PublicID`, no `ID` interno.
- Cada módulo expone `Register`.
- En módulos vertical slice, `Register` recibe el container del módulo y registra handlers de slices.
- En módulos planos legacy, `Register` puede recibir un handler único.
- El router/root container monta módulos completos; nunca slices individuales.

## 10) Respuesta API y validación

Contrato estándar:

```json
{
  "success": true,
  "message": "string",
  "data": {},
  "meta": null,
  "errors": null
}
```

- Listados paginados usan `meta.pagination` y `meta.filters`.
- Errores de validación usan `errors` por campo: `{ "field": ["mensaje"] }`.
- No usar `go-playground/validator`; usar `internal/platform/validator`.
- El formato de respuesta no se duplica por módulo; vive en `platform/response`.

## 11) Auth, autorización y multitenancy

- Un usuario tiene un solo rol.
- Roles base: `root`, `admin`, `user`.
- Autorización por permisos (`users.create`, `roles.view`, etc.), no por nombre de rol.
- Autenticación y autorización son middlewares separados.
- Root tiene todos los permisos y puede operar globalmente.
- Admin y user siempre operan con `company_id`.
- Queries tenant-scoped filtran por `company_id`.
- Usar helpers de tenant cuando aplique (`WithCompany`, `ApplyTenantScope`, contexto tenant).

## 12) Prohibiciones duras

1. No exponer `ID` interno en API.
2. No hardcodear contraseñas, keys o secrets.
3. No devolver `password_hash`.
4. No revelar si falló email o password por separado en login.
5. No importar módulos entre sí.
6. No importar repositories ni modelos GORM de otro módulo.
7. No usar `gin.H` para respuestas estándar.
8. No lógica de negocio en `cmd`, `infra`, `shared`, `core` ni containers.
9. No poner queries reutilizables en `core`; van en `queries/` con tests.
10. No crear slices para endpoints inexistentes.
