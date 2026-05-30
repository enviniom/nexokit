package update_permission

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/gin-gonic/gin"
)

type svcResp struct{ err error }

func (s svcResp) Update(string, core.UpdatePermissionRequest) (*core.PermissionResponse, error) {
	if s.err != nil { return nil, s.err }
	return &core.PermissionResponse{}, nil
}

func TestHandlerUpdateStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct{ err error; code int }{{nil, http.StatusOK}, {core.ErrNotFound, http.StatusNotFound}, {core.ErrSystemImmutable, http.StatusConflict}} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/permissions/p1", bytes.NewBufferString(`{"name":"x"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "p1"}}
		NewHandler(svcResp{err: tc.err}).Update(c)
		if w.Code != tc.code { t.Fatalf("expected %d got %d", tc.code, w.Code) }
	}
}
