package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func TestRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))
	var captured []*gin.Error

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Next()
		captured = c.Errors
	})
	r.Use(RequestID())
	r.Use(ErrorLogger(log))
	r.Use(Recovery())
	r.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}

	var resp response.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Success {
		t.Error("expected success to be false")
	}
	if resp.Message != messages.MsgInternalError {
		t.Errorf("unexpected message: %s", resp.Message)
	}
	if len(captured) != 1 {
		t.Fatalf("expected exactly one error in context, got %d", len(captured))
	}

	// Recovery must not emit its own log line; ErrorLogger owns the record.
	// The captured buffer should contain exactly the ErrorLogger record.
	if buf.Len() == 0 {
		t.Error("expected ErrorLogger to emit the panic record")
	}
	if !bytes.Contains(buf.Bytes(), []byte("\"code\":\"internal\"")) {
		t.Errorf("expected ErrorLogger record with code, got %q", buf.String())
	}
}

func TestRecovery_ErrorLoggerLogsPanic(t *testing.T) {
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

	if buf.Len() == 0 {
		t.Fatal("expected ErrorLogger to emit a log line for the panic")
	}
	if !bytes.Contains(buf.Bytes(), []byte("panic: boom")) {
		t.Errorf("expected log to contain panic value, got %q", buf.String())
	}
}

func TestRecovery_NormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	r := gin.New()
	r.Use(Recovery())
	r.Use(ErrorLogger(log))
	r.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ok", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no error log output for normal request, got %q", buf.String())
	}
}

func TestRecovery_PanicErrorIsAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))
	var captured []*gin.Error

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Next()
		captured = c.Errors
	})
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

	var ae *apperror.AppError
	for _, err := range captured {
		if errors.As(err.Err, &ae) {
			break
		}
	}
	if ae == nil {
		t.Fatal("expected panic error to be an *AppError")
	}
	if ae.Code != apperror.CodeInternal {
		t.Errorf("Code = %q, want %q", ae.Code, apperror.CodeInternal)
	}
}
