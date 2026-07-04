package middleware

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

// ErrorLogger is the single owner of error log records. It runs after the
// request handler and any inner middleware (including Recovery) have finished,
// then emits one structured slog.Error record per entry in c.Errors.
func ErrorLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		if len(c.Errors) == 0 {
			return
		}

		rid, _ := c.Get(messages.CtxRequestID)
		ridStr, _ := rid.(string)

		tenantID := ""
		if tenantCtx, ok := tenant.FromGin(c); ok {
			tenantID = tenantCtx.CompanySlug
		}

		actorID := authctx.PublicIDFromGin(c)
		status := c.Writer.Status()

		for _, err := range c.Errors {
			code, publicMsg, internalChain := extractErrorLogInfo(err.Err)

			log.Error(messages.MsgHTTPRequest,
				slog.String(messages.CtxRequestID, ridStr),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int("status", status),
				slog.Int64("latency_ms", latency.Milliseconds()),
				slog.String("tenant_id", tenantID),
				slog.String("actor_id", actorID),
				slog.String("code", code),
				slog.String("public_message", publicMsg),
				slog.String("internal_chain", internalChain),
			)
		}
	}
}

// extractErrorLogInfo pulls the fields needed for structured error logging.
// When the error is (or wraps) an *AppError, it uses Code/PublicMessage and the
// Internal chain; otherwise it logs the raw error text as internal_chain.
func extractErrorLogInfo(err error) (code string, publicMsg string, internalChain string) {
	var ae *apperror.AppError
	if errors.As(err, &ae) {
		code = string(ae.Code)
		publicMsg = ae.PublicMessage
		if ae.Internal != nil {
			internalChain = unwrapChain(ae.Internal)
		}
		return
	}
	internalChain = unwrapChain(err)
	return
}

func unwrapChain(err error) string {
	if err == nil {
		return ""
	}

	parts := []string{err.Error()}
	for unwrapped := errors.Unwrap(err); unwrapped != nil; unwrapped = errors.Unwrap(unwrapped) {
		parts = append(parts, unwrapped.Error())
	}

	return strings.Join(parts, " -> ")
}
