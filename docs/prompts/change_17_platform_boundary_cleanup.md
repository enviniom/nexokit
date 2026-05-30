> Lee también `_context.md` antes de implementar este change.

Quiero iniciar un nuevo change SDD/OpenSpec para reforzar fronteras de `platform` sin cambio funcional.

Change: `17-platform-boundary-cleanup`

## Objetivo

Limpiar y endurecer los límites de `internal/platform/*` para que solo contenga contratos/utilidades cross-application, moviendo lenguaje de dominio (mensajes/errores/constantes) a sus módulos dueños.

## Scope

- Auditar `internal/platform/*` y detectar elementos con semántica de dominio.
- Mover mensajes/errores/constantes de dominio a `internal/modules/<mod>/core/*`.
- Definir explícitamente qué sí puede vivir en `platform` (response, apperror, validator, query, tenant, authctx, password, token, identity, mensajes genéricos API/validación/middleware).
- Mantener un único contrato de respuesta en `platform/response` (sin duplicaciones por módulo).
- Actualizar documentación/arquitectura del change con decisiones y reglas resultantes.

## Out of scope (explícito)

- Cambios funcionales de handlers, servicios o reglas de negocio.
- Cambios en shape JSON de respuestas API.
- Reescritura masiva de módulos no impactados.
- Cualquier cambio en `internal/modules/permissions/**` salvo aprobación explícita.

## Antes de implementar

1. Mapear contenidos actuales de `platform` en: `genérico` vs `dominio`.
2. Proponer destino módulo-por-módulo para cada elemento de dominio.
3. Identificar riesgos de imports circulares o acoplamientos nuevos.
4. Proponer plan incremental sin romper contrato API.
5. Crear `proposal`, `spec`, `design` y `tasks`.
6. Esperar revisión antes de aplicar.

## Criterios de aceptación

- `platform` queda acotado a responsabilidades cross-application.
- Mensajes/errores/constantes de dominio quedan en `modules/<mod>/core/*`.
- `platform/response` sigue siendo única fuente del contrato de respuesta.
- No hay cambio funcional observable (mismo comportamiento HTTP).
- No se modifica `internal/modules/permissions/**`; si aparece dependencia, documentar como follow-up o riesgo.
