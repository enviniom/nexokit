package messages

// API response messages
const (
	MsgSuccess         = "Operación exitosa"
	MsgCreated         = "Recurso creado exitosamente"
	MsgDeleted         = "Recurso eliminado exitosamente"
	MsgHealthy         = "API is healthy"
	MsgValidationError = "Error de validación"
	MsgInternalError   = "Error interno del servidor"
	MsgNotFound        = "Recurso no encontrado"
	MsgUnauthorized    = "No autorizado"
	MsgForbidden       = "Acceso denegado"
	MsgConflict        = "El recurso ya existe"
	MsgBadRequest      = "Solicitud inválida"
)

// Validation rule messages
const (
	MsgRequired       = "es requerido"
	MsgMinLength      = "debe tener al menos %d caracteres"
	MsgMaxLength      = "no puede superar %d caracteres"
	MsgValidEmail     = "debe ser un email válido"
	MsgHasUppercase   = "debe contener al menos una mayúscula"
	MsgHasDigit       = "debe contener al menos un número"
	MsgHasSpecialChar = "debe contener al menos un carácter especial"
	MsgMinWords       = "debe contener al menos %d palabras"
	MsgNoNumbers      = "no debe contener números"
	MsgInvalidFormat  = "formato inválido"
)

// Middleware messages
const (
	MsgPanicRecovered = "internal server error"
	MsgPanicLog       = "panic recovered"
)
