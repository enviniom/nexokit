# NexoKit — Contexto base (obligatorio en todos los prompts)

Este contexto define reglas NO negociables para cualquier change SDD/OpenSpec.

## 1) Identidad y stack fijo

- NexoKit: starter modular y opinionado en Go para APIs SaaS.
- Stack fijo: Go, Gin, GORM, PostgreSQL, Goose, PASETO v4.local + refresh token opaco, argon2id, Redis/Valkey opcional, slog + lumberjack, testing estándar + `httptest` (+ `testify` opcional).

## 2) Fronteras de carpetas

- `cmd/*`: solo entrypoints (sin negocio).
- `internal/infra/*`: integración externa (DB, cache, logger).
- `internal/shared/*`: **mínimo** y sin negocio (tipos/base models).
- `internal/modules/*`: dominio y casos de uso.
- `internal/platform/*`: contratos/utilidades **cross-application**.

## 3) Límite estricto de `platform`

`platform` existe para piezas genéricas reutilizables por múltiples módulos, por ejemplo:

- `response`
- `apperror`
- `validator`
- `query`
- `tenant`
- `authctx`
- `password`
- `token`
- `identity`
- mensajes genéricos de API/validación/middleware

Reglas:

1. Mensajes, errores y constantes de **dominio** viven en su módulo dueño, no en `platform`.
2. `platform/response` define el contrato JSON estándar y helpers de respuesta; **no duplicar** formato de respuesta por módulo.
3. Si una constante/error no aplica a más de un dominio, no pertenece a `platform`.

## 4) Responsabilidad del dominio por módulo

Cada módulo debe centralizar su lenguaje de dominio en `core/`:

- `core/error.go`
- `core/constants.go` (Go no tiene enums reales; usar typed constants cuando se necesite type-safety)
- `core/dto.go`
- `core/model.go`
- `core/contracts.go`

## 5) Autonomía entre módulos (obligatorio)

1. Un módulo **no importa** otro módulo.
2. Un módulo no usa repositories/modelos GORM de otro módulo.
3. Preferir modelos locales parciales en `core/model.go` antes que dependencia cross-module.
4. Si se requiere capacidad externa, el módulo define contrato consumidor propio (`core/contracts.go`) y se inyecta desde el root app container.
5. `internal/app/container.go` wirea módulos; no conoce slices internos.
6. `internal/modules/<mod>/container.go` importa solo slices de su módulo y es composition root puro (wiring + registro de rutas, sin negocio).

## 6) IDs duales (obligatorio)

- `ID` interno (`uint`) para PK/FK/joins: nunca exponer en API.
- `PublicID` (`string`) expuesto como `id`:
  - ULID (26 chars) para entidades normales.
  - UUIDv4/nanoid para recursos sensibles.

## 7) Contrato de respuesta API (obligatorio)

```json
{
  "success": true,
  "message": "string",
  "data": {},
  "meta": null,
  "errors": null
}
```

- Todos los handlers responden vía `platform/response`.
- Prohibido responder con `gin.H` directo para contrato estándar.

## 8) Validador

- No usar `go-playground/validator`.
- Usar `internal/platform/validator` + `validator.RespondIfInvalid(...)`.

## 9) Auth, autorización y tenant

- Un usuario tiene un rol (`root`, `admin`, `user`).
- Autorización por permisos (`users.create`, etc.), no por nombre de rol.
- `AuthMiddleware`, `RequirePermission`, `RequireRole` separados.
- Aislamiento tenant por `company_id`; root global, admin/user tenant-scoped.

## 10) Vertical slice (cuando aplique)

- Un endpoint existente = un caso de uso = un slice.
- Nombrar por intención de negocio (`view_company`, no `get_company`).
- Cada slice con `handler/service/repository` + tests.
- `queries/` para acceso reusable a datos (con tests propios).
- No crear slices para endpoints inexistentes.

## 11) Prohibiciones duras

1. No exponer `ID` interno en API.
2. No filtrar `password_hash` ni secretos.
3. No revelar qué parte del login falló (email vs password).
4. No lógica de negocio en `cmd`, `infra`, `shared`, ni containers.
5. No dependencias cross-module directas.
6. No duplicar contrato de respuesta fuera de `platform/response`.
