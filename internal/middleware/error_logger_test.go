package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

func TestErrorLogger_HandledErrorProducesOneLogLine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	r := gin.New()
	r.Use(RequestID())
	r.Use(DebugErrors(true))
	r.Use(ErrorLogger(log))
	r.GET("/fail", func(c *gin.Context) {
		response.HandleError(c, apperror.Internal(apperror.CodeInternal, "boom", errors.New("kaboom")))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	req.Header.Set(messages.HeaderRequestID, "req-123")
	r.ServeHTTP(w, req)

	lines := errorLogLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 error log line, got %d", len(lines))
	}
	line := lines[0]
	assertField(t, line, "request_id", "req-123")
	assertField(t, line, "method", "GET")
	assertField(t, line, "path", "/fail")
	assertField(t, line, "status", float64(500))
	assertField(t, line, "code", "internal")
	assertField(t, line, "public_message", "boom")
	assertField(t, line, "internal_chain", "kaboom")
	assertField(t, line, "tenant_id", "")
	assertField(t, line, "actor_id", "")
	if _, ok := line["latency_ms"]; !ok {
		t.Error("expected latency_ms field")
	}
}

func TestErrorLogger_PanicProducesOneLogLine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	r := gin.New()
	r.Use(RequestID())
	r.Use(ErrorLogger(log))
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	lines := errorLogLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 error log line, got %d", len(lines))
	}
	line := lines[0]
	assertField(t, line, "code", "internal")
	assertField(t, line, "public_message", messages.MsgInternalError)
	internal, ok := line["internal_chain"].(string)
	if !ok || !strings.Contains(internal, "panic: boom") {
		t.Errorf("expected internal_chain to contain 'panic: boom', got %v", line["internal_chain"])
	}
}

func TestErrorLogger_MissingContextFieldsAreEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	r := gin.New()
	r.Use(DebugErrors(true))
	r.Use(ErrorLogger(log))
	r.GET("/fail", func(c *gin.Context) {
		response.HandleError(c, apperror.Internal(apperror.CodeInternal, "boom", errors.New("kaboom")))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)

	lines := errorLogLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 error log line, got %d", len(lines))
	}
	assertField(t, lines[0], "tenant_id", "")
	assertField(t, lines[0], "actor_id", "")
}

func TestErrorLogger_TenantAndActorPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	r := gin.New()
	r.Use(DebugErrors(true))
	r.Use(func(c *gin.Context) {
		tenant.SetGin(c, tenant.NewScoped(1, "acme"))
		authctx.SetGin(c, &authctx.User{PublicID: "user-42"})
		c.Next()
	})
	r.Use(ErrorLogger(log))
	r.GET("/fail", func(c *gin.Context) {
		response.HandleError(c, apperror.Internal(apperror.CodeInternal, "boom", errors.New("kaboom")))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	r.ServeHTTP(w, req)

	lines := errorLogLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("expected 1 error log line, got %d", len(lines))
	}
	assertField(t, lines[0], "tenant_id", "acme")
	assertField(t, lines[0], "actor_id", "user-42")
}

func TestErrorLogger_NoErrorsNoLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	r := gin.New()
	r.Use(ErrorLogger(log))
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	if buf.Len() != 0 {
		t.Errorf("expected no error log output, got %q", buf.String())
	}
}

func errorLogLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var lines []map[string]any
	for _, raw := range bytes.Split(buf.Bytes(), []byte("\n")) {
		if len(raw) == 0 {
			continue
		}
		var line map[string]any
		if err := json.Unmarshal(raw, &line); err != nil {
			t.Fatalf("failed to unmarshal log line %q: %v", raw, err)
		}
		lines = append(lines, line)
	}
	return lines
}

func assertField(t *testing.T, line map[string]any, key string, want any) {
	t.Helper()
	got, ok := line[key]
	if !ok {
		t.Errorf("missing field %q in log line", key)
		return
	}
	if got != want {
		t.Errorf("%s = %v (%T), want %v (%T)", key, got, got, want, want)
	}
}
