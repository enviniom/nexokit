package revoke_token

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeService struct{ err error }

func (f fakeService) Revoke(req core.RefreshRequest) error { return f.err }

func jsonRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns success", func(t *testing.T) {
		h := NewHandler(fakeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/auth/logout", core.RefreshRequest{RefreshToken: "refresh"})

		h.Handle(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("returns validation error", func(t *testing.T) {
		h := NewHandler(fakeService{})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/auth/logout", core.RefreshRequest{})

		h.Handle(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}
	})

	t.Run("maps unauthorized error", func(t *testing.T) {
		h := NewHandler(fakeService{err: core.ErrInvalidRefreshToken})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/auth/logout", core.RefreshRequest{RefreshToken: "refresh"})

		h.Handle(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}

		assertUnauthorizedEnvelope(t, w)
	})
}

func assertUnauthorizedEnvelope(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	var resp response.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if resp.Success != false {
		t.Errorf("Success = %v, want false", resp.Success)
	}
	if resp.Message != messages.MsgUnauthorized {
		t.Errorf("Message = %q, want %q", resp.Message, messages.MsgUnauthorized)
	}
	if resp.Data != nil {
		t.Errorf("Data = %v, want nil", resp.Data)
	}
	if resp.Errors != nil {
		t.Errorf("Errors = %v, want nil", resp.Errors)
	}
	if resp.Debug != "" {
		t.Errorf("Debug = %q, want empty", resp.Debug)
	}
}
