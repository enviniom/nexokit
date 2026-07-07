# Prompt de Instrucción: Change 18 - Vertical-Slice & Platform/Shared Review (module-by-module)

> [!IMPORTANT]
> **REGLA DE CONTEXTO GLOBAL**: Antes de ejecutar cualquier instrucción, lee y memoriza obligatoriamente la especificación oficial de base de datos en [openspec/database_schema.md](file:///home/enviniom/Proyectos/Go/nexokit/openspec/database_schema.md) y las reglas de programación globales en [openspec/core_context.md](file:///home/enviniom/Proyectos/Go/nexokit/openspec/core_context.md).

---

### Tipo de cambio
**SDD de capacidad transversal** (refactor iterativo, un módulo por sub-cambio).
Sigue el flujo completo: explore → propose → spec → design → tasks → apply → verify → archive — **aplicado módulo por módulo** dentro del mismo change, con un PR por módulo.

### ROL
Senior Architect & Go Specialist, foco en boundary discipline, convenciones y consistencia.

### CONTEXTO DE ENTRADA
- Proyecto: KatalejoShop SaaS.
- REGLAS DE CODIFICACIÓN: Sigue estrictamente [openspec/core_context.md](file:///home/enviniom/Proyectos/Go/nexokit/openspec/core_context.md).
- ESQUEMA DE BASE DE DATOS: Sigue [openspec/database_schema.md](file:///home/enviniom/Proyectos/Go/nexokit/openspec/database_schema.md).
- DOCUMENTACIÓN DE BACKEND OBLIGATORIA (es la base de toda la auditoría módulo por módulo). El entry point es [docs/modules.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules.md) y la fuente de verdad por área es:
  - [docs/modules/module-structure.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/module-structure.md) — forma estándar de módulo y responsabilidades por carpeta.
  - [docs/modules/vertical-slices.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/vertical-slices.md) — shape de slice, single-entity vs multi-entity, checklist.
  - [docs/modules/boundaries-and-dependencies.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/boundaries-and-dependencies.md) — reglas de coupling, `platform/shared`, wiring por app container.
  - [docs/modules/validation-and-errors.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/validation-and-errors.md) — `core/errors.go` con `apperror`, `response.HandleError`, envelope.
  - [docs/modules/queries-and-persistence.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/queries-and-persistence.md) — extracción a `queries/`, `TableName()` para partial GORM models.
  - [docs/modules/testing.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/testing.md) — cobertura mínima por tipo de archivo y table-driven tests.
  - Documento histórico: `docs/vertical-slice-modules.md` queda como **stub de compatibilidad**; no usarlo como referencia canónica, sólo como índice de la nueva estructura.
- DEPENDENCIAS: este change **consume** los outputs de **Change 16** (nuevo `AppError` con helpers y middleware) y se apoya en el desacoplamiento de **Change 17** para no re-mezclar uploads.

### OBJETIVO DEL CAMBIO
Revisar **módulo por módulo** y aplicar cuatro mejoras sistemáticas:

1. **Subir a `platform/shared` lo que se reutiliza entre módulos**: interfaces pequeñas, helpers puros, value objects transversales (p. ej. `Slugify`, normalización de teléfono, sanitización de strings para slugs públicos, helpers de paginación canónica).
2. **Migrar los errores de cada módulo al nuevo patrón** de **Change 16**: cada módulo declara sus errores reusables en `core/errors.go` usando los helpers de `apperror`; los services/repositories devuelven errores de `core`; los handlers delegan en `response.HandleError`.
3. **Documentar las convenciones de error por módulo** en `openspec/specs/<module>/errors.md` o en el spec general: tabla de errores, mapping a `Code` y `HTTPStatus`, ejemplos de uso.
4. **Eliminar duplicaciones**, en particular el `slugify` (movido a `platform/shared/slug` con tests), normalizadores de strings y cualquier helper de dominio repetido en más de un módulo.

### INSTRUCCIONES PASO A PASO

#### 1. Exploración previa (en `explore.md`)
- Inventariar módulos (`internal/modules/*`) y para cada uno:
  - Helpers en `core/` candidatos a `platform/shared` (usados por ≥ 2 módulos).
  - Errores definidos fuera de `core/errors.go` o construidos ad-hoc.
  - Duplicaciones detectadas (slugify, normalizadores, validadores).
  - Slice que rompe convención de `slices/` (entidad junto a `core/`, slice en raíz, etc.).
- Resultado: un plan de iteración **M1..MN** (un módulo por iteración) con criterios de aceptación.

#### 2. Convenciones a documentar (en `specs.md` y/o `docs/module-error-conventions.md`)
- Estructura de `core/errors.go` por módulo:

  ```go
  package core

  import "github.com/.../platform/apperror"

  var (
      ErrCustomerNotFound = apperror.NotFound("customer not found", nil)
      ErrEmailInUse       = apperror.Conflict("email already in use", nil)
  )
  ```

- Política de wrapping: services/repos pueden hacer `fmt.Errorf("get customer %d: %w", id, err)` para errores técnicos, pero **no** deben construir `apperror.*` ad-hoc.
- Handlers: deben evitar construir `apperror.*` ad-hoc; para errores de negocio usan los errores de `core` y llaman `response.HandleError(c, err)`.
- DTOs: `Validate()` devuelve `response.ValidationErrors` para errores por campo; el handler responde con `response.ValidationError(c, errs)` y HTTP 422.

#### 3. Iteración módulo por módulo (un PR por módulo)
Por cada módulo del plan M1..MN:

- Mover helpers reusables a `platform/shared/<helper>/<helper>.go` con tests propios.
- Reescribir `core/errors.go` con sentinels `apperror.*`.
- Sustituir cada construcción ad-hoc en services/repos/handlers.
- Ajustar tests: añadir un table-driven test por módulo que recorra cada sentinel y verifique `Status`, `Code` y `PublicMessage`.
- Eliminar duplicaciones (slugify primero, luego el resto).
- Validar que el módulo respeta [docs/modules.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules.md) y los doc específicos de cada área (ver CONTEXTO DE ENTRADA): slices bajo `slices/`, errores en `core/errors.go` con `apperror`, etc.

#### 4. PR por módulo
- Un branch y un PR por módulo para mantener review pequeño (alineado con el skill `chained-pr` cuando el change se vuelva grande).
- Cada PR referencia este change y enlaza al spec para que el revisor no tenga que reconstruir el contexto.

#### 5. Cierre del change
- Cuando todos los módulos pasaron, publicar `docs/module-error-conventions.md` con la tabla canónica "Error de `core` → `Code` → `HTTPStatus` → ejemplo de uso".
- Añadir un linter arquitectónico simple (regex/grep en CI o un `//go:build` check documentado) que prohíba `import "...platform/apperror"` desde `internal/modules/*/slices/**/service.go` y `.../repository.go`.

### CRITERIOS DE VERIFICACIÓN
- `go vet ./...`, `go build ./...`, `go test ./...` limpios por cada PR de módulo.
- Tabla final "errores por módulo" firmada en el spec, sin filas duplicadas.
- `grep -R "apperror\." api/internal/modules/ | grep -v _test.go` devuelve ocurrencias únicamente en `handler.go` y en `core/errors.go`; nunca en `service.go` ni `repository.go`.
- `slugify` existe en exactamente un sitio: `platform/shared/slug`, con tests.

### ENTREGABLES DEL CAMBIO
- Conventional commit por PR de módulo: `refactor(<module>): migrate errors to AppError, extract shared helpers, follow slice conventions`.
- Conventional commit汇总 al cierre: `docs(backend): publish module error conventions and platform/shared index`.
- Archivos a tocar: `api/internal/platform/shared/**` (nuevo/crecido), `api/internal/modules/*/core/errors.go`, handlers/services/repos de cada módulo, `docs/module-error-conventions.md` (nuevo), [docs/modules/validation-and-errors.md](file:///home/enviniom/Proyectos/Go/nexokit/docs/modules/validation-and-errors.md) (referencia canónica de mapping).
- Outputs SDD: un set de `proposal.md`, `specs.md`, `design.md`, `tasks.md` global + un `apply-report.md` por PR de módulo + `verify-report.md` final.

### NO-OBJETIVOS
- No introducir nuevos features de negocio; el change es de boundaries y consistencia.
- No romper el contrato HTTP: los `Code` y mensajes pueden evolucionar, pero los status code deben ser backwards-compatible (o documentar el breaking change en `verify-report.md`).
- No rehacer migraciones; los datos no cambian.
