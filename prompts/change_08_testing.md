> Lee también `_context.md` antes de implementar este change.

# Change 8: Testing, calidad y CI básico

## Objetivo

Agregar una estrategia de testing y calidad para NexoKit, asegurando que las funcionalidades base del framework puedan verificarse automáticamente y que los proyectos creados con NexoKit tengan una base confiable de pruebas.

## Principio clave

En Go, los unit tests deben vivir junto al código que prueban, usando archivos `_test.go` dentro del mismo paquete.

La carpeta `/tests` debe reservarse principalmente para:

- Integration tests.
- Test helpers.
- Fixtures.
- Setup de base de datos de prueba.

## Alcance

Implementar:

- Unit tests.
- Integration tests.
- Test database.
- Test helpers.
- Factories o fixtures.
- Mocks para servicios externos cuando aplique.
- Tests de middlewares.
- Tests de respuesta API estándar.
- Tests de auth.
- Tests de RBAC.
- Tests de tenant scope.
- Tests de paginación y filtros.
- Tests de cache adapter.
- Tests de rate limit.
- Coverage.
- Makefile con comandos de test.
- GitHub Actions básico.

## Herramientas sugeridas

- `testing` estándar de Go.
- `httptest` para endpoints HTTP.
- `testify` opcional para assertions cuando aporta claridad.
- PostgreSQL de test usando Docker Compose.
- Redis/Valkey de test usando Docker Compose.
- GitHub Actions para CI.

## Estructura recomendada

### Unit tests junto al código

```txt
internal/
  platform/response/
    response.go
    response_test.go

  modules/auth/
    service.go
    service_test.go
    password.go
    password_test.go
    token.go
    token_test.go

  middleware/
    auth.go
    auth_test.go
    rbac.go
    rbac_test.go
    tenant.go
    tenant_test.go
    rate_limit.go
    rate_limit_test.go

  platform/query/
    pagination.go
    pagination_test.go
    filters.go
    filters_test.go

  infra/cache/
    noop.go
    noop_test.go
    redis.go
    redis_test.go
```

### Integration tests separados

```txt
tests/
  integration/
    auth_test.go
    users_test.go
    tenant_test.go
    rbac_test.go
    health_test.go

  helpers/
    app.go
    database.go
    auth.go
    fixtures.go

  fixtures/
    users.go
    companies.go
    roles.go
```

## Paquetes de testing

Usar el mismo paquete cuando se necesite acceder a funciones no exportadas:

```go
package auth
```

Usar paquete externo cuando se quiera probar como consumidor del paquete:

```go
package auth_test
```

## Buenas prácticas obligatorias

Los tests de NexoKit deben seguir estas reglas:

- Los unit tests deben estar junto al código probado.
- Los archivos deben terminar en `_test.go`.
- Los tests deben usar nombres descriptivos.
- Usar subtests con `t.Run`.
- Usar table-driven tests para casos repetitivos.
- Revisar errores explícitamente.
- Usar `t.Fatal` o `t.Fatalf` para errores de setup.
- Usar `t.Error` o `t.Errorf` para aserciones que permiten continuar.
- Evitar panics en código productivo.
- Preferir interfaces pequeñas para facilitar mocks.
- No abusar de frameworks pesados de testing.
- Usar `testing` estándar de Go como base.
- `testify` puede usarse para assertions si aporta claridad, pero no debe ocultar la lógica.

## Ejemplo de table-driven test

```go
func TestPagination_Normalize(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		perPage  int
		wantPage int
		wantPer  int
	}{
		{
			name:     "default values",
			page:     0,
			perPage:  0,
			wantPage: 1,
			wantPer:  15,
		},
		{
			name:     "negative page becomes first page",
			page:     -1,
			perPage:  20,
			wantPage: 1,
			wantPer:  20,
		},
		{
			name:     "per page above max is capped",
			page:     1,
			perPage:  500,
			wantPage: 1,
			wantPer:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePagination(tt.page, tt.perPage)

			if got.Page != tt.wantPage {
				t.Errorf("expected page %d, got %d", tt.wantPage, got.Page)
			}

			if got.PerPage != tt.wantPer {
				t.Errorf("expected per_page %d, got %d", tt.wantPer, got.PerPage)
			}
		})
	}
}
```

## Testing por capas

### Unit tests

Para lógica pura y rápida:

```txt
response
errors
pagination
filters
validators
password hashing
PASETO generation/parsing
RBAC permission checks
tenant context
cache interface noop
rate limit logic
```

### Handler tests

Usando `httptest`:

```txt
login endpoint
me endpoint
protected route without token
protected route with invalid token
protected route without permission
paginated list endpoint
```

### Integration tests

Para probar componentes conectados:

```txt
login real con DB de test
refresh token real con DB
CRUD users
tenant isolation con DB
RBAC sobre rutas HTTP
migraciones
repositories GORM
cache Redis/Valkey
```

## Casos mínimos de prueba

### Response

- Respuesta exitosa.
- Respuesta de error.
- Respuesta de validación.
- Respuesta paginada.

### Auth

- Login exitoso.
- Login con credenciales inválidas.
- Usuario inactivo no puede iniciar sesión.
- Refresh token válido.
- Refresh token revocado.
- Logout revoca refresh token.

### RBAC

- Usuario con permiso accede.
- Usuario sin permiso recibe 403.
- Usuario no autenticado recibe 401.
- Root tiene todos los permisos.

### Multitenancy

- Admin solo accede a datos de su company.
- Admin no puede acceder a datos de otra company.
- Root puede acceder globalmente.
- Tenant se resuelve correctamente desde contexto.

### Paginación y filtros

- Paginación por defecto.
- Cambio de `page`.
- Cambio de `per_page`.
- Sorting.
- Search.
- Filtro por estado.

### Cache

- Cache disabled usa NoopCache.
- Redis/Valkey set/get/delete.
- TTL funciona correctamente.

### Rate limit

- Request dentro del límite pasa.
- Request por encima del límite responde 429.
- Login tiene rate limit propio.

## Comandos esperados

```txt
make test
make test-unit
make test-integration
make test-coverage
```

## CI básico

Crear workflow de GitHub Actions que ejecute:

```txt
go test ./...
go vet ./...
go fmt check
```

Opcional:

```txt
golangci-lint
coverage report
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe estructura base de tests.
2. Los unit tests viven junto al código probado.
3. Los integration tests viven en `/tests/integration`.
4. Existen tests unitarios para response.
5. Existen tests para validaciones.
6. Existen tests para auth.
7. Existen tests para RBAC.
8. Existen tests para tenant scope.
9. Existen tests para paginación.
10. Existen tests para filtros.
11. Existen tests para cache.
12. Existen tests para rate limit.
13. Existe comando `make test`.
14. Existe comando `make test-unit`.
15. Existe comando `make test-integration`.
16. Existe comando `make test-coverage`.
17. Existe CI básico con GitHub Actions.
18. La documentación explica cómo correr tests.
19. Los tests pueden ejecutarse en ambiente local.
20. Los tests siguen buenas prácticas idiomáticas de Go.
