> Lee también `_context.md` antes de implementar este change.

Quiero iniciar un nuevo change SDD/OpenSpec para migrar el módulo `permissions` a arquitectura vertical slice.
Change: `18-vertical-slice-permissions`
Objetivo:
Refactorizar únicamente `internal/modules/permissions/` hacia vertical slice, manteniendo el módulo como frontera principal y haciéndolo más autocontenido. No migrar otros módulos en este change.
Reglas de arquitectura:
- Un endpoint existente = un caso de uso = un slice.
- Cada slice debe tener su propia carpeta con:
  - handler.go + handler_test.go
  - service.go + service_test.go
  - repository.go + repository_test.go
- Usar nombres de slices por intención de negocio, no por detalle técnico. Ejemplo: `view_company`, no `get_company`.
- No crear slices para endpoints que no existen actualmente.
Reglas de módulo autocontenido:
- El módulo NO debe depender de repositories de otros módulos.
- Si necesita leer/escribir datos de una tabla relacionada con otro módulo, debe definir su propio modelo local con solo los campos que necesita.
- Está permitido repetir modelos parciales entre módulos si eso evita acoplamiento.
- Los modelos Go del módulo NO son la fuente del esquema real de la base de datos.
- La fuente del esquema real son las migraciones.
- Los modelos del módulo sirven para mapear/leer/escribir los campos que ese módulo necesita.
- Ejemplo: `auth` no debería depender del repository de `users`; puede tener su propio modelo parcial de user con solo los campos necesarios para autenticar.
- Esta regla busca evitar dependencias cruzadas entre módulos y mantener cada módulo como unidad autónoma.
Estructura esperada:
- La raíz del módulo debe conservar solo archivos transversales:
  - container.go
  - routes.go
  - archivos de compatibilidad si son necesarios
  - tests transversales del módulo si protegen comportamiento de rutas/resolver/migración
- Crear `core/` para elementos compartidos sin lógica testeable directa:
  - modelos
  - DTOs/contratos
  - enums/constants
  - errors
  - valores compartidos del módulo
- Crear `queries/` para queries reutilizables que se repiten entre repositories:
  - una query por archivo cuando sea práctico
  - cada query debe tener su propio `_test.go`
- Los repositories de slices pueden delegar en `queries/`.
- Si un repository de slice solo wrappea una query, igual debe tener `repository_test.go`, documentando que la lógica fuerte se cubre en `queries/` y que el test del slice valida delegación/wiring.
- El root app container solo debe llamar al container del módulo.
- El container del módulo debe ser solo composition root: wiring y registro de rutas; sin lógica de negocio y sin service locator.
- Preservar comportamiento actual. Este change es refactor arquitectónico, no cambio funcional.
- Ejecutar tests del módulo y `go test ./...`.
Antes de implementar:
1. Explorar el módulo actual.
2. Proponer el mapeo endpoint → slice.
3. Detectar dependencias hacia otros módulos y proponer cómo eliminarlas con modelos locales parciales.
4. Detectar queries duplicadas candidatas para `queries/`.
5. Proponer riesgos y plan de migración.
6. Crear proposal, specs, design y tasks.
7. Esperar mi revisión antes de aplicar.
Entrega esperada:
- Un solo módulo migrado por change.
- El módulo queda autocontenido y sin dependencias a repositories de otros módulos.
- Specs/docs actualizadas con la realidad final.
- No hacer commit hasta que yo revise el resultado.