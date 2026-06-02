> Lee también `_context.md` antes de implementar este change.

Quiero iniciar un nuevo change SDD/OpenSpec para auditar toda la app y verificar cumplimiento de las reglas de vertical slice documentadas en `docs/vertical-slice-modules.md`.
Change: `vertical-slice-rules-audit`
Objetivo:
Revisar todos los módulos de la app para detectar desviaciones contra las convenciones vertical slice actuales y proponer/implementar correcciones de forma segura, módulo por módulo y slice por slice.
Contexto:
Ya existe `docs/vertical-slice-modules.md` como contrato arquitectónico. El módulo IAM fue ajustado siguiendo estas reglas:
- cada slice tiene `handler-service-repository`;
- no existen repositories compartidos por entidad;
- queries reutilizables van en `queries/`, una query por archivo y con tests;
- queries single-use quedan en el repository del slice;
- repositories traducen errores GORM/DB a errores de dominio;
- services no importan GORM ni `platform/apperror`;
- handlers traducen errores de dominio a HTTP/API;
- modelos GORM parciales deben declarar `TableName()` cuando mapean tablas existentes.
Alcance:
1. Auditar módulos bajo `internal/modules/`.
2. Identificar:
   - `shared/repository.go` o repositories compartidos por entidad;
   - services importando GORM;
   - services importando `platform/apperror`;
   - handlers con lógica de negocio;
   - repositories devolviendo errores HTTP/API;
   - queries reutilizadas que deberían vivir en `queries/`;
   - queries single-use mal extraídas;
   - modelos parciales sin `TableName()`;
   - slices sin tests de handler/service/repository.
3. Generar reporte por módulo y prioridad.
4. Proponer plan de corrección por slices, no en bloque.
5. Implementar solo después de revisión.
Reglas:
- No refactorizar todo de una vez.
- Corregir módulo por módulo y slice por slice.
- No cambiar rutas públicas, payloads ni status codes salvo que el audit detecte un bug y se apruebe explícitamente.
- Mantener tests con el slice que verifican.
- Si una corrección toca comportamiento, frenar para revisión antes de continuar.
- No hacer commit hasta revisión.
Entrega esperada:
- OpenSpec proposal/spec/design/tasks.
- Reporte de auditoría con tabla:
  - módulo;
  - slice;
  - violación;
  - severidad;
  - corrección sugerida;
  - tests requeridos.
- Plan de implementación incremental.
- `go test ./...` y `go build ./...` como verificación final.
Fuera de alcance:
- Reescribir módulos completos sin necesidad.
- Cambiar arquitectura fuera de `internal/modules/`.
- Eliminar legacy sin un change dedicado.