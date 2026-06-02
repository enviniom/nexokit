> Lee también `_context.md` antes de implementar este change.
Change: 23-remove-legacy-iam-modules
Objetivo:
Eliminar los módulos legacy `internal/modules/users/`, `internal/modules/roles/` e `internal/modules/permissions/` después de que `internal/modules/iam/` quedó como frontera principal.
Alcance:
- Eliminar `internal/modules/users/`
- Eliminar `internal/modules/roles/`
- Eliminar `internal/modules/permissions/`
- Eliminar imports, tests o wiring residual si existen
- Actualizar OpenSpec/docs si todavía mencionan preservación legacy como estado activo
- Verificar que IAM sigue compilando y que la app usa solo IAM
Validaciones:
- `go list ./...` sin referencias a los módulos eliminados
- `go test ./...`
- `go build ./...`
- Verificar rutas públicas users/roles/permissions siguen funcionando vía IAM
- Verificar auth/login/session sigue resolviendo usuarios vía IAM
- Verificar seeds/bootstrap/sync permissions no dependen de legacy
Fuera de alcance:
- Cambiar rutas públicas
- Cambiar payloads
- Cambiar schema/migraciones
- Refactor adicional de IAM no relacionado al borrado
Punto importante: antes de borrar, haría una búsqueda explícita de imports/referencias a:
- internal/modules/users
- internal/modules/roles
- internal/modules/permissions
- constructores legacy
- tests que importen handlers/services legacy
Si eso da limpio, el borrado debería ser seguro.