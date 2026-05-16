package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func jsonRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	r := httptest.NewRequest(method, path, &buf)
	r.Header.Set("Content-Type", "application/json")
	return r
}

// fakeService is a test double for the service.
type fakeService struct {
	users           []UserResponse
	user            *UserResponse
	total           int64
	err             error
	created         *UserResponse
	updated         *UserResponse
	deletedPID      string
	changePasswordErr error
	toggled         *UserResponse
}

func (f *fakeService) List(page, perPage int) ([]UserResponse, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	return f.users, f.total, nil
}

func (f *fakeService) GetByPublicID(publicID string) (*UserResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.user, nil
}

func (f *fakeService) Create(req CreateUserRequest) (*UserResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.created, nil
}

func (f *fakeService) Update(publicID string, req UpdateUserRequest) (*UserResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.updated != nil {
		return f.updated, nil
	}
	return f.user, nil
}

func (f *fakeService) Delete(publicID string) error {
	f.deletedPID = publicID
	return f.err
}

func (f *fakeService) ChangePassword(publicID string, req ChangePasswordRequest) error {
	return f.changePasswordErr
}

func (f *fakeService) ToggleStatus(publicID string, req UpdateStatusRequest) (*UserResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.toggled, nil
}

func setupHandler(svc Service) (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(svc)
	return gin.New(), h
}

func TestHandler_List(t *testing.T) {
	t.Run("returns paginated users list", func(t *testing.T) {
		svc := &fakeService{
			users: []UserResponse{
				{PublicID: "user1", Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1, RoleName: "admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{PublicID: "user2", Name: "Bob", Email: "bob@example.com", IsActive: true, RoleID: 2, RoleName: "user", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
			total: 2,
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users?page=1&per_page=10", nil)
		h.List(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[[]UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		metaMap, ok := resp.Meta.(map[string]any)
		if !ok {
			t.Fatal("expected meta to be a map")
		}
		pgRaw, ok := metaMap["pagination"]
		if !ok {
			t.Fatal("expected pagination in meta")
		}
		pgBytes, _ := json.Marshal(pgRaw)
		var pg response.PaginationMeta
		if err := json.Unmarshal(pgBytes, &pg); err != nil {
			t.Fatalf("failed to unmarshal pagination: %v", err)
		}
		if pg.Total != 2 {
			t.Errorf("expected total 2, got %d", pg.Total)
		}
		if len(resp.Data) != 2 {
			t.Errorf("expected 2 users, got %d", len(resp.Data))
		}
	})

	t.Run("returns error on service failure", func(t *testing.T) {
		svc := &fakeService{err: errors.New("boom")}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/users", nil)
		h.List(c)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected status 500, got %d", w.Code)
		}
	})
}

func TestHandler_GetByPublicID(t *testing.T) {
	t.Run("returns user when found", func(t *testing.T) {
		svc := &fakeService{
			user: &UserResponse{PublicID: "user1", Name: "Alice", Email: "alice@example.com", IsActive: true, RoleID: 1, RoleName: "admin", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		h.GetByPublicID(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		if resp.Data.PublicID != "user1" {
			t.Errorf("expected public_id 'user1', got %s", resp.Data.PublicID)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		h.GetByPublicID(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_Create(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		svc := &fakeService{
			created: &UserResponse{PublicID: "user3", Name: "Charlie", Email: "charlie@example.com", RoleID: 2, RoleName: "user"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", CreateUserRequest{Name: "Charlie", Email: "charlie@example.com", Password: "Password1", RoleID: 2})
		h.Create(c)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !resp.Success {
			t.Error("expected success response")
		}
		if resp.Data.PublicID != "user3" {
			t.Errorf("expected public_id 'user3', got %s", resp.Data.PublicID)
		}
	})

	t.Run("returns bad request on invalid body", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", map[string]string{"email": "bad"})
		h.Create(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("returns conflict on duplicate email", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrConflict}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = jsonRequest(http.MethodPost, "/users", CreateUserRequest{Name: "Charlie", Email: "charlie@example.com", Password: "Password1", RoleID: 2})
		h.Create(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})
}

func TestHandler_Update(t *testing.T) {
	t.Run("updates user successfully", func(t *testing.T) {
		svc := &fakeService{
			updated: &UserResponse{PublicID: "user1", Name: "Alice Updated", Email: "alice-new@example.com", RoleID: 2, RoleName: "user"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPut, "/users/user1", UpdateUserRequest{Name: "Alice Updated", Email: "alice-new@example.com", RoleID: 2})
		h.Update(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.Name != "Alice Updated" {
			t.Errorf("expected name 'Alice Updated', got %s", resp.Data.Name)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = jsonRequest(http.MethodPut, "/users/missing", UpdateUserRequest{Name: "Alice", Email: "alice@example.com", RoleID: 1})
		h.Update(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("returns conflict on duplicate email", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrConflict}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPut, "/users/user1", UpdateUserRequest{Name: "Alice", Email: "bob@example.com", RoleID: 1})
		h.Update(c)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409, got %d", w.Code)
		}
	})
}

func TestHandler_Delete(t *testing.T) {
	t.Run("deletes user successfully", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/users/user1", nil)
		h.Delete(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
		if svc.deletedPID != "user1" {
			t.Errorf("expected deleted public_id 'user1', got %s", svc.deletedPID)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/users/missing", nil)
		h.Delete(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_ChangePassword(t *testing.T) {
	t.Run("changes password successfully", func(t *testing.T) {
		svc := &fakeService{}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/user1/password", ChangePasswordRequest{CurrentPassword: "old", NewPassword: "NewPassword1"})
		h.ChangePassword(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("returns unauthorized on wrong password", func(t *testing.T) {
		svc := &fakeService{changePasswordErr: apperror.ErrUnauthorized}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/user1/password", ChangePasswordRequest{CurrentPassword: "wrong", NewPassword: "NewPassword1"})
		h.ChangePassword(c)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{changePasswordErr: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/missing/password", ChangePasswordRequest{CurrentPassword: "old", NewPassword: "NewPassword1"})
		h.ChangePassword(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}

func TestHandler_ToggleStatus(t *testing.T) {
	t.Run("toggles user status to inactive", func(t *testing.T) {
		svc := &fakeService{
			toggled: &UserResponse{PublicID: "user1", Name: "Alice", Email: "alice@example.com", IsActive: false, RoleID: 1, RoleName: "admin"},
		}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "user1"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/user1/status", UpdateStatusRequest{IsActive: false})
		h.ToggleStatus(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		var resp response.APIResponse[UserResponse]
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if resp.Data.IsActive {
			t.Error("expected user to be inactive")
		}
	})

	t.Run("returns not found when missing", func(t *testing.T) {
		svc := &fakeService{err: apperror.ErrNotFound}
		_, h := setupHandler(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "missing"}}
		c.Request = jsonRequest(http.MethodPatch, "/users/missing/status", UpdateStatusRequest{IsActive: false})
		h.ToggleStatus(c)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", w.Code)
		}
	})
}
