package revoke_token

import (
	"bytes"
	"encoding/json"
	"github.com/enviniom/nexokit/internal/modules/auth/core/refresh_token"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	err   error
	calls *int
}

func (f fakeService) Revoke(req refresh_token.RefreshRequest) error {
	if f.calls != nil {
		*f.calls++
	}
	return f.err
}

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
		c.Request = jsonRequest(http.MethodPost, "/auth/logout", refresh_token.RefreshRequest{RefreshToken: "refresh"})

		h.Handle(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("returns validation error", func(t *testing.T) {
		calls := 0
		h := NewHandler(fakeService{calls: &calls})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/auth/logout", refresh_token.RefreshRequest{})

		h.Handle(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}
		assertRejectedRequest(t, w, calls, []string{"refresh_token"})
	})

	t.Run("maps unauthorized error", func(t *testing.T) {
		h := NewHandler(fakeService{err: core.ErrInvalidRefreshToken})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/auth/logout", refresh_token.RefreshRequest{RefreshToken: "refresh"})

		h.Handle(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}

		assertUnauthorizedEnvelope(t, w)
	})
}

func assertRejectedRequest(t *testing.T, w *httptest.ResponseRecorder, calls int, keys []string) {
	t.Helper()
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if calls != 0 {
		t.Fatalf("service called %d times", calls)
	}
	if len(envelope) != 5 || string(envelope["success"]) != "false" || string(envelope["data"]) != "null" || string(envelope["meta"]) != "null" {
		t.Fatalf("unexpected envelope: %s", w.Body.String())
	}
	var message string
	_ = json.Unmarshal(envelope["message"], &message)
	if keys == nil {
		if message != messages.MsgBadRequest || string(envelope["errors"]) != "null" {
			t.Fatalf("unexpected bad request: %s", w.Body.String())
		}
		return
	}
	var errors map[string][]string
	if err := json.Unmarshal(envelope["errors"], &errors); err != nil {
		t.Fatal(err)
	}
	if message != messages.MsgValidationError || len(errors) != len(keys) {
		t.Fatalf("unexpected validation envelope: %s", w.Body.String())
	}
	for _, key := range keys {
		if len(errors[key]) == 0 {
			t.Fatalf("missing validation error %q: %s", key, w.Body.String())
		}
	}
}

func TestHandlerMalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	calls := 0
	h := NewHandler(fakeService{calls: &calls})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBufferString(`{"refresh_token":`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Handle(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	assertRejectedRequest(t, w, calls, nil)
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
