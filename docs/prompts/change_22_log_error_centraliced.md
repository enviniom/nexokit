Quiero iniciar un nuevo change SDD/OpenSpec para agregar logging centralizado de errores manejados por la API.
Change: `handled-api-error-logging`
Objetivo:
Agregar observabilidad para errores procesados por `response.HandleError` sin exponer detalles internos al cliente.
Contexto:
Actualmente `response.HandleError(c, err)` construye un `response.Error` usando `apperror.Status(err)` y `apperror.PublicMessage(err, gin.Mode())`. En producción los detalles técnicos se ocultan correctamente, pero no queda claro que el error original se registre en logs. Esto puede hacer que errores de dominio/técnicos se pierdan operacionalmente.
Reglas:
- No exponer `err.Error()` directamente al cliente en production.
- Mantener mensajes públicos seguros.
- Registrar el error original completo para observabilidad.
- Preservar el contrato actual de responses.
- Evitar que handlers individuales tengan que loguear manualmente.
- No mezclar logging de dominio dentro de services/repositories.
Explorar:
1. Cómo se usa actualmente `response.HandleError`.
2. Si existe middleware que procese `gin.Context.Errors`.
3. Si conviene que `HandleError` haga `c.Error(err)` y que un middleware loguee.
4. Si conviene inyectar logger en response/middleware.
5. Cómo incluir request_id, method, path, status, public message y error original.
6. Cómo testear que el error se preserva para logging sin filtrarse al cliente.
Entrega esperada:
- Proposal, specs, design y tasks.
- Implementación con tests.
- `go test ./...` pasando.
Fuera de alcance:
- Cambiar mensajes públicos de API.
- Exponer detalles técnicos al cliente.
- Refactor masivo de todos los handlers.