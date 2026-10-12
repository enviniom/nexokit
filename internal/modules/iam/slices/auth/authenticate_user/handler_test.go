package authenticate_user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/modules/iam/core/auth"
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type fakeService struct {
	result *user.LoginResponse
	err    error
	calls  *int
}

func (f fakeService) Login(req user.LoginRequest) (*user.LoginResponse, error) {
	if f.calls != nil {
		*f.calls++
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func authJSONRequest(method, path string, body any) *http.Request {
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

	t.Run("returns tokens and sanitized user", func(t *testing.T) {
		h := NewHandler(fakeService{result: &user.LoginResponse{AccessToken: "access", RefreshToken: "refresh", User: user.AuthUserResponse{PublicID: "user1", Email: "user@example.com", RoleName: "admin"}}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", user.LoginRequest{Email: "user@example.com", Password: "Secret1!"})

		h.Handle(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "password") || strings.Contains(w.Body.String(), "password_hash") {
			t.Fatalf("response leaked password fields: %s", w.Body.String())
		}
		var resp response.APIResponse[user.LoginResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.AccessToken != "access" || resp.Data.RefreshToken != "refresh" || resp.Data.User.Email != "user@example.com" {
			t.Fatalf("unexpected login response: %#v", resp.Data)
		}
	})

	t.Run("returns validation errors before service call", func(t *testing.T) {
		calls := 0
		h := NewHandler(fakeService{calls: &calls})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", user.LoginRequest{Email: "bad"})

		h.Handle(c)

		if w.Code != http.StatusUnprocessableEntity {
			t.Fatalf("expected status 422, got %d", w.Code)
		}
		assertRejectedRequest(t, w, calls, []string{"email", "password"})
	})

	t.Run("returns generic unauthorized for invalid credentials", func(t *testing.T) {
		h := NewHandler(fakeService{err: core.ErrInvalidCredentials})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = authJSONRequest(http.MethodPost, "/auth/login", user.LoginRequest{Email: "user@example.com", Password: "Secret1!"})

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

func TestHandlerRejectedBodies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		status     int
		keys       []string
	}{
		{"malformed JSON", `{"email":`, http.StatusBadRequest, nil},
		{"missing fields", `{}`, http.StatusUnprocessableEntity, []string{"email", "password"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			h := NewHandler(fakeService{calls: &calls})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			h.Handle(c)
			if w.Code != tc.status {
				t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
			}
			assertRejectedRequest(t, w, calls, tc.keys)
		})
	}
}

func TestHandlerOptionalUserFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uint(7)
	for _, tc := range []struct {
		name     string
		optional *uint
	}{
		{name: "absent"}, {name: "present", optional: &id},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(fakeService{result: &user.LoginResponse{User: user.AuthUserResponse{CompanyID: tc.optional, CreatedBy: tc.optional, UpdatedBy: tc.optional}}})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = authJSONRequest(http.MethodPost, "/auth/login", user.LoginRequest{Email: "user@example.com", Password: "secret"})
			h.Handle(c)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			var envelope struct {
				Data struct {
					User map[string]json.RawMessage `json:"user"`
				} `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			wantCount := 8
			if tc.optional != nil {
				wantCount += 3
			}
			if len(envelope.Data.User) != wantCount {
				t.Fatalf("unexpected user fields: %s", w.Body.String())
			}
			for _, key := range []string{"company_id", "created_by", "updated_by"} {
				value, present := envelope.Data.User[key]
				if present != (tc.optional != nil) || (present && string(value) != "7") {
					t.Errorf("%s = %s, present=%v", key, value, present)
				}
			}
		})
	}
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
