package list_permissions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandlerListReturnsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(NewService(fakeRepo{}))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/permissions", nil)
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
