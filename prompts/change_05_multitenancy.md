> Lee también `_context.md` antes de implementar este change.

# Change 5: Multitenancy por company_id

## Objetivo

Implementar multitenancy por `company_id` como funcionalidad base de NexoKit.

El sistema debe permitir que una misma API y una misma base de datos sirvan a múltiples empresas, manteniendo aislamiento de datos.

## Contexto

NexoKit será usado para aplicaciones SaaS. En el caso de la tienda virtual, cada empresa tendrá sus propios productos, pedidos, clientes, usuarios y configuración.

Por eso NexoKit debe traer multitenancy como característica base, no como agregado posterior.

## Alcance de este change

Implementar:

- Modelo Company.
- Company ID en usuarios.
- Tenant context.
- Middleware de tenant.
- Repositorios filtrados por company_id.
- Helpers para aplicar scope por tenant.
- Protección contra acceso cruzado.
- Root con acceso global.
- Admin/user limitado a su company.
- Resolución de tenant por header, dominio o subdominio.

## Tabla companies

```txt
id
public_id
name
slug
domain nullable
subdomain nullable
status
created_at
updated_at
deleted_at nullable
created_by nullable
updated_by nullable
```

## Ajuste en users

```txt
company_id nullable
```

Reglas:

- Root puede tener `company_id` nulo.
- Admin y user deben tener `company_id`.
- El sistema debe validar esto.

## Resolución de tenant

Soportar inicialmente:

```txt
X-Company-ID
Host header
Subdomain
Domain
```

Orden sugerido para APIs privadas:

1. Usuario autenticado.
2. Company del usuario.
3. Header solo si root o modo desarrollo.

Orden sugerido para APIs públicas:

1. Host header.
2. Domain configurado.
3. Subdomain configurado.
4. Header `X-Tenant` solo en desarrollo.

## Tenant context

Crear un contexto reusable:

```txt
TenantContext
- company_id
- company_slug
- is_root_scope
```

## GORM scopes

Crear helpers:

```go
func WithCompany(db *gorm.DB, companyID uint) *gorm.DB
func ApplyTenantScope(db *gorm.DB, ctx TenantContext) *gorm.DB
```

Regla:

- Si el usuario es root y está en modo global, puede consultar todo.
- Si el usuario no es root, todas las consultas deben filtrar por `company_id`.

## Endpoints de companies

```txt
GET    /api/v1/companies
POST   /api/v1/companies
GET    /api/v1/companies/:id
PUT    /api/v1/companies/:id
DELETE /api/v1/companies/:id
```

Solo root debe poder crear empresas.

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe tabla companies.
2. Los usuarios pueden estar asociados a company.
3. El root puede crear companies.
4. El admin no puede crear companies.
5. El tenant se carga en contexto.
6. Las consultas se filtran por company_id.
7. Un admin no puede leer datos de otra company.
8. Un admin no puede modificar datos de otra company.
9. El root puede operar en modo global.
10. El root puede operar en contexto de una company específica.
11. El middleware de tenant funciona en rutas privadas.
12. El middleware de tenant funciona en rutas públicas.
13. Existe helper o scope reusable de GORM para aplicar tenant.
14. Existe documentación de cómo crear nuevos modelos multitenant.
