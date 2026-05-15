package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/gin-gonic/gin"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	r := gin.New()
	r.Use(Logger(log))
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	output := buf.String()
	if !strings.Contains(output, messages.MsgHTTPRequest) {
		t.Errorf("expected log output to contain %q", messages.MsgHTTPRequest)
	}
	if !strings.Contains(output, "GET") {
		t.Error("expected log output to contain method GET")
	}
}
