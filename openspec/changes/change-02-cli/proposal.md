# Proposal: CLI interno y developer experience

## Intent

NexoKit necesita herramientas internas de developer experience para que los proyectos clonados desde la plantilla puedan ejecutar tareas repetitivas (migraciones, seeds, generación de módulos) sin depender de un CLI global instalable. Este change entrega un CLI interno mínimo (`cmd/nexokit`) y un Makefile funcional.

## Scope

### In Scope

- Makefile con comandos de desarrollo frecuentes.
- CLI interno en `cmd/nexokit` con subcomandos `serve`, `create-root`, `migrate`, `make module`, `make migration`, `make seed`, `status`, `config`.
- Generador de módulos con estructura plana NexoKit y migración opcional.
- Documentación de comandos disponibles.

### Out of Scope

- CLI global instalable (`nexokit new app-name`).
- Sincronización automática de permisos (`permissions sync`).

## Capabilities

### New Capabilities

- `cli-commands`: subcomandos del CLI interno (`serve`, `create-root`, `migrate`, `make`, `status`, `config`).
- `module-generator`: generación de módulos CRUD con estructura plana y migración SQL opcional.
- `dev-tooling`: Makefile con targets `dev`, `build`, `test`, `migrate-up`, `migrate-down`, `migrate-create`, `seed`, `create-root`, `lint`, `fmt`.

### Modified Capabilities

- None.

## Approach

Implementar un CLI interno con la biblioteca estándar `flag` o `cobra` (a definir en diseño) bajo `cmd/nexokit/main.go`. Los comandos `migrate` usarán Goose como librería. El generador de módulos usará plantillas Go (`text/template`) para producir archivos consistentes con las convenciones de NexoKit. El Makefile invocará comandos del CLI y de herramientas estándar de Go.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `cmd/nexokit/` | New | Punto de entrada del CLI interno. |
| `internal/cli/` | New | Comandos, generadores y plantillas del CLI. |
| `Makefile` | Modified | Nuevos targets de desarrollo y migraciones. |
| `docs/` | Modified | Documentación de comandos del CLI. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| El generador de módulos queda desfasado si cambian convenciones | Med | Plantillas versionadas y tests de golden file para el generador. |
| CLI interno crea confusión con futuro CLI global | Low | Documentar explícitamente que `cmd/nexokit` es interno al proyecto. |
| Dependencia circular al importar `internal/` desde `cmd/nexokit` | Low | `cmd/` solo importa `app/` y `config/`; lógica del CLI vive en `internal/cli/`. |

## Rollback Plan

Eliminar `cmd/nexokit/` y `internal/cli/`. Revertir `Makefile` a versión anterior. No hay cambios en base de datos ni en la API en ejecución.

## Dependencies

- Goose debe estar disponible como librería Go (ya en stack fijo).
- `internal/app/container.go` debe permitir inicializar dependencias sin levantar el servidor HTTP.

## Success Criteria

- [ ] Makefile funcional con al menos 10 targets.
- [ ] Comandos `serve`, `create-root`, `migrate up/down/create`, `make module`, `make migration`, `make seed`, `status`, `config` ejecutan sin errores.
- [ ] El generador crea módulos con estructura plana NexoKit y función `Register`.
- [ ] El generador soporta flags `--crud`, `--migration`, `--tenant`.
- [ ] Documentación lista qué comandos existen y cuáles quedan para versiones posteriores.
