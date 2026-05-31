> Lee también `_context.md` antes de implementar este change.

Quiero iniciar un nuevo change SDD/OpenSpec para revisar frontera arquitectónica de ownership de errores de validación, sin cambio de comportamiento.

Change: `20-validation-errors-boundary`

## Objetivo

Evaluar y refactorizar (si corresponde) la propiedad de `ValidationErrors` para que no dependa conceptualmente de `platform/response` cuando su ubicación natural sea `platform/validator`, preservando contrato API y comportamiento de handlers.

## Scope

- Analizar dónde vive hoy `ValidationErrors` y sus dependencias.
- Evaluar alternativas de ownership (`platform/response` vs `platform/validator`) con tradeoffs.
- Si aplica, mover/tipar `ValidationErrors` en `platform/validator`.
- Ajustar imports/adaptadores necesarios sin cambiar shape JSON ni semántica HTTP.
- Documentar la decisión final y reglas para futuros cambios.

## Out of scope (explícito)

- Cambiar contrato JSON de `APIResponse`.
- Cambiar comportamiento funcional de validaciones de negocio.
- Refactors no relacionados con ownership de errores de validación.

## Antes de implementar

1. Trazar flujo completo de validación: validator → handler → response.
2. Definir criterio arquitectónico de ownership (conceptual y de dependencia).
3. Validar que la opción elegida no genere acoplamiento inverso ni imports cíclicos.
4. Proponer migración mínima con impacto acotado.
5. Crear `proposal`, `spec`, `design` y `tasks`.
6. Esperar revisión antes de aplicar.

## Criterios de aceptación

- Ownership de `ValidationErrors` queda explícito y coherente con fronteras de plataforma.
- Si hay refactor, handlers mantienen el mismo comportamiento observable.
- Se preserva exactamente el shape JSON de respuestas API.
- No se introduce duplicación de contrato de respuesta por módulo.
- Resultado documentado para evitar regresiones arquitectónicas.
