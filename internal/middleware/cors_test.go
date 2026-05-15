package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: "https://example.com"},
	}

	r := gin.New()
	r.Use(CORS(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	origin := w.Header().Get("Access-Control-Allow-Origin")
	if origin != "https://example.com" {
		t.Errorf("expected CORS origin 'https://example.com', got %s", origin)
	}
}

func TestCORS_Preflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}

	r := gin.New()
	r.Use(CORS(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 for preflight, got %d", w.Code)
	}
}
