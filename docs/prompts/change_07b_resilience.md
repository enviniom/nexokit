> Lee también `_context.md` antes de implementar este change.

# Change 7b: Infraestructura de resiliencia: cache y rate limit

## Objetivo

Implementar cache y rate limiting como capacidades de resiliencia opcionales de NexoKit.

## Alcance de este change

Implementar:

- Cache adapter con interfaz.
- Cliente Redis/Valkey.
- `NoopCache` para proyectos sin cache.
- Rate limiter por IP.
- Rate limiter para endpoints sensibles (login, refresh).

## Cache

Soportar Redis o Valkey mediante `go-redis`, compatible con ambos.

Variables:

```txt
CACHE_DRIVER
REDIS_ADDR
REDIS_PASSWORD
REDIS_DB
CACHE_DEFAULT_TTL_SECONDS
```

El sistema debe permitir:

```txt
CACHE_DRIVER=none
```

para proyectos que no requieran cache.

## Cache adapter

```go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, error)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

Implementaciones:

```txt
RedisCache
NoopCache
```

## Rate limit

Implementar rate limit para:

- Login.
- Refresh token.
- Endpoints públicos sensibles.
- API general opcional.

Implementación:

- Memoria local por defecto (suficiente para instancia única).
- Redis si está configurado (para rate limit distribuido).

Variables:

```txt
RATE_LIMIT_ENABLED
RATE_LIMIT_REQUESTS
RATE_LIMIT_WINDOW_SECONDS
LOGIN_RATE_LIMIT_REQUESTS
LOGIN_RATE_LIMIT_WINDOW_SECONDS
```

## Criterios de aceptación

Este change se considera completo cuando:

1. Existe cliente Redis/Valkey configurable.
2. Existe interfaz `Cache`.
3. Existe `RedisCache`.
4. Existe `NoopCache` si cache está deshabilitada.
5. Existe rate limit global opcional.
6. Existe rate limit específico para login.
7. Los endpoints rate-limited responden 429 cuando corresponde.
8. El `docker-compose.yml` incluye Redis para desarrollo local.
