package rotate_token

import (
	"bytes"
	"encoding/json"
	"github.com/enviniom/nexokit/internal/modules/iam/core/refresh_token"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	result *refresh_token.TokenPairResponse
	err    error
	calls  *int
}

func (f fakeService) Rotate(req refresh_token.RefreshRequest) (*refresh_token.TokenPairResponse, error) {
	if f.calls != nil {
		*f.calls++
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
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

	t.Run("returns rotated token pair", func(t *testing.T) {
		h := NewHandler(fakeService{result: &refresh_token.TokenPairResponse{AccessToken: "new-access", RefreshToken: "new-refresh"}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/auth/refresh", refresh_token.RefreshRequest{RefreshToken: "old-refresh"})

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
		c.Request = jsonRequest(http.MethodPost, "/auth/refresh", refresh_token.RefreshRequest{})

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
		c.Request = jsonRequest(http.MethodPost, "/auth/refresh", refresh_token.RefreshRequest{RefreshToken: "old-refresh"})

		h.Handle(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}

		assertUnauthorizedEnvelope(t, w)
	})
}

func assertRejectedRequest(t *testing.T, w *httptest.ResponseRecorder, calls int, keys []string) {
	t.Helper()
	var envelope response.ValidationErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(w.Body.Bytes(), &fields)
	if calls != 0 || len(fields) != 5 || envelope.Success || envelope.Data != nil || envelope.Meta != nil {
		t.Fatalf("calls=%d envelope=%s", calls, w.Body.String())
	}
	if keys == nil {
		if envelope.Message != messages.MsgBadRequest || string(fields["errors"]) != "null" {
			t.Fatalf("unexpected bad request: %s", w.Body.String())
		}
		return
	}
	if envelope.Message != messages.MsgValidationError || len(envelope.Errors) != len(keys) {
		t.Fatalf("unexpected validation: %s", w.Body.String())
	}
	for _, key := range keys {
		if len(envelope.Errors[key]) == 0 {
			t.Fatalf("missing error %q", key)
		}
	}
}

func TestHandlerMalformedJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	calls := 0
	h := NewHandler(fakeService{calls: &calls})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBufferString(`{"refresh_token":`))
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
