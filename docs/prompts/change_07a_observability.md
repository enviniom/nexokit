> Lee también `_context.md` antes de implementar este change.

# Change 7a: Infraestructura de observabilidad: logger, log rotator, health checks y graceful shutdown

## Objetivo

Implementar las funcionalidades de observabilidad y ciclo de vida del servidor que deben estar listas en NexoKit para cualquier proyecto.

Este change se separa del Change 7b (cache y rate limit) para evitar que decisiones sobre el logger bloqueen el trabajo de infraestructura de resiliencia. Son responsabilidades independientes.

## Alcance de este change

Implementar:

- Logger estructurado con `slog`.
- Middleware de request logging.
- Log rotator con `lumberjack`.
- Configuración de archivo de logs.
- Health check de base de datos.
- Health checks extendidos (`/health/live`, `/health/ready`).
- Graceful shutdown.
- Recovery middleware refinado.

## Logger

NexoKit usa `slog` (estándar de Go desde 1.21) para minimizar dependencias externas.

Si en un proyecto concreto se necesita máximo rendimiento o estructura avanzada, se puede reemplazar por `zap`, pero `slog` es el default del framework.

## Log rotation

Usar `lumberjack` para que funcione igual en cualquier entorno sin depender del sistema operativo.

Variables:

```txt
LOG_LEVEL
LOG_FORMAT
LOG_DIR
LOG_FILE
LOG_MAX_SIZE_MB
LOG_MAX_BACKUPS
LOG_MAX_AGE_DAYS
LOG_COMPRESS
```

## Health checks

Endpoints:

```txt
GET /health
GET /health/live
GET /health/ready
```

`/health/live`:

- API está viva.

`/health/ready`:

- DB conectada.
- Cache conectada si está habilitada.

## Graceful shutdown

Implementar cierre controlado:

- Detener servidor HTTP.
- Cerrar conexión DB.
- Cerrar cache si está activa.
- Respetar timeout.

Variables:

```txt
SHUTDOWN_TIMEOUT_SECONDS
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe logger estructurado con `slog`.
2. Los requests se registran con método, path, status y duración.
3. Los errores se registran con contexto.
4. Existe rotación de logs con `lumberjack`.
5. Los logs se guardan en archivo si está configurado.
6. Los logs pueden imprimirse en consola en desarrollo.
7. `/health/live` funciona.
8. `/health/ready` valida DB y cache.
9. Existe graceful shutdown.
10. Existe recovery middleware refinado.
