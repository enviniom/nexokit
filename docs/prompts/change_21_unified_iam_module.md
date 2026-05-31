> Lee también `_context.md` antes de implementar este change. 

Quiero iniciar un nuevo change SDD/OpenSpec para explorar e implementar una unificación progresiva de `users`, `roles` y `permissions` en un único módulo IAM/RBAC, sin eliminar todavía los módulos existentes.
Change: `21-unified-iam-module`
Objetivo:
Crear un nuevo módulo autocontenido que unifique la responsabilidad de usuarios, roles y permisos bajo una única frontera de dominio, usando arquitectura vertical slice, pero dejando intactos los módulos actuales `internal/modules/users/`, `internal/modules/roles/` e `internal/modules/permissions/` para poder revisar, contrastar y comparar antes de eliminar código legacy.
Nombre sugerido del nuevo módulo:
- Preferido: `iam`
- Alternativas aceptables si el explore lo justifica: `access`, `identity_access`
Regla principal:
- NO eliminar los módulos actuales.
- NO mover destructivamente archivos desde `users`, `roles` o `permissions`.
- NO borrar código legacy.
- El nuevo módulo debe convivir temporalmente con los módulos actuales.
- El único reemplazo funcional esperado en este change es el wiring del root app container para usar el nuevo módulo IAM donde corresponda.
Alcance:
- Crear `internal/modules/iam/`.
- Reproducir dentro de `iam` las capacidades actualmente necesarias de:
  - usuarios
  - roles
  - permisos
  - resolución de permisos para authz
  - sincronización de permisos registrados
- Mantener las rutas públicas y comportamiento actual.
- Reemplazar el wiring en `internal/app/container.go` para que la app use el nuevo módulo IAM.
- Mantener los módulos legacy como referencia viva para revisión.
Reglas de arquitectura:
- El nuevo módulo `iam` debe ser autocontenido.
- No debe depender de repositories de `users`, `roles` ni `permissions`.
- Puede definir modelos locales propios para users/roles/permissions con solo los campos necesarios.
- Las migraciones siguen siendo la fuente real del esquema de base de datos.
- Los modelos Go del módulo IAM son solo mapeos locales.
- Usar vertical slice por intención de negocio, no por detalle técnico.
- Un endpoint existente = un caso de uso = un slice.
- Los métodos internos usados por authz o bootstrap también deben modelarse como slices/casos de uso internos.
Estructura esperada:
- `internal/modules/iam/container.go`
- `internal/modules/iam/routes.go`
- `internal/modules/iam/core/`
- `internal/modules/iam/queries/`
- slices por intención de negocio, por ejemplo:
  - `list_users`
  - `view_user`
  - `create_user`
  - `update_user`
  - `list_roles`
  - `view_role`
  - `create_role`
  - `update_role`
  - `list_permissions`
  - `view_permission`
  - `update_permission`
  - `resolve_user_permissions`
  - `sync_permissions`
  - `assign_role_to_user`
  - `assign_permissions_to_role`
- Ajustar los nombres reales según los endpoints y casos existentes.
Compatibilidad:
- Mantener rutas HTTP actuales.
- Mantener payloads actuales.
- Mantener códigos de estado actuales.
- Mantener contratos usados por middleware/authz.
- Si hace falta, crear aliases/adapters temporales, pero documentarlos como compatibilidad transitoria.
- No cambiar migraciones salvo que el explore demuestre que es estrictamente necesario.
Antes de implementar:
1. Explorar `internal/modules/users/`, `internal/modules/roles/`, `internal/modules/permissions/`.
2. Mapear endpoints actuales → slices IAM.
3. Mapear métodos internos actuales → slices internos IAM.
4. Detectar dependencias cruzadas actuales y cómo desaparecen dentro de IAM.
5. Proponer estrategia de convivencia con módulos legacy.
6. Proponer qué cambia exactamente en `internal/app/container.go`.
7. Proponer riesgos, plan de migración y plan de rollback.
8. Crear proposal, specs, design y tasks.
9. Esperar mi revisión antes de aplicar.
Entrega esperada:
- Nuevo módulo `internal/modules/iam/` funcional.
- App container usando IAM como frontera principal.
- Módulos legacy conservados sin borrado destructivo.
- Tests del nuevo módulo.
- Tests existentes pasando.
- `go test ./...` pasando.
- OpenSpec actualizado con la realidad final.
- No hacer commit hasta que yo revise el resultado.
Fuera de alcance:
- Eliminar `internal/modules/users/`.
- Eliminar `internal/modules/roles/`.
- Eliminar `internal/modules/permissions/`.
- Renombrar rutas públicas.
- Cambiar comportamiento funcional.
- Hacer limpieza definitiva del legacy; eso queda para un change posterior.