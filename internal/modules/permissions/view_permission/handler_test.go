package view_permission

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/gin-gonic/gin"
)

type okSvc struct{}

func (okSvc) GetByPublicID(string) (*core.PermissionResponse, error) { return &core.PermissionResponse{}, nil }

type nfSvc struct{}

func (nfSvc) GetByPublicID(string) (*core.PermissionResponse, error) { return nil, core.ErrNotFound }

func TestHandlerGetByPublicID200And404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		svc  Service
		code int
	}{{okSvc{}, http.StatusOK}, {nfSvc{}, http.StatusNotFound}} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/permissions/p1", nil)
		c.Params = gin.Params{{Key: "id", Value: "p1"}}
		NewHandler(tc.svc).GetByPublicID(c)
		if w.Code != tc.code {
			t.Fatalf("expected %d got %d", tc.code, w.Code)
		}
	}
}
