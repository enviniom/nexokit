package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/gin-gonic/gin"
)

func TestDebugErrors_StoresEnabledFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(DebugErrors(true))
	r.GET("/flag", func(c *gin.Context) {
		v, ok := c.Get(messages.CtxDebugErrors)
		if !ok {
			t.Fatal("expected debug_errors to be set in context")
		}
		enabled, ok := v.(bool)
		if !ok {
			t.Fatalf("debug_errors value is not bool: %T", v)
		}
		if !enabled {
			t.Errorf("debug_errors = %v; want true", enabled)
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/flag", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestDebugErrors_StoresDisabledFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(DebugErrors(false))
	r.GET("/flag", func(c *gin.Context) {
		v, ok := c.Get(messages.CtxDebugErrors)
		if !ok {
			t.Fatal("expected debug_errors to be set in context")
		}
		enabled, ok := v.(bool)
		if !ok {
			t.Fatalf("debug_errors value is not bool: %T", v)
		}
		if enabled {
			t.Errorf("debug_errors = %v; want false", enabled)
		}
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/flag", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
