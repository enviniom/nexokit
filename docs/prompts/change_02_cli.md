> Lee también `_context.md` antes de implementar este change.

# Change 2: CLI interno y developer experience para nuevos módulos

## Objetivo

Agregar herramientas de developer experience para que NexoKit sea realmente útil al trabajar dentro de proyectos creados desde la plantilla y al crear módulos repetibles.

Este change no debe tratarse como accesorio tardío. Pero en la primera versión el foco no será un CLI global tipo Laravel/Angular, sino un CLI interno mínimo que valide las convenciones de la plantilla clonable.

La estrategia es:

```txt
Primero: template clonable y ejecutable.
Después: CLI interno para tareas repetitivas.
Más adelante: CLI instalable con `nexokit new` interactivo.
```

## Alcance

Implementar un CLI o comandos simples para:

- Crear usuario root.
- Ejecutar migraciones.
- Revertir migraciones.
- Crear migración.
- Crear módulo base.
- Crear seed.
- Ver configuración actual.
- Ver estado de la aplicación.

## Opciones

### Opción A: comandos dentro del binario principal

Ejemplo:

```txt
go run cmd/api/main.go serve
go run cmd/api/main.go create-root
go run cmd/api/main.go migrate up
go run cmd/api/main.go migrate down
go run cmd/api/main.go make module products
```

### Opción B: CLI separado dentro del mismo repo

```txt
cmd/nexokit/main.go
```

Ejemplo:

```txt
go run ./cmd/nexokit make module products
```

### Opción C: Makefile + scripts

Ejemplo:

```txt
make dev
make migrate-up
make migrate-down
make make-module name=products
make create-root
```

## Recomendación

Para la primera versión:

- Usar Makefile para comandos de desarrollo frecuentes.
- Crear CLI mínimo dentro del proyecto (`cmd/nexokit`).
- Incluir desde temprano `serve`, `create-root`, `migrate` y `make module`.
- Evitar que la generación de módulos quede como mejora tardía, porque define la arquitectura real del framework.
- Dejar `nexokit new` interactivo para una versión posterior, cuando la plantilla ya esté validada.

Comandos esperados para la primera versión como CLI interno:

```bash
go run ./cmd/nexokit make module products
go run ./cmd/nexokit make migration create_products_table
go run ./cmd/nexokit create-root
```

Comandos esperados para una versión posterior instalable:

```bash
nexokit new app-name
nexokit new app-name --auth --tenant --cache=redis --log-rotation=false
```

## Generador de módulos

El generador de módulos debe crear una estructura plana consistente con la convención de NexoKit:

```txt
internal/modules/products/
  handler.go
  service.go
  repository.go
  dto.go
  model.go
  routes.go
  validation.go
```

Y debe poder generar opcionalmente:

```txt
migrations/YYYYMMDDHHMMSS_create_products_table.sql
```

### Contrato del generador de módulos

El comando base será:

```bash
nexokit make module products
```

Flags sugeridos:

```bash
nexokit make module products --crud --migration --tenant
```

El flag `--permissions` queda fuera de la primera versión. La sincronización de permisos es una operación compleja que requiere introspección del sistema y se implementará cuando el CLI esté más maduro.

El generador debe poder crear:

```txt
Modelo con BaseModel.
DTOs de create, update, response y filtros.
Repository con búsqueda por PublicID.
Service con reglas de negocio mínimas.
Handler HTTP.
Routes del módulo con función Register.
Validaciones.
Migración SQL.
Scope de tenant si se usa --tenant.
```

El CLI no debe modificar lógica existente de forma silenciosa. Si necesita registrar rutas en archivos globales, debe hacerlo de forma explícita, idempotente y documentada.

## Makefile sugerido

```txt
make dev
make build
make test
make test-unit
make test-integration
make test-coverage
make migrate-up
make migrate-down
make migrate-create name=create_users_table
make seed
make create-root
make lint
make fmt
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe Makefile funcional.
2. Existe comando para correr la API en desarrollo.
3. Existe comando para compilar.
4. Existe comando para correr tests.
5. Existe comando para crear migraciones.
6. Existe comando para ejecutar migraciones.
7. Existe comando para revertir migraciones.
8. Existe comando seguro para crear root.
9. Existe documentación de comandos.
10. Existe generador básico de módulos.
11. El generador puede crear módulo CRUD con migración opcional.
12. El generador usa `BaseModel` con `ID` interno y `PublicID` externo.
13. El generador produce la función `Register` de rutas estándar.
14. El documento explica que `nexokit new` global y `permissions sync` quedan para versiones posteriores.
