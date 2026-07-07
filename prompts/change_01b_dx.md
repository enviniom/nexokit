> Lee también `_context.md` antes de implementar este change.

# Change 9: Developer Experience — constantes de mensajes, pre-commit y herramientas de desarrollo

## Objetivo

Mejorar la experiencia de desarrollo de NexoKit sin afectar el runtime de la aplicación. Este change se enfoca en consistencia de mensajes, calidad del código y automatizaciones que evitan errores comunes antes de cada commit.

## Alcance

### 1. Constantes de mensajes

Centralizar todos los magic strings de mensajes que hoy están dispersos en el código:

- Mensajes de respuesta HTTP genéricos (`"Operación exitosa"`, `"Datos obtenidos correctamente"`, etc.)
- Mensajes de error de validación de campos (`"es requerido"`, `"debe ser un email válido"`, etc.)
- Mensajes de middleware (recovery, request ID, etc.)
- Mensajes de logger

Las constantes de `apperror` ya son la fuente de verdad de los errores de aplicación — no se mueven.

Ubicación propuesta:

```txt
internal/platform/messages/
  messages.go     <- constantes de mensajes de respuesta API
```

O alternativamente como constantes dentro de cada paquete que las usa (evaluar en el change). La decisión debe documentarse.

### 2. Pre-commit hook

Implementar un script de pre-commit en `scripts/pre-commit.sh` que verifique antes de cada commit:

1. **Binarios**: detecta si hay archivos binarios (ejecutables) siendo agregados al commit y los rechaza con mensaje claro.
2. **Tamaño de archivos**: alerta si algún archivo del commit supera un umbral configurable (default: 1MB). No bloquea, solo advierte.
3. **Paridad de variables de entorno**: compara las keys de `.env` (si existe) con las de `.env.example` y reporta:
   - Variables en `.env` que no están en `.env.example` (olvidadas documentar).
   - Variables en `.env.example` que no están en `.env` (olvidadas configurar localmente).
4. **`go vet`**: corre `go vet ./...` y rechaza el commit si hay errores.
5. **`go fmt`**: detecta archivos Go sin formatear y rechaza el commit (no reformatea automáticamente — el dev debe correr `make fmt`).

El hook debe:
- Ser instalable con un solo comando: `make install-hooks`
- Tener salida clara y con colores (verde ✓, rojo ✗, amarillo ⚠)
- Fallar rápido: si un check crítico falla, no continúa con los siguientes

### 3. Makefile

Agregar targets:

```makefile
install-hooks   # instala el pre-commit hook en .git/hooks/
uninstall-hooks # elimina el hook instalado
check-env       # corre solo la verificación de paridad de .env vs .env.example
```

### 4. Documentación

Actualizar `README.md` con:
- Sección de setup del pre-commit hook
- Descripción de qué verifica cada check
- Instrucciones para saltarse el hook en casos excepcionales (`git commit --no-verify`)

## Variables de entorno

No se agregan variables de entorno nuevas. El umbral de tamaño de archivo del pre-commit puede ser configurable directamente en el script con una variable al tope del archivo.

## Criterios de aceptación

1. No existe ningún magic string de mensaje en `platform/response`, `platform/validator`, `middleware/` — todos usan constantes.
2. `make install-hooks` instala el pre-commit en `.git/hooks/pre-commit`.
3. Hacer commit con un binario falla con mensaje claro.
4. Hacer commit con una variable en `.env` que no está en `.env.example` muestra advertencia.
5. Hacer commit con código sin formatear falla con mensaje claro.
6. `make uninstall-hooks` elimina el hook limpiamente.
7. README documenta el flujo completo del pre-commit.
