package middleware

import (
	"fmt"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

// Recovery catches panics, pushes the panic value into c.Errors as an
// *AppError, writes a structured 500 response, and aborts. Logging is owned by
// ErrorLogger, which runs after Recovery on the Gin unwind.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				err := apperror.Internal(
					apperror.CodeInternal,
					messages.MsgInternalError,
					fmt.Errorf("panic: %v", r),
				)
				_ = c.Error(err)
				response.InternalServerError(c, messages.MsgInternalError)
				c.Abort()
			}
		}()
		c.Next()
	}
}
