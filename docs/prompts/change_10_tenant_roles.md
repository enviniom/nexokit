> Lee también `_context.md` antes de implementar este change.

# Change 10: Roles por tenant y root global

## Objetivo

Hacer que los roles administrables pertenezcan a una company, manteniendo el rol `root` como rol global del sistema.

Este change corrige el modelo RBAC para SaaS: `admin`, `user` y roles custom no deben ser globales, porque sus permisos y usuarios dependen del tenant.

## Contexto

Actualmente existen roles base y permisos globales. Para multitenancy real, cada company debe poder tener sus propios roles:

```txt
company A -> admin, user, seller
company B -> admin, user, support
```

El rol `root` es especial: no pertenece a ninguna company, no necesita permisos asignados y conserva bypass global.

## Alcance de este change

Implementar:

- `company_id` nullable en `roles`.
- Scope tenant para roles no-root.
- Seed inicial con solo rol `root` global.
- Eliminación o ajuste del seed global de role-permissions.
- Bypass de permisos para root sin depender de registros en `role_permissions`.
- Restricción para no crear roles reservados desde API.
- Tests de aislamiento tenant y protección de root.

## Reglas

- El rol `root` debe tener `company_id = null` e `is_system = true`.
- Los roles `admin`, `user` y custom deben tener `company_id`.
- No se puede crear desde API un rol con slug reservado: `root`, `admin`, `user`, salvo los roles creados por el proceso de onboarding de company.
- No se puede crear un rol `root` desde API.
- No se puede editar ni borrar el rol `root` desde API.
- Un usuario no-root solo puede ver roles de su company.
- Root puede operar globalmente y también en contexto de una company cuando corresponda.
- El root no necesita registros en `role_permissions`; el middleware debe resolver bypass antes de consultar permisos.
- El seed inicial no debe crear permisos ni asignaciones de permisos para root.

## Campos

Confirmar o agregar en `roles`:

```txt
company_id nullable
```

Reglas esperadas:

```txt
root  -> company_id null
admin -> company_id requerido
user  -> company_id requerido
custom roles -> company_id requerido
```

## Impacto en seeds

El seed inicial debe dejar el sistema en estado mínimo seguro:

```txt
roles:
- root global

permissions:
- ninguno, o solo catálogo si el sistema lo requiere explícitamente

role_permissions:
- ninguno para root
```

Los roles `admin` y `user` de cada tenant se crearán más adelante durante el onboarding de company.

## Criterios de aceptación

Este change se considera completo cuando:

1. `roles.company_id` existe y permite `null` solo para root/global.
2. El seed inicial crea solo el rol `root` global.
3. Root tiene bypass de permisos sin depender de `role_permissions`.
4. No se puede crear un rol `root` desde API.
5. No se puede editar el rol `root` desde API.
6. No se puede borrar el rol `root` desde API.
7. No se puede crear un rol con slug reservado desde API.
8. Un usuario no-root solo lista roles de su company.
9. Un usuario no-root no puede acceder a roles de otra company.
10. Los tests prueban root global, roles tenant y aislamiento por company.
